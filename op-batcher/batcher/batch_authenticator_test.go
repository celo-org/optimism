package batcher

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/bindings/batchauthenticator"
	"github.com/ethereum-optimism/optimism/op-service/bindings/systemconfig"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

var (
	testAuthAddr         = common.HexToAddress("0x00000000000000000000000000000000000000aa")
	testSystemConfigAddr = common.HexToAddress("0x00000000000000000000000000000000000000cc")
)

// mockAuthBackend is a minimal bind.ContractCaller standing in for the L1
// client, serving both the BatchAuthenticator and the SystemConfig.
type mockAuthBackend struct {
	authABI         *abi.ABI
	systemConfigABI *abi.ABI

	// code is returned by CodeAt. Empty means "not deployed yet": like a real
	// node, CallContract then returns no data.
	code []byte
	// codeErr, when set, fails CodeAt instead of returning code.
	codeErr error
	// callErr, when set, fails CallContract.
	callErr error
	// failGetter, when set, fails only the call to that getter.
	failGetter string

	activeIsEspresso bool
	espressoBatcher  common.Address
	fallbackBatcher  common.Address
	teeVerifier      common.Address
	// systemConfigAddr is the address the BatchAuthenticator reports as its
	// SystemConfig, and the only address serving batcherHash.
	systemConfigAddr common.Address

	codeAtCalls int
	callCalls   int
	// lastCallHadDeadline records whether the context reaching CallContract
	// carried a deadline, i.e. whether the reader applied NetworkTimeout.
	lastCallHadDeadline bool
}

func newMockAuthBackend(t *testing.T) *mockAuthBackend {
	t.Helper()
	authABI, err := batchauthenticator.BatchAuthenticatorMetaData.GetAbi()
	require.NoError(t, err)
	systemConfigABI, err := systemconfig.SystemConfigMetaData.GetAbi()
	require.NoError(t, err)
	return &mockAuthBackend{
		authABI:          authABI,
		systemConfigABI:  systemConfigABI,
		code:             []byte{0x60, 0x00},
		systemConfigAddr: testSystemConfigAddr,
	}
}

func (m *mockAuthBackend) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	m.codeAtCalls++
	if m.codeErr != nil {
		return nil, m.codeErr
	}
	return m.code, nil
}

func (m *mockAuthBackend) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	m.callCalls++
	_, m.lastCallHadDeadline = ctx.Deadline()
	if m.callErr != nil {
		return nil, m.callErr
	}
	if len(m.code) == 0 {
		return nil, nil
	}
	if call.To == nil {
		return nil, errors.New("call without a recipient")
	}
	// Each address serves only its own ABI, so a getter read from the wrong
	// contract fails the lookup.
	contractABI := m.authABI
	if *call.To == m.systemConfigAddr {
		contractABI = m.systemConfigABI
	}
	method, err := contractABI.MethodById(call.Data)
	if err != nil {
		return nil, err
	}
	if method.Name == m.failGetter {
		return nil, fmt.Errorf("rpc boom on %s", method.Name)
	}
	switch method.Name {
	case "activeIsEspresso":
		return method.Outputs.Pack(m.activeIsEspresso)
	case "espressoBatcher":
		return method.Outputs.Pack(m.espressoBatcher)
	case "espressoTEEVerifier":
		return method.Outputs.Pack(m.teeVerifier)
	case "systemConfig":
		return method.Outputs.Pack(m.systemConfigAddr)
	case "batcherHash":
		var batcherHash [32]byte
		copy(batcherHash[12:], m.fallbackBatcher.Bytes())
		return method.Outputs.Pack(batcherHash)
	}
	return nil, fmt.Errorf("unexpected method call %s", method.Name)
}

func newTestReader(t *testing.T, backend *mockAuthBackend) *batchAuthenticatorReader {
	t.Helper()
	r, err := newBatchAuthenticatorReader(testAuthAddr, backend, time.Second)
	require.NoError(t, err)
	return r
}

