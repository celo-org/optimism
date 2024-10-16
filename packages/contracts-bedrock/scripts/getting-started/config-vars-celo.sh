#!/usr/bin/env bash

# This script is used to generate the getting-started.json configuration file
# used in the Getting Started quickstart guide on the docs site. Avoids the
# need to have the getting-started.json committed to the repo since it's an
# invalid JSON file when not filled in, which is annoying.

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
CONTRACTS_BASE=$(dirname "$(dirname "$SCRIPT_DIR")")
reqenv() {
    if [ -z "${!1}" ]; then
        echo "Error: environment variable '$1' is undefined"
        exit 1
    fi
}

append_with_default() {
    json_key="$1"
    env_var_name="$2"
    default_value="$3"
    var_value="${!env_var_name}"

    if [ -z "$var_value" ] || [ "$var_value" == "None" ]; then
        var_value="$default_value"
    fi

    echo "  \"$json_key\": \"$var_value\"," >> tmp_config.json
}

# Check required environment variables
reqenv "DEPLOYMENT_CONTEXT"
reqenv "GS_ADMIN_ADDRESS"
reqenv "GS_BATCHER_ADDRESS"
reqenv "GS_PROPOSER_ADDRESS"
reqenv "GS_SEQUENCER_ADDRESS"
reqenv "L1_RPC_URL"
reqenv "L1_CHAIN_ID"
reqenv "L2_CHAIN_ID"
reqenv "L1_BLOCK_TIME"
reqenv "L2_BLOCK_TIME"
reqenv "FUNDS_DEV_ACCOUNTS"
reqenv "USE_ALTDA"
reqenv "DEPLOY_CELO_CONTRACTS"
reqenv "USE_CUSTOM_GAS_TOKEN"
reqenv "CUSTOM_GAS_TOKEN_ADDRESS"

# Get the finalized block timestamp and hash
block=$(cast block finalized --rpc-url "$L1_RPC_URL")
timestamp=$(echo "$block" | awk '/timestamp/ { print $2 }')
blockhash=$(echo "$block" | awk '/hash/ { print $2 }')
batchInboxAddressSuffix=$(printf "%0$((37 - ${#L2_CHAIN_ID}))d" 0)$L2_CHAIN_ID
batchInboxAddress=0xfff$batchInboxAddressSuffix

# Start generating the config file in a temporary file

