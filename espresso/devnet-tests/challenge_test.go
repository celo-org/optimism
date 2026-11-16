package devnet_tests

import (
	"context"
	"testing"
	"time"

	gametypes "github.com/ethereum-optimism/optimism/op-challenger/game/types"
	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/require"
)

// TestChallengeGame verifies that the succinct proposer creates dispute games,
// that they can be queried from the DisputeGameFactory, and that the honest
// proposer wins them. The succinct proposer needs finalized L2 blocks before
// creating games.
func TestChallengeGame(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	d := NewDevnet(ctx, t, WithProofs())
	require.NoError(t, d.Up())
	defer func() {
		require.NoError(t, d.Down())
	}()

	require.NoError(t, d.WaitForBatcher(ctx))

	// Check that the devnet works and generate some L2 activity.
	require.NoError(t, d.RunSimpleL2Burn())

	t.Log("waiting for the succinct proposer to create a dispute game")
	var games []ChallengeGame
	require.Eventually(t, func() bool {
		var err error
		games, err = d.ListChallengeGames()
		if err != nil {
			t.Logf("error listing games (will retry): %v", err)
			return false
		}
		return len(games) > 0
	}, 5*time.Minute, 5*time.Second, "no dispute game created")

	game := games[0]
	t.Logf("game created: index=%d, address=%s, claims=%d", game.Index, game.Address, game.Claims)
	require.GreaterOrEqual(t, game.Claims, uint64(1), "the game should have the proposer's root claim")

	disputeGame, err := bindings.NewFaultDisputeGame(game.Address, d.L1)
	require.NoError(t, err)
	callOpts := &bind.CallOpts{Context: ctx}
	gameStatus := func() (gametypes.GameStatus, error) {
		raw, err := disputeGame.Status(callOpts)
		if err != nil {
			return 0, err
		}
		return gametypes.GameStatusFromUint8(raw)
	}

	status, err := gameStatus()
	require.NoError(t, err)
	t.Logf("dispute game initial status: %s", status)
	require.Contains(t, []gametypes.GameStatus{gametypes.GameStatusInProgress, gametypes.GameStatusDefenderWon}, status)

	// The game resolves once its challenge window (MAX_CHALLENGE_DURATION in
	// the compose file) has passed; the succinct challenger resolves it.
	t.Logf("waiting for dispute game %s to resolve", game.Address)
	require.Eventually(t, func() bool {
		s, err := gameStatus()
		if err != nil {
			t.Logf("error reading the game status (will retry): %v", err)
			return false
		}
		status = s
		return status != gametypes.GameStatusInProgress
	}, 15*time.Minute, 10*time.Second, "dispute game did not resolve")
	require.Equal(t, gametypes.GameStatusDefenderWon, status, "the honest proposer should win the dispute game")
}
