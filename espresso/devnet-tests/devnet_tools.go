package devnet_tests

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	hdwallet "github.com/ethereum-optimism/go-ethereum-hdwallet"
	"github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum-optimism/optimism/op-e2e/config/secrets"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	opclient "github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/sources"
)

// opSuccinctGameType is the dispute game type of OPSuccinctFaultDisputeGame.
const opSuccinctGameType = 42

// Devnet drives the compose devnet in espresso/docker-compose.yml.
type Devnet struct {
	ctx         context.Context
	secrets     secrets.Secrets
	operator    *ecdsa.PrivateKey
	composeFile string
	profiles    []string
	outageTime  time.Duration
	successTime time.Duration

	// Teardown runs at most once per instance: Down is called by the test, by
	// Up on failure and when ctx ends. A repeated `docker compose down` could
	// otherwise remove the containers of a devnet started after this one.
	downOnce sync.Once
	downErr  error

	L1            *ethclient.Client
	L2Seq         *ethclient.Client
	L2SeqRollup   *sources.RollupClient
	L2Verif       *ethclient.Client
	L2VerifRollup *sources.RollupClient
}

// DevnetOption configures a Devnet.
type DevnetOption func(*Devnet)

// WithProofs also starts the op-succinct fault-proof services (compose profile
// "proofs"): the OPSuccinctFaultDisputeGame deployer, proposer and challenger.
func WithProofs() DevnetOption {
	return func(d *Devnet) { d.profiles = append(d.profiles, "proofs") }
}

// NewDevnet returns a handle on the compose devnet. It does not start it; see Up.
//
// The devnet needs the images built by `docker compose build` and the
// deployment written by espresso/devnet/prepare.sh. The test is skipped in
// short mode.
func NewDevnet(ctx context.Context, t *testing.T, opts ...DevnetOption) *Devnet {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping devnet test in short mode")
	}

	espressoDir, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("failed to resolve the espresso directory: %v", err)
	}
	deployment := filepath.Join(espressoDir, "deployment", "l1-config", "state.json")
	if _, err := os.Stat(deployment); err != nil {
		t.Fatalf("devnet deployment not found (%v); run espresso/devnet/prepare.sh first", err)
	}

	s, err := secrets.DefaultMnemonicConfig.Secrets()
	if err != nil {
		t.Fatalf("failed to create default secrets: %v", err)
	}
	operator, err := operatorKey()
	if err != nil {
		t.Fatalf("failed to derive the operator key: %v", err)
	}

	d := &Devnet{
		ctx:         ctx,
		secrets:     *s,
		operator:    operator,
		composeFile: filepath.Join(espressoDir, "docker-compose.yml"),
		outageTime:  durationFromEnv(t, "ESPRESSO_DEVNET_TESTS_OUTAGE_PERIOD", 10*time.Second),
		successTime: durationFromEnv(t, "ESPRESSO_DEVNET_TESTS_LIVENESS_PERIOD", 10*time.Second),
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func durationFromEnv(t *testing.T, name string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(name)
	if !ok {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		t.Fatalf("invalid value for %s: %v", name, err)
	}
	return d
}

// Service identifies a Docker Compose service of the devnet.
type Service string

const (
	OpBatcher         Service = "op-batcher"
	OpBatcherFallback Service = "op-batcher-fallback"
	OpNodeSequencer   Service = "op-node-sequencer"
	OpGethSequencer   Service = "op-geth-sequencer"
	OpNodeVerifier    Service = "op-node-verifier"
	OpGethVerifier    Service = "op-geth-verifier"
	L1Anvil           Service = "l1-anvil"
)

// composeCommand returns a `docker compose` invocation on the devnet's compose
// file with the devnet's profiles enabled.
func (d *Devnet) composeCommand(args ...string) *exec.Cmd {
	cmd := exec.CommandContext(d.ctx, "docker", append([]string{"compose", "-f", d.composeFile}, args...)...)
	cmd.Env = append(os.Environ(), "COMPOSE_PROFILES="+strings.Join(d.profiles, ","))
	return cmd
}

func (d *Devnet) isRunning() bool {
	cmd := d.composeCommand("ps", "-q")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		log.Error("failed to check if devnet is running", "error", err)
		return false
	}
	return strings.TrimSpace(buf.String()) != ""
}

