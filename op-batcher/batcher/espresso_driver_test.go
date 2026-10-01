package batcher

import (
	"context"
	"sync"
	"testing"

	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/stretchr/testify/require"
)

// Stopping drops the whole session (streamer, submitter and verifier address),
// so no per-start Espresso state survives into the next start. A zero-value
// Streamer is Stop()-safe, which makes this directly testable.
func TestStopBatchSubmittingDropsEspressoSession(t *testing.T) {
	l := newAuthSubmitter(t)
	l.shutdownCtx, l.cancelShutdownCtx = context.WithCancel(context.Background())
	l.killCtx, l.cancelKillCtx = context.WithCancel(context.Background())
	l.wg = &sync.WaitGroup{}
	l.running = true
	l.espresso = &espressoSession{streamer: &espressoStreamers.Streamer{}}

	require.NoError(t, l.StopBatchSubmitting(context.Background()))
	require.Nil(t, l.espresso)
}
