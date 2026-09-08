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
// The SystemConfig binding comes from RollupConfig.L1SystemConfigAddress
// rather than from BatchAuthenticator.systemConfig(), which the batcher would
// otherwise re-read on every publish tick to reach the same address.
//
// The deployment probe is lazy and latching: it runs on each call until it
// first observes code, then never again. Lazy so a batcher started before the
// contract is deployed keeps skipping publishes per tick rather than failing
// to start; latching keeps the probe off the steady-state publish path, which
// leaves the gate at two eth_calls per tick in either mode.
//
// A zero BatchAuthenticator address is a nil reader, not a reader bound to the
// zero address.
type batchAuthenticatorReader struct {
	addr         common.Address
	auth         *batchauthenticator.BatchAuthenticatorCaller
	systemConfig *systemconfig.SystemConfigCaller
	backend      bind.ContractCaller
	timeout      time.Duration

	mu       sync.Mutex
	haveCode bool
}

// newBatchAuthenticatorReader binds a reader to addr, resolving the fallback
// batcher through the SystemConfig at systemConfigAddr.
func newBatchAuthenticatorReader(addr, systemConfigAddr common.Address, backend bind.ContractCaller, timeout time.Duration) (*batchAuthenticatorReader, error) {
	auth, err := batchauthenticator.NewBatchAuthenticatorCaller(addr, backend)
	if err != nil {
		return nil, fmt.Errorf("failed to bind BatchAuthenticator at %s: %w", addr, err)
	}
	systemConfig, err := systemconfig.NewSystemConfigCaller(systemConfigAddr, backend)
	if err != nil {
		return nil, fmt.Errorf("failed to bind SystemConfig at %s: %w", systemConfigAddr, err)
	}
	return &batchAuthenticatorReader{
		addr:         addr,
		auth:         auth,
		systemConfig: systemConfig,
		backend:      backend,
		timeout:      timeout,
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
	opts, cancel := r.callOpts(ctx)
	defer cancel()
	batcherHash, err := r.systemConfig.BatcherHash(opts)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to read batcherHash: %w", err)
	}
	// batcherHash stores the batcher address in its low 20 bytes,
	// which is why we use bytes to address
	return common.BytesToAddress(batcherHash[:]), nil
}

// EspressoTEEVerifier returns the contract's configured EspressoTEEVerifier
// address. Read once at startup.
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
	return addr, nil
}
