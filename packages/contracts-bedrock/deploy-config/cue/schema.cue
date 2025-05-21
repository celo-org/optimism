package deployconfig

// Type constraints
#Address:        string & =~"^0x[a-fA-F0-9]{40}$"
#Hex64:          string & =~"^0x[a-fA-F0-9]{64}$"
#Hex:            string & =~"^0x[a-fA-F0-9]*$"
#CommitmentType: "GenericCommitment" | "KeccakCommitment"

// Global vars
ZeroAddress:     "0x0000000000000000000000000000000000000000"
MinuteInSeconds: 60
HourInSeconds:   60 * MinuteInSeconds
DayInSeconds:    24 * HourInSeconds
gWei:            1000000000

#Superchain: {
	// indicates the protocol version that
	// nodes are required to adopt, to stay in sync with the network.
	requiredProtocolVersion: #Hex
	// indicates the protocol version that
	// nodes are recommended to adopt, to stay in sync with the network.
	recommendedProtocolVersion: #Hex
	// GUARDIAN account in the SuperchainConfig.
	// has the ability to pause withdrawals.
	superchainConfigGuardian: #Address

	protocolVersionsProxy?: #Address
	...
}

#Core: {
	// anchors the L2 at an L1 block.
	// The timestamp of the block referenced by l1StartingBlockTag is used
	// in the L2 genesis block, rollup-config, and L1 output-oracle contract.
	// The Output oracle deploy script may use it if the L2 starting timestamp is nil, assuming the L2 genesis is set up with this.
	// The L2 genesis timestamp does not affect the initial L2 account state:
	// the storage of the L1Block contract at genesis is zeroed, since the adoption of
	// the L2-genesis allocs-generation through solidity script.
	l1StartingBlockTag: #Hex64 | null
	// chain ID of the L1 chain.
	l1ChainID: uint64
	// chain ID of the L2 chain.
	l2ChainID: uint64 & !=l1ChainID
	// this is only used for a deploy-config sanity check and internally for generating
	// some config parameters with CUE so that we can set them in a more human-readable form.
	l1BlockTime: uint64 & >l2BlockTime
	// number of seconds between each L2 block.
	l2BlockTime: uint64 & >=1
	// number of seconds before an output is considered
	// finalized. This impacts the amount of time that withdrawals take to finalize and is
	// generally set to 1 week.
	// This is only relevant for the OutputOracle <v3, so for non-fault-proofs
	// The contracts are deployed, even when the fault-proof system is used,
	// so we need to set something here.
	finalizationPeriodSeconds: uint64
	// number of seconds after the L1 timestamp of the end of the
	// sequencing window that batches must be included, otherwise L2 blocks including
	// deposits are force included.
	// With Fjord, the MaxSequencerDrift becomes a constant. Use the ChainSpec
	// instead of reading this rollup configuration field directly to determine
	// the max sequencer drift for a given block based on the block's L1 origin.
	// Chains that activate Fjord at genesis may leave this field empty.
	maxSequencerDrift?: uint64 & >0
	// number of L1 blocks per sequencing window.
	sequencerWindowSize: uint64
	// number of L1 blocks that a frame stays valid when included in L1.
	// This is a pre-Granite parameter, for post-Granite, this
	// is hardcoded to 50 l1 blocks!
	channelTimeout: uint64
	// L1 account that batches are sent to.
	batchInboxAddress: #Address
	// block at which the op-node should start syncing
	// from. It is an override to set this value on legacy networks where it is not set by
	// default. It can be removed once all networks have this value set in their storage.
	// This is only relevant for older proxies that get upgraded, if this is set to
	// 0, the SystemConfig will set this to it's initialization block automatically.
	// The value itself is then read from the op-node from the contract storage.
	systemConfigStartBlock: uint64
	// address of the key the sequencer uses to sign blocks on the P2P layer
	p2pSequencerAddress: #Address
	// initial sequencer account that authorizes batches.
	// transactions sent from this account to the batch inbox address are considered valid.
	// this value is "initial" only since it will be set in the genesis blocks
	// system-config.
	batchSenderAddress: #Address
	// owner of the ProxyAdmin predeploy on L2.
	proxyAdminOwner: #Address
	// owner of the system on L1. Any L1 contract that is ownable has
	// this account set as its owner.
	finalSystemOwner: #Address

	// the base-fee-vault related values are passed to the fee-vault contract.
	// This contract is pre-deployed, but doesn't receive the fees in Cel2.

	// Recipient must not be the ZeroAddress as per the docs.
	baseFeeVaultRecipient: #Address
	//  minimum withdrawal amount for the BaseFeeVault
	baseFeeVaultMinimumWithdrawalAmount: #Hex
	// withdrawal network for the BaseFeeVault
	baseFeeVaultWithdrawalNetwork: uint64

	// the l1-fee vault related values are passed to the fee-vault contract.
	// This contract is pre-deployed, but doesn't receive the l1-fee in Cel2, because
	// we set our l1-fee scalars to zero.

	// recipient of fees accumulated in the L1FeeVault.
	// can be an account on L1 or L2, depending on the L1FeeVaultWithdrawalNetwork value
	// Recipient must not be the ZeroAddress as per the docs.
	l1FeeVaultRecipient: #Address
	// minimum withdrawal amount for the L1FeeVault
	l1FeeVaultMinimumWithdrawalAmount: #Hex
	// withdrawal network for the L1FeeVault
	l1FeeVaultWithdrawalNetwork: uint64

	// the sequencer-fee-vault related values are passed to the fee-vault contract.
	// This contract is pre-deployed, but it doesn't receive the fee-tip in Cel2.
	// The value is set in the genesis block as well as in the payload-attributes
	// as the Coinbase.

	// recipient of fees accumulated in the SequencerFeeVault.
	// can be an account on L1 or L2, depending on the SequencerFeeVaultWithdrawalNetwork value
	sequencerFeeVaultRecipient: #Address
	// minimum withdrawal amount for the SequencerFeeVault
	sequencerFeeVaultMinimumWithdrawalAmount: #Hex
	// withdrawal network for the SequencerFeeVault
	sequencerFeeVaultWithdrawalNetwork: uint64
	...
}

