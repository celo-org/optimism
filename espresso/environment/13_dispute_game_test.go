package environment_test

import (
	"context"
	"testing"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-challenger/game/types"
	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/challenger"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/disputegame"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum/go-ethereum/common"
)

// TestOutputAlphabetGameWithEspresso_ChallengerWins verifies that fraud proof challenges work correctly
// with Espresso integration enabled. It ensures a challenger can successfully detect and win
// against a malicious proposer, validating that the dispute resolution process remains intact
// when using Espresso for transaction finalization.
//
// This test mirrors the logic from TestOutputAlphabetGame_ChallengerWins in the non-Espresso
// implementation (op-e2e/faultproofs/output_alphabet_test.go), but runs with the Espresso-mode
// batcher enabled.
//
// Test structure:
//   - Setup: Initialize Sequencer and Batcher in Espresso mode
//   - Action: Deploy fault dispute system and trigger challenger response
//   - Assert: Verify challenger successfully wins the dispute game
func TestOutputAlphabetGameWithEspresso_ChallengerWins(t *testing.T) {
	op_e2e.InitParallel(t)
	ctx := context.Background()

	// Start a Espresso Dev Node
	launcher := new(env.EspressoDevNodeLauncherDocker)

	// Start a Fault Dispute System with Espresso Dev Node
	sys, espressoDevNode, err := launcher.StartE2eDevnet(
		ctx,
		t,
		env.WithFaultDisputeSystem(),
		env.WithL1FinalizedDistance(0),
		env.WithSequencerUseFinalized(true),
	)

	// Signal the testnet to shut down
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	l1Client := sys.NodeClient("l1")

	// Close the system and stop the Espresso Dev Node
	defer sys.Close()
	defer func() {
		err = espressoDevNode.Stop()
		if err != nil {
			t.Fatalf("failed to stop espresso dev node: %v", err)
		}
	}()

	// All the following testing code is pasted from `TestOutputAlphabetGame_ChallengerWins` in `op-e2e/faultproofs/output_alphabet_test.go`
	disputeGameFactory := disputegame.NewFactoryHelper(t, ctx, sys)
	game := disputeGameFactory.StartOutputAlphabetGame(ctx, "sequencer", 3, common.Hash{0xff})
	correctTrace := game.CreateHonestActor(ctx, "sequencer")
	game.LogGameData(ctx)

	opts := challenger.WithPrivKey(sys.Cfg.Secrets.Alice)
	game.StartChallenger(ctx, "sequencer", "Challenger", opts)
	game.LogGameData(ctx)

	// Challenger should post an output root to counter claims down to the leaf level of the top game
	claim := game.RootClaim(ctx)
	for claim.IsOutputRoot(ctx) && !claim.IsOutputRootLeaf(ctx) {
		if claim.AgreesWithOutputRoot() {
			// If the latest claim agrees with the output root, expect the honest challenger to counter it
			claim = claim.WaitForCounterClaim(ctx)
			game.LogGameData(ctx)
			claim.RequireCorrectOutputRoot(ctx)
		} else {
			// Otherwise we should counter
			claim = claim.Attack(ctx, common.Hash{0xaa})
			game.LogGameData(ctx)
		}
	}

	// Wait for the challenger to post the first claim in the cannon trace
	claim = claim.WaitForCounterClaim(ctx)
	game.LogGameData(ctx)

	// Attack the root of the alphabet trace subgame
	claim = correctTrace.AttackClaim(ctx, claim)
	for !claim.IsMaxDepth(ctx) {
		if claim.AgreesWithOutputRoot() {
			// If the latest claim supports the output root, wait for the honest challenger to respond
			claim = claim.WaitForCounterClaim(ctx)
			game.LogGameData(ctx)
		} else {
			// Otherwise we need to counter the honest claim
			claim = correctTrace.AttackClaim(ctx, claim)
			game.LogGameData(ctx)
		}
	}
	// Challenger should be able to call step and counter the leaf claim.
	claim.WaitForCountered(ctx)
	game.LogGameData(ctx)

	sys.TimeTravelClock.AdvanceTime(game.MaxClockDuration(ctx))
	require.NoError(t, wait.ForNextBlock(ctx, l1Client))
	game.WaitForGameStatus(ctx, types.GameStatusChallengerWon)
	game.LogGameData(ctx)
}
