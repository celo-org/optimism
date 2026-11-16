package devnet_tests

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/stretchr/testify/require"
)

// TestBatcherRestart stops the Espresso batcher, checks that a transaction
// submitted meanwhile is not derived by the verifier, then restarts the batcher
// and checks that the transaction is derived.
func TestBatcherRestart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	d := NewDevnet(ctx, t)
	require.NoError(t, d.Up())
	defer func() {
		require.NoError(t, d.Down())
	}()

	require.NoError(t, d.WaitForBatcher(ctx))

	// Send a transaction just to check that everything has started up ok.
	require.NoError(t, d.RunSimpleL2Burn())

	// Shut down the batcher and have another transaction submitted while it is down.
	require.NoError(t, d.ServiceDown(OpBatcher))
	d.SleepOutageDuration()

	receipt, err := d.SubmitSimpleL2Burn()
	require.NoError(t, err)

	// While the batcher is down, the verifier must not derive the transaction.
	d.SleepOutageDuration()
	_, err = d.L2Verif.TransactionReceipt(ctx, receipt.Receipt.TxHash)
	require.ErrorIs(t, err, ethereum.NotFound)

	// Bring the batcher back up and check that it batches the transaction
	// submitted while it was down.
	require.NoError(t, d.ServiceUp(OpBatcher))
	require.NoError(t, d.WaitForBatcher(ctx))
	require.NoError(t, d.VerifySimpleL2Burn(receipt))

	// Submit another transaction at the end just to check that things stay working.
	d.SleepRecoveryDuration()
	require.NoError(t, d.RunSimpleL2Burn())
}
