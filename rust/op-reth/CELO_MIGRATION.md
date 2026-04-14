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

## Storage V2 problems

During import I saw these problems, which might or might not be fixed by the last commits:

```
paul_lange_clabs_co@paul-reth-import-test:/mnt/ssd/optimism/rust$ ./target/release/op-reth init-state --chain celo --storage.v2 --datadir=/mnt/ssd/data2 --without-ovm ../../celo/l1-final-state.json.with-allocs.jsonl
2026-04-13T17:47:03.297813Z  INFO Initialized tracing, debug log directory: /home/paul_lange_clabs_co/.cache/reth/logs/celo
2026-04-13T17:47:03.298556Z  INFO Reth init-state starting chain_name="Celo mainnet"
2026-04-13T17:47:03.300401Z  INFO Opening storage db_path="/mnt/ssd/data2/db" sf_path="/mnt/ssd/data2/static_files"
2026-04-13T17:47:03.304392Z  INFO check_consistency{read_only=false}: Verifying storage consistency.
2026-04-13T17:47:03.329411Z  INFO Setting up dummy EVM chain before importing state. new_tip=NumHash { number: 31056500, hash: 0x7586014e20c69b3fa7c9070baf1a7edd95833db57853126f32593b455da2e5c5 }
2026-04-13T17:47:42.477127Z  INFO Appending first valid block.
2026-04-13T17:47:42.496685Z  INFO Set up finished.
2026-04-13T17:47:42.504544Z  INFO Initiating state dump
2026-04-13T17:47:45.104737Z  INFO parsed_new_accounts=285228
2026-04-13T17:47:45.711377Z  INFO parsed_new_accounts=570456
2026-04-13T17:47:57.261171Z  INFO parsed_new_accounts=855684
2026-04-13T17:47:58.496431Z  INFO parsed_new_accounts=1140912
2026-04-13T17:47:59.430044Z  INFO parsed_new_accounts=1426140
2026-04-13T17:48:02.646269Z  INFO parsed_new_accounts=1711368
2026-04-13T17:48:13.589369Z  INFO parsed_new_accounts=1996596
2026-04-13T17:48:15.385231Z  INFO parsed_new_accounts=2281824
2026-04-13T17:48:16.705019Z  INFO parsed_new_accounts=2567052
2026-04-13T17:48:17.462988Z  INFO parsed_new_accounts=2852280
2026-04-13T17:48:23.275126Z  INFO parsed_new_accounts=3137508
2026-04-13T17:48:32.768201Z  INFO parsed_new_accounts=3422736
2026-04-13T17:48:34.661063Z  INFO parsed_new_accounts=3707964
2026-04-13T17:48:38.533981Z  INFO parsed_new_accounts=3993192
2026-04-13T17:48:41.768584Z  INFO parsed_new_accounts=4278420
2026-04-13T17:48:47.033822Z  INFO parsed_new_accounts=4563648
2026-04-13T17:48:50.381858Z  INFO parsed_new_accounts=4848876
2026-04-13T17:48:51.819680Z  INFO parsed_new_accounts=5134104
2026-04-13T17:48:55.970964Z  INFO parsed_new_accounts=5419332
2026-04-13T17:48:57.385912Z  INFO parsed_new_accounts=5704560
2026-04-13T17:48:58.125682Z  INFO parsed_new_accounts=5989788
2026-04-13T17:49:04.232916Z  INFO parsed_new_accounts=6275016
2026-04-13T17:49:09.814462Z  INFO parsed_new_accounts=6560244
2026-04-13T17:49:15.727191Z  INFO parsed_new_accounts=6845472
2026-04-13T17:49:22.318107Z  INFO parsed_new_accounts=7130700
2026-04-13T17:49:24.123644Z  INFO parsed_new_accounts=7415928
2026-04-13T17:49:25.129018Z  INFO parsed_new_accounts=7701156
2026-04-13T17:49:29.504810Z  INFO parsed_new_accounts=7986384
2026-04-13T17:49:31.402493Z  INFO parsed_new_accounts=8271612
2026-04-13T17:49:33.407966Z  INFO parsed_new_accounts=8556840
2026-04-13T17:49:35.641795Z  INFO parsed_new_accounts=8842068
2026-04-13T17:49:37.385763Z  INFO parsed_new_accounts=9127296
2026-04-13T17:49:39.493995Z  INFO parsed_new_accounts=9412524
2026-04-13T17:49:40.459440Z  INFO parsed_new_accounts=9697752
2026-04-13T17:49:46.032842Z  INFO parsed_new_accounts=9982980
2026-04-13T17:49:47.401970Z  INFO parsed_new_accounts=10268208
2026-04-13T17:49:49.418314Z  INFO parsed_new_accounts=10553436
2026-04-13T17:49:51.530573Z  INFO parsed_new_accounts=10838664
2026-04-13T17:49:52.442792Z  INFO parsed_new_accounts=11123892
2026-04-13T17:50:09.219271Z  INFO parsed_new_accounts=11409120
2026-04-13T17:50:20.278613Z  INFO parsed_new_accounts=11694348
2026-04-13T17:50:23.224165Z  INFO parsed_new_accounts=11979576
2026-04-13T17:50:27.012773Z  INFO parsed_new_accounts=12264804
2026-04-13T17:50:28.399742Z  INFO parsed_new_accounts=12550032
2026-04-13T17:50:30.683336Z  INFO parsed_new_accounts=12835260
2026-04-13T17:50:32.762288Z  INFO parsed_new_accounts=13120488
2026-04-13T17:50:34.190738Z  INFO parsed_new_accounts=13405716
2026-04-13T17:51:24.810656Z  INFO parsed_new_accounts=13690944
2026-04-13T17:51:25.755298Z  INFO parsed_new_accounts=13976172
2026-04-13T17:51:27.244202Z  INFO parsed_new_accounts=14261400
2026-04-13T17:51:31.164009Z  INFO parsed_new_accounts=14546628
2026-04-13T17:51:35.853427Z  INFO parsed_new_accounts=14831856
2026-04-13T17:51:37.652219Z  INFO parsed_new_accounts=15117084
2026-04-13T17:51:40.039563Z  INFO parsed_new_accounts=15402312
2026-04-13T17:51:47.758014Z  INFO parsed_new_accounts=15687540
2026-04-13T17:51:57.334543Z  INFO parsed_new_accounts=15972768
2026-04-13T17:52:03.216040Z  INFO parsed_new_accounts=16257996
2026-04-13T17:52:04.754878Z  INFO parsed_new_accounts=16543224
2026-04-13T17:52:08.651693Z  INFO parsed_new_accounts=16828452
2026-04-13T17:52:09.570126Z  INFO parsed_new_accounts=17113680
2026-04-13T17:52:12.884692Z  INFO parsed_new_accounts=17398908
2026-04-13T17:52:15.069946Z  INFO parsed_new_accounts=17684136
2026-04-13T17:52:25.383157Z  INFO parsed_new_accounts=17969364
2026-04-13T17:52:30.954210Z  INFO parsed_new_accounts=18254592
2026-04-13T17:52:32.045141Z  INFO parsed_new_accounts=18539820
2026-04-13T17:52:34.105039Z  INFO parsed_new_accounts=18825048
2026-04-13T17:52:36.552111Z  INFO parsed_new_accounts=19110276
2026-04-13T17:52:38.884921Z  INFO parsed_new_accounts=19395504
2026-04-13T17:52:41.235617Z  INFO parsed_new_accounts=19680732
2026-04-13T17:52:42.089006Z  INFO parsed_new_accounts=19965960
2026-04-13T17:52:43.592820Z  INFO parsed_new_accounts=20251188
2026-04-13T17:52:47.069978Z  INFO parsed_new_accounts=20536416
2026-04-13T17:52:50.628766Z  INFO parsed_new_accounts=20821644
2026-04-13T17:52:56.821686Z  INFO parsed_new_accounts=21106872
2026-04-13T17:52:58.242562Z  INFO parsed_new_accounts=21392100
2026-04-13T17:52:59.466898Z  INFO parsed_new_accounts=21677328
2026-04-13T17:53:28.901741Z  INFO parsed_new_accounts=21962556
2026-04-13T17:53:30.993659Z  INFO parsed_new_accounts=22247784
2026-04-13T17:53:34.306684Z  INFO parsed_new_accounts=22533012
2026-04-13T17:53:39.811133Z  INFO parsed_new_accounts=22818240
2026-04-13T17:53:41.428333Z  INFO parsed_new_accounts=23103468
2026-04-13T17:53:43.875438Z  INFO parsed_new_accounts=23388696
2026-04-13T17:53:45.168184Z  INFO parsed_new_accounts=23673924
2026-04-13T17:53:49.540768Z  INFO Writing accounts to db total_inserted_accounts=285229
Error: trying to append data to StorageChangeSets as block #31056500 but expected block #1

Location:
    /mnt/ssd/rust/cargo/git/checkouts/reth-e231042ee7db3fb7/564ffa5/crates/storage/db-common/src/init.rs:808:13

```

