package enclave_tests

import (
	"context"
	"os"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

// Covers EIF build, enclave boot and the first batch reaching L1.
const enclaveBatcherTimeout = 10 * time.Minute

// TestTEEBatcherUpgrade runs the Espresso batcher in a Nitro enclave built from
// TEE_BATCHER_APP_IMAGE, replaces it with one built from TEE_BATCHER_UPGRADE_IMAGE, and
// checks that the new batcher picks up the L2 blocks the old one left unsubmitted.
func TestTEEBatcherUpgrade(t *testing.T) {
	baseImage := os.Getenv("TEE_BATCHER_APP_IMAGE")
	upgradeImage := os.Getenv("TEE_BATCHER_UPGRADE_IMAGE")
	if baseImage == "" || upgradeImage == "" {
		t.Skip("TEE_BATCHER_APP_IMAGE and TEE_BATCHER_UPGRADE_IMAGE must be set")
	}

	ctx := t.Context()
	sys, _, err := new(env.EspressoDevNodeLauncherDocker).StartE2eDevnet(ctx, t, env.WithBatcherStoppedInitially())
	require.NoError(t, err)
	defer env.Stop(t, sys)

	args := enclaveBatcherArgs(t, sys)

	base := runEnclaveBatcher(t, baseImage, args)
	env.RunSimpleL2BurnWithTimeout(ctx, t, sys, enclaveBatcherTimeout)

	require.NoError(t, stopEnclaveBatcher(base))
	l2Seq := sys.NodeClient(e2esys.RoleSeq)
	l2Verif := sys.NodeClient(e2esys.RoleVerif)
	alice := sys.Cfg.Secrets.Alice
	nonce, err := l2Seq.NonceAt(ctx, crypto.PubkeyToAddress(alice.PublicKey), nil)
	require.NoError(t, err)
	pending := helpers.SendL2Tx(t, sys.Cfg, l2Seq, alice, env.L2TxWithNonce(nonce))
	_, err = l2Verif.TransactionReceipt(ctx, pending.TxHash)
	require.ErrorIs(t, err, ethereum.NotFound)

	runEnclaveBatcher(t, upgradeImage, args)
	waitCtx, cancel := context.WithTimeout(ctx, enclaveBatcherTimeout)
	defer cancel()
	_, err = wait.ForReceiptOK(waitCtx, l2Verif, pending.TxHash)
	require.NoError(t, err)
	env.RunSimpleL2Burn(ctx, t, sys)
}
