package batcher

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/espresso/bindings"
	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

var errSendFailed = errors.New("send failed")

// recordedSend captures a single queued transaction, whether it arrived via
// Send or as a leg of SendPair.
type recordedSend struct {
	candidate txmgr.TxCandidate
	receiptCh chan txmgr.TxReceipt[txRef]
	viaPair   bool
}

// fakeTxSender records queued transactions in order and immediately delivers a
// canned response (by index) on the receipt channel, mimicking the txmgr Queue,
// which forwards exactly one receipt per queued transaction. SendPair records
// its two legs as two consecutive sends consuming two consecutive responses.
type fakeTxSender struct {
	sends     []recordedSend
	pairCalls int
	responses []txmgr.TxReceipt[txRef]
}

func (f *fakeTxSender) deliver(id txRef, candidate txmgr.TxCandidate, receiptCh chan txmgr.TxReceipt[txRef], viaPair bool) {
	idx := len(f.sends)
	f.sends = append(f.sends, recordedSend{candidate: candidate, receiptCh: receiptCh, viaPair: viaPair})
	resp := f.responses[idx]
	resp.ID = id
	receiptCh <- resp
}

func (f *fakeTxSender) Send(id txRef, candidate txmgr.TxCandidate, receiptCh chan txmgr.TxReceipt[txRef]) {
	f.deliver(id, candidate, receiptCh, false)
}

func (f *fakeTxSender) SendPair(firstID txRef, first txmgr.TxCandidate, firstCh chan txmgr.TxReceipt[txRef], secondID txRef, second txmgr.TxCandidate, secondCh chan txmgr.TxReceipt[txRef]) {
	f.pairCalls++
	f.deliver(firstID, first, firstCh, true)
	f.deliver(secondID, second, secondCh, true)
}

// fallbackAuthMetricsSpy counts the auth failure metrics so tests can assert the
// right signal fired for each failure mode.
type fallbackAuthMetricsSpy struct {
	metrics.Metricer
	windowExceeded int
}

func (s *fallbackAuthMetricsSpy) RecordFallbackAuthWindowExceeded() { s.windowExceeded++ }

func newAuthSubmitter(t *testing.T) *BatchSubmitter {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.Metr = metrics.NoopMetrics
	l.RollupConfig = &rollup.Config{
		BatchAuthenticatorAddress: common.HexToAddress("0x00000000000000000000000000000000000000aa"),
	}
	return l
}

func testAuthTxData() txData {
	return singleFrameTxData(frameData{data: []byte("frame-data")})
}

// testBlobCandidate returns a tx candidate carrying one (zero) blob, which routes
// submitAuthenticatedBatch onto the serialized blob path.
func testBlobCandidate() *txmgr.TxCandidate {
	return &txmgr.TxCandidate{Blobs: []*eth.Blob{{}}}
}

func receiptWithBlock(num int64) *types.Receipt {
	return &types.Receipt{BlockNumber: big.NewInt(num), Status: types.ReceiptStatusSuccessful}
}

func revertedReceiptWithBlock(num int64) *types.Receipt {
	return &types.Receipt{BlockNumber: big.NewInt(num), Status: types.ReceiptStatusFailed}
}

// authTxRef builds the txRef the auth paths key their receipts under.
func authTxRef(txdata txData) txRef {
	return txRef{id: txdata.ID(), isBlob: txdata.daType == DaTypeBlob, daType: txdata.daType, size: txdata.Len()}
}

// The following suite drives submitAuthenticatedBatch / watchAuthReceipts directly. This is the
// flow shared by both the fallback and TEE auth paths (they differ only in how the auth calldata
// is built; see TestFallbackAuth_Calldata / TestEspressoAuth_Calldata for that distinction).

