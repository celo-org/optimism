package batcher

import (
	"bytes"
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
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

// errTipL1Client fails every tip fetch, for the fail-closed path of the
// publish gate.
type errTipL1Client struct {
	mockFixedTimeL1Client
}

func (e *errTipL1Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return nil, errors.New("l1 tip unavailable")
}

var (
	batchAuthTestABI = mustParseABI(bindings.BatchAuthenticatorMetaData)
	sysCfgTestABI    = mustParseABI(bindings.SystemConfigMetaData)
)

func mustParseABI(md *bind.MetaData) *abi.ABI {
	parsed, err := md.GetAbi()
	if err != nil {
		panic(err)
	}
	return parsed
}

// fakeBatchAuthBackend extends mockFixedTimeL1Client with a CodeAt and
// CallContract that serve the read-only BatchAuthenticator/SystemConfig calls
// isBatcherActive makes, dispatching on (contract address, method ID) resolved
// from the same abigen bindings the production code calls through.
type fakeBatchAuthBackend struct {
	mockFixedTimeL1Client
	authAddr         common.Address
	sysCfgAddr       common.Address
	activeIsEspresso bool
	espressoBatcher  common.Address
	fallbackBatcher  common.Address
	callErr          error // when set, every CallContract fails
}

func (f *fakeBatchAuthBackend) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return []byte{0x01}, nil
}

func (f *fakeBatchAuthBackend) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if f.callErr != nil {
		return nil, f.callErr
	}
	responses := []struct {
		to     common.Address
		abi    *abi.ABI
		method string
		value  any
	}{
		{f.authAddr, batchAuthTestABI, "activeIsEspresso", f.activeIsEspresso},
		{f.authAddr, batchAuthTestABI, "espressoBatcher", f.espressoBatcher},
		{f.authAddr, batchAuthTestABI, "systemConfig", f.sysCfgAddr},
		{f.sysCfgAddr, sysCfgTestABI, "batcherHash", [32]byte(common.BytesToHash(f.fallbackBatcher.Bytes()))},
	}
	for _, r := range responses {
		if *call.To == r.to && bytes.Equal(call.Data, r.abi.Methods[r.method].ID) {
			return r.abi.Methods[r.method].Outputs.Pack(r.value)
		}
	}
	return nil, fmt.Errorf("unexpected contract call to %s with data %x", call.To, call.Data)
}

// contractState is the post-enforcement BatchAuthenticator view a test case
// runs against; a nil *contractState means the case must not consult the
// contract at all.
type contractState struct {
	activeIsEspresso bool
	callErr          error
}

