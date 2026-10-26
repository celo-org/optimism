package batcher

// These tests cover block-range selection, sync reconciliation, and safe-head
// re-anchoring during resets. Draining is covered in espresso_loading_test.go;
// session lifecycle and setup are covered in espresso_driver_test.go.

import (
	"context"
	"errors"
	"math/big"
	"math/rand"
	"testing"
	"time"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	derivetest "github.com/ethereum-optimism/optimism/op-node/rollup/derive/test"
	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

type reanchorTestStreamer struct {
	espressoStreamer
	positions []eth.L2BlockRef
}

func (s *reanchorTestStreamer) SetBatchPosition(position eth.L2BlockRef) {
	s.positions = append(s.positions, position)
}

func newTestBlockLoader(t *testing.T) *BlockLoader {
	return &BlockLoader{
		batcher: &BatchSubmitter{
			DriverSetup: DriverSetup{Log: testlog.Logger(t, log.LevelDebug)},
			degradedLog: oplog.NewRepeatStateLogger(),
		},
	}
}

// op-node has transiently reported sync statuses with individual fields zeroed
// while the rest are populated (see the "LocalSafeL2=0,SafeL2>0" case in
// sync_actions_test.go). The loader must retry rather than treat a zero
// LocalSafeL2 as the queue floor and enqueue the whole pre-caffeination history.
func TestNextBlockRangeZeroStatusFields(t *testing.T) {
	populated := eth.SyncStatus{
		HeadL1:      eth.L1BlockRef{Number: 5, Hash: common.Hash{0x01}},
		CurrentL1:   eth.L1BlockRef{Number: 2, Hash: common.Hash{0x02}},
		LocalSafeL2: eth.L2BlockRef{Number: 104, Hash: common.Hash{0x03}},
		SafeL2:      eth.L2BlockRef{Number: 103, Hash: common.Hash{0x04}},
		UnsafeL2:    eth.L2BlockRef{Number: 109, Hash: common.Hash{0x05}},
	}

	t.Run("fully populated status enqueues from local-safe", func(t *testing.T) {
		loader := newTestBlockLoader(t)
		r, action := loader.nextBlockRange(&populated)
		require.Equal(t, EnqueueBlockAction(ActionEnqueue), action)
		require.Equal(t, inclusiveBlockRange{105, 109}, r)
	})

	zeroings := map[string]func(*eth.SyncStatus){
		"LocalSafeL2": func(s *eth.SyncStatus) { s.LocalSafeL2 = eth.L2BlockRef{} },
		"UnsafeL2":    func(s *eth.SyncStatus) { s.UnsafeL2 = eth.L2BlockRef{} },
		"HeadL1":      func(s *eth.SyncStatus) { s.HeadL1 = eth.L1BlockRef{} },
		"CurrentL1":   func(s *eth.SyncStatus) { s.CurrentL1 = eth.L1BlockRef{} },
	}
	for field, zero := range zeroings {
		t.Run("zero "+field+" retries", func(t *testing.T) {
			status := populated
			zero(&status)
			loader := newTestBlockLoader(t)
			r, action := loader.nextBlockRange(&status)
			require.Equal(t, EnqueueBlockAction(ActionRetry), action)
			require.Equal(t, inclusiveBlockRange{}, r)
			require.Nil(t, loader.prevSyncStatus,
				"a zeroed status must not become the comparison baseline for later ticks")
		})
	}
}

// espressoSyncChannelManager must report an out-of-sync status so that
// espressoBatchLoadingLoop skips draining against an untrustworthy local-safe
// floor: with LocalSafeL2 zeroed, the stale-batch re-anchor check in the drain
// loop never fires and already-derived blocks would be republished.
func TestEspressoSyncChannelManagerReportsOutOfSync(t *testing.T) {
	newTestSubmitter := func(t *testing.T) *BatchSubmitter {
		lgr := testlog.Logger(t, log.LevelDebug)
		return &BatchSubmitter{
			DriverSetup: DriverSetup{Log: lgr},
			channelMgr:  NewChannelManager(lgr, metrics.NoopMetrics, ChannelConfig{}, &rollup.Config{}),
			degradedLog: oplog.NewRepeatStateLogger(),
		}
	}

	populated := eth.SyncStatus{
		HeadL1:      eth.L1BlockRef{Number: 5, Hash: common.Hash{0x01}},
		CurrentL1:   eth.L1BlockRef{Number: 2, Hash: common.Hash{0x02}},
		LocalSafeL2: eth.L2BlockRef{Number: 104, Hash: common.Hash{0x03}},
		SafeL2:      eth.L2BlockRef{Number: 103, Hash: common.Hash{0x04}},
		UnsafeL2:    eth.L2BlockRef{Number: 109, Hash: common.Hash{0x05}},
	}

	t.Run("fully populated status is in sync", func(t *testing.T) {
		l := newTestSubmitter(t)
		require.False(t, l.espressoSyncChannelManager(&populated, 0))
		require.Equal(t, populated.CurrentL1, l.prevCurrentL1)
	})

	t.Run("zeroed LocalSafeL2 is out of sync", func(t *testing.T) {
		l := newTestSubmitter(t)
		status := populated
		status.LocalSafeL2 = eth.L2BlockRef{}
		require.True(t, l.espressoSyncChannelManager(&status, 0))
		require.Zero(t, l.prevCurrentL1,
			"an out-of-sync status must not advance the CurrentL1 baseline")
	})

	t.Run("reversed CurrentL1 is out of sync", func(t *testing.T) {
		l := newTestSubmitter(t)
		l.prevCurrentL1 = eth.L1BlockRef{Number: 3, Hash: common.Hash{0x06}}
		require.True(t, l.espressoSyncChannelManager(&populated, 0))
	})
}

func TestEspressoSyncChannelManagerRejectsStaleEpoch(t *testing.T) {
	l, _ := setup(t, nil)
	block := newMiniL2BlockWithChainIDAndNumberParent(0, l.RollupConfig.L2ChainID, big.NewInt(20), common.Hash{})
	require.NoError(t, l.channelMgr.AddL2Block(block))
	streamer := &reanchorTestStreamer{}
	l.espressoSession = &espressoSession{streamer: streamer}
	l.espressoClearEpoch = 1
	l.prevCurrentL1 = eth.L1BlockRef{Number: 5}
	status := &eth.SyncStatus{
		HeadL1:      eth.L1BlockRef{Number: 6},
		CurrentL1:   eth.L1BlockRef{Number: 6},
		LocalSafeL2: eth.L2BlockRef{Number: 1, L1Origin: eth.BlockID{Number: 4}},
		UnsafeL2:    eth.L2BlockRef{Number: 21},
	}

	// This old safe head would clear the nonempty manager if the epoch were
	// ignored. A stale snapshot must leave both state and its baseline intact.
	require.True(t, l.espressoSyncChannelManager(status, 0))
	require.Equal(t, eth.L1BlockRef{Number: 5}, l.prevCurrentL1)
	require.Equal(t, 1, l.channelMgr.blocks.Len())
	require.Equal(t, block.Hash(), l.channelMgr.blocks[0].Hash())
	require.Empty(t, streamer.positions)

	// A snapshot taken in the current epoch can reconcile normally.
	require.False(t, l.espressoSyncChannelManager(status, 1))
	require.Equal(t, status.CurrentL1, l.prevCurrentL1)
	require.Zero(t, l.channelMgr.blocks.Len())
	require.Equal(t, []eth.L2BlockRef{status.LocalSafeL2}, streamer.positions)
}

// clearState must not clear the channel manager unless the streamer re-anchor
// target is already in hand: espressoReanchorTarget reporting ok=false makes
// the caller retry the whole clear instead of performing it partially, which
// would leave an emptied channel manager with the streamer at its old cursor.
func TestEspressoReanchorTarget(t *testing.T) {
	newReanchorSubmitter := func(t *testing.T, session *espressoSession) (*BatchSubmitter, *mockL2EndpointProvider) {
		ep := newEndpointProvider()
		return &BatchSubmitter{
			DriverSetup: DriverSetup{
				Log:              testlog.Logger(t, log.LevelDebug),
				Config:           BatcherConfig{NetworkTimeout: time.Second},
				EndpointProvider: ep,
			},
			espressoSession: session,
		}, ep
	}

	t.Run("no session needs no target", func(t *testing.T) {
		l, _ := newReanchorSubmitter(t, nil)
		target, ok := l.espressoReanchorTarget(t.Context())
		require.True(t, ok)
		require.Nil(t, target)
	})

	t.Run("sync status failure blocks the clear", func(t *testing.T) {
		l, ep := newReanchorSubmitter(t, &espressoSession{streamer: &espressoStreamers.Streamer{}})
		ep.rollupClient.ExpectSyncStatus(nil, errors.New("transient RPC failure"))
		target, ok := l.espressoReanchorTarget(t.Context())
		require.False(t, ok)
		require.Nil(t, target)
	})

	t.Run("zeroed LocalSafeL2 blocks the clear", func(t *testing.T) {
		l, ep := newReanchorSubmitter(t, &espressoSession{streamer: &espressoStreamers.Streamer{}})
		ep.rollupClient.ExpectSyncStatus(&eth.SyncStatus{
			HeadL1: eth.L1BlockRef{Number: 5, Hash: common.Hash{0x01}},
		}, nil)
		target, ok := l.espressoReanchorTarget(t.Context())
		require.False(t, ok)
		require.Nil(t, target)
	})

	t.Run("populated status yields the local-safe target", func(t *testing.T) {
		l, ep := newReanchorSubmitter(t, &espressoSession{streamer: &espressoStreamers.Streamer{}})
		localSafe := eth.L2BlockRef{Number: 104, Hash: common.Hash{0x03}}
		ep.rollupClient.ExpectSyncStatus(&eth.SyncStatus{
			HeadL1:      eth.L1BlockRef{Number: 5, Hash: common.Hash{0x01}},
			LocalSafeL2: localSafe,
		}, nil)
		target, ok := l.espressoReanchorTarget(t.Context())
		require.True(t, ok)
		require.NotNil(t, target)
		require.Equal(t, localSafe, *target)
	})
}

func TestEspressoClearStateUsesOneSyncSnapshot(t *testing.T) {
	origin := eth.BlockID{Number: 7, Hash: common.Hash{7}}
	localSafe := eth.L2BlockRef{Number: 10, Hash: common.Hash{10}, L1Origin: origin}
	for _, scenario := range []struct {
		name      string
		session   bool
		localSafe eth.L2BlockRef
	}{
		{name: "active session", session: true, localSafe: localSafe},
		{name: "active session at genesis origin", session: true, localSafe: eth.L2BlockRef{Number: 10, Hash: common.Hash{10}}},
		{name: "no session", localSafe: localSafe},
		{name: "no session before local-safe head", localSafe: eth.L2BlockRef{}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			batcher, endpoints := setup(t, nil)
			batcher.Config.NetworkTimeout = time.Second
			streamer := &reanchorTestStreamer{}
			if scenario.session {
				batcher.espressoSession = &espressoSession{streamer: streamer}
			}
			endpoints.rollupClient.ExpectSyncStatus(&eth.SyncStatus{
				HeadL1:      eth.L1BlockRef{Number: 20},
				LocalSafeL2: scenario.localSafe,
			}, nil)
			require.NoError(t, batcher.channelMgr.AddL2Block(newMiniL2Block(0)))

			batcher.clearState(t.Context())

			wantOrigin := scenario.localSafe.L1Origin
			if wantOrigin.Number == 0 {
				wantOrigin = batcher.RollupConfig.Genesis.L1
			}
			require.Equal(t, wantOrigin, batcher.channelMgr.l1OriginLastSubmittedChannel)
			require.Zero(t, batcher.channelMgr.blocks.Len())
			require.Zero(t, batcher.channelMgr.tip)
			require.EqualValues(t, 1, batcher.espressoClearEpoch)
			if scenario.session {
				require.Equal(t, []eth.L2BlockRef{scenario.localSafe}, streamer.positions)
			}
			endpoints.rollupClient.AssertNumberOfCalls(t, "SyncStatus", 1)
		})
	}
}