```
paul_lange_clabs_co@paul-reth-import-test:/mnt/ssd/optimism/rust$ ./target/release/op-reth init-state --chain celo --storage.v2 --datadir=/mnt/ssd/data2 --without-ovm ../../celo/l1-final-state.json.with-allocs.jsonl
2026-04-13T18:03:25.576184Z  INFO Initialized tracing, debug log directory: /home/paul_lange_clabs_co/.cache/reth/logs/celo
2026-04-13T18:03:25.576959Z  INFO Reth init-state starting chain_name="Celo mainnet"
2026-04-13T18:03:25.578916Z  INFO Opening storage db_path="/mnt/ssd/data2/db" sf_path="/mnt/ssd/data2/static_files"
2026-04-13T18:03:25.692962Z  INFO check_consistency{read_only=false}: Verifying storage consistency.
2026-04-13T18:03:25.695331Z  INFO check_consistency{read_only=false}:Checking consistency for segment{segment=Headers initial_highest_block=31056500 initial_highest_block=31056500 highest_block=31056500 highest_block=31056500  }:ensure_invariants{highest_static_file_entry=Some(31056500) highest_static_file_block=Some(31056500) table="Headers"}: Unwinding static file segment. from=31056500 to=0
2026-04-13T18:03:25.763329Z  INFO check_consistency{read_only=false}:Checking consistency for segment{segment=Transactions initial_highest_block=31056500 initial_highest_block=31056500 highest_block=31056500 highest_block=31056500  }:ensure_invariants{highest_static_file_entry=None highest_static_file_block=Some(31056500) table="Transactions"}: Unwinding static file segment. from=31056500 to=0
2026-04-13T18:03:25.788310Z  INFO check_consistency{read_only=false}:Checking consistency for segment{segment=Receipts initial_highest_block=31056500 initial_highest_block=31056500 highest_block=31056500 highest_block=31056500  }:ensure_invariants{highest_static_file_entry=None highest_static_file_block=Some(31056500) table="Receipts"}: Unwinding static file segment. from=31056500 to=0
2026-04-13T18:03:25.808574Z  INFO check_consistency{read_only=false}:Checking consistency for segment{segment=TransactionSenders initial_highest_block=31056500 initial_highest_block=31056500 highest_block=31056500 highest_block=31056500  }:ensure_invariants{highest_static_file_entry=None highest_static_file_block=Some(31056500) table="TransactionSenders"}: Unwinding static file segment. from=31056500 to=0
2026-04-13T18:03:25.828522Z  INFO check_consistency{read_only=false}:Checking consistency for segment{segment=AccountChangeSets initial_highest_block=0 initial_highest_block=0 highest_block=31056499 highest_block=31056499}: Setting unwind target. unwind_target=31056499
2026-04-13T18:03:25.828549Z  INFO check_consistency{read_only=false}:Checking consistency for segment{segment=AccountChangeSets initial_highest_block=0 initial_highest_block=0 highest_block=31056499 highest_block=31056499  }:ensure_invariants{highest_static_file_entry=None highest_static_file_block=Some(31056499) table="AccountChangeSets"}: Unwinding static file segment. from=31056499 to=0
Error: not able to find AccountChangeSets static file for block number 0
```

