#!/bin/sh
set -euo pipefail

# Sync local copies of target Optimism and op-geth branches

OPTIMISM_CONTRACT_DIRECTORY="optimism-contracts"
OP_GETH_DIRECTORY="op-geth"

DEFAULT_OPTIMISM_CONTRACT_BRANCH="Kourin1996/celo-rebase-13-contracts-devnet"
DEFAULT_OP_GETH_BRANCH="celo-rebase-13"
OPTIMISN_CONTRACT_BRANCH="${1:-$DEFAULT_OPTIMISM_CONTRACT_BRANCH}"
OP_GETH_BRANCH="${2:-$DEFAULT_OP_GETH_BRANCH}"

# optimism
echo "Preparing optimism repository in $OPTIMISM_CONTRACT_DIRECTORY"
if [ ! -d $OPTIMISM_CONTRACT_DIRECTORY ]; then
    echo "Cloning optimism into $OPTIMISM_CONTRACT_DIRECTORY"
    git clone https://github.com/celo-org/optimism.git $OPTIMISM_CONTRACT_DIRECTORY
fi

(
    echo "Check out / pull $OPTIMISN_CONTRACT_BRANCH"
    cd $OPTIMISM_CONTRACT_DIRECTORY
    git fetch origin $OPTIMISN_CONTRACT_BRANCH
    git checkout $OPTIMISN_CONTRACT_BRANCH
    git pull --ff-only origin $OPTIMISN_CONTRACT_BRANCH
)

# op-geth
echo "Preparing op-geth reository in $OP_GETH_DIRECTORY"
if [ ! -d $OP_GETH_DIRECTORY ]; then
    echo "Cloning op-geth into $OP_GETH_DIRECTORY"
    git clone https://github.com/celo-org/op-geth.git $OP_GETH_DIRECTORY
fi

(
    echo "Check out / pull $OP_GETH_BRANCH"
    cd $OP_GETH_DIRECTORY
    git fetch origin $OP_GETH_BRANCH
    git checkout $OP_GETH_BRANCH
    git pull --ff-only origin $OP_GETH_BRANCH
)
