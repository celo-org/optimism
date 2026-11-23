package sequencing

import (
	"context"
	"encoding/binary"
	"fmt"
	"maps"
	"math/big"
	"math/rand" // nosemgrep
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/attributes"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/rollup/engine"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/event"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

// TestCgtV2ActivationBlockConsolidation runs derivation over a CGTv2 activation block that the
// sequencer built and gossiped as celo-reth's deposits-only fallback. The derived attributes
// carry the upgrade group, so they never match that block, and op-node rebuilds the block on
// its parent. The real attributes handler, engine controller and sequencer run against an EL
// that follows reth's forkchoice rules and celo-reth's payload builder.
func TestCgtV2ActivationBlockConsolidation(t *testing.T) {
	// AttributesMatchBlock logs the transactions that differ, and the attributes handler logs
	// the mismatch that sends the block to be rebuilt.
	mismatch := []string{"mismatched hashes", "L2 reorg: existing unsafe block does not match derived attributes from L1"}

	t.Run("identical rebuild keeps the unsafe chain", func(t *testing.T) {
		h := newCgtV2Consolidation(t, false)
		head := h.head()

		h.deriveActivationBlock()

		rebuild := forkchoiceCall{
			state:     eth.ForkchoiceState{HeadBlockHash: h.parent.Hash, SafeBlockHash: h.parent.Hash, FinalizedBlockHash: h.parent.Hash},
			withAttrs: true,
		}
		require.Equal(t, []forkchoiceCall{rebuild}, h.el.forkchoiceCalls, "the EL only rebuilt the activation block on its parent")
		require.Equal(t, head, h.ec.UnsafeL2Head(), "unsafe chain kept")
		require.Equal(t, h.activation, h.ec.PendingSafeL2Head())
		require.Equal(t, h.activation, h.ec.SafeL2Head())
		require.Equal(t, head, h.announcedHeads[len(h.announcedHeads)-1], "listeners end on the kept head")
		require.Equal(t, head.Hash, h.el.tip)

		// Without gossip, the new safe head reaches the EL with the next batched forkchoice
		// update, which names the kept head.
		h.derivation.Emit(h.ctx, derive.DeriverIdleEvent{})
		h.drain()
		require.Equal(t, []forkchoiceCall{rebuild, {state: eth.ForkchoiceState{
			HeadBlockHash:      head.Hash,
			SafeBlockHash:      h.activation.Hash,
			FinalizedBlockHash: h.parent.Hash,
		}}}, h.el.forkchoiceCalls)
		require.Equal(t, head.Hash, h.el.tip)

		// The rest of the batch consolidates against the kept chain without another rebuild.
		h.deriveDescendant(0, h.activation)
		require.Equal(t, 1, h.el.payloadJobs())
		require.Equal(t, h.ref(h.descendants[0]), h.ec.SafeL2Head())
		require.Equal(t, head, h.ec.UnsafeL2Head())

		// The sequencer's next block applies directly on the kept head, and the forkchoice
		// update that makes it canonical carries the new safe head to the EL.
		next := h.el.build(head.Hash, h.nextAttributes(head))
		h.ec.AddUnsafePayload(h.ctx, next)
		h.drain()
		require.Equal(t, h.ref(next), h.ec.UnsafeL2Head())
		require.Equal(t, eth.ForkchoiceState{
			HeadBlockHash:      next.ExecutionPayload.BlockHash,
			SafeBlockHash:      h.descendants[0].ExecutionPayload.BlockHash,
			FinalizedBlockHash: h.parent.Hash,
		}, h.el.forkchoiceCalls[len(h.el.forkchoiceCalls)-1].state)
		require.Empty(t, h.el.reorgDepths)
		h.requireProblemLogs(mismatch...)
	})

	t.Run("sequencer continues on the kept head", func(t *testing.T) {
		h := newCgtV2Consolidation(t, true)
		head := h.head()
		// The sequencer is building on the head when derivation reaches the activation block.
		h.seq.RunAction()
		h.drain()
		require.Equal(t, head, h.seq.building.Onto)

		h.deriveActivationBlock()

		// Replaying its inbox, the sequencer drops its job when the rebuild announces the
		// activation block's parent as head, pauses for the derived build, and then resumes
		// on the announced kept head.
		h.seq.drainInbox()
		_, armed := h.seq.NextAction()
		require.True(t, armed)
		require.Equal(t, head, h.seq.unsafeHead)
		h.seq.RunAction()
		h.drain()
		require.Equal(t, head, h.seq.building.Onto)
		h.seq.RunAction()
		h.drain()
		require.NotNil(t, h.gossip.payload, "the sequencer sealed, inserted and gossiped its next block")
		require.Equal(t, head.Hash, h.gossip.payload.ExecutionPayload.ParentHash)
		require.Equal(t, h.gossip.payload.ExecutionPayload.BlockHash, h.ec.UnsafeL2Head().Hash)
		require.Empty(t, h.el.reorgDepths, "no gossiped block was reorged out")
		h.requireProblemLogs(append(mismatch, "Detected new block-building from L1 derivation, avoiding sequencing for now.")...)
	})

	t.Run("rebuild with the upgrade replaces the unsafe chain", func(t *testing.T) {
		h := newCgtV2Consolidation(t, false)
		h.el.upgradeSucceeds = true

		h.deriveActivationBlock()

		upgraded := h.ec.UnsafeL2Head()
		require.Equal(t, h.activation.Number, upgraded.Number)
		require.NotEqual(t, h.activation.Hash, upgraded.Hash)
		require.Equal(t, upgraded, h.ec.SafeL2Head())
		require.Len(t, h.el.blocks[upgraded.Hash].ExecutionPayload.Transactions, len(h.activationAttrs.Transactions),
			"the derived block carries the upgrade group")
		require.Equal(t, upgraded.Hash, h.el.tip)
		require.Equal(t, []uint64{cgtV2Descendants + 1}, h.el.reorgDepths,
			"the fallback block and its descendants are reorged out")
		h.requireProblemLogs(mismatch...)
	})
}

// cgtV2ChainID selects the CGTv2 upgrade bundle the tests derive.
var cgtV2ChainID = big.NewInt(1337)

// cgtV2Descendants is how many gossiped blocks the unsafe chain holds above the activation block.
const cgtV2Descendants = 5

// rethEngine is an in-memory EL with reth's forkchoice rules and celo-reth's CGTv2 payload
// builder.
//
// Forkchoice follows reth's engine tree for OP stack chains with the default
// allow_unwind_canonical_header=false (crates/engine/tree/src/tree/mod.rs):
//   - an update to a canonical ancestor keeps the canonical tip, and with attributes starts a
//     payload job on that ancestor;
//   - an update to a known block off the canonical chain reorgs the canonical chain to it;
//   - an update whose safe or finalized block would not be canonical afterwards is refused
//     with InvalidForkchoiceState, before anything changes;
//   - newPayload of a known block reports VALID.
//
// The builder seals the attributes' transactions plus, unless NoTxPool is set, one mempool
// transaction. Attributes that carry the CGTv2 upgrade group are sealed as celo-reth seals a
// failed upgrade: without the group, and with the upgrade gas allowance taken off the gas
// limit. With upgradeSucceeds set, the group is sealed as given.
type rethEngine struct {
	t               *testing.T
	cfg             *rollup.Config
	rng             *rand.Rand
	groupLen        int
	allowance       uint64
	upgradeSucceeds bool

	blocks    map[common.Hash]*eth.ExecutionPayloadEnvelope
	canonical map[uint64]common.Hash
	tip       common.Hash

	jobs    map[eth.PayloadID]*eth.ExecutionPayloadEnvelope
	nextJob uint64

	forkchoiceCalls []forkchoiceCall
	// reorgDepths holds, per reorg, how many canonical blocks it removed.
	reorgDepths []uint64
}

type forkchoiceCall struct {
	state     eth.ForkchoiceState
	withAttrs bool
}

func (r *rethEngine) number(hash common.Hash) uint64 {
	return uint64(r.blocks[hash].ExecutionPayload.BlockNumber)
}

// appendCanonical stores a block as the new canonical tip.
func (r *rethEngine) appendCanonical(env *eth.ExecutionPayloadEnvelope) {
	hash := env.ExecutionPayload.BlockHash
	r.blocks[hash] = env
	r.canonical[uint64(env.ExecutionPayload.BlockNumber)] = hash
	r.tip = hash
}

func (r *rethEngine) payloadJobs() int {
	jobs := 0
	for _, call := range r.forkchoiceCalls {
		if call.withAttrs {
			jobs++
		}
	}
	return jobs
}

func (r *rethEngine) build(parentHash common.Hash, attrs *eth.PayloadAttributes) *eth.ExecutionPayloadEnvelope {
	txs := append([]eth.Data(nil), attrs.Transactions...)
	gasLimit := uint64(*attrs.GasLimit)
	if derive.ContainsCgtV2Activation(txs) && !r.upgradeSucceeds {
		// The group closes the transaction list, after the deposits the block keeps.
		txs = txs[:len(txs)-r.groupLen]
		gasLimit -= r.allowance
	}
	if !attrs.NoTxPool {
		tx, err := testutils.RandomLegacyTxNotProtected(r.rng).MarshalBinary()
		require.NoError(r.t, err)
		txs = append(txs, tx)
	}
	number := r.number(parentHash) + 1
	parts := [][]byte{
		parentHash[:],
		binary.BigEndian.AppendUint64(nil, number),
		binary.BigEndian.AppendUint64(nil, uint64(attrs.Timestamp)),
		binary.BigEndian.AppendUint64(nil, gasLimit),
		attrs.PrevRandao[:],
		attrs.SuggestedFeeRecipient[:],
	}
	for _, tx := range txs {
		parts = append(parts, crypto.Keccak256(tx))
	}
	return &eth.ExecutionPayloadEnvelope{
		ParentBeaconBlockRoot: attrs.ParentBeaconBlockRoot,
		ExecutionPayload: &eth.ExecutionPayload{
			ParentHash:   parentHash,
			FeeRecipient: attrs.SuggestedFeeRecipient,
			PrevRandao:   attrs.PrevRandao,
			BlockNumber:  eth.Uint64Quantity(number),
			GasLimit:     eth.Uint64Quantity(gasLimit),
			Timestamp:    attrs.Timestamp,
			Transactions: txs,
			Withdrawals:  attrs.Withdrawals,
			BlockHash:    crypto.Keccak256Hash(parts...),
		},
	}
}

func (r *rethEngine) ForkchoiceUpdate(_ context.Context, state *eth.ForkchoiceState, attrs *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	r.forkchoiceCalls = append(r.forkchoiceCalls, forkchoiceCall{state: *state, withAttrs: attrs != nil})
	if _, ok := r.blocks[state.HeadBlockHash]; !ok {
		return &eth.ForkchoiceUpdatedResult{PayloadStatus: eth.PayloadStatusV1{Status: eth.ExecutionSyncing}}, nil
	}
	canonical, removed := r.canonicalAfter(state.HeadBlockHash)
	for _, hash := range []common.Hash{state.SafeBlockHash, state.FinalizedBlockHash} {
		if hash == (common.Hash{}) {
			continue // a zero hash leaves that label unset
		}
		if _, known := r.blocks[hash]; !known || canonical[r.number(hash)] != hash {
			return nil, eth.InputError{
				Inner: fmt.Errorf("forkchoice block %s is not canonical", hash),
				Code:  eth.InvalidForkchoiceState,
			}
		}
	}
	if r.canonical[r.number(state.HeadBlockHash)] != state.HeadBlockHash {
		r.tip = state.HeadBlockHash
	}
	r.canonical = canonical
	if removed > 0 {
		r.reorgDepths = append(r.reorgDepths, removed)
	}
	result := &eth.ForkchoiceUpdatedResult{PayloadStatus: eth.PayloadStatusV1{Status: eth.ExecutionValid}}
	if attrs != nil {
		r.nextJob++
		var id eth.PayloadID
		binary.BigEndian.PutUint64(id[:], r.nextJob)
		r.jobs[id] = r.build(state.HeadBlockHash, attrs)
		result.PayloadID = &id
	}
	return result, nil
}

// canonicalAfter returns the canonical chain after a forkchoice update to head, and how many
// canonical blocks the update removes. A canonical head leaves the chain as it is; any other
// head replaces the canonical blocks above its fork point.
func (r *rethEngine) canonicalAfter(head common.Hash) (map[uint64]common.Hash, uint64) {
	if r.canonical[r.number(head)] == head {
		return r.canonical, 0
	}
	canonical := maps.Clone(r.canonical)
	fork := head
	for r.canonical[r.number(fork)] != fork {
		canonical[r.number(fork)] = fork
		fork = r.blocks[fork].ExecutionPayload.ParentHash
	}
	oldTip := r.number(r.tip)
	for n := r.number(head) + 1; n <= oldTip; n++ {
		delete(canonical, n)
	}
	return canonical, oldTip - r.number(fork)
}

func (r *rethEngine) GetPayload(_ context.Context, info eth.PayloadInfo) (*eth.ExecutionPayloadEnvelope, error) {
	env, ok := r.jobs[info.ID]
	if !ok {
		return nil, fmt.Errorf("unknown payload %s", info.ID)
	}
	return env, nil
}

func (r *rethEngine) NewPayload(_ context.Context, payload *eth.ExecutionPayload, _ *common.Hash) (*eth.PayloadStatusV1, error) {
	if _, ok := r.blocks[payload.BlockHash]; ok {
		return &eth.PayloadStatusV1{Status: eth.ExecutionValid}, nil
	}
	if _, ok := r.blocks[payload.ParentHash]; !ok {
		return &eth.PayloadStatusV1{Status: eth.ExecutionSyncing}, nil
	}
	r.blocks[payload.BlockHash] = &eth.ExecutionPayloadEnvelope{ExecutionPayload: payload}
	return &eth.PayloadStatusV1{Status: eth.ExecutionValid}, nil
}

func (r *rethEngine) ref(hash common.Hash) (eth.L2BlockRef, error) {
	env, ok := r.blocks[hash]
	if !ok {
		return eth.L2BlockRef{}, ethereum.NotFound
	}
	return derive.PayloadToBlockRef(r.cfg, env.ExecutionPayload)
}

// L2BlockRefByLabel is not modelled: the tests set every head on the engine controller.
func (r *rethEngine) L2BlockRefByLabel(_ context.Context, label eth.BlockLabel) (eth.L2BlockRef, error) {
	r.t.Fatalf("unexpected head lookup by label %s", label)
	return eth.L2BlockRef{}, nil
}

func (r *rethEngine) L2BlockRefByHash(_ context.Context, hash common.Hash) (eth.L2BlockRef, error) {
	return r.ref(hash)
}

func (r *rethEngine) L2BlockRefByNumber(_ context.Context, number uint64) (eth.L2BlockRef, error) {
	hash, ok := r.canonical[number]
	if !ok {
		return eth.L2BlockRef{}, ethereum.NotFound
	}
	return r.ref(hash)
}

func (r *rethEngine) PayloadByNumber(_ context.Context, number uint64) (*eth.ExecutionPayloadEnvelope, error) {
	hash, ok := r.canonical[number]
	if !ok {
		return nil, ethereum.NotFound
	}
	return r.blocks[hash], nil
}

// l1InfoAttributes prepares payload attributes with a real L1-info deposit, so that op-node
// can read the L1 origin back from the blocks built from them.
type l1InfoAttributes struct {
	cfg        *rollup.Config
	sysCfg     eth.SystemConfig
	l1         map[common.Hash]*testutils.MockBlockInfo
	beaconRoot common.Hash
}

func (b *l1InfoAttributes) PreparePayloadAttributes(_ context.Context, parent eth.L2BlockRef, epoch eth.BlockID) (*eth.PayloadAttributes, error) {
	seqNumber := uint64(0)
	if parent.L1Origin.Hash == epoch.Hash {
		seqNumber = parent.SequenceNumber + 1
	}
	info := b.l1[epoch.Hash]
	timestamp := parent.Time + b.cfg.BlockTime
	l1InfoTx, err := derive.L1InfoDepositBytes(b.cfg, params.MergedTestChainConfig, b.sysCfg, seqNumber, info, timestamp)
	if err != nil {
		return nil, err
	}
	gasLimit := eth.Uint64Quantity(b.sysCfg.GasLimit)
	withdrawals := make(types.Withdrawals, 0)
	beaconRoot := b.beaconRoot
	return &eth.PayloadAttributes{
		Timestamp:             eth.Uint64Quantity(timestamp),
		PrevRandao:            eth.Bytes32(info.MixDigest()),
		SuggestedFeeRecipient: predeploys.SequencerFeeVaultAddr,
		Withdrawals:           &withdrawals,
		ParentBeaconBlockRoot: &beaconRoot,
		Transactions:          []eth.Data{l1InfoTx},
		GasLimit:              &gasLimit,
	}, nil
}

// cgtV2Consolidation is a node at the moment derivation reaches a CGTv2 activation block that
// celo-reth sealed as a deposits-only fallback, with gossiped blocks already on top of it. Its
// components are wired as op-node's driver wires them.
type cgtV2Consolidation struct {
	t     *testing.T
	ctx   context.Context
	cfg   *rollup.Config
	el    *rethEngine
	ec    *engine.EngineController
	exec  *event.GlobalSyncExec
	logs  *testlog.CapturingHandler
	attrs *l1InfoAttributes
	// derivation emits derived attributes in place of the derivation pipeline.
	derivation event.Emitter

	seq    *Sequencer
	gossip *FakeAsyncGossip

	epoch       eth.L1BlockRef
	derivedFrom eth.L1BlockRef
	// parent is the safe head, and activation the gossiped fallback block on top of it.
	parent, activation eth.L2BlockRef
	// activationAttrs are the attributes derivation produces for the activation block.
	activationAttrs *eth.PayloadAttributes
	// descendants are the gossiped blocks above the activation block, oldest first.
	descendants []*eth.ExecutionPayloadEnvelope
	// announcedHeads are the unsafe heads of every ForkchoiceUpdateEvent, in order.
	announcedHeads []eth.L2BlockRef
}

func newCgtV2Consolidation(t *testing.T, sequencing bool) *cgtV2Consolidation {
	ctx := context.Background()
	logger, logs := testlog.CaptureLogger(t, log.LevelInfo)

	prevEpoch := eth.L1BlockRef{Hash: common.Hash{0xe0}, Number: 100, Time: 1_000_000}
	epoch := eth.L1BlockRef{Hash: common.Hash{0xe1}, Number: 101, ParentHash: prevEpoch.Hash, Time: prevEpoch.Time + 12}
	derivedFrom := eth.L1BlockRef{Hash: common.Hash{0xe9}, Number: 130, Time: epoch.Time + 29*12}
	l1Info := func(ref eth.L1BlockRef) *testutils.MockBlockInfo {
		return &testutils.MockBlockInfo{
			InfoHash: ref.Hash, InfoParentHash: ref.ParentHash, InfoNum: ref.Number, InfoTime: ref.Time,
			InfoBaseFee: big.NewInt(7), InfoBlobBaseFee: big.NewInt(1),
			InfoReceiptRoot: types.EmptyRootHash, InfoMixDigest: [32]byte{byte(ref.Number)},
		}
	}
	sysCfg := eth.SystemConfig{
		BatcherAddr: common.Address{42},
		Overhead:    [32]byte{31: 123},
		Scalar:      [32]byte{31: 42},
		GasLimit:    30_000_000,
	}
	cfg := &rollup.Config{
		Genesis: rollup.Genesis{
			L1:           prevEpoch.ID(),
			L2:           eth.BlockID{Hash: common.Hash{0x99}, Number: 0},
			L2Time:       prevEpoch.Time - 100,
			SystemConfig: sysCfg,
		},
		BlockTime:         1,
		SeqWindowSize:     3600,
		MaxSequencerDrift: 600,
		L2ChainID:         cgtV2ChainID,
		RegolithTime:      new(uint64),
		CanyonTime:        new(uint64),
		EcotoneTime:       new(uint64),
	}
	group, _, err := derive.CgtV2UpgradeTransactions(cgtV2ChainID, big.NewInt(600))
	require.NoError(t, err)
	allowance, err := derive.CgtV2UpgradeGas(cgtV2ChainID)
	require.NoError(t, err)

	el := &rethEngine{
		t:         t,
		cfg:       cfg,
		rng:       rand.New(rand.NewSource(7)),
		groupLen:  len(group),
		allowance: allowance,
		blocks:    map[common.Hash]*eth.ExecutionPayloadEnvelope{},
		canonical: map[uint64]common.Hash{},
		jobs:      map[eth.PayloadID]*eth.ExecutionPayloadEnvelope{},
	}
	attrs := &l1InfoAttributes{
		cfg:        cfg,
		sysCfg:     sysCfg,
		l1:         map[common.Hash]*testutils.MockBlockInfo{prevEpoch.Hash: l1Info(prevEpoch), epoch.Hash: l1Info(epoch)},
		beaconRoot: common.Hash{0xbe},
	}
	h := &cgtV2Consolidation{t: t, ctx: ctx, cfg: cfg, el: el, logs: logs, attrs: attrs, epoch: epoch, derivedFrom: derivedFrom}

	// The safe parent closes the previous epoch. A stub block stands in for the chain below it.
	stub := &eth.ExecutionPayloadEnvelope{ExecutionPayload: &eth.ExecutionPayload{BlockNumber: 9, BlockHash: common.Hash{0x09}}}
	el.appendCanonical(stub)
	stubRef := eth.L2BlockRef{Hash: stub.ExecutionPayload.BlockHash, Number: 9, Time: epoch.Time - 2, L1Origin: prevEpoch.ID(), SequenceNumber: 5}
	parentAttrs, err := attrs.PreparePayloadAttributes(ctx, stubRef, prevEpoch.ID())
	require.NoError(t, err)
	parentAttrs.NoTxPool = true
	parentEnv := el.build(stubRef.Hash, parentAttrs)
	el.appendCanonical(parentEnv)
	h.parent = h.ref(parentEnv)

	// The activation block opens the new epoch. Derivation gives it an origin deposit, the
	// upgrade group and a gas limit raised by the upgrade allowance. The sequencer built it
	// from the same attributes and celo-reth sealed the fallback.
	activationAttrs, err := attrs.PreparePayloadAttributes(ctx, h.parent, epoch.ID())
	require.NoError(t, err)
	to := common.Address{0xde}
	deposit, err := types.NewTx(&types.DepositTx{
		SourceHash: common.Hash{0x11}, From: common.Address{0x22}, To: &to, Gas: 50_000,
	}).MarshalBinary()
	require.NoError(t, err)
	activationAttrs.Transactions = append(activationAttrs.Transactions, deposit)
	activationAttrs.Transactions = append(activationAttrs.Transactions, group...)
	raised := eth.Uint64Quantity(sysCfg.GasLimit + allowance)
	activationAttrs.GasLimit = &raised
	activationAttrs.NoTxPool = true
	activationEnv := el.build(h.parent.Hash, activationAttrs)
	el.appendCanonical(activationEnv)
	h.activation = h.ref(activationEnv)
	h.activationAttrs = activationAttrs

	head := h.activation
	for i := 0; i < cgtV2Descendants; i++ {
		env := el.build(head.Hash, h.nextAttributes(head))
		el.appendCanonical(env)
		h.descendants = append(h.descendants, env)
		head = h.ref(env)
	}

	exec := event.NewGlobalSynchronous(ctx)
	sys := event.NewSystem(logger, exec)
	ec := engine.NewEngineController(ctx, el, logger, metrics.NoopMetrics, cfg, &sync.Config{},
		&testutils.MockL1Source{}, sys.Register("engine-controller", nil), nil)
	sys.Register("attributes-handler", attributes.NewAttributesHandler(logger, cfg, ctx, el, ec))
	sys.Register("engine", ec)
	sys.Register("forkchoice-listener", event.DeriverFunc(func(_ context.Context, ev event.Event) bool {
		if fc, ok := ev.(engine.ForkchoiceUpdateEvent); ok {
			h.announcedHeads = append(h.announcedHeads, fc.UnsafeL2Head)
		}
		return false
	}))
	h.derivation = sys.Register("derivation", nil)
	h.ec, h.exec = ec, exec

	ec.SetUnsafeHead(head)
	ec.SetPendingSafeL2Head(h.parent)
	ec.SetLocalSafeHead(h.parent)
	ec.SetFinalizedHead(h.parent)

	if sequencing {
		h.gossip = &FakeAsyncGossip{}
		origin := &FakeL1OriginSelector{l1OriginFn: func(eth.L2BlockRef) (eth.L1BlockRef, error) { return epoch, nil }}
		h.seq = NewSequencer(ctx, logger, cfg, defaultSealingDuration, attrs, origin, &BasicSequencerStateListener{},
			&FakeConductor{leader: true}, h.gossip, metrics.NoopMetrics, ec)
		// The next block is due now.
		now := time.Unix(int64(head.Time+cfg.BlockTime), 0)
		h.seq.timeNow = func() time.Time { return now }
		sys.Register("sequencer", h.seq)
		require.NoError(t, h.seq.Init(ctx, true))
	}
	h.drain()
	return h
}

func (h *cgtV2Consolidation) drain() {
	require.NoError(h.t, h.exec.Drain())
}

func (h *cgtV2Consolidation) ref(env *eth.ExecutionPayloadEnvelope) eth.L2BlockRef {
	ref, err := derive.PayloadToBlockRef(h.cfg, env.ExecutionPayload)
	require.NoError(h.t, err)
	return ref
}

// head is the tip of the gossiped chain.
func (h *cgtV2Consolidation) head() eth.L2BlockRef {
	return h.ref(h.descendants[len(h.descendants)-1])
}

// nextAttributes are the sequencer's attributes for an ordinary block on top of parent.
func (h *cgtV2Consolidation) nextAttributes(parent eth.L2BlockRef) *eth.PayloadAttributes {
	attrs, err := h.attrs.PreparePayloadAttributes(h.ctx, parent, h.epoch.ID())
	require.NoError(h.t, err)
	return attrs
}

func (h *cgtV2Consolidation) deriveActivationBlock() {
	h.deriveBlock(h.activationAttrs, h.parent)
}

// deriveDescendant derives descendant i from the batch that carries it, with the same
// transactions the sequencer included.
func (h *cgtV2Consolidation) deriveDescendant(i int, parent eth.L2BlockRef) {
	env := h.descendants[i]
	payload := env.ExecutionPayload
	gasLimit := payload.GasLimit
	beaconRoot := *env.ParentBeaconBlockRoot
	h.deriveBlock(&eth.PayloadAttributes{
		Timestamp:             payload.Timestamp,
		PrevRandao:            payload.PrevRandao,
		SuggestedFeeRecipient: payload.FeeRecipient,
		Withdrawals:           payload.Withdrawals,
		ParentBeaconBlockRoot: &beaconRoot,
		Transactions:          append([]eth.Data(nil), payload.Transactions...),
		NoTxPool:              true,
		GasLimit:              &gasLimit,
	}, parent)
}

// deriveBlock hands derived attributes to the attributes handler and runs every resulting event.
func (h *cgtV2Consolidation) deriveBlock(attrs *eth.PayloadAttributes, parent eth.L2BlockRef) {
	h.derivation.Emit(h.ctx, derive.DerivedAttributesEvent{Attributes: &derive.AttributesWithParent{
		Attributes:  attrs,
		Parent:      parent,
		Concluding:  true,
		DerivedFrom: h.derivedFrom,
	}})
	h.drain()
}

// requireProblemLogs checks that the node logged exactly these warnings and errors, in order.
func (h *cgtV2Consolidation) requireProblemLogs(messages ...string) {
	var logged []string
	for _, record := range h.logs.FindLogs(func(r *testlog.CapturedRecord) bool { return r.Record.Level >= log.LevelWarn }) {
		logged = append(logged, record.Message)
	}
	require.Equal(h.t, messages, logged)
}
