package finalized

import (
	"context"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type finalized struct {
	derive.L1Fetcher
	l1Finalized func() eth.L1BlockRef
}

func NewFinalized(l1Finalized func() eth.L1BlockRef, fetcher derive.L1Fetcher) *finalized {
	return &finalized{L1Fetcher: fetcher, l1Finalized: l1Finalized}
}

func (f *finalized) L1BlockRefByNumber(ctx context.Context, num uint64) (eth.L1BlockRef, error) {
	l1Finalized := f.l1Finalized()
	if num == 0 || num <= l1Finalized.Number {
		return f.L1Fetcher.L1BlockRefByNumber(ctx, num)
	}
	return eth.L1BlockRef{}, ethereum.NotFound
}

var _ derive.L1Fetcher = (*finalized)(nil)
