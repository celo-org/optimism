package environment

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	"github.com/ethereum/go-ethereum/ethclient"
)

// runSimpleL2Transfer runs a simple L2 burn transaction and verifies it on the
// L2 Verifier.
func RunSimpleL2Transfer(
	ctx context.Context,
	t *testing.T,
	system *e2esys.System,
	nonce uint64,
	amount big.Int,
	l2Seq *ethclient.Client,
) common.Hash {
	_, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	privateKey := system.Cfg.Secrets.Bob

	t.Log("Sending tx", "nonce", nonce)

	destAddress := system.Cfg.Secrets.Addresses().Alice

	receipt := helpers.SendL2TxWithID(t, system.Cfg.L2ChainIDBig(), l2Seq, privateKey, func(opts *helpers.TxOpts) {
		opts.Nonce = nonce
		opts.ToAddr = &destAddress
		opts.Value = &amount
	})

	t.Log("Receipt", receipt)

	txHash := receipt.TxHash

	return txHash
}

// runSimpleL1TransferAndVerifier runs a simple L1 transfer and verifies it on
// the L2 Verifier.
func RunSimpleL1TransferAndVerifier(ctx context.Context, t *testing.T, system *e2esys.System) {
	privateKey := system.Cfg.Secrets.Bob

	l1Client := system.NodeClient(e2esys.RoleL1)
	l2Verif := system.NodeClient(e2esys.RoleVerif)

	fromAddress := system.Cfg.Secrets.Addresses().Bob

	// Send Transaction on L1, and wait for verification on the L2 Verifier
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Get the Starting Balance of the Address
	startBalance, err := l2Verif.BalanceAt(ctx, fromAddress, nil)
	if have, want := err, error(nil); have != want {
		t.Errorf("attempt to get starting balance for %s failed:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", fromAddress, have, want)
	}

	// Create a new Keyed Transaction
	options, err := bind.NewKeyedTransactorWithChainID(privateKey, system.Cfg.L1ChainIDBig())
	require.NoError(t, err, "failed to create keyed transaction with chain ID %d", system.Cfg.L1ChainIDBig())

	// Send a Deposit Transaction
	mintAmount := big.NewInt(1_000_000_000_000)
	options.Value = mintAmount
	_ = helpers.SendDepositTx(t, system.Cfg, l1Client, l2Verif, options, nil)

	endBalance, err := wait.ForBalanceChange(ctx, l2Verif, fromAddress, startBalance)
	require.NoError(t, err, "waiting for balance change failed")

	diff := new(big.Int).Sub(endBalance, startBalance)
	require.Equal(t, diff, mintAmount, "balance change does not match mint amount")

	cancel()
}

// RunSimpleL2Burn runs a simple L2 burn transaction and verifies it on the
// L2 Verifier with a 2-minute timeout.
func RunSimpleL2Burn(ctx context.Context, t *testing.T, system *e2esys.System) {
	RunSimpleL2BurnWithTimeout(ctx, t, system, 2*time.Minute)
}

// RunSimpleL2BurnWithTimeout runs a simple L2 burn and verifies on the verifier,
// using the given timeout for the overall operation. Use a longer timeout (e.g. 5*time.Minute)
// when the verifier may be slow to derive, e.g. after a batcher switch.
func RunSimpleL2BurnWithTimeout(ctx context.Context, t *testing.T, system *e2esys.System, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	l2Seq := system.NodeClient(e2esys.RoleSeq)
	l2Verif := system.NodeClient(e2esys.RoleVerif)

	senderKey := system.Cfg.Secrets.Bob
	senderAddress := system.Cfg.Secrets.Addresses().Bob
	amountToBurn := big.NewInt(1234)
	burnAddress := common.Address{0xff, 0xff}

	nonce, err := l2Seq.NonceAt(ctx, senderAddress, nil)
	require.NoError(t, err, "failed to get nonce for account %s", senderAddress)

	initialBurnAddressBalance, err := l2Seq.BalanceAt(ctx, burnAddress, nil)
	require.NoError(t, err, "failed to get initial balance for burn address %s", burnAddress)

	// Use the ctx-scoped sender (not helpers.SendL2TxWithID, which hardcodes its
	// own 30s timeout) so the caller's timeout governs the verifier-receipt wait;
	// after a batcher switch the verifier can take longer than 30s to derive.
	sendL2TxAndVerify(
		ctx,
		t,
		system.Cfg.L2ChainIDBig(),
		l2Seq,
		senderKey,
		L2TxWithOptions(
			L2TxWithAmount(amountToBurn),
			L2TxWithNonce(nonce),
			L2TxWithToAddress(&burnAddress),
			L2TxWithVerifyOnClients(l2Verif),
		),
	)

	// Check the balance of hte burn address using the L2 Verifier
	burnAddressBalance, err := wait.ForBalanceChange(ctx, l2Verif, burnAddress, initialBurnAddressBalance)
	require.NoError(t, err, "burn address balance didn't change")

	// Make sure that these match
	require.Equal(t, new(big.Int).Sub(burnAddressBalance, initialBurnAddressBalance), amountToBurn, "burn address balance doesn't match the amount burned")

	cancel()
}

