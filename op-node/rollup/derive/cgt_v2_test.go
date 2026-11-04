package derive

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"os"
	"testing"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/require"
)

func cgtPortalLog(t *testing.T, portal common.Address, nonce byte) *types.Log {
	t.Helper()
	target := common.Address{nonce}
	event, err := MarshalDepositLogEvent(portal, &optypes.DepositTx{From: common.Address{0x11}, To: &target, Mint: big.NewInt(0), Value: big.NewInt(0), Gas: 100000})
	require.NoError(t, err)
	return event
}
func cgtTestConfig() *rollup.Config {
	return &rollup.Config{BlockTime: 2, L1ChainID: big.NewInt(1), L2ChainID: big.NewInt(1337), L1SystemConfigAddress: common.Address{0x33}, DepositContractAddress: common.Address{0x44}}
}
func cgtTestOrigin(t *testing.T, receipts optypes.Receipts) (*types.Header, eth.L2BlockRef, eth.BlockID) {
	t.Helper()
	parent := eth.L2BlockRef{Hash: common.Hash{1}, Number: 100, Time: 100, L1Origin: eth.BlockID{Hash: common.Hash{2}, Number: 10}}
	header := &types.Header{Number: big.NewInt(11), ParentHash: parent.L1Origin.Hash, Time: 100, BaseFee: big.NewInt(1), ReceiptHash: types.DeriveSha(receipts, trie.NewStackTrie(nil))}
	var index uint
	for _, receipt := range receipts {
		for _, log := range receipt.Logs {
			log.BlockHash, log.Index = header.Hash(), index
			index++
		}
	}
	return header, parent, eth.BlockID{Hash: header.Hash(), Number: 11}
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
	header, parent, origin := cgtTestOrigin(t, receipts)
	cfg := cgtTestConfig()
	builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, &cgtL1Stub{eth.HeaderBlockInfo(header), receipts}, cgtL2Stub{})
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.Len(t, attrs.Transactions, 17)
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
	fresh := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, &cgtL1Stub{eth.HeaderBlockInfo(header), receipts}, cgtL2Stub{})
	replay, err := fresh.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.Equal(t, attrs, replay)
}

func TestCgtV2SameOriginAndReorgReplay(t *testing.T) {
	receipts := cgtSeededReceipts()
	header, parent, origin := cgtTestOrigin(t, receipts)
	l1 := &cgtL1Stub{eth.HeaderBlockInfo(header), receipts}
	builder := NewFetchingAttributesBuilder(cgtTestConfig(), params.MergedTestChainConfig, nil, l1, cgtL2Stub{})
	first, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.Len(t, first.Transactions, 17)
	child := eth.L2BlockRef{Hash: common.Hash{3}, ParentHash: parent.Hash, Number: 101, Time: 102, L1Origin: origin}
	same, err := builder.PreparePayloadAttributes(context.Background(), child, origin)
	require.NoError(t, err)
	require.False(t, ContainsCgtV2Activation(same.Transactions))
	laterHeader := &types.Header{Number: big.NewInt(12), ParentHash: origin.Hash, Time: 102, BaseFee: big.NewInt(1), ReceiptHash: types.EmptyReceiptsHash}
	l1.info, l1.receipts = eth.HeaderBlockInfo(laterHeader), nil
	later, err := builder.PreparePayloadAttributes(context.Background(), child, eth.ToBlockID(l1.info))
	require.NoError(t, err)
	require.False(t, ContainsCgtV2Activation(later.Transactions))
	l1.info, l1.receipts = eth.HeaderBlockInfo(header), receipts
	replay, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.Equal(t, first, replay)
}

func TestCgtV2UpgradeGasRoundTrip(t *testing.T) {
	receipts := cgtSeededReceipts()
	header, parent, origin := cgtTestOrigin(t, receipts)
	cfg := cgtTestConfig()
	builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, &cgtL1Stub{eth.HeaderBlockInfo(header), receipts}, cgtL2Stub{})
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

func TestCgtV2DeprecatedChainsAreNotEnabled(t *testing.T) {
	for _, id := range []int64{44787, 62320} {
		cfg := cgtTestConfig()
		cfg.L2ChainID = big.NewInt(id)
		upgrades, err := cgtV2UpgradeTransactionsFromReceipts(cfg, cgtSeededReceipts())
		require.NoError(t, err)
		require.Empty(t, upgrades)
	}
}

func TestCgtV2IgnoresForeignEscrowEmitter(t *testing.T) {
	cfg := cgtTestConfig()
	cfg.L2ChainID = big.NewInt(params.CeloMainnetChainID)
	group, err := cgtV2UpgradeTransactionsFromReceipts(cfg, cgtSeededReceipts())
	require.NoError(t, err)
	require.Empty(t, group)
}