func TestEspressoClearStateRejectsZeroLocalSafeHead(t *testing.T) {
	batcher, endpoints := setup(t, nil)
	batcher.Config.NetworkTimeout = time.Second
	streamer := &reanchorTestStreamer{}
	batcher.espressoSession = &espressoSession{streamer: streamer}
	endpoints.rollupClient.ExpectSyncStatus(&eth.SyncStatus{
		HeadL1: eth.L1BlockRef{Number: 20},
	}, nil)
	require.NoError(t, batcher.channelMgr.AddL2Block(newMiniL2Block(0)))

	// Cancel the retry immediately. The invalid attempt must not re-anchor;
	// only the cancellation fallback may clear the manager to an empty origin.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	batcher.clearState(ctx)

	require.Empty(t, streamer.positions)
	require.Zero(t, batcher.channelMgr.l1OriginLastSubmittedChannel)
	require.Zero(t, batcher.channelMgr.blocks.Len())
	require.EqualValues(t, 1, batcher.espressoClearEpoch)
	endpoints.rollupClient.AssertNumberOfCalls(t, "SyncStatus", 1)
}

// numberedRef builds an L2BlockRef whose hash encodes its number, so hash
// comparisons in the loader are meaningful without building real blocks.
func numberedRef(n uint64) eth.L2BlockRef {
	return eth.L2BlockRef{Number: n, Hash: common.Hash{byte(n)}}
}

