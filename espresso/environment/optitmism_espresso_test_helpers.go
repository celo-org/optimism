package environment

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/faultproofs"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type EspressoAllocAccount struct {
	State types.Account `json:"state"`
	Name  string        `json:"name"`
}

//go:embed allocs.json
var ESPRESSO_ALLOCS_RAW string
var ESPRESSO_ALLOCS map[common.Address]EspressoAllocAccount

func init() {
	// Unmarshal allocs to set up the dockerConfig environment variables
	ESPRESSO_ALLOCS = make(map[common.Address]EspressoAllocAccount)

	if err := json.Unmarshal([]byte(ESPRESSO_ALLOCS_RAW), &ESPRESSO_ALLOCS); err != nil {
		panic(fmt.Sprintf("failed to unmarshal ESPRESSO_ALLOCS: %v", err))
	}
}

// mockDummyLightClientAddr is a non-zero placeholder light-client address used when
// running against the in-memory mock Espresso client, which does not model a light
// client. No contract is deployed there; the streamer tolerates the resulting error.
var mockDummyLightClientAddr = common.HexToAddress("0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0")

// mockLightClientAddr returns the configured light-client address if
// ESPRESSO_SEQUENCER_LIGHT_CLIENT_PROXY_ADDRESS is set, otherwise a fixed dummy.
func mockLightClientAddr() common.Address {
	if v, ok := os.LookupEnv("ESPRESSO_SEQUENCER_LIGHT_CLIENT_PROXY_ADDRESS"); ok && common.IsHexAddress(v) {
		return common.HexToAddress(v)
	}
	return mockDummyLightClientAddr
}

// This is the mnemonic that we use to create the private key for deploying
// contacts on the L1
const ESPRESSO_MNEMONIC = "giant issue aisle success illegal bike spike question tent bar rely arctic volcano long crawl hungry vocal artwork sniff fantasy very lucky have athlete"

// This is the Mnemonic Index that we use to create the private key for deploying
// contracts on the L1
const ESPRESSO_MNEMONIC_INDEX = "0"

const ESPRESSO_TESTING_BATCHER_KEY = "0xfad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19"

// This is address that corresponds to the menmonic we pass to the espresso-dev-node
var ESPRESSO_CONTRACT_ACCOUNT = common.HexToAddress("0x8943545177806ed17b9f23f0a21ee5948ecaa776")

// EigenDA consstants
const (
	EIGENDA_DOCKER_PORT  = "3100"
	EIGENDA_DOCKER_IMAGE = "ghcr.io/layr-labs/eigenda-proxy:2.2.1"
)

// ErrEspressoBlockHeightDidNotIncrease is a sentinel error that occurs when
// the Espresso Block Height does not increase within the alloted context
// allowance.
var ErrEspressoBlockHeightDidNotIncrease = errors.New("espresso block height did not increase")

// ErrFailedToParseNumber is a sentinel error that occurs when we are unable
// to parse a number from a string
var ErrFailedToParseNumber = errors.New("failed to parse number from string")

// WaitForEspressoBlockHeightToBePositive waits for the Espresso Block Height to
// increase beyond 0.
func WaitForEspressoBlockHeightToBePositive(ctx context.Context, url string) error {
	for {
		select {
		case <-ctx.Done():
			// We've timed out
			return ErrEspressoBlockHeightDidNotIncrease
		default:
		}

		time.Sleep(time.Millisecond * 10)

		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			// Service may not yet be available?
			continue
		}

		if response.StatusCode != http.StatusOK {
			// Service may not yet be available?
			continue
		}

		// Alright, presumably, we have a block height

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, response.Body); err != nil {
			return err
		}
		if err := response.Body.Close(); err != nil {
			return err
		}

		blockHeight, ok := new(big.Int).SetString(buf.String(), 10)
		if !ok {
			return ErrFailedToParseNumber
		}

		if blockHeight.Cmp(big.NewInt(0)) > 0 {
			// We have a positive block height! That means we're
			// committing blocks, and we're progressing.  We
			// **SHOULD** be good to continue"
			return nil
		}
	}
}

