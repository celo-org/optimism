package derive

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

func TestCgtV2SignalMatchesRustVector(t *testing.T) {
	var vector struct {
		Origin          common.Hash
		Amount          string
		Encoded         hexutil.Bytes
		SourceHash      common.Hash
		TransactionHash common.Hash
	}
	data, err := os.ReadFile("testdata/cgt_v2_signal_v1.json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &vector))
	amount, ok := new(big.Int).SetString(vector.Amount, 10)
	require.True(t, ok)
	tx, err := CgtV2ActivationDeposit(vector.Origin, amount)
	require.NoError(t, err)
	raw, err := tx.MarshalBinary()
	require.NoError(t, err)
	require.Equal(t, []byte(vector.Encoded), raw)
	require.Equal(t, vector.SourceHash, tx.SourceHash)
	require.Equal(t, vector.TransactionHash, tx.Hash())
	require.True(t, ContainsCgtV2Activation([]hexutil.Bytes{raw}))
	for _, amount := range []*big.Int{nil, big.NewInt(-1), new(big.Int).Lsh(big.NewInt(1), 256)} {
		_, err := CgtV2ActivationDeposit(vector.Origin, amount)
		require.Error(t, err)
	}
}

type cgtInputStub struct {
	cgtStateStub
	info                        eth.BlockInfo
	amount                      uint64
	seeded                      bool
	snapshotReads, receiptReads int
}

func (s *cgtInputStub) L1Storage(ctx context.Context, b eth.BlockID, a common.Address, slots []common.Hash) ([]common.Hash, error) {
	if len(slots) == 1 {
		return s.cgtStateStub.L1Storage(ctx, b, a, slots)
	}
	s.snapshotReads++
	if s.err != nil {
		return nil, s.err
	}
	if b != eth.ToBlockID(s.info) {
		return nil, errors.New("snapshot at wrong origin")
	}
	seeded := uint64(0)
	if s.seeded {
		seeded = 1
	}
	return []common.Hash{common.BigToHash(new(big.Int).SetUint64(s.amount)), common.BigToHash(new(big.Int).SetUint64(seeded))}, nil
}
func (s *cgtInputStub) L1BlockInfo(_ context.Context, b eth.BlockID) (eth.BlockInfo, error) {
	s.receiptReads++
	if eth.ToBlockID(s.info) != b {
		return nil, errors.New("wrong receipt root")
	}
	return s.info, nil
}

func cgtFeatureLog(system common.Address) *types.Log {
	var feature common.Hash
	copy(feature[:], []byte("CUSTOM_GAS_TOKEN"))
	return &types.Log{Address: system, Topics: []common.Hash{crypto.Keccak256Hash([]byte("FeatureSet(bytes32,bool)")), feature, common.BigToHash(big.NewInt(1))}, BlockHash: common.Hash{0xff}, Index: 999, TxIndex: 999}
}
func cgtPortalLog(t *testing.T, portal common.Address, nonce byte) *types.Log {
	t.Helper()
	target := common.Address{nonce}
	log, err := MarshalDepositLogEvent(portal, &optypes.DepositTx{From: common.Address{0x11}, To: &target, Mint: big.NewInt(0), Value: big.NewInt(0), Gas: 100000})
	require.NoError(t, err)
	log.BlockHash = common.Hash{0xee}
	log.Index = 777
	log.TxIndex = 888
	return log
}
func cgtTestDeriver(t *testing.T, rs optypes.Receipts) (*cgtV2Derivation, *cgtInputStub, eth.L2BlockRef, eth.BlockID) {
	t.Helper()
	parent := eth.L2BlockRef{Hash: common.Hash{1}, Number: 100, Time: 100, L1Origin: eth.BlockID{Hash: common.Hash{2}, Number: 10}}
	header := &types.Header{Number: big.NewInt(11), ParentHash: parent.L1Origin.Hash, Time: 100, BaseFee: big.NewInt(1), ReceiptHash: types.DeriveSha(rs, trie.NewStackTrie(nil))}
	source := &cgtInputStub{info: eth.HeaderBlockInfo(header), amount: 600, seeded: true}
	cfg := &rollup.Config{L2ChainID: big.NewInt(1337), L1SystemConfigAddress: common.Address{0x33}, DepositContractAddress: common.Address{0x44}}
	d := newCgtV2Derivation(cfg, source)
	d.snapshot = &cgtV2SnapshotLayout{bridge: common.HexToAddress("0xcc"), amountSlot: common.Hash{}, seededSlot: common.Hash{31: 1}}
	return d, source, parent, eth.ToBlockID(source.info)
}

func TestCgtV2DepositOrdering(t *testing.T) {
	portal, system := common.Address{0x44}, common.Address{0x33}
	// Put all three logs in one receipt. RPC indices deliberately disagree with canonical order.
	rs := optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{cgtPortalLog(t, portal, 1), cgtFeatureLog(system), cgtPortalLog(t, portal, 2)}}})
	d, source, parent, origin := cgtTestDeriver(t, rs)
	got, active, err := d.deposits(context.Background(), parent, origin, rs)
	require.NoError(t, err)
	require.True(t, active)
	require.Len(t, got, 3)
	for i, index := range []uint64{0, 2} {
		tx, err := optypes.UnmarshalDepositTx(got[i*2])
		require.NoError(t, err)
		require.Equal(t, (&UserDepositSource{L1BlockHash: origin.Hash, LogIndex: index}).SourceHash(), tx.SourceHash)
		require.Equal(t, common.Address{byte(i + 1)}, *tx.To)
	}
	signal, err := optypes.UnmarshalDepositTx(got[1])
	require.NoError(t, err)
	require.Equal(t, origin.Hash.Bytes(), signal.Data[36:68])
	require.Equal(t, big.NewInt(600), new(big.Int).SetBytes(signal.Data[68:]))
	require.Equal(t, 1, source.snapshotReads)
	again, active, err := d.deposits(context.Background(), parent, origin, rs)
	require.NoError(t, err)
	require.True(t, active)
	require.Equal(t, got, again)
	require.Equal(t, 1, source.snapshotReads, "same parent retries reuse authenticated inputs")
}

