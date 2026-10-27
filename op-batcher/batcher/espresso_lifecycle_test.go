package batcher

import (
	"context"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sdkclient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	taggedbase64 "github.com/EspressoSystems/espresso-network/sdks/go/tagged-base64"
	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum-optimism/optimism/op-service/clock"
	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/signer"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// Sleep notifications are sent after the timer is registered, so tests can
// advance time without racing the goroutine that is about to start sleeping.
type espressoTestClock struct {
	*clock.DeterministicClock
	sleeps chan time.Duration
}

func newEspressoTestClock() *espressoTestClock {
	return &espressoTestClock{clock.NewDeterministicClock(time.Unix(1000, 0)), make(chan time.Duration, 8192)}
}

func (c *espressoTestClock) SleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := c.NewTimer(d)
	defer timer.Stop()
	c.sleeps <- d
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.Ch():
		return nil
	}
}

type controlledEspressoClient struct {
	contextErrorEspressoClient
	submit  func(context.Context, espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error)
	receipt func(context.Context, *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error)
	height  func(context.Context) (uint64, error)
}

func (c *controlledEspressoClient) SubmitTransaction(ctx context.Context, tx espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
	if c.submit != nil {
		return c.submit(ctx, tx)
	}
	return taggedbase64.New("TX", []byte{byte(tx.Namespace)})
}
func (c *controlledEspressoClient) FetchTransactionByHash(ctx context.Context, hash *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
	if c.receipt != nil {
		return c.receipt(ctx, hash)
	}
	return espressoCommon.TransactionQueryData{}, sdkclient.ErrEphemeral
}
func (c *controlledEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	if c.height != nil {
		return c.height(ctx)
	}
	return 100, nil
}

func newControlledSubmitter(t *testing.T, client espressoSubmissionClient, options ...EspressoTransactionSubmitterOption) (*espressoTransactionSubmitter, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	opts := []EspressoTransactionSubmitterOption{WithContext(ctx), WithEspressoSubmissionClient(client)}
	s := NewEspressoTransactionSubmitter(append(opts, options...)...)
	t.Cleanup(func() {
		cancel()
		done := make(chan struct{})
		go func() { s.wg.Wait(); close(done) }()
		receiveTestValue(t, done)
	})
	return s, cancel
}

func waitEspressoGroup(t *testing.T, s *espressoTransactionSubmitter) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- s.eg.Wait() }()
	require.NoError(t, receiveTestValue(t, done))
}

func withTestClock(c clock.Clock) EspressoTransactionSubmitterOption {
	return func(cfg *EspressoTransactionSubmitterConfig) { cfg.Clock = c }
}

func withTestInFlightLimit(n int) EspressoTransactionSubmitterOption {
	return func(cfg *EspressoTransactionSubmitterConfig) { cfg.MaxInFlightJobs = n }
}

// Submitter observability.

func TestEspressoSubmitterUsesInjectedLogger(t *testing.T) {
	for _, scenario := range []struct {
		name                  string
		submitErr, receiptErr error
		messages              []string
	}{
		{"confirmed", nil, nil, []string{"Submitted transaction to Espresso", "Transaction confirmed on Espresso"}},
		{"submit failure", sdkclient.ErrPermanent, nil, []string{"Espresso request failed permanently"}},
		{"receipt failure", nil, sdkclient.ErrPermanent, []string{"Submitted transaction to Espresso", "Espresso request failed permanently"}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			logger, logs := testlog.CaptureLogger(t, log.LevelDebug)
			client := &controlledEspressoClient{
				submit: func(context.Context, espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
					if scenario.submitErr != nil {
						return nil, scenario.submitErr
					}
					return taggedbase64.New("TX", []byte{1})
				},
				receipt: func(context.Context, *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
					return espressoCommon.TransactionQueryData{}, scenario.receiptErr
				},
			}
			s, _ := newControlledSubmitter(t, client, WithEspressoLogger(logger.New("component", "espresso-submitter-test")))
			require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{}))
			waitEspressoGroup(t, s) // Read captured records only after the logging goroutine exits.
			records := logs.FindLogs(testlog.NewAttributesFilter("component", "espresso-submitter-test"))
			require.Len(t, records, len(scenario.messages))
			for i, message := range scenario.messages {
				require.Equal(t, message, records[i].Message)
			}
		})
	}
}

