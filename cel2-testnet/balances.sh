#!/bin/bash
set -eo pipefail

source .envrc

# SEQUENCER does not need any balance
for wallet in ADMIN PROPOSER BATCHER
do
    varname=${wallet}_ADDR
    printf "%-12s" ${wallet}:
    echo $(cast balance -r $L1_RPC --ether ${!varname})
done
