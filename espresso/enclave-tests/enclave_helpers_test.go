package enclave_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoTypes "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/ethereum-optimism/optimism/espresso"
	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-service/endpoint"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	espressoDevNodeImage = "ghcr.io/espressosystems/espresso-sequencer/espresso-dev-node:release-20251120-lip2p-tcp-3855"
	espressoAPIPort      = 24000

	// Must match NC_PORT and READY_PORT in enclave-entrypoint.bash.
	enclaveArgsPort  = 8337
	enclaveReadyPort = 8338
)

// enclaveBatcherArgs starts the services the enclave batcher depends on and returns its flags.
func enclaveBatcherArgs(t *testing.T, sys *e2esys.System) []string {
	l1 := sys.NodeEndpoint(e2esys.RoleL1).(endpoint.HttpRPC).HttpRPC()
	espressoURL, lightClient, originHeight := startEspressoDevNode(t, l1)

	// The e2e allocs use mock TEE contracts, which accept any journal of at least 20 bytes.
	attestation := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, `{"raw_proof":{"journal":"%s"},"onchain_proof":"00"}`, strings.Repeat("00", 20))
	}))
	t.Cleanup(attestation.Close)

	return []string{
		"--l1-eth-rpc=" + l1,
		"--l2-eth-rpc=" + sys.NodeEndpoint(e2esys.RoleSeq).(endpoint.HttpRPC).HttpRPC(),
		"--rollup-rpc=" + sys.RollupEndpoint(e2esys.RoleSeq).(endpoint.HttpRPC).HttpRPC(),
		"--private-key=" + hexutil.Encode(crypto.FromECDSA(sys.Cfg.Secrets.AccountAtIdx(6))),
		"--num-confirmations=1",
		"--sub-safety-margin=4",
		"--max-channel-duration=1",
		"--poll-interval=1s",
		"--data-availability-type=calldata",
		"--" + espresso.EnabledFlagName + "=true",
		"--" + espresso.QueryServiceUrlsFlagName + "=" + espressoURL,
		"--" + espresso.QueryServiceUrlsFlagName + "=" + espressoURL,
		"--" + espresso.L1UrlFlagName + "=" + l1,
		"--" + espresso.CaffeinationHeightEspresso + "=" + strconv.FormatUint(originHeight, 10),
		"--" + espresso.LightClientAddrFlagName + "=" + lightClient,
		"--" + espresso.AttestationServiceFlagName + "=" + attestation.URL,
	}
}

// startEspressoDevNode runs an espresso-dev-node against l1, using the Espresso contracts
// that the e2e system predeploys from espresso/environment/allocs.json. It returns the
// query service URL, light client address and a usable Espresso starting height.
func startEspressoDevNode(t *testing.T, l1 string) (string, string, uint64) {
	envVars := map[string]string{
		"ESPRESSO_DEV_NODE_L1_DEPLOYMENT": "skip",
		"ESPRESSO_SEQUENCER_L1_PROVIDER":  l1,
		"ESPRESSO_SEQUENCER_ETH_MNEMONIC": env.ESPRESSO_MNEMONIC,
		"ESPRESSO_DEPLOYER_ACCOUNT_INDEX": env.ESPRESSO_MNEMONIC_INDEX,
		"ESPRESSO_SEQUENCER_STORAGE_PATH": "/data/espresso",
		"ESPRESSO_SEQUENCER_API_PORT":     fmt.Sprint(espressoAPIPort),
		"ESPRESSO_DEV_NODE_VERSION":       "0.4",
		// Even a local 0.4 node has epochs (300 blocks by default), which require
		// finalized L1 staking events. We skip deployment and reuse genesis allocs
		// rather than setting up those events. Keep epoch transitions outside this
		// batcher-upgrade test; they are not the behavior under test here.
		"ESPRESSO_DEV_NODE_EPOCH_HEIGHT":         "18446744073709551615", // u64::MAX
		"ESPRESSO_SEQUENCER_L1_POLLING_INTERVAL": "30ms",
	}
	for address, account := range env.ESPRESSO_ALLOCS {
		if account.Name != "" {
			envVars[account.Name] = hexutil.Encode(address[:])
		}
	}
	args := []string{"run", "-d", "--rm", "--network=host"}
	for k, v := range envVars {
		args = append(args, "--env", k+"="+v)
	}
	out, err := exec.Command("docker", append(args, espressoDevNodeImage)...).Output()
	require.NoError(t, err)
	container := strings.TrimSpace(string(out))
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", container).Run() })

	url := fmt.Sprintf("http://127.0.0.1:%d", espressoAPIPort)
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	// No batcher has started yet, so the earlier blocks contain no batches to replay.
	// Start from a header with an L1 finality reference: the streamer rejects the
	// dev node's genesis header and cannot advance past it on retries.
	originHeight, err := waitForEspressoStartHeight(ctx, espressoClient.NewClient(url))
	require.NoError(t, err)
	t.Logf("Espresso dev node ready at starting height %d", originHeight)
	return url, envVars["ESPRESSO_SEQUENCER_LIGHT_CLIENT_PROXY_ADDRESS"], originHeight
}

type espressoStartHeightSource interface {
	FetchLatestBlockHeight(context.Context) (uint64, error)
	FetchHeaderByHeight(context.Context, uint64) (espressoTypes.HeaderImpl, error)
}

