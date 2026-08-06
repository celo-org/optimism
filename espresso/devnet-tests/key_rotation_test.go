package devnet_tests

import "testing"

// TestChangeBatchAuthenticatorOwner exercises key rotation on the
// BatchAuthenticator contract: transferring ownership and verifying the new
// owner (and only the new owner) can administrate the contract while the
// devnet keeps progressing.
//
// TODO: port TestChangeBatchAuthenticatorOwner from
// optimism-espresso-integration espresso/devnet-tests/key_rotation_test.go
// once the compose devnet harness is in place.
func TestChangeBatchAuthenticatorOwner(t *testing.T) {
	t.Skip(skipNotPorted)
}
