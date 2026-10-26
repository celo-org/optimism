package derive

import (
	"fmt"
	"math/big"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

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
