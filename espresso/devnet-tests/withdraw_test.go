package devnet_tests

import "testing"

// TestWithdrawal runs a full L2→L1 withdrawal against the compose devnet:
// initiate on L2, wait for a dispute game covering the withdrawal block,
// prove against its output root, and finalize on L1 after the proof window.
//
// TODO: port TestWithdrawal from optimism-espresso-integration
// espresso/devnet-tests/withdraw_test.go once the compose devnet harness
// (including the succinct-proposer service) is in place.
func TestWithdrawal(t *testing.T) {
	t.Skip(skipNotPorted)
}
