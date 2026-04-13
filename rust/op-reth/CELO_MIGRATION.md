# Celo L2 Migration: reth State Import

This document describes how to initialize an op-reth node with the Celo L1 state dump for the Cel2 migration.

## Overview

The Celo L2 migration imports the full Celo L1 state (pre-migration) into reth at migration block `31,056,500`. The state dump contains all existing Celo L1 accounts, and the L2 allocs (OP Stack predeploys and Celo-specific contracts injected at migration time) must be appended to it.

## Prerequisites

- A built `op-reth` binary with the Celo migration header support (branch `palango/reth-import`)
- A Celo chain spec file

## Step 0: Download and Decompress the L1 State Dump

The compressed state dump (~5GB zstd) is available at:
https://storage.googleapis.com/cel2-rollup-files/celo/l1-final-state.json.zst

Download and decompress (~50GB JSONL):

```bash
curl -O https://storage.googleapis.com/cel2-rollup-files/celo/l1-final-state.json.zst
zstd -d l1-final-state.json.zst
```

## Step 1: Fix the Zero Address Entry

The L1 state dump has a bug where the zero address (`0x0000000000000000000000000000000000000000`) is missing its `address` field and instead has a `key` field containing the keccak256 hash of the address. This must be fixed before import.

```bash
./scripts/fix_dump_zero_address.sh /path/to/celo-l1-dump-final-state.json
```

This modifies the file in place.

You can verify the fix with `scripts/check_dump_addresses.py`.

## Step 2: Prepare the State Dump

The state dump contains Celo L1 accounts but is missing the L2 allocs (OP Stack predeploys, bridge contracts, etc.) that are injected during the migration. These must be appended.

The L2 allocs are published at:
https://storage.googleapis.com/cel2-rollup-files/celo/l2-allocs.json

Run the script to download the allocs and append them to a copy of the state dump. The script also updates the state root on line 1 to match `CEL2_HEADER.state_root`, which reth verifies before importing.

```bash
python3 scripts/append_l2_allocs.py /path/to/celo-l1-dump-final-state.json
```

This creates `/path/to/celo-l1-dump-final-state.json.with-allocs.jsonl` without modifying the original. You can also specify a custom output path:

```bash
python3 scripts/append_l2_allocs.py /path/to/celo-l1-dump-final-state.json /path/to/output.jsonl
```

### State dump format

The dump is JSONL with:
- Line 1: `{"root": "0x..."}` — must match `CEL2_HEADER.state_root` (reth checks this before importing)
- Lines 2+: one account per line with `address`, `balance`, `nonce`, `code`, `storage` fields

Extra fields from the L1 dump (`root`, `codeHash`) are silently ignored due to serde `flatten` behavior. Balance can be a decimal string (e.g. `"157500000000000"`) or hex (`"0x..."`). Nonce can be a JSON integer or hex string.

After importing all accounts, reth computes the state root from the trie and verifies it also matches `CEL2_HEADER.state_root` (`0xed980641...`).

## Step 3: Initialize reth

The dummy chain creation opens many static file segments. Increase the file descriptor limit before running:

```bash
ulimit -n 10240
```

Then run `op-reth init-state` with the `--without-ovm` flag and the prepared state dump:

```bash
op-reth init-state \
  --chain /path/to/celo-chainspec.json \
  --datadir=/path/to/datadir \
  --without-ovm \
  /path/to/celo-l1-dump-final-state.json.with-allocs.jsonl
```

The `--without-ovm` flag with Celo mainnet chain ID (`42220`) will:

1. Create a dummy chain up to block `31,056,499`
2. Append the hardcoded Cel2 migration header at block `31,056,500`
3. Import all accounts from the state dump
4. Compute the state root and verify it matches the migration header

## Network Config & Assets

All Celo mainnet migration artifacts are available at:

| Asset | URL |
|-------|-----|
| L1 state dump | https://storage.googleapis.com/cel2-rollup-files/celo/l1-final-state.json.zst |
| Deploy config | https://storage.googleapis.com/cel2-rollup-files/celo/config.json |
| L1 contract addresses | https://storage.googleapis.com/cel2-rollup-files/celo/deployment-l1.json |
| L2 allocs | https://storage.googleapis.com/cel2-rollup-files/celo/l2-allocs.json |
| Rollup config | https://storage.googleapis.com/cel2-rollup-files/celo/rollup.json |
| Genesis (snap sync) | https://storage.googleapis.com/cel2-rollup-files/celo/genesis.json |
| Migrated chaindata | https://storage.googleapis.com/cel2-rollup-files/celo/celo-mainnet-migrated-chaindata.tar.zst |

## Open Questions

- The state dump may have storage values without `0x` prefix — needs verification that reth's `Bytes` deserializer handles non-prefixed hex strings correctly.
