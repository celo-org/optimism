# Celo L2 Migration Script

## Overview

This script migrates a Celo L1 database into one compatible with Celo L2. It consists of 3 main processes that respectively migrate ancient blocks, non-ancient blocks and state. Migrated data is copied into a new datadir, leaving the old datadir unchanged. To minimize migration downtime, the script is designed to run in two stages. The `pre migration` stage can be run ahead of the `full migration` and will process as much of the migration as possible up to that point. The `full migration` can then be run to finish migrating new blocks that were created after the `pre migration` and apply necessary state changes on top of the migration block.

### Pre migration

The `pre migration` itself consists of two parts that are run in parallel: It copies and transforms the ancient / frozen blocks (i.e. all blocks before the last 90000) and also copies over the rest of the database using `rsync`. The ancients db is migrated sequentially because it is append-only, while the rest of the database is copied and then transformed in-place. We use `rsync` because it has flags for ignoring the ancients directory, skipping any already copied files and deleting any extra files in the new db, ensuring that we can run the script multiple times and only copy over actual updates.

The `pre migration` step is still run during a `full migration` but it will be much quicker as only newly frozen blocks and recent file changes need to be migrated.

### Full migration

During the `full migration` we re-run the `pre migration` step to capture any updates since the last `pre migration` and then apply in-place changes to non-ancient blocks and state. While this is happening, the script also checks for any ancient blocks that have remained in leveldb despite being frozen and removes them from the new db. Non-ancient blocks are then transformed to ensure compatibility with the L2 codebase.

Finally after all blocks have been migrated, the script performs a series of modifications to the state db. First, it deploys the L2 smart contracts by iterating through the genesis allocs passed to the script and setting the nonce, balance, code and storage for each address accordingly, overwritting existing data if necessary. Finally, these changes are committed to the state db to produce a new state root and create the first Celo L2 block.

### Notes

The longest running section of the script is the ancients migration followed by the `rsync` command. By running these together in a `pre migration` we greatly reduce how long they will take during the `full migration`. Changes made to non-ancient blocks and state during a `full migration` are erased by the next `rsync` command.

The script outputs a `rollup-config.json` file that is passed to the sequencer in order to start the L2 network.

See `--help` for how to run each portion of the script individually, along with other configuration options.

### Running the script

First, build the script by running

```bash
make celo-migrate
```

from the `op-chain-ops` directory.

You can then run the script as follows.

```bash
go run ./cmd/celo-migrate --help
```

NOTE: You will need `rsync` to run this script if it's not already installed

#### Running with local test setup (Alfajores / Holesky)

To test the script locally, we can migrate an alfajores database and use Holesky as our L1. The input files needed for this can be found in `./testdata`. The necessary smart contracts have already been deployed on Holesky.

##### Pull down the latest alfajores database snapshot

```bash
gcloud alpha storage cp gs://celo-chain-backup/alfajores/chaindata-latest.tar.zst alfajores.tar.zst
```

Unzip and rename

```bash
tar --use-compress-program=unzstd -xvf alfajores.tar.zst
mv chaindata ./data/alfajores_old
```

##### Generate test allocs file

The state migration takes in a allocs file that specifies the l2 state changes to be made during the migration. This file can be generated from the deploy config and l1 contract addresses by running the following from the `contracts-bedrock` directory.

```bash
CONTRACT_ADDRESSES_PATH=../../op-chain-ops/cmd/celo-migrate/testdata/deployment-l1-holesky.json \
DEPLOY_CONFIG_PATH=../../op-chain-ops/cmd/celo-migrate/testdata/deploy-config-holesky-alfajores.json \
STATE_DUMP_PATH=../../op-chain-ops/cmd/celo-migrate/testdata/l2-allocs-alfajores.json \
forge script ./scripts/L2Genesis.s.sol:L2Genesis \
--sig 'runWithStateDump()'
```

This should output the allocs file to `./testdata/l2-allocs-alfajores.json`. If you encounter difficulties with this and want to just continue testing the script, you can alternatively find the allocs file [here](https://gist.github.com/jcortejoso/7f90ba9b67c669791014661ccb6de81a).

##### Run script with test configuration

```bash
go run ./cmd/celo-migrate pre \
--old-db ./data/alfajores_old \
--new-db ./data/alfajores_new
```

Running the pre-migration script should take ~5 minutes. This script copies and transforms ancient blocks and, in parallel, copies over all other chaindata without transforming it. This can be re-run mutliple times leading up to the full migration, and should only migrate updates to the old db between re-runs.

```bash
go run ./cmd/celo-migrate full \
--deploy-config ./cmd/celo-migrate/testdata/deploy-config-holesky-alfajores.json \
--l1-deployments ./cmd/celo-migrate/testdata/deployment-l1-holesky.json \
--l1-rpc https://ethereum-holesky-rpc.publicnode.com  \
--l2-allocs ./cmd/celo-migrate/testdata/l2-allocs-alfajores.json \
--outfile.rollup-config ./cmd/celo-migrate/testdata/rollup-config.json \
--old-db ./data/alfajores_old \
--new-db ./data/alfajores_new
```

Running the full migration script re-runs the pre-migration script once to migrate any new changes to the old db that have occurred since the last pre-migration. It then performs in-place transformations on the non-ancient blocks and performs the state migration as well.

#### Running for Cel2 migration

##### Generate allocs file

You can generate the allocs file needed to run the migration with the following script in `contracts-bedrock`

```bash
CONTRACT_ADDRESSES_PATH=<PATH_TO_CONTRACT_ADDRESSES> \
DEPLOY_CONFIG_PATH=<PATH_TO_MY_DEPLOY_CONFIG> \
STATE_DUMP_PATH=<PATH_TO_WRITE_L2_ALLOCS> \
forge script scripts/L2Genesis.s.sol:L2Genesis \
--sig 'runWithStateDump()'
```

##### Dress rehearsal / pre-migration

To minimize downtime caused by the migration, node operators can prepare their Cel2 databases by running the pre-migration command a day ahead of the actual migration. This will pre-populate the new database with most of the ancient blocks needed for the final migration and copy over other chaindata without transforming it.

If node operators would like to practice a `full migration` they can do so and reset their databases to the correct state by running another `pre migration` afterward.

IMPORTANT: The pre-migration should be run using a chaindata snapshot, rather than a db that is being used by a node. To avoid network downtime, we recommend that node operators do not stop any nodes in order to perform the pre-migration.

Node operators should inspect their migration logs after the dress rehearsal to ensure the migration completed succesfully and direct any questions to the Celo developer community on Discord before the actual migration.

##### Final migration

On the day of the actual cel2 migration, the `full migration` script can be run using the datadir of a Celo L1 node that has halted on the migration block. Far in advance of the migration, a version of `celo-blockchain` will be distributed where a flag can specify a block to halt on. When the Celo community aligns on a migration block, node operators will start / restart their nodes with this flag specifying the migration block. Their nodes will halt when this block is reached, at which point they will be able to run `full migration` and begin syncing with the Celo L2 network.