#EIP1559: {
	// denominator of EIP1559 base fee market.
	eip1559Denominator: uint64 & >0
	// elasticity of the EIP1559 fee market.
	eip1559Elasticity: uint64
	// denominator of EIP1559 base fee market when Canyon is active.
	eip1559DenominatorCanyon: uint64 & eip1559Denominator
	// EIP1559BaseFeeFloor fixed floor for the EIP1559 base fee market.
	eip1559BaseFeeFloor: uint64
	...
}

#Governance: {
	// configures whether or not include governance token predeploy
	enableGovernance: bool
	//  ERC20 symbol of the GovernanceToken
	governanceTokenName: "Optimism"
	// ERC20 name of the GovernanceToken
	governanceTokenSymbol: "OP"
	// owner of the GovernanceToken.
	// Has the ability to mint and burn tokens.
	governanceTokenOwner: #Address
	...
}

#GasPriceOracle: {
	// value of the base fee scalar used for fee calculations.
	// part of the system-config, so this takes effect in the genesis
	// block until it is overwritten by system events.
	gasPriceOracleBaseFeeScalar: uint64
	// value of the blob base fee scalar used for fee calculations.
	// part of the system-config, so this takes effect in the genesis
	// block until it is overwritten by system events.
	gasPriceOracleBlobBaseFeeScalar: uint64

	// initial value of the gas overhead in the GasPriceOracle predeploy.
	// deprecated: Since Ecotone, this field is superseded by gasPriceOracleBaseFeeScalar and GasPriceOracleBlobBaseFeeScalar.
	// part of the system-config, so this takes effect in the genesis
	// block until it is overwritten by system events.
	gasPriceOracleOverhead?: uint64
	// initial value of the gas scalar in the GasPriceOracle predeploy.
	// deprecated: Since Ecotone, this field is superseded by gasPriceOracleBaseFeeScalar and GasPriceOracleBlobBaseFeeScalar.
	// part of the system-config, so this takes effect in the genesis
	// block until it is overwritten by system events.
	gasPriceOracleScalar?: uint64
	...
}

#Dev: {
	// configures whether to fund the dev accounts.
	// This should only be used during devnet deployments.
	fundDevAccounts: bool
	// deploys the celo contracts within the l2 genesis alloction.
	// This should only be used during devnet deployments.
	deployCeloContracts: bool
	...
}

