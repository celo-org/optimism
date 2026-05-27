package batcher

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// hasBatchAuthenticator returns true if the rollup config has a non-zero
// BatchAuthenticatorAddress, indicating that the BatchAuthenticator-based
// authentication path is in use.
func (l *BatchSubmitter) hasBatchAuthenticator() bool {
	return l.RollupConfig.BatchAuthenticatorAddress != (common.Address{})
}

// isFallbackAuthRequired reports whether the fallback (non-TEE) batcher must
// route its batch txs through BatchAuthenticator.authenticateBatchInfo before
// posting to the BatchInbox.
//
// This decision must align with the verifier's per-L1-block fork gate
// (DataSourceConfig.isEspressoEnforcement, which evaluates the hardfork
// activation predicate against the *containing* L1 block's timestamp). Since
// the tx is not yet mined at decision time, its eventual containing block
// has a strictly greater timestamp than the L1 tip the batcher observes:
//
//	l1Tip.Time (batcher's view) < l1OriginTime (block containing the tx)
//
// Without compensation, in the window [forkTime − maxL1InclusionDelay, forkTime)
// the batcher would skip authenticateBatchInfo while the verifier — once the
// tx lands in a post-fork block — would require the resulting
// BatchInfoAuthenticated event, silently dropping the batch.
//
// To prevent this, we add Config.FallbackAuthLeadTime to the L1 tip's
// timestamp before evaluating the fork predicate. This makes the batcher
// start authenticating slightly before the verifier requires it. The reverse
// asymmetry (authenticated tx lands pre-fork) is harmless: pre-fork the
// verifier uses sender-based authorization and the auth event is just an
// unrelated L1 tx that does not affect derivation.
func (l *BatchSubmitter) isFallbackAuthRequired(ctx context.Context) (bool, error) {
	tip, err := l.l1Tip(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to fetch L1 tip for fallback-auth gate: %w", err)
	}
	leadSec := uint64(l.Config.FallbackAuthLeadTime / time.Second)
	return l.RollupConfig.IsEspresso(tip.Time + leadSec), nil
}
