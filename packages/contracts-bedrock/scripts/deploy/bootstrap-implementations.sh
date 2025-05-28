#!/usr/bin/env bash
set -euo pipefail

[ -z "$NETWORK" ] && echo "Need to set the NETWORK via env" && exit 1;
[ -z "$DEPLOY_ETH_RPC_URL" ] && echo "Need to set the DEPLOY_ETH_RPC_URL via env" && exit 1;
[ -z "$DEPLOY_PRIVATE_KEY" ] && echo "Need to set the DEPLOY_PRIVATE_KEY via env" && exit 1;

echo "> Bootstraping contracts"
if [ "${NETWORK}" == "alfajores" ]; then
echo "Performing bootstrap for Alfajores!"
echo "Alfajores is not yet configured from this script!"
exit 1;
elif [ "${NETWORK}" == "baklava" ]; then
echo "Performing bootstrap for Baklava!"
forge script -vvv scripts/deploy/DeployImplementations.s.sol:BaklavaDeployImplementations --rpc-url "$DEPLOY_ETH_RPC_URL" --broadcast --private-key "$DEPLOY_PRIVATE_KEY"
else
  echo "Unsupported network! Choose from 'alfajores' or 'baklava'"
fi
