# Celo devnet + mainnet deployment scripts

## Generate Devnet L1 and L2 Allocations

1. Run
    ```bash
    make devnet-clean && DEVNET_CELO=true SAFE_AS_OWNER=true make devnet-up
    ```
    This will generate and run full devnet in local docker
2. Copy `.devnet/genesis-L1.json`, `.devnet/genesis-L2.json`, and `.devnet/rollup.json` to the `.devnet` folder in other branch
3. Run on other branch
    ```bash
    DEVNET_CELO=true SAFE_AS_OWNER=true make devnet-up
    ```
4. To double check that everything is working correctly, search for

    ``` L1 genesis already generated.```

    ``` L2 genesis and rollup configs already generated.```

    in logs


## Generate L2 Allocations for real network (Alfajores, Baklava, Mainnet)

### Environment Variables

- `CELO_MONOREPO_DIR`
  Root directory of the monorepo. Defaults to current directory.

- **Directory Paths**
  - `CELO_OP_NODE_DIR`: `$(CELO_MONOREPO_DIR)/op-node`
  - `CELO_DEVNET_CONFIG_PATH`: `packages/contracts-bedrock/deploy-config/devnetL1.json`
  - `CELO_L2_ALLOCS_PATH`: `packages/contracts-bedrock/.devnet/allocs-l2.json`
  - `CELO_ADDRESSES_JSON_PATH`: `packages/contracts-bedrock/.devnet/addresses.json`
  - `CELO_GENESIS_L2_PATH`: `packages/contracts-bedrock/.devnet/genesis-l2.json`
  - `CELO_ROLLUP_CONFIG_PATH`: `packages/contracts-bedrock/.devnet/rollup.json`
  - `CELO_CONTRACTS_BEDROCK_DIR`: `packages/contracts-bedrock`

- **RPC**
  - `CELO_L1_RPC_URL`: `http://localhost:8545`

- **Deployment**
  - `CELO_DEPLOYMENT_CONTEXT` (default: `devnetL1`)
  - `CELO_L1_CHAIN_ID` (default: `1`)

- **Derived Paths**
  - `CELO_DEPLOY_CONFIG_PATH`: `deploy-config/$(CELO_DEPLOYMENT_CONTEXT).json`
  - `CELO_STATE_DUMP_PATH`: `deployments/l2-allocs.json`
  - `CELO_CONTRACT_ADDRESSES_PATH`: `deployments/$(CELO_L1_CHAIN_ID)-deploy.json`

## Makefile Targets

- **generate_l2_allocs**
  - Runs Forge script `L2Genesis` with necessary paths
  - Generates and saves l2 allocations to `$(CELO_STATE_DUMP_PATH)`

## Usage

1. **Create L1 deploy script**
  It is necessary to add L1 deployment file into `deployments` directory with the name convetion `$(CELO_L1_CHAIN_ID)-deploy.json` eg `1-deploy.json` (example of such file https://storage.googleapis.com/cel2-rollup-files/baklava/deployment-l1.json)
1. **Set Environment Variables**

   ```bash
   export CELO_MONOREPO_DIR=/path/to/monorepo
   export CELO_DEPLOYMENT_CONTEXT=devnetL1
   export CELO_L1_CHAIN_ID=1
   ```
1. **Generate L2 Allocations**

   ```bash
   make generate_l2_allocs
    ```
