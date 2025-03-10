package addresses

import (
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/predeploys"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type CeloAddresses struct {
	SuggestedFeeRecipient common.Address
}

var (
	MainnetAddresses = &CeloAddresses{
		SuggestedFeeRecipient: common.HexToAddress("0xaa"), // TODO: replace it with the actual one
	}

	AlfajoresAddresses = &CeloAddresses{
		SuggestedFeeRecipient: predeploys.SequencerFeeVaultAddr,
	}

	BaklavaAddresses = &CeloAddresses{
		SuggestedFeeRecipient: predeploys.SequencerFeeVaultAddr,
	}
)

// getAddresses returns the addresses for the given chainID or
// nil if not found.
func getAddresses(chainID *big.Int) *CeloAddresses {
	// ChainID can be uninitialized in some tests
	if chainID == nil {
		return nil
	}
	switch chainID.Uint64() {
	case params.CeloAlfajoresChainID:
		return AlfajoresAddresses
	case params.CeloBaklavaChainID:
		return BaklavaAddresses
	case params.CeloMainnetChainID:
		return MainnetAddresses
	default:
		return nil
	}
}

// GetAddressesOrDefault returns the addresses for the given chainID or
// the default addresses if none are found.
func GetAddressesOrDefault(chainID *big.Int, defaultValue *CeloAddresses) *CeloAddresses {
	addresses := getAddresses(chainID)
	if addresses == nil {
		return defaultValue
	}
	return addresses
}