// numberedRefs builds the contiguous ref chain [from, to].
func numberedRefs(from, to uint64) []eth.L2BlockRef {
	refs := make([]eth.L2BlockRef, 0, to-from+1)
	for n := from; n <= to; n++ {
		refs = append(refs, numberedRef(n))
	}
	return refs
}

// TestNextBlockRangeBranches locks the reset/reorg/prune branches of
// nextBlockRange. The safe-chain-reorg hash check is the sole defense against
// re-queueing orphaned blocks to Espresso, and every reset branch feeds
// loader.reset() -> requestClearState, so a silently wrong action here either
// resubmits derived history or drops blocks.
func TestNextBlockRangeBranches(t *testing.T) {
	newStatus := func(localSafe eth.L2BlockRef, unsafeNum uint64) *eth.SyncStatus {
		return &eth.SyncStatus{
			HeadL1:      eth.L1BlockRef{Number: 50, Hash: common.Hash{0xa1}},
			CurrentL1:   eth.L1BlockRef{Number: 20, Hash: common.Hash{0xa2}},
			LocalSafeL2: localSafe,
			UnsafeL2:    numberedRef(unsafeNum),
		}
	}

	tests := []struct {
		name       string
		queued     []eth.L2BlockRef
		safe       eth.L2BlockRef
		unsafe     uint64
		wantAction EnqueueBlockAction
		wantRange  inclusiveBlockRange
		wantQueue  []eth.L2BlockRef // nil: unchanged
	}{
		{
			name:       "queue caught up with the unsafe head retries",
			queued:     numberedRefs(105, 109),
			safe:       numberedRef(104),
			unsafe:     109,
			wantAction: ActionRetry,
		},
		{
			name:       "derivation ahead of the whole queue resets",
			queued:     numberedRefs(105, 106),
			safe:       numberedRef(107),
			unsafe:     110,
			wantAction: ActionReset,
		},
		{
			name:       "safe head below the oldest queued block resets",
			queued:     numberedRefs(105, 108),
			safe:       numberedRef(103),
			unsafe:     110,
			wantAction: ActionReset,
		},
		{
			// A queue with a hole (106 missing, e.g. after a partial enqueue
			// failure) makes the count-based check the only guard.
			name:       "gapped queue with the safe head above its span resets",
			queued:     []eth.L2BlockRef{numberedRef(105), numberedRef(107)},
			safe:       numberedRef(107),
			unsafe:     110,
			wantAction: ActionReset,
		},
		{
			// Same height we queued at 106, different (canonical) hash: our
			// queued chain was orphaned and must not keep feeding Espresso.
			name:       "safe-chain reorg is detected by hash mismatch",
			queued:     numberedRefs(105, 108),
			safe:       eth.L2BlockRef{Number: 106, Hash: common.Hash{0xde, 0xad}},
			unsafe:     110,
			wantAction: ActionReset,
		},
		{
			// Blocks strictly below the safe head are derived and must be dropped.
			name:       "matching safe hash prunes derived blocks and enqueues the rest",
			queued:     numberedRefs(105, 108),
			safe:       numberedRef(106),
			unsafe:     110,
			wantAction: ActionEnqueue,
			wantRange:  inclusiveBlockRange{109, 110},
			wantQueue:  numberedRefs(106, 108),
		},
		{
			// Hash deliberately unrelated to the queue: nothing of ours sits at
			// the safe height yet, so no hash comparison (and no prune) may fire.
			name:       "safe head just below the queue enqueues without pruning",
			queued:     numberedRefs(105, 108),
			safe:       eth.L2BlockRef{Number: 104, Hash: common.Hash{0xbe, 0xef}},
			unsafe:     110,
			wantAction: ActionEnqueue,
			wantRange:  inclusiveBlockRange{109, 110},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loader := newTestBlockLoader(t)
			loader.queuedBlocks = test.queued
			r, action := loader.nextBlockRange(newStatus(test.safe, test.unsafe))
			require.Equal(t, test.wantAction, action)
			require.Equal(t, test.wantRange, r)
			if test.wantQueue == nil {
				test.wantQueue = test.queued
			}
			require.Equal(t, test.wantQueue, loader.queuedBlocks)
		})
	}

	t.Run("reversed CurrentL1 retries and keeps the old baseline", func(t *testing.T) {
		loader := newTestBlockLoader(t)
		prev := newStatus(numberedRef(104), 109)
		prev.CurrentL1.Number = 30 // ahead of the new status' CurrentL1 (20)
		loader.prevSyncStatus = prev
		r, action := loader.nextBlockRange(newStatus(numberedRef(104), 109))
		require.Equal(t, EnqueueBlockAction(ActionRetry), action)
		require.Equal(t, inclusiveBlockRange{}, r)
		require.Same(t, prev, loader.prevSyncStatus,
			"a reversed status must not become the comparison baseline")
	})
}