// espressoInFlightGauge reads the exported Prometheus series, rather than a test spy.
func espressoInFlightGauge(t *testing.T, metr *metrics.Metrics) float64 {
	t.Helper()
	return opmetrics.NewMetricChecker(t, metr.Registry()).
		FindByName("op_batcher_test_espresso_transactions_in_flight").FindByLabels(nil).Gauge.GetValue()
}

func TestEspressoInFlightGaugeTracksLifecycle(t *testing.T) {
	c := newEspressoTestClock()
	metr := metrics.NewMetrics("test")
	var submissions atomic.Int64
	requests := make(chan receiptRequest, 1)
	heights := make(chan struct{}, 8)
	client := &controlledEspressoClient{
		submit: func(context.Context, espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
			if submissions.Add(1) == 1 {
				return nil, sdkclient.ErrEphemeral
			}
			return taggedbase64.New("TX", []byte{1})
		},
		receipt: func(ctx context.Context, hash *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
			request := receiptRequest{hash, make(chan error, 1)}
			requests <- request
			select {
			case err := <-request.reply:
				return espressoCommon.TransactionQueryData{}, err
			case <-ctx.Done():
				return espressoCommon.TransactionQueryData{}, ctx.Err()
			}
		},
		height: func(context.Context) (uint64, error) {
			heights <- struct{}{}
			return 100, nil
		},
	}
	s, cancel := newControlledSubmitter(t, client, withTestClock(c), withTestInFlightLimit(1), WithEspressoMetrics(metr))
	require.Zero(t, espressoInFlightGauge(t, metr))
	s.Start()
	receiveTestValue(t, heights)
	require.Zero(t, espressoInFlightGauge(t, metr), "the height tracker is not a transaction")
	require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{}))
	delay := receiveTestValue(t, c.sleeps)
	require.Equal(t, 1.0, espressoInFlightGauge(t, metr), "submission backoff retains the lifecycle")
	c.AdvanceTime(delay)
	first := receiveTestValue(t, requests)
	require.Equal(t, 1.0, espressoInFlightGauge(t, metr), "a receipt RPC retains the same lifecycle")
	first.reply <- sdkclient.ErrEphemeral
	delay = receiveTestValue(t, c.sleeps)
	require.Equal(t, 1.0, espressoInFlightGauge(t, metr), "receipt retry waits retain the lifecycle")
	c.AdvanceTime(delay)
	receiveTestValue(t, requests).reply <- nil
	waitEspressoGroup(t, s)
	require.Zero(t, espressoInFlightGauge(t, metr), "confirmation releases the lifecycle")
	cancel()
	s.wg.Wait()
	require.Zero(t, espressoInFlightGauge(t, metr))
}

// Admission and transaction retries.

func TestEspressoRejectsInvalidLimits(t *testing.T) {
	for _, n := range []int{0, -1} {
		require.Panics(t, func() {
			NewEspressoTransactionSubmitter(WithEspressoSubmissionClient(&controlledEspressoClient{}), withTestInFlightLimit(n))
		})
	}
}

// This is concurrent admission, rather than 4096 sequential calls from distinct
// goroutines. Every admitted job stays blocked until all producers have returned.
func TestEspressoConcurrentAdmissionLimit(t *testing.T) {
	calls := make(chan context.Context, 4096)
	metr := metrics.NewMetrics("test")
	s, cancel := newControlledSubmitter(t, &contextErrorEspressoClient{calls: calls}, withTestInFlightLimit(3), WithEspressoNetworkTimeout(time.Minute), WithEspressoMetrics(metr))
	var accepted atomic.Int64
	errors := make(chan error, 4096)
	var producers sync.WaitGroup
	for range 4096 {
		producers.Add(1)
		go func() {
			defer producers.Done()
			if err := s.SubmitTransaction(&espressoCommon.Transaction{}); err != nil {
				errors <- err
			} else {
				accepted.Add(1)
			}
		}()
	}
	done := make(chan struct{})
	go func() { producers.Wait(); close(done) }()
	receiveTestValue(t, done)
	require.EqualValues(t, 3, accepted.Load())
	close(errors)
	for err := range errors {
		var full ErrTooManyInFlightRequests
		require.ErrorAs(t, err, &full)
	}
	for range 3 {
		require.NoError(t, receiveTestValue(t, calls).Err())
	}
	require.Equal(t, 3.0, espressoInFlightGauge(t, metr), "rejected admissions must not change the gauge")
	cancel()
	require.ErrorIs(t, s.SubmitTransaction(&espressoCommon.Transaction{}), context.Canceled)
	s.wg.Wait()
	require.Zero(t, espressoInFlightGauge(t, metr), "cancellation must release every active lifecycle")
}

