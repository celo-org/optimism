package espresso

import (
	"context"

	op "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
)

// EspressoStreamer defines the interface for the Espresso streamer.
type EspressoStreamer[B op.Batch] interface {
	// Update will update the `EspressoStreamer“ by attempting to ensure that
	// the next call to the `Next` method will return a `Batch`.
	//
	// It attempts to ensure the existence of a next batch, provided no errors
	// occur when communicating with HotShot, by processing Blocks retrieved
	// from `HotShot` in discreet batches. If each processing of a batch of
	// blocks will not yield a new `Batch`, then it will continue to process
	// the next batch of blocks from HotShot until it runs out of blocks to
	// process.
	//
	//	NOTE: this method is best effort.  It is unable to guarantee that the
	//	next call to `Next` will return a batch.  However, the only things
	//	that will prevent the next call to `Next` from returning a batch is if
	//	there are no more HotShot blocks to process currently, or if an error
	//	occurs when communicating with HotShot.
	Update(ctx context.Context) error

	// Refresh updates the local references of the EspressoStreamer to the
	// specified values.
	//
	// These values can be used to help determine whether the Streamer needs
	// to be reset or not.
	//
	// NOTE: This will only automatically reset the Streamer if the
	// `safeBatchNumber` moves backwards.
	Refresh(ctx context.Context, finalizedL1 eth.L1BlockRef, safeBatchNumber uint64, safeL1Origin eth.BlockID) error

	// RefreshSafeL1Origin updates the safe L1 origin for the streamer. This is
	// used to help the streamer determine if it needs to be reset or not based
	// on the safe L1 origin moving backwards.
	//
	// NOTE: This will only automatically reset the Streamer if the
	// `safeL1Origin` moves backwards.
	RefreshSafeL1Origin(safeL1Origin eth.BlockID)

	// Reset will reset the Streamer to the last known good safe state.
	// This generally means resetting to the last know good safe batch
	// position, but in the case of consuming blocks from Espresso, it will
	// also reset the starting Espresso block position to the last known
	// good safe block position there as well.
	Reset()

	// UnmarshalBatch is a convenience method that allows the caller to
	// attempt to unmarshal a batch from the provided byte slice.
	UnmarshalBatch(b []byte) (*B, error)

	// HasNext checks to see if there are any batches left to read in the
	// streamer.
	HasNext(ctx context.Context) bool

	// Next attempts to return the next batch from the streamer.  If there
	// are no batches left to read, at the moment of the call, it will return
	// nil.
	Next(ctx context.Context) *B

	// Peek attempts to return the next batch from the streamer without advancing the streamer's position.
	// If there are no batches left to read, at the moment of the call, it will return nil.
	Peek(ctx context.Context) *B

	// SetProperHead drains stale/wrong-fork entries from the buffer front,
	// positioning it at the correct fork for the next Peek call. Should be called
	// when Peek returns a batch whose parentHash doesn't match the current chain tip.
	// No-ops if headBatch's block number doesn't match the expected next batch position.
	SetProperHead(parentHash common.Hash)
}
