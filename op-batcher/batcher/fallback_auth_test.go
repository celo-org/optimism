package batcher

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

var errSendFailed = errors.New("send failed")

// recordedSend captures a single queue.Send invocation.
type recordedSend struct {
	candidate txmgr.TxCandidate
	receiptCh chan txmgr.TxReceipt[txRef]
}

// fakeTxSender records Send calls in order and immediately delivers a canned
// response (by index) on the receipt channel, mimicking the txmgr Queue, which
// forwards exactly one receipt per Send.
type fakeTxSender struct {
	sends     []recordedSend
	responses []txmgr.TxReceipt[txRef]
}

func (f *fakeTxSender) Send(id txRef, candidate txmgr.TxCandidate, receiptCh chan txmgr.TxReceipt[txRef]) {
	idx := len(f.sends)
	f.sends = append(f.sends, recordedSend{candidate: candidate, receiptCh: receiptCh})
	resp := f.responses[idx]
	resp.ID = id
	receiptCh <- resp
}

type windowExceededSpy struct {
	metrics.Metricer
	count int
}

func (s *windowExceededSpy) RecordFallbackAuthWindowExceeded() { s.count++ }

func newFallbackAuthSubmitter(t *testing.T) *BatchSubmitter {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.Metr = metrics.NoopMetrics
	l.RollupConfig = &rollup.Config{
		BatchAuthenticatorAddress: common.HexToAddress("0x00000000000000000000000000000000000000aa"),
	}
	return l
}

func testFallbackTxData(t *testing.T) txData {
	return singleFrameTxData(frameData{data: []byte("frame-data")})
}

func receiptWithBlock(num int64) *types.Receipt {
	return &types.Receipt{BlockNumber: big.NewInt(num), Status: types.ReceiptStatusSuccessful}
}

func revertedReceiptWithBlock(num int64) *types.Receipt {
	return &types.Receipt{BlockNumber: big.NewInt(num), Status: types.ReceiptStatusFailed}
}

// TestFallbackAuth_OrderingAndSuccess verifies the auth tx is submitted before
// the batch tx (so it takes the lower nonce and lands first, as Espresso
// requires) and that a single success receipt for the batch txData is emitted
// when both txs land within the lookback window.
func TestFallbackAuth_OrderingAndSuccess(t *testing.T) {
	l := newFallbackAuthSubmitter(t)
	txdata := testFallbackTxData(t)
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)}, // auth
			{Receipt: receiptWithBlock(101)}, // batch
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)

	require.Len(t, queue.sends, 2)
	// First send must target the BatchAuthenticator (the auth tx), giving it the
	// lower, earlier-mined nonce.
	require.NotNil(t, queue.sends[0].candidate.To)
	require.Equal(t, l.RollupConfig.BatchAuthenticatorAddress, *queue.sends[0].candidate.To)
	// Second send is the batch tx itself.
	require.Equal(t, candidate.TxData, queue.sends[1].candidate.TxData)

	got := <-receiptsCh
	require.NoError(t, got.Err)
	require.Equal(t, receiptWithBlock(101).BlockNumber, got.Receipt.BlockNumber)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

// TestFallbackAuth_AuthFailureRetried verifies that an auth-tx send failure produces an error
// receipt keyed to the batch txData so the frames are re-queued. The pair is submitted
// back-to-back (pipelined), so the batch tx is also sent; if its auth never lands the batch is
// simply dropped by derivation — its commitment has no matching auth event in the lookback
// window — so the orphaned send is harmless to safety.
func TestFallbackAuth_AuthFailureRetried(t *testing.T) {
	l := newFallbackAuthSubmitter(t)
	txdata := testFallbackTxData(t)
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Err: errSendFailed},             // auth fails
			{Receipt: receiptWithBlock(101)}, // batch send (result discarded on auth failure)
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
	// Both txs are submitted back-to-back; the auth failure does not gate the batch send.
	require.Len(t, queue.sends, 2)
}

// TestFallbackAuth_AuthFailureTxRefType verifies that an auth-tx failure is
// reported under a calldata-typed txRef even when the batch txdata is blob.
// The auth tx is always calldata; if its ErrAlreadyReserved failure were
// labeled with the batch's blob type, cancelBlockingTx would send a calldata
// cancel against a blobpool reservation, which is rejected the same way,
// looping forever without ever displacing the stuck blob tx.
func TestFallbackAuth_AuthFailureTxRefType(t *testing.T) {
	l := newFallbackAuthSubmitter(t)
	txdata := testFallbackTxData(t)
	txdata.daType = DaTypeBlob
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Err: errSendFailed},             // auth fails
			{Receipt: receiptWithBlock(101)}, // batch send (result discarded on auth failure)
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
	require.Len(t, queue.sends, 2)
	require.False(t, got.ID.isBlob)
	require.Equal(t, DaTypeCalldata, got.ID.daType)
}

