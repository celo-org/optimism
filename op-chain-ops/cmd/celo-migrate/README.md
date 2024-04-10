# Celo L1 -> Cel2 migrator

This tool allows migrating the state of a Celo chain to a genesis block for a CeL2 chain.

## Test Setup

Create a local chain (assuming this in run in the `celo-blockchain` repository):

```sh
build/bin/mycelo genesis --buildpath compiled-system-contracts --dev.accounts 2 --newenv tmp/testenv --mnemonic "miss fire behind decide egg buyer honey seven advance uniform profit renew"
build/bin/mycelo validator-init tmp/testenv/
build/bin/mycelo validator-run tmp/testenv/
```

Create some data

```sh
build/bin/mycelo load-bot tmp/testenv
```

To run the migration, run in `op-chain-ops` (set `CELO_DATADIR` if the `celo-blockchain` repo is not located at `~/celo-blockchain`):

```sh
make && ./migrate.sh
```

## Tasks

- Import and change `BuildL2Genesis`
  - Don't set balances for predeploys/precompiles
