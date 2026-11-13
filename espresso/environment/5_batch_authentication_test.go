package environment_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum/go-ethereum/crypto"
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

// TestE2eDevnetWithUnattestedBatcherKey verifies that when a batcher key is
// not properly attested, the L2 chain can still produce unsafe blocks but
// cannot progress to safe L2 blocks. The original body used the batcher key
// for 0xe16d5c4080C0faD6D2Ef4eb07C657674a217271C, which the
// MockEspressoNitroTEEVerifier contract special-cases as unattested.
//
// TODO: restore TestE2eDevnetWithUnattestedBatcherKey (original body in git
// history at 3dd70bc97d^) once the mock dev node — or a dedicated test
// service — can emulate the SP1 zk attestation verifier.
func TestE2eDevnetWithUnattestedBatcherKey(t *testing.T) {
	t.Skip("attestation tests not ported: the in-memory mock Espresso dev node does not emulate the SP1 zk attestation verifier")
}
