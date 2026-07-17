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

// pairSend drives one SendPairAsync submission from broadcast through repair
// and response delivery.
type pairSend struct {
	txManager *SimpleTxManager
	parentCtx context.Context
	leg1      leg
	leg2      leg
}

// sendPair drives the pair to resolution on its own goroutine: it broadcasts both
// legs concurrently, cancels the second leg if the first fails permanently,
// repairs any nonce the pair would otherwise leave unmined, and delivers
// exactly one response per channel once the pair's on-chain state is clean.
func (p *pairSend) sendPair() {
	m := p.txManager
	defer func() { m.metr.RecordPendingTx(m.pending.Add(-2)) }()
	defer p.leg1.cancelFunc()
	defer p.leg2.cancelFunc()

	leg2Res := make(chan legResult, 1)
	go func() {
		leg2Res <- p.leg2.send(m)
	}()

	legResult1 := p.leg1.send(m)

	// A mined-but-reverted first leg also fails the pair, not just a failed
	// send: the second leg depends on the first leg's successful execution.
	leg1Failed := legResult1.err != nil || (legResult1.receipt != nil && legResult1.receipt.Status != types.ReceiptStatusSuccessful)
	if leg1Failed {
		m.l.Warn("pair first leg failed; cancelling second leg",
			"firstNonce", p.leg1.tx.Nonce(), "secondNonce", p.leg2.tx.Nonce(),
			"err", legResult1.err, "reverted", legResult1.err == nil)
		p.leg2.cancelFunc()
	}

	legResult2 := <-leg2Res

	p.repair(legResult1.err, legResult2.err)

	// A second leg that errored after the first leg failed did not fail on
	// its own — it was cancelled because the pair failed; report it as such.
	if leg1Failed && legResult2.err != nil {
		legResult2.err = fmt.Errorf("%w: %w", ErrPairLegCancelled, legResult2.err)
	}
	p.leg1.respond(legResult1)
	p.leg2.respond(legResult2)
}

// repair consumes any nonce the pair would otherwise leave unmined: a leg
// that resolved with an error never mined, so its nonce gets a cancellation
// no-op (see cancelPairNonce)
func (p *pairSend) repair(err1 error, err2 error) {
	// Nothing to repair when both legs mined
	if err1 == nil && err2 == nil {
		return
	}

	m := p.txManager
	if p.parentCtx.Err() != nil {
		// Canceled parent (shutdown, or a sibling send failing the queue's
		// error group): don't publish no-ops, but reset the nonce so the next
		// send re-queries the chain and refills any gap the pair leaves.
		m.l.Info("pair send context cancelled; resetting nonce and leaving in-flight legs to resolve in the pool",
			"firstNonce", p.leg1.tx.Nonce(), "secondNonce", p.leg2.tx.Nonce())
		m.resetNonce()
		return
	}

	repairCtx, cancel := getContext(context.WithoutCancel(p.parentCtx), m.cfg.TxSendTimeout)
	defer cancel()

	var repairs sync.WaitGroup
	if err1 != nil {
		repairs.Add(1)
		go func() {
			defer repairs.Done()
			m.cancelPairNonce(repairCtx, p.leg1.tx)
		}()
	}
	if err2 != nil {
		repairs.Add(1)
		go func() {
			defer repairs.Done()
			m.cancelPairNonce(repairCtx, p.leg2.tx)
		}()
	}
	repairs.Wait()
}

// SendPairAsync implements TxManager.SendPairAsync — see the interface doc
// for the pair contract. It crafts both legs synchronously and hands them to
// pairSend.sendPair; cancellation and nonce repair live in repair and
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
	if err != nil {
		m.resetNonce()
		cancel()
		respondErr(err, fmt.Errorf("%w: first leg preparation failed: %w", ErrPairLegCancelled, err))
		return
	}

	prepareCtx, cancel = getContext(ctx, m.cfg.TxSendTimeout)
	tx2, err := m.prepare(prepareCtx, secondCandidate)
	if err != nil {
		m.resetNonce()
		cancel()
		respondErr(fmt.Errorf("pair aborted before broadcast: second leg preparation failed: %w", err), err)
		return
	}

	leg1 := leg{responseCh: firstRespCh, tx: tx1}
	leg2 := leg{responseCh: secondRespCh, tx: tx2}
	leg1.sendContext, leg1.cancelFunc = getContext(ctx, m.cfg.TxSendTimeout)
	leg2.sendContext, leg2.cancelFunc = getContext(ctx, m.cfg.TxSendTimeout)

	m.metr.RecordPendingTx(m.pending.Add(2))

	p := &pairSend{
		txManager: m,
		parentCtx: ctx,
		leg1:      leg1,
		leg2:      leg2,
	}
	go p.sendPair()
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
