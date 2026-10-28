package derive

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
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
	require.Len(t, first, 16)
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

func TestCgtV2DeprecatedChainsAreNotEnabled(t *testing.T) {
	for _, id := range []int64{44787, 62320} {
		cfg := cgtTestConfig()
		cfg.L2ChainID = big.NewInt(id)
		require.Nil(t, newCgtV2Derivation(cfg, &cgtInputStub{}))
	}
}

func TestCgtV2IgnoresForeignEscrowEmitter(t *testing.T) {
	receipts := cgtSeededReceipts()
	d, _, parent, origin := cgtTestDeriver(t, receipts)
	cfg := cgtTestConfig()
	cfg.L2ChainID = big.NewInt(params.CeloMainnetChainID)
	d = newCgtV2Derivation(cfg, d.source)
	_, _, group, err := d.originDeposits(context.Background(), parent, origin, receipts)
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
			_, source, parent, origin := cgtTestDeriver(t, receipts)
			parent.L1Origin.Hash = common.Hash{0xff}
			builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, &cgtL1Stub{source.info, receipts}, cgtL2Stub{}, source)
			_, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
			require.ErrorIs(t, err, ErrReset)
		})
	}
}

func TestCgtV2HeaderBinding(t *testing.T) {
	for _, name := range []string{"valid", "wrong hash", "wrong number", "nil number", "nil header", "RPC error", "missing RPC", "oversized number"} {
		t.Run(name, func(t *testing.T) {
			_, source, _, origin := cgtTestDeriver(t, nil)
			var rpc CgtV2InputSource = source
			switch name {
			case "wrong hash":
				source.header.Extra = []byte{1}
			case "wrong number":
				source.header.Number = big.NewInt(12)
			case "nil number":
				source.header.Number = nil
			case "nil header":
				source.header = nil
			case "RPC error":
				source.err = errors.New("offline")
			case "missing RPC":
				rpc = nil
			case "oversized number":
				source.header.Number = new(big.Int).Lsh(big.NewInt(1), 65)
			}
			header, err := cgtV2Header(context.Background(), rpc, origin)
			if name == "valid" {
				require.NoError(t, err)
				require.IsType(t, &types.Header{}, header)
			} else {
				require.Error(t, err)
				if name == "RPC error" {
					require.ErrorIs(t, err, ErrTemporary)
				} else {
					require.ErrorIs(t, err, ErrCritical)
				}
			}
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
	d, _, parent, origin := cgtTestDeriver(t, receipts)
	_, deposits, upgrades, err := d.originDeposits(context.Background(), parent, origin, receipts)
	txs := append(deposits, upgrades...)
	require.NoError(t, err)
	require.NotEmpty(t, upgrades)
	require.Len(t, txs, 18)
	for i, index := range []uint64{0, 2} {
		tx, err := optypes.UnmarshalDepositTx(txs[i])
		require.NoError(t, err)
		require.Equal(t, (&UserDepositSource{L1BlockHash: origin.Hash, LogIndex: index}).SourceHash(), tx.SourceHash)
	}
	require.False(t, containsCgtV2Group(txs[:2]))
	require.True(t, containsCgtV2Group(txs[2:]))
}

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
			require.Len(t, group, 16)
			fund, err := optypes.UnmarshalDepositTx(group[len(group)-1])
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
	var receipt types.Receipt
	require.NoError(t, rlp.DecodeBytes(vector.HeaderRlp, &header))
	require.NoError(t, rlp.DecodeBytes(vector.ReceiptRlp, &receipt))
	cfg := cgtTestConfig()
	cfg.DepositContractAddress = common.HexToAddress("0x44")
	cfg.L1SystemConfigAddress = common.Address{}
	zero := uint64(0)
	cfg.RegolithTime, cfg.CanyonTime, cfg.DeltaTime = &zero, &zero, &zero
	cfg.EcotoneTime, cfg.FjordTime, cfg.GraniteTime = &zero, &zero, &zero
	info := eth.HeaderBlockInfo(&header)
	source := &cgtInputStub{header: &header, info: info}
	receipts := optypes.FromGethReceipts(types.Receipts{&receipt})
	parent := eth.L2BlockRef{Hash: vector.ParentL2Hash, Time: 100, L1Origin: eth.BlockID{Hash: header.ParentHash, Number: 10}}
	builder := NewFetchingAttributesBuilder(cfg, params.MainnetChainConfig, nil, &cgtL1Stub{info, receipts}, cgtL2Stub{}, source)
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
