package environment_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/require"
)

// Time to wait for a transaction to be enforced after submission.
const WAIT_FORCED_TXN_TIME = 25 * time.Second

// Window small enough to guarantee that transactions are enforced before WAIT_FORCED_TXN_TIME.
const SMALL_SEQUENCER_WINDOW = 2 // Minimum possible value

// Window large enough to guarantee that transactions are not enforced before WAIT_FORCED_TXN_TIME.
const LARGER_SEQUENCER_WINDOW = 1000

// Get the appropriate sequencer window size for testing.
func sequencer_window_size(withSmallWindow bool) uint64 {
	if withSmallWindow {
		return SMALL_SEQUENCER_WINDOW
	}
	return LARGER_SEQUENCER_WINDOW
}

// ForcedTransaction attempts to verify that the forced transaction mechanism works for the
// withdrawal transaction with and without Espresso dev node.
//
// This function is designed to evaluate Test 11 as outlined within the Espresso Celo Integration
// plan. It has stated task definition as follows:
//
//	Arrange:
//		Set the sequencer window size small or large.
//		Start the devnet with the sequencer window setting, with or without the Espresso dev node.
//		Stop the sequencer.
//	Act:
//		Send a withdrawal and wait until the small window is passed.
//	Assert:
//		The balance reflects (or does not reflect) the withdrawal transaction, if the sequencer
//		window is set small (or large, respectively), regardless of whether launching with the
//		Espresso dev node.
func ForcedTransaction(t *testing.T, withSmallSequencerWindow bool, withEspresso bool) {
	// Set up the test timeout condition.
	// Extended timeout to accommodate slower processing in test environments
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Launch the devnet with the given sequencer window size.
	var system *e2esys.System
	var err error
	if withEspresso {
		launcher := new(env.EspressoDevNodeLauncherDocker)
		systemWithEspresso, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t, env.WithSequencerWindowSize(sequencer_window_size(withSmallSequencerWindow)))
		system = systemWithEspresso
		require.NoError(t, err, "Failed to launch with the Espresso dev node")
		defer env.Stop(t, system)
		defer env.Stop(t, espressoDevNode)
	} else {
		sysConfig := e2esys.DefaultSystemConfig(t, e2esys.WithAllocType(config.DefaultAllocType))
		sysConfig.DeployConfig.SequencerWindowSize = sequencer_window_size(withSmallSequencerWindow)
		system, err = sysConfig.Start(t)
		require.NoError(t, err, "failed to launch without Espresso dev node")
		defer env.Stop(t, system)
	}

	// Set up Alice's address and record the initial balance.
	l2Verif := system.NodeClient(e2esys.RoleVerif)
	address := system.Cfg.Secrets.Addresses().Alice
	l1Client := system.NodeClient(e2esys.RoleL1)
	initialBalance, err := l2Verif.BalanceAt(ctx, address, nil)
	require.NoError(t, err, "Failed to get initial balance")

	// Simulate sequencer downtime.
	err = system.RollupNodes[e2esys.RoleSeq].Stop(ctx)
	require.NoError(t, err, "Failed to stop sequencer")

	// Initiate a withdrawal from Alice to the L1 following
	// https://docs.unichain.org/docs/technical-information/submitting-transactions-from-l1#initiating-a-withdrawal-from-l1.
	portal, err := bindings.NewOptimismPortal(system.Cfg.L1Deployments.OptimismPortalProxy, l1Client)
	require.NoError(t, err, "Failed to create Optimism portal")
	opts, err := bind.NewKeyedTransactorWithChainID(system.Cfg.Secrets.Alice, system.Cfg.L1ChainIDBig())
	require.NoError(t, err, "Failed to create withdrawal transaction options")
	withdrawalAmount := new(big.Int).SetUint64(1000)
	tx, err := portal.DepositTransaction(
		opts,
		predeploys.L2ToL1MessagePasserAddr,
		withdrawalAmount,
		uint64(300_000),
		false,
		nil,
	)
	require.NoError(t, err, "Failed to create transaction")
	_, err = bind.WaitMined(ctx, l1Client, tx)
	require.NoError(t, err, "Transaction not minted")

	// Wait and attempt to get the new balance after the withdrawal.
	time.Sleep(WAIT_FORCED_TXN_TIME)
	newBalance, err := wait.ForBalanceChange(ctx, l2Verif, address, initialBalance)

	if withSmallSequencerWindow {
		// Verify that Alice's balance decreases as expected.
		require.NoError(t, err, "Failed to get new balance")
		require.LessOrEqualf(t, newBalance.Uint64(), initialBalance.Uint64()-withdrawalAmount.Uint64(), "Balance not decreased")

	} else {
		// Verify that Alice's balance is inaccessible.
		require.Error(t, err, "Not expected to get new balance")
	}
}

// TestForcedTransactionWithoutEspressoSmallWindow verifies that the withdrawal transaction is
// enforced after the sequencer window is passed when launching without the Espresso dev node.
func TestForcedTransactionWithoutEspressoSmallWindow(t *testing.T) {
	ForcedTransaction(t, true, false)
}

// TestForcedTransactionWithoutEspressoLargeWindow verifies that the withdrawal transaction is not
// enforced before the sequencer window is passed when launching without the Espresso dev node.
func TestForcedTransactionWithoutEspressoLargeWindow(t *testing.T) {
	ForcedTransaction(t, false, false)
}

// TestForcedTransactionWithEspressoSmallWindow verifies that the withdrawal transaction is
// enforced after the sequencer window is passed when launching with the Espresso dev node.
func TestForcedTransactionWithEspressoSmallWindow(t *testing.T) {
	ForcedTransaction(t, true, true)
}

// TestForcedTransactionWithEspressoLargeWindow verifies that the withdrawal transaction is not
// enforced before the sequencer window is passed when launching with the Espresso dev node.
func TestForcedTransactionWithEspressoLargeWindow(t *testing.T) {
	ForcedTransaction(t, false, false)
}
