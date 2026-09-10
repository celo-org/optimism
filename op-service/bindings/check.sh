#!/usr/bin/env bash
# Regenerates the bindings and fails if the result differs from what is committed.
# Expects built contracts; `just check-bindings` builds them.

set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

op-service/bindings/gen.sh

if changed=$(git status --porcelain -- ':(glob)op-service/bindings/*/*.go') && [[ -n $changed ]]; then
  echo "bindings are out of date; review and commit the regenerated files:" >&2
  echo "$changed" >&2
  exit 1
fi
echo "bindings are up to date"
