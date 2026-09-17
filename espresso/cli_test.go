package espresso

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// validEnabledConfig returns a CLIConfig that passes Check with Espresso enabled;
// tests break one field at a time off this baseline.
func validEnabledConfig() CLIConfig {
	return CLIConfig{
		Enabled:                    true,
		PollInterval:               250 * time.Millisecond,
		QueryServiceURLs:           []string{"http://localhost:1234"},
		LightClientAddr:            common.HexToAddress("0x1"),
		L1URL:                      "http://localhost:8545",
		EspressoAttestationService: "http://localhost:5678",
		VerifyReceiptMaxBlocks:     DefaultVerifyReceiptMaxBlocks,
		VerifyReceiptSafetyTimeout: DefaultVerifyReceiptSafetyTimeout,
		VerifyReceiptRetryDelay:    DefaultVerifyReceiptRetryDelay,
	}
}

func TestCheck(t *testing.T) {
	require.NoError(t, validEnabledConfig().Check())

	// Disabled configs are not validated at all.
	require.NoError(t, CLIConfig{Enabled: false}.Check())
}
