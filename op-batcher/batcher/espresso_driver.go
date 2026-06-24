package batcher

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

// waitForAuthGroup blocks until all in-flight fallback-auth watcher goroutines
// have finished. publishingLoop calls it before closing receiptsCh: each watcher
// is a sender on receiptsCh, so the receipts loop must still be draining
// receiptsCh at this point or a watcher's final send would block forever. Each
// watcher always terminates because the txmgr Queue emits exactly one receipt per
// Send, even on context cancellation.
func (l *BatchSubmitter) waitForAuthGroup() {
	l.authGroup.Wait()
}

// dispatchAuthenticatedSendTx routes sendTx through the fallback-batcher
// post-fork auth path, returning true when the tx has been handed off to
// authGroup. Returns false to mean "fall through to the upstream queue.Send
// path" — pre-fork operation and any cancel tx.
//
// The fallback batcher consults isFallbackAuthRequired to gate authentication
// behind the EspressoTime hardfork: pre-fork the verifier accepts plain
// sender-authenticated batches, and the BatchAuthenticator contract is
// irrelevant.
func (l *BatchSubmitter) dispatchAuthenticatedSendTx(txdata txData, isCancel bool, candidate *txmgr.TxCandidate, queue TxSender[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) bool {
	if isCancel {
		return false
	}
	if !l.hasBatchAuthenticator() {
		return false
	}
	fallbackAuthRequired, err := l.isFallbackAuthRequired(l.killCtx)
	if err != nil {
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  txRef{id: txdata.ID(), isCancel: isCancel, isBlob: txdata.daType == DaTypeBlob, daType: txdata.daType, size: txdata.Len()},
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