// EspressoDevNodeLauncherDocker is an implementation of EspressoDevNodeLauncher
// that uses Docker to launch the Espresso Dev Node
type EspressoDevNodeLauncherDocker struct{}

var _ EspressoE2eDevnetLauncher = (*EspressoDevNodeLauncherDocker)(nil)

// FailedToDetermineL1RPCURL represents a class of errors that occur when we
// are unable to correctly form our L1 RPC URL
type FailedToDetermineL1RPCURL struct {
	Cause error
}

// Error implements error
func (f FailedToDetermineL1RPCURL) Error() string {
	return fmt.Sprintf("failed to determine the L1 RPC URL: %v", f.Cause)
}

// FailedToLoadEspressoAccount represents a class of errors that occur when we
// are unable to load the espresso account
type FailedToLoadEspressoAccount struct {
	Cause error
}

// Error implements error
func (f FailedToLoadEspressoAccount) Error() string {
	return fmt.Sprintf("failed to load the espresso account: %v", f.Cause)
}

// FailedToLaunchDockerContainer represents a class of errors that occur when
// we are unable to launch a docker container
type FailedToLaunchDockerContainer struct {
	Cause error
}

// Error implements error
func (f FailedToLaunchDockerContainer) Error() string {
	return fmt.Sprintf("failed to launch docker container: %v", f.Cause)
}

// EspressoNodeFailedToBecomeReady represents a class of errors that indicate
// that the espresso-dev-node failed to become ready.
type EspressoNodeFailedToBecomeReady struct {
	Cause error
}

// Error implements error
func (e EspressoNodeFailedToBecomeReady) Error() string {
	return fmt.Sprintf("espresso node failed to become ready: %v", e.Cause)
}

// ErrUnableToDetermineEspressoDevNodeSequencerHost is a sentinel error that
// indicates that we were unable to determine what the Sequencer API host
// is meant to be.
var ErrUnableToDetermineEspressoDevNodeSequencerHost = errors.New("unable to determine the host for the espresso-dev-node sequencer api")

// defaultSystemConfigBuilder is the default SystemConfigBuilder utilized by
// the GetE2eDevnetSysConfig method.
func defaultSystemConfigBuilder(t *testing.T, options ...e2esys.SystemConfigOpt) e2esys.SystemConfig {
	return e2esys.DefaultSystemConfig(t, options...)
}

// GetE2eDevnetSysConfig returns a configuration for an E2E devnet.
func (l *EspressoDevNodeLauncherDocker) GetE2eDevnetSysConfig(ctx context.Context, t *testing.T, options ...E2eSystemOption) e2esys.SystemConfig {
	systemConfigsOpts := []e2esys.SystemConfigOpt{
		e2esys.WithAllocType(config.AllocTypeEspressoWithoutEnclave),
	}

	sysConfigBuilder := defaultSystemConfigBuilder
	for _, opt := range options {
		if sysConfigOption := opt.SystemConfigOpt; sysConfigOption != nil {
			systemConfigsOpts = append(systemConfigsOpts, sysConfigOption)
		}

		if builder := opt.SysConfigBuilder; builder != nil {
			sysConfigBuilder = builder
		}
	}

	sysConfig := sysConfigBuilder(t, systemConfigsOpts...)

	// Set a short L1 block time and finalized distance to make tests faster and reach finality sooner
	sysConfig.DeployConfig.L1BlockTime = 2

	// Activate the Espresso hardfork at genesis.
	espressoOffset := hexutil.Uint64(0)
	sysConfig.DeployConfig.L2GenesisEspressoTimeOffset = &espressoOffset

	// Ensure that we fund the dev accounts
	sysConfig.DeployConfig.FundDevAccounts = true

	millionEthers := new(big.Int).Mul(new(big.Int).SetUint64(1_000_000), new(big.Int).SetUint64(params.Ether))

	sysConfig.L1Allocs[ESPRESSO_CONTRACT_ACCOUNT] = types.Account{
		Nonce:   100000,        // Set the nonce to avoid collisions with predeployed contracts
		Balance: millionEthers, // Pre-fund Espresso deployer acount with 1M Ether
	}

	// Set up the L1Allocs in the system config
	for address, account := range ESPRESSO_ALLOCS {
		sysConfig.L1Allocs[address] = account.State
	}

	for _, opt := range options {
		if sysConfigOption := opt.SystemConfigOption; sysConfigOption != nil {
			sysConfigOption(&sysConfig)
		}
	}

	return sysConfig
}

