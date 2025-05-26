#!/usr/bin/env bash
set -euo pipefail

[ -z "$DEPLOY_ETH_RPC_URL" ] && echo "Need to set the DEPLOY_ETH_RPC_URL via env" && exit 1;
[ -z "$DEPLOY_PRIVATE_KEY" ] && echo "Need to set the DEPLOY_PRIVATE_KEY via env" && exit 1;

verify_flag=""
if [ -n "${DEPLOY_VERIFY:-}" ]; then
  verify_flag="--verify"
fi

echo "> Deploying contracts"
forge script -vvv scripts/deploy/Deploy.s.sol:Deploy --rpc-url "$DEPLOY_ETH_RPC_URL" --broadcast --private-key "$DEPLOY_PRIVATE_KEY" $verify_flag

if [ -n "${DEPLOY_GENERATE_HARDHAT_ARTIFACTS:-}" ]; then
  echo "> Generating hardhat artifacts"
  forge script -vvv scripts/deploy/Deploy.s.sol:Deploy --sig 'sync()' --rpc-url "$DEPLOY_ETH_RPC_URL" --broadcast --private-key "$DEPLOY_PRIVATE_KEY"
fi
