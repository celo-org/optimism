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
	transactionReference := newTxRef(txdata, isCancel)
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

	l.submitAuthenticatedBatch(transactionReference, verifyCandidate, candidate, queue, receiptsCh)
}

// submitAuthenticatedBatch submits an authenticateBatchInfo tx (verifyCandidate) followed by the
// batch inbox tx (candidate), then spawns a watcher (tracked by authGroup so the publishing loop
// drains it before closing receiptsCh) that validates both receipts and emits a single synthetic
// receipt for the batch txData.
//
// Both the TEE and fallback auth paths share this submission flow; they differ only in how the
// authenticateBatchInfo calldata in verifyCandidate is built (TEE-attested signature vs empty
// signature relying on the contract's msg.sender check).
//
// The auth leg must land at or before the batch leg (the verifier scans a lookback window ending
// at the batch tx's L1 block), so the auth tx must take the lower, earlier-mined nonce. Calldata
// pairs are pipelined via queue.SendPair (auth leg first); blob batches fall back to a serialized
// submission because geth's cross-pool account reservation cannot hold a calldata auth tx and a
// blob batch tx at once.
func (l *BatchSubmitter) submitAuthenticatedBatch(transactionReference txRef, verifyCandidate txmgr.TxCandidate, candidate *txmgr.TxCandidate, queue TxSender[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) {
	// The auth tx shares the batch txdata's identity (so a failure requeues the right frames)
	// but is always a calldata tx. Auth failures must carry its real type: an ErrAlreadyReserved
	// receipt labeled with the batch's blob type would make cancelBlockingTx cancel the wrong
	// pool, leaving the reserving tx stuck.
	authReference := transactionReference
	authReference.isBlob = false
	authReference.daType = DaTypeCalldata

	l.Log.Debug(
		"Sending authenticateBatchInfo transaction",
		"txRef", transactionReference,
		"address", l.RollupConfig.BatchAuthenticatorAddress.String(),
	)

	if len(candidate.Blobs) > 0 {
		// SendPair doesn't support blobs. Unreachable while calldata-only DA is
		// enforced: checkEspressoDataAvailability refuses to start a blob- or
		// auto-configured batcher on a chain with EspressoTime set.
		l.sendAuthSerialized(transactionReference, authReference, verifyCandidate, candidate, queue, receiptsCh)
		return
	}

	// Private buffered channels — Queue.SendPair forwards exactly one receipt to
	// each, so the watcher reads exactly once per channel. These never reach
	// handleReceipt; only the synthetic receipt from the watcher does.
	authReceiptCh := make(chan txmgr.TxReceipt[txRef], 1)
	batchReceiptCh := make(chan txmgr.TxReceipt[txRef], 1)

	// Calldata pair: pipelined, auth leg first; see TxManager.SendPairAsync for the
	// contract (ordering, slot accounting, cancellation of the batch leg on failure).
	queue.SendPair(authReference, verifyCandidate, authReceiptCh, transactionReference, *candidate, batchReceiptCh)

	l.authGroup.Add(1)
	go func() {
		defer l.authGroup.Done()
		l.watchAuthReceipts(transactionReference, authReference, authReceiptCh, batchReceiptCh, receiptsCh)
	}()
}

// sendAuthSerialized submits an auth+batch pair serially: the auth tx is sent
// and confirmed first, then the batch tx. It runs on the publishing-loop goroutine and
// blocks it for the pair's full confirmation cycle, so consecutive pairs never overlap.
// Its only caller is the blob branch of submitAuthenticatedBatch, so it cannot run while
// calldata-only DA is enforced.
func (l *BatchSubmitter) sendAuthSerialized(transactionReference, authReference txRef, verifyCandidate txmgr.TxCandidate, candidate *txmgr.TxCandidate, queue TxSender[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) {
	authReceiptCh := make(chan txmgr.TxReceipt[txRef], 1)
	queue.Send(authReference, verifyCandidate, authReceiptCh)
	authResult := <-authReceiptCh
	if err := authResultError(authResult); err != nil {
		l.Log.Error("authenticateBatchInfo transaction failed", "txRef", transactionReference, "err", err)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  authReference,
			Err: err,
		}
		return
	}

	// The auth tx is confirmed on L1, releasing the account's reservation in the
	// calldata pool. Now send the blob batch tx.
	batchReceiptCh := make(chan txmgr.TxReceipt[txRef], 1)
	queue.Send(transactionReference, *candidate, batchReceiptCh)
	batchResult := <-batchReceiptCh
	l.resolveAuthPair(transactionReference, authReference, authResult, batchResult, receiptsCh)
}

// watchAuthReceipts collects the auth and batch receipts of a pipelined auth+batch
// pair and resolves them into a single synthetic receipt (see resolveAuthPair).
func (l *BatchSubmitter) watchAuthReceipts(transactionReference, authReference txRef, authReceiptCh, batchReceiptCh chan txmgr.TxReceipt[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) {
	authResult := <-authReceiptCh
	batchResult := <-batchReceiptCh
	l.resolveAuthPair(transactionReference, authReference, authResult, batchResult, receiptsCh)
}

// authResultError interprets the auth leg's receipt: a send failure or a
// mined-but-reverted auth tx both fail the pair, since a reverted authenticateBatchInfo call
// emits no BatchInfoAuthenticated event and the verifier would drop the batch (txmgr returns a
// receipt as soon as the tx is mined, regardless of execution status). The batch inbox tx
// needs no such status check: derivation reads its data by L1 inclusion, not by execution
// status.
func authResultError(authResult txmgr.TxReceipt[txRef]) error {
	if authResult.Err != nil {
		return fmt.Errorf("failed to send authenticateBatchInfo transaction: %w", authResult.Err)
	}
	if authResult.Receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("authenticateBatchInfo transaction reverted: %s", authResult.Receipt.TxHash)
	}
	return nil
}

// resolveAuthPair validates the receipts of an auth+batch pair — auth success, batch
// success, and that the batch tx landed within the lookback window of the auth tx — and
// forwards a single synthetic receipt keyed to the batch txData onto receiptsCh. Any failure
// produces an error receipt so the channel manager rewinds and resubmits the frame set. Auth
// failures are reported under authReference (see its construction in submitAuthenticatedBatch
// for why the calldata typing matters).
func (l *BatchSubmitter) resolveAuthPair(transactionReference, authReference txRef, authResult, batchResult txmgr.TxReceipt[txRef], receiptsCh chan txmgr.TxReceipt[txRef]) {
	if err := authResultError(authResult); err != nil {
		l.Log.Error("authenticateBatchInfo transaction failed", "txRef", transactionReference, "err", err)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  authReference,
			Err: err,
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
		l.Metr.RecordFallbackAuthWindowExceeded()
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
