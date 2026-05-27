package batcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	altda "github.com/ethereum-optimism/optimism/op-alt-da"
	"github.com/ethereum-optimism/optimism/op-batcher/compressor"
	"github.com/ethereum-optimism/optimism/op-batcher/config"
	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockL2EndpointProvider struct {
	ethClient       *testutils.MockL2Client
	ethClientErr    error
	rollupClient    *testutils.MockRollupClient
	rollupClientErr error
}

func newEndpointProvider() *mockL2EndpointProvider {
	return &mockL2EndpointProvider{
		ethClient:    new(testutils.MockL2Client),
		rollupClient: new(testutils.MockRollupClient),
	}
}

func (p *mockL2EndpointProvider) EthClient(context.Context) (dial.EthClientInterface, error) {
	return p.ethClient, p.ethClientErr
}

func (p *mockL2EndpointProvider) RollupClient(context.Context) (dial.RollupClientInterface, error) {
	return p.rollupClient, p.rollupClientErr
}

func (p *mockL2EndpointProvider) Close() {}

const genesisL1Origin = uint64(123)

func setup(t *testing.T, closeAppFn context.CancelCauseFunc) (*BatchSubmitter, *mockL2EndpointProvider) {
	ep := newEndpointProvider()

	cfg := defaultTestRollupConfig
	cfg.Genesis.L1.Number = genesisL1Origin

	if closeAppFn == nil {
		closeAppFn = func(cause error) {
			t.Fatalf("closeAppFn called, batcher hit a critical error: %v", cause)
		}
	}

	return NewBatchSubmitter(DriverSetup{
		closeApp:     closeAppFn,
		Log:          testlog.Logger(t, log.LevelDebug),
		Metr:         metrics.NoopMetrics,
		RollupConfig: cfg,
		Config: BatcherConfig{
			ThrottleParams: config.ThrottleParams{
				ControllerType: config.StepControllerType,
			},
		},
		ChannelConfig:    defaultTestChannelConfig(),
		EndpointProvider: ep,
	}), ep
}

func TestBatchSubmitter_SafeL1Origin(t *testing.T) {
	bs, ep := setup(t, nil)

	tests := []struct {
		name                   string
		currentSafeOrigin      uint64
		failsToFetchSyncStatus bool
		expectResult           uint64
		expectErr              bool
	}{
		{
			name:              "ExistingSafeL1Origin",
			currentSafeOrigin: 999,
			expectResult:      999,
		},
		{
			name:              "NoExistingSafeL1OriginUsesGenesis",
			currentSafeOrigin: 0,
			expectResult:      genesisL1Origin,
		},
		{
			name:                   "ErrorFetchingSyncStatus",
			failsToFetchSyncStatus: true,
			expectErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.failsToFetchSyncStatus {
				ep.rollupClient.ExpectSyncStatus(&eth.SyncStatus{}, errors.New("failed to fetch sync status"))
			} else {
				ep.rollupClient.ExpectSyncStatus(&eth.SyncStatus{
					LocalSafeL2: eth.L2BlockRef{
						L1Origin: eth.BlockID{
							Number: tt.currentSafeOrigin,
						},
					},
				}, nil)
			}

			id, err := bs.safeL1Origin(context.Background())

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectResult, id.Number)
			}
		})
	}
}

func TestBatchSubmitter_SafeL1Origin_FailsToResolveRollupClient(t *testing.T) {
	bs, ep := setup(t, nil)

	ep.rollupClientErr = errors.New("failed to resolve rollup client")

	_, err := bs.safeL1Origin(context.Background())
	log := testlog.Logger(t, log.LevelDebug)
	log.Debug("Err", err)
	require.Error(t, err)
}

type MockTxQueue struct {
	m sync.Map
}

func (q *MockTxQueue) Send(ref txRef, candidate txmgr.TxCandidate, receiptCh chan txmgr.TxReceipt[txRef]) {
	q.m.Store(ref.id.String(), candidate)
}