// faultDisputeSystemConfigBuilder id a SystemConfigBuilder that configures
// the system for use with the Fault Dispute System.
func faultDisputeSystemConfigBuilder(t *testing.T, options ...e2esys.SystemConfigOpt) e2esys.SystemConfig {
	return faultproofs.GetFaultDisputeSystemConfigForEspresso(t, options)
}

// WithFaultDisputeSystem will modify the default SysConfigBuilder utilized
// to be one that configures the FaultDisputeSsytem for Espresso.
func WithFaultDisputeSystem() E2eDevnetLauncherOption {
	return func(launcherCtx *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			SysConfigBuilder: faultDisputeSystemConfigBuilder,
		}
	}
}

// WithAltDa is an E2eDevnetLauncherOption that adjusts the SystemConfig
// to be configured for use as a Alt Da.
func WithAltDa() E2eDevnetLauncherOption {
	return func(_ *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			SystemConfigOption: func(sysConfig *e2esys.SystemConfig) {
				sysConfig.DeployConfig.UseAltDA = true
				sysConfig.DeployConfig.DACommitmentType = "KeccakCommitment"
				sysConfig.DeployConfig.DAChallengeWindow = 16
				sysConfig.DeployConfig.DAResolveWindow = 16
				sysConfig.DeployConfig.DABondSize = 1000000
				sysConfig.DeployConfig.DAResolverRefundPercentage = 0
				sysConfig.BatcherMaxPendingTransactions = 0
				sysConfig.BatcherBatchType = 0
				sysConfig.DataAvailabilityType = flags.CalldataType
			},
		}
	}
}

// GetE2eDevnetStartOptions returns the start options for the E2E devnet.
func (l *EspressoDevNodeLauncherDocker) GetE2eDevnetStartOptions(originalCtx context.Context, t *testing.T, launchContext *E2eDevnetLauncherContext, options ...E2eDevnetLauncherOption) []e2esys.StartOption {
	initialOptions := []E2eDevnetLauncherOption{
		launchEspressoDevNode(),
	}

	allOptions := append(initialOptions, options...)

	startOptions := []e2esys.StartOption{}

	for _, opt := range allOptions {
		options := opt(launchContext)

		if gethOption := options.GethOptions; gethOption != nil {
			for k, v := range gethOption {
				launchContext.SystemCfg.GethOptions[k] = append(launchContext.SystemCfg.GethOptions[k], v...)
			}
		}

		if startOption := options.StartOptions; startOption != nil {
			startOptions = append(startOptions, startOption...)
		}
	}

	return startOptions
}

func expandLauncherOptionsToSystemOptions(launchContext *E2eDevnetLauncherContext, options []E2eDevnetLauncherOption) []E2eSystemOption {
	e2eSystemOption := make([]E2eSystemOption, 0, len(options))
	for _, opt := range options {
		e2eSystemOption = append(e2eSystemOption, opt(launchContext))
	}

	return e2eSystemOption
}

