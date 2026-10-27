package derive

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

func TestCgtV2GroupMatchesSharedVectors(t *testing.T) {
	raw, gas, err := CgtV2UpgradeTransactions(big.NewInt(1337), big.NewInt(600))
	require.NoError(t, err)
	require.Len(t, raw, 20)
	require.EqualValues(t, 48_100_000, gas)
	var vectors []struct {
		SourceHash      common.Hash
		TransactionHash common.Hash
		File            string
	}
	data, err := os.ReadFile("testdata/cgtv2/cgtv2-vectors.json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &vectors))
	for i, encoded := range raw {
		frozen, err := os.ReadFile("testdata/cgtv2/" + vectors[i].File)
		require.NoError(t, err)
		require.Equal(t, frozen, []byte(encoded))
		tx, err := optypes.UnmarshalDepositTx(encoded)
		require.NoError(t, err)
		require.Equal(t, vectors[i].SourceHash, tx.SourceHash)
		require.Equal(t, vectors[i].TransactionHash, crypto.Keccak256Hash(encoded))
	}
	for _, amount := range []*big.Int{nil, big.NewInt(-1), big.NewInt(1001)} {
		_, _, err := CgtV2UpgradeTransactions(big.NewInt(1337), amount)
		require.Error(t, err)
	}
	raw, _, err = CgtV2UpgradeTransactions(big.NewInt(1337), big.NewInt(1000))
	require.NoError(t, err)
	tx, err := optypes.UnmarshalDepositTx(raw[len(raw)-1])
	require.NoError(t, err)
	require.Zero(t, tx.Mint.Sign())
	require.Zero(t, tx.Value.Sign())
	_, _, err = CgtV2UpgradeTransactions(big.NewInt(42220), big.NewInt(600))
	require.Error(t, err)
}

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
