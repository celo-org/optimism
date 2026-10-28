package derive

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

func TestCgtV2GroupMatchesSharedVectors(t *testing.T) {
	raw, gas, err := CgtV2UpgradeTransactions(big.NewInt(1337), big.NewInt(600))
	require.NoError(t, err)
	require.Len(t, raw, 16)
	require.EqualValues(t, 40_600_000, gas)
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

func TestCgtV2BundleGasRoundTrip(t *testing.T) {
	bundle, err := cgtV2NUTBundle(cgtV2FixtureChainID)
	require.NoError(t, err)
	bundle.Transactions[0].GasLimit += 111
	original := cgtV2FixtureBundle
	t.Cleanup(func() { cgtV2FixtureBundle = original })
	cgtV2FixtureBundle, err = json.Marshal(bundle)
	require.NoError(t, err)
	receipts := cgtSeededReceipts()
	_, source, parent, origin := cgtTestDeriver(t, receipts)
	cfg := cgtTestConfig()
	builder := NewFetchingAttributesBuilder(cfg, params.MergedTestChainConfig, nil, &cgtL1Stub{source.info, receipts}, cgtL2Stub{}, source)
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, origin)
	require.NoError(t, err)
	require.EqualValues(t, 70_600_111, *attrs.GasLimit)
	payload := &eth.ExecutionPayload{BlockNumber: 101, Timestamp: attrs.Timestamp, GasLimit: *attrs.GasLimit, Transactions: attrs.Transactions}
	recovered, err := PayloadToSystemConfig(cfg, payload)
	require.NoError(t, err)
	require.EqualValues(t, 30_000_000, recovered.GasLimit)
}

func TestCgtV2ProductionBundleSelection(t *testing.T) {
	for _, id := range []uint64{params.CeloMainnetChainID, params.CeloSepoliaChainID, params.CeloChaosChainID} {
		t.Run(fmt.Sprint(id), func(t *testing.T) {
			original := cgtV2Networks[id]
			t.Cleanup(func() { cgtV2Networks[id] = original })
			// Install a mock artifact through the production profile, without changing the fixture profile.
			network := original
			network.bundle = cgtV2FixtureBundle
			cgtV2Networks[id] = network
			txs, gas, err := CgtV2UpgradeTransactions(new(big.Int).SetUint64(id), big.NewInt(600))
			require.NoError(t, err)
			require.Len(t, txs, 16)
			require.EqualValues(t, 40_600_000, gas)
			funding, err := optypes.UnmarshalDepositTx(txs[15])
			require.NoError(t, err)
			supply, _ := new(big.Int).SetString("1000000000000000000000000000", 10)
			require.Equal(t, new(big.Int).Sub(supply, big.NewInt(600)), funding.Value)
			network.bundle = []byte(`{"metadata":{"version":"1.0.0"},"transactions":[]}`)
			cgtV2Networks[id] = network
			_, _, err = CgtV2UpgradeTransactions(new(big.Int).SetUint64(id), big.NewInt(600))
			require.ErrorContains(t, err, "unresolved")
		})
	}
}

func TestCgtV2BundleRejectsInvalidShapeAndGas(t *testing.T) {
	for _, kind := range []string{"count", "intent", "zero-gas", "overflow"} {
		t.Run(kind, func(t *testing.T) {
			bundle, err := cgtV2NUTBundle(cgtV2FixtureChainID)
			require.NoError(t, err)
			switch kind {
			case "count":
				bundle.Transactions = bundle.Transactions[:14]
			case "intent":
				bundle.Transactions[0].Intent = "wrong"
			case "zero-gas":
				bundle.Transactions[0].GasLimit = 0
			case "overflow":
				bundle.Transactions[0].GasLimit = ^uint64(0)
			}
			original := cgtV2FixtureBundle
			t.Cleanup(func() { cgtV2FixtureBundle = original })
			cgtV2FixtureBundle, err = json.Marshal(bundle)
			require.NoError(t, err)
			_, _, err = CgtV2UpgradeTransactions(big.NewInt(1337), big.NewInt(600))
			require.Error(t, err)
		})
	}
}
