package batcher

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

const timeoutWarning = "Espresso request timed out"

func newCapturingTimeoutWarner(t *testing.T) (*espressoTimeoutWarner, *testlog.CapturingHandler) {
	logger, logs := testlog.CaptureLogger(t, slog.LevelWarn)
	return newEspressoTimeoutWarner(logger, "test op", time.Second), logs
}

func TestEspressoTimeoutWarnerClassification(t *testing.T) {
	expired, cancelExpired := context.WithTimeout(context.Background(), 0)
	defer cancelExpired()
	<-expired.Done()
	shutdown, cancelShutdown := context.WithCancel(context.Background())
	shutdownCall, cancelShutdownCall := context.WithTimeout(shutdown, time.Hour)
	defer cancelShutdownCall()
	cancelShutdown()
	pending, cancelPending := context.WithTimeout(context.Background(), time.Hour)
	defer cancelPending()
	// The SDK formats context errors, so classification must ignore err's chain.
	sdkErr := errors.New("request failed: context deadline exceeded")

	for _, tc := range []struct {
		name      string
		parent    context.Context
		callCtx   context.Context
		err       error
		wantsWarn bool
	}{
		{"deadline expired", context.Background(), expired, sdkErr, true},
		{"success at deadline", context.Background(), expired, nil, false},
		{"shutdown cancellation", shutdown, shutdownCall, sdkErr, false},
		{"other failure", context.Background(), pending, context.DeadlineExceeded, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w, logs := newCapturingTimeoutWarner(t)
			w.check(tc.parent, tc.callCtx, tc.err)
			rec := logs.FindLog(testlog.NewMessageContainsFilter(timeoutWarning))
			if !tc.wantsWarn {
				require.Nil(t, rec)
				return
			}
			require.NotNil(t, rec)
			require.Equal(t, "test op", rec.AttrValue("operation"))
			require.Equal(t, time.Second, rec.AttrValue("timeout"))
		})
	}
}

func TestEspressoTimeoutWarnerRateLimitsConcurrentWorkers(t *testing.T) {
	w, logs := newCapturingTimeoutWarner(t)
	expired, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	<-expired.Done()

	const workers = 50
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.check(context.Background(), expired, context.DeadlineExceeded)
		}()
	}
	wg.Wait()
	logs.RequireMessageContainedOnce(t, timeoutWarning)

	// The next warning reports the timeouts suppressed since the previous one,
	// so together the two warnings account for every timeout. A zero Sometimes
	// runs its first Do, allowing exactly one more warning.
	w.limiter = rate.Sometimes{}
	w.check(context.Background(), expired, context.DeadlineExceeded)
	recs := logs.FindLogs(testlog.NewMessageContainsFilter(timeoutWarning))
	require.Len(t, recs, 2)
	require.EqualValues(t, workers+1, recs[0].AttrValue("timeouts").(int64)+recs[1].AttrValue("timeouts").(int64))
}

func TestEspressoRequestsWarnOnTimeout(t *testing.T) {
	newSubmitter := func(t *testing.T) (*espressoTransactionSubmitter, chan context.Context, *testlog.CapturingHandler) {
		logger, logs := testlog.CaptureLogger(t, slog.LevelWarn)
		calls := make(chan context.Context, 1)
		s := NewEspressoTransactionSubmitter(WithContext(t.Context()), WithEspressoLogger(logger),
			WithEspressoSubmissionClient(&contextErrorEspressoClient{calls: calls}),
			WithEspressoNetworkTimeout(time.Millisecond), WithVerifyReceiptRetryDelay(time.Millisecond))
		return s, calls, logs
	}

	t.Run("submit", func(t *testing.T) {
		s, _, logs := newSubmitter(t)
		_, err := s.submitAttempt(s.ctx, &espressoCommon.Transaction{})
		require.Error(t, err)
		logs.RequireMessageContainedOnce(t, timeoutWarning, testlog.NewAttributesFilter("operation", "submit transaction"))
	})

	t.Run("verify", func(t *testing.T) {
		s, calls, logs := newSubmitter(t)
		ctx, cancel := context.WithCancel(s.ctx)
		done := make(chan verifyOutcome, 1)
		go func() { done <- s.confirmReceipt(ctx, new(espressoCommon.TaggedBase64), true) }()
		receiveTestValue(t, calls)
		// The retry lookup starts only after the first one has been checked.
		receiveTestValue(t, calls)
		cancel()
		require.Equal(t, verifyAbort, receiveTestValue(t, done))
		logs.RequireMessageContainedOnce(t, timeoutWarning, testlog.NewAttributesFilter("operation", "verify transaction"))
	})
}
