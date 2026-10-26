package sequencing

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/rollup/engine"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type cgtAttributesBuilder struct {
	derive.AttributesBuilder
	active bool
}

func (b *cgtAttributesBuilder) PreparePayloadAttributes(ctx context.Context, p eth.L2BlockRef, h eth.BlockID) (*eth.PayloadAttributes, error) {
	attrs, err := b.AttributesBuilder.PreparePayloadAttributes(ctx, p, h)
	if err != nil || !b.active {
		return attrs, err
	}
	group, _, err := derive.CgtV2UpgradeTransactions(big.NewInt(1337), big.NewInt(600))
	if err != nil {
		return nil, err
	}
	attrs.Transactions = append(attrs.Transactions, group...)
	return attrs, nil
}

func TestCgtV2SequencerNoTxPool(t *testing.T) {
	s := newSeqSetup(t)
	builder := &cgtAttributesBuilder{AttributesBuilder: s.seq.attrBuilder, active: true}
	s.seq.attrBuilder = builder
	var noTxPool []bool
	s.deps.eng.startBuildFn = func(_ context.Context, a *derive.AttributesWithParent) (*engine.BuildStartResult, error) {
		noTxPool = append(noTxPool, a.Attributes.NoTxPool)
		return &engine.BuildStartResult{Info: eth.PayloadInfo{ID: eth.PayloadID{1}}, Parent: a.Parent, BuildStarted: s.clock.Now()}, nil
	}
	s.seq.RunAction()
	require.Equal(t, []bool{true}, noTxPool)
	// Next ordinary build resumes pool selection, without a sticky activation policy.
	builder.active = false
	s.seq.building = BuildingState{}
	s.seq.lastSealed = eth.L2BlockRef{}
	deliver(s.seq, engine.ForkchoiceUpdateEvent{UnsafeL2Head: eth.L2BlockRef{Hash: common.Hash{9}, Number: s.head.Number + 1, Time: s.head.Time + s.deps.cfg.BlockTime, L1Origin: s.l1Origin.ID()}})
	s.seq.RunAction()
	require.Equal(t, []bool{true, false}, noTxPool)
}