func (l *EspressoDevNodeLauncherDocker) StartE2eDevnet(ctx context.Context, t *testing.T, options ...E2eDevnetLauncherOption) (*e2esys.System, EspressoDevNode, error) {
	launchContext := E2eDevnetLauncherContext{
		Ctx:       ctx,
		T:         t,
		SystemCfg: nil,
	}

	e2eSystemOption := expandLauncherOptionsToSystemOptions(&launchContext, options)

	sysConfig := l.GetE2eDevnetSysConfig(ctx, t, e2eSystemOption...)
	originalCtx := ctx
	launchContext.SystemCfg = &sysConfig

	startOptions := l.GetE2eDevnetStartOptions(originalCtx, t, &launchContext, options...)

	// We want to run the espresso-dev-node.  But we need it to be able to
	// access the L1 node.

	system, err := sysConfig.Start(
		t,

		startOptions...,
	)
	if err != nil {
		if system != nil {
			// We don't want the system running in a partial / incomplete
			// state. So we'll tell it to stop here, just in case.
			system.Close()
		}

		return system, nil, err
	}

	// Auto System Cleanup tied to the passed in context.
	{
		// We want to ensure that the lifecycle of the system node is tied to
		// the context we were given.  So if the context is canceled, or
		// otherwise closed, it will automatically clean up the system.
		go (func(ctx context.Context) {
			<-ctx.Done()

			// The system is guaranteed to not be null here.
			system.Close()
		})(originalCtx)
	}

	// Back the EspressoDevNode with the in-memory mock client owned by the system.
	if system.EspressoClient != nil {
		launchContext.EspressoDevNode = &mockEspressoDevNode{client: system.EspressoClient}
	}

	return system, launchContext.EspressoDevNode, launchContext.Error
}

// This code is adapted from a gist file:
// https://gist.github.com/sevkin/96bdae9274465b2d09191384f86ef39d
func determineFreePort() (port int, err error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer func() {
		err = listener.Close()
	}()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

func SetBatcherKey(privateKey ecdsa.PrivateKey) E2eDevnetLauncherOption {
	return func(ct *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				{
					Role: "set-batcher-key",
					BatcherMod: func(c *batcher.CLIConfig, sys *e2esys.System) {
						c.Espresso.TestingBatcherPrivateKey = &privateKey
					},
				},
			},
		}
	}
}

// *c will be set to batcher config. Any devnet launcher options that modify the batcher config
// should be called before this one.
func GetBatcherConfig(c *batcher.CLIConfig) E2eDevnetLauncherOption {
	return func(ct *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				{
					Role: "get-batcher-config",
					BatcherMod: func(cfg *batcher.CLIConfig, sys *e2esys.System) {
						cfg.TargetNumFrames = 10
						cfg.MaxL1TxSize = 250
						cfg.MaxChannelDuration = 1000
						*c = *cfg
					},
				},
			},
		}
	}
}

// SetEspressoUrls allows to set the list of urls for the Espresso client in such a way that N of them are "good" and M of them are "bad".
// Good urls are the urls defined by this test framework repeated M times. The bad url is provided to the function
// This function is introduced for testing purposes. It allows to check the enforcement of the majority rule (Test 12)
func SetEspressoUrls(numGood int, numBad int, badServerUrl string) E2eDevnetLauncherOption {
	return func(ct *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				{
					BatcherMod: func(c *batcher.CLIConfig, sys *e2esys.System) {
						goodUrl := c.Espresso.QueryServiceURLs[0]
						var urls []string

						for i := 0; i < numGood; i++ {
							urls = append(urls, goodUrl)
						}

						for i := 0; i < numBad; i++ {
							urls = append(urls, badServerUrl)
						}
						c.Espresso.QueryServiceURLs = urls
					},
				},
			},
		}
	}
}

// SystemConfigOptionDisableBatcher is a SystemConfigOption that disables
// the Batcher.
//
// | NOTE: This doesn't actually stop the Batcher from being created entirely.
//
//	Instead, it prevents the Batcher from "Starting".  The Batcher still
//	exists in the local context, it just won't be running initially. But
//	it can still be started programatically via its API.  This is most
//	easily done by calling `StartBatchSubmitting` on the `TestDriver` of
//	the system.
func SystemConfigOptionDisableBatcher(cfg *e2esys.SystemConfig) {
	cfg.DisableBatcher = true
}

