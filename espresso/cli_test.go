package espresso

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/urfave/cli/v2"
)

// validEnabledConfig returns a CLIConfig that passes Check with Espresso
// enabled; individual tests break one field at a time.
func validEnabledConfig() CLIConfig {
	return CLIConfig{
		Enabled:                       true,
		PollInterval:                  250 * time.Millisecond,
		QueryServiceURLs:              []string{"http://localhost:1234"},
		LightClientAddr:               common.HexToAddress("0x1"),
		L1URL:                         "http://localhost:8545",
		CaffeinationHeightEspressoSet: true,
		EspressoAttestationService:    "http://localhost:5678",
		VerifyReceiptMaxBlocks:        DefaultVerifyReceiptMaxBlocks,
		VerifyReceiptSafetyTimeout:    DefaultVerifyReceiptSafetyTimeout,
		VerifyReceiptRetryDelay:       DefaultVerifyReceiptRetryDelay,
	}
}

func TestCheckRequiresExplicitOriginHeightEspresso(t *testing.T) {
	cfg := validEnabledConfig()
	require.NoError(t, cfg.Check())

	cfg.CaffeinationHeightEspressoSet = false
	require.ErrorContains(t, cfg.Check(), "origin-height-espresso")

	// An explicit zero is a valid scan start (fresh chains); only the
	// inherited default is rejected.
	cfg.CaffeinationHeightEspressoSet = true
	cfg.CaffeinationHeightEspresso = 0
	require.NoError(t, cfg.Check())

	// Disabled configs are not validated at all.
	cfg = CLIConfig{Enabled: false}
	require.NoError(t, cfg.Check())
}

// configForArgs runs ReadCLIConfig through a real cli.App so that IsSet
// reflects genuine flag/env parsing rather than hand-set struct fields.
func configForArgs(t *testing.T, args ...string) CLIConfig {
	t.Helper()
	app := cli.NewApp()
	app.Flags = CLIFlags("TEST_BATCHER", "espresso")
	var config CLIConfig
	app.Action = func(ctx *cli.Context) error {
		config = ReadCLIConfig(ctx)
		return nil
	}
	require.NoError(t, app.Run(append([]string{"espresso-test"}, args...)))
	return config
}

func TestReadCLIConfigOriginHeightEspressoSet(t *testing.T) {
	t.Run("omitted", func(t *testing.T) {
		cfg := configForArgs(t)
		require.False(t, cfg.CaffeinationHeightEspressoSet)
		require.Zero(t, cfg.CaffeinationHeightEspresso)
	})

	t.Run("explicit zero via flag", func(t *testing.T) {
		cfg := configForArgs(t, fmt.Sprintf("--%s=0", CaffeinationHeightEspresso))
		require.True(t, cfg.CaffeinationHeightEspressoSet)
		require.Zero(t, cfg.CaffeinationHeightEspresso)
	})

	t.Run("non-zero via flag", func(t *testing.T) {
		cfg := configForArgs(t, fmt.Sprintf("--%s=1337", CaffeinationHeightEspresso))
		require.True(t, cfg.CaffeinationHeightEspressoSet)
		require.EqualValues(t, 1337, cfg.CaffeinationHeightEspresso)
	})

	t.Run("via env var", func(t *testing.T) {
		t.Setenv("TEST_BATCHER_ESPRESSO_ORIGIN_HEIGHT_ESPRESSO", "42")
		cfg := configForArgs(t)
		require.True(t, cfg.CaffeinationHeightEspressoSet)
		require.EqualValues(t, 42, cfg.CaffeinationHeightEspresso)
	})
}
