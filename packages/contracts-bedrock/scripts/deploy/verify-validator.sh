#!/usr/bin/env bash
set -euo pipefail

# V2
# VALIDATOR=0xbaf913435088651429154dc2e2caa031f3b75c90

# V3
# VALIDATOR=0x9df52e41189e89485bb7aee1e5cc93874dd89712

# Require env vars
[ -z "${VERSION:-}" ] && echo "Need to set the VERSION via env" && exit 1;
[ -z "${VALIDATOR:-}" ] && echo "Need to set the VERSION via env" && exit 1;

# Optional env vars
if [ -z "${CHAIN_ID:-}" ]; then
    # Fallback to Holesky
    CHAIN_ID=17000
fi

# Check version
case $VERSION in
    "v2")
        echo "Detected supported version: $VERSION"
        CONTRACT="StandardValidatorV200"
    ;;
    "v3")
        echo "Detected supported version: $VERSION"
        CONTRACT="StandardValidatorV300"
    ;;
    *)
    echo "Invalid version: $VERSION" && exit 1
    ;;
esac

if [ "${BLOCKSCOUT_API_KEY:-}" ]; then
    forge verify-contract $VALIDATOR $CONTRACT \
        --chain-id $CHAIN_ID \
        --etherscan-api-key=$BLOCKSCOUT_API_KEY \
        --verifier-url=https://eth-holesky.blockscout.com/api/ \
        --watch
fi
if [ "${ETHERSCAN_API_KEY:-}" ]; then
    echo "Etherscan not yet supported! Missing constructor verification";
    # forge verify-contract $VALIDATOR $CONTRACT \
    #     --chain-id $CHAIN_ID \
    #     --etherscan-api-key=$ETHERSCAN_API_KEY \
    #     --watch
fi
if [ "${TENDERLY_URL:-}" && "${TENDERLY_API_KEY:-}" ]; then
    forge verify-contract $VALIDATOR $CONTRACT \
        --chain-id $CHAIN_ID \
        --verifier-url=$TENDERLY_URL \
        --etherscan-api-key=$TENDERLY_API_KEY \
        --watch
fi