// TestNewBatchAuthenticatorReader_ZeroAddressIsNilReader pins the invariant
// every call site reads: a chain without a BatchAuthenticator gets no reader,
// never one bound to the zero address that would answer every question with a
// failed call.
func TestNewBatchAuthenticatorReader_ZeroAddressIsNilReader(t *testing.T) {
	r, err := newBatchAuthenticatorReader(common.Address{}, newMockAuthBackend(t), time.Second)
	require.NoError(t, err)
	require.Nil(t, r)
}

// TestBatchAuthenticatorReader_ReadsDoNotProbe pins the steady-state cost of a
// read: one eth_call and no CodeAt, since a deployed contract always returns
// data and the bindings only look for code when it does not. It also checks
// that the reader applies the network timeout itself, so no call site can
// forget it.
func TestBatchAuthenticatorReader_ReadsDoNotProbe(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.activeIsEspresso = true
	r := newTestReader(t, backend)

	for i := 0; i < 5; i++ {
		active, err := r.ActiveIsEspresso(context.Background())
		require.NoError(t, err)
		require.True(t, active)
	}

	require.Zero(t, backend.codeAtCalls, "reads of a deployed contract should not check for code")
	require.Equal(t, 5, backend.callCalls, "each read is still one eth_call")
	require.True(t, backend.lastCallHadDeadline, "reader must bound reads by NetworkTimeout")
}

// TestBatchAuthenticatorReader_NotDeployedIsRetried covers a batcher that
// starts before the BatchAuthenticator is deployed, which is legitimate for the
// fallback batcher. Its reads must fail with bind.ErrNoCode and recover once the
// contract appears, rather than caching the failure for the life of the process.
func TestBatchAuthenticatorReader_NotDeployedIsRetried(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.code = nil // not deployed yet
	r := newTestReader(t, backend)

	_, err := r.ActiveIsEspresso(context.Background())
	require.ErrorIs(t, err, bind.ErrNoCode)

	backend.codeErr = errors.New("rpc boom")
	_, err = r.ActiveIsEspresso(context.Background())
	require.ErrorContains(t, err, "rpc boom")

	// Contract shows up; the reader recovers without needing a restart.
	backend.codeErr = nil
	backend.code = []byte{0x60, 0x00}
	backend.activeIsEspresso = true
	active, err := r.ActiveIsEspresso(context.Background())
	require.NoError(t, err)
	require.True(t, active)

	require.Equal(t, 2, backend.codeAtCalls, "only the reads that got no data checked for code")
	require.Equal(t, 3, backend.callCalls)
}

// TestBatchAuthenticatorReader_EspressoTEEVerifier pins the address the batcher
// signs against: it is whatever the contract reports, unchanged.
func TestBatchAuthenticatorReader_EspressoTEEVerifier(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.teeVerifier = common.HexToAddress("0x00000000000000000000000000000000000000bb")

	got, err := newTestReader(t, backend).EspressoTEEVerifier(context.Background())
	require.NoError(t, err)
	require.Equal(t, backend.teeVerifier, got)
}

// TestBatchAuthenticatorReader_ZeroTEEVerifierIsRejected covers the reader's
// guard against a zero espressoTEEVerifier. No supported deployment yields one,
// but accepting it would have the batcher sign every authentication against the
// zero address as EIP-712 verifying contract, which the real verifier rejects.
// Startup must fail instead.
func TestBatchAuthenticatorReader_ZeroTEEVerifierIsRejected(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.teeVerifier = common.Address{}
	r := newTestReader(t, backend)

	_, err := r.EspressoTEEVerifier(context.Background())
	require.ErrorContains(t, err, "zero espressoTEEVerifier address")
}

// TestBatchAuthenticatorReader_FallbackBatcher locks in that the fallback
// batcher address is read through the SystemConfig the BatchAuthenticator
// itself names, which is the address the contract checks msg.sender against,
// and that the reader learns that address once.
func TestBatchAuthenticatorReader_FallbackBatcher(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.fallbackBatcher = common.HexToAddress("0x00000000000000000000000000000000000000dd")
	r := newTestReader(t, backend)

	got, err := r.FallbackBatcher(context.Background())
	require.NoError(t, err)
	require.Equal(t, backend.fallbackBatcher, got)
	require.Equal(t, 2, backend.callCalls, "first read resolves the SystemConfig address, then reads batcherHash")

	got, err = r.FallbackBatcher(context.Background())
	require.NoError(t, err)
	require.Equal(t, backend.fallbackBatcher, got)
	require.Equal(t, 3, backend.callCalls, "the SystemConfig address is resolved once")
}

