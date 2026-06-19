package environment

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/espresso"
	"github.com/ethereum-optimism/optimism/espresso/bindings"
	altda "github.com/ethereum-optimism/optimism/op-alt-da"
	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
	batcherCfg "github.com/ethereum-optimism/optimism/op-batcher/config"
	"github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-service/endpoint"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/oppprof"
	"github.com/ethereum-optimism/optimism/op-service/rpc"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

const (
	ENCLAVE_INTERMEDIATE_IMAGE_TAG = "op-batcher-enclave:tests"
	ENCLAVE_IMAGE_TAG              = "op-batcher-enclaver:tests"
	ESPRESSO_RUN_ENCLAVE_TESTS     = "ESPRESSO_RUN_ENCLAVE_TESTS"

	// TeeTypeNitro corresponds to IEspressoTEEVerifier.TeeType.NITRO enum value
	TeeTypeNitro uint8 = 0
)

func HasTee() (bool, error) {
	_, hasTee := os.LookupEnv(ESPRESSO_RUN_ENCLAVE_TESTS)
	if hasTee {
		if _, err := os.Stat("/dev/nitro_enclaves"); os.IsNotExist(err) {
			return false, fmt.Errorf("/dev/nitro_enclaves does not exist; cannot run enclave tests without Nitro Enclaves support")
		}
	}
	return hasTee, nil
}

// Skips the calling test if `ESPRESSO_RUN_ENCLAVE_TESTS` is not set.
func RunOnlyWithEnclave(t *testing.T) {
	_, doRun := os.LookupEnv(ESPRESSO_RUN_ENCLAVE_TESTS)
	if !doRun {
		t.SkipNow()
	}
}

// Formats a configuration flag name and it's value for use in commandline,
// then adds to the args slice.
// Example: appendArg(&args, "people", []{"Alice", "Bob"}) will append
// {'--people', 'Alice,Bob'} to args
func appendArg(args *[]string, flagName string, value any) {
	boolValue, isBool := value.(bool)
	if isBool {
		if boolValue {
			*args = append(*args, fmt.Sprintf("--%s", flagName))
		}
		return
	}

	strSliceValue, isStrSlice := value.([]string)
	if isStrSlice {
		if len(strSliceValue) > 0 {
			*args = append(*args, fmt.Sprintf("--%s", flagName), strings.Join(strSliceValue, ","))
		}
		return
	}

	formattedValue := fmt.Sprintf("%v", value)
	if formattedValue != "" {
		*args = append(*args, fmt.Sprintf("--%s", flagName), formattedValue)
	}
}

