#!/bin/bash
#shellcheck disable=SC2034  # unused vars make sense in a shared file

export ETH_RPC_URL_L1=http://127.0.0.1:8545
export ETH_RPC_URL_L2=http://127.0.0.1:9545

# pre-funded with native token (first default anvil account)
ACC_FUNDED_PRIVKEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
ACC_FUNDED_ADDR=$(cast wallet address $ACC_FUNDED_PRIVKEY)

ACC_PRIVKEY=0x92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e
ACC_ADDR=$(cast wallet address $ACC_PRIVKEY)
REGISTRY_ADDR=0x000000000000000000000000000000000000ce10
TOKEN_ADDR=0x471ece3750da237f93b8e339c536989b8978a438
L1_BRIDGE_ADDR=$(jq '.L1StandardBridgeProxy' ../../.devnet/addresses.json -r)
CUSTOM_GAS_TOKEN_ADDR=$(jq '.CustomGasToken' ../../.devnet/addresses.json -r)
OPTIMISM_PORTAL_PROXY_ADDR=$(jq '.OptimismPortalProxy' ../../.devnet/addresses.json -r)
L2_BRIDGE_ADDR=0x4200000000000000000000000000000000000010
BRIDGED_ETH_ADDR=0x4200000000000000000000000000000000001023