// Up starts the devnet and opens RPC clients on its nodes. The devnet is shut
// down when ctx ends, or immediately if Up fails.
func (d *Devnet) Up() (err error) {
	if d.isRunning() {
		// A leftover devnet, not this instance's: tear it down without
		// consuming this instance's Down.
		if err := d.down(); err != nil {
			return err
		}
		return errors.New("devnet was already running; it has been shut down, start from a clean state")
	}

	cmd := d.composeCommand("up", "-d", "--wait")
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	cmd.Stderr = buf
	upErr := cmd.Run()

	// From here on the caller only owns a running devnet if Up succeeds.
	defer func() {
		if err != nil {
			if testing.Verbose() {
				d.dumpLogs()
			}
			if downErr := d.Down(); downErr != nil {
				log.Error("error shutting down devnet after encountering another error", "error", downErr)
			}
		}
	}()
	if upErr != nil {
		return fmt.Errorf("failed to start docker compose (%w): %s", upErr, buf.String())
	}

	go func() {
		<-d.ctx.Done()
		if err := d.Down(); err != nil {
			log.Error("error shutting down devnet asynchronously", "error", err)
		}
	}()

	if testing.Verbose() {
		// Stream logs while the test runs; the command ends with ctx.
		go func() {
			cmd := d.composeCommand("logs", "-f")
			cmd.Stdout = os.Stdout
			_ = cmd.Run()
		}()
	}

	if d.L2Seq, err = d.serviceClient(OpGethSequencer, 8546); err != nil {
		return err
	}
	if d.L2SeqRollup, err = d.rollupClient(OpNodeSequencer, 9545); err != nil {
		return err
	}
	if d.L2Verif, err = d.serviceClient(OpGethVerifier, 8546); err != nil {
		return err
	}
	if d.L2VerifRollup, err = d.rollupClient(OpNodeVerifier, 9546); err != nil {
		return err
	}
	if d.L1, err = d.serviceClient(L1Anvil, 8545); err != nil {
		return err
	}
	return nil
}

func (d *Devnet) dumpLogs() {
	cmd := d.composeCommand("logs", "--no-color")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

// Down stops the devnet and deletes its volumes. Only the first call does the
// work; later calls return its result.
func (d *Devnet) Down() error {
	d.downOnce.Do(func() { d.downErr = d.down() })
	return d.downErr
}

func (d *Devnet) down() error {
	for _, c := range []*ethclient.Client{d.L1, d.L2Seq, d.L2Verif} {
		if c != nil {
			c.Close()
		}
	}
	for _, c := range []*sources.RollupClient{d.L2SeqRollup, d.L2VerifRollup} {
		if c != nil {
			c.Close()
		}
	}
	// The devnet is throwaway, so give containers one second to stop. Use a
	// fresh context: Down also runs once d.ctx has ended.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	args := []string{"compose", "-f", d.composeFile, "down", "-v", "--remove-orphans", "--timeout", "1"}
	cmd := exec.CommandContext(ctx, "docker", args...)
	// Every profile, so that services of any profile are removed.
	cmd.Env = append(os.Environ(), "COMPOSE_PROFILES=*")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker compose down (%w): %s", err, out)
	}
	return nil
}

