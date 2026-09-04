package devnet_tests

import "testing"

// TestBatcherRestart stops the batcher container mid-run, checks that the
// verifier does not process transactions submitted during the outage, then
// brings the batcher back and checks the backlog is processed and the chain
// keeps progressing.
//
// TODO: port TestBatcherRestart from optimism-espresso-integration
// espresso/devnet-tests/batcher_restart_test.go once the compose devnet
// harness is in place.
func TestBatcherRestart(t *testing.T) {
	t.Skip(skipNotPorted)
}
