# CGTv2 production bundle settings

Generate the production bundle through `cgtv2_bundle.py`. By default it resolves `latest` once, then reads legacy vault getter values at that fixed block hash and passes them together with explicit token metadata to `GenerateCgtV2NUTBundle.runProductionFromJson()`. It performs no broadcasts or on-chain writes.

Use an L2 RPC that supports historical `eth_call` with an EIP-1898 block selector. Every getter request uses `{ "blockHash": H, "requireCanonical": true }`, even if the head advances during execution. A failed pinned read stops the tool; it never retries at a newer block or substitutes guessed defaults. RPC responses are trusted build inputs, not state proofs.

## Generate

Create a JSON file containing the expected L2 chain ID, the four finalized deployment addresses, and the intended token name and symbol. Do not include vault settings; the tool obtains those from the legacy vault. Token metadata is required explicitly because the legacy Celo L1Block getters return Ether/ETH rather than the intended Celo/CELO metadata.

```json
{
  "chainId": 42220,
  "bridgeProxy": "<reserved L2 bridge proxy>",
  "l1Bridge": "<L1 gas bridge>",
  "l1Token": "<L1 CELO token>",
  "owner": "<controller governance owner>",
  "gasPayingTokenName": "Celo",
  "gasPayingTokenSymbol": "CELO"
}
```

From `packages/contracts-bedrock`, run:

```bash
mise x -- just cgtv2-bundle generate \
  --rpc-url "$L2_RPC_URL" \
  --config .testdata/cgtv2-mainnet-deployment.json \
  --output snapshots/upgrades/cgtv2-mainnet-nut-bundle.json
```

The tool reads `RECIPIENT()`, `MIN_WITHDRAWAL_AMOUNT()` and `WITHDRAWAL_NETWORK()` from the sequencer fee vault at `0x4200000000000000000000000000000000000011`. A missing, reverting or malformed getter fails generation. Metadata comes only from the deployment config, is passed through JSON, and is checked against the generated initializer bytes; no L1Block metadata getter is queried. Missing, empty or non-string metadata fails before any RPC reads. The Solidity generator remains the low-level writer and the synthetic fixture entry point remains unchanged.

The output has a sibling `cgtv2-mainnet-nut-bundle.snapshot.json` containing the source block hash/number, chain ID, captured configuration and the bundle SHA-256. It does not store the RPC URL. Review and freeze both files, register the reviewed production bundle in Go/Rust, and retain the final contract/artifact and gas checks described in the activation design. This tool does not register a production bundle or resolve deployment addresses.

## First initialization

The bundle deploys six implementations and performs nine calls, for 15 static deposits plus the dynamic reserve-funding deposit. It does not deploy StorageSetter or reset initializer guards. The controller and reserved bridge proxy must be unused, and the FeeVault's OZ v5 initializer slot must be zero. An already-initialized target reverts instead of being reset.

Celo mainnet block 81287284 (`0x6e4a1f177fa8518b9b6765b6340af5a36fe46a652ebd5c865d176f949176d990`) has FeeVault version `1.5.0-beta.3` and a zero OZ v5 initializer slot (`0xf0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a00`). Its populated slot zero is `totalProcessed`, which is preserved. Recheck these assumptions for each final network and activation parent.

## Check before migration

Before activation, compare the current legacy settings with the frozen bundle. This resolves `latest` once for this verification run:

```bash
mise x -- just cgtv2-bundle verify \
  --rpc-url "$L2_RPC_URL" \
  --bundle snapshots/upgrades/cgtv2-mainnet-nut-bundle.json
```

To reproduce either command at a specific L2 block, add `--block-hash "$L2_BLOCK_HASH"`. An explicit hash bypasses the `latest` lookup. The resolved hash is printed, and generation records it in the snapshot.

The sibling snapshot path is selected automatically; `--snapshot` can select it explicitly. Verification checks the bundle digest, decodes the frozen initializer calldata to check the snapshot settings, then checks the chain, block hash/number and all three legacy vault values. The configured token name and symbol must match the snapshot and initializer calldata; verification does not replace them with legacy L1Block values. It returns nonzero if anything differs or a read fails. It neither rewrites the bundle nor updates the snapshot. If a setting changed, regenerate and review the bundle and client artifacts before migration. Keep the relevant implementation/configuration unchanged between this check and activation. Run this check before the legacy contracts have been upgraded.

## Tests

```bash
mise x -- python3 -m unittest discover -s scripts/cgt -p test_cgtv2_bundle.py -v
mise x -- just test --offline --match-contract CgtV2Upgrade_Upgrade_Test
```

The Python tests use a local JSON-RPC server and run the real Solidity generator and ABI decoder. They check captured initializer values, explicit metadata preservation despite legacy Ether/ETH getters, missing metadata, pinned requests, configuration changes and read/digest failures. They do not require production RPC access or change the frozen Go/Rust fixtures.
