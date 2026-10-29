package attributes

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// Characterize the existing safe/unsafe comparison: expected derivation
// already authenticates the group, and consolidation must compare every step.
func TestCgtV2SafeAttributesRejectMissingOrMisplacedGroup(t *testing.T) {
	cfg := &rollup.Config{ChainOpConfig: defaultOpConfig}
	cfg.ActivateAtGenesis(forks.Jovian)
	for _, name := range []string{"valid", "missing", "misplaced", "wrong amount"} {
		t.Run(name, func(t *testing.T) {
			args := jovianArgs()
			group, _, err := derive.CgtV2UpgradeTransactions(big.NewInt(1337), big.NewInt(600))
			require.NoError(t, err)
			// Origin deposits precede the complete upgrade group.
			before, deposit := args.attrs.Transactions[0], eth.Data{0xc0}
			args.attrs.Transactions = append([]eth.Data{before, deposit}, group...)
			txs := append([]eth.Data(nil), args.attrs.Transactions...)
			switch name {
			case "missing":
				txs = txs[:len(txs)-1]
			case "misplaced":
				txs[1], txs[2] = txs[2], txs[1]
			case "wrong amount":
				funding, err := optypes.UnmarshalDepositTx(txs[len(txs)-1])
				require.NoError(t, err)
				funding.Mint.Add(funding.Mint, big.NewInt(1))
				funding.Value.Add(funding.Value, big.NewInt(1))
				changed, err := funding.MarshalBinary()
				require.NoError(t, err)
				txs[len(txs)-1] = changed
			}
			args.envelope.ExecutionPayload.Transactions = txs
			err = AttributesMatchBlock(cfg, args.attrs, args.parentHash, args.envelope, testlog.Logger(t, log.LevelCrit))
			if name == "valid" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
