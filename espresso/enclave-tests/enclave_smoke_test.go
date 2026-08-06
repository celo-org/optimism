// Package enclave_tests holds the Espresso enclave test tier from
// EspressoSystems/optimism-espresso-integration (espresso/enclave-tests):
// e2e tests that run the op-batcher inside a real AWS Nitro enclave.
//
// The tests are currently skipped stubs. The real bodies require a
// Nitro-enabled EC2 machine, the op-batcher enclave docker image
// (`just op-batcher-enclave-image` in the reference repo's kurtosis-devnet)
// and ESPRESSO_RUN_ENCLAVE_TESTS=true; in CI they run through the disabled
// .github/workflows/espresso-devnet-tests-tee.yaml / ec2-devnet-test.yaml
// workflows (see their TODO headers for the infrastructure checklist).
//
// Reference:
// https://github.com/EspressoSystems/optimism-espresso-integration/tree/celo-integration-rebase-17/espresso/enclave-tests
package enclave_tests

import "testing"

// TestE2eDevnetWithEspressoAndEnclaveSimpleTransactions is the enclave smoke
// test: it runs the standard simple-transactions e2e flow with the batcher
// executing inside a real AWS Nitro enclave, attesting its key via the real
// attestation path instead of the mock TEE verifier.
//
// TODO: port TestE2eDevnetWithEspressoAndEnclaveSimpleTransactions from
// optimism-espresso-integration espresso/enclave-tests/enclave_smoke_test.go
// once the Nitro-enclave batcher image and EC2 runner infrastructure are
// available in this repo.
func TestE2eDevnetWithEspressoAndEnclaveSimpleTransactions(t *testing.T) {
	t.Skip("enclave test tier not ported yet: needs a Nitro-enabled EC2 runner, the op-batcher enclave image and the real attestation flow (see package doc)")
}
