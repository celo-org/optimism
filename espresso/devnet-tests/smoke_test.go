package devnet_tests

import "testing"

// TestSmoke brings the compose devnet up, waits for the batcher to be
// healthy, and sends a simple L2 transaction to check that everything has
// started up ok. Runs in both the standard and the TEE (Nitro enclave)
// compose profiles.
//
// TODO: port TestSmoke from optimism-espresso-integration
// espresso/devnet-tests/smoke_test.go once the compose devnet harness is in
// place.
func TestSmoke(t *testing.T) {
	t.Skip(skipNotPorted)
}
