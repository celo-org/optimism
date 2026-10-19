package batcher

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdkclient "github.com/EspressoSystems/espresso-network/sdks/go/client"
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

// contextError fails the fake call if it receives an active context.
func contextError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return errors.New("expected a canceled or expired call context")
}

type contextErrorEspressoClient struct{}

func (c *contextErrorEspressoClient) sdkContextError(ctx context.Context) error {
	return fmt.Errorf("%w: %v", sdkclient.ErrEphemeral, contextError(ctx)) //nolint:errorlint // Match the SDK: the context error is formatted, not wrapped.
}

func (c *contextErrorEspressoClient) SubmitTransaction(ctx context.Context, _ espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
	return nil, c.sdkContextError(ctx)
}
func (c *contextErrorEspressoClient) FetchTransactionByHash(ctx context.Context, _ *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
	return espressoCommon.TransactionQueryData{}, c.sdkContextError(ctx)
}
func (c *contextErrorEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	// MultipleNodesClient loses the sentinel when it formats its collected errors.
	return 0, fmt.Errorf("fetch latest block height failed with all nodes, Errors: %v", []error{c.sdkContextError(ctx)})
}
func (c *contextErrorEspressoClient) FetchNamespaceTransactionsInRange(ctx context.Context, _, _, _ uint64) ([]espressoCommon.NamespaceTransactionsRangeData, error) {
	return nil, c.sdkContextError(ctx)
}
func (c *contextErrorEspressoClient) FetchHeadersByRange(ctx context.Context, _, _ uint64) ([]espressoCommon.HeaderImpl, error) {
	return nil, c.sdkContextError(ctx)
}

var espressoClientMethods = []struct {
	name string
	call func(context.Context, espressoRuntimeClient) error
}{
	{"SubmitTransaction", func(ctx context.Context, c espressoRuntimeClient) error {
		_, err := c.SubmitTransaction(ctx, espressoCommon.Transaction{})
		return err
	}},
	{"FetchTransactionByHash", func(ctx context.Context, c espressoRuntimeClient) error {
		_, err := c.FetchTransactionByHash(ctx, new(espressoCommon.TaggedBase64))
		return err
	}},
	{"FetchLatestBlockHeight", func(ctx context.Context, c espressoRuntimeClient) error {
		_, err := c.FetchLatestBlockHeight(ctx)
		return err
	}},
	{"FetchNamespaceTransactionsInRange", func(ctx context.Context, c espressoRuntimeClient) error {
		_, err := c.FetchNamespaceTransactionsInRange(ctx, 0, 1, 2)
		return err
	}},
	{"FetchHeadersByRange", func(ctx context.Context, c espressoRuntimeClient) error {
		_, err := c.FetchHeadersByRange(ctx, 0, 1)
		return err
	}},
}

func newExpiredDeadlineTestClient() *espressoTimeoutClient {
	return &espressoTimeoutClient{
		client: &contextErrorEspressoClient{},
		networkTimeoutCtx: func(ctx context.Context) (context.Context, context.CancelFunc) {
			// An expired deadline checks context propagation without waiting for a timer.
			return context.WithDeadline(ctx, time.Unix(1, 0))
		},
	}
}

func newFutureTimeoutTestClient() *espressoTimeoutClient {
	return &espressoTimeoutClient{
		client: &contextErrorEspressoClient{},
		networkTimeoutCtx: func(ctx context.Context) (context.Context, context.CancelFunc) {
			return context.WithTimeout(ctx, time.Minute)
		},
	}
}

func requireEspressoTimeout(t *testing.T, method string, err error) {
	t.Helper()
	require.ErrorContains(t, err, context.DeadlineExceeded.Error())
	require.NotErrorIs(t, err, context.DeadlineExceeded, "SDK formats the underlying context error with %%v")
	if method == "FetchLatestBlockHeight" {
		require.NotErrorIs(t, err, sdkclient.ErrEphemeral)
	} else {
		require.ErrorIs(t, err, sdkclient.ErrEphemeral)
	}
}

func TestEspressoTimeoutClientBoundsCalls(t *testing.T) {
	for _, method := range espressoClientMethods {
		t.Run(method.name, func(t *testing.T) {
			client := newExpiredDeadlineTestClient()
			err := method.call(t.Context(), client)
			requireEspressoTimeout(t, method.name, err)
		})
	}
}

func TestEspressoTimeoutClientHonorsCallerContext(t *testing.T) {
	for _, method := range espressoClientMethods {
		t.Run(method.name, func(t *testing.T) {
			client := newFutureTimeoutTestClient()
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			require.ErrorContains(t, method.call(ctx, client), context.Canceled.Error())
			ctx, cancel = context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
			defer cancel()
			require.ErrorContains(t, method.call(ctx, client), context.DeadlineExceeded.Error())
		})
	}
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

func TestInitEspressoBoundsSubmission(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bound cleanup even if the client timeout is removed.
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
	}))
	t.Cleanup(server.Close)

	service, cfg := espressoServiceTestSetup(t, server.URL)
	require.NoError(t, service.initEspresso(t.Context(), cfg))
	setup := &DriverSetup{}
	service.applyEspressoDriverSetup(setup)

	_, err := setup.Espresso.Client.SubmitTransaction(t.Context(), espressoCommon.Transaction{})
	require.ErrorIs(t, err, sdkclient.ErrEphemeral)
	require.ErrorContains(t, err, context.DeadlineExceeded.Error())
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
