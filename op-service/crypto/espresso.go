package crypto

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	hdwallet "github.com/ethereum-optimism/go-ethereum-hdwallet"
	opsigner "github.com/ethereum-optimism/optimism/op-service/signer"
)

// ChainSignerFactory creates a SignerFn that is bound to a specific ChainID
type ChainSignerFactory func(chainID *big.Int, from common.Address) ChainSigner

// ChainSigner is a generic interface for signing transactions or arbitrary data.
type ChainSigner interface {

	// SignTransaction signs a transaction with the given address.
	SignTransaction(ctx context.Context, addr common.Address, tx *types.Transaction) (*types.Transaction, error)

	// Sign signs the hash of arbitrary data.
	Sign(ctx context.Context, hash []byte) ([]byte, error)
}

// clientSigner is a ChainSigner that utilizes a remote signer to perform
// Sign and SignTransaction
type clientSigner struct {
	signerClient *opsigner.SignerClient
	fromAddress  common.Address
	chainID      *big.Int
}

// Sign implements Signer.
func (c *clientSigner) Sign(ctx context.Context, data []byte) ([]byte, error) {
	return c.signerClient.Sign(ctx, c.fromAddress, data)
}

// VerifySignature verifies that the signature was produced by the expected address.
// data is the original message (e.g., txdata.CallData()) and signature is the result
// from eth_sign.
func Verify(data []byte, signature []byte, expected common.Address) error {

	pubKey, err := crypto.SigToPub(data, signature)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	// Convert the ecdsa.PublicKey to an Address
	address := crypto.PubkeyToAddress(*pubKey)

	// Ensure that the derived address matches the expected address.
	if !bytes.Equal(address.Bytes(), expected.Bytes()) {
		return fmt.Errorf("address mismatch: got %s, expected %s", address.Hex(), expected.Hex())
	}

	return nil
}

// SignTransaction implements Signer.
func (c *clientSigner) SignTransaction(ctx context.Context, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
	if !bytes.Equal(address[:], c.fromAddress[:]) {
		return nil, fmt.Errorf("attempting to sign for %s, expected %s: ", address, c.fromAddress)
	}
	return c.signerClient.SignTransaction(ctx, c.chainID, address, tx)
}

var _ ChainSigner = &clientSigner{}

// privateKeySigner is a ChainSigner that delegates to the stored
// functions for performing Sign and SignTransaction. In general these stored
// functions are expected to have access to a private key that is not
// explicitly stored within the structure itself.
type privateKeySigner struct {
	chainID     *big.Int
	st          bind.SignerFn
	fromAddress common.Address
	s           func(common.Address, []byte) ([]byte, error)
}

// Sign implements Signer.
func (p *privateKeySigner) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	return p.s(p.fromAddress, hash)
}

// SignTransaction implements Signer.
func (p *privateKeySigner) SignTransaction(ctx context.Context, addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
	return p.st(addr, tx)
}

var _ ChainSigner = &privateKeySigner{}

// ChainSignerFactoryFromConfig considers three ways that signers are created & then creates single factory from those config options.
// It can either take a remote signer (via opsigner.CLIConfig) or it can be provided either a mnemonic + derivation path or a private key.
// It prefers the remote signer, then the mnemonic or private key (only one of which can be provided).
func ChainSignerFactoryFromConfig(l log.Logger, privateKey, mnemonic, hdPath string, signerConfig opsigner.CLIConfig) (ChainSignerFactory, common.Address, error) {
	var signer ChainSignerFactory
	var fromAddress common.Address
	if signerConfig.Enabled() {
		signerClient, err := opsigner.NewSignerClientFromConfig(l, signerConfig)
		if err != nil {
			l.Error("Unable to create Signer Client", "error", err)
			return nil, common.Address{}, fmt.Errorf("failed to create the signer client: %w", err)
		}
		fromAddress = common.HexToAddress(signerConfig.Address)
		signer = func(chainID *big.Int, _ common.Address) ChainSigner {
			return &clientSigner{
				signerClient: signerClient,
				fromAddress:  fromAddress,
				chainID:      chainID,
			}
		}
	} else {
		var privKey *ecdsa.PrivateKey
		var err error

		if privateKey != "" && mnemonic != "" {
			return nil, common.Address{}, errors.New("cannot specify both a private key and a mnemonic")
		}
		if privateKey == "" {
			// Parse l2output wallet private key and L2OO contract address.
			wallet, err := hdwallet.NewFromMnemonic(mnemonic)
			if err != nil {
				return nil, common.Address{}, fmt.Errorf("failed to parse mnemonic: %w", err)
			}

			privKey, err = wallet.PrivateKey(accounts.Account{
				URL: accounts.URL{
					Path: hdPath,
				},
			})
			if err != nil {
				return nil, common.Address{}, fmt.Errorf("failed to create a wallet: %w", err)
			}
		} else {
			privKey, err = crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
			if err != nil {
				return nil, common.Address{}, fmt.Errorf("failed to parse the private key: %w", err)
			}
		}
		// we force the curve to Geth's instance, because Geth does an equality check in the nocgo version:
		// https://github.com/ethereum/go-ethereum/blob/723b1e36ad6a9e998f06f74cc8b11d51635c6402/crypto/signature_nocgo.go#L82
		privKey.PublicKey.Curve = crypto.S256()
		fromAddress = crypto.PubkeyToAddress(privKey.PublicKey)
		signer = func(chainID *big.Int, from common.Address) ChainSigner {
			s := PrivateKeySignerFn(privKey, chainID)
			return &privateKeySigner{
				chainID: chainID,
				st:      s,
				s: func(addr common.Address, hash []byte) ([]byte, error) {
					return crypto.Sign(hash, privKey)
				},
			}
		}
	}

	return signer, fromAddress, nil
}
