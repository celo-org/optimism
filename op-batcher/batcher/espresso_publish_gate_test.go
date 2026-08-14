package batcher

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

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

func sel4(sig string) (s [4]byte) {
	copy(s[:], gethcrypto.Keccak256([]byte(sig)))
	return
}

// fakeBatchAuthBackend extends mockFixedTimeL1Client with a CodeAt and
// CallContract that serve the read-only BatchAuthenticator/SystemConfig calls
// isBatcherActive makes, dispatching on (contract address, 4-byte selector).
// Pre-enforcement test cases deliberately use the bare mockFixedTimeL1Client
// instead: its embedded ContractBackend is nil, so any contract consult
// panics the test — the publish decision inside the grace window must not
// depend on contract state.
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
	var sel [4]byte
	copy(sel[:], call.Data)
	switch {
	case *call.To == f.authAddr && sel == sel4("activeIsEspresso()"):
		var out common.Hash
		if f.activeIsEspresso {
			out[31] = 1
		}
		return out[:], nil
	case *call.To == f.authAddr && sel == sel4("espressoBatcher()"):
		return common.BytesToHash(f.espressoBatcher.Bytes()).Bytes(), nil
	case *call.To == f.authAddr && sel == sel4("systemConfig()"):
		return common.BytesToHash(f.sysCfgAddr.Bytes()).Bytes(), nil
	case *call.To == f.sysCfgAddr && sel == sel4("batcherHash()"):
		return common.BytesToHash(f.fallbackBatcher.Bytes()).Bytes(), nil
	}
	return nil, fmt.Errorf("unexpected contract call to %s with selector %x", call.To, sel)
}

