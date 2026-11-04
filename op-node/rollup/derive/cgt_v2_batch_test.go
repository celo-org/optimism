package derive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"testing"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

type cgtAttributesBuilderStub struct {
	active bool
	err    error
}

func (d *cgtAttributesBuilderStub) PreparePayloadAttributes(_ context.Context, parent eth.L2BlockRef, _ eth.BlockID) (*eth.PayloadAttributes, error) {
	if d.err != nil {
		return nil, d.err
	}
	attrs := &eth.PayloadAttributes{Timestamp: eth.Uint64Quantity(parent.Time + 2)}
	if d.active {
		upgrades, _, err := CgtV2UpgradeTransactions(big.NewInt(1337), big.NewInt(600))
		if err != nil {
			return nil, err
		}
		attrs.Transactions = upgrades
	}
	return attrs, nil
}

func TestCgtV2BatchRejection(t *testing.T) {
	for _, span := range []bool{false, true} {
		for _, kind := range []string{"native", "CIP-64", "empty", "post-exec"} {
			t.Run(fmt.Sprintf("span=%v/%s", span, kind), func(t *testing.T) {
				parent := eth.L2BlockRef{Hash: common.Hash{1}, Number: 100, Time: 100, L1Origin: eth.BlockID{Hash: common.Hash{2}, Number: 10}}
				b := eth.L1BlockRef{Hash: parent.L1Origin.Hash, Number: 10, Time: 88}
				h := eth.L1BlockRef{Hash: common.Hash{3}, Number: 11, ParentHash: b.Hash, Time: 100}
				cfg := &rollup.Config{L2ChainID: big.NewInt(1337), BlockTime: 2, MaxSequencerDrift: 600, SeqWindowSize: 10, DeltaTime: new(uint64), HoloceneTime: new(uint64), LagoonTime: new(uint64)}
				batch := &SingularBatch{ParentHash: parent.Hash, EpochHash: h.Hash, EpochNum: 11, Timestamp: 102}
				switch kind {
				case "native":
					batch.Transactions = []hexutil.Bytes{{0xc0}}
				case "CIP-64":
					batch.Transactions = []hexutil.Bytes{{0x7b, 0xc0}}
				case "post-exec":
					batch.Transactions = []hexutil.Bytes{{optypes.PostExecTxType, 0xc0}}
				}
				var input Batch = batch
				if span {
					sp := &SpanBatch{ChainID: cfg.L2ChainID, Batches: []*SpanBatchElement{singularBatchToElement(batch)}}
					copy(sp.ParentCheck[:], parent.Hash[:20])
					copy(sp.L1OriginCheck[:], h.Hash[:20])
					input = sp
				}
				prev := &fakeBatchQueueInput{batches: []Batch{input}, errors: []error{nil}, origin: h}
				builder := &cgtAttributesBuilderStub{active: true}
				mux := NewBatchMux(testlog.Logger(t, log.LevelError), cfg, prev, newFakeSafeBlockFetcher())
				require.ErrorIs(t, mux.Reset(context.Background(), b, eth.SystemConfig{}), io.EOF)
				queue := NewAttributesQueue(testlog.Logger(t, log.LevelError), cfg, builder, mux)
				got, err := queue.NextAttributes(context.Background(), parent)
				if kind != "empty" {
					require.ErrorIs(t, err, NotEnoughData)
					require.Nil(t, got)
					require.Empty(t, prev.batches)
				} else {
					require.NoError(t, err)
					require.NotNil(t, got)
				}
			})
		}
	}
}

func TestCgtV2BatchRPCErrorRetriesCandidate(t *testing.T) {
	parent := eth.L2BlockRef{Hash: common.Hash{1}, Number: 100, Time: 100, L1Origin: eth.BlockID{Hash: common.Hash{2}, Number: 10}}
	b := eth.L1BlockRef{Hash: parent.L1Origin.Hash, Number: 10, Time: 88}
	h := eth.L1BlockRef{Hash: common.Hash{3}, Number: 11, Time: 100}
	cfg := &rollup.Config{BlockTime: 2, MaxSequencerDrift: 600, SeqWindowSize: 10, HoloceneTime: new(uint64)}
	batch := &SingularBatch{ParentHash: parent.Hash, EpochHash: h.Hash, EpochNum: 11, Timestamp: 102, Transactions: []hexutil.Bytes{{0xc0}}}
	prev := &fakeBatchQueueInput{batches: []Batch{batch}, errors: []error{nil}, origin: h}
	builder := &cgtAttributesBuilderStub{err: NewTemporaryError(errors.New("offline"))}
	mux := NewBatchMux(testlog.Logger(t, log.LevelError), cfg, prev, nil)
	require.ErrorIs(t, mux.Reset(context.Background(), b, eth.SystemConfig{}), io.EOF)
	queue := NewAttributesQueue(testlog.Logger(t, log.LevelError), cfg, builder, mux)
	_, err := queue.NextAttributes(context.Background(), parent)
	require.ErrorIs(t, err, ErrTemporary)
	builder.err = nil
	got, err := queue.NextAttributes(context.Background(), parent)
	require.NoError(t, err)
	require.Equal(t, batch.Transactions, got.Attributes.Transactions)
	require.Equal(t, 1, prev.i, "retry must not read another batch")
}