// TestFallbackAuth_BatchFailureTxRefType verifies the converse: a batch-tx
// failure keeps the batch txdata's own type on the forwarded receipt.
func TestFallbackAuth_BatchFailureTxRefType(t *testing.T) {
	l := newFallbackAuthSubmitter(t)
	txdata := testFallbackTxData(t)
	txdata.daType = DaTypeBlob
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)}, // auth lands
			{Err: errSendFailed},             // batch fails
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
	require.Len(t, queue.sends, 2)
	require.True(t, got.ID.isBlob)
	require.Equal(t, DaTypeBlob, got.ID.daType)
}

func TestFallbackAuth_BatchFailureRetried(t *testing.T) {
	l := newFallbackAuthSubmitter(t)
	txdata := testFallbackTxData(t)
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)}, // auth lands
			{Err: errSendFailed},             // batch fails
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

// TestFallbackAuth_AuthRevertedRetried verifies that an authenticateBatchInfo tx
// that mines but reverts (no event emitted for the verifier) produces an error
// receipt so the frames are re-queued, rather than being confirmed as success.
func TestFallbackAuth_AuthRevertedRetried(t *testing.T) {
	l := newFallbackAuthSubmitter(t)
	txdata := testFallbackTxData(t)
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: revertedReceiptWithBlock(100)}, // auth mined but reverted
			{Receipt: receiptWithBlock(101)},         // batch send (result discarded on auth revert)
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
	// A reverted auth tx emits no BatchInfoAuthenticated event, so the batch is unverifiable and
	// an error receipt is forwarded to re-queue the frames. Both txs are still submitted
	// (pipelined); the orphaned batch is dropped by derivation for lack of a matching auth event.
	require.Len(t, queue.sends, 2)
}

// TestFallbackAuth_WindowViolationRetried verifies that a batch tx landing
// outside the lookback window of the auth tx produces an error receipt (so the
// channel manager rewinds and resubmits), rather than being confirmed.
func TestFallbackAuth_WindowViolationRetried(t *testing.T) {
	l := newFallbackAuthSubmitter(t)
	metr := &windowExceededSpy{Metricer: metrics.NoopMetrics}
	l.Metr = metr
	txdata := testFallbackTxData(t)
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	tooFar := int64(100 + derive.BatchAuthLookbackWindow + 1)
	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)},    // auth
			{Receipt: receiptWithBlock(tooFar)}, // batch too far away
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
	require.Equal(t, 1, metr.count, "window violation should record the fallback_auth_window_exceeded metric")
}

// TestFallbackAuth_WindowBoundaryAccepted pins the inclusive bound of the lookback
// window: a batch landing exactly BatchAuthLookbackWindow blocks after the auth tx is
// still accepted by the verifier (CollectAuthenticatedBatches scans
// [batchBlock - BatchAuthLookbackWindow, batchBlock]), so the batcher must not
// re-queue it.
func TestFallbackAuth_WindowBoundaryAccepted(t *testing.T) {
	l := newFallbackAuthSubmitter(t)
	txdata := testFallbackTxData(t)
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	boundary := int64(100 + derive.BatchAuthLookbackWindow)
	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)},      // auth
			{Receipt: receiptWithBlock(boundary)}, // batch at the exact edge of the window
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)

	got := <-receiptsCh
	require.NoError(t, got.Err)
	require.Equal(t, receiptWithBlock(boundary).BlockNumber, got.Receipt.BlockNumber)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

// TestComputeCommitment_Parity locks the batcher's batch-commitment computation to
// the verifier's. The batcher must hash exactly what op-node derivation hashes, or
// post-fork batches fail the commitment match and are silently dropped. It checks
// both paths against derive.ComputeCalldataBatchHash / derive.ComputeBlobBatchHash;
// for blobs it uses the real versioned hashes the tx will carry (via MakeSidecar,
// the same hashes the verifier reads from tx.BlobHashes()).
func TestComputeCommitment_Parity(t *testing.T) {
	t.Run("calldata", func(t *testing.T) {
		for _, data := range [][]byte{[]byte("batch calldata payload"), {}, nil} {
			got, err := computeCommitment(&txmgr.TxCandidate{TxData: data})
			require.NoError(t, err)
			require.Equal(t, derive.ComputeCalldataBatchHash(data), common.Hash(got))
		}
	})

	t.Run("blobs", func(t *testing.T) {
		for _, n := range []int{1, 3} {
			blobs := make([]*eth.Blob, n)
			for i := range blobs {
				var blob eth.Blob
				// Distinct first byte per blob so the versioned hashes differ and the
				// concatenation order is actually exercised.
				require.NoError(t, blob.FromData(eth.Data{byte(i), 0xab, 0xcd}))
				blobs[i] = &blob
			}
			_, blobHashes, err := txmgr.MakeSidecar(blobs, false)
			require.NoError(t, err)

			got, err := computeCommitment(&txmgr.TxCandidate{Blobs: blobs})
			require.NoError(t, err)
			require.Equal(t, derive.ComputeBlobBatchHash(blobHashes), common.Hash(got))
		}
	})
}
