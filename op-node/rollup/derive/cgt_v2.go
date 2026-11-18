package derive

import (
	"fmt"
	"math/big"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// ContainsCgtV2Activation recognizes any reserved group source. Execution checks
// every required transaction field, order, success and group completeness.
func ContainsCgtV2Activation(txs []hexutil.Bytes) bool { return containsCgtV2Group(txs) }

// FetchReceipts authenticates the receipt root and log metadata. These additional
// checks ensure the post-Byzantium receipt fields used by the event scan are canonical.
func cgtV2UpgradeTransactionsFromReceipts(cfg *rollup.Config, receipts optypes.Receipts) ([]hexutil.Bytes, error) {
	if cfg.L2ChainID == nil || !cfg.L2ChainID.IsUint64() {
		return nil, nil
	}
	network, ok := cgtV2Params(bigs.Uint64Strict(cfg.L2ChainID))
	if !ok {
		return nil, nil
	}
	for _, receipt := range receipts {
		if receipt == nil {
			return nil, fmt.Errorf("CGTv2 missing origin receipt")
		}
		if (receipt.Status != types.ReceiptStatusFailed && receipt.Status != types.ReceiptStatusSuccessful) || len(receipt.PostState) != 0 {
			return nil, fmt.Errorf("CGTv2 origin receipt has noncanonical status or poststate")
		}
		for _, log := range receipt.Logs {
			if log == nil {
				return nil, fmt.Errorf("CGTv2 missing origin log")
			}
		}
	}
	amount, err := cgtV2EscrowAmount(receipts, network.bridge, cfg.DepositContractAddress)
	if err != nil || amount == nil {
		return nil, err
	}
	upgrades, _, err := CgtV2UpgradeTransactions(cfg.L2ChainID, amount)
	return upgrades, err
}

var cgtV2EscrowEvent = crypto.Keccak256Hash([]byte("EscrowSeeded(address,uint256)"))
var cgtV2RetryEvent = crypto.Keccak256Hash([]byte("CgtV2RetryRequested(bytes32,uint256)"))

// The initial activation scan validates the receipts before this call. A retry
// is considered only at the first L2 block of a new origin.
func cgtV2RetryTransactions(cfg *rollup.Config, receipts optypes.Receipts) ([]hexutil.Bytes, error) {
	chainID := cfg.L2ChainID
	if chainID == nil || !chainID.IsUint64() {
		return nil, nil
	}
	network, ok := cgtV2Params(bigs.Uint64Strict(chainID))
	if !ok {
		return nil, nil
	}
	for _, receipt := range receipts {
		if receipt.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, log := range receipt.Logs {
			if log.Address != network.bridge || len(log.Topics) == 0 || log.Topics[0] != cgtV2RetryEvent {
				continue
			}
			if len(log.Topics) != 2 || len(log.Data) != 32 {
				return nil, NewCriticalError(fmt.Errorf("malformed CGTv2 retry event"))
			}
			if network.retryBundleHash == (common.Hash{}) || log.Topics[1] != network.retryBundleHash {
				oplog.Warn("Skipping CGTv2 retry request with an unrecognized bundle hash",
					"l1_block_hash", log.BlockHash,
					"l1_tx_hash", log.TxHash,
					"requested_bundle_hash", log.Topics[1],
					"expected_bundle_hash", network.retryBundleHash)
				continue
			}
			txs, _, err := CgtV2UpgradeTransactions(chainID, new(big.Int).SetBytes(log.Data))
			if err != nil {
				return nil, NewCriticalError(err)
			}
			// The static RLP artifact matches Kona. Funding is derived separately from the event.
			static, err := rlp.EncodeToBytes(txs[:len(txs)-1])
			if err != nil {
				return nil, NewCriticalError(err)
			}
			if crypto.Keccak256Hash(static) != network.retryBundleHash {
				return nil, NewCriticalError(fmt.Errorf("local CGTv2 retry artifact does not match protocol hash %s", network.retryBundleHash))
			}
			return txs, nil
		}
	}
	return nil, nil
}

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
