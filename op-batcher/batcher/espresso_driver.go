package batcher

import (
	"context"
	"fmt"
	"math/big"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoLightClient "github.com/EspressoSystems/espresso-network/sdks/go/light-client"
	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum-optimism/optimism/op-service/bigs"
	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

// EspressoDriverSetup groups all TEE-batcher-specific runtime state plumbed
// from BatcherService into DriverSetup. Defined here to keep the upstream
// Optimism DriverSetup field block compact (see driver.go).
//
// All fields are nil/zero when --espresso.enabled is false except for the
// fallback batcher's ChainSigner/SequencerAddress, which are always populated
// by applyEspressoDriverSetup.
type EspressoDriverSetup struct {
	Client           *espressoClient.MultipleNodesClient
	LightClient      *espressoLightClient.LightclientCaller
	ChainSigner      opcrypto.ChainSigner
	SequencerAddress common.Address
	Attestation      []byte
}

// batcherL1Adapter wraps the batcher's L1Client to implement espresso.L1Client
// (HeaderHashByNumber + HeaderByNumber + bind.ContractCaller).
type batcherL1Adapter struct {
	L1Client L1Client
}

func (a *batcherL1Adapter) HeaderHashByNumber(ctx context.Context, number *big.Int) (common.Hash, error) {
	h, err := a.L1Client.HeaderByNumber(ctx, number)
	if err != nil {
		return common.Hash{}, err
	}
	return h.Hash(), nil
}

func (a *batcherL1Adapter) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return a.L1Client.HeaderByNumber(ctx, number)
}

func (a *batcherL1Adapter) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return a.L1Client.CodeAt(ctx, contract, blockNumber)
}

func (a *batcherL1Adapter) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return a.L1Client.CallContract(ctx, call, blockNumber)
}

// batcherL2Adapter wraps the batcher's L2 eth client to implement espresso.L2Client.
// The streamer uses it once, at construction, to resolve the block hash of the
// batch position it is anchored to. dial.EthClientInterface exposes no
// header-only accessor, so this fetches the full block and takes its hash.
type batcherL2Adapter struct {
	EthClient dial.EthClientInterface
}

func (a *batcherL2Adapter) HeaderHashByNumber(ctx context.Context, number *big.Int) (common.Hash, error) {
	block, err := a.EthClient.BlockByNumber(ctx, number)
	if err != nil {
		return common.Hash{}, err
	}
	return block.Hash(), nil
}

// networkTimeoutCtx bounds a single external call with the configured network
// timeout. Every raw RPC read on the Espresso startup path must go through it:
// StartBatchSubmitting holds the start mutex, and StopBatchSubmitting needs that
// mutex before it can cancel anything, so an unbounded call on a stalled endpoint
// would wedge the batcher beyond even a graceful shutdown. Calls with their own
// timeout regime (the attestation service client, Txmgr.Send) are exempt.
func (l *BatchSubmitter) networkTimeoutCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, l.Config.NetworkTimeout)
}

// setupEspressoStreamer constructs the Espresso streamer for a BatchSubmitter that
// is starting up; no-op when --espresso.enabled is false.
//
// Called from StartBatchSubmitting rather than NewBatchSubmitter: it waits for the
// rollup node to report a local-safe L2 head, so it needs a context and an L2 node
// that has reached genesis. It also returns an error rather than panicking, which
// construction inside NewBatchSubmitter could not do.
//
// The streamer is anchored at the local-safe head, which waitForLocalSafeHead
// guarantees is at or past --espresso.origin-height-l2, so the streamer never starts
// inside pre-Espresso history (those blocks are the fallback batcher's to submit).
// Gating startup on the caffeination point being local-safe - rather than anchoring
// on the configured height directly - means the anchor hash always comes from the
// derived chain: an unsafe caffeination block could reorg after being cached and
// permanently wedge fork selection. The configured height is never looked up either:
// NewStreamer resolves its anchor block's hash from the L2 client, which fails on
// endpoints that no longer serve the (ever aging) configured height and would wedge
// restarts. --espresso.origin-height-espresso keeps its role: it decides where the
// streamer starts polling HotShot.
func (l *BatchSubmitter) setupEspressoStreamer(ctx context.Context) error {
	if !l.Config.Espresso.Enabled {
		return nil
	}

	anchor, err := l.waitForLocalSafeHead(ctx)
	if err != nil {
		return err
	}

	ethClient, err := l.EndpointProvider.EthClient(ctx)
	if err != nil {
		return fmt.Errorf("getting the L2 eth client for the Espresso streamer: %w", err)
	}

	// Convert typed nil pointer to untyped nil interface to avoid typed-nil interface panic
	// in confirmEspressoBlockHeight when EspressoLightClient is not configured.
	var lightClientIface espressoStreamers.LightClientCallerInterface
	if l.Espresso.LightClient != nil {
		lightClientIface = l.Espresso.LightClient
	}

	// NewStreamer performs one L2 lookup (its anchor block's hash), so it gets the
	// same per-call bound as every other startup RPC.
	streamerCtx, cancel := l.networkTimeoutCtx(ctx)
	defer cancel()
	streamer, err := espressoStreamers.NewStreamer(
		streamerCtx,
		l.Espresso.Client,
		&batcherL1Adapter{L1Client: l.L1Client},
		&batcherL2Adapter{EthClient: ethClient},
		lightClientIface,
		l.RollupConfig.BatchAuthenticatorAddress,
		bigs.Uint64Strict(l.RollupConfig.L2ChainID),
		derivation.CreateEspressoBatchUnmarshaler(),
		l.getSyncStatus,
		l.Config.Espresso.PollInterval,
		l.Log,
		l.Config.Espresso.CaffeinationHeightEspresso,
		anchor.Number,
	)
	if err != nil {
		return fmt.Errorf("failed to create Espresso streamer: %w", err)
	}
	l.espressoStreamer = streamer

	// Re-anchor on the exact ref we resolved: NewStreamer resolved a hash for the
	// same height, but the sync status is the authoritative view.
	streamer.SetBatchPosition(anchor)
	l.Log.Info("Anchored the Espresso streamer at the local-safe L2 head", "anchor", anchor)
	return nil
}

