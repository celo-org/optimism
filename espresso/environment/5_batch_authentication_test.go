package environment_test

import "testing"

// The two tests in this file verify batch authentication against a real
// attestation flow. Their original implementations gated on the SP1 zk
// attestation-verifier Docker container, which the in-memory mock Espresso
// dev node does not emulate, so they were dropped when the suite moved off
// the dockerized espresso-dev-node (see commit 3dd70bc97d for the removed
// bodies, and the espresso/devnet-tests package doc for the wider real-TEE
// tier that is also pending a port).

// TestE2eDevnetWithInvalidAttestation verifies that the batcher correctly
// fails to register when provided with an invalid attestation, ensuring the
// BatchAuthenticator contract properly validates attestations.
//
// TODO: restore TestE2eDevnetWithInvalidAttestation (original body in git
// history at 3dd70bc97d^) once the mock dev node — or a dedicated test
// service — can emulate the SP1 zk attestation verifier.
func TestE2eDevnetWithInvalidAttestation(t *testing.T) {
	t.Skip("attestation tests not ported: the in-memory mock Espresso dev node does not emulate the SP1 zk attestation verifier")
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
