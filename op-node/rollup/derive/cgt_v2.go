package derive

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

var (
	cgtV2SourceHash   = (&UpgradeDepositSource{Intent: "Celo: CGTv2 Activation"}).SourceHash()
	cgtV2Selector     = [4]byte{0xc6, 0xd5, 0x30, 0x16}
	cgtV2Feature      = func() (out common.Hash) { copy(out[:], "CUSTOM_GAS_TOKEN"); return }()
	cgtV2FeatureEvent = crypto.Keccak256Hash([]byte("FeatureSet(bytes32,bool)"))
	// Provisional v6 layout, compiler snapshot at 20bb4480868c. Reconfirm this
	// mapping slot with the final L1 migration artifact before release.
	cgtV2FeatureSlot = crypto.Keccak256Hash(cgtV2Feature[:], common.LeftPadBytes([]byte{109}, 32))
)

// CgtV2ActivationDeposit uses the existing upgrade domain and matches the shared
// Rust executor's version-1 signal. Amounts are uint256 values in base units.
func CgtV2ActivationDeposit(origin common.Hash, amount *big.Int) (*optypes.DepositTx, error) {
	if amount == nil || amount.Sign() < 0 || amount.BitLen() > 256 {
		return nil, fmt.Errorf("CGTv2 migrated amount is not uint256")
	}
	data := make([]byte, 100)
	copy(data[:4], cgtV2Selector[:])
	data[35] = 1
	copy(data[36:68], origin[:])
	amount.FillBytes(data[68:100])
	target := predeploys.L1BlockAddr
	return &optypes.DepositTx{
		SourceHash: cgtV2SourceHash, From: L1InfoDepositerAddress, To: &target,
		Mint: big.NewInt(0), Value: big.NewInt(0), Gas: 100_000, Data: data,
	}, nil
}

// ContainsCgtV2Activation identifies the reserved upgrade source in attributes
// produced by derivation. Execution independently validates all signal fields.
func ContainsCgtV2Activation(txs []hexutil.Bytes) bool {
	for _, raw := range txs {
		if len(raw) == 0 || raw[0] != optypes.DepositTxType {
			continue
		}
		tx, err := optypes.UnmarshalDepositTx(raw)
		if err == nil && tx.SourceHash == cgtV2SourceHash {
			return true
		}
	}
	return false
}

type CgtV2InputSource interface {
	CgtV2StateSource
	L1BlockInfo(context.Context, eth.BlockID) (eth.BlockInfo, error)
}

func (s *CgtV2RPCSource) L1BlockInfo(ctx context.Context, block eth.BlockID) (eth.BlockInfo, error) {
	header, err := cgtV2Header(ctx, s.l1, block)
	if err != nil {
		return nil, err
	}
	return eth.HeaderBlockInfo(header), nil
}

// This layout is compiled protocol data, never a CLI/rollup-JSON override.
// Populate production addresses/slots from reviewed deployment/compiler output.
type cgtV2SnapshotLayout struct {
	bridge       common.Address
	amountSlot   common.Hash
	seededSlot   common.Hash
	seededOffset uint8
}

type cgtV2Derivation struct {
	source   CgtV2InputSource
	tracker  *cgtV2Tracker
	portal   common.Address
	snapshot *cgtV2SnapshotLayout
	parent   common.Hash
	origin   eth.BlockID
	info     eth.BlockInfo
	txs      []hexutil.Bytes
	active   bool
}

func newCgtV2Derivation(cfg *rollup.Config, source CgtV2InputSource) *cgtV2Derivation {
	if source == nil || cfg.L2ChainID == nil || !cfg.L2ChainID.IsUint64() {
		return nil
	}
	switch bigs.Uint64Strict(cfg.L2ChainID) {
	case 1337, 42220, 44787, 62320, 11142220, 11162320:
	default:
		return nil
	}
	return &cgtV2Derivation{
		source: source, portal: cfg.DepositContractAddress,
		tracker: newCgtV2Tracker(source, cfg.L1SystemConfigAddress, cgtV2FeatureSlot),
		// The final migratedAmount field is pending L1 contract work. Until its
		// proxy/layout is compiled here, an enable event fails explicitly.
		snapshot: nil,
	}
}

func (d *cgtV2Derivation) Reset() {
	d.tracker.Reset()
	d.parent, d.origin, d.info, d.txs, d.active = common.Hash{}, eth.BlockID{}, nil, nil, false
}

func cgtV2EnableLog(log *types.Log, systemConfig common.Address) bool {
	return log.Address == systemConfig && len(log.Topics) == 3 && len(log.Data) == 0 &&
		log.Topics[0] == cgtV2FeatureEvent && log.Topics[1] == cgtV2Feature &&
		log.Topics[2] == common.BigToHash(big.NewInt(1))
}