// TestBatchAuthenticatorReader_ZeroSystemConfigIsNotLatched covers the reader's
// guard against a zero SystemConfig address. No supported deployment yields
// one, but binding it would pin the reader to a contract that never answers, so
// the reader must reject it and resolve again on the next read.
func TestBatchAuthenticatorReader_ZeroSystemConfigIsNotLatched(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.fallbackBatcher = common.HexToAddress("0x00000000000000000000000000000000000000dd")
	backend.systemConfigAddr = common.Address{}
	r := newTestReader(t, backend)

	_, err := r.FallbackBatcher(context.Background())
	require.ErrorContains(t, err, "zero systemConfig address")

	backend.systemConfigAddr = testSystemConfigAddr
	got, err := r.FallbackBatcher(context.Background())
	require.NoError(t, err)
	require.Equal(t, backend.fallbackBatcher, got)
}

// Fragments of the publish gate's skip warnings, matched by requireWarnedOnce.
const (
	warnModeInactive    = "Batcher is not the active batcher"
	warnKeyUnauthorized = "Configured batcher key is not the authorized batcher"
	warnFallbackGateErr = "Failed to evaluate fallback-auth gate"
	warnActiveCheckErr  = "Failed to check if batcher is active"
)

// requireWarnedOnce asserts the captured logs hold exactly one Warn and that it
// contains want, or no Warn at all when want is empty.
func requireWarnedOnce(t *testing.T, logs *testlog.CapturingHandler, want string) {
	t.Helper()
	warns := logs.FindLogs(testlog.NewLevelFilter(log.LevelWarn))
	if want == "" {
		require.Empty(t, warns)
		return
	}
	require.Len(t, warns, 1)
	require.Contains(t, warns[0].Record.Message, want)
}

// TestIsBatcherActive covers the publish gate's two checks - mode, then
// identity - and pins their cost, since the gate runs before every batch
// transaction. A mode mismatch must not pay for the identity read. Each skip
// warns once and is throttled on the next check.
func TestIsBatcherActive(t *testing.T) {
	espressoAddr := common.HexToAddress("0x00000000000000000000000000000000000000e1")
	fallbackAddr := common.HexToAddress("0x00000000000000000000000000000000000000e2")
	otherAddr := common.HexToAddress("0x00000000000000000000000000000000000000e3")

	tests := []struct {
		name             string
		activeIsEspresso bool
		espressoEnabled  bool
		from             common.Address
		want             bool
		wantCalls        int
		wantWarn         string
	}{
		{"espresso batcher, espresso active, authorized", true, true, espressoAddr, true, 2, ""},
		{"fallback batcher, fallback active, authorized", false, false, fallbackAddr, true, 3, ""},
		{"espresso batcher, espresso active, wrong key", true, true, otherAddr, false, 2, warnKeyUnauthorized},
		{"fallback batcher, fallback active, wrong key", false, false, otherAddr, false, 3, warnKeyUnauthorized},
		// Mode mismatch short-circuits before the identity read.
		{"espresso batcher while fallback active", false, true, espressoAddr, false, 1, warnModeInactive},
		{"fallback batcher while espresso active", true, false, fallbackAddr, false, 1, warnModeInactive},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backend := newMockAuthBackend(t)
			backend.activeIsEspresso = test.activeIsEspresso
			backend.espressoBatcher = espressoAddr
			backend.fallbackBatcher = fallbackAddr

			logger, logs := testlog.CaptureLogger(t, log.LevelDebug)
			l := &BatchSubmitter{}
			l.Log = logger
			l.degradedLog = oplog.NewRepeatStateLogger()
			l.Txmgr = &testutils.FakeTxMgr{FromAddr: test.from}
			l.Config.Espresso.Enabled = test.espressoEnabled
			l.batchAuth = newTestReader(t, backend)

			got, err := l.isBatcherActive(context.Background())
			require.NoError(t, err)
			require.Equal(t, test.want, got)
			require.Equal(t, test.wantCalls, backend.callCalls)
			requireWarnedOnce(t, logs, test.wantWarn)

			// The same skip again is throttled.
			_, err = l.isBatcherActive(context.Background())
			require.NoError(t, err)
			requireWarnedOnce(t, logs, test.wantWarn)
		})
	}
}