// Config is a convenience function that allows for the initial modification
// of the SystemConfig only.
func Config(fn func(*e2esys.SystemConfig)) E2eDevnetLauncherOption {
	return func(ct *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			SystemConfigOption: fn,
		}
	}
}

// WithBatcherStoppedInitially is an E2eDevNetLauncherOption that ensures that
// the locally created Batcher is not running initially.
//
// The Batcher can still be started locally with a call to the TestDriver's
// method: `StartBatchSubmitting`.
func WithBatcherStoppedInitially() E2eDevnetLauncherOption {
	return Config(SystemConfigOptionDisableBatcher)
}

// getContainerRemappedHostPort is a helper function that takes the
// containerListeningHostPort and returns the remapped host port
// that the container is listening on.
//
// By default the mapped hosts and ports are in the form of
// - 0.0.0.0:<port> for IPv4
// - [::]:<port> for IPv6
//
// So this function will replace the host with "localhost" to allow
// for communication with the host system.
func getContainerRemappedHostPort(containerListeningHostPort string) (string, error) {
	_, port, err := net.SplitHostPort(containerListeningHostPort)
	if err != nil {
		return "", ErrUnableToDetermineEspressoDevNodeSequencerHost
	}

	hostPort := net.JoinHostPort("localhost", port)

	return hostPort, nil
}

// determineDockerNetworkMode is a helper function that determines the
// docker network mode to use for the container.
//
// We launch in network mode host on linux, otherwise the container is not able
// to communicate with the host system. We use host.docker.internal to do this
// on platforms that are not running natively on linux, as this special address
// achieves the same result.  But on linux, this does not work, and we need to
// run on the host instead.
func determineDockerNetworkMode() string {
	if isRunningOnLinux {
		return "host"
	}

	return ""
}

// launchEspressoDevNodeStartOption is E2eDevnetLauncherOption that launches the
// Espresso Dev Node within a Docker container.  It also ensures that the
// Espresso Dev Node is actively producing blocks before returning.
func launchEspressoDevNodeStartOption(ct *E2eDevnetLauncherContext) e2esys.StartOption {
	return e2esys.StartOption{
		Role: "launch-espresso-dev-node",
		BatcherMod: func(c *batcher.CLIConfig, sys *e2esys.System) {
			// Fail early if there was a prior setup failure.
			if ct.Error != nil {
				ct.T.Fatalf("devnet setup failed before espresso dev node could start: %v", ct.Error)
				return
			}

			// The in-memory mock Espresso client is created and injected into the
			// batchers by e2esys.SystemConfig.Start (sys.EspressoClient). Here we only
			// finish configuring the Espresso CLI config; the QueryServiceURLs placeholder
			// is set by e2esys and is never dialed because the mock client overrides it.
			c.Espresso.Enabled = true
			c.LogConfig.Level = slog.LevelDebug
			// The light client is not modeled by the in-memory mock. The batcher still
			// requires a non-zero address to construct its LightclientCaller, but the
			// streamer tolerates the resulting "no contract" error and continues without
			// pinning the HotShot height. Use the configured address if present, else a
			// fixed dummy so tests don't require ESPRESSO_SEQUENCER_LIGHT_CLIENT_PROXY_ADDRESS.
			c.Espresso.LightClientAddr = mockLightClientAddr()
			c.Espresso.AllowEmptyAttestationService()
		},
	}
}

// launchEspressoDevNode is an E2eDevnetLauncherOption that configures the system to
// use the in-memory mock Espresso client in place of a dockerized espresso-dev-node.
func launchEspressoDevNode() E2eDevnetLauncherOption {
	return func(ct *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				launchEspressoDevNodeStartOption(ct),
			},
		}
	}
}

// StopConfig represents the configuration options for the Stop function.
// The configuration options help to define how the Stop function should
// to failure types.
type StopConfig struct {
	IgnoreErrors bool
	Ctx          context.Context
}