// waitForEspressoStartHeight waits for the local dev node's latest header to have L1 finality.
func waitForEspressoStartHeight(ctx context.Context, client espressoStartHeightSource) (uint64, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		height, err := client.FetchLatestBlockHeight(ctx)
		if err == nil && height > 0 {
			// The status endpoint returns a block count, so the latest header is one below it.
			origin := height - 1
			header, err := client.FetchHeaderByHeight(ctx, origin)
			if err == nil {
				if header.Header == nil || header.Header.GetBlockHeight() != origin {
					return 0, fmt.Errorf("unexpected Espresso header at height %d", origin)
				}
				if header.Header.GetL1Finalized() != nil {
					return origin, nil
				}
			}
		}
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("waiting for an Espresso header with L1 finality: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

type espressoStartHeightStub struct {
	requested []uint64
}

func (s *espressoStartHeightStub) FetchLatestBlockHeight(context.Context) (uint64, error) {
	if len(s.requested) == 0 {
		return 1, nil // Only the genesis header is available initially.
	}
	return 8, nil
}

func (s *espressoStartHeightStub) FetchHeaderByHeight(_ context.Context, height uint64) (espressoTypes.HeaderImpl, error) {
	s.requested = append(s.requested, height)
	header := &espressoTypes.Header0_3{Height: height}
	if height > 0 {
		header.L1Finalized = &espressoTypes.L1BlockInfo{}
	}
	return espressoTypes.HeaderImpl{Header: header}, nil
}

func TestEspressoStartHeightWaitsForUsableLatestHeader(t *testing.T) {
	source := new(espressoStartHeightStub)
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	height, err := waitForEspressoStartHeight(ctx, source)
	require.NoError(t, err)
	require.Equal(t, uint64(7), height)
	require.Equal(t, []uint64{0, 7}, source.requested)
}

// runEnclaveBatcher wraps appImage into an EIF, boots it in a Nitro enclave, sends it the
// batcher args and returns the enclave's container name.
func runEnclaveBatcher(t *testing.T, appImage string, args []string) string {
	ctx := t.Context()
	eif := "op-batcher-eif:" + uuid.New().String()[:8]
	buildEif(t, appImage, eif)

	name := "batcher-enclaver-" + uuid.New().String()[:8]
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "-d", "--privileged", "--net=host",
		"--name="+name, "--device=/dev/nitro_enclaves", eif)
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
	t.Cleanup(func() { _ = stopEnclaveBatcher(name) })
	if testing.Verbose() {
		logs := exec.Command("docker", "logs", "-f", name)
		logs.Stdout, logs.Stderr = os.Stdout, os.Stderr
		require.NoError(t, logs.Start())
	}

	require.NoError(t, waitForEnclaveReady(ctx))
	var payload bytes.Buffer
	for _, arg := range args {
		payload.WriteString(arg)
		payload.WriteByte(0)
	}
	payload.WriteByte(0)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", enclaveArgsPort), 5*time.Second)
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.Write(payload.Bytes())
	require.NoError(t, err)
	return name
}

// stopEnclaveBatcher stops gracefully so enclaver can shut the enclave down; the allocator
// only has room for one enclave at a time.
func stopEnclaveBatcher(name string) error {
	err := exec.Command("docker", "stop", name).Run()
	_ = exec.Command("nitro-cli", "terminate-enclave", "--all").Run()
	return err
}

func buildEif(t *testing.T, appImage, eif string) {
	memoryMiB, err := strconv.Atoi(os.Getenv("TEE_ENCLAVE_MEMORY_MIB"))
	require.NoError(t, err, "TEE_ENCLAVE_MEMORY_MIB must be a positive integer")
	require.Positive(t, memoryMiB, "TEE_ENCLAVE_MEMORY_MIB must be a positive integer")
	cpuCount, err := strconv.Atoi(os.Getenv("TEE_ENCLAVE_CPU_COUNT"))
	require.NoError(t, err, "TEE_ENCLAVE_CPU_COUNT must be a positive integer")
	require.Positive(t, cpuCount, "TEE_ENCLAVE_CPU_COUNT must be a positive integer")

	// enclaver reads YAML, which accepts JSON.
	manifest, err := json.Marshal(map[string]any{
		"version":  "v1",
		"name":     "op-batcher",
		"target":   eif,
		"sources":  map[string]any{"app": appImage},
		"defaults": map[string]any{"cpu_count": cpuCount, "memory_mb": memoryMiB},
		"egress":   map[string]any{"proxy_port": 10000, "allow": []string{"0.0.0.0/0", "**", "::/0"}},
		"ingress":  []map[string]any{{"listen_port": enclaveArgsPort}, {"listen_port": enclaveReadyPort}},
	})
	require.NoError(t, err)
	manifestFile := t.TempDir() + "/enclaver.yaml"
	require.NoError(t, os.WriteFile(manifestFile, manifest, 0o644))

	var stdout bytes.Buffer
	cmd := exec.CommandContext(t.Context(), "enclaver", "build", "--file", manifestFile)
	cmd.Stdout, cmd.Stderr = &stdout, os.Stderr
	require.NoError(t, cmd.Run())

	var output struct{ Measurements struct{ PCR0 string } }
	measurements := regexp.MustCompile(`\{[\s\S]*"Measurements"[\s\S]*\}`).Find(stdout.Bytes())
	require.NoError(t, json.Unmarshal(measurements, &output))
	t.Logf("built %s from %s with PCR0 %s", eif, appImage, output.Measurements.PCR0)
}

// waitForEnclaveReady polls until the enclave entrypoint sends "READY". Enclaver accepts
// connections before the vsock bridge is up, so an accepted connection alone is not enough.
func waitForEnclaveReady(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	for {
		if conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", enclaveReadyPort), 5*time.Second); err == nil {
			buf := make([]byte, 16)
			n, _ := conn.Read(buf)
			conn.Close()
			if bytes.Contains(buf[:n], []byte("READY")) {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("enclave did not become ready: %w", ctx.Err())
		case <-time.After(2 * time.Second):
		}
	}
}
