package attributes

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// Characterize the existing safe/unsafe comparison: expected derivation
// already authenticates the signal, and consolidation must compare its position.
func TestCgtV2SafeAttributesRejectMissingOrMisplacedSignal(t *testing.T) {
	cfg := &rollup.Config{ChainOpConfig: defaultOpConfig}
	cfg.ActivateAtGenesis(forks.Jovian)
	for _, name := range []string{"valid", "missing", "misplaced", "wrong amount"} {
		t.Run(name, func(t *testing.T) {
			args := jovianArgs()
			signal, err := derive.CgtV2ActivationDeposit(common.Hash{1}, big.NewInt(600))
			require.NoError(t, err)
			raw, err := signal.MarshalBinary()
			require.NoError(t, err)
			// Opaque non-signal txs stand for the surrounding L1-info/deposits.
			before := args.attrs.Transactions[0]
			after := eth.Data{0xc0}
			args.attrs.Transactions = []eth.Data{before, raw, after}
			txs := []eth.Data{before, raw, after}
			switch name {
			case "missing":
				txs = []eth.Data{before, after}
			case "misplaced":
				txs = []eth.Data{before, after, raw}
			case "wrong amount":
				changed := append(eth.Data(nil), raw...)
				changed[len(changed)-1]++
				txs[1] = changed
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
