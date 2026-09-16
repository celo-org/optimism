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

// The two contract ABIs the active-batcher gate reads from, parsed once.
var (
	authABI   = mustGetABI(bindings.BatchAuthenticatorMetaData)
	sysCfgABI = mustGetABI(bindings.SystemConfigMetaData)
)

func mustGetABI(md *bind.MetaData) *abi.ABI {
	parsed, err := md.GetAbi()
	if err != nil {
		panic(err)
	}
	return parsed
}

// gateMethodBySelector resolves the calldata's 4-byte selector to a method of
// one of the two gate ABIs.
func gateMethodBySelector(data []byte) (abi.Method, error) {
	if m, err := authABI.MethodById(data[:4]); err == nil {
		return *m, nil
	}
	if m, err := sysCfgABI.MethodById(data[:4]); err == nil {
		return *m, nil
	}
	return abi.Method{}, fmt.Errorf("unknown method selector %#x", data[:4])
}

// contractCallHandler produces the raw return data (or error) for one
// invocation of the given contract method.
type contractCallHandler func(method abi.Method) ([]byte, error)

// contractHandlers routes contract calls per contract address, then per method
// name.
type contractHandlers map[common.Address]map[string]contractCallHandler

// mockContractL1Client is an L1Client whose CallContract dispatches on the
// target contract address and the method named by the calldata selector.
// Keying on the address matters: the two ABIs share several selectors (owner,
// version, ...), and it makes the tests prove a value was read from the right
// contract, not merely that the method was called somewhere. Calls with no
// registered handler fail loudly: the tests below also use that to prove a
// method was NOT consulted on a given path.
type mockContractL1Client struct {
	bind.ContractBackend
	code     []byte
	codeErr  error
	tipTime  uint64
	tipErr   error
	handlers contractHandlers
}

func (m *mockContractL1Client) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return m.code, m.codeErr
}

func (m *mockContractL1Client) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if call.To == nil {
		return nil, errors.New("unexpected contract-creation call")
	}
	method, err := gateMethodBySelector(call.Data)
	if err != nil {
		return nil, err
	}
	handler, ok := m.handlers[*call.To][method.Name]
	if !ok {
		return nil, fmt.Errorf("unexpected call to method %s on contract %s", method.Name, call.To)
	}
	return handler(method)
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

// returns builds a handler that packs vals as the called method's return
// values.
func returns(vals ...any) contractCallHandler {
	return func(method abi.Method) ([]byte, error) {
		return method.Outputs.Pack(vals...)
	}
}

func fails() contractCallHandler {
	return func(abi.Method) ([]byte, error) { return nil, errContractCallFailed }
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
	tests := []struct {
		name            string
		espressoEnabled bool
		handlers        contractHandlers
		want            bool
		wantErr         bool
	}{
		{
			name:            "espresso batcher active and authorized",
			espressoEnabled: true,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(true),
					"espressoBatcher":  returns(testSenderAddr),
				},
			},
			want: true,
		},
		{
			name:            "espresso batcher active but key not authorized",
			espressoEnabled: true,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(true),
					"espressoBatcher":  returns(testOtherAddr),
				},
			},
			want: false,
		},
		{
			// No espressoBatcher handler: the identity of the inactive mode
			// must not even be consulted once the mode gate fails.
			name:            "espresso batcher while fallback mode is active",
			espressoEnabled: true,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(false),
				},
			},
			want: false,
		},
		{
			// The counterpart quadrant: an inverted mode check would return
			// true here and make the fallback batcher publish during an
			// Espresso handoff.
			name:            "fallback batcher while espresso mode is active",
			espressoEnabled: false,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(true),
				},
			},
			want: false,
		},
		{
			name:            "fallback batcher active and authorized via batcherHash",
			espressoEnabled: false,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(false),
					"systemConfig":     returns(testSysCfgAddr),
				},
				testSysCfgAddr: {
					"batcherHash": returns(eth.AddressAsLeftPaddedHash(testSenderAddr)),
				},
			},
			want: true,
		},
		{
			name:            "fallback batcher active but key not authorized",
			espressoEnabled: false,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(false),
					"systemConfig":     returns(testSysCfgAddr),
				},
				testSysCfgAddr: {
					"batcherHash": returns(eth.AddressAsLeftPaddedHash(testOtherAddr)),
				},
			},
			want: false,
		},
		{
			// batcherHash stores the batcher address in its low 20 bytes; a
			// future versioned batcherHash may set high bytes. The gate must
			// slice the low 20 bytes rather than compare the whole word.
			name:            "batcherHash high bytes do not obscure the low-20-byte address",
			espressoEnabled: false,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(false),
					"systemConfig":     returns(testSysCfgAddr),
				},
				testSysCfgAddr: {
					"batcherHash": func(method abi.Method) ([]byte, error) {
						versioned := eth.AddressAsLeftPaddedHash(testSenderAddr)
						for i := 0; i < 12; i++ {
							versioned[i] = 0xff
						}
						return method.Outputs.Pack(versioned)
					},
				},
			},
			want: true,
		},
		{
			name:            "activeIsEspresso read failure is an error",
			espressoEnabled: true,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": fails(),
				},
			},
			wantErr: true,
		},
		{
			name:            "espressoBatcher read failure is an error",
			espressoEnabled: true,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(true),
					"espressoBatcher":  fails(),
				},
			},
			wantErr: true,
		},
		{
			name:            "systemConfig read failure is an error",
			espressoEnabled: false,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(false),
					"systemConfig":     fails(),
				},
			},
			wantErr: true,
		},
		{
			name:            "batcherHash read failure is an error",
			espressoEnabled: false,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(false),
					"systemConfig":     returns(testSysCfgAddr),
				},
				testSysCfgAddr: {
					"batcherHash": fails(),
				},
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

	tests := []struct {
		name            string
		espressoEnabled bool
		tipTime         uint64
		tipErr          error
		handlers        contractHandlers
		wantSkip        bool
	}{
		{
			name:            "active espresso batcher publishes",
			espressoEnabled: true,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(true),
					"espressoBatcher":  returns(testSenderAddr),
				},
			},
			wantSkip: false,
		},
		{
			name:            "espresso batcher skips while fallback mode is active",
			espressoEnabled: true,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(false),
				},
			},
			wantSkip: true,
		},
		{
			name:            "espresso batcher fails closed on gate error",
			espressoEnabled: true,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": fails(),
				},
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
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(false),
					"systemConfig":     returns(testSysCfgAddr),
				},
				testSysCfgAddr: {
					"batcherHash": returns(eth.AddressAsLeftPaddedHash(testSenderAddr)),
				},
			},
			wantSkip: false,
		},
		{
			name:            "post-fork fallback batcher skips while espresso mode is active",
			espressoEnabled: false,
			tipTime:         espressoTime,
			handlers: contractHandlers{
				testAuthAddr: {
					"activeIsEspresso": returns(true),
				},
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
		// The typed-nil client panics on any RPC use, proving neither gate is
		// evaluated.
		l := newActiveGateSubmitter(t, true, nil)
		l.RollupConfig.BatchAuthenticatorAddress = common.Address{}

		require.False(t, l.shouldSkipPublishForActiveSeq(context.Background()))
	})
}
