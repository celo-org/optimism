package driver

import (
	"context"
	"testing"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// TestOriginSelectorSameFinalized ensures that the origin selector
// advances only to the finalized block when Cel2 is enabled.
func TestOriginSelectorSameFinalized(t *testing.T) {
	log := testlog.Logger(t, log.LevelCrit)
	cfg := &rollup.Config{
		MaxSequencerDrift: 500,
		BlockTime:         2,
		Cel2Time:          new(uint64),
	}
	l1 := &testutils.MockL1Source{}
	defer l1.AssertExpectations(t)
	a := eth.L1BlockRef{
		Hash:   common.Hash{'a'},
		Number: 10,
		Time:   20,
	}
	l2Head := eth.L2BlockRef{
		L1Origin: a.ID(),
		Time:     24,
	}

	l1.ExpectL1BlockRefByHash(a.Hash, a, nil)
	l1.ExpectL1BlockRefByLabel(eth.Finalized, a, nil)

	s := NewL1OriginSelector(log, cfg, l1)
	next, err := s.FindL1Origin(context.Background(), l2Head)
	require.Nil(t, err)
	require.Equal(t, a, next)
}

// TestOriginSelectorNewFinalized ensures that the origin selector
// advances by only one block when a new finalized block is available.
func TestOriginSelectorNewFinalized(t *testing.T) {
	log := testlog.Logger(t, log.LevelCrit)
	cfg := &rollup.Config{
		MaxSequencerDrift: 500,
		BlockTime:         2,
		Cel2Time:          new(uint64),
	}
	l1 := &testutils.MockL1Source{}
	defer l1.AssertExpectations(t)
	a := eth.L1BlockRef{
		Hash:   common.Hash{'a'},
		Number: 10,
		Time:   20,
	}
	b := eth.L1BlockRef{
		Hash:       common.Hash{'b'},
		Number:     11,
		Time:       25,
		ParentHash: a.Hash,
	}
	c := eth.L1BlockRef{
		Hash:       common.Hash{'c'},
		Number:     12,
		Time:       30,
		ParentHash: b.Hash,
	}
	l2Head := eth.L2BlockRef{
		L1Origin: a.ID(),
		Time:     24,
	}

	l1.ExpectL1BlockRefByHash(a.Hash, a, nil)
	l1.ExpectL1BlockRefByNumber(b.Number, b, nil)
	l1.ExpectL1BlockRefByLabel(eth.Finalized, c, nil)

	s := NewL1OriginSelector(log, cfg, l1)
	next, err := s.FindL1Origin(context.Background(), l2Head)
	require.Nil(t, err)
	require.Equal(t, b, next)
}
