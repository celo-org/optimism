package derive_test

import (
	"bytes"
	"math/big"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	derive "github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	dtest "github.com/ethereum-optimism/optimism/op-node/rollup/derive/test"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

var defaultTestRollUpConfig = &rollup.Config{
	Genesis:   rollup.Genesis{L2: eth.BlockID{Number: 0}},
	L2ChainID: big.NewInt(1234),
}

// compareHash is a helper function that compares two hashes.
func compareHash(a, b common.Hash) int {
	if c := bytes.Compare(a[:], b[:]); c != 0 {
		return c
	}
	return 0
}

// compareTransaction is a helper function that compares two transactions
// by only inspecting their hashes.
func compareTransaction(a, b *gethTypes.Transaction) int {
	return compareHash(a.Hash(), b.Hash())
}

// compareHeader is a helper function that compares two headers
// by only inspecting their hashes.
func compareHeader(a, b *gethTypes.Header) int {
	return compareHash(a.Hash(), b.Hash())
}

// compareWithdrawl is a helper function that compares two withdrawals
// by checking that their slice members compare equivalently.
func compareWithdrawl(a, b *gethTypes.Withdrawal) int {
	if c := a.Index - b.Index; c != 0 {
		return int(c)
	}

	if c := a.Validator - b.Validator; c != 0 {
		return int(c)
	}

	if c := a.Address.Cmp(b.Address); c != 0 {
		return c
	}

	if c := a.Amount - b.Amount; c != 0 {
		return int(c)
	}

	return 0
}

// compareBody is a helper function that compares two bodies
// by checking that their slice members compare equivalently.
func compareBody(a, b *gethTypes.Body) int {
	if c := slices.CompareFunc(a.Transactions, b.Transactions, compareTransaction); c != 0 {
		return c
	}

	if c := slices.CompareFunc(a.Uncles, b.Uncles, compareHeader); c != 0 {
		return c
	}

	if c := slices.CompareFunc(a.Withdrawals, b.Withdrawals, compareWithdrawl); c != 0 {
		return c
	}

	return 0
}

// TestUnmarshalEspressoTransactionTooShort verifies that UnmarshalEspressoTransaction
// returns an error (rather than panicking) when the input is shorter than a signature.
func TestUnmarshalEspressoTransactionTooShort(t *testing.T) {
	cases := [][]byte{
		nil,
		{},
		make([]byte, crypto.SignatureLength-1),
	}
	for _, data := range cases {
		_, err := derive.UnmarshalEspressoTransaction(data)
		require.Error(t, err, "expected error for %d-byte input", len(data))
	}
}

// TestEspressoBatchConversion tests the conversion of a block to an Espresso
// Batch, and ensures that the recovery of the original Block is possible with
// the contents of the Espresso Batch.
func TestEspressoBatchConversion(t *testing.T) {
	rng := rand.New(rand.NewSource(4982432))
	ti := time.Now()

	originalBlock := dtest.RandomL2BlockWithChainIdAndTime(rng, rng.Intn(32), defaultTestRollUpConfig.L2ChainID, ti)

	espressoBatch, err := derive.BlockToEspressoBatch(defaultTestRollUpConfig, originalBlock)
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to convert block to batch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	decodedBlock, err := espressoBatch.ToBlock(defaultTestRollUpConfig)
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to decode batch back to block:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	// Let's perform a sanity check on the decoded block to ensure that all of
	// the fields match the original block.

	if have, want := decodedBlock.BaseFee(), originalBlock.BaseFee(); have.Cmp(want) != 0 {
		t.Errorf("decoded block base fee mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.BeaconRoot(), originalBlock.BeaconRoot(); have != want {
		t.Errorf("decoded block beacon root mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.BlobGasUsed(), originalBlock.BlobGasUsed(); have != want {
		t.Errorf("decoded block blob gas used mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Bloom(), originalBlock.Bloom(); have != want {
		t.Errorf("decoded block bloom mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Body(), originalBlock.Body(); compareBody(have, want) != 0 {
		t.Errorf("decoded block body mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Coinbase(), originalBlock.Coinbase(); have != want {
		t.Errorf("decoded block coinbase mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Difficulty(), originalBlock.Difficulty(); have.Cmp(want) != 0 {
		t.Errorf("decoded block difficulty mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.ExcessBlobGas(), originalBlock.ExcessBlobGas(); have != want {
		t.Errorf("decoded block excess blob gas mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.ExecutionWitness(), originalBlock.ExecutionWitness(); have != want {
		t.Errorf("decoded block execution witness mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Extra(), originalBlock.Extra(); !bytes.Equal(have, want) {
		t.Errorf("decoded block extra mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.GasLimit(), originalBlock.GasLimit(); have != want {
		t.Errorf("decoded block gas limit mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.GasUsed(), originalBlock.GasUsed(); have != want {
		t.Errorf("decoded block gas used mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Hash(), originalBlock.Hash(); have != want {
		t.Errorf("decoded block hash mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Header(), originalBlock.Header(); compareHeader(have, want) != 0 {
		t.Errorf("decoded block header mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.MixDigest(), originalBlock.MixDigest(); have != want {
		t.Errorf("decoded block mix digest mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Nonce(), originalBlock.Nonce(); have != want {
		t.Errorf("decoded block nonce mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Number(), originalBlock.Number(); have.Cmp(want) != 0 {
		t.Errorf("decoded block number mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.NumberU64(), originalBlock.NumberU64(); have != want {
		t.Errorf("decoded block number u64 mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.ParentHash(), originalBlock.ParentHash(); have != want {
		t.Errorf("decoded block parent hash mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.ReceiptHash(), originalBlock.ReceiptHash(); have != want {
		t.Errorf("decoded block receipt hash mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.RequestsHash(), originalBlock.RequestsHash(); have != want {
		t.Errorf("decoded block requests hash mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Root(), originalBlock.Root(); have != want {
		t.Errorf("decoded block root mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Size(), originalBlock.Size(); have != want {
		t.Errorf("decoded block size mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Time(), originalBlock.Time(); have != want {
		t.Errorf("decoded block time mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Transactions(), originalBlock.Transactions(); slices.CompareFunc(have, want, compareTransaction) != 0 {
		t.Errorf("decoded block transactions mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.TxHash(), originalBlock.TxHash(); have != want {
		t.Errorf("decoded block tx hash mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.UncleHash(), originalBlock.UncleHash(); have != want {
		t.Errorf("decoded block uncle hash mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.Withdrawals(), originalBlock.Withdrawals(); slices.CompareFunc(have, want, compareWithdrawl) != 0 {
		t.Errorf("decoded block withdrawals mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	if have, want := decodedBlock.WithdrawalsRoot(), originalBlock.WithdrawalsRoot(); have != want {
		t.Errorf("decoded block withdrawals root mismatch:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}
}
