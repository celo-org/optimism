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
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

// noContractL1Client is an L1Client that serves the tip but fails the test on any
// contract read. Pre-enforcement publish decisions must not depend on contract
// state (the #492 regression consulted activeIsEspresso in-window), and a mock
// with a nil embedded ContractBackend would enforce that by segfaulting the whole
// binary rather than naming the offending call.
type noContractL1Client struct {
	mockFixedTimeL1Client
	t *testing.T
}

func (c *noContractL1Client) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	c.t.Fatal("publish gate read contract code pre-enforcement; the decision must not depend on contract state")
	return nil, nil
}

func (c *noContractL1Client) CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
	c.t.Fatal("publish gate called the contract pre-enforcement; the decision must not depend on contract state")
	return nil, nil
}

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
	t                *testing.T
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
	// Fail the test rather than only the call: the gate turns any error into a
	// skip, so a call dispatched to the wrong address or method would otherwise
	// be indistinguishable from a legitimate fail-closed decision and leave the
	// skip-expecting cases passing for the wrong reason.
	err := fmt.Errorf("unexpected contract call to %s with data %x", call.To, call.Data)
	f.t.Error(err)
	return nil, err
}

// Fixture addresses shared by the publish-gate tests.
var (
	gateAuthAddr    = common.HexToAddress("0x00000000000000000000000000000000000000aa")
	gateSysCfgAddr  = common.HexToAddress("0x00000000000000000000000000000000000000bb")
	gateTeeKey      = common.HexToAddress("0x00000000000000000000000000000000000000c1")
	gateFallbackKey = common.HexToAddress("0x00000000000000000000000000000000000000c2")
)

// newAuthBackend builds a fake BatchAuthenticator view at the given tip time.
func newAuthBackend(t *testing.T, tipTime uint64, activeIsEspresso bool, callErr error) fakeBatchAuthBackend {
	return fakeBatchAuthBackend{
		mockFixedTimeL1Client: mockFixedTimeL1Client{time: tipTime},
		t:                     t,
		authAddr:              gateAuthAddr,
		sysCfgAddr:            gateSysCfgAddr,
		activeIsEspresso:      activeIsEspresso,
		espressoBatcher:       gateTeeKey,
		fallbackBatcher:       gateFallbackKey,
		callErr:               callErr,
	}
}

// newGateSubmitter builds a BatchSubmitter carrying the fields the publish gate
// reads, standing in for the ones NewBatchSubmitter would have populated. A zero
// authAddr or a nil espressoTime leaves that part of the rollup config unset.
func newGateSubmitter(
	t *testing.T,
	l1 L1Client,
	espressoEnabled bool,
	from common.Address,
	espressoTime *uint64,
	authAddr common.Address,
) *BatchSubmitter {
	logger := testlog.Logger(t, log.LevelDebug)
	l := &BatchSubmitter{}
	l.Log = logger
	l.degradedLog = oplog.NewRepeatStateLogger()
	l.Metr = metrics.NoopMetrics
	l.RollupConfig = &rollup.Config{
		EspressoTime:              espressoTime,
		BatchAuthenticatorAddress: authAddr,
	}
	l.Config.NetworkTimeout = time.Second
	l.Config.Espresso.Enabled = espressoEnabled
	l.L1Client = l1
	l.Txmgr = testutils.NewFakeTxMgr(logger, from, eth.ChainIDFromUInt64(0))
	return l
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

	wrongKey := common.HexToAddress("0x00000000000000000000000000000000000000c9")

	tests := []struct {
		name            string
		noAuthenticator bool // leave BatchAuthenticatorAddress zero
		unscheduled     bool // leave EspressoTime nil
		tipTime         uint64
		tipErr          bool
		// contract is nil for pre-enforcement cases: they run against
		// noContractL1Client, which fails the test on any contract read —
		// in-window publish decisions must not depend on contract state (the
		// #492 regression consulted activeIsEspresso in-window).
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
			// publishing whatever activeIsEspresso says, and the TEE batcher
			// stands down instead of burning fees on dropped batches.
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
					l1Client = &noContractL1Client{
						mockFixedTimeL1Client: mockFixedTimeL1Client{time: test.tipTime},
						t:                     t,
					}
				default:
					backend := newAuthBackend(t, test.tipTime, test.contract.activeIsEspresso, test.contract.callErr)
					l1Client = &backend
				}

				from, override, wantSkip := gateFallbackKey, test.fallbackFrom, test.wantSkipFallback
				if role.espresso {
					from, override, wantSkip = gateTeeKey, test.teeFrom, test.wantSkipTee
				}
				if override != (common.Address{}) {
					from = override
				}

				var cfgEspressoTime *uint64
				if !test.unscheduled {
					cfgEspressoTime = u64(espressoTime)
				}
				var cfgAuthAddr common.Address
				if !test.noAuthenticator {
					cfgAuthAddr = gateAuthAddr
				}
				l := newGateSubmitter(t, l1Client, role.espresso, from, cfgEspressoTime, cfgAuthAddr)

				require.Equal(t, wantSkip, l.shouldSkipPublishForActiveSeq(context.Background()))
			})
		}
	}
}