#Celo: {
	// TODO: do we want to deploy with the script,
	// or deploy in advance and configure the address
	//
	// flag to indicate that a custom gas token should be used
	useCustomGasToken: bool
	// address of the ERC20 token to be used to pay for gas on L2
	customGasTokenAddress:      #Address
	proxyAdminOwnerIsMultisig?: bool
	externalSuperchainConfig?:  #Address
	...
}

#DataAvailability: {
	// flag that indicates if the system is using op-alt-da
	useAltDA: bool
	// specifies the allowed commitment
	daCommitmentType: #CommitmentType
	// block interval during which the availability of a data commitment can be challenged.
	daChallengeWindow: uint64 & >0
	// block interval during which a data availability challenge can be resolved.
	daResolveWindow: uint64 & >0

	// TODO:
	// Javi: probably doesn't make sense to set this in permissioned fault proofs

	// percentage of the resolving cost to be refunded to the resolver
	// such as 100 means 100% refund.
	daResolverRefundPercentage?: uint8 & >=0 & <=100
	// required bond size to initiate a data availability challenge.
	daBondSize?: uint64
	// represents the L1 address of the DataAvailabilityChallenge contract.
	daChallengeProxy?: #Address
	...
}
#UpgradeSchedule: {
	// number of seconds after genesis block that Regolith hard fork activates.
	// Set it to 0 to activate at genesis. Nil to disable Regolith.
	l2GenesisRegolithTimeOffset?: #Hex
	// number of seconds after genesis block that Canyon hard fork activates.
	// Set it to 0 to activate at genesis. Nil to disable Canyon.
	l2GenesisCanyonTimeOffset?: #Hex
	// number of seconds after genesis block that Delta hard fork activates.
	// Set it to 0 to activate at genesis. Nil to disable Delta.
	l2GenesisDeltaTimeOffset?: #Hex
	// number of seconds after genesis block that Ecotone hard fork activates.
	// Set it to 0 to activate at genesis. Nil to disable Ecotone.
	l2GenesisEcotoneTimeOffset?: #Hex
	// number of seconds after genesis block that Fjord hard fork activates.
	// Set it to 0 to activate at genesis. Nil to disable Fjord.
	l2GenesisFjordTimeOffset?: #Hex
	// number of seconds after genesis block that Granite hard fork activates.
	// Set it to 0 to activate at genesis. Nil to disable Granite.
	l2GenesisGraniteTimeOffset?: #Hex
	// number of seconds after genesis block that the Holocene hard fork activates.
	// Set it to 0 to activate at genesis. Nil to disable Holocene.
	l2GenesisHoloceneTimeOffset?: #Hex | *null
	// number of seconds after genesis block that the Isthmus hard fork activates.
	// Set it to 0 to activate at genesis. Nil to disable Isthmus.
	l2GenesisIsthmusTimeOffset?: #Hex | *null
	// number of seconds after genesis block that the Interop hard fork activates.
	// Set it to 0 to activate at genesis. Nil to disable Interop.
	l2GenesisInteropTimeOffset?: #Hex
	// When Cancun activates. Relative to L1 genesis.
	// NOTE: probably set accoring to mainnet value
	l1CancunTimeOffset?: #Hex
	// When Prague activates. Relative to L1 genesis.
	// NOTE: probably set accoring to mainnet value
	l1PragueTimeOffset?: #Hex
	//  is a flag that indicates if the system is using interop
	useInterop?: bool
	...
}

#L2GenesisBlock: {
	// TODO: those values are currently all nulled or using a default, should we set any of those?
	// L2GenesisBlockNonce         hexutil.Uint64 `json:"l2GenesisBlockNonce"`
	// L2GenesisBlockDifficulty    *hexutil.Big   `json:"l2GenesisBlockDifficulty"`
	// L2GenesisBlockMixHash       common.Hash    `json:"l2GenesisBlockMixHash"`
	// L2GenesisBlockNumber        hexutil.Uint64 `json:"l2GenesisBlockNumber"`
	// L2GenesisBlockGasUsed       hexutil.Uint64 `json:"l2GenesisBlockGasUsed"`
	// L2GenesisBlockParentHash    common.Hash    `json:"l2GenesisBlockParentHash"`

	// part of the system-config, so this takes effect in the genesis
	// block until it is overwritten by system events.
	l2GenesisBlockGasLimit: #Hex
	// part of the system-config, so this takes effect in the genesis
	// block until it is overwritten by system events.
	l2GenesisBlockBaseFeePerGas: #Hex
	...
}

