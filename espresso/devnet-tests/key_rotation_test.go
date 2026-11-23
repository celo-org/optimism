package devnet_tests

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// TestChangeBatchAuthenticatorOwner transfers the ownership of the
// BatchAuthenticator to Bob with its two-step transfer, checks that the
// previous owner lost it, and that the chain keeps working.
//
// Unlike in the reference repo, the BatchAuthenticator proxy is administered
// by the chain's shared ProxyAdmin (see op-deployer's DeployEspresso), whose
// ownership is independent of the BatchAuthenticator's: the test checks that
// it is left unchanged rather than transferring it too.
func TestChangeBatchAuthenticatorOwner(t *testing.T) {
	ctx, d := StartDevnet(t, 20*time.Minute)

	// Send a transaction just to check that everything has started up ok.
	require.NoError(t, d.RunSimpleL2Burn())

	batchAuthenticator, owner, err := d.BatchAuthenticator(ctx)
	require.NoError(t, err)
	callOpts := &bind.CallOpts{Context: ctx}

	currentOwner, err := batchAuthenticator.Owner(callOpts)
	require.NoError(t, err)
	require.Equal(t, owner.From, currentOwner, "BATCH_AUTHENTICATOR_OWNER should own the BatchAuthenticator")
	bob := d.secrets.Addresses().Bob
	require.NotEqual(t, bob, currentOwner)

	proxyAdminOwner, err := batchAuthenticator.ProxyAdminOwner(callOpts)
	require.NoError(t, err)
	require.NotEqual(t, common.Address{}, proxyAdminOwner, "ProxyAdmin owner should not be zero")

	// Step 1: the current owner nominates Bob.
	tx, err := batchAuthenticator.TransferOwnership(owner, bob)
	require.NoError(t, err)
	_, err = wait.ForReceiptOK(ctx, d.L1, tx.Hash())
	require.NoError(t, err)

	pendingOwner, err := batchAuthenticator.PendingOwner(callOpts)
	require.NoError(t, err)
	require.Equal(t, bob, pendingOwner)

	// Step 2: Bob accepts.
	bobOpts, err := d.L1Transactor(ctx, d.secrets.Bob)
	require.NoError(t, err)
	tx, err = batchAuthenticator.AcceptOwnership(bobOpts)
	require.NoError(t, err)
	_, err = wait.ForReceiptOK(ctx, d.L1, tx.Hash())
	require.NoError(t, err)

	newOwner, err := batchAuthenticator.Owner(callOpts)
	require.NoError(t, err)
	require.Equal(t, bob, newOwner, "BatchAuthenticator owner should be Bob")

	// The previous owner can no longer act as the owner. The call fails at gas
	// estimation, so nothing is sent.
	_, err = batchAuthenticator.TransferOwnership(owner, owner.From)
	require.Error(t, err, "the previous owner should no longer own the BatchAuthenticator")

	// The ProxyAdmin is not affected.
	newProxyAdminOwner, err := batchAuthenticator.ProxyAdminOwner(callOpts)
	require.NoError(t, err)
	require.Equal(t, proxyAdminOwner, newProxyAdminOwner, "ProxyAdmin owner should be unchanged")

	// Check that everything still functions.
	require.NoError(t, d.RunSimpleL2Burn())
}
