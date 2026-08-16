package batcher

import (
	"context"
	"errors"
	"testing"
	"time"

	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

const testCaffeinationL2 = uint64(500)

func newAnchorTestSubmitter(t *testing.T, caffeinationL2 uint64) (*BatchSubmitter, *mockL2EndpointProvider) {
	ep := newEndpointProvider()
	l := &BatchSubmitter{
		DriverSetup: DriverSetup{
			Log: testlog.Logger(t, log.LevelDebug),
			Config: BatcherConfig{
				NetworkTimeout: time.Second,
				Espresso: EspressoBatcherConfig{
					CaffeinationHeightL2: caffeinationL2,
				},
			},
			EndpointProvider: ep,
		},
	}
	return l, ep
}

// anchorStatus returns a sync status whose HeadL1 is populated (getSyncStatus
// backs off internally on an empty HeadL1) and whose LocalSafeL2 sits at the
// given height.
func anchorStatus(localSafe uint64) *eth.SyncStatus {
	status := &eth.SyncStatus{
		HeadL1: eth.L1BlockRef{Number: 50, Hash: common.Hash{0xa1}},
	}
	if localSafe > 0 {
		status.LocalSafeL2 = eth.L2BlockRef{Number: localSafe, Hash: common.Hash{byte(localSafe)}}
	}
	return status
}

// TestWaitForLocalSafeHead locks the streamer's anchor gate: the Espresso
// batcher must refuse to run until the caffeination point is local-safe (the
// fallback batcher's pre-activation channels have derived), must retry through
// transient sync-status trouble, and must give up with a retryable error
// rather than anchoring below the caffeination point.
func TestWaitForLocalSafeHead(t *testing.T) {
	t.Run("anchors once the caffeination point is local-safe", func(t *testing.T) {
		l, ep := newAnchorTestSubmitter(t, testCaffeinationL2)
		ep.rollupClient.ExpectSyncStatus(anchorStatus(testCaffeinationL2), nil)

		anchor, err := l.waitForLocalSafeHeadWithTiming(t.Context(), time.Second, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, testCaffeinationL2, anchor.Number)
	})

	t.Run("anchors past the caffeination point", func(t *testing.T) {
		l, ep := newAnchorTestSubmitter(t, testCaffeinationL2)
		ep.rollupClient.ExpectSyncStatus(anchorStatus(testCaffeinationL2+7), nil)

		anchor, err := l.waitForLocalSafeHeadWithTiming(t.Context(), time.Second, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, testCaffeinationL2+7, anchor.Number)
	})

	t.Run("gives up while the caffeination point stays underived", func(t *testing.T) {
		l, ep := newAnchorTestSubmitter(t, testCaffeinationL2)
		ep.rollupClient.Mock.On("SyncStatus").Return(anchorStatus(testCaffeinationL2-1), nil)

		_, err := l.waitForLocalSafeHeadWithTiming(t.Context(), 50*time.Millisecond, 5*time.Millisecond)
		require.ErrorContains(t, err, "caffeination point (500)",
			"the give-up error must name the height the operator is waiting on")
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("retries through a transient sync-status failure", func(t *testing.T) {
		l, ep := newAnchorTestSubmitter(t, testCaffeinationL2)
		ep.rollupClient.ExpectSyncStatus(nil, errors.New("transient RPC failure"))
		ep.rollupClient.ExpectSyncStatus(anchorStatus(testCaffeinationL2), nil)

		anchor, err := l.waitForLocalSafeHeadWithTiming(t.Context(), 5*time.Second, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, testCaffeinationL2, anchor.Number)
	})

	t.Run("retries past an empty local-safe head", func(t *testing.T) {
		l, ep := newAnchorTestSubmitter(t, testCaffeinationL2)
		ep.rollupClient.ExpectSyncStatus(anchorStatus(0), nil) // populated status, zero LocalSafeL2
		ep.rollupClient.ExpectSyncStatus(anchorStatus(testCaffeinationL2), nil)

		anchor, err := l.waitForLocalSafeHeadWithTiming(t.Context(), 5*time.Second, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, testCaffeinationL2, anchor.Number)
	})

	t.Run("zero caffeination height anchors on any populated local-safe head", func(t *testing.T) {
		// --espresso.origin-height-l2 unset: no floor beyond a populated
		// head. Steady-state restarts rely on this; handoffs must set the flag.
		l, ep := newAnchorTestSubmitter(t, 0)
		ep.rollupClient.ExpectSyncStatus(anchorStatus(1), nil)

		anchor, err := l.waitForLocalSafeHeadWithTiming(t.Context(), time.Second, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, uint64(1), anchor.Number)
	})
}

// TestRollbackFailedStart locks the startup rollback: a failed Espresso setup
// must undo everything StartBatchSubmitting set up — clear the running flag so
// a later start attempt is not rejected with "batcher is already running", and
// cancel both contexts so nothing keeps running behind the failed start.
func TestRollbackFailedStart(t *testing.T) {
	tests := []struct {
		name     string
		streamer *espressoStreamers.Streamer
	}{
		{
			// Setup can fail before setupEspressoStreamer constructs one.
			name:     "without a streamer",
			streamer: nil,
		},
		{
			// A constructed-but-never-started streamer makes Stop a no-op.
			name:     "with a constructed streamer",
			streamer: &espressoStreamers.Streamer{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := &BatchSubmitter{}
			l.Log = testlog.Logger(t, log.LevelDebug)
			l.shutdownCtx, l.cancelShutdownCtx = context.WithCancel(context.Background())
			l.killCtx, l.cancelKillCtx = context.WithCancel(context.Background())
			l.running = true
			l.espressoStreamer = test.streamer

			l.rollbackFailedStart()

			require.False(t, l.running, "a failed start must not wedge later start attempts")
			require.ErrorIs(t, l.shutdownCtx.Err(), context.Canceled)
			require.ErrorIs(t, l.killCtx.Err(), context.Canceled)
		})
	}
}