// RunSimpleMultiTransactions sends numTransactions simple L2 transactions
// from Bob's account and returns the receipts.
//
// This is all attempted in parallel, as it will spawn a separate goroutine
// for each transaction submission.  Each transaction will be provided its
// own nonce, based on the currently understood value of the nonce for
// Bob.
//
// This will return once all receipts have been returned.
func RunSimpleMultiTransactions(ctx context.Context, t *testing.T, system *e2esys.System, numTransactions int) ([]*types.Receipt, error) {
	ctx, cancel := context.WithTimeoutCause(ctx, 2*time.Minute, fmt.Errorf("failed to submit all transactions within time frame: %w", context.DeadlineExceeded))
	defer cancel()

	senderKey := system.Cfg.Secrets.Bob
	senderAddress := system.Cfg.Secrets.Addresses().Bob
	l2Seq := system.NodeClient(e2esys.RoleSeq)
	nonce, err := l2Seq.NonceAt(ctx, senderAddress, nil)
	if err != nil {
		require.NoError(t, err, "failed to get nonce for account %s", senderAddress)
	}

	ch := make(chan *types.Receipt, numTransactions)
	defer close(ch)
	for i := range numTransactions {
		go (func(ch chan *types.Receipt, i int, nonce uint64) {
			receipt := helpers.SendL2Tx(t, system.Cfg, l2Seq, senderKey, func(opts *helpers.TxOpts) {
				opts.Nonce = nonce + uint64(i)
				// We need to explicitly increase the gas beyond some threshold
				// for an unknown reason.  We'll set it high enough so that
				// it hopefully won't cause a problem
				opts.Gas = 100_000
			})
			ch <- receipt
		})(ch, i, nonce)
	}

	var receipts []*types.Receipt
	for range numTransactions {
		select {
		case <-ctx.Done():
			return receipts, ctx.Err()
		case receipt := <-ch:
			receipts = append(receipts, receipt)
		}
	}
	return receipts, nil
}

// sendL2TxAndVerify is an Espresso-local variant of helpers.SendL2TxWithID that
// honours the supplied ctx for the receipt waits instead of imposing its own
// fixed 30s deadline. The Espresso batcher-switch tests need a longer window
// because the verifier can lag well past 30s while it re-derives after a switch.
// Behaviour is otherwise identical: it signs and sends the tx, waits for an OK
// receipt on l2Client, asserts the expected status, and verifies the same
// receipt on every TxOpts.VerifyClients client.
func sendL2TxAndVerify(ctx context.Context, t *testing.T, chainID *big.Int, l2Client *ethclient.Client, privKey *ecdsa.PrivateKey, applyTxOpts helpers.TxOptsFn) *types.Receipt {
	opts := &helpers.TxOpts{
		Value:          common.Big0,
		GasTipCap:      big.NewInt(10),
		GasFeeCap:      big.NewInt(200),
		Gas:            21_000,
		ExpectedStatus: types.ReceiptStatusSuccessful,
	}
	applyTxOpts(opts)

	tx := types.MustSignNewTx(privKey, types.LatestSignerForChainID(chainID), &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     opts.Nonce,
		To:        opts.ToAddr,
		Value:     opts.Value,
		GasTipCap: opts.GasTipCap,
		GasFeeCap: opts.GasFeeCap,
		Gas:       opts.Gas,
		Data:      opts.Data,
	})

	require.NoError(t, l2Client.SendTransaction(ctx, tx), "Sending L2 tx")

	receipt, err := wait.ForReceiptOK(ctx, l2Client, tx.Hash())
	require.NoError(t, err, "Waiting for L2 tx")
	require.Equal(t, opts.ExpectedStatus, receipt.Status, "TX should have expected status")

	for i, client := range opts.VerifyClients {
		t.Logf("Waiting for tx %v on verification client %d", tx.Hash(), i)
		receiptVerif, err := wait.ForReceiptOK(ctx, client, tx.Hash())
		require.NoErrorf(t, err, "Waiting for L2 tx on verification client %d", i)
		require.Equalf(t, receipt, receiptVerif, "Receipts should be the same on sequencer and verification client %d", i)
	}
	return receipt
}
