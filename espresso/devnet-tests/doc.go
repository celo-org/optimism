// Package devnet_tests holds the Espresso devnet test tier from
// EspressoSystems/optimism-espresso-integration (espresso/devnet-tests).
//
// Unlike espresso/environment, which runs in-process against an in-memory mock
// Espresso client, this tier drives the docker-compose devnet in
// espresso/docker-compose.yml: anvil as L1, op-geth and op-node for a
// sequencer and a verifier, the Espresso batcher and the fallback batcher, a
// real espresso-dev-node, and EigenDA in memstore mode. Devnet (devnet_tools.go)
// starts and stops it and wraps common interactions.
//
// Running locally, from the repository root:
//
//	just build-contracts
//	espresso/devnet/prepare.sh
//	docker compose -f espresso/docker-compose.yml build
//	go test -v -run TestSmoke ./espresso/devnet-tests/...
//
// ESPRESSO_DEVNET_TESTS_LIVENESS_PERIOD and ESPRESSO_DEVNET_TESTS_OUTAGE_PERIOD
// trade test duration for strictness: how long the devnet must stay healthy
// before a run counts as a success, and how long an unhealthy state is left
// before recovery (default 10s each). CI runs
// .github/workflows/espresso-devnet-tests.yaml.
//
// TestSmoke is ported. The other tests are stubs that skip immediately; their
// bodies are in the reference repo. TestWithdrawal and TestChallengeGame need
// the op-succinct services, started with WithProofs. The TEE compose profile
// (a batcher inside an AWS Nitro enclave) is not ported.
//
// Reference:
// https://github.com/EspressoSystems/optimism-espresso-integration/tree/celo-integration-rebase-17/espresso/devnet-tests
package devnet_tests

// skipNotPorted is the reason every not-yet-ported test in this package skips.
const skipNotPorted = "devnet test not ported yet: its body is in optimism-espresso-integration espresso/devnet-tests (see package doc)"
