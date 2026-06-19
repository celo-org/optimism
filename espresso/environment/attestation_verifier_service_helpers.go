package environment

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
)

// ErrorAttestationConfigFieldNotSet is an error that indicates that specific
// configuration value for the AttestationVerifierServiceConfig struct
// is not set.
type ErrorAttestationConfigFieldNotSet struct {
	FieldName string
}

// Error implements error
func (e ErrorAttestationConfigFieldNotSet) Error() string {
	return fmt.Sprintf("\"%s\" is not set for \"AttestationVerifierServiceConfig\"", e.FieldName)
}

// AttestationVerifierServiceConfig is a struct that contatins all of the
// configuration options / values for the Attestation Verifier Service
// configuration.
type AttestationVerifierServiceConfig struct {
	networkRPCURL         string
	sp1Prover             string
	nitroVerifierAddress  string
	networkUseDocker      string
	skipTimeValidityCheck string
	rustLog               string
	networkPrivateKey     string
	rpcURL                string
	host                  string
	port                  string
	dockerImage           string
}

// applyOptions is a convenience method for quickly applying options against
// the configuration.
func (c *AttestationVerifierServiceConfig) applyOptions(options ...AttestationVerifierServiceOption) {
	for _, opt := range options {
		opt(c)
	}
}

// Verify performs a basic verification of the values stored within the
// AttestationVerifierServiceConfig struct.
func (c *AttestationVerifierServiceConfig) Verify(ct *E2eDevnetLauncherContext) {
	if ct.Error != nil {
		// Early Return if we already have an Error set
		return
	}

	// Now launch the attestation verifier zk server
	// Now we need to launch the attestation verifier zk server
	fmt.Println("Starting attestation verifier zk server...")

	if c.networkRPCURL == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"networkRPCURL"}
		return
	}

	if c.sp1Prover == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"sp1Prover"}
		return
	}

	if c.nitroVerifierAddress == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"nitroVerifierAddress"}
		return
	}

	if c.networkUseDocker == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"networkUseDocker"}
		return
	}

	if c.skipTimeValidityCheck == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"skipTimeValidityCheck"}
		return
	}

	if c.rustLog == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"rustLog"}
		return
	}

	if c.networkPrivateKey == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"networkPrivateKey"}
		return
	}

	if c.rpcURL == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"rpcURL"}
		return
	}

	if c.host == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"host"}
		return
	}

	if c.port == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"port"}
		return
	}

	if c.dockerImage == "" {
		ct.Error = ErrorAttestationConfigFieldNotSet{"dockerImage"}
		return
	}
}

// AttestationVerifierServiceOption represents a functional option that allows
// for the modification / configuration of the Attestation Verifier Service
// in a flexible manner.
type AttestationVerifierServiceOption func(*AttestationVerifierServiceConfig)

// WithAttestationServiceVerifierNetworkRPCURL configures the Network RPC URL
// for the Attestation Verifier Service.
func WithAttestationServiceVerifierNetworkRPCURL(networkRPCURL string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.networkRPCURL = networkRPCURL
	}
}

// WithAttestationServiceVerifierSP1Prover configures the SP1 Provider
// for the Attstation Verifier Service.
func WithAttestationServiceVerifierSP1Prover(sp1Prover string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.sp1Prover = sp1Prover
	}
}

// WithAttestationServiceVerifierNitroVerifierAddress configures the
// Nitro Verifier Address for the Attestation Verifier Service
func WithAttestationServiceVerifierNitroVerifierAddress(nitroVerifierAddress string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.nitroVerifierAddress = nitroVerifierAddress
	}
}

// WithAttestationServiceVerifierNetworkUseDocker configures the
// Network Use Docker configuration for the Attestation Verifier Service.
func WithAttestationServiceVerifierNetworkUseDocker(networkUseDocker string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.networkUseDocker = networkUseDocker
	}
}

// WithAttestationServiceVerifierSkipTimeValidityCheck configures the
// Skip Time Validity Check configuration for the Attestation Verifier
// Service.
func WithAttestationServiceVerifierSkipTimeValidityCheck(skipTimeValidityCheck string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.skipTimeValidityCheck = skipTimeValidityCheck
	}
}

// WithAttestationServiceVerifierRustLog configures the Rust Log
// configuration for the Attestation Verifier Service.
func WithAttestationServiceVerifierRustLog(rustLog string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.rustLog = rustLog
	}
}

// WithAttestationServiceVerifierNetworkPrivateKey configurs the network
// private key for the Attestation Verifier Service.
func WithAttestationServiceVerifierNetworkPrivateKey(networkPrivateKey string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.networkPrivateKey = networkPrivateKey
	}
}

