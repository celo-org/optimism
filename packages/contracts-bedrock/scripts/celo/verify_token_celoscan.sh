#!/bin/bash

if [ -z "$*" ]; then
  echo "Verify L2 bridged tokens on Celoscan"
  echo
  echo "Usage: $0 <token_address> [<token_address> ...]"
  exit 1
fi

for BRIDGED_TOKEN in "$@"; do
  # cast_call <address> <signature>
  function cast_call() {
    cast call --json --rpc-url https://forno.celo.org "$1" "$2" | jq -r ".[0]"
  }

  REMOTE_TOKEN=$(cast_call "$BRIDGED_TOKEN" "REMOTE_TOKEN()(address)")
  NAME=$(cast_call "$BRIDGED_TOKEN" "name()(string)")
  SYMBOL=$(cast_call "$BRIDGED_TOKEN" "symbol()(string)")
  DECIMALS=$(cast_call "$BRIDGED_TOKEN" "decimals()(uint8)")

  CONSTRUCTOR_ARGS=$(cast abi-encode "constructor(address,address,string,string,uint8)"  0x4200000000000000000000000000000000000010 "$REMOTE_TOKEN" "$NAME" "$SYMBOL" "$DECIMALS")
  CONSTRUCTOR_ARGS=${CONSTRUCTOR_ARGS#0x}

  forge verify-contract \
    --verifier=etherscan \
    --verifier-url=https://api.celoscan.io/api/ \
    --constructor-args="$CONSTRUCTOR_ARGS" \
    --skip-is-verified-check \
    "$BRIDGED_TOKEN" \
    src/universal/OptimismMintableERC20.sol:OptimismMintableERC20
done
