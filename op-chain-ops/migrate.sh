set -o errexit   # abort on nonzero exitstatus
set -o nounset   # abort on unbound variable
set -o pipefail  # don't hide errors within pipes
set -x

echo "Starting Migration"

./bin/op-migrate \
  --db-path=${CELO_DATADIR:-~/celo-blockchain/tmp/testenv/validator-00/} \
  --dry-run