// fakeChainSigner returns a constant signature; EnqueueBlocks only needs
// signing to succeed, not to verify.
type fakeChainSigner struct{ opcrypto.ChainSigner }

func (fakeChainSigner) Sign(context.Context, []byte) ([]byte, error) {
	return make([]byte, 65), nil
}

// enqueueTest is a BatchSubmitter that can run EnqueueBlocks end to end: fetch
// a block from the (mock) sequencer, convert it, sign it, and hand it to the
// Espresso transaction submitter, which reports each submission on submitted.
type enqueueTest struct {
	bs        *BatchSubmitter
	ep        *mockL2EndpointProvider
	streamer  *reanchorTestStreamer
	submitted chan espressoCommon.Transaction
}

func newEnqueueTest(t *testing.T) *enqueueTest {
	bs, ep := setup(t, nil)
	bs.Config.NetworkTimeout = time.Second
	bs.Espresso.ChainSigner = fakeChainSigner{}
	submitted := make(chan espressoCommon.Transaction, 8)
	submitter, _ := newControlledSubmitter(t, &controlledEspressoClient{
		submit: func(ctx context.Context, tx espressoCommon.Transaction) (*espressoCommon.TaggedBase64, error) {
			submitted <- tx
			<-ctx.Done() // park the lifecycle until the test's cleanup cancels it
			return nil, ctx.Err()
		},
	})
	streamer := &reanchorTestStreamer{}
	bs.espressoSession = &espressoSession{submitter: submitter, streamer: streamer}
	return &enqueueTest{bs: bs, ep: ep, streamer: streamer, submitted: submitted}
}

