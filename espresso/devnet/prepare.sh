#!/usr/bin/env bash
# Generates the deployment consumed by espresso/docker-compose.yml:
#
#   espresso/deployment/l1-config/state.json    L1 chain for anvil (--load-state), with
#                                               the Espresso and OP contracts deployed
#   espresso/deployment/l2-config/genesis.json  L2 genesis (timestamp patched at startup)
#   espresso/deployment/l2-config/rollup.json   rollup config (L2 genesis patched at startup)
#   espresso/deployment/deployer/               op-deployer workdir (intent, state)
#   espresso/deployment/deployer/succinct.env   contract addresses for the op-succinct services
#
# An anvil chain is seeded with the Espresso contracts of
# espresso/environment/allocs.json, the OP contracts are deployed on it with
# op-deployer, and its full state, blocks included, is dumped. The devnet L1
# loads that dump and continues the same chain, so everything that refers to
# L1 blocks of the deployment stays valid: the rollup config's L1 genesis, the
# BatchAuthenticator's batcher history, and derivation's batch-auth lookback
# over L1 ancestors.
#
# Requires: go, anvil, cast, jq, and compiled contracts in
# packages/contracts-bedrock/forge-artifacts (`just build-contracts`).
set -euo pipefail

ESPRESSO_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
OP_ROOT=$(cd "${ESPRESSO_DIR}/.." && pwd)

# shellcheck source-path=SCRIPTDIR source=../.env
source "${ESPRESSO_DIR}/.env"

DEPLOYMENT_DIR="${ESPRESSO_DIR}/deployment"
DEPLOYER_DIR="${DEPLOYMENT_DIR}/deployer"
L1_CONFIG_DIR="${DEPLOYMENT_DIR}/l1-config"
L2_CONFIG_DIR="${DEPLOYMENT_DIR}/l2-config"
BIN_DIR="${DEPLOYMENT_DIR}/bin"
ANVIL_STATE_FILE="${DEPLOYMENT_DIR}/anvil_state.json"
ARTIFACTS_DIR="file://${OP_ROOT}/packages/contracts-bedrock/forge-artifacts"

ANVIL_PORT=${PREPARE_ANVIL_PORT:-28545}
ANVIL_URL="http://127.0.0.1:${ANVIL_PORT}"

for tool in go anvil cast jq; do
    command -v "${tool}" >/dev/null || { echo "missing required tool: ${tool}" >&2; exit 1; }
done
if [[ ! -d "${OP_ROOT}/packages/contracts-bedrock/forge-artifacts" ]]; then
    echo "missing forge artifacts; run 'just build-contracts' first" >&2
    exit 1
fi

rm -rf "${DEPLOYMENT_DIR}"
mkdir -p "${DEPLOYER_DIR}" "${L1_CONFIG_DIR}" "${L2_CONFIG_DIR}" "${BIN_DIR}"

echo "Building op-deployer and dasel..."
(cd "${OP_ROOT}" && go build -o "${BIN_DIR}/op-deployer" ./op-deployer/cmd/op-deployer)
# dasel v2 edits the op-deployer intent, which is TOML.
GOBIN="${BIN_DIR}" go install github.com/tomwright/dasel/v2/cmd/dasel@v2.8.1
OP_DEPLOYER="${BIN_DIR}/op-deployer"
DASEL="${BIN_DIR}/dasel"

ESPRESSO_ALLOCS="${ESPRESSO_DIR}/environment/allocs.json"

anvil --port "${ANVIL_PORT}" --chain-id "${L1_CHAIN_ID}" \
    --disable-gas-limit --disable-code-size-limit \
    --dump-state "${ANVIL_STATE_FILE}" >"${DEPLOYMENT_DIR}/anvil.log" 2>&1 &
ANVIL_PID=$!
trap 'kill "${ANVIL_PID}" 2>/dev/null || true' EXIT

for _ in $(seq 1 50); do
    cast bn --rpc-url "${ANVIL_URL}" >/dev/null 2>&1 && break
    sleep 0.2
done

