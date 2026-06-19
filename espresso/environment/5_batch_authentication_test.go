package environment_test

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum/go-ethereum/crypto"
)

// TestE2eDevnetWithInvalidAttestation verifies that the batcher correctly fails to register
// when provided with an invalid attestation. This test ensures that the BatchAuthenticator
// contract properly validates attestations.
func TestE2eDevnetWithInvalidAttestation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate private key")
	}

	system, _, err := launcher.StartE2eDevnet(ctx, t,
		env.SetBatcherKey(*privateKey),
		env.WithBatcherStoppedInitially(),
		env.WithEspressoAttestationVerifierService(),
	)

	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	batchDriver := system.BatchSubmitter.TestDriver()
	batchDriver.Espresso.Attestation = []byte("this is an invalid attestation")
	err = batchDriver.StartBatchSubmitting()

	if err == nil {
		t.Fatalf("batcher should've failed to register with invalid attestation but got nil error")
	}

	// Check for the key part of the error message
	expectedMsg := "could not register with BatchAuthenticator contract"
	errMsg := err.Error()
	if !strings.Contains(errMsg, expectedMsg) {
		t.Fatalf("error message does not contain expected message %q:\ngot: %q", expectedMsg, errMsg)
	}
}

// TestE2eDevnetWithUnattestedBatcherKey verifies that when a batcher key is not properly
// attested, the L2 chain can still produce unsafe blocks but cannot progress to safe L2 blocks.
func TestE2eDevnetWithUnattestedBatcherKey(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	// This is a random private key belonging to address 0xe16d5c4080C0faD6D2Ef4eb07C657674a217271C that will result in Mock Nitro verifier to return `false`
	// because the given key is not registered as an attested batcher.
	// Check the following code in Mock Espresso Nitro verifier:
	//    if (signer == address(0xe16d5c4080C0faD6D2Ef4eb07C657674a217271C)) {
	//        return false;
	//   }
	privateKey, err := crypto.HexToECDSA("841c29acb9520a7ea8a48e7686cd825b93e8a3ecd966b62cb396ff8a2cd7e80e")
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	system, _, err := launcher.StartE2eDevnet(ctx, t,
		env.SetBatcherKey(*privateKey),
		env.WithEspressoAttestationVerifierService(),
	)
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	l2Seq := system.NodeClient("sequencer")

	// Check that unsafe L2 is progressing...
	_, err = geth.WaitForBlock(big.NewInt(15), l2Seq)
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to wait for block:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	// ...but safe L2 is not
	_, err = geth.WaitForBlockToBeSafe(big.NewInt(1), l2Seq, 2*time.Minute)
	if err == nil {
		t.Fatalf("block shouldn't be safe")
	}

	_ = system
}
