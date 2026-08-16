package batcher

// Espresso driver tests cover starting and stopping a batcher session and
// creating its streamer, including setup failures and RPC deadlines. Batch
// draining is covered in espresso_loading_test.go; range selection, channel
// reconciliation, and resets are covered in espresso_block_loader_test.go.

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	espressoLightClient "github.com/EspressoSystems/espresso-network/sdks/go/light-client"
	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

// stopRecordingStreamer records calls while delegating to the wrapped streamer.
type stopRecordingStreamer struct {
	espressoStreamer
	stopCalls int
}

func (s *stopRecordingStreamer) Stop() {
	s.stopCalls++
	if s.espressoStreamer != nil {
		s.espressoStreamer.Stop()
	}
}

func (s *stopRecordingStreamer) SetBatchPosition(position eth.L2BlockRef) {
	if s.espressoStreamer != nil {
		s.espressoStreamer.SetBatchPosition(position)
	}
}

func newEspressoLifecycleTestBatcher(t *testing.T) *BatchSubmitter {
	t.Helper()
	batcher, endpoints := setup(t, nil)
	rollupConfig := *batcher.RollupConfig
	rollupConfig.BatchAuthenticatorAddress = testAuthAddr
	batcher.RollupConfig = &rollupConfig
	batcher.Config.NetworkTimeout = time.Second
	batcher.Config.PollInterval = time.Hour
	batcher.Config.Espresso = EspressoBatcherConfig{
		Enabled:                    true,
		PollInterval:               time.Hour,
		VerifyReceiptMaxBlocks:     10,
		VerifyReceiptSafetyTimeout: time.Minute,
		VerifyReceiptRetryDelay:    time.Hour,
	}
	batcher.L1Client = &fakeL1Client{}
	batcher.Txmgr = testutils.NewFakeTxMgr(batcher.Log, common.Address{}, eth.ChainIDFromUInt64(1))
	batcher.Espresso.Client = &contextErrorEspressoClient{}

	backend := newMockAuthBackend(t)
	backend.teeVerifier = common.Address{1}
	batcher.batchAuth = newTestBatchAuthenticator(t, backend)
	lightClient, err := espressoLightClient.NewLightclientCaller(common.Address{2}, backend)
	require.NoError(t, err)
	batcher.Espresso.LightClient = lightClient

	block := types.NewBlockWithHeader(&types.Header{Number: big.NewInt(1)})
	endpoints.ethClient.Mock.On("BlockByNumber", big.NewInt(1)).Return(block, nil)
	endpoints.rollupClient.Mock.On("SyncStatus").Return(&eth.SyncStatus{
		HeadL1: eth.L1BlockRef{Number: genesisL1Origin},
		LocalSafeL2: eth.L2BlockRef{
			Hash:     block.Hash(),
			Number:   1,
			L1Origin: batcher.RollupConfig.Genesis.L1,
		},
	}, nil)
	return batcher
}

func TestBatchSubmitterEspressoStopStart(t *testing.T) {
	batcher := newEspressoLifecycleTestBatcher(t)

	var previous espressoStreamer
	for range 2 {
		require.NoError(t, batcher.StartBatchSubmitting())
		require.NotNil(t, batcher.espressoSession)
		require.Equal(t, batcher.Config.NetworkTimeout, batcher.espressoSession.submitter.networkTimeout)
		require.Equal(t, batcher.Log, batcher.espressoSession.submitter.logger)
		require.Equal(t, batcher.Metr, batcher.espressoSession.submitter.metr)
		current := batcher.espressoSession.streamer
		require.NotNil(t, current)
		if previous != nil {
			require.NotSame(t, previous, current, "restart must construct a fresh streamer")
		}
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = batcher.StopBatchSubmittingIfRunning(ctx)
			current.Stop()
		})
		recorder := &stopRecordingStreamer{espressoStreamer: current}
		batcher.espressoSession.streamer = recorder
		require.True(t, batcher.running)

		require.NoError(t, batcher.StopBatchSubmitting(t.Context()))
		require.Equal(t, 1, recorder.stopCalls, "driver stop must call streamer Stop")
		require.Nil(t, batcher.espressoSession, "driver stop must drop the session")
		require.False(t, batcher.running)
		previous = current
	}
}

// Stopping drops the whole session (streamer, submitter and verifier address),
// so no per-start Espresso state survives into the next start. A zero-value
// Streamer is Stop()-safe, which makes this directly testable.
func TestStopBatchSubmittingDropsEspressoSession(t *testing.T) {
	l := newAuthSubmitter(t)
	l.shutdownCtx, l.cancelShutdownCtx = context.WithCancel(context.Background())
	l.killCtx, l.cancelKillCtx = context.WithCancel(context.Background())
	l.wg = &sync.WaitGroup{}
	l.running = true
	l.espressoSession = &espressoSession{streamer: &espressoStreamers.Streamer{}}

	require.NoError(t, l.StopBatchSubmitting(context.Background()))
	require.Nil(t, l.espressoSession)
}

// stalledAnchorClient exposes the context that NewStreamer passes through the
// L2 adapter, then stalls until that context expires or its parent is canceled.
type stalledAnchorClient struct {
	dial.EthClientInterface
	calls chan context.Context
}

func (c *stalledAnchorClient) BlockByNumber(ctx context.Context, _ *big.Int) (*types.Block, error) {
	c.calls <- ctx
	<-ctx.Done()
	return nil, ctx.Err()
}

type anchorEndpointProvider struct {
	dial.L2EndpointProvider
	client dial.EthClientInterface
}

func (p *anchorEndpointProvider) EthClient(context.Context) (dial.EthClientInterface, error) {
	return p.client, nil
}