func TestAttributesParentMismatchPrecedesMalformedDeposit(t *testing.T) {
	for _, chainID := range []int64{10, 1337} {
		t.Run(big.NewInt(chainID).String(), func(t *testing.T) {
			cfg := cgtTestConfig()
			cfg.L2ChainID = big.NewInt(chainID)
			deposit := cgtPortalLog(t, cfg.DepositContractAddress, 1)
			deposit.Data = nil
			receipts := optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{deposit}}})
			header, parent, origin := cgtTestOrigin(t, receipts)
			parent.L1Origin.Hash = common.Hash{0xff}
			builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, &cgtL1Stub{eth.HeaderBlockInfo(header), receipts}, cgtL2Stub{})
			_, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
			require.ErrorIs(t, err, ErrReset)
		})
	}
}

func cgtEscrowLog(bridge, portal common.Address, amount int64) *types.Log {
	return &types.Log{Address: bridge, Topics: []common.Hash{crypto.Keccak256Hash([]byte("EscrowSeeded(address,uint256)")), common.BytesToHash(portal.Bytes())}, Data: common.LeftPadBytes(big.NewInt(amount).Bytes(), 32)}
}

func TestCgtV2EscrowEvent(t *testing.T) {
	bridge, portal := common.HexToAddress("0xcc"), common.HexToAddress("0x44")
	for _, name := range []string{"valid", "zero", "wrong bridge", "genesis seeding", "failed receipt", "wrong topic", "missing indexed portal", "short amount", "extra topic"} {
		t.Run(name, func(t *testing.T) {
			event := cgtEscrowLog(bridge, portal, 600)
			receipt := &types.Receipt{Status: 1, Logs: []*types.Log{event}}
			want, bad := int64(600), false
			switch name {
			case "zero":
				event.Data = make([]byte, 32)
				want = 0
			case "wrong bridge":
				event.Address = common.HexToAddress("0xee")
				want = -1
			case "genesis seeding":
				event.Topics[1] = common.HexToHash("0xaa")
				want = -1
			case "failed receipt":
				receipt.Status = 0
				want = -1
			case "wrong topic":
				event.Topics[0] = common.Hash{}
				want = -1
			case "missing indexed portal":
				event.Topics = event.Topics[:1]
				bad = true
			case "short amount":
				event.Data = event.Data[:31]
				bad = true
			case "extra topic":
				event.Topics = append(event.Topics, common.Hash{})
				bad = true
			}
			amount, err := cgtV2EscrowAmount(optypes.FromGethReceipts(types.Receipts{receipt}), bridge, portal)
			if bad {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if want < 0 {
				require.Nil(t, amount)
			} else {
				require.Zero(t, amount.Cmp(big.NewInt(want)))
			}
		})
	}
}

func TestCgtV2EscrowAppendsGroupAfterAllOriginDeposits(t *testing.T) {
	portal := common.Address{0x44}
	receipts := optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{
		cgtPortalLog(t, portal, 1), cgtEscrowLog(common.HexToAddress("0xcc"), portal, 600), cgtPortalLog(t, portal, 2),
	}}})
	header, parent, origin := cgtTestOrigin(t, receipts)
	builder := NewFetchingAttributesBuilder(cgtTestConfig(), params.MergedTestChainConfig, nil, &cgtL1Stub{eth.HeaderBlockInfo(header), receipts}, cgtL2Stub{})
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.Len(t, attrs.Transactions, 19)
	for i, index := range []uint64{0, 2} {
		tx, err := optypes.UnmarshalDepositTx(attrs.Transactions[i+1])
		require.NoError(t, err)
		require.Equal(t, (&UserDepositSource{L1BlockHash: origin.Hash, LogIndex: index}).SourceHash(), tx.SourceHash)
	}
	require.False(t, containsCgtV2Group(attrs.Transactions[:3]))
	require.True(t, containsCgtV2Group(attrs.Transactions[3:]))
}

