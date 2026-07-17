package txmgr

import (
	"context"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/txmgr/metrics"
)

// pairTestHarness runs a real SimpleTxManager against the mock backend and
// records every broadcast so tests can assert on cancellation no-ops and nonce
// hygiene — the failure modes SendPairAsync exists to contain.
type pairTestHarness struct {
	mgr     *SimpleTxManager
	backend *mockBackendWithNonce

	mu         sync.Mutex
	broadcasts []*types.Transaction
}

// pairLegBehavior tells the harness's tx sender what to do with a leg,
// identified by its first data byte.
type pairLegBehavior struct {
	sendErr    error // returned from every SendTransaction call for this leg
	mine       bool  // mine the tx (at its current fees) as soon as it is sent
	mineStatus uint64
}

func newPairTestHarness(t *testing.T, behaviors map[byte]pairLegBehavior) *pairTestHarness {
	backend := newMockBackendWithNonce(newGasPricer(1))
	conf := configWithNumConfs(1)
	conf.Backend = backend

	mgr, err := NewSimpleTxManagerFromConfig("TEST", testlog.Logger(t, log.LevelCrit), &metrics.NoopTxMetrics{}, conf)
	require.NoError(t, err)

	h := &pairTestHarness{mgr: mgr, backend: backend}
	backend.setTxSender(func(ctx context.Context, tx *types.Transaction) error {
		h.mu.Lock()
		h.broadcasts = append(h.broadcasts, tx)
		h.mu.Unlock()

		// Cancellation no-ops (empty-data self-sends crafted by
		// cancelPairNonce) are always accepted and mined immediately.
		if len(tx.Data()) == 0 {
			hash := tx.Hash()
			backend.mine(&hash, tx.GasFeeCap(), nil)
			return nil
		}
		b, ok := behaviors[tx.Data()[0]]
		require.True(t, ok, "unexpected tx broadcast with marker %d", tx.Data()[0])
		if b.sendErr != nil {
			return b.sendErr
		}
		if b.mine {
			hash := tx.Hash()
			backend.mineWithStatus(&hash, tx.GasFeeCap(), nil, b.mineStatus)
		}
		return nil
	})
	return h
}

// noopsAt returns the nonces of broadcast cancellation no-ops (empty-data
// self-sends), sorted.
func (h *pairTestHarness) noopsAt() []uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	var nonces []uint64
	for _, tx := range h.broadcasts {
		if len(tx.Data()) == 0 && tx.To() != nil && *tx.To() == h.mgr.cfg.From {
			nonces = append(nonces, tx.Nonce())
		}
	}
	slices.Sort(nonces)
	return slices.Compact(nonces)
}

func pairCandidate(marker byte) TxCandidate {
	return TxCandidate{TxData: []byte{marker}, To: &common.Address{0xbb}}
}

func sendPair(mgr *SimpleTxManager) (SendResponse, SendResponse) {
	ch1 := make(chan SendResponse, 1)
	ch2 := make(chan SendResponse, 1)
	mgr.SendPairAsync(context.Background(), pairCandidate(1), pairCandidate(2), ch1, ch2)
	return <-ch1, <-ch2
}

// mockBackendWithNonce initializes the manager's nonce from len(minedTxs), so
// the first pair takes nonces 0 and 1 and a follow-up send takes 2.
const pairBaseNonce = uint64(0)

// TestSendPair_Success: both legs mine; consecutive nonces in argument order;
// no cancellation no-ops are broadcast.
func TestSendPair_Success(t *testing.T) {
	h := newPairTestHarness(t, map[byte]pairLegBehavior{
		1: {mine: true, mineStatus: types.ReceiptStatusSuccessful},
		2: {mine: true, mineStatus: types.ReceiptStatusSuccessful},
	})

	r1, r2 := sendPair(h.mgr)

	require.NoError(t, r1.Err)
	require.NoError(t, r2.Err)
	require.Equal(t, pairBaseNonce, r1.Nonce)
	require.Equal(t, pairBaseNonce+1, r2.Nonce, "pair legs must take consecutive nonces in argument order")
	require.Empty(t, h.noopsAt(), "a healthy pair must not broadcast cancellation no-ops")
}

