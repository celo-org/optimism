package environment_test

import (
	"context"
	"io"
	"log/slog"
	"math/big"
	"testing"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum-optimism/optimism/op-service/sources/mocks"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

// TestPipelineEnhancement ensures the derivation pipeline does not include batches that lack
// a BatchInfoAuthenticated event from the BatchAuthenticator contract.
//
// When batch authentication is enabled (BatchAuthenticatorAddress is set), the pipeline uses
// event-based authentication: it scans L1 receipts in a lookback window for a
// BatchInfoAuthenticated event matching the batch hash. Transactions without a corresponding
// auth event are filtered out, regardless of sender or receipt status.
//
//	Arrange:
//		Running Sequencer, Batcher in Espresso mode (with BatchAuthenticator deployed)
//	Act:
//		Send a transaction from a non-batcher account to the batch inbox contract
//		with a specific payload (0x42424242424242424242).
//		Fetch N, the block number where the transaction was included.
//	Assert:
//		Instantiate a data source with block N via the DataSourceFactory.
//		Verify that the pipeline returns no data, because the transaction has no
//		corresponding BatchInfoAuthenticated event.

func TestPipelineEnhancement(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t)
	require.NoError(t, err, "failed to start dev environment with espresso dev node")

	// Stop the batcher to ensure no valid batch is posted to L1.
	driver := system.BatchSubmitter.TestDriver()
	err = driver.StopBatchSubmitting(ctx)
	require.NoError(t, err, "failed to stop batch submitter")

	l1Client := system.NodeClient(e2esys.RoleL1)

	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	// Send a transaction from a non-batcher account to the BatchInbox EOA.
	// This tx will not have a BatchInfoAuthenticated event, so the pipeline should reject it.
	txData := []byte("42424242424242424242")

	tx := gethTypes.MustSignNewTx(system.Cfg.Secrets.Bob, system.RollupConfig.L1Signer(), &gethTypes.DynamicFeeTx{
		ChainID:   system.Cfg.L1ChainIDBig(),
		Nonce:     0,
		GasTipCap: big.NewInt(1 * params.GWei),
		GasFeeCap: big.NewInt(10 * params.GWei),
		Gas:       5_000_000,
		To:        &system.RollupConfig.BatchInboxAddress,
		Value:     big.NewInt(0),
		Data:      txData,
	})

	l := log.NewLogger(slog.Default().Handler())
	err = l1Client.SendTransaction(ctx, tx)
	require.NoError(t, err)

	// BatchInbox is an EOA, so the tx succeeds on L1. The pipeline rejects it because
	// there is no matching BatchInfoAuthenticated event.
	receipt, err := wait.ForReceiptMaybe(ctx, l1Client, tx.Hash(), gethTypes.ReceiptStatusSuccessful, true)
	require.NoError(t, err, "Waiting for receipt on transaction", tx)

	l1ClientFetching, err := client.NewRPC(ctx, nil, system.NodeEndpoint(e2esys.RoleL1).RPC())
	require.NoError(t, err)
	l1RefClient, err := sources.NewL1Client(l1ClientFetching, l, nil, sources.L1ClientDefaultConfig(system.RollupConfig, true, sources.RPCKindStandard))
	require.NoError(t, err)

	// Mock the L1 Beacon client as by default system.RollupConfig.EcotoneTime = 0
	p := mocks.NewBeaconClient(t)
	f := mocks.NewBeaconClient(t)
	c := sources.NewL1BeaconClient(p, sources.L1BeaconClientConfig{}, f)

	factory := derive.NewDataSourceFactory(l, system.RollupConfig, l1RefClient, c, nil)

	batcherAddress := crypto.PubkeyToAddress(*system.BatchSubmitter.Espresso.BatcherPublicKey)
	l1Block, err := l1Client.BlockByNumber(ctx, receipt.BlockNumber)
	require.NoError(t, err)
	l1BlockRef := eth.L1BlockRef{
		Hash:       l1Block.Hash(),
		Number:     l1Block.NumberU64(),
		ParentHash: l1Block.ParentHash(),
		Time:       l1Block.Time(),
	}
	datas, err := factory.OpenData(ctx, l1BlockRef, batcherAddress)
	require.NoError(t, err)

	data, err := datas.Next(ctx)

	// The pipeline returns no data because the tx has no matching BatchInfoAuthenticated event
	require.Equal(t, data, eth.Data(nil))
	require.Equal(t, err, io.EOF)
}
