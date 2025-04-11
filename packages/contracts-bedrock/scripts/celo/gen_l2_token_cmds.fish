#!/usr/bin/env fish

if [ -z "$argv" ];
  echo Create commands to deploy L2 tokens for bridging from Ethereum
  echo
  echo "Usage: $(status filename) <l1_token_address> [<l1_token_address> ...]"
  return
end

echo
echo "Commands to deploy L2 tokens for bridging from Ethereum:"
echo

set -x ETH_RPC_URL https://ethereum-rpc.publicnode.com

for address in $argv
	set symbol (cast call $address "symbol() returns (string)" --json | jq -r '.[0]')
	set name (cast call $address "name() returns (string)" --json | jq -r '.[0]')
	set decimals (cast call $address "decimals() returns (uint256)" --json | jq -r '.[0]')
	echo cast send 0x4200000000000000000000000000000000000012 "\"createOptimismMintableERC20WithDecimals(address,string,string,uint8)\"" $address "\"$name (Celo native bridge)\"" \"$symbol\" $decimals --private-key \$PRIVKEY
end
