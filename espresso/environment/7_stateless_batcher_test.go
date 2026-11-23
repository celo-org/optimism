package environment_test

import (
	"context"
	"math/big"
	"math/rand/v2"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatelessBatcher is a test that verifies a batcher can operate (especially restart) correctly and efficiently without persistent storage.
//
// This test is designed to evaluate Test 7 as outlined within the
// Espresso Celo Integration plan.  It has stated task definition as follows:
// Run the rollup and randomly restart the batcher. Check the liveness of the rollup, and the consistency of Espresso confirmations and L1 confirmations.
// We don't need to clear persistent storage because the original Optimism code isn't and our integration work shouldn't use any.
// More specifically the test is defined as follows
//	Arrange:
//		Running Sequencer, Batcher in Espresso mode, OP node.
//	Act:
//		Loop over n iterations
//      Randomly pick one iteration to stop the batcher and another to start the batcher
//      For all the other iterations send one coin to Alice.
//	Assert:
//		Query the OP node to check that Alice balance has been increased by n-2

func TestStatelessBatcher(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t)
	// Signal the testnet to shut down
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	addressAlice := system.Cfg.Secrets.Addresses().Alice
	rollupClient := system.RollupClient(e2esys.RoleVerif)
	l2Seq := system.NodeClient(e2esys.RoleSeq)
	l2Verif := system.NodeClient(e2esys.RoleVerif)

	// Fund Alice
	env.RunSimpleL1TransferAndVerifier(ctx, t, system)

	balanceAliceInitial, err := l2Verif.BalanceAt(ctx, addressAlice, nil)
	if have, want := err, error(nil); have != want {
		t.Fatalf("Failed to fetch Alice's balance:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	// Setup Bob for sending coins to Alice
	privateKey := system.Cfg.Secrets.Bob
	bobOptions, err := bind.NewKeyedTransactorWithChainID(privateKey, system.Cfg.L1ChainIDBig())
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to create transaction options for bob:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	amount := new(big.Int).SetUint64(1)
	numTransfers := 0
	bobOptions.Value = amount

	driver := system.BatchSubmitter.TestDriver()
	safeBlockInclusionDuration := time.Duration(6*system.Cfg.DeployConfig.L1BlockTime) * time.Second

	numIterations := 8

	var txHashes []common.Hash

	// We select a range of iterations when the batcher is turned off.
	restartIteration := 1 + rand.IntN(numIterations-1)
	for i := 0; i < numIterations; i++ {

		// +1 because of the deposit transaction above
		nonce := uint64(numTransfers + 1)

		t.Log("******************* Iteration: ", i)
		//Let us stop the batcher
		if i == restartIteration {

			t.Log("+++++++++++++++++ Iteration with batcher restart: ", i)
			// Stop the batcher
			err = driver.StopBatchSubmitting(ctx)
			require.NoError(t, err)

			// wait for any old safe blocks being submitted / derived
			time.Sleep(safeBlockInclusionDuration)

			// get the initial sync status
			seqStatus, err := rollupClient.SyncStatus(context.Background())
			require.NoError(t, err)

			// ensure that the safe chain does not advance while the batcher is stopped
			newSeqStatus, err := rollupClient.SyncStatus(ctx)
			require.NoError(t, err)
			require.Equal(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain advanced while batcher was stopped")

			// Send a transaction while the batcher is down. This transaction should still be processed correctly by the sequencer and at some point be
			// inserted in a safe L2 block
			receipt := helpers.SendL2TxWithID(t, system.Cfg.L2ChainIDBig(), l2Seq, system.Cfg.Secrets.Bob, func(opts *helpers.TxOpts) {
				opts.Nonce = nonce
				opts.ToAddr = &addressAlice
				opts.Value = new(big.Int).SetUint64(1)
			})

			// Store the hash to check later if the transaction has been submitted successfully to the L2
			tx_hash := receipt.TxHash

			// Start again
			err = driver.StartBatchSubmitting()
			require.NoError(t, err)
			time.Sleep(safeBlockInclusionDuration)
			t.Log("Batcher restarting....")

			// Ensure that the safe chain does advance while the batcher is stopped
			_, err = geth.WaitForBlockToBeSafe(new(big.Int).SetUint64(seqStatus.SafeL2.Number+1), l2Verif, 2*time.Minute)
			require.NoError(t, err)
			newSeqStatus, err = rollupClient.SyncStatus(ctx)
			require.NoError(t, err)
			require.Greater(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain does not make progress")

			// Ensure the transaction sent while the batcher was down did go through
			_, err = wait.ForReceiptOK(ctx, l2Verif, tx_hash)
			require.NoError(t, err)

		} else {
			// The batcher is up, we can send coins
			txHash := env.RunSimpleL2Transfer(ctx, t, system, nonce, *amount, l2Seq)
			txHashes = append(txHashes, txHash)
		}

		// There should be a transfer for each iteration
		numTransfers++
	}

	var numDepositsBigInt big.Int
	numDepositsBigInt.SetInt64(int64(numTransfers))

	expectedAmount := new(big.Int).Mul(new(big.Int).Add(balanceAliceInitial, &numDepositsBigInt), amount)

	// Wait for the transfers to be processed
	for _, hash := range txHashes {
		_, err := wait.ForReceiptOK(ctx, l2Verif, hash)
		require.NoError(t, err)
	}

	l2BalanceNew, _ := l2Verif.BalanceAt(ctx, addressAlice, nil)

	assert.Equal(t, expectedAmount, l2BalanceNew)
}
