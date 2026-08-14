package batcher

import (
	"context"
	"sync"
	"testing"

	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// startedSubmitter returns a BatchSubmitter in the state StartBatchSubmitting
// leaves it in right after taking ownership (running, fresh contexts, empty
// waitgroup) with a streamer attached. A zero-value Streamer is safe here:
// Stop() on it is a no-op (nil cancel), and these tests only exercise
// teardown, never the streamer itself.
func startedSubmitter(t *testing.T) *BatchSubmitter {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.running = true
	l.shutdownCtx, l.cancelShutdownCtx = context.WithCancel(context.Background())
	l.killCtx, l.cancelKillCtx = context.WithCancel(context.Background())
	l.wg = &sync.WaitGroup{}
	l.espressoStreamer = &espressoStreamers.Streamer{}
	return l
}

// TestStopBatchSubmittingDropsStreamer pins that a stop/start cycle cannot put
// the next start's clearState behind the streamer re-anchor gate: that gate
// retries with no deadline while StartBatchSubmitting holds l.mutex — the very
// mutex the stop that could cancel it would need — so the stopped run's
// streamer must not survive into the next start.
func TestStopBatchSubmittingDropsStreamer(t *testing.T) {
	l := startedSubmitter(t)
	require.NoError(t, l.StopBatchSubmitting(context.Background()))
	require.Nil(t, l.espressoStreamer)
	require.False(t, l.running)
}

// TestRollbackFailedStartDropsStreamer pins the same invariant for the failed
// start path: a start that constructed a streamer and then failed must not
// leave it behind for the next attempt's clearState to re-anchor against.
func TestRollbackFailedStartDropsStreamer(t *testing.T) {
	l := startedSubmitter(t)
	l.rollbackFailedStart()
	require.Nil(t, l.espressoStreamer)
	require.False(t, l.running)
}
