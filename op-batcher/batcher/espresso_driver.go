package batcher

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

// dispatchAuthenticatedSendTx routes sendTx through the fallback-batcher
// post-fork auth path, returning true when the tx has been fully handled
// (sendTxWithFallbackAuth blocks until the pair is resolved and a receipt has
// been forwarded). Returns false to mean "fall through to the upstream
// queue.Send path" — pre-fork operation and any cancel tx.
//
// The fallback batcher consults isFallbackAuthRequired to gate authentication
// behind the EspressoTime hardfork: pre-fork the verifier accepts plain
// sender-authenticated batches, and the BatchAuthenticator contract is
// irrelevant.
func (l *BatchSubmitter) dispatchAuthenticatedSendTx(txdata txData, isCancel bool, candidate *txmgr.TxCandidate, queue TxSender[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) bool {
	if isCancel {
		return false
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