// LaunchBatcherInEnclave is an E2eDevnetLauncherOption that configures the
// batcher to not launch in the standard e2e devnet, and to launch the batcher
// externally in an Enclave.
//
// Adding this option automatically configures the Espresso Attestation
// Service Service with its default configuration. This is required to run in
// an Enclave, as such, it is better to pre-configure the option instead of
// allowing for the potential of an error to occur due to not including the
// other Option.
//
// This LauncherOption explicitly creates a Batcher to run in the Enclave based
// on the configuration of the batcher that would be created and started
// locally. The locally created Batcher in the E2e System is never meant to
// actually run with this option, and instead the External Batcher is meant
// to be run instead.
func LaunchBatcherInEnclave() E2eDevnetLauncherOption {
	return func(ct *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			// | NOTE:  while this option initially disables the batcher for
			//   the purposes of being started later, it is the intention of
			//   this Launchger Option to tie the Batcher as an external
			//   connection, rather than the local testing one.  As a result
			//   The local Batcher should not be accessed / inspecting /
			//   interacted with for the purposes of any tests that are
			//   utilizing this Launcher Option.
			SystemConfigOption: SystemConfigOptionDisableBatcher,
			SystemConfigOpt:    e2esys.WithAllocType(config.AllocTypeEspressoWithEnclave),
			StartOptions: []e2esys.StartOption{
				launchEspressoAttestationVerifierServiceDockerContainer(ct),
				{
					Role: "launch-batcher-in-enclave",

					BatcherMod: func(c *batcher.CLIConfig, sys *e2esys.System) {
						// We will manually convert CLIConfig back to commandline arguments
						var args []string

						// Enclave batcher requires valid throttle config (upper > lower). System config
						// often has zero throttle; use flag defaults only for the enclave so integration
						// tests (non-enclave batcher) are unchanged.
						throttle := c.ThrottleConfig
						if throttle.UpperThreshold <= throttle.LowerThreshold {
							throttle.ControllerType = batcherCfg.ThrottleControllerType(flags.DefaultThrottleControllerType)
							throttle.LowerThreshold = flags.DefaultThrottleLowerThreshold
							throttle.UpperThreshold = flags.DefaultThrottleUpperThreshold
							throttle.TxSizeLowerLimit = flags.DefaultThrottleTxSizeLowerLimit
							throttle.TxSizeUpperLimit = flags.DefaultThrottleTxSizeUpperLimit
							throttle.BlockSizeLowerLimit = flags.DefaultThrottleBlockSizeLowerLimit
							throttle.BlockSizeUpperLimit = flags.DefaultThrottleBlockSizeUpperLimit
						}

						// We don't want to stop this batcher
						appendArg(&args, flags.StoppedFlag.Name, false)

						// These flags require separate handling: we want to use HTTP endpoints,
						// as Odyn proxy inside the enclave doesn't support websocket
						l1Rpc := sys.L1.UserRPC().(endpoint.HttpRPC).HttpRPC()
						appendArg(&args, flags.L1EthRpcFlag.Name, l1Rpc)
						appendArg(&args, txmgr.L1RPCFlagName, l1Rpc)
						appendArg(&args, espresso.L1UrlFlagName, l1Rpc)
						appendArg(&args, espresso.RollupL1UrlFlagName, l1Rpc)
						l2EthRpc := sys.EthInstances[e2esys.RoleSeq].UserRPC().(endpoint.HttpRPC).HttpRPC()
						appendArg(&args, flags.L2EthRpcFlag.Name, l2EthRpc)
						rollupRpc := sys.RollupNodes[e2esys.RoleSeq].UserRPC().(endpoint.HttpRPC).HttpRPC()
						appendArg(&args, flags.RollupRpcFlag.Name, rollupRpc)

						// Batcher flags
						appendArg(&args, flags.ActiveSequencerCheckDurationFlag.Name, c.ActiveSequencerCheckDuration)
						appendArg(&args, flags.ApproxComprRatioFlag.Name, c.ApproxComprRatio)
						appendArg(&args, flags.BatchTypeFlag.Name, c.BatchType)
						appendArg(&args, flags.CheckRecentTxsDepthFlag.Name, c.CheckRecentTxsDepth)
						appendArg(&args, flags.CompressionAlgoFlag.Name, c.CompressionAlgo.String())
						appendArg(&args, flags.CompressorFlag.Name, c.Compressor)
						appendArg(&args, flags.DataAvailabilityTypeFlag.Name, c.DataAvailabilityType.String())
						appendArg(&args, flags.MaxBlocksPerSpanBatch.Name, c.MaxBlocksPerSpanBatch)
						appendArg(&args, flags.MaxChannelDurationFlag.Name, c.MaxChannelDuration)
						appendArg(&args, flags.MaxL1TxSizeBytesFlag.Name, c.MaxL1TxSize)
						appendArg(&args, flags.MaxPendingTransactionsFlag.Name, c.MaxPendingTransactions)
						appendArg(&args, flags.PollIntervalFlag.Name, c.PollInterval)
						appendArg(&args, flags.AdditionalThrottlingEndpointsFlag.Name, strings.Join(throttle.AdditionalEndpoints, ","))
						appendArg(&args, flags.SubSafetyMarginFlag.Name, c.SubSafetyMargin)
						appendArg(&args, flags.TargetNumFramesFlag.Name, c.TargetNumFrames)
						appendArg(&args, flags.ThrottleBlockSizeLowerLimitFlag.Name, throttle.BlockSizeLowerLimit)
						appendArg(&args, flags.ThrottleBlockSizeUpperLimitFlag.Name, throttle.BlockSizeUpperLimit)
						appendArg(&args, flags.ThrottleUsafeDABytesLowerThresholdFlag.Name, throttle.LowerThreshold)
						appendArg(&args, flags.ThrottleUsafeDABytesUpperThresholdFlag.Name, throttle.UpperThreshold)
						appendArg(&args, flags.ThrottleTxSizeLowerLimitFlag.Name, throttle.TxSizeLowerLimit)
						appendArg(&args, flags.ThrottleTxSizeUpperLimitFlag.Name, throttle.TxSizeUpperLimit)
						appendArg(&args, flags.ThrottleControllerTypeFlag.Name, string(throttle.ControllerType))
						appendArg(&args, flags.WaitNodeSyncFlag.Name, c.WaitNodeSync)

						// TxMgr flags
						appendArg(&args, txmgr.MnemonicFlagName, c.TxMgrConfig.Mnemonic)
						appendArg(&args, txmgr.HDPathFlagName, c.TxMgrConfig.HDPath)
						appendArg(&args, txmgr.SequencerHDPathFlag.Name, c.TxMgrConfig.SequencerHDPath)
						appendArg(&args, txmgr.L2OutputHDPathFlag.Name, c.TxMgrConfig.L2OutputHDPath)
						appendArg(&args, txmgr.PrivateKeyFlagName, c.TxMgrConfig.PrivateKey)
						appendArg(&args, txmgr.NumConfirmationsFlagName, c.TxMgrConfig.NumConfirmations)
						appendArg(&args, txmgr.SafeAbortNonceTooLowCountFlagName, c.TxMgrConfig.SafeAbortNonceTooLowCount)
						appendArg(&args, txmgr.FeeLimitMultiplierFlagName, c.TxMgrConfig.FeeLimitMultiplier)
						appendArg(&args, txmgr.FeeLimitThresholdFlagName, c.TxMgrConfig.FeeLimitThresholdGwei)
						appendArg(&args, txmgr.MinBaseFeeFlagName, c.TxMgrConfig.MinBaseFeeGwei)
						appendArg(&args, txmgr.MinTipCapFlagName, c.TxMgrConfig.MinTipCapGwei)
						appendArg(&args, txmgr.ResubmissionTimeoutFlagName, c.TxMgrConfig.ResubmissionTimeout)
						appendArg(&args, txmgr.ReceiptQueryIntervalFlagName, c.TxMgrConfig.ReceiptQueryInterval)
						appendArg(&args, txmgr.NetworkTimeoutFlagName, c.TxMgrConfig.NetworkTimeout)
						appendArg(&args, txmgr.TxNotInMempoolTimeoutFlagName, c.TxMgrConfig.TxNotInMempoolTimeout)
						appendArg(&args, txmgr.TxSendTimeoutFlagName, c.TxMgrConfig.TxSendTimeout)

						// Log flags
						appendArg(&args, log.LevelFlagName, c.LogConfig.Level)
						appendArg(&args, log.ColorFlagName, c.LogConfig.Color)
						appendArg(&args, log.FormatFlagName, c.LogConfig.Format.String())
						appendArg(&args, log.PidFlagName, c.LogConfig.Pid)

						// Metrics flags
						appendArg(&args, metrics.EnabledFlagName, c.MetricsConfig.Enabled)
						appendArg(&args, metrics.ListenAddrFlagName, c.MetricsConfig.ListenAddr)
						appendArg(&args, metrics.PortFlagName, c.MetricsConfig.ListenPort)

						// Pprof flags
						appendArg(&args, oppprof.EnabledFlagName, c.PprofConfig.ListenEnabled)
						appendArg(&args, oppprof.ListenAddrFlagName, c.PprofConfig.ListenAddr)
						appendArg(&args, oppprof.PortFlagName, c.PprofConfig.ListenPort)
						appendArg(&args, oppprof.ProfileTypeFlagName, c.PprofConfig.ProfileType.String())
						appendArg(&args, oppprof.ProfilePathFlagName, c.PprofConfig.ProfileDir+"/"+c.PprofConfig.ProfileFilename)

						// RPC flags
						appendArg(&args, rpc.ListenAddrFlagName, c.RPC.ListenAddr)
						appendArg(&args, rpc.PortFlagName, c.RPC.ListenPort)
						appendArg(&args, rpc.EnableAdminFlagName, c.RPC.EnableAdmin)

						// AltDA flags
						appendArg(&args, altda.EnabledFlagName, c.AltDA.Enabled)
						appendArg(&args, altda.DaServerAddressFlagName, c.AltDA.DAServerURL)
						appendArg(&args, altda.VerifyOnReadFlagName, c.AltDA.VerifyOnRead)
						appendArg(&args, altda.PutTimeoutFlagName, c.AltDA.PutTimeout)
						appendArg(&args, altda.GetTimeoutFlagName, c.AltDA.GetTimeout)
						appendArg(&args, altda.MaxConcurrentRequestsFlagName, c.AltDA.MaxConcurrentRequests)

						// Espresso flags
						appendArg(&args, espresso.EnabledFlagName, c.Espresso.Enabled)
						appendArg(&args, espresso.PollIntervalFlagName, c.Espresso.PollInterval)
						appendArg(&args, espresso.LightClientAddrFlagName, c.Espresso.LightClientAddr)
						appendArg(&args, espresso.TestingBatcherPrivateKeyFlagName, hexutil.Encode(crypto.FromECDSA(c.Espresso.TestingBatcherPrivateKey)))
						for _, url := range c.Espresso.QueryServiceURLs {
							appendArg(&args, espresso.QueryServiceUrlsFlagName, url)
						}
						appendArg(&args, espresso.AttestationServiceFlagName, c.Espresso.EspressoAttestationService)
						appendArg(&args, espresso.BatchAuthenticatorAddrFlagName, c.Espresso.BatchAuthenticatorAddr)

						err := SetupEnclaver(ct.Ctx, sys, args...)
						if err != nil {
							panic(fmt.Sprintf("failed to setup enclaver: %v", err))
						}

						cli := new(EnclaverCli)
						cli.RunEnclave(ct.Ctx, ENCLAVE_IMAGE_TAG)
					},
				},
			},
		}
	}
}