// Transient failures retain the admitted transaction and back off until the
// endpoint recovers or the session stops. There is no retry-count limit.
func TestEspressoTransientSubmissionRetriesUntilConfirmed(t *testing.T) {
	c := newEspressoTestClock()
	calls := make(chan struct{}, 16)
	var recovered atomic.Bool
	client := &controlledEspressoClient{
		submit: func(context.Context, espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
			calls <- struct{}{}
			if recovered.Load() {
				return taggedbase64.New("TX", []byte{1})
			}
			return nil, sdkclient.ErrEphemeral
		},
		receipt: func(context.Context, *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
			return espressoCommon.TransactionQueryData{}, nil
		},
	}
	s, _ := newControlledSubmitter(t, client, withTestClock(c), withTestInFlightLimit(1))
	require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{}))
	for attempt := range 10 {
		receiveTestValue(t, calls)
		delay := receiveTestValue(t, c.sleeps)
		require.Positive(t, delay, "submission retries must not spin")
		var full ErrTooManyInFlightRequests
		require.ErrorAs(t, s.SubmitTransaction(&espressoCommon.Transaction{}), &full)
		if attempt == 9 {
			recovered.Store(true)
		}
		c.AdvanceTime(delay)
	}
	receiveTestValue(t, calls)
	waitEspressoGroup(t, s)
}

func TestEspressoPermanentFailureReleasesSlotWithoutCancellingOthers(t *testing.T) {
	for _, stage := range []string{"submit", "verify"} {
		t.Run(stage, func(t *testing.T) {
			blocked := make(chan context.Context, 1)
			client := &controlledEspressoClient{
				submit: func(ctx context.Context, tx espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
					if tx.Namespace == 1 {
						if stage == "submit" {
							return nil, sdkclient.ErrPermanent
						}
						return taggedbase64.New("TX", []byte{1})
					}
					blocked <- ctx
					<-ctx.Done()
					return nil, sdkclient.ErrEphemeral
				},
				receipt: func(context.Context, *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
					return espressoCommon.TransactionQueryData{}, sdkclient.ErrPermanent
				},
			}
			metr := metrics.NewMetrics("test")
			s, cancel := newControlledSubmitter(t, client, withTestInFlightLimit(2), WithEspressoMetrics(metr))
			require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{Namespace: 1}))
			require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{Namespace: 2}))
			active := receiveTestValue(t, blocked)
			// The failed transaction must release its slot; the healthy one stays active.
			require.Eventually(t, func() bool {
				err := s.SubmitTransaction(&espressoCommon.Transaction{Namespace: 3})
				if err == nil {
					return true
				}
				var full ErrTooManyInFlightRequests
				require.ErrorAs(t, err, &full)
				return false
			}, time.Second, time.Millisecond)
			replacement := receiveTestValue(t, blocked)
			require.NoError(t, replacement.Err())
			require.NoError(t, active.Err())
			require.NoError(t, s.ctx.Err())
			require.Equal(t, 2.0, espressoInFlightGauge(t, metr), "permanent failure must release its gauge contribution")
			cancel()
			s.wg.Wait()
			require.Zero(t, espressoInFlightGauge(t, metr))
		})
	}
}

// Receipt verification windows.

type receiptRequest struct {
	hash  *espressoCommon.TaggedBase64
	reply chan error
}

