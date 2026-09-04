package crypto

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func PrivateKeySignerFn(key *ecdsa.PrivateKey, chainID *big.Int) bind.SignerFn {
	from := crypto.PubkeyToAddress(key.PublicKey)
	signer := types.LatestSignerForChainID(chainID)
	return func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != from {
			return nil, bind.ErrNotAuthorized
		}
		signature, err := crypto.Sign(signer.Hash(tx).Bytes(), key)
		if err != nil {
			return nil, err
		}
		return tx.WithSignature(signer, signature)
	}
}

func SignerFnFromBind(fn bind.SignerFn) SignerFn {
	return func(_ context.Context, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		return fn(address, tx)
	}
}

// SignerFn is a generic transaction signing function. It may be a remote signer so it takes a context.
// It also takes the address that should be used to sign the transaction with.
type SignerFn func(context.Context, common.Address, *types.Transaction) (*types.Transaction, error)
