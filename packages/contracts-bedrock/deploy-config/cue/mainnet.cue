package deployconfig

// this is the mainnet template for all values
// we know before any L1 deployment, privkey or safe generation
#MainnetTemplate: #Common & {
	_params: sequencerWindowSizeHours: 24

	// required for mainnet
	proxyAdminOwnerIsMultisig!: bool
	externalSuperchainConfig!:  #Address
	finalSystemOwner!:          #Address

	if proxyAdminOwnerIsMultisig == true {
		proxyAdminOwner: !=finalSystemOwner
	}

	externalSuperchainConfig: "0x95703e0982140D16f8ebA6d158FccEde42f04a4C"
	protocolVersionsProxy:    "0x1b6dEB2197418075AB314ac4D52Ca1D104a8F663"
	// set this to null,
	// and also we don't need to overwrite this later anymore
	// since we derive the starting-block based on the
	// beacon-api deterministically now.
	l1StartingBlockTag: null
	l1ChainID:          1
	l2ChainID:          42220
	// this parameter is only relevant when l1 origin is
	// less than l2 start time, just after migration.
	// afterwards this is hardcoded to the value 2892.
	maxSequencerDrift: 2892
	batchInboxAddress: "0xff00000000000000000000000000000000042220"

	l2GenesisBlockGasLimit:      "0x1c9c380"   // 30 000 000
	l2GenesisBlockBaseFeePerGas: "0x5d21dba00" // 25 000 000 000

	// L2OO related, those values are used in
	// the (later stage) initialization of the L2OO
	finalizationPeriodSeconds:        (7 * DayInSeconds)
	l2OutputOracleSubmissionInterval: (30 * MinuteInSeconds)

	// this has to be overwritten later after l2 genesis,
	// but before the fault-game initialization
	faultGameGenesisBlock: uint64 | *31056500

	preimageOracleMinProposalSize: uint64 | *126000
	preimageOracleChallengePeriod: uint64 | *(24 * HourInSeconds)

	useAltDA:       true
	useFaultProofs: true
}

network: production: mainnet: #MainnetTemplate & {
	_params: l1CeloSafeAddress: "0x4092A77bAF58fef0309452cEaCb09221e556E112"
	// multisig on future l2, so we can't withdraw fees to L1:
	_params: feeRecipientAddress: "0x7A1E98FC9a008107DbD1f430a05Ace8cf6f3FE19"
	_params: withdrawFeesOnL2:    true

	//has to be the aliased version of the finalSystemOwner!
	proxyAdminOwnerIsMultisig: true
	proxyAdminOwner:           "0x51a3a77baf58fef0309452ceacb09221e556f223"

	p2pSequencerAddress: "0xA6F1c6c24De8b112dd3867dB907d187d490e6ddF"
	batchSenderAddress:  "0x0cd08c7f7a96aa9635f761b49216b9ea74c5ca60"
	// those are important, since they are also used for the faultproofs
	l2OutputOracleProposer:   "0x1204884e697efd929729b9a717ea14496298a689"
	l2OutputOracleChallenger: "0x6b145ebf66602ec524b196426b46631259689583"
	superchainConfigGuardian: "0x6E226fa22e5F19363d231D3FA048aaBa73CC1f47"

	// When are we able to set the correct values?
	faultGameGenesisOutputRoot:        "0x3c736a83458982ae1f6b62284e9af2687333e17625c7147b9af4758fa84952e8"
	faultGameAbsolutePrestate:         "0x0364010a7b2be12b8583c8bc2c610ef5b77bb52161cac1dd4f8cbe47edc05afd"
	faultGameGenesisBlock:             31056500
	l2OutputOracleStartingBlockNumber: 0
	l2OutputOracleStartingTimestamp:   0
}

network: dryrun: mainnet: #MainnetTemplate & {
	// not a safe, but we can use the parameter with an EOA
	_params: l1CeloSafeAddress:   "0x3d98acBC85D3252DFfd6b500D94341F4774256F0"
	_params: feeRecipientAddress: "0x22EaF69162ae49605441229EdbEF7D9FC5f4f094"
	_params: withdrawFeesOnL2:    true

	// instead of the aliased address, use an EOA
	proxyAdminOwner:           "0x4ea9acbc85d3252dffd6b500d94341f477426801"
	proxyAdminOwnerIsMultisig: true

	l2OutputOracleStartingBlockNumber: 30941000
	l2OutputOracleStartingTimestamp:   1742366039

	p2pSequencerAddress:      "0xc0fD4a912b7aC8D5a3ABDbef23c88c67Cfb528Cb"
	batchSenderAddress:       "0x4e8b8dd9611845f5fb80f43662dbeefbb47a75f6"
	l2OutputOracleProposer:   "0xc01061d4cc5b98965d2aa4b1dfcc1d77bb0d29f3"
	l2OutputOracleChallenger: "0x9e21944b9dd761e7a89ecb17be005e955e750f2b"
	superchainConfigGuardian: "0x1174B5f5Dd8fA3be9549b131E9810703D15f153d"

	faultGameGenesisBlock:           30941000
	disputeGameFinalityDelaySeconds: (1 * HourInSeconds)
	faultGameWithdrawalDelay:        (2 * disputeGameFinalityDelaySeconds)
	proofMaturityDelaySeconds:       (2 * disputeGameFinalityDelaySeconds)
	faultGameClockExtension:         (30 * MinuteInSeconds)
	faultGameMaxClockDuration:       faultGameWithdrawalDelay
	preimageOracleChallengePeriod:   (10 * MinuteInSeconds)

	faultGameAbsolutePrestate:  "0x0364010a7b2be12b8583c8bc2c610ef5b77bb52161cac1dd4f8cbe47edc05afd"
	faultGameGenesisOutputRoot: "0x3c736a83458982ae1f6b62284e9af2687333e17625c7147b9af4758fa84952e8"
}