// TestSendPair_FirstLegSendFailure is the nonce-gap hazard the pair exists to
// contain: the auth-style first leg fails permanently (a critical nonce error
// — txmgr treats unknown RPC errors as transient and retries them forever, so
// permanent failures are SendState critical errors, timeouts, or
// cancellations) after the second leg was already broadcast at the next
// nonce. The pair must (a) cancel the orphaned second leg with a no-op at its
// exact nonce, (b) fill the first leg's own nonce hole the same way, and (c)
// leave the manager unwedged: the next send gets the next fresh nonce with no
// reset and confirms normally.
func TestSendPair_FirstLegSendFailure(t *testing.T) {
	h := newPairTestHarness(t, map[byte]pairLegBehavior{
		1: {sendErr: core.ErrNonceTooLow},
		2: {},                                                      // accepted into the pool, never mined (gapped behind leg 1)
		3: {mine: true, mineStatus: types.ReceiptStatusSuccessful}, // follow-up send
	})

	r1, r2 := sendPair(h.mgr)

	require.Error(t, r1.Err)
	require.ErrorIs(t, r2.Err, ErrPairLegCancelled)
	require.Equal(t, []uint64{pairBaseNonce, pairBaseNonce + 1}, h.noopsAt(),
		"both pair nonces must be consumed by cancellation no-ops")

	// No wedge and no nonce reset: the next send continues the sequence at the
	// nonce after the repaired pair and confirms.
	ch := make(chan SendResponse, 1)
	h.mgr.SendAsync(context.Background(), pairCandidate(3), ch)
	r3 := <-ch
	require.NoError(t, r3.Err)
	require.Equal(t, pairBaseNonce+2, r3.Nonce)
}

// TestSendPair_FirstLegReverted: a mined-but-reverted first leg also fails the
// pair (its event never fired), so the second leg is cancelled — but the first
// leg's nonce was consumed on chain, so only the second leg's nonce gets a
// no-op. The first leg's response keeps SendAsync semantics: a receipt with
// failed status and no error.
func TestSendPair_FirstLegReverted(t *testing.T) {
	h := newPairTestHarness(t, map[byte]pairLegBehavior{
		1: {mine: true, mineStatus: types.ReceiptStatusFailed},
		2: {},                                                      // accepted, never mined; must be cancelled
		3: {mine: true, mineStatus: types.ReceiptStatusSuccessful}, // follow-up send
	})

	r1, r2 := sendPair(h.mgr)

	require.NoError(t, r1.Err, "a mined-but-reverted leg is a receipt, not a send error")
	require.NotNil(t, r1.Receipt)
	require.Equal(t, types.ReceiptStatusFailed, r1.Receipt.Status)
	require.ErrorIs(t, r2.Err, ErrPairLegCancelled)
	require.Equal(t, []uint64{pairBaseNonce + 1}, h.noopsAt(),
		"only the unmined second leg's nonce needs a cancellation no-op")

	ch := make(chan SendResponse, 1)
	h.mgr.SendAsync(context.Background(), pairCandidate(3), ch)
	r3 := <-ch
	require.NoError(t, r3.Err)
	require.Equal(t, pairBaseNonce+2, r3.Nonce)
}

// TestSendPair_ParentContextCanceled: a pair canceled via its parent context
// (shutdown, or a sibling failing the queue's error group) must not broadcast
// no-ops, but must reset the cached nonce so the next send refills the gap.
func TestSendPair_ParentContextCanceled(t *testing.T) {
	h := newPairTestHarness(t, map[byte]pairLegBehavior{
		1: {}, // broadcast accepted, never mines
		2: {},
	})

	ctx, cancel := context.WithCancel(context.Background())
	ch1 := make(chan SendResponse, 1)
	ch2 := make(chan SendResponse, 1)
	h.mgr.SendPairAsync(ctx, pairCandidate(1), pairCandidate(2), ch1, ch2)
	cancel()

	r1, r2 := <-ch1, <-ch2
	require.Error(t, r1.Err)
	require.Error(t, r2.Err)

	require.Empty(t, h.noopsAt(), "no cancellation no-ops may be broadcast from a canceled parent context")

	h.mgr.nonceLock.RLock()
	nonce := h.mgr.nonce
	h.mgr.nonceLock.RUnlock()
	require.Nil(t, nonce, "nonce must be reset so the next send re-queries the chain and refills the gap")
}

