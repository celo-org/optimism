package devnet_tests

import "testing"

// TestBatcherSwitching tests that the batcher can be switched from the
// Espresso batcher to a fallback batcher using the BatchAuthenticator
// contract. This is the devnet equivalent of the espresso/environment
// TestBatcherSwitching: op-batcher (Espresso, initially active) and
// op-batcher-fallback (initially stopped) run as separate containers.
//
// TODO: port TestBatcherSwitching from optimism-espresso-integration
// espresso/devnet-tests/batcher_switching_test.go once the compose devnet
// harness is in place.
func TestBatcherSwitching(t *testing.T) {
	t.Skip(skipNotPorted)
}
