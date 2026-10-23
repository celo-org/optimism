package derive

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

var (
	cgtV2BridgeL2 = common.HexToAddress("0x4200000000000000000000000000000000001023")
	cgtV2FlagSlot = common.HexToHash("0xd2ff82c9b477ff6a09f530b1c627ffb4b0b81e2ae2ba427f824162e8dad020aa")
)

// CgtV2ProofClient supplies untrusted RPC data. Verification below is mandatory,
// including when the client's general trustRPC setting skips other checks.
type CgtV2ProofClient interface {
	HeaderByHash(context.Context, common.Hash) (*types.Header, error)
	GetProof(context.Context, common.Address, []common.Hash, string) (*eth.AccountResult, error)
}

type CgtV2Completion struct {
	MarkerPresent bool
	FlagEnabled   bool
}

type CgtV2StateSource interface {
	L1Storage(context.Context, eth.BlockID, common.Address, []common.Hash) ([]common.Hash, error)
	L2Completion(context.Context, eth.L2BlockRef) (CgtV2Completion, error)
}

// CgtV2RPCSource authenticates exact-hash L1 and L2 state without trusting RPC
// account metadata, independently returned roots, or latest block labels.
type CgtV2RPCSource struct {
	l1 CgtV2ProofClient
	l2 CgtV2ProofClient
}

func NewCgtV2RPCSource(l1, l2 CgtV2ProofClient) *CgtV2RPCSource {
	return &CgtV2RPCSource{l1: l1, l2: l2}
}

func cgtV2Header(ctx context.Context, rpc CgtV2ProofClient, block eth.BlockID) (*types.Header, error) {
	if rpc == nil {
		return nil, NewCriticalError(fmt.Errorf("CGTv2 requires header and account-proof RPC support"))
	}
	header, err := rpc.HeaderByHash(ctx, block.Hash)
	if err != nil {
		return nil, NewTemporaryError(fmt.Errorf("CGTv2 header %s: %w", block, err))
	}
	if header == nil || header.Number == nil || !header.Number.IsUint64() || bigs.Uint64Strict(header.Number) != block.Number || header.Hash() != block.Hash {
		return nil, NewCriticalError(fmt.Errorf("CGTv2 header does not match requested block %s", block))
	}
	return header, nil
}

func cgtV2Account(ctx context.Context, rpc CgtV2ProofClient, block eth.BlockID, root common.Hash, address common.Address, slots []common.Hash) (*eth.AccountResult, bool, error) {
	result, err := rpc.GetProof(ctx, address, slots, block.Hash.Hex())
	if err != nil {
		return nil, false, NewTemporaryError(fmt.Errorf("CGTv2 proof at %s for %s: %w", block, address, err))
	}
	if result == nil || result.Address != address || len(result.StorageProof) != len(slots) {
		return nil, false, NewCriticalError(fmt.Errorf("CGTv2 proof address or slot count mismatch"))
	}
	for i, slot := range slots {
		entry := result.StorageProof[i]
		if !bytes.Equal(entry.Key, slot[:]) || entry.Value.ToInt().Sign() < 0 || entry.Value.ToInt().BitLen() > 256 {
			return nil, false, NewCriticalError(fmt.Errorf("CGTv2 proof slot/value mismatch at %s", slot))
		}
	}

	// AccountResult.Verify only accepts inclusion proofs. Check existence first
	// so an authenticated exclusion can identify an undeployed L2 bridge.
	nodes := memorydb.New()
	defer nodes.Close()
	for _, node := range result.AccountProof {
		if err := nodes.Put(crypto.Keccak256(node), node); err != nil {
			return nil, false, NewCriticalError(fmt.Errorf("CGTv2 account proof node: %w", err))
		}
	}
	if root == types.EmptyRootHash {
		return result, false, nil
	}
	account, err := trie.VerifyProof(root, crypto.Keccak256(address[:]), nodes)
	if err != nil {
		return nil, false, NewCriticalError(fmt.Errorf("CGTv2 account proof: %w", err))
	}
	if account == nil {
		return result, false, nil
	}
	if result.Balance == nil || result.Balance.ToInt().Sign() < 0 || result.Balance.ToInt().BitLen() > 256 {
		return nil, false, NewCriticalError(fmt.Errorf("CGTv2 invalid account balance"))
	}
	verified := *result
	if result.StorageHash == types.EmptyRootHash {
		for _, entry := range result.StorageProof {
			if entry.Value.ToInt().Sign() != 0 {
				return nil, false, NewCriticalError(fmt.Errorf("CGTv2 nonzero storage in empty trie"))
			}
		}
		// The account proof authenticates the empty storage root, which proves
		// every slot is zero without a storage-trie node.
		verified.StorageProof = nil
	}
	if err := verified.Verify(root); err != nil {
		return nil, false, NewCriticalError(fmt.Errorf("CGTv2 account/storage proof: %w", err))
	}
	return result, true, nil
}

