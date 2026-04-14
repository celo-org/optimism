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

## Step 1: Prepare the State Dump

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

## Step 2: Initialize reth

The dummy chain creation opens many static file segments. Increase the file descriptor limit before running:

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

Previous run (Mon 13 April) — state root mismatch, before alloc merge fix:

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

```
paul_lange_clabs_co@paul-reth-import-test:/mnt/ssd/optimism/rust$ ./target/release/op-reth init-state --chain celo --datadir=/mnt/ssd/chain
-data --without-ovm ../../celo/l1-final-state.json.with-allocs.jsonl
2026-04-13T16:42:18.812534Z  INFO Initialized tracing, debug log directory: /home/paul_lange_clabs_co/.cache/reth/logs/celo
2026-04-13T16:42:18.815639Z  INFO Reth init-state starting chain_name="Celo mainnet"
2026-04-13T16:42:18.820852Z  INFO Opening storage db_path="/mnt/ssd/chain-data/db" sf_path="/mnt/ssd/chain-data/static_files"
2026-04-13T16:42:18.831917Z  INFO check_consistency{read_only=false}: Verifying storage consistency.
2026-04-13T16:42:18.848072Z  INFO Setting up dummy EVM chain before importing state. new_tip=NumHash { number: 31056500, hash: 0x7586014e20c69b3fa7c9070baf1a7edd95833db57853126f32593b455da2e5c5 }
2026-04-13T16:42:56.827581Z  INFO Appending first valid block.
2026-04-13T16:42:56.848294Z  INFO Set up finished.
2026-04-13T16:42:56.850124Z  INFO Initiating state dump
2026-04-13T16:42:59.503582Z  INFO parsed_new_accounts=285228
2026-04-13T16:43:00.123970Z  INFO parsed_new_accounts=570456
2026-04-13T16:43:12.435430Z  INFO parsed_new_accounts=855684
2026-04-13T16:43:13.720443Z  INFO parsed_new_accounts=1140912
2026-04-13T16:43:14.683206Z  INFO parsed_new_accounts=1426140
2026-04-13T16:43:17.947835Z  INFO parsed_new_accounts=1711368
2026-04-13T16:43:28.915803Z  INFO parsed_new_accounts=1996596
2026-04-13T16:43:30.585443Z  INFO parsed_new_accounts=2281824
2026-04-13T16:43:31.989111Z  INFO parsed_new_accounts=2567052
2026-04-13T16:43:32.752545Z  INFO parsed_new_accounts=2852280
2026-04-13T16:43:38.433548Z  INFO parsed_new_accounts=3137508
2026-04-13T16:43:48.029585Z  INFO parsed_new_accounts=3422736
2026-04-13T16:43:49.896711Z  INFO parsed_new_accounts=3707964
2026-04-13T16:43:52.651012Z  INFO parsed_new_accounts=3993192
2026-04-13T16:43:54.843911Z  INFO parsed_new_accounts=4278420
2026-04-13T16:44:00.074544Z  INFO parsed_new_accounts=4563648
2026-04-13T16:44:04.177806Z  INFO parsed_new_accounts=4848876
2026-04-13T16:44:05.611900Z  INFO parsed_new_accounts=5134104
2026-04-13T16:44:11.530471Z  INFO parsed_new_accounts=5419332
2026-04-13T16:44:15.545082Z  INFO parsed_new_accounts=5704560
2026-04-13T16:44:16.486635Z  INFO parsed_new_accounts=5989788
2026-04-13T16:44:22.514600Z  INFO parsed_new_accounts=6275016
2026-04-13T16:44:27.429958Z  INFO parsed_new_accounts=6560244
2026-04-13T16:44:32.588978Z  INFO parsed_new_accounts=6845472
2026-04-13T16:44:35.638466Z  INFO parsed_new_accounts=7130700
2026-04-13T16:44:37.554623Z  INFO parsed_new_accounts=7415928
2026-04-13T16:44:39.176799Z  INFO parsed_new_accounts=7701156
2026-04-13T16:44:41.472247Z  INFO parsed_new_accounts=7986384
2026-04-13T16:44:43.858357Z  INFO parsed_new_accounts=8271612
2026-04-13T16:44:48.757872Z  INFO parsed_new_accounts=8556840
2026-04-13T16:44:52.484596Z  INFO parsed_new_accounts=8842068
2026-04-13T16:44:54.204385Z  INFO parsed_new_accounts=9127296
2026-04-13T16:44:56.270181Z  INFO parsed_new_accounts=9412524
2026-04-13T16:44:57.221695Z  INFO parsed_new_accounts=9697752
2026-04-13T16:45:02.768897Z  INFO parsed_new_accounts=9982980
2026-04-13T16:45:04.131802Z  INFO parsed_new_accounts=10268208
2026-04-13T16:45:05.862055Z  INFO parsed_new_accounts=10553436
2026-04-13T16:45:07.952379Z  INFO parsed_new_accounts=10838664
2026-04-13T16:45:08.844751Z  INFO parsed_new_accounts=11123892
2026-04-13T16:45:27.148172Z  INFO parsed_new_accounts=11409120
2026-04-13T16:45:34.429270Z  INFO parsed_new_accounts=11694348
2026-04-13T16:45:37.374610Z  INFO parsed_new_accounts=11979576
2026-04-13T16:45:41.062672Z  INFO parsed_new_accounts=12264804
2026-04-13T16:45:42.439363Z  INFO parsed_new_accounts=12550032
2026-04-13T16:45:44.713347Z  INFO parsed_new_accounts=12835260
2026-04-13T16:45:47.052955Z  INFO parsed_new_accounts=13120488
2026-04-13T16:45:48.486462Z  INFO parsed_new_accounts=13405716
2026-04-13T16:46:40.782331Z  INFO parsed_new_accounts=13690944
2026-04-13T16:46:41.731557Z  INFO parsed_new_accounts=13976172
2026-04-13T16:46:43.233270Z  INFO parsed_new_accounts=14261400
2026-04-13T16:46:47.344231Z  INFO parsed_new_accounts=14546628
2026-04-13T16:46:52.164761Z  INFO parsed_new_accounts=14831856
2026-04-13T16:46:53.917277Z  INFO parsed_new_accounts=15117084
2026-04-13T16:46:56.287892Z  INFO parsed_new_accounts=15402312
2026-04-13T16:47:03.937836Z  INFO parsed_new_accounts=15687540
2026-04-13T16:47:13.109870Z  INFO parsed_new_accounts=15972768
2026-04-13T16:47:18.574494Z  INFO parsed_new_accounts=16257996
2026-04-13T16:47:20.473239Z  INFO parsed_new_accounts=16543224
2026-04-13T16:47:24.387829Z  INFO parsed_new_accounts=16828452
2026-04-13T16:47:25.282035Z  INFO parsed_new_accounts=17113680
2026-04-13T16:47:28.572755Z  INFO parsed_new_accounts=17398908
2026-04-13T16:47:30.714713Z  INFO parsed_new_accounts=17684136
2026-04-13T16:47:40.000098Z  INFO parsed_new_accounts=17969364
2026-04-13T16:47:43.471973Z  INFO parsed_new_accounts=18254592
2026-04-13T16:47:44.565534Z  INFO parsed_new_accounts=18539820
2026-04-13T16:47:48.171336Z  INFO parsed_new_accounts=18825048
2026-04-13T16:47:53.268887Z  INFO parsed_new_accounts=19110276
2026-04-13T16:47:55.599853Z  INFO parsed_new_accounts=19395504
2026-04-13T16:47:57.916394Z  INFO parsed_new_accounts=19680732
2026-04-13T16:47:58.750556Z  INFO parsed_new_accounts=19965960
2026-04-13T16:48:00.229423Z  INFO parsed_new_accounts=20251188
2026-04-13T16:48:03.678579Z  INFO parsed_new_accounts=20536416
2026-04-13T16:48:06.687550Z  INFO parsed_new_accounts=20821644
2026-04-13T16:48:10.952793Z  INFO parsed_new_accounts=21106872
2026-04-13T16:48:13.010899Z  INFO parsed_new_accounts=21392100
2026-04-13T16:48:16.197875Z  INFO parsed_new_accounts=21677328
2026-04-13T16:48:46.673103Z  INFO parsed_new_accounts=21962556
2026-04-13T16:48:48.736134Z  INFO parsed_new_accounts=22247784
2026-04-13T16:48:52.021656Z  INFO parsed_new_accounts=22533012
2026-04-13T16:48:57.506038Z  INFO parsed_new_accounts=22818240
2026-04-13T16:48:59.110278Z  INFO parsed_new_accounts=23103468
2026-04-13T16:49:01.542361Z  INFO parsed_new_accounts=23388696
2026-04-13T16:49:02.849429Z  INFO parsed_new_accounts=23673924
2026-04-13T16:49:07.178136Z  INFO Writing accounts to db total_inserted_accounts=285229
2026-04-13T16:49:17.500647Z  INFO Writing accounts to db total_inserted_accounts=570457
2026-04-13T16:49:28.108600Z  INFO Writing accounts to db total_inserted_accounts=855685
2026-04-13T16:50:03.923589Z  INFO Writing accounts to db total_inserted_accounts=1140913
2026-04-13T16:52:35.831660Z  INFO Writing accounts to db total_inserted_accounts=1426141
2026-04-13T16:52:45.883713Z  INFO Writing accounts to db total_inserted_accounts=1711369
2026-04-13T16:52:59.472283Z  INFO Writing accounts to db total_inserted_accounts=1996597
2026-04-13T16:53:08.372623Z  INFO Writing accounts to db total_inserted_accounts=2281825
2026-04-13T16:53:14.781745Z  INFO Writing accounts to db total_inserted_accounts=2567053
2026-04-13T16:53:26.389300Z  INFO Writing accounts to db total_inserted_accounts=2852281
2026-04-13T16:53:45.282166Z  INFO Writing accounts to db total_inserted_accounts=3137509
2026-04-13T16:53:55.603621Z  INFO Writing accounts to db total_inserted_accounts=3422737
2026-04-13T16:54:01.525335Z  INFO Writing accounts to db total_inserted_accounts=3707965
2026-04-13T16:54:08.428505Z  INFO Writing accounts to db total_inserted_accounts=3993193
2026-04-13T16:54:17.078504Z  INFO Writing accounts to db total_inserted_accounts=4278421
2026-04-13T16:54:27.033591Z  INFO Writing accounts to db total_inserted_accounts=4563649
2026-04-13T16:54:36.092951Z  INFO Writing accounts to db total_inserted_accounts=4848877
2026-04-13T16:54:45.267075Z  INFO Writing accounts to db total_inserted_accounts=5134105
2026-04-13T16:55:09.267964Z  INFO Writing accounts to db total_inserted_accounts=5419333
2026-04-13T16:55:59.100860Z  INFO Writing accounts to db total_inserted_accounts=5704561
2026-04-13T16:56:10.258350Z  INFO Writing accounts to db total_inserted_accounts=5989789
2026-04-13T16:56:19.970851Z  INFO Writing accounts to db total_inserted_accounts=6275017
2026-04-13T16:56:45.127827Z  INFO Writing accounts to db total_inserted_accounts=6560245
2026-04-13T16:58:27.649731Z  INFO Writing accounts to db total_inserted_accounts=6845473
2026-04-13T16:58:48.402822Z  INFO Writing accounts to db total_inserted_accounts=7130701
2026-04-13T16:59:02.370346Z  INFO Writing accounts to db total_inserted_accounts=7415929
2026-04-13T16:59:18.178311Z  INFO Writing accounts to db total_inserted_accounts=7701157
2026-04-13T16:59:41.031120Z  INFO Writing accounts to db total_inserted_accounts=7986385
2026-04-13T16:59:45.080597Z  INFO Writing accounts to db total_inserted_accounts=8271613
2026-04-13T16:59:51.030033Z  INFO Writing accounts to db total_inserted_accounts=8556841
2026-04-13T17:00:00.362954Z  INFO Writing accounts to db total_inserted_accounts=8842069
2026-04-13T17:00:48.945788Z  INFO Writing accounts to db total_inserted_accounts=9127297
2026-04-13T17:01:35.158928Z  INFO Writing accounts to db total_inserted_accounts=9412525
2026-04-13T17:01:50.173355Z  INFO Writing accounts to db total_inserted_accounts=9697753
2026-04-13T17:02:02.994997Z  INFO Writing accounts to db total_inserted_accounts=9982981
2026-04-13T17:02:08.880774Z  INFO Writing accounts to db total_inserted_accounts=10268209
2026-04-13T17:02:14.654114Z  INFO Writing accounts to db total_inserted_accounts=10553437
2026-04-13T17:02:19.467126Z  INFO Writing accounts to db total_inserted_accounts=10838665
2026-04-13T17:02:29.333504Z  INFO Writing accounts to db total_inserted_accounts=11123893
2026-04-13T17:02:59.614413Z  INFO Writing accounts to db total_inserted_accounts=11409121
2026-04-13T17:03:07.361987Z  INFO Writing accounts to db total_inserted_accounts=11694349
2026-04-13T17:03:23.012917Z  INFO Writing accounts to db total_inserted_accounts=11979577
2026-04-13T17:03:33.465679Z  INFO Writing accounts to db total_inserted_accounts=12264805
2026-04-13T17:03:41.258517Z  INFO Writing accounts to db total_inserted_accounts=12550033
2026-04-13T17:03:50.245527Z  INFO Writing accounts to db total_inserted_accounts=12835261
2026-04-13T17:04:02.235349Z  INFO Writing accounts to db total_inserted_accounts=13120489
2026-04-13T17:04:07.842991Z  INFO Writing accounts to db total_inserted_accounts=13405717
2026-04-13T17:04:17.457631Z  INFO Writing accounts to db total_inserted_accounts=13690945
2026-04-13T17:04:31.578721Z  INFO Writing accounts to db total_inserted_accounts=13976173
2026-04-13T17:04:38.845099Z  INFO Writing accounts to db total_inserted_accounts=14261401
2026-04-13T17:04:46.594605Z  INFO Writing accounts to db total_inserted_accounts=14546629
2026-04-13T17:04:53.272062Z  INFO Writing accounts to db total_inserted_accounts=14831857
2026-04-13T17:05:06.760919Z  INFO Writing accounts to db total_inserted_accounts=15117085
2026-04-13T17:05:27.484118Z  INFO Writing accounts to db total_inserted_accounts=15402313
2026-04-13T17:05:51.557930Z  INFO Writing accounts to db total_inserted_accounts=15687541
2026-04-13T17:06:04.043248Z  INFO Writing accounts to db total_inserted_accounts=15972769
2026-04-13T17:06:15.434413Z  INFO Writing accounts to db total_inserted_accounts=16257997
2026-04-13T17:06:21.631663Z  INFO Writing accounts to db total_inserted_accounts=16543225
2026-04-13T17:06:34.033365Z  INFO Writing accounts to db total_inserted_accounts=16828453
2026-04-13T17:06:41.789633Z  INFO Writing accounts to db total_inserted_accounts=17113681
2026-04-13T17:06:50.300120Z  INFO Writing accounts to db total_inserted_accounts=17398909
2026-04-13T17:07:01.432326Z  INFO Writing accounts to db total_inserted_accounts=17684137
2026-04-13T17:07:16.606154Z  INFO Writing accounts to db total_inserted_accounts=17969365
2026-04-13T17:07:28.581218Z  INFO Writing accounts to db total_inserted_accounts=18254593
2026-04-13T17:07:54.371080Z  INFO Writing accounts to db total_inserted_accounts=18539821
2026-04-13T17:08:10.016478Z  INFO Writing accounts to db total_inserted_accounts=18825049
2026-04-13T17:08:18.033917Z  INFO Writing accounts to db total_inserted_accounts=19110277
2026-04-13T17:08:25.679326Z  INFO Writing accounts to db total_inserted_accounts=19395505
2026-04-13T17:08:35.017122Z  INFO Writing accounts to db total_inserted_accounts=19680733
2026-04-13T17:08:49.689781Z  INFO Writing accounts to db total_inserted_accounts=19965961
2026-04-13T17:09:00.907484Z  INFO Writing accounts to db total_inserted_accounts=20251189
2026-04-13T17:09:15.153160Z  INFO Writing accounts to db total_inserted_accounts=20536417
2026-04-13T17:09:43.232637Z  INFO Writing accounts to db total_inserted_accounts=20821645
2026-04-13T17:10:03.417596Z  INFO Writing accounts to db total_inserted_accounts=21106873
2026-04-13T17:10:17.799564Z  INFO Writing accounts to db total_inserted_accounts=21392101
2026-04-13T17:10:31.665755Z  INFO Writing accounts to db total_inserted_accounts=21677329
2026-04-13T17:10:40.229743Z  INFO Writing accounts to db total_inserted_accounts=21962557
2026-04-13T17:10:52.117399Z  INFO Writing accounts to db total_inserted_accounts=22247785
2026-04-13T17:11:18.432483Z  INFO Writing accounts to db total_inserted_accounts=22533013
2026-04-13T17:11:37.933215Z  INFO Writing accounts to db total_inserted_accounts=22818241
2026-04-13T17:11:50.539144Z  INFO Writing accounts to db total_inserted_accounts=23103469
2026-04-13T17:12:18.830799Z  INFO Writing accounts to db total_inserted_accounts=23388697
2026-04-13T17:12:35.140155Z  INFO Writing accounts to db total_inserted_accounts=23673925
2026-04-13T17:12:47.473810Z  INFO Writing accounts to db total_inserted_accounts=23855330
2026-04-13T17:12:50.644928Z  INFO All accounts written to database, starting state root computation (may take some time)
2026-04-13T17:24:05.391624Z ERROR Computed state root does not match state root in state dump computed_state_root=0xd84e693fd4e8ce5ad0b4c7e45b088d4119b7d0e9b8d76ee74ee83a5c7adea6a1 expected_state_root=0xed980641a4bd4d2e84c6c8db980b7f05e95733c92be2e0045db3735efeb1d807
Error: state root mismatch: got 0xd84e693fd4e8ce5ad0b4c7e45b088d4119b7d0e9b8d76ee74ee83a5c7adea6a1, expected 0xed980641a4bd4d2e84c6c8db980b7f05e95733c92be2e0045db3735efeb1d807

Location:
    /mnt/ssd/rust/cargo/git/checkouts/reth-e231042ee7db3fb7/564ffa5/crates/storage/db-common/src/init.rs:684:10
```
