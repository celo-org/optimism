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
)

// ContainsCgtV2Activation recognizes any reserved group source. Execution checks
// every required transaction field, order, success and group completeness.
func ContainsCgtV2Activation(txs []hexutil.Bytes) bool { return containsCgtV2Group(txs) }

// FetchReceipts authenticates the receipt root and log metadata. These additional
// checks ensure the post-Byzantium receipt fields used by the event scan are canonical.
func cgtV2UpgradeTransactionsFromReceipts(cfg *rollup.Config, receipts optypes.Receipts) ([]hexutil.Bytes, uint64, error) {
	if cfg.L2ChainID == nil || !cfg.L2ChainID.IsUint64() {
		return nil, 0, nil
	}
	chainID := bigs.Uint64Strict(cfg.L2ChainID)
	network, ok := cgtV2Params(chainID)
	if !ok {
		return nil, 0, nil
	}
	for _, receipt := range receipts {
		if receipt == nil {
			return nil, 0, fmt.Errorf("CGTv2 missing origin receipt")
		}
		if (receipt.Status != types.ReceiptStatusFailed && receipt.Status != types.ReceiptStatusSuccessful) || len(receipt.PostState) != 0 {
			return nil, 0, fmt.Errorf("CGTv2 origin receipt has noncanonical status or poststate")
		}
		for _, log := range receipt.Logs {
			if log == nil {
				return nil, 0, fmt.Errorf("CGTv2 missing origin log")
			}
		}
	}
	amount, retry, err := cgtV2EscrowAmount(receipts, network, cfg.DepositContractAddress)
	if err != nil || amount == nil {
		return nil, 0, err
	}
	return network.upgradeTransactions(chainID, amount, retry)
}

var cgtV2EscrowEvent = crypto.Keccak256Hash([]byte("EscrowSeeded(address,uint256)"))
var cgtV2RetryEvent = crypto.Keccak256Hash([]byte("CgtV2RetryRequested(bytes32,uint256)"))

// Receipts must be authenticated before using this result. RPC log metadata is irrelevant.
// seedEscrowGenesis emits a different indexed caller, so only the configured Portal matches.
// Retry requests must name a registered bundle. The first matching event supplies the migrated amount.
func cgtV2EscrowAmount(receipts optypes.Receipts, network cgtV2NetworkParams, portal common.Address) (amount *big.Int, retry *cgtV2RetryBundle, err error) {
	expectedPortal := common.BytesToHash(portal.Bytes())
	for _, receipt := range receipts {
		if receipt == nil {
			return nil, nil, fmt.Errorf("CGTv2 missing origin receipt")
		}
		if receipt.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, log := range receipt.Logs {
			if log == nil {
				return nil, nil, fmt.Errorf("CGTv2 missing origin log")
			}
			if log.Address != network.bridge || len(log.Topics) == 0 {
				continue
			}
			switch log.Topics[0] {
			case cgtV2EscrowEvent:
				if len(log.Topics) != 2 || len(log.Data) != 32 {
					return nil, nil, fmt.Errorf("malformed CGTv2 EscrowSeeded event")
				}
				if log.Topics[1] == expectedPortal {
					return new(big.Int).SetBytes(log.Data), nil, nil
				}
			case cgtV2RetryEvent:
				if len(log.Topics) != 2 || len(log.Data) != 32 {
					return nil, nil, fmt.Errorf("malformed CGTv2 retry event")
				}
				retry := network.retryBundle(log.Topics[1])
				if retry == nil {
					oplog.Error("Skipping CGTv2 retry request for an unregistered bundle",
						"l1_origin", log.BlockNumber,
						"l1_block_hash", log.BlockHash,
						"l1_tx_hash", log.TxHash,
						"requested_bundle_hash", log.Topics[1])
					continue
				}
				return new(big.Int).SetBytes(log.Data), retry, nil
			}
		}
	}
	return nil, nil, nil
}
