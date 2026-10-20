package batcher

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdkclient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	taggedbase64 "github.com/EspressoSystems/espresso-network/sdks/go/tagged-base64"
	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/espresso"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	txmgrmocks "github.com/ethereum-optimism/optimism/op-service/txmgr/mocks"
)

// contextErrorEspressoClient optionally blocks so tests can inspect the context
// supplied by the real call site. Without calls it fails immediately for lifecycle tests.
type contextErrorEspressoClient struct{ calls chan context.Context }

func (c *contextErrorEspressoClient) sdkContextError(ctx context.Context) error {
	if c.calls == nil {
		return sdkclient.ErrEphemeral
	}
	c.calls <- ctx
	<-ctx.Done()
	return fmt.Errorf("%w: %v", sdkclient.ErrEphemeral, ctx.Err()) //nolint:errorlint // Match the SDK's formatted context error.
}
func (c *contextErrorEspressoClient) SubmitTransaction(ctx context.Context, _ espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
	return nil, c.sdkContextError(ctx)
}
func (c *contextErrorEspressoClient) FetchTransactionByHash(ctx context.Context, _ *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
	return espressoCommon.TransactionQueryData{}, c.sdkContextError(ctx)
}
func (c *contextErrorEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	return 0, fmt.Errorf("fetch latest block height failed with all nodes, Errors: %v", []error{c.sdkContextError(ctx)})
}
func (c *contextErrorEspressoClient) FetchNamespaceTransactionsInRange(ctx context.Context, _, _, _ uint64) ([]espressoCommon.NamespaceTransactionsRangeData, error) {
	return nil, c.sdkContextError(ctx)
}
func (c *contextErrorEspressoClient) FetchHeadersByRange(ctx context.Context, _, _ uint64) ([]espressoCommon.HeaderImpl, error) {
	return nil, c.sdkContextError(ctx)
}

func receiveTestValue[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for test value")
		var zero T
		return zero
	}
}

// Only starting the call differs between the three tests. The assertions cover
// the configured budget, an earlier caller deadline, and caller cancellation.
func testEspressoCallContexts(t *testing.T, start func(*testing.T, *espressoTransactionSubmitter)) {
	t.Helper()
	for _, scenario := range []string{"timeout", "caller deadline", "caller cancellation"} {
		t.Run(scenario, func(t *testing.T) {
			timeout := 20 * time.Millisecond
			ctx, cancel := context.WithCancel(t.Context())
			if scenario == "caller deadline" {
				cancel()
				ctx, cancel = context.WithTimeout(t.Context(), 100*time.Millisecond)
			}
			if scenario != "timeout" {
				timeout = time.Minute
			}
			client := &contextErrorEspressoClient{calls: make(chan context.Context, 1)}
			s := NewEspressoTransactionSubmitter(WithContext(ctx), WithEspressoSubmissionClient(client),
				WithEspressoNetworkTimeout(timeout), WithVerifyReceiptRetryDelay(time.Hour))
			t.Cleanup(func() { cancel(); s.wg.Wait() })
			started := time.Now()
			start(t, s)
			callCtx := receiveTestValue(t, client.calls)
			deadline, ok := callCtx.Deadline()
			require.True(t, ok)
			if scenario == "caller deadline" {
				parentDeadline, _ := ctx.Deadline()
				require.Equal(t, parentDeadline, deadline)
			} else {
				require.False(t, deadline.Before(started.Add(timeout)), "call must receive the full configured budget")
				require.LessOrEqual(t, time.Until(deadline), timeout)
			}
			if scenario == "caller cancellation" {
				cancel()
			}
			receiveTestValue(t, callCtx.Done())
			if scenario == "caller cancellation" {
				require.ErrorIs(t, callCtx.Err(), context.Canceled)
			} else {
				require.ErrorIs(t, callCtx.Err(), context.DeadlineExceeded)
			}
			if scenario == "timeout" {
				require.NoError(t, ctx.Err(), "request timeout must not cancel the parent context")
			}
		})
	}
}

func TestEspressoSubmitWorkerContexts(t *testing.T) {
	testEspressoCallContexts(t, func(t *testing.T, s *espressoTransactionSubmitter) {
		s.SpawnWorkers(1, 0)
		worker := receiveTestValue(t, s.submitWorkerQueue)
		worker <- espressoTransactionJobAttempt{job: espressoSubmitTransactionJob{transaction: &espressoCommon.Transaction{}}, resp: make(chan espressoSubmitTransactionJobResponse, 1)}
	})
}

func TestEspressoVerifyWorkerContexts(t *testing.T) {
	testEspressoCallContexts(t, func(t *testing.T, s *espressoTransactionSubmitter) {
		s.SpawnWorkers(0, 1)
		worker := receiveTestValue(t, s.verifyReceiptWorkerQueue)
		worker <- espressoVerifyReceiptJobAttempt{job: espressoVerifyReceiptJob{hash: new(espressoCommon.TaggedBase64)}, resp: make(chan espressoVerifyReceiptJobResponse, 1)}
	})
}

