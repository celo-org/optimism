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
		{"HTTP endpoint with path", true, "http://localhost:8080/query", false},
		{"IPv6 endpoint", true, "http://[::1]:8080", false},
		{"missing endpoint", true, "", true},
		{"blank endpoint", true, " ", true},
		{"comma list", true, "https://a.example,https://b.example", true},
		{"duplicate URL list", true, "https://a.example,https://a.example", true},
		{"missing scheme", true, "127.0.0.1:8080", true},
		{"relative URL", true, "//espresso-lb.example", true},
		{"missing host", true, "https:///query", true},
		{"unsupported scheme", true, "ftp://espresso-lb.example", true},
		{"invalid escape", true, "https://espresso-lb.example/%zz", true},
		{"unnormalized URL", true, "https://espresso-lb.example\n", true},
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
				require.ErrorContains(t, cfg.Check(), "Espresso query service URL")
			} else {
				require.NoError(t, cfg.Check())
			}
		})
	}
}

func TestReadCLIConfigQueryServiceURL(t *testing.T) {
	const endpoint = "https://espresso-lb.example"
	for _, test := range []struct {
		name string
		args []string
		env  string
	}{
		{"flag", []string{"--espresso.query-service-url", endpoint}, ""},
		{"environment", nil, endpoint},
		{"trim flag", []string{"--espresso.query-service-url", " \t" + endpoint + "\n"}, ""},
		{"trim environment", nil, " \t" + endpoint + "\n"},
		{"repeated flag uses last value", []string{"--espresso.query-service-url=https://other.example", "--espresso.query-service-url=" + endpoint}, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("TEST_ESPRESSO_QUERY_SERVICE_URL", test.env)
			app := &cli.App{
				Flags: CLIFlags("TEST", ""),
				Action: func(ctx *cli.Context) error {
					require.Equal(t, endpoint, ReadCLIConfig(ctx).QueryServiceURL)
					return nil
				},
			}
			require.NoError(t, app.Run(append([]string{"test"}, test.args...)))
		})
	}
}
