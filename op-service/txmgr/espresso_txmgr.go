package txmgr

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-service/retry"
)

// ErrPairLegCancelled is delivered as the second leg's response error when the
// first leg of a SendPairAsync pair fails permanently and the second leg is
// cancelled as a result.
var ErrPairLegCancelled = errors.New("paired transaction cancelled: first leg failed")

// errPairBlobsUnsupported is returned when a pair candidate carries blobs;
// see the blob paragraph in SendPairAsync's doc for why.
var errPairBlobsUnsupported = errors.New("SendPairAsync does not support blob candidates")

func getContext(ctx context.Context, txTimeout time.Duration) (context.Context, context.CancelFunc) {
	if txTimeout == 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, txTimeout)
}

// sendPair drives the pair to resolution on its own goroutine: it broadcasts
// both legs concurrently, cancels the second leg if the first fails permanently,
// and delivers exactly one response per channel. A second leg cancelled this way
// has its error wrapped in ErrPairLegCancelled so callers can tell it did not
// fail on its own.
//
// Before any response is delivered, the pair repairs the nonce of each leg that
// resolved with an error (repairPair): the nonce is consumed on chain by a
// fee-bumped no-op, and sendPair blocks until that settles. Why: an errored leg
// is in an unknown state — possibly never broadcast, possibly in the pool at
// heavily bumped fees, possibly even mined unseen. Stock recovery (resetNonce,
// then re-signing the same nonces) recovers by collision with those orphans,
// which is safe for self-contained txs (a stale winner is at worst a duplicate)
// but not for pairs: a failure makes the batcher rewind and re-pack, so resent
// legs carry a different commitment than the orphans they contest. A stale leg
// winning one nonce while a fresh leg wins the adjacent one mines a torn pair —
// auth and batch with mismatched commitments, ignored by derivation — and the
// recovery round has itself half-failed, leaving fresh orphans behind.
//
// Repairing before responding keeps the failure self-contained: the error
// cannot cancel the queue's error group, and so cannot trigger resubmission,
// until this pair's nonces are already consumed — the resubmission's nonce
// query starts above everything the pair left behind.
//
// Aborted siblings get no such guarantee. A pair cancelled by the group abort
// (parent context already dead) only resets the nonce — see repairPair's
// parent-context branch — leaving its legs wherever cancellation caught them:
// That spoilage is self-limiting: a stale leg can only cause damage by
// contesting a slot, and a contested slot that fails a round becomes an
// errored leg of that round's pair — so repair settles the very slots the
// damage occurred at, and each abort's leftovers are cleaned up by the first
// round they hurt. It is accepted because settling every aborted sibling
// eagerly would stall the queue's drain barrier a block per abort, scaling
// with the pipeline depth.
func (m *SimpleTxManager) sendPair(parentCtx context.Context, tx1, tx2 *types.Transaction, ch1, ch2 chan SendResponse) {
	defer func() { m.metr.RecordPendingTx(m.pending.Add(-2)) }()
	ctx1, cancel1 := getContext(parentCtx, m.cfg.TxSendTimeout)
	ctx2, cancel2 := getContext(parentCtx, m.cfg.TxSendTimeout)
	defer cancel1()
	defer cancel2()

	res2 := make(chan SendResponse, 1)
	go func() {
		receipt, err := m.sendTx(ctx2, tx2)
		res2 <- SendResponse{Receipt: receipt, Nonce: tx2.Nonce(), Err: err}
	}()

	receipt1, err1 := m.sendTx(ctx1, tx1)

	// A mined-but-reverted first leg also fails the pair, not just a failed
	// send: the second leg depends on the first leg's successful execution.
	leg1Failed := err1 != nil || (receipt1 != nil && receipt1.Status != types.ReceiptStatusSuccessful)
	if leg1Failed {
		m.l.Warn("pair first leg failed; cancelling second leg",
			"firstNonce", tx1.Nonce(), "secondNonce", tx2.Nonce(),
			"err", err1, "reverted", err1 == nil)
		cancel2()
	}

	resp2 := <-res2

	m.repairPair(parentCtx, tx1, err1, tx2, resp2.Err)

	// A second leg that errored after the first leg failed did not fail on
	// its own — it was cancelled because the pair failed; report it as such.
	if leg1Failed && resp2.Err != nil {
		resp2.Err = fmt.Errorf("%w: %w", ErrPairLegCancelled, resp2.Err)
	}
	ch1 <- SendResponse{Receipt: receipt1, Nonce: tx1.Nonce(), Err: err1}
	ch2 <- resp2
}

// repairPair consumes any nonce the pair would otherwise leave unmined: a leg
// that resolved with an error never mined, so its nonce gets a cancellation
// no-op (see cancelPairNonce)
func (m *SimpleTxManager) repairPair(parentCtx context.Context, tx1 *types.Transaction, err1 error, tx2 *types.Transaction, err2 error) {
	// Nothing to repair when both legs mined
	if err1 == nil && err2 == nil {
		return
	}

	if parentCtx.Err() != nil {
		// Canceled parent (shutdown, or a sibling send failing the queue's
		// error group): don't publish no-ops, but reset the nonce so the next
		// send re-queries the chain and refills any gap the pair leaves.
		m.l.Info("pair send context cancelled; resetting nonce and leaving in-flight legs to resolve in the pool",
			"firstNonce", tx1.Nonce(), "secondNonce", tx2.Nonce())
		m.resetNonce()
		return
	}

	repairCtx, cancel := getContext(context.WithoutCancel(parentCtx), m.cfg.TxSendTimeout)
	defer cancel()

	var repairs sync.WaitGroup
	if err1 != nil {
		repairs.Add(1)
		go func() {
			defer repairs.Done()
			m.cancelPairNonce(repairCtx, tx1)
		}()
	}
	if err2 != nil {
		repairs.Add(1)
		go func() {
			defer repairs.Done()
			m.cancelPairNonce(repairCtx, tx2)
		}()
	}
	repairs.Wait()
}

