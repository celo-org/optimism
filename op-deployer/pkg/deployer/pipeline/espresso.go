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

	// Read the underlying AWS NitroEnclaveVerifier address (from Automata).
	// If not set, the zero address triggers mock verifier deployment — dev/test only.
	var nitroEnclaveVerifierAddress common.Address
	if envVar := os.Getenv("NITRO_ENCLAVE_VERIFIER_ADDRESS"); envVar != "" {
		nitroEnclaveVerifierAddress = common.HexToAddress(envVar)
		lgr.Info("using nitro enclave verifier from NITRO_ENCLAVE_VERIFIER_ADDRESS", "address", nitroEnclaveVerifierAddress.Hex())
	} else {
		lgr.Info("NITRO_ENCLAVE_VERIFIER_ADDRESS not set — deploying mock TEE verifiers")
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
