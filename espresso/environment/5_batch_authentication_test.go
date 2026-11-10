package environment_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum/go-ethereum/crypto"
)

// TestE2eDevnetWithInvalidAttestation verifies that the batcher correctly
// fails to register when provided with an invalid attestation, ensuring the
// BatchAuthenticator contract properly validates attestations.
//
// Its original implementation gated on the SP1 zk attestation-verifier Docker
// container, which the in-memory mock Espresso dev node does not emulate, so
// it was dropped when the suite moved off the dockerized espresso-dev-node
// (see commit acfa2e13c0 for the removed body and the
// attestation_verifier_service_helpers.go it depended on).
//
// TODO: restore TestE2eDevnetWithInvalidAttestation (original body in git
// history at acfa2e13c0^) once the attestation verifier service is restored.
func TestE2eDevnetWithInvalidAttestation(t *testing.T) {
	t.Skip("attestation test not ported: the in-memory mock Espresso dev node does not emulate the SP1 zk attestation verifier")
}

// unattestedBatcherKeyHex is the private key for
// 0xe16d5c4080C0faD6D2Ef4eb07C657674a217271C, the signer that
// MockEspressoNitroTEEVerifier.isSignerValid special-cases to return false
// (see packages/contracts-bedrock/test/mocks/MockEspressoTEEVerifiers.sol).
const unattestedBatcherKeyHex = "841c29acb9520a7ea8a48e7686cd825b93e8a3ecd966b62cb396ff8a2cd7e80e"

// TestE2eDevnetWithUnattestedBatcherKey verifies that when a batcher key is
// not properly attested, the L2 chain can still produce unsafe blocks but
// cannot progress to safe L2 blocks.
//
// This test does not need the SP1 attestation verifier service: outside a
// Nitro enclave the batcher has no attestation, so registerBatcher returns
// before ever contacting the service. The rejection comes from the mock TEE
// verifier deployed by DeployEspresso (no NITRO_ENCLAVE_VERIFIER_ADDRESS),
// which refuses every batch signature produced by the unattested key, so the
// BatchAuthenticator never authenticates a batch and safe L2 stays put.
func TestE2eDevnetWithUnattestedBatcherKey(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	privateKey, err := crypto.HexToECDSA(unattestedBatcherKeyHex)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t,
		env.SetBatcherKey(*privateKey),
	)
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	// Signal the testnet to shut down
	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	l2Seq := system.NodeClient(e2esys.RoleSeq)

	// Check that unsafe L2 is progressing...
	if _, err := geth.WaitForBlock(big.NewInt(15), l2Seq); err != nil {
		t.Fatalf("failed to wait for unsafe block 15: %v", err)
	}

	// ...but safe L2 is not.
	if _, err := geth.WaitForBlockToBeSafe(big.NewInt(1), l2Seq, 2*time.Minute); err == nil {
		t.Fatalf("L2 block 1 became safe even though the batcher key is unattested")
	}
}
