package batcher

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

// dispatchAuthenticatedSendTx routes sendTx through the Espresso (TEE) auth
// path or the fallback-batcher post-fork auth path, returning true when the tx
// has been handed off to authGroup. Returns false to mean "fall through to
// the upstream queue.Send path" — pre-fork fallback batcher and any cancel tx.
//
// The TEE batcher (Config.Espresso.Enabled == true) always authenticates. The
// fallback batcher consults isFallbackAuthRequired to gate authentication
// behind the EspressoTime hardfork: pre-fork the verifier accepts plain
// sender-authenticated batches, and the BatchAuthenticator contract is
// irrelevant. From EspressoTime onward the fallback authenticates every batch,
// which requires activeIsEspresso to be false: the contract's fallback branch is
// the only one its key can satisfy, and a rejected auth leg cancels the paired
// batch leg. See op-batcher/readme.md.
func (l *BatchSubmitter) dispatchAuthenticatedSendTx(txdata txData, isCancel bool, candidate *txmgr.TxCandidate, queue TxSender[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) bool {
	if isCancel {
		return false
	}
	// Espresso batcher: authenticate via BatchAuthenticator. sendTxWithEspresso runs on the
	// caller's goroutine and submits through the ordered queue (submitAuthenticatedBatch spawns
	// its own authGroup-tracked watcher), so it must not be wrapped in a separate goroutine.
	if l.Config.Espresso.Enabled {
		l.sendTxWithEspresso(txdata, isCancel, candidate, queue, receiptsCh)
		return true
	}
	fallbackAuthRequired, err := l.isFallbackAuthRequired(l.killCtx)
	if err != nil {
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  newTxRef(txdata, isCancel),
			Err: fmt.Errorf("failed to evaluate fallback-auth gate: %w", err),
		}
		return true
	}
	if !fallbackAuthRequired {
		return false
	}
	l.sendTxWithFallbackAuth(txdata, isCancel, candidate, queue, receiptsCh)
	return true
}

// isFallbackAuthRequired reports whether the fallback (non-TEE) batcher must
// route its batch txs through BatchAuthenticator.authenticateBatchInfo before
// posting to the BatchInbox. It returns false if the rollup config has a
// zero BatchAuthenticatorAddress, indicating that the BatchAuthenticator-based
// authentication path is not in use.
//
// The gate switches at plain Espresso activation (tip.Time >= EspressoTime), a full grace
// period before the verifier enforces event-based authentication. See
// derive.BatchAuthEnforcementDelaySecs for the full grace-period mechanism and why the
// batcher and verifier gates are offset.
func (l *BatchSubmitter) isFallbackAuthRequired(ctx context.Context) (bool, error) {
	if l.RollupConfig.BatchAuthenticatorAddress == (common.Address{}) {
		return false, nil
	}
	tip, err := l.l1Tip(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to fetch L1 tip for fallback-auth gate: %w", err)
	}
	return l.RollupConfig.IsEspresso(tip.Time), nil
}

// waitForAuthGroup blocks until all in-flight auth goroutines (fallback watchers
// or TEE submissions) have finished. publishingLoop calls it before closing receiptsCh: each watcher
// is a sender on receiptsCh, so the receipts loop must still be draining
// receiptsCh at this point or a watcher's final send would block forever. Each
// watcher always terminates because the txmgr Queue emits exactly one receipt per
// Send, even on context cancellation.
func (l *BatchSubmitter) waitForAuthGroup() {
	l.authGroup.Wait()
}
