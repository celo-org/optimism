package batcher

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/bindings/batchauthenticator"
	"github.com/ethereum-optimism/optimism/op-service/bindings/systemconfig"
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

	// code is returned by CodeAt. Empty means "not deployed yet".
	code []byte
	// codeErr, when set, fails CodeAt instead of returning code.
	codeErr error

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
	if len(call.Data) < 4 {
		return nil, errors.New("short calldata")
	}
	if call.To == nil {
		return nil, errors.New("call without a recipient")
	}
	selector := call.Data[:4]
	switch *call.To {
	case testAuthAddr:
		switch {
		case bytes.Equal(selector, m.authABI.Methods["activeIsEspresso"].ID):
			return m.authABI.Methods["activeIsEspresso"].Outputs.Pack(m.activeIsEspresso)
		case bytes.Equal(selector, m.authABI.Methods["espressoBatcher"].ID):
			return m.authABI.Methods["espressoBatcher"].Outputs.Pack(m.espressoBatcher)
		case bytes.Equal(selector, m.authABI.Methods["espressoTEEVerifier"].ID):
			return m.authABI.Methods["espressoTEEVerifier"].Outputs.Pack(m.teeVerifier)
		case bytes.Equal(selector, m.authABI.Methods["systemConfig"].ID):
			return m.authABI.Methods["systemConfig"].Outputs.Pack(m.systemConfigAddr)
		}
	case m.systemConfigAddr:
		if bytes.Equal(selector, m.systemConfigABI.Methods["batcherHash"].ID) {
			var batcherHash [32]byte
			copy(batcherHash[12:], m.fallbackBatcher.Bytes())
			return m.systemConfigABI.Methods["batcherHash"].Outputs.Pack(batcherHash)
		}
	}
	return nil, errors.New("unexpected method call")
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

// TestBatchAuthenticatorReader_ProbeLatches is the core claim of the shared
// reader: the deployment probe is paid once, not once per publish tick. It also
// checks that the reader applies the network timeout itself, so no call site
// can forget it.
func TestBatchAuthenticatorReader_ProbeLatches(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.activeIsEspresso = true
	r := newTestReader(t, backend)

	for i := 0; i < 5; i++ {
		active, err := r.ActiveIsEspresso(context.Background())
		require.NoError(t, err)
		require.True(t, active)
	}

	require.Equal(t, 1, backend.codeAtCalls, "deployment probe should be paid once, not per read")
	require.Equal(t, 5, backend.callCalls, "each read is still one eth_call")
	require.True(t, backend.lastCallHadDeadline, "reader must bound reads by NetworkTimeout")
}

// TestBatchAuthenticatorReader_ProbeFailureIsNotLatched covers the reason the
// probe stays lazy: a batcher may legitimately start before the
// BatchAuthenticator is deployed. Such a batcher must keep retrying and recover
// once the contract appears, rather than caching the failure for the life of
// the process.
func TestBatchAuthenticatorReader_ProbeFailureIsNotLatched(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.code = nil // not deployed yet
	r := newTestReader(t, backend)

	_, err := r.ActiveIsEspresso(context.Background())
	require.ErrorContains(t, err, "no contract code at BatchAuthenticator address")
	require.Zero(t, backend.callCalls, "should not read from an undeployed address")

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

	require.Equal(t, 3, backend.codeAtCalls)
	require.Equal(t, 1, backend.callCalls)
}

// TestBatchAuthenticatorReader_EspressoTEEVerifier covers what the reader adds
// over the generated binding: it refuses to read from an undeployed address,
// and it bounds the call with NetworkTimeout. The resolved address feeds the
// EIP-712 VerifyingContract domain field, so a silently zeroed result would
// sign against the wrong verifier.
func TestBatchAuthenticatorReader_EspressoTEEVerifier(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.teeVerifier = common.HexToAddress("0x00000000000000000000000000000000000000bb")
	backend.code = nil // not deployed yet
	r := newTestReader(t, backend)

	_, err := r.EspressoTEEVerifier(context.Background())
	require.ErrorContains(t, err, "no contract code at BatchAuthenticator address")
	require.Zero(t, backend.callCalls, "should not read from an undeployed address")

	backend.code = []byte{0x60, 0x00}
	got, err := r.EspressoTEEVerifier(context.Background())
	require.NoError(t, err)
	require.Equal(t, backend.teeVerifier, got)
	require.True(t, backend.lastCallHadDeadline, "reader must bound reads by NetworkTimeout")
}