// TestEnqueueBlocksReorgDetection locks EnqueueBlocks' parent-hash check — the
// loader-side reorg defense — and the reset it triggers: the queued chain is
// dropped, the channel manager is cleared, and the streamer is re-anchored to
// the local-safe head before anything from the orphaned chain reaches Espresso.
func TestEnqueueBlocksReorgDetection(t *testing.T) {
	rng := rand.New(rand.NewSource(1234))
	// A queued tip that no fetched block extends.
	orphanedTip := []eth.L2BlockRef{{Number: 105, Hash: common.Hash{0xee}}}

	t.Run("a reorged block resets the loader and clears state", func(t *testing.T) {
		et := newEnqueueTest(t)
		loader := &BlockLoader{batcher: et.bs}
		loader.queuedBlocks = orphanedTip
		loader.prevSyncStatus = &eth.SyncStatus{}
		block := derivetest.RandomL2BlockWithChainId(rng, 1, defaultTestRollupConfig.L2ChainID)
		et.ep.ethClient.ExpectBlockByNumber(new(big.Int).SetUint64(106), block, nil)
		localSafe := eth.L2BlockRef{Number: 104, Hash: common.Hash{0x04}, L1Origin: eth.BlockID{Number: 40}}
		et.ep.rollupClient.ExpectSyncStatus(&eth.SyncStatus{
			HeadL1:      eth.L1BlockRef{Number: 50, Hash: common.Hash{0xa1}},
			LocalSafeL2: localSafe,
		}, nil)

		loader.EnqueueBlocks(t.Context(), inclusiveBlockRange{106, 106})

		require.Nil(t, loader.queuedBlocks, "reorg must drop the queued chain")
		require.Nil(t, loader.prevSyncStatus, "reorg must drop the sync-status baseline")
		require.EqualValues(t, 1, et.bs.espressoClearEpoch, "reorg must clear the channel manager")
		require.Equal(t, []eth.L2BlockRef{localSafe}, et.streamer.positions,
			"reorg must re-anchor the streamer to the local-safe head")
		require.Empty(t, et.submitted, "the reorged block must not reach Espresso")
	})

	t.Run("a block extending the queue is signed, submitted to espresso, and appended", func(t *testing.T) {
		et := newEnqueueTest(t)
		block := derivetest.RandomL2BlockWithChainId(rng, 1, defaultTestRollupConfig.L2ChainID)
		loader := &BlockLoader{batcher: et.bs}
		loader.queuedBlocks = []eth.L2BlockRef{{Number: block.NumberU64() - 1, Hash: block.ParentHash()}}
		et.ep.ethClient.ExpectBlockByNumber(new(big.Int).SetUint64(106), block, nil)

		loader.EnqueueBlocks(t.Context(), inclusiveBlockRange{106, 106})

		require.Len(t, loader.queuedBlocks, 2)
		appended := loader.queuedBlocks[1]
		require.Equal(t, block.Hash(), appended.Hash)
		require.Equal(t, block.NumberU64(), appended.Number)
		require.Zero(t, et.bs.espressoClearEpoch)
		receiveTestValue(t, et.submitted)
	})

	t.Run("a fetch failure stops the scan without resetting", func(t *testing.T) {
		et := newEnqueueTest(t)
		loader := &BlockLoader{batcher: et.bs}
		loader.queuedBlocks = orphanedTip
		et.ep.ethClient.ExpectBlockByNumber(new(big.Int).SetUint64(106), (*types.Block)(nil), errors.New("rpc down"))

		// The range spans two blocks; the mock would reject a fetch of 107,
		// so completing without a mock failure also proves the scan stopped.
		loader.EnqueueBlocks(t.Context(), inclusiveBlockRange{106, 107})

		require.Equal(t, orphanedTip, loader.queuedBlocks,
			"a transient fetch failure must keep the queue for the next tick")
		require.Zero(t, et.bs.espressoClearEpoch)
		require.Empty(t, et.submitted)
	})
}
