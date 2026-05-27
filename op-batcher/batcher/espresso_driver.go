package batcher

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

// authGroup serializes in-flight fallback-auth submissions so the
// publishingLoop can drain them on shutdown. Initialized in
// NewBatchSubmitter and lifted in waitForAuthGroup. The TEE batcher follow-up
// PR reuses the same group.
//
// Bounded to a fixed concurrency limit to cap the number of BatchInbox
// transactions simultaneously waiting on an authenticateBatchInfo
// transaction to be confirmed.
const fallbackAuthGroupLimit = 128

// initAuthGroup applies the concurrency limit. Called from NewBatchSubmitter.
func (l *BatchSubmitter) initAuthGroup() {
	l.authGroup.SetLimit(fallbackAuthGroupLimit)
}

// waitForAuthGroup blocks until all in-flight fallback-auth submissions have
// completed. Called from publishingLoop's tail; blocks until killCtx is
// cancelled if any auth retries are still in flight.
func (l *BatchSubmitter) waitForAuthGroup() {
	if err := l.authGroup.Wait(); err != nil {
		if !errors.Is(err, context.Canceled) {
			l.Log.Error("error waiting for fallback-auth transactions to complete", "err", err)
		}
	}
}

// dispatchAuthenticatedSendTx routes sendTx through the fallback-batcher
// post-fork auth path, returning true when the tx has been handed off to
// authGroup. Returns false to mean "fall through to the upstream queue.Send
// path" — pre-fork operation and any cancel tx.
//
// The fallback batcher consults isFallbackAuthRequired to gate authentication
// behind the EspressoTime hardfork: pre-fork the verifier accepts plain
// sender-authenticated batches, and the BatchAuthenticator contract is
// irrelevant; calling authenticateBatchInfo pre-fork would also revert against
// the default activeIsEspresso=true contract state.
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
	l.authGroup.Go(
		func() error {
			l.sendTxWithFallbackAuth(txdata, isCancel, candidate, queue, receiptsCh)
			return nil
		},
	)
	return true
}
