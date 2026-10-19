package batcher

import (
	"context"
	"math/big"
	"testing"
	"time"

	espressoLightClient "github.com/EspressoSystems/espresso-network/sdks/go/light-client"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

type contextErrorL1Client struct{ L1Client }

func (*contextErrorL1Client) HeaderByNumber(ctx context.Context, _ *big.Int) (*types.Header, error) {
	return nil, contextError(ctx)
}
func (*contextErrorL1Client) CodeAt(ctx context.Context, _ common.Address, _ *big.Int) ([]byte, error) {
	return nil, contextError(ctx)
}
func (*contextErrorL1Client) CallContract(ctx context.Context, _ ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	return nil, contextError(ctx)
}

type contextErrorL2Client struct{ dial.EthClientInterface }

func (*contextErrorL2Client) BlockByNumber(ctx context.Context, _ *big.Int) (*types.Block, error) {
	return nil, contextError(ctx)
}

func TestBatcherRPCAdaptersBoundCalls(t *testing.T) {
	networkTimeoutCtx := func(ctx context.Context) (context.Context, context.CancelFunc) {
		return context.WithDeadline(ctx, time.Unix(1, 0))
	}
	l1 := &batcherL1Adapter{L1Client: &contextErrorL1Client{}, networkTimeoutCtx: networkTimeoutCtx}
	l2 := &batcherL2Adapter{EthClient: &contextErrorL2Client{}, networkTimeoutCtx: networkTimeoutCtx}
	methods := []struct {
		name string
		call func(context.Context) error
	}{
		{"L1.HeaderHashByNumber", func(ctx context.Context) error { _, err := l1.HeaderHashByNumber(ctx, nil); return err }},
		{"L1.HeaderByNumber", func(ctx context.Context) error { _, err := l1.HeaderByNumber(ctx, nil); return err }},
		{"L1.CodeAt", func(ctx context.Context) error { _, err := l1.CodeAt(ctx, common.Address{}, nil); return err }},
		{"L1.CallContract", func(ctx context.Context) error { _, err := l1.CallContract(ctx, ethereum.CallMsg{}, nil); return err }},
		{"L2.HeaderHashByNumber", func(ctx context.Context) error { _, err := l2.HeaderHashByNumber(ctx, nil); return err }},
	}
	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			err := method.call(t.Context())
			require.ErrorIs(t, err, context.DeadlineExceeded)
		})
	}
}

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
