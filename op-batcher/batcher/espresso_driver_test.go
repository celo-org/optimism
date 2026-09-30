package batcher

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	espressoLightClient "github.com/EspressoSystems/espresso-network/sdks/go/light-client"
	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
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
		require.NotNil(t, batcher.espresso)
		require.Equal(t, batcher.Config.NetworkTimeout, batcher.espresso.submitter.networkTimeout)
		current := batcher.espresso.streamer
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
		batcher.espresso.streamer = recorder
		require.True(t, batcher.running)

		require.NoError(t, batcher.StopBatchSubmitting(t.Context()))
		require.Equal(t, 1, recorder.stopCalls, "driver stop must call streamer Stop")
		require.Nil(t, batcher.espresso, "driver stop must drop the session")
		require.False(t, batcher.running)
		previous = current
	}
}

// failingCallL1Client fails every contract call, so resolving the TEE verifier
// address fails.
type failingCallL1Client struct {
	fakeL1Client
}

func (f *failingCallL1Client) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return nil, errors.New("contract call failed")
}

func newLifecycleSubmitter(t *testing.T) *BatchSubmitter {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.Config.NetworkTimeout = time.Second
	l.RollupConfig = &rollup.Config{
		BatchAuthenticatorAddress: common.HexToAddress("0x00000000000000000000000000000000000000aa"),
	}
	l.shutdownCtx, l.cancelShutdownCtx = context.WithCancel(context.Background())
	l.killCtx, l.cancelKillCtx = context.WithCancel(context.Background())
	l.wg = &sync.WaitGroup{}
	return l
}

// A start that fails partway must not publish a session: the next start would
// otherwise run clearState against this run's streamer and submitter.
func TestStartEspressoLoopsFailureLeavesNoSession(t *testing.T) {
	l := newLifecycleSubmitter(t)
	l.L1Client = &failingCallL1Client{}

	err := l.startEspressoLoops(&espressoStreamers.Streamer{}, make(chan txmgr.TxReceipt[txRef]), make(chan pubInfo, 1), make(chan int64, 1))
	require.ErrorContains(t, err, "could not resolve TEE verifier address")
	require.Nil(t, l.espresso)
}

// Stopping drops the whole session (streamer, submitter and verifier address),
// so no per-start Espresso state survives into the next start. A zero-value
// Streamer is Stop()-safe, which makes this directly testable.
func TestStopBatchSubmittingDropsEspressoSession(t *testing.T) {
	l := newLifecycleSubmitter(t)
	l.running = true
	l.espresso = &espressoSession{
		streamer:           &espressoStreamers.Streamer{},
		submitter:          &espressoTransactionSubmitter{},
		teeVerifierAddress: common.HexToAddress("0x00000000000000000000000000000000000000bb"),
	}

	require.NoError(t, l.StopBatchSubmitting(context.Background()))
	require.Nil(t, l.espresso)
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
	batcher.espresso = &espressoSession{streamer: streamer}
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
	require.Nil(t, batcher.espresso, "a failed lookup must not publish a session")
}