const (
	espressoAnchorTimeout       = 1 * time.Minute
	espressoAnchorRetryInterval = 1 * time.Second
)

// waitForLocalSafeHead polls the sync status until it reports a local-safe L2 head at
// or past the caffeination point to anchor the streamer on, giving up after
// espressoAnchorTimeout.
//
// LocalSafeL2 rather than SafeL2 (cross-safe): every streamer anchor must share the
// base computeSyncActions and safeL1Origin derive clearing/pruning from, and cross-safe
// can lag local-safe (see the note in computeSyncActions).
//
// Requiring the caffeination point to be local-safe is what enforces the anchor floor:
// the Espresso batcher refuses to run - with a clean, retryable startup failure - until
// the fallback batcher's pre-activation channels have derived. Blocks at or below
// --espresso.origin-height-l2 are thereby never this batcher's to track, and the
// anchor hash always comes from the derived chain rather than a reorgable unsafe block.
func (l *BatchSubmitter) waitForLocalSafeHead(ctx context.Context) (eth.L2BlockRef, error) {
	ctx, cancel := context.WithTimeout(ctx, espressoAnchorTimeout)
	defer cancel()

	ticker := time.NewTicker(espressoAnchorRetryInterval)
	defer ticker.Stop()

	caffeinationL2 := l.Config.Espresso.CaffeinationHeightL2
	for {
		syncStatus, err := l.getSyncStatus(ctx)
		switch {
		case err != nil:
			l.Log.Warn("Failed to fetch sync status to anchor the Espresso streamer, retrying", "err", err)
		case syncStatus.LocalSafeL2 == (eth.L2BlockRef{}):
			l.Log.Warn("Sync status has no local-safe L2 head yet, retrying")
		case syncStatus.LocalSafeL2.Number < caffeinationL2:
			l.Log.Warn("Local-safe L2 head has not reached the caffeination point yet, retrying",
				"localSafeL2", syncStatus.LocalSafeL2.Number, "caffeinationL2", caffeinationL2)
		default:
			return syncStatus.LocalSafeL2, nil
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return eth.L2BlockRef{}, fmt.Errorf(
				"no local-safe L2 head at or past the caffeination point (%d) to anchor the Espresso streamer within %s (safe to retry once activation has been derived from L1): %w",
				caffeinationL2, espressoAnchorTimeout, ctx.Err())
		}
	}
}

// rollbackFailedStart undoes the startup state set at the top of StartBatchSubmitting
// (running flag, shutdown/kill contexts) so a failed start does not wedge later
// attempts behind "batcher is already running". Cancelling shutdownCtx also winds down
// the streamer's poll loops if they were started, and Stop waits for them; a streamer
// that was constructed but never started makes Stop a no-op. Only the Espresso setup
// error paths roll back: the upstream error paths (waitForL2Genesis, waitNodeSync)
// deliberately keep upstream's behavior. Caller must hold l.mutex.
func (l *BatchSubmitter) rollbackFailedStart() {
	l.cancelShutdownCtx()
	l.cancelKillCtx()
	if l.espressoStreamer != nil {
		l.espressoStreamer.Stop()
	}
	l.running = false
}

