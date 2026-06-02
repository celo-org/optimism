package derive

import (
	"context"
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-service/eth"
)

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

// NewBatchAuthCaches creates caches sized for the given lookback window.
func NewBatchAuthCaches(lookbackWindow uint64) *BatchAuthCaches {
	// lookbackWindow past blocks + 1 current block + 1 LRU overhead.
	// lru.New only errors on size <= 0.
	size := int(lookbackWindow) + 2
	authCache, _ := lru.New[common.Hash, map[common.Hash]common.Address](size)
	refCache, _ := lru.New[common.Hash, eth.L1BlockRef](size)
	return &BatchAuthCaches{AuthCache: authCache, RefCache: refCache}
}

// ComputeCalldataBatchHash computes keccak256(calldata), matching the BatchAuthenticator
// contract's calldata batch validation path.
func ComputeCalldataBatchHash(data []byte) common.Hash {
	return crypto.Keccak256Hash(data)
}

// ComputeBlobBatchHash computes keccak256(concat(blobHashes)), matching the BatchAuthenticator
// contract's blob batch validation path.
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
	lookbackWindow uint64,
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
			if err != nil {
				return nil, NewTemporaryError(fmt.Errorf("batch auth: failed to fetch receipts for block %d: %w", currentBlock.Number, err))
			}
			events := collectAuthEventsFromReceipts(receipts, authenticatorAddr)
			cache.Add(currentBlock.Hash, events)
			mergeNewest(events)
		}

		if currentBlock.Number == 0 || ref.Number-currentBlock.Number >= lookbackWindow {
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
			if err != nil {
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
