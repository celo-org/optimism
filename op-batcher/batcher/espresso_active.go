package batcher

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
)

// checkBatcherActive queries the BatchAuthenticator contract to decide whether
// this batcher should be publishing. It applies two checks:
//  1. Mode: the contract's activeIsEspresso flag must match this node's role
//     (Config.Espresso.Enabled). A mismatch reports PublishGateModeInactive.
//  2. Identity: the sender key (Txmgr.From) must be the authorized batcher for
//     that mode, otherwise every authenticateBatchInfo call reverts. A wrong
//     key reports PublishGateKeyUnauthorized.
//
// It runs before every publishTxToL1 call. A mode mismatch costs one eth_call
// and a matching mode at most two in steady state. A batcher that is not the
// active one takes a skip branch every time, so both warnings are throttled.
func (l *BatchSubmitter) checkBatcherActive(ctx context.Context) (metrics.PublishGateState, error) {
	if l.batchAuth == nil {
		return "", errors.New("no BatchAuthenticator configured")
	}

	activeIsEspresso, err := l.batchAuth.ActiveIsEspresso(ctx)
	if err != nil {
		return "", err
	}

	batcherAddr := l.Txmgr.From()

	if activeIsEspresso != l.Config.Espresso.Enabled {
		l.degradedLog.Warn(l.Log, "batcherModeInactive", "Batcher is not the active batcher, skipping publish",
			"batcherAddr", batcherAddr,
			"activeIsEspresso", activeIsEspresso,
			"EspressoEnabled", l.Config.Espresso.Enabled,
		)
		// The key is only checked while our mode is active. Reset its state so a
		// still-wrong key warns afresh when the mode comes back, rather than staying
		// throttled and later reporting a duration that spans the inactive period.
		l.degradedLog.Clear(l.Log, "batcherKeyUnauthorized", "Batcher mode is inactive, suspending the batcher key check")
		return metrics.PublishGateModeInactive, nil
	}
	l.degradedLog.Clear(l.Log, "batcherModeInactive", "Batcher mode is active again")

	// Our mode is active; make sure our sender key is the authorized batcher for it,
	// otherwise every publish reverts (Unauthorized*Batcher) in a loop.
	var expected common.Address
	if activeIsEspresso {
		expected, err = l.batchAuth.EspressoBatcher(ctx)
	} else {
		expected, err = l.batchAuth.FallbackBatcher(ctx)
	}
	if err != nil {
		return "", err
	}

	if batcherAddr != expected {
		l.degradedLog.Warn(l.Log, "batcherKeyUnauthorized", "Configured batcher key is not the authorized batcher for the active mode, skipping publish",
			"batcherAddr", batcherAddr,
			"expected", expected,
			"activeIsEspresso", activeIsEspresso,
		)
		return metrics.PublishGateKeyUnauthorized, nil
	}
	l.degradedLog.Clear(l.Log, "batcherKeyUnauthorized", "Configured batcher key is authorized again")

	return metrics.PublishGateActive, nil
}
