package derive

import (
	"context"
	"testing"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestCgtV2EscrowAppendsGroupAfterAllOriginDeposits(t *testing.T) {
	portal := common.Address{0x44}
	receipts := optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{
		cgtPortalLog(t, portal, 1), cgtEscrowLog(common.HexToAddress("0xcc"), portal, 600), cgtPortalLog(t, portal, 2),
	}}})
	d, _, parent, origin := cgtTestDeriver(t, receipts)
	_, deposits, upgrades, err := d.originDeposits(context.Background(), parent, origin, receipts)
	txs := append(deposits, upgrades...)
	require.NoError(t, err)
	require.NotEmpty(t, upgrades)
	require.Len(t, txs, 15)
	for i, index := range []uint64{0, 2} {
		tx, err := optypes.UnmarshalDepositTx(txs[i])
		require.NoError(t, err)
		require.Equal(t, (&UserDepositSource{L1BlockHash: origin.Hash, LogIndex: index}).SourceHash(), tx.SourceHash)
	}
	require.False(t, containsCgtV2Group(txs[:2]))
	require.True(t, containsCgtV2Group(txs[2:]))
}
