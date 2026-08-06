package devnet_tests

import "testing"

// TestForcedTransaction verifies that the forced-transaction (deposit
// inclusion) mechanism works on the compose devnet: a transaction submitted
// via the OptimismPortal must be enforced on L2 within the inclusion window.
// This is the devnet equivalent of the espresso/environment forced
// transaction tests.
//
// TODO: port TestForcedTransaction from optimism-espresso-integration
// espresso/devnet-tests/forced_transaction_test.go once the compose devnet
// harness is in place.
func TestForcedTransaction(t *testing.T) {
	t.Skip(skipNotPorted)
}