```
paul_lange_clabs_co@paul-reth-import-test:/mnt/ssd/optimism/rust$ ./target/release/op-reth init-state --chain celo --storage.v2 --datadir=/mnt/ssd/data2 --without-ovm ../../celo/l1-final-state.json.with-allocs.jsonl
2026-04-13T18:05:19.604988Z  INFO Initialized tracing, debug log directory: /home/paul_lange_clabs_co/.cache/reth/logs/celo
2026-04-13T18:05:19.605816Z  INFO Reth init-state starting chain_name="Celo mainnet"
2026-04-13T18:05:19.607707Z  INFO Opening storage db_path="/mnt/ssd/data2/db" sf_path="/mnt/ssd/data2/static_files"
2026-04-13T18:05:19.610502Z  INFO check_consistency{read_only=false}: Verifying storage consistency.
2026-04-13T18:05:19.619625Z  INFO Setting up dummy EVM chain before importing state. new_tip=NumHash { number: 31056500, hash: 0x7586014e20c69b3fa7c9070baf1a7edd95833db57853126f32593b455da2e5c5 }
2026-04-13T18:05:58.511018Z  INFO Appending first valid block.
2026-04-13T18:05:58.530085Z  INFO Set up finished.
2026-04-13T18:05:58.532919Z  WARN Failed to load static file jar path="/mnt/ssd/data2/static_files/static_file_account-change-sets_31000000_31499999" e=No such file or directory (os error 2)
Error: failed to open file "/mnt/ssd/data2/static_files/static_file_account-change-sets_31000000_31499999.conf": No such file or directory (os error 2)

Caused by:
    No such file or directory (os error 2)
```
