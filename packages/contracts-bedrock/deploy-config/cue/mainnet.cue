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

	protocolVersionsProxy: "0x1b6dEB2197418075AB314ac4D52Ca1D104a8F663"
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
	l2OutputOracleSubmissionInterval: 1800

	preimageOracleMinProposalSize: 126000
	preimageOracleChallengePeriod: 86400

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
	externalSuperchainConfig: "0x95703e0982140D16f8ebA6d158FccEde42f04a4C"

	// When are we able to set the correct values?
	faultGameGenesisOutputRoot:        "0x3c736a83458982ae1f6b62284e9af2687333e17625c7147b9af4758fa84952e8"
	faultGameAbsolutePrestate:         "0x0364010a7b2be12b8583c8bc2c610ef5b77bb52161cac1dd4f8cbe47edc05afd"
	l2OutputOracleStartingBlockNumber: 0
	l2OutputOracleStartingTimestamp:   0
}

network: dryrun: mainnet: #MainnetTemplate & {
	_params: l1CeloSafeAddress:   "0x1174B5f5Dd8fA3be9549b131E9810703D15f153d"
	_params: feeRecipientAddress: "0x22EaF69162ae49605441229EdbEF7D9FC5f4f094"
	_params: withdrawFeesOnL2:    false

	// instead of the aliased address, use the EOA
	proxyAdminOwner:           _params.l1CeloSafeAddress
	proxyAdminOwnerIsMultisig: false
	externalSuperchainConfig:  ZeroAddress

	l2OutputOracleStartingBlockNumber: 0
	l2OutputOracleStartingTimestamp:   0

	p2pSequencerAddress:      "0x8478dB1A971C003f5Fe8eb4160C696f20B1FF6B6"
	batchSenderAddress:       "0x4e8b8dd9611845f5fb80f43662dbeefbb47a75f6"
	l2OutputOracleProposer:   "0xc01061d4cc5b98965d2aa4b1dfcc1d77bb0d29f3"
	l2OutputOracleChallenger: "0x9e21944b9dd761e7a89ecb17be005e955e750f2b"
	superchainConfigGuardian: "0x1174B5f5Dd8fA3be9549b131E9810703D15f153d"

	faultGameAbsolutePrestate:  "0x0318d12b4f68c79bd937d480326031ceffbefad5934e431a3b430c058c1c9e1b"
	faultGameGenesisOutputRoot: "0x3c736a83458982ae1f6b62284e9af2687333e17625c7147b9af4758fa84952e8"
}
