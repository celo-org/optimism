package batcher

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum-optimism/optimism/espresso/bindings"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

// computeCommitment computes the batch commitment hash from a transaction
// candidate. It delegates to the same functions the verifier uses so the two
// provably agree on the bytes that get authenticated:
//   - calldata transactions: keccak256(calldata).
//   - blob transactions: keccak256(concat(blobVersionedHashes)).
func computeCommitment(candidate *txmgr.TxCandidate) ([32]byte, error) {
	if len(candidate.Blobs) == 0 {
		return derive.ComputeCalldataBatchHash(candidate.TxData), nil
	}

	blobHashes := make([]common.Hash, len(candidate.Blobs))
	for i, blob := range candidate.Blobs {
		blobCommitment, err := blob.ComputeKZGCommitment()
		if err != nil {
			return [32]byte{}, fmt.Errorf("failed to compute KZG commitment for blob: %w", err)
		}
		blobHashes[i] = eth.KZGToVersionedHash(blobCommitment)
	}
	return derive.ComputeBlobBatchHash(blobHashes), nil
}

// sendTxWithFallbackAuth authenticates a batch transaction via the BatchAuthenticator contract
// using the fallback batcher's sender identity (msg.sender check on-chain), then sends the
// batch data to the BatchInbox address.
//
// The contract's fallback path checks msg.sender against systemConfig.batcherHash(), so no
// separate signature is needed — the L1 transaction is already signed by the TxManager's key.
func (l *BatchSubmitter) sendTxWithFallbackAuth(txdata txData, isCancel bool, candidate *txmgr.TxCandidate, queue TxSender[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) {
	transactionReference := txRef{id: txdata.ID(), isCancel: isCancel, isBlob: txdata.daType == DaTypeBlob, daType: txdata.daType, size: txdata.Len()}
	l.Log.Debug("Sending fallback-authenticated L1 transaction", "txRef", transactionReference)

	commitment, err := computeCommitment(candidate)
	if err != nil {
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to compute commitment: %w", err),
		}
		return
	}
	l.Log.Debug("Computed fallback batch commitment", "txRef", transactionReference, "commitment", hexutil.Encode(commitment[:]))

	batchAuthenticatorAbi, err := bindings.BatchAuthenticatorMetaData.GetAbi()
	if err != nil {
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to get batch authenticator ABI: %w", err),
		}
		return
	}

	// Pass an empty signature — the contract checks msg.sender for the fallback path.
	authenticateBatchCalldata, err := batchAuthenticatorAbi.Pack("authenticateBatchInfo", commitment, []byte{})
	if err != nil {
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to pack authenticateBatchInfo calldata: %w", err),
		}
		return
	}

	verifyCandidate := txmgr.TxCandidate{
		TxData: authenticateBatchCalldata,
		To:     &l.RollupConfig.BatchAuthenticatorAddress,
	}

	// Private buffered channels: queue.Send forwards exactly one receipt to each, so the watcher
	// reads exactly once per channel (even on context cancellation, the queue still emits a
	// ctx-error receipt). These never reach handleReceipt; only the synthetic receipt below does.
	authReceiptCh := make(chan txmgr.TxReceipt[txRef], 1)
	batchReceiptCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.Log.Debug(
		"Sending fallback authenticateBatchInfo transaction",
		"txRef", transactionReference,
		"commitment", hexutil.Encode(commitment[:]),
		"address", l.RollupConfig.BatchAuthenticatorAddress.String(),
	)
	// Submit the auth tx then the batch tx, in order, on the publishing-loop goroutine so their
	// nonces are assigned in submission order. Each Send blocks here when the queue is at its
	// MaxPendingTransactions limit.
	queue.Send(transactionReference, verifyCandidate, authReceiptCh)
	queue.Send(transactionReference, *candidate, batchReceiptCh)

	l.authGroup.Go(func() error {
		l.watchFallbackAuthReceipts(transactionReference, authReceiptCh, batchReceiptCh, receiptsCh)
		return nil
	})
}

// watchFallbackAuthReceipts collects the auth and batch receipts for a fallback-auth pair,
// validates that the batch tx landed within the lookback window of the auth tx, and forwards a
// single synthetic receipt keyed to the batch txData onto receiptsCh. Any failure produces an
// error receipt so the channel manager rewinds and resubmits the frame set.
func (l *BatchSubmitter) watchFallbackAuthReceipts(transactionReference txRef, authReceiptCh, batchReceiptCh chan txmgr.TxReceipt[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) {
	authResult := <-authReceiptCh
	batchResult := <-batchReceiptCh

	if authResult.Err != nil {
		l.Log.Error("Failed to send fallback authenticateBatchInfo transaction", "txRef", transactionReference, "err", authResult.Err)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to send fallback authenticateBatchInfo transaction: %w", authResult.Err),
		}
		return
	}

	// txmgr returns a receipt as soon as the tx is mined, regardless of execution status. A
	// reverted authenticateBatchInfo call emits no BatchInfoAuthenticated event, so the verifier
	// drops the batch and the safe head stalls; report failure so the frames are re-queued. The
	// batch inbox tx needs no such check: derivation reads its data by L1 inclusion, not by
	// execution status.
	if authResult.Receipt.Status != types.ReceiptStatusSuccessful {
		l.Log.Error("Fallback authenticateBatchInfo transaction reverted", "txRef", transactionReference, "txHash", authResult.Receipt.TxHash)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("fallback authenticateBatchInfo transaction reverted: %s", authResult.Receipt.TxHash),
		}
		return
	}

	if batchResult.Err != nil {
		l.Log.Error("Failed to send batch inbox transaction", "txRef", transactionReference, "err", batchResult.Err)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to send batch inbox transaction: %w", batchResult.Err),
		}
		return
	}

	distance := new(big.Int).Sub(batchResult.Receipt.BlockNumber, authResult.Receipt.BlockNumber)
	lookbackWindow := new(big.Int).SetUint64(derive.BatchAuthLookbackWindow)
	if distance.Sign() < 0 || distance.Cmp(lookbackWindow) > 0 {
		l.Log.Error("authenticateBatchInfo transaction too far from batch inbox transaction", "txRef", transactionReference, "distance", distance)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("authenticateBatchInfo transaction too far from batch inbox transaction: %s", distance),
		}
		return
	}

	receiptsCh <- txmgr.TxReceipt[txRef]{
		ID:      transactionReference,
		Receipt: batchResult.Receipt,
		Err:     nil,
	}
}