anvil_rpc() { cast rpc --rpc-url "${ANVIL_URL}" "$@" >/dev/null; }

# Seed the Espresso contracts. This also advances OPERATOR's nonce past the
# Espresso deployments, so the OP contracts it deploys with CREATE do not land
# on Espresso addresses.
jq -r 'to_entries[] | [.key, (.value.state.code // "0x"), (.value.state.nonce // 0), (.value.state.balance // "0x0")] | @tsv' \
    "${ESPRESSO_ALLOCS}" |
    while IFS=$'\t' read -r address code nonce balance; do
        anvil_rpc anvil_setCode "${address}" "${code}"
        anvil_rpc anvil_setNonce "${address}" "$(printf '0x%x' "${nonce}")"
        anvil_rpc anvil_setBalance "${address}" "${balance}"
    done
jq -r 'to_entries[] | .key as $address | (.value.state.storage // {}) | to_entries[] | [$address, .key, .value] | @tsv' \
    "${ESPRESSO_ALLOCS}" |
    while IFS=$'\t' read -r address slot value; do
        anvil_rpc anvil_setStorageAt "${address}" "${slot}" "${value}"
    done

anvil_rpc anvil_setBalance "${OPERATOR_ADDRESS}" 0x100000000000000000000000000000000000
anvil_rpc anvil_setBalance "${PROPOSER_ADDRESS}" 0x100000000000000000000000000000000000

"${OP_DEPLOYER}" bootstrap superchain \
    --l1-rpc-url="${ANVIL_URL}" \
    --private-key="${OPERATOR_PRIVATE_KEY}" \
    --artifacts-locator="${ARTIFACTS_DIR}" \
    --outfile="${DEPLOYER_DIR}/bootstrap_superchain.json" \
    --superchain-proxy-admin-owner="${OPERATOR_ADDRESS}" \
    --guardian="${OPERATOR_ADDRESS}"

"${OP_DEPLOYER}" bootstrap implementations \
    --l1-rpc-url="${ANVIL_URL}" \
    --private-key="${OPERATOR_PRIVATE_KEY}" \
    --artifacts-locator="${ARTIFACTS_DIR}" \
    --superchain-config-proxy="$(jq -r .superchainConfigProxyAddress "${DEPLOYER_DIR}/bootstrap_superchain.json")" \
    --superchain-proxy-admin="$(jq -r .proxyAdminAddress "${DEPLOYER_DIR}/bootstrap_superchain.json")" \
    --upgrade-controller="${OPERATOR_ADDRESS}" \
    --challenger="${OPERATOR_ADDRESS}" \
    --proof-maturity-delay-seconds=12 \
    --dispute-game-finality-delay-seconds=6 \
    --outfile="${DEPLOYER_DIR}/bootstrap_implementations.json"

"${OP_DEPLOYER}" init \
    --l1-chain-id "${L1_CHAIN_ID}" \
    --l2-chain-ids "${L2_CHAIN_ID}" \
    --intent-type standard-overrides \
    --outdir "${DEPLOYER_DIR}"

INTENT="${DEPLOYER_DIR}/intent.toml"
intent() { "${DASEL}" put -f "${INTENT}" "$@"; }

intent -s .l1ContractsLocator -v "${ARTIFACTS_DIR}"
intent -s .l2ContractsLocator -v "${ARTIFACTS_DIR}"
intent -s .opcmAddress -v "$(jq -r .opcmV2Address "${DEPLOYER_DIR}/bootstrap_implementations.json")"
intent -s .superchainConfigProxy -v "$(jq -r .superchainConfigProxyAddress "${DEPLOYER_DIR}/bootstrap_superchain.json")"
intent -s .fundDevAccounts -t bool -v true
intent -s .globalDeployOverrides.faultGameMaxClockDuration -t int -v 302400
intent -s .globalDeployOverrides.faultGameClockExtension -t int -v 10800
intent -s .globalDeployOverrides.preimageOracleChallengePeriod -t int -v 0
intent -s .globalDeployOverrides.dangerouslyAllowCustomDisputeParameters -t bool -v true
intent -s .globalDeployOverrides.proofMaturityDelaySeconds -t int -v 12
intent -s .globalDeployOverrides.disputeGameFinalityDelaySeconds -t int -v 6

