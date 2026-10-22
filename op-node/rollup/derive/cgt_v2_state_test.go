package derive

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type cgtProofRPC struct {
	proofs           map[common.Address]*eth.AccountResult
	header           *types.Header
	proof            *eth.AccountResult
	err              error
	requestedHash    common.Hash
	requestedAddress common.Address
	requestedSlots   []common.Hash
}

func (r *cgtProofRPC) HeaderByHash(_ context.Context, hash common.Hash) (*types.Header, error) {
	r.requestedHash = hash
	return r.header, r.err
}
func (r *cgtProofRPC) GetProof(_ context.Context, addr common.Address, slots []common.Hash, tag string) (*eth.AccountResult, error) {
	if tag != r.requestedHash.Hex() {
		return nil, errors.New("proof not requested at exact header hash")
	}
	r.requestedAddress, r.requestedSlots = addr, slots
	if r.proofs != nil {
		return r.proofs[addr], r.err
	}
	return r.proof, r.err
}

// Build real inclusion/exclusion proofs; changing an RPC claim cannot change the authenticated root.
func cgtProofFixture(t *testing.T, addr common.Address, slots []common.Hash, values []uint64, exists bool) (*types.Header, *eth.AccountResult) {
	t.Helper()
	db := triedb.NewDatabase(rawdb.NewMemoryDatabase(), nil)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	storage := trie.NewEmpty(db)
	for i, slot := range slots {
		if values[i] != 0 {
			encoded, err := rlp.EncodeToBytes(new(big.Int).SetUint64(values[i]).Bytes())
			require.NoError(t, err)
			require.NoError(t, storage.Update(crypto.Keccak256(slot[:]), encoded))
		}
	}
	storageRoot := storage.Hash()
	state := trie.NewEmpty(db)
	account := types.StateAccount{Balance: uint256.NewInt(0), Root: storageRoot, CodeHash: crypto.Keccak256([]byte{0x00})}
	accountBytes, err := rlp.EncodeToBytes(&account)
	require.NoError(t, err)
	if exists {
		require.NoError(t, state.Update(crypto.Keccak256(addr[:]), accountBytes))
	} else {
		// A real nonempty state trie proves the queried account's absence.
		require.NoError(t, state.Update(crypto.Keccak256(common.Address{0x99}.Bytes()), accountBytes))
	}
	proofNodes := func(tr *trie.Trie, key []byte) []hexutil.Bytes {
		nodes := memorydb.New()
		require.NoError(t, tr.Prove(key, nodes))
		it := nodes.NewIterator(nil, nil)
		defer it.Release()
		var out []hexutil.Bytes
		for it.Next() {
			out = append(out, append([]byte(nil), it.Value()...))
		}
		require.NoError(t, it.Error())
		return out
	}
	result := &eth.AccountResult{
		Address: addr, Balance: (*hexutil.Big)(big.NewInt(0)),
		CodeHash: common.BytesToHash(account.CodeHash), StorageHash: storageRoot,
		AccountProof: proofNodes(state, crypto.Keccak256(addr[:])),
	}
	for i, slot := range slots {
		result.StorageProof = append(result.StorageProof, eth.StorageProofEntry{
			Key: eth.StorageKey(slot.Bytes()), Value: hexutil.Big(*new(big.Int).SetUint64(values[i])),
			Proof: proofNodes(storage, crypto.Keccak256(slot[:])),
		})
	}
	return &types.Header{Number: big.NewInt(10), Root: state.Hash()}, result
}

func TestCgtV2ProofBinding(t *testing.T) {
	address, slot := common.Address{0xcc}, common.Hash{0x42}
	for _, name := range []string{"valid", "wrong hash", "wrong number", "wrong root", "wrong proxy", "wrong slot", "wrong value", "missing proof", "nil proof", "RPC failure"} {
		t.Run(name, func(t *testing.T) {
			header, proof := cgtProofFixture(t, address, []common.Hash{slot}, []uint64{600}, true)
			block := eth.BlockID{Hash: header.Hash(), Number: 10}
			rpc := &cgtProofRPC{header: header, proof: proof}
			switch name {
			case "wrong hash":
				block.Hash = common.Hash{0xff}
			case "wrong number":
				block.Number++
			case "wrong root":
				header.Root = common.Hash{0xff}
				block.Hash = header.Hash()
			case "wrong proxy":
				proof.Address = common.Address{0xdd}
			case "wrong slot":
				proof.StorageProof[0].Key = common.Hash{0x43}.Bytes()
			case "wrong value":
				proof.StorageProof[0].Value = hexutil.Big(*big.NewInt(601))
			case "missing proof":
				proof.AccountProof = nil
			case "nil proof":
				rpc.proof = nil
			case "RPC failure":
				rpc.err = errors.New("offline")
			}
			source := NewCgtV2RPCSource(rpc, rpc)
			got, err := source.L1Storage(context.Background(), block, address, []common.Hash{slot})
			if name != "valid" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, []common.Hash{common.BigToHash(big.NewInt(600))}, got)
			require.Equal(t, address, rpc.requestedAddress)
			require.Equal(t, []common.Hash{slot}, rpc.requestedSlots)
		})
	}
}