// Hold each receipt response until the test has selected its observed height.
// Sleep notifications let the test advance time after the retry timer is ready.
type verificationTest struct {
	submitter   *espressoTransactionSubmitter
	clock       *espressoTestClock
	requests    chan receiptRequest
	submissions *atomic.Int64
}

func newVerificationTest(t *testing.T, options ...EspressoTransactionSubmitterOption) *verificationTest {
	t.Helper()
	c := newEspressoTestClock()
	requests := make(chan receiptRequest, 1)
	submits := new(atomic.Int64)
	client := &controlledEspressoClient{
		submit: func(context.Context, espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
			return taggedbase64.New("TX", []byte{byte(submits.Add(1))})
		},
		receipt: func(ctx context.Context, hash *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
			req := receiptRequest{hash, make(chan error, 1)}
			requests <- req
			select {
			case err := <-req.reply:
				return espressoCommon.TransactionQueryData{}, err
			case <-ctx.Done():
				return espressoCommon.TransactionQueryData{}, ctx.Err()
			}
		},
	}
	opts := []EspressoTransactionSubmitterOption{withTestClock(c), WithVerifyReceiptMaxBlocks(5), WithVerifyReceiptSafetyTimeout(500 * time.Millisecond)}
	s, _ := newControlledSubmitter(t, client, append(opts, options...)...)
	require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{}))
	return &verificationTest{s, c, requests, submits}
}

func (v *verificationTest) retry(t *testing.T, advance time.Duration, height uint64, submissions int64) receiptRequest {
	t.Helper()
	require.Equal(t, 100*time.Millisecond, receiveTestValue(t, v.clock.sleeps))
	v.clock.AdvanceTime(advance)
	request := receiveTestValue(t, v.requests)
	require.Equal(t, submissions, v.submissions.Load())
	v.submitter.latestBlockHeight.Store(height)
	return request
}

// Signal when confirmReceipt reaches its RPC-slot select, so the test can
// advance height and time while capacity is occupied without scheduling sleeps.
type receiptSlotWaitContext struct {
	context.Context
	waiting chan struct{}
	once    sync.Once
}

func (c *receiptSlotWaitContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}

func TestEspressoVerificationWindowsWhileWaitingForSlot(t *testing.T) {
	for _, safetyExpired := range []bool{false, true} {
		name := "height window starts at first lookup"
		if safetyExpired {
			name = "safety deadline includes slot wait"
		}
		t.Run(name, func(t *testing.T) {
			c := newEspressoTestClock()
			s, _ := newControlledSubmitter(t, &controlledEspressoClient{}, withTestClock(c),
				WithVerifyReceiptMaxBlocks(5), WithVerifyReceiptSafetyTimeout(time.Second))
			s.latestBlockHeight.Store(100)
			for range cap(s.verifySlots) {
				s.verifySlots <- struct{}{}
			}
			ctx := &receiptSlotWaitContext{Context: s.ctx, waiting: make(chan struct{})}
			result := make(chan verifyOutcome, 1)
			go func() { result <- s.confirmReceipt(ctx, nil, true) }()
			receiveTestValue(t, ctx.waiting)
			s.latestBlockHeight.Store(105)
			if safetyExpired {
				c.AdvanceTime(time.Second + time.Nanosecond)
			}
			<-s.verifySlots // Allow the first lookup, which returns ErrEphemeral.
			if !safetyExpired {
				// Five blocks passed while waiting, but verification must still
				// get its own five-block window before requesting resubmission.
				delay := receiveTestValue(t, c.sleeps)
				require.Equal(t, 100*time.Millisecond, delay)
				s.latestBlockHeight.Store(110)
				c.AdvanceTime(delay)
			}
			require.Equal(t, verifyResubmit, receiveTestValue(t, result))
		})
	}
}

