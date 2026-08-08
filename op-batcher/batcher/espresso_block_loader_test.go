package batcher

import (
	"testing"

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
