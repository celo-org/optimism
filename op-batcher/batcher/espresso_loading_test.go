package batcher

// Espresso loading tests cover reading batches from the streamer and adding
// their blocks to the channel manager for publication to L1. They check batch
// draining, cursor advancement, reset exclusion, stale snapshots, Peek timeouts,
// and publication signals. See espresso_block_loader_test.go for range selection
// and reset recovery, and espresso_driver_test.go for session lifecycle and setup.

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/EspressoSystems/espresso-streamers/op/derivation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
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

// loadingBarrier pauses an operation at an observable boundary. Release is
// idempotent so cleanup can also release a goroutine after a failed assertion.
type loadingBarrier struct {
	entered     chan struct{}
	released    chan struct{}
	enterOnce   sync.Once
	releaseOnce sync.Once
}

func newLoadingBarrier() *loadingBarrier {
	return &loadingBarrier{entered: make(chan struct{}), released: make(chan struct{})}
}

func (b *loadingBarrier) wait() {
	b.enterOnce.Do(func() { close(b.entered) })
	<-b.released
}

func (b *loadingBarrier) release() {
	b.releaseOnce.Do(func() { close(b.released) })
}

// gatedLoadingStreamer records concurrent calls independently of the production
// lock, including when a regression lets a reset run during an in-flight batch.
// Gates run outside mu so the fake does not supply the exclusion being tested.
type gatedLoadingStreamer struct {
	espressoStreamer
	mu           sync.Mutex
	batch        *derivation.EspressoBatch
	peekCalls    int
	advanceCalls int
	positions    []eth.L2BlockRef
	operations   []string
	onPeek       func()
	onAdvance    func()
	onPosition   func()
}

func (s *gatedLoadingStreamer) Peek(context.Context) *derivation.EspressoBatch {
	s.mu.Lock()
	s.peekCalls++
	batch := s.batch
	s.mu.Unlock()
	if s.onPeek != nil {
		s.onPeek()
	}
	return batch
}

