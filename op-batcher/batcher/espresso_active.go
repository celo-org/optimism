package batcher

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/espresso/bindings"
)

// isBatcherActive checks if the current batcher is the active one by querying
// the BatchAuthenticator contract. Returns true if this batcher instance should
// be publishing batches, false if it should stay idle.
//
// Only meaningful once event-based batch auth is enforced: the caller
// (shouldSkipPublishForActiveSeq) must gate on derive.IsEspressoAuthEnforced
// first, since pre-enforcement the contract may be undeployed and its
// activeIsEspresso default of true does not confer ownership.
//
// It applies two gates:
//  1. Mode: the contract's activeIsEspresso flag must match this node's role
//     (Config.Espresso.Enabled). activeIsEspresso==true means the Espresso batcher
//     is active; false means the fallback batcher is active.
//  2. Identity: once the mode matches, the configured sender key (Txmgr.From) must
//     be the authorized batcher for that mode, otherwise every authenticateBatchInfo
//     call reverts (Unauthorized{Espresso,Fallback}Batcher) and the batcher loops.
func (l *BatchSubmitter) isBatcherActive(ctx context.Context) (bool, error) {
	// Check if contract code exists at the address
	codeCtx, codeCancel := context.WithTimeout(ctx, l.Config.NetworkTimeout)
	defer codeCancel()
	code, err := l.L1Client.CodeAt(codeCtx, l.RollupConfig.BatchAuthenticatorAddress, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check code at BatchAuthenticator address: %w", err)
	}
	if len(code) == 0 {
		return false, fmt.Errorf("no contract code at BatchAuthenticator address %s", l.RollupConfig.BatchAuthenticatorAddress.Hex())
	}

	batchAuthenticator, err := bindings.NewBatchAuthenticator(l.RollupConfig.BatchAuthenticatorAddress, l.L1Client)
	if err != nil {
		return false, fmt.Errorf("failed to create BatchAuthenticator binding: %w", err)
	}

	cCtx, cancel := context.WithTimeout(ctx, l.Config.NetworkTimeout)
	defer cancel()

	callOpts := &bind.CallOpts{Context: cCtx}

	activeIsEspresso, err := batchAuthenticator.ActiveIsEspresso(callOpts)
	if err != nil {
		return false, fmt.Errorf("failed to check activeIsEspresso: %w", err)
	}

	batcherAddr := l.Txmgr.From()

	modeActive := (activeIsEspresso && l.Config.Espresso.Enabled) ||
		(!activeIsEspresso && !l.Config.Espresso.Enabled)
	if !modeActive {
		l.Log.Warn("Batcher is not the active batcher, skipping publish",
			"batcherAddr", batcherAddr,
			"activeIsEspresso", activeIsEspresso,
			"EspressoEnabled", l.Config.Espresso.Enabled,
		)
		return false, nil
	}

	// Our mode is active; make sure our sender key is the authorized batcher for it,
	// otherwise every publish reverts (Unauthorized*Batcher) in a loop.
	var expected common.Address
	if activeIsEspresso {
		expected, err = batchAuthenticator.EspressoBatcher(callOpts)
		if err != nil {
			return false, fmt.Errorf("failed to read espressoBatcher: %w", err)
		}
	} else {
		expected, err = l.fallbackBatcherAddr(batchAuthenticator, callOpts)
		if err != nil {
			return false, err
		}
	}

	if batcherAddr != expected {
		l.Log.Warn("Configured batcher key is not the authorized batcher for the active mode, skipping publish",
			"batcherAddr", batcherAddr,
			"expected", expected,
			"activeIsEspresso", activeIsEspresso,
		)
		return false, nil
	}

	return true, nil
}

// fallbackBatcherAddr returns the address the BatchAuthenticator's fallback batcher
func (l *BatchSubmitter) fallbackBatcherAddr(batchAuthenticator *bindings.BatchAuthenticator, opts *bind.CallOpts) (common.Address, error) {
	systemConfigAddr, err := batchAuthenticator.SystemConfig(opts)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to read systemConfig address: %w", err)
	}
	systemConfig, err := bindings.NewSystemConfigCaller(systemConfigAddr, l.L1Client)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to create SystemConfig binding: %w", err)
	}
	batcherHash, err := systemConfig.BatcherHash(opts)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to read batcherHash: %w", err)
	}
	// batcherHash stores the batcher address in its low 20 bytes,
	// which is why we use bytes to address
	return common.BytesToAddress(batcherHash[:]), nil
}
