package deployconfig

network: test: baklava: #Common & {
	_params: {
		l1CeloSafeAddress:        "0x22EaF69162ae49605441229EdbEF7D9FC5f4f094"
		feeRecipientAddress:      l1CeloSafeAddress
		sequencerWindowSizeHours: 24
		withdrawFeesOnL2:         false
	}

	l1StartingBlockTag: null
	l1ChainID:          17000
	l2ChainID:          62320
	maxSequencerDrift:  10 * MinuteInSeconds
	// this was used incorrectly, or at least it doesn't have
	// any effect since we also activate the granite hardfork
	// which has a hardcoded value of 50
	channelTimeout: 300

	p2pSequencerAddress:      "0x3Cd8072cbC235246c684ab9BD76Bb6f3813Df2CD"
	batchInboxAddress:        "0xff00000000000000000000000000000000062320"
	batchSenderAddress:       "0x242C6e6eA8e910A1835eFA4CaF8641769C27B595"
	proxyAdminOwner:          _params.l1CeloSafeAddress
	superchainConfigGuardian: _params.l1CeloSafeAddress

	l2OutputOracleSubmissionInterval:  600
	l2OutputOracleStartingBlockNumber: 28308600
	l2OutputOracleStartingTimestamp:   1739987136
	l2OutputOracleProposer:            "0x85c7AC265419359806B147Fba6Ea654229928333"
	l2OutputOracleChallenger:          "0xDc94436A193a827786270dD4F6cD4b35c3f0C8f8"

	finalizationPeriodSeconds: 12

	l2GenesisBlockGasLimit:      "0x1c9c380"
	l2GenesisBlockBaseFeePerGas: "0x3b9aca00"

	useFaultProofs:             true
	faultGameAbsolutePrestate:  "0x0364010a7b2be12b8583c8bc2c610ef5b77bb52161cac1dd4f8cbe47edc05afd"
	faultGameGenesisOutputRoot: "0x2fafa02f4d94e20796afac0bae793bcc5c3cbb1244bc5a8c730153def18b0f3f"

	preimageOracleMinProposalSize: 126000
	preimageOracleChallengePeriod: 86400

	useAltDA: true
}
