package txmgr

import (
	"context"

	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// SignTransaction is a function that provides the ability to sign a transaction
func (m *SimpleTxManager) SignTransaction(ctx context.Context, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
	return m.cfg.ChainSigner.SignTransaction(ctx, address, tx)
}

// Sign is a function that provides the ability to sign a hash
func (m *SimpleTxManager) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	return m.cfg.ChainSigner.Sign(ctx, hash)
}

// Ensure adherence to the interface
var _ opcrypto.ChainSigner = &SimpleTxManager{}