func (s *gatedLoadingStreamer) AdvancePosition() {
	if s.onAdvance != nil {
		s.onAdvance()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.advanceCalls++
	s.operations = append(s.operations, "advance")
	s.batch = nil
}

func (s *gatedLoadingStreamer) SetBatchPosition(position eth.L2BlockRef) {
	if s.onPosition != nil {
		s.onPosition()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.positions = append(s.positions, position)
	s.operations = append(s.operations, "re-anchor")
	s.batch = nil
}

func (s *gatedLoadingStreamer) setBatch(batch *derivation.EspressoBatch) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batch = batch
}

// Probe without starting a mutex waiter, which would prevent synctest quiescence.
func requireEspressoStreamerLocked(t *testing.T, batcher *BatchSubmitter) {
	t.Helper()
	available := batcher.espressoStreamerMutex.TryLock()
	if available {
		batcher.espressoStreamerMutex.Unlock()
	}
	require.False(t, available, "the operation must hold the streamer lock at this boundary")
}

func newEspressoLoadingTest(t *testing.T) (*BatchSubmitter, *loadingTestStreamer, *eth.SyncStatus) {
	t.Helper()
	batcher, endpoints, status := newEspressoLoadingTestBatcher(t)
	endpoints.rollupClient.Mock.On("SyncStatus").Return(status, nil)
	streamer := &loadingTestStreamer{}
	batcher.espressoSession = &espressoSession{streamer: streamer}
	return batcher, streamer, status
}

func newEspressoLoadingTestBatcher(t *testing.T) (*BatchSubmitter, *mockL2EndpointProvider, *eth.SyncStatus) {
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
	return batcher, endpoints, status
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
	require.True(t, batcher.espressoStreamerMutex.TryLock(), "the operation must release the streamer lock before returning")
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
	for _, scenario := range []string{"empty", "stale"} {
		t.Run(scenario, func(t *testing.T) {
			batcher, streamer, status := newEspressoLoadingTest(t)
			batch, _ := newEspressoLoadingBatch(t, batcher, 11, status.LocalSafeL2.Hash)
			if scenario == "stale" {
				batch.BatchHeader.Number = big.NewInt(9)
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
			require.Equal(t, 1, streamer.peekCalls)
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

func TestEspressoLoaderResetWaitsForInFlightBatch(t *testing.T) {
	batcher, endpoints, status := newEspressoLoadingTestBatcher(t)
	batch11, block11 := newEspressoLoadingBatch(t, batcher, 11, status.LocalSafeL2.Hash)
	peek := newLoadingBarrier()
	advance := newLoadingBarrier()
	streamer := &gatedLoadingStreamer{batch: batch11, onPeek: peek.wait, onAdvance: advance.wait}
	batcher.espressoSession = &espressoSession{streamer: streamer}
	resetStatus := *status
	resetStatus.LocalSafeL2 = eth.L2BlockRef{
		Number: 8, Hash: common.Hash{8}, L1Origin: status.LocalSafeL2.L1Origin,
	}
	endpoints.rollupClient.Mock.On("SyncStatus").Return(&resetStatus, nil).Once()
	loader := &BlockLoader{
		batcher:        batcher,
		queuedBlocks:   []eth.L2BlockRef{status.LocalSafeL2},
		prevSyncStatus: status,
	}

	loaded := make(chan struct{})
	var result espressoLoadResult
	t.Cleanup(func() { peek.release(); advance.release(); receiveTestValue(t, loaded) })
	go func() {
		defer close(loaded)
		result = batcher.espressoLoadBatch(t.Context(), status, 0)
	}()
	receiveTestValue(t, peek.entered)
	requireEspressoStreamerLocked(t, batcher)
	peek.release()
	receiveTestValue(t, advance.entered)
	// Only the loader is running here, so it must own the streamer lock until
	// the block already added to the manager has its matching cursor advance.
	requireEspressoStreamerLocked(t, batcher)
	require.Equal(t, block11.Hash(), batcher.channelMgr.tip)

	reset := make(chan struct{})
	resetStarted := make(chan struct{})
	t.Cleanup(func() { peek.release(); advance.release(); receiveTestValue(t, reset) })
	go func() {
		defer close(reset)
		close(resetStarted)
		loader.reset(t.Context())
	}()
	receiveTestValue(t, resetStarted)
	advance.release()
	receiveTestValue(t, loaded)
	receiveTestValue(t, reset)

	require.Equal(t, espressoLoadAdded, result)
	require.Equal(t, []string{"advance", "re-anchor"}, streamer.operations,
		"the complete batch must precede the reset's cursor change")
	require.Equal(t, 1, streamer.advanceCalls)
	require.Equal(t, []eth.L2BlockRef{resetStatus.LocalSafeL2}, streamer.positions)
	require.Empty(t, batcher.channelMgr.blocks)
	require.Equal(t, common.Hash{}, batcher.channelMgr.tip)
	require.EqualValues(t, 1, batcher.espressoClearEpoch)
	require.Nil(t, loader.queuedBlocks)
	require.Nil(t, loader.prevSyncStatus)
	requireEspressoStreamerUnlocked(t, batcher)

	// Continuing the old drain must not skip a canonical replacement below its
	// stale safe floor or overwrite the reset's cursor with the old safe head.
	replacement, block9 := newEspressoLoadingBatch(t, batcher, 9, resetStatus.LocalSafeL2.Hash)
	streamer.setBatch(replacement)
	require.Equal(t, espressoLoadStop, batcher.espressoLoadBatch(t.Context(), status, 0))
	require.Equal(t, 1, streamer.peekCalls)
	require.Equal(t, 1, streamer.advanceCalls)
	require.Equal(t, []eth.L2BlockRef{resetStatus.LocalSafeL2}, streamer.positions)
	require.Empty(t, batcher.channelMgr.blocks)

	// A fresh snapshot can load that same replacement normally.
	require.Equal(t, espressoLoadAdded, batcher.espressoLoadBatch(t.Context(), &resetStatus, 1))
	require.Equal(t, 2, streamer.peekCalls)
	require.Equal(t, 2, streamer.advanceCalls)
	require.Equal(t, []eth.L2BlockRef{resetStatus.LocalSafeL2}, streamer.positions)
	require.Len(t, batcher.channelMgr.blocks, 1)
	require.Equal(t, block9.Hash(), batcher.channelMgr.blocks[0].Hash())
	require.Equal(t, block9.Hash(), batcher.channelMgr.tip)
	requireEspressoStreamerUnlocked(t, batcher)
}

func TestEspressoPendingResetInvalidatesLoadingAndReconciliation(t *testing.T) {
	for _, scenario := range []string{"success", "cancellation"} {
		t.Run(scenario, func(t *testing.T) {
			batcher, endpoints, status := newEspressoLoadingTestBatcher(t)
			batch11, _ := newEspressoLoadingBatch(t, batcher, 11, status.LocalSafeL2.Hash)
			streamer := &gatedLoadingStreamer{batch: batch11}
			batcher.espressoSession = &espressoSession{streamer: streamer}
			resetCtx, cancelReset := context.WithCancel(t.Context())
			rpc := newLoadingBarrier()
			var rpcStatus *eth.SyncStatus = status
			var rpcErr error
			if scenario == "cancellation" {
				rpcStatus, rpcErr = nil, context.Canceled
			}
			endpoints.rollupClient.Mock.On("SyncStatus").Return(rpcStatus, rpcErr).Once().
				Run(func(mock.Arguments) { rpc.wait() })

			reset := make(chan struct{})
			t.Cleanup(func() { cancelReset(); rpc.release(); receiveTestValue(t, reset) })
			go func() { defer close(reset); batcher.clearState(resetCtx) }()
			receiveTestValue(t, rpc.entered)
			requireEspressoStreamerLocked(t, batcher)

			loaded := make(chan struct{})
			loadStarted := make(chan struct{})
			var result espressoLoadResult
			t.Cleanup(func() { cancelReset(); rpc.release(); receiveTestValue(t, loaded) })
			go func() {
				defer close(loaded)
				close(loadStarted)
				result = batcher.espressoLoadBatch(t.Context(), status, 0)
			}()
			reconciled := make(chan struct{})
			syncStarted := make(chan struct{})
			var skipDrain bool
			t.Cleanup(func() { cancelReset(); rpc.release(); receiveTestValue(t, reconciled) })
			go func() {
				defer close(reconciled)
				close(syncStarted)
				skipDrain = batcher.espressoSyncChannelManager(status, 0)
			}()
			receiveTestValue(t, loadStarted)
			receiveTestValue(t, syncStarted)
			if scenario == "cancellation" {
				cancelReset()
			}
			rpc.release()
			receiveTestValue(t, reset)
			receiveTestValue(t, loaded)
			receiveTestValue(t, reconciled)

			require.Equal(t, espressoLoadStop, result)
			require.True(t, skipDrain)
			require.Zero(t, streamer.peekCalls, "loading cannot inspect a candidate using the pre-reset snapshot")
			require.Zero(t, streamer.advanceCalls)
			if scenario == "cancellation" {
				require.Empty(t, streamer.positions, "a cancelled reset must not invent a canonical cursor")
				require.Equal(t, eth.BlockID{}, batcher.channelMgr.l1OriginLastSubmittedChannel)
			} else {
				require.Equal(t, []eth.L2BlockRef{status.LocalSafeL2}, streamer.positions)
				require.Equal(t, status.LocalSafeL2.L1Origin, batcher.channelMgr.l1OriginLastSubmittedChannel)
			}
			require.Zero(t, batcher.prevCurrentL1, "stale reconciliation must not update its L1 baseline")
			require.Empty(t, batcher.channelMgr.blocks)
			require.Equal(t, common.Hash{}, batcher.channelMgr.tip)
			require.EqualValues(t, 1, batcher.espressoClearEpoch)
			requireEspressoStreamerUnlocked(t, batcher)
		})
	}
}

func TestEspressoReconciliationHoldsStreamerLock(t *testing.T) {
	batcher, _, status := newEspressoLoadingTestBatcher(t)
	_, block12 := newEspressoLoadingBatch(t, batcher, 12, common.Hash{11})
	require.NoError(t, batcher.channelMgr.AddL2Block(block12))
	position := newLoadingBarrier()
	streamer := &gatedLoadingStreamer{onPosition: position.wait}
	batcher.espressoSession = &espressoSession{streamer: streamer}

	reconciled := make(chan struct{})
	var skipDrain bool
	t.Cleanup(func() { position.release(); receiveTestValue(t, reconciled) })
	go func() {
		defer close(reconciled)
		skipDrain = batcher.espressoSyncChannelManager(status, 0)
	}()
	receiveTestValue(t, position.entered)
	requireEspressoStreamerLocked(t, batcher)

	position.release()
	receiveTestValue(t, reconciled)

	require.False(t, skipDrain)
	require.Equal(t, []eth.L2BlockRef{status.LocalSafeL2}, streamer.positions)
	require.Empty(t, batcher.channelMgr.blocks)
	require.Zero(t, batcher.espressoClearEpoch,
		"reconciliation's own clear does not invalidate its current snapshot")
	requireEspressoStreamerUnlocked(t, batcher)
}

func TestEspressoLoadingLoopInvalidatesSnapshotFetchedDuringReset(t *testing.T) {
	batcher, endpoints, status := newEspressoLoadingTestBatcher(t)
	oldStatus := *status
	oldStatus.LocalSafeL2 = eth.L2BlockRef{
		Number: 12, Hash: common.Hash{0xff}, L1Origin: status.LocalSafeL2.L1Origin,
	}
	replacement, _ := newEspressoLoadingBatch(t, batcher, 11, status.LocalSafeL2.Hash)
	streamer := &gatedLoadingStreamer{}
	batcher.espressoSession = &espressoSession{streamer: streamer}
	ctx, cancel := context.WithCancel(t.Context())
	firstStatus := newLoadingBarrier()
	endpoints.rollupClient.Mock.On("SyncStatus").Return(&oldStatus, nil).Once().Run(func(mock.Arguments) { firstStatus.wait() })
	endpoints.rollupClient.Mock.On("SyncStatus").Return(status, nil).Once()
	// Hold later ticks until cancellation so they cannot obscure what happened
	// to the snapshot that was already in flight when the reset landed.
	endpoints.rollupClient.Mock.On("SyncStatus").Return((*eth.SyncStatus)(nil), context.Canceled).
		Run(func(mock.Arguments) { <-ctx.Done() })

	var wg sync.WaitGroup
	wg.Add(1)
	published := make(chan pubInfo, 1)
	done := make(chan struct{})
	t.Cleanup(func() { firstStatus.release(); cancel(); receiveTestValue(t, done) })
	go batcher.espressoBatchLoadingLoop(ctx, &wg, published, make(chan int64, 1))
	go func() { wg.Wait(); close(done) }()
	receiveTestValue(t, firstStatus.entered)

	batcher.clearState(t.Context())
	streamer.setBatch(replacement)
	firstStatus.release()
	receiveTestValue(t, published)
	cancel()
	receiveTestValue(t, done)

	require.Zero(t, streamer.peekCalls, "a reset during SyncStatus must invalidate that tick before draining")
	require.Zero(t, streamer.advanceCalls)
	require.Equal(t, []eth.L2BlockRef{status.LocalSafeL2}, streamer.positions,
		"the stale safe head must not replace the reset's canonical cursor")
	require.Empty(t, batcher.channelMgr.blocks)
	require.EqualValues(t, 1, batcher.espressoClearEpoch)
	requireEspressoStreamerUnlocked(t, batcher)
}