// Builds docker and enclaver EIF image for op-batcher and registers EIF's PCR0 with
// EspressoNitroTEEVerifier. args... are command-line arguments to op-batcher
// to be baked into the image.
func SetupEnclaver(ctx context.Context, sys *e2esys.System, args ...string) error {
	// Build underlying batcher docker image with baked-in arguments
	dockerCli := new(DockerCli)
	err := dockerCli.Build(ctx,
		ENCLAVE_INTERMEDIATE_IMAGE_TAG,
		"../../ops/docker/op-stack-go/Dockerfile",
		"op-batcher-enclave-target",
		"../../",
		DockerBuildArg{
			Name:  "ENCLAVE_BATCHER_ARGS",
			Value: strings.Join(args, " "),
		})
	if err != nil {
		return fmt.Errorf("failed to build docker image: %w", err)
	}

	// Build EIF image based on the docker image we just built
	enclaverCli := new(EnclaverCli)
	manifest := DefaultManifest("op-batcher", ENCLAVE_IMAGE_TAG, ENCLAVE_INTERMEDIATE_IMAGE_TAG)
	measurements, err := enclaverCli.BuildEnclave(ctx, manifest)
	if err != nil {
		return fmt.Errorf("failed to build enclave image: %w", err)
	}
	pcr0Bytes, err := hexutil.Decode("0x" + measurements.PCR0)
	if err != nil {
		return fmt.Errorf("failed to decode PCR0: %w", err)
	}

	return RegisterEnclaveHash(ctx, sys, pcr0Bytes)
}