// startEspressoLoops registers the batcher with the BatchAuthenticator
// contract, resolves the TEE verifier address, spawns the Espresso transaction
// submitter, and starts the four Espresso-specific batcher goroutines (in
// addition to the upstream receiptsLoop and publishingLoop). Replaces the
// upstream three-goroutine pattern when --espresso.enabled is set.
func (l *BatchSubmitter) startEspressoLoops(receiptsCh chan txmgr.TxReceipt[txRef], publishSignal chan pubInfo, unsafeBytesUpdated chan int64) error {
	if err := l.registerBatcher(l.killCtx); err != nil {
		return fmt.Errorf("could not register with BatchAuthenticator contract: %w", err)
	}

	// Resolve the TEE verifier address from the BatchAuthenticator contract.
	if err := l.resolveTEEVerifierAddress(l.killCtx); err != nil {
		return fmt.Errorf("could not resolve TEE verifier address: %w", err)
	}

	// The streamer drives itself from its own poll loops, so it is started here rather
	// than being pumped by espressoBatchLoadingLoop. Bound to shutdownCtx so it stops
	// fetching before the publish path winds down. Kept as the last setup step that
	// can fail, so a setup error never has running poll loops to unwind
	// (rollbackFailedStart would stop them anyway).
	if err := l.espressoStreamer.Start(l.shutdownCtx); err != nil {
		return fmt.Errorf("could not start the Espresso streamer: %w", err)
	}

	l.espressoSubmitter = NewEspressoTransactionSubmitter(
		WithContext(l.shutdownCtx),
		WithWaitGroup(l.wg),
		WithEspressoClient(l.Espresso.Client),
		WithVerifyReceiptMaxBlocks(l.Config.Espresso.VerifyReceiptMaxBlocks),
		WithVerifyReceiptSafetyTimeout(l.Config.Espresso.VerifyReceiptSafetyTimeout),
		WithVerifyReceiptRetryDelay(l.Config.Espresso.VerifyReceiptRetryDelay),
	)
	l.espressoSubmitter.SpawnWorkers(4, 4)
	l.espressoSubmitter.Start()

	l.wg.Add(4)
	go l.receiptsLoop(l.wg, receiptsCh) // ranges over receiptsCh channel
	go l.espressoBatchQueueingLoop(l.shutdownCtx, l.wg)
	go l.espressoBatchLoadingLoop(l.shutdownCtx, l.wg, publishSignal, unsafeBytesUpdated) // sends on unsafeBytesUpdated (if throttling enabled) and publishSignal. Closes them both when done
	go l.publishingLoop(l.killCtx, l.wg, receiptsCh, publishSignal)                       // ranges over publishSignal, spawns routines which send on receiptsCh. Closes receiptsCh when done.
	return nil
}

// shouldSkipPublishForActiveSeq returns true if publishStateToL1 should skip
// publishing because this batcher is not the on-chain "active" batcher
// (BatchAuthenticator.activeIsEspresso). The Espresso TEE batcher always
// honors the flag (it is fundamentally a post-fork actor); the fallback
// batcher honors it only once fallback auth is required (pre-fork it must run
// as a vanilla upstream Optimism batcher with no BatchAuthenticator coupling).
// Fails closed: if either gate cannot be evaluated, publishing is skipped for
// this tick and retried on the next.
func (l *BatchSubmitter) shouldSkipPublishForActiveSeq(ctx context.Context) bool {
	if l.RollupConfig.BatchAuthenticatorAddress == (common.Address{}) {
		return false
	}
	consultActiveFlag := l.Config.Espresso.Enabled
	if !consultActiveFlag {
		fallbackAuthRequired, err := l.isFallbackAuthRequired(ctx)
		if err != nil {
			l.Log.Warn("Failed to evaluate fallback-auth gate, skipping publish", "err", err)
			return true
		}
		consultActiveFlag = fallbackAuthRequired
	}
	if !consultActiveFlag {
		return false
	}
	isActive, err := l.isBatcherActive(ctx)
	if err != nil {
		l.Log.Warn("Failed to check if batcher is active, skipping publish", "err", err)
		return true
	}
	return !isActive
}

// espressoReanchorTarget fetches the local-safe L2 head the Espresso streamer
// must be re-anchored to when clearState resets the channel manager. clearState
// calls it BEFORE clearing anything so the clear and the re-anchor are atomic:
// clearing first and fetching after would, on a transient fetch failure, leave
// an emptied channel manager with the streamer still at its old cursor, and the
// blocks in between would only be recovered once computeSyncActions notices the
// gap - or not at all while the streamer has nothing to serve.
//
// Reports ok=false when the sync status cannot be fetched: the caller must
// retry the whole clear rather than perform it partially. Reports a nil target
// (and ok=true) when there is nothing to re-anchor: --espresso.enabled unset,
// or the startup path, where clearState runs before the streamer is constructed.
func (l *BatchSubmitter) espressoReanchorTarget(ctx context.Context) (target *eth.L2BlockRef, ok bool) {
	if !l.Config.Espresso.Enabled || l.espressoStreamer == nil {
		return nil, true
	}
	syncStatus, err := l.getSyncStatus(ctx)
	if err != nil {
		l.Log.Warn("Failed to fetch sync status to re-anchor the Espresso streamer, will retry the clear", "err", err)
		return nil, false
	}
	return &syncStatus.LocalSafeL2, true
}
