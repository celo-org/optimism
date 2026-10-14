#!/usr/bin/env bash
# Checks the change-detection patterns in routing.yml against known paths.
# collect-params.sh matches each changed file with `grep -E`; this script runs
# the same match on single paths and asserts which patterns they hit.
#
# Usage:
#   bash .circleci/scripts/test-change-patterns.sh
#
# Requires: yq (same version used in CI)
set -euo pipefail

ROUTING="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/routing.yml"

PASS=0
FAIL=0

# expect <section> <pattern name> <path> <true|false>
expect() {
  local section="${1}" name="${2}" path="${3}" want="${4}"
  local pattern got
  pattern=$(yq -r ".change_patterns.${section}.\"${name}\"" "${ROUTING}")
  if grep -qE "${pattern}" <<< "${path}"; then got=true; else got=false; fi
  if [[ "${got}" == "${want}" ]]; then
    PASS=$((PASS + 1))
  else
    FAIL=$((FAIL + 1))
    echo "FAIL: ${section}.${name} on '${path}': got ${got}, want ${want} (pattern: ${pattern})"
  fi
}

# rust-ci runs the kona and op-reth snapshot checks, which regenerate their
# committed files from the superchain-registry submodule. A bump of the
# submodule alone must still run them.
expect any rust_changes_detected "superchain-registry" true
expect any rust_changes_detected "rust/kona/crates/protocol/registry/etc/configs.json" true
expect any rust_changes_detected "op-e2e/config/init.go" true
expect any rust_changes_detected ".circleci/routing.yml" true
expect any rust_changes_detected "mise.toml" true
expect any rust_changes_detected "superchain-registry/superchain/configs/mainnet/celo.toml" false
expect any rust_changes_detected "op-core/superchain/configs.go" false
expect any rust_changes_detected "justfile" false
expect any rust_changes_detected "op-node/rollup/types.go" false
expect any rust_changes_detected "docs/public-docs/index.md" false

expect any contracts_changed "packages/contracts-bedrock/src/L1/OptimismPortal2.sol" true
expect any contracts_changed ".github/workflows/renovate.yaml" true
expect any contracts_changed "op-node/rollup/types.go" false

expect all only_docs_changes "docs/public-docs/index.md" true
expect all only_docs_changes "docs/ai/derivation.md" false

echo "${PASS} passed, ${FAIL} failed"
[[ "${FAIL}" -eq 0 ]]
