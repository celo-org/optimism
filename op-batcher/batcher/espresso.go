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
	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"golang.org/x/sync/errgroup"

	"github.com/ethereum-optimism/optimism/espresso"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum-optimism/optimism/op-service/bindings/batchauthenticator"
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

// espressoTransactionSubmitter submits transactions to Espresso and confirms
// their receipts. Each accepted transaction runs its whole lifecycle (submit,
// then poll for the receipt) in its own goroutine, bounded by an errgroup so at
// most maxInFlight run at once.
type espressoTransactionSubmitter struct {
	ctx                        context.Context
	wg                         *sync.WaitGroup
	eg                         *errgroup.Group // bounds concurrent submitAndConfirm goroutines
	espresso                   espressoClient.EspressoClient
	latestBlockHeight          atomic.Uint64 // shared HotShot block height, updated by trackBlockHeight
	verifyReceiptMaxBlocks     uint64
	verifyReceiptSafetyTimeout time.Duration
	verifyReceiptRetryDelay    time.Duration
	maxInFlight                int
}

// EspressoTransactionSubmitterConfig is a configuration struct for the
// EspressoTransactionSubmitter. It contains the configurable details for
// creating the EspressoTransactionSubmitter.
type EspressoTransactionSubmitterConfig struct {
	Ctx                        context.Context
	EspressoClient             espressoClient.EspressoClient
	Wg                         *sync.WaitGroup
	VerifyReceiptMaxBlocks     uint64
	VerifyReceiptSafetyTimeout time.Duration
	VerifyReceiptRetryDelay    time.Duration
	MaxInFlightJobs            int
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
// After returning, the caller should call Start to launch the block-height
// tracker, after which transactions can be submitted via SubmitTransaction.
func NewEspressoTransactionSubmitter(options ...EspressoTransactionSubmitterOption) *espressoTransactionSubmitter {
	config := EspressoTransactionSubmitterConfig{
		Ctx:                        context.Background(),
		Wg:                         new(sync.WaitGroup),
		VerifyReceiptMaxBlocks:     espresso.DefaultVerifyReceiptMaxBlocks,
		VerifyReceiptSafetyTimeout: espresso.DefaultVerifyReceiptSafetyTimeout,
		VerifyReceiptRetryDelay:    espresso.DefaultVerifyReceiptRetryDelay,
		MaxInFlightJobs:            espresso.DefaultMaxInFlightRequestsToEspresso,
	}

	for _, option := range options {
		option(&config)
	}

	if config.EspressoClient == nil {
		panic("Espresso client is required")
	}

	eg := new(errgroup.Group)
	eg.SetLimit(config.MaxInFlightJobs)

	return &espressoTransactionSubmitter{
		ctx:                        config.Ctx,
		wg:                         config.Wg,
		eg:                         eg,
		espresso:                   config.EspressoClient,
		verifyReceiptMaxBlocks:     config.VerifyReceiptMaxBlocks,
		verifyReceiptSafetyTimeout: config.VerifyReceiptSafetyTimeout,
		verifyReceiptRetryDelay:    config.VerifyReceiptRetryDelay,
		maxInFlight:                config.MaxInFlightJobs,
	}
}

// ErrTooManyInFlightRequests is returned when the maximum number of in-flight
// transactions is already running, so a new one cannot be started right now.
type ErrTooManyInFlightRequests struct {
	MaxInFlightRequests int
}

// Error implements error
func (e ErrTooManyInFlightRequests) Error() string {
	return fmt.Sprintf("too many requests in flight to espresso, maximum allowed: %d", e.MaxInFlightRequests)
}

// SubmitTransaction starts a goroutine that submits the transaction to Espresso
// and confirms its receipt, retrying as needed.
//
// It does NOT block: if the maximum number of in-flight transactions is already
// running, it returns ErrTooManyInFlightRequests immediately so the caller can
// apply back pressure and try again later.
func (s *espressoTransactionSubmitter) SubmitTransaction(tx *espressoCommon.Transaction) error {
	// wg tracks the goroutine for shutdown; the errgroup bounds concurrency.
	s.wg.Add(1)
	started := s.eg.TryGo(func() error {
		defer s.wg.Done()
		s.submitAndConfirm(s.ctx, tx)
		return nil
	})
	if !started {
		s.wg.Done()
		return ErrTooManyInFlightRequests{MaxInFlightRequests: s.maxInFlight}
	}
	return nil
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
func evaluateSubmission(err error) JobEvaluation {
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

// submitAndConfirm runs the full lifecycle for one transaction: submit it to
// Espresso, then poll until its receipt is verifiable. A verification timeout
// loops back to re-submit. It returns once the transaction is confirmed,
// permanently fails, or the context is cancelled.
func (s *espressoTransactionSubmitter) submitAndConfirm(ctx context.Context, tx *espressoCommon.Transaction) {
	for {
		hash, ok := s.submitToEspresso(ctx, tx)
		if !ok {
			return // context cancelled or a permanent submit failure
		}

		switch s.confirmReceipt(ctx, hash) {
		case verifyDone:
			log.Info("Transaction confirmed on Espresso", "hash", hash.String())
			return
		case verifyResubmit:
			continue // verification timed out; submit a fresh copy
		default: // verifyAbort
			return
		}
	}
}

// submitToEspresso submits the transaction, retrying ephemeral failures (with a
// short delay so a persistently failing endpoint is not hammered). It returns
// the transaction hash on success, or ok=false if the failure is permanent or
// the context is cancelled.
func (s *espressoTransactionSubmitter) submitToEspresso(ctx context.Context, tx *espressoCommon.Transaction) (*espressoCommon.TaggedBase64, bool) {
	for {
		if ctx.Err() != nil {
			return nil, false
		}

		hash, err := s.espresso.SubmitTransaction(ctx, *tx)
		if err == nil {
			log.Info("Submitted transaction to Espresso", "hash", hash)
		}

		switch evaluateSubmission(err) {
		case Handle:
			return hash, true
		case Skip:
			return nil, false
		default: // RetrySubmission
			if !sleep(ctx, s.verifyReceiptRetryDelay) {
				return nil, false
			}
		}
	}
}

// verifyOutcome is the result of confirmReceipt.
type verifyOutcome int

const (
	verifyDone     verifyOutcome = iota // receipt is verifiable; we are done
	verifyResubmit                      // verification timed out; re-submit
	verifyAbort                         // permanent failure or context cancelled
)

// confirmReceipt polls for the transaction's receipt until it is verifiable,
// re-submitting (verifyResubmit) if it takes too long. The first attempt is
// immediate; subsequent attempts wait verifyReceiptRetryDelay.
func (s *espressoTransactionSubmitter) confirmReceipt(ctx context.Context, hash *espressoCommon.TaggedBase64) verifyOutcome {
	startTime := time.Now()
	// Snapshot the current height so we can measure how many blocks pass.
	startHeight := s.latestBlockHeight.Load()

	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			if !sleep(ctx, s.verifyReceiptRetryDelay) {
				return verifyAbort
			}
		}
		if ctx.Err() != nil {
			return verifyAbort
		}

		_, err := s.espresso.FetchTransactionByHash(ctx, hash)
		currentHeight := s.latestBlockHeight.Load()

		switch s.evaluateVerification(err, startHeight, currentHeight, startTime) {
		case Handle:
			return verifyDone
		case Skip:
			return verifyAbort
		case RetrySubmission:
			return verifyResubmit
		default: // RetryVerification
			continue
		}
	}
}

// sleep waits for d or until ctx is cancelled. It returns false if ctx was
// cancelled first.
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
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
func (s *espressoTransactionSubmitter) evaluateVerification(err error, startHeight, currentHeight uint64, startTime time.Time) JobEvaluation {
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
	if startHeight > 0 && currentHeight >= startHeight+s.verifyReceiptMaxBlocks {
		log.Info("Verification timed out by block count, re-submitting",
			"startHeight", startHeight,
			"currentHeight", currentHeight,
			"maxBlocks", s.verifyReceiptMaxBlocks)
		return RetrySubmission
	}

	// Wall-clock safety backstop in case the block height tracker is stale
	// or broken (e.g., query service returning old data).
	if elapsed := time.Since(startTime); elapsed > s.verifyReceiptSafetyTimeout {
		log.Warn("Verification timed out by safety timeout, re-submitting",
			"elapsed", elapsed,
			"safetyTimeout", s.verifyReceiptSafetyTimeout)
		return RetrySubmission
	}

	// Otherwise, retry the verification.
	return RetryVerification
}

// trackBlockHeight periodically polls FetchLatestBlockHeight and stores
// the result in s.latestBlockHeight for verify jobs to compare against.
// This avoids redundant height queries from individual verify goroutines.
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

// Start launches the shared block-height tracker. Per-transaction work is
// started on demand by SubmitTransaction.
func (s *espressoTransactionSubmitter) Start() {
	go s.trackBlockHeight()
}

// Converts a block to an EspressoBatch and starts a goroutine that publishes it to Espresso
// Returns error only if batch conversion fails, otherwise it is infallible, as the goroutine
// will retry publishing until successful.
func (l *BatchSubmitter) queueBlockToEspresso(ctx context.Context, block *types.Block) error {
	espressoBatch, err := derivation.BlockToEspressoBatch(l.RollupConfig, block)
	if err != nil {
		l.Log.Warn("Failed to derive batch from block", "err", err)
		return fmt.Errorf("failed to derive batch from block: %w", err)
	}

	transaction, err := espressoBatch.ToEspressoTransaction(ctx, bigs.Uint64Strict(l.RollupConfig.L2ChainID), l.Espresso.ChainSigner)
	if err != nil {
		l.Log.Warn("Failed to create Espresso transaction from a batch", "err", err)
		return fmt.Errorf("failed to create Espresso transaction from a batch: %w", err)
	}

	commitment := transaction.Commit()
	hash, _ := tagged_base64.New("TX", commitment[:])
	l.Log.Info("Created Espresso transaction from batch", "hash", hash, "batchNr", bigs.Uint64Strict(espressoBatch.BatchHeader.Number))

	if err := l.espressoSubmitter.SubmitTransaction(transaction); err != nil {
		return fmt.Errorf("failed to submit job to espresso: %w", err)
	}

	return nil
}

// espressoSyncChannelManager reconciles the channel manager with the latest sync
// status, reporting whether the sequencer is out of sync (malformed status or
// reversed CurrentL1). The streamer no longer needs pumping here: it refreshes L1
// finality and fetches HotShot blocks from its own poll loops.
func (l *BatchSubmitter) espressoSyncChannelManager(newSyncStatus *eth.SyncStatus) (outOfSync bool) {
	l.channelMgrMutex.Lock()
	defer l.channelMgrMutex.Unlock()
	syncActions, outOfSync := computeSyncActions(*newSyncStatus, l.prevCurrentL1, l.channelMgr.blocks, l.channelMgr.channelQueue, l.Log)
	if outOfSync {
		l.degradedLog.Warn(l.Log, "sequencerOutOfSync", "Sequencer is out of sync, retrying next tick.")
		return true
	}
	l.degradedLog.Clear(l.Log, "sequencerOutOfSync", "Sequencer back in sync")
	l.prevCurrentL1 = newSyncStatus.CurrentL1
	if syncActions.clearState != nil {
		l.channelMgr.Clear(*syncActions.clearState)
		// LocalSafeL2, matching the base computeSyncActions derived clearState from:
		// the channel manager and the streamer must not be reset onto different heads.
		// Always at or past the caffeination point: startup gates on that.
		l.espressoStreamer.SetBatchPosition(newSyncStatus.LocalSafeL2)
	} else {
		l.channelMgr.PruneSafeBlocks(syncActions.blocksToPrune)
		l.channelMgr.PruneChannels(syncActions.channelsToPrune)
	}
	return false
}

// requestClearState asks the batch loading loop to perform `l.clearState`.
func (l *BatchSubmitter) requestClearState() {
	l.clearStateRequested.Store(true)
}

// performClearState runs clearState if it was requested via requestClearState,
// reporting whether a clear was performed.
func (l *BatchSubmitter) performClearState(ctx context.Context) bool {
	if !l.clearStateRequested.CompareAndSwap(true, false) {
		return false
	}
	l.Log.Info("Clearing state as requested by the block queueing loop")
	l.clearState(ctx)
	return true
}

// Periodically refreshes the sync status and drains the Espresso streamer of any
// batches that extend the tip it is tracking.
// Owns publishSignal and unsafeBytesUpdated: it is their only closer, so the loops
// ranging over them (publishingLoop, throttlingLoop) terminate when this loop exits.
func (l *BatchSubmitter) espressoBatchLoadingLoop(ctx context.Context, wg *sync.WaitGroup, publishSignal chan pubInfo, unsafeBytesUpdated chan int64) {
	l.Log.Info("Starting EspressoBatchLoadingLoop", "polling interval", l.Config.Espresso.PollInterval)

	defer wg.Done()
	ticker := time.NewTicker(l.Config.Espresso.PollInterval)
	defer ticker.Stop()
	defer close(publishSignal)
	defer close(unsafeBytesUpdated)

	for {
		select {
		case <-ticker.C:
			// Check if block loader requested to clear state
			l.performClearState(ctx)

			newSyncStatus, err := l.getSyncStatus(ctx)
			if err != nil {
				l.degradedLog.Warn(l.Log, "syncStatusErr/espressoBatchLoading", "failed to refresh sync status", "err", err)
				continue
			}
			l.degradedLog.Clear(l.Log, "syncStatusErr/espressoBatchLoading", "sync status fetch recovered")

			// An out-of-sync status (zeroed fields or reversed CurrentL1) cannot
			// be trusted as the drain floor: with LocalSafeL2 zeroed, the
			// stale-batch re-anchor check below never fires and already-derived
			// blocks would be republished. Skip the tick, mirroring how the base
			// driver skips loading when computeSyncActions reports out-of-sync.
			if l.espressoSyncChannelManager(newSyncStatus) {
				continue
			}

			blocksAdded := 0

			for {
				// Check if block loader requested to clear state
				if l.performClearState(ctx) {
					break
				}

				batch := l.espressoStreamer.Peek(ctx)
				if batch == nil {
					break
				}

				// A batch at or below the local-safe head is already derived from L1,
				// and adding it would make the publish path resubmit it. The cursor
				// falls behind the safe head when previously submitted channels finish
				// deriving while the streamer backfills (e.g. after a restart), and an
				// empty channel manager gives computeSyncActions nothing to reconcile.
				// Jump past the whole stale range in one re-anchor: the sync status ref
				// is canonical, unlike a stale candidate's own hash, which advancing
				// batch-by-batch would promote to the streamer's tip.
				if batch.Number() <= newSyncStatus.LocalSafeL2.Number {
					l.Log.Info("Peeked batch at or below the local-safe head, re-anchoring the streamer",
						"batchNr", batch.Number(), "localSafeL2", newSyncStatus.LocalSafeL2)
					l.espressoStreamer.SetBatchPosition(newSyncStatus.LocalSafeL2)
					break
				}

				// This should happen ONLY if the batch is malformed. ToBlock has to guarantee no
				// transient errors. Advancing past it would promote a block the channel manager
				// never received to the streamer's tip, stalling every later batch, so re-anchor
				// instead of skipping.
				block, err := batch.ToBlock(l.RollupConfig)
				if err != nil {
					l.Log.Error("failed to convert singular batch to block", "err", err)
					l.clearState(ctx)
					break
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
					// clearState re-anchors the streamer to the safe head.
					l.clearState(ctx)
					break
				}

				l.espressoStreamer.AdvancePosition()
				l.Log.Info("Added L2 block to channel manager", "blockNr", block.NumberU64())

				// During a large drain, signal periodically so throttling can engage
				// before the whole backlog is consumed (mirrors loadBlocksIntoState).
				blocksAdded++
				if blocksAdded%100 == 0 {
					l.sendToThrottlingLoop(unsafeBytesUpdated)
				}
			}

			l.sendToThrottlingLoop(unsafeBytesUpdated)
			l.tryPublishSignal(publishSignal, pubInfo{})

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

func (l *BlockLoader) reset() {
	l.prevSyncStatus = nil
	l.queuedBlocks = nil
	l.batcher.requestClearState()
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
			l.reset()
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
	// Mirror computeSyncActions' zero-field guard: op-node has transiently
	// reported statuses with individual fields zeroed while the rest are
	// populated (see sync_actions_test.go). Treating a zero LocalSafeL2 as the
	// queue floor would enqueue the entire pre-caffeination history, so retry
	// until the status is fully populated.
	if isZero(newSyncStatus.LocalSafeL2) ||
		isZero(newSyncStatus.UnsafeL2) ||
		isZero(newSyncStatus.HeadL1) ||
		isZero(newSyncStatus.CurrentL1) {
		l.batcher.degradedLog.Warn(l.batcher.Log, "emptySyncStatusField", "empty BlockRef in sync status",
			"localSafeL2", newSyncStatus.LocalSafeL2, "unsafeL2", newSyncStatus.UnsafeL2,
			"headL1", newSyncStatus.HeadL1, "currentL1", newSyncStatus.CurrentL1)
		return inclusiveBlockRange{}, ActionRetry
	}
	l.batcher.degradedLog.Clear(l.batcher.Log, "emptySyncStatusField", "sync status fully populated")

	if l.prevSyncStatus != nil && newSyncStatus.CurrentL1.Number < l.prevSyncStatus.CurrentL1.Number {
		// Sequencer restarted and hasn't caught up yet
		l.batcher.degradedLog.Warn(l.batcher.Log, "sequencerCurrentL1Reversed", "sequencer currentL1 reversed", "new currentL1", newSyncStatus.CurrentL1.Number, "previous currentL1", l.prevSyncStatus.CurrentL1.Number)
		return inclusiveBlockRange{}, ActionRetry
	}
	l.batcher.degradedLog.Clear(l.batcher.Log, "sequencerCurrentL1Reversed", "sequencer currentL1 caught up")

	l.prevSyncStatus = newSyncStatus

	// LocalSafeL2 rather than SafeL2 (cross-safe), for the same reason as
	// computeSyncActions: cross-safe can lag local-safe, and blocks at or below
	// local-safe are already derived from L1 so they must not be re-enqueued.
	// A populated LocalSafeL2 is always at or past the caffeination point
	// (startup gates on that; the guard above rejects zeroed statuses), so
	// pre-caffeination blocks - the fallback batcher's - are never enqueued.
	safeL2 := newSyncStatus.LocalSafeL2

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
				if enqueued := numEnqueuedBlocksAfter - numEnqueuedBlocksBefore; enqueued < int(blocksToQueue.numBlocks()) {
					// We weren't able to submit all of the blocks to Espresso
					// that we were attempting to.
					//
					// TODO: We should probably throttle a bit.
					l.Log.Debug("Could not enqueue all blocks to Espresso", "enqueued", enqueued, "attempted", blocksToQueue.numBlocks())
				}
			} else if action == ActionReset {
				loader.reset()
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
func (l *BatchSubmitter) resolveTEEVerifierAddress(ctx context.Context) error {
	if l.RollupConfig.BatchAuthenticatorAddress == (common.Address{}) {
		// If batcher authenticator address is nil, we will keep teeVerifierAddress to nil as well
		return nil
	}
	auth, err := batchauthenticator.NewBatchAuthenticatorCaller(l.RollupConfig.BatchAuthenticatorAddress, l.L1Client)
	if err != nil {
		return fmt.Errorf("failed to create BatchAuthenticator caller: %w", err)
	}
	callCtx, cancel := l.networkTimeoutCtx(ctx)
	defer cancel()
	addr, err := auth.EspressoTEEVerifier(&bind.CallOpts{Context: callCtx})
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
	codeCtx, cancel := l.networkTimeoutCtx(ctx)
	code, err := l.L1Client.CodeAt(codeCtx, l.RollupConfig.BatchAuthenticatorAddress, nil)
	cancel()
	if err != nil {
		return fmt.Errorf("failed to check code at contract address: %w", err)
	}
	if len(code) == 0 {
		return fmt.Errorf("no contract deployed at this address %w", err)
	}

	abi, err := batchauthenticator.BatchAuthenticatorMetaData.GetAbi()
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

// sendTxWithEspresso authenticates a batch transaction via the BatchAuthenticator contract using
// a TEE-attested EIP-712 signature, then sends the batch data to the BatchInbox address. Both txs
// are submitted through the ordered txmgr queue (auth first, batch second) so they are mined in
// submission order, as required under Holocene, and both stay under MaxPendingTransactions
// (queue.Send blocks when the queue is full). A watcher goroutine collects both receipts and
// emits a single synthetic receipt for the batch txData.
func (l *BatchSubmitter) sendTxWithEspresso(txdata txData, isCancel bool, candidate *txmgr.TxCandidate, queue TxSender[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) {
	transactionReference := txRef{id: txdata.ID(), isCancel: isCancel, isBlob: txdata.daType == DaTypeBlob, daType: txdata.daType, size: txdata.Len()}
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

	batchAuthenticatorAbi, err := batchauthenticator.BatchAuthenticatorMetaData.GetAbi()
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

	l.submitAuthenticatedBatch(transactionReference, verifyCandidate, candidate, queue, receiptsCh)
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