// WithAttestationServiceVerifierRPCURL configures the RPC URL for the
// AttestationVerifier Service.
func WithAttestationServiceVerifierRPCURL(rpcURL string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.rpcURL = rpcURL
	}
}

// WithAttestationServiceVerifierHost configures the Host configuration for
// the Attestation Verifier Service.
func WithAttestationServiceVerifierHost(host string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.host = host
	}
}

// WithAttestationServiceVerifierPort configures the Port configuration for
// the Attestation Verifier Service.
func WithAttestationServiceVerifierPort(port string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.port = port
	}
}

// WithAttestationServiceVerifierDockerImage configures the Docker Image
// configuration for the Attestation Verifier SErvice.
func WithAttestationServiceVerifierDockerImage(dockerImage string) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.dockerImage = dockerImage
	}
}

// WithAttestationServiceVerifierOptions applies multiple options as a single
// option.  This is provided for convenience, and nothing else.
func WithAttestationServiceVerifierOptions(options ...AttestationVerifierServiceOption) AttestationVerifierServiceOption {
	return func(c *AttestationVerifierServiceConfig) {
		c.applyOptions(options...)
	}
}

// WithAttestationConfigFromENV is an option that will populate and overwrite
// the configuration for the Attestation Service Verifier with values taken
// from the ENV variables, should they be present.
func WithAttestationConfigFromENV() AttestationVerifierServiceOption {
	var options []AttestationVerifierServiceOption

	if networkRPCURL := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_NETWORK_RPC_URL"); networkRPCURL != "" {
		options = append(options, WithAttestationServiceVerifierNetworkRPCURL(networkRPCURL))
	}

	if sp1Prover := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_SP1_PROVER"); sp1Prover != "" {
		options = append(options, WithAttestationServiceVerifierSP1Prover(sp1Prover))
	}

	if nitroVerifierAddress := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_NITRO_VERIFIER_ADDRESS"); nitroVerifierAddress != "" {
		options = append(options, WithAttestationServiceVerifierNitroVerifierAddress(nitroVerifierAddress))
	}

	if useDocker := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_NETWORK_USE_DOCKER"); useDocker != "" {
		options = append(options, WithAttestationServiceVerifierNetworkUseDocker(useDocker))
	}

	if skipTimeValidityCheck := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_SKIP_TIME_VALIDITY_CHECK"); skipTimeValidityCheck != "" {
		options = append(options, WithAttestationServiceVerifierSkipTimeValidityCheck(skipTimeValidityCheck))
	}

	if rustLog := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_RUST_LOG"); rustLog != "" {
		options = append(options, WithAttestationServiceVerifierRustLog(rustLog))
	}

	if networkPrivateKey := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_NETWORK_PRIVATE_KEY"); networkPrivateKey != "" {
		options = append(options, WithAttestationServiceVerifierNetworkPrivateKey(networkPrivateKey))
	}

	if rpcURL := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_RPC_URL"); rpcURL != "" {
		options = append(options, WithAttestationServiceVerifierRPCURL(rpcURL))
	}

	if host := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_HOST"); host != "" {
		options = append(options, WithAttestationServiceVerifierHost(host))
	}

	if port := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_PORT"); port != "" {
		options = append(options, WithAttestationServiceVerifierPort(port))
	}

	if dockerImage := os.Getenv("ESPRESSO_ATTESTATION_VERIFIER_DOCKER_IMAGE"); dockerImage != "" {
		options = append(options, WithAttestationServiceVerifierDockerImage(dockerImage))
	}

	return WithAttestationServiceVerifierOptions(options...)
}

