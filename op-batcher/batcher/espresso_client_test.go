package batcher

import (
	"context"
	"testing"
	"time"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/stretchr/testify/require"
)

type testEspressoClient struct{}

func (c *testEspressoClient) waitForDeadline(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (c *testEspressoClient) SubmitTransaction(ctx context.Context, _ espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
	return nil, c.waitForDeadline(ctx)
}

func (c *testEspressoClient) FetchTransactionByHash(ctx context.Context, _ *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
	return espressoCommon.TransactionQueryData{}, c.waitForDeadline(ctx)
}

func (c *testEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	return 0, c.waitForDeadline(ctx)
}

func (c *testEspressoClient) FetchNamespaceTransactionsInRange(ctx context.Context, _, _, _ uint64) ([]espressoCommon.NamespaceTransactionsRangeData, error) {
	return nil, c.waitForDeadline(ctx)
}

func (c *testEspressoClient) FetchHeadersByRange(ctx context.Context, _, _ uint64) ([]espressoCommon.HeaderImpl, error) {
	return nil, c.waitForDeadline(ctx)
}

func TestEspressoClientTimeout(t *testing.T) {
	const timeout = 10 * time.Millisecond
	client := &espressoClient{
		networkTimeoutCtx: func(ctx context.Context) (context.Context, context.CancelFunc) {
			return context.WithTimeout(ctx, timeout)
		},
		client: &testEspressoClient{},
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	start := time.Now()
	_, submitErr := client.SubmitTransaction(ctx, espressoCommon.Transaction{})
	_, transactionErr := client.FetchTransactionByHash(ctx, nil)
	_, heightErr := client.FetchLatestBlockHeight(ctx)
	_, transactionsErr := client.FetchNamespaceTransactionsInRange(ctx, 0, 1, 2)
	_, headersErr := client.FetchHeadersByRange(ctx, 0, 1)
	require.Less(t, time.Since(start), 500*time.Millisecond)

	for _, err := range []error{submitErr, transactionErr, heightErr, transactionsErr, headersErr} {
		require.ErrorIs(t, err, context.DeadlineExceeded)
	}
}