# The Espresso batcher (HD index 6) is registered in the BatchAuthenticator;
# the fallback batcher (HD index 2) is the SystemConfig batcher.
intent -s '.chains.[0].espressoEnabled' -t bool -v true
intent -s '.chains.[0].espressoBatcher' -v "${ESPRESSO_BATCHER_ADDRESS}"
intent -s '.chains.[0].roles.batcher' -v "${FALLBACK_BATCHER_ADDRESS}"
intent -s '.chains.[0].roles.proposer' -v "${PROPOSER_ADDRESS}"
for role in systemConfigOwner unsafeBlockSigner l1ProxyAdminOwner l2ProxyAdminOwner challenger; do
    intent -s ".chains.[0].roles.${role}" -v "${OPERATOR_ADDRESS}"
done
for recipient in baseFeeVaultRecipient l1FeeVaultRecipient sequencerFeeVaultRecipient operatorFeeVaultRecipient chainFeesRecipient; do
    intent -s ".chains.[0].${recipient}" -v "${OPERATOR_ADDRESS}"
done
intent -s '.chains.[0].dangerousAltDAConfig.useAltDA' -t bool -v true
intent -s '.chains.[0].dangerousAltDAConfig.daCommitmentType' -v "GenericCommitment"
intent -s '.chains.[0].dangerousAltDAConfig.daChallengeWindow' -t int -v 6
intent -s '.chains.[0].dangerousAltDAConfig.daResolveWindow' -t int -v 1

# A fixed create2 salt keeps the contract addresses deterministic.
jq '.create2Salt = "0xaecea4f57fadb2097ccd56594f2f22715ac52f92971c5913b70a7f1134b68feb"' \
    "${DEPLOYER_DIR}/state.json" >"${DEPLOYER_DIR}/state.json.tmp"
mv "${DEPLOYER_DIR}/state.json.tmp" "${DEPLOYER_DIR}/state.json"

BATCH_AUTHENTICATOR_OWNER_ADDRESS="${BATCH_AUTHENTICATOR_OWNER_ADDRESS}" "${OP_DEPLOYER}" apply \
    --l1-rpc-url "${ANVIL_URL}" \
    --workdir "${DEPLOYER_DIR}" \
    --private-key="${OPERATOR_PRIVATE_KEY}"

"${OP_DEPLOYER}" inspect genesis --workdir "${DEPLOYER_DIR}" --outfile "${L2_CONFIG_DIR}/genesis.json" "${L2_CHAIN_ID}"
"${OP_DEPLOYER}" inspect rollup --workdir "${DEPLOYER_DIR}" --outfile "${L2_CONFIG_DIR}/rollup.json" "${L2_CHAIN_ID}"

# Contract addresses for the fdg-deployer, succinct-proposer and
# succinct-challenger services. op-succinct reads FACTORY_ADDRESS.
deployment() { jq -r ".opChainDeployments[0].$1" "${DEPLOYER_DIR}/state.json"; }
cat >"${DEPLOYER_DIR}/succinct.env" <<EOF
FACTORY_ADDRESS=$(deployment DisputeGameFactoryProxy)
DISPUTE_GAME_FACTORY_ADDRESS=$(deployment DisputeGameFactoryProxy)
OPTIMISM_PORTAL2_ADDRESS=$(deployment OptimismPortalProxy)
ANCHOR_STATE_REGISTRY_ADDRESS=$(deployment AnchorStateRegistryProxy)
EOF

# Stop anvil; it writes its state to ANVIL_STATE_FILE on SIGTERM.
kill -TERM "${ANVIL_PID}"
wait "${ANVIL_PID}" || true
trap - EXIT
mv "${ANVIL_STATE_FILE}" "${L1_CONFIG_DIR}/state.json"

echo "Devnet deployment written to ${DEPLOYMENT_DIR}"
