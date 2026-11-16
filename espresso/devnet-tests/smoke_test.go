package devnet_tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestSmoke brings the compose devnet up, waits for the batcher to post a
// batch that the verifier derives, and sends a simple L2 transaction to check
// that everything has started up ok.
func TestSmoke(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	d := NewDevnet(ctx, t)
	require.NoError(t, d.Up())
	defer func() {
		require.NoError(t, d.Down())
	}()

	require.NoError(t, d.WaitForBatcher(ctx))
	require.NoError(t, d.RunSimpleL2Burn())
}
