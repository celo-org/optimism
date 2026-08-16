package batcher

import (
	"context"
	"errors"
	"math/big"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	derivetest "github.com/ethereum-optimism/optimism/op-node/rollup/derive/test"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	t.Run("zeroed LocalSafeL2 blocks the clear", func(t *testing.T) {
		l, ep := newReanchorSubmitter(t, true, &espressoStreamers.Streamer{})
		ep.rollupClient.ExpectSyncStatus(&eth.SyncStatus{
			HeadL1: eth.L1BlockRef{Number: 5, Hash: common.Hash{0x01}},
		}, nil)
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

	t.Run("queue caught up with the unsafe head retries", func(t *testing.T) {
		loader := newTestBlockLoader(t)
		loader.queuedBlocks = numberedRefs(105, 109)
		r, action := loader.nextBlockRange(newStatus(numberedRef(104), 109))
		require.Equal(t, EnqueueBlockAction(ActionRetry), action)
		require.Equal(t, inclusiveBlockRange{}, r)
		require.Equal(t, numberedRefs(105, 109), loader.queuedBlocks)
	})

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

	t.Run("derivation ahead of the whole queue resets", func(t *testing.T) {
		loader := newTestBlockLoader(t)
		loader.queuedBlocks = numberedRefs(105, 106)
		_, action := loader.nextBlockRange(newStatus(numberedRef(107), 110))
		require.Equal(t, EnqueueBlockAction(ActionReset), action)
	})

	t.Run("safe head below the oldest queued block resets", func(t *testing.T) {
		loader := newTestBlockLoader(t)
		loader.queuedBlocks = numberedRefs(105, 108)
		_, action := loader.nextBlockRange(newStatus(numberedRef(103), 110))
		require.Equal(t, EnqueueBlockAction(ActionReset), action)
	})

	t.Run("gapped queue with the safe head above its span resets", func(t *testing.T) {
		loader := newTestBlockLoader(t)
		// A queue with a hole (106 missing, e.g. after a partial enqueue
		// failure) makes the count-based check the only guard.
		loader.queuedBlocks = []eth.L2BlockRef{numberedRef(105), numberedRef(107)}
		_, action := loader.nextBlockRange(newStatus(numberedRef(107), 110))
		require.Equal(t, EnqueueBlockAction(ActionReset), action)
	})

	t.Run("safe-chain reorg is detected by hash mismatch", func(t *testing.T) {
		loader := newTestBlockLoader(t)
		loader.queuedBlocks = numberedRefs(105, 108)
		// Same height we queued at 106, different (canonical) hash: our
		// queued chain was orphaned and must not keep feeding Espresso.
		reorgedSafe := eth.L2BlockRef{Number: 106, Hash: common.Hash{0xde, 0xad}}
		_, action := loader.nextBlockRange(newStatus(reorgedSafe, 110))
		require.Equal(t, EnqueueBlockAction(ActionReset), action)
	})

	t.Run("matching safe hash prunes derived blocks and enqueues the rest", func(t *testing.T) {
		loader := newTestBlockLoader(t)
		loader.queuedBlocks = numberedRefs(105, 108)
		r, action := loader.nextBlockRange(newStatus(numberedRef(106), 110))
		require.Equal(t, EnqueueBlockAction(ActionEnqueue), action)
		require.Equal(t, inclusiveBlockRange{109, 110}, r)
		require.Equal(t, numberedRefs(106, 108), loader.queuedBlocks,
			"blocks strictly below the safe head are derived and must be dropped")
	})

	t.Run("safe head just below the queue enqueues without pruning", func(t *testing.T) {
		loader := newTestBlockLoader(t)
		loader.queuedBlocks = numberedRefs(105, 108)
		// Hash deliberately unrelated to the queue: with zero blocks to
		// enqueue-check there is nothing of ours at the safe height yet, so
		// no hash comparison (and no prune) may fire.
		safe := eth.L2BlockRef{Number: 104, Hash: common.Hash{0xbe, 0xef}}
		r, action := loader.nextBlockRange(newStatus(safe, 110))
		require.Equal(t, EnqueueBlockAction(ActionEnqueue), action)
		require.Equal(t, inclusiveBlockRange{109, 110}, r)
		require.Equal(t, numberedRefs(105, 108), loader.queuedBlocks)
	})
}

// nopEspressoClient satisfies the EspressoClient interface for wiring up an
// espressoTransactionSubmitter whose workers are never spawned; any actual use
// of the client would panic on the embedded nil interface.
type nopEspressoClient struct {
	espressoClient.EspressoClient
}

// fakeChainSigner returns a constant signature; EnqueueBlocks only needs
// signing to succeed, not to verify.
type fakeChainSigner struct{}

func (fakeChainSigner) SignTransaction(ctx context.Context, addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
	return tx, nil
}

func (fakeChainSigner) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	return make([]byte, 65), nil
}

// newEnqueueTestSubmitter builds a BatchSubmitter that can run EnqueueBlocks
// end to end: fetch a block from the (mock) sequencer, convert it, sign it,
// and hand it to the espresso transaction submitter's job queue.
func newEnqueueTestSubmitter(t *testing.T) (*BatchSubmitter, *mockL2EndpointProvider) {
	bs, ep := setup(t, nil)
	bs.Espresso.ChainSigner = fakeChainSigner{}
	bs.espressoSubmitter = NewEspressoTransactionSubmitter(
		WithContext(t.Context()),
		WithEspressoClient(nopEspressoClient{}),
	)
	return bs, ep
}

// TestEnqueueBlocksReorgDetection locks EnqueueBlocks' parent-hash check — the
// loader-side reorg defense — and the reset it triggers: the reset must be
// requested via the clearState handshake, never performed inline on the
// queueing loop (the data race the handshake was introduced to fix).
func TestEnqueueBlocksReorgDetection(t *testing.T) {
	rng := rand.New(rand.NewSource(1234))

	t.Run("a reorged block resets the loader and requests a state clear", func(t *testing.T) {
		bs, ep := newEnqueueTestSubmitter(t)
		loader := &BlockLoader{batcher: bs}
		loader.queuedBlocks = []eth.L2BlockRef{{Number: 105, Hash: common.Hash{0xee}}}
		loader.prevSyncStatus = &eth.SyncStatus{}

		block := derivetest.RandomL2BlockWithChainId(rng, 1, defaultTestRollupConfig.L2ChainID)
		require.NotEqual(t, loader.queuedBlocks[0].Hash, block.ParentHash(),
			"sanity: the fetched block must not extend the queued tip")
		ep.ethClient.ExpectBlockByNumber(new(big.Int).SetUint64(106), block, nil)

		loader.EnqueueBlocks(t.Context(), inclusiveBlockRange{106, 106})

		require.Nil(t, loader.queuedBlocks, "reorg must drop the queued chain")
		require.Nil(t, loader.prevSyncStatus, "reorg must drop the sync-status baseline")
		require.True(t, bs.clearStateRequested.Load(),
			"reorg must request a state clear from the loading loop")
		require.Empty(t, bs.espressoSubmitter.submitJobQueue,
			"the reorged block must not reach Espresso")
	})

	t.Run("a block extending the queue is signed, queued to espresso, and appended", func(t *testing.T) {
		bs, ep := newEnqueueTestSubmitter(t)
		block := derivetest.RandomL2BlockWithChainId(rng, 1, defaultTestRollupConfig.L2ChainID)
		loader := &BlockLoader{batcher: bs}
		loader.queuedBlocks = []eth.L2BlockRef{{Number: block.NumberU64() - 1, Hash: block.ParentHash()}}
		ep.ethClient.ExpectBlockByNumber(new(big.Int).SetUint64(106), block, nil)

		loader.EnqueueBlocks(t.Context(), inclusiveBlockRange{106, 106})

		require.Len(t, loader.queuedBlocks, 2)
		appended := loader.queuedBlocks[1]
		require.Equal(t, block.Hash(), appended.Hash)
		require.Equal(t, block.NumberU64(), appended.Number)
		require.False(t, bs.clearStateRequested.Load())
		require.Len(t, bs.espressoSubmitter.submitJobQueue, 1,
			"exactly one espresso submission job for the one new block")
	})

	t.Run("a fetch failure stops the scan without resetting", func(t *testing.T) {
		bs, ep := newEnqueueTestSubmitter(t)
		loader := &BlockLoader{batcher: bs}
		loader.queuedBlocks = []eth.L2BlockRef{{Number: 105, Hash: common.Hash{0xee}}}
		ep.ethClient.ExpectBlockByNumber(new(big.Int).SetUint64(106), (*types.Block)(nil), errors.New("rpc down"))

		// The range spans two blocks; the mock would reject a fetch of 107,
		// so completing without a mock failure also proves the scan stopped.
		loader.EnqueueBlocks(t.Context(), inclusiveBlockRange{106, 107})

		require.Equal(t, []eth.L2BlockRef{{Number: 105, Hash: common.Hash{0xee}}}, loader.queuedBlocks,
			"a transient fetch failure must keep the queue for the next tick")
		require.False(t, bs.clearStateRequested.Load())
	})
}

// TestClearStateHandshake locks the requestClearState/performClearState CAS
// handshake between the batch queueing loop (requester) and the batch loading
// loop (performer). The handshake exists so the queueing loop never runs
// clearState itself — doing so raced the loading loop on the channel manager
// and streamer (the earlier data-race fix); this is its regression test and
// must keep running under -race.
func TestClearStateHandshake(t *testing.T) {
	populatedStatus := &eth.SyncStatus{
		HeadL1:      eth.L1BlockRef{Number: 50, Hash: common.Hash{0xa1}},
		LocalSafeL2: eth.L2BlockRef{Number: 104, L1Origin: eth.BlockID{Number: 40}},
	}

	t.Run("perform without a request is a no-op", func(t *testing.T) {
		bs, _ := setup(t, nil)
		// No SyncStatus expectation is registered: if performClearState ran
		// clearState anyway, the mock would reject the unexpected call.
		require.False(t, bs.performClearState(t.Context()))
	})

	t.Run("a request is performed exactly once", func(t *testing.T) {
		bs, ep := setup(t, nil)
		ep.rollupClient.ExpectSyncStatus(populatedStatus, nil)

		bs.requestClearState()
		require.True(t, bs.performClearState(t.Context()), "the requested clear must be performed")
		require.False(t, bs.performClearState(t.Context()), "the request must be consumed by the CAS")
		require.False(t, bs.clearStateRequested.Load())
	})

	t.Run("concurrent requesters and one performer race cleanly", func(t *testing.T) {
		bs, ep := setup(t, nil)
		ep.rollupClient.Mock.On("SyncStatus").Return(populatedStatus, nil)

		const requesters = 4
		const requestsEach = 64

		var performed atomic.Int64
		stop := make(chan struct{})
		var performerWg sync.WaitGroup
		performerWg.Add(1)
		go func() {
			defer performerWg.Done()
			for {
				if bs.performClearState(t.Context()) {
					performed.Add(1)
				}
				select {
				case <-stop:
					return
				default:
				}
			}
		}()

		var requesterWg sync.WaitGroup
		for i := 0; i < requesters; i++ {
			requesterWg.Add(1)
			go func() {
				defer requesterWg.Done()
				for j := 0; j < requestsEach; j++ {
					bs.requestClearState()
				}
			}()
		}
		requesterWg.Wait()
		close(stop)
		performerWg.Wait()

		// The performer may have exited between the last request and its
		// final check; one drain settles it.
		if bs.performClearState(t.Context()) {
			performed.Add(1)
		}

		require.False(t, bs.clearStateRequested.Load(), "no request may be left dangling")
		require.GreaterOrEqual(t, performed.Load(), int64(1))
		require.LessOrEqual(t, performed.Load(), int64(requesters*requestsEach),
			"the CAS must never perform more clears than were requested")
	})
}
