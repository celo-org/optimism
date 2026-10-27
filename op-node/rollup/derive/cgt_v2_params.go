package derive

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

// CeloTokenL1.TOTAL_MARKET_CAP: 1 billion CELO, with 18 decimals.
// Source: celo-contracts/v6.0.0 at 0296412d673a666af6236efa1f615032b1f1c424.
const cgtV2TotalSupply = "1000000000000000000000000000"

// TODO(CGTv2): replace for each network after its L1 bridge deployment.
// This temporary address is not a deployed production bridge.
var cgtV2TemporaryL1Bridge = common.HexToAddress("0x00000000000000000000000000000000c6700001")

type cgtV2NetworkParams struct {
	bridge      common.Address
	totalSupply string
}

// Production profiles never select the synthetic NUT bundle.
var cgtV2Networks = map[uint64]cgtV2NetworkParams{
	params.CeloMainnetChainID: {bridge: cgtV2TemporaryL1Bridge, totalSupply: cgtV2TotalSupply},
	params.CeloSepoliaChainID: {bridge: cgtV2TemporaryL1Bridge, totalSupply: cgtV2TotalSupply},
	params.CeloChaosChainID:   {bridge: cgtV2TemporaryL1Bridge, totalSupply: cgtV2TotalSupply},
}

func cgtV2Params(chainID uint64) (cgtV2NetworkParams, bool) {
	if chainID == cgtV2FixtureChainID {
		return cgtV2FixtureParams, true
	}
	network, ok := cgtV2Networks[chainID]
	return network, ok
}

func (p cgtV2NetworkParams) reserveSeed(migratedAmount *big.Int) (*big.Int, bool) {
	supply, ok := new(big.Int).SetString(p.totalSupply, 10)
	if !ok || migratedAmount == nil || migratedAmount.Sign() < 0 || migratedAmount.Cmp(supply) > 0 {
		return nil, false
	}
	return supply.Sub(supply, migratedAmount), true
}