func TestSetupEspressoStreamerBoundsAnchorLookup(t *testing.T) {
	batcher := newEspressoLifecycleTestBatcher(t)
	batcher.Config.NetworkTimeout = 20 * time.Millisecond
	calls := make(chan context.Context, 1)
	batcher.EndpointProvider = &anchorEndpointProvider{
		L2EndpointProvider: batcher.EndpointProvider,
		client:             &stalledAnchorClient{calls: calls},
	}
	ctx, cancel := context.WithCancel(t.Context())
	results := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	started := time.Now()
	var streamer *espressoStreamers.Streamer
	go func() {
		defer wg.Done()
		var err error
		streamer, err = batcher.setupEspressoStreamer(ctx)
		results <- err
	}()
	t.Cleanup(func() { cancel(); wg.Wait() })
	callCtx := receiveTestValue(t, calls)
	deadline, ok := callCtx.Deadline()
	require.True(t, ok, "the anchor lookup must have a deadline")
	require.False(t, deadline.Before(started.Add(batcher.Config.NetworkTimeout)))
	require.LessOrEqual(t, time.Until(deadline), batcher.Config.NetworkTimeout, "deadline must not exceed the configured timeout")
	require.ErrorIs(t, receiveTestValue(t, results), context.DeadlineExceeded)
	require.NoError(t, ctx.Err(), "the lookup timeout must not cancel its caller")
	require.Nil(t, streamer, "a failed lookup must not return a streamer")
	require.Nil(t, batcher.espressoSession, "a failed lookup must not publish a session")
}

// TestSetupEspressoStreamer_NoAuthenticator guards the nil binding case: the
// streamer filters batches by the BatchAuthenticator address, so without one it
// must fail before it waits for the local-safe head.
func TestSetupEspressoStreamer_NoAuthenticator(t *testing.T) {
	batcher := newEspressoLifecycleTestBatcher(t)
	batcher.batchAuth = nil

	streamer, err := batcher.setupEspressoStreamer(t.Context())
	require.ErrorContains(t, err, "no BatchAuthenticator configured")
	require.Nil(t, streamer)
}

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

		anchor, err := l.waitForLocalSafeHead(t.Context(), time.Second, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, testCaffeinationL2, anchor.Number)
	})

	t.Run("anchors past the caffeination point", func(t *testing.T) {
		l, ep := newAnchorTestSubmitter(t, testCaffeinationL2)
		ep.rollupClient.ExpectSyncStatus(anchorStatus(testCaffeinationL2+7), nil)

		anchor, err := l.waitForLocalSafeHead(t.Context(), time.Second, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, testCaffeinationL2+7, anchor.Number)
	})

	t.Run("gives up while the caffeination point stays underived", func(t *testing.T) {
		l, ep := newAnchorTestSubmitter(t, testCaffeinationL2)
		ep.rollupClient.Mock.On("SyncStatus").Return(anchorStatus(testCaffeinationL2-1), nil)

		_, err := l.waitForLocalSafeHead(t.Context(), 50*time.Millisecond, 5*time.Millisecond)
		require.ErrorContains(t, err, "caffeination point (500)",
			"the give-up error must name the height the operator is waiting on")
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("retries through a transient sync-status failure", func(t *testing.T) {
		l, ep := newAnchorTestSubmitter(t, testCaffeinationL2)
		ep.rollupClient.ExpectSyncStatus(nil, errors.New("transient RPC failure"))
		ep.rollupClient.ExpectSyncStatus(anchorStatus(testCaffeinationL2), nil)

		anchor, err := l.waitForLocalSafeHead(t.Context(), 5*time.Second, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, testCaffeinationL2, anchor.Number)
	})

	t.Run("retries past an empty local-safe head", func(t *testing.T) {
		l, ep := newAnchorTestSubmitter(t, testCaffeinationL2)
		ep.rollupClient.ExpectSyncStatus(anchorStatus(0), nil) // populated status, zero LocalSafeL2
		ep.rollupClient.ExpectSyncStatus(anchorStatus(testCaffeinationL2), nil)

		anchor, err := l.waitForLocalSafeHead(t.Context(), 5*time.Second, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, testCaffeinationL2, anchor.Number)
	})

	t.Run("zero caffeination height anchors on any populated local-safe head", func(t *testing.T) {
		// --espresso.origin-height-l2 unset: no floor beyond a populated
		// head. Steady-state restarts rely on this; handoffs must set the flag.
		l, ep := newAnchorTestSubmitter(t, 0)
		ep.rollupClient.ExpectSyncStatus(anchorStatus(1), nil)

		anchor, err := l.waitForLocalSafeHead(t.Context(), time.Second, time.Millisecond)
		require.NoError(t, err)
		require.Equal(t, uint64(1), anchor.Number)
	})
}

// TestRollbackFailedStart locks the startup rollback: a failed Espresso setup
// must undo everything StartBatchSubmitting set up — clear the running flag so
// a later start attempt is not rejected with "batcher is already running", and
// cancel both contexts so nothing keeps running behind the failed start. The
// session itself needs no unwinding: newEspressoSession publishes it only once
// fully built, so a failed start never has one.
func TestRollbackFailedStart(t *testing.T) {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.shutdownCtx, l.cancelShutdownCtx = context.WithCancel(context.Background())
	l.killCtx, l.cancelKillCtx = context.WithCancel(context.Background())
	l.running = true

	l.rollbackFailedStart()

	require.False(t, l.running, "a failed start must not wedge later start attempts")
	require.ErrorIs(t, l.shutdownCtx.Err(), context.Canceled)
	require.ErrorIs(t, l.killCtx.Err(), context.Canceled)
	require.Nil(t, l.espressoSession, "a failed start must not leave a session behind")
}
