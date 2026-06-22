package environment_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/espresso/bindings"
	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/require"
)

// TestEspressoEnforcementHardfork exercises the EspressoTime
// hardfork transition while the chain is running on the fallback batcher.
//
// Pre-fork the fallback batcher posts plain BatchInbox txs (no
// BatchAuthenticator interaction) and the verifier accepts them via upstream
// sender-based authorization. Across the boundary the batcher's gate flips
// (driven by `isFallbackAuthRequired`'s lead-time mechanism) and it starts
// calling `authenticateBatchInfo`; the verifier post-fork requires the
// resulting `BatchInfoAuthenticated` events. The lead time prevents a window
// where the batcher omits authentication while the verifier requires it.
//
// `activeIsEspresso` is flipped to false before Phase 1 (modeling a chain
// that experienced a fallback-batcher event before the hardfork) and back to
// true before Phase 3 so the subsequently-started TEE batcher is the
// on-chain active batcher.
func TestEspressoEnforcementHardfork(t *testing.T) {
	// 5 minutes covers Espresso devnet startup plus pre-fork test ops.
	const enforcementOffset = 5 * time.Minute
	// Smaller than enforcementOffset so the batcher actually exercises the
	// pre-fork no-auth path, but >> L1BlockTime to absorb inclusion delay.
	const fallbackAuthLeadTime = 30 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Minute)
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	// Captured for the post-fork TEE batcher restart.
	espressoBatcherConfig := &batcher.CLIConfig{}

	// The batcher-config options run before GetBatcherConfig so the snapshot it
	// takes into espressoBatcherConfig reflects them. Small frames + a long
	// channel duration force multi-frame channels split across L1 blocks, which
	// makes a batch tx land in an L1 block at/after the fork boundary likely so
	// the lead-time auth gate is actually exercised.
	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t,
		env.WithEspressoEnforcementOffset(enforcementOffset),
		env.WithFallbackAuthLeadTime(fallbackAuthLeadTime),
		env.WithL1FinalizedDistance(0),
		env.WithSequencerUseFinalized(true),
		env.WithBatcherStoppedInitially(),
		env.WithBatcherTargetNumFrames(10),
		env.WithBatcherMaxL1TxSize(250),
		env.WithBatcherMaxChannelDuration(1000),
		// Unbounded pending L1 txs so the Espresso auth+batch pairs (routed
		// through the ordered txmgr queue) publish concurrently instead of
		// one-per-L1-block; otherwise L1 data availability lags far behind the
		// sequencer and the verifier cannot derive recent blocks within the
		// test's confirmation windows.
		env.WithBatcherMaxPendingTransactions(0),
		env.GetBatcherConfig(espressoBatcherConfig),
	)
	require.NoError(t, err)
	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	require.NotNil(t, system.RollupConfig.EspressoTime,
		"test requires EspressoTime to be set on the rollup config")
	forkTime := *system.RollupConfig.EspressoTime
	nowSec := uint64(time.Now().Unix())
	require.Greater(t, forkTime, nowSec, "fork timestamp must be in the future at test start")
	t.Logf("EspressoTime = %d (now = %d, in %ds)", forkTime, nowSec, forkTime-nowSec)

	l1Client := system.NodeClient(e2esys.RoleL1)
	verifClient := system.NodeClient(e2esys.RoleVerif)
	verifRollup := system.RollupClient(e2esys.RoleVerif)
	espClient := espressoDevNode.Client()

	deployerTransactor, err := bind.NewKeyedTransactorWithChainID(
		system.Config().Secrets.Deployer, system.Cfg.L1ChainIDBig())
	require.NoError(t, err)
	batchAuthenticator, err := bindings.NewBatchAuthenticator(
		system.RollupConfig.BatchAuthenticatorAddress, l1Client)
	require.NoError(t, err)

	activeIsEspresso, err := batchAuthenticator.ActiveIsEspresso(nil)
	require.NoError(t, err)
	require.True(t, activeIsEspresso, "BatchAuthenticator default should be activeIsEspresso=true")

	// Flip to fallback before Phase 1 so the fallback batcher's
	// `isBatcherActive` check (consulted post-lead-time) sees itself as
	// active and continues publishing across the boundary.
	switchTx, err := batchAuthenticator.SetActiveIsEspresso(deployerTransactor, false)
	require.NoError(t, err)
	_, err = wait.ForReceiptOK(ctx, l1Client, switchTx.Hash())
	require.NoError(t, err)
	activeIsEspresso, err = batchAuthenticator.ActiveIsEspresso(nil)
	require.NoError(t, err)
	require.False(t, activeIsEspresso, "first SetActiveIsEspresso(false) should set activeIsEspresso=false")

	// Phase 1 (pre-fork): fallback batcher publishes plain BatchInbox txs.
	require.NoError(t, system.FallbackBatchSubmitter.TestDriver().StartBatchSubmitting())
	env.RunSimpleL1TransferAndVerifier(ctx, t, system)
	env.RunSimpleL2Burn(ctx, t, system)

	status, err := verifRollup.SyncStatus(ctx)
	require.NoError(t, err)
	require.Less(t, status.SafeL2.Time, forkTime, "verifier should still be pre-fork after initial L2 burn")

	// Phase 2 (boundary): fallback batcher continues across forkTime,
	// switching to authenticateBatchInfo at forkTime - leadTime.
	preBoundaryL1Head, err := l1Client.BlockNumber(ctx)
	require.NoError(t, err)
	require.NoError(t, wait.ForBlockWithTimestamp(ctx, l1Client, forkTime))
	t.Logf("L1 reached fork timestamp; pre-boundary L1 head was %d", preBoundaryL1Head)

	// A stall here would indicate dropped batches (unauthenticated tx in a
	// post-fork L1 block) — i.e. lead-time too short.
	require.NoError(t, wait.ForBlockWithTimestamp(ctx, verifClient, forkTime+30))
	require.NoError(t, wait.ForNextBlock(ctx, l1Client))
	require.NoError(t, wait.ForNextBlock(ctx, l1Client))

	status, err = verifRollup.SyncStatus(ctx)
	require.NoError(t, err)
	require.Greater(t, status.SafeL2.Time, forkTime,
		"verifier safe head should have advanced past forkTime (lead-time too short?)")

	postBoundaryL1Head, err := l1Client.BlockNumber(ctx)
	require.NoError(t, err)

	// Confirm the lead-time path actually executed.
	authIter, err := batchAuthenticator.FilterBatchInfoAuthenticated(&bind.FilterOpts{
		Context: ctx,
		Start:   preBoundaryL1Head,
		End:     &postBoundaryL1Head,
	}, nil)
	require.NoError(t, err)
	defer authIter.Close()
	authEventCount := 0
	for authIter.Next() {
		authEventCount++
	}
	require.NoError(t, authIter.Error())
	require.Greater(t, authEventCount, 0,
		"fallback batcher should have emitted >=1 BatchInfoAuthenticated event around the boundary")
	t.Logf("fallback batcher emitted %d BatchInfoAuthenticated events around the boundary", authEventCount)

	// Phase 3: stop fallback, flip back to TEE, start TEE batcher.
	require.NoError(t, system.FallbackBatchSubmitter.TestDriver().StopBatchSubmitting(ctx))

	switchTx, err = batchAuthenticator.SetActiveIsEspresso(deployerTransactor, true)
	require.NoError(t, err)
	_, err = wait.ForReceiptOK(ctx, l1Client, switchTx.Hash())
	require.NoError(t, err)
	activeIsEspresso, err = batchAuthenticator.ActiveIsEspresso(nil)
	require.NoError(t, err)
	require.True(t, activeIsEspresso, "second SetActiveIsEspresso(true) should set activeIsEspresso=true")

	// Stream from the live head, not from genesis.
	l2Height, err := waitForRollupToMovePastL1Block(ctx, verifRollup, status.CurrentL1.Number)
	require.NoError(t, err)
	espHeight, err := espClient.FetchLatestBlockHeight(ctx)
	require.NoError(t, err)

	espressoBatcherConfig.MaxChannelDuration = 10
	espressoBatcherConfig.TargetNumFrames = 1
	espressoBatcherConfig.MaxL1TxSize = 120_000
	// Caffeinate at espHeight-1 (last already-sealed block) so the streamer reads from
	// espHeight inclusive and picks up the batches this batcher re-submits there.
	espressoBatcherConfig.Espresso.CaffeinationHeightEspresso = espHeight - 1
	espressoBatcherConfig.Espresso.CaffeinationHeightL2 = l2Height
	// Inherited Stopped=true from WithBatcherStoppedInitially.
	espressoBatcherConfig.Stopped = false

	teeCtx, teeCancel := context.WithCancelCause(ctx)
	defer teeCancel(nil)
	teeBatcher, err := batcher.BatcherServiceFromCLIConfig(
		teeCtx, teeCancel, "0.0.1", espressoBatcherConfig, system.BatchSubmitter.Log,
		batcher.WithEspressoClientOverride(system.EspressoClient))
	require.NoError(t, err)
	require.NoError(t, teeBatcher.Start(teeCtx))

	// Phase 4 (post-fork): TEE batcher takes over, verifier keeps advancing.
	env.RunSimpleL2BurnWithTimeout(ctx, t, system, 5*time.Minute)

	status, err = verifRollup.SyncStatus(ctx)
	require.NoError(t, err)
	require.Greater(t, status.SafeL2.Time, forkTime, "verifier should have advanced post-fork")
}