// TestSubmitAuthenticatedBatch_OrderingAndSuccess verifies a calldata batch is submitted as a
// single pipelined pair — the auth tx first (so it takes the lower nonce and lands first, as
// required under Holocene), the batch tx second — and that a single success receipt for the batch
// txData is emitted when both txs land within the lookback window.
func TestSubmitAuthenticatedBatch_OrderingAndSuccess(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()
	ref := authTxRef(txdata)
	verifyCandidate := txmgr.TxCandidate{TxData: []byte("auth-calldata"), To: &l.RollupConfig.BatchAuthenticatorAddress}
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)}, // auth
			{Receipt: receiptWithBlock(101)}, // batch
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.submitAuthenticatedBatch(ref, verifyCandidate, candidate, queue, receiptsCh)
	l.authGroup.Wait()

	require.Len(t, queue.sends, 2)
	require.Equal(t, 1, queue.pairCalls, "calldata pair must be submitted via SendPair")
	require.True(t, queue.sends[0].viaPair)
	require.True(t, queue.sends[1].viaPair)
	// First send is the auth tx, giving it the lower, earlier-mined nonce.
	require.Equal(t, verifyCandidate.TxData, queue.sends[0].candidate.TxData)
	// Second send is the batch tx itself.
	require.Equal(t, candidate.TxData, queue.sends[1].candidate.TxData)

	got := <-receiptsCh
	require.NoError(t, got.Err)
	require.Equal(t, receiptWithBlock(101).BlockNumber, got.Receipt.BlockNumber)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

func TestSubmitAuthenticatedBatch_AuthSendFailureRetried(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Err: errSendFailed},             // auth fails to send
			{Err: txmgr.ErrPairLegCancelled}, // batch leg cancelled by the pair
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.submitAuthenticatedBatch(authTxRef(txdata), txmgr.TxCandidate{}, &txmgr.TxCandidate{}, queue, receiptsCh)
	l.authGroup.Wait()

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

func TestSubmitAuthenticatedBatch_BatchSendFailureRetried(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)}, // auth lands
			{Err: errSendFailed},             // batch fails to send
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.submitAuthenticatedBatch(authTxRef(txdata), txmgr.TxCandidate{}, &txmgr.TxCandidate{}, queue, receiptsCh)
	l.authGroup.Wait()

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

// TestSubmitAuthenticatedBatch_AuthRevertedRetried verifies that an authenticateBatchInfo tx that
// mines but reverts (no event emitted for the verifier) produces an error receipt so the frames
// are re-queued, rather than being confirmed as success.
func TestSubmitAuthenticatedBatch_AuthRevertedRetried(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: revertedReceiptWithBlock(100)}, // auth mined but reverted
			{Err: txmgr.ErrPairLegCancelled},         // batch leg cancelled by the pair
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.submitAuthenticatedBatch(authTxRef(txdata), txmgr.TxCandidate{}, &txmgr.TxCandidate{}, queue, receiptsCh)
	l.authGroup.Wait()

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

// TestSubmitAuthenticatedBatch_WindowViolationRetried verifies that a batch tx landing outside the
// lookback window of the auth tx produces an error receipt (so the channel manager rewinds and
// resubmits), rather than being confirmed, and records the window-exceeded metric.
func TestSubmitAuthenticatedBatch_WindowViolationRetried(t *testing.T) {
	l := newAuthSubmitter(t)
	metr := &fallbackAuthMetricsSpy{Metricer: metrics.NoopMetrics}
	l.Metr = metr
	txdata := testAuthTxData()

	tooFar := int64(100 + derive.BatchAuthLookbackWindow + 1)
	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)},    // auth
			{Receipt: receiptWithBlock(tooFar)}, // batch too far away
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.submitAuthenticatedBatch(authTxRef(txdata), txmgr.TxCandidate{}, &txmgr.TxCandidate{}, queue, receiptsCh)
	l.authGroup.Wait()

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
	require.Equal(t, 1, metr.windowExceeded, "window violation should record the window-exceeded metric")
}

// TestSubmitAuthenticatedBatch_WindowBoundaryAccepted pins the inclusive bound of the
// lookback window: a batch landing exactly BatchAuthLookbackWindow blocks after the auth
// tx is still accepted by the verifier (CollectAuthenticatedBatches scans
// [batchBlock - BatchAuthLookbackWindow, batchBlock]), so the batcher must not
// re-queue it.
func TestSubmitAuthenticatedBatch_WindowBoundaryAccepted(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()

	boundary := int64(100 + derive.BatchAuthLookbackWindow)
	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)},      // auth
			{Receipt: receiptWithBlock(boundary)}, // batch at the exact edge of the window
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.submitAuthenticatedBatch(authTxRef(txdata), txmgr.TxCandidate{}, &txmgr.TxCandidate{}, queue, receiptsCh)
	l.authGroup.Wait()

	got := <-receiptsCh
	require.NoError(t, got.Err)
	require.Equal(t, receiptWithBlock(boundary).BlockNumber, got.Receipt.BlockNumber)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

