package devnet_tests

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/require"
)

// TestBatcherSwitching tests that the batcher can be switched from the
// Espresso batcher to the fallback batcher using the BatchAuthenticator
// contract. This is the devnet equivalent of the espresso/environment
// TestBatcherSwitching.
//
// Both batchers run from the start: op-batcher (Espresso) and
// op-batcher-fallback, which stays idle while the BatchAuthenticator's
// activeIsEspresso is true. The test stops the Espresso batcher, switches the
// BatchAuthenticator to the fallback batcher and checks that the chain keeps
// being batched.
func TestBatcherSwitching(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	d := NewDevnet(ctx, t)
	require.NoError(t, d.Up())
	defer func() {
		require.NoError(t, d.Down())
	}()

	require.NoError(t, d.WaitForBatcher(ctx))

	// Send an initial transaction to check that everything has started up ok.
	require.NoError(t, d.RunSimpleL2Burn())

	batchAuthenticator, owner, err := d.BatchAuthenticator(ctx)
	require.NoError(t, err)
	callOpts := &bind.CallOpts{Context: ctx}

	activeIsEspresso, err := batchAuthenticator.ActiveIsEspresso(callOpts)
	require.NoError(t, err)
	require.True(t, activeIsEspresso, "the devnet should start with the Espresso batcher active")

	// Stop the Espresso batcher, so that only the fallback batcher can batch.
	require.NoError(t, d.StopBatcherSubmitting(OpBatcher))

	// Switch the active batcher to the fallback batcher.
	tx, err := batchAuthenticator.SetActiveIsEspresso(owner, false)
	require.NoError(t, err)
	receipt, err := wait.ForReceiptOK(ctx, d.L1, tx.Hash())
	require.NoError(t, err)
	t.Logf("switched to the fallback batcher in L1 block %d", receipt.BlockNumber)

	activeIsEspresso, err = batchAuthenticator.ActiveIsEspresso(callOpts)
	require.NoError(t, err)
	require.False(t, activeIsEspresso, "activeIsEspresso should be false after the switch")

	// Check that the fallback batcher batches new transactions.
	require.NoError(t, d.RunSimpleL2Burn())

	// Submit another transaction later to check that things stay working.
	d.SleepRecoveryDuration()
	require.NoError(t, d.RunSimpleL2Burn())
}