// TestIsBatcherActive_SharedKeyWrongMode pins the mode check on its own. With
// one key authorized for both roles the identity check passes in either mode,
// so only the mode check can keep a batcher idle while the other mode is active.
func TestIsBatcherActive_SharedKeyWrongMode(t *testing.T) {
	sharedAddr := common.HexToAddress("0x00000000000000000000000000000000000000e1")

	for _, espressoEnabled := range []bool{true, false} {
		t.Run(fmt.Sprintf("espressoEnabled=%v", espressoEnabled), func(t *testing.T) {
			backend := newMockAuthBackend(t)
			backend.activeIsEspresso = !espressoEnabled
			backend.espressoBatcher = sharedAddr
			backend.fallbackBatcher = sharedAddr

			l := &BatchSubmitter{}
			l.Log = testlog.Logger(t, log.LevelDebug)
			l.degradedLog = oplog.NewRepeatStateLogger()
			l.Txmgr = &testutils.FakeTxMgr{FromAddr: sharedAddr}
			l.Config.Espresso.Enabled = espressoEnabled
			l.batchAuth = newTestReader(t, backend)

			got, err := l.isBatcherActive(context.Background())
			require.NoError(t, err)
			require.False(t, got)
		})
	}
}

// TestIsBatcherActive_KeyWarningResetsOnModeSwitch checks that leaving our mode
// resets the throttled wrong-key warning: when the mode comes back with the key
// still wrong, the operator gets a fresh Warn instead of silence.
func TestIsBatcherActive_KeyWarningResetsOnModeSwitch(t *testing.T) {
	espressoAddr := common.HexToAddress("0x00000000000000000000000000000000000000e1")
	otherAddr := common.HexToAddress("0x00000000000000000000000000000000000000e3")

	backend := newMockAuthBackend(t)
	backend.activeIsEspresso = true
	backend.espressoBatcher = espressoAddr

	logger, logs := testlog.CaptureLogger(t, log.LevelDebug)
	l := &BatchSubmitter{}
	l.Log = logger
	l.degradedLog = oplog.NewRepeatStateLogger()
	l.Txmgr = &testutils.FakeTxMgr{FromAddr: otherAddr}
	l.Config.Espresso.Enabled = true
	l.batchAuth = newTestReader(t, backend)

	check := func() {
		t.Helper()
		active, err := l.isBatcherActive(context.Background())
		require.NoError(t, err)
		require.False(t, active)
	}

	// Wrong key in our mode: warn once, then stay quiet.
	check()
	check()
	logs.RequireMessageContainedOnce(t, warnKeyUnauthorized)

	// The other mode takes over, then ours comes back with the key still wrong.
	backend.activeIsEspresso = false
	check()
	backend.activeIsEspresso = true
	check()
	logs.RequireMessageContainedNTimes(t, warnKeyUnauthorized, 2)
	require.Nil(t, logs.FindLog(testlog.NewMessageContainsFilter("authorized again")),
		"the key was never authorized, so nothing may report it recovered")
}

// TestIsBatcherActive_NoAuthenticator guards the nil reader case: without a
// configured BatchAuthenticator the gate must report an error rather than
// silently treating this batcher as active.
func TestIsBatcherActive_NoAuthenticator(t *testing.T) {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.degradedLog = oplog.NewRepeatStateLogger()

	_, err := l.isBatcherActive(context.Background())
	require.ErrorContains(t, err, "no BatchAuthenticator configured")
}