// contractAcceptsAuth mirrors the sender check in
// BatchAuthenticator.authenticateBatchInfo (src/L1/BatchAuthenticator.sol): with
// activeIsEspresso set, only espressoBatcher() may authenticate; otherwise only the
// SystemConfig batcher may. Pinned on the Solidity side by BatchAuthenticator.t.sol —
// mirrored here so a cell can be evaluated without a chain.
func contractAcceptsAuth(activeIsEspresso bool, sender, espressoBatcher, fallbackBatcher common.Address) bool {
	if activeIsEspresso {
		return sender == espressoBatcher
	}
	return sender == fallbackBatcher
}

// batchLands reports whether a batch published by sender would reach the safe chain,
// composing two rules the batcher has to satisfy at once:
//
//   - derivation (isBatchTxAuthorized): pre-enforcement it authorizes on the L1 sender
//     alone, so only the SystemConfig batcher's batches derive no matter what was
//     authenticated; once enforced the batch needs a BatchInfoAuthenticated event from
//     its own sender.
//   - the txmgr pairing contract (SendPairAsync): a reverted auth leg cancels the batch
//     leg, so an auth call the contract rejects stops the batch reaching L1 at all —
//     even where derivation would not have required it.
func batchLands(enforced, authenticates, authAccepted bool, sender, fallbackBatcher common.Address) bool {
	if authenticates && !authAccepted {
		return false // cancelled batch leg
	}
	if enforced {
		return authenticates
	}
	return sender == fallbackBatcher
}

// TestPublishAuthInvariant_RoleTimeFlag checks, across role x L1-tip-region x
// activeIsEspresso, that exactly one role publishes and its batch lands (see
// batchLands). Batches are judged at their decision-time tip, not landing time, so
// the straddle case stays with the fork-boundary tests.
//
// The excluded cell — grace window, activeIsEspresso true — is ruled out by
// deployment, not code: the fallback publishes, its auth call reverts, and the
// cancelled batch leg stalls the safe head. Asserted here as a stall to record the
// dependency on the deploy scripts' false initializer; see op-batcher/readme.md.
func TestPublishAuthInvariant_RoleTimeFlag(t *testing.T) {
	const espressoTime uint64 = 1_000_000

	regions := []struct {
		name string
		tip  uint64
	}{
		{"pre-fork", espressoTime - 1},
		{"grace window", espressoTime},
		{"enforced", espressoTime + derive.BatchAuthEnforcementDelaySecs},
	}
	roles := []struct {
		name     string
		espresso bool
		from     common.Address
	}{
		{"fallback", false, gateFallbackKey},
		{"tee", true, gateTeeKey},
	}

	for _, region := range regions {
		for _, activeIsEspresso := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/activeIsEspresso=%v", region.name, activeIsEspresso), func(t *testing.T) {
				var publishing, landing []string
				for _, role := range roles {
					backend := newAuthBackend(t, region.tip, activeIsEspresso, nil)
					l := newGateSubmitter(t, &backend, role.espresso, role.from, u64(espressoTime), gateAuthAddr)

					if l.shouldSkipPublishForActiveSeq(context.Background()) {
						continue
					}
					publishing = append(publishing, role.name)

					// dispatchAuthenticatedSendTx always authenticates for the TEE
					// role; the fallback consults the auth-dispatch gate.
					authenticates := true
					if !role.espresso {
						required, err := l.isFallbackAuthRequired(context.Background())
						require.NoError(t, err)
						authenticates = required
					}
					accepted := contractAcceptsAuth(activeIsEspresso, role.from, gateTeeKey, gateFallbackKey)
					enforced := region.tip >= espressoTime+derive.BatchAuthEnforcementDelaySecs
					if batchLands(enforced, authenticates, accepted, role.from, gateFallbackKey) {
						landing = append(landing, role.name)
					}
				}

				if region.name == "grace window" && activeIsEspresso {
					// Excluded by deployment, not by code — see the doc comment.
					require.Equal(t, []string{"fallback"}, publishing, "the fallback is the only role that may publish in the window")
					require.Empty(t, landing, "no role can make progress in this cell")
					return
				}
				require.Len(t, publishing, 1, "exactly one role must publish")
				require.Equal(t, publishing, landing, "the publishing role's batch must land")
			})
		}
	}
}