func TestEspressoVerificationAdoptsHeightAndRestartsWindow(t *testing.T) {
	v := newVerificationTest(t, WithVerifyReceiptSafetyTimeout(time.Hour))
	first := receiveTestValue(t, v.requests)
	first.reply <- sdkclient.ErrEphemeral // The tracker has not produced a height yet.
	v.retry(t, 100*time.Millisecond, 100, 1).reply <- sdkclient.ErrEphemeral
	v.retry(t, 100*time.Millisecond, 104, 1).reply <- sdkclient.ErrEphemeral
	v.retry(t, 100*time.Millisecond, 105, 1).reply <- sdkclient.ErrEphemeral
	second := receiveTestValue(t, v.requests)
	require.EqualValues(t, 2, v.submissions.Load())
	require.NotEqual(t, first.hash.String(), second.hash.String())
	second.reply <- sdkclient.ErrEphemeral
	// The next round starts at 105, rather than retaining the previous window.
	v.retry(t, 100*time.Millisecond, 108, 2).reply <- sdkclient.ErrEphemeral
	v.retry(t, 100*time.Millisecond, 109, 2).reply <- sdkclient.ErrEphemeral
	v.retry(t, 100*time.Millisecond, 110, 2).reply <- sdkclient.ErrEphemeral
	third := receiveTestValue(t, v.requests)
	require.EqualValues(t, 3, v.submissions.Load())
	third.reply <- nil
	waitEspressoGroup(t, v.submitter)
}

func TestEspressoVerificationRestartsSafetyTimeout(t *testing.T) {
	v := newVerificationTest(t)
	first := receiveTestValue(t, v.requests)
	first.reply <- sdkclient.ErrEphemeral
	v.retry(t, 500*time.Millisecond+time.Nanosecond, 0, 1).reply <- sdkclient.ErrEphemeral
	second := receiveTestValue(t, v.requests)
	require.EqualValues(t, 2, v.submissions.Load())
	require.NotEqual(t, first.hash.String(), second.hash.String())
	second.reply <- sdkclient.ErrEphemeral
	v.retry(t, 100*time.Millisecond, 0, 2).reply <- sdkclient.ErrEphemeral
	v.retry(t, 400*time.Millisecond+time.Nanosecond, 0, 2).reply <- sdkclient.ErrEphemeral
	third := receiveTestValue(t, v.requests)
	require.EqualValues(t, 3, v.submissions.Load())
	third.reply <- nil
	waitEspressoGroup(t, v.submitter)
}

// RPC concurrency and receipt polling budget.

// Keep this expectation independent of the implementation constant so an
// accidental change to the previous four-worker limit fails the tests.
const expectedEspressoRPCConcurrency = 4

// Initial queries stay responsive during catch-up while repeated polling has one
// session-wide token budget of four retries per configured interval.
func TestEspressoReceiptRetryBudgetAndFirstLookup(t *testing.T) {
	for _, interval := range []time.Duration{100 * time.Millisecond, time.Second} {
		t.Run(interval.String(), func(t *testing.T) {
			c := newEspressoTestClock()
			queries := make(chan string, 1024)
			client := &controlledEspressoClient{receipt: func(ctx context.Context, hash *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
				queries <- hash.String()
				return espressoCommon.TransactionQueryData{}, sdkclient.ErrEphemeral
			}}
			s, _ := newControlledSubmitter(t, client, withTestClock(c), WithVerifyReceiptRetryDelay(interval))
			for i := 1; i <= 127; i++ {
				require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{Namespace: uint64(i)}))
			}
			for range 127 {
				receiveTestValue(t, queries)
				require.Equal(t, interval, receiveTestValue(t, c.sleeps))
			}
			c.AdvanceTime(interval)
			for range 127 {
				receiveTestValue(t, c.sleeps)
			}
			require.Len(t, queries, expectedEspressoRPCConcurrency)
			for range expectedEspressoRPCConcurrency {
				receiveTestValue(t, queries)
			}
			require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{Namespace: 255}))
			hash, err := taggedbase64.New("TX", []byte{255})
			require.NoError(t, err)
			require.Equal(t, hash.String(), receiveTestValue(t, queries), "new first lookup must bypass queued retry reservations without advancing time")
			require.Equal(t, interval, receiveTestValue(t, c.sleeps))
			c.AdvanceTime(interval)
			// The sustained budget is four more retries per configured interval.
			for range expectedEspressoRPCConcurrency {
				receiveTestValue(t, queries)
			}
			require.Never(t, func() bool { return len(queries) > 0 }, 30*time.Millisecond, time.Millisecond)
		})
	}
}

