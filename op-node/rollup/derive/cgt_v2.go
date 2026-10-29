package derive

import (
	"context"
	"fmt"
	"math/big"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
)

// ContainsCgtV2Activation recognizes any reserved group source. Execution checks
// every required transaction field, order, success and group completeness.
func ContainsCgtV2Activation(txs []hexutil.Bytes) bool { return containsCgtV2Group(txs) }

// cgtV2Derivation holds fixed protocol inputs only. There is no migration-status
// cache or L1/L2 state bootstrap to reset during retries, restart or reorg.
type cgtV2Derivation struct {
	source  CgtV2InputSource
	portal  common.Address
	bridge  common.Address
	chainID *big.Int
}

func newCgtV2Derivation(cfg *rollup.Config, source CgtV2InputSource) *cgtV2Derivation {
	if cfg.L2ChainID == nil || !cfg.L2ChainID.IsUint64() {
		return nil
	}
	network, ok := cgtV2Params(bigs.Uint64Strict(cfg.L2ChainID))
	if !ok {
		return nil
	}
	return &cgtV2Derivation{
		source: source, portal: cfg.DepositContractAddress,
		bridge: network.bridge, chainID: new(big.Int).Set(cfg.L2ChainID),
	}
}

func (d *cgtV2Derivation) originDeposits(ctx context.Context, parent eth.L2BlockRef, origin eth.BlockID, receipts optypes.Receipts) (eth.BlockInfo, []hexutil.Bytes, []hexutil.Bytes, error) {
	header, err := cgtV2Header(ctx, d.source, origin)
	if err != nil {
		return nil, nil, nil, err
	}
	if origin.Number != parent.L1Origin.Number+1 || header.ParentHash != parent.L1Origin.Hash {
		return nil, nil, nil, NewResetError(fmt.Errorf("CGTv2 origin is not the next authenticated epoch"))
	}
	for _, receipt := range receipts {
		if receipt == nil {
			return nil, nil, nil, NewCriticalError(fmt.Errorf("CGTv2 missing origin receipt"))
		}
		// Celo's L1 origins are post-Byzantium. geth encodes every nonzero
		// Status as success and prefers PostState when present, so checking
		// receiptsRoot alone does not authenticate arbitrary RPC field values.
		if (receipt.Status != types.ReceiptStatusFailed && receipt.Status != types.ReceiptStatusSuccessful) || len(receipt.PostState) != 0 {
			return nil, nil, nil, NewCriticalError(fmt.Errorf("CGTv2 origin receipt has noncanonical status or poststate"))
		}
		for _, log := range receipt.Logs {
			if log == nil {
				return nil, nil, nil, NewCriticalError(fmt.Errorf("CGTv2 missing origin log"))
			}
		}
	}
	if types.DeriveSha(receipts, trie.NewStackTrie(nil)) != header.ReceiptHash {
		return nil, nil, nil, NewCriticalError(fmt.Errorf("CGTv2 receipts do not match origin %s", origin))
	}
	amount, err := cgtV2EscrowAmount(receipts, d.bridge, d.portal)
	if err != nil {
		return nil, nil, nil, NewCriticalError(err)
	}
	// BlockHash and Index are RPC metadata outside receiptsRoot. Reconstruct
	// them from the authenticated header and canonical receipt/log traversal.
	canonical := receipts.Geth()
	var index uint
	for i, receipt := range canonical {
		copyReceipt := *receipt
		copyReceipt.Logs = make([]*types.Log, len(receipt.Logs))
		for j, log := range receipt.Logs {
			copyLog := *log
			copyLog.BlockHash, copyLog.Index = origin.Hash, index
			index++
			copyReceipt.Logs[j] = &copyLog
		}
		canonical[i] = &copyReceipt
	}
	deposits, err := DeriveDeposits(canonical, d.portal)
	if err != nil {
		return nil, nil, nil, NewCriticalError(err)
	}
	var upgrades []hexutil.Bytes
	if amount != nil {
		upgrades, _, err = CgtV2UpgradeTransactions(d.chainID, amount)
		if err != nil {
			return nil, nil, nil, NewCriticalError(err)
		}
	}
	// Every Portal deposit in this origin executes before the CGTv2 group,
	// including zero-mint calls emitted after EscrowSeeded in L1 log order.
	return eth.HeaderBlockInfo(header), deposits, upgrades, nil
}

// CgtV2ActivationBlock checks the first L2 block of an epoch. All new origins
// use normal authenticated receipt data regardless of migration status.
func (ba *FetchingAttributesBuilder) CgtV2ActivationBlock(ctx context.Context, parent eth.L2BlockRef, epoch eth.BlockID) (bool, error) {
	if ba.cgt == nil || parent.L1Origin == epoch {
		return false, nil
	}
	_, receipts, err := ba.l1.FetchReceipts(ctx, epoch.Hash)
	if err != nil {
		return false, NewTemporaryError(fmt.Errorf("CGTv2 origin receipts: %w", err))
	}
	_, _, upgrades, err := ba.cgt.originDeposits(ctx, parent, epoch, receipts)
	return len(upgrades) > 0, err
}

// CgtV2InputSource supplies untrusted header bytes. No account/storage proof or
// L2 completion RPC is required. HeaderByHash normally reuses the L1 client's cache.
type CgtV2InputSource interface {
	HeaderByHash(context.Context, common.Hash) (*types.Header, error)
}

func cgtV2Header(ctx context.Context, rpc CgtV2InputSource, block eth.BlockID) (*types.Header, error) {
	if rpc == nil {
		return nil, NewCriticalError(fmt.Errorf("CGTv2 requires L1 header RPC support"))
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

var cgtV2EscrowEvent = crypto.Keccak256Hash([]byte("EscrowSeeded(address,uint256)"))

// Receipts must be authenticated before using this result. RPC log metadata is irrelevant.
// seedEscrowGenesis emits a different indexed caller, so only the configured Portal matches.
// The approved L1 migration seeds once; the first matching event supplies A.
func cgtV2EscrowAmount(receipts optypes.Receipts, bridge, portal common.Address) (*big.Int, error) {
	expectedPortal := common.BytesToHash(portal.Bytes())
	for _, receipt := range receipts {
		if receipt == nil {
			return nil, fmt.Errorf("CGTv2 missing origin receipt")
		}
		if receipt.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, log := range receipt.Logs {
			if log == nil {
				return nil, fmt.Errorf("CGTv2 missing origin log")
			}
			if log.Address != bridge || len(log.Topics) == 0 || log.Topics[0] != cgtV2EscrowEvent {
				continue
			}
			if len(log.Topics) != 2 || len(log.Data) != 32 {
				return nil, fmt.Errorf("malformed CGTv2 EscrowSeeded event")
			}
			if log.Topics[1] == expectedPortal {
				return new(big.Int).SetBytes(log.Data), nil
			}
		}
	}
	return nil, nil
}
