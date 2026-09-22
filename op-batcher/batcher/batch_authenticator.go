package batcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/op-service/bindings/batchauthenticator"
	"github.com/ethereum-optimism/optimism/op-service/bindings/systemconfig"
)

// batchAuthenticatorReader is the batcher's single read-only view of the
// BatchAuthenticator contract, bound once at construction and shared by
// registerBatcher, resolveTEEVerifierAddress and isBatcherActive.
//
// The SystemConfig binding is resolved from BatchAuthenticator.systemConfig(),
// the address the contract itself resolves the fallback batcher through. Taking
// it from anywhere else would let this gate and the on-chain check disagree.
//
// The deployment probe is lazy and latching: it runs on each call until it
// first observes code, then never again. Lazy so the fallback batcher, which
// never registers with the contract, keeps skipping publishes rather than
// failing to start against a BatchAuthenticator deployed after it; latching
// keeps the probe off the steady-state publish path. The SystemConfig address
// latches the same way.
//
// A zero address is rejected wherever the reader keeps what it read, because a
// contract holding code but no initialized state answers every address getter
// with one, and a latched zero never recovers. The batcher identities are
// compared and discarded, so a zero there matches no configured key and simply
// skips the publish.
type batchAuthenticatorReader struct {
	addr    common.Address
	auth    *batchauthenticator.BatchAuthenticatorCaller
	backend bind.ContractCaller
	timeout time.Duration

	mu           sync.Mutex
	haveCode     bool
	systemConfig *systemconfig.SystemConfigCaller
}

// newBatchAuthenticatorReader binds a reader to addr. A zero addr means the
// chain has no BatchAuthenticator and yields no reader, never one bound to the
// zero address; callers read that nil as "no BatchAuthenticator configured".
func newBatchAuthenticatorReader(addr common.Address, backend bind.ContractCaller, timeout time.Duration) (*batchAuthenticatorReader, error) {
	if addr == (common.Address{}) {
		return nil, nil
	}
	auth, err := batchauthenticator.NewBatchAuthenticatorCaller(addr, backend)
	if err != nil {
		return nil, fmt.Errorf("failed to bind BatchAuthenticator at %s: %w", addr, err)
	}
	return &batchAuthenticatorReader{
		addr:    addr,
		auth:    auth,
		backend: backend,
		timeout: timeout,
	}, nil
}

// ensureDeployed verifies that code exists at the bound address, skipping the
// check once it has succeeded. Failures are not latched, so a transient RPC
// error or a not-yet-deployed contract is retried on the next call.
func (r *batchAuthenticatorReader) ensureDeployed(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.haveCode {
		return nil
	}
	cCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	code, err := r.backend.CodeAt(cCtx, r.addr, nil)
	if err != nil {
		return fmt.Errorf("failed to check code at BatchAuthenticator address %s: %w", r.addr, err)
	}
	if len(code) == 0 {
		return fmt.Errorf("no contract code at BatchAuthenticator address %s", r.addr)
	}
	r.haveCode = true
	return nil
}

// readContract performs one read against a deployed BatchAuthenticator, bounded
// by the network timeout. getter names the Solidity getter, for the error.
func readContract[T any](ctx context.Context, r *batchAuthenticatorReader, getter string, call func(*bind.CallOpts) (T, error)) (T, error) {
	var zero T
	if err := r.ensureDeployed(ctx); err != nil {
		return zero, err
	}
	cCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	value, err := call(&bind.CallOpts{Context: cCtx})
	if err != nil {
		return zero, fmt.Errorf("failed to read %s: %w", getter, err)
	}
	return value, nil
}

// ActiveIsEspresso reports the contract's activeIsEspresso flag: true when the
// Espresso (TEE) batcher is active, false when the fallback batcher is.
func (r *batchAuthenticatorReader) ActiveIsEspresso(ctx context.Context) (bool, error) {
	return readContract(ctx, r, "activeIsEspresso", r.auth.ActiveIsEspresso)
}

// EspressoBatcher returns the address authorized to authenticate batches while
// activeIsEspresso is true.
func (r *batchAuthenticatorReader) EspressoBatcher(ctx context.Context) (common.Address, error) {
	return readContract(ctx, r, "espressoBatcher", r.auth.EspressoBatcher)
}

// FallbackBatcher returns the address authorized to authenticate batches while
// activeIsEspresso is false, read from the SystemConfig's batcherHash.
func (r *batchAuthenticatorReader) FallbackBatcher(ctx context.Context) (common.Address, error) {
	if err := r.ensureDeployed(ctx); err != nil {
		return common.Address{}, err
	}
	systemConfig, err := r.systemConfigCaller(ctx)
	if err != nil {
		return common.Address{}, err
	}
	batcherHash, err := readContract(ctx, r, "batcherHash", systemConfig.BatcherHash)
	if err != nil {
		return common.Address{}, err
	}
	// batcherHash stores the batcher address in its low 20 bytes,
	// which is why we use bytes to address
	return common.BytesToAddress(batcherHash[:]), nil
}

// systemConfigCaller binds the SystemConfig named by the BatchAuthenticator,
// reading the address once and keeping the binding.
func (r *batchAuthenticatorReader) systemConfigCaller(ctx context.Context) (*systemconfig.SystemConfigCaller, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.systemConfig != nil {
		return r.systemConfig, nil
	}
	cCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	addr, err := r.auth.SystemConfig(&bind.CallOpts{Context: cCtx})
	if err != nil {
		return nil, fmt.Errorf("failed to read systemConfig: %w", err)
	}
	if addr == (common.Address{}) {
		return nil, fmt.Errorf("BatchAuthenticator at %s has a zero systemConfig address", r.addr)
	}
	systemConfig, err := systemconfig.NewSystemConfigCaller(addr, r.backend)
	if err != nil {
		return nil, fmt.Errorf("failed to bind SystemConfig at %s: %w", addr, err)
	}
	r.systemConfig = systemConfig
	return systemConfig, nil
}

// EspressoTEEVerifier returns the contract's configured EspressoTEEVerifier
// address, the EIP-712 verifying contract every authentication is signed
// against.
func (r *batchAuthenticatorReader) EspressoTEEVerifier(ctx context.Context) (common.Address, error) {
	addr, err := readContract(ctx, r, "espressoTEEVerifier", r.auth.EspressoTEEVerifier)
	if err != nil {
		return common.Address{}, err
	}
	if addr == (common.Address{}) {
		return common.Address{}, fmt.Errorf("BatchAuthenticator at %s has a zero espressoTEEVerifier address", r.addr)
	}
	return addr, nil
}