// SendPairAsync implements TxManager.SendPairAsync — see the interface doc
// for the pair contract. It crafts both legs synchronously and hands them to
// sendPair; cancellation and nonce repair live in repairPair and
// cancelPairNonce.
func (m *SimpleTxManager) SendPairAsync(
	ctx context.Context,
	firstCandidate TxCandidate,
	secondCandidate TxCandidate,
	firstRespCh chan SendResponse,
	secondRespCh chan SendResponse,
) {
	if cap(firstRespCh) == 0 || cap(secondRespCh) == 0 {
		panic("SendPairAsync: channels must be buffered")
	}

	respondErr := func(err1 error, err2 error) {
		firstRespCh <- SendResponse{Err: err1}
		secondRespCh <- SendResponse{Err: err2}
	}

	if len(firstCandidate.Blobs) > 0 || len(secondCandidate.Blobs) > 0 {
		respondErr(errPairBlobsUnsupported, errPairBlobsUnsupported)
		return
	}

	// refuse new requests if the tx manager is closed
	if m.closed.Load() {
		respondErr(ErrClosed, ErrClosed)
		return
	}

	prepareCtx, cancel := getContext(ctx, m.cfg.TxSendTimeout)
	tx1, err := m.prepare(prepareCtx, firstCandidate)
	cancel()
	if err != nil {
		m.resetNonce()
		respondErr(err, fmt.Errorf("%w: first leg preparation failed: %w", ErrPairLegCancelled, err))
		return
	}

	prepareCtx, cancel = getContext(ctx, m.cfg.TxSendTimeout)
	tx2, err := m.prepare(prepareCtx, secondCandidate)
	cancel()
	if err != nil {
		m.resetNonce()
		respondErr(fmt.Errorf("pair aborted before broadcast: second leg preparation failed: %w", err), err)
		return
	}

	m.metr.RecordPendingTx(m.pending.Add(2))

	go m.sendPair(ctx, tx1, tx2, firstRespCh, secondRespCh)
}

// cancelPairNonce consumes target's nonce on chain by publishing a fee-bumped
// no-op (zero-value self-send) at that exact nonce, and waits for the nonce
// to be consumed — by the no-op, or by target itself if it wins the race
// (both outcomes end with the nonce deterministically spent). This is what
// lets a failed pair guarantee it leaves no gap and no live orphan behind
// without a global nonce reset.
//
// Cancellation is a replacement: the no-op's fees start bumped above target's
// original fees (satisfying the pool's replacement rules), and sendTx's
// underpriced handling keeps bumping if a higher-fee version of target is
// still floating. If the fee limit prevents out-pricing target, target itself
// is high-fee enough to mine, which also consumes the nonce.
//
// On failure (e.g. RPC down for the whole attempt) it falls back to
// resetNonce, restoring the legacy behavior where the next send re-queries
// the chain nonce and refills the gap.
func (m *SimpleTxManager) cancelPairNonce(ctx context.Context, target *types.Transaction) {
	l := m.l.New("nonce", target.Nonce(), "replaces", target.Hash())
	noopTx, err := retry.Do(ctx, 5, retry.Fixed(2*time.Second), func() (*types.Transaction, error) {
		return m.craftPairCancelTx(ctx, target)
	})
	if err != nil {
		l.Error("failed to craft pair cancellation tx; falling back to nonce reset", "err", err)
		m.resetNonce()
		return
	}

	if _, err := m.sendTx(ctx, noopTx); err != nil {
		if errors.Is(err, core.ErrNonceTooLow) {
			// The nonce was consumed by another transaction — target itself,
			// or an earlier cancellation. Either way it is spent, which is
			// all the pair needs.
			l.Info("pair nonce consumed by another transaction before cancellation")
			return
		}
		l.Error("failed to cancel pair nonce; falling back to nonce reset", "err", err)
		m.resetNonce()
		return
	}
	l.Info("pair nonce cancelled", "cancelTx", noopTx.Hash())
}

// craftPairCancelTx builds and signs the no-op replacement
func (m *SimpleTxManager) craftPairCancelTx(ctx context.Context, target *types.Transaction) (*types.Transaction, error) {
	tip, baseFee, _, _, err := m.SuggestGasPriceCaps(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price info: %w", err)
	}
	bumpedTip, bumpedFee := updateFees(target.GasTipCap(), target.GasFeeCap(), tip, baseFee, false, m.l)
	if err := m.checkLimits(tip, baseFee, bumpedTip, bumpedFee); err != nil {
		return nil, fmt.Errorf("fee limits exceeded for pair cancellation tx: %w", err)
	}

	to := m.cfg.From
	msg := &types.DynamicFeeTx{
		ChainID:   m.chainID,
		Nonce:     target.Nonce(),
		To:        &to,
		GasTipCap: bumpedTip,
		GasFeeCap: bumpedFee,
		Gas:       params.TxGas,
	}
	signerCtx, cancel := context.WithTimeout(ctx, m.cfg.NetworkTimeout)
	defer cancel()
	return m.cfg.Signer(signerCtx, m.cfg.From, types.NewTx(msg))
}
