package crypto

import (
	"testing"

	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/signer"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

// should be run with CGO_ENABLED=0

func TestVerify(t *testing.T) {
	// happy path
	batcherSignature := []byte{
		109, 206, 105, 108, 152, 110, 156, 111, 239, 153, 224, 182, 140, 49, 105, 120,
		153, 163, 162, 47, 119, 34, 68, 128, 118, 33, 143, 79, 101, 212, 75, 161,
		124, 77, 236, 159, 70, 167, 95, 51, 92, 127, 236, 253, 4, 211, 222, 117,
		54, 27, 214, 232, 135, 87, 33, 77, 16, 155, 164, 116, 220, 116, 31, 208, 1,
	}
	sequencerBatchesByte := []byte{
		166, 136, 91, 55, 49, 112, 45, 166,
		46, 142, 74, 143, 88, 74, 196, 106,
		127, 104, 34, 244, 226, 186, 80, 251,
		169, 2, 246, 123, 21, 136, 210, 59,
	}

	expected := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	err := Verify(sequencerBatchesByte, batcherSignature, expected)
	require.NoError(t, err)

	// wrong length batcher signature
	wrongLengthBatcherSignature := []byte{
		1,
	}
	err = Verify(sequencerBatchesByte, wrongLengthBatcherSignature, expected)
	// check it returns an correct error: address mismatch
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to recover public key: invalid signature length")

	// wrong batcher signature
	wrongBatcherSignature := []byte{
		1, 1, 1, 1, 152, 110, 156, 111, 239, 153, 224, 182, 140, 49, 105, 120,
		153, 163, 162, 47, 119, 34, 68, 128, 118, 33, 143, 79, 101, 212, 75, 161,
		124, 77, 236, 159, 70, 167, 95, 51, 92, 127, 236, 253, 4, 211, 222, 117,
		54, 27, 214, 232, 135, 87, 33, 77, 16, 155, 164, 116, 220, 116, 31, 208, 1,
	}
	err = Verify(sequencerBatchesByte, wrongBatcherSignature, expected)
	// check it returns an correct error: address mismatch
	require.Error(t, err)
	require.Contains(t, err.Error(), "address mismatch")

	// wrong expected address
	wrongExpected := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C9")
	err = Verify(sequencerBatchesByte, batcherSignature, wrongExpected)
	require.Error(t, err)
	require.Contains(t, err.Error(), "address mismatch")

}

func TestChainSignerFactoryFromMnemonic(t *testing.T) {
	mnemonic := "test test test test test test test test test test test junk"
	hdPath := "m/44'/60'/0'/0/1"
	testChainSignerSignTransaction(t, "", mnemonic, hdPath, signer.CLIConfig{})
	testChainSignerSign(t, "", mnemonic, hdPath, signer.CLIConfig{})
}

func TestChainSignerFactoryFromKey(t *testing.T) {
	priv := "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	testChainSignerSignTransaction(t, priv, "", "", signer.CLIConfig{})
	testChainSignerSign(t, priv, "", "", signer.CLIConfig{})
}

func testChainSignerSignTransaction(t *testing.T, priv, mnemonic, hdPath string, cfg signer.CLIConfig) {
	logger := testlog.Logger(t, log.LevelDebug)

	factoryFn, addr, err := ChainSignerFactoryFromConfig(logger, priv, mnemonic, hdPath, cfg)
	require.NoError(t, err)
	expectedAddr := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	require.Equal(t, expectedAddr, addr)
	chainID := big.NewInt(10)
	chainSigner := factoryFn(chainID, addr) // for chain ID 10
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     0,
		GasTipCap: big.NewInt(1),
		GasFeeCap: big.NewInt(1),
		Gas:       21000,
		To:        nil,
		Value:     big.NewInt(0),
		Data:      []byte("test"),
	})
	signedTx, err := chainSigner.SignTransaction(context.Background(), addr, tx)
	require.NoError(t, err)
	gethSigner := types.LatestSignerForChainID(chainID)
	sender, err := gethSigner.Sender(signedTx)
	require.NoError(t, err)
	require.Equal(t, expectedAddr, sender)
}

func testChainSignerSign(t *testing.T, priv, mnemonic, hdPath string, cfg signer.CLIConfig) {
	logger := testlog.Logger(t, log.LevelDebug)

	factoryFn, addr, err := ChainSignerFactoryFromConfig(logger, priv, mnemonic, hdPath, cfg)
	require.NoError(t, err)
	expectedAddr := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	require.Equal(t, expectedAddr, addr)
	chainID := big.NewInt(10)
	chainSigner := factoryFn(chainID, addr) // for chain ID 10

	payload := []byte{0x01, 0x02, 0x03, 0x04}
	hash := crypto.Keccak256(payload)
	signed, err := chainSigner.Sign(context.Background(), hash)
	require.NoError(t, err)

	// Recover the public key from the signature and hash
	pubKey, err := crypto.SigToPub(hash, signed)
	require.NoError(t, err)

	// Convert the ecdsa.PublicKey to an Address
	address := crypto.PubkeyToAddress(*pubKey)

	// Ensure that the derived address matches the expected address.
	require.Equal(t, expectedAddr, address)
}