// WaitForBatcher waits until the verifier's safe head advances, i.e. until a
// batch has been posted to L1 and derived.
func (d *Devnet) WaitForBatcher(ctx context.Context) error {
	safe := big.NewInt(rpc.SafeBlockNumber.Int64())

	current := new(big.Int)
	if head, err := d.L2Verif.HeaderByNumber(ctx, safe); err == nil {
		current = head.Number
	} else if !isNotFound(err) {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	for {
		head, err := d.L2Verif.HeaderByNumber(ctx, safe)
		switch {
		case err == nil && head.Number.Cmp(current) > 0:
			return nil
		case err != nil && !isNotFound(err):
			return err
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("verifier safe head did not advance past %d: %w", current, ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}

// ServiceUp starts a service of the devnet.
func (d *Devnet) ServiceUp(service Service) error {
	log.Info("bringing up service", "service", service)
	return d.composeCommand("up", "-d", string(service)).Run()
}

// ServiceDown stops and removes a service of the devnet.
func (d *Devnet) ServiceDown(service Service) error {
	log.Info("shutting down service", "service", service)
	return d.composeCommand("down", string(service)).Run()
}

// callBatcherRPC calls an admin RPC method of a running batcher service.
func (d *Devnet) callBatcherRPC(service Service, method string) error {
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","params":[],"id":1}`, method)
	cmd := d.composeCommand(
		"exec", "-T", string(service),
		"wget", "-q", "-O-", "--header=Content-Type: application/json", "--post-data="+body, "http://localhost:8545",
	)
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to call %s (%w): %s%s", method, err, stdout.String(), stderr.String())
	}
	// A JSON-RPC error comes with HTTP 200, so wget succeeds: check the body.
	var resp struct {
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return fmt.Errorf("failed to decode the %s response (%w): %s", method, err, stdout.String())
	}
	if resp.Error != nil {
		return fmt.Errorf("%s failed: %s (code %d)", method, resp.Error.Message, resp.Error.Code)
	}
	log.Info("RPC call successful", "service", service, "method", method, "response", stdout.String())
	return nil
}

// StartBatcherSubmitting starts batch submission on a running batcher service.
func (d *Devnet) StartBatcherSubmitting(service Service) error {
	return d.callBatcherRPC(service, "admin_startBatcher")
}

// StopBatcherSubmitting stops batch submission on a running batcher service.
func (d *Devnet) StopBatcherSubmitting(service Service) error {
	return d.callBatcherRPC(service, "admin_stopBatcher")
}

func (d *Devnet) RollupConfig(ctx context.Context) (*rollup.Config, error) {
	return d.L2SeqRollup.RollupConfig(ctx)
}

// SystemConfig returns the L1 SystemConfig contract and a transactor for its owner.
func (d *Devnet) SystemConfig(ctx context.Context) (*bindings.SystemConfig, *bind.TransactOpts, error) {
	config, err := d.RollupConfig(ctx)
	if err != nil {
		return nil, nil, err
	}
	contract, err := bindings.NewSystemConfig(config.L1SystemConfigAddress, d.L1)
	if err != nil {
		return nil, nil, err
	}
	owner, err := bind.NewKeyedTransactorWithChainID(d.operator, config.L1ChainID)
	if err != nil {
		return nil, nil, err
	}
	return contract, owner, nil
}

// operatorKey derives OPERATOR_PRIVATE_KEY from espresso/.env: HD index 0 of
// the default test mnemonic, which espresso/devnet/prepare.sh makes the owner
// of SystemConfig and the other chain roles. secrets.Secrets has no field for
// that index.
func operatorKey() (*ecdsa.PrivateKey, error) {
	wallet, err := hdwallet.NewFromMnemonic(secrets.DefaultMnemonicConfig.Mnemonic)
	if err != nil {
		return nil, err
	}
	return wallet.PrivateKey(accounts.Account{URL: accounts.URL{Path: "m/44'/60'/0'/0/0"}})
}

// SubmitL2Tx sends a transaction from Alice to the sequencer and waits until
// the sequencer (but not necessarily the verifier) includes it.
func (d *Devnet) SubmitL2Tx(applyTxOpts helpers.TxOptsFn) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(d.ctx, 3*time.Minute)
	defer cancel()

	chainID, err := d.L2Seq.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	privKey := d.secrets.Alice
	address := crypto.PubkeyToAddress(privKey.PublicKey)
	balance, err := d.L2Seq.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("getting initial sender balance: %w", err)
	}
	if balance.Sign() <= 0 {
		return nil, errors.New("sender account empty")
	}
	nonce, err := d.L2Seq.NonceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting nonce: %w", err)
	}

	opts := &helpers.TxOpts{
		Nonce:          nonce,
		Value:          common.Big0,
		GasTipCap:      big.NewInt(10),
		GasFeeCap:      big.NewInt(1_000_000_000),
		Gas:            21_000,
		ExpectedStatus: types.ReceiptStatusSuccessful,
	}
	applyTxOpts(opts)

	tx := types.MustSignNewTx(privKey, types.LatestSignerForChainID(chainID), &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     opts.Nonce,
		To:        opts.ToAddr,
		Value:     opts.Value,
		GasTipCap: opts.GasTipCap,
		GasFeeCap: opts.GasFeeCap,
		Gas:       opts.Gas,
		Data:      opts.Data,
	})
	log.Info("send transaction", "from", address, "hash", tx.Hash())
	if err := d.L2Seq.SendTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("sending L2 tx: %w", err)
	}

	receipt, err := wait.ForReceiptOK(ctx, d.L2Seq, tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("waiting for L2 tx: %w", err)
	}
	if opts.ExpectedStatus != receipt.Status {
		return nil, fmt.Errorf("wrong status: have %d, want %d", receipt.Status, opts.ExpectedStatus)
	}
	return receipt, nil
}