func TestCgtV2CompletionProof(t *testing.T) {
	for _, exists := range []bool{false, true} {
		t.Run(map[bool]string{false: "absent", true: "present"}[exists], func(t *testing.T) {
			markerHeader, marker := cgtProofFixture(t, cgtV2BridgeL2, nil, nil, exists)
			// Completion uses proofs for two accounts at the SAME parent state. The
			// absence case needs no flag read; the combined-state case is covered below.
			rpc := &cgtProofRPC{header: markerHeader, proof: marker}
			source := NewCgtV2RPCSource(rpc, rpc)
			parent := eth.L2BlockRef{Hash: markerHeader.Hash(), Number: 10}
			completion, err := source.L2Completion(context.Background(), parent)
			if exists {
				require.Error(t, err, "a marker proof cannot substitute for the L1Block flag proof")
				return
			}
			require.NoError(t, err)
			require.False(t, completion.MarkerPresent)
			// Forged marker metadata cannot turn exclusion into completion.
			marker.CodeHash = common.Hash{0x77}
			completion, err = source.L2Completion(context.Background(), parent)
			require.NoError(t, err)
			require.False(t, completion.MarkerPresent)
		})
	}
}

type cgtStateStub struct {
	completion       CgtV2Completion
	feature          common.Hash
	err              error
	l1Reads, l2Reads int
	lastOrigin       eth.BlockID
}

func (s *cgtStateStub) L1Storage(_ context.Context, origin eth.BlockID, _ common.Address, slots []common.Hash) ([]common.Hash, error) {
	s.l1Reads++
	s.lastOrigin = origin
	if s.err != nil {
		return nil, s.err
	}
	out := make([]common.Hash, len(slots))
	for i := range out {
		out[i] = s.feature
	}
	return out, nil
}
func (s *cgtStateStub) L2Completion(context.Context, eth.L2BlockRef) (CgtV2Completion, error) {
	s.l2Reads++
	return s.completion, s.err
}

func TestCgtV2Bootstrap(t *testing.T) {
	for _, tc := range []struct {
		name       string
		completion CgtV2Completion
		feature    uint64
		wantErr    bool
	}{
		{"before activation", CgtV2Completion{}, 0, false},
		{"completed", CgtV2Completion{true, true}, 1, false},
		{"inconsistent marker", CgtV2Completion{true, false}, 0, true},
		{"missed migration", CgtV2Completion{}, 1, true},
		{"invalid bool", CgtV2Completion{}, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := &cgtStateStub{completion: tc.completion, feature: common.BigToHash(new(big.Int).SetUint64(tc.feature))}
			tracker := newCgtV2Tracker(state, common.Address{0x11}, common.Hash{0x22})
			parent := eth.L2BlockRef{Hash: common.Hash{0x33}, Number: 100, L1Origin: eth.BlockID{Hash: common.Hash{0x44}, Number: 10}}
			err := tracker.Bootstrap(context.Background(), parent)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.completion.MarkerPresent, tracker.complete)
			if !tracker.complete {
				require.Equal(t, parent.L1Origin, state.lastOrigin)
			}
		})
	}
}

func TestCgtV2BootstrapRequestCounts(t *testing.T) {
	state := &cgtStateStub{completion: CgtV2Completion{true, true}}
	tracker := newCgtV2Tracker(state, common.Address{}, common.Hash{})
	parent := eth.L2BlockRef{Hash: common.Hash{1}, Number: 100}
	require.NoError(t, tracker.prepareParent(context.Background(), parent))
	require.NoError(t, tracker.prepareParent(context.Background(), parent))
	child := eth.L2BlockRef{Hash: common.Hash{2}, ParentHash: parent.Hash, Number: 101}
	require.NoError(t, tracker.prepareParent(context.Background(), child))
	require.Equal(t, 1, state.l2Reads)
	require.Zero(t, state.l1Reads)
	tracker.Reset()
	require.NoError(t, tracker.prepareParent(context.Background(), child))
	require.Equal(t, 2, state.l2Reads)
	// Same height, different branch must not inherit completion.
	state.completion = CgtV2Completion{}
	child.Hash = common.Hash{3}
	require.NoError(t, tracker.prepareParent(context.Background(), child))
	require.False(t, tracker.complete)
	require.Equal(t, 3, state.l2Reads)
	require.Equal(t, 1, state.l1Reads)
	state.err = errors.New("state unavailable")
	tracker.Reset()
	require.Error(t, tracker.prepareParent(context.Background(), child))
	require.False(t, tracker.initialized)
}

