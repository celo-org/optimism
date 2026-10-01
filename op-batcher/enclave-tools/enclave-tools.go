package enclave_tools

import (
	"context"
	"crypto/ecdsa"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/bindings/batchauthenticator"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EnclaveMeasurements struct {
	PCR0 string `json:"PCR0"`
	PCR1 string `json:"PCR1"`
	PCR2 string `json:"PCR2"`
}

// Builds docker and enclaver EIF image for op-batcher and registers EIF's PCR0 with
// EspressoNitroTEEVerifier. args... are command-line arguments to op-batcher
// to be baked into the image.
func BuildBatcherImage(ctx context.Context, opRoot string, tag string, cpuCount uint, memoryMb uint, args ...string) (EnclaveMeasurements, error) {
	intermediateTag := tag + "intermediate"

	cmd := exec.CommandContext(ctx, "docker",
		"build",
		"--tag", intermediateTag,
		"--file", filepath.Join(opRoot, "ops/docker/op-stack-go/Dockerfile"),
		"--target", "op-batcher-enclave-target",
		"--build-arg", "ENCLAVE_BATCHER_ARGS="+strings.Join(args, " "),
		opRoot,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return EnclaveMeasurements{}, fmt.Errorf("failed to build intermediate docker image: %w", err)
	}

	// Build EIF image based on the docker image we just built
	enclaverCli := new(EnclaverCli)
	manifest := DefaultManifest("op-batcher", tag, intermediateTag, cpuCount, memoryMb)
	measurements, err := enclaverCli.BuildEnclave(ctx, manifest)
	return measurements, err
}

// BuildEifFromImage builds an EIF image by wrapping a pre-built app Docker image with Enclaver.
func BuildEifFromImage(ctx context.Context, appImage string, eifTag string, cpuCount uint, memoryMb uint) (EnclaveMeasurements, error) {
	manifest := DefaultManifest("op-batcher", eifTag, appImage, cpuCount, memoryMb)
	return new(EnclaverCli).BuildEnclave(ctx, manifest)
}

// TeeType enum values from IEspressoTEEVerifier (NITRO=0).
const teeTypeNitro uint8 = 0

const teeVerifierABI = `[
  {"inputs":[{"name":"enclaveHash","type":"bytes32"},{"name":"valid","type":"bool"},{"name":"teeType","type":"uint8"}],"name":"setEnclaveHash","outputs":[],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"name":"enclaveHash","type":"bytes32"},{"name":"teeType","type":"uint8"}],"name":"registeredEnclaveHashes","outputs":[{"name":"","type":"bool"}],"stateMutability":"view","type":"function"}
]`

// getVerifier retrieves the EspressoTEEVerifier instance and L1 client by traversing the contract chain.
func getVerifier(ctx context.Context, authenticatorAddress common.Address, L1Url string) (*bind.BoundContract, *ethclient.Client, error) {
	l1Client, err := ethclient.DialContext(ctx, L1Url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to L1 client: %w", err)
	}

	authenticator, err := batchauthenticator.NewBatchAuthenticator(authenticatorAddress, l1Client)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create batch authenticator: %w", err)
	}

	verifierAddress, err := authenticator.EspressoTEEVerifier(&bind.CallOpts{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get verifier address: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(teeVerifierABI))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse verifier ABI: %w", err)
	}

	return bind.NewBoundContract(verifierAddress, parsedABI, l1Client, l1Client, l1Client), l1Client, nil
}

// RegisterEnclaveHash registers the enclave PCR0 hash with the EspressoTEEVerifier.
func RegisterEnclaveHash(ctx context.Context, authenticatorAddress common.Address, L1Url string, key *ecdsa.PrivateKey, pcr0Bytes []byte) error {
	verifier, l1Client, err := getVerifier(ctx, authenticatorAddress, L1Url)
	if err != nil {
		return err
	}

	chainID, err := l1Client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	opts, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}
	// Call EspressoTEEVerifier.setEnclaveHash, not EspressoNitroTEEVerifier.setEnclaveHash
	// directly. The nitro verifier's setEnclaveHash is onlyTEEVerifier, so only the TEE verifier
	// contract can call it, not an EOA operator key.
	registrationTx, err := verifier.Transact(opts, "setEnclaveHash", crypto.Keccak256Hash(pcr0Bytes), true, teeTypeNitro)
	if err != nil {
		return fmt.Errorf("failed to create registration transaction: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	receipt, err := bind.WaitMined(waitCtx, l1Client, registrationTx)
	if err != nil {
		return fmt.Errorf("failed to wait for registration transaction: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("registration transaction failed")
	}

	return nil
}

// IsEnclaveHashRegistered checks if the given PCR0 hash is already registered with the EspressoTEEVerifier
func IsEnclaveHashRegistered(ctx context.Context, authenticatorAddress common.Address, L1Url string, pcr0Bytes []byte) (bool, error) {
	verifier, _, err := getVerifier(ctx, authenticatorAddress, L1Url)
	if err != nil {
		return false, fmt.Errorf("failed to get verifier: %w", err)
	}

	var result []any
	err = verifier.Call(&bind.CallOpts{Context: ctx}, &result, "registeredEnclaveHashes", crypto.Keccak256Hash(pcr0Bytes), teeTypeNitro)
	if err != nil {
		return false, fmt.Errorf("failed to call registeredEnclaveHashes function: %w", err)
	}
	if len(result) != 1 {
		return false, fmt.Errorf("registeredEnclaveHashes returned %d values", len(result))
	}

	return *abi.ConvertType(result[0], new(bool)).(*bool), nil
}