func (q *MockTxQueue) Load(id string) txmgr.TxCandidate {
	c, _ := q.m.Load(id)
	return c.(txmgr.TxCandidate)
}

func TestBatchSubmitter_sendTx_FloorDataGas(t *testing.T) {
	bs, _ := setup(t, nil)

	q := new(MockTxQueue)

	txData := txData{
		frames: []frameData{
			{
				data: []byte{0x01, 0x02, 0x03}, // 3 nonzero bytes = 12 tokens https://github.com/ethereum/EIPs/blob/master/EIPS/eip-7623.md
			},
		},
	}
	candidate := txmgr.TxCandidate{
		To:     &bs.RollupConfig.BatchInboxAddress,
		TxData: txData.CallData(),
	}

	bs.sendTx(txData,
		false,
		&candidate,
		q,
		make(chan txmgr.TxReceipt[txRef]))

	candidateOut := q.Load(txData.ID().String())

	expectedFloorDataGas := uint64(21_000 + 12*10)
	require.GreaterOrEqual(t, candidateOut.GasLimit, expectedFloorDataGas)
}

type handlerFailureMode string

const (
	noFailure      handlerFailureMode = "none"
	internalError  handlerFailureMode = "internal_error"
	methodNotFound handlerFailureMode = "method_not_found"
)

