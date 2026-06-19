package environment_test

import (
	"context"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
)

// TestEspressoDockerDevNodeSmokeTest is a smoke test for the Espresso Dev Node
// Docker implementation. It starts the dev node and then stops it. And tries
// to ensure that the e2e system, and the docker container stop correctly.
func TestEspressoDockerDevNodeSmokeTest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t)
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	defer env.Stop(t, espressoDevNode, env.IgnoreStopErrors)
	defer env.Stop(t, system, env.IgnoreStopErrors)

	{
		// Stop the Docker Container
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		espressoClose := make(chan struct{})

		var err error

		go (func(ch chan struct{}) {
			err = espressoDevNode.Stop()
			close(ch)
		})(espressoClose)

		select {
		case <-ctx.Done():
			t.Errorf("espresso dev node failed to stop in the anticipated time given: %v", ctx.Err())
		case <-espressoClose:
			// Espresso Dev Node stopped in the anticipated time
			if err != nil {
				t.Fatalf("failed to stop espresso dev node: %v", err)
			}
		}

		// One last sanity check to ensure that the container is not still
		// running.

		err = espressoDevNode.Stop()
		if err == nil {
			t.Fatalf("espresso dev node should return an error indicating that it cannot be stopped, as it is not running")
		}

		if _, castOk := err.(env.DockerContainerNotRunningError); !castOk {
			t.Fatalf("espresso dev node should return a DockerContainerNotRunningError, but received: %v", err)
		}
	}

	{
		// Stop the e2e system
		sysClose := make(chan struct{})

		go (func(ch chan struct{}) {
			system.Close()
			close(ch)
		})(sysClose)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		select {
		case <-ctx.Done():
			t.Errorf("system failed to close in the anticipated time given: %v", ctx.Err())

		case <-sysClose:
			// System closed in the anticipated time
		}
	}
}

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

	// Start a temporary EigenDA Docker instance for this test
	eigenda, err := env.StartEigenDA(ctx)
	if err != nil {
		t.Fatalf("failed to start EigenDA: %v", err)
	}
	// Stopped when the test exits
	defer env.StopDockerContainer(eigenda.ContainerID)

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
