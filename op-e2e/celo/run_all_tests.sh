#!/bin/bash
#shellcheck disable=SC1091
set -eo pipefail

SCRIPT_DIR=$(readlink -f "$(dirname "$0")")
TEST_GLOB=$1
spawn_devnet=${SPAWN_DEVNET:-true}

if [[ $spawn_devnet != false ]]; then
  ## Start geth
  cd "$SCRIPT_DIR/../.." || exit 1
  trap 'cd "$SCRIPT_DIR/../.." && make devnet-down' EXIT # kill bg job at exit
  DEVNET_L2OO=false GENERIC_ALTDA=true DEVNET_ALTDA=true DEVNET_CELO=true make devnet-up
fi

cd "$SCRIPT_DIR" || exit 1
source "$SCRIPT_DIR/shared.sh"

# Wait for geth to be ready
for _ in {1..10}; do
  if cast block &>/dev/null; then
    echo geth ready
    break
  fi
  sleep 0.2
done

# There's a problem with geth return errors on the first transaction sent.
# See https://github.com/ethereum/web3.py/issues/3212
# To work around this, send a transaction before running tests
cast send --async --json --private-key "$ACC_PRIVKEY" "$TOKEN_ADDR" 'transfer(address to, uint256 value) returns (bool)' 0x000000000000000000000000000000000000dEaD 100
sleep 1

## Run tests
echo Start tests
failures=0
tests=0
for f in test_*"$TEST_GLOB"*; do
  echo -e "\nRun $f"
  if "./$f"; then
    tput setaf 2 || true
    echo "PASS $f"
  else
    tput setaf 1 || true
    echo "FAIL $f ❌"
    ((failures++)) || true
  fi
  tput sgr0 || true
  ((tests++)) || true
done

## Final summary
echo
if [[ $failures -eq 0 ]]; then
  tput setaf 2 || true
  echo All tests succeeded!
else
  tput setaf 1 || true
  echo "$failures/$tests" failed.
fi
tput sgr0 || true
exit "$failures"
