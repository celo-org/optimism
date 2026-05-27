package batcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	tagged_base64 "github.com/EspressoSystems/espresso-network/sdks/go/tagged-base64"
	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/ethereum-optimism/optimism/espresso"
	"github.com/ethereum-optimism/optimism/espresso/bindings"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

// EspressoOnchainProof is the proof structure returned by the attestation service for onchain verification.
type EspressoOnchainProof struct {
	Proof    json.RawMessage `json:"proof,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
	RawProof struct {
		Journal string `json:"journal"`
	} `json:"raw_proof"`
	OnchainProof string `json:"onchain_proof"`
}

// espressoSubmitTransactionJob is a struct that holds the state required to
// submit a transaction to Espresso.
// It contains the transaction to be submitted itself, and a number to
// track the total number of attempts to submit this transaction to Espresso.
type espressoSubmitTransactionJob struct {
	attempts    int
	transaction *espressoCommon.Transaction
}

// espressoSubmitTransactionJobResponse is a struct that holds the
// response from the Espresso client after submitting a transaction.
// It contains the job that was submitted, the hash of the transaction
// that was submitted (if successful), and any error that occurred during the
// submission (if unsuccessful).
type espressoSubmitTransactionJobResponse struct {
	job  espressoSubmitTransactionJob
	hash *espressoCommon.TaggedBase64
	err  error
}

// espressoTransactionJobAttempt is a struct that holds the job and
// response channel for a transaction submission job.
//
// This is the unit of work that is submitted to the worker to process
// for transaction submissions.
type espressoTransactionJobAttempt struct {
	job  espressoSubmitTransactionJob
	resp chan espressoSubmitTransactionJobResponse
}

// espressoVerifyReceiptJob is a struct that holds the state required to
// verify a receipt for a transaction that was submitted to Espresso.
// It contains the transaction that was submitted, the hash of the
// transaction, and the number of attempts to verify the receipt.
type espressoVerifyReceiptJob struct {
	attempts    int
	startHeight uint64    // HotShot block height when verification began (set on first attempt)
	startTime   time.Time // wall-clock time when verification began (safety backstop)
	transaction espressoSubmitTransactionJob
	hash        *espressoCommon.TaggedBase64
}

// espressoVerifyReceiptJobResponse is a struct that holds the
// response from the Espresso client after verifying a receipt.
// It contains the job that was submitted, and any error that occurred
// during the verification (if unsuccessful).
type espressoVerifyReceiptJobResponse struct {
	job           espressoVerifyReceiptJob
	err           error
	currentHeight uint64 // latest known HotShot block height at time of verification attempt
}

// espressoVerifyReceiptJobAttempt is a struct that holds the job and
// response channel for a receipt verification job.
//
// This is the unit of work that is submitted to the worker to process
// for receipt verifications.
type espressoVerifyReceiptJobAttempt struct {
	job  espressoVerifyReceiptJob
	resp chan espressoVerifyReceiptJobResponse
}

// espressoTransactionSubmitter is a struct that holds the state that governs
// the worker queue processing details for submitting transactions to Espresso
// without spawning arbitrarily many goroutines.
type espressoTransactionSubmitter struct {
	ctx                        context.Context
	wg                         *sync.WaitGroup
	submitJobQueue             chan espressoSubmitTransactionJob
	submitRespQueue            chan espressoSubmitTransactionJobResponse
	submitWorkerQueue          chan chan espressoTransactionJobAttempt
	verifyReceiptJobQueue      chan espressoVerifyReceiptJob
	verifyReceiptRespQueue     chan espressoVerifyReceiptJobResponse
	verifyReceiptWorkerQueue   chan chan espressoVerifyReceiptJobAttempt
	espresso                   espressoClient.EspressoClient
	latestBlockHeight          atomic.Uint64 // shared HotShot block height, updated by trackBlockHeight
	verifyReceiptMaxBlocks     uint64
	verifyReceiptSafetyTimeout time.Duration
	verifyReceiptRetryDelay    time.Duration
	numInFlightJobs            atomic.Int64
	numMaxInFlightJobs         int
}

// EspressoTransactionSubmitterConfig is a configuration struct for the
// EspressoTransactionSubmitter. It contains the configurable details for
// creating the EspressoTransactionSubmitter.
type EspressoTransactionSubmitterConfig struct {
	Ctx                                context.Context
	EspressoClient                     espressoClient.EspressoClient
	Wg                                 *sync.WaitGroup
	SubmitJobQueueCapacity             int
	SubmitResponseQueueCapacity        int
	VerifyReceiptJobQueueCapacity      int
	VerifyReceiptResponseQueueCapacity int
	VerifyReceiptMaxBlocks             uint64
	VerifyReceiptSafetyTimeout         time.Duration
	VerifyReceiptRetryDelay            time.Duration
	MaxInFlightJobs                    int
}

// EspressoTransactionSubmitterOption is a function that can be used to
// configure the EspressoTransactionSubmitterConfig.
type EspressoTransactionSubmitterOption func(*EspressoTransactionSubmitterConfig)

// WithContext is an option that can be used to set the Espresso client
// for the EspressoTransactionSubmitterConfig.
func WithContext(ctx context.Context) EspressoTransactionSubmitterOption {
	return func(config *EspressoTransactionSubmitterConfig) {
		config.Ctx = ctx
	}
}

// WithEspressoClient is an option that can be used to set the Espresso client
// for the EspressoTransactionSubmitterConfig.
func WithEspressoClient(client espressoClient.EspressoClient) EspressoTransactionSubmitterOption {
	return func(config *EspressoTransactionSubmitterConfig) {
		config.EspressoClient = client
	}
}

// WithWaitGroup is an option that can be used to set the wait group
// for the EspressoTransactionSubmitterConfig.
func WithWaitGroup(wg *sync.WaitGroup) EspressoTransactionSubmitterOption {
	return func(config *EspressoTransactionSubmitterConfig) {
		config.Wg = wg
	}
}

// WithVerifyReceiptMaxBlocks sets the number of HotShot blocks to wait for a
// submitted transaction to become queryable before re-submitting.
func WithVerifyReceiptMaxBlocks(n uint64) EspressoTransactionSubmitterOption {
	return func(config *EspressoTransactionSubmitterConfig) {
		config.VerifyReceiptMaxBlocks = n
	}
}

// WithVerifyReceiptSafetyTimeout sets the wall-clock backstop for receipt
// verification. If the block height tracker is stale or broken, re-submission
// is triggered after this duration.
func WithVerifyReceiptSafetyTimeout(d time.Duration) EspressoTransactionSubmitterOption {
	return func(config *EspressoTransactionSubmitterConfig) {
		config.VerifyReceiptSafetyTimeout = d
	}
}

// WithVerifyReceiptRetryDelay sets the delay between receipt verification retries.
func WithVerifyReceiptRetryDelay(d time.Duration) EspressoTransactionSubmitterOption {
	return func(config *EspressoTransactionSubmitterConfig) {
		config.VerifyReceiptRetryDelay = d
	}
}

// WithMaxInFlightJobs sets the maximum number of inflight requests to
// have at once.  Once at capacity all new submission attempts will
// automatically fail.
func WithMaxInFlightJobs(n int) EspressoTransactionSubmitterOption {
	return func(config *EspressoTransactionSubmitterConfig) {
		config.MaxInFlightJobs = n
	}
}

// NewEspressoTransactionSubmitter creates a new EspressoTransactionSubmitter
// throttle with the given context and espresso client.  It will create a new
// transaction submitter with some default options, and apply those options to
// the configuration.
//
// The resulting instance should reflect the given configuration.
// After returning, the caller should call SpawnWorkers to start the workers,
// and Start to start the job scheduling and response handling portions of the
// transaction submitter. After that, the user should be able to submit
// transactions to the submitter via the SubmitTransaction method.
func NewEspressoTransactionSubmitter(options ...EspressoTransactionSubmitterOption) *espressoTransactionSubmitter {
	config := EspressoTransactionSubmitterConfig{
		Ctx:                                context.Background(),
		Wg:                                 new(sync.WaitGroup),
		SubmitJobQueueCapacity:             espresso.DefaultMaxInFlightRequestsToEspresso,
		SubmitResponseQueueCapacity:        10,
		VerifyReceiptJobQueueCapacity:      espresso.DefaultMaxInFlightRequestsToEspresso,
		VerifyReceiptResponseQueueCapacity: 10,
		VerifyReceiptMaxBlocks:             espresso.DefaultVerifyReceiptMaxBlocks,
		VerifyReceiptSafetyTimeout:         espresso.DefaultVerifyReceiptSafetyTimeout,
		VerifyReceiptRetryDelay:            espresso.DefaultVerifyReceiptRetryDelay,
		MaxInFlightJobs:                    espresso.DefaultMaxInFlightRequestsToEspresso,
	}

	for _, option := range options {
		option(&config)
	}

	if config.EspressoClient == nil {
		panic("Espresso client is required")
	}

	return &espressoTransactionSubmitter{
		ctx:                        config.Ctx,
		wg:                         config.Wg,
		submitJobQueue:             make(chan espressoSubmitTransactionJob, config.SubmitJobQueueCapacity),
		submitRespQueue:            make(chan espressoSubmitTransactionJobResponse, config.SubmitResponseQueueCapacity),
		submitWorkerQueue:          make(chan chan espressoTransactionJobAttempt),
		verifyReceiptJobQueue:      make(chan espressoVerifyReceiptJob, config.VerifyReceiptJobQueueCapacity),
		verifyReceiptRespQueue:     make(chan espressoVerifyReceiptJobResponse, config.VerifyReceiptResponseQueueCapacity),
		verifyReceiptWorkerQueue:   make(chan chan espressoVerifyReceiptJobAttempt),
		espresso:                   config.EspressoClient,
		verifyReceiptMaxBlocks:     config.VerifyReceiptMaxBlocks,
		verifyReceiptSafetyTimeout: config.VerifyReceiptSafetyTimeout,
		verifyReceiptRetryDelay:    config.VerifyReceiptRetryDelay,
		numInFlightJobs:            atomic.Int64{},
		numMaxInFlightJobs:         config.MaxInFlightJobs,
	}
}

// ErrTooManyInFlightRequests is an error that is returned when there are
// too many requests in flight at once right now.
type ErrTooManyInFlightRequests struct {
	NumInFlightRequests int
	MaxInFlightRequests int
}

// Error implements error
func (e ErrTooManyInFlightRequests) Error() string {
	return fmt.Sprintf("too many requests in flight to espresso, in flight requests: %d, maximum allowed: %d", e.NumInFlightRequests, e.MaxInFlightRequests)
}

// ErrSubmitToEspressoChannelFull is an error that is returned when the channel
// to submit a transaction to the espresso job channel is full.
//
// This ultimately means that the channel buffer size of the job channel is
// smaller than the max number of in flight requests allowed.
type ErrSubmitToEspressoChannelFull struct {
	Capacity int
	Len      int
}

// Error implements error
func (e ErrSubmitToEspressoChannelFull) Error() string {
	return fmt.Sprintf("submit transaction to espresso job channel is full, len: %d, capacity: %d", e.Len, e.Capacity)
}

// SubmitTransaction will submit a transaction to the Job queue.
//
// NOTE: There is a limit on the maximum number of inflight requests we allow
// at once.  If we're over or at capacity, we'll return an error indicating so.
// If we have capacity available, we'll attempt to submit to the channel, and
// if we're unable to, we'll return an error.  This will **NOT** block.
func (s *espressoTransactionSubmitter) SubmitTransaction(job *espressoCommon.Transaction) error {
	// Check to see if we're over capacity, and if we are, then immediately
	// return an error.
	if numInFlightRequests, numMaxInFlightRequests := s.numInFlightJobs.Load(), int64(s.numMaxInFlightJobs); numInFlightRequests >= numMaxInFlightRequests {
		return ErrTooManyInFlightRequests{
			NumInFlightRequests: int(numInFlightRequests),
			MaxInFlightRequests: int(numMaxInFlightRequests),
		}
	}

	// Construct the job submission
	jobSubmission := espressoSubmitTransactionJob{
		transaction: job,
	}

	select {
	default:
		return ErrSubmitToEspressoChannelFull{
			Len:      len(s.submitJobQueue),
			Capacity: cap(s.submitJobQueue),
		}
	case s.submitJobQueue <- jobSubmission:
		// increment in flight requests
		s.numInFlightJobs.Add(1)
		return nil
	}
}

// Evaluation result for a job.
type JobEvaluation int

const (
	// Continue handling the current job.
	Handle JobEvaluation = iota
	// Retry the submission.
	RetrySubmission
	// Retry the verification.
	RetryVerification
	// Skip the current job and proceed to the next one.
	Skip
)

// Evaluate the submission job.
//
// # Returns
//
// * If there is no error: Handle.
//
// * If there is a permanent issue that won't be fixed by a retry: Skip.
//
// * Otherwise: RetrySubmission.
func evaluateSubmission(jobResp espressoSubmitTransactionJobResponse) JobEvaluation {
	err := jobResp.err

	// If there's no error, continue handling the submission.
	if err == nil {
		return Handle
	}

	if errors.Is(err, espressoClient.ErrPermanent) {
		return Skip
	}

	if !errors.Is(err, espressoClient.ErrEphemeral) {
		// Log the warning for a potentially missed error handling, but still retry it.
		log.Warn("error not explicitly marked as retryable or not", "err", err)
	}

	// Otherwise, retry the submission.
	return RetrySubmission
}

// handleTransactionSubmitJobResponse is a function that is meant to be run in a
// goroutine.
//
// It handles the responses from the submit transaction jobs.  It will
// determine if the transaction was successfully submitted to Espresso, and
// if not, it will retry the transaction.  If the transaction was successfully
// submitted, it will then submit a job to the verify receipt job queue to
// verify the receipt of the transaction.
func (s *espressoTransactionSubmitter) handleTransactionSubmitJobResponse() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		var jobResp espressoSubmitTransactionJobResponse
		var ok bool

		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			log.Debug("Espresso transaction submitter queue status",
				"submitJobQueue", len(s.submitJobQueue),
				"submitRespQueue", len(s.submitRespQueue),
				"verifyReceiptJobQueue", len(s.verifyReceiptJobQueue),
				"verifyReceiptRespQueue", len(s.verifyReceiptRespQueue))
			continue
		case jobResp, ok = <-s.submitRespQueue:
			if !ok {
				// Our channel is closed, and we are done
				return
			}
		}

		switch evaluation := evaluateSubmission(jobResp); evaluation {
		case Skip:
			s.numInFlightJobs.Add(-1)
			continue
		case RetrySubmission:
			s.submitJobQueue <- jobResp.job
			continue
		}

		verifyJob := espressoVerifyReceiptJob{
			startTime:   time.Now(),
			transaction: jobResp.job,
			hash:        jobResp.hash,
		}

		select {
		case <-s.ctx.Done():
			return
		// Move to verifying the receipt
		case s.verifyReceiptJobQueue <- verifyJob:
		}
	}
}

// Default values for receipt verification tuning are defined as exported
// constants in the espresso package (espresso.DefaultVerifyReceipt*) so that
// the CLI flag defaults and this batcher logic share a single source of truth.

// evaluateVerification evaluates the verification job response.
//
// # Returns
//
// * If there is no error: Handle.
//
// * If there is a permanent issue that won't be fixed by a retry: Skip.
//
// * If enough HotShot blocks have passed since verification started: RetrySubmission.
//
// * If the wall-clock safety timeout is exceeded: RetrySubmission.
//
// * Otherwise: RetryVerification.
func (s *espressoTransactionSubmitter) evaluateVerification(jobResp espressoVerifyReceiptJobResponse) JobEvaluation {
	err := jobResp.err

	// If there's no error, continue handling the verification.
	if err == nil {
		return Handle
	}

	if errors.Is(err, espressoClient.ErrPermanent) {
		return Skip
	}

	if !errors.Is(err, espressoClient.ErrEphemeral) {
		// Log the warning for a potentially missed error handling, but still retry it.
		log.Warn("error not explicitly marked as retryable or not", "err", err)
	}

	// Block-count-based timeout: re-submit if enough HotShot blocks have
	// passed since verification started. The startHeight guard handles the
	// edge case where the height tracker hasn't fetched its first value yet.
	if jobResp.job.startHeight > 0 && jobResp.currentHeight >= jobResp.job.startHeight+s.verifyReceiptMaxBlocks {
		log.Info("Verification timed out by block count, re-submitting",
			"startHeight", jobResp.job.startHeight,
			"currentHeight", jobResp.currentHeight,
			"maxBlocks", s.verifyReceiptMaxBlocks)
		return RetrySubmission
	}

	// Wall-clock safety backstop in case the block height tracker is stale
	// or broken (e.g., query service returning old data).
	if elapsed := time.Since(jobResp.job.startTime); elapsed > s.verifyReceiptSafetyTimeout {
		log.Warn("Verification timed out by safety timeout, re-submitting",
			"elapsed", elapsed,
			"safetyTimeout", s.verifyReceiptSafetyTimeout)
		return RetrySubmission
	}

	// Otherwise, retry the verification.
	return RetryVerification
}

// handleVerifyReceiptJobResponse is a function that is meant to be run in a
// goroutine.
//
// This function handles responses from the verify receipt job queue.  It will
// check the results for any errors, and if there are any errors that are
// applicable to retry, it will requeue the job for another attempt.
// If the the job is successful, no further processing is needed and it is
// considered complete.
// If the job has taken too long to verify, then it will re-submit the job
// back to the submit transaction queue for another attempt.
//
// NOTE: This function currently will loop forever if the transaction is
// never going to be available.
func (s *espressoTransactionSubmitter) handleVerifyReceiptJobResponse() {
	for {
		var jobResp espressoVerifyReceiptJobResponse
		var ok bool

		select {
		case <-s.ctx.Done():
			return
		case jobResp, ok = <-s.verifyReceiptRespQueue:
			if !ok {
				// Our channel is closed, and we are done
				return
			}
		}

		switch evaluation := s.evaluateVerification(jobResp); evaluation {
		case Skip:
			// decrement in flight jobs on skip, since we're done with this job
			s.numInFlightJobs.Add(-1)
			continue
		case RetrySubmission:
			s.submitJobQueue <- jobResp.job.transaction
			continue
		case RetryVerification:
			s.verifyReceiptJobQueue <- jobResp.job
			continue
		}

		s.numInFlightJobs.Add(-1)
		// We're done with this job and transaction, we have successfully
		// confirmed that the transaction was submitted to Espresso
		commitment := jobResp.job.transaction.transaction.Commit()
		hash, _ := tagged_base64.New("TX", commitment[:])
		log.Info("Transaction confirmed on Espresso", "hash", hash.String())
	}
}

// scheduleSubmitTransactionJobs is a function that is meant to be run in a
// goroutine.
//
// It handles the scheduling of submit transaction jobs so that the submit
// transaction workers can process them.
func (s *espressoTransactionSubmitter) scheduleSubmitTransactionJobs() {
	for {
		var ok bool

		// Get a worker from the worker queue
		var worker chan espressoTransactionJobAttempt
		select {
		case <-s.ctx.Done():
			return

		case worker, ok = <-s.submitWorkerQueue:
			if !ok {
				// Our channel is closed, and we are done
				return
			}
		}

		// Get a job from the job queue
		var job espressoSubmitTransactionJob
		select {
		case <-s.ctx.Done():
			return
		case job, ok = <-s.submitJobQueue:
			if !ok {
				// Our channel is closed, and we are done
				return
			}
		}

		// Submit the job to the worker
		select {
		case <-s.ctx.Done():
			return

		case worker <- espressoTransactionJobAttempt{job: job, resp: s.submitRespQueue}:
		}
	}
}

// scheduleVerifyReceiptJobs is a function that is meant to be run in a
// goroutine.
//
// It handles the scheduling of verify receipt jobs so that the verify receipt
// workers can process them.
func (s *espressoTransactionSubmitter) scheduleVerifyReceiptsJobs() {
	for {
		var ok bool

		// Get a worker from the worker queue
		var worker chan espressoVerifyReceiptJobAttempt
		select {
		case <-s.ctx.Done():
			return

		case worker, ok = <-s.verifyReceiptWorkerQueue:
			if !ok {
				// Our channel is closed, and we are done
				return
			}
		}

		// Get a job from the job queue
		var job espressoVerifyReceiptJob
		select {
		case <-s.ctx.Done():
			return
		case job, ok = <-s.verifyReceiptJobQueue:
			if !ok {
				// Our channel is closed, and we are done
				return
			}
		}

		// Submit the job to the worker
		select {
		case <-s.ctx.Done():
			return

		case worker <- espressoVerifyReceiptJobAttempt{job: job, resp: s.verifyReceiptRespQueue}:
		}
	}
}

// espressoSubmitTransactionWorker is a function that is meant to be run as a
// goroutine.  It will create a channel for it's job queue, and submit those to
// the worker queue in order to wait for work.  It will then take that job and
// attempt to submit the transaction contained within to espresso using the
// given espresso client. It will submit the response back to the channel
// contained within the job attempt it received.
//
// It's lifetime is governed by the context passed to it, and it will stop
// processing when that context is cancelled.
//
// NOTE: If the context is cancelled after a job has been received, but before
// it is able to submit the transaction, or report about it's result, the job
// may be lost.
func espressoSubmitTransactionWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	cli espressoClient.EspressoClient,
	workerQueue chan<- chan espressoTransactionJobAttempt,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer wg.Done()
	ch := make(chan espressoTransactionJobAttempt)
	defer close(ch)

	for {
		var ok bool
		select {
		case <-ctx.Done():
			return

			// Queue our job queue, asking for work
		case workerQueue <- ch:
		}

		// Wait for a job to run
		var jobAttempt espressoTransactionJobAttempt
		select {
		case <-ctx.Done():
			return
		case jobAttempt, ok = <-ch:
			if !ok {
				// Our channel is closed, and we are done
				return
			}
		}

		// Submit the transaction to Espresso
		hash, err := cli.SubmitTransaction(ctx, *jobAttempt.job.transaction)
		if err == nil {
			log.Info("Submitted transaction to Espresso", "hash", hash)
		}

		jobAttempt.job.attempts++
		resp := espressoSubmitTransactionJobResponse{
			job:  jobAttempt.job,
			hash: hash,
			err:  err,
		}

		select {
		case <-ctx.Done():
			return

		// Send the response back via the channel in the job attempt struct
		case jobAttempt.resp <- resp:
		}
	}
}

// espressoVerifyTransactionWorker is a function that is meant to be run as a
// goroutine.  It will create a channel for it's job queue, and submit those to
// the worker queue in order to wait for work.  It will then take that job and
// attempt to verify the transaction contained within to espresso using the
// given espresso client. It will submit the response back to the channel
// contained within the job attempt it received.
func espressoVerifyTransactionWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	cli espressoClient.EspressoClient,
	workerQueue chan<- chan espressoVerifyReceiptJobAttempt,
	latestHeight *atomic.Uint64,
	retryDelay time.Duration,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer wg.Done()
	ch := make(chan espressoVerifyReceiptJobAttempt)
	defer close(ch)

	for {
		var ok bool
		select {
		case <-ctx.Done():
			return

			// Queue our job queue, asking for work
		case workerQueue <- ch:
		}

		// Wait for a job to run
		var jobAttempt espressoVerifyReceiptJobAttempt
		select {
		case <-ctx.Done():
			return
		case jobAttempt, ok = <-ch:
			if !ok {
				// Our channel is closed, and we are done
				return
			}
		}

		// On the first attempt, snapshot the current block height so we
		// can measure how many blocks pass during verification.
		if jobAttempt.job.attempts == 0 {
			jobAttempt.job.startHeight = latestHeight.Load()
		}

		if jobAttempt.job.attempts > 0 {
			// We have already attempted this job, so we will wait a bit
			// NOTE: this prevents this worker from being able to process
			// other jobs while we wait for this delay.
			time.Sleep(retryDelay)
		}

		_, err := cli.FetchTransactionByHash(ctx, jobAttempt.job.hash)

		jobAttempt.job.attempts++
		resp := espressoVerifyReceiptJobResponse{
			job:           jobAttempt.job,
			err:           err,
			currentHeight: latestHeight.Load(),
		}

		select {
		case <-ctx.Done():
			return

		case jobAttempt.resp <- resp:
		}
	}
}

// SpawnWorkers spawns the given number of workers to process the
// submit transaction jobs and verify receipt jobs.
func (s *espressoTransactionSubmitter) SpawnWorkers(numSubmitTransactionWorkers, numVerifyReceiptWorkers int) {
	workersCtx := s.ctx

	for i := 0; i < numSubmitTransactionWorkers; i++ {
		s.wg.Add(1)
		go espressoSubmitTransactionWorker(workersCtx, s.wg, s.espresso, s.submitWorkerQueue)
	}

	for i := 0; i < numVerifyReceiptWorkers; i++ {
		s.wg.Add(1)
		go espressoVerifyTransactionWorker(workersCtx, s.wg, s.espresso, s.verifyReceiptWorkerQueue, &s.latestBlockHeight, s.verifyReceiptRetryDelay)
	}
}

// trackBlockHeight periodically polls FetchLatestBlockHeight and stores
// the result in s.latestBlockHeight for verify jobs to compare against.
// This avoids redundant height queries from individual verify workers.
func (s *espressoTransactionSubmitter) trackBlockHeight() {
	for {
		height, err := s.espresso.FetchLatestBlockHeight(s.ctx)
		if err == nil {
			s.latestBlockHeight.Store(height)
		} else if s.ctx.Err() == nil {
			log.Debug("failed to fetch latest block height for verification tracking", "err", err)
		}

		// Wait for the next interval or until context is done.
		select {
		case <-time.After(s.verifyReceiptRetryDelay):
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *espressoTransactionSubmitter) Start() {
	// Block height tracker for verify receipt timeout
	go s.trackBlockHeight()

	// Submit Transaction Jobs
	go s.scheduleSubmitTransactionJobs()
	go s.handleTransactionSubmitJobResponse()

	// Verify Receipt Jobs
	go s.scheduleVerifyReceiptsJobs()
	go s.handleVerifyReceiptJobResponse()
}

// Converts a block to an EspressoBatch and starts a goroutine that publishes it to Espresso
// Returns error only if batch conversion fails, otherwise it is infallible, as the goroutine
// will retry publishing until successful.
func (l *BatchSubmitter) queueBlockToEspresso(ctx context.Context, block *types.Block) error {
	espressoBatch, err := derive.BlockToEspressoBatch(l.RollupConfig, block)
	if err != nil {
		l.Log.Warn("Failed to derive batch from block", "err", err)
		return fmt.Errorf("failed to derive batch from block: %w", err)
	}

	transaction, err := espressoBatch.ToEspressoTransaction(ctx, l.RollupConfig.L2ChainID.Uint64(), l.Espresso.ChainSigner)
	if err != nil {
		l.Log.Warn("Failed to create Espresso transaction from a batch", "err", err)
		return fmt.Errorf("failed to create Espresso transaction from a batch: %w", err)
	}

	commitment := transaction.Commit()
	hash, _ := tagged_base64.New("TX", commitment[:])
	l.Log.Info("Created Espresso transaction from batch", "hash", hash, "batchNr", espressoBatch.BatchHeader.Number.Uint64())

	if err := l.espressoSubmitter.SubmitTransaction(transaction); err != nil {
		return fmt.Errorf("failed to submit job to espresso: %w", err)
	}

	return nil
}

func (l *BatchSubmitter) espressoSyncAndRefresh(ctx context.Context, newSyncStatus *eth.SyncStatus) {
	err := l.EspressoStreamer().Refresh(ctx, newSyncStatus.FinalizedL1, newSyncStatus.SafeL2.Number, newSyncStatus.FinalizedL2.L1Origin)
	if err != nil {
		l.degradedLog.Warn(l.Log, "espressoStreamerRefreshErr", "Failed to refresh Espresso streamer", "err", err)
	} else {
		l.degradedLog.Clear(l.Log, "espressoStreamerRefreshErr", "Espresso streamer refresh recovered")
	}

	l.channelMgrMutex.Lock()
	defer l.channelMgrMutex.Unlock()
	syncActions, outOfSync := computeSyncActions(*newSyncStatus, l.prevCurrentL1, l.channelMgr.blocks, l.channelMgr.channelQueue, l.Log)
	if outOfSync {
		l.degradedLog.Warn(l.Log, "sequencerOutOfSync", "Sequencer is out of sync, retrying next tick.")
		return
	}
	l.degradedLog.Clear(l.Log, "sequencerOutOfSync", "Sequencer back in sync")
	l.prevCurrentL1 = newSyncStatus.CurrentL1
	if syncActions.clearState != nil {
		l.channelMgr.Clear(*syncActions.clearState)
		l.EspressoStreamer().Reset()
	} else {
		l.channelMgr.PruneSafeBlocks(syncActions.blocksToPrune)
		l.channelMgr.PruneChannels(syncActions.channelsToPrune)
	}
}

// peekNextBatch returns the next batch from the streamer, performing a fork check
// against an expected parent hash.
//
// The expected parent is tip when tip is set. When tip is zero (channel manager was
// just cleared), we fall back to safeL2.Hash if the batch is at exactly safeL2+1 —
// the one position where we can set tip to the known safe head. Otherwise we accept the batch as-is.
func (l *BatchSubmitter) peekNextBatch(ctx context.Context, syncStatus *eth.SyncStatus) *derive.EspressoBatch {
	l.channelMgrMutex.Lock()
	tip := l.channelMgr.tip
	l.channelMgrMutex.Unlock()

	batch := l.EspressoStreamer().Peek(ctx)
	if batch == nil {
		return nil
	}

	// Check if we can set the tip if not set
	if tip == (common.Hash{}) && (*batch).Number() == syncStatus.SafeL2.Number+1 {
		l.Log.Info(
			"setting tip to safe l2 hash",
			"batchNr", (*batch).Number(),
			"batchParent", (*batch).Header().ParentHash.Hex(),
			"tip", tip,
		)
		tip = syncStatus.SafeL2.Hash
	}

	if tip == (common.Hash{}) {
		l.Log.Warn(
			"tip is not set, taking available batch",
			"blockParentHash", (*batch).Header().ParentHash.Hex(),
			"blockHash", (*batch).Header().Hash().Hex(),
		)
		return batch
	}

	if (*batch).Header().ParentHash != tip {
		l.Log.Warn(
			"head batch fork mismatch, seeking to proper head",
			"batchNr", (*batch).Number(),
			"batchParent", (*batch).Header().ParentHash,
			"tip", tip,
		)
		l.EspressoStreamer().SetProperHead(tip)
		return nil
	}

	return batch
}

// Periodically refreshes the sync status and polls Espresso streamer for new batches
func (l *BatchSubmitter) espressoBatchLoadingLoop(ctx context.Context, wg *sync.WaitGroup, publishSignal chan pubInfo) {
	l.Log.Info("Starting EspressoBatchLoadingLoop", "polling interval", l.Config.Espresso.PollInterval)

	defer wg.Done()
	ticker := time.NewTicker(l.Config.Espresso.PollInterval)
	defer ticker.Stop()
	defer close(publishSignal)

	for {
		select {
		case <-ticker.C:
			newSyncStatus, err := l.getSyncStatus(ctx)
			if err != nil {
				l.degradedLog.Warn(l.Log, "syncStatusErr/espressoBatchLoading", "failed to refresh sync status", "err", err)
				continue
			}
			l.degradedLog.Clear(l.Log, "syncStatusErr/espressoBatchLoading", "sync status fetch recovered")

			l.espressoSyncAndRefresh(ctx, newSyncStatus)

			err = l.EspressoStreamer().Update(ctx)

			var batch *derive.EspressoBatch

			for {

				batch = l.peekNextBatch(ctx, newSyncStatus)

				if batch == nil {
					break
				}

				// This should happen ONLY if the batch is malformed. ToBlock has to guarantee no
				// transient errors.
				block, err := batch.ToBlock(l.RollupConfig)
				if err != nil {
					l.Log.Error("failed to convert singular batch to block", "err", err)
					l.EspressoStreamer().Next(ctx)
					continue
				}

				l.Log.Info(
					"Received block from Espresso",
					"blockNr", block.NumberU64(),
					"blockHash", block.Hash(),
					"parentHash", block.ParentHash(),
				)

				l.channelMgrMutex.Lock()
				err = l.channelMgr.AddL2Block(block)
				l.channelMgrMutex.Unlock()

				if err != nil {
					l.Log.Error("failed to add L2 block to channel manager", "err", err)
					l.clearState(ctx)
					l.EspressoStreamer().Reset()
					break
				}

				l.EspressoStreamer().Next(ctx)
				l.Log.Info("Added L2 block to channel manager", "blockNr", block.NumberU64())
			}

			l.tryPublishSignal(publishSignal, pubInfo{})

			// A failure in the streamer Update can happen after the buffer has been partially filled
			if err != nil {
				l.degradedLog.Warn(l.Log, "espressoStreamerUpdateErr", "failed to update Espresso streamer", "err", err)
				continue
			}
			l.degradedLog.Clear(l.Log, "espressoStreamerUpdateErr", "Espresso streamer update recovered")

		case <-ctx.Done():
			l.Log.Info("espressoBatchLoadingLoop returning")
			return
		}
	}
}

type BlockLoader struct {
	queuedBlocks   []eth.L2BlockRef
	prevSyncStatus *eth.SyncStatus
	batcher        *BatchSubmitter
}

func (l *BlockLoader) reset(ctx context.Context) {
	l.prevSyncStatus = nil
	l.queuedBlocks = nil
	l.batcher.clearState(ctx)
}

func (l *BlockLoader) EnqueueBlocks(ctx context.Context, blocksToQueue inclusiveBlockRange) {
	l.batcher.Log.Debug("Loading and queueing blocks", "range", blocksToQueue)
	for i := blocksToQueue.start; i <= blocksToQueue.end; i++ {
		block, err := l.batcher.fetchBlock(ctx, i)
		if err != nil {
			l.batcher.degradedLog.Warn(l.batcher.Log, "fetchBlockErr", "Failed to fetch block", "err", err)
			break
		}
		l.batcher.degradedLog.Clear(l.batcher.Log, "fetchBlockErr", "Block fetching recovered")

		for _, txn := range block.Transactions() {
			l.batcher.Log.Debug("tx hash before submitting to Espresso", "hash", txn.Hash().String())
		}

		if len(l.queuedBlocks) > 0 && block.ParentHash() != l.queuedBlocks[len(l.queuedBlocks)-1].Hash {
			l.batcher.Log.Warn("Found L2 reorg", "block_number", i)
			l.reset(ctx)
			break
		}

		blockRef, err := derive.L2BlockToBlockRef(l.batcher.RollupConfig, block)
		if err != nil {
			// NOTE: if we fail to convert an L2Block to a BlockRef, it's
			// unlikely that breaking here, and waiting for resubmission would
			// actually ever result in it succeeding.
			// For now, we add a log, but this may be a Fatal unrecoverable
			// error if it ever occurs.
			l.batcher.Log.Warn("failed to convert block to block reference", "err", err)
			break
		}

		err = l.batcher.queueBlockToEspresso(ctx, block)
		if err != nil {
			l.batcher.Log.Debug("queue block to espresso failed", "err", err)
			break
		}

		l.queuedBlocks = append(l.queuedBlocks, blockRef)
	}
}

type EnqueueBlockAction uint

const (
	ActionEnqueue = iota
	ActionRetry
	ActionReset
)

// This function is an analogue of `computeSyncActions` for Espresso batcher mode
//
// It computes the next block range to enqueue to Espresso based on new newSyncStatus and
// does a number of checks to ensure consistency of the chain.
//
// If reorg is detected, empty range and ActionReset is returned.
// If there isn't enough information or no blocks to load yet, empty range and ActionRetry is returned.
func (l *BlockLoader) nextBlockRange(newSyncStatus *eth.SyncStatus) (inclusiveBlockRange, EnqueueBlockAction) {
	if newSyncStatus.HeadL1 == (eth.L1BlockRef{}) {
		// empty sync status
		return inclusiveBlockRange{}, ActionRetry
	}

	if l.prevSyncStatus == nil {
		l.prevSyncStatus = newSyncStatus
	}

	if newSyncStatus.CurrentL1.Number < l.prevSyncStatus.CurrentL1.Number {
		// sequencer restarted and hasn't caught up yet
		l.batcher.degradedLog.Warn(l.batcher.Log, "sequencerCurrentL1Reversed", "sequencer currentL1 reversed", "new currentL1", newSyncStatus.CurrentL1.Number, "previous currentL1", l.prevSyncStatus.CurrentL1.Number)
		return inclusiveBlockRange{}, ActionRetry
	}
	l.batcher.degradedLog.Clear(l.batcher.Log, "sequencerCurrentL1Reversed", "sequencer currentL1 caught up")

	safeL2 := newSyncStatus.SafeL2

	// State empty, just enqueue all unsafe blocks
	if len(l.queuedBlocks) == 0 {
		return inclusiveBlockRange{safeL2.Number + 1, newSyncStatus.UnsafeL2.Number}, ActionEnqueue
	}

	lastQueuedBlock := l.queuedBlocks[len(l.queuedBlocks)-1]
	firstQueuedBlock := l.queuedBlocks[0]
	nextSafeBlockNum := safeL2.Number + 1

	if lastQueuedBlock.Number >= newSyncStatus.UnsafeL2.Number {
		// nothing to enqueue, unsafe block number is not higher than safe
		return inclusiveBlockRange{}, ActionRetry
	}

	if lastQueuedBlock.Number < safeL2.Number {
		// derivation pipeline is somehow ahead of us, reset
		return inclusiveBlockRange{}, ActionReset
	}

	if nextSafeBlockNum < firstQueuedBlock.Number {
		l.batcher.Log.Warn("next safe block is below oldest block in state")
		return inclusiveBlockRange{}, ActionReset
	}

	numBlocksToEnqueue := nextSafeBlockNum - firstQueuedBlock.Number

	if numBlocksToEnqueue > uint64(len(l.queuedBlocks)) {
		l.batcher.Log.Warn("safe head above newest block in state, resetting loader")
		return inclusiveBlockRange{}, ActionReset
	}

	if numBlocksToEnqueue > 0 && l.queuedBlocks[numBlocksToEnqueue-1].Hash != safeL2.Hash {
		l.batcher.Log.Warn("safe chain reorg, resetting loader")
		return inclusiveBlockRange{}, ActionReset
	}

	if safeL2.Number > firstQueuedBlock.Number {
		numFinalizedBlocksInQueue := safeL2.Number - firstQueuedBlock.Number
		l.batcher.Log.Warn(
			"Removing finalized blocks from queued",
			"numFinalizedBlocksInQueue", numFinalizedBlocksInQueue,
			"safeL2", safeL2,
			"firstQueuedBlock", firstQueuedBlock)
		l.queuedBlocks = l.queuedBlocks[numFinalizedBlocksInQueue:]
	}

	return inclusiveBlockRange{lastQueuedBlock.Number + 1, newSyncStatus.UnsafeL2.Number}, ActionEnqueue
}

// numBlocks is a convenience method for inclusiveBlockRange that can be
// utilized to quickly determine how many blocks are being referenced within
// the range.
func (i inclusiveBlockRange) numBlocks() uint64 {
	return 1 + i.end - i.start
}

// LARGE_BLOCK_GAP_THRESHOLD is a threshold for the number of blocks that we
// consider to be a "large gap" when queueing blocks to Espresso.  We're
// interested in being alerted when we're falling behind.
const LARGE_BLOCK_GAP_THRESHOLD = 30 * 60

// blockLoadingLoop
// -  polls the sequencer,
// -  queues unsafe blocks from the sequencer to Espresso
func (l *BatchSubmitter) espressoBatchQueueingLoop(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(l.Config.PollInterval)
	defer ticker.Stop()
	defer wg.Done()

	loader := BlockLoader{
		batcher: l,
	}

	// *
	// * BEFORE we start:
	// * - scan batchInbox from batchInbox.lastBackfilled
	// * - enqueue all batches from batchInbox that are _by fallback batcher_ to Espresso
	// * - wait for espresso queue to clear
	// * - set lastBackfilled to block height of the last of such batches
	// *

	for {
		select {
		case <-ticker.C:
			newSyncStatus, err := l.getSyncStatus(ctx)
			if err != nil {
				l.degradedLog.Warn(l.Log, "syncStatusErr/espressoBatchQueueing", "Couldn't get sync status", "error", err)
				continue
			}
			l.degradedLog.Clear(l.Log, "syncStatusErr/espressoBatchQueueing", "sync status fetch recovered")

			blocksToQueue, action := loader.nextBlockRange(newSyncStatus)

			// We add a check here to add visibility to us that we've exceeded
			// a threshold.
			if numBlocks := blocksToQueue.numBlocks(); numBlocks >= LARGE_BLOCK_GAP_THRESHOLD {
				l.Log.Warn("Large gap of blocks to enqueue to Espresso detected", "numBlocks", numBlocks, "blocksToQueue", blocksToQueue)
			}

			if action == ActionEnqueue {
				numEnqueuedBlocksBefore := len(loader.queuedBlocks)
				loader.EnqueueBlocks(ctx, blocksToQueue)
				numEnqueuedBlocksAfter := len(loader.queuedBlocks)

				// This is a check to help us determine whether we're able to
				// push through all of the blocks we've attempted to or not.
				if (numEnqueuedBlocksAfter - numEnqueuedBlocksBefore) < int(blocksToQueue.numBlocks()) {
					// If we're in this conditional, it means we weren't able
					// submit all of the blocks to Espresso that we were
					// attempting to.
					//
					// TODO: We should probably throttle a bit.
				}
			} else if action == ActionReset {
				loader.reset(ctx)
			}

		case <-ctx.Done():
			l.Log.Info("blockLoadingLoop returning")
			return
		}
	}
}

func (l *BatchSubmitter) fetchBlock(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	l2Client, err := l.EndpointProvider.EthClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting L2 client: %w", err)
	}

	cCtx, cancel := context.WithTimeout(ctx, l.Config.NetworkTimeout)
	defer cancel()

	block, err := l2Client.BlockByNumber(cCtx, new(big.Int).SetUint64(blockNumber))
	if err != nil {
		return nil, fmt.Errorf("getting L2 block: %w", err)
	}

	return block, nil
}

// resolveTEEVerifierAddress queries the BatchAuthenticator contract to get the
// EspressoTEEVerifier address.
func (l *BatchSubmitter) resolveTEEVerifierAddress() error {
	if l.RollupConfig.BatchAuthenticatorAddress == (common.Address{}) {
		// If batcher authenticator address is nil, we will keep teeVerifierAddress to nil as well
		return nil
	}
	auth, err := bindings.NewBatchAuthenticatorCaller(l.RollupConfig.BatchAuthenticatorAddress, l.L1Client)
	if err != nil {
		return fmt.Errorf("failed to create BatchAuthenticator caller: %w", err)
	}
	addr, err := auth.EspressoTEEVerifier(nil)
	if err != nil {
		return fmt.Errorf("failed to query EspressoTEEVerifier address: %w", err)
	}
	l.teeVerifierAddress = addr
	l.Log.Info("Resolved TEE verifier address", "address", addr.Hex())
	return nil
}

func (l *BatchSubmitter) registerBatcher(ctx context.Context) error {
	if len(l.Espresso.Attestation) == 0 {
		l.Log.Warn("Attestation is empty, skipping registration")
		return nil
	}

	if l.Config.Espresso.AttestationService == "" {
		l.Log.Warn("EspressoAttestationServices is not set, skipping registration")
		return nil
	}

	l.Log.Info("Batch authenticator address", "value", l.RollupConfig.BatchAuthenticatorAddress)
	code, err := l.L1Client.CodeAt(ctx, l.RollupConfig.BatchAuthenticatorAddress, nil)
	if err != nil {
		return fmt.Errorf("failed to check code at contract address: %w", err)
	}
	if len(code) == 0 {
		return fmt.Errorf("no contract deployed at this address %w", err)
	}

	abi, err := bindings.BatchAuthenticatorMetaData.GetAbi()
	if err != nil {
		return fmt.Errorf("failed to get Batch Authenticator ABI: %w", err)
	}

	onchainProof, err := l.GenerateZKProof(ctx, l.Espresso.Attestation)
	if err != nil {
		l.Log.Error("failed to generate zk proof from nitro attestation", "err", err)
		return fmt.Errorf("failed to generate zk proof from nitro attestation: %w", err)
	}

	journalBytes, err := hex.DecodeString(stripHexPrefix(onchainProof.RawProof.Journal))
	if err != nil {
		l.Log.Error("failed to decode journal hex string", "err", err)
		return fmt.Errorf("failed to decode journal hex string: %w", err)
	}
	onchainProofBytes, err := hex.DecodeString(stripHexPrefix(onchainProof.OnchainProof))
	if err != nil {
		l.Log.Error("failed to decode onchain proof hex string", "err", err)
		return fmt.Errorf("failed to decode onchain proof hex string: %w", err)
	}
	log.Info("successfully generated zk proof from nitro attestation")

	txData, err := abi.Pack("registerSigner", journalBytes, onchainProofBytes)
	if err != nil {
		return fmt.Errorf("failed to create registerSigner transaction: %w", err)
	}

	candidate := txmgr.TxCandidate{
		TxData: txData,
		To:     &l.RollupConfig.BatchAuthenticatorAddress,
	}

	l.Log.Info("Registering batcher with the BatchAuthenticator contract")
	_, err = l.Txmgr.Send(ctx, candidate)
	if err != nil {
		return fmt.Errorf("failed to send registerBatcher transaction: %w", err)
	}

	l.Log.Info("Registered batcher with the BatchAuthenticator contract")

	return nil
}

func (l *BatchSubmitter) GenerateZKProof(ctx context.Context, attestationBytes []byte) (*EspressoOnchainProof, error) {
	attestationServiceURL := strings.TrimSuffix(l.Config.Espresso.AttestationService, "/")
	url := attestationServiceURL + "/generate_proof"
	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(attestationBytes))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/octet-stream")
	client := http.Client{
		Timeout: 5 * time.Minute,
	}
	res, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer (func() {
		_ = res.Body.Close()
	})()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", res.StatusCode)
	}

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var zkProof EspressoOnchainProof
	err = json.Unmarshal(responseData, &zkProof)
	if err != nil {
		return nil, err
	}

	return &zkProof, nil
}

// sendTxWithEspresso uses the txmgr queue to send the given transaction candidate after setting
// its gaslimit. It will block if the txmgr queue has reached its MaxPendingTransactions limit.
func (l *BatchSubmitter) sendTxWithEspresso(txdata txData, isCancel bool, candidate *txmgr.TxCandidate, queue TxSender[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) {
	transactionReference := txRef{id: txdata.ID(), isCancel: isCancel, isBlob: txdata.daType == DaTypeBlob}
	l.Log.Debug("Sending Espresso-enabled L1 transaction", "txRef", transactionReference)

	commitment, err := computeCommitment(candidate)
	if err != nil {
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: err,
		}
		return
	}
	l.Log.Debug("Computed batch commitment", "txRef", transactionReference, "commitment", hexutil.Encode(commitment[:]))

	signature, err := l.signEIP712Commitment(commitment)
	if err != nil {
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to sign transaction: %w", err),
		}
		return
	}

	l.Log.Debug("Signed transaction", "txRef", transactionReference, "commitment", hexutil.Encode(commitment[:]), "sig", hexutil.Encode(signature))

	batchAuthenticatorAbi, err := bindings.BatchAuthenticatorMetaData.GetAbi()
	if err != nil {
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to get batch authenticator ABI: %w", err),
		}
		return
	}

	authenticateBatchCalldata, err := batchAuthenticatorAbi.Pack("authenticateBatchInfo", commitment, signature)
	if err != nil {
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to pack authenticateBatch calldata: %w", err),
		}
		return
	}

	verifyCandidate := txmgr.TxCandidate{
		TxData: authenticateBatchCalldata,
		To:     &l.RollupConfig.BatchAuthenticatorAddress,
	}

	l.Log.Debug(
		"Sending authenticateBatch transaction",
		"txRef", transactionReference,
		"commitment", hexutil.Encode(commitment[:]),
		"sig", hexutil.Encode(signature),
		"address", l.RollupConfig.BatchAuthenticatorAddress.String(),
	)
	verificationReceipt, err := l.Txmgr.Send(l.killCtx, verifyCandidate)
	if err != nil {
		l.Log.Error("Failed to send authenticateBatch transaction", "txRef", transactionReference, "err", err)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to send authenticateBatch transaction: %w", err),
		}
		return
	}

	receipt, err := l.Txmgr.Send(l.killCtx, *candidate)
	if err != nil {
		l.Log.Error("Failed to send batch inbox transaction", "txRef", transactionReference, "err", err)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to send batch inbox transaction: %w", err),
		}
		return
	}

	distance := new(big.Int).Sub(receipt.BlockNumber, verificationReceipt.BlockNumber)
	lookbackWindow := new(big.Int).SetUint64(derive.BatchAuthLookbackWindow)
	if distance.Sign() < 0 || distance.Cmp(lookbackWindow) >= 0 {
		l.Log.Error("authenticateBatch transaction too far from batch inbox transaction", "txRef", transactionReference, "distance", distance)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("authenticateBatch transaction too far from batch inbox transaction: %s", distance),
		}
		return
	}

	receiptsCh <- txmgr.TxReceipt[txRef]{
		ID:      transactionReference,
		Receipt: receipt,
		Err:     nil,
	}
}

// signEIP712Commitment creates an EIP-712 signature for the given commitment using the batcher's private key.
func (l *BatchSubmitter) signEIP712Commitment(commitment [32]byte) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"EspressoTEEVerifier": []apitypes.Type{
				{Name: "commitment", Type: "bytes32"},
			},
		},
		PrimaryType: "EspressoTEEVerifier",
		Domain: apitypes.TypedDataDomain{
			Name:              "EspressoTEEVerifier",
			Version:           "1",
			ChainId:           (*math.HexOrDecimal256)(l.RollupConfig.L1ChainID),
			VerifyingContract: l.teeVerifierAddress.String(),
		},
		Message: map[string]interface{}{
			"commitment": commitment,
		},
	}
	// Calculate the hash using go-ethereum's EIP-712 implementation
	hash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate EIP-712 hash: %w", err)
	}

	signature, err := crypto.Sign(hash, l.Config.Espresso.BatcherPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign EIP-712 hash: %w", err)
	}

	// Normalize the recovery ID (v) from 0/1 to 27/28 for Solidity's ECDSA.recover
	// See: https://github.com/ethereum/go-ethereum/issues/19751#issuecomment-504900739
	if signature[64] < 27 {
		signature[64] += 27
	}
	return signature, nil
}

func stripHexPrefix(hexStr string) string {
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		return hexStr[2:]
	}
	return hexStr
}