#OutputOracle: {
	// number of L2 blocks between outputs that are submitted
	// to the L2OutputOracle contract located on L1.
	l2OutputOracleSubmissionInterval: uint64
	// starting timestamp for the L2OutputOracle.
	// MUST be the same as the timestamp of the L2OO start block.
	// FIXME: this is somewhat problematic (at least for L2OO systems),
	// since the computeL2Time takes that as a starting time for l2 block time calculations.
	// However we previously set that based on the calculated Celo L1 migration block time,
	// although this has to be the actual timestamp of the first L2 block
	// specified in l2OutputOracleStartingBlockNumber
	// /// @notice Returns the L2 timestamp corresponding to a given L2 block number.
	// /// @param _l2BlockNumber The L2 block number of the target block.
	// /// @return L2 timestamp of the given block.
	// function computeL2Timestamp(uint256 _l2BlockNumber) public view returns (uint256) {
	//     return startingTimestamp + ((_l2BlockNumber - startingBlockNumber) * l2BlockTime);
	// }
	l2OutputOracleStartingTimestamp: uint64
	// starting block number for the L2OutputOracle.
	// Must be greater than or equal to the first Bedrock block.
	// The first L2 output will correspond to this value plus the submission interval.
	l2OutputOracleStartingBlockNumber: uint64
	// address of the account that proposes L2 outputs.
	l2OutputOracleProposer: #Address
	// address of the account that challenges L2 outputs.
	l2OutputOracleChallenger: #Address
	...
}

#FaultProofs: {
	// flag that indicates if the system is using fault
	// proofs instead of the older output oracle mechanism.
	useFaultProofs: bool
	// absolute prestate of Cannon. This is computed
	// by generating a proof from the 0th -> 1st instruction and grabbing the prestate from
	// the output JSON. All honest challengers should agree on the setup state of the program.
	faultGameAbsolutePrestate: #Hex64
	// maximum depth of the position tree within the fault dispute game.
	// `2^{FaultGameMaxDepth}` is how many instructions the execution trace bisection game
	// supports. Ideally, this should be conservatively set so that there is always enough
	// room for a full Cannon trace.
	faultGameMaxDepth: uint64
	// amount of time that the dispute game will set the potential grandchild claim's,
	// clock to, if the remaining time is less than this value at the time of a claim's creation.
	faultGameClockExtension: uint64
	// maximum amount of time that may accumulate on a team's chess clock before they
	// may no longer respond.
	faultGameMaxClockDuration: uint64
	// the block number for genesis.
	faultGameGenesisBlock: uint64
	// output root for the genesis block.
	faultGameGenesisOutputRoot: #Hex64
	// depth at which the fault dispute game splits from output roots to execution trace claims.
	faultGameSplitDepth: uint64
	// number of seconds that users must wait before withdrawing ETH from a fault game.
	faultGameWithdrawalDelay: uint64
	// minimum number of bytes that a large preimage oracle proposal can be.
	preimageOracleMinProposalSize: uint64
	// number of seconds that challengers have to challenge a large preimage proposal.
	preimageOracleChallengePeriod: uint64
	// number of seconds that a proof must be
	// mature before it can be used to finalize a withdrawal.
	proofMaturityDelaySeconds: uint64
	// is an additional number of seconds a
	// dispute game must wait before it can be used to finalize a withdrawal.
	disputeGameFinalityDelaySeconds: uint64
	// dispute game type that the OptimismPortal
	// contract will respect for finalizing withdrawals.
	respectedGameType: uint64
	...
}

#DeployConfig: #L2GenesisBlock &
	#OutputOracle &
	#FaultProofs &
	#Superchain &
	#Core &
	#EIP1559 &
	#GasPriceOracle &
	#Governance &
	#Dev &
	#Celo &
	#DataAvailability &
	#UpgradeSchedule