// launchEspressoAttestationVerifierServiceDockerContainer is a StartOption that
// ensures that the Espresso Attestation Verifier Service is launched in its
// own docker container.
//
// It will launch the service, and modify the Batcher CLIConfig with the
// configured parameters.
func launchEspressoAttestationVerifierServiceDockerContainer(ct *E2eDevnetLauncherContext, options ...AttestationVerifierServiceOption) e2esys.StartOption {
	return e2esys.StartOption{
		Role: "launch-espresso-attestation-verifier",
		BatcherMod: func(c *batcher.CLIConfig, sys *e2esys.System) {
			if ct.Error != nil {
				// Early Return if we already have an Error set
				return
			}

			// These are the default configuration values.
			// These values are based on those contained within the
			// ".env" file.
			cfg := AttestationVerifierServiceConfig{
				networkRPCURL:         "https://rpc.mainnet.succinct.xyz",
				sp1Prover:             "mock",
				nitroVerifierAddress:  "0x2D7fbBAD6792698Ba92e67b7e180f8010B9Ec788",
				networkUseDocker:      "1",
				skipTimeValidityCheck: "true",
				rustLog:               "string",
				networkPrivateKey:     "0x71f8e55f7555c946eadd5a2b5897465a9813b3ee493d6ef4ba6f1505a6e97af3",
				host:                  "0.0.0.0",
				port:                  "8080",
			}

			// Apply all Environment Variable modifications
			cfg.applyOptions(WithAttestationConfigFromENV())

			// Apply the options
			cfg.applyOptions(options...)

			// Verify the options
			cfg.Verify(ct)

			if ct.Error != nil {
				// Early return, as we have an error in our configuration
				return
			}

			dockerConfig := DockerContainerConfig{
				Image:   cfg.dockerImage,
				Network: determineDockerNetworkMode(),
				Ports: []string{
					cfg.port,
				},
				Name: "attestation-verifier-zk",
				Environment: map[string]string{
					"NETWORK_RPC_URL":          cfg.networkRPCURL,
					"SP1_PROVER":               cfg.sp1Prover,
					"NITRO_VERIFIER_ADDRESS":   cfg.nitroVerifierAddress,
					"USE_DOCKER":               cfg.networkUseDocker,
					"SKIP_TIME_VALIDITY_CHECK": cfg.skipTimeValidityCheck,
					"RUST_LOG":                 cfg.rustLog,
					"NETWORK_PRIVATE_KEY":      cfg.networkPrivateKey,
					"RPC_URL":                  cfg.rpcURL,
					"HOST":                     cfg.host,
					"PORT":                     cfg.port,
				},
			}
			containerCli := new(DockerCli)

			attestationVerifierInfo, err := containerCli.LaunchContainer(ct.Ctx, dockerConfig)
			if err != nil {
				ct.Error = FailedToLaunchDockerContainer{Cause: err}
				return
			}

			// Get the actual mapped port
			ports := attestationVerifierInfo.PortMap[cfg.port]
			if len(ports) == 0 {
				ct.Error = fmt.Errorf("no port mapping found for attestation verifier")
				return
			}

			healthCheckCtx, cancel := context.WithTimeout(ct.Ctx, 60*time.Second)
			defer cancel()

			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			attestationHostPort, err := getContainerRemappedHostPort(ports[0])
			if err != nil {
				ct.Error = err
				return
			}

			// Use the actual host:port for health check
			attestationURL := "http://" + attestationHostPort

			// Replace the EspressoDevNode with the wrapped Dev Node,
			// so we can tie into the cleanup stage.
			ct.EspressoDevNode = &EspressoDevNodeWithAttestationVerifier{
				EspressoDevNode:            ct.EspressoDevNode,
				AttestationVerifierService: attestationVerifierInfo,
			}

			c.Espresso.EspressoAttestationService = attestationURL
			healthCheckURL := attestationURL + "/health"
			for {
				select {
				case <-healthCheckCtx.Done():
					ct.Error = fmt.Errorf("attestation verifier did not become healthy: %w", healthCheckCtx.Err())
					return
				case <-ticker.C:
					resp, err := http.Get(healthCheckURL)
					if resp != nil {
						_ = resp.Body.Close()
					}

					if err == nil && resp.StatusCode == http.StatusOK {
						// We are done waiting, we have a good response, and
						// the service seems to be healthy
						return
					}
				}
			}
		},
	}
}

// WithEspressoAttestationVerifierService is a Devnet option that ensures that
// the Docker Container image is up and running before the Batcher is
// launched.
func WithEspressoAttestationVerifierService() E2eDevnetLauncherOption {
	return func(ct *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				launchEspressoAttestationVerifierServiceDockerContainer(ct),
			},
		}
	}
}

// EspressoDevNodeWithAttestationVerifier  is a simple struct meant to wrap
// an existing EspressoDevNode and add its own container for reference and
// removal on cleanup.
type EspressoDevNodeWithAttestationVerifier struct {
	EspressoDevNode
	AttestationVerifierService DockerContainerInfo
}

// Stop overwrites and implements EspressoDevNode
func (w *EspressoDevNodeWithAttestationVerifier) Stop() error {
	dockerCli := new(DockerCli)

	err := dockerCli.StopContainer(context.Background(), w.AttestationVerifierService.ContainerID)

	// Always try to shut down the Espresso Dev Node
	if err := w.EspressoDevNode.Stop(); err != nil {
		return err
	}

	return err
}
