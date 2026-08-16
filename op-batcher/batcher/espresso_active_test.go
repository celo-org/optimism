package batcher

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/espresso/bindings"
	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

var (
	testAuthAddr          = common.HexToAddress("0x00000000000000000000000000000000000000a0")
	testSysCfgAddr        = common.HexToAddress("0x00000000000000000000000000000000000000b0")
	testSenderAddr        = common.HexToAddress("0x00000000000000000000000000000000000000c0")
	testOtherAddr         = common.HexToAddress("0x00000000000000000000000000000000000000d0")
	errContractCallFailed = errors.New("contract call failed")
)

// contractCallHandler produces the raw return data (or error) for one contract
// method invocation.
type contractCallHandler func() ([]byte, error)

// mockContractL1Client is an L1Client whose CallContract dispatches on the
// 4-byte method selector of the incoming call. Selectors are unique across the
// BatchAuthenticator and SystemConfig ABIs, so one dispatch table serves both
// contracts. Calls with no registered handler fail loudly: the tests below also
// use that to prove a method was NOT consulted on a given path.
type mockContractL1Client struct {
	bind.ContractBackend
	code     []byte
	codeErr  error
	tipTime  uint64
	tipErr   error
	handlers map[string]contractCallHandler // hex selector -> handler
}

func (m *mockContractL1Client) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return m.code, m.codeErr
}

func (m *mockContractL1Client) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	sel := hexutil.Encode(call.Data[:4])
	handler, ok := m.handlers[sel]
	if !ok {
		return nil, fmt.Errorf("unexpected contract call with selector %s", sel)
	}
	return handler()
}

func (m *mockContractL1Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if m.tipErr != nil {
		return nil, m.tipErr
	}
	return &types.Header{Number: big.NewInt(0), Time: m.tipTime}, nil
}

func (m *mockContractL1Client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return 0, nil
}

func batchAuthenticatorABI(t *testing.T) *abi.ABI {
	a, err := bindings.BatchAuthenticatorMetaData.GetAbi()
	require.NoError(t, err)
	return a
}

func systemConfigABI(t *testing.T) *abi.ABI {
	a, err := bindings.SystemConfigMetaData.GetAbi()
	require.NoError(t, err)
	return a
}

func methodSelector(t *testing.T, a *abi.ABI, method string) string {
	m, ok := a.Methods[method]
	require.True(t, ok, "ABI has no method %q", method)
	return hexutil.Encode(m.ID)
}

// returns registers a handler that packs vals as the method's return values.
func returns(t *testing.T, a *abi.ABI, method string, vals ...any) contractCallHandler {
	m := a.Methods[method]
	return func() ([]byte, error) {
		out, err := m.Outputs.Pack(vals...)
		require.NoError(t, err)
		return out, nil
	}
}

func fails() contractCallHandler {
	return func() ([]byte, error) { return nil, errContractCallFailed }
}

// addressAsBytes32 left-pads an address into the low 20 bytes of a bytes32,
// mirroring how SystemConfig stores batcherHash.
func addressAsBytes32(addr common.Address) [32]byte {
	var out [32]byte
	copy(out[12:], addr[:])
	return out
}

func newActiveGateSubmitter(t *testing.T, espressoEnabled bool, client *mockContractL1Client) *BatchSubmitter {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.Metr = metrics.NoopMetrics
	l.RollupConfig = &rollup.Config{
		BatchAuthenticatorAddress: testAuthAddr,
	}
	l.Config.NetworkTimeout = time.Second
	l.Config.Espresso.Enabled = espressoEnabled
	l.Txmgr = testutils.NewFakeTxMgr(l.Log, testSenderAddr, eth.ChainIDFromUInt64(1))
	l.L1Client = client
	return l
}