// TestShouldSkipPublish_EnforcementBoundary locks batch-submission ownership to
// the verifier's event-auth enforcement boundary for both batcher roles (issue
// #492); see shouldSkipPublishForActiveSeq for the ownership rationale. Every
// case runs once per role, so the pre-enforcement invariant — the fallback
// always publishes, the TEE batcher never does — reads directly off the two
// want columns.
func TestShouldSkipPublish_EnforcementBoundary(t *testing.T) {
	const espressoTime uint64 = 1_000_000
	enforcementTime := espressoTime + derive.BatchAuthEnforcementDelaySecs

	authAddr := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	sysCfgAddr := common.HexToAddress("0x00000000000000000000000000000000000000bb")
	teeKey := common.HexToAddress("0x00000000000000000000000000000000000000c1")
	fallbackKey := common.HexToAddress("0x00000000000000000000000000000000000000c2")
	wrongKey := common.HexToAddress("0x00000000000000000000000000000000000000c9")

	tests := []struct {
		name            string
		noAuthenticator bool // leave BatchAuthenticatorAddress zero
		unscheduled     bool // leave EspressoTime nil
		tipTime         uint64
		tipErr          bool
		// contract is nil for pre-enforcement cases: they run against the bare
		// mockFixedTimeL1Client, whose embedded ContractBackend is nil, so any
		// contract consult panics the test — in-window publish decisions must
		// not depend on contract state (the #492 regression consulted
		// activeIsEspresso in-window).
		contract *contractState
		// teeFrom/fallbackFrom override the role's authorized key (teeKey /
		// fallbackKey) to exercise the identity checks.
		teeFrom          common.Address
		fallbackFrom     common.Address
		wantSkipTee      bool
		wantSkipFallback bool
	}{
		{
			name:             "no authenticator configured",
			noAuthenticator:  true,
			tipTime:          enforcementTime,
			wantSkipTee:      false,
			wantSkipFallback: false,
		},
		{
			name:             "espresso unscheduled",
			unscheduled:      true,
			tipTime:          enforcementTime,
			wantSkipTee:      true,
			wantSkipFallback: false,
		},
		{
			name:             "pre-fork",
			tipTime:          espressoTime - 1,
			wantSkipTee:      true,
			wantSkipFallback: false,
		},
		{
			// The #492 regression: inside the grace window the fallback keeps
			// publishing despite activeIsEspresso defaulting true, and the TEE
			// batcher stands down instead of burning fees on dropped batches.
			name:             "at fork time, grace window opens",
			tipTime:          espressoTime,
			wantSkipTee:      true,
			wantSkipFallback: false,
		},
		{
			name:             "one second before enforcement",
			tipTime:          enforcementTime - 1,
			wantSkipTee:      true,
			wantSkipFallback: false,
		},
		{
			name:             "at enforcement, espresso flagged active",
			tipTime:          enforcementTime,
			contract:         &contractState{activeIsEspresso: true},
			wantSkipTee:      false,
			wantSkipFallback: true,
		},
		{
			name:             "at enforcement, fallback flagged active",
			tipTime:          enforcementTime,
			contract:         &contractState{activeIsEspresso: false},
			wantSkipTee:      true,
			wantSkipFallback: false,
		},
		{
			name:             "at enforcement, espresso active but tee key unauthorized",
			tipTime:          enforcementTime,
			contract:         &contractState{activeIsEspresso: true},
			teeFrom:          wrongKey,
			wantSkipTee:      true,
			wantSkipFallback: true, // flag-skipped, identity not reached
		},
		{
			name:             "at enforcement, fallback active but fallback key unauthorized",
			tipTime:          enforcementTime,
			contract:         &contractState{activeIsEspresso: false},
			fallbackFrom:     wrongKey,
			wantSkipTee:      true, // flag-skipped, identity not reached
			wantSkipFallback: true,
		},
		{
			name:             "at enforcement, contract unavailable fails closed",
			tipTime:          enforcementTime,
			contract:         &contractState{activeIsEspresso: true, callErr: errors.New("contract unavailable")},
			wantSkipTee:      true,
			wantSkipFallback: true,
		},
		{
			name:             "tip fetch failure fails closed",
			tipErr:           true,
			wantSkipTee:      true,
			wantSkipFallback: true,
		},
	}

	roles := []struct {
		name     string
		espresso bool
	}{
		{name: "fallback", espresso: false},
		{name: "tee", espresso: true},
	}

	for _, test := range tests {
		for _, role := range roles {
			t.Run(test.name+"/"+role.name, func(t *testing.T) {
				var l1Client L1Client
				switch {
				case test.tipErr:
					l1Client = &errTipL1Client{}
				case test.contract == nil:
					l1Client = &mockFixedTimeL1Client{time: test.tipTime}
				default:
					l1Client = &fakeBatchAuthBackend{
						mockFixedTimeL1Client: mockFixedTimeL1Client{time: test.tipTime},
						authAddr:              authAddr,
						sysCfgAddr:            sysCfgAddr,
						activeIsEspresso:      test.contract.activeIsEspresso,
						espressoBatcher:       teeKey,
						fallbackBatcher:       fallbackKey,
						callErr:               test.contract.callErr,
					}
				}

				from, override, wantSkip := fallbackKey, test.fallbackFrom, test.wantSkipFallback
				if role.espresso {
					from, override, wantSkip = teeKey, test.teeFrom, test.wantSkipTee
				}
				if override != (common.Address{}) {
					from = override
				}

				logger := testlog.Logger(t, log.LevelDebug)
				l := &BatchSubmitter{}
				l.Log = logger
				l.Metr = metrics.NoopMetrics
				l.RollupConfig = &rollup.Config{}
				if !test.noAuthenticator {
					l.RollupConfig.BatchAuthenticatorAddress = authAddr
				}
				if !test.unscheduled {
					l.RollupConfig.EspressoTime = u64(espressoTime)
				}
				l.Config.NetworkTimeout = time.Second
				l.Config.Espresso.Enabled = role.espresso
				l.L1Client = l1Client
				l.Txmgr = testutils.NewFakeTxMgr(logger, from, eth.ChainIDFromUInt64(0))

				require.Equal(t, wantSkip, l.shouldSkipPublishForActiveSeq(context.Background()))
			})
		}
	}
}
