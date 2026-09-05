package batcher

import (
	"context"
	"testing"
	"time"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/stretchr/testify/require"
)

// blockingEspressoClient hangs every call until its context is cancelled,
// modeling the SDK's timeout-less HTTP client on a black-holed connection.
type blockingEspressoClient struct{}

func (blockingEspressoClient) SubmitTransaction(ctx context.Context, tx espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (blockingEspressoClient) FetchTransactionByHash(ctx context.Context, hash *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
	<-ctx.Done()
	return espressoCommon.TransactionQueryData{}, ctx.Err()
}

func (blockingEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	<-ctx.Done()
	return 0, ctx.Err()
}

// TestBoundedEspressoClientAppliesDeadline pins that every wrapped method
// carries a per-call deadline even when the caller's context has none (as the
// workers' long-lived loop contexts do not).
func TestBoundedEspressoClientAppliesDeadline(t *testing.T) {
	c := newBoundedEspressoClient(blockingEspressoClient{}, 10*time.Millisecond)

	_, err := c.SubmitTransaction(context.Background(), espressoCommon.Transaction{})
	require.ErrorIs(t, err, context.DeadlineExceeded)

	_, err = c.FetchTransactionByHash(context.Background(), nil)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	_, err = c.FetchLatestBlockHeight(context.Background())
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
