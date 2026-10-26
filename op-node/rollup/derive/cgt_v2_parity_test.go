package derive

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"testing"

	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

func TestCgtV2FullOriginMatchesNativeProofClient(t *testing.T) {
	raw, err := os.ReadFile("testdata/cgtv2-origin.json")
	require.NoError(t, err)
	var vector struct {
		HeaderRlp    hexutil.Bytes   `json:"headerRlp"`
		ReceiptRlp   hexutil.Bytes   `json:"receiptRlp"`
		ParentL2Hash common.Hash     `json:"parentL2Hash"`
		Transactions []hexutil.Bytes `json:"transactions"`
		GasLimit     uint64          `json:"gasLimit"`
	}
	require.NoError(t, json.Unmarshal(raw, &vector))
	var header types.Header
	var receipt types.Receipt
	require.NoError(t, rlp.DecodeBytes(vector.HeaderRlp, &header))
	require.NoError(t, rlp.DecodeBytes(vector.ReceiptRlp, &receipt))
	cfg := cgtTestConfig()
	cfg.DepositContractAddress = common.HexToAddress("0x44")
	cfg.L1SystemConfigAddress = common.Address{}
	zero := uint64(0)
	cfg.RegolithTime, cfg.CanyonTime, cfg.DeltaTime = &zero, &zero, &zero
	cfg.EcotoneTime, cfg.FjordTime, cfg.GraniteTime = &zero, &zero, &zero
	info := eth.HeaderBlockInfo(&header)
	source := &cgtInputStub{header: &header, info: info}
	receipts := optypes.FromGethReceipts(types.Receipts{&receipt})
	parent := eth.L2BlockRef{Hash: vector.ParentL2Hash, Time: 100, L1Origin: eth.BlockID{Hash: header.ParentHash, Number: 10}}
	builder := NewFetchingAttributesBuilder(cfg, params.MainnetChainConfig, nil, &cgtL1Stub{info, receipts}, cgtL2Stub{}, source)
	attrs, err := builder.PreparePayloadAttributes(context.Background(), parent, eth.BlockID{Hash: header.Hash(), Number: 11})
	require.NoError(t, err)
	require.Equal(t, vector.Transactions, attrs.Transactions)
	require.EqualValues(t, vector.GasLimit, *attrs.GasLimit)
	require.Len(t, attrs.Transactions, 16)
	for _, index := range []int{1, 2} {
		deposit, err := optypes.UnmarshalDepositTx(attrs.Transactions[index])
		require.NoError(t, err)
		if index == 1 {
			require.Zero(t, deposit.Mint.Cmp(big.NewInt(5)))
		} else {
			require.Zero(t, deposit.Mint.Sign())
		}
	}
}