func TestCgtV2EmptyStateProof(t *testing.T) {
	header := &types.Header{Number: big.NewInt(10), Root: types.EmptyRootHash}
	rpc := &cgtProofRPC{header: header, proof: &eth.AccountResult{Address: cgtV2BridgeL2}}
	got, err := NewCgtV2RPCSource(rpc, rpc).L2Completion(context.Background(), eth.L2BlockRef{Hash: header.Hash(), Number: 10})
	require.NoError(t, err)
	require.False(t, got.MarkerPresent)
}

func TestCgtV2EmptyStorageProof(t *testing.T) {
	address, slot := common.Address{0xcc}, common.Hash{0x42}
	header, proof := cgtProofFixture(t, address, []common.Hash{slot}, []uint64{0}, true)
	rpc := &cgtProofRPC{header: header, proof: proof}
	got, err := NewCgtV2RPCSource(rpc, rpc).L1Storage(context.Background(), eth.BlockID{Hash: header.Hash(), Number: 10}, address, []common.Hash{slot})
	require.NoError(t, err)
	require.Equal(t, []common.Hash{{}}, got)
}

// Merge independently built storage tries into one actual parent state root.
func cgtProofWorld(t *testing.T, accounts map[common.Address]*eth.AccountResult) *types.Header {
	t.Helper()
	db := triedb.NewDatabase(rawdb.NewMemoryDatabase(), nil)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	state := trie.NewEmpty(db)
	for addr, p := range accounts {
		encoded, err := rlp.EncodeToBytes(&types.StateAccount{Nonce: uint64(p.Nonce), Balance: uint256.MustFromBig(p.Balance.ToInt()), Root: p.StorageHash, CodeHash: p.CodeHash.Bytes()})
		require.NoError(t, err)
		require.NoError(t, state.Update(crypto.Keccak256(addr[:]), encoded))
	}
	for addr, p := range accounts {
		nodes := memorydb.New()
		require.NoError(t, state.Prove(crypto.Keccak256(addr[:]), nodes))
		it := nodes.NewIterator(nil, nil)
		p.AccountProof = nil
		for it.Next() {
			p.AccountProof = append(p.AccountProof, append([]byte(nil), it.Value()...))
		}
		require.NoError(t, it.Error())
		it.Release()
	}
	return &types.Header{Number: big.NewInt(10), Root: state.Hash()}
}

func TestCgtV2CompletedParentProof(t *testing.T) {
	for _, enabled := range []uint64{0, 1, 2} {
		_, marker := cgtProofFixture(t, cgtV2BridgeL2, nil, nil, true)
		_, flag := cgtProofFixture(t, predeploys.L1BlockAddr, []common.Hash{cgtV2FlagSlot}, []uint64{enabled}, true)
		proofs := map[common.Address]*eth.AccountResult{cgtV2BridgeL2: marker, predeploys.L1BlockAddr: flag}
		header := cgtProofWorld(t, proofs)
		rpc := &cgtProofRPC{header: header, proofs: proofs}
		got, err := NewCgtV2RPCSource(rpc, rpc).L2Completion(context.Background(), eth.L2BlockRef{Hash: header.Hash(), Number: 10})
		if enabled == 2 {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)
		require.True(t, got.MarkerPresent)
		require.Equal(t, enabled == 1, got.FlagEnabled)
	}
}

func TestCgtV2PackedBoolOffset(t *testing.T) {
	word := common.Hash{}
	word[11] = 1
	word[31] = 2
	enabled, err := cgtV2Bool(word, 20)
	require.NoError(t, err)
	require.True(t, enabled)
	_, err = cgtV2Bool(word, 0)
	require.Error(t, err)
	_, err = cgtV2Bool(word, 32)
	require.Error(t, err)
}

func TestCgtV2DiscardedActivationParent(t *testing.T) {
	state := &cgtStateStub{}
	tracker := newCgtV2Tracker(state, common.Address{}, common.Hash{})
	parent := eth.L2BlockRef{Hash: common.Hash{1}, Number: 100, L1Origin: eth.BlockID{Hash: common.Hash{2}, Number: 10}}
	require.NoError(t, tracker.prepareParent(context.Background(), parent))
	tracker.proposed = eth.BlockID{Hash: common.Hash{3}, Number: 11}
	tracker.activation = true
	// Another accepted child stays in the old epoch; our proposed activation was discarded.
	child := eth.L2BlockRef{Hash: common.Hash{4}, ParentHash: parent.Hash, Number: 101, L1Origin: parent.L1Origin}
	require.NoError(t, tracker.prepareParent(context.Background(), child))
	require.False(t, tracker.complete)
	require.False(t, tracker.activation)
}
