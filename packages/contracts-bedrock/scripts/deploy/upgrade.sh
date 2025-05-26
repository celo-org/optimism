#!/usr/bin/env bash
set -euo pipefail

[ -z "$DEPLOY_ETH_RPC_URL" ] && echo "Need to set the DEPLOY_ETH_RPC_URL via env" && exit 1;
[ -z "$DEPLOY_PRIVATE_KEY" ] && echo "Need to set the DEPLOY_PRIVATE_KEY via env" && exit 1;

echo "> Upgrading contracts"
forge script -vvv scripts/deploy/UpgradeOPChain.s.sol:CeloUpgradeOPChain --rpc-url "$DEPLOY_ETH_RPC_URL" --broadcast --private-key "$DEPLOY_PRIVATE_KEY"
