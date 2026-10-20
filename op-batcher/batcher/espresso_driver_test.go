package batcher

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	espressoLightClient "github.com/EspressoSystems/espresso-network/sdks/go/light-client"
	"github.com/EspressoSystems/espresso-streamers/op/derivation"
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
	batcher.batchAuth = newTestReader(t, backend)
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
		require.Equal(t, batcher.Config.NetworkTimeout, batcher.espressoSubmitter.networkTimeout)
		current := batcher.espressoStreamer
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
		batcher.espressoStreamer = recorder
		require.True(t, batcher.running)

		require.NoError(t, batcher.StopBatchSubmitting(t.Context()))
		require.Equal(t, 1, recorder.stopCalls, "driver stop must call streamer Stop")
		require.Nil(t, batcher.espressoStreamer, "driver stop must clear the streamer reference")
		require.False(t, batcher.running)
		previous = current
	}
}

func TestBatchSubmitterFailedStartStopsAndClearsEspressoStreamer(t *testing.T) {
	batcher := &BatchSubmitter{DriverSetup: DriverSetup{Log: testlog.Logger(t, log.LevelDebug)}}
	streamer := &stopRecordingStreamer{}
	batcher.espressoStreamer = streamer
	batcher.shutdownCtx, batcher.cancelShutdownCtx = context.WithCancel(t.Context())
	batcher.killCtx, batcher.cancelKillCtx = context.WithCancel(t.Context())
	batcher.running = true

	batcher.mutex.Lock()
	batcher.rollbackFailedStart()
	batcher.mutex.Unlock()

	require.Equal(t, 1, streamer.stopCalls)
	require.Nil(t, batcher.espressoStreamer)
	require.False(t, batcher.running)
}

// stalledPeekStreamer waits for the caller to cancel or time out each Peek.
type stalledPeekStreamer struct {
	espressoStreamer
	calls chan context.Context
}

func (s *stalledPeekStreamer) Peek(ctx context.Context) *derivation.EspressoBatch {
	s.calls <- ctx
	<-ctx.Done()
	return nil
}
func (*stalledPeekStreamer) SetBatchPosition(eth.L2BlockRef) {}

func TestEspressoLoadingLoopBoundsPeekAndRetries(t *testing.T) {
	batcher, endpoints := setup(t, nil)
	batcher.Config.NetworkTimeout = 20 * time.Millisecond
	batcher.Config.Espresso.PollInterval = time.Millisecond
	endpoints.rollupClient.Mock.On("SyncStatus").Return(&eth.SyncStatus{
		HeadL1:      eth.L1BlockRef{Number: genesisL1Origin},
		CurrentL1:   eth.L1BlockRef{Number: genesisL1Origin},
		LocalSafeL2: eth.L2BlockRef{Number: 1},
		SafeL2:      eth.L2BlockRef{Number: 1},
		UnsafeL2:    eth.L2BlockRef{Number: 1},
	}, nil)
	streamer := &stalledPeekStreamer{calls: make(chan context.Context, 4)}
	batcher.espressoStreamer = streamer
	ctx, cancel := context.WithCancel(t.Context())
	var wg sync.WaitGroup
	wg.Add(1)
	started := time.Now()
	go batcher.espressoBatchLoadingLoop(ctx, &wg, make(chan pubInfo, 1), make(chan int64, 1))
	t.Cleanup(func() { cancel(); wg.Wait() })
	var previous context.Context
	for range 2 {
		callCtx := receiveTestValue(t, streamer.calls)
		deadline, ok := callCtx.Deadline()
		require.True(t, ok)
		if previous == nil {
			require.False(t, deadline.Before(started.Add(batcher.Config.NetworkTimeout)), "Peek must receive the full configured budget")
		} else {
			previousDeadline, _ := previous.Deadline()
			require.True(t, deadline.After(previousDeadline), "each Peek needs a fresh deadline")
		}
		require.LessOrEqual(t, time.Until(deadline), batcher.Config.NetworkTimeout, "deadline must not exceed the configured timeout")
		if previous != nil {
			require.ErrorIs(t, previous.Err(), context.DeadlineExceeded)
		}
		require.NoError(t, ctx.Err())
		previous = callCtx
	}
	cancel()
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
	go func() { defer wg.Done(); results <- batcher.setupEspressoStreamer(ctx) }()
	t.Cleanup(func() { cancel(); wg.Wait() })
	callCtx := receiveTestValue(t, calls)
	deadline, ok := callCtx.Deadline()
	require.True(t, ok, "the anchor lookup must have a deadline")
	require.False(t, deadline.Before(started.Add(batcher.Config.NetworkTimeout)))
	require.LessOrEqual(t, time.Until(deadline), batcher.Config.NetworkTimeout, "deadline must not exceed the configured timeout")
	require.ErrorIs(t, receiveTestValue(t, results), context.DeadlineExceeded)
	require.NoError(t, ctx.Err(), "the lookup timeout must not cancel its caller")
	require.Nil(t, batcher.espressoStreamer, "a failed lookup must not install a streamer")
}
