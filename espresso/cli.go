package espresso

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/urfave/cli/v2"
)

// espressoFlags returns the flag names for espresso
func espressoFlags(v string) string {
	return "espresso." + v
}

func espressoEnvs(envprefix, v string) []string {
	return []string{envprefix + "_ESPRESSO_" + v}
}

// Default values for batch submission receipt verification tuning.
// Defined here so that both the CLI flag defaults and the batcher logic
// can reference a single source of truth.
//
// Note: DefaultBatchAuthLookbackWindow lives in constants.go (mips64-clean
// build target shared with the derivation pipeline).
const (
	DefaultVerifyReceiptMaxBlocks        uint64        = 5
	DefaultVerifyReceiptSafetyTimeout    time.Duration = 5 * time.Minute
	DefaultVerifyReceiptRetryDelay       time.Duration = 100 * time.Millisecond
	DefaultMaxInFlightRequestsToEspresso               = 128
)

var (
	EnabledFlagName                    = espressoFlags("enabled")
	PollIntervalFlagName               = espressoFlags("poll-interval")
	QueryServiceUrlsFlagName           = espressoFlags("urls")
	LightClientAddrFlagName            = espressoFlags("light-client-addr")
	L1UrlFlagName                      = espressoFlags("l1-url")
	TestingBatcherPrivateKeyFlagName   = espressoFlags("testing-batcher-private-key")
	CaffeinationHeightEspresso         = espressoFlags("origin-height-espresso")
	CaffeinationHeightL2               = espressoFlags("origin-height-l2")
	AttestationServiceFlagName         = espressoFlags("espresso-attestation-service")
	VerifyReceiptMaxBlocksFlagName     = espressoFlags("verify-receipt-max-blocks")
	VerifyReceiptSafetyTimeoutFlagName = espressoFlags("verify-receipt-safety-timeout")
	VerifyReceiptRetryDelayFlagName    = espressoFlags("verify-receipt-retry-delay")
)

func CLIFlags(envPrefix string, category string) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:     EnabledFlagName,
			Usage:    "Enable Espresso mode",
			Value:    false,
			EnvVars:  espressoEnvs(envPrefix, "ENABLED"),
			Category: category,
		},
		&cli.DurationFlag{
			Name:     PollIntervalFlagName,
			Usage:    "Polling interval for Espresso queries",
			Value:    250 * time.Millisecond,
			EnvVars:  espressoEnvs(envPrefix, "POLL_INTERVAL"),
			Category: category,
		},
		&cli.StringSliceFlag{
			Name:     QueryServiceUrlsFlagName,
			Usage:    "Comma-separated list of Espresso query service URLs",
			EnvVars:  espressoEnvs(envPrefix, "URLS"),
			Category: category,
		},
		&cli.StringFlag{
			Name:     LightClientAddrFlagName,
			Usage:    "Address of the Espresso light client",
			EnvVars:  espressoEnvs(envPrefix, "LIGHT_CLIENT_ADDR"),
			Category: category,
		},
		&cli.StringFlag{
			Name:     L1UrlFlagName,
			Usage:    "L1 RPC URL the Espresso light client is deployed on; defaults to the batcher's L1 RPC when unset",
			EnvVars:  espressoEnvs(envPrefix, "L1_URL"),
			Category: category,
		},
		&cli.StringFlag{
			Name:     TestingBatcherPrivateKeyFlagName,
			Usage:    "Pre-approved batcher ephemeral key (testing only)",
			EnvVars:  espressoEnvs(envPrefix, "TESTING_BATCHER_PRIVATE_KEY"),
			Category: category,
		},
		&cli.Uint64Flag{
			Name: CaffeinationHeightEspresso,
			Usage: "HotShot block height the Espresso streamer starts scanning for batches from " +
				"(the Espresso-side caffeination point). Every start - including restarts - scans " +
				"forward from this height in bounded fetches, so on a long-lived chain a stale " +
				"value makes restarts replay HotShot history before any batch is posted. " +
				"Zero scans from HotShot genesis. Required when Espresso is enabled; pass 0 " +
				"explicitly if scanning from genesis is intended.",
			EnvVars:  espressoEnvs(envPrefix, "ORIGIN_HEIGHT_ESPRESSO"),
			Category: category,
		},
		&cli.Uint64Flag{
			Name: CaffeinationHeightL2,
			Usage: "L2 batch position at which the Espresso streamer starts emitting batches. " +
				"Operational parameter for restarting batchers mid-chain; set it explicitly when " +
				"taking over from the fallback batcher. " +
				"When zero, no floor is enforced and the streamer anchors at the current local-safe head. " +
				"Independent of the EspressoTime hardfork, which gates derivation semantics.",
			Value:    0,
			EnvVars:  espressoEnvs(envPrefix, "ORIGIN_HEIGHT_L2"),
			Category: category,
		},
		&cli.StringFlag{
			Name:     AttestationServiceFlagName,
			Usage:    "URL of the Espresso attestation service",
			EnvVars:  espressoEnvs(envPrefix, "ESPRESSO_ATTESTATION_SERVICE"),
			Category: category,
		},
		&cli.Uint64Flag{
			Name:     VerifyReceiptMaxBlocksFlagName,
			Usage:    "Number of HotShot blocks to wait for a submitted transaction to become queryable before re-submitting",
			Value:    DefaultVerifyReceiptMaxBlocks,
			EnvVars:  espressoEnvs(envPrefix, "VERIFY_RECEIPT_MAX_BLOCKS"),
			Category: category,
		},
		&cli.DurationFlag{
			Name:     VerifyReceiptSafetyTimeoutFlagName,
			Usage:    "Wall-clock backstop for receipt verification; re-submits the transaction if this duration is exceeded",
			Value:    DefaultVerifyReceiptSafetyTimeout,
			EnvVars:  espressoEnvs(envPrefix, "VERIFY_RECEIPT_SAFETY_TIMEOUT"),
			Category: category,
		},
		&cli.DurationFlag{
			Name:     VerifyReceiptRetryDelayFlagName,
			Usage:    "Delay between receipt verification retries",
			Value:    DefaultVerifyReceiptRetryDelay,
			EnvVars:  espressoEnvs(envPrefix, "VERIFY_RECEIPT_RETRY_DELAY"),
			Category: category,
		},
		// Note: --espresso.fallback-auth-lead-time is registered by the
		// fallback batcher in op-batcher/flags/flags.go; it is read by both
		// the fallback and the TEE batcher paths.
	}
}

