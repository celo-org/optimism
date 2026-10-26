package derive

import (
	"math/big"
	"testing"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func cgtEscrowLog(bridge, portal common.Address, amount int64) *types.Log {
	return &types.Log{Address: bridge, Topics: []common.Hash{crypto.Keccak256Hash([]byte("EscrowSeeded(address,uint256)")), common.BytesToHash(portal.Bytes())}, Data: common.LeftPadBytes(big.NewInt(amount).Bytes(), 32)}
}

func TestCgtV2EscrowEvent(t *testing.T) {
	bridge, portal := common.HexToAddress("0xcc"), common.HexToAddress("0x44")
	for _, name := range []string{"valid", "zero", "wrong bridge", "genesis seeding", "failed receipt", "wrong topic", "missing indexed portal", "short amount", "extra topic"} {
		t.Run(name, func(t *testing.T) {
			event := cgtEscrowLog(bridge, portal, 600)
			receipt := &types.Receipt{Status: 1, Logs: []*types.Log{event}}
			want, bad := int64(600), false
			switch name {
			case "zero":
				event.Data = make([]byte, 32)
				want = 0
			case "wrong bridge":
				event.Address = common.HexToAddress("0xee")
				want = -1
			case "genesis seeding":
				event.Topics[1] = common.HexToHash("0xaa")
				want = -1
			case "failed receipt":
				receipt.Status = 0
				want = -1
			case "wrong topic":
				event.Topics[0] = common.Hash{}
				want = -1
			case "missing indexed portal":
				event.Topics = event.Topics[:1]
				bad = true
			case "short amount":
				event.Data = event.Data[:31]
				bad = true
			case "extra topic":
				event.Topics = append(event.Topics, common.Hash{})
				bad = true
			}
			amount, err := cgtV2EscrowAmount(optypes.FromGethReceipts(types.Receipts{receipt}), bridge, portal)
			if bad {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if want < 0 {
				require.Nil(t, amount)
			} else {
				require.Zero(t, amount.Cmp(big.NewInt(want)))
			}
		})
	}
}