func TestCgtV2CanonicalReceiptFields(t *testing.T) {
	for _, name := range []string{"valid", "nil receipt", "nil log", "noncanonical success", "noncanonical poststate", "invalid receipt after event"} {
		t.Run(name, func(t *testing.T) {
			receipts := cgtSeededReceipts()
			switch name {
			case "nil receipt":
				receipts[0] = nil
			case "nil log":
				receipts[0].Logs[0] = nil
			case "noncanonical success":
				receipts[0].Status = 2
			case "noncanonical poststate":
				receipts[0].PostState = []byte{1}
			case "invalid receipt after event":
				receipts = append(receipts, optypes.FromGethReceipt(&types.Receipt{Status: 2}))
			}
			group, err := cgtV2UpgradeTransactionsFromReceipts(cgtTestConfig(), receipts)
			if name != "valid" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, group, 16)
			fund, err := optypes.UnmarshalDepositTx(group[len(group)-1])
			require.NoError(t, err)
			require.EqualValues(t, 400, bigs.Uint64Strict(fund.Mint))
		})
	}
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
	header, parent, origin := cgtTestOrigin(t, receipts)
	l1 := &cgtL1Stub{eth.HeaderBlockInfo(header), receipts}
	l2 := &cgtConfigL2{eth.SystemConfig{GasLimit: 30_000_000}}
	builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, l1, l2)
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.EqualValues(t, 9_000_000+cgtV2UpgradeGas, *attrs.GasLimit)
	payload := &eth.ExecutionPayload{BlockNumber: 101, Timestamp: attrs.Timestamp, GasLimit: *attrs.GasLimit, Transactions: attrs.Transactions}
	l2.config, err = PayloadToSystemConfig(cfg, payload)
	require.NoError(t, err)
	require.EqualValues(t, 9_000_000, l2.config.GasLimit)
	// A later L1 gas update overrides the recovered normal limit, without retaining group gas.
	nextReceipts := optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{cgtGasUpdate(cfg, 12_000_000)}}})
	header = &types.Header{Number: big.NewInt(12), ParentHash: origin.Hash, Time: 102, BaseFee: big.NewInt(1), ReceiptHash: types.DeriveSha(nextReceipts, trie.NewStackTrie(nil))}
	l1.info, l1.receipts = eth.HeaderBlockInfo(header), nextReceipts
	child := eth.L2BlockRef{Hash: common.Hash{3}, Number: 101, Time: 102, L1Origin: origin}
	next, err := builder.PreparePayloadAttributes(context.Background(), child, eth.ToBlockID(l1.info))
	require.NoError(t, err)
	require.False(t, ContainsCgtV2Activation(next.Transactions))
	require.EqualValues(t, 12_000_000, *next.GasLimit)
	// Overflow of the temporary gas allowance is a critical error.
	l2.config.GasLimit = ^uint64(0)
	overflowReceipts := cgtSeededReceipts()
	header = &types.Header{Number: big.NewInt(11), ParentHash: parent.L1Origin.Hash, Time: 100, BaseFee: big.NewInt(1), ReceiptHash: types.DeriveSha(overflowReceipts, trie.NewStackTrie(nil))}
	l1.info, l1.receipts = eth.HeaderBlockInfo(header), overflowReceipts
	_, err = builder.PreparePayloadAttributes(context.Background(), parent, eth.ToBlockID(eth.HeaderBlockInfo(header)))
	require.ErrorIs(t, err, ErrCritical)
}

func TestCgtV2FullOriginMatchesNativeProofClient(t *testing.T) {
	raw, err := os.ReadFile("testdata/cgtv2-origin.json")
	require.NoError(t, err)
	var vector struct {
		HeaderRlp    hexutil.Bytes   `json:"headerRlp"`
		ReceiptRlp   hexutil.Bytes   `json:"receiptRlp"`
		ParentL2Hash common.Hash     `json:"parentL2Hash"`
		Transactions []hexutil.Bytes `json:"transactions"`
		GasLimit     uint64          `json:"gasLimit"`
	}
	require.NoError(t, json.Unmarshal(raw, &vector))
	var header types.Header
	require.NoError(t, rlp.DecodeBytes(vector.HeaderRlp, &header))
	cfg := cgtTestConfig()
	cfg.DepositContractAddress = common.HexToAddress("0x44")
	cfg.L1SystemConfigAddress = common.Address{}
	zero := uint64(0)
	cfg.RegolithTime, cfg.CanyonTime, cfg.DeltaTime = &zero, &zero, &zero
	cfg.EcotoneTime, cfg.FjordTime, cfg.GraniteTime = &zero, &zero, &zero
	info := eth.HeaderBlockInfo(&header)
	receipts, err := eth.DecodeRawReceipts(eth.ToBlockID(info), []hexutil.Bytes{vector.ReceiptRlp}, []common.Hash{{1}})
	require.NoError(t, err)
	parent := eth.L2BlockRef{Hash: vector.ParentL2Hash, Time: 100, L1Origin: eth.BlockID{Hash: header.ParentHash, Number: 10}}
	builder := NewFetchingAttributesBuilder(cfg, params.MainnetChainConfig, nil, &cgtL1Stub{info, receipts}, cgtL2Stub{})
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, eth.BlockID{Hash: header.Hash(), Number: 11})
	require.NoError(t, err)
	require.Equal(t, vector.Transactions, attrs.Transactions)
	require.EqualValues(t, vector.GasLimit, *attrs.GasLimit)
	require.Len(t, attrs.Transactions, 19)
	for _, index := range []int{1, 2} {
		deposit, err := optypes.UnmarshalDepositTx(attrs.Transactions[index])
		require.NoError(t, err)
		if index == 1 {
			require.Zero(t, deposit.Mint.Cmp(big.NewInt(5)))
		} else {
			require.Zero(t, deposit.Mint.Sign())
		}
	}
}
