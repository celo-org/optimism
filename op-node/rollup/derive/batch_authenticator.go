package derive

import (
	"context"
	"errors"
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// IsEspressoAuthEnforced returns true once event-based batch authentication is enforced
// at the given L1 origin time: the Espresso fork is active AND at least
// BatchAuthEnforcementDelaySecs has elapsed since activation. Before that, derivation
// keeps accepting sender-authenticated batches. See BatchAuthEnforcementDelaySecs
// (params.go) for the full grace-period mechanism.
//
// Exported because the op-batcher publish gate (shouldSkipPublishForActiveSeq) keys
// batch-submission ownership on the exact predicate the verifier enforces with: the
// fallback batcher owns publishing before enforcement, the TEE batcher after.
func IsEspressoAuthEnforced(cfg *rollup.Config, l1OriginTime uint64) bool {
	return cfg.IsEspresso(l1OriginTime) && l1OriginTime >= *cfg.EspressoTime+BatchAuthEnforcementDelaySecs
}

var (
	// BatchInfoAuthenticatedABI is the event signature for
	// BatchInfoAuthenticated(bytes32 commitment, address indexed caller).
	// The commitment is an unindexed (data) argument; only caller is indexed.
	BatchInfoAuthenticatedABI     = "BatchInfoAuthenticated(bytes32,address)"
	BatchInfoAuthenticatedABIHash = crypto.Keccak256Hash([]byte(BatchInfoAuthenticatedABI))
)

// BatchAuthCaches holds the LRU caches used by CollectAuthenticatedBatches.
// Keyed by block hash so they are naturally reorg-safe: after a reorg the
// parent-hash traversal follows a different chain and stale entries are
// never hit. Thread-safe via lru.Cache's internal mutex.
type BatchAuthCaches struct {
	// AuthCache maps L1 block hash to the set of authenticated batch
	// commitments found in that block's receipts, where each commitment maps to
	// the caller (the address that emitted the auth event).
	AuthCache *lru.Cache[common.Hash, map[common.Hash]common.Address]
	// RefCache maps L1 block hash to its L1BlockRef, avoiding redundant
	// L1BlockRefByHash RPC calls during lookback window traversal.
	RefCache *lru.Cache[common.Hash, eth.L1BlockRef]
}

// NewBatchAuthCaches creates caches sized for the BatchAuthLookbackWindow.
func NewBatchAuthCaches() *BatchAuthCaches {
	// The lookback window covers 101 blocks (the ref block plus 100 ancestors),
	// so 101 entries are live during any single traversal. We add +2 (not +1)
	// because the traversal reads newest-to-oldest: the ref block is touched
	// first and so becomes the LRU entry. With exactly 101 slots, inserting the
	// next block's ref would evict the previous ref (its parent) — the very block
	// we're about to read — triggering a cascade of evict-and-refetch through the
	// whole window. The extra slot leaves room for the 101 new window entries plus
	// one stale entry (the block that just fell out of the lookback window). That
	// stale entry, untouched in the current traversal, is the LRU and gets evicted
	// instead, so no cascade occurs.
	// lru.New only errors on size <= 0.
	size := int(BatchAuthLookbackWindow) + 2
	authCache, _ := lru.New[common.Hash, map[common.Hash]common.Address](size)
	refCache, _ := lru.New[common.Hash, eth.L1BlockRef](size)
	return &BatchAuthCaches{AuthCache: authCache, RefCache: refCache}
}

// ComputeCalldataBatchHash computes keccak256(calldata), the commitment a calldata batch
// is authenticated under. BatchAuthenticator.authenticateBatchInfo takes the commitment as
// an opaque bytes32 and never inspects how it was derived, so this encoding is agreed
// off-chain between the batcher and derivation.
func ComputeCalldataBatchHash(data []byte) common.Hash {
	return crypto.Keccak256Hash(data)
}

// ComputeBlobBatchHash computes keccak256(concat(blobHashes)), the same commitment for a
// blob batch, agreed off-chain as above. No live path consumes it while calldata-only DA
// is enforced: derivation drops blob batches before hashing them, and the batcher cannot
// start with blob DA once espresso_time is set.
func ComputeBlobBatchHash(blobHashes []common.Hash) common.Hash {
	concatenated := make([]byte, 32*len(blobHashes))
	for i, h := range blobHashes {
		copy(concatenated[i*32:(i+1)*32], h[:])
	}
	return crypto.Keccak256Hash(concatenated)
}

// collectAuthEventsFromReceipts extracts all authenticated batch commitments from
// the given receipts, mapping each commitment to the caller that emitted the
// BatchInfoAuthenticated event (the indexed Topics[1]). The caller is later
// matched against the batch transaction's L1 sender, so a batch is only accepted
// if the same address both authenticated and submitted it.
func collectAuthEventsFromReceipts(receipts types.Receipts, authenticatorAddr common.Address) map[common.Hash]common.Address {
	result := make(map[common.Hash]common.Address)
	for _, receipt := range receipts {
		if receipt.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, lg := range receipt.Logs {
			if lg.Address != authenticatorAddr {
				continue
			}
			if len(lg.Topics) >= 2 && lg.Topics[0] == BatchInfoAuthenticatedABIHash && len(lg.Data) >= 32 {
				commitment := common.BytesToHash(lg.Data[:32])
				caller := common.BytesToAddress(lg.Topics[1][:])
				result[commitment] = caller
			}
		}
	}
	return result
}

// CollectAuthenticatedBatches scans L1 receipts in the range
// [ref.Number - lookbackWindow, ref.Number] and returns a map from each batch
// commitment hash that was authenticated via a BatchInfoAuthenticated event to
// the caller that emitted it (the event's indexed `caller`). Callers use this to
// require that a batch transaction's L1 sender matches the address that
// authenticated the batch.
//
// This is called once per L1 block by the data source, and the returned set is checked
// against each candidate batch transaction. This avoids rescanning the lookback window
// for every individual batch transaction.
//
// The scan walks newest block to oldest; when the same commitment is authenticated
// in more than one block, the newest event's caller is retained.
//
// Results are cached per block hash in the provided BatchAuthCaches. For consecutive
// L1 blocks the lookback windows overlap by ~99 blocks, so only one new block's
// receipts need to be fetched on each call. The cache is keyed by block hash (not
// number) so it is naturally reorg-safe.
//
// Using event scanning (rather than L1 contract state reads) keeps the derivation
// pipeline compatible with the op-program fault proof environment, which can only
// access L1 block headers, transactions, receipts, and blobs.
func CollectAuthenticatedBatches(
	ctx context.Context,
	fetcher L1Fetcher,
	ref eth.L1BlockRef,
	authenticatorAddr common.Address,
	caches *BatchAuthCaches,
	logger log.Logger,
) (map[common.Hash]common.Address, error) {
	cache := caches.AuthCache
	refCache := caches.RefCache

	// Cache the starting block ref so future calls that traverse through this
	// block (as part of their lookback window) can resolve it without an RPC call.
	refCache.Add(ref.Hash, ref)

	// Traversal is newest-block-first, so a commitment already in the map was
	// seen in a newer block; mergeNewest keeps that newer caller (see doc above).
	allAuthenticated := make(map[common.Hash]common.Address)
	mergeNewest := func(src map[common.Hash]common.Address) {
		for commitment, caller := range src {
			if _, seen := allAuthenticated[commitment]; !seen {
				allAuthenticated[commitment] = caller
			}
		}
	}

	currentBlock := ref
	receiptCacheHits := 0
	refCacheHits := 0

	for {
		// Check receipt cache first
		if cached, ok := cache.Get(currentBlock.Hash); ok {
			mergeNewest(cached)
			receiptCacheHits++
		} else {
			// Cache miss: fetch receipts, extract events, cache the result
			_, receipts, err := fetcher.FetchReceipts(ctx, currentBlock.Hash)
			if errors.Is(err, ethereum.NotFound) {
				// A block in the lookback window is no longer available, e.g. an L1
				// reorg orphaned it. Treat this like the data sources treat a missing
				// ref block (calldata_source.go / blob_data_source.go): signal a reset so
				// the pipeline re-derives from a canonical origin, rather than retrying
				// the same step forever as a temporary error.
				return nil, NewResetError(fmt.Errorf("batch auth: receipts for block %d not found: %w", currentBlock.Number, err))
			} else if err != nil {
				return nil, NewTemporaryError(fmt.Errorf("batch auth: failed to fetch receipts for block %d: %w", currentBlock.Number, err))
			}
			events := collectAuthEventsFromReceipts(receipts, authenticatorAddr)
			cache.Add(currentBlock.Hash, events)
			mergeNewest(events)
		}

		if currentBlock.Number == 0 || ref.Number-currentBlock.Number >= BatchAuthLookbackWindow {
			break
		}

		// Resolve parent block ref, using the cache to avoid redundant RPC calls.
		// Consecutive L1 blocks share ~99 blocks in their lookback windows, so
		// after the first full traversal almost every parent lookup is a cache hit.
		parentHash := currentBlock.ParentHash
		if cachedRef, ok := refCache.Get(parentHash); ok {
			currentBlock = cachedRef
			refCacheHits++
		} else {
			parentRef, err := fetcher.L1BlockRefByHash(ctx, parentHash)
			if errors.Is(err, ethereum.NotFound) {
				// See the FetchReceipts NotFound case above: a missing ancestor means the
				// lookback window crossed a reorg, so reset rather than retry forever.
				return nil, NewResetError(fmt.Errorf("batch auth: L1 block ref %s not found: %w", parentHash.String(), err))
			} else if err != nil {
				return nil, NewTemporaryError(fmt.Errorf("batch auth: failed to fetch L1 block ref %s: %w", parentHash.String(), err))
			}
			refCache.Add(parentHash, parentRef)
			currentBlock = parentRef
		}
	}

	logger.Debug("collected authenticated batches from lookback window",
		"count", len(allAuthenticated), "fromBlock", currentBlock.Number, "toBlock", ref.Number,
		"receiptCacheHits", receiptCacheHits, "refCacheHits", refCacheHits)
	return allAuthenticated, nil
}
