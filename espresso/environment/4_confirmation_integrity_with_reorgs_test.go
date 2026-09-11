package environment_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Computes the hash of the content of a batch. Introduced for testing purposes only.
// @param b batch
// @return string containing the hash of a batch
func BatchHash(b *derive.SingularBatch) string {

	// Concatenate the transactions and other relevant metadata of the batch
	str := ""

	for _, tx := range b.Transactions {
		str = str + tx.String()
	}
	str = str + b.EpochHash.String()
	str = str + strconv.Itoa(int(b.Timestamp))
	str = str + b.ParentHash.String()

	h := sha256.New()
	h.Write([]byte(str))
	hash := h.Sum(nil)

	res := hex.EncodeToString(hash)

	return res
}

// This function computes a list where several batches are sent to unfinalized L1 blocks.
// It works as follows:
//  1. Pick the current L1 block number. Store it so that later we can reorg back to this point
//  2. Start from the height of the first L2 block unsafe block.
//  3. Pick the blocks from this height and the subsequent.
//     This first block is initially unsafe, but we wait for it to become safe inside the L1 finality window and likewise for the others.
//  4. Do this until the "finality" window closes i.e. before the number of new L1 blocks is bigger than L1FinalizedDistance.
//  5. While doing this also send transactions to the sequencer to avoid dealing with empty blocks.
//  6. Return the list of batch hashes, the L1 block number to reorg to and also the L2 block height determined in step 2.
//     @param ctx,t,system standard parameters to execute a test
//     @param l1Client L1 client used to fetch L1 block numbers
//     @param l2Seq sequencer used to send transactions
//     @param l2Verif OP node we monitor for unsafe blocks
//     @return batches list of hashes of the batches
//     @return l1HeightStart L1 block number collected in step 1.
//     @return unsafeL2BlockNumber L2 block number collected in step 2.
func collectBatchesPublishedOnUnfinalizedL1Blocks(ctx context.Context, t *testing.T, system *e2esys.System, l1Client *ethclient.Client, l2Seq *ethclient.Client, l2Verif *ethclient.Client) ([]string, uint64, uint64) {
	var batches []string

	l1Height, err := l1Client.BlockNumber(ctx)
	require.NoError(t, err)
	l1HeightStart := l1Height
	log.Info("L1 height to reorg to", "height", l1HeightStart)
	// Keep monitoring L2 blocks while L1 is producing unfinalized blocks
	i := int64(0)
	// Fetch height of the most recent block which is unsafe and which batch will be sent to L1 at a later stage
	rollupClient := system.RollupClient(e2esys.RoleVerif)

	status, err := rollupClient.SyncStatus(ctx)
	require.NoError(t, err)
	unsafeL2BlockNumber := status.SafeL2.Number + 1

	nonce := uint64(0)
	addressAlice := system.Cfg.Secrets.Addresses().Alice

	for (l1Height - l1HeightStart) < system.Cfg.L1FinalizedDistance {
		height := uint64(i) + unsafeL2BlockNumber

		//Send some transactions to fill the batches
		receipt := helpers.SendL2TxWithID(t, system.Cfg.L2ChainIDBig(), l2Seq, system.Cfg.Secrets.Bob, func(opts *helpers.TxOpts) {
			opts.Nonce = nonce
			opts.ToAddr = &addressAlice
			opts.Value = new(big.Int).SetUint64(1)
		})
		nonce++
		log.Info("Receipt", "value", receipt)

		l2Head, err := geth.WaitForBlockToBeSafe(new(big.Int).SetUint64(height), l2Verif, 10*time.Second)
		require.NoError(t, err)

		if err == nil { // Insert new batch in the list

			batch, l2HeadL1Info, err := derive.BlockToSingularBatch(system.RollupCfg(), l2Head)
			require.NoError(t, err)
			log.Info("l2HeadL1Info", "value", l2HeadL1Info)

			batchHash := BatchHash(batch)

			t.Log("New element inserted", "value", batchHash, "list length", len(batches))
			batches = append(batches, batchHash)

			i++
		}

		l1Height, err = l1Client.BlockNumber(ctx)
		require.NoError(t, err)
	}

	return batches, l1HeightStart, unsafeL2BlockNumber
}

