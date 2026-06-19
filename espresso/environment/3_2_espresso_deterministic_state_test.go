package environment_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	geth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

// TestDeterministicDerivationExecutionStateWithInvalidTransaction is a test that
// attempts to make sure that the op-node continues to derive and finalize a
// valid chain even when malicious transactions are submitted.
//
// This test is designed to evaluate Test 3.2 as outlined within the
// Espresso Celo Integration plan.  It has stated task definition as follows:
//
//	Arrange:
//		Running Sequencer, Batcher in Espresso mode, and OP node.
//	Act:
//		Send some transactions from Bob to Alice and some regular L2 transactions.
//		While sending regular L2 transactions to the sequencer also send transactions to Espresso using an invalid batcher address, and transactions directly to L1 (e.g. transactions that were not previously posted to Espresso).
//	Assert:
//		The op-node continues to finalize valid blocks on L1, ignoring the malicious transactions.

func TestDeterministicDerivationExecutionStateWithInvalidTransaction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	// Start the devnet with the sequencer using finalized blocks
	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t, env.WithL1FinalizedDistance(0), env.WithSequencerUseFinalized(true))
	// Signal the testnet to shut down
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	// We want to setup our test
	addressAlice := system.Cfg.Secrets.Addresses().Alice
	espressoClient, err := espressoClient.NewMultipleNodesClient(espressoDevNode.EspressoUrls())
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to create Espresso client:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}
	l1Client := system.NodeClient(e2esys.RoleL1)
	l2Verif := system.NodeClient(e2esys.RoleVerif)
	l2Seq := system.NodeClient(e2esys.RoleSeq)

	// We want to send some transactions from Bob to Alice
	{
		privateKey := system.Cfg.Secrets.Bob
		bobOptions, err := bind.NewKeyedTransactorWithChainID(privateKey, system.Cfg.L1ChainIDBig())
		if have, want := err, error(nil); have != want {
			t.Fatalf("failed to create transaction options for bob:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}

		mintAmount := new(big.Int).SetUint64(1)
		bobOptions.Value = mintAmount
		_ = helpers.SendDepositTx(t, system.Cfg, l1Client, l2Verif, bobOptions, func(l2Opts *helpers.DepositTxOpts) {
			// Send from Bob to Alice
			l2Opts.ToAddr = addressAlice
		})
	}

	// Send some regular L2 transactions in each iteration and there are 10 rounds in total
	// Since we wait for valid transactions sent before attackRoundEspresso and after attackRoundL1 to be included,
	// we can be confident that the malicious transactions are included on L1/Espresso while the L2 chain is making progress.
	// The reason is that the iterations of the test are executed sequentially.
	numIterations := 10
	attackRoundEspresso := 5 // the round where we send transactions directly to Espresso outside without running the batcher code.
	attackRoundL1 := 7       // the round where we send a transaction directly to the BatchInbox EOA.
	// Compare states between nodes for multiple latest blocks
	// We don't compare states for every individual block as any diff in block x will be reflected in block x + n
	for i := 0; i < numIterations; i++ {

		// Send some regular L2 transactions in each iteration
		tx := geth_types.MustSignNewTx(system.Cfg.Secrets.Bob, geth_types.LatestSignerForChainID(system.Cfg.L2ChainIDBig()), &geth_types.DynamicFeeTx{
			ChainID:   system.Cfg.L2ChainIDBig(),
			Nonce:     uint64(i + 1), // +1 because of the deposit transaction above
			To:        &addressAlice,
			Value:     big.NewInt(1),
			GasTipCap: big.NewInt(10),
			GasFeeCap: big.NewInt(200),
			Gas:       21_000,
		})
		err := l2Seq.SendTransaction(ctx, tx)
		if have, want := err, error(nil); have != want {
			t.Fatalf("Sending L2 tx:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}
		// Wait for the receipt
		_, err = wait.ForReceiptOK(ctx, l2Seq, tx.Hash())
		if have, want := err, error(nil); have != want {
			t.Fatalf("Waiting for L2 tx:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}

		// When it is the attack round, send some Espresso transactions using fakeBatcherPrivateKey directly to Espresso.
		// The L2 batch embedded in the Espresso transaction is well formed but will be ignored as the  transaction is not signed by the batcher and the batch information is not authenticated to the batch authentication contract either.

		switch i {
		case attackRoundEspresso:
			// Create a fake Espresso transaction
			fakeBatcherPrivateKey, err := forgedBatcherPrivateKey()
			if err != nil {
				t.Fatalf("Failed to get fake batcher private key:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", err, nil)
			}
			fakeEspressoTransaction, err := createEspressoTransaction(TEST_ESPRESSO_TRANSACTION, system.Cfg.L2ChainIDBig(), fakeBatcherPrivateKey)
			if err != nil {
				t.Fatalf("Failed to create fake Espresso transaction:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", err, nil)
			}

			// Send transaction directly to Espresso to bypass the batcher
			txHash, err := espressoClient.SubmitTransaction(ctx, *fakeEspressoTransaction)
			if err != nil {
				t.Fatalf("Failed to submit transaction:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", err, nil)
			}

			err = env.WaitForEspressoTx(ctx, txHash, espressoClient)
			if err != nil {
				t.Fatalf("Espresso transaction failed to be confirmed:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", err, nil)
			}

		case attackRoundL1:
			// create a transaction
			tx := geth_types.MustSignNewTx(system.Cfg.Secrets.Bob, system.RollupConfig.L1Signer(), &geth_types.DynamicFeeTx{
				ChainID:   system.Cfg.L1ChainIDBig(),
				Nonce:     1,
				To:        &system.RollupConfig.BatchInboxAddress,
				Value:     big.NewInt(1),
				GasTipCap: big.NewInt(1 * params.GWei),
				GasFeeCap: big.NewInt(10 * params.GWei),
				Gas:       5_000_000,
			})
			// Send a transaction directly to L1
			err = l1Client.SendTransaction(ctx, tx)
			if have, want := err, error(nil); have != want {
				t.Fatalf("failed to send transaction directly to L1:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
			}

			// BatchInbox is an EOA, so the tx succeeds on L1. The pipeline
			// ignores it because there is no matching BatchInfoAuthenticated event.
			_, err = wait.ForReceiptOK(ctx, l1Client, tx.Hash())
			if have, want := err, error(nil); have != want {
				t.Fatalf("failed to get receipt for transaction:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
			}
		}

		// Get latest finalized block from op node, ensuring that the chain
		// continues to finalize despite the malicious transactions submitted
		// directly to Espresso and L1, which the derivation pipeline ignores.
		// We use BlockByNumber to get the states as the engine state will be reflected in the block.
		_, err = l2Verif.BlockByNumber(ctx, big.NewInt(rpc.FinalizedBlockNumber.Int64()))
		if err != nil {
			t.Fatalf("failed to get block from opBlock: %v", err)
		}
	}

}

// forgeBatcherPrivateKey is a helper function that forge a batcher private key
func forgedBatcherPrivateKey() (*ecdsa.PrivateKey, error) {
	return crypto.GenerateKey()
}

func realBatcherPrivateKey(system *e2esys.System) (*ecdsa.PrivateKey, error) {
	return system.Cfg.Secrets.Batcher, nil
}

const TEST_ESPRESSO_TRANSACTION = "0xf90388f9023da00d68b82fa254b7d23a8584bcaa67be241a269c86aac05a2a6fc805a672bb910ea01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347944200000000000000000000000000000000000011a0d6cc9c002bc6a8d1c8501c57301b6b2f037494e1e0f61e417411e17f4e80b5afa028881bc4fc4c5fa67f26462837f88937961b6667ae4af043218a0c1b72a5f53ca0d8056577b8ef8e580c0ebc96def906b3699ddc8d91e15abf9c7a7e7bb4f85c96b901000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080018401c9c380830272ca84681d98b780a0000000000000000000000000000000000000000000000000000000000000000088000000000000000001a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b4218080a00000000000000000000000000000000000000000000000000000000000000000f849a00d68b82fa254b7d23a8584bcaa67be241a269c86aac05a2a6fc805a672bb910e80a0d7d069186bed40982ca7e7747d61c78718d0dda165d74e62164c1bba165001f784681d98b7c0b8fb7ef8f8a07a2aa57f213dfe5e61ceaebcd45c61252157b4e3c1e82e1ec0dca455b1173ad894deaddeaddeaddeaddeaddeaddeaddeaddead00019442000000000000000000000000000000000000158080830f424080b8a4440a5e20000f424000000000000000000000000100000000681d98b60000000000000000000000000000000000000000000000000000000000000000000000003b9aca000000000000000000000000000000000000000000000000000000000000000001d7d069186bed40982ca7e7747d61c78718d0dda165d74e62164c1bba165001f70000000000000000000000003c44cdddb6a900fa2b585dd299e03d12fa4293bc"

// createEspressoTransaction creates a Espresso transaction with a FAKE or REAL batcher private key
func createEspressoTransaction(transactionString string, chainID *big.Int, batcherKey *ecdsa.PrivateKey) (*espressoCommon.Transaction, error) {
	// This is the genesis Espresso transaction that created by honest sequencer
	bufData, err := hexutil.Decode(transactionString)
	if err != nil {
		log.Error("failed to decode Espresso transaction in the test", "error", err)
		return nil, err
	}
	buf := bytes.NewBuffer(bufData)

	// Sign the encoded batch with FAKE or REAL batcher private key
	batcherSignature, err := crypto.Sign(crypto.Keccak256(buf.Bytes()), batcherKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create batcher signature: %w", err)
	}

	// Combine signature and batch data
	payload := append(batcherSignature, buf.Bytes()...)

	// Create and return Espresso Transaction
	return &espressoCommon.Transaction{
		Namespace: chainID.Uint64(),
		Payload:   payload,
	}, nil
}

// espressoTransactionDataSkippingUnmarshal extract the L1 info deposit from Espresso transaction without checking whether the unmarshal could work
func espressoTransactionDataSkippingUnmarshal(transactionString string) (*geth_types.Transaction, error) {
	bufData, err := hexutil.Decode(transactionString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Espresso transaction in the test: %w", err)
	}
	buf := bytes.NewBuffer(bufData)

	batchData := buf.Bytes()

	var batch derive.EspressoBatch
	if err := rlp.DecodeBytes(batchData, &batch); err != nil {
		return nil, fmt.Errorf("failed to decode Espresso batch: %w", err)
	}

	return batch.L1InfoDeposit, nil
}

// TestValidEspressoTransactionCreation is a test that
// make sure we have correct way to create a Espresso transaction.
// This test is a unit test to serve the correctness of TestDeterministicDerivationExecutionStateWithInvalidTransaction.
func TestValidEspressoTransactionCreation(t *testing.T) {
	// Ignore it by default as it takes a long time to run
	t.Skip("skipping test")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	// once this StartE2eDevnet returns, we have a running Espresso Dev Node
	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t)
	// Signal the testnet to shut down
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	// We want to setup our test
	espressoClient, err := espressoClient.NewMultipleNodesClient(espressoDevNode.EspressoUrls())
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to create Espresso client:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}
	l2Verif := system.NodeClient(e2esys.RoleVerif)
	// create a real Espresso transaction and make sure it can go through
	{
		// Create a real Espresso transaction
		realBatcherPrivateKey, err := realBatcherPrivateKey(system)
		if err != nil {
			t.Fatalf("Failed to get real batcher private key:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", err, nil)
		}
		realEspressoTransaction, err := createEspressoTransaction(TEST_ESPRESSO_TRANSACTION, system.Cfg.L2ChainIDBig(), realBatcherPrivateKey)
		if err != nil {
			t.Fatalf("Failed to create real Espresso transaction:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", err, nil)
		}

		// Send transaction directly to Espresso
		txHash, err := espressoClient.SubmitTransaction(ctx, *realEspressoTransaction)
		if err != nil {
			t.Fatalf("Failed to submit transaction:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", err, nil)
		}

		// Parameters for transaction fetching loop, which waits for transactions
		// to be sequenced on Espresso
		transactionFetchTimeout := 4 * time.Second
		transactionFetchInterval := 100 * time.Millisecond

		// Check Espresso will accept the transaction
		timer := time.NewTimer(transactionFetchTimeout)
		defer timer.Stop()

		ticker := time.NewTicker(transactionFetchInterval)
		defer ticker.Stop()

		transactionFound := false
		for !transactionFound {
			select {
			case <-ticker.C:
				_, err := espressoClient.FetchTransactionByHash(ctx, txHash)
				if err == nil {
					// test pass
					transactionFound = true
				}
			case <-timer.C:
				t.Fatalf("Failed to fetch transaction by hash after multiple attempts")
			case <-ctx.Done():
				t.Fatalf("Cancelling transaction publishing")
			}
		}

		// Make sure the transaction will go through to op node by checking it will go through batch submitter's streamer
		batchSubmitter := system.BatchSubmitter
		_, err = batchSubmitter.EspressoStreamer().UnmarshalBatch(realEspressoTransaction.Payload)
		if have, want := err, error(nil); have != want {
			t.Fatalf("Failed to unmarshal batch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}

		// Extract L1 info deposit transaction from Espresso transaction
		l1InfoDeposit, err := espressoTransactionDataSkippingUnmarshal(TEST_ESPRESSO_TRANSACTION)
		if err != nil {
			t.Fatalf("Failed to get L1 info deposit:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", err, nil)
		}

		// Make sure the transaction will really go through to verifier by waiting for its hash
		_, err = wait.ForReceiptOK(ctx, l2Verif, l1InfoDeposit.Hash())
		require.NoError(t, err, "deposit didn't arrive on Decaf node")
	}

}
