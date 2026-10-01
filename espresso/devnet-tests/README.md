# Espresso Devnet Tests

Test various end-to-end functionalities in a locally running devnet.

## Running

`go test ./espresso/devnet-tests/...`

Configure how long it takes to run the tests vs how stringent the tests are by setting
`ESPRESSO_DEVNET_TESTS_LIVENESS_PERIOD` and `ESPRESSO_DEVNET_TESTS_OUTAGE_PERIOD`. These determine
how long we need the devnet to run in a healthy state before considering a run successful, and how
long to let an unhealthy state persist before attempting recovery, respectively. For the fullest
test these are set to `1m` and `10m`, but for quick testing, more reasonable values would be around
`10s`.