type CLIConfig struct {
	Enabled                    bool
	PollInterval               time.Duration
	QueryServiceURLs           []string
	LightClientAddr            common.Address
	L1URL                      string
	TestingBatcherPrivateKey   *ecdsa.PrivateKey
	CaffeinationHeightEspresso uint64
	// CaffeinationHeightEspressoSet records whether origin-height-espresso was
	// provided explicitly (flag or env var). The flag's zero value means "scan
	// HotShot from genesis", which is almost never intended on a live chain, so
	// Check rejects an Enabled config that merely inherited the default.
	CaffeinationHeightEspressoSet bool
	CaffeinationHeightL2          uint64
	EspressoAttestationService    string

	// Batch submission receipt verification tuning
	VerifyReceiptMaxBlocks     uint64
	VerifyReceiptSafetyTimeout time.Duration
	VerifyReceiptRetryDelay    time.Duration

	// Non directly configurable option
	allowEmptyAttestationService bool `json:"-"`
}

// AllowEmptyAttestationService allows the attestation service URL to be
// empty. This is set explicitly from a public method, and isn't derivable
// from serialization or any other form other than this method.  This allows
// this setting to be configured via the code, but not externally.
func (c *CLIConfig) AllowEmptyAttestationService() {
	c.allowEmptyAttestationService = true
}

func (c CLIConfig) Check() error {
	if c.Enabled {
		// Check required fields when Espresso is enabled
		if len(c.QueryServiceURLs) == 0 {
			return fmt.Errorf("query service URLs are required when Espresso is enabled")
		}
		if c.LightClientAddr == (common.Address{}) {
			return fmt.Errorf("light client address is required when Espresso is enabled")
		}
		if c.L1URL == "" {
			return fmt.Errorf("L1 URL is required when Espresso is enabled")
		}
		if !c.allowEmptyAttestationService && c.EspressoAttestationService == "" {
			return fmt.Errorf("attestation service URL is required when Espresso is enabled")
		}
		if c.PollInterval <= 0 {
			return fmt.Errorf("poll interval must be > 0")
		}
		if !c.CaffeinationHeightEspressoSet {
			return fmt.Errorf("origin-height-espresso is required when Espresso is enabled: " +
				"it is the HotShot height batch scanning starts from, and the implicit zero " +
				"default would silently scan from HotShot genesis (pass 0 explicitly if that is intended)")
		}
		if c.VerifyReceiptMaxBlocks == 0 {
			return fmt.Errorf("verify-receipt-max-blocks must be > 0")
		}
		if c.VerifyReceiptSafetyTimeout <= 0 {
			return fmt.Errorf("verify-receipt-safety-timeout must be > 0")
		}
		if c.VerifyReceiptRetryDelay <= 0 {
			return fmt.Errorf("verify-receipt-retry-delay must be > 0")
		}
	}
	return nil
}

func ReadCLIConfig(c *cli.Context) CLIConfig {
	config := CLIConfig{
		Enabled:                       c.Bool(EnabledFlagName),
		PollInterval:                  c.Duration(PollIntervalFlagName),
		L1URL:                         c.String(L1UrlFlagName),
		CaffeinationHeightEspresso:    c.Uint64(CaffeinationHeightEspresso),
		CaffeinationHeightEspressoSet: c.IsSet(CaffeinationHeightEspresso),
		CaffeinationHeightL2:          c.Uint64(CaffeinationHeightL2),
		EspressoAttestationService:    c.String(AttestationServiceFlagName),
		VerifyReceiptMaxBlocks:        c.Uint64(VerifyReceiptMaxBlocksFlagName),
		VerifyReceiptSafetyTimeout:    c.Duration(VerifyReceiptSafetyTimeoutFlagName),
		VerifyReceiptRetryDelay:       c.Duration(VerifyReceiptRetryDelayFlagName),
	}

	config.QueryServiceURLs = c.StringSlice(QueryServiceUrlsFlagName)

	addrStr := c.String(LightClientAddrFlagName)
	config.LightClientAddr = common.HexToAddress(addrStr)

	pkStr := c.String(TestingBatcherPrivateKeyFlagName)
	pkStr = strings.TrimPrefix(pkStr, "0x")
	pk, err := crypto.HexToECDSA(pkStr)
	if err == nil {
		config.TestingBatcherPrivateKey = pk
	}

	return config
}