// TestAuth_BlobSerialized verifies a blob batch takes the serialized path: the auth tx is sent
// alone (not as a pair) and the blob batch tx is only sent after the auth tx confirms, because
// geth's cross-pool account reservation cannot hold a calldata auth tx and a blob batch tx at once.
func TestAuth_BlobSerialized(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()
	txdata.daType = DaTypeBlob
	candidate := testBlobCandidate()

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)}, // auth
			{Receipt: receiptWithBlock(101)}, // batch, sent only after auth confirmed
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)
	l.authGroup.Wait()

	require.Len(t, queue.sends, 2)
	require.Equal(t, 0, queue.pairCalls, "blob pair must not be pipelined")
	require.False(t, queue.sends[0].viaPair)
	require.False(t, queue.sends[1].viaPair)
	require.Equal(t, l.RollupConfig.BatchAuthenticatorAddress, *queue.sends[0].candidate.To)

	got := <-receiptsCh
	require.NoError(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

// TestAuth_BlobAuthFailureGatesBatch verifies the serialized blob path never sends the batch tx
// when the auth tx fails: the batch would be crafted at the next nonce while the failed auth send
// resets the txmgr nonce, leaving a nonce gap (and an unauthenticated blob batch) behind.
func TestAuth_BlobAuthFailureGatesBatch(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()
	txdata.daType = DaTypeBlob
	candidate := testBlobCandidate()

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Err: errSendFailed}, // auth fails. No batch response, it must never be sent
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)
	l.authGroup.Wait()

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
	require.Len(t, queue.sends, 1)
}

// TestAuth_BlobAuthRevertGatesBatch is the revert variant: a mined but reverted auth emits no
// BatchInfoAuthenticated event, so the blob batch would be unverifiable and must not be submitted
// at all.
func TestAuth_BlobAuthRevertGatesBatch(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()
	txdata.daType = DaTypeBlob
	candidate := testBlobCandidate()

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: revertedReceiptWithBlock(100)}, // auth mined but reverted
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)
	l.authGroup.Wait()

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
	require.Len(t, queue.sends, 1)
}

// TestAuth_AuthFailureTxRefType verifies that an auth-tx failure is reported under a
// calldata-typed txRef even when the batch txdata is blob. The auth tx is always calldata; if its
// ErrAlreadyReserved failure were labeled with the batch's blob type, cancelBlockingTx would send
// a calldata cancel against a blobpool reservation, which is rejected the same way, looping
// forever without ever displacing the stuck blob tx.
func TestAuth_AuthFailureTxRefType(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()
	txdata.daType = DaTypeBlob
	candidate := testBlobCandidate()

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Err: errSendFailed}, // auth fails. No batch response, it must never be sent
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)
	l.authGroup.Wait()

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
	require.Len(t, queue.sends, 1)
	require.False(t, got.ID.isBlob)
	require.Equal(t, DaTypeCalldata, got.ID.daType)
}

// TestAuth_BatchFailureTxRefType verifies the converse: a batch-tx failure keeps the batch
// txdata's own type on the forwarded receipt.
func TestAuth_BatchFailureTxRefType(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()
	txdata.daType = DaTypeBlob
	candidate := testBlobCandidate()

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)}, // auth lands
			{Err: errSendFailed},             // batch fails
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)
	l.authGroup.Wait()

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
	require.Len(t, queue.sends, 2)
	require.True(t, got.ID.isBlob)
	require.Equal(t, DaTypeBlob, got.ID.daType)
}