func TestCgtV2IgnoresUnrelatedFeatureLogs(t *testing.T) {
	system := common.Address{0x33}
	for _, name := range []string{"other emitter", "other feature", "failed receipt", "disabled", "malformed topics", "malformed data"} {
		t.Run(name, func(t *testing.T) {
			log := cgtFeatureLog(system)
			status := uint64(1)
			switch name {
			case "other emitter":
				log.Address = common.Address{0x55}
			case "other feature":
				log.Topics[1] = common.Hash{0x12}
			case "failed receipt":
				status = 0
			case "disabled":
				log.Topics[2] = common.Hash{}
			case "malformed topics":
				log.Topics = log.Topics[:2]
			case "malformed data":
				log.Data = []byte{1}
			}
			rs := optypes.FromGethReceipts(types.Receipts{{Status: status, Logs: []*types.Log{log}}})
			d, source, parent, origin := cgtTestDeriver(t, rs)
			txs, active, err := d.deposits(context.Background(), parent, origin, rs)
			require.NoError(t, err)
			require.False(t, active)
			require.Empty(t, txs)
			require.Zero(t, source.snapshotReads)
		})
	}
}

func TestCgtV2ActivationOriginAndReset(t *testing.T) {
	rs := optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{cgtFeatureLog(common.Address{0x33})}}})
	d, source, parent, origin := cgtTestDeriver(t, rs)
	_, active, err := d.deposits(context.Background(), parent, origin, rs)
	require.NoError(t, err)
	require.True(t, active)
	source.completion = CgtV2Completion{true, true}
	child := eth.L2BlockRef{Hash: common.Hash{3}, ParentHash: parent.Hash, Number: 101, Time: 102, L1Origin: origin}
	txs, active, err := d.deposits(context.Background(), child, origin, nil)
	require.NoError(t, err)
	require.False(t, active)
	require.Empty(t, txs)
	require.Equal(t, 1, source.snapshotReads)
	d.Reset()
	source.completion = CgtV2Completion{}
	_, active, err = d.deposits(context.Background(), parent, origin, rs)
	require.NoError(t, err)
	require.True(t, active)
	require.Equal(t, 2, source.snapshotReads)
	d.Reset()
	source.feature = common.BigToHash(big.NewInt(1))
	_, _, err = d.deposits(context.Background(), child, origin, nil)
	require.Error(t, err)
}

func TestCgtV2SnapshotSeededAndConfiguration(t *testing.T) {
	rs := optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{cgtFeatureLog(common.Address{0x33})}}})
	for _, name := range []string{"seeded zero", "unseeded zero", "unresolved layout", "RPC failure"} {
		t.Run(name, func(t *testing.T) {
			d, s, p, h := cgtTestDeriver(t, rs)
			s.amount = 0
			switch name {
			case "unseeded zero":
				s.seeded = false
			case "unresolved layout":
				d.snapshot = nil
			case "RPC failure":
				s.err = errors.New("offline")
			}
			txs, active, err := d.deposits(context.Background(), p, h, rs)
			if name != "seeded zero" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.True(t, active)
			tx, e := optypes.UnmarshalDepositTx(txs[0])
			require.NoError(t, e)
			require.Zero(t, new(big.Int).SetBytes(tx.Data[68:]).Sign())
		})
	}
}

// The same builder prepares sequencer attributes and safe attributes; this
// drives actual PreparePayloadAttributes, including L1-info and epoch rules.
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
	rs := optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{cgtFeatureLog(common.Address{0x33})}}})
	d, s, parent, origin := cgtTestDeriver(t, rs)
	cfg := &rollup.Config{BlockTime: 2, L1ChainID: big.NewInt(1), L2ChainID: big.NewInt(1337), L1SystemConfigAddress: common.Address{0x33}, DepositContractAddress: common.Address{0x44}}
	builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, &cgtL1Stub{s.info, rs}, cgtL2Stub{}, s)
	builder.cgt.snapshot = d.snapshot
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.Len(t, attrs.Transactions, 2)
	require.True(t, attrs.NoTxPool)
	require.True(t, ContainsCgtV2Activation(attrs.Transactions))
	tx, err := optypes.UnmarshalDepositTx(attrs.Transactions[0])
	require.NoError(t, err)
	require.Equal(t, predeploys.L1BlockAddr, *tx.To)
	batch := &SingularBatch{ParentHash: parent.Hash, EpochNum: rollup.Epoch(origin.Number), EpochHash: origin.Hash, Timestamp: 102}
	aq := NewAttributesQueue(testlog.Logger(t, log.LevelError), cfg, builder, nil)
	safe, err := aq.createNextAttributes(context.Background(), batch, parent)
	require.NoError(t, err)
	require.Equal(t, attrs.Transactions, safe.Transactions)
}
