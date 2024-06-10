#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail

source shared.sh

function get_balance {
  cast call $L1_BRIDGE_ADDR 'function balanceOf(address) returns (uint256)' $ACC_ADDR
}

balance_before=$(get_balance)
cast send -r $ETH_RPC_URL_L1 --private-key $ACC_PRIVKEY --value 100 $L1_BRIDGE_ADDR
sleep 5
balance_after=$(get_balance)
echo "Balance change: $balance_before -> $balance_after"
[[ $((balance_before + 100)) -eq $balance_after ]] || (
  echo "Balance did not change as expected"
  exit 1
)
