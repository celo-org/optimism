# ================================
# Variables Configuration
# ================================

# Default to current directory if CELO_MONOREPO_DIR is not set
CELO_MONOREPO_DIR ?= $(shell pwd)

# Directory Paths
CELO_OP_NODE_DIR             := $(CELO_MONOREPO_DIR)/op-node
CELO_DEVNET_CONFIG_PATH      := $(CELO_MONOREPO_DIR)/packages/contracts-bedrock/deploy-config/devnetL1.json
CELO_L2_ALLOCS_PATH          := $(CELO_MONOREPO_DIR)/packages/contracts-bedrock/.devnet/allocs-l2.json
CELO_ADDRESSES_JSON_PATH     := $(CELO_MONOREPO_DIR)/packages/contracts-bedrock/.devnet/addresses.json
CELO_GENESIS_L2_PATH         := $(CELO_MONOREPO_DIR)/packages/contracts-bedrock/.devnet/genesis-l2.json
CELO_ROLLUP_CONFIG_PATH      := $(CELO_MONOREPO_DIR)/packages/contracts-bedrock/.devnet/rollup.json
CELO_CONTRACTS_BEDROCK_DIR   := $(CELO_MONOREPO_DIR)/packages/contracts-bedrock

# RPC Configuration
CELO_L1_RPC_URL              := http://localhost:8545

# Deployment Configuration
CELO_DEPLOYMENT_CONTEXT ?= devnetL1
CELO_L1_CHAIN_ID ?= 1

# Derived Paths based on Deployment Context and Chain ID
CELO_DEPLOY_CONFIG_PATH      := $(CELO_CONTRACTS_BEDROCK_DIR)/deploy-config/$(CELO_DEPLOYMENT_CONTEXT).json
CELO_STATE_DUMP_PATH         := $(CELO_CONTRACTS_BEDROCK_DIR)/deployments/l2-allocs.json
CELO_CONTRACT_ADDRESSES_PATH := $(CELO_CONTRACTS_BEDROCK_DIR)/deployments/$(CELO_L1_CHAIN_ID)-deploy.json

# ================================
# Phony Targets
# ================================
.PHONY: run_Genesis generate_l2_allocs deploy clean help

# ================================
# generate_l2_allocs Target
# ================================
generate_l2_allocs:
	@# Validate Environment Variables
	@if [ -z "$(CELO_DEPLOYMENT_CONTEXT)" ]; then \
		echo "Error: CELO_DEPLOYMENT_CONTEXT is not set."; \
		echo "Set it using 'make deploy CELO_DEPLOYMENT_CONTEXT=<context>' or export it as an environment variable."; \
		exit 1; \
	fi
	@if [ -z "$(CELO_L1_CHAIN_ID)" ]; then \
		echo "Error: CELO_L1_CHAIN_ID is not set."; \
		echo "Set it using 'make deploy CELO_L1_CHAIN_ID=<chain_id>' or export it as an environment variable."; \
		exit 1; \
	fi

	@echo "Running Forge Deployment Script..."

	# Ensure deployment directory exists
	@mkdir -p $(CELO_CONTRACTS_BEDROCK_DIR)/deployments

	# Run Forge script with absolute paths
	@cd $(CELO_CONTRACTS_BEDROCK_DIR) && \
	DEPLOY_CONFIG_PATH=$(CELO_DEPLOY_CONFIG_PATH) \
	STATE_DUMP_PATH=$(CELO_STATE_DUMP_PATH) \
	CONTRACT_ADDRESSES_PATH=$(CELO_CONTRACT_ADDRESSES_PATH) \
	forge script scripts/L2Genesis.s.sol:L2Genesis --sig 'runWithStateDump()'

	@echo "Saved L2 allocations to $(CELO_STATE_DUMP_PATH)"