// RegisterEnclaveHash registers the enclave PCR0 hash with the EspressoNitroTEEVerifier.
func RegisterEnclaveHash(ctx context.Context, sys *e2esys.System, pcr0Bytes []byte) error {
	l1Client := sys.NodeClient(e2esys.RoleL1)
	authenticator, err := bindings.NewBatchAuthenticator(sys.RollupConfig.BatchAuthenticatorAddress, l1Client)
	if err != nil {
		return fmt.Errorf("failed to create batch authenticator: %w", err)
	}

	verifierAddress, err := authenticator.EspressoTEEVerifier(&bind.CallOpts{})
	if err != nil {
		return fmt.Errorf("failed to get verifier address: %w", err)
	}

	verifier, err := bindings.NewEspressoTEEVerifier(verifierAddress, l1Client)
	if err != nil {
		return fmt.Errorf("failed to create verifier: %w", err)
	}

	opts, err := bind.NewKeyedTransactorWithChainID(sys.Cfg.Secrets.Deployer, sys.Cfg.L1ChainIDBig())
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	// SetEnclaveHash must be called through EspressoTEEVerifier wrapper because
	// NitroTEEVerifier.setEnclaveHash has onlyTEEVerifier modifier, restricting calls
	// to only the TEEVerifier contract. The wrapper has onlyGuardianOrOwner permissions.
	registrationTx, err := verifier.SetEnclaveHash(opts, crypto.Keccak256Hash(pcr0Bytes), true, TeeTypeNitro)
	if err != nil {
		return fmt.Errorf("failed to create registration transaction: %w", err)
	}

	receipt, err := geth.WaitForTransaction(registrationTx.Hash(), l1Client, 2*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to wait for registration transaction: %w", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("registration transaction failed")
	}

	return nil
}