// countingTipL1Client serves a tip whose time and failure mode can be changed
// between calls and counts the tip fetches the gate performs, so the enforcement
// memo's effect is observable. It embeds fakeBatchAuthBackend so isBatcherActive
// still has contract state to read once enforcement is reached.
type countingTipL1Client struct {
	fakeBatchAuthBackend
	tipFetches int
	tipErr     error
}

func (c *countingTipL1Client) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	c.tipFetches++
	if c.tipErr != nil {
		return nil, c.tipErr
	}
	return &types.Header{Number: big.NewInt(0), Time: c.time}, nil
}

// TestIsBatchAuthEnforcedAtTip_Memoization pins the memo in
// isBatchAuthEnforcedAtTip, which the boundary table above cannot reach: every
// case there builds a fresh BatchSubmitter and calls the gate once, so the memo
// is always false on entry.
//
// Enforcement is monotone in wall-clock time, so the first true observation is
// latched. The consequences are that steady state stops fetching the tip, that a
// tip going backwards does not un-enforce, and that a tip fetch failure stops
// failing closed at this gate once latched — the last is a real behaviour change
// and is pinned here deliberately. It is safe because it only shifts the decision
// onto isBatcherActive, whose own contract reads still fail closed.
func TestIsBatchAuthEnforcedAtTip_Memoization(t *testing.T) {
	const espressoTime uint64 = 1_000_000
	enforcementTime := espressoTime + derive.BatchAuthEnforcementDelaySecs

	// The fallback role against activeIsEspresso=false: post-enforcement this is
	// the active batcher, so the gate's answer turns on the boundary alone.
	newSubmitter := func(t *testing.T, tipTime uint64) (*BatchSubmitter, *countingTipL1Client) {
		client := &countingTipL1Client{fakeBatchAuthBackend: newAuthBackend(t, tipTime, false, nil)}
		l := newGateSubmitter(t, client, false, gateFallbackKey, u64(espressoTime), gateAuthAddr)
		return l, client
	}

	t.Run("enforced tip is fetched once, then memoized", func(t *testing.T) {
		l, client := newSubmitter(t, enforcementTime)
		for range 3 {
			enforced, err := l.isBatchAuthEnforcedAtTip(context.Background())
			require.NoError(t, err)
			require.True(t, enforced)
		}
		require.Equal(t, 1, client.tipFetches, "later calls must short-circuit on the memo")
	})

	t.Run("pre-enforcement tip is re-fetched every call", func(t *testing.T) {
		l, client := newSubmitter(t, enforcementTime-1)
		for range 3 {
			enforced, err := l.isBatchAuthEnforcedAtTip(context.Background())
			require.NoError(t, err)
			require.False(t, enforced)
		}
		require.Equal(t, 3, client.tipFetches, "nothing is latched until the boundary is observed")
	})

	t.Run("tip falling back below the boundary stays enforced", func(t *testing.T) {
		l, client := newSubmitter(t, enforcementTime)
		enforced, err := l.isBatchAuthEnforcedAtTip(context.Background())
		require.NoError(t, err)
		require.True(t, enforced)

		// A deep L1 reorg could put the tip back inside the grace window. The
		// verifier judges each batch by its own L1 origin time, so un-enforcing
		// here would only hand publishing back to a role whose batches the
		// post-reorg chain may already require events for.
		client.time = espressoTime
		enforced, err = l.isBatchAuthEnforcedAtTip(context.Background())
		require.NoError(t, err)
		require.True(t, enforced)
		require.Equal(t, 1, client.tipFetches)
	})

	t.Run("memoized enforcement no longer fails closed on a tip fetch failure", func(t *testing.T) {
		l, client := newSubmitter(t, enforcementTime)
		require.False(t, l.shouldSkipPublishForActiveSeq(context.Background()),
			"the fallback is the active batcher at enforcement")

		client.tipErr = errors.New("l1 tip unavailable")
		require.False(t, l.shouldSkipPublishForActiveSeq(context.Background()),
			"the memo answers the boundary without a tip fetch, so the gate stays on isBatcherActive")
		require.Equal(t, 1, client.tipFetches)
	})
}
