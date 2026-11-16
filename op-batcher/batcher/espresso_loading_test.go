package batcher

// Espresso loading tests cover reading batches from the streamer and adding
// their blocks to the channel manager for publication to L1. They check batch
// draining, cursor advancement, lock release, Peek timeouts, and publication
// signals. See espresso_block_loader_test.go for range selection and reset targets,
// and espresso_driver_test.go for session lifecycle and streamer setup.

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// loadingTestStreamer leaves the current batch in place until it is advanced.
type loadingTestStreamer struct {
	espressoStreamer
	batches      []*derivation.EspressoBatch
	peekCalls    int
	advanceCalls int
	positions    []eth.L2BlockRef
}

func (s *loadingTestStreamer) Peek(context.Context) *derivation.EspressoBatch {
	s.peekCalls++
	if s.advanceCalls >= len(s.batches) {
		return nil
	}
	return s.batches[s.advanceCalls]
}

func (s *loadingTestStreamer) AdvancePosition() {
	s.advanceCalls++
}

func (s *loadingTestStreamer) SetBatchPosition(position eth.L2BlockRef) {
	s.positions = append(s.positions, position)
	// Re-anchoring invalidates the scripted candidate; a real streamer would
	// fetch again from its new position.
	s.batches = nil
}

func newEspressoLoadingTest(t *testing.T) (*BatchSubmitter, *loadingTestStreamer, *eth.SyncStatus) {
	t.Helper()
	batcher, endpoints := setup(t, nil)
	batcher.Config.NetworkTimeout = time.Second
	batcher.Config.Espresso.PollInterval = time.Millisecond
	status := &eth.SyncStatus{
		HeadL1:    eth.L1BlockRef{Number: genesisL1Origin},
		CurrentL1: eth.L1BlockRef{Number: genesisL1Origin},
		LocalSafeL2: eth.L2BlockRef{
			Number:   10,
			Hash:     common.Hash{10},
			L1Origin: batcher.RollupConfig.Genesis.L1,
		},
		UnsafeL2: eth.L2BlockRef{Number: 12, Hash: common.Hash{12}},
	}
	endpoints.rollupClient.Mock.On("SyncStatus").Return(status, nil)
	streamer := &loadingTestStreamer{}
	batcher.espressoSession = &espressoSession{streamer: streamer}
	return batcher, streamer, status
}

func newEspressoLoadingBatch(t *testing.T, batcher *BatchSubmitter, number int64, parent common.Hash) (*derivation.EspressoBatch, *types.Block) {
	t.Helper()
	block := newMiniL2BlockWithChainIDAndNumberParent(0, batcher.RollupConfig.L2ChainID, big.NewInt(number), parent)
	batch, err := derivation.BlockToEspressoBatch(batcher.RollupConfig, block)
	require.NoError(t, err)
	return batch, block
}

func requireEspressoStreamerUnlocked(t *testing.T, batcher *BatchSubmitter) {
	t.Helper()
	require.True(t, batcher.espressoStreamerMutex.TryLock(), "loading a batch must release the streamer lock before returning")
	batcher.espressoStreamerMutex.Unlock()
}

func runEspressoLoadingUntilPublish(t *testing.T, batcher *BatchSubmitter) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	var wg sync.WaitGroup
	wg.Add(1)
	published := make(chan pubInfo, 1)
	go batcher.espressoBatchLoadingLoop(ctx, &wg, published, make(chan int64, 1))
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	t.Cleanup(func() { cancel(); receiveTestValue(t, done) })

	receiveTestValue(t, published)
	cancel()
	receiveTestValue(t, done)
}

// stalledPeekStreamer waits for the caller to cancel or time out each Peek.
type stalledPeekStreamer struct {
	espressoStreamer
	calls chan context.Context
}

func (s *stalledPeekStreamer) Peek(ctx context.Context) *derivation.EspressoBatch {
	s.calls <- ctx
	<-ctx.Done()
	return nil
}
func (*stalledPeekStreamer) SetBatchPosition(eth.L2BlockRef) {}

