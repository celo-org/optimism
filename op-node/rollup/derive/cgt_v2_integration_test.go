package derive

import (
	"context"
	"encoding/binary"
	"math/big"
	"testing"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/require"
)

func TestCgtV2AuthenticatedDerivation(t *testing.T) {
	for _, name := range []string{"valid", "forged amount", "missing event", "forged header", "wrong parent", "nil receipt", "nil log", "noncanonical success", "noncanonical poststate"} {
		t.Run(name, func(t *testing.T) {
			receipts := cgtSeededReceipts()
			d, source, parent, origin := cgtTestDeriver(t, receipts)
			switch name {
			case "forged amount":
				receipts[0].Logs[0].Data[31]++
			case "missing event":
				receipts = nil
			case "forged header":
				source.header.ReceiptHash = types.EmptyReceiptsHash
			case "wrong parent":
				source.header.ParentHash = common.Hash{9}
				origin.Hash = source.header.Hash()
			case "nil receipt":
				receipts[0] = nil
			case "nil log":
				receipts[0].Logs[0] = nil
			case "noncanonical success":
				receipts[0].Status = 2
			case "noncanonical poststate":
				receipts[0].PostState = []byte{1}
				receipts[0].Status = 0
			}
			_, _, group, err := d.originDeposits(context.Background(), parent, origin, receipts)
			if name != "valid" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, group, 13)
			fund, err := optypes.UnmarshalDepositTx(group[12])
			require.NoError(t, err)
			require.EqualValues(t, 400, bigs.Uint64Strict(fund.Mint))
		})
	}
}

func TestCgtV2RetryReceiptsCannotChangeSystemConfig(t *testing.T) {
	receipts := cgtSeededReceipts()
	_, source, parent, origin := cgtTestDeriver(t, receipts)
	cfg := cgtTestConfig()
	l1 := &cgtL1Stub{source.info, receipts}
	builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, l1, cgtL2Stub{}, source)
	active, err := builder.CgtV2ActivationBlock(context.Background(), parent, origin)
	require.NoError(t, err)
	require.True(t, active)
	data := make([]byte, 96)
	data[31] = 32
	data[63] = 32
	binary.BigEndian.PutUint64(data[88:], 9_000_000)
	forged := &types.Log{Address: cfg.L1SystemConfigAddress, Topics: []common.Hash{ConfigUpdateEventABIHash, ConfigUpdateEventVersion0, SystemConfigUpdateGasLimit}, Data: data}
	l1.receipts = optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{receipts[0].Logs[0], forged}}})
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.ErrorIs(t, err, ErrCritical)
	require.Nil(t, attrs)
	l1.receipts = receipts
	attrs, err = builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.EqualValues(t, 30_000_000+cgtV2UpgradeGas, *attrs.GasLimit)
	// L1 data authentication remains required on every retry, not only on first detection.
	require.Equal(t, 3, source.headerReads)
}

type cgtConfigL2 struct{ config eth.SystemConfig }

func (s *cgtConfigL2) SystemConfigByL2Hash(context.Context, common.Hash) (eth.SystemConfig, error) {
	return s.config, nil
}
func cgtGasUpdate(cfg *rollup.Config, gas uint64) *types.Log {
	data := make([]byte, 96)
	data[31] = 32
	data[63] = 32
	binary.BigEndian.PutUint64(data[88:], gas)
	return &types.Log{Address: cfg.L1SystemConfigAddress, Topics: []common.Hash{ConfigUpdateEventABIHash, ConfigUpdateEventVersion0, SystemConfigUpdateGasLimit}, Data: data}
}
func TestCgtV2GasUpdateAndFollowingOrigin(t *testing.T) {
	cfg := cgtTestConfig()
	receipts := cgtSeededReceipts()
	receipts[0].Logs = append(receipts[0].Logs, cgtGasUpdate(cfg, 9_000_000))
	_, source, parent, origin := cgtTestDeriver(t, receipts)
	l1 := &cgtL1Stub{source.info, receipts}
	l2 := &cgtConfigL2{eth.SystemConfig{GasLimit: 30_000_000}}
	builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, l1, l2, source)
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.EqualValues(t, 9_000_000+cgtV2UpgradeGas, *attrs.GasLimit)
	payload := &eth.ExecutionPayload{BlockNumber: 101, Timestamp: attrs.Timestamp, GasLimit: *attrs.GasLimit, Transactions: attrs.Transactions}
	l2.config, err = PayloadToSystemConfig(cfg, payload)
	require.NoError(t, err)
	require.EqualValues(t, 9_000_000, l2.config.GasLimit)
	// A later L1 gas update overrides the recovered normal limit, without retaining group gas.
	nextReceipts := optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{cgtGasUpdate(cfg, 12_000_000)}}})
	source.header = &types.Header{Number: big.NewInt(12), ParentHash: origin.Hash, Time: 102, BaseFee: big.NewInt(1), ReceiptHash: types.DeriveSha(nextReceipts, trie.NewStackTrie(nil))}
	l1.info, l1.receipts = eth.HeaderBlockInfo(source.header), nextReceipts
	child := eth.L2BlockRef{Hash: common.Hash{3}, Number: 101, Time: 102, L1Origin: origin}
	next, err := builder.PreparePayloadAttributes(context.Background(), child, eth.ToBlockID(l1.info))
	require.NoError(t, err)
	require.False(t, ContainsCgtV2Activation(next.Transactions))
	require.EqualValues(t, 12_000_000, *next.GasLimit)
	// Overflow of the temporary gas allowance is a critical error.
	l2.config.GasLimit = ^uint64(0)
	overflowReceipts := cgtSeededReceipts()
	source.header = &types.Header{Number: big.NewInt(11), ParentHash: parent.L1Origin.Hash, Time: 100, BaseFee: big.NewInt(1), ReceiptHash: types.DeriveSha(overflowReceipts, trie.NewStackTrie(nil))}
	l1.info, l1.receipts = eth.HeaderBlockInfo(source.header), overflowReceipts
	_, err = builder.PreparePayloadAttributes(context.Background(), parent, eth.ToBlockID(eth.HeaderBlockInfo(source.header)))
	require.ErrorIs(t, err, ErrCritical)
}
