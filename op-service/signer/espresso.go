package signer

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Sign represents the interface for signing things via eth_sign.
func (s *SignerClient) Sign(ctx context.Context, address common.Address, data []byte) ([]byte, error) {
	var result hexutil.Bytes
	if err := s.client.CallContext(ctx, &result, "eth_sign", address, data); err != nil {
		return nil, fmt.Errorf("eth_sign failed: %w", err)
	}
	return result, nil
}
