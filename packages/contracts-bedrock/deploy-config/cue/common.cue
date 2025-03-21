package deployconfig

#Parameters: {
	_params: {
		l1CeloSafeAddress:        #Address
		feeRecipientAddress:      #Address
		sequencerWindowSizeHours: uint64
		withdrawFeesOnL2!:        bool
		isLocalDevnet!:           bool
	}
	...
}

#Common: #DeployConfig & #Parameters & {
	_params: {
		// we set this to double the OP-suggested value (of 12h),
		// because we had issues with EigenDA before and
		// this was the only thing preventing a long reorg.
		// Now we have the fallback mechanism, so this is not as
		// relevant any longer. We still agreed that we want to be
		// on the safe side here.
		sequencerWindowSizeHours: uint64 | *24
		isLocalDevnet:            bool | *false
		withdrawFeesOnL2:         bool | true
	}

	deployCeloContracts: false
	fundDevAccounts:     false
	if _params.isLocalDevnet == true {
		deployCeloContracts: true
		fundDevAccounts:     true
	}

	l2BlockTime: 1
	l1BlockTime: 12
	// has not effect anyway if not pre-Granite,
	// but set it to a default of 50, which is the Granite hardcoded value
	channelTimeout:         uint64 | *50
	sequencerWindowSize:    div((_params.sequencerWindowSizeHours * HourInSeconds), l1BlockTime)
	systemConfigStartBlock: 0

	finalSystemOwner: _params.l1CeloSafeAddress

	// superchain-related
	requiredProtocolVersion:    "0x0000000000000000000000000000000000000003000000010000000000000000"
	recommendedProtocolVersion: "0x0000000000000000000000000000000000000003000000010000000000000000"

	// those values were chosen to match the gasprice parameters of ethereum,
	// while taking into account our shorter blocktime
	eip1559Denominator:       400
	eip1559Elasticity:        5
	eip1559DenominatorCanyon: eip1559Denominator
	eip1559Denominator:       400
	eip1559Elasticity:        5
	eip1559BaseFeeFloor:      25 * gWei

	gasPriceOracleOverhead:          0
	gasPriceOracleScalar:            0
	gasPriceOracleBaseFeeScalar:     0
	gasPriceOracleBlobBaseFeeScalar: 0

	l2GenesisFjordTimeOffset:    #Hex64 | *"0x0"
	l2GenesisRegolithTimeOffset: #Hex64 | *"0x0"
	l2GenesisEcotoneTimeOffset:  #Hex64 | *"0x0"
	l2GenesisDeltaTimeOffset:    #Hex64 | *"0x0"
	l2GenesisCanyonTimeOffset:   #Hex64 | *"0x0"
	l2GenesisGraniteTimeOffset:  #Hex64 | *"0x0"

	// those two will be deployed but don't receive funds
	baseFeeVaultRecipient: _params.feeRecipientAddress
	l1FeeVaultRecipient:   _params.feeRecipientAddress
	// for now this will be deployed and receive the tips
	// in Celo but also fee-currency.
	// The Celo is withdrawable to this address by everyone,
	// the fee-currencies are owned by this address but the
	// contract implementation is not compatible with ERC20 for now.
	sequencerFeeVaultRecipient:               _params.feeRecipientAddress
	baseFeeVaultMinimumWithdrawalAmount:      "0x8ac7230489e80000" // 10000000000000000000
	l1FeeVaultMinimumWithdrawalAmount:        "0x8ac7230489e80000" // 10000000000000000000
	sequencerFeeVaultMinimumWithdrawalAmount: "0x8ac7230489e80000" // 10000000000000000000

	if _params.withdrawFeesOnL2 == false {
		sequencerFeeVaultWithdrawalNetwork: 0
		baseFeeVaultWithdrawalNetwork:      0
		l1FeeVaultWithdrawalNetwork:        0
	}
	if _params.withdrawFeesOnL2 == true {
		sequencerFeeVaultWithdrawalNetwork: 1
		baseFeeVaultWithdrawalNetwork:      1
		l1FeeVaultWithdrawalNetwork:        1
	}

	enableGovernance:     false
	governanceTokenOwner: ZeroAddress

	// AltDA
	daCommitmentType:  "GenericCommitment"
	daChallengeWindow: 1
	daResolveWindow:   1

	// FaultProofs
	faultGameMaxDepth:       uint64 | *73
	faultGameClockExtension: uint64 | *(3 * HourInSeconds)
	faultGameSplitDepth:     uint64 | *30
	// we are following op and base here
	faultGameWithdrawalDelay: uint64 | *(7 * DayInSeconds)

	proofMaturityDelaySeconds:       uint64 | *(7 * DayInSeconds)
	disputeGameFinalityDelaySeconds: uint64 | *(div(7*DayInSeconds, 2))
	faultGameMaxClockDuration:       uint64 | *(div(7*DayInSeconds, 2))
	// different than OP's "0" type
	respectedGameType: 1

	useCustomGasToken:     true
	customGasTokenAddress: ZeroAddress
}
