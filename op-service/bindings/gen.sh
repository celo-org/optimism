#!/usr/bin/env bash
#
# Regenerates the ABI-only Go bindings under op-service/bindings from the
# contracts-bedrock forge artifacts. Run via `just gen-bindings`, which builds
# the contracts first.
#
# These are the bindings production services import; they live here so that no
# production binary depends on op-e2e. One package per contract, so the struct
# types abigen emits for shared Solidity types cannot collide. ABI-only because
# nothing in Go deploys these contracts, and BatchAuthenticator's bytecode is
# not reproducible across checkouts (it compiles in a source from a submodule
# of a submodule); its ABI is.

set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

# To bind another contract, add it here and run `just gen-bindings`.
CONTRACTS=(BatchAuthenticator SystemConfig)

ARTIFACTS=packages/contracts-bedrock/forge-artifacts
OUT=op-service/bindings

# Start from nothing, so a contract removed from CONTRACTS disappears too.
rm -rf "${OUT:?}"/*/

for contract in "${CONTRACTS[@]}"; do
  artifact=$ARTIFACTS/$contract.sol/$contract.json
  if [[ ! -f $artifact ]]; then
    echo "error: $artifact not found; build the contracts first" >&2
    exit 1
  fi
  pkg=$(tr '[:upper:]' '[:lower:]' <<< "$contract")
  mkdir -p "$OUT/$pkg"
  jq .abi "$artifact" | abigen --abi - --pkg "$pkg" --type "$contract" --out "$OUT/$pkg/$pkg.go"
done

echo "regenerated ${#CONTRACTS[@]} bindings in $OUT"
