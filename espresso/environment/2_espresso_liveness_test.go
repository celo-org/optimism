package environment_test

import (
	"context"
	"math/big"
	"math/rand"
	"testing"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	geth_types "github.com/ethereum/go-ethereum/core/types"
)

// TestE2eDevnetWithEspressoDegradedLiveness is a test that checks that
// the rollup will continue to make progress even in the event of intermittent
// Espresso system failures.
//
// The Criteria for this test is as follows:
//
//	Requirement: Resubmission to Espresso.
//		Randomly turn the Espresso builder off and on. Check that the rollup
//		continues to make progress, including progressing settlement on the
//		base layer.
//
// We don't have any direct way of turning the Espresso builder off and on via
// the Dev node API at the moment.  However, we do have the ability to turn
// the consensus layer on and off via turning hotshot on and off.
//
// This is **NOT** the same thing, nor would it result in the same behavior as
// turning the Builder off and on. For the following reasons:
//
//	1 HotShot being off means no new blocks are being produced
//	2 The Builder being off means that only empty blocks are being produced
//	3 Turning the Builder off potentially means losing pool information,
//	  requiring re-submission so that the builder can include the transaction
//	  in the next block.
//
// With these caveats in mind, we may be able to simulate the behavior of 2
// at the very least, if we intercept the client submitting transactions to
// Espresso, and simulating the client being unable to submit transactions.
// Likewise, we might be able to simulate 3 by falsely reporting to the
// submitter that the transaction was submitted successfully, and withholding
// the submission itself.
func TestE2eDevnetWithEspressoDegradedLiveness(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	// Start a Server to proxy requests to Espresso
	_, server, option := env.SetupQueryServiceIntercept(
		// This decider will randomly report successful submissions of
		// transactions to Espresso, but will not actually submit them.
		// This will approximately occur 10% of the time, given the
		// criteria to roll a number 0-9 and only to occur if the rolled
		// number is 0.
		env.SetDecider(env.NewRandomRollFakeSubmitTransactionSuccess(
			10,
			0,
			1,
			rand.New(rand.NewSource(0)),
		)),
	)

	defer server.Close()
	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t, option)

	// Signal the testnet to shut down
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	defer system.Close()
	defer func() {
		err = espressoDevNode.Stop()
		if err != nil {
			t.Fatalf("failed to stop espresso dev node: %v", err)
		}
	}()

	addressAlice := system.Cfg.Secrets.Addresses().Alice

	l2Seq := system.NodeClient(e2esys.RoleSeq)
	l2Verif := system.NodeClient(e2esys.RoleVerif)

	balanceAliceInitial, err := l2Verif.BalanceAt(ctx, addressAlice, nil)
	if have, want := err, error(nil); have != want {
		t.Fatalf("Failed to fetch Alice's balance:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	const N = 10
	{
		var receipts []*geth_types.Receipt

		for i := range N {
			receipt := helpers.SendL2TxWithID(t, system.Cfg.L2ChainIDBig(), l2Seq, system.Cfg.Secrets.Bob, func(opts *helpers.TxOpts) {
				opts.Nonce = uint64(i)
				opts.ToAddr = &addressAlice
				opts.Value = big.NewInt(1)
			})

			receipts = append(receipts, receipt)
		}

		// Let's verify that all of our transactions came through successfully
		for _, receipt := range receipts {
			_, err := wait.ForReceiptOK(ctx, l2Verif, receipt.TxHash)
			if have, want := err, error(nil); have != want {
				t.Fatalf("Waiting for L2 tx on verification client:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
			}
		}

		// Alice's balance should have increased by N
		balanceAliceFinal, err := l2Verif.BalanceAt(ctx, addressAlice, nil)
		if have, want := err, error(nil); have != want {
			t.Fatalf("Failed to fetch Alice's balance:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}

		expectedBalance := new(big.Int).Add(balanceAliceInitial, big.NewInt(int64(N)))
		if balanceAliceFinal.Cmp(expectedBalance) != 0 {
			t.Fatalf("Alice's balance did not increase as expected:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", balanceAliceFinal, expectedBalance)
		}
	}
}
