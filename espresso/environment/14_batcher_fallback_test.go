package environment_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/espresso/bindings"
	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-batcher/compressor"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive/params"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

// waitForRollupToMovePastL1Block waits until the targeted rollup cli moves
// past the reference l1BlockNumber.  This indicates that the rollupCli is
// still receiveing information that indicates that the L1 has progressed
// past the desired height.
//
// For convenience, this also returns the Local Safe L2 Height of the last
// call to the Sync Status on the Rollup Client.  If this wait passes, it
// will be the LocalSafeL2 height of the SyncStatus that exceeded the
// referenced l1BlockNumber.
func waitForRollupToMovePastL1Block(ctx context.Context, rollupCli *sources.RollupClient, l1BlockNumber uint64) (uint64, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	var localSafeL2Height uint64
	defer cancel()
	err := wait.For(timeoutCtx, 100*time.Millisecond, func() (bool, error) {
		status, err := rollupCli.SyncStatus(ctx)
		if err != nil {
			return false, err
		}

		localSafeL2Height = status.LocalSafeL2.Number
		return status.CurrentL1.Number > l1BlockNumber, nil
	})

	return localSafeL2Height, err
}

// TestBatcherSwitching is a test case that is meant to verify the correct
// expected behavior when switching between an Espresso batcher and a
// fallback batcher, ensuring seamless transitions in both directions.
//
// In this scenario the test starts with the batcher running in Espresso
// mode and verifies transactions work correctly. It then stops the Espresso batcher,
// sends switch action to the Batch Authenticator contract and switches to the
// fallback batcher, verifies transactions continue to go through. Next, it switches
// back to the Espresso batcher by restarting it with proper caffeination heights
// (both Espresso and L2 heights set to ensure correct sync points) and verifies
// that transactions continue to go through.
func TestBatcherSwitching(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	// We will need this config to start a new instance of "TEE" batcher
	// with parameters tweaked.
	batcherConfig := &batcher.CLIConfig{}
	// L1FinalizedDistance(0) to avoid long delays after batcher switch.
	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t,
		env.WithL1FinalizedDistance(0),
		env.WithSequencerUseFinalized(true),
		env.GetBatcherConfig(batcherConfig))
	require.NoError(t, err)

	l1Client := system.NodeClient(e2esys.RoleL1)
	espClient := espressoDevNode.Client()

	deployerTransactor, err := bind.NewKeyedTransactorWithChainID(system.Config().Secrets.Deployer, system.Cfg.L1ChainIDBig())
	require.NoError(t, err)

	batchAuthenticator, err := bindings.NewBatchAuthenticator(system.RollupConfig.BatchAuthenticatorAddress, l1Client)
	require.NoError(t, err)

	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	// Send Transaction on L1, and wait for verification on the L2 Verifier
	env.RunSimpleL1TransferAndVerifier(ctx, t, system)

	// Verify everything works
	env.RunSimpleL2Burn(ctx, t, system)

	// Stop the "TEE" batcher
	err = system.BatchSubmitter.TestDriver().StopBatchSubmitting(ctx)
	require.NoError(t, err)

	// Switch active batcher to the fallback (non-Espresso) path
	tx, err := batchAuthenticator.SetActiveIsEspresso(deployerTransactor, false)
	require.NoError(t, err)
	_, err = wait.ForReceiptOK(ctx, l1Client, tx.Hash())
	require.NoError(t, err)

	// Start the fallback batcher
	err = system.FallbackBatchSubmitter.TestDriver().StartBatchSubmitting()
	require.NoError(t, err)

	// Everything should still work
	env.RunSimpleL2Burn(ctx, t, system)

	// Stop the fallback batcher
	err = system.FallbackBatchSubmitter.TestDriver().StopBatchSubmitting(ctx)
	require.NoError(t, err)

	// Switch batcher back to the "TEE" (Espresso) batcher
	tx, err = batchAuthenticator.SetActiveIsEspresso(deployerTransactor, true)
	require.NoError(t, err)
	switchReceipt, err := wait.ForReceiptOK(ctx, l1Client, tx.Hash())
	require.NoError(t, err)

	// Give things time to settle
	l2Height, err := waitForRollupToMovePastL1Block(ctx, system.RollupClient(e2esys.RoleVerif), switchReceipt.BlockNumber.Uint64())
	require.NoError(t, err)

	espHeight, err := espClient.FetchLatestBlockHeight(ctx)
	require.NoError(t, err)

	// Start a new "TEE" batcher
	// Use moderate channel settings so the new batcher submits batches promptly without posting
	// every L1 block.
	batcherConfig.MaxChannelDuration = 10
	batcherConfig.TargetNumFrames = 1
	batcherConfig.MaxL1TxSize = 120_000
	batcherConfig.Espresso.CaffeinationHeightEspresso = espHeight
	batcherConfig.Espresso.CaffeinationHeightL2 = l2Height
	batcherCtx, cancelBatcher := context.WithCancelCause(ctx)
	defer cancelBatcher(nil)
	newBatcher, err := batcher.BatcherServiceFromCLIConfig(batcherCtx, cancelBatcher, "0.0.1", batcherConfig, system.BatchSubmitter.Log)
	require.NoError(t, err)
	err = newBatcher.Start(batcherCtx)
	require.NoError(t, err)

	// Everything should still work (use longer timeout after batcher switch)
	env.RunSimpleL2BurnWithTimeout(ctx, t, system, 5*time.Minute)
}