// TestBatchAuthenticatorReader_ZeroTEEVerifierIsRejected covers a
// BatchAuthenticator proxy holding code but no state. Its espressoTEEVerifier
// reads as zero, and accepting it would have the batcher sign every
// authentication against the zero address as EIP-712 verifying contract, which
// the real verifier rejects. Startup must fail instead.
func TestBatchAuthenticatorReader_ZeroTEEVerifierIsRejected(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.teeVerifier = common.Address{} // not initialized yet
	r := newTestReader(t, backend)

	_, err := r.EspressoTEEVerifier(context.Background())
	require.ErrorContains(t, err, "zero espressoTEEVerifier address")
}

// TestBatchAuthenticatorReader_FallbackBatcher locks in that the fallback
// batcher address is read through the SystemConfig the BatchAuthenticator
// itself names, which is the address the contract checks msg.sender against.
// Resolving it costs one extra read the first time and nothing afterwards.
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

// TestBatchAuthenticatorReader_ZeroSystemConfigIsNotLatched covers a
// BatchAuthenticator whose proxy holds code but has not been initialized: it
// reports a zero SystemConfig address. Binding that address would pin the
// reader to a contract that never answers, so the reader must reject it and
// resolve again once the contract is initialized.
func TestBatchAuthenticatorReader_ZeroSystemConfigIsNotLatched(t *testing.T) {
	backend := newMockAuthBackend(t)
	backend.fallbackBatcher = common.HexToAddress("0x00000000000000000000000000000000000000dd")
	backend.systemConfigAddr = common.Address{} // not initialized yet
	r := newTestReader(t, backend)

	_, err := r.FallbackBatcher(context.Background())
	require.ErrorContains(t, err, "zero systemConfig address")

	backend.systemConfigAddr = testSystemConfigAddr
	got, err := r.FallbackBatcher(context.Background())
	require.NoError(t, err)
	require.Equal(t, backend.fallbackBatcher, got)
}

// TestIsBatcherActive covers the publish gate's two checks - mode, then
// identity - and pins their cost. Both modes settle at two eth_calls; the
// fallback mode pays one more on its first evaluation to learn which
// SystemConfig the BatchAuthenticator resolves the fallback batcher through.
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
	}{
		{"espresso batcher, espresso active, authorized", true, true, espressoAddr, true, 2},
		{"fallback batcher, fallback active, authorized", false, false, fallbackAddr, true, 3},
		{"espresso batcher, espresso active, wrong key", true, true, otherAddr, false, 2},
		{"fallback batcher, fallback active, wrong key", false, false, otherAddr, false, 3},
		// Mode mismatch short-circuits before the identity read.
		{"espresso batcher while fallback active", false, true, espressoAddr, false, 1},
		{"fallback batcher while espresso active", true, false, fallbackAddr, false, 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backend := newMockAuthBackend(t)
			backend.activeIsEspresso = test.activeIsEspresso
			backend.espressoBatcher = espressoAddr
			backend.fallbackBatcher = fallbackAddr

			l := &BatchSubmitter{}
			l.Log = testlog.Logger(t, log.LevelDebug)
			l.Txmgr = &testutils.FakeTxMgr{FromAddr: test.from}
			l.Config.Espresso.Enabled = test.espressoEnabled
			l.batchAuth = newTestReader(t, backend)

			got, err := l.isBatcherActive(context.Background())
			require.NoError(t, err)
			require.Equal(t, test.want, got)
			require.Equal(t, test.wantCalls, backend.callCalls)
		})
	}
}

// TestIsBatcherActive_NoAuthenticator guards the nil reader case: without a
// configured BatchAuthenticator the gate must report an error rather than
// silently treating this batcher as active.
func TestIsBatcherActive_NoAuthenticator(t *testing.T) {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)

	_, err := l.isBatcherActive(context.Background())
	require.ErrorContains(t, err, "no BatchAuthenticator configured")
}
