package batcher

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

// isBatcherActive checks if the current batcher is the active one by querying
// the BatchAuthenticator contract. Returns true if this batcher instance should
// be publishing batches, false if it should stay idle.
//
// It applies two gates:
//  1. Mode: the contract's activeIsEspresso flag must match this node's role
//     (Config.Espresso.Enabled). activeIsEspresso==true means the Espresso batcher
//     is active; false means the fallback batcher is active.
//  2. Identity: once the mode matches, the configured sender key (Txmgr.From) must
//     be the authorized batcher for that mode, otherwise every authenticateBatchInfo
//     call reverts (Unauthorized{Espresso,Fallback}Batcher) and the batcher loops.
//
// This runs on every publish tick and costs two eth_calls in either mode.
func (l *BatchSubmitter) isBatcherActive(ctx context.Context) (bool, error) {
	if l.batchAuth == nil {
		return false, errors.New("no BatchAuthenticator configured")
	}

	activeIsEspresso, err := l.batchAuth.ActiveIsEspresso(ctx)
	if err != nil {
		return false, err
	}

	batcherAddr := l.Txmgr.From()

	if activeIsEspresso != l.Config.Espresso.Enabled {
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
		expected, err = l.batchAuth.EspressoBatcher(ctx)
	} else {
		expected, err = l.batchAuth.FallbackBatcher(ctx)
	}
	if err != nil {
		return false, err
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