// TestQueue_SendPair_Pipelines proves a pair holds exactly one maxPending
// slot, in both directions. At most one: with maxPending=2, two pairs (four
// txs) are all broadcast before anything confirms — per-tx slot accounting or
// per-pair serialization would block this. At least one: a third pair is not
// admitted (SendPair stays blocked, no nonces assigned) until a first pair
// confirms and frees its slot.
func TestQueue_SendPair_Pipelines(t *testing.T) {
	const maxPending = 2 // queue slots
	const numPairs = 2   // one slot each; 2*numPairs txs in flight at once
	const numTxs = 2 * numPairs

	wg := sync.WaitGroup{}
	backend := newMockBackendWithConfirmationDelay(newGasPricer(3), &wg)
	conf := configWithNumConfs(1)
	conf.Backend = backend
	// Keep resubmission fee-bumps out of the test's timing windows: a bumped
	// tx is a new hash, which the delay backend would count against the
	// WaitGroup unexpectedly.
	conf.RebroadcastInterval.Store(int64(time.Minute))
	conf.ResubmissionTimeout.Store(int64(time.Minute))
	mgr, err := NewSimpleTxManagerFromConfig("TEST", testlog.Logger(t, log.LevelCrit), &metrics.NoopTxMetrics{}, conf)
	require.NoError(t, err)

	q := NewQueue[int](context.Background(), mgr, maxPending)

	emptyCandidate := TxCandidate{To: &common.Address{}}

	receiptChs := make([]chan TxReceipt[int], numTxs)
	wg.Add(numTxs)
	for pair := range numPairs {
		id := 2 * pair
		receiptChs[id] = make(chan TxReceipt[int], 1)
		receiptChs[id+1] = make(chan TxReceipt[int], 1)
		q.SendPair(
			id, emptyCandidate, receiptChs[id],
			id+1, emptyCandidate, receiptChs[id+1],
		)
	}

	// cachedNonces returns the sorted nonces of every tx the backend has seen;
	// nonceSeq builds the expected contiguous sequence from startingNonce.
	cachedNonces := func() []uint64 {
		nonces := make([]uint64, 0, len(backend.cachedTxs))
		for _, tx := range backend.cachedTxs {
			nonces = append(nonces, tx.Nonce())
		}
		slices.Sort(nonces)
		return nonces
	}
	nonceSeq := func(n int) []uint64 {
		seq := make([]uint64, n)
		for i := range seq {
			seq[i] = startingNonce + uint64(i)
		}
		return seq
	}

	// All four txs reach the backend while nothing has confirmed yet: the two
	// pairs are genuinely in flight concurrently.
	wg.Wait()
	require.Equal(t, nonceSeq(numTxs), cachedNonces())

	// Both slots are held, so a third pair must block in SendPair without
	// being admitted (and without nonces assigned — they would land at +4/+5
	// only after admission).
	var thirdAdmitted atomic.Bool
	thirdCh1 := make(chan TxReceipt[int], 1)
	thirdCh2 := make(chan TxReceipt[int], 1)
	wg.Add(2)
	go func() {
		q.SendPair(
			numTxs, emptyCandidate, thirdCh1,
			numTxs+1, emptyCandidate, thirdCh2,
		)
		thirdAdmitted.Store(true)
	}()
	require.Never(t, thirdAdmitted.Load, 1*time.Second, 10*time.Millisecond,
		"a third pair was admitted while both slots were held: a pair must cost one slot")

	// Confirming the first two pairs frees their slots; the third pair is then
	// admitted and broadcasts at the next nonces.
	backend.MineAll()
	require.Eventually(t, thirdAdmitted.Load, 10*time.Second, 10*time.Millisecond,
		"third pair was not admitted after slots freed")
	wg.Wait() // third pair's txs reach the backend
	backend.MineAll()
	require.NoError(t, q.Wait())

	for i, ch := range append(receiptChs, thirdCh1, thirdCh2) {
		select {
		case r := <-ch:
			require.NoError(t, r.Err, "receipt %d", i)
		case <-time.After(10 * time.Second):
			t.Fatalf("timed out waiting for receipt %d", i)
		}
	}

	// The third pair's nonces continue the sequence, assigned only after
	// admission.
	require.Equal(t, nonceSeq(numTxs+2), cachedNonces())
}
