package engine

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/event"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

// TestDerivedPayloadBelowUnsafeHead covers a derived block that lands below the unsafe head,
// which happens when consolidation rebuilds a block whose attributes did not match. If the EL
// has that exact block canonical under a still-canonical unsafe head, the unsafe chain is kept
// and only the safe heads advance. In every other case, a missing block or failed lookup
// included, the derived block replaces the unsafe chain, as does a derived block at the unsafe
// head's own height.
func TestDerivedPayloadBelowUnsafeHead(t *testing.T) {
	finalized := eth.L2BlockRef{Hash: common.Hash{0xf0}, Number: 5}
	parent := eth.L2BlockRef{Hash: common.Hash{0x09}, Number: 9}
	derived := eth.L2BlockRef{Hash: common.Hash{0x0a}, Number: 10, ParentHash: parent.Hash}
	unsafe := eth.L2BlockRef{Hash: common.Hash{0x0d}, Number: 13}
	l1 := eth.L1BlockRef{Hash: common.Hash{0x1a}, Number: 100}
	envelope := &eth.ExecutionPayloadEnvelope{ExecutionPayload: &eth.ExecutionPayload{
		ParentHash:  parent.Hash,
		BlockHash:   derived.Hash,
		BlockNumber: eth.Uint64Quantity(derived.Number),
	}}
	const lookupWarning = "Cannot check whether the unsafe chain builds on the derived block, replacing the unsafe chain"

	tests := []struct {
		name string
		// unsafeHead overrides the default unsafe head above the derived block.
		unsafeHead  *eth.L2BlockRef
		expectEL    func(el *testutils.MockEngine)
		keepsUnsafe bool
		lookupFails bool
	}{
		{
			name: "canonical block keeps the unsafe chain",
			expectEL: func(el *testutils.MockEngine) {
				el.ExpectL2BlockRefByNumber(derived.Number, derived, nil)
				el.ExpectL2BlockRefByNumber(unsafe.Number, unsafe, nil)
			},
			keepsUnsafe: true,
		},
		{
			name: "different canonical block replaces the unsafe chain",
			expectEL: func(el *testutils.MockEngine) {
				el.ExpectL2BlockRefByNumber(derived.Number, eth.L2BlockRef{Hash: common.Hash{0xaa}, Number: derived.Number}, nil)
			},
		},
		{
			name: "non-canonical unsafe head replaces the unsafe chain",
			expectEL: func(el *testutils.MockEngine) {
				el.ExpectL2BlockRefByNumber(derived.Number, derived, nil)
				el.ExpectL2BlockRefByNumber(unsafe.Number, eth.L2BlockRef{Hash: common.Hash{0xdd}, Number: unsafe.Number}, nil)
			},
		},
		{
			name: "missing canonical block replaces the unsafe chain",
			expectEL: func(el *testutils.MockEngine) {
				el.ExpectL2BlockRefByNumber(derived.Number, eth.L2BlockRef{},
					fmt.Errorf("failed to determine L2BlockRef of height %d: %w", derived.Number, ethereum.NotFound))
			},
		},
		{
			name: "failed lookup replaces the unsafe chain",
			expectEL: func(el *testutils.MockEngine) {
				el.ExpectL2BlockRefByNumber(derived.Number, eth.L2BlockRef{}, errors.New("connection refused"))
			},
			lookupFails: true,
		},
		{
			name:       "block at the unsafe head height takes the regular path",
			unsafeHead: &derived,
			expectEL:   func(el *testutils.MockEngine) {},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger, logs := testlog.CaptureLogger(t, log.LevelInfo)
			el := &testutils.MockEngine{}
			// An unexpected engine call fails this case instead of panicking.
			el.Test(t)
			tc.expectEL(el)
			head := unsafe
			if tc.unsafeHead != nil {
				head = *tc.unsafeHead
			}
			var announced []ForkchoiceUpdateEvent
			emitter := &testutils.MockEmitter{}
			emitter.ExpectMaybeRun(func(ev event.Event) {
				if fc, ok := ev.(ForkchoiceUpdateEvent); ok {
					announced = append(announced, fc)
				}
			})
			ec := NewEngineController(context.Background(), el, logger, metrics.NoopMetrics,
				&rollup.Config{}, &sync.Config{}, &testutils.MockL1Source{}, emitter, nil)
			// Derivation is behind the unsafe chain, with the EL in sync with op-node.
			ec.unsafeHead = head
			ec.pendingSafeHead = parent
			ec.localSafeHead = parent
			ec.deprecatedSafeHead = parent
			ec.localFinalizedHead = finalized
			ec.deprecatedFinalizedHead = finalized
			ec.lastForkchoice = eth.ForkchoiceState{
				HeadBlockHash:      head.Hash,
				SafeBlockHash:      parent.Hash,
				FinalizedBlockHash: finalized.Hash,
			}
			wantHead := head
			if !tc.keepsUnsafe {
				wantHead = derived
				el.ExpectForkchoiceUpdate(&eth.ForkchoiceState{
					HeadBlockHash:      derived.Hash,
					SafeBlockHash:      derived.Hash,
					FinalizedBlockHash: finalized.Hash,
				}, nil, &eth.ForkchoiceUpdatedResult{PayloadStatus: eth.PayloadStatusV1{Status: eth.ExecutionValid}}, nil)
			}

			ec.OnEvent(context.Background(), PayloadSuccessEvent{
				Concluding:    true,
				DerivedFrom:   l1,
				InsertStarted: time.Now(),
				Envelope:      envelope,
				Ref:           derived,
			})

			// The mock fails on any engine call it does not expect, so on the kept chain this
			// also checks that the EL received no forkchoice update, and at the unsafe head's
			// height that no canonical lookup was made.
			el.AssertExpectations(t)
			require.Equal(t, wantHead, ec.UnsafeL2Head())
			require.Equal(t, derived, ec.PendingSafeL2Head())
			require.Equal(t, derived, ec.SafeL2Head())
			require.Equal(t, []ForkchoiceUpdateEvent{{
				UnsafeL2Head:    wantHead,
				SafeL2Head:      derived,
				FinalizedL2Head: finalized,
			}}, announced, "listeners learn the resulting unsafe head once")
			var problems, wantProblems []string
			for _, record := range logs.FindLogs(func(r *testlog.CapturedRecord) bool { return r.Record.Level >= log.LevelWarn }) {
				problems = append(problems, record.Message)
			}
			if tc.lookupFails {
				wantProblems = []string{lookupWarning}
			}
			require.Equal(t, wantProblems, problems)
		})
	}
}
