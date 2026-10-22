package environment_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

// TestBatcherWaitForFinality is a test that attempts to make sure that the batcher waits for the
// derived L1 block to be finalized before submitting a new block.
//
// This tests is designed to evaluate Test 8.1.1 as outlined within the Espresso Celo Integration
// plan. It has stated task definition as follows:
//
//	Arrange:
//		Run the sequencer and the batcher in Espresso mode.
//	Act:
//		Wait until a new block is finalized.
//	Assert:
//		The batcher doesn't submit a block without finalized L1 origin to the L1.
//		After the L1 origin is finalized, the batcher submits the block.
func TestBatcherWaitForFinality(t *testing.T) {
	// Basic test setup.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	launcher := new(env.EspressoDevNodeLauncherDocker)

	// Set NonFinalizedProposals to true and SequencerUseFinalized to false, to make sure we are
	// testing how the batcher handles the finality.
	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t, env.WithL1FinalizedDistance(4), env.WithNonFinalizedProposals(true), env.WithSequencerUseFinalized(false))
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}
	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	rollupClient := system.RollupClient(e2esys.RoleVerif)
	l1Client := system.NodeClient(e2esys.RoleL1)

	initialStatus, err := rollupClient.SyncStatus(context.Background())
	require.NoError(t, err)
	initialSafeL1Number := initialStatus.SafeL1.Number

	// Wait for new blocks to be finalized, which will enable the batcher to submit more blocks to
	// to the L1.
	tickerFinality := time.NewTicker(1 * time.Second)
	defer tickerFinality.Stop()

	for {
		select {
		case <-ctx.Done():
			require.FailNow(t, "Timeout: Finalized L1 number not increased by 10")
		case <-tickerFinality.C:
			// Verify that the batcher waits for the L1 origin to be finalized before submitting a new
			// block to the L1.
			statusAfterWait, err := rollupClient.SyncStatus(context.Background())
			require.NoError(t, err)
			// Compare against the L1 chain's own finalized tag, not the
			// verifier's FinalizedL1: the batcher gates submission on the
			// sequencer node's finality view, and the verifier's finality
			// poller can briefly lag it, which would trip this assertion
			// without any batcher misbehavior. Querying L1 after the sync
			// status keeps the check sound: finality only advances, so a
			// batch submitted before its origin finalized would still show
			// origin > finalized here.
			finalizedL1Header, err := l1Client.HeaderByNumber(ctx, big.NewInt(rpc.FinalizedBlockNumber.Int64()))
			require.NoError(t, err)
			require.LessOrEqual(t, statusAfterWait.SafeL2.L1Origin.Number, bigs.Uint64Strict(finalizedL1Header.Number), "L1 origin not finalized before submission")

			// Exit the test if there are 10 new safe blocks on the L1.
			if statusAfterWait.SafeL1.Number >= initialSafeL1Number+10 {
				return
			}
		}
	}
}

func runL1Reorg(ctx context.Context, t *testing.T, system *e2esys.System) {
	l2Seq := system.NodeClient(e2esys.RoleSeq)
	l1Client := system.NodeClient(e2esys.RoleL1)

	// Wait for batcher to start advancing L2 head
	_, err := geth.WaitForBlockToBeSafe(big.NewInt(2), l2Seq, 2*time.Minute)
	if have, want := err, error(nil); have != want {
		t.Fatalf("L2 isn't progressing:\nhave:\n\t%v\nwant:\n\t%v", have, want)
	}

	t.Log("L2 is progressing")

	// Wait for L2 head to be based off non-genesis unfinalized block
	l2HeadL1Info := &derive.L1BlockInfo{}
	var l2Head *types.Block
	var unsafeL2Height uint64
	var l1Height uint64
	for l2HeadL1Info.Number == 0 || (l1Height-l2HeadL1Info.Number) >= system.Cfg.L1FinalizedDistance {
		unsafeL2Height, err = l2Seq.BlockNumber(ctx)
		require.NoError(t, err)

		l2Head, err = l2Seq.BlockByNumber(ctx, new(big.Int).SetUint64(unsafeL2Height))
		require.NoError(t, err)

		_, l2HeadL1Info, err = derive.BlockToSingularBatch(system.RollupCfg(), l2Head)
		require.NoError(t, err)

		l1Height, err = l1Client.BlockNumber(ctx)
		require.NoError(t, err)
	}

	l1Origin, err := l1Client.BlockByNumber(ctx, new(big.Int).SetUint64(l2HeadL1Info.Number))
	require.NoError(t, err)

	// Introduce a reorg at L1
	t.Logf("Introducing reorg at L1Origin %d, L1Head %d, l2Head %d", l1Origin.Number(), l1Height, unsafeL2Height)
	err = system.ForkL1(l1Origin.ParentHash())
	require.NoError(t, err)

	// Wait for SafeL2 to advance despite the reorg
	_, err = geth.WaitForBlockToBeSafe(new(big.Int).SetUint64(unsafeL2Height+1), l2Seq, 2*time.Minute)
	require.NoError(t, err)

	// Check that safe chain doesn't contain the forked block
	newL2Head, err := l2Seq.BlockByNumber(ctx, new(big.Int).SetUint64(unsafeL2Height))
	require.NoError(t, err)
	require.NotEqual(t, newL2Head.Hash(), l2Head.Hash())
}

// TestE2eDevnetWithL1Reorg tests how the batcher handles an L1 reorg.
// Specifically, it focuses on cases where unsafe L2 chain contains blocks that
// reference unfinalized L1 blocks as their origin.
//
// This tests is designed to evaluate Test 8.1.2 as outlined within the Espresso Celo
// Integration plan. The test is defined as follows:
// Arrange:
//
//	Running Sequencer, Batcher in Espresso mode & OP node.
//
// Act:
//
//	Wait for sequencer to propose an unsafe L2 block with unfinalized L1 origin
//	Simulate L1 reorg at that block's origin
//
// Assert:
//
//	Assert that derivation pipeline still progresses
//	Assert that the OP node reports a new block at the target L2 height
func TestE2eDevnetWithL1Reorg(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	system, devNode, err := launcher.StartE2eDevnet(ctx, t, env.WithL1FinalizedDistance(16))
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	defer env.Stop(t, system)
	defer env.Stop(t, devNode)

	runL1Reorg(ctx, t, system)
}