func (s *CgtV2RPCSource) L1Storage(ctx context.Context, block eth.BlockID, address common.Address, slots []common.Hash) ([]common.Hash, error) {
	header, err := cgtV2Header(ctx, s.l1, block)
	if err != nil {
		return nil, err
	}
	result, exists, err := cgtV2Account(ctx, s.l1, block, header.Root, address, slots)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NewCriticalError(fmt.Errorf("CGTv2 required L1 proxy %s is absent at %s", address, block))
	}
	values := make([]common.Hash, len(slots))
	for i, entry := range result.StorageProof {
		values[i] = common.BigToHash(entry.Value.ToInt())
	}
	return values, nil
}

func (s *CgtV2RPCSource) L2Completion(ctx context.Context, parent eth.L2BlockRef) (CgtV2Completion, error) {
	block := parent.ID()
	header, err := cgtV2Header(ctx, s.l2, block)
	if err != nil {
		return CgtV2Completion{}, err
	}
	marker, exists, err := cgtV2Account(ctx, s.l2, block, header.Root, cgtV2BridgeL2, nil)
	if err != nil {
		return CgtV2Completion{}, err
	}
	if !exists || marker.CodeHash == types.EmptyCodeHash {
		return CgtV2Completion{}, nil
	}
	flag, exists, err := cgtV2Account(ctx, s.l2, block, header.Root, predeploys.L1BlockAddr, []common.Hash{cgtV2FlagSlot})
	if err != nil {
		return CgtV2Completion{}, err
	}
	if !exists {
		return CgtV2Completion{}, NewCriticalError(fmt.Errorf("CGTv2 completed parent has no L1Block account"))
	}
	enabled, err := cgtV2Bool(common.BigToHash(flag.StorageProof[0].Value.ToInt()), 0)
	return CgtV2Completion{MarkerPresent: true, FlagEnabled: enabled}, err
}

// offset is the compiler storage-layout byte offset, counted from the low end.
func cgtV2Bool(word common.Hash, offset uint8) (bool, error) {
	if offset >= 32 || word[31-offset] > 1 {
		return false, NewCriticalError(fmt.Errorf("CGTv2 invalid bool at byte offset %d", offset))
	}
	return word[31-offset] == 1, nil
}

type cgtV2Tracker struct {
	source       CgtV2StateSource
	systemConfig common.Address
	featureSlot  common.Hash
	initialized  bool
	complete     bool
	parent       eth.L2BlockRef
	proposed     eth.BlockID
	activation   bool
}

func newCgtV2Tracker(source CgtV2StateSource, systemConfig common.Address, featureSlot common.Hash) *cgtV2Tracker {
	return &cgtV2Tracker{source: source, systemConfig: systemConfig, featureSlot: featureSlot}
}

func (t *cgtV2Tracker) Reset() {
	t.initialized, t.complete, t.activation = false, false, false
	t.parent, t.proposed = eth.L2BlockRef{}, eth.BlockID{}
}

func (t *cgtV2Tracker) Bootstrap(ctx context.Context, parent eth.L2BlockRef) error {
	t.Reset()
	completion, err := t.source.L2Completion(ctx, parent)
	if err != nil {
		return err
	}
	if completion.MarkerPresent {
		if !completion.FlagEnabled {
			return NewResetError(fmt.Errorf("CGTv2 parent has bridge code but the CGT flag is unset"))
		}
		t.complete = true
	} else {
		values, err := t.source.L1Storage(ctx, parent.L1Origin, t.systemConfig, []common.Hash{t.featureSlot})
		if err != nil {
			return err
		}
		if len(values) != 1 {
			return NewCriticalError(fmt.Errorf("CGTv2 missing bootstrap feature slot"))
		}
		enabled, err := cgtV2Bool(values[0], 0)
		if err != nil {
			return err
		}
		if enabled {
			return NewResetError(fmt.Errorf("CGTv2 parent origin is enabled but migration is absent"))
		}
	}
	t.parent, t.initialized = parent, true
	return nil
}

func (t *cgtV2Tracker) prepareParent(ctx context.Context, parent eth.L2BlockRef) error {
	if !t.initialized || (parent != t.parent && (parent.Number != t.parent.Number+1 || parent.ParentHash != t.parent.Hash)) {
		return t.Bootstrap(ctx, parent)
	}
	if parent == t.parent {
		return nil
	}
	// A proposed signal does not prove execution committed. Authenticate the
	// first actual child before caching completion and stopping L1 tracking.
	if !t.complete && parent.L1Origin != t.parent.L1Origin && parent.L1Origin != t.proposed {
		return t.Bootstrap(ctx, parent)
	}
	if !t.complete && t.activation && parent.L1Origin == t.proposed {
		if err := t.Bootstrap(ctx, parent); err != nil {
			return err
		}
		if !t.complete {
			return NewResetError(fmt.Errorf("CGTv2 activation child has no completion marker"))
		}
		return nil
	}
	t.parent = parent
	t.proposed, t.activation = eth.BlockID{}, false
	return nil
}
