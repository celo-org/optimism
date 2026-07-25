package rollup

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
)

// celoEspressoParamsJSON is the canonical source of truth for the Espresso batch-authentication
// parameters of the known Celo chains, keyed by L2 chain ID. The Espresso parameters are
// consensus-critical: espresso_time switches batch authorization from sender-based to
// BatchAuthenticator event-based, so op-node and the fault-proof program must agree on them or
// the two derivation pipelines produce divergent outputs from the same inputs.
//
// The file is mirrored byte-for-byte in celo-kona at crates/kona/proof/src/celo_espresso_params.json,
// where a unit test asserts the program-baked constants (CELO_MAINNET_ESPRESSO /
// CELO_SEPOLIA_ESPRESSO / CELO_CHAOS_ESPRESSO in crates/kona/proof/src/boot.rs) match it. When
// scheduling Espresso on a chain, update the file in both repositories together with the baked
// constants in celo-kona.
//
//go:embed celo_espresso_params.json
var celoEspressoParamsJSON []byte

// celoEspressoParams is one chain's entry in celo_espresso_params.json. Both fields are nil for
// a chain on which Espresso is not (yet) scheduled.
type celoEspressoParams struct {
	EspressoTime              *uint64         `json:"espresso_time"`
	BatchAuthenticatorAddress *common.Address `json:"batch_authenticator_address"`
}

// celoEspressoParamsByChainID is celo_espresso_params.json parsed at package load. The embedded
// file is part of the build, so any malformed or internally inconsistent content is a programming
// error and panics rather than being reported at config-load time.
var celoEspressoParamsByChainID = func() map[uint64]celoEspressoParams {
	var raw map[string]celoEspressoParams
	if err := json.Unmarshal(celoEspressoParamsJSON, &raw); err != nil {
		panic(fmt.Errorf("invalid embedded celo_espresso_params.json: %w", err))
	}
	byChainID := make(map[uint64]celoEspressoParams, len(raw))
	for key, params := range raw {
		chainID, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			panic(fmt.Errorf("invalid chain ID %q in embedded celo_espresso_params.json: %w", key, err))
		}
		if (params.EspressoTime == nil) != (params.BatchAuthenticatorAddress == nil) {
			panic(fmt.Errorf("chain %d in embedded celo_espresso_params.json must set espresso_time and batch_authenticator_address together", chainID))
		}
		if params.BatchAuthenticatorAddress != nil && *params.BatchAuthenticatorAddress == (common.Address{}) {
			panic(fmt.Errorf("chain %d in embedded celo_espresso_params.json has a zero batch_authenticator_address", chainID))
		}
		byChainID[chainID] = params
	}
	return byChainID
}()

// ErrCeloEspressoParamsMismatch is returned by Check when the loaded rollup config carries
// Espresso parameters that differ from the canonical celo_espresso_params.json entry for the
// chain. The fault-proof program derives the known Celo chains with the canonical values baked
// in, so a diverging op-node config would derive a different chain than the proofs attest to.
var ErrCeloEspressoParamsMismatch = errors.New("espresso params diverge from the canonical celo_espresso_params.json entry for this chain")

// checkCeloEspressoParams validates the config's Espresso parameters against the canonical
// per-chain values embedded from celo_espresso_params.json. A no-op for chains without a
// canonical entry (devnets, non-Celo chains), which keep whatever the config carries.
func (cfg *Config) checkCeloEspressoParams() error {
	if cfg.L2ChainID == nil || !cfg.L2ChainID.IsUint64() {
		return nil
	}
	canonical, ok := celoEspressoParamsByChainID[cfg.L2ChainID.Uint64()]
	if !ok {
		return nil
	}
	fmtTime := func(t *uint64) string {
		if t == nil {
			return "unset"
		}
		return strconv.FormatUint(*t, 10)
	}
	if (cfg.EspressoTime == nil) != (canonical.EspressoTime == nil) ||
		(cfg.EspressoTime != nil && *cfg.EspressoTime != *canonical.EspressoTime) {
		return fmt.Errorf("%w: espresso_time is %s, canonical is %s",
			ErrCeloEspressoParamsMismatch, fmtTime(cfg.EspressoTime), fmtTime(canonical.EspressoTime))
	}
	canonicalAddr := common.Address{} // canonical entries carry no address while Espresso is unscheduled
	if canonical.BatchAuthenticatorAddress != nil {
		canonicalAddr = *canonical.BatchAuthenticatorAddress
	}
	if cfg.BatchAuthenticatorAddress != canonicalAddr {
		return fmt.Errorf("%w: batch_authenticator_address is %s, canonical is %s",
			ErrCeloEspressoParamsMismatch, cfg.BatchAuthenticatorAddress, canonicalAddr)
	}
	return nil
}
