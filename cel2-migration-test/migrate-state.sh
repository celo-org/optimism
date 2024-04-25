set -o errexit   # abort on nonzero exitstatus
set -o nounset   # abort on unbound variable
set -o pipefail  # don't hide errors within pipes

SOURCE_DIR="data/l2/source"
TARGET_DIR="data/l2/migrated"

if [ -d "$TARGET_DIR" ]; then rm -Rf $TARGET_DIR; fi

cp -r ${SOURCE_DIR} ${TARGET_DIR}

MIGRATION_BIN="../op-chain-ops/bin/celo-migrate"

${MIGRATION_BIN} \
  --deploy-config config/config.json \
  --l1-deployments config/deployment-l1.json \
  --l1-rpc http://localhost:8545 \
  --db-path ${TARGET_DIR} \
  --outfile.l2 config/op-state-log.json \
  --outfile.rollup config/rollup-config.json \

# Move datadir to the place where op-geth expects it
mv ${TARGET_DIR}/celo ${TARGET_DIR}/geth
