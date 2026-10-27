package derive

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

func TestCgtV2NetworkSupplyAndBounds(t *testing.T) {
	supply, ok := new(big.Int).SetString("1000000000000000000000000000", 10)
	require.True(t, ok)
	for _, id := range []uint64{params.CeloMainnetChainID, params.CeloSepoliaChainID, params.CeloChaosChainID} {
		network, ok := cgtV2Params(id)
		require.True(t, ok)
		require.NotEqual(t, common.Address{}, network.bridge)
		require.NotEqual(t, cgtV2FixtureParams.bridge, network.bridge)
		reserve, ok := network.reserveSeed(big.NewInt(0))
		require.True(t, ok)
		require.Equal(t, supply, reserve)
		reserve.SetInt64(0)
		reserve, ok = network.reserveSeed(supply)
		require.True(t, ok)
		require.Zero(t, reserve.Sign())
		_, ok = network.reserveSeed(new(big.Int).Add(supply, big.NewInt(1)))
		require.False(t, ok)
		_, _, err := CgtV2UpgradeTransactions(new(big.Int).SetUint64(id), big.NewInt(0))
		require.ErrorContains(t, err, "upgrade bundle is unresolved")
	}
	fixture, ok := cgtV2Params(cgtV2FixtureChainID)
	require.True(t, ok)
	reserve, ok := fixture.reserveSeed(big.NewInt(600))
	require.True(t, ok)
	require.Equal(t, big.NewInt(400), reserve)
	for _, id := range []uint64{1, 44787, 62320} {
		_, ok := cgtV2Params(id)
		require.False(t, ok)
	}
}