// VerifyL2Tx waits until the verifier derives a transaction already included
// by the sequencer, and checks both agree on its receipt.
func (d *Devnet) VerifyL2Tx(receipt *types.Receipt) error {
	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Minute)
	defer cancel()

	verified, err := wait.ForReceiptOK(ctx, d.L2Verif, receipt.TxHash)
	if err != nil {
		return fmt.Errorf("waiting for L2 tx on verification client: %w", err)
	}
	if !reflect.DeepEqual(receipt, verified) {
		return fmt.Errorf("verification client returned incorrect receipt\nSeq:  %v\nVerif: %v", receipt, verified)
	}
	return nil
}

// RunL2Tx submits a transaction and waits for the verifier to derive it.
func (d *Devnet) RunL2Tx(applyTxOpts helpers.TxOptsFn) error {
	receipt, err := d.SubmitL2Tx(applyTxOpts)
	if err != nil {
		return err
	}
	return d.VerifyL2Tx(receipt)
}

type BurnReceipt struct {
	InitialBurnBalance *big.Int
	BurnAmount         *big.Int
	BurnAddress        common.Address
	Receipt            *types.Receipt
}

// SubmitSimpleL2Burn sends 1 wei to a burn address and waits until the
// sequencer (but not necessarily the verifier) includes it.
func (d *Devnet) SubmitSimpleL2Burn() (*BurnReceipt, error) {
	receipt := &BurnReceipt{
		BurnAddress: common.Address{0xff, 0xff},
		BurnAmount:  big.NewInt(1),
	}

	var err error
	receipt.InitialBurnBalance, err = d.L2Verif.BalanceAt(d.ctx, receipt.BurnAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("getting initial burn address balance: %w", err)
	}

	tx := environment.L2TxWithOptions(
		environment.L2TxWithAmount(receipt.BurnAmount),
		environment.L2TxWithToAddress(&receipt.BurnAddress),
		environment.L2TxWithVerifyOnClients(d.L2Verif),
	)
	if receipt.Receipt, err = d.SubmitL2Tx(tx); err != nil {
		return nil, err
	}
	return receipt, nil
}

// VerifySimpleL2Burn waits for the verifier to derive a burn submitted with
// SubmitSimpleL2Burn and checks the burned amount.
func (d *Devnet) VerifySimpleL2Burn(receipt *BurnReceipt) error {
	ctx, cancel := context.WithTimeout(d.ctx, 2*time.Minute)
	defer cancel()

	if err := d.VerifyL2Tx(receipt.Receipt); err != nil {
		return err
	}

	final, err := wait.ForBalanceChange(ctx, d.L2Verif, receipt.BurnAddress, receipt.InitialBurnBalance)
	if err != nil {
		return fmt.Errorf("waiting for balance change for burn address %s: %w", receipt.BurnAddress, err)
	}
	burned := new(big.Int).Sub(final, receipt.InitialBurnBalance)
	if burned.Cmp(receipt.BurnAmount) != 0 {
		return fmt.Errorf("incorrect amount burned (have %s, want %s)", burned, receipt.BurnAmount)
	}
	return nil
}

// RunSimpleL2Burn runs a simple L2 burn and checks the verifier derives it.
func (d *Devnet) RunSimpleL2Burn() error {
	receipt, err := d.SubmitSimpleL2Burn()
	if err != nil {
		return err
	}
	return d.VerifySimpleL2Burn(receipt)
}

