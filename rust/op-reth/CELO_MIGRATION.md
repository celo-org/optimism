# Celo L2 Migration: reth State Import

This document describes how to initialize an op-reth node with the Celo L1 state dump for the Cel2 migration.

## Overview

The Celo L2 migration imports the full Celo L1 state (pre-migration) into reth at migration block `31,056,500`. The state dump contains all existing Celo L1 accounts, and the L2 allocs (OP Stack predeploys and Celo-specific contracts injected at migration time) must be appended to it.

## Prerequisites

- A built `op-reth` binary with the Celo migration header support (branch `palango/reth-import`)

## Step 1: Download and Decompress the L1 State Dump

The compressed state dump (~15GB zstd) is available at:
https://storage.googleapis.com/cel2-rollup-files/celo/l1-final-state.json.zst

Download and decompress (~56GB JSONL):

```bash
curl -O https://storage.googleapis.com/cel2-rollup-files/celo/l1-final-state.json.zst
zstd -d l1-final-state.json.zst
```

## Step 2: Prepare the State Dump

The L1 state dump must be merged with the L2 migration allocs before import. The merge script applies the same logic as the Go migration (`applyAllocsToState` + `setupUnreleasedTreasury`):

- Fixes the zero address dump bug (missing `address` field)
- Merges L2 alloc accounts with existing L1 accounts using allowlist rules
- Sets the treasury balance to the post-migration value
- Updates the state root to `CEL2_HEADER.state_root`

```bash
python3 scripts/append_l2_allocs.py /path/to/l1-final-state.json
```

This creates `/path/to/l1-final-state.json.with-allocs.jsonl` without modifying the original. You can also specify a custom output path:

```bash
python3 scripts/append_l2_allocs.py /path/to/l1-final-state.json /path/to/output.jsonl
```

### State dump format

The dump is JSONL with:
- Line 1: `{"root": "0x..."}` — must match `CEL2_HEADER.state_root` (reth checks this before importing)
- Lines 2+: one account per line with `address`, `balance`, `nonce`, `code`, `storage` fields

Extra fields from the L1 dump (`root`, `codeHash`) are silently ignored due to serde `flatten` behavior. Balance can be a decimal string (e.g. `"157500000000000"`) or hex (`"0x..."`). Nonce can be a JSON integer or hex string.

After importing all accounts, reth computes the state root from the trie and verifies it also matches `CEL2_HEADER.state_root` (`0xed980641...`).

## Step 3: Initialize reth

The dummy chain creation opens many static file segments. On macOS, increase the file descriptor limit before running:

```bash
ulimit -n 10240
```

Then run `op-reth init-state` with the `--without-ovm` flag and the prepared state dump:

```bash
op-reth init-state \
  --chain celo \
  --datadir=/path/to/datadir \
  --without-ovm \
  /path/to/celo-l1-dump-final-state.json.with-allocs.jsonl
```

The `--without-ovm` flag with Celo mainnet chain ID (`42220`) will:

1. Create a dummy chain up to block `31,056,499`
2. Append the hardcoded Cel2 migration header at block `31,056,500`
3. Import all accounts from the state dump
4. Compute the state root and verify it matches the migration header

## Logs

Successful import (14 April 2026):

