package batcher

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum-optimism/optimism/espresso/bindings"
)

// isBatcherActive checks if the current batcher is the active one by querying
// the BatchAuthenticator contract. Returns true if this batcher instance should
// be publishing batches, false if it should stay idle.
//
// The active batcher is determined by the contract's activeIsEspresso flag:
//   - If activeIsEspresso is true, the Espresso batcher address is active
//   - If activeIsEspresso is false, the fallback batcher address is active
//
// This method compares the batcher's own role (Config.Espresso.Enabled)
// against the contract's activeIsEspresso flag.
func (l *BatchSubmitter) isBatcherActive(ctx context.Context) (bool, error) {
	// Check if contract code exists at the address
	code, err := l.L1Client.CodeAt(ctx, l.RollupConfig.BatchAuthenticatorAddress, nil)
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

	isActive := (activeIsEspresso && l.Config.Espresso.Enabled) ||
		(!activeIsEspresso && !l.Config.Espresso.Enabled)

	if !isActive {
		l.Log.Warn("Batcher is not the active batcher, skipping publish",
			"batcherAddr", batcherAddr,
			"activeIsEspresso", activeIsEspresso,
			"EspressoEnabled", l.Config.Espresso.Enabled,
		)
	}

	return isActive, nil
}
