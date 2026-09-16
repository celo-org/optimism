package batcher

import (
	"errors"
	"fmt"
	"testing"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	"github.com/stretchr/testify/require"
)

// TestEvaluateSubmission covers evaluateSubmission, the pure decision function
// submitToEspresso uses to classify the result of submitting an Espresso
// transaction: no error -> Handle (proceed to receipt verification), a permanent
// error -> Skip (drop the tx), an ephemeral or unclassified error ->
// RetrySubmission (submit again).
func TestEvaluateSubmission(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want JobEvaluation
	}{
		{"success", nil, Handle},
		{"permanent", espressoClient.ErrPermanent, Skip},
		{"permanent wrapped", fmt.Errorf("boom: %w", espressoClient.ErrPermanent), Skip},
		{"ephemeral", espressoClient.ErrEphemeral, RetrySubmission},
		{"ephemeral wrapped", fmt.Errorf("boom: %w", espressoClient.ErrEphemeral), RetrySubmission},
		{"unclassified errors are retried", errors.New("something else"), RetrySubmission},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, evaluateSubmission(tt.err))
		})
	}
}

// TestEvaluateVerification covers evaluateVerification, the pure decision
// function confirmReceipt uses after each receipt-poll attempt: receipt found ->
// Handle (confirmed, done), a permanent error -> Skip, verification that has
// outlasted its block-count or wall-clock budget -> RetrySubmission (re-submit a
// fresh copy), otherwise -> RetryVerification (keep polling).
func TestEvaluateVerification(t *testing.T) {
	const maxBlocks = 5
	const safetyTimeout = 5 * time.Minute
	s := &espressoTransactionSubmitter{
		verifyReceiptMaxBlocks:     maxBlocks,
		verifyReceiptSafetyTimeout: safetyTimeout,
	}

	// now is the reference "started" time; recent enough not to trip the safety timeout.
	now := time.Now()

	tests := []struct {
		name          string
		err           error
		startHeight   uint64
		currentHeight uint64
		startTime     time.Time
		want          JobEvaluation
	}{
		{"success", nil, 100, 100, now, Handle},
		{"permanent", espressoClient.ErrPermanent, 100, 100, now, Skip},
		{
			name: "within both windows keeps polling",
			err:  espressoClient.ErrEphemeral, startHeight: 100, currentHeight: 104, startTime: now,
			want: RetryVerification,
		},
		{
			name: "block-count timeout re-submits",
			err:  espressoClient.ErrEphemeral, startHeight: 100, currentHeight: 105, startTime: now,
			want: RetrySubmission,
		},
		{
			name: "zero startHeight disables the block-count check",
			// Height tracker hasn't produced a value yet: even a huge currentHeight
			// must not trip the block-count timeout, so we keep polling.
			err: espressoClient.ErrEphemeral, startHeight: 0, currentHeight: 1_000_000, startTime: now,
			want: RetryVerification,
		},
		{
			name: "wall-clock safety timeout re-submits",
			err:  espressoClient.ErrEphemeral, startHeight: 100, currentHeight: 101,
			startTime: now.Add(-2 * safetyTimeout),
			want:      RetrySubmission,
		},
		{
			name: "unclassified error is retried like ephemeral",
			err:  errors.New("something else"), startHeight: 100, currentHeight: 104, startTime: now,
			want: RetryVerification,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, s.evaluateVerification(tt.err, tt.startHeight, tt.currentHeight, tt.startTime))
		})
	}
}
