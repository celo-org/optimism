package batcher

import (
	"errors"
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
		require.False(t, l.espressoSyncChannelManager(&populated))
		require.Equal(t, populated.CurrentL1, l.prevCurrentL1)
	})

	t.Run("zeroed LocalSafeL2 is out of sync", func(t *testing.T) {
		l := newTestSubmitter(t)
		status := populated
		status.LocalSafeL2 = eth.L2BlockRef{}
		require.True(t, l.espressoSyncChannelManager(&status))
		require.Zero(t, l.prevCurrentL1,
			"an out-of-sync status must not advance the CurrentL1 baseline")
	})

	t.Run("reversed CurrentL1 is out of sync", func(t *testing.T) {
		l := newTestSubmitter(t)
		l.prevCurrentL1 = eth.L1BlockRef{Number: 3, Hash: common.Hash{0x06}}
		require.True(t, l.espressoSyncChannelManager(&populated))
	})
}

// clearState must not clear the channel manager unless the streamer re-anchor
// target is already in hand: espressoReanchorTarget reporting ok=false makes
// the caller retry the whole clear instead of performing it partially, which
// would leave an emptied channel manager with the streamer at its old cursor.
func TestEspressoReanchorTarget(t *testing.T) {
	newReanchorSubmitter := func(t *testing.T, enabled bool, streamer *espressoStreamers.Streamer) (*BatchSubmitter, *mockL2EndpointProvider) {
		ep := newEndpointProvider()
		return &BatchSubmitter{
			DriverSetup: DriverSetup{
				Log: testlog.Logger(t, log.LevelDebug),
				Config: BatcherConfig{
					NetworkTimeout: time.Second,
					Espresso:       EspressoBatcherConfig{Enabled: enabled},
				},
				EndpointProvider: ep,
			},
			espressoStreamer: streamer,
		}, ep
	}

	t.Run("espresso disabled needs no target", func(t *testing.T) {
		l, _ := newReanchorSubmitter(t, false, nil)
		target, ok := l.espressoReanchorTarget(t.Context())
		require.True(t, ok)
		require.Nil(t, target)
	})

	t.Run("startup before streamer construction needs no target", func(t *testing.T) {
		l, _ := newReanchorSubmitter(t, true, nil)
		target, ok := l.espressoReanchorTarget(t.Context())
		require.True(t, ok)
		require.Nil(t, target)
	})

	t.Run("sync status failure blocks the clear", func(t *testing.T) {
		l, ep := newReanchorSubmitter(t, true, &espressoStreamers.Streamer{})
		ep.rollupClient.ExpectSyncStatus(nil, errors.New("transient RPC failure"))
		target, ok := l.espressoReanchorTarget(t.Context())
		require.False(t, ok)
		require.Nil(t, target)
	})

	t.Run("populated status yields the local-safe target", func(t *testing.T) {
		l, ep := newReanchorSubmitter(t, true, &espressoStreamers.Streamer{})
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
