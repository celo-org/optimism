package engine

import (
	"context"
	"errors"
	"math/big"
	mrand "math/rand"
	gosync "sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/event"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type staleParentEngine struct {
	testutils.MockEngine
	buildAttempts int
}

func (e *staleParentEngine) ForkchoiceUpdate(context.Context, *eth.ForkchoiceState, *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	e.buildAttempts++
	return nil, eth.InputError{Inner: errors.New("missing parent header"), Code: eth.InvalidPayloadAttributes}
}

func TestInvalidPayloadDropsHead(t *testing.T) {
	emitter := &testutils.MockEmitter{}
	ec := NewEngineController(context.Background(), nil, testlog.Logger(t, 0), metrics.NoopMetrics, &rollup.Config{}, &sync.Config{}, &testutils.MockL1Source{}, emitter)

	payload := &eth.ExecutionPayloadEnvelope{ExecutionPayload: &eth.ExecutionPayload{
		BlockHash: common.Hash{0x01},
	}}

	emitter.ExpectOnce(PayloadInvalidEvent{})
	emitter.ExpectOnce(ForkchoiceUpdateEvent{})

	// Add an unsafe payload requests a forkchoice update via engine controller
	ec.AddUnsafePayload(context.Background(), payload)

	require.NotNil(t, ec.unsafePayloads.Peek())

	// Mark it invalid; it should be dropped if it matches the queue head
	ec.OnEvent(context.Background(), PayloadInvalidEvent{Envelope: payload})
	require.Nil(t, ec.unsafePayloads.Peek())
}

func TestBuildStartRejectsStaleSequencerParent(t *testing.T) {
	staleParent := eth.L2BlockRef{Hash: common.Hash{0x40}, Number: 40, Time: 80}
	currentUnsafe := eth.L2BlockRef{Hash: common.Hash{0x41}, Number: 41, Time: 82}
	safe := eth.L2BlockRef{Hash: common.Hash{0x38}, Number: 38, Time: 76}
	finalized := eth.L2BlockRef{Hash: common.Hash{0x01}, Number: 1, Time: 2}
	attributes := &derive.AttributesWithParent{
		Attributes: &eth.PayloadAttributes{
			Timestamp:    eth.Uint64Quantity(82),
			Transactions: []eth.Data{{gethtypes.DepositTxType}},
			NoTxPool:     true,
		},
		Parent: staleParent,
	}

	var emitted []event.Event
	emitter := event.EmitterFunc(func(_ context.Context, ev event.Event) {
		emitted = append(emitted, ev)
	})
	engine := &staleParentEngine{}
	ec := NewEngineController(context.Background(), engine, testlog.Logger(t, 0), metrics.NoopMetrics, &rollup.Config{}, &sync.Config{}, &testutils.MockL1Source{}, emitter)
	ec.SetUnsafeHead(currentUnsafe)
	ec.SetSafeHead(safe)
	ec.SetFinalizedHead(finalized)

	ec.OnEvent(context.Background(), BuildStartEvent{Attributes: attributes})
	for _, ev := range emitted {
		if invalid, ok := ev.(BuildInvalidEvent); ok {
			ec.OnEvent(context.Background(), invalid)
		}
	}

	require.Zero(t, engine.buildAttempts, "must not ask the engine to build on an orphaned parent")
	require.Equal(t, []event.Event{ForkchoiceUpdateEvent{
		UnsafeL2Head:    currentUnsafe,
		SafeL2Head:      safe,
		FinalizedL2Head: finalized,
	}}, emitted)
}

type sealingEngine struct {
	testutils.MockEngine
	envelope    *eth.ExecutionPayloadEnvelope
	getPayloads int
	newPayloads int
}

func (e *sealingEngine) GetPayload(context.Context, eth.PayloadInfo) (*eth.ExecutionPayloadEnvelope, error) {
	e.getPayloads++
	return e.envelope, nil
}

func (e *sealingEngine) NewPayload(context.Context, *eth.ExecutionPayload, *common.Hash) (*eth.PayloadStatusV1, error) {
	e.newPayloads++
	return &eth.PayloadStatusV1{Status: eth.ExecutionValid}, nil
}

func TestSealBuildRejectsStaleParent(t *testing.T) {
	cfg, parent, sealed, envelope := buildSimpleCfgAndPayload(t)
	movedOn := eth.L2BlockRef{Hash: common.Hash{0x41}, Number: sealed.Number, Time: sealed.Time}

	var emitted []event.Event
	emitter := event.EmitterFunc(func(_ context.Context, ev event.Event) { emitted = append(emitted, ev) })
	engine := &sealingEngine{envelope: envelope}
	ec := NewEngineController(context.Background(), engine, testlog.Logger(t, 0), metrics.NoopMetrics, cfg, &sync.Config{}, &testutils.MockL1Source{}, emitter)
	ec.SetUnsafeHead(movedOn)

	result, err := ec.SealBuild(context.Background(), eth.PayloadInfo{ID: eth.PayloadID{0x01}}, time.Now())
	require.ErrorIs(t, err, ErrStaleBuild)
	require.Nil(t, result)
	require.Equal(t, 1, engine.getPayloads, "the job can only be checked once it is sealed")
	require.Len(t, emitted, 1, "a forkchoice update is requested so the caller can re-plan")
	require.IsType(t, ForkchoiceUpdateEvent{}, emitted[0])

	emitted = nil
	ec.SetUnsafeHead(parent)
	result, err = ec.SealBuild(context.Background(), eth.PayloadInfo{ID: eth.PayloadID{0x01}}, time.Now())
	require.NoError(t, err)
	require.Equal(t, sealed.Hash, result.Ref.Hash)
	require.Empty(t, emitted, "the direct seal path emits no events on success")
}

func TestProcessPayloadRejectsStaleParent(t *testing.T) {
	cfg, _, sealed, envelope := buildSimpleCfgAndPayload(t)
	movedOn := eth.L2BlockRef{Hash: common.Hash{0x41}, Number: sealed.Number, Time: sealed.Time}

	var emitted []event.Event
	emitter := event.EmitterFunc(func(_ context.Context, ev event.Event) { emitted = append(emitted, ev) })
	engine := &sealingEngine{envelope: envelope}
	ec := NewEngineController(context.Background(), engine, testlog.Logger(t, 0), metrics.NoopMetrics, cfg, &sync.Config{}, &testutils.MockL1Source{}, emitter)
	ec.SetUnsafeHead(movedOn)

	err := ec.ProcessPayload(context.Background(), envelope, sealed, time.Now())
	require.ErrorIs(t, err, ErrStaleBuild)
	require.Zero(t, engine.newPayloads, "the stale payload must not reach the engine")
	require.Equal(t, movedOn, ec.UnsafeL2Head(), "the head must not be reorged backwards")
	require.Len(t, emitted, 1)
	require.IsType(t, ForkchoiceUpdateEvent{}, emitted[0])
}

func TestUnsafeL2HeadConcurrentAccess(t *testing.T) {
	ec := NewEngineController(context.Background(), nil, testlog.Logger(t, 0), metrics.NoopMetrics, &rollup.Config{}, &sync.Config{}, &testutils.MockL1Source{}, &testutils.MockEmitter{})
	refs := [2]eth.L2BlockRef{
		{Hash: common.Hash{0x01}, Number: 1, Time: 2},
		{Hash: common.Hash{0x02}, Number: 2, Time: 4},
	}
	ec.SetUnsafeHead(refs[0])

	start := make(chan struct{})
	var wg gosync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 10_000; i++ {
			ec.mu.Lock()
			ec.SetUnsafeHead(refs[i%len(refs)])
			ec.mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 10_000; i++ {
			_ = ec.UnsafeL2Head()
		}
	}()

	close(start)
	wg.Wait()
}

// buildSimpleCfgAndPayload creates a minimal rollup config and a valid payload (A1) on top of A0.
func buildSimpleCfgAndPayload(t *testing.T) (*rollup.Config, eth.L2BlockRef, eth.L2BlockRef, *eth.ExecutionPayloadEnvelope) {
	t.Helper()
	rng := mrand.New(mrand.NewSource(1234))
	refA := testutils.RandomBlockRef(rng)

	refA0 := eth.L2BlockRef{
		Hash:           testutils.RandomHash(rng),
		Number:         0,
		ParentHash:     common.Hash{},
		Time:           refA.Time,
		L1Origin:       refA.ID(),
		SequenceNumber: 0,
	}

	cfg := &rollup.Config{
		Genesis: rollup.Genesis{
			L1:     refA.ID(),
			L2:     refA0.ID(),
			L2Time: refA0.Time,
			SystemConfig: eth.SystemConfig{
				BatcherAddr: common.Address{42},
				Overhead:    [32]byte{123},
				Scalar:      [32]byte{42},
				GasLimit:    20_000_000,
			},
		},
		BlockTime:     1,
		SeqWindowSize: 2,
	}

	refA1 := eth.L2BlockRef{
		Hash:           testutils.RandomHash(rng),
		Number:         refA0.Number + 1,
		ParentHash:     refA0.Hash,
		Time:           refA0.Time + cfg.BlockTime,
		L1Origin:       refA.ID(),
		SequenceNumber: 1,
	}

	// Populate necessary L1 info fields
	aL1Info := &testutils.MockBlockInfo{
		InfoParentHash:  refA.ParentHash,
		InfoNum:         refA.Number,
		InfoTime:        refA.Time,
		InfoHash:        refA.Hash,
		InfoBaseFee:     big.NewInt(1),
		InfoBlobBaseFee: big.NewInt(1),
		InfoReceiptRoot: gethtypes.EmptyRootHash,
		InfoRoot:        testutils.RandomHash(rng),
		InfoGasUsed:     rng.Uint64(),
	}
	a1L1Info, err := derive.L1InfoDepositBytes(cfg, params.SepoliaChainConfig, cfg.Genesis.SystemConfig, refA1.SequenceNumber, aL1Info, refA1.Time)
	require.NoError(t, err)

	payloadA1 := &eth.ExecutionPayloadEnvelope{ExecutionPayload: &eth.ExecutionPayload{
		ParentHash:   refA1.ParentHash,
		BlockNumber:  eth.Uint64Quantity(refA1.Number),
		Timestamp:    eth.Uint64Quantity(refA1.Time),
		BlockHash:    refA1.Hash,
		Transactions: []eth.Data{a1L1Info},
	}}
	return cfg, refA0, refA1, payloadA1
}

func TestOnUnsafePayload_EnqueueEmit(t *testing.T) {
	cfg, _, _, payloadA1 := buildSimpleCfgAndPayload(t)

	emitter := &testutils.MockEmitter{}
	ec := NewEngineController(context.Background(), nil, testlog.Logger(t, 0), metrics.NoopMetrics, cfg, &sync.Config{}, &testutils.MockL1Source{}, emitter)

	emitter.ExpectOnce(PayloadInvalidEvent{})
	emitter.ExpectOnce(ForkchoiceUpdateEvent{})

	ec.AddUnsafePayload(context.Background(), payloadA1)

	got := ec.unsafePayloads.Peek()
	require.NotNil(t, got)
	require.Equal(t, payloadA1, got)
}

func TestOnForkchoiceUpdate_ProcessRetryAndPop(t *testing.T) {
	cfg, refA0, refA1, payloadA1 := buildSimpleCfgAndPayload(t)

	emitter := &testutils.MockEmitter{}
	mockEngine := &testutils.MockEngine{}
	cl := NewEngineController(context.Background(), mockEngine, testlog.Logger(t, 0), metrics.NoopMetrics, cfg, &sync.Config{SyncMode: sync.CLSync}, &testutils.MockL1Source{}, emitter)

	// queue payload A1
	emitter.ExpectOnceType("UnsafeUpdateEvent")
	emitter.ExpectOnceType("PayloadInvalidEvent")
	emitter.ExpectOnceType("ForkchoiceUpdateEvent")
	emitter.ExpectOnceType("ForkchoiceUpdateEvent")
	cl.AddUnsafePayload(context.Background(), payloadA1)

	// applicable forkchoice -> process once
	mockEngine.ExpectGetPayload(eth.PayloadID{}, payloadA1, nil)
	mockEngine.ExpectNewPayload(payloadA1.ExecutionPayload, nil, &eth.PayloadStatusV1{Status: eth.ExecutionValid}, nil)
	mockEngine.ExpectForkchoiceUpdate(&eth.ForkchoiceState{HeadBlockHash: refA1.Hash, SafeBlockHash: common.Hash{}, FinalizedBlockHash: common.Hash{}}, nil, &eth.ForkchoiceUpdatedResult{PayloadStatus: eth.PayloadStatusV1{Status: eth.ExecutionValid}}, nil)
	cl.OnEvent(context.Background(), ForkchoiceUpdateEvent{UnsafeL2Head: refA0, SafeL2Head: refA0, FinalizedL2Head: refA0})
	require.NotNil(t, cl.unsafePayloads.Peek(), "should not pop yet")

	// same forkchoice -> retry
	cl.OnEvent(context.Background(), ForkchoiceUpdateEvent{UnsafeL2Head: refA0, SafeL2Head: refA0, FinalizedL2Head: refA0})
	require.NotNil(t, cl.unsafePayloads.Peek(), "still pending")

	// after applied (unsafe head == A1) -> pop
	cl.OnEvent(context.Background(), ForkchoiceUpdateEvent{UnsafeL2Head: refA1, SafeL2Head: refA0, FinalizedL2Head: refA0})
	require.Nil(t, cl.unsafePayloads.Peek())
}

func TestPeekUnsafePayload(t *testing.T) {
	cfg, _, _, payloadA1 := buildSimpleCfgAndPayload(t)

	emitter := &testutils.MockEmitter{}
	ec := NewEngineController(context.Background(), nil, testlog.Logger(t, 0), metrics.NoopMetrics, cfg, &sync.Config{SyncMode: sync.CLSync}, &testutils.MockL1Source{}, emitter)

	// empty -> zero
	_, ref := ec.PeekUnsafePayload()
	require.Equal(t, eth.L2BlockRef{}, ref)

	// queue -> returns derived ref
	_ = ec.unsafePayloads.Push(payloadA1)
	want, err := derive.PayloadToBlockRef(cfg, payloadA1.ExecutionPayload)
	require.NoError(t, err)

	_, ref = ec.PeekUnsafePayload()
	require.Equal(t, want, ref)
}

func TestPeekUnsafePayload_OnDeriveErrorReturnsZero(t *testing.T) {
	// missing L1-info in txs will cause derive error
	emitter := &testutils.MockEmitter{}
	ec := NewEngineController(context.Background(), nil, testlog.Logger(t, 0), metrics.NoopMetrics, &rollup.Config{}, &sync.Config{SyncMode: sync.CLSync}, &testutils.MockL1Source{}, emitter)

	bad := &eth.ExecutionPayloadEnvelope{ExecutionPayload: &eth.ExecutionPayload{BlockNumber: 1, BlockHash: common.Hash{0xaa}}}
	_ = ec.unsafePayloads.Push(bad)
	_, ref := ec.PeekUnsafePayload()
	require.Equal(t, eth.L2BlockRef{}, ref)
}

func TestInvalidPayloadForNonHead_NoDrop(t *testing.T) {
	emitter := &testutils.MockEmitter{}
	ec := NewEngineController(context.Background(), nil, testlog.Logger(t, 0), metrics.NoopMetrics, &rollup.Config{}, &sync.Config{SyncMode: sync.CLSync}, &testutils.MockL1Source{}, emitter)

	// Head payload (lower block number)
	head := &eth.ExecutionPayloadEnvelope{ExecutionPayload: &eth.ExecutionPayload{
		BlockNumber: 1,
		BlockHash:   common.Hash{0x01},
	}}
	// Non-head payload (higher block number)
	other := &eth.ExecutionPayloadEnvelope{ExecutionPayload: &eth.ExecutionPayload{
		BlockNumber: 2,
		BlockHash:   common.Hash{0x02},
	}}

	emitter.ExpectOnce(PayloadInvalidEvent{})
	emitter.ExpectOnce(ForkchoiceUpdateEvent{})
	ec.AddUnsafePayload(context.Background(), head)

	emitter.ExpectOnce(PayloadInvalidEvent{})
	emitter.ExpectOnce(ForkchoiceUpdateEvent{})
	ec.AddUnsafePayload(context.Background(), other)

	// Invalidate non-head should not drop head
	ec.OnEvent(context.Background(), PayloadInvalidEvent{Envelope: other})
	require.Equal(t, 2, ec.unsafePayloads.Len())
	require.Equal(t, head, ec.unsafePayloads.Peek())
}

// note: nil-envelope behavior is not tested to match current implementation
