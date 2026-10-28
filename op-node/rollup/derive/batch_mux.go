package derive

import (
	"context"
	"fmt"
	"slices"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/log"
)

type CgtV2ActivationDetector interface {
	CgtV2ActivationBlock(context.Context, eth.L2BlockRef, eth.BlockID) (bool, error)
}

// BatchMux multiplexes between different batch stages.
// Stages are swapped on demand during Reset calls, or explicitly with Transform.
// It currently chooses the BatchQueue pre-Holocene and the BatchStage post-Holocene.
type BatchMux struct {
	cgt               CgtV2ActivationDetector
	pending           *SingularBatch
	pendingConcluding bool
	log               log.Logger
	cfg               *rollup.Config
	prev              NextBatchProvider
	l2                SafeBlockFetcher

	// embedded active stage
	SingularBatchProvider
}

var _ SingularBatchProvider = (*BatchMux)(nil)

// NewBatchMux returns an uninitialized BatchMux. Reset has to be called before
// calling other methods, to activate the right stage for a given L1 origin.
func NewBatchMux(lgr log.Logger, cfg *rollup.Config, prev NextBatchProvider, l2 SafeBlockFetcher, cgt ...CgtV2ActivationDetector) *BatchMux {
	b := &BatchMux{log: lgr, cfg: cfg, prev: prev, l2: l2}
	if len(cgt) > 0 {
		b.cgt = cgt[0]
	}
	return b
}

func (b *BatchMux) Reset(ctx context.Context, base eth.L1BlockRef, sysCfg eth.SystemConfig) error {
	b.pending, b.pendingConcluding = nil, false
	switch {
	default:
		if _, ok := b.SingularBatchProvider.(*BatchQueue); !ok {
			b.log.Info("BatchMux: activating pre-Holocene stage during reset", "origin", base)
			b.SingularBatchProvider = NewBatchQueue(b.log, b.cfg, b.prev, b.l2)
		}
	case b.cfg.IsHolocene(base.Time):
		if _, ok := b.SingularBatchProvider.(*BatchStage); !ok {
			b.log.Info("BatchMux: activating Holocene stage during reset", "origin", base)
			b.SingularBatchProvider = NewBatchStage(b.log, b.cfg, b.prev, b.l2)
		}
	}
	return b.SingularBatchProvider.Reset(ctx, base, sysCfg)
}

func (b *BatchMux) Transform(f forks.Name) {
	switch f {
	case forks.Holocene:
		b.TransformHolocene()
	}
}

func (b *BatchMux) TransformHolocene() {
	b.pending, b.pendingConcluding = nil, false
	switch bp := b.SingularBatchProvider.(type) {
	case *BatchQueue:
		b.log.Info("BatchMux: transforming to Holocene stage")
		bs := NewBatchStage(b.log, b.cfg, b.prev, b.l2)
		// Even though any ongoing span batch or queued batches are dropped at Holocene activation, the
		// post-Holocene batch stage still needs access to the collected l1Blocks pre-Holocene because
		// the first Holocene channel will contain pre-Holocene batches.
		bs.l1Blocks = slices.Clone(bp.l1Blocks)
		bs.origin = bp.origin
		b.SingularBatchProvider = bs
	case *BatchStage:
		// Even if the pipeline is Reset to the activation block, the previous origin will be the
		// same, so transformStages isn't called.
		panic(fmt.Sprintf("Holocene BatchStage already active, old origin: %v", bp.Origin()))
	default:
		panic(fmt.Sprintf("unknown batch stage type: %T", bp))
	}
}

// NextBatch applies the origin-state rule only after the normal batch checks.
// Span batches have already been split, so a valid prefix is never invalidated
// just because a later block in the span is the activation block.
func (b *BatchMux) NextBatch(ctx context.Context, parent eth.L2BlockRef) (*SingularBatch, bool, error) {
	if b.pending == nil {
		batch, concluding, err := b.SingularBatchProvider.NextBatch(ctx, parent)
		if err != nil {
			return nil, false, err
		}
		b.pending, b.pendingConcluding = batch, concluding
	}
	batch, concluding := b.pending, b.pendingConcluding
	if b.cgt != nil && len(batch.Transactions) > 0 {
		active, err := b.cgt.CgtV2ActivationBlock(ctx, parent, batch.Epoch())
		if err != nil {
			// Keep the candidate across RPC retries.
			return nil, false, err
		}
		if active {
			var origins []eth.L1BlockRef
			var inclusion eth.L1BlockRef
			switch stage := b.SingularBatchProvider.(type) {
			case *BatchStage:
				origins, inclusion = stage.l1Blocks, stage.origin
			case *BatchQueue:
				origins, inclusion = stage.l1Blocks, stage.origin
			}
			if checkSingularBatch(b.cfg, b.log, origins, parent, batch, inclusion, true) != BatchAccept {
				b.pending, b.pendingConcluding = nil, false
				switch stage := b.SingularBatchProvider.(type) {
				case *BatchStage:
					stage.FlushChannel()
				case *BatchQueue:
					stage.nextSpan = nil
				}
				return nil, false, NotEnoughData
			}
		}
	}
	b.pending, b.pendingConcluding = nil, false
	return batch, concluding, nil
}

func (b *BatchMux) FlushChannel() {
	b.pending, b.pendingConcluding = nil, false
	b.SingularBatchProvider.FlushChannel()
}