func TestEspressoLoadingLoopBoundsPeekAndRetries(t *testing.T) {
	batcher, endpoints := setup(t, nil)
	batcher.Config.NetworkTimeout = 20 * time.Millisecond
	batcher.Config.Espresso.PollInterval = time.Millisecond
	endpoints.rollupClient.Mock.On("SyncStatus").Return(&eth.SyncStatus{
		HeadL1:      eth.L1BlockRef{Number: genesisL1Origin},
		CurrentL1:   eth.L1BlockRef{Number: genesisL1Origin},
		LocalSafeL2: eth.L2BlockRef{Number: 1},
		SafeL2:      eth.L2BlockRef{Number: 1},
		UnsafeL2:    eth.L2BlockRef{Number: 1},
	}, nil)
	streamer := &stalledPeekStreamer{calls: make(chan context.Context, 4)}
	batcher.espressoSession = &espressoSession{streamer: streamer}
	ctx, cancel := context.WithCancel(t.Context())
	var wg sync.WaitGroup
	wg.Add(1)
	started := time.Now()
	go batcher.espressoBatchLoadingLoop(ctx, &wg, make(chan pubInfo, 1), make(chan int64, 1))
	t.Cleanup(func() { cancel(); wg.Wait() })
	var previous context.Context
	for range 2 {
		callCtx := receiveTestValue(t, streamer.calls)
		deadline, ok := callCtx.Deadline()
		require.True(t, ok)
		if previous == nil {
			require.False(t, deadline.Before(started.Add(batcher.Config.NetworkTimeout)), "Peek must receive the full configured budget")
		} else {
			previousDeadline, _ := previous.Deadline()
			require.True(t, deadline.After(previousDeadline), "each Peek needs a fresh deadline")
		}
		require.LessOrEqual(t, time.Until(deadline), batcher.Config.NetworkTimeout, "deadline must not exceed the configured timeout")
		if previous != nil {
			require.ErrorIs(t, previous.Err(), context.DeadlineExceeded)
		}
		require.NoError(t, ctx.Err())
		previous = callCtx
	}
	cancel()
}

func TestEspressoLoadingLoopDrainsConsecutiveBatches(t *testing.T) {
	batcher, streamer, status := newEspressoLoadingTest(t)
	batch11, block11 := newEspressoLoadingBatch(t, batcher, 11, status.LocalSafeL2.Hash)
	batch12, block12 := newEspressoLoadingBatch(t, batcher, 12, block11.Hash())
	streamer.batches = []*derivation.EspressoBatch{batch11, batch12}

	runEspressoLoadingUntilPublish(t, batcher)

	require.Equal(t, 2, streamer.advanceCalls)
	require.Empty(t, streamer.positions)
	require.Len(t, batcher.channelMgr.blocks, 2)
	require.Equal(t, block11.Hash(), batcher.channelMgr.blocks[0].Hash())
	require.Equal(t, block12.Hash(), batcher.channelMgr.blocks[1].Hash())
	require.Equal(t, block12.Hash(), batcher.channelMgr.tip)
	requireEspressoStreamerUnlocked(t, batcher)
}

func TestEspressoLoadingLoopPublishesWhileOutOfSync(t *testing.T) {
	for _, scenario := range []string{"zero local-safe head", "reversed CurrentL1"} {
		t.Run(scenario, func(t *testing.T) {
			batcher, streamer, status := newEspressoLoadingTest(t)
			_, block11 := newEspressoLoadingBatch(t, batcher, 11, status.LocalSafeL2.Hash)
			require.NoError(t, batcher.channelMgr.AddL2Block(block11))
			batch12, _ := newEspressoLoadingBatch(t, batcher, 12, block11.Hash())
			streamer.batches = []*derivation.EspressoBatch{batch12}
			previousL1 := status.CurrentL1
			if scenario == "zero local-safe head" {
				status.LocalSafeL2 = eth.L2BlockRef{}
			} else {
				previousL1.Number++
			}
			batcher.prevCurrentL1 = previousL1

			runEspressoLoadingUntilPublish(t, batcher)

			require.Zero(t, streamer.peekCalls, "out-of-sync ticks must publish existing work without loading")
			require.Zero(t, streamer.advanceCalls)
			require.Empty(t, streamer.positions)
			require.Len(t, batcher.channelMgr.blocks, 1)
			require.Equal(t, block11.Hash(), batcher.channelMgr.blocks[0].Hash())
			require.Equal(t, block11.Hash(), batcher.channelMgr.tip)
			require.Equal(t, previousL1, batcher.prevCurrentL1)
			require.Zero(t, batcher.espressoClearEpoch)
		})
	}
}