func TestCgtV2PipelineConstructorWiring(t *testing.T) {
	rs := cgtSeededReceipts()
	header, parent, origin := cgtTestOrigin(t, rs)
	cfg := &rollup.Config{BlockTime: 2, L1ChainID: big.NewInt(1), L2ChainID: big.NewInt(1337), L1SystemConfigAddress: common.Address{0x33}, DepositContractAddress: common.Address{0x44}}
	l1 := &testutils.MockL1Source{}
	l1.ExpectFetchReceipts(origin.Hash, eth.HeaderBlockInfo(header), rs, nil)
	l2 := &testutils.MockL2Client{}
	l2.ExpectSystemConfigByL2Hash(parent.Hash, eth.SystemConfig{GasLimit: 30000000}, nil)
	pipeline := NewDerivationPipeline(testlog.Logger(t, log.LevelError), cfg, nil, l1, nil, nil, l2, nil, params.MergedTestChainConfig)
	batch := &SingularBatch{ParentHash: parent.Hash, EpochHash: origin.Hash, EpochNum: rollup.Epoch(origin.Number), Timestamp: 102}
	attrs, err := pipeline.attrib.createNextAttributes(context.Background(), batch, parent)
	require.NoError(t, err)
	require.True(t, ContainsCgtV2Activation(attrs.Transactions))
	require.True(t, attrs.NoTxPool)
	pipeline.attrib.reset()
	l1.AssertExpectations(t)
	l2.AssertExpectations(t)
}

func TestCgtV2SpanPreservesValidPrefix(t *testing.T) {
	parent := eth.L2BlockRef{Hash: common.Hash{1}, Number: 100, Time: 100, L1Origin: eth.BlockID{Hash: common.Hash{2}, Number: 10}}
	b := eth.L1BlockRef{Hash: parent.L1Origin.Hash, Number: 10, Time: 88}
	h := eth.L1BlockRef{Hash: common.Hash{3}, Number: 11, Time: 100}
	cfg := &rollup.Config{L2ChainID: big.NewInt(1337), BlockTime: 2, MaxSequencerDrift: 600, SeqWindowSize: 10, DeltaTime: new(uint64), HoloceneTime: new(uint64)}
	sp := &SpanBatch{ChainID: cfg.L2ChainID, Batches: []*SpanBatchElement{
		{EpochNum: 10, Timestamp: 102, Transactions: []hexutil.Bytes{{0xc0}}},
		{EpochNum: 11, Timestamp: 104, Transactions: []hexutil.Bytes{{0xc0}}},
	}}
	copy(sp.ParentCheck[:], parent.Hash[:20])
	copy(sp.L1OriginCheck[:], h.Hash[:20])
	prev := &fakeBatchQueueInput{batches: []Batch{sp}, errors: []error{nil}, origin: h}
	builder := &cgtAttributesBuilderStub{}
	mux := NewBatchMux(testlog.Logger(t, log.LevelError), cfg, prev, nil)
	require.ErrorIs(t, mux.Reset(context.Background(), b, eth.SystemConfig{}), io.EOF)
	queue := NewAttributesQueue(testlog.Logger(t, log.LevelError), cfg, builder, mux)
	first, err := queue.NextAttributes(context.Background(), parent)
	require.NoError(t, err)
	require.EqualValues(t, 102, first.Attributes.Timestamp)
	builder.active = true
	child := eth.L2BlockRef{Hash: common.Hash{4}, ParentHash: parent.Hash, Number: 101, Time: 102, L1Origin: b.ID()}
	next, err := queue.NextAttributes(context.Background(), child)
	require.ErrorIs(t, err, NotEnoughData)
	require.Nil(t, next)
	require.EqualValues(t, 102, first.Attributes.Timestamp, "dropping the activation block must preserve the accepted prefix")
	require.Empty(t, mux.SingularBatchProvider.(*BatchStage).nextSpan)
	retry, err := queue.DepositsOnlyAttributes(parent.ID(), first.DerivedFrom)
	require.NoError(t, err, "rejecting a later batch must preserve the last returned attributes")
	require.EqualValues(t, 102, retry.Attributes.Timestamp)
}

func TestBatchMuxFlushPreservesLegacyQueuedBatches(t *testing.T) {
	queue := &BatchQueue{batches: []*BatchWithL1InclusionBlock{{}}}
	queue.nextSpan = []*SingularBatch{{Timestamp: 102}, {Timestamp: 104}}
	mux := &BatchMux{SingularBatchProvider: queue}
	mux.FlushChannel()
	require.Empty(t, queue.nextSpan)
	require.Len(t, queue.batches, 1)
}
