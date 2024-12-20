#!/bin/bash
#shellcheck disable=SC2034  # unused vars make sense in a shared file

export ETH_RPC_URL=http://localhost:9545
export ETH_RPC_URL_L1=http://localhost:8545

export ACC_PRIVKEY=ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
ACC_ADDR=$(cast wallet address $ACC_PRIVKEY)
export ACC_ADDR
export REGISTRY_ADDR=0x000000000000000000000000000000000000ce10
export TOKEN_ADDR=0x471ece3750da237f93b8e339c536989b8978a438
export FEE_CURRENCY_DIRECTORY_ADDR=0x9212Fb72ae65367A7c887eC4Ad9bE310BAC611BF
