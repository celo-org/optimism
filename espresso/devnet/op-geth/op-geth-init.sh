#!/bin/bash
# Entrypoint of the op-geth image of the Espresso compose devnet.
#
# The L2 genesis and rollup config are generated ahead of time by
# espresso/devnet/prepare.sh and mounted read-only at /deployment/l2-config.
# The one-shot "genesis" and "rollup" modes copy them into the shared /config
# volume and patch in the values that are only known once the devnet runs.
#
# MODE=genesis  set the L2 genesis timestamp to now, write the JWT secret, and
#               wait until L1 (L1_RPC) has a finalized block
# MODE=rollup   point the rollup config at the L2 genesis block of the
#               sequencer's op-geth (OP_RPC). Its L1 genesis, the deployment
#               block, is already right: the devnet L1 continues that chain.
# MODE=geth     (default) initialize the datadir from /config/genesis.json and run geth
set -euo pipefail

OP_HTTP_PORT=${OP_HTTP_PORT:-8546}
OP_ENGINE_PORT=${OP_ENGINE_PORT:-8552}
L2_CHAIN_ID=${L2_CHAIN_ID:-22266222}
MODE=${MODE:-geth}

# Prints the requested field of a block, e.g. `block_field "$L1_RPC" 0x0 hash`.
block_field() {
    curl -sf -X POST -H 'Content-Type: application/json' \
        --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$2\",false],\"id\":1}" \
        "$1" | jq -r ".result.$3 // empty"
}

case "$MODE" in
genesis)
    timestamp=$(printf '0x%x' "$(date +%s)")
    jq --arg ts "$timestamp" '.timestamp = $ts' /deployment/l2-config/genesis.json >/config/genesis.json
    # Fixed dev secret, shared by op-geth and op-node through /config.
    printf '2692310708e4207ecd73bf5597a59ab9cd085380108a7787b3d6be22840e37f0' >/config/jwt.txt

    echo "Waiting for an L1 finalized block..."
    until [[ -n "$(block_field "$L1_RPC" finalized number)" ]]; do
        sleep 3
    done
    echo "L2 genesis ready (timestamp ${timestamp})"
    ;;

rollup)
    l2_hash=$(block_field "$OP_RPC" 0x0 hash)
    l2_time=$(block_field "$OP_RPC" 0x0 timestamp)
    if [[ -z "$l2_hash" || -z "$l2_time" ]]; then
        echo "failed to read the L2 genesis block" >&2
        exit 1
    fi
    # daFootprintGasScalar is unknown to the op-succinct images.
    jq --arg l2 "$l2_hash" --argjson time "$((l2_time))" '
        .genesis.l2 = {hash: $l2, number: 0}
        | .genesis.l2_time = $time
        | del(.genesis.system_config.daFootprintGasScalar)
    ' /deployment/l2-config/rollup.json >/config/rollup.json
    echo "Rollup config ready (L2 genesis ${l2_hash})"
    ;;

geth)
    if [[ ! -d /data/geth ]]; then
        geth --gcmode=archive init --state.scheme=hash --datadir=/data/geth /config/genesis.json
    fi
    exec geth \
        --datadir=/data/geth \
        --gcmode=archive \
        --state.scheme=hash \
        --networkid="${L2_CHAIN_ID}" \
        --http \
        --http.addr=0.0.0.0 \
        --http.port="${OP_HTTP_PORT}" \
        --http.api=eth,net,web3,debug,admin,txpool \
        --http.vhosts='*' \
        --http.corsdomain='*' \
        --authrpc.addr=0.0.0.0 \
        --authrpc.port="${OP_ENGINE_PORT}" \
        --authrpc.vhosts='*' \
        --authrpc.jwtsecret=/config/jwt.txt \
        --rollup.disabletxpoolgossip=true \
        --rollup.halt=major \
        --nodiscover \
        --syncmode=full
    ;;

*)
    echo "unknown MODE: $MODE (want genesis, rollup or geth)" >&2
    exit 1
    ;;
esac
