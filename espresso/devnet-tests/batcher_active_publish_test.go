package devnet_tests

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

// publishWait is how long TestBatcherActivePublishOnly lets the batchers run
// before checking which of them posted batches: long enough for the active
// batcher to post one, and, after a switch, for the batcher just deactivated
// to drain the transactions it already queued.
const publishWait = 60 * time.Second

// hasBatchTransactions reports whether sender sent a transaction to the batch
// inbox in L1 blocks [startBlock, endBlock].
func hasBatchTransactions(ctx context.Context, client *ethclient.Client, batchInbox, sender common.Address, startBlock, endBlock uint64) (bool, error) {
	for i := startBlock; i <= endBlock; i++ {
		block, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(i))
		if err != nil {
			return false, fmt.Errorf("failed to get block %d: %w", i, err)
		}
		for _, tx := range block.Transactions() {
			if tx.To() == nil || *tx.To() != batchInbox {
				continue
			}
			from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
			if err != nil {
				return false, fmt.Errorf("failed to recover the sender of %s: %w", tx.Hash(), err)
			}
			if from == sender {
				return true, nil
			}
		}
	}
	return false, nil
}

// TestBatcherActivePublishOnly tests that only the batcher that the
// BatchAuthenticator marks active posts batches to L1, before and after a
// switch. Both batchers run throughout.
func TestBatcherActivePublishOnly(t *testing.T) {
	ctx, d := StartDevnet(t, 20*time.Minute)

	// Send an initial transaction to check that everything has started up ok.
	require.NoError(t, d.RunSimpleL2Burn())

	config, err := d.RollupConfig(ctx)
	require.NoError(t, err)
	batchAuthenticator, owner, err := d.BatchAuthenticator(ctx)
	require.NoError(t, err)
	callOpts := &bind.CallOpts{Context: ctx}

	espressoBatcher, err := batchAuthenticator.EspressoBatcher(callOpts)
	require.NoError(t, err)
	fallbackBatcher := config.Genesis.SystemConfig.BatcherAddr

	activeIsEspresso, err := batchAuthenticator.ActiveIsEspresso(callOpts)
	require.NoError(t, err)
	t.Logf("initial state: activeIsEspresso = %v", activeIsEspresso)

	verifyPublishing := func(expectEspressoActive bool) {
		t.Logf("verifying publishing with expectEspressoActive = %v", expectEspressoActive)

		// Only count blocks mined after this point: the current head may
		// hold batches posted before this check (e.g. by a batcher that was
		// just deactivated).
		head, err := d.L1.BlockNumber(ctx)
		require.NoError(t, err)
		startBlock := head + 1

		// Generate L2 traffic for the active batcher to post.
		burn, err := d.SubmitSimpleL2Burn()
		require.NoError(t, err)
		t.Logf("generated L2 transaction %s in L2 block %d", burn.Receipt.TxHash, burn.Receipt.BlockNumber)

		time.Sleep(publishWait)

		endBlock, err := d.L1.BlockNumber(ctx)
		require.NoError(t, err)
		t.Logf("checking L1 blocks %d-%d", startBlock, endBlock)

		espressoPublished, err := hasBatchTransactions(ctx, d.L1, config.BatchInboxAddress, espressoBatcher, startBlock, endBlock)
		require.NoError(t, err)
		fallbackPublished, err := hasBatchTransactions(ctx, d.L1, config.BatchInboxAddress, fallbackBatcher, startBlock, endBlock)
		require.NoError(t, err)
		t.Logf("Espresso batcher published: %v, fallback batcher published: %v", espressoPublished, fallbackPublished)

		if expectEspressoActive {
			require.True(t, espressoPublished, "the Espresso batcher should publish when active")
			require.False(t, fallbackPublished, "the fallback batcher should not publish when inactive")
		} else {
			require.True(t, fallbackPublished, "the fallback batcher should publish when active")
			require.False(t, espressoPublished, "the Espresso batcher should not publish when inactive")
		}
	}

	verifyPublishing(activeIsEspresso)

	// Switch the active batcher.
	tx, err := batchAuthenticator.SetActiveIsEspresso(owner, !activeIsEspresso)
	require.NoError(t, err)
	_, err = wait.ForReceiptOK(ctx, d.L1, tx.Hash())
	require.NoError(t, err)
	activeIsEspresso = !activeIsEspresso
	t.Logf("switched to activeIsEspresso = %v", activeIsEspresso)

	// The batcher just deactivated may still have batch transactions in
	// flight: let them land before the next check starts counting.
	time.Sleep(publishWait)

	verifyPublishing(activeIsEspresso)
}
