package zk

import (
	"fmt"
	"os"
	"testing"
)

// TestMain skips this package on Celo. Its tests drive the ZK dispute game with kona-sp1-proposer,
// which Celo does not run: upstream kona has no Celo support and re-executes a Celo block
// that holds a user transaction to a different output root.
// Run serially against a Celo execution client, the package also exceeds the acceptance test timeout.
func TestMain(m *testing.M) {
	fmt.Println("Skipped: kona-sp1-proposer has no Celo support")
	os.Exit(0)
}
