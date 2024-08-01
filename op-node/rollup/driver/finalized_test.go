package driver

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

type finalizedTest struct {
	name  string
	final uint64
	hash  common.Hash // hash of finalized block
	req   uint64
	pass  bool
}

func (ft *finalizedTest) Run(t *testing.T) {
	l1Fetcher := &testutils.MockL1Source{}
	l1Finalized := eth.L1BlockRef{Number: ft.final, Hash: ft.hash}
	l1FinalizedGetter := func() eth.L1BlockRef { return l1Finalized }

	f := NewFinalized(l1FinalizedGetter, l1Fetcher)

	if ft.pass {
		// no calls to the l1Fetcher are made if the block number is not finalized yet
		l1Fetcher.ExpectL1BlockRefByNumber(ft.req, eth.L1BlockRef{Number: ft.req}, nil)
	}

	out, err := f.L1BlockRefByNumber(context.Background(), ft.req)
	l1Fetcher.AssertExpectations(t)

	if ft.pass {
		require.NoError(t, err)
		require.Equal(t, out, eth.L1BlockRef{Number: ft.req})
	} else {
		require.Equal(t, ethereum.NotFound, err)
	}
}

func TestFinalized(t *testing.T) {
	testCases := []finalizedTest{
		{name: "finalized", final: 10, hash: exHash, req: 10, pass: true},
		{name: "finalized past", final: 10, hash: exHash, req: 8, pass: true},
		{name: "not finalized", final: 10, hash: exHash, req: 11, pass: false},
		{name: "no L1 state", req: 10, pass: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, tc.Run)
	}
}