// TxManagerIntercept is a txmgr.TxManager that wraps another txmgr.TxManager
// and intercepts calls to Send and SendAsync.
//
// The purpose of this intercept is to simulate a failure in the tx submission
// process such that when activated a single frame of a multi-frame channel is
// sent to L1, and the remaining frames fail to be sent, triggering the fallback
// batcher to take over.
type TxManagerIntercept struct {
	txmgr.TxManager
	sync.Mutex

	// shouldFail indicates whether to simulate a failure on Send/SendAsync.
	shouldFail bool

	// triggerAfterOne indicates whether to start failing after a single
	// successful Send/SendAsync.
	triggerAfterOne bool

	// failureCount tracks the number of failures that have occurred.
	failureCount int

	successfulFrames   map[derive.ChannelID][]derive.Frame
	unsuccessfulFrames map[derive.ChannelID][]derive.Frame
}

func NewTxManagerIntercept(base txmgr.TxManager) *TxManagerIntercept {
	return &TxManagerIntercept{
		TxManager:          base,
		successfulFrames:   make(map[derive.ChannelID][]derive.Frame),
		unsuccessfulFrames: make(map[derive.ChannelID][]derive.Frame),
	}
}

// ErrSimulatedTxSubmissionFailure is the sentinel error returned when a
// simulated tx submission failure is triggered.
//
// We utilize this as a placeholder error to indicate that the tx submission
// failure was intentional for testing purposes.
var ErrSimulatedTxSubmissionFailure = errors.New("simulated tx submission failure")

// decodeFrameInformation takes a txmgr.TxCandidate and attempts to decode
// frames contained within either the Blob fields, or the TxData field.
func decodeFrameInformation(candidate txmgr.TxCandidate) ([]derive.Frame, error) {
	if len(candidate.TxData) > 0 {
		// We have a CallData tx, so we can decode the frame information from
		// the tx data.
		return decodeFrameInformationFromTxData(candidate)
	}

	if len(candidate.Blobs) > 0 {
		// We have a Blob tx, so we can decode the frame information from
		// the blobs.
		return decodeFrameInformationFromBlobs(candidate)
	}

	return nil, fmt.Errorf("tx candidate has neither tx data nor blobs to decode frame information from")
}

// decodeFrameInformationFromData takes a byte slice and decodes each frame
// until it can no longer decode any frames. It returns a slice of all
// decoded frames, and any error encountered.
func decodeFrameInformationFromData(data []byte) ([]derive.Frame, error) {
	if data[0] != params.DerivationVersion0 {
		// Not a supported derivation version
		return nil, fmt.Errorf("unsupported derivation version: %d", data[0])
	}

	var frames []derive.Frame
	reader := bytes.NewBuffer(data[1:])
	for {
		var frame derive.Frame
		err := frame.UnmarshalBinary(reader)
		if errors.Is(err, io.EOF) {
			// We've consumed all of the frames.
			break
		}

		// If this is any other error, it indicates that there was an
		// error decoding the frame.
		if err != nil {
			return frames, fmt.Errorf("error decoding frame: %w", err)
		}

		frames = append(frames, frame)
	}

	return frames, nil
}

