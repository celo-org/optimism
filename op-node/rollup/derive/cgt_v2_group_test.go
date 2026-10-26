package derive

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestCgtV2GroupMatchesSharedVectors(t *testing.T) {
	raw, gas, err := CgtV2UpgradeTransactions(big.NewInt(1337), big.NewInt(600))
	require.NoError(t, err)
	require.Len(t, raw, 13)
	require.EqualValues(t, 56_300_000, gas)
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
	tx, err := optypes.UnmarshalDepositTx(raw[12])
	require.NoError(t, err)
	require.Zero(t, tx.Mint.Sign())
	require.Zero(t, tx.Value.Sign())
	_, _, err = CgtV2UpgradeTransactions(big.NewInt(42220), big.NewInt(600))
	require.Error(t, err)
}
