package batcher_test

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/stretchr/testify/require"
)

const NUMBER_OF_TRANSACTIONS_TO_SUBMIT = 1024 * 4

// TestEspressoTransactionSubmitterDeadlock is a test that is meant to test
// the underlying conditions in the espressoTransactionSubmitter that may
// result in a deadlock.
//
// The basic idea is to intentionally trigger the circumstances that can
// trigger the deadlock to occur in the espressoTransactionSubmitter. This
// test with this description **SHOULD** remain as a regression test.
//
// The specific suspected criteria for triggering a deadlock in the
// implementation of the espressoTransactionSubmitter  as of 2026-04-23 is
// as follows:
//
// - Assume that the Espresso Service is not ever returning successfully
// - Assume we are receiving a consistent stream of of new blocks coming in
//
// If we have both of these criteria, it should be possible to fill up the
// underlying channels of `espressoTransactionSubmitter` for both Submission
// jobs, and verification jobs.
//
// The way we *should* be able to detect the deadlock is if a call to
// SubmitTransaction blocks.
//
// NOTE: After this has been fixed, it is unlikely that this will fail in the
// future. This will primarily be due to `SubmitTransaction` being modified
// to longer explicitly block.
func TestEspressoTransactionSubmitterDeadlock(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	submitter := batcher.NewEspressoTransactionSubmitter(
		batcher.WithEspressoClient(new(AlwaysFailingEspressoClient)),
		batcher.WithContext(ctx),
	)

	submitter.SpawnWorkers(4, 4)
	submitter.Start()

	for i := 0; i < NUMBER_OF_TRANSACTIONS_TO_SUBMIT; i++ {
		txn := espressoCommon.Transaction{
			Namespace: 1,
			Payload:   make([]byte, 8),
		}
		binary.LittleEndian.PutUint64(txn.Payload, uint64(i))

		submitCtx, submitCancel := context.WithCancel(context.Background())
		go (func(cancel context.CancelFunc, txn espressoCommon.Transaction) {
			submitter.SubmitTransaction(&txn)
			cancel()
		})(submitCancel, txn)

		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)

		select {
		case <-submitCtx.Done():
			// Things progressed without issue.
			timeoutCancel()
			submitCancel()
			continue
		case <-timeoutCtx.Done():
			// The call to SubmitTransaction did not return within the expected
			// time frame.

			timeoutCancel()
			submitCancel()
			t.Fatalf("SubmitTransaction has blocked for longer than 200 milliseconds, on transaction %d, for error: %s\n", i, timeoutCtx.Err())
		}
	}

	// If we've gotten here, we have passed without deadlocking.
}

// TestEspressoTransactionSubmitterProgress is a test that is meant to test
// and verify the modified behavior of the espressoTransactionSubmitter after
// the behavior change to explicitly address the potential for the previous
// deadlock condition.
//
// The resolution is to target and limit the number of active inflight requests
// pending to be submitted and verified on Espresso at a given time. We target
// this behavior explicitly by being able to configure a maximum number of
// pending or "inflight" requests that are being waited on.  By setting this
// limit, and being able to track things going through the pipeline, we can
// ensure that we continue to make progress, and put back pressure on the
// submitter themselves, so that they don't keep trying to submit new
// transactions if we are unable to effectively handle them at our current
// capacity.
//
// This test ensures that this workflow can be processed by first starting with
// a failing EspressoClient, and once we hit the threshold of having too
// many in flight requests, we swap the Client over to a succeeding client,
// and ensure that we're able to get through our pending backlog, and all new
// transactions to submit without stalling.
func TestEspressoTransactionSubmitterProgress(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	failingClient := new(AlwaysFailingEspressoClient)
	succeedingClient := new(FakeSubmissionSucceedingEspressoClient)
	succeedingClient.Init()
	succeedingClient.EspressoClient = failingClient
	espClient := new(EspressoClientSwappableImplementation)
	espClient.SetEspressoClient(failingClient)
	submitter := batcher.NewEspressoTransactionSubmitter(
		batcher.WithEspressoClient(espClient),
		batcher.WithContext(ctx),
	)

	submitter.SpawnWorkers(4, 4)
	submitter.Start()

	i := 0
	for ; i < NUMBER_OF_TRANSACTIONS_TO_SUBMIT; i++ {
		txn := espressoCommon.Transaction{
			Namespace: 1,
			Payload:   make([]byte, 8),
		}
		binary.LittleEndian.PutUint64(txn.Payload, uint64(i))
		err := submitter.SubmitTransaction(&txn)

		if err == nil {
			continue
		}

		// We've triggered the initial condition, and we're no longer
		if _, ok := err.(batcher.ErrTooManyInFlightRequests); ok {
			// able to submit new transactions to the queue, as we've filled
			// the in flight capacity.
			// NOTE: we decrement `i` here, as we didn't successfully submit it.
			i--
			break
		}

		require.NoError(t, err, "unexpected error encountered while submit transaction has been called")
	}

	// Now we trigger the EspressoClient to start working again.
	// This will effectively simulate that the external network has
	// no recovered.

	espClient.SetEspressoClient(succeedingClient)

	// We need to wait a little bit to give the submitter time to process
	// some of its backlog.

	time.Sleep(10 * time.Millisecond)

	for ; i < NUMBER_OF_TRANSACTIONS_TO_SUBMIT; i++ {
		txn := espressoCommon.Transaction{
			Namespace: 1,
			Payload:   make([]byte, 8),
		}
		binary.LittleEndian.PutUint64(txn.Payload, uint64(i))
		err := submitter.SubmitTransaction(&txn)

		if _, ok := err.(batcher.ErrTooManyInFlightRequests); ok {
			// Slow down a bit, and decrement `i` so we try it again.
			i--
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// No further errors should occur
		require.NoError(t, err, "unexpected error encountered while submit transaction has been called")
	}
}
