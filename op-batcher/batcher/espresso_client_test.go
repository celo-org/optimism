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
	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
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

// testEspressoCallContexts checks the configured budget, earlier caller deadlines,
// and caller cancellation. start returns a function that waits for the call to finish.
func testEspressoCallContexts(t *testing.T, start func(*testing.T, context.Context, time.Duration, chan context.Context) func(error)) {
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
			expectedErr := context.DeadlineExceeded
			if scenario == "caller cancellation" {
				expectedErr = context.Canceled
			}
			calls := make(chan context.Context, 1)
			var wait func(error)
			t.Cleanup(func() {
				cancel()
				if wait != nil {
					wait(expectedErr)
				}
			})
			started := time.Now()
			wait = start(t, ctx, timeout, calls)
			callCtx := receiveTestValue(t, calls)
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
			require.ErrorIs(t, callCtx.Err(), expectedErr)
			if scenario == "timeout" {
				require.NoError(t, ctx.Err(), "request timeout must not cancel the parent context")
			}
		})
	}
}

func testEspressoWorkerContexts(t *testing.T, start func(*testing.T, *espressoTransactionSubmitter)) {
	t.Helper()
	testEspressoCallContexts(t, func(t *testing.T, ctx context.Context, timeout time.Duration, calls chan context.Context) func(error) {
		s := NewEspressoTransactionSubmitter(WithContext(ctx),
			WithEspressoSubmissionClient(&contextErrorEspressoClient{calls: calls}),
			WithEspressoNetworkTimeout(timeout), WithVerifyReceiptRetryDelay(time.Hour))
		start(t, s)
		return func(error) { s.wg.Wait() }
	})
}

func TestEspressoSubmitWorkerContexts(t *testing.T) {
	testEspressoWorkerContexts(t, func(t *testing.T, s *espressoTransactionSubmitter) {
		s.SpawnWorkers(1, 0)
		s.submitJobQueue <- espressoSubmitTransactionJob{transaction: &espressoCommon.Transaction{}}
	})
}

func TestEspressoVerifyWorkerContexts(t *testing.T) {
	testEspressoWorkerContexts(t, func(t *testing.T, s *espressoTransactionSubmitter) {
		s.SpawnWorkers(0, 1)
		s.verifyReceiptJobQueue <- espressoVerifyReceiptJob{hash: new(espressoCommon.TaggedBase64)}
	})
}

// TestEspressoWorkersProcessJobsConcurrently proves that each shared job queue
// feeds multiple workers: both client calls must start while the first is still
// blocked. Start is intentionally omitted so response handlers cannot retry jobs.
func TestEspressoWorkersProcessJobsConcurrently(t *testing.T) {
	for _, stage := range []struct {
		name          string
		submitWorkers int
		verifyWorkers int
		enqueue       func(*espressoTransactionSubmitter)
	}{
		{
			name: "submit", submitWorkers: 2,
			enqueue: func(s *espressoTransactionSubmitter) {
				s.submitJobQueue <- espressoSubmitTransactionJob{transaction: &espressoCommon.Transaction{}}
			},
		},
		{
			name: "verify", verifyWorkers: 2,
			enqueue: func(s *espressoTransactionSubmitter) {
				s.verifyReceiptJobQueue <- espressoVerifyReceiptJob{hash: new(espressoCommon.TaggedBase64)}
			},
		},
	} {
		t.Run(stage.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			calls := make(chan context.Context, 2)
			s := NewEspressoTransactionSubmitter(WithContext(ctx),
				WithEspressoSubmissionClient(&contextErrorEspressoClient{calls: calls}),
				WithEspressoNetworkTimeout(time.Minute))
			t.Cleanup(func() { cancel(); s.wg.Wait() })
			s.SpawnWorkers(stage.submitWorkers, stage.verifyWorkers)
			stage.enqueue(s)
			stage.enqueue(s)

			first := receiveTestValue(t, calls)
			second := receiveTestValue(t, calls)
			require.NoError(t, first.Err(), "first call must remain blocked while the second starts")
			require.NoError(t, second.Err())
		})
	}
}

func TestEspressoHeightTrackerContexts(t *testing.T) {
	testEspressoWorkerContexts(t, func(t *testing.T, s *espressoTransactionSubmitter) {
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
	// Exactly one worker processes both jobs. Leave Start uncalled so this test
	// is the sole consumer of submitRespQueue, without a competing response handler.
	submitter.SpawnWorkers(1, 0)
	t.Cleanup(func() { cancel(); submitter.wg.Wait() })
	responses := submitter.submitRespQueue
	job := espressoSubmitTransactionJob{transaction: &espressoCommon.Transaction{}}
	submitter.submitJobQueue <- job
	receiveTestValue(t, requestReceived) // Proves initialization and propagation reached the real SDK/server.
	response := receiveTestValue(t, responses)
	require.ErrorIs(t, response.err, sdkclient.ErrEphemeral)
	require.ErrorContains(t, response.err, context.DeadlineExceeded.Error())
	require.NoError(t, ctx.Err(), "request timeout must not cancel the parent context")

	// The sole worker must process the next job after its first request times out.
	close(serverRecovered)
	submitter.submitJobQueue <- job
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

// stalledChainSigner records the request context and waits for cancellation.
type stalledChainSigner struct {
	opcrypto.ChainSigner
	calls chan context.Context
}

func (s *stalledChainSigner) Sign(ctx context.Context, _ []byte) ([]byte, error) {
	s.calls <- ctx
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestQueueBlockToEspressoSigningContext(t *testing.T) {
	testEspressoCallContexts(t, func(t *testing.T, ctx context.Context, timeout time.Duration, calls chan context.Context) func(error) {
		batcher, _ := setup(t, nil)
		batcher.Config.NetworkTimeout = timeout
		batcher.Espresso.ChainSigner = &stalledChainSigner{calls: calls}
		block := newMiniL2BlockWithChainID(0, batcher.RollupConfig.L2ChainID)
		result := make(chan error, 1)
		go func() { result <- batcher.queueBlockToEspresso(ctx, block) }()
		return func(expectedErr error) {
			require.ErrorIs(t, receiveTestValue(t, result), expectedErr)
		}
	})
}
