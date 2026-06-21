package environment_test

import (
	"context"
	"testing"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
)

// TestE2eDevnetWithEspressoSimpleTransactions launches the e2e Dev Net with the Espresso Dev Node
// and runs a couple of simple transactions to it.
func TestE2eDevnetWithEspressoSimpleTransactions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t)
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	// Signal the testnet to shut down on exit
	defer env.Stop(t, espressoDevNode)
	defer env.Stop(t, system)
	// Send Transaction on L1, and wait for verification on the L2 Verifier
	env.RunSimpleL1TransferAndVerifier(ctx, t, system)

	// Submit a Transaction on the L2 Sequencer node, to a Burn Address
	env.RunSimpleL2Burn(ctx, t, system)
}

// TestE2eDevnetWithEspressoAndAltDaSimpleTransactions launches the e2e Dev Net with the Espresso
// Dev Node in AltDA mode and runs a couple of simple transactions to it.
func TestE2eDevnetWithEspressoAndAltDaSimpleTransactions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	// WithAltDa enables UseAltDA, which wires the e2e system to the in-process
	// altda.FakeDAServer; no external EigenDA proxy is contacted.
	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t, env.WithAltDa())
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	// Signal the testnet to shut down on exit
	defer env.Stop(t, espressoDevNode)
	defer env.Stop(t, system)
	// Send Transaction on L1, and wait for verification on the L2 Verifier
	env.RunSimpleL1TransferAndVerifier(ctx, t, system)

	// Submit a Transaction on the L2 Sequencer node, to a Burn Address
	env.RunSimpleL2Burn(ctx, t, system)
}

// TestE2eDevnetWithoutEspressoSimpleTransactions launches the e2e Dev Net
// without the Espresso Dev Node and runs a couple of simple transactions to it.
func TestE2eDevnetWithoutEspressoSimpleTransaction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sysConfig := e2esys.DefaultSystemConfig(t, e2esys.WithAllocType(config.DefaultAllocType))

	system, err := sysConfig.Start(t)
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start e2e dev environment:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}
	// Shut down the test net on exit
	defer env.Stop(t, system)

	// Send Transaction on L1, and wait for verification on the L2 Verifier
	env.RunSimpleL1TransferAndVerifier(ctx, t, system)

	// Submit a Transaction on the L2 Sequencer node, to a Burn Address
	env.RunSimpleL2Burn(ctx, t, system)
}
