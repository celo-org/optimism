package environment_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

// fakeAttestationVerifier stands in for the SP1 zk attestation-verifier
// service. It rejects every attestation with a 400, as the real verifier does
// for bytes that do not parse as a Nitro attestation document, and records
// what it received so the test can tell a verifier rejection apart from the
// batcher failing before it ever dialed the service.
type fakeAttestationVerifier struct {
	mu       sync.Mutex
	received [][]byte
}

func (f *fakeAttestationVerifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/generate_proof" {
		http.NotFound(w, r)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f.mu.Lock()
	f.received = append(f.received, body)
	f.mu.Unlock()
	http.Error(w, "invalid attestation document", http.StatusBadRequest)
}

func (f *fakeAttestationVerifier) requests() [][]byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([][]byte(nil), f.received...)
}

// TestE2eDevnetWithInvalidAttestation verifies that the batcher refuses to
// start when the attestation verifier rejects its attestation: registration
// with the BatchAuthenticator contract must fail rather than be skipped.
//
// The verifier is faked in-process. The e2e L1 deploys
// MockEspressoTEEVerifier, whose registerService accepts any input, so the
// verifier is the only place an invalid attestation can be rejected here.
func TestE2eDevnetWithInvalidAttestation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	verifier := new(fakeAttestationVerifier)
	verifierServer := httptest.NewServer(verifier)
	defer verifierServer.Close()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t,
		env.SetBatcherKey(*privateKey),
		env.WithBatcherStoppedInitially(),
		env.WithAttestationService(verifierServer.URL),
	)
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	// Signal the testnet to shut down
	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	invalidAttestation := []byte("this is an invalid attestation")

	batchDriver := system.BatchSubmitter.TestDriver()
	batchDriver.Espresso.Attestation = invalidAttestation
	err = batchDriver.StartBatchSubmitting()
	if err == nil {
		t.Fatalf("batcher should've failed to register with invalid attestation but got nil error")
	}

	// Several registration failures (proof decoding, building or sending the
	// registerSigner tx) share the outer message, so also pin the cause to the
	// verifier rejecting the attestation.
	for _, expectedMsg := range []string{
		"could not register with BatchAuthenticator contract",
		"failed to generate zk proof from nitro attestation",
		"received non-200 response: 400",
	} {
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Fatalf("error message does not contain expected message %q:\ngot: %q", expectedMsg, err.Error())
		}
	}

	requests := verifier.requests()
	if have, want := len(requests), 1; have != want {
		t.Fatalf("attestation verifier received %d proof requests, want %d", have, want)
	}
	if !bytes.Equal(requests[0], invalidAttestation) {
		t.Fatalf("attestation verifier received %q, want %q", requests[0], invalidAttestation)
	}
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
	case !isTimeout(err):
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

// isTimeout reports whether err is a timeout. When the deadline expires during
// an RPC call, the client reports its transport's timeout (e.g. "write pipe:
// i/o timeout") rather than context.DeadlineExceeded.
func isTimeout(err error) bool {
	var timeout interface{ Timeout() bool }
	return errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &timeout) && timeout.Timeout())
}