// createHTTPHandler creates a mock HTTP handler for testing, it accepts a callback which
// is invoked when the expected request is received.
func createHTTPHandler(t *testing.T, cb func(), failureMode handlerFailureMode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var req struct {
				JSONRPC string        `json:"jsonrpc"`
				Method  string        `json:"method"`
				Params  []interface{} `json:"params"`
				ID      interface{}   `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
				cb()

				switch failureMode {
				case noFailure:
					w.Header().Set("Content-Type", "application/json")
					_, err := w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":true}`))
					if err != nil {
						t.Logf("Error writing response: %v", err)
					}
					return
				case internalError:
					http.Error(w, "Simulated failure", http.StatusInternalServerError)
					return
				case methodNotFound:
					w.Header().Set("Content-Type", "application/json")
					_, err := w.Write([]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"error":{"code":%d,"message":"method not found"}}`, eth.MethodNotFound)))
					if err != nil {
						t.Logf("Error writing response: %v", err)
					}
					return
				}
			}
		}
		http.Error(w, "Unexpected request", http.StatusBadRequest)
	}
}

func TestBatchSubmitter_ThrottlingEndpoints(t *testing.T) {
	// Set a very long timeout to avoid flakiness
	timeout := time.Second * 120
	testThrottlingEndpoints := func(numHealthyServers, numUnhealthyServers int, throttlingEnabled bool) func(t *testing.T) {

		return func(t *testing.T) {
			healthyCalls := make([]int, numHealthyServers)
			unHealthyCalls := make([]int, numUnhealthyServers)

			healthyServers := make([]*httptest.Server, numHealthyServers)
			unhealthyServers := make([]*httptest.Server, numUnhealthyServers)

			urls := make([]string, 0, numHealthyServers+numUnhealthyServers)

			for i := range healthyCalls {
				healthyServers[i] = httptest.NewServer(createHTTPHandler(t, func() { healthyCalls[i]++ }, noFailure))
				urls = append(urls, healthyServers[i].URL)
				defer healthyServers[i].Close()
			}
			for i := range unHealthyCalls {
				unhealthyServers[i] = httptest.NewServer(createHTTPHandler(t, func() { unHealthyCalls[i]++ }, internalError))
				urls = append(urls, unhealthyServers[i].URL)
				defer unhealthyServers[i].Close()
			}

			// Setup test context
			ctx, cancel := context.WithCancel(context.Background())

			// Add in an endpoint with no server at all, representing an "always down" endpoint (only when throttling enabled)
			if throttlingEnabled {
				urls = append(urls, "http://invalid/")
			}

			t.Log("Throttling endpoints:", urls)
			t.Logf("Throttling enabled: %v", throttlingEnabled)

			var batcherShutdownError error

			// Create real metrics instead of NoopMetrics so we can verify metric recording
			metr := metrics.NewMetrics("test")

			ep := newEndpointProvider()
			cfg := defaultTestRollupConfig
			cfg.Genesis.L1.Number = genesisL1Origin

			// Set threshold values based on whether throttling is enabled
			lowerThreshold := uint64(0)
			upperThreshold := uint64(0)
			if throttlingEnabled {
				lowerThreshold = 10000
				upperThreshold = 20000
			}

			// Create test BatchSubmitter
			bs := NewBatchSubmitter(DriverSetup{
				closeApp:     func(cause error) { batcherShutdownError = cause },
				Log:          testlog.Logger(t, log.LevelDebug),
				Metr:         metr, // Use real metrics
				RollupConfig: cfg,
				Config: BatcherConfig{
					ThrottleParams: config.ThrottleParams{
						ControllerType:      config.StepControllerType,
						LowerThreshold:      lowerThreshold,
						UpperThreshold:      upperThreshold,
						TxSizeLowerLimit:    5000,
						TxSizeUpperLimit:    10000,
						BlockSizeLowerLimit: 20000,
						BlockSizeUpperLimit: 30000,
						Endpoints:           urls,
					},
					NetworkTimeout: time.Second,
				},
				ChannelConfig:    defaultTestChannelConfig(),
				EndpointProvider: ep,
			})

			bs.shutdownCtx = ctx

			// Test the throttling loop
			pendingBytesUpdated := make(chan int64, 1)
			wg1 := sync.WaitGroup{}

			// Start throttling loop in a goroutine only if throttling is enabled
			if throttlingEnabled {
				wg1.Add(1)
				go bs.throttlingLoop(&wg1, pendingBytesUpdated)
			}

			// Add a block to the channel manager so unsafeDABytes() returns > 0
			testBlock := newMiniL2Block(5) // Create a block with 5 transactions
			err := bs.channelMgr.AddL2Block(testBlock)
			require.NoError(t, err, "Should be able to add block to channel manager")

			// Simulate block loading by calling sendToThrottlingLoop periodically
			wg2 := sync.WaitGroup{}
			blockLoadingCtx, cancelBlockLoading := context.WithCancel(context.Background())
			go func() {
				defer wg2.Done()
				// Simulate block loading
				for range time.NewTicker(100 * time.Millisecond).C {
					select {
					case <-blockLoadingCtx.Done():
						return
					default:
						// Simulate block loading - use sendToThrottlingLoop which records metrics
						// and sends to the channel (this is what the real block loading loop does)
						bs.sendToThrottlingLoop(pendingBytesUpdated)
					}
				}

			}()
			wg2.Add(1)

			t.Cleanup(func() {
				cancelBlockLoading()
				wg2.Wait()
				close(pendingBytesUpdated)
				wg1.Wait()
				cancel()
			})

			// Verify metrics: unsafe_da_bytes metric should be recorded in all cases
			time.Sleep(200 * time.Millisecond) // Wait for metric updates
			c := opmetrics.NewMetricChecker(t, metr.Registry())
			prefix := "op_batcher_test_"
			unsafeDABytesFamily := c.FindByName(prefix + "unsafe_da_bytes")
			require.NotNil(t, unsafeDABytesFamily, "unsafe_da_bytes metric should exist")
			unsafeDABytesMetric := unsafeDABytesFamily.FindByLabels(map[string]string{})
			require.NotNil(t, unsafeDABytesMetric, "unsafe_da_bytes metric should be queryable")
			metricValue := unsafeDABytesMetric.Gauge.GetValue()
			require.Greater(t, metricValue, 0.0, "unsafe_da_bytes should be > 0 after adding blocks")
			t.Logf("unsafe_da_bytes metric value: %.0f", metricValue)

			if throttlingEnabled {
				// Only check endpoint calls when throttling is enabled
				require.Eventually(t,
					func() bool {
						// Check that all endpoints were called
						if slices.Contains(healthyCalls, 0) || slices.Contains(unHealthyCalls, 0) {
							return false
						}
						return true
					}, time.Second*10, time.Millisecond*10, "All endpoints should have been called within 10s")

				startTestServerAtAddr := func(addr string, handler http.HandlerFunc) *httptest.Server {
					ln, err := net.Listen("tcp", addr)
					require.NoError(t, err, "Failed to create new listener for test server")

					s := &httptest.Server{
						Listener: ln,
						Config:   &http.Server{Handler: handler},
					}
					s.Start()
					return s
				}

				// Take one of the healthy servers down, wait 2s and restart. Check it is called again.
				if len(healthyServers) > 0 {
					restartedServerCalled := false

					addr := healthyServers[0].Listener.Addr().String()
					healthyServers[0].Close()
					time.Sleep(time.Second * 2)
					startTestServerAtAddr(addr, createHTTPHandler(t, func() { restartedServerCalled = true }, noFailure))
					defer healthyServers[0].Close()
					t.Log("restarted server at", addr)

					require.Eventually(t, func() bool {
						return restartedServerCalled
					}, timeout, time.Millisecond*10, "Restarted server should have been called within 2s")
				}

				// Take an unhealthy server down, wait 2s and bring it back up with misconfiguration. Check the batcher exits.
				if len(unhealthyServers) > 0 {
					restartedServerCalled := false

					addr := unhealthyServers[0].Listener.Addr().String()
					unhealthyServers[0].Close()
					time.Sleep(time.Second * 2)
					startTestServerAtAddr(addr, createHTTPHandler(t, func() { restartedServerCalled = true }, methodNotFound))
					defer unhealthyServers[0].Close()
					t.Log("restarted server at", addr)

					require.Eventually(t, func() bool {
						return restartedServerCalled
					}, timeout, time.Millisecond*10, "Restarted server should have been called within 2s")

					require.Eventually(t, func() bool {
						return batcherShutdownError != nil
					}, timeout, time.Millisecond*10, "Batcher should have triggered self shutdown within 2s")

					require.Equal(t, batcherShutdownError.Error(), ErrSetMaxDASizeRPCMethodUnavailable("http://"+addr, errors.New("method not found")).Error(), "Batcher shutdown error should be the same as the expected error")
				}
			} else {
				// When throttling is disabled, verify endpoints were NOT called
				time.Sleep(time.Second * 2) // Wait to ensure no calls are made
				for i := range healthyCalls {
					require.Equal(t, 0, healthyCalls[i], "No endpoint calls should be made when throttling is disabled")
				}
				for i := range unHealthyCalls {
					require.Equal(t, 0, unHealthyCalls[i], "No endpoint calls should be made when throttling is disabled")
				}
				t.Log("Verified: no endpoint calls when throttling disabled")
			}
		}
	}
	t.Run("two normal endpoints", testThrottlingEndpoints(2, 0, true))
	t.Run("two failing endpoints", testThrottlingEndpoints(0, 2, true))
	t.Run("one normal endpoint, one failing endpoint", testThrottlingEndpoints(1, 1, true))
	t.Run("throttling disabled", testThrottlingEndpoints(1, 0, false))
}

func TestBatchSubmitter_CriticalError(t *testing.T) {
	criticalErrors := []error{
		eth.InputError{
			Code: eth.MethodNotFound,
		},
		eth.InputError{
			Code: eth.InvalidParams,
		},
	}

	for _, e := range criticalErrors {
		assert.True(t, isCriticalThrottlingRPCError(e), "false positive: %s", e)
	}

	nonCriticalErrors := []error{
		eth.InputError{
			Code: eth.UnsupportedFork,
		},
		errors.New("timeout"),
	}

	for _, e := range nonCriticalErrors {
		assert.False(t, isCriticalThrottlingRPCError(e), "false negative: %s", e)
	}
}

// ======= ALTDA TESTS =======

// fakeL1Client is just a dummy struct. All fault injection is done via the fakeTxMgr (which doesn't interact with this fakeL1Client).
type fakeL1Client struct {
	// Embed bind.ContractBackend so the type satisfies the L1Client interface
	// (which requires it for the BatchAuthenticator binding used by the
	// fallback batcher). AltDA tests never exercise these methods.
	bind.ContractBackend
}

func (f *fakeL1Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if number == nil {
		number = big.NewInt(0)
	}
	return &types.Header{
		Number:     number,
		ParentHash: common.Hash{},
		Time:       0,
	}, nil
}
func (f *fakeL1Client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return 0, nil
}

func altDASetup(_ *testing.T, log log.Logger) (*BatchSubmitter, *mockL2EndpointProvider, *altda.MockDAClient, *testutils.FakeTxMgr) {
	ep := newEndpointProvider()

	rollupCfg := &rollup.Config{
		Genesis:   rollup.Genesis{L2: eth.BlockID{Number: 0}, L1: eth.BlockID{Number: genesisL1Origin}},
		L2ChainID: big.NewInt(1234),
	}
	batcherCfg := BatcherConfig{
		PollInterval: 10 * time.Millisecond,
		UseAltDA:     true,
	}

	fakeTxMgr := testutils.NewFakeTxMgr(log.With("subsystem", "fake-txmgr"), common.Address{0}, eth.ChainIDFromUInt64(0))
	l1Client := &fakeL1Client{}

	channelCfg := ChannelConfig{
		// SeqWindowSize:      15,
		// SubSafetyMargin:    4,
		ChannelTimeout:  10,
		MaxFrameSize:    400, // so that each channel has exactly 1 frame (output is ~291 bytes)
		TargetNumFrames: 1,
		BatchType:       derive.SingularBatchType,
		CompressorConfig: compressor.Config{
			Kind: compressor.NoneKind,
		},
		DaType: DaTypeAltDA,
	}
	mockAltDAClient := altda.NewCountingGenericCommitmentMockDAClient(log.With("subsystem", "da-client"))
	return NewBatchSubmitter(DriverSetup{
		Log:              log,
		Metr:             metrics.NoopMetrics,
		RollupConfig:     rollupCfg,
		ChannelConfig:    channelCfg,
		Config:           batcherCfg,
		EndpointProvider: ep,
		Txmgr:            fakeTxMgr,
		L1Client:         l1Client,
		AltDA:            mockAltDAClient,
	}), ep, mockAltDAClient, fakeTxMgr
}

func fakeSyncStatus(unsafeL2BlockNum uint64, L1BlockRef eth.L1BlockRef) *eth.SyncStatus {
	return &eth.SyncStatus{
		UnsafeL2: eth.L2BlockRef{
			Hash:   common.HexToHash("0x1234"),
			Number: unsafeL2BlockNum,
			L1Origin: eth.BlockID{
				Number: 0,
			},
		},
		SafeL2: eth.L2BlockRef{
			Hash:   common.HexToHash("0x5678"),
			Number: 0,
			L1Origin: eth.BlockID{
				Number: 0,
			},
		},
		LocalSafeL2: eth.L2BlockRef{
			Hash:   common.HexToHash("0x5678"),
			Number: 0,
			L1Origin: eth.BlockID{
				Number: 0,
			},
		},
		CurrentL1: L1BlockRef,
		HeadL1:    L1BlockRef,
	}
}

// There are 4 failure cases (unhappy paths) that the op-batcher has to deal with.
// They are outlined in https://github.com/ethereum-optimism/optimism/tree/develop/op-batcher#happy-path
// This test suite covers these 4 cases in the context of AltDA.
func TestBatchSubmitter_AltDA_FailureCase1_L2Reorg(t *testing.T) {
	t.Parallel()
	log := testlog.Logger(t, log.LevelDebug)
	bs, ep, mockAltDAClient, fakeTxMgr := altDASetup(t, log)

	L1Block0 := types.NewBlock(&types.Header{
		Number: big.NewInt(0),
	}, nil, nil, nil, types.DefaultBlockConfig)
	L1Block0Ref := eth.L1BlockRef{
		Hash:   L1Block0.Hash(),
		Number: L1Block0.NumberU64(),
	}
	// We return incremental syncStatuses to force the op-batcher to entirely process each L2 block one by one.
	// To test multi channel behavior, we could return a sync status that is multiple blocks ahead of the current L2 block.
	ep.rollupClient.Mock.On("SyncStatus").Times(10).Return(fakeSyncStatus(1, L1Block0Ref), nil)
	ep.rollupClient.Mock.On("SyncStatus").Times(10).Return(fakeSyncStatus(2, L1Block0Ref), nil)
	ep.rollupClient.Mock.On("SyncStatus").Times(10).Return(fakeSyncStatus(3, L1Block0Ref), nil)
	ep.rollupClient.Mock.On("SyncStatus").Times(10).Return(fakeSyncStatus(1, L1Block0Ref), nil)
	ep.rollupClient.Mock.On("SyncStatus").Times(10).Return(fakeSyncStatus(2, L1Block0Ref), nil)
	ep.rollupClient.Mock.On("SyncStatus").Return(fakeSyncStatus(3, L1Block0Ref), nil)

	L2Block0 := newMiniL2BlockWithNumberParent(1, big.NewInt(0), common.HexToHash("0x0"))
	L2Block1 := newMiniL2BlockWithNumberParent(1, big.NewInt(1), L2Block0.Hash())
	L2Block2 := newMiniL2BlockWithNumberParent(1, big.NewInt(2), L2Block1.Hash())
	L2Block2Prime := newMiniL2BlockWithNumberParentAndL1Information(1, big.NewInt(2), L2Block1.Hash(), 101, 0)
	L2Block3Prime := newMiniL2BlockWithNumberParent(1, big.NewInt(3), L2Block2Prime.Hash())

	// L2block0 is the genesis block which is considered safe, so never loaded into the state.
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(1)).Twice().Return(L2Block1, nil)
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(2)).Once().Return(L2Block2, nil)
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(2)).Once().Return(L2Block2Prime, nil)
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(3)).Twice().Return(L2Block3Prime, nil)

	err := bs.StartBatchSubmitting()
	require.NoError(t, err)
	time.Sleep(1 * time.Second) // 1 second is enough to process all blocks at 10ms poll interval
	err = bs.StopBatchSubmitting(context.Background())
	require.NoError(t, err)

	// After the reorg, block 1 needs to be reprocessed, hence why we see 5 store calls: 1, 2, 1, 2', 3'
	require.Equal(t, 5, mockAltDAClient.StoreCount)
	require.Equal(t, uint64(5), fakeTxMgr.Nonce)

}

func TestBatchSubmitter_AltDA_FailureCase2_FailedL1Tx(t *testing.T) {
	t.Parallel()
	log := testlog.Logger(t, log.LevelDebug)
	bs, ep, mockAltDAClient, fakeTxMgr := altDASetup(t, log)

	L1Block0 := types.NewBlock(&types.Header{
		Number: big.NewInt(0),
	}, nil, nil, nil, types.DefaultBlockConfig)
	L1Block0Ref := eth.L1BlockRef{
		Hash:   L1Block0.Hash(),
		Number: L1Block0.NumberU64(),
	}
	// We return incremental syncStatuses to force the op-batcher to entirely process each L2 block one by one.
	// To test multi channel behavior, we could return a sync status that is multiple blocks ahead of the current L2 block.
	ep.rollupClient.Mock.On("SyncStatus").Times(10).Return(fakeSyncStatus(1, L1Block0Ref), nil)
	ep.rollupClient.Mock.On("SyncStatus").Times(10).Return(fakeSyncStatus(2, L1Block0Ref), nil)
	ep.rollupClient.Mock.On("SyncStatus").Times(10).Return(fakeSyncStatus(3, L1Block0Ref), nil)
	ep.rollupClient.Mock.On("SyncStatus").Return(fakeSyncStatus(4, L1Block0Ref), nil)

	L2Block0 := newMiniL2BlockWithNumberParent(1, big.NewInt(0), common.HexToHash("0x0"))
	L2Block1 := newMiniL2BlockWithNumberParent(1, big.NewInt(1), L2Block0.Hash())
	L2Block2 := newMiniL2BlockWithNumberParent(1, big.NewInt(2), L2Block1.Hash())
	L2Block3 := newMiniL2BlockWithNumberParent(1, big.NewInt(3), L2Block2.Hash())
	L2Block4 := newMiniL2BlockWithNumberParent(1, big.NewInt(4), L2Block3.Hash())

	// L2block0 is the genesis block which is considered safe, so never loaded into the state.
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(1)).Once().Return(L2Block1, nil)
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(2)).Once().Return(L2Block2, nil)
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(3)).Once().Return(L2Block3, nil)
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(4)).Once().Return(L2Block4, nil)

	fakeTxMgr.ErrorEveryNthSend(2)
	err := bs.StartBatchSubmitting()
	require.NoError(t, err)
	time.Sleep(1 * time.Second) // 1 second is enough to process all blocks at 10ms poll interval
	err = bs.StopBatchSubmitting(context.Background())
	require.NoError(t, err)

	require.Equal(t, 4, mockAltDAClient.StoreCount)
	// TODO: we should prob also check that the commitments are in order?
	require.Equal(t, uint64(4), fakeTxMgr.Nonce)
}

func TestBatchSubmitter_AltDA_FailureCase3_ChannelTimeout(t *testing.T) {
	// This function is not implemented because the batcher channel logic makes it very difficult to inject faults.
	// A version of this test was implemented here: https://github.com/Layr-Labs/optimism/blob/4b79c981a13bf096ae2984634d976956fbbfddff/op-batcher/batcher/driver_test.go#L300
	// However we opted to not merge it into the main branch because it has an external dependency on the https://github.com/pingcap/failpoint package,
	// and requires a lot of custom test setup and failpoint code injection into the batcher's codebase.
	// See https://github.com/ethereum-optimism/optimism/commit/4b79c981a13bf096ae2984634d976956fbbfddff for the full implementation.
}

func TestBatchSubmitter_AltDA_FailureCase4_FailedBlobSubmission(t *testing.T) {
	t.Parallel()
	log := testlog.Logger(t, log.LevelDebug)
	bs, ep, mockAltDAClient, fakeTxMgr := altDASetup(t, log)

	L1Block0 := types.NewBlock(&types.Header{
		Number: big.NewInt(0),
	}, nil, nil, nil, types.DefaultBlockConfig)
	L1Block0Ref := eth.L1BlockRef{
		Hash:   L1Block0.Hash(),
		Number: L1Block0.NumberU64(),
	}
	ep.rollupClient.Mock.On("SyncStatus").Return(fakeSyncStatus(4, L1Block0Ref), nil)

	L2Block0 := newMiniL2BlockWithNumberParent(1, big.NewInt(0), common.HexToHash("0x0"))
	L2Block1 := newMiniL2BlockWithNumberParent(1, big.NewInt(1), L2Block0.Hash())
	L2Block2 := newMiniL2BlockWithNumberParent(1, big.NewInt(2), L2Block1.Hash())
	L2Block3 := newMiniL2BlockWithNumberParent(1, big.NewInt(3), L2Block2.Hash())
	L2Block4 := newMiniL2BlockWithNumberParent(1, big.NewInt(4), L2Block3.Hash())

	// L2block0 is the genesis block which is considered safe, so never loaded into the state.
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(1)).Once().Return(L2Block1, nil)
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(2)).Once().Return(L2Block2, nil)
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(3)).Once().Return(L2Block3, nil)
	ep.ethClient.Mock.On("BlockByNumber", big.NewInt(4)).Once().Return(L2Block4, nil)

	mockAltDAClient.DropEveryNthPut(2)

	err := bs.StartBatchSubmitting()
	require.NoError(t, err)
	time.Sleep(1 * time.Second) // 1 second is enough to process all blocks at 10ms poll interval
	err = bs.StopBatchSubmitting(context.Background())
	require.NoError(t, err)

	require.Equal(t, 4, mockAltDAClient.StoreCount)
	require.Equal(t, uint64(4), fakeTxMgr.Nonce)
}
