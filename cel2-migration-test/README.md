# Migration testing setup

This directory contains files for running a local L2 with the purpose of testing
the migration. Most services are run via docker compose, but there are some
scripts that need to be run manually (right now).

In general services can be started by `docker compose up <service_name>`. If you
want to start in a detached mode, add `-d` to the command. Logs can be watched
with `docker compose logs <service_name>`. Add `-f` to follow them.

All commands are expected to be run from inside this directory.

## Walkthrough

1. **Start the L1**

   The L1 is a `geth` client running with the clique consensus engine in the
   `l1` service. Is uses the genesis file at `genesis-l1.json` to initialize the
   chain and fund important accounts that will be used in later steps.

   Start the L1 by running `docker compose up l1`. This will initialize the
   chain, the data dir is `data/l1`.

1. **Deploy the OP L1 contracts**

   Now it is necessary to deploy the L1 contracts. This is done by the
   `deploy-l1-contracts.sh` script.

   Start the script `./deploy-l1-contracts.sh`. This creates a config file
   (`config/config.json`) and a file containing L1 addresses
   (`deployment-l1.json`).

   Running the script overwrites prior deployments. If it fails without a
   readable error message, it might be caused by re-using an old salt. In this
   case you can create a new one with `direnv allow`.

1. **Setup Celo datadir and run migration**

   First we need to build the migration tool.

   ```sh
   cd ../op-chain-ops  # Assuming you're in cel2-migration-test
   make celo-migrate
   ```

   Now it possible to setup the migration of a Celo datadir.  Copy the datadir
   of the node you want into `data/l2` and name it `source`.  This directory
   will not be touched in the migration process.

   **Important**: Then update the `config/config.json` file under the
   `l2ChainID` field with the chain id of the celo chain.

   Run the migration by executing `./migrate-state.sh`. It should finish with
   the message "Finished migration successfully!" and a new directory `migrated`
   in `data/l2`. Additionally two files will have been created:

   - `config/rollup-config.json` is the config required by `op-node`.
   - `config/op-state-log.json` is purely informational and contains all state
   that was written into the migrated state database.

1. **Run `op-geth` on migrated state**

   Start `op-geth` with `docker compose up l2`.

   Make sure this prints the correct chain id and *Optimism* as the consensus
   engine. Additionally, the merge should be configured and the *Cel2* hardfork
   enabled.

   The execution client is now running and waiting for command from the
   consensus client, `op-node`.

1. **Run `op-node` on migrated state**

   Finally, we can start `op-node`: `docker compose up op-node`.

   This should show logs indicating that blocks are created on both the `l2` and
   `op-node` services.

1. **Run `op-batcher` and `op-proposer`**

   Run `docker compose up op-batcher op-proposer`. If this fails with unset
   environment variables, reload the `.envrc` file with `direnv allow`.