// decodeFrameInformationFromTxData takes a txmgr.TxCandidate and will assume
// that the frame data is encoded within the TxData. This data will be taken
// and decoded into frames and returned.
func decodeFrameInformationFromTxData(candidate txmgr.TxCandidate) ([]derive.Frame, error) {
	data := candidate.TxData

	return decodeFrameInformationFromData(data)
}

// decodeFrameInformationFromBlobs() takes a txmgr.TxCandidate and will assume
// that the frame data is encoded within the Blobs.  The blobs will be
// converted back to txData, and the data will be decoded into frames.
func decodeFrameInformationFromBlobs(candidate txmgr.TxCandidate) ([]derive.Frame, error) {
	var frames []derive.Frame
	for _, blob := range candidate.Blobs {
		data, err := blob.ToData()
		if err != nil {
			return frames, fmt.Errorf("error converting blob to data: %w", err)
		}

		newFrames, err := decodeFrameInformationFromData(data)
		if err != nil {
			return frames, err
		}
		frames = append(frames, newFrames...)
	}

	return frames, nil
}

func (t *TxManagerIntercept) markFramesAsSuccessful(frames []derive.Frame) {
	t.Lock()
	defer t.Unlock()
	for _, frame := range frames {
		t.successfulFrames[frame.ID] = append(t.successfulFrames[frame.ID], frame)
	}
}

func (t *TxManagerIntercept) markFramesAsUnsuccessful(frames []derive.Frame) {
	t.Lock()
	defer t.Unlock()
	for _, frame := range frames {
		t.unsuccessfulFrames[frame.ID] = append(t.unsuccessfulFrames[frame.ID], frame)
	}
}

// Send implements txmgr.TxManager.
//
// Send is overridden to simulate a failure when shouldFail is true, and to
// allow for one final transaction to be sent before failures begin when
// triggerAfterOne is true.
func (t *TxManagerIntercept) Send(ctx context.Context, candidate txmgr.TxCandidate) (*types.Receipt, error) {
	frames, err := decodeFrameInformation(candidate)
	if err != nil {
		// Not a batch frame transaction (e.g. authenticateBatch call) — pass through
		return t.TxManager.Send(ctx, candidate)
	}

	if t.shouldFail {
		t.failureCount++
		t.markFramesAsUnsuccessful(frames)
		time.Sleep(50 * time.Millisecond) // Simulate some delay
		return nil, ErrSimulatedTxSubmissionFailure
	}

	if t.triggerAfterOne {
		t.shouldFail = true
	}

	t.markFramesAsSuccessful(frames)

	return t.TxManager.Send(ctx, candidate)
}

// SendAsync implements txmgr.TxManager.
//
// SendAsync is overridden to simulate a failure when shouldFail is true, and
// to allow for one final transaction to be sent before failures begin when
// triggerAfterOne is true.
func (t *TxManagerIntercept) SendAsync(ctx context.Context, candidate txmgr.TxCandidate, ch chan txmgr.SendResponse) {
	frames, err := decodeFrameInformation(candidate)
	if err != nil {
		// Not a batch frame transaction (e.g. authenticateBatch call) — pass through
		t.TxManager.SendAsync(ctx, candidate, ch)
		return
	}

	if t.shouldFail {
		t.failureCount++
		t.markFramesAsUnsuccessful(frames)
		time.Sleep(50 * time.Millisecond) // Simulate some delay
		ch <- txmgr.SendResponse{Err: ErrSimulatedTxSubmissionFailure}
		return
	}

	if t.triggerAfterOne {
		t.shouldFail = true
	}

	t.markFramesAsSuccessful(frames)
	t.TxManager.SendAsync(ctx, candidate, ch)
}

type partialFrameData struct {
	channelID          derive.ChannelID
	successfulFrames   []derive.Frame
	unsuccessfulFrames []derive.Frame
}

