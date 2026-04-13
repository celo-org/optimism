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
  --chain celo \
  --datadir=/path/to/datadir \
  --storage.v2 \
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

## Logs

From run Mon, 13 April

```
~/P/optimism (palango/reth-import) $ ./scripts/fix_dump_zero_address.sh  ~/Downloads/l1-final-state.json
Before: {"balance":"324222713619168179923","nonce":0,"root":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","codeHash":"0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470","key":"0x5380c7b7ae81a58eb98d9c78de4a1fd7fd9535fc953ed2be602daaa41767312a"}
After:  {"balance":"324222713619168179923","nonce":0,"root":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421","codeHash":"0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470","address":"0x0000000000000000000000000000000000000000"}
mise WARN  missing: golangci-lint@2.8.0 rust@1.92.0 op-acceptor@op-acceptor/v3.8.3
~/P/optimism (palango/reth-import) $ python3 ./scripts/append_l2_allocs.py --update-state-root ~/Downloads/l1-final-state.json
Copying /Users/paul/Downloads/l1-final-state.json to /Users/paul/Downloads/l1-final-state.json.with-allocs.jsonl (45.8 GB)...
  45.8 / 45.8 GB
Downloading l2-allocs.json from https://storage.googleapis.com/cel2-rollup-files/celo/l2-allocs.json...
Appending 2080 accounts to /Users/paul/Downloads/l1-final-state.json.with-allocs.jsonl...
Replacing state root: 0x3817f87742aaeb96219d6c1fed8b32a562b59065116c49a53693abfd118838bc -> 0xed980641a4bd4d2e84c6c8db980b7f05e95733c92be2e0045db3735efeb1d807
Old first line: b'{"root":"0x3817f87742aaeb96219d6c1fed8b32a562b59065116c49a53693abfd118838bc"}\n'
New first line: b'{"root":"0xed980641a4bd4d2e84c6c8db980b7f05e95733c92be2e0045db3735efeb1d807"}\n'
Done.
```
