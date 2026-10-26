package derive

import (
	"context"
	"math/big"
	"testing"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/require"
)

type cgtInputStub struct {
	header      *types.Header
	info        eth.BlockInfo
	err         error
	headerReads int
}

func (s *cgtInputStub) HeaderByHash(context.Context, common.Hash) (*types.Header, error) {
	s.headerReads++
	return s.header, s.err
}
func cgtPortalLog(t *testing.T, portal common.Address, nonce byte) *types.Log {
	t.Helper()
	target := common.Address{nonce}
	event, err := MarshalDepositLogEvent(portal, &optypes.DepositTx{From: common.Address{0x11}, To: &target, Mint: big.NewInt(0), Value: big.NewInt(0), Gas: 100000})
	require.NoError(t, err)
	event.BlockHash, event.Index, event.TxIndex = common.Hash{0xee}, 777, 888
	return event
}
func cgtTestConfig() *rollup.Config {
	return &rollup.Config{BlockTime: 2, L1ChainID: big.NewInt(1), L2ChainID: big.NewInt(1337), L1SystemConfigAddress: common.Address{0x33}, DepositContractAddress: common.Address{0x44}}
}
func cgtTestDeriver(t *testing.T, receipts optypes.Receipts) (*cgtV2Derivation, *cgtInputStub, eth.L2BlockRef, eth.BlockID) {
	t.Helper()
	parent := eth.L2BlockRef{Hash: common.Hash{1}, Number: 100, Time: 100, L1Origin: eth.BlockID{Hash: common.Hash{2}, Number: 10}}
	header := &types.Header{Number: big.NewInt(11), ParentHash: parent.L1Origin.Hash, Time: 100, BaseFee: big.NewInt(1), ReceiptHash: types.DeriveSha(receipts, trie.NewStackTrie(nil))}
	source := &cgtInputStub{header: header, info: eth.HeaderBlockInfo(header)}
	return newCgtV2Derivation(cgtTestConfig(), source), source, parent, eth.ToBlockID(source.info)
}
func cgtSeededReceipts() optypes.Receipts {
	return optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{cgtEscrowLog(common.HexToAddress("0xcc"), common.Address{0x44}, 600)}}})
}

type cgtL1Stub struct {
	info     eth.BlockInfo
	receipts optypes.Receipts
}

func (s *cgtL1Stub) InfoByHash(context.Context, common.Hash) (eth.BlockInfo, error) {
	return s.info, nil
}
func (s *cgtL1Stub) FetchReceipts(context.Context, common.Hash) (eth.BlockInfo, optypes.Receipts, error) {
	return s.info, s.receipts, nil
}

type cgtL2Stub struct{}

func (cgtL2Stub) SystemConfigByL2Hash(context.Context, common.Hash) (eth.SystemConfig, error) {
	return eth.SystemConfig{GasLimit: 30000000}, nil
}

func TestCgtV2AttributesBuilderWiring(t *testing.T) {
	receipts := cgtSeededReceipts()
	_, source, parent, origin := cgtTestDeriver(t, receipts)
	cfg := cgtTestConfig()
	builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, &cgtL1Stub{source.info, receipts}, cgtL2Stub{}, source)
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.Len(t, attrs.Transactions, 14)
	require.True(t, attrs.NoTxPool)
	require.True(t, ContainsCgtV2Activation(attrs.Transactions))
	require.EqualValues(t, 30_000_000+cgtV2UpgradeGas, *attrs.GasLimit)
	tx, err := optypes.UnmarshalDepositTx(attrs.Transactions[0])
	require.NoError(t, err)
	require.Equal(t, L1InfoDepositerAddress, tx.From)
	batch := &SingularBatch{ParentHash: parent.Hash, EpochNum: rollup.Epoch(origin.Number), EpochHash: origin.Hash, Timestamp: 102}
	queue := NewAttributesQueue(testlog.Logger(t, log.LevelError), cfg, builder, nil)
	safe, err := queue.createNextAttributes(context.Background(), batch, parent)
	require.NoError(t, err)
	require.Equal(t, attrs.Transactions, safe.Transactions)
	// Rebuilding on the same valid parent requires no process-local migration state.
	fresh := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, &cgtL1Stub{source.info, receipts}, cgtL2Stub{}, source)
	replay, err := fresh.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.Equal(t, attrs, replay)
}

func TestCgtV2SameOriginAndReorgReplay(t *testing.T) {
	receipts := cgtSeededReceipts()
	d, source, parent, origin := cgtTestDeriver(t, receipts)
	_, _, first, err := d.originDeposits(context.Background(), parent, origin, receipts)
	require.NoError(t, err)
	require.Len(t, first, 13)
	child := eth.L2BlockRef{Hash: common.Hash{3}, ParentHash: parent.Hash, Number: 101, Time: 102, L1Origin: origin}
	builder := NewFetchingAttributesBuilder(cgtTestConfig(), params.MergedTestChainConfig, nil, &cgtL1Stub{source.info, receipts}, cgtL2Stub{}, source)
	reads := source.headerReads
	active, err := builder.CgtV2ActivationBlock(context.Background(), child, origin)
	require.NoError(t, err)
	require.False(t, active)
	require.Equal(t, reads, source.headerReads)
	original := source.header
	source.header = &types.Header{Number: big.NewInt(12), ParentHash: origin.Hash, ReceiptHash: types.EmptyReceiptsHash}
	_, _, later, err := d.originDeposits(context.Background(), child, eth.BlockID{Hash: source.header.Hash(), Number: 12}, nil)
	require.NoError(t, err)
	require.Empty(t, later)
	source.header = original
	_, _, replay, err := d.originDeposits(context.Background(), parent, origin, receipts)
	require.NoError(t, err)
	require.Equal(t, first, replay)
}

func TestCgtV2UpgradeGasRoundTrip(t *testing.T) {
	receipts := cgtSeededReceipts()
	_, source, parent, origin := cgtTestDeriver(t, receipts)
	cfg := cgtTestConfig()
	builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, &cgtL1Stub{source.info, receipts}, cgtL2Stub{}, source)
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	payload := &eth.ExecutionPayload{BlockNumber: 101, Timestamp: attrs.Timestamp, GasLimit: *attrs.GasLimit, Transactions: attrs.Transactions}
	system, err := PayloadToSystemConfig(cfg, payload)
	require.NoError(t, err)
	require.EqualValues(t, 30_000_000, system.GasLimit)
	payload.GasLimit = eth.Uint64Quantity(cgtV2UpgradeGas - 1)
	_, err = PayloadToSystemConfig(cfg, payload)
	require.Error(t, err)
}
