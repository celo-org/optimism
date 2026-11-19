package pipeline

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/opcm"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum/go-ethereum/common"
)

// DeployEspresso deploys the Espresso BatchAuthenticator + TEE verifier contracts for a chain
// whose intent has EspressoEnabled. It is a no-op for chains without Espresso enabled.
//
// It targets the redesigned scripts/deploy/DeployEspresso.s.sol (PR #455), whose inputs are
// espressoOwner + sharedProxyAdmin (the Espresso proxies are handed to the existing shared
// OP Stack ProxyAdmin rather than dedicated ones).
func DeployEspresso(env *Env, intent *state.Intent, st *state.State, chainID common.Hash) error {
	lgr := env.Logger.New("stage", "deploy-espresso")

	chainIntent, err := intent.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain intent: %w", err)
	}

	chainState, err := st.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain state: %w", err)
	}

	if !chainIntent.EspressoEnabled {
		lgr.Info("espresso not enabled, skipping BatchAuthenticator deployment")
		return nil
	}

	lgr.Info("deploying espresso contracts")

	nitroEnclaveVerifierAddress, err := nitroEnclaveVerifier(env)
	if err != nil {
		return err
	}
	if nitroEnclaveVerifierAddress == (common.Address{}) {
		lgr.Warn("NITRO_ENCLAVE_VERIFIER_ADDRESS not set — deploying mock TEE verifiers (dev/test only)")
	} else {
		lgr.Info("using nitro enclave verifier from NITRO_ENCLAVE_VERIFIER_ADDRESS", "address", nitroEnclaveVerifierAddress.Hex())
	}

	// The application-level owner gating operational setters. Defaults to the deployer.
	espressoOwner := env.Deployer
	if envVar := os.Getenv("BATCH_AUTHENTICATOR_OWNER_ADDRESS"); envVar != "" {
		espressoOwner = common.HexToAddress(envVar)
		lgr.Info("using espresso owner from BATCH_AUTHENTICATOR_OWNER_ADDRESS", "address", espressoOwner.Hex())
	} else {
		lgr.Info("using deployer as espresso owner", "address", espressoOwner.Hex())
	}

	// The Espresso proxies are handed to the existing shared OP Stack ProxyAdmin.
	sharedProxyAdmin := chainState.OpChainProxyAdminImpl
	if sharedProxyAdmin == (common.Address{}) {
		return fmt.Errorf("shared OP Stack ProxyAdmin not deployed for chain %s", chainID.Hex())
	}

	eo, err := opcm.DeployEspresso(env.L1ScriptHost, opcm.DeployEspressoInput{
		NitroEnclaveVerifier: nitroEnclaveVerifierAddress,
		EspressoBatcher:      chainIntent.EspressoBatcher,
		SystemConfig:         chainState.SystemConfigProxy,
		EspressoOwner:        espressoOwner,
		SharedProxyAdmin:     sharedProxyAdmin,
	}, espressoOwner)
	if err != nil {
		return fmt.Errorf("failed to deploy espresso contracts: %w", err)
	}

	chainState.BatchAuthenticatorAddress = eo.BatchAuthenticatorAddress
	lgr.Info("espresso contracts deployed",
		"batchAuthenticator", eo.BatchAuthenticatorAddress,
		"teeVerifier", eo.TeeVerifierProxy,
		"nitroTEEVerifier", eo.NitroTEEVerifier,
	)
	return nil
}

// allowMockTEEVerifierEnv opts a non-genesis deployment into the mock TEE verifiers.
const allowMockTEEVerifierEnv = "ESPRESSO_ALLOW_MOCK_TEE_VERIFIER"

// nitroEnclaveVerifier returns the underlying AWS NitroEnclaveVerifier address (from Automata)
// read from NITRO_ENCLAVE_VERIFIER_ADDRESS. The zero address makes DeployEspresso.s.sol deploy
// permissive mock verifiers, which disables the Nitro attestation check; it is only returned
// when the variable is unset and either the deployment is in-memory (genesis target) or
// ESPRESSO_ALLOW_MOCK_TEE_VERIFIER=true. A malformed or zero address is always an error.
func nitroEnclaveVerifier(env *Env) (common.Address, error) {
	raw := os.Getenv("NITRO_ENCLAVE_VERIFIER_ADDRESS")
	if raw == "" {
		if env.AllowMockTEEVerifier || os.Getenv(allowMockTEEVerifierEnv) == "true" {
			return common.Address{}, nil
		}
		return common.Address{}, fmt.Errorf("NITRO_ENCLAVE_VERIFIER_ADDRESS is not set; set it to the Nitro enclave verifier address, or set %s=true to deploy mock TEE verifiers (dev/test only)", allowMockTEEVerifierEnv)
	}
	if !common.IsHexAddress(raw) {
		return common.Address{}, fmt.Errorf("NITRO_ENCLAVE_VERIFIER_ADDRESS is not a valid address: %q", raw)
	}
	addr := common.HexToAddress(raw)
	if addr == (common.Address{}) {
		return common.Address{}, fmt.Errorf("NITRO_ENCLAVE_VERIFIER_ADDRESS must not be the zero address; unset it and set %s=true to deploy mock TEE verifiers", allowMockTEEVerifierEnv)
	}
	return addr, nil
}
