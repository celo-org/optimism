package derive

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// CgtV2InputSource supplies untrusted header bytes. No account/storage proof or
// L2 completion RPC is required. HeaderByHash normally reuses the L1 client's cache.
type CgtV2InputSource interface {
	HeaderByHash(context.Context, common.Hash) (*types.Header, error)
}

func cgtV2Header(ctx context.Context, rpc CgtV2InputSource, block eth.BlockID) (*types.Header, error) {
	if rpc == nil {
		return nil, NewCriticalError(fmt.Errorf("CGTv2 requires L1 header RPC support"))
	}
	header, err := rpc.HeaderByHash(ctx, block.Hash)
	if err != nil {
		return nil, NewTemporaryError(fmt.Errorf("CGTv2 header %s: %w", block, err))
	}
	if header == nil || header.Number == nil || !header.Number.IsUint64() || bigs.Uint64Strict(header.Number) != block.Number || header.Hash() != block.Hash {
		return nil, NewCriticalError(fmt.Errorf("CGTv2 header does not match requested block %s", block))
	}
	return header, nil
}
