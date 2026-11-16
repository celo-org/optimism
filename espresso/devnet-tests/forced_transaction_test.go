package devnet_tests

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

// TestForcedTransaction checks that, with the sequencer down, a user can still
// submit an L2 transaction through an L1 deposit to the OptimismPortal.
//
// It only checks that the deposit is accepted on L1, not that the verifier
// derives it once the sequencing window expires: that would take too long on
// this devnet, and shrinking the sequencing window breaks the other tests.
func TestForcedTransaction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	d := NewDevnet(ctx, t)
	require.NoError(t, d.Up())
	defer func() {
		require.NoError(t, d.Down())
	}()

	require.NoError(t, d.WaitForBatcher(ctx))

	alice := crypto.PubkeyToAddress(d.secrets.Alice.PublicKey)
	initialBalance, err := d.L2Verif.BalanceAt(ctx, alice, nil)
	require.NoError(t, err)
	require.Positive(t, initialBalance.Sign(), "Alice should have a positive L2 balance")
	t.Logf("Alice's initial L2 balance: %s", initialBalance)

	// Read the OptimismPortal address before the sequencer (and its rollup
	// RPC) goes down.
	systemConfig, _, err := d.SystemConfig(ctx)
	require.NoError(t, err)
	portalAddr, err := systemConfig.OptimismPortal(&bind.CallOpts{Context: ctx})
	require.NoError(t, err)
	l1ChainID, err := d.L1.ChainID(ctx)
	require.NoError(t, err)

	t.Log("stopping the sequencer to force inclusion through L1")
	require.NoError(t, d.ServiceDown(OpNodeSequencer))

	portal, err := bindings.NewOptimismPortal(portalAddr, d.L1)
	require.NoError(t, err)
	opts, err := bind.NewKeyedTransactorWithChainID(d.secrets.Alice, l1ChainID)
	require.NoError(t, err)
	opts.Context = ctx
	// No msg.value: the deposit spends Alice's existing L2 balance rather
	// than minting new ETH. An explicit gas limit avoids estimation issues in
	// ResourceMetering.
	opts.GasLimit = 500_000

	tx, err := portal.DepositTransaction(
		opts,
		predeploys.L2ToL1MessagePasserAddr,
		big.NewInt(1000), // L2 value
		100_000,          // L2 gas limit
		false,
		nil,
	)
	require.NoError(t, err)
	t.Logf("deposit transaction submitted: %s", tx.Hash())

	receipt, err := wait.ForReceiptOK(ctx, d.L1, tx.Hash())
	require.NoError(t, err)
	t.Logf("deposit transaction mined in L1 block %d", receipt.BlockNumber)
}
