set -o errexit   # abort on nonzero exitstatus
set -o nounset   # abort on unbound variable
set -o pipefail  # don't hide errors within pipes

# Create config file
./config/config.sh
cp config/config.json ../packages/contracts-bedrock/deploy-config/${DEPLOYMENT_CONTEXT}.json

# Deploy CREATE2 contract
codesize=$(cast codesize 0x4e59b44847b379578588920cA78FbF26c0B4956C --rpc-url $L1_RPC_URL)
if [[ $codesize =~ 0 ]]; then
  cast publish --rpc-url $L1_RPC_URL 0xf8a58085174876e800830186a08080b853604580600e600039806000f350fe7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe03601600081602082378035828234f58015156039578182fd5b8082525050506014600cf31ba02222222222222222222222222222222222222222222222222222222222222222a02222222222222222222222222222222222222222222222222222222222222222
fi

# Deploy OP contracts
pushd ../packages/contracts-bedrock
forge script scripts/Deploy.s.sol:Deploy --private-key $GS_ADMIN_PRIVATE_KEY --broadcast --rpc-url $L1_RPC_URL --priority-gas-price 1
popd

# Copy deployment information
cp  ../packages/contracts-bedrock/deployments/${DEPLOYMENT_CONTEXT}/.deploy config/deployment-l1.json