```
$ python3 scripts/append_l2_allocs.py /mnt/ssd/celo/l1-final-state.json
Downloading l2-allocs.json from https://storage.googleapis.com/cel2-rollup-files/celo/l2-allocs.json...
Loaded 2080 L2 alloc accounts.
Processing /mnt/ssd/celo/l1-final-state.json -> /mnt/ssd/celo/l1-final-state.json.with-allocs.jsonl...
State root: 0x3817f87742aaeb96219d6c1fed8b32a562b59065116c49a53693abfd118838bc -> 0xed980641a4bd4d2e84c6c8db980b7f05e95733c92be2e0045db3735efeb1d807
  Merge 0xfb1bffc9d739b8d520daf37df666da4c687191ea: case 2 (allowlist skip, keep L1)
  Merge 0x000000000022d473030f116ddee9f6b43ac78ba3: case 2 (allowlist skip, keep L1)
  Merge 0x914d7fec6aac8cd542e72bca78b30650d45643d7: case 2 (allowlist skip, keep L1)
  5000000 accounts processed...
  Merge 0xefc2c1444ebcc4db75e7613d20c6a62ff67a167c: case 2 (allowlist skip, keep L1)
  Merge 0x4200000000000000000000000000000000000006: case 5 (balance-only L1, balance=23086547216520198)
  Merge 0x4e59b44847b379578588920ca78fbf26c0b4956c: case 2 (allowlist skip, keep L1)
  Merge 0x0000000071727de22e5e9d8baf0edac6f37da032: case 2 (allowlist skip, keep L1)
Fixed zero address entry (key -> address)
  10000000 accounts processed...
  Merge 0xba5ed099633d3b313e4d5f7bdc1305d3c28ba5ed: case 2 (allowlist skip, keep L1)
  Merge 0x69f4d1788e39c87893c980c06edf4b7f686e2938: case 2 (allowlist skip, keep L1)
  Merge 0xa1dabef33b3b82c7814b6d82a79e50f4ac44102b: case 2 (allowlist skip, keep L1)
  Merge 0x998739bfdaadde7c933b942a68053933098f9eda: case 2 (allowlist skip, keep L1)
  Merge 0x5ff137d4b0fdcd49dca30c7cf57e578a026d2789: case 2 (allowlist skip, keep L1)
  15000000 accounts processed...
  Merge 0xca11bde05977b3631167028862be2a173976ca11: case 2 (allowlist skip, keep L1)
  Merge 0x13b0d85ccb8bf860b6b79af3029fca081ae9bef2: case 3 (allowlist, overwrite code only)
  20000000 accounts processed...
CeloToken totalSupply: 707425076124471169833280651
Treasury L1 balance: 0
  Merge 0x7fc98430eaedbb6070b35b39d798725049088348: case 2 (allowlist skip, keep L1)
Treasury 0x7a8c7a833565fc428cdfba20fe03fafb178a434f: L1 only, set balance

Processed 23853250 L1 accounts
Merged 15 existing accounts with L2 allocs
Appended 2065 new accounts from L2 allocs
Treasury balance: 292574923875528830166719349 (0xf20326676e9b4664ff5b75)
Treasury balance verified: 1000000000000000000000000000 - 707425076124471169833280651 + 0 = 292574923875528830166719349

Done.
paul_lange_clabs_co@paul-reth-import-test:/mnt/ssd/optimism/rust$ ./target/release/op-reth init-state --chain celo --datadir=/mnt/ssd/data --without-ovm ../../celo/l1-final-state.json.with-allocs.jsonl
2026-04-15T10:05:34.713366Z  INFO Initialized tracing, debug log directory: /home/paul_lange_clabs_co/.cache/reth/logs/celo
2026-04-15T10:05:34.714155Z  INFO Reth init-state starting chain_name="Celo mainnet"
2026-04-15T10:05:34.716359Z  INFO Opening storage db_path="/mnt/ssd/data/db" sf_path="/mnt/ssd/data/static_files"
2026-04-15T10:05:34.719588Z  INFO check_consistency{read_only=false}: Verifying storage consistency.
2026-04-15T10:05:34.724666Z  INFO Setting up dummy EVM chain before importing state. new_tip=NumHash { number: 31056500, hash: 0x7586014e20c69b3fa7c9070baf1a7edd95833db57853126f32593b455da2e5c5 }
2026-04-15T10:06:15.278653Z  INFO Appending first valid block.
2026-04-15T10:06:15.298941Z  INFO Set up finished.
2026-04-15T10:06:15.302261Z  INFO Initiating state dump
2026-04-15T10:06:18.441280Z  INFO parsed_new_accounts=285228
2026-04-15T10:06:19.105206Z  INFO parsed_new_accounts=570456
...
2026-04-15T10:12:19.693707Z  INFO parsed_new_accounts=23388696
2026-04-15T10:12:21.001389Z  INFO parsed_new_accounts=23673924
2026-04-15T10:12:25.469480Z  INFO Writing accounts to db total_inserted_accounts=285229
2026-04-15T10:12:35.931102Z  INFO Writing accounts to db total_inserted_accounts=570457
...
2026-04-15T10:35:57.113326Z  INFO Writing accounts to db total_inserted_accounts=23855315
2026-04-15T10:36:02.674107Z  INFO All accounts written to database, starting state root computation (may take some time)
2026-04-15T10:47:18.447497Z  INFO Computed state root matches state root in state dump computed_state_root=0xed980641a4bd4d2e84c6c8db980b7f05e95733c92be2e0045db3735efeb1d807
2026-04-15T10:47:18.660820Z  INFO Genesis block written hash=0x7586014e20c69b3fa7c9070baf1a7edd95833db57853126f32593b455da2e5c5
```
