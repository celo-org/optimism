package derive

import (
	"bytes"
	"context"
	"fmt"

	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
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
	return b.BatchHeader.Number.Uint64()
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

	return &batch, nil
}

// NOTE: This function MUST guarantee no transient errors. It is allowed to fail only on
// invalid batches or in case of misconfiguration of the batcher, in which case it should fail
// for all batches.
func (b *EspressoBatch) ToBlock(rollupCfg *rollup.Config) (*types.Block, error) {
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
