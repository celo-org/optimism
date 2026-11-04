package derive

import (
	"context"
	"errors"
	"io"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

// TestAttributesQueue checks that it properly uses the PreparePayloadAttributes function
// (which is well tested) and that it properly sets NoTxPool and adds in the candidate
// transactions.
func TestAttributesQueue(t *testing.T) {
	// test config, only init the necessary fields
	cfg := &rollup.Config{
		BlockTime:              2,
		L1ChainID:              big.NewInt(101),
		L2ChainID:              big.NewInt(102),
		DepositContractAddress: common.Address{0xbb},
		L1SystemConfigAddress:  common.Address{0xcc},
	}
	rng := rand.New(rand.NewSource(1234))
	l1Info := testutils.RandomBlockInfo(rng)

	l1Fetcher := &testutils.MockL1Source{}
	defer l1Fetcher.AssertExpectations(t)

	l1Fetcher.ExpectInfoByHash(l1Info.InfoHash, l1Info, nil)

	safeHead := testutils.RandomL2BlockRef(rng)
	safeHead.L1Origin = l1Info.ID()
	safeHead.Time = l1Info.InfoTime

	batch := SingularBatch{
		ParentHash:   safeHead.Hash,
		EpochNum:     rollup.Epoch(l1Info.InfoNum),
		EpochHash:    l1Info.InfoHash,
		Timestamp:    safeHead.Time + cfg.BlockTime,
		Transactions: []eth.Data{eth.Data("foobar"), eth.Data("example")},
	}

	parentL1Cfg := eth.SystemConfig{
		BatcherAddr: common.Address{42},
		Overhead:    [32]byte{},
		Scalar:      [32]byte{},
		GasLimit:    1234,
	}
	expectedL1Cfg := eth.SystemConfig{
		BatcherAddr: common.Address{42},
		Overhead:    [32]byte{},
		Scalar:      [32]byte{},
		GasLimit:    1234,
	}

	l2Fetcher := &testutils.MockL2Client{}
	l2Fetcher.ExpectSystemConfigByL2Hash(safeHead.Hash, parentL1Cfg, nil)

	rollupCfg := rollup.Config{}
	l1InfoTx, err := L1InfoDepositBytes(&rollupCfg, params.MergedTestChainConfig, expectedL1Cfg, safeHead.SequenceNumber+1, l1Info, 0)
	require.NoError(t, err)
	attrs := eth.PayloadAttributes{
		Timestamp:             eth.Uint64Quantity(safeHead.Time + cfg.BlockTime),
		PrevRandao:            eth.Bytes32(l1Info.InfoMixDigest),
		SuggestedFeeRecipient: predeploys.SequencerFeeVaultAddr,
		Transactions:          []eth.Data{l1InfoTx, eth.Data("foobar"), eth.Data("example")},
		NoTxPool:              true,
		GasLimit:              (*eth.Uint64Quantity)(&expectedL1Cfg.GasLimit),
	}
	attrBuilder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, l1Fetcher, l2Fetcher)

	aq := NewAttributesQueue(testlog.Logger(t, log.LevelError), cfg, attrBuilder, nil)

	actual, err := aq.createNextAttributes(context.Background(), &batch, safeHead)

	require.NoError(t, err)
	require.Equal(t, attrs, *actual)
}

func TestCgtV2AttributesQueueDropsBatchAfterReceiptRetry(t *testing.T) {
	for _, retry := range []bool{false, true} {
		t.Run(map[bool]string{false: "immediate", true: "RPC retry"}[retry], func(t *testing.T) {
			receipts := cgtSeededReceipts()
			header, parent, origin := cgtTestOrigin(t, receipts)
			cfg := cgtTestConfig()
			cfg.HoloceneTime = new(uint64)
			cfg.MaxSequencerDrift = 600
			cfg.SeqWindowSize = 10
			base := eth.L1BlockRef{Hash: parent.L1Origin.Hash, Number: parent.L1Origin.Number, Time: 88}
			head := eth.L1BlockRef{Hash: origin.Hash, ParentHash: base.Hash, Number: origin.Number, Time: 100}
			batch := &SingularBatch{ParentHash: parent.Hash, EpochHash: origin.Hash, EpochNum: rollup.Epoch(origin.Number), Timestamp: 102, Transactions: []hexutil.Bytes{{0xc0}}}
			prev := &fakeBatchQueueInput{batches: []Batch{batch}, errors: []error{nil}, origin: head}
			mux := NewBatchMux(testlog.Logger(t, log.LevelError), cfg, prev, nil)
			require.ErrorIs(t, mux.Reset(context.Background(), base, eth.SystemConfig{}), io.EOF)
			l1 := &testutils.MockL1Source{}
			defer l1.AssertExpectations(t)
			if retry {
				l1.ExpectFetchReceipts(origin.Hash, nil, nil, errors.New("offline"))
			}
			l1.ExpectFetchReceipts(origin.Hash, eth.HeaderBlockInfo(header), receipts, nil)
			builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, l1, cgtL2Stub{})
			queue := NewAttributesQueue(testlog.Logger(t, log.LevelError), cfg, builder, mux)
			if retry {
				_, err := queue.NextAttributes(context.Background(), parent)
				require.ErrorIs(t, err, ErrTemporary)
				require.Same(t, batch, queue.batch)
				require.Equal(t, 1, prev.i)
			}
			attrs, err := queue.NextAttributes(context.Background(), parent)
			require.ErrorIs(t, err, NotEnoughData)
			require.Nil(t, attrs)
			require.Nil(t, queue.batch)
			require.Empty(t, prev.batches)
			_, err = queue.NextAttributes(context.Background(), parent)
			require.ErrorIs(t, err, io.EOF, "the rejected batch must not be retried")
		})
	}
}
