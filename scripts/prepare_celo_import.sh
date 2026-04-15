#!/usr/bin/env bash
# Prepare and run the Celo L1 state import into op-reth.
# See rust/op-reth/CELO_MIGRATION.md for details.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DUMP_URL="https://storage.googleapis.com/cel2-rollup-files/celo/l1-final-state.json.zst"

if [ $# -ne 2 ]; then
    echo "Usage: $0 <workdir> <datadir>" >&2
    exit 1
fi

workdir="$1"
datadir="$2"

mkdir -p "$workdir"
dump_zst="$workdir/l1-final-state.json.zst"
dump="$workdir/l1-final-state.json"
dump_allocs="$dump.with-allocs.jsonl"

# Step 0: Download and decompress
if [ ! -f "$dump" ]; then
    if [ ! -f "$dump_zst" ]; then
        echo "==> Downloading L1 state dump..."
        curl -o "$dump_zst" "$DUMP_URL"
    fi
    echo "==> Decompressing..."
    zstd -d "$dump_zst"
else
    echo "==> L1 state dump already exists: $dump"
fi

# Step 1: Append L2 allocs and update state root
if [ ! -f "$dump_allocs" ]; then
    echo "==> Appending L2 allocs..."
    python3 "$SCRIPT_DIR/append_l2_allocs.py" "$dump"
else
    echo "==> Allocs file already exists: $dump_allocs"
fi

# Step 2: Initialize reth
echo "==> Initializing reth (ulimit -n 10240)..."
ulimit -n 10240
op-reth init-state \
    --chain celo \
    --datadir="$datadir" \
    --without-ovm \
    "$dump_allocs"

echo "==> Done."