func (d *cgtV2Derivation) deposits(ctx context.Context, parent eth.L2BlockRef, origin eth.BlockID, receipts optypes.Receipts) ([]hexutil.Bytes, bool, error) {
	if err := d.tracker.prepareParent(ctx, parent); err != nil {
		return nil, false, err
	}
	if origin == parent.L1Origin {
		return nil, false, nil
	}
	if d.tracker.complete {
		txs, err := DeriveDeposits(receipts.Geth(), d.portal)
		return txs, false, err
	}
	cached := d.info != nil && d.parent == parent.Hash && d.origin == origin
	info := d.info
	if !cached {
		var err error
		info, err = d.source.L1BlockInfo(ctx, origin)
		if err != nil {
			return nil, false, err
		}
	}
	if info == nil || eth.ToBlockID(info) != origin || origin.Number != parent.L1Origin.Number+1 || info.ParentHash() != parent.L1Origin.Hash {
		return nil, false, NewResetError(fmt.Errorf("CGTv2 origin is not the next authenticated epoch"))
	}
	for _, receipt := range receipts {
		if receipt == nil {
			return nil, false, NewCriticalError(fmt.Errorf("CGTv2 missing origin receipt"))
		}
		for _, log := range receipt.Logs {
			if log == nil {
				return nil, false, NewCriticalError(fmt.Errorf("CGTv2 missing origin log"))
			}
		}
	}
	if types.DeriveSha(receipts, trie.NewStackTrie(nil)) != info.ReceiptHash() {
		return nil, false, NewCriticalError(fmt.Errorf("CGTv2 receipts do not match origin %s", origin))
	}
	// Attributes also use these fresh receipts for SystemConfig updates. A
	// cached signal authenticates neither a replacement receipt list nor its
	// independently returned RPC header. Always bind that list to this root.
	if cached {
		return d.txs, d.active, nil
	}

	var txs []hexutil.Bytes
	var active bool
	var index uint64
	for _, receipt := range receipts {
		for _, log := range receipt.Logs {
			// Block hash and log index are RPC metadata, outside the receipt root.
			// Derive them from the authenticated origin and canonical traversal.
			canonical := *log
			canonical.BlockHash, canonical.Index = origin.Hash, uint(index)
			index++
			if receipt.Status != types.ReceiptStatusSuccessful {
				continue
			}
			if !active && cgtV2EnableLog(log, d.tracker.systemConfig) {
				if d.snapshot == nil {
					return nil, false, NewCriticalError(fmt.Errorf("CGTv2 L1 snapshot address and layout are unresolved"))
				}
				values, err := d.source.L1Storage(ctx, origin, d.snapshot.bridge, []common.Hash{d.snapshot.amountSlot, d.snapshot.seededSlot})
				if err != nil {
					return nil, false, err
				}
				if len(values) != 2 {
					return nil, false, NewCriticalError(fmt.Errorf("CGTv2 missing snapshot storage"))
				}
				seeded, err := cgtV2Bool(values[1], d.snapshot.seededOffset)
				if err != nil {
					return nil, false, err
				}
				if !seeded {
					return nil, false, NewCriticalError(fmt.Errorf("CGTv2 snapshot is not seeded"))
				}
				signal, err := CgtV2ActivationDeposit(origin.Hash, values[0].Big())
				if err != nil {
					return nil, false, NewCriticalError(err)
				}
				raw, err := signal.MarshalBinary()
				if err != nil {
					return nil, false, NewCriticalError(err)
				}
				txs, active = append(txs, raw), true
			}
			if log.Address == d.portal && len(log.Topics) > 0 && log.Topics[0] == DepositEventABIHash {
				deposit, err := UnmarshalDepositLogEvent(&canonical)
				if err != nil {
					return nil, false, NewCriticalError(fmt.Errorf("CGTv2 ordered deposit: %w", err))
				}
				raw, err := deposit.MarshalBinary()
				if err != nil {
					return nil, false, NewCriticalError(err)
				}
				txs = append(txs, raw)
			}
		}
	}
	d.parent, d.origin, d.info, d.txs, d.active = parent.Hash, origin, info, txs, active
	d.tracker.proposed, d.tracker.activation = origin, active
	return txs, active, nil
}

// CgtV2ActivationBlock is shared by singular and span-derived per-block checks.
// No CGT-specific receipts are read after completion or within the same epoch.
func (ba *FetchingAttributesBuilder) CgtV2ActivationBlock(ctx context.Context, parent eth.L2BlockRef, epoch eth.BlockID) (bool, error) {
	if ba.cgt == nil {
		return false, nil
	}
	if err := ba.cgt.tracker.prepareParent(ctx, parent); err != nil {
		return false, err
	}
	if ba.cgt.tracker.complete || parent.L1Origin == epoch {
		return false, nil
	}
	_, receipts, err := ba.l1.FetchReceipts(ctx, epoch.Hash)
	if err != nil {
		return false, NewTemporaryError(fmt.Errorf("CGTv2 origin receipts: %w", err))
	}
	_, active, err := ba.cgt.deposits(ctx, parent, epoch, receipts)
	return active, err
}

// ResetCgtV2 invalidates branch-bound state even when a reset resumes at the
// same parent hash. Other builders need not implement this optional interface.
func (ba *FetchingAttributesBuilder) ResetCgtV2() {
	if ba.cgt != nil {
		ba.cgt.Reset()
	}
}
