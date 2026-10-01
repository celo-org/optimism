package devnet_tests

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/stretchr/testify/require"
)

func TestTEEBatcherUpgrade(t *testing.T) {
	baseImage := os.Getenv("TEE_BATCHER_APP_IMAGE")
	upgradeImage := os.Getenv("TEE_BATCHER_UPGRADE_IMAGE")
	if baseImage == "" || upgradeImage == "" {
		t.Skip("TEE batcher image variables are only set by the upgrade workflow")
	}
	require.NotEqual(t, baseImage, upgradeImage)

	profile := ProfileFromEnv(t)
	require.Equal(t, TEE, profile)

	t.Run(string(profile), func(t *testing.T) {
		ctx := t.Context()
		devnet := NewDevnet(ctx, t, profile)
		require.NoError(t, devnet.Up())
		defer func() { require.NoError(t, devnet.Down()) }()

		require.NoError(t, devnet.WaitForBatcher(ctx))
		require.NoError(t, devnet.RunSimpleL2Burn())

		require.NoError(t, devnet.ServiceDown(OpBatcher))
		pending, err := devnet.SubmitSimpleL2Burn()
		require.NoError(t, err)
		_, err = devnet.L2Verif.TransactionReceipt(ctx, pending.Receipt.TxHash)
		require.ErrorIs(t, err, ethereum.NotFound)

		t.Setenv("TEE_BATCHER_APP_IMAGE", upgradeImage)
		require.NoError(t, devnet.ServiceUp(OpBatcher))
		require.NoError(t, devnet.WaitForBatcher(ctx))
		require.NoError(t, devnet.VerifySimpleL2Burn(pending))
		require.NoError(t, devnet.RunSimpleL2Burn())
	})
}