// TestShouldSkipPublishForActiveSeq covers the publish loop's gate in front of
// isBatcherActive: the Espresso batcher always consults the active flag, the
// fallback batcher only once fallback auth is required, and either gate failing
// to evaluate skips the publish (fails closed). Each skip warns once and is
// throttled on the next check.
func TestShouldSkipPublishForActiveSeq(t *testing.T) {
	const espressoTime = 1000
	espressoAddr := common.HexToAddress("0x00000000000000000000000000000000000000e1")
	fallbackAddr := common.HexToAddress("0x00000000000000000000000000000000000000e2")

	tests := []struct {
		name             string
		espressoEnabled  bool
		tipTime          uint64
		tipErr           error
		callErr          error
		activeIsEspresso bool
		from             common.Address
		wantSkip         bool
		wantWarn         string
	}{
		// Pre-fork the fallback batcher ignores the flag, even when it names the Espresso batcher.
		{name: "fallback batcher before the fork ignores the flag", tipTime: espressoTime - 1, activeIsEspresso: true, from: fallbackAddr, wantSkip: false},
		{name: "fallback batcher, fallback-auth gate fails", tipErr: errors.New("l1 down"), from: fallbackAddr, wantSkip: true, wantWarn: warnFallbackGateErr},
		{name: "fallback batcher after the fork, fallback active", tipTime: espressoTime, from: fallbackAddr, wantSkip: false},
		{name: "fallback batcher after the fork, espresso active", tipTime: espressoTime, activeIsEspresso: true, from: fallbackAddr, wantSkip: true, wantWarn: warnModeInactive},
		// The Espresso batcher consults the flag without reading the L1 tip.
		{name: "espresso batcher, espresso active", espressoEnabled: true, tipErr: errors.New("unused"), activeIsEspresso: true, from: espressoAddr, wantSkip: false},
		{name: "espresso batcher, fallback active", espressoEnabled: true, from: espressoAddr, wantSkip: true, wantWarn: warnModeInactive},
		{name: "espresso batcher, active check fails", espressoEnabled: true, callErr: errors.New("l1 down"), activeIsEspresso: true, from: espressoAddr, wantSkip: true, wantWarn: warnActiveCheckErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backend := newMockAuthBackend(t)
			backend.callErr = test.callErr
			backend.activeIsEspresso = test.activeIsEspresso
			backend.espressoBatcher = espressoAddr
			backend.fallbackBatcher = fallbackAddr

			logger, logs := testlog.CaptureLogger(t, log.LevelDebug)
			l := &BatchSubmitter{}
			l.Log = logger
			l.degradedLog = oplog.NewRepeatStateLogger()
			l.RollupConfig = &rollup.Config{EspressoTime: u64(espressoTime)}
			l.Config.NetworkTimeout = time.Second
			l.Config.Espresso.Enabled = test.espressoEnabled
			l.L1Client = &mockFixedTimeL1Client{time: test.tipTime, err: test.tipErr}
			l.Txmgr = &testutils.FakeTxMgr{FromAddr: test.from}
			l.batchAuth = newTestReader(t, backend)

			require.Equal(t, test.wantSkip, l.shouldSkipPublishForActiveSeq(context.Background()))
			requireWarnedOnce(t, logs, test.wantWarn)

			// The same skip again is throttled.
			require.Equal(t, test.wantSkip, l.shouldSkipPublishForActiveSeq(context.Background()))
			requireWarnedOnce(t, logs, test.wantWarn)
		})
	}
}

// TestShouldSkipPublishForActiveSeq_NoAuthenticator guards the nil reader case:
// a chain without a BatchAuthenticator runs a plain upstream batcher, which
// never skips a publish.
func TestShouldSkipPublishForActiveSeq_NoAuthenticator(t *testing.T) {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.degradedLog = oplog.NewRepeatStateLogger()

	require.False(t, l.shouldSkipPublishForActiveSeq(context.Background()))
}

// TestIsBatcherActive_ReadErrors fails each contract read the gate makes, one
// per row. isBatcherActive must return the error rather than swallow it, and
// the error must name the getter that failed.
func TestIsBatcherActive_ReadErrors(t *testing.T) {
	tests := []struct {
		failGetter      string
		espressoEnabled bool
	}{
		{"activeIsEspresso", true},
		{"espressoBatcher", true},
		{"systemConfig", false},
		{"batcherHash", false},
	}

	for _, test := range tests {
		t.Run(test.failGetter, func(t *testing.T) {
			backend := newMockAuthBackend(t)
			backend.activeIsEspresso = test.espressoEnabled
			backend.failGetter = test.failGetter

			l := &BatchSubmitter{}
			l.Log = testlog.Logger(t, log.LevelDebug)
			l.degradedLog = oplog.NewRepeatStateLogger()
			l.Txmgr = &testutils.FakeTxMgr{}
			l.Config.Espresso.Enabled = test.espressoEnabled
			l.batchAuth = newTestReader(t, backend)

			_, err := l.isBatcherActive(context.Background())
			require.ErrorContains(t, err, "failed to read "+test.failGetter)
		})
	}
}

