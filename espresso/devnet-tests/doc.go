// Package devnet_tests holds the Espresso devnet test tier from
// EspressoSystems/optimism-espresso-integration (espresso/devnet-tests).
//
// Unlike espresso/environment — which runs in-process against an in-memory
// mock Espresso client — this tier drives a docker-compose devnet and, in the
// TEE profile, a batcher running inside a real AWS Nitro enclave with the SP1
// zk attestation verifier. Every test in this package is currently a stub
// that skips immediately; the real bodies live in the reference repo and
// cannot run here until the supporting infrastructure is ported:
//
//   - the espresso/ docker-compose devnet (compose profiles, including `tee`)
//     and the devnet_tools.go harness (NewDevnet, per-service up/down,
//     liveness/outage tuning via ESPRESSO_DEVNET_TESTS_{LIVENESS,OUTAGE}_PERIOD)
//   - .github/workflows/espresso-devnet-tests.yaml — the standard profile on
//     GitHub-hosted runners
//   - .github/workflows/espresso-devnet-tests-tee.yaml plus the reusable
//     .github/workflows/ec2-devnet-test.yaml — each test group on an
//     ephemeral AWS Nitro-enclave-enabled EC2 runner
//     (ESPRESSO_RUN_ENCLAVE_TESTS=true, COMPOSE_PROFILES=tee). Requires the
//     EspressoSystems/ec2-github-runner enclave fork, repository AWS secrets
//     (IAM role, subnet, security group) and the Nix+Docker+nitro-cli AMI.
//
// Reference:
// https://github.com/EspressoSystems/optimism-espresso-integration/tree/celo-integration-rebase-17/espresso/devnet-tests
package devnet_tests

// skipNotPorted is the reason every stub in this package skips; see the
// package doc for the full porting checklist.
const skipNotPorted = "devnet test tier not ported yet: needs the docker-compose devnet and the EC2/TEE workflows from optimism-espresso-integration (see package doc)"
