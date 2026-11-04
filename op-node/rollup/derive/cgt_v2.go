package derive

import (
	"fmt"
	"math/big"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
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
