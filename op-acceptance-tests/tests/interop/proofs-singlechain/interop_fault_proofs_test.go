package proofs_singlechain

import (
	"testing"

	sfp "github.com/ethereum-optimism/optimism/op-acceptance-tests/tests/superfaultproofs"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
)

func TestInteropSingleChainFaultProofs(gt *testing.T) {
	gt.Skip("Skipped: fault proof program has no Celo support (cannot parse cel2_time in RollupConfig)")
	t := devtest.SerialT(gt)
	sys := presets.NewSingleChainInteropSupernodeProofs(t)
	sfp.RunSingleChainSuperFaultProofSmokeTest(t, sys)
}
