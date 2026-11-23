package opcm

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-chain-ops/script"
	"github.com/ethereum/go-ethereum/common"
)

// DeployEspressoInput mirrors the DeployEspressoInput contract in
// scripts/deploy/DeployEspresso.s.sol. The field names (and thus the generated
// setter selectors) must match the contract's getters.
type DeployEspressoInput struct {
	// NitroEnclaveVerifier is the underlying AWS Nitro enclave verifier (from Automata).
	// Set to the zero address to deploy mock verifiers (dev/test only).
	NitroEnclaveVerifier common.Address
	// EspressoBatcher is the TEE batcher EOA authorized to call the BatchAuthenticator.
	EspressoBatcher common.Address
	// SystemConfig is the chain's SystemConfig proxy address.
	SystemConfig common.Address
	// EspressoOwner is the application-level (OZ Ownable) owner gating operational setters.
	EspressoOwner common.Address
	// SharedProxyAdmin is the existing OP Stack ProxyAdmin the Espresso proxies are handed to.
	SharedProxyAdmin common.Address
}

// DeployEspressoOutput mirrors the DeployEspressoOutput contract.
type DeployEspressoOutput struct {
	BatchAuthenticatorAddress common.Address
	TeeVerifierProxy          common.Address
	NitroTEEVerifier          common.Address
}

// DeployEspressoScript is the typed handle to the DeployEspresso script's run
// function, which takes (input, output, deployerAddress).
type DeployEspressoScript struct {
	Run func(input, output, deployerAddress common.Address) error
}

func DeployEspresso(
	host *script.Host,
	input DeployEspressoInput,
	deployerAddress common.Address,
) (DeployEspressoOutput, error) {
	var output DeployEspressoOutput
	inputAddr := host.NewScriptAddress()
	outputAddr := host.NewScriptAddress()

	cleanupInput, err := script.WithPrecompileAtAddress(host, inputAddr, &input)
	if err != nil {
		return output, fmt.Errorf("failed to insert DeployEspressoInput precompile: %w", err)
	}
	defer cleanupInput()

	cleanupOutput, err := script.WithPrecompileAtAddress(host, outputAddr, &output,
		script.WithFieldSetter[*DeployEspressoOutput])
	if err != nil {
		return output, fmt.Errorf("failed to insert DeployEspressoOutput precompile: %w", err)
	}
	defer cleanupOutput()

	implContract := "DeployEspresso"
	deployScript, cleanupDeploy, err := script.WithScript[DeployEspressoScript](host, "DeployEspresso.s.sol", implContract)
	if err != nil {
		return output, fmt.Errorf("failed to load %s script: %w", implContract, err)
	}
	defer cleanupDeploy()

	if err := deployScript.Run(inputAddr, outputAddr, deployerAddress); err != nil {
		return output, fmt.Errorf("failed to run %s script: %w", implContract, err)
	}

	return output, nil
}
