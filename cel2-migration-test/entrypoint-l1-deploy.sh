#!/bin/bash
set -exu

VERBOSITY=${GETH_VERBOSITY:-3}
CONFIG_DIR=/config
CONFIG_OUT=/config/config.json

export GS_ADMIN_ADDRESS="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
export GS_ADMIN_PRIVATE_KEY="0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

export GS_BATCHER_ADDRESS="0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
export GS_BATCHER_PRIVATE_KEY="0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"

export GS_PROPOSER_ADDRESS="0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"
export GS_PROPOSER_PRIVATE_KEY="0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"

export GS_SEQUENCER_ADDRESS="0x90F79bf6EB2c4f870365E785982E1f101E93b906"
export GS_SEQUENCER_PRIVATE_KEY="0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6"

export L1_RPC_URL="http://l1:8545"
export DEPLOYMENT_CONTEXT=cel2-migration


if [ ! -f "$CONFIG_OUT" ]; then
	echo "$CONFIG_OUT missing, running OP deployment"

  # Create config file
  /config/config.sh

  # Deplot CREATE2 contract
  codesize=$(cast codesize 0x4e59b44847b379578588920cA78FbF26c0B4956C --rpc-url $L1_RPC_URL)
  if [[ $codesize =~ 0 ]]; then
    cast publish --rpc-url $L1_RPC_URL 0xf8a58085174876e800830186a08080b853604580600e600039806000f350fe7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe03601600081602082378035828234f58015156039578182fd5b8082525050506014600cf31ba02222222222222222222222222222222222222222222222222222222222222222a02222222222222222222222222222222222222222222222222222222222222222
  fi

  # Deploy OP contracts
  cd /repo/packages/contracts-bedrock
  forge script scripts/Deploy.s.sol:Deploy --private-key $GS_ADMIN_PRIVATE_KEY --broadcast --rpc-url $L1_RPC_URL

  cp deployments/getting-started/.deploy /config/deployment-l1.json
else
	echo "$CONFIG_OUT exists."
fi
