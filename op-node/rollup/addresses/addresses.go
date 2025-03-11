package addresses

import (
	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	"github.com/ethereum-optimism/optimism/op-service/predeploys"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type CeloAddresses struct {
	SuggestedFeeRecipient common.Address
}

// Map chainIDs to their respective addresses
var chainAddressMap = map[uint64]*CeloAddresses{
	params.CeloMainnetChainID: {
		SuggestedFeeRecipient: common.HexToAddress("0x7A1E98FC9a008107DbD1f430a05Ace8cf6f3FE19"),
	},
	params.CeloAlfajoresChainID: {
		SuggestedFeeRecipient: predeploys.SequencerFeeVaultAddr,
	},
	params.CeloBaklavaChainID: {
		SuggestedFeeRecipient: predeploys.SequencerFeeVaultAddr,
	},
	// for op-program tests
	params.OPMainnetChainID: {
		SuggestedFeeRecipient: predeploys.SequencerFeeVaultAddr,
	},
	chaincfg.OPSepolia().L2ChainID.Uint64(): {
		SuggestedFeeRecipient: predeploys.SequencerFeeVaultAddr,
	},
}

// GetAddressesOrDefault returns the addresses for the given chainID or
// the Mainnet addresses if none are found for the chainID.
func GetAddressesOrDefault(chainID uint64) *CeloAddresses {
	if addresses := chainAddressMap[chainID]; addresses != nil {
		return addresses
	}
	return chainAddressMap[params.CeloMainnetChainID]
}
