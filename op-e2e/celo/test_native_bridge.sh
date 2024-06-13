#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail

# NOTE: This requires some native token on the L2,
# which are all prefunded from the genesis

source shared.sh

function get_balance_l1 {
  cast call -r $ETH_RPC_URL_L1 $CUSTOM_GAS_TOKEN_ADDR "balanceOf(address _account)" $ACC_ADDR
}

function get_balance_l2 {
  cast balance $ACC_ADDR
}

cast send -r $ETH_RPC_URL_L1 --private-key $ACC_FUNDED_PRIVKEY --value 10 $ACC_ADDR
balance_before=$(get_balance_l1)
echo $balance_before

# txhash=$(
#   cast send --json --private-key $ACC_PRIVKEY $OPTIMISM_PORTAL_PROXY_ADDR \
#     "initiateWithdrawal(address _target, uint256 _gasLimit, bytes memory _data)" \
#     $ACC_ADDR 21000 0x |
#     jq .transactionHash -r
# )
# echo $txhash
npm test ./testsuite
# balance_after=$(get_balance_l1)
# echo "Balance change: $balance_before -> $balance_after"
# [[ $((balance_before + 9)) -eq $balance_after ]] || (
#   echo "Balance did not change as expected"
#   exit 1
# )