// unpackAuthenticateBatchInfo decodes an authenticateBatchInfo(bytes32,bytes) calldata blob into
// its commitment and signature arguments.
func unpackAuthenticateBatchInfo(t *testing.T, calldata []byte) (commitment [32]byte, signature []byte) {
	abi, err := bindings.BatchAuthenticatorMetaData.GetAbi()
	require.NoError(t, err)
	method, ok := abi.Methods["authenticateBatchInfo"]
	require.True(t, ok)
	require.Equal(t, method.ID, calldata[:4], "calldata is not an authenticateBatchInfo call")

	args, err := method.Inputs.Unpack(calldata[4:])
	require.NoError(t, err)
	require.Len(t, args, 2)
	return args[0].([32]byte), args[1].([]byte)
}

// TestFallbackAuth_Calldata verifies the distinguishing behavior of the fallback path: the auth tx
// it submits is authenticateBatchInfo(commitment, emptySig), relying on the contract's msg.sender
// check rather than a signature.
func TestFallbackAuth_Calldata(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData()
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)},
			{Receipt: receiptWithBlock(101)},
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)
	l.authGroup.Wait()

	require.Len(t, queue.sends, 2)
	require.NotNil(t, queue.sends[0].candidate.To)
	require.Equal(t, l.RollupConfig.BatchAuthenticatorAddress, *queue.sends[0].candidate.To)

	expectedCommitment, err := computeCommitment(candidate)
	require.NoError(t, err)
	commitment, signature := unpackAuthenticateBatchInfo(t, queue.sends[0].candidate.TxData)
	require.Equal(t, expectedCommitment, commitment)
	require.Empty(t, signature, "fallback path must send an empty signature")
}

// TestEspressoAuth_Calldata verifies the distinguishing behavior of the TEE path: the auth tx it
// submits is authenticateBatchInfo(commitment, sig) where sig is a 65-byte EIP-712 signature over
// the commitment that recovers to the batcher's key.
func TestEspressoAuth_Calldata(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	l := newAuthSubmitter(t)
	l.RollupConfig.L1ChainID = big.NewInt(1)
	l.teeVerifierAddress = common.HexToAddress("0x00000000000000000000000000000000000000bb")
	l.Config.Espresso.BatcherPrivateKey = key

	txdata := testAuthTxData()
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)},
			{Receipt: receiptWithBlock(101)},
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithEspresso(txdata, false, candidate, queue, receiptsCh)
	l.authGroup.Wait()

	require.Len(t, queue.sends, 2)
	require.NotNil(t, queue.sends[0].candidate.To)
	require.Equal(t, l.RollupConfig.BatchAuthenticatorAddress, *queue.sends[0].candidate.To)

	expectedCommitment, err := computeCommitment(candidate)
	require.NoError(t, err)
	commitment, signature := unpackAuthenticateBatchInfo(t, queue.sends[0].candidate.TxData)
	require.Equal(t, expectedCommitment, commitment)
	require.Len(t, signature, 65, "TEE path must send a 65-byte EIP-712 signature")

	// Reconstruct the EIP-712 digest the batcher signed and confirm the signature recovers to the
	// batcher's key.
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"EspressoTEEVerifier": []apitypes.Type{
				{Name: "commitment", Type: "bytes32"},
			},
		},
		PrimaryType: "EspressoTEEVerifier",
		Domain: apitypes.TypedDataDomain{
			Name:              "EspressoTEEVerifier",
			Version:           "1",
			ChainId:           (*math.HexOrDecimal256)(l.RollupConfig.L1ChainID),
			VerifyingContract: l.teeVerifierAddress.String(),
		},
		Message: map[string]interface{}{
			"commitment": commitment,
		},
	}
	digest, _, err := apitypes.TypedDataAndHash(typedData)
	require.NoError(t, err)

	// Denormalize v (27/28 -> 0/1) before recovering, undoing the Solidity-compat normalization.
	recoverSig := make([]byte, 65)
	copy(recoverSig, signature)
	require.Contains(t, []byte{27, 28}, recoverSig[64])
	recoverSig[64] -= 27

	recovered, err := crypto.SigToPub(digest, recoverSig)
	require.NoError(t, err)
	require.Equal(t, crypto.PubkeyToAddress(key.PublicKey), crypto.PubkeyToAddress(*recovered))
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
