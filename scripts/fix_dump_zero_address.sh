#!/usr/bin/env bash
# Fix the zero address entry in the Celo L1 state dump.
#
# The dump tool omits the "address" field for the zero address
# (0x0000000000000000000000000000000000000000) and instead includes
# a "key" field with its keccak256 hash. This replaces the "key"
# with the correct "address" field.

set -euo pipefail

if [ $# -ne 1 ]; then
    echo "Usage: $0 <dump.json>" >&2
    exit 1
fi

input="$1"

if [ ! -f "$input" ]; then
    echo "Error: file not found: $input" >&2
    exit 1
fi

sed -i '' 's/"key":"0x5380c7b7ae81a58eb98d9c78de4a1fd7fd9535fc953ed2be602daaa41767312a"/"address":"0x0000000000000000000000000000000000000000"/' \
    "$input"

echo "Fixed zero address in: $input"
