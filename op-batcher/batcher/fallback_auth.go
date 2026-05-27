package batcher

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum-optimism/optimism/espresso/bindings"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

// computeCommitment computes the batch commitment hash from a transaction candidate.
// For calldata transactions, it returns keccak256(calldata).
// For blob transactions, it returns keccak256(concat(blobVersionedHashes)).
func computeCommitment(candidate *txmgr.TxCandidate) ([32]byte, error) {
	if len(candidate.Blobs) == 0 {
		return crypto.Keccak256Hash(candidate.TxData), nil
	}

	concatenatedBlobHashes := make([]byte, 0)
	for _, blob := range candidate.Blobs {
		blobCommitment, err := blob.ComputeKZGCommitment()
		if err != nil {
			return [32]byte{}, fmt.Errorf("failed to compute KZG commitment for blob: %w", err)
		}
		blobHash := eth.KZGToVersionedHash(blobCommitment)
		concatenatedBlobHashes = append(concatenatedBlobHashes, blobHash.Bytes()...)
	}
	return crypto.Keccak256Hash(concatenatedBlobHashes), nil
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

	l.Log.Debug(
		"Sending fallback authenticateBatchInfo transaction",
		"txRef", transactionReference,
		"commitment", hexutil.Encode(commitment[:]),
		"address", l.RollupConfig.BatchAuthenticatorAddress.String(),
	)
	verificationReceipt, err := l.Txmgr.Send(l.killCtx, verifyCandidate)
	if err != nil {
		l.Log.Error("Failed to send fallback authenticateBatchInfo transaction", "txRef", transactionReference, "err", err)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to send fallback authenticateBatchInfo transaction: %w", err),
		}
		return
	}

	receipt, err := l.Txmgr.Send(l.killCtx, *candidate)
	if err != nil {
		l.Log.Error("Failed to send batch inbox transaction", "txRef", transactionReference, "err", err)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("failed to send batch inbox transaction: %w", err),
		}
		return
	}

	distance := new(big.Int).Sub(receipt.BlockNumber, verificationReceipt.BlockNumber)
	lookbackWindow := new(big.Int).SetUint64(derive.BatchAuthLookbackWindow)
	if distance.Sign() < 0 || distance.Cmp(lookbackWindow) >= 0 {
		l.Log.Error("authenticateBatchInfo transaction too far from batch inbox transaction", "txRef", transactionReference, "distance", distance)
		receiptsCh <- txmgr.TxReceipt[txRef]{
			ID:  transactionReference,
			Err: fmt.Errorf("authenticateBatchInfo transaction too far from batch inbox transaction: %s", distance),
		}
		return
	}

	receiptsCh <- txmgr.TxReceipt[txRef]{
		ID:      transactionReference,
		Receipt: receipt,
		Err:     nil,
	}
}
