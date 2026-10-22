package derive

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type cgtRPCBlocks map[common.Hash]*cgtProofRPC

func (r cgtRPCBlocks) HeaderByHash(ctx context.Context, h common.Hash) (*types.Header, error) {
	block, ok := r[h]
	if !ok {
		return nil, errors.New("unknown header")
	}
	return block.HeaderByHash(ctx, h)
}
func (r cgtRPCBlocks) GetProof(ctx context.Context, a common.Address, slots []common.Hash, tag string) (*eth.AccountResult, error) {
	block, ok := r[common.HexToHash(tag)]
	if !ok {
		return nil, errors.New("unknown proof block")
	}
	return block.GetProof(ctx, a, slots, tag)
}

// The full derivation consumes real account/storage proofs and authenticated
// receipt contents. The bridge layout here is synthetic, not a deployment value.
func TestCgtV2AuthenticatedDerivation(t *testing.T) {
	for _, name := range []string{"valid", "forged amount", "forged receipts", "wrong packed offset", "transport failure"} {
		t.Run(name, func(t *testing.T) {
			rs := optypes.FromGethReceipts(types.Receipts{{Status: 1, Logs: []*types.Log{cgtFeatureLog(common.Address{0x33})}}})
			d, stub, parent, _ := cgtTestDeriver(t, rs)
			system := d.tracker.systemConfig
			bHeader, bProof := cgtProofFixture(t, system, []common.Hash{cgtV2FeatureSlot}, []uint64{0}, true)
			parent.L1Origin = eth.BlockID{Hash: bHeader.Hash(), Number: 10}
			_, hProof := cgtProofFixture(t, d.snapshot.bridge, []common.Hash{d.snapshot.amountSlot, d.snapshot.seededSlot}, []uint64{600, 1}, true)
			hHeader := cgtProofWorld(t, map[common.Address]*eth.AccountResult{d.snapshot.bridge: hProof})
			hHeader.Number = big.NewInt(11)
			hHeader.ParentHash = bHeader.Hash()
			hHeader.ReceiptHash = stub.info.ReceiptHash()
			origin := eth.BlockID{Hash: hHeader.Hash(), Number: 11}
			l2Header, l2Proof := cgtProofFixture(t, cgtV2BridgeL2, nil, nil, false)
			l2Header.Number = new(big.Int).SetUint64(parent.Number)
			parent.Hash = l2Header.Hash()
			l1 := cgtRPCBlocks{bHeader.Hash(): {header: bHeader, proof: bProof}, hHeader.Hash(): {header: hHeader, proof: hProof}}
			l2 := cgtRPCBlocks{parent.Hash: {header: l2Header, proof: l2Proof}}
			source := NewCgtV2RPCSource(l1, l2)
			d.source = source
			d.tracker.source = source
			switch name {
			case "forged amount":
				hProof.StorageProof[0].Value = hexutil.Big(*big.NewInt(601))
			case "forged receipts":
				rs[0].Logs[0].Data = []byte{1}
			case "wrong packed offset":
				d.snapshot.seededOffset = 1
			case "transport failure":
				l1[hHeader.Hash()].err = errors.New("offline")
			}
			txs, active, err := d.deposits(context.Background(), parent, origin, rs)
			if name != "valid" {
				require.Error(t, err)
				if name == "transport failure" {
					require.ErrorIs(t, err, ErrTemporary)
				}
				return
			}
			require.NoError(t, err)
			require.True(t, active)
			require.Len(t, txs, 1)
			tx, err := optypes.UnmarshalDepositTx(txs[0])
			require.NoError(t, err)
			require.Zero(t, new(big.Int).SetBytes(tx.Data[68:]).Cmp(big.NewInt(600)))
			require.Equal(t, origin.Hash.Bytes(), tx.Data[36:68])
		})
	}
}
