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
// latches the same way, leaving the gate at two eth_calls in either mode.
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

// Address returns the BatchAuthenticator address this reader is bound to.
func (r *batchAuthenticatorReader) Address() common.Address {
	return r.addr
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

// callOpts bounds a single contract read by the network timeout. The caller
// must invoke the returned cancel func.
func (r *batchAuthenticatorReader) callOpts(ctx context.Context) (*bind.CallOpts, context.CancelFunc) {
	cCtx, cancel := context.WithTimeout(ctx, r.timeout)
	return &bind.CallOpts{Context: cCtx}, cancel
}

// ActiveIsEspresso reports the contract's activeIsEspresso flag: true when the
// Espresso (TEE) batcher is active, false when the fallback batcher is.
func (r *batchAuthenticatorReader) ActiveIsEspresso(ctx context.Context) (bool, error) {
	if err := r.ensureDeployed(ctx); err != nil {
		return false, err
	}
	opts, cancel := r.callOpts(ctx)
	defer cancel()
	active, err := r.auth.ActiveIsEspresso(opts)
	if err != nil {
		return false, fmt.Errorf("failed to check activeIsEspresso: %w", err)
	}
	return active, nil
}

// EspressoBatcher returns the address authorized to authenticate batches while
// activeIsEspresso is true.
func (r *batchAuthenticatorReader) EspressoBatcher(ctx context.Context) (common.Address, error) {
	if err := r.ensureDeployed(ctx); err != nil {
		return common.Address{}, err
	}
	opts, cancel := r.callOpts(ctx)
	defer cancel()
	addr, err := r.auth.EspressoBatcher(opts)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to read espressoBatcher: %w", err)
	}
	return addr, nil
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
	opts, cancel := r.callOpts(ctx)
	defer cancel()
	batcherHash, err := systemConfig.BatcherHash(opts)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to read batcherHash: %w", err)
	}
	// batcherHash stores the batcher address in its low 20 bytes,
	// which is why we use bytes to address
	return common.BytesToAddress(batcherHash[:]), nil
}

// systemConfigCaller binds the SystemConfig named by the BatchAuthenticator,
// reading the address once and keeping the binding. Only a usable address is
// kept, so a contract read before its initialization is retried rather than
// pinning the reader to the zero address.
func (r *batchAuthenticatorReader) systemConfigCaller(ctx context.Context) (*systemconfig.SystemConfigCaller, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.systemConfig != nil {
		return r.systemConfig, nil
	}
	opts, cancel := r.callOpts(ctx)
	defer cancel()
	addr, err := r.auth.SystemConfig(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to read systemConfig address: %w", err)
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
// address. A zero address is an error: it is the EIP-712 verifying contract
// every authentication is signed against, and initialize rejects it, so the
// contract reporting one means it has code but no state yet.
func (r *batchAuthenticatorReader) EspressoTEEVerifier(ctx context.Context) (common.Address, error) {
	if err := r.ensureDeployed(ctx); err != nil {
		return common.Address{}, err
	}
	opts, cancel := r.callOpts(ctx)
	defer cancel()
	addr, err := r.auth.EspressoTEEVerifier(opts)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to query EspressoTEEVerifier address: %w", err)
	}
	if addr == (common.Address{}) {
		return common.Address{}, fmt.Errorf("BatchAuthenticator at %s has a zero espressoTEEVerifier address", r.addr)
	}
	return addr, nil
}
