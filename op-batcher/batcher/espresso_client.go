package batcher

import (
	"context"
	"time"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
)

// EspressoSubmitClient is the slice of the Espresso SDK client consumed by the
// transaction submitter and its workers. Narrowed from the SDK's full
// EspressoClient so the per-call deadline wrapper below only has to cover
// methods that are actually called (the full interface includes streaming
// endpoints, which a per-call deadline would break).
type EspressoSubmitClient interface {
	SubmitTransaction(ctx context.Context, tx espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error)
	FetchTransactionByHash(ctx context.Context, hash *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error)
	FetchLatestBlockHeight(ctx context.Context) (uint64, error)
}

// boundedEspressoClient enforces a per-call deadline on every Espresso SDK
// call. The SDK issues plain HTTP requests with no client-side timeout, and the
// submit/verify workers otherwise call it with their long-lived loop contexts —
// a black-holed connection would hold a worker forever, and with every worker
// wedged, submission stays stopped even after the endpoint recovers.
type boundedEspressoClient struct {
	inner   EspressoSubmitClient
	timeout time.Duration
}

func newBoundedEspressoClient(inner EspressoSubmitClient, timeout time.Duration) *boundedEspressoClient {
	return &boundedEspressoClient{inner: inner, timeout: timeout}
}

func (c *boundedEspressoClient) SubmitTransaction(ctx context.Context, tx espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.inner.SubmitTransaction(ctx, tx)
}

func (c *boundedEspressoClient) FetchTransactionByHash(ctx context.Context, hash *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.inner.FetchTransactionByHash(ctx, hash)
}

func (c *boundedEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.inner.FetchLatestBlockHeight(ctx)
}