// TestIsBatcherActive locks the two gates of the active-batcher switch: the
// mode gate (the contract's activeIsEspresso flag must match this node's role)
// and the identity gate (the configured sender key must be the authorized
// batcher for that mode). This switch decides whether a batcher publishes at
// all during a handoff, and an inverted mode check compiles and passes CI, so
// every quadrant of (activeIsEspresso x Espresso.Enabled) is pinned explicitly.
func TestIsBatcherActive(t *testing.T) {
	authABI := batchAuthenticatorABI(t)
	sysABI := systemConfigABI(t)

	selActive := methodSelector(t, authABI, "activeIsEspresso")
	selEspressoBatcher := methodSelector(t, authABI, "espressoBatcher")
	selSystemConfig := methodSelector(t, authABI, "systemConfig")
	selBatcherHash := methodSelector(t, sysABI, "batcherHash")

	tests := []struct {
		name            string
		espressoEnabled bool
		handlers        map[string]contractCallHandler
		want            bool
		wantErr         bool
	}{
		{
			name:            "espresso batcher active and authorized",
			espressoEnabled: true,
			handlers: map[string]contractCallHandler{
				selActive:          returns(t, authABI, "activeIsEspresso", true),
				selEspressoBatcher: returns(t, authABI, "espressoBatcher", testSenderAddr),
			},
			want: true,
		},
		{
			name:            "espresso batcher active but key not authorized",
			espressoEnabled: true,
			handlers: map[string]contractCallHandler{
				selActive:          returns(t, authABI, "activeIsEspresso", true),
				selEspressoBatcher: returns(t, authABI, "espressoBatcher", testOtherAddr),
			},
			want: false,
		},
		{
			// No espressoBatcher handler: the identity of the inactive mode
			// must not even be consulted once the mode gate fails.
			name:            "espresso batcher while fallback mode is active",
			espressoEnabled: true,
			handlers: map[string]contractCallHandler{
				selActive: returns(t, authABI, "activeIsEspresso", false),
			},
			want: false,
		},
		{
			// The counterpart quadrant: an inverted mode check would return
			// true here and make the fallback batcher publish during an
			// Espresso handoff.
			name:            "fallback batcher while espresso mode is active",
			espressoEnabled: false,
			handlers: map[string]contractCallHandler{
				selActive: returns(t, authABI, "activeIsEspresso", true),
			},
			want: false,
		},
		{
			name:            "fallback batcher active and authorized via batcherHash",
			espressoEnabled: false,
			handlers: map[string]contractCallHandler{
				selActive:       returns(t, authABI, "activeIsEspresso", false),
				selSystemConfig: returns(t, authABI, "systemConfig", testSysCfgAddr),
				selBatcherHash:  returns(t, sysABI, "batcherHash", addressAsBytes32(testSenderAddr)),
			},
			want: true,
		},
		{
			name:            "fallback batcher active but key not authorized",
			espressoEnabled: false,
			handlers: map[string]contractCallHandler{
				selActive:       returns(t, authABI, "activeIsEspresso", false),
				selSystemConfig: returns(t, authABI, "systemConfig", testSysCfgAddr),
				selBatcherHash:  returns(t, sysABI, "batcherHash", addressAsBytes32(testOtherAddr)),
			},
			want: false,
		},
		{
			// batcherHash stores the batcher address in its low 20 bytes; a
			// future versioned batcherHash may set high bytes. The gate must
			// slice the low 20 bytes rather than compare the whole word.
			name:            "batcherHash high bytes do not obscure the low-20-byte address",
			espressoEnabled: false,
			handlers: map[string]contractCallHandler{
				selActive:       returns(t, authABI, "activeIsEspresso", false),
				selSystemConfig: returns(t, authABI, "systemConfig", testSysCfgAddr),
				selBatcherHash: func() ([]byte, error) {
					versioned := addressAsBytes32(testSenderAddr)
					for i := 0; i < 12; i++ {
						versioned[i] = 0xff
					}
					return sysABI.Methods["batcherHash"].Outputs.Pack(versioned)
				},
			},
			want: true,
		},
		{
			name:            "activeIsEspresso read failure is an error",
			espressoEnabled: true,
			handlers: map[string]contractCallHandler{
				selActive: fails(),
			},
			wantErr: true,
		},
		{
			name:            "espressoBatcher read failure is an error",
			espressoEnabled: true,
			handlers: map[string]contractCallHandler{
				selActive:          returns(t, authABI, "activeIsEspresso", true),
				selEspressoBatcher: fails(),
			},
			wantErr: true,
		},
		{
			name:            "systemConfig read failure is an error",
			espressoEnabled: false,
			handlers: map[string]contractCallHandler{
				selActive:       returns(t, authABI, "activeIsEspresso", false),
				selSystemConfig: fails(),
			},
			wantErr: true,
		},
		{
			name:            "batcherHash read failure is an error",
			espressoEnabled: false,
			handlers: map[string]contractCallHandler{
				selActive:       returns(t, authABI, "activeIsEspresso", false),
				selSystemConfig: returns(t, authABI, "systemConfig", testSysCfgAddr),
				selBatcherHash:  fails(),
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &mockContractL1Client{
				code:     []byte{0x01},
				handlers: test.handlers,
			}
			l := newActiveGateSubmitter(t, test.espressoEnabled, client)

			got, err := l.isBatcherActive(context.Background())
			if test.wantErr {
				require.Error(t, err)
				require.False(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}

	t.Run("no contract code at the authenticator address is an error", func(t *testing.T) {
		client := &mockContractL1Client{code: nil}
		l := newActiveGateSubmitter(t, true, client)
		got, err := l.isBatcherActive(context.Background())
		require.ErrorContains(t, err, "no contract code")
		require.False(t, got)
	})

	t.Run("CodeAt failure is an error", func(t *testing.T) {
		client := &mockContractL1Client{codeErr: errContractCallFailed}
		l := newActiveGateSubmitter(t, true, client)
		got, err := l.isBatcherActive(context.Background())
		require.Error(t, err)
		require.False(t, got)
	})
}

// TestShouldSkipPublishForActiveSeq locks the publish gate wired into
// publishStateToL1: the Espresso batcher always honors the on-chain active
// flag, the fallback batcher honors it only once fallback auth is required
// (post EspressoTime), and every gate-evaluation failure fails closed by
// skipping the tick.
func TestShouldSkipPublishForActiveSeq(t *testing.T) {
	const espressoTime = 1000
	authABI := batchAuthenticatorABI(t)
	sysABI := systemConfigABI(t)

	selActive := methodSelector(t, authABI, "activeIsEspresso")
	selEspressoBatcher := methodSelector(t, authABI, "espressoBatcher")
	selSystemConfig := methodSelector(t, authABI, "systemConfig")
	selBatcherHash := methodSelector(t, sysABI, "batcherHash")

	tests := []struct {
		name            string
		espressoEnabled bool
		tipTime         uint64
		tipErr          error
		handlers        map[string]contractCallHandler
		wantSkip        bool
	}{
		{
			name:            "active espresso batcher publishes",
			espressoEnabled: true,
			handlers: map[string]contractCallHandler{
				selActive:          returns(t, authABI, "activeIsEspresso", true),
				selEspressoBatcher: returns(t, authABI, "espressoBatcher", testSenderAddr),
			},
			wantSkip: false,
		},
		{
			name:            "espresso batcher skips while fallback mode is active",
			espressoEnabled: true,
			handlers: map[string]contractCallHandler{
				selActive: returns(t, authABI, "activeIsEspresso", false),
			},
			wantSkip: true,
		},
		{
			name:            "espresso batcher fails closed on gate error",
			espressoEnabled: true,
			handlers: map[string]contractCallHandler{
				selActive: fails(),
			},
			wantSkip: true,
		},
		{
			// No handlers at all: consulting the contract would fail and flip
			// the result to skip, so wantSkip=false also proves the pre-fork
			// fallback batcher never touches the BatchAuthenticator.
			name:            "pre-fork fallback batcher publishes without consulting the contract",
			espressoEnabled: false,
			tipTime:         espressoTime - 1,
			handlers:        nil,
			wantSkip:        false,
		},
		{
			name:            "post-fork fallback batcher publishes while active and authorized",
			espressoEnabled: false,
			tipTime:         espressoTime,
			handlers: map[string]contractCallHandler{
				selActive:       returns(t, authABI, "activeIsEspresso", false),
				selSystemConfig: returns(t, authABI, "systemConfig", testSysCfgAddr),
				selBatcherHash:  returns(t, sysABI, "batcherHash", addressAsBytes32(testSenderAddr)),
			},
			wantSkip: false,
		},
		{
			name:            "post-fork fallback batcher skips while espresso mode is active",
			espressoEnabled: false,
			tipTime:         espressoTime,
			handlers: map[string]contractCallHandler{
				selActive: returns(t, authABI, "activeIsEspresso", true),
			},
			wantSkip: true,
		},
		{
			name:            "fallback batcher fails closed on fallback-auth gate error",
			espressoEnabled: false,
			tipErr:          errContractCallFailed,
			wantSkip:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &mockContractL1Client{
				code:     []byte{0x01},
				tipTime:  test.tipTime,
				tipErr:   test.tipErr,
				handlers: test.handlers,
			}
			l := newActiveGateSubmitter(t, test.espressoEnabled, client)
			l.RollupConfig.EspressoTime = u64(espressoTime)

			require.Equal(t, test.wantSkip, l.shouldSkipPublishForActiveSeq(context.Background()))
		})
	}

	t.Run("no authenticator configured never skips", func(t *testing.T) {
		// A nil L1 client proves neither gate is evaluated: any RPC use would panic.
		l := newActiveGateSubmitter(t, true, nil)
		l.L1Client = nil
		l.RollupConfig.BatchAuthenticatorAddress = common.Address{}

		require.False(t, l.shouldSkipPublishForActiveSeq(context.Background()))
	})
}
