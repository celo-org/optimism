package derive

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// count the tagging info as 200 in terms of buffer size.
const frameOverhead = 200

// frameSize calculates the size of the frame + overhead for
// storing the frame. The sum of the frame size of each frame in
// a channel determines the channel's size. The sum of the channel
// sizes is used for pruning & compared against `MaxChannelBankSize`
func frameSize(frame Frame) uint64 {
	return uint64(len(frame.Data)) + frameOverhead
}

// MaxSpanBatchElementCount is the maximum number of blocks, transactions in total,
// or transaction per block allowed in a span batch.
const MaxSpanBatchElementCount = 10_000_000

// BatchAuthLookbackWindow is the maximum number of L1 blocks before a batch submission to
// scan for a BatchInfoAuthenticated event. The authentication transaction must land
// in this window (or in the same block as the batch submission) for the batch to be
// considered valid post-Espresso.
//
// At ~12s per L1 block, 100 blocks ≈ 20 minutes. This gives the batcher time to land
// the batch data transaction on L1 after the authentication transaction, even under
// L1 congestion or batcher restarts.
const BatchAuthLookbackWindow uint64 = 100

// BatchAuthEnforcementDelaySecs is the number of seconds after the EspressoTime activation
// during which derivation still accepts sender-authenticated batches. Event-based batch
// authentication is only enforced for L1 blocks with origin time >= EspressoTime +
// BatchAuthEnforcementDelaySecs.
//
// This is the canonical description of the grace-period mechanism; isEspressoAuthEnforced
// (batch_authenticator.go), the op-batcher fallback-auth gate (isFallbackAuthRequired in
// espresso_active.go), and the fork-boundary tests all defer here.
//
// Behavior by L1 origin time t relative to EspressoTime (E):
//   - before enforcement (t < E+delay, i.e. pre-fork or within the grace window):
//     upstream sender-based authorization — the batch tx's L1 sender must equal the
//     configured batcher address. Auth events are not scanned.
//   - enforced (t >= E+delay): event-based authentication only — the batch's commitment
//     must have a BatchInfoAuthenticated event within BatchAuthLookbackWindow AND the tx's
//     L1 sender must equal the caller that emitted it. Sender-only authorization is rejected.
//
// The grace period lets the batcher switch to authenticated submission at activation
// without a configured lead time. The batcher's gate flips at plain activation (L1 tip
// time >= EspressoTime), a full grace period before the verifier enforces. Because the
// batcher decides on the L1 tip time while the verifier judges by the batch tx's
// containing-block time, a batch decided pre-fork can land in a post-activation block;
// accepting sender auth through the window keeps it valid as long as its inclusion delay
// stays below the grace period. The reverse asymmetry (an authenticated tx landing before
// enforcement) is harmless: pre-enforcement the verifier uses sender-based authorization
// and the auth event is just an unrelated L1 tx that does not affect derivation.
//
// Sized to the duration of one full BatchAuthLookbackWindow at the nominal 12s L1 slot
// time (~20 minutes) — far above any realistic L1 inclusion delay. Under missed L1
// slots the delay covers fewer than BatchAuthLookbackWindow blocks, which is fine: the
// bound that matters is time (inclusion delay), not block count.
const BatchAuthEnforcementDelaySecs uint64 = BatchAuthLookbackWindow * 12

// DuplicateErr is returned when a newly read frame is already known
var DuplicateErr = errors.New("duplicate frame")

// ChannelIDLength defines the length of the channel IDs
const ChannelIDLength = 16

// ChannelID is an opaque identifier for a channel. It is 128 bits to be globally unique.
type ChannelID [ChannelIDLength]byte

func (id ChannelID) String() string {
	return fmt.Sprintf("%x", id[:])
}

// TerminalString implements log.TerminalStringer, formatting a string for console output during logging.
func (id ChannelID) TerminalString() string {
	return fmt.Sprintf("%x..%x", id[:3], id[13:])
}

func (id ChannelID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *ChannelID) UnmarshalText(text []byte) error {
	h, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}
	if len(h) != ChannelIDLength {
		return errors.New("invalid length")
	}
	copy(id[:], h)
	return nil
}
