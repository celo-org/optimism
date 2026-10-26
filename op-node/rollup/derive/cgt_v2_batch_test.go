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

type cgtDetectorStub struct {
	active bool
	err    error
	calls  int
}

func (d *cgtDetectorStub) CgtV2ActivationBlock(context.Context, eth.L2BlockRef, eth.BlockID) (bool, error) {
	d.calls++
	return d.active, d.err
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
				detector := &cgtDetectorStub{active: true}
				mux := NewBatchMux(testlog.Logger(t, log.LevelError), cfg, prev, newFakeSafeBlockFetcher(), detector)
				require.ErrorIs(t, mux.Reset(context.Background(), b, eth.SystemConfig{}), io.EOF)
				got, _, err := mux.NextBatch(context.Background(), parent)
				if kind == "native" || kind == "CIP-64" {
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
	detector := &cgtDetectorStub{err: NewTemporaryError(errors.New("offline"))}
	mux := NewBatchMux(testlog.Logger(t, log.LevelError), cfg, prev, nil, detector)
	require.ErrorIs(t, mux.Reset(context.Background(), b, eth.SystemConfig{}), io.EOF)
	_, _, err := mux.NextBatch(context.Background(), parent)
	require.ErrorIs(t, err, ErrTemporary)
	detector.err = nil
	got, _, err := mux.NextBatch(context.Background(), parent)
	require.NoError(t, err)
	require.Equal(t, batch, got)
	require.Equal(t, 1, prev.i, "retry must not read another batch")
}

func TestCgtV2PipelineConstructorWiring(t *testing.T) {
	rs := cgtSeededReceipts()
	_, s, parent, origin := cgtTestDeriver(t, rs)
	cfg := &rollup.Config{BlockTime: 2, L1ChainID: big.NewInt(1), L2ChainID: big.NewInt(1337), L1SystemConfigAddress: common.Address{0x33}, DepositContractAddress: common.Address{0x44}}
	l1 := &testutils.MockL1Source{}
	l1.ExpectFetchReceipts(origin.Hash, s.info, rs, nil)
	l2 := &testutils.MockL2Client{}
	l2.ExpectSystemConfigByL2Hash(parent.Hash, eth.SystemConfig{GasLimit: 30000000}, nil)
	pipeline := NewDerivationPipeline(testlog.Logger(t, log.LevelError), cfg, nil, l1, nil, nil, l2, nil, params.MergedTestChainConfig, s)
	builder := pipeline.attrib.builder.(*FetchingAttributesBuilder)
	batch := &SingularBatch{ParentHash: parent.Hash, EpochHash: origin.Hash, EpochNum: rollup.Epoch(origin.Number), Timestamp: 102}
	attrs, err := pipeline.attrib.createNextAttributes(context.Background(), batch, parent)
	require.NoError(t, err)
	require.True(t, ContainsCgtV2Activation(attrs.Transactions))
	require.True(t, attrs.NoTxPool)
	mux := pipeline.attrib.prev.(*BatchMux)
	require.Same(t, builder, mux.cgt, "batch validation and attributes must use the same authenticated derivation")
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
	detector := &cgtDetectorStub{}
	mux := NewBatchMux(testlog.Logger(t, log.LevelError), cfg, prev, nil, detector)
	require.ErrorIs(t, mux.Reset(context.Background(), b, eth.SystemConfig{}), io.EOF)
	first, _, err := mux.NextBatch(context.Background(), parent)
	require.NoError(t, err)
	require.Equal(t, uint64(102), first.Timestamp)
	detector.active = true
	child := eth.L2BlockRef{Hash: common.Hash{4}, ParentHash: parent.Hash, Number: 101, Time: 102, L1Origin: b.ID()}
	next, _, err := mux.NextBatch(context.Background(), child)
	require.ErrorIs(t, err, NotEnoughData)
	require.Nil(t, next)
	require.Equal(t, uint64(102), first.Timestamp, "dropping the activation block must preserve the accepted prefix")
	require.Empty(t, mux.SingularBatchProvider.(*BatchStage).nextSpan)
}