// StopOption is a functional option that allows for the modification of the
// Stop Config
type StopOption func(*StopConfig)

// IgnoreStopErrors is a functional option that ignores errors encountered
// by the stop function, so that they do not cause test failure
func IgnoreStopErrors(c *StopConfig) {
	c.IgnoreErrors = true
}

// Stop is a convenience method to handle the graceful shutdown, and the errors
// thereof of any node that should be stopped on test exit.
// There are different type signatures for the shutdown methods, and this
// aims to handle each of them as gracefully as possible while still ensuring
// that any returned errors are handled accordingly.
func Stop(t *testing.T, toStop any, options ...StopOption) {
	config := StopConfig{
		Ctx: context.Background(),
	}

	for _, opt := range options {
		opt(&config)
	}

	ctx := config.Ctx
	if cast, castOk := toStop.(interface{ Stop() error }); castOk {
		if have, want := cast.Stop(), error(nil); have != want && !config.IgnoreErrors {
			t.Fatalf("failed to stop node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}

		return

	}

	if cast, castOk := toStop.(interface{ Stop(context.Context) error }); castOk {
		if have, want := cast.Stop(ctx), error(nil); have != want && !config.IgnoreErrors {
			t.Fatalf("failed to stop node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}

		return
	}

	if cast, castOk := toStop.(interface{ Close() }); castOk {
		cast.Close()
		return
	}

	if cast, castOk := toStop.(interface{ Close(context.Context) }); castOk {
		cast.Close(ctx)
		return
	}

	if cast, castOk := toStop.(interface{ Close(context.Context) error }); castOk {
		if have, want := cast.Close(ctx), error(nil); have != want && !config.IgnoreErrors {
			t.Fatalf("failed to stop node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}

		return
	}

	t.Fatalf("unable to determine how to stop the given node")
}

// Waits for an Espresso transaction to be confirmed using its hash.
func WaitForEspressoTx(ctx context.Context, txHash *espressoCommon.TaggedBase64, espressoClient espressoClient.EspressoClient) error {
	const transactionFetchTimeout = 4 * time.Second
	const transactionFetchInterval = 100 * time.Millisecond

	timer := time.NewTimer(transactionFetchTimeout)
	defer timer.Stop()

	ticker := time.NewTicker(transactionFetchInterval)
	defer ticker.Stop()

	var err error
	for {
		select {
		case <-ticker.C:
			_, err := espressoClient.FetchTransactionByHash(ctx, txHash)
			if err == nil {
				return nil
			}
		case <-timer.C:
			return fmt.Errorf("failed to fetch transaction by hash: %w", err)
		case <-ctx.Done():
			return nil
		}
	}
}

// --- EigenDA test helpers ---

// StartEigenDA launches a temporary EigenDA proxy in Docker for use in tests.
// It blocks until the proxy port is reachable or the context times out.
func StartEigenDA(ctx context.Context) (*DockerContainerInfo, error) {
	cli := new(DockerCli)

	cfg := DockerContainerConfig{
		Image:   EIGENDA_DOCKER_IMAGE,
		Network: determineDockerNetworkMode(),
		Environment: map[string]string{
			"EIGENDA_PROXY_MEMSTORE_ENABLED": "true",
			"PORT":                           EIGENDA_DOCKER_PORT,
		},
		Ports: []string{EIGENDA_DOCKER_PORT},
	}

	container, err := cli.LaunchContainer(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Wait for port to be reachable
	timeout, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	for {
		select {
		case <-timeout.Done():
			return nil, fmt.Errorf("EigenDA proxy did not become ready")
		default:
			conn, err := net.DialTimeout("tcp", "localhost:"+EIGENDA_DOCKER_PORT, time.Second)
			if err == nil {
				conn.Close()
				return &container, nil
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// StopDockerContainer stops a Docker container by ID.
// Errors are ignored as this is best-effort test cleanup.
func StopDockerContainer(id string) {
	_ = new(DockerCli).StopContainer(context.Background(), id)
}