func TestEspressoHeightTrackerContexts(t *testing.T) {
	testEspressoCallContexts(t, func(t *testing.T, s *espressoTransactionSubmitter) {
		s.wg.Add(1)
		go func() { defer s.wg.Done(); s.trackBlockHeight() }()
	})
}

// espressoServiceTestSetup supplies valid signing and configuration for client
// initialization without an L1 connection or enclave.
func espressoServiceTestSetup(t *testing.T, queryURL string) (*BatcherService, *CLIConfig) {
	t.Helper()
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	manager := new(txmgrmocks.TxManager)
	manager.On("From").Return(crypto.PubkeyToAddress(key.PublicKey))
	manager.On("ChainID").Return(eth.ChainIDFromUInt64(1))

	service := &BatcherService{Log: testlog.Logger(t, log.LevelDebug), TxManager: manager}
	cfg := &CLIConfig{
		L1EthRpc: "http://unused.invalid",
		TxMgrConfig: txmgr.CLIConfig{
			PrivateKey:     hex.EncodeToString(crypto.FromECDSA(key)),
			NetworkTimeout: 10 * time.Millisecond,
		},
		Espresso: espresso.CLIConfig{
			Enabled:                    true,
			QueryServiceURLs:           []string{queryURL, queryURL},
			L1URL:                      "http://unused.invalid",
			LightClientAddr:            common.Address{1},
			EspressoAttestationService: "http://unused.invalid",
			PollInterval:               time.Second,
			VerifyReceiptMaxBlocks:     10,
			VerifyReceiptSafetyTimeout: time.Minute,
			VerifyReceiptRetryDelay:    time.Second,
			TestingBatcherPrivateKey:   key,
		},
	}
	return service, cfg
}

func TestInitEspressoSubmissionTimeoutAndRecovery(t *testing.T) {
	expectedHash, err := taggedbase64.New("TX", []byte("second job"))
	require.NoError(t, err)
	serverRecovered := make(chan struct{})
	requestReceived := make(chan struct{}, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived <- struct{}{}
		select {
		case <-serverRecovered:
			if err := json.NewEncoder(w).Encode(expectedHash); err != nil {
				t.Error(err)
			}
			return
		default:
		}
		// Give cleanup a backstop if a client bound regresses.
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
	}))
	t.Cleanup(server.Close)
	service, cfg := espressoServiceTestSetup(t, server.URL)
	cfg.TxMgrConfig.NetworkTimeout = 100 * time.Millisecond
	require.NoError(t, service.initEspresso(t.Context(), cfg))
	setup := &DriverSetup{}
	service.applyEspressoDriverSetup(setup)
	ctx, cancel := context.WithCancel(t.Context())
	submitter := NewEspressoTransactionSubmitter(WithContext(ctx),
		WithEspressoSubmissionClient(setup.Espresso.Client), WithEspressoNetworkTimeout(cfg.TxMgrConfig.NetworkTimeout))
	submitter.SpawnWorkers(1, 0)
	t.Cleanup(func() { cancel(); submitter.wg.Wait() })
	responses := make(chan espressoSubmitTransactionJobResponse, 1)
	job := espressoTransactionJobAttempt{
		job:  espressoSubmitTransactionJob{transaction: &espressoCommon.Transaction{}},
		resp: responses,
	}
	worker := receiveTestValue(t, submitter.submitWorkerQueue)
	worker <- job
	receiveTestValue(t, requestReceived) // Proves initialization and propagation reached the real SDK/server.
	response := receiveTestValue(t, responses)
	require.ErrorIs(t, response.err, sdkclient.ErrEphemeral)
	require.ErrorContains(t, response.err, context.DeadlineExceeded.Error())
	require.NoError(t, ctx.Err(), "request timeout must not cancel the parent context")

	// A timed-out request must not prevent the same worker from processing another job.
	close(serverRecovered)
	nextWorker := receiveTestValue(t, submitter.submitWorkerQueue)
	require.Equal(t, worker, nextWorker)
	nextWorker <- job
	response = receiveTestValue(t, responses)
	require.NoError(t, response.err)
	require.Equal(t, expectedHash, response.hash)
}

func TestInitEspressoDisabledClientIsNil(t *testing.T) {
	service := &BatcherService{}
	require.NoError(t, service.initEspresso(t.Context(), &CLIConfig{}))
	manager := new(txmgrmocks.TxManager)
	manager.On("From").Return(common.Address{})
	service.TxManager = manager

	setup := &DriverSetup{}
	service.applyEspressoDriverSetup(setup)
	require.Nil(t, setup.Espresso.Client)
}
