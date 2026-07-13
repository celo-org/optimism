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
// The gate switches at plain Espresso activation (tip.Time >= EspressoTime), a full grace
// period before the verifier enforces event-based authentication. See
// derive.BatchAuthEnforcementDelaySecs for the full grace-period mechanism and why the
// batcher and verifier gates are offset.
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
