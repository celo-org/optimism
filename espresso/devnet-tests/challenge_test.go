package devnet_tests

import "testing"

// TestChallengeGame verifies that the succinct proposer creates dispute games
// and that games can be queried from the DisputeGameFactory contract. The
// succinct proposer needs finalized L2 blocks before creating games. In the
// reference repo this test self-skips in the TEE compose profile, which does
// not run the succinct proposer/challenger.
//
// TODO: port TestChallengeGame from optimism-espresso-integration
// espresso/devnet-tests/challenge_test.go once the compose devnet harness
// (including the succinct-proposer service) is in place.
func TestChallengeGame(t *testing.T) {
	t.Skip(skipNotPorted)
}
