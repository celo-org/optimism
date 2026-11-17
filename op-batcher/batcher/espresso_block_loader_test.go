package batcher

// These tests cover block-range selection, sync reconciliation, and safe-head
// re-anchoring during resets. Draining is covered in espresso_loading_test.go;
// session lifecycle and setup are covered in espresso_driver_test.go.

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common"
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
