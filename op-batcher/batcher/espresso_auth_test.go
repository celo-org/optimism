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

func newAuthSubmitter(t *testing.T) *BatchSubmitter {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.RollupConfig = &rollup.Config{
		BatchAuthenticatorAddress: common.HexToAddress("0x00000000000000000000000000000000000000aa"),
	}
	return l
}

func testAuthTxData(t *testing.T) txData {
	return singleFrameTxData(frameData{data: []byte("frame-data")})
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

// TestSubmitAuthenticatedBatch_OrderingAndSuccess verifies the auth tx is submitted before the
// batch tx (so it takes the lower nonce and lands first, as required under Holocene) and that a
// single success receipt for the batch txData is emitted when both txs land within the lookback
// window.
func TestSubmitAuthenticatedBatch_OrderingAndSuccess(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData(t)
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
	require.NoError(t, l.authGroup.Wait())

	require.Len(t, queue.sends, 2)
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
	txdata := testAuthTxData(t)

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Err: errSendFailed},             // auth fails to send
			{Receipt: receiptWithBlock(101)}, // batch lands anyway
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.submitAuthenticatedBatch(authTxRef(txdata), txmgr.TxCandidate{}, &txmgr.TxCandidate{}, queue, receiptsCh)
	require.NoError(t, l.authGroup.Wait())

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

func TestSubmitAuthenticatedBatch_BatchSendFailureRetried(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData(t)

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)}, // auth lands
			{Err: errSendFailed},             // batch fails to send
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.submitAuthenticatedBatch(authTxRef(txdata), txmgr.TxCandidate{}, &txmgr.TxCandidate{}, queue, receiptsCh)
	require.NoError(t, l.authGroup.Wait())

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

// TestSubmitAuthenticatedBatch_AuthRevertedRetried verifies that an authenticateBatchInfo tx that
// mines but reverts (no event emitted for the verifier) produces an error receipt so the frames
// are re-queued, rather than being confirmed as success.
func TestSubmitAuthenticatedBatch_AuthRevertedRetried(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData(t)

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: revertedReceiptWithBlock(100)}, // auth mined but reverted
			{Receipt: receiptWithBlock(101)},         // batch lands
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.submitAuthenticatedBatch(authTxRef(txdata), txmgr.TxCandidate{}, &txmgr.TxCandidate{}, queue, receiptsCh)
	require.NoError(t, l.authGroup.Wait())

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
}

// TestSubmitAuthenticatedBatch_WindowViolationRetried verifies that a batch tx landing outside the
// lookback window of the auth tx produces an error receipt (so the channel manager rewinds and
// resubmits), rather than being confirmed.
func TestSubmitAuthenticatedBatch_WindowViolationRetried(t *testing.T) {
	l := newAuthSubmitter(t)
	txdata := testAuthTxData(t)

	tooFar := int64(100 + derive.BatchAuthLookbackWindow)
	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)},    // auth
			{Receipt: receiptWithBlock(tooFar)}, // batch too far away
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.submitAuthenticatedBatch(authTxRef(txdata), txmgr.TxCandidate{}, &txmgr.TxCandidate{}, queue, receiptsCh)
	require.NoError(t, l.authGroup.Wait())

	got := <-receiptsCh
	require.Error(t, got.Err)
	require.Equal(t, txdata.ID().String(), got.ID.id.String())
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
	txdata := testAuthTxData(t)
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)},
			{Receipt: receiptWithBlock(101)},
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithFallbackAuth(txdata, false, candidate, queue, receiptsCh)
	require.NoError(t, l.authGroup.Wait())

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

	txdata := testAuthTxData(t)
	candidate := &txmgr.TxCandidate{TxData: []byte("batch-calldata")}

	queue := &fakeTxSender{
		responses: []txmgr.TxReceipt[txRef]{
			{Receipt: receiptWithBlock(100)},
			{Receipt: receiptWithBlock(101)},
		},
	}
	receiptsCh := make(chan txmgr.TxReceipt[txRef], 1)

	l.sendTxWithEspresso(txdata, false, candidate, queue, receiptsCh)
	require.NoError(t, l.authGroup.Wait())

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
