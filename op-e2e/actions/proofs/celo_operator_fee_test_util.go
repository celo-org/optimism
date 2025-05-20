package proofs

import (
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/proofs/helpers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/addresses"
)

// In the celo op-geth state transition function we issue the base fee to the fee handler
// if running in a cel2 context, otherwise it is issued to the base fee vault.
// We need to account for this here so that we can correctly account for all funds.
func getBaseFeeRecipientAddress(env *helpers.L2FaultProofEnv) common.Address {
	if env.Sd.L2Cfg.Config.IsCel2(env.Sequencer.L2Unsafe().Time) {
		return addresses.GetAddressesOrDefault(env.Sd.RollupCfg.L2ChainID, addresses.MainnetAddresses).FeeHandler
	}
	return predeploys.BaseFeeVaultAddr
}
