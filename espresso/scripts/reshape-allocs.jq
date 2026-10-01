#!/bin/bash
# Converts output of espresso-dev-node launched with
# 'ESPRESSO_DEV_NODE_L1_DEPLOYMENT=dump' to form suitable
# for e2e testing harness.
# Usage:
# ./scripts/reshape-allocs.jq /path/to/devnode/generated/allocs.json > environment/allocs.json

jq -S '
# pad hex-encoded U256 with leading zeroes to full
# 32 bytes (e.g. "0x1" -> "0x0000..0001" with 63 zeroes)
def pad_hex: .[2:] as $hex
  | (64 - ($hex | length)) as $padding
  | "0x" + ("0" * $padding) + $hex ;

# Reshape the input
. | map_values({
    state:  {
      nonce: .nonce,
      code: .code,
      balance: .balance,
      storage: .storage | with_entries({key: .key|pad_hex, value : .value|pad_hex}),
    },
    name: .name,
})
' "$@"
