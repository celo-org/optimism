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

// TestBoundedEspressoClientAppliesDeadline pins that every method of the
// wrapper the submit/verify workers call through carries a per-call deadline:
// a hung endpoint must yield a prompt DeadlineExceeded instead of consuming a
// worker forever.
func TestBoundedEspressoClientAppliesDeadline(t *testing.T) {
	c := newBoundedEspressoClient(blockingEspressoClient{}, 10*time.Millisecond)

	calls := map[string]func(context.Context) error{
		"SubmitTransaction": func(ctx context.Context) error {
			_, err := c.SubmitTransaction(ctx, espressoCommon.Transaction{})
			return err
		},
		"FetchTransactionByHash": func(ctx context.Context) error {
			_, err := c.FetchTransactionByHash(ctx, nil)
			return err
		},
		"FetchLatestBlockHeight": func(ctx context.Context) error {
			_, err := c.FetchLatestBlockHeight(ctx)
			return err
		},
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			// The parent context has no deadline, like the workers' loop
			// contexts; only the wrapper can unblock the call.
			err := call(context.Background())
			require.ErrorIs(t, err, context.DeadlineExceeded)
		})
	}
}