func (t *TxManagerIntercept) partialFrameData() []partialFrameData {
	var partials []partialFrameData

	for channelID, unsuccessfulFrames := range t.unsuccessfulFrames {
		successfulFrames, ok := t.successfulFrames[channelID]
		if !ok {
			continue
		}

		partials = append(partials, partialFrameData{
			channelID:          channelID,
			successfulFrames:   successfulFrames,
			unsuccessfulFrames: unsuccessfulFrames,
		})
	}

	return partials
}

// Compile time assertion to ensure TxManagerIntercept implements
// txmgr.TxManager.
var _ txmgr.TxManager = (*TxManagerIntercept)(nil)

// retryWaitNTimes retries the given function up to n times until it
// succeeds.
func retryWaitNTimes(fn func() error, n int) error {
	var lastErr error
	for range n {
		lastErr = fn()
		if lastErr == nil {
			break
		}
	}

	return lastErr
}

// TestFallbackMechanismIntegrationTestChannelNotClosed is a test case that is
// meant to verify the correct expected behavior in the event that the Espresso
// Batcher encounters an error mid L1 Batch submission that prevents the full
// channel from being submitted to the L1.
//
// In this scenario this issue is expected to send a single frame of a
// multi-frame channel to the contract. At this point the batch should be
// switched to the fallback and the fallback batcher should continue
// submitting the remaining frames of the channel without any issues.
func TestFallbackMechanismIntegrationTestChannelNotClosed(t *testing.T) {
	ctx, cancel := context.WithTimeoutCause(context.Background(), time.Minute*10, fmt.Errorf("test did not complete within expected time allotment: %w", context.DeadlineExceeded))
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	// In order to force a multi-frame channel with the e2e system setup,
	// we need to multimately modify the channel config that will be utilized
	// by the batcher.
	//
	// This may seem a bit convoluted, but we have to contend with a few
	// different settings in order to ensure that the behavior we are
	// targeting is achieved.
	//
	// All of the options given below are utilized with the specific purpose
	// of triggering multi-frame channels.
	//
	// NOTE: Some of the configuration options pull double-duty.  They are
	// utilized in both the creation of the frames, and the sending of the
	// frames.  In both scenarios, they may behave differently. I will make
	// an effor to note them where they occur.

	system, espressoDevNode, err := launcher.StartE2eDevnet(
		ctx,
		t,
		// Make Sure that the Batcher does not start Running
		env.WithBatcherStoppedInitially(),

		// Explicitly disable using any sort of compression.  This is
		// necessary as we will be specifying that we will be targeting
		// a specific frame size, and we don't want compression to indirectly
		// interfere with this process.
		env.WithBatcherCompressor(compressor.NoneKind),

		// This sets the Target Number of Frames that each channel is aiming
		// to achieve. In this case we specify 3 so that we can ensure that
		// our channel will always be a multi frame channel.
		//
		// Coupling this with the max channel duration helps us to ensure that
		// each channel will always aim for the same number of channels.
		//
		// NOTE: This has different behavior when constructing the frames on
		// the Batcher preparation than it does on the L1 submission.
		// Specifically concerning the Da Type DaTypeCalldata.  When utilizing
		// call data, the L1 Submission will **ALWAYS** utilize 1 frame instead
		// of this passed value.  Yet channel construction will utilize this
		// provided value as appropriate.
		env.WithBatcherTargetNumFrames(3),

		// We set the MaxL1TxSize to some value that will hold our L2
		// Transaction size comfortably.
		env.WithBatcherMaxL1TxSize(1200),

		// We set the MaxChannelDuration to 0 specifically to disable premature
		// channel closing before we have enough frames.  The default behavior
		// is to create a new Channel with the specified window of L1 Blocks.
		// The idea is that you can prevent eager channel production when
		// you are not producing blocks with transactions.
		//
		// Setting this to 0 explicitly disables the feature, and as a result
		// it will only send the data when the previous conditions are met.
		env.WithBatcherMaxChannelDuration(0),
	)

	require.NoError(t, err)
	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	// We create an intercept around the existing tx manager so we have
	// control over when our failures start to occur.

	interceptTxManager := NewTxManagerIntercept(
		system.BatchSubmitter.TxManager,
	)

	// Replace the existing TxManager with our intercept
	system.BatchSubmitter.TestDriver().Txmgr = interceptTxManager

	// Start the Batcher again, so the publishingLoop picks up the TxManager
	// when creating its queue.
	err = system.BatchSubmitter.TestDriver().StartBatchSubmitting()
	require.NoError(t, err)

	// Wait for the Next L2 Block to be verified by ensure everything is
	// working and progressing without issue
	err = wait.ForProcessingFullBatch(ctx, system.RollupClient(e2esys.RoleVerif))
	require.NoError(t, err)

	l2Seq := system.NodeClient(e2esys.RoleSeq)
	l1Client := system.NodeClient(e2esys.RoleL1)
	l2Verif := system.NodeClient(e2esys.RoleVerif)

	// Let's make sure that the system is progressing initially for both
	// the Sequencer, the Verifier, and the L1Node
	err = wait.ForBlock(ctx, l2Seq, 3)
	require.NoError(t, err)
	err = retryWaitNTimes(func() error {
		return wait.ForNextBlock(ctx, l2Verif)
	}, 3)
	require.NoError(t, err)

	// Verify everything works
	env.RunSimpleL2Burn(ctx, t, system)

	// We want to trigger the failure mode now.
	interceptTxManager.triggerAfterOne = true

	// Now we need to submit a multi-frame channel to L1 to trigger the
	// failure. We can do this by adjusting the batcher config to use a very
	// small MaxL1TxSize such that even a small L2 transaction will result in
	// multiple frames.

	// We want enough L2 Transactions to ensure we have multiple frames.
	const n = 10

	receipts, err := env.RunSimpleMultiTransactions(ctx, t, system, n)
	require.NoError(t, err)

	// We want to wait until we know that the intercept tx manager has
	// trigger the failure mode successfully, and that all n transactions
	// have been attempted.

	// Wait until at least 2 L2 blocks have been mined (one for the
	// a block with successful frames, and one for a block with failed frames).
	err = wait.ForNextBlock(ctx, l2Seq)
	require.NoError(t, err)
	err = wait.ForNextBlock(ctx, l2Seq)
	require.NoError(t, err)

	// Make sure that at least one failure has occurred, as this should
	// indicate that the submission process should have failed a multiframe
	// channel submission.
	err = wait.For(ctx, 10*time.Second, func() (bool, error) {
		return interceptTxManager.failureCount >= 1, nil
	})
	require.NoError(t, err)

	// Stop the "TEE" batcher
	err = system.BatchSubmitter.TestDriver().StopBatchSubmitting(ctx)
	require.NoError(t, err)

	// Switch active batcher
	options, err := bind.NewKeyedTransactorWithChainID(system.Config().Secrets.Deployer, system.Cfg.L1ChainIDBig())
	require.NoError(t, err)

	batchAuthenticator, err := bindings.NewBatchAuthenticator(system.RollupConfig.BatchAuthenticatorAddress, l1Client)
	require.NoError(t, err)

	tx, err := batchAuthenticator.SetActiveIsEspresso(options, false)
	require.NoError(t, err)
	_, err = wait.ForReceiptOK(ctx, l1Client, tx.Hash())
	require.NoError(t, err)

	// Start the fallback batcher
	err = system.FallbackBatchSubmitter.TestDriver().StartBatchSubmitting()
	require.NoError(t, err)

	// There should be some failure recorded in the intercept tx manager that
	// has a corresponding success in the intercept tx manager.

	partialFrameData := interceptTxManager.partialFrameData()
	require.Greaterf(t, len(partialFrameData), 0, "expected to find at least one partially submitted frame")

	// Verify that our previous receipts were also recorded on L1
	for i, receipt := range receipts {
		_, err := wait.ForReceiptOK(ctx, l2Verif, receipt.TxHash)
		require.NoError(t, err, "failed to find receipt %d for tx %s on L2 Verifier", i, receipt.TxHash)
	}

	// Everything should still work
	env.RunSimpleL2Burn(ctx, t, system)
}
