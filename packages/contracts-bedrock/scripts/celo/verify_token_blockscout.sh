#!/bin/bash

if [ -z "$*" ]; then
  echo "Verify L2 bridged tokens on Blockscout"
  echo
  echo "Usage: $0 <token_address> [<token_address> ...]"
  exit 1
fi

for BRIDGED_TOKEN in "$@"; do
  forge verify-contract \
    --verifier=blockscout \
    --verifier-url=https://celo.blockscout.com/api/ \
    "$BRIDGED_TOKEN" \
    src/universal/OptimismMintableERC20.sol:OptimismMintableERC20
done
