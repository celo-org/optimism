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

// batchAuthenticatorReader is the batcher's read-only view of the
// BatchAuthenticator contract. It is bound once at construction and shared by
// registerBatcher, resolveTEEVerifierAddress and isBatcherActive.
//
// The SystemConfig address is read from BatchAuthenticator.systemConfig(),
// because that is where the contract itself looks up the fallback batcher.
// Reading it from anywhere else could make this gate and the on-chain check
// disagree.
//
// Reads do not check for code first: when an eth_call returns no data, the
// bindings look for code at the address and fail with bind.ErrNoCode. So the
// fallback batcher, which never registers with the contract, can start before
// the BatchAuthenticator is deployed and skip publishes until it appears. The
// SystemConfig address is read once and cached.
//
// A zero SystemConfig or TEE verifier address is rejected, although no
// supported deployment yields one: the proxy reverts until its implementation
// is set, the deploy scripts set it and call initialize in one upgradeToAndCall,
// and initialize rejects both zeros. The checks are kept as a cheap guard
// because both addresses are read once and kept, so a bad value would never be
// re-read. The batcher addresses need no such guard: they are re-read and
// compared with the batcher's own address on every check.
type batchAuthenticatorReader struct {
	addr    common.Address
	auth    *batchauthenticator.BatchAuthenticatorCaller
	backend bind.ContractCaller
	timeout time.Duration

	// mu guards systemConfig, so the reader is safe to share.
	mu           sync.Mutex
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

// ensureDeployed verifies that code exists at the bound address. Reads do not
// need it, see batchAuthenticatorReader. registerBatcher does: it sends a
// transaction, and one sent to an address with no code succeeds as a no-op.
func (r *batchAuthenticatorReader) ensureDeployed(ctx context.Context) error {
	cCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	code, err := r.backend.CodeAt(cCtx, r.addr, nil)
	if err != nil {
		return fmt.Errorf("failed to check code at BatchAuthenticator address %s: %w", r.addr, err)
	}
	if len(code) == 0 {
		return fmt.Errorf("no contract code at BatchAuthenticator address %s", r.addr)
	}
	return nil
}

// readContract performs one contract read, bounded by the network timeout.
// getter names the Solidity getter, for the error.
func readContract[T any](ctx context.Context, r *batchAuthenticatorReader, getter string, call func(*bind.CallOpts) (T, error)) (T, error) {
	var zero T
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
	addr, err := readContract(ctx, r, "systemConfig", r.auth.SystemConfig)
	if err != nil {
		return nil, err
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
