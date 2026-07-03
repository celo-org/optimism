package batcher

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
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

func newFallbackAuthSubmitter(t *testing.T) *BatchSubmitter {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
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
	require.NoError(t, l.authGroup.Wait())

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

func TestFallbackAuth_AuthFailureRetried(t *testing.T) {
	l := newFallbackAuthSubmitter(t)
	txdata := testFallbackTxData(t)
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Err: errSendFailed},             // auth fails
			{Receipt: receiptWithBlock(101)}, // batch lands anyway
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)
	require.NoError(t, l.authGroup.Wait())

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
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
	require.NoError(t, l.authGroup.Wait())

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
			{Receipt: receiptWithBlock(101)},         // batch lands
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)
	require.NoError(t, l.authGroup.Wait())

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

// TestFallbackAuth_WindowViolationRetried verifies that a batch tx landing
// outside the lookback window of the auth tx produces an error receipt (so the
// channel manager rewinds and resubmits), rather than being confirmed.
func TestFallbackAuth_WindowViolationRetried(t *testing.T) {
	l := newFallbackAuthSubmitter(t)
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
	require.NoError(t, l.authGroup.Wait())

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
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
	require.NoError(t, l.authGroup.Wait())

	got := <-receiptsCh
	require.NoError(t, got.Err)
	require.Equal(t, receiptWithBlock(boundary).BlockNumber, got.Receipt.BlockNumber)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}