// Also preserves the overlap regression: all four calls must enter while
// earlier calls remain blocked, then additional calls must wait for a slot.
func TestEspressoRPCConcurrencyLimits(t *testing.T) {
	for _, stage := range []string{"submit", "verify"} {
		t.Run(stage, func(t *testing.T) {
			entered := make(chan struct{}, 128)
			var active, peak atomic.Int64
			block := func(ctx context.Context) {
				n := active.Add(1)
				defer active.Add(-1)
				for previous := peak.Load(); n > previous; previous = peak.Load() {
					if peak.CompareAndSwap(previous, n) {
						break
					}
				}
				entered <- struct{}{}
				<-ctx.Done()
			}
			client := &controlledEspressoClient{}
			if stage == "submit" {
				client.submit = func(ctx context.Context, _ espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
					block(ctx)
					return nil, ctx.Err()
				}
			} else {
				client.receipt = func(ctx context.Context, _ *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
					block(ctx)
					return espressoCommon.TransactionQueryData{}, ctx.Err()
				}
			}
			s, _ := newControlledSubmitter(t, client, WithEspressoNetworkTimeout(time.Minute))
			for range 128 {
				require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{}))
			}
			for range expectedEspressoRPCConcurrency {
				receiveTestValue(t, entered)
			}
			select {
			case <-entered:
				t.Fatal("client calls exceeded the RPC concurrency budget")
			case <-time.After(30 * time.Millisecond):
			}
			require.EqualValues(t, expectedEspressoRPCConcurrency, peak.Load())
		})
	}
}

// Shutdown during RPCs and retry waits.

func stopEspressoTestBatcher(t *testing.T, s *espressoTransactionSubmitter, cancel context.CancelFunc) {
	t.Helper()
	batcher := &BatchSubmitter{
		DriverSetup: DriverSetup{Log: testlog.Logger(t, log.LevelDebug)},
		wg:          s.wg, running: true, cancelShutdownCtx: cancel, cancelKillCtx: func() {},
		espressoSession: &espressoSession{submitter: s, streamer: &stopRecordingStreamer{}},
	}
	stopped := make(chan error, 1)
	go func() { stopped <- batcher.StopBatchSubmitting(t.Context()) }()
	require.NoError(t, receiveTestValue(t, stopped))
	require.Nil(t, batcher.espressoSession)
}

func TestEspressoStopCancelsSubmissionBackoff(t *testing.T) {
	c := newEspressoTestClock()
	s, cancel := newControlledSubmitter(t, &controlledEspressoClient{
		submit: func(context.Context, espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
			return nil, sdkclient.ErrEphemeral
		},
	}, withTestClock(c))
	require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{}))
	require.Positive(t, receiveTestValue(t, c.sleeps))
	stopEspressoTestBatcher(t, s, cancel)
}

func TestEspressoStopCancelsSubmissionRPC(t *testing.T) {
	calls := make(chan context.Context, 1)
	s, cancel := newControlledSubmitter(t, &contextErrorEspressoClient{calls: calls})
	require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{}))
	receiveTestValue(t, calls)
	stopEspressoTestBatcher(t, s, cancel)
}

func TestEspressoStopCancelsReceiptBudgetWait(t *testing.T) {
	c := newEspressoTestClock()
	s, cancel := newControlledSubmitter(t, &controlledEspressoClient{}, withTestClock(c))
	s.receiptRetries = rate.NewLimiter(rate.Every(time.Hour), 1)
	s.receiptRetries.ReserveN(c.Now(), 1)
	require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{}))
	require.Equal(t, 100*time.Millisecond, receiveTestValue(t, c.sleeps))
	c.AdvanceTime(100 * time.Millisecond)
	require.Greater(t, receiveTestValue(t, c.sleeps), time.Minute)
	stopEspressoTestBatcher(t, s, cancel)
}