// This collect the first N L2 blocks from a specific height. Collected blocks are guaranteed to be safe
// @param ctx,t,system standard parameters to execute a test.
// @param l2Verif OP node to fetch the safe blocks.
// @apram n number of blocks to fetch.
// @param startIndex initial height of the L2 chain to start fetching blocks from.
func collectFirstNL2SafeBlocks(ctx context.Context, t *testing.T, system *e2esys.System, l2Verif *ethclient.Client, n int, startIndex uint64) []string {
	var batches []string

	for i := 0; i < n; i++ {
		height := startIndex + uint64(i)
		_, err := geth.WaitForBlockToBeSafe(big.NewInt(int64(height)), l2Verif, 2*time.Minute)
		require.NoError(t, err)

		l2Head, err := l2Verif.BlockByNumber(ctx, new(big.Int).SetUint64(height))
		require.NoError(t, err)

		batch, _, err := derive.BlockToSingularBatch(system.RollupCfg(), l2Head)
		require.NoError(t, err)
		batchHash := BatchHash(batch)
		batches = append(batches, batchHash)

	}
	return batches
}

// Main logic of the test:
// 1. Collect some unsafe batches in list L.
// 2. Do the reorg.
// 3. Collect the batches into list L'.
// 4. Check that L=L'.
func run(ctx context.Context, t *testing.T, system *e2esys.System) {
	l2Seq := system.NodeClient(e2esys.RoleSeq)
	l2Verif := system.NodeClient(e2esys.RoleVerif)
	l1Client := system.NodeClient(e2esys.RoleL1)

	var unsafeL2Height uint64
	var l1Height uint64

	// Wait for batcher to start advancing L2 head
	_, err := geth.WaitForBlockToBeSafe(big.NewInt(2), l2Seq, 2*time.Minute)
	require.NoError(t, err, "L2 isn't progressing as expected")

	t.Log("L2 is progressing")

	// Fetch batches before reorg
	batchesBefore, L1BlockHeightToReorgTo, startIndex := collectBatchesPublishedOnUnfinalizedL1Blocks(ctx, t, system, l1Client, l2Seq, l2Verif)

	l1Origin, err := l1Client.BlockByNumber(ctx, new(big.Int).SetUint64(L1BlockHeightToReorgTo))
	require.NoError(t, err)

	log.Info("+++ L2 blocks before reorg", "value", batchesBefore)
	// Introduce a reorg at L1
	l1Height, err = l1Client.BlockNumber(ctx)
	require.NoError(t, err)
	t.Logf("Introducing reorg at L1Origin %d, L1Head %d, l2Head %d", l1Origin.Number(), l1Height, unsafeL2Height)
	err = system.ForkL1(l1Origin.ParentHash())
	require.NoError(t, err)

	n := len(batchesBefore)
	batchesAfter := collectFirstNL2SafeBlocks(ctx, t, system, l2Verif, n, startIndex)

	log.Info("+++ L2 blocks after reorg", "value", batchesAfter)

	assert.Equal(t, batchesAfter, batchesBefore)

}

// TestConfirmationIntegrityWithReorgs
// Post batches to both Espresso and the L1 then force the L1 to reorg back to an earlier state in which those batches have not been posted.
// Wait for some time and check that the batches are eventually posted to the L1 again, in the same order as they were originally sequenced,
// as if the reorg did not happen.
// More specifically the test is defined as follows
//
//	Arrange:
//		Running Sequencer, Batcher in Espresso mode, OP node.
//	Act:
//		Store the (unfinalized) head of the L1 in variable h.
//		Wait for the first n batches to be posted on possibly unfinalized L1 blocks.
//		Collect the batches of the corresponding batches and store them in list L.
//		Reorg the L1 to height h.
//		Wait for the L2 to reach safe height n again as the corresponding batches are submitted again to L1.
//		Store these n batches in L'
//	Assert:
//		L == L'
func TestConfirmationIntegrityWithReorgs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	system, _, err := launcher.StartE2eDevnet(ctx, t)
	require.NoError(t, err, "failed to start dev environment with espresso dev node")

	run(ctx, t, system)
}
