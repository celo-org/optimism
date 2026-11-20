package devnet_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestSmoke brings the compose devnet up, waits for the batcher to post a
// batch that the verifier derives, and sends a simple L2 transaction to check
// that everything has started up ok.
func TestSmoke(t *testing.T) {
	_, d := StartDevnet(t, 20*time.Minute)
	require.NoError(t, d.RunSimpleL2Burn())
}
