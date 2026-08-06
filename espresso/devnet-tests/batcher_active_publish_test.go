package devnet_tests

import "testing"

// TestBatcherActivePublishOnly tests that only the on-chain active batcher
// publishes to L1: with both the Espresso batcher and the fallback batcher
// running, the BatchInbox must only receive transactions from the batcher
// that BatchAuthenticator.activeIsEspresso designates as active.
//
// TODO: port TestBatcherActivePublishOnly from optimism-espresso-integration
// espresso/devnet-tests/batcher_active_publish_test.go once the compose
// devnet harness is in place.
func TestBatcherActivePublishOnly(t *testing.T) {
	t.Skip(skipNotPorted)
}