cat << EOL > tmp_config.json
 {
  "l1StartingBlockTag": "$blockhash",

  "l1ChainID": $L1_CHAIN_ID,
  "l2ChainID": $L2_CHAIN_ID,
  "l2BlockTime": $L2_BLOCK_TIME,
  "l1BlockTime": $L1_BLOCK_TIME,

  "maxSequencerDrift": 600,
  "sequencerWindowSize": 3600,
  "channelTimeout": 300,

  "p2pSequencerAddress": "$GS_SEQUENCER_ADDRESS",
  "batchInboxAddress": "$batchInboxAddress",
  "batchSenderAddress": "$GS_BATCHER_ADDRESS",

  "l2OutputOracleSubmissionInterval": 300,
  "l2OutputOracleStartingBlockNumber": 0,
  "l2OutputOracleStartingTimestamp": $timestamp,

  "l2OutputOracleProposer": "$GS_PROPOSER_ADDRESS",
  "l2OutputOracleChallenger": "$GS_ADMIN_ADDRESS",

  "finalizationPeriodSeconds": 12,

  "proxyAdminOwner": "$GS_ADMIN_ADDRESS",
  "baseFeeVaultRecipient": "$GS_ADMIN_ADDRESS",
  "l1FeeVaultRecipient": "$GS_ADMIN_ADDRESS",
  "sequencerFeeVaultRecipient": "$GS_ADMIN_ADDRESS",
  "finalSystemOwner": "$GS_ADMIN_ADDRESS",
  "superchainConfigGuardian": "$GS_ADMIN_ADDRESS",

  "baseFeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000",
  "l1FeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000",
  "sequencerFeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000",
  "baseFeeVaultWithdrawalNetwork": 0,
  "l1FeeVaultWithdrawalNetwork": 0,
  "sequencerFeeVaultWithdrawalNetwork": 0,

  "gasPriceOracleOverhead": 0,
  "gasPriceOracleScalar": 1000000,

  "deployCeloContracts": $DEPLOY_CELO_CONTRACTS,

  "enableGovernance": $ENABLE_GOVERNANCE,
  "governanceTokenSymbol": "OP",
  "governanceTokenName": "Optimism",
  "governanceTokenOwner": "$GS_ADMIN_ADDRESS",

  "l2GenesisBlockGasLimit": "0x1c9c380",
  "l2GenesisBlockBaseFeePerGas": "0x3b9aca00",

  "eip1559Denominator": 400,
  "eip1559DenominatorCanyon": 400,
  "eip1559Elasticity": 5,
  "eip1559BaseFeeFloor": 5000000000,
EOL

# Append conditional environment variables with their corresponding default values
# Activate granite fork
if [ -n "${GRANITE_TIME_OFFSET}" ]; then
    append_with_default "l2GenesisGraniteTimeOffset" "GRANITE_TIME_OFFSET" "0x0"
fi
# Activate holocene fork
if [ -n "${HOLOCENE_TIME_OFFSET}" ]; then
    append_with_default "l2GenesisHoloceneTimeOffset" "HOLOCENE_TIME_OFFSET" "0x0"
fi

# Activate the interop fork
if [ -n "${INTEROP_TIME_OFFSET}" ]; then
    append_with_default "l2GenesisInteropTimeOffset" "INTEROP_TIME_OFFSET" "0x0"
fi

# Already forked updates
append_with_default "l2GenesisFjordTimeOffset" "FJORD_TIME_OFFSET" "0x0"
append_with_default "l2GenesisRegolithTimeOffset" "REGOLITH_TIME_OFFSET" "0x0"
append_with_default "l2GenesisEcotoneTimeOffset" "ECOTONE_TIME_OFFSET" "0x0"
append_with_default "l2GenesisDeltaTimeOffset" "DELTA_TIME_OFFSET" "0x0"
append_with_default "l2GenesisCanyonTimeOffset" "CANYON_TIME_OFFSET" "0x0"
append_with_default "l2GenesisGraniteTimeOffset" "GRANITE_TIME_OFFSET" "0x0"

# Continue generating the config file
cat << EOL >> tmp_config.json
  "systemConfigStartBlock": 0,

  "requiredProtocolVersion": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "recommendedProtocolVersion": "0x0000000000000000000000000000000000000000000000000000000000000000",

  "faultGameAbsolutePrestate": "0x037ef3c1a487960b0e633d3e513df020c43432769f41a634d18a9595cbf53c55",
  "faultGameMaxDepth": 73,
  "faultGameClockExtension": 10800,
  "faultGameMaxClockDuration": 302400,
  "faultGameGenesisBlock": 0,
  "faultGameGenesisOutputRoot": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "faultGameSplitDepth": 30,
  "faultGameWithdrawalDelay": 604800,

  "preimageOracleMinProposalSize": 126000,
  "preimageOracleChallengePeriod": 86400,

  "fundDevAccounts": $FUNDS_DEV_ACCOUNTS,
  "useFaultProofs": false,
  "initializeAnchorStateRegistry": false,
  "proofMaturityDelaySeconds": 604800,
  "disputeGameFinalityDelaySeconds": 302400,
  "respectedGameType": 0,

  "useAltDA": $USE_ALTDA,
  "daCommitmentType": "GenericCommitment",
  "daChallengeWindow": 1,
  "daResolveWindow": 1,

  "useCustomGasToken": $USE_CUSTOM_GAS_TOKEN,
  "customGasTokenAddress": "$CUSTOM_GAS_TOKEN_ADDRESS"
}
EOL

# Write the final config file
mv tmp_config.json "$CONTRACTS_BASE/deploy-config/$DEPLOYMENT_CONTEXT.json"

echo "Wrote config file to $CONTRACTS_BASE/deploy-config/$DEPLOYMENT_CONTEXT.json"
