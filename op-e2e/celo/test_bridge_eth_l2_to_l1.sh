#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail

# NOTE: This requires some tokens on the L2, so this test will only work if test_bridge_eth_l1_to_l2.sh has been run before

source shared.sh

function get_balance {
  cast balance -r $ETH_RPC_URL_L1 $ACC_ADDR
}

balance_before=$(get_balance)
txhash=$(
  cast send --json --private-key $ACC_PRIVKEY $L2_BRIDGE_ADDR \
      'function withdraw(address _l2Token, uint256 _amount, uint32 _minGasLimit, bytes calldata _extraData)' \
      $BRIDGED_ETH_ADDR 9 21000 0x \
    | jq .transactionHash -r)
npx ./prove_bridged_message $txhash
balance_after=$(get_balance)
echo "Balance change: $balance_before -> $balance_after"
[[ $((balance_before + 9)) -eq $balance_after ]] || (echo "Balance did not change as expected"; exit 1)
