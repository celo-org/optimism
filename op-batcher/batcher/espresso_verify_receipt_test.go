package batcher

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	tagged_base64 "github.com/EspressoSystems/espresso-network/sdks/go/tagged-base64"
	espressoTypes "github.com/EspressoSystems/espresso-network/sdks/go/types"
	espressoTypesCommon "github.com/EspressoSystems/espresso-network/sdks/go/types/common"
	"github.com/stretchr/testify/require"
)

// TestEvaluateVerification locks the receipt-verification decision table, in
// particular the two re-submission timeouts: the HotShot block-count timeout
// (with its zero-startHeight guard for a tracker that has not produced a
// height yet) and the wall-clock safety backstop for a stale or broken
// tracker. A wrong verdict here either drops a batch (skipping a retryable
// job) or re-submits forever.
func TestEvaluateVerification(t *testing.T) {
	const maxBlocks = 10
	const safetyTimeout = time.Hour

	ephemeralErr := fmt.Errorf("%w: receipt not found", espressoClient.ErrEphemeral)
	permanentErr := fmt.Errorf("%w: malformed transaction", espressoClient.ErrPermanent)

	tests := []struct {
		name          string
		err           error
		startHeight   uint64
		currentHeight uint64
		startedAgo    time.Duration
		want          JobEvaluation
	}{
		{
			name: "success is handled",
			want: Handle,
		},
		{
			name: "permanent error is skipped",
			err:  permanentErr,
			// Both timeouts exceeded: a permanent error must win over retry.
			startHeight:   100,
			currentHeight: 100 + maxBlocks,
			startedAgo:    2 * safetyTimeout,
			want:          Skip,
		},
		{
			name:          "ephemeral error within both budgets retries verification",
			err:           ephemeralErr,
			startHeight:   100,
			currentHeight: 100 + maxBlocks - 1,
			want:          RetryVerification,
		},
		{
			name:          "exactly maxBlocks elapsed re-submits",
			err:           ephemeralErr,
			startHeight:   100,
			currentHeight: 100 + maxBlocks,
			want:          RetrySubmission,
		},
		{
			name:          "far past maxBlocks re-submits",
			err:           ephemeralErr,
			startHeight:   100,
			currentHeight: 100 + 100*maxBlocks,
			want:          RetrySubmission,
		},
		{
			// The tracker had not stored a height when the job started, so
			// block counting is meaningless: the guard must fall through to
			// verification retry, not treat height 0 as "maxBlocks passed".
			name:          "zero startHeight never triggers the block-count timeout",
			err:           ephemeralErr,
			startHeight:   0,
			currentHeight: 1 << 40,
			want:          RetryVerification,
		},
		{
			// Tracker stale or broken (currentHeight stuck at startHeight):
			// the wall clock is the backstop.
			name:          "wall-clock safety timeout re-submits",
			err:           ephemeralErr,
			startHeight:   100,
			currentHeight: 100,
			startedAgo:    safetyTimeout + time.Minute,
			want:          RetrySubmission,
		},
		{
			name:          "unclassified error is retried like an ephemeral one",
			err:           errors.New("connection reset"),
			startHeight:   100,
			currentHeight: 100 + maxBlocks - 1,
			want:          RetryVerification,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &espressoTransactionSubmitter{
				verifyReceiptMaxBlocks:     maxBlocks,
				verifyReceiptSafetyTimeout: safetyTimeout,
			}
			resp := espressoVerifyReceiptJobResponse{
				job: espressoVerifyReceiptJob{
					startHeight: test.startHeight,
					startTime:   time.Now().Add(-test.startedAgo),
				},
				err:           test.err,
				currentHeight: test.currentHeight,
			}
			require.Equal(t, test.want, s.evaluateVerification(resp))
		})
	}
}

// stuckReceiptEspressoClient accepts submissions and serves an advancing
// HotShot block height, but never finds a submitted transaction: the shape of
// an Espresso node that swallowed a transaction. The embedded nil interface
// panics on any other method, proving no other client call is involved.
type stuckReceiptEspressoClient struct {
	espressoClient.EspressoClient
	height      atomic.Uint64
	heightStep  uint64
	fetches     atomic.Int64
	submissions atomic.Int64
}

func (c *stuckReceiptEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	c.fetches.Add(1)
	return c.height.Add(c.heightStep), nil
}

func (c *stuckReceiptEspressoClient) SubmitTransaction(ctx context.Context, tx espressoTypesCommon.Transaction) (*espressoTypesCommon.TaggedBase64, error) {
	c.submissions.Add(1)
	return tagged_base64.New("TX", tx.Payload)
}

func (c *stuckReceiptEspressoClient) FetchTransactionByHash(ctx context.Context, hash *espressoTypes.TaggedBase64) (espressoTypes.TransactionQueryData, error) {
	return espressoTypes.TransactionQueryData{}, fmt.Errorf("%w: transaction not yet in a block", espressoClient.ErrEphemeral)
}

// TestVerifyReceiptBlockCountTimeoutResubmits runs the whole submitter
// pipeline against a client whose transactions never become queryable while
// HotShot keeps producing blocks, and requires the transaction to be
// re-submitted. This is the end-to-end path of the block-count timeout: the
// tracker feeding latestBlockHeight, the worker snapshotting startHeight on
// the first attempt, and evaluateVerification routing the job back to the
// submission queue.
func TestVerifyReceiptBlockCountTimeoutResubmits(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	client := &stuckReceiptEspressoClient{heightStep: 10}
	client.height.Store(100)

	submitter := NewEspressoTransactionSubmitter(
		WithContext(ctx),
		WithEspressoClient(client),
		WithVerifyReceiptMaxBlocks(5),
		// Keep the wall-clock backstop out of play so only the block-count
		// path can trigger the re-submission this test requires.
		WithVerifyReceiptSafetyTimeout(time.Hour),
		WithVerifyReceiptRetryDelay(time.Millisecond),
	)
	submitter.SpawnWorkers(1, 1)
	submitter.Start()

	// Wait until the tracker has completed a fetch-and-store round (the
	// second fetch proves the first store), so the verify job's startHeight
	// snapshot is nonzero — a zero snapshot disables the block-count timeout
	// by design.
	require.Eventually(t, func() bool { return client.fetches.Load() >= 2 }, 10*time.Second, time.Millisecond,
		"the block height tracker never polled the client")

	txn := &espressoTypesCommon.Transaction{Namespace: 1, Payload: make([]byte, 8)}
	require.NoError(t, submitter.SubmitTransaction(txn))

	require.Eventually(t, func() bool { return client.submissions.Load() >= 2 }, 30*time.Second, time.Millisecond,
		"the verification timeout never re-submitted the stuck transaction")
}
