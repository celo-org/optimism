package txmgr

import (
	"context"

	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// SignTransaction signs a transaction using the configured ChainSigner.
func (c *Config) SignTransaction(ctx context.Context, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
	return c.ChainSigner.SignTransaction(ctx, address, tx)
}

// Sign signs the hash of arbitrary data using the configured ChainSigner.
func (c *Config) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	return c.ChainSigner.Sign(ctx, hash)
}

// Ensure adherence to the interface
var _ opcrypto.ChainSigner = &Config{}

// SignTransaction signs a transaction using the underlying config's ChainSigner.
func (m *SimpleTxManager) SignTransaction(ctx context.Context, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
	return m.cfg.SignTransaction(ctx, address, tx)
}

// Sign signs the hash of arbitrary data using the underlying config's ChainSigner.
func (m *SimpleTxManager) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	return m.cfg.Sign(ctx, hash)
}

// Ensure adherence to the interface
var _ opcrypto.ChainSigner = &SimpleTxManager{}
