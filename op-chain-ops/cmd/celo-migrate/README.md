# Celo L2 Migration Tool

Tool for preparing a pre-L2 Celo database for use in an L2 Celo node.

> ⚠️ The instructions in this README are for illustrative purposes only. For the most complete and up-to-date information on how to participate in the Celo L2 hardfork, please see the [Celo Docs](https://docs.celo.org/cel2/notices/l2-migration).

See also the [celo-l2-node-docker-compose](https://github.com/celo-org/celo-l2-node-docker-compose) repo, which provides tooling to make migrating and running Celo L2 nodes easy. We recommend migrating data using the tooling provided there, as it simplifies the migration interface significantly and provides necessary configuration artifacts.

## Migration Script

This script migrates a pre-L2 database into one compatible with Celo L2. It consists of 3 main processes for migrating ancient blocks, non-ancient blocks and state. Migrated data is written to a new datadir, leaving the old datadir unchanged.

To minimize migration downtime, the script is designed to run in two stages:

1. The `pre-migration` stage can be run ahead of the `full migration` and will process as much of the migration as possible up to that point.
2. The `full migration` can then be run to finish migrating new blocks that were created after the `pre-migration` and apply necessary state changes on top of the migration block.

### Pre-migration

The `pre-migration` consists of two steps that are run in parallel:

1. Copy and transform the ancient / frozen blocks (i.e. all blocks before the last 90000).
2. Copy over the rest of the database using `rsync`.

The ancients db is migrated sequentially because it is append-only, while the rest of the database is copied and then transformed in-place. We use `rsync` because it has flags for ignoring the ancients directory, skipping any already copied files and deleting any extra files in the new db, ensuring that we can run the script multiple times and only copy over actual updates.

The `pre-migration` step is still run during a `full migration` but it will be much quicker as only newly frozen blocks and recent file changes need to be migrated.

### Full migration

During the `full migration`, we repeat the `pre-migration` step to capture any updates since the last `pre-migration` and then apply in-place changes to non-ancient blocks and state. While this is happening, the script also checks for any stray ancient blocks that have remained in leveldb despite being frozen and removes them from the new db. Non-ancient blocks are then transformed to ensure compatibility with the L2 codebase.

Finally after all blocks have been migrated, the script performs a series of modifications to the state db:

1. First, it deploys the L2 smart contracts by iterating through the genesis allocs passed to the script and setting the nonce, balance, code and storage for each address accordingly, overwritting existing data if necessary.
2. Finally, these changes are committed to the state db to produce a new state root and create the first Celo L2 block.

### Notes

> [!TIP]
> See `--help` for how to run each portion of the script individually, along with other configuration options.

- The script outputs a `rollup-config.json` file that is passed to the sequencer in order to start the L2 network.

- The longest running section of the script is the ancients migration, followed by the `rsync` command. By running these together in a `pre-migration` we greatly reduce how long they will take during the `full migration`. Changes made to non-ancient blocks and state during a `full migration` are erased by the next `rsync` command.

> ⚠️ **Do not migrate archive data, only full node data**. Because we use `rsync` with checksums, the command will take a very long time if run on archive data. All the historical state stored by an archive node will be checksummed even if a `pre-migration` has already been performed to copy over the data. This is slow and memory instensive. Moreover, Celo L2 nodes cannot use pre-hardfork state, so all the state data will be copied over and stored for no reason. Therefore, we do not recommend running the migration script on an archive datadir.

### Running the script

> [!NOTE]
> You will need `rsync` to run this script if it's not already installed.

From the `op-chain-ops` directory, first build the script by running:

```bash
make celo-migrate
```

You can then run the script as follows:

```bash
go run ./cmd/celo-migrate --help
```

#### Running with local test setup (Alfajores / Holesky)

To test the script locally, we can migrate an Alfajores database and use Holesky as our L1. The input files needed for this can be found in `./testdata`. The necessary smart contracts have already been deployed on Holesky.

##### Pull down the latest Alfajores database snapshot

```bash
gcloud alpha storage cp gs://celo-chain-backup/alfajores/chaindata-latest.tar.zst alfajores.tar.zst
```

Unzip and rename

```bash
tar --use-compress-program=unzstd -xvf alfajores.tar.zst
mv chaindata ./data/alfajores_old
```

##### Generate test allocs file

The state migration takes in an allocs file that specifies the l2 state changes to be made during the migration. This file can be generated from the deploy config and l1 contract addresses by running the following from the `contracts-bedrock` directory.

```bash
CONTRACT_ADDRESSES_PATH=../../op-chain-ops/cmd/celo-migrate/testdata/deployment-l1-dango.json \
DEPLOY_CONFIG_PATH=../../op-chain-ops/cmd/celo-migrate/testdata/deploy-config-dango.json \
STATE_DUMP_PATH=../../op-chain-ops/cmd/celo-migrate/testdata/l2-allocs-dango.json \
forge script ./scripts/L2Genesis.s.sol:L2Genesis \
--sig 'runWithStateDump()'
```

This should output the allocs file to `./testdata/l2-allocs-dango.json`. If you encounter difficulties with this and want to just continue testing the script, you can alternatively find the allocs file [here](https://storage.googleapis.com/cel2-rollup-files/alfajores-mvp/l2-allocs.json).

##### Run script with test configuration

```bash
go run ./cmd/celo-migrate pre \
--old-db ./data/alfajores_old \
--new-db ./data/alfajores_new
```

Running the pre-migration script should take ~5 minutes. This script copies and transforms ancient blocks and, in parallel, copies over all other chaindata without transforming it. This can be re-run mutliple times leading up to the full migration, and should only migrate updates to the old db between re-runs.

```bash
go run ./cmd/celo-migrate full \
--deploy-config ./cmd/celo-migrate/testdata/deploy-config-dango.json \
--l1-deployments ./cmd/celo-migrate/testdata/deployment-l1-dango.json \
--l1-rpc https://ethereum-holesky-rpc.publicnode.com  \
--l2-allocs ./cmd/celo-migrate/testdata/l2-allocs-dango.json \
--outfile.rollup-config ./cmd/celo-migrate/testdata/rollup-config-dango.json \
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

##### Dry-run / pre-migration

To minimize downtime caused by the migration, node operators can prepare their Cel2 databases by running the pre-migration command a day ahead of the actual migration. This will pre-populate the new database with most of the ancient blocks needed for the final migration and copy over other chaindata without transforming it.

If node operators would like to practice a `full migration` they can do so and reset their databases to the correct state by running another `pre-migration` afterward.

> [!IMPORTANT]
> The pre-migration should be run using a chaindata snapshot, rather than a db that is being used by a node. To avoid network downtime, we recommend that node operators do not stop any nodes in order to perform the pre-migration.

Node operators should inspect their migration logs after the dry-run to ensure the migration completed successfully and direct any questions to the Celo developer community on Discord before the actual migration.

##### Final migration

On the day of the actual Cel2 migration, the `full migration` script can be run using the datadir of a Celo L1 node that has halted on the migration block. Far in advance of the migration, a version of `celo-blockchain` will be distributed where a flag can specify a block to halt on. When the Celo community aligns on a migration block, node operators will start / restart their nodes with this flag specifying the migration block. Their nodes will halt when this block is reached, at which point they will be able to run `full migration` and begin syncing with the Celo L2 network.
