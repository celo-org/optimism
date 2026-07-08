package batcher

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// isFallbackAuthRequired reports whether the fallback (non-TEE) batcher must
// route its batch txs through BatchAuthenticator.authenticateBatchInfo before
// posting to the BatchInbox. It returns false if the rollup config has a
// zero BatchAuthenticatorAddress, indicating that the BatchAuthenticator-based
// authentication path is not in use.
//
// The batcher switches at Espresso activation (tip.Time >= EspressoTime), while the
// verifier only starts enforcing event-based authentication one grace period later
// (derive.BatchAuthEnforcementDelay).
// The gap absorbs the delay between the batcher's gate decision (based on L1 tip time)
// and the batch tx's eventual inclusion (based on containing-block time): a batch
// decided pre-fork that lands post-activation is still accepted under sender
// authorization as long as its inclusion delay stays below the grace period (~20 min).
// The reverse asymmetry (authenticated tx lands before enforcement) is harmless:
// pre-enforcement the verifier uses sender-based authorization and the auth event is
// just an unrelated L1 tx that does not affect derivation.
func (l *BatchSubmitter) isFallbackAuthRequired(ctx context.Context) (bool, error) {
	if l.RollupConfig.BatchAuthenticatorAddress == (common.Address{}) {
		return false, nil
	}
	tip, err := l.l1Tip(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to fetch L1 tip for fallback-auth gate: %w", err)
	}
	return l.RollupConfig.IsEspresso(tip.Time), nil
}
