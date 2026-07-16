package derive

import (
	"bytes"
	"context"
	"fmt"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	opCrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// A SingularBatch with block number attached to restore ordering
// when fetching from Espresso
type EspressoBatch struct {
	BatchHeader   *types.Header
	Batch         SingularBatch
	L1InfoDeposit *types.Transaction
	SignerAddress common.Address
}

func (b EspressoBatch) Number() uint64 {
	return bigs.Uint64Strict(b.BatchHeader.Number)
}

func (b EspressoBatch) L1Origin() eth.BlockID {
	return b.Batch.Epoch()
}

func (b EspressoBatch) Header() *types.Header {
	return b.BatchHeader
}

func (b EspressoBatch) Hash() common.Hash {
	hash := crypto.Keccak256Hash(b.BatchHeader.Hash().Bytes(), b.L1InfoDeposit.Hash().Bytes())
	return hash
}

func (b EspressoBatch) Signer() common.Address {
	return b.SignerAddress
}

func (b *EspressoBatch) ToEspressoTransaction(ctx context.Context, namespace uint64, signer opCrypto.ChainSigner) (*espressoCommon.Transaction, error) {
	buf := new(bytes.Buffer)
	err := rlp.Encode(buf, *b)
	if err != nil {
		return nil, fmt.Errorf("failed to encode batch: %w", err)
	}

	batcherSignature, err := signer.Sign(ctx, crypto.Keccak256(buf.Bytes()))

	if err != nil {
		return nil, fmt.Errorf("failed to create batcher signature: %w", err)
	}

	payload := append(batcherSignature, buf.Bytes()...)

	return &espressoCommon.Transaction{Namespace: namespace, Payload: payload}, nil

}

func BlockToEspressoBatch(rollupCfg *rollup.Config, block *types.Block) (*EspressoBatch, error) {
	if len(block.Transactions()) == 0 {
		return nil, fmt.Errorf("Block doesn't contain any transactions")
	}

	l1InfoDeposit := block.Transactions()[0]
	if !l1InfoDeposit.IsDepositTx() {
		return nil, fmt.Errorf("First transaction is not L1 info deposit")
	}

	batch, _, err := BlockToSingularBatch(rollupCfg, block)
	if err != nil {
		return nil, err
	}

	return &EspressoBatch{
		BatchHeader:   block.Header(),
		Batch:         *batch,
		L1InfoDeposit: l1InfoDeposit,
	}, nil
}

// CreateEspressoBatchUnmarshaler returns a function that can be used to
// unmarshal an Espresso transaction into an EspressoBatch.
// The signer address is recovered from the signature and stored on the batch
// for later verification in CheckBatch (two-phase verification).
func CreateEspressoBatchUnmarshaler() func(data []byte) (*EspressoBatch, error) {
	return func(data []byte) (*EspressoBatch, error) {
		return UnmarshalEspressoTransaction(data)
	}
}

func UnmarshalEspressoTransaction(data []byte) (*EspressoBatch, error) {
	if len(data) < crypto.SignatureLength {
		return nil, fmt.Errorf("transaction data too short: %d bytes, need at least %d", len(data), crypto.SignatureLength)
	}
	signatureData, batchData := data[:crypto.SignatureLength], data[crypto.SignatureLength:]
	batchHash := crypto.Keccak256(batchData)

	signerKey, err := crypto.SigToPub(batchHash, signatureData)
	if err != nil {
		return nil, err
	}
	signer := crypto.PubkeyToAddress(*signerKey)

	var batch EspressoBatch
	if err := rlp.DecodeBytes(batchData, &batch); err != nil {
		return nil, err
	}
	batch.SignerAddress = signer

	if batch.BatchHeader == nil || batch.BatchHeader.Number == nil {
		return nil, fmt.Errorf("batch header is missing a block number")
	}
	if !batch.BatchHeader.Number.IsUint64() {
		return nil, fmt.Errorf("batch header number %s does not fit in uint64", batch.BatchHeader.Number)
	}
	if batch.L1InfoDeposit == nil {
		return nil, fmt.Errorf("batch is missing the L1 info deposit transaction")
	}

	return &batch, nil
}

// NOTE: This function MUST guarantee no transient errors. It is allowed to fail only on
// invalid batches or in case of misconfiguration of the batcher, in which case it should fail
// for all batches.
func (b *EspressoBatch) ToBlock(rollupCfg *rollup.Config) (*types.Block, error) {
	// The produced block must round-trip through BlockToSingularBatch when the channel
	// manager encodes it, so enforce that function's requirements up front, plus
	// consistency between the header and the batch body it claims to describe.
	if b.BatchHeader == nil {
		return nil, fmt.Errorf("batch has no header")
	}
	if b.L1InfoDeposit == nil || !b.L1InfoDeposit.IsDepositTx() {
		return nil, fmt.Errorf("first transaction is not an L1 info deposit")
	}
	l1Info, err := L1BlockInfoFromBytes(rollupCfg, b.BatchHeader.Time, b.L1InfoDeposit.Data())
	if err != nil {
		return nil, fmt.Errorf("could not parse the L1 info deposit: %w", err)
	}
	if b.Batch.ParentHash != b.BatchHeader.ParentHash {
		return nil, fmt.Errorf("batch parent hash %s does not match header parent hash %s", b.Batch.ParentHash, b.BatchHeader.ParentHash)
	}
	if b.Batch.Timestamp != b.BatchHeader.Time {
		return nil, fmt.Errorf("batch timestamp %d does not match header timestamp %d", b.Batch.Timestamp, b.BatchHeader.Time)
	}
	if uint64(b.Batch.EpochNum) != l1Info.Number || b.Batch.EpochHash != l1Info.BlockHash {
		return nil, fmt.Errorf("batch epoch %d (%s) does not match L1 info deposit epoch %d (%s)",
			b.Batch.EpochNum, b.Batch.EpochHash, l1Info.Number, l1Info.BlockHash)
	}

	// Re-insert the deposit transaction
	txs := []*types.Transaction{b.L1InfoDeposit}
	for i, opaqueTx := range b.Batch.Transactions {
		var tx types.Transaction
		err := tx.UnmarshalBinary(opaqueTx)
		if err != nil {
			return nil, fmt.Errorf("could not decode tx %d: %w", i, err)
		}
		txs = append(txs, &tx)
	}
	return types.NewBlockWithHeader(b.BatchHeader).WithBody(types.Body{
		Transactions: txs,
	}), nil
}