func TestEspressoStopCancelsReceiptAndWaitsForTracker(t *testing.T) {
	receiptStarted := make(chan struct{}, 1)
	heightStarted := make(chan struct{}, 1)
	heightCancelled := make(chan struct{}, 1)
	releaseHeight := make(chan struct{})
	var release sync.Once
	t.Cleanup(func() { release.Do(func() { close(releaseHeight) }) })
	client := &controlledEspressoClient{
		receipt: func(ctx context.Context, _ *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
			receiptStarted <- struct{}{}
			<-ctx.Done()
			return espressoCommon.TransactionQueryData{}, ctx.Err()
		},
		height: func(ctx context.Context) (uint64, error) {
			heightStarted <- struct{}{}
			<-ctx.Done()
			heightCancelled <- struct{}{}
			<-releaseHeight
			return 0, ctx.Err()
		},
	}
	s, cancel := newControlledSubmitter(t, client)
	s.Start()
	require.NoError(t, s.SubmitTransaction(&espressoCommon.Transaction{}))
	receiveTestValue(t, receiptStarted)
	receiveTestValue(t, heightStarted)
	batcher := &BatchSubmitter{
		DriverSetup: DriverSetup{Log: testlog.Logger(t, log.LevelDebug)},
		wg:          s.wg, running: true, cancelShutdownCtx: cancel, cancelKillCtx: func() {},
		espressoSession: &espressoSession{submitter: s, streamer: &stopRecordingStreamer{}},
	}
	stopped := make(chan error, 1)
	go func() { stopped <- batcher.StopBatchSubmitting(t.Context()) }()
	receiveTestValue(t, heightCancelled)
	waitEspressoGroup(t, s)
	select {
	case <-stopped:
		t.Fatal("stop returned before the height tracker exited")
	case <-time.After(30 * time.Millisecond):
	}
	release.Do(func() { close(releaseHeight) })
	require.NoError(t, receiveTestValue(t, stopped))
	require.Nil(t, batcher.espressoSession)
}

// Queue-loop retry integration.

// newQueueRetryBatcher provides two consecutive unsafe blocks and a real
// local signer. The returned hashes identify their expected signed transactions.
func newQueueRetryBatcher(t *testing.T) (*BatchSubmitter, []string) {
	t.Helper()
	batcher, endpoints := setup(t, nil)
	batcher.Config.PollInterval = time.Millisecond
	batcher.Config.NetworkTimeout = time.Second
	factory, address, err := opcrypto.ChainSignerFactoryFromConfig(batcher.Log,
		"0000000000000000000000000000000000000000000000000000000000000001", "", "", signer.CLIConfig{})
	require.NoError(t, err)
	batcher.Espresso.ChainSigner = factory(batcher.RollupConfig.L2ChainID, address)

	safe := eth.L2BlockRef{Number: 10, Hash: common.Hash{10}}
	block11 := newMiniL2BlockWithChainIDAndNumberParent(0, batcher.RollupConfig.L2ChainID, big.NewInt(11), safe.Hash)
	block12 := newMiniL2BlockWithChainIDAndNumberParent(0, batcher.RollupConfig.L2ChainID, big.NewInt(12), block11.Hash())
	endpoints.rollupClient.Mock.On("SyncStatus").Return(&eth.SyncStatus{
		HeadL1:      eth.L1BlockRef{Number: genesisL1Origin},
		CurrentL1:   eth.L1BlockRef{Number: genesisL1Origin},
		LocalSafeL2: safe,
		UnsafeL2:    eth.L2BlockRef{Number: 12, Hash: block12.Hash()},
	}, nil)
	hashes := make([]string, 0, 2)
	for _, block := range []*types.Block{block11, block12} {
		endpoints.ethClient.Mock.On("BlockByNumber", block.Number()).Return(block, nil)
		batch, err := derivation.BlockToEspressoBatch(batcher.RollupConfig, block)
		require.NoError(t, err)
		tx, err := batch.ToEspressoTransaction(t.Context(), bigs.Uint64Strict(batcher.RollupConfig.L2ChainID), batcher.Espresso.ChainSigner)
		require.NoError(t, err)
		commitment := tx.Commit()
		hash, err := taggedbase64.New("TX", commitment[:])
		require.NoError(t, err)
		hashes = append(hashes, hash.String())
	}
	return batcher, hashes
}

