#!/bin/sh

if [ -z "$1" ]; then
  echo "Create commands to deploy L2 tokens for bridging from Ethereum"
  echo
  echo "Usage: $(basename "$0") <l1_token_address> [<l1_token_address> ...]"
  exit 1
fi

echo
echo "Commands to deploy L2 tokens for bridging from Ethereum:"
echo

ETH_RPC_URL=https://ethereum-rpc.publicnode.com
export ETH_RPC_URL

for address in "$@"; do
  symbol=$(cast call "$address" "symbol() returns (string)" --json | jq -r '.[0]')
  name=$(cast call "$address" "name() returns (string)" --json | jq -r '.[0]')
  decimals=$(cast call "$address" "decimals() returns (uint256)" --json | jq -r '.[0]')
  echo "cast send 0x4200000000000000000000000000000000000012 \"createOptimismMintableERC20WithDecimals(address,string,string,uint8)\" $address \"$name (Celo native bridge)\" \"$symbol\" $decimals --private-key \$PRIVKEY"
done
