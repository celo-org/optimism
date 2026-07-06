package batcher

import (
	"context"
	"fmt"
	"math/big"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoLightClient "github.com/EspressoSystems/espresso-network/sdks/go/light-client"
	op "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/espresso"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
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
// (HeaderHashByNumber + bind.ContractCaller).
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

func (a *batcherL1Adapter) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return a.L1Client.CodeAt(ctx, contract, blockNumber)
}

func (a *batcherL1Adapter) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return a.L1Client.CallContract(ctx, call, blockNumber)
}

// EspressoStreamer returns the Espresso batch streamer for use by the service and tests.
func (l *BatchSubmitter) EspressoStreamer() espresso.EspressoStreamer[derive.EspressoBatch] {
	return l.espressoStreamer
}

// setupEspressoStreamer constructs the Espresso streamer (and its buffered
// wrapper) for a freshly-built BatchSubmitter. Called from NewBatchSubmitter
// only when --espresso.enabled is set; no-op otherwise. Panics on streamer
// construction failure to mirror the existing NewBatchSubmitter behavior.
func (l *BatchSubmitter) setupEspressoStreamer() {
	if !l.Config.Espresso.Enabled {
		return
	}
	l1Adapter := &batcherL1Adapter{L1Client: l.L1Client}
	// Convert typed nil pointer to untyped nil interface to avoid typed-nil interface panic
	// in confirmEspressoBlockHeight when EspressoLightClient is not configured.
	var lightClientIface op.LightClientCallerInterface
	if l.Espresso.LightClient != nil {
		lightClientIface = l.Espresso.LightClient
	}
	unbufferedStreamer, err := op.NewEspressoStreamer(
		l.RollupConfig.L2ChainID.Uint64(),
		l1Adapter,
		l1Adapter,
		l.Espresso.Client,
		lightClientIface,
		l.Log,
		derive.CreateEspressoBatchUnmarshaler(),
		l.Config.Espresso.CaffeinationHeightEspresso,
		l.Config.Espresso.CaffeinationHeightL2,
		l.RollupConfig.BatchAuthenticatorAddress,
		false,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create Espresso streamer: %v", err))
	}
	l.espressoStreamer = op.NewBufferedEspressoStreamer(unbufferedStreamer)
	l.Log.Info("Streamer started", "streamer", l.espressoStreamer)
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

// resetEspressoStreamer resets the Espresso streamer when --espresso.enabled
// is set; no-op otherwise. Called from clearState alongside the upstream
// channel-manager reset so the streamer's view of "next batch" matches the
// freshly-cleared channel state.
func (l *BatchSubmitter) resetEspressoStreamer() {
	if l.Config.Espresso.Enabled {
		l.EspressoStreamer().Reset()
	}
}