// queueRetryClient transiently fails the first block once and holds the second in
// flight until the first confirms, exposing any reset-driven duplicate work.
type queueRetryClient struct {
	controlledEspressoClient
	failedHash          string
	failStage           string
	submissions         atomic.Int64
	otherSubmissions    atomic.Int64
	otherReceiptStarted chan struct{}
	releaseOther        chan struct{}
	firstConfirmed      sync.Once
	failed              atomic.Bool
	confirmed           chan string
}

func (c *queueRetryClient) SubmitTransaction(_ context.Context, tx espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
	commitment := tx.Commit()
	hash, err := taggedbase64.New("TX", commitment[:])
	if err != nil {
		return nil, err
	}
	if hash.String() == c.failedHash {
		c.submissions.Add(1)
		if c.failStage == "submit" && c.failed.CompareAndSwap(false, true) {
			return nil, sdkclient.ErrEphemeral
		}
	} else {
		c.otherSubmissions.Add(1)
	}
	return hash, nil
}

func (c *queueRetryClient) FetchTransactionByHash(ctx context.Context, hash *espressoCommon.TaggedBase64) (espressoCommon.TransactionQueryData, error) {
	if hash.String() == c.failedHash && c.failStage == "receipt" && c.failed.CompareAndSwap(false, true) {
		return espressoCommon.TransactionQueryData{}, sdkclient.ErrEphemeral
	}
	if hash.String() == c.failedHash {
		c.firstConfirmed.Do(func() { close(c.releaseOther) })
	} else {
		c.otherReceiptStarted <- struct{}{}
		select {
		case <-c.releaseOther:
		case <-ctx.Done():
			return espressoCommon.TransactionQueryData{}, ctx.Err()
		}
	}
	c.confirmed <- hash.String()
	return espressoCommon.TransactionQueryData{}, nil
}

func startEspressoQueueLoop(t *testing.T, batcher *BatchSubmitter, client espressoSubmissionClient, options ...EspressoTransactionSubmitterOption) *espressoTransactionSubmitter {
	t.Helper()
	s, cancel := newControlledSubmitter(t, client, options...)
	batcher.wg = s.wg
	batcher.running = true
	batcher.cancelShutdownCtx = cancel
	batcher.cancelKillCtx = func() {}
	batcher.espressoSession = &espressoSession{submitter: s, streamer: &stopRecordingStreamer{}}
	s.Start()
	s.wg.Add(1)
	go batcher.espressoBatchQueueingLoop(s.ctx, s.wg)
	return s
}

// Exercise fetching, signing, retrying and confirmation without resetting the
// queue or replaying a healthy block that is still being verified.
func TestEspressoQueueLoopRetriesWithoutReplayingOtherBlocks(t *testing.T) {
	for _, scenario := range []struct {
		stage           string
		wantSubmissions int64
	}{
		{"submit", 2}, {"receipt", 1},
	} {
		t.Run(scenario.stage, func(t *testing.T) {
			batcher, hashes := newQueueRetryBatcher(t)
			c := newEspressoTestClock()
			client := &queueRetryClient{
				failedHash: hashes[0], failStage: scenario.stage, confirmed: make(chan string, 16),
				otherReceiptStarted: make(chan struct{}, 16), releaseOther: make(chan struct{}),
			}
			s := startEspressoQueueLoop(t, batcher, client, withTestClock(c))

			receiveTestValue(t, client.otherReceiptStarted)
			c.AdvanceTime(receiveTestValue(t, c.sleeps))
			seen := make(map[string]bool)
			for !seen[hashes[0]] || !seen[hashes[1]] {
				seen[receiveTestValue(t, client.confirmed)] = true
			}
			waitEspressoGroup(t, s)
			stopped := make(chan error, 1)
			go func() { stopped <- batcher.StopBatchSubmitting(t.Context()) }()
			require.NoError(t, receiveTestValue(t, stopped))

			require.True(t, client.failed.Load(), "the injected transient error must have been exercised")
			require.Equal(t, scenario.wantSubmissions, client.submissions.Load())
			require.EqualValues(t, 1, client.otherSubmissions.Load(), "the healthy in-flight block must not be replayed")
			require.False(t, batcher.clearStateRequested.Load(), "an SDK error must not reset unrelated work")
			require.Nil(t, batcher.espressoSession)
		})
	}
}
