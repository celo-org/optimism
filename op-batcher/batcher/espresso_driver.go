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

// setupEspressoStreamer constructs the Espresso streamer for a BatchSubmitter that
// is starting up; no-op when --espresso.enabled is false.
//
// Called from StartBatchSubmitting rather than NewBatchSubmitter: the streamer
// resolves its anchor batch's hash from the L2 client while constructing, so it
// needs a context and an L2 node that has reached genesis. It also returns an error
// rather than panicking, which construction inside NewBatchSubmitter could not do.
func (l *BatchSubmitter) setupEspressoStreamer(ctx context.Context) error {
	if !l.Config.Espresso.Enabled {
		return nil
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

	streamer, err := espressoStreamers.NewStreamer(
		ctx,
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
		l.Config.Espresso.CaffeinationHeightL2,
	)
	if err != nil {
		return fmt.Errorf("failed to create Espresso streamer: %w", err)
	}
	l.espressoStreamer = streamer

	// Re-anchor to the safe L2 head.
	return l.anchorEspressoStreamerAtSafeHead(ctx)
}

const (
	espressoAnchorTimeout       = 1 * time.Minute
	espressoAnchorRetryInterval = 1 * time.Second
)

// anchorEspressoStreamerAtSafeHead repositions a freshly-constructed streamer from
// its configured origin onto the current local-safe L2 head, waiting for the sync
// status to report one.
//
// LocalSafeL2 rather than SafeL2 (cross-safe): every streamer anchor must share the
// base computeSyncActions and safeL1Origin derive clearing/pruning from, and cross-safe
// can lag local-safe (see the note in computeSyncActions).
func (l *BatchSubmitter) anchorEspressoStreamerAtSafeHead(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, espressoAnchorTimeout)
	defer cancel()

	ticker := time.NewTicker(espressoAnchorRetryInterval)
	defer ticker.Stop()

	for {
		syncStatus, err := l.getSyncStatus(ctx)
		switch {
		case err != nil:
			l.Log.Warn("Failed to fetch sync status to anchor the Espresso streamer, retrying", "err", err)
		case syncStatus.LocalSafeL2 == (eth.L2BlockRef{}):
			l.Log.Warn("Sync status has no local-safe L2 head yet, retrying")
		default:
			l.espressoStreamer.SetBatchPosition(syncStatus.LocalSafeL2)
			l.Log.Info("Anchored the Espresso streamer at the local-safe L2 head", "localSafeL2", syncStatus.LocalSafeL2)
			return nil
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return fmt.Errorf("could not anchor the Espresso streamer at the safe L2 head within %s: %w", espressoAnchorTimeout, ctx.Err())
		}
	}
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
	if err := l.resolveTEEVerifierAddress(); err != nil {
		return fmt.Errorf("could not resolve TEE verifier address: %w", err)
	}

	// The streamer drives itself from its own poll loops, so it is started here rather
	// than being pumped by espressoBatchLoadingLoop. Bound to shutdownCtx so it stops
	// fetching before the publish path winds down. Must stay the last setup step that
	// can fail: a failed StartBatchSubmitting returns without cancelling shutdownCtx,
	// so an error after this point would leave the streamer's poll loops running.
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

// resetEspressoStreamer re-anchors the Espresso streamer to the local-safe L2 head
// when --espresso.enabled is set; no-op otherwise. Called from clearState alongside the
// upstream channel-manager reset so the streamer's view of "next batch" matches the
// freshly-cleared channel state (which is likewise derived from LocalSafeL2).
//
// The nil check covers the startup path: clearState runs before the streamer is
// constructed, so the first call of a start cycle finds nothing to re-anchor.
func (l *BatchSubmitter) resetEspressoStreamer(ctx context.Context) {
	if !l.Config.Espresso.Enabled || l.espressoStreamer == nil {
		return
	}
	syncStatus, err := l.getSyncStatus(ctx)
	if err != nil {
		l.Log.Warn("Failed to fetch sync status to re-anchor the Espresso streamer, keeping the current position", "err", err)
		return
	}
	l.espressoStreamer.SetBatchPosition(syncStatus.LocalSafeL2)
}
