package espresso

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCLIConfigQueryServiceURL(t *testing.T) {
	for _, test := range []struct {
		name    string
		enabled bool
		url     string
		wantErr bool
	}{
		{"disabled", false, "", false},
		{"single endpoint", true, "https://espresso-lb.example", false},
		{"missing endpoint", true, "", true},
		{"blank endpoint", true, " ", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg := CLIConfig{
				Enabled:                    test.enabled,
				QueryServiceURL:            test.url,
				LightClientAddr:            common.Address{1},
				L1URL:                      "https://l1.example",
				EspressoAttestationService: "https://attestation.example",
				PollInterval:               time.Second,
				VerifyReceiptMaxBlocks:     DefaultVerifyReceiptMaxBlocks,
				VerifyReceiptSafetyTimeout: DefaultVerifyReceiptSafetyTimeout,
				VerifyReceiptRetryDelay:    DefaultVerifyReceiptRetryDelay,
			}
			if test.wantErr {
				require.ErrorContains(t, cfg.Check(), "Espresso query service URL is required")
			} else {
				require.NoError(t, cfg.Check())
			}
		})
	}
}

func TestReadCLIConfigQueryServiceURL(t *testing.T) {
	const endpoint = "https://espresso-lb.example"
	for _, source := range []string{"flag", "environment"} {
		t.Run(source, func(t *testing.T) {
			t.Setenv("TEST_ESPRESSO_URL", "")
			args := []string{"test"}
			if source == "flag" {
				args = append(args, "--espresso.url", endpoint)
			} else {
				t.Setenv("TEST_ESPRESSO_URL", endpoint)
			}
			app := &cli.App{
				Flags: CLIFlags("TEST", ""),
				Action: func(ctx *cli.Context) error {
					require.Equal(t, endpoint, ReadCLIConfig(ctx).QueryServiceURL)
					return nil
				},
			}
			require.NoError(t, app.Run(args))
		})
	}
}
