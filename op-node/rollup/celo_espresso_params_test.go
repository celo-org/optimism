package rollup

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
)

// chaosBatchAuthenticator is the canonical BatchAuthenticator address for Celo Chaos, duplicated
// here as a tripwire: changing celo_espresso_params.json must be a deliberate act that also
// updates this test (and the baked constants in celo-kona's crates/kona/proof/src/boot.rs).
var chaosBatchAuthenticator = common.HexToAddress("0xb4B5343d9635b05cA4FbdB09BB4929E21A1A8B37")

// TestCeloEspressoParams_CanonicalFile pins the parsed content of the embedded
// celo_espresso_params.json: exactly the three known Celo chains, with Espresso scheduled only
// on Chaos.
func TestCeloEspressoParams_CanonicalFile(t *testing.T) {
	require.Len(t, celoEspressoParamsByChainID, 3)

	for _, chainID := range []uint64{42220, 11142220} {
		params, ok := celoEspressoParamsByChainID[chainID]
		require.True(t, ok, "missing canonical entry for chain %d", chainID)
		require.Nil(t, params.EspressoTime, "Espresso must be unscheduled on chain %d", chainID)
		require.Nil(t, params.BatchAuthenticatorAddress, "no authenticator expected on chain %d", chainID)
	}

	chaos, ok := celoEspressoParamsByChainID[11162320]
	require.True(t, ok, "missing canonical entry for Celo Chaos")
	require.NotNil(t, chaos.EspressoTime)
	require.EqualValues(t, 1782910800, *chaos.EspressoTime)
	require.NotNil(t, chaos.BatchAuthenticatorAddress)
	require.Equal(t, chaosBatchAuthenticator, *chaos.BatchAuthenticatorAddress)
}

// TestConfig_Check_CeloEspressoParams verifies that Check rejects Espresso parameters that
// diverge from the canonical celo_espresso_params.json entry for the known Celo chain IDs, and
// leaves other chains alone.
func TestConfig_Check_CeloEspressoParams(t *testing.T) {
	// configFor builds an otherwise-valid config for the given chain with the given Espresso
	// parameters. Ecotone (and its predecessor forks) activate at genesis, as on the real Celo
	// chains, so the espresso-after-ecotone rule is satisfied whenever espressoTime is set.
	configFor := func(chainID uint64, espressoTime *uint64, authenticator common.Address) *Config {
		cfg := randConfig()
		zero := uint64(0)
		cfg.RegolithTime = &zero
		cfg.CanyonTime = &zero
		cfg.DeltaTime = &zero
		cfg.EcotoneTime = &zero
		cfg.L2ChainID = new(big.Int).SetUint64(chainID)
		cfg.EspressoTime = espressoTime
		cfg.BatchAuthenticatorAddress = authenticator
		return cfg
	}

	chaosTime := uint64(1782910800)

	// The canonical parameters pass for each known chain.
	require.NoError(t, configFor(42220, nil, common.Address{}).Check())
	require.NoError(t, configFor(11142220, nil, common.Address{}).Check())
	require.NoError(t, configFor(11162320, &chaosTime, chaosBatchAuthenticator).Check())

	// Chaos: missing espresso_time, wrong espresso_time, or wrong authenticator all diverge.
	err := configFor(11162320, nil, chaosBatchAuthenticator).Check()
	require.ErrorIs(t, err, ErrCeloEspressoParamsMismatch)
	wrongTime := chaosTime + 1
	err = configFor(11162320, &wrongTime, chaosBatchAuthenticator).Check()
	require.ErrorIs(t, err, ErrCeloEspressoParamsMismatch)
	err = configFor(11162320, &chaosTime, common.Address{0x01}).Check()
	require.ErrorIs(t, err, ErrCeloEspressoParamsMismatch)

	// Mainnet/Sepolia: scheduling Espresso (or carrying an authenticator) without updating the
	// canonical file is rejected.
	err = configFor(42220, &chaosTime, chaosBatchAuthenticator).Check()
	require.ErrorIs(t, err, ErrCeloEspressoParamsMismatch)
	err = configFor(11142220, nil, chaosBatchAuthenticator).Check()
	require.ErrorIs(t, err, ErrCeloEspressoParamsMismatch)

	// A chain without a canonical entry keeps whatever the config carries.
	require.NoError(t, configFor(901, &chaosTime, chaosBatchAuthenticator).Check())
	require.NoError(t, configFor(901, nil, common.Address{}).Check())

	// A non-uint64 L2 chain ID cannot have a canonical entry and is not rejected here.
	huge, ok := new(big.Int).SetString("340282366920938463463374607431768211456", 10) // 2^128
	require.True(t, ok)
	cfg := configFor(901, nil, common.Address{})
	cfg.L2ChainID = huge
	require.NoError(t, cfg.checkCeloEspressoParams())
}
