package derive

import "github.com/ethereum/go-ethereum/common"

// Synthetic profile matching the frozen contract/NUT fixtures, not a production default.
const cgtV2FixtureChainID uint64 = 1337

var cgtV2FixtureParams = cgtV2NetworkParams{
	bridge:      common.HexToAddress("0xcc"),
	totalSupply: "1000",
}