func TestEspressoLoadBatchStopsWithoutAdvancing(t *testing.T) {
	for _, scenario := range []string{"empty", "stale", "changed epoch"} {
		t.Run(scenario, func(t *testing.T) {
			batcher, streamer, status := newEspressoLoadingTest(t)
			batch, _ := newEspressoLoadingBatch(t, batcher, 11, status.LocalSafeL2.Hash)
			switch scenario {
			case "stale":
				batch.BatchHeader.Number = big.NewInt(9)
				streamer.batches = []*derivation.EspressoBatch{batch}
			case "changed epoch":
				batcher.espressoClearEpoch = 1
				streamer.batches = []*derivation.EspressoBatch{batch}
			}

			require.Equal(t, espressoLoadStop, batcher.espressoLoadBatch(t.Context(), status, 0))
			requireEspressoStreamerUnlocked(t, batcher)
			require.Zero(t, streamer.advanceCalls)
			require.Empty(t, batcher.channelMgr.blocks)
			require.Equal(t, common.Hash{}, batcher.channelMgr.tip)
			if scenario == "stale" {
				require.Equal(t, []eth.L2BlockRef{status.LocalSafeL2}, streamer.positions,
					"a stale batch must re-anchor to the canonical safe reference")
			} else {
				require.Empty(t, streamer.positions)
			}
			if scenario == "changed epoch" {
				require.Zero(t, streamer.peekCalls, "an invalidated snapshot must not reach the streamer")
				require.Equal(t, espressoLoadAdded, batcher.espressoLoadBatch(t.Context(), status, 1))
				requireEspressoStreamerUnlocked(t, batcher)
				require.Equal(t, 1, streamer.advanceCalls, "a refreshed snapshot must allow loading to resume")
			} else {
				require.Equal(t, 1, streamer.peekCalls)
			}
		})
	}
}

func TestEspressoLoadingLoopClearsRejectedBatches(t *testing.T) {
	for _, scenario := range []string{"malformed batch", "channel manager reorg"} {
		t.Run(scenario, func(t *testing.T) {
			batcher, streamer, status := newEspressoLoadingTest(t)
			_, block11 := newEspressoLoadingBatch(t, batcher, 11, status.LocalSafeL2.Hash)
			require.NoError(t, batcher.channelMgr.AddL2Block(block11))
			parent := block11.Hash()
			if scenario == "channel manager reorg" {
				parent = common.Hash{0xff}
			}
			batch12, _ := newEspressoLoadingBatch(t, batcher, 12, parent)
			if scenario == "malformed batch" {
				batch12.L1InfoDeposit = nil
			}
			streamer.batches = []*derivation.EspressoBatch{batch12}

			runEspressoLoadingUntilPublish(t, batcher)

			requireEspressoStreamerUnlocked(t, batcher)
			require.Zero(t, streamer.advanceCalls, "a rejected block must not advance the cursor")
			require.Equal(t, []eth.L2BlockRef{status.LocalSafeL2}, streamer.positions,
				"the loading loop must execute the requested clear")
			require.EqualValues(t, 1, batcher.espressoClearEpoch)
			require.Empty(t, batcher.channelMgr.blocks)
			require.Equal(t, common.Hash{}, batcher.channelMgr.tip)
		})
	}
}

func TestEspressoLoadBatchResetReleasesStreamer(t *testing.T) {
	for _, scenario := range []string{"malformed batch", "channel manager reorg"} {
		t.Run(scenario, func(t *testing.T) {
			batcher, streamer, status := newEspressoLoadingTest(t)
			_, block11 := newEspressoLoadingBatch(t, batcher, 11, status.LocalSafeL2.Hash)
			require.NoError(t, batcher.channelMgr.AddL2Block(block11))
			parent := block11.Hash()
			if scenario == "channel manager reorg" {
				parent = common.Hash{0xff}
			}
			batch12, _ := newEspressoLoadingBatch(t, batcher, 12, parent)
			if scenario == "malformed batch" {
				batch12.L1InfoDeposit = nil
			}
			streamer.batches = []*derivation.EspressoBatch{batch12}

			require.Equal(t, espressoLoadReset, batcher.espressoLoadBatch(t.Context(), status, 0))
			requireEspressoStreamerUnlocked(t, batcher)
			require.Zero(t, streamer.advanceCalls, "a rejected block must not advance the cursor")
			require.Empty(t, streamer.positions)
			require.Len(t, batcher.channelMgr.blocks, 1, "the caller must perform the clear after the helper unlocks")
			require.Equal(t, block11.Hash(), batcher.channelMgr.tip)
		})
	}
}
