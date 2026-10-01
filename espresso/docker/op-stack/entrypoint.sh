#!/bin/sh

GAME_FACTORY_ADDRESS=$(jq -r '.opChainDeployments[0].DisputeGameFactoryProxy' ./deployer/state.json)
export "${ENV_PREFIX}_GAME_FACTORY_ADDRESS=${GAME_FACTORY_ADDRESS}"

"$@"
