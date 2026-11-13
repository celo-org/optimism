package environment_test

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

// TestE2eDevnetWithInvalidAttestation verifies that the batcher correctly
// fails to register when provided with an invalid attestation, ensuring the
// BatchAuthenticator contract properly validates attestations.
//
// TODO: restore TestE2eDevnetWithInvalidAttestation (original body in git
// history at acfa2e13c0^).
func TestE2eDevnetWithInvalidAttestation(t *testing.T) {
	t.Skip("TestE2eDevnetWithInvalidAttestation is not ported yet")
}

// unattestedBatcherKeyHex is the private key for
// 0xe16d5c4080C0faD6D2Ef4eb07C657674a217271C, the signer that
// MockEspressoNitroTEEVerifier.isSignerValid special-cases to return false
// (see packages/contracts-bedrock/test/mocks/MockEspressoTEEVerifiers.sol).
const unattestedBatcherKeyHex = "841c29acb9520a7ea8a48e7686cd825b93e8a3ecd966b62cb396ff8a2cd7e80e"

// TestE2eDevnetWithUnattestedBatcherKey verifies that when a batcher key is
// not properly attested, the L2 chain can still produce unsafe blocks but
// cannot progress to safe L2 blocks.
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

	// ...but safe L2 is not. Only a timeout proves the safe head stalled; any
	// other error (e.g. the sequencer RPC going away) must fail the test.
	_, err = geth.WaitForBlockToBeSafe(big.NewInt(1), l2Seq, 2*time.Minute)
	switch {
	case err == nil:
		t.Fatalf("L2 block 1 became safe even though the batcher key is unattested")
	case !errors.Is(err, context.DeadlineExceeded):
		t.Fatalf("unexpected error while waiting for L2 block 1 to be safe: %v", err)
	}

	safeHead, err := l2Seq.HeaderByNumber(ctx, big.NewInt(rpc.SafeBlockNumber.Int64()))
	if err != nil {
		t.Fatalf("failed to fetch L2 safe head: %v", err)
	}
	if have, want := bigs.Uint64Strict(safeHead.Number), uint64(0); have != want {
		t.Fatalf("L2 safe head advanced with an unattested batcher key:\nhave:\n\t%d\nwant:\n\t%d\n", have, want)
	}
}
