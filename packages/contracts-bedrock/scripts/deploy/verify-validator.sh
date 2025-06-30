#!/usr/bin/env bash
set -euo pipefail

# Baklava: V2
# VALIDATOR=0xbaf913435088651429154dc2e2caa031f3b75c90

# Baklava: V3
# VALIDATOR=0x9df52e41189e89485bb7aee1e5cc93874dd89712

# Alfajores: V2
# VALIDATOR=0x14271661be501d640701a9a59bdc105f6ab4bce9

# Alfajores: V3
# VALIDATOR=0xc6bacfa8421117677e03c3eb81d44b37a9ceef31

# Mainnet: V2
# VALIDATOR=0x3c7433651845b795cf45eda81dc5ba20ebacc5e6

# Mainnet: V3
VALIDATOR=0x1ef2aa0d3d08d320a9d6bf23661bfcb6bf1ce6fa

# Require env vars
[ -z "${VERSION:-}" ] && echo "Need to set the VERSION via env" && exit 1;
[ -z "${VALIDATOR:-}" ] && echo "Need to set the VERSION via env" && exit 1;

# Optional env vars
if [ -z "${CHAIN_ID:-}" ]; then
    # Fallback to Holesky
    CHAIN_ID=17000
fi

# BLOCKSCOUT_URL=https://eth-holesky.blockscout.com/api/
BLOCKSCOUT_URL=https://eth.blockscout.com/api/

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
        --verifier-url=$BLOCKSCOUT_URL \
        --watch
fi
if [ "${ETHERSCAN_API_KEY:-}" ]; then
    echo "Etherscan not yet supported! Missing constructor verification";
    # forge verify-contract $VALIDATOR $CONTRACT \
    #     --chain-id $CHAIN_ID \
    #     --etherscan-api-key=$ETHERSCAN_API_KEY \
    #     --watch
fi
if [ "${TENDERLY_URL:-}" ] && [ "${TENDERLY_API_KEY:-}" ]; then
    forge verify-contract $VALIDATOR $CONTRACT \
        --chain-id $CHAIN_ID \
        --verifier-url=$TENDERLY_URL \
        --etherscan-api-key=$TENDERLY_API_KEY \
        --watch
fi