// TestResolveTEEVerifierAddress covers the startup step that fixes the EIP-712
// verifying contract: any failure must surface and leave the address unset,
// rather than letting the batcher sign against a zero verifier.
func TestResolveTEEVerifierAddress(t *testing.T) {
	teeVerifier := common.HexToAddress("0x00000000000000000000000000000000000000bb")

	tests := []struct {
		name        string
		teeVerifier common.Address
		failGetter  string
		wantErr     string
	}{
		{name: "resolved", teeVerifier: teeVerifier},
		{name: "read fails", teeVerifier: teeVerifier, failGetter: "espressoTEEVerifier", wantErr: "failed to read espressoTEEVerifier"},
		{name: "zero verifier", teeVerifier: common.Address{}, wantErr: "zero espressoTEEVerifier address"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backend := newMockAuthBackend(t)
			backend.teeVerifier = test.teeVerifier
			backend.failGetter = test.failGetter

			l := &BatchSubmitter{}
			l.Log = testlog.Logger(t, log.LevelDebug)
			l.batchAuth = newTestReader(t, backend)

			err := l.resolveTEEVerifierAddress(context.Background())
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				require.Zero(t, l.teeVerifierAddress)
				return
			}
			require.NoError(t, err)
			require.Equal(t, teeVerifier, l.teeVerifierAddress)
		})
	}
}

// TestRegisterBatcher_DeploymentCheck covers registerBatcher's checks on the
// BatchAuthenticator before it asks for a proof and sends the registration: a
// registration sent to an address with no code would succeed as a no-op, so
// each failure must stop it before the attestation service is called.
func TestRegisterBatcher_DeploymentCheck(t *testing.T) {
	tests := []struct {
		name    string
		codeErr error
		wantErr string
	}{
		{name: "code check fails", codeErr: errors.New("rpc boom"), wantErr: "failed to check code at BatchAuthenticator address"},
		{name: "not deployed", wantErr: "no contract code at BatchAuthenticator address"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backend := newMockAuthBackend(t)
			backend.code = nil // not deployed, or the check fails before it can tell
			backend.codeErr = test.codeErr

			l := newRegisteringBatchSubmitter(t)
			l.batchAuth = newTestReader(t, backend)

			require.ErrorContains(t, l.registerBatcher(context.Background()), test.wantErr)
		})
	}
}

// TestBatchAuthenticatorCallers_NoAuthenticator guards the nil reader case for
// the two startup steps: without a configured BatchAuthenticator both must
// fail, and registerBatcher must not ask the attestation service for a proof.
func TestBatchAuthenticatorCallers_NoAuthenticator(t *testing.T) {
	l := newRegisteringBatchSubmitter(t)

	require.ErrorContains(t, l.resolveTEEVerifierAddress(context.Background()), "no BatchAuthenticator configured")
	require.ErrorContains(t, l.registerBatcher(context.Background()), "no BatchAuthenticator configured")
}

// newRegisteringBatchSubmitter returns a BatchSubmitter with no reader that is
// set up to register: it has an attestation and an attestation service, which
// fails the test if called, since each caller must stop before asking for a
// proof.
func newRegisteringBatchSubmitter(t *testing.T) *BatchSubmitter {
	t.Helper()
	attestationService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("attestation service called despite a failed BatchAuthenticator check")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(attestationService.Close)

	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.RollupConfig = &rollup.Config{BatchAuthenticatorAddress: testAuthAddr}
	l.Espresso.Attestation = []byte{0x01}
	l.Config.Espresso.AttestationService = attestationService.URL
	return l
}
