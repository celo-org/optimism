package batcher

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// fakeL1 serves a linear chain of L1 headers; HeaderByNumber(nil) returns the last one.
type fakeL1 struct {
	headers []*types.Header
}

func newFakeL1(n int, blockTime uint64) *fakeL1 {
	l1 := &fakeL1{}
	parent := common.Hash{}
	for i := 0; i < n; i++ {
		h := &types.Header{Number: big.NewInt(int64(i)), Time: 1000 + uint64(i)*blockTime, ParentHash: parent}
		l1.headers = append(l1.headers, h)
		parent = h.Hash()
	}
	return l1
}

func (f *fakeL1) HeaderByNumber(_ context.Context, number *big.Int) (*types.Header, error) {
	if number == nil {
		return f.headers[len(f.headers)-1], nil
	}
	if number.Uint64() >= uint64(len(f.headers)) {
		return nil, ethereum.NotFound
	}
	return f.headers[number.Uint64()], nil
}

func (f *fakeL1) ref(n uint64) eth.L1BlockRef {
	return eth.InfoToL1BlockRef(eth.HeaderBlockInfo(f.headers[n]))
}

func TestCheckDerivationRules(t *testing.T) {
	zero := uint64(0)
	cfg := &rollup.Config{
		BlockTime:         2,
		SeqWindowSize:     10,
		MaxSequencerDrift: 600,
		RegolithTime:      &zero,
		CanyonTime:        &zero,
		DeltaTime:         &zero,
		EcotoneTime:       &zero,
		FjordTime:         &zero,
		GraniteTime:       &zero,
		HoloceneTime:      &zero,
	}
	const l1BlockTime = 12
	userTx := hexutil.Bytes{types.DynamicFeeTxType, 0x01}

	l1 := newFakeL1(6, l1BlockTime)
	origin := l1.ref(3)
	parent := eth.L2BlockRef{
		Hash:     common.Hash{0xaa},
		Number:   100,
		Time:     origin.Time + 4,
		L1Origin: origin.ID(),
	}
	// validBatch builds on parent and keeps its epoch.
	validBatch := func() *derive.SingularBatch {
		return &derive.SingularBatch{
			ParentHash:   parent.Hash,
			EpochNum:     rollup.Epoch(origin.Number),
			EpochHash:    origin.Hash,
			Timestamp:    parent.Time + cfg.BlockTime,
			Transactions: []hexutil.Bytes{userTx},
		}
	}

	tests := []struct {
		name   string
		l1     *fakeL1
		parent eth.L2BlockRef
		mutate func(b *derive.SingularBatch)
		check  func(t *testing.T, err error)
	}{
		{
			name:   "same epoch",
			mutate: func(b *derive.SingularBatch) {},
			check:  func(t *testing.T, err error) { require.NoError(t, err) },
		},
		{
			name: "next epoch",
			mutate: func(b *derive.SingularBatch) {
				next := l1.ref(4)
				b.EpochNum, b.EpochHash = rollup.Epoch(next.Number), next.Hash
			},
			parent: eth.L2BlockRef{Hash: parent.Hash, Number: parent.Number, Time: l1.ref(4).Time - cfg.BlockTime, L1Origin: origin.ID()},
			check:  func(t *testing.T, err error) { require.NoError(t, err) },
		},
		{
			name: "next epoch not visible on L1 yet",
			l1:   newFakeL1(4, l1BlockTime),
			mutate: func(b *derive.SingularBatch) {
				b.EpochNum, b.EpochHash = rollup.Epoch(origin.Number+1), common.Hash{0x01}
			},
			check: func(t *testing.T, err error) { require.ErrorIs(t, err, errBatchUndecided) },
		},
		{
			name: "epoch skip",
			mutate: func(b *derive.SingularBatch) {
				skip := l1.ref(5)
				b.EpochNum, b.EpochHash = rollup.Epoch(skip.Number), skip.Hash
			},
			parent: eth.L2BlockRef{Hash: parent.Hash, Number: parent.Number, Time: l1.ref(5).Time - cfg.BlockTime, L1Origin: origin.ID()},
			check:  requireRejected,
		},
		{
			name:   "epoch too old",
			mutate: func(b *derive.SingularBatch) { b.EpochNum, b.EpochHash = rollup.Epoch(2), l1.ref(2).Hash },
			check:  requireRejected,
		},
		{
			name:   "wrong epoch hash",
			mutate: func(b *derive.SingularBatch) { b.EpochHash = common.Hash{0x01} },
			check:  requireRejected,
		},
		{
			name:   "timestamp gap",
			mutate: func(b *derive.SingularBatch) { b.Timestamp += cfg.BlockTime },
			check:  requireRejected,
		},
		{
			name:   "timestamp not advancing",
			mutate: func(b *derive.SingularBatch) { b.Timestamp = parent.Time },
			check:  requireRejected,
		},
		{
			name:   "wrong parent hash",
			mutate: func(b *derive.SingularBatch) { b.ParentHash = common.Hash{0x01} },
			check:  requireRejected,
		},
		{
			name: "L2 timestamp before L1 origin",
			mutate: func(b *derive.SingularBatch) {
				next := l1.ref(4)
				b.EpochNum, b.EpochHash = rollup.Epoch(next.Number), next.Hash
			},
			check: requireRejected,
		},
		{
			name:   "deposit tx in batch",
			mutate: func(b *derive.SingularBatch) { b.Transactions = []hexutil.Bytes{{types.DepositTxType, 0x01}} },
			check:  requireRejected,
		},
		{
			name:   "empty tx in batch",
			mutate: func(b *derive.SingularBatch) { b.Transactions = []hexutil.Bytes{{}} },
			check:  requireRejected,
		},
		{
			name:   "sequencer drift exceeded with user txs",
			mutate: func(b *derive.SingularBatch) { b.Timestamp = origin.Time + 1802 },
			parent: eth.L2BlockRef{Hash: parent.Hash, Number: parent.Number, Time: origin.Time + 1800, L1Origin: origin.ID()},
			check:  requireRejected,
		},
		{
			name:   "sequencing window expired",
			l1:     newFakeL1(int(origin.Number+cfg.SeqWindowSize+2), l1BlockTime),
			mutate: func(b *derive.SingularBatch) {},
			check:  requireRejected,
		},
		{
			name:   "parent L1 origin not canonical",
			parent: eth.L2BlockRef{Hash: parent.Hash, Number: parent.Number, Time: parent.Time, L1Origin: eth.BlockID{Number: origin.Number, Hash: common.Hash{0x01}}},
			mutate: func(b *derive.SingularBatch) {},
			check:  func(t *testing.T, err error) { require.ErrorIs(t, err, errBatchUndecided) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fetcher := l1
			if tc.l1 != nil {
				fetcher = tc.l1
			}
			p := parent
			if tc.parent != (eth.L2BlockRef{}) {
				p = tc.parent
			}
			b := validBatch()
			b.Timestamp = p.Time + cfg.BlockTime
			tc.mutate(b)
			err := checkDerivationRules(context.Background(), cfg, testlog.Logger(t, log.LevelDebug), fetcher, time.Second, p, b)
			tc.check(t, err)
		})
	}
}

func requireRejected(t *testing.T, err error) {
	var rejected *ErrBatchRejected
	require.ErrorAs(t, err, &rejected)
}