type EnclaverManifestSources struct {
	App string `yaml:"app"`
}

type EnclaverManifestDefaults struct {
	CpuCount uint `yaml:"cpu_count"`
	MemoryMb uint `yaml:"memory_mb"`
}

type EnclaverManifestKmsProxy struct {
	ListenPort uint16 `yaml:"listen_port,omitempty"`
}

type EnclaverManifestEgress struct {
	Allow     []string `yaml:"allow"`
	Deny      []string `yaml:"deny"`
	ProxyPort uint16   `yaml:"proxy_port,omitempty"`
}

type EnclaverManifestIngress struct {
	ListenPort uint16 `yaml:"listen_port"`
}

type EnclaverManifest struct {
	Version  string                    `yaml:"version"`
	Name     string                    `yaml:"name"`
	Target   string                    `yaml:"target"`
	Sources  *EnclaverManifestSources  `yaml:"sources,omitempty"`
	Defaults *EnclaverManifestDefaults `yaml:"defaults,omitempty"`
	KmsProxy *EnclaverManifestKmsProxy `yaml:"kms_proxy,omitempty"`
	Egress   *EnclaverManifestEgress   `yaml:"egress,omitempty"`
	Ingress  []EnclaverManifestIngress `yaml:"ingress"`
}

func DefaultManifest(name string, target string, source string) EnclaverManifest {
	return EnclaverManifest{
		Version: "v1",
		Name:    name,
		Target:  target,
		Sources: &EnclaverManifestSources{
			App: source,
		},
		Defaults: &EnclaverManifestDefaults{
			CpuCount: 2,
			MemoryMb: 4096,
		},
		Egress: &EnclaverManifestEgress{
			ProxyPort: 10000,
			Allow:     []string{"0.0.0.0/0", "**", "::/0"},
		},
	}
}

type EnclaveMeasurements struct {
	PCR0 string `json:"PCR0"`
	PCR1 string `json:"PCR1"`
	PCR2 string `json:"PCR2"`
}

type EnclaverBuildOutput struct {
	Measurements EnclaveMeasurements `json:"Measurements"`
}

type EnclaverCli struct{}

// BuildEnclave builds an enclaver EIF image using the provided manifest. If build is successful,
// it returns the image's Measurements.
func (*EnclaverCli) BuildEnclave(ctx context.Context, manifest EnclaverManifest) (*EnclaveMeasurements, error) {
	tempfile, err := os.CreateTemp("", "enclaver-manifest")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempfile.Name())

	if err := yaml.NewEncoder(tempfile).Encode(manifest); err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	cmd := exec.CommandContext(
		ctx,
		"enclaver",
		"build",
		"--file",
		tempfile.Name(),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	// Find measurements in the output
	re := regexp.MustCompile(`\{[\s\S]*"Measurements"[\s\S]*\}`)
	jsonMatch := re.Find(stdout.Bytes())
	if jsonMatch == nil {
		return nil, fmt.Errorf("could not find measurements JSON in output")
	}

	var output EnclaverBuildOutput
	if err := json.Unmarshal(jsonMatch, &output); err != nil {
		return nil, fmt.Errorf("failed to parse measurements JSON: %w", err)
	}

	return &output.Measurements, nil
}

// RunEnclave runs an enclaver EIF image `name`. Stdout and stderr are redirected to the parent process.
func (*EnclaverCli) RunEnclave(ctx context.Context, name string) {
	// We'll append this to container name to avoid conflicts
	nameSuffix := uuid.New().String()[:8]

	// We don't use 'enclaver run' here, because it doesn't
	// support --net=host, which is required for Odyn to
	// correctly resolve 'host' to parent machine's localhost
	cmd := exec.CommandContext(
		ctx,
		"docker",
		"run",
		"--rm",
		"--privileged",
		"--net=host",
		fmt.Sprintf("--name=batcher-enclaver-%s", nameSuffix),
		"--device=/dev/nitro_enclaves",
		name,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	go func() {
		err := cmd.Run()
		if err != nil {
			panic(fmt.Errorf("enclave exited with an error: %w", err))
		}
	}()
}