// SleepOutageDuration waits for ESPRESSO_DEVNET_TESTS_OUTAGE_PERIOD while
// simulating an outage.
func (d *Devnet) SleepOutageDuration() {
	log.Info("sleeping during simulated outage", "duration", d.outageTime)
	time.Sleep(d.outageTime)
}

// SleepRecoveryDuration waits for ESPRESSO_DEVNET_TESTS_LIVENESS_PERIOD before
// a run is considered a success.
func (d *Devnet) SleepRecoveryDuration() {
	log.Info("sleeping to check that things stay working", "duration", d.successTime)
	time.Sleep(d.successTime)
}

type ChallengeGame struct {
	Index      uint64
	Address    common.Address
	OutputRoot []byte
	Claims     uint64
}

// ListChallengeGames lists the OPSuccinctFaultDisputeGame games of the
// DisputeGameFactory. Needs WithProofs.
func (d *Devnet) ListChallengeGames() ([]ChallengeGame, error) {
	systemConfig, _, err := d.SystemConfig(d.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system config: %w", err)
	}
	factoryAddr, err := systemConfig.DisputeGameFactory(&bind.CallOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to get dispute game factory address: %w", err)
	}
	factory, err := bindings.NewDisputeGameFactory(factoryAddr, d.L1)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to dispute game factory: %w", err)
	}
	gameCount, err := factory.GameCount(&bind.CallOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to get game count: %w", err)
	}

	var games []ChallengeGame
	for i := range bigs.Uint64Strict(gameCount) {
		gameInfo, err := factory.GameAtIndex(&bind.CallOpts{}, new(big.Int).SetUint64(i))
		if err != nil {
			return nil, fmt.Errorf("failed to get game at index %d: %w", i, err)
		}
		if gameInfo.GameType != opSuccinctGameType {
			continue
		}
		// OPSuccinctFaultDisputeGame only has a root claim, no claim tree.
		game, err := bindings.NewFaultDisputeGame(gameInfo.Proxy, d.L1)
		if err != nil {
			return nil, fmt.Errorf("failed to bind to dispute game %s: %w", gameInfo.Proxy, err)
		}
		rootClaim, err := game.RootClaim(&bind.CallOpts{})
		if err != nil {
			return nil, fmt.Errorf("failed to get root claim of %s: %w", gameInfo.Proxy, err)
		}
		games = append(games, ChallengeGame{
			Index:      i,
			Address:    gameInfo.Proxy,
			OutputRoot: rootClaim[:],
			Claims:     1,
		})
	}
	return games, nil
}

// hostPort returns the host port that privatePort of a service is published on.
func (d *Devnet) hostPort(service Service, privatePort uint16) (uint16, error) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	cmd := d.composeCommand("port", string(service), strconv.Itoa(int(privatePort)))
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("docker compose port (%w): %s", err, stderr.String())
	}
	out := strings.TrimSpace(stdout.String())
	_, portStr, found := strings.Cut(out, ":")
	if !found {
		return 0, fmt.Errorf("invalid output from docker compose port: %q", out)
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid output from docker compose port: %q (%w)", out, err)
	}
	return uint16(port), nil
}

// serviceClient opens an Ethereum RPC client on a service's RPC port.
func (d *Devnet) serviceClient(service Service, port uint16) (*ethclient.Client, error) {
	hostPort, err := d.hostPort(service, port)
	if err != nil {
		return nil, fmt.Errorf("could not get %s port: %w", service, err)
	}
	client, err := ethclient.DialContext(d.ctx, fmt.Sprintf("http://127.0.0.1:%d", hostPort))
	if err != nil {
		return nil, fmt.Errorf("could not open %s RPC client: %w", service, err)
	}
	return client, nil
}

// rollupClient opens an op-node rollup RPC client on a service's RPC port.
func (d *Devnet) rollupClient(service Service, port uint16) (*sources.RollupClient, error) {
	hostPort, err := d.hostPort(service, port)
	if err != nil {
		return nil, fmt.Errorf("could not get %s port: %w", service, err)
	}
	rpc, err := opclient.NewRPC(d.ctx, log.Root(), fmt.Sprintf("http://127.0.0.1:%d", hostPort), opclient.WithDialAttempts(10))
	if err != nil {
		return nil, fmt.Errorf("could not open %s RPC client: %w", service, err)
	}
	return sources.NewRollupClient(rpc), nil
}