// TestShouldSkipPublish_EnforcementBoundary locks batch-submission ownership to
// the verifier's event-auth enforcement boundary (issue #492): the fallback
// batcher owns publishing before enforcement — including the whole grace window,
// regardless of activeIsEspresso — and the TEE batcher owns it after, subject to
// the activeIsEspresso flag and the sender-identity check. A divergence here
// either burns L1 fees on batches every verifier drops (TEE publishing
// in-window) or stalls the safe head with no publisher at all (fallback standing
// down in-window).
func TestShouldSkipPublish_EnforcementBoundary(t *testing.T) {
	const espressoTime uint64 = 1_000_000
	enforcementTime := espressoTime + derive.BatchAuthEnforcementDelaySecs

	authAddr := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	sysCfgAddr := common.HexToAddress("0x00000000000000000000000000000000000000bb")
	teeKey := common.HexToAddress("0x00000000000000000000000000000000000000c1")
	fallbackKey := common.HexToAddress("0x00000000000000000000000000000000000000c2")
	wrongKey := common.HexToAddress("0x00000000000000000000000000000000000000c9")

	// contractBackend builds the post-enforcement L1 client, parameterized on
	// the tip time and the contract's flag/error state.
	contractBackend := func(tipTime uint64, activeIsEspresso bool, callErr error) *fakeBatchAuthBackend {
		return &fakeBatchAuthBackend{
			mockFixedTimeL1Client: mockFixedTimeL1Client{time: tipTime},
			authAddr:              authAddr,
			sysCfgAddr:            sysCfgAddr,
			activeIsEspresso:      activeIsEspresso,
			espressoBatcher:       teeKey,
			fallbackBatcher:       fallbackKey,
			callErr:               callErr,
		}
	}

	tests := []struct {
		name            string
		espressoEnabled bool
		authAddr        common.Address
		espressoTime    *uint64
		l1Client        L1Client
		from            common.Address
		wantSkip        bool
	}{
		// No authenticator configured: gate disabled for both roles.
		{
			name:            "fallback without authenticator never skips",
			espressoEnabled: false,
			authAddr:        common.Address{},
			espressoTime:    u64(espressoTime),
			l1Client:        &mockFixedTimeL1Client{time: enforcementTime},
			wantSkip:        false,
		},
		{
			name:            "tee without authenticator never skips",
			espressoEnabled: true,
			authAddr:        common.Address{},
			espressoTime:    u64(espressoTime),
			l1Client:        &mockFixedTimeL1Client{time: enforcementTime},
			wantSkip:        false,
		},
		// Pre-enforcement: the bare mockFixedTimeL1Client panics on any
		// contract consult, so these cases also prove the contract is not read.
		{
			name:            "fallback publishes when espresso is unscheduled",
			espressoEnabled: false,
			authAddr:        authAddr,
			espressoTime:    nil,
			l1Client:        &mockFixedTimeL1Client{time: enforcementTime},
			wantSkip:        false,
		},
		{
			name:            "tee skips when espresso is unscheduled",
			espressoEnabled: true,
			authAddr:        authAddr,
			espressoTime:    nil,
			l1Client:        &mockFixedTimeL1Client{time: enforcementTime},
			wantSkip:        true,
		},
		{
			name:            "fallback publishes pre-fork",
			espressoEnabled: false,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        &mockFixedTimeL1Client{time: espressoTime - 1},
			wantSkip:        false,
		},
		{
			name:            "tee skips pre-fork",
			espressoEnabled: true,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        &mockFixedTimeL1Client{time: espressoTime - 1},
			wantSkip:        true,
		},
		// The #492 regression: inside the grace window the fallback must keep
		// publishing without consulting activeIsEspresso, and the TEE batcher
		// must stand down instead of burning fees on batches derivation drops.
		{
			name:            "fallback publishes at fork time despite activeIsEspresso",
			espressoEnabled: false,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        &mockFixedTimeL1Client{time: espressoTime},
			wantSkip:        false,
		},
		{
			name:            "tee skips at fork time",
			espressoEnabled: true,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        &mockFixedTimeL1Client{time: espressoTime},
			wantSkip:        true,
		},
		{
			name:            "fallback publishes one second before enforcement",
			espressoEnabled: false,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        &mockFixedTimeL1Client{time: enforcementTime - 1},
			wantSkip:        false,
		},
		{
			name:            "tee skips one second before enforcement",
			espressoEnabled: true,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        &mockFixedTimeL1Client{time: enforcementTime - 1},
			wantSkip:        true,
		},
		// Post-enforcement: activeIsEspresso plus the identity check decide.
		{
			name:            "tee publishes exactly at enforcement when active and authorized",
			espressoEnabled: true,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        contractBackend(enforcementTime, true, nil),
			from:            teeKey,
			wantSkip:        false,
		},
		{
			name:            "fallback honors activeIsEspresso once enforced",
			espressoEnabled: false,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        contractBackend(enforcementTime, true, nil),
			from:            fallbackKey,
			wantSkip:        true,
		},
		{
			name:            "tee skips once enforced when fallback is active",
			espressoEnabled: true,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        contractBackend(enforcementTime+1, false, nil),
			from:            teeKey,
			wantSkip:        true,
		},
		{
			name:            "fallback publishes once enforced when flagged active and authorized",
			espressoEnabled: false,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        contractBackend(enforcementTime+1, false, nil),
			from:            fallbackKey,
			wantSkip:        false,
		},
		{
			name:            "tee skips once enforced when its key is not the espresso batcher",
			espressoEnabled: true,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        contractBackend(enforcementTime, true, nil),
			from:            wrongKey,
			wantSkip:        true,
		},
		{
			name:            "fallback skips once enforced when its key is not the systemconfig batcher",
			espressoEnabled: false,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        contractBackend(enforcementTime, false, nil),
			from:            wrongKey,
			wantSkip:        true,
		},
		// Fail closed on unavailable gate inputs.
		{
			name:            "tee fails closed on contract errors once enforced",
			espressoEnabled: true,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        contractBackend(enforcementTime, true, errors.New("contract unavailable")),
			from:            teeKey,
			wantSkip:        true,
		},
		{
			name:            "fallback fails closed on contract errors once enforced",
			espressoEnabled: false,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        contractBackend(enforcementTime, false, errors.New("contract unavailable")),
			from:            fallbackKey,
			wantSkip:        true,
		},
		{
			name:            "fallback fails closed on tip fetch errors",
			espressoEnabled: false,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        &errTipL1Client{},
			wantSkip:        true,
		},
		{
			name:            "tee fails closed on tip fetch errors",
			espressoEnabled: true,
			authAddr:        authAddr,
			espressoTime:    u64(espressoTime),
			l1Client:        &errTipL1Client{},
			wantSkip:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := testlog.Logger(t, log.LevelDebug)
			l := &BatchSubmitter{}
			l.Log = logger
			l.Metr = metrics.NoopMetrics
			l.RollupConfig = &rollup.Config{
				BatchAuthenticatorAddress: test.authAddr,
				EspressoTime:              test.espressoTime,
			}
			l.Config.NetworkTimeout = time.Second
			l.Config.Espresso.Enabled = test.espressoEnabled
			l.L1Client = test.l1Client
			l.Txmgr = testutils.NewFakeTxMgr(logger, test.from, eth.ChainIDFromUInt64(0))

			require.Equal(t, test.wantSkip, l.shouldSkipPublishForActiveSeq(context.Background()))
		})
	}
}
