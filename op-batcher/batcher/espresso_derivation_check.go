package batcher

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// errBatchUndecided means the block can't be checked yet (e.g. the L1 block it
// adopts as its origin isn't visible from our L1 endpoint yet). The block must
// be retried later, not signed.
var errBatchUndecided = errors.New("batch validity undecided, retry later")

// ErrBatchRejected means the block violates a derivation batch rule and must
// never be signed.
type ErrBatchRejected struct {
	Validity derive.BatchValidity
}

func (e *ErrBatchRejected) Error() string {
	return fmt.Sprintf("batch violates derivation rules (validity: %s)", batchValidityName(e.Validity))
}

func batchValidityName(v derive.BatchValidity) string {
	switch v {
	case derive.BatchDrop:
		return "drop"
	case derive.BatchAccept:
		return "accept"
	case derive.BatchUndecided:
		return "undecided"
	case derive.BatchFuture:
		return "future"
	case derive.BatchPast:
		return "past"
	default:
		return fmt.Sprintf("unknown(%d)", v)
	}
}

// l1HeaderFetcher is the subset of L1Client needed to check derivation rules.
type l1HeaderFetcher interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

// checkDerivationRules runs the same batch checks as op-node's derivation
// pipeline on batch, treating parent as the safe head and the current L1 head
// as the L1 inclusion block. It returns nil if the batch would be accepted,
// errBatchUndecided if it can't be decided yet, ErrBatchRejected if it
// violates a rule, or a fetch error.
func checkDerivationRules(ctx context.Context, cfg *rollup.Config, lgr log.Logger, l1 l1HeaderFetcher,
	timeout time.Duration, parent eth.L2BlockRef, batch *derive.SingularBatch,
) error {
	fetch := func(number *big.Int) (*types.Header, error) {
		cCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return l1.HeaderByNumber(cCtx, number)
	}

	l1Head, err := fetch(nil)
	if err != nil {
		return fmt.Errorf("fetching L1 head: %w", err)
	}

	// The parent's L1 origin is the current epoch. It must be canonical on our
	// L1 endpoint, otherwise we can't tell which successor the batch may adopt.
	originHeader, err := fetch(new(big.Int).SetUint64(parent.L1Origin.Number))
	if err != nil {
		return fmt.Errorf("fetching parent L1 origin %d: %w", parent.L1Origin.Number, err)
	}
	if originHeader.Hash() != parent.L1Origin.Hash {
		lgr.Warn("Parent's L1 origin is not canonical on L1",
			"parent", parent.ID(), "l1Origin", parent.L1Origin, "canonicalHash", originHeader.Hash())
		return errBatchUndecided
	}
	l1Blocks := []eth.L1BlockRef{eth.InfoToL1BlockRef(eth.HeaderBlockInfo(originHeader))}

	// The successor is only needed if the batch advances the epoch or exceeds
	// the sequencer drift. If it isn't available yet, CheckBatch returns
	// BatchUndecided when it needs it.
	nextHeader, err := fetch(new(big.Int).SetUint64(parent.L1Origin.Number + 1))
	switch {
	case errors.Is(err, ethereum.NotFound):
	case err != nil:
		return fmt.Errorf("fetching L1 block %d: %w", parent.L1Origin.Number+1, err)
	case nextHeader.ParentHash != originHeader.Hash():
		lgr.Warn("L1 reorg while checking batch", "l1Origin", parent.L1Origin, "next", nextHeader.Number)
		return errBatchUndecided
	default:
		l1Blocks = append(l1Blocks, eth.InfoToL1BlockRef(eth.HeaderBlockInfo(nextHeader)))
	}

	// Singular batches never need the L2 fetcher, which is only used for span batches.
	validity := derive.CheckBatch(ctx, cfg, lgr, l1Blocks, parent, &derive.BatchWithL1InclusionBlock{
		Batch:            batch,
		L1InclusionBlock: eth.InfoToL1BlockRef(eth.HeaderBlockInfo(l1Head)),
	}, nil)
	switch validity {
	case derive.BatchAccept:
		return nil
	case derive.BatchUndecided:
		return errBatchUndecided
	default:
		return &ErrBatchRejected{Validity: validity}
	}
}
