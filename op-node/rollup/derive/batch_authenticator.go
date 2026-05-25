package derive

import (
	"context"
	"fmt"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-service/eth"
)

var (
	// BatchInfoAuthenticatedABI is the event signature for BatchInfoAuthenticated(bytes32 indexed commitment).
	BatchInfoAuthenticatedABI     = "BatchInfoAuthenticated(bytes32)"
	BatchInfoAuthenticatedABIHash = crypto.Keccak256Hash([]byte(BatchInfoAuthenticatedABI))

	// batchAuthCache is a global LRU cache mapping L1 block hash to the set of
	// authenticated batch commitment hashes found in that block's receipts.
	// Keyed by block hash so it is naturally reorg-safe: after a reorg the
	// parent-hash traversal follows a different chain and stale entries are
	// never hit. Thread-safe via lru.Cache's internal mutex.
	batchAuthCache     *lru.Cache[common.Hash, map[common.Hash]bool]
	batchAuthCacheOnce sync.Once

	// blockRefCache is a global LRU cache mapping L1 block hash to its L1BlockRef.
	// This avoids redundant L1BlockRefByHash RPC calls during the lookback window
	// traversal: consecutive L1 blocks share ~99 blocks in their lookback windows,
	// so almost every parent-hash lookup hits the cache after the first full traversal.
	// Keyed by block hash for natural reorg safety (same rationale as batchAuthCache).
	blockRefCache     *lru.Cache[common.Hash, eth.L1BlockRef]
	blockRefCacheOnce sync.Once
)

// resetBatchAuthCaches resets both global caches (receipt and block ref).
// This is only intended for use in tests to ensure isolation between test cases.
func resetBatchAuthCaches() {
	batchAuthCache = nil
	batchAuthCacheOnce = sync.Once{}
	blockRefCache = nil
	blockRefCacheOnce = sync.Once{}
}

func getCache[T any](cache **lru.Cache[common.Hash, T], once *sync.Once, size int) *lru.Cache[common.Hash, T] {
	once.Do(func() {
		// lookbackWindow past blocks + 1 current block + 1 LRU overhead.
		// lru.New only errors on size <= 0.
		*cache, _ = lru.New[common.Hash, T](size + 2)
	})
	return *cache
}

func getBatchAuthCache(lookbackWindow uint64) *lru.Cache[common.Hash, map[common.Hash]bool] {
	return getCache(&batchAuthCache, &batchAuthCacheOnce, int(lookbackWindow))
}

func getBlockRefCache(lookbackWindow uint64) *lru.Cache[common.Hash, eth.L1BlockRef] {
	return getCache(&blockRefCache, &blockRefCacheOnce, int(lookbackWindow))
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

// FindBatchAuthEvent scans the given receipts for a BatchInfoAuthenticated event
// emitted by authenticatorAddr with a commitment matching batchHash.
// Returns true if such an event is found.
func FindBatchAuthEvent(receipts types.Receipts, authenticatorAddr common.Address, batchHash common.Hash) bool {
	for _, receipt := range receipts {
		if receipt.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, lg := range receipt.Logs {
			if lg.Address != authenticatorAddr {
				continue
			}
			// BatchInfoAuthenticated has 2 topics: event sig, indexed commitment
			if len(lg.Topics) >= 2 &&
				lg.Topics[0] == BatchInfoAuthenticatedABIHash &&
				lg.Topics[1] == batchHash {
				return true
			}
		}
	}
	return false
}

// collectAuthEventsFromReceipts extracts all authenticated batch hashes from the given receipts.
// It returns the set of commitment hashes that have been authenticated by the given authenticator.
func collectAuthEventsFromReceipts(receipts types.Receipts, authenticatorAddr common.Address) map[common.Hash]bool {
	result := make(map[common.Hash]bool)
	for _, receipt := range receipts {
		if receipt.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, lg := range receipt.Logs {
			if lg.Address != authenticatorAddr {
				continue
			}
			if len(lg.Topics) >= 2 && lg.Topics[0] == BatchInfoAuthenticatedABIHash {
				result[lg.Topics[1]] = true
			}
		}
	}
	return result
}

// CollectAuthenticatedBatches scans L1 receipts in the range
// [ref.Number - lookbackWindow, ref.Number] and returns the set of all
// batch commitment hashes that were authenticated via BatchInfoAuthenticated events.
//
// This is called once per L1 block by the data source, and the returned set is checked
// against each candidate batch transaction. This avoids rescanning the lookback window
// for every individual batch transaction.
//
// Results are cached per block hash in a global LRU cache. For consecutive L1 blocks
// the lookback windows overlap by ~99 blocks, so only one new block's receipts need
// to be fetched on each call. The cache is keyed by block hash (not number) so it is
// naturally reorg-safe.
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
	logger log.Logger,
) (map[common.Hash]bool, error) {
	cache := getBatchAuthCache(lookbackWindow)
	refCache := getBlockRefCache(lookbackWindow)

	// Cache the starting block ref so future calls that traverse through this
	// block (as part of their lookback window) can resolve it without an RPC call.
	refCache.Add(ref.Hash, ref)

	allAuthenticated := make(map[common.Hash]bool)
	currentBlock := ref
	receiptCacheHits := 0
	refCacheHits := 0

	for {
		// Check receipt cache first
		if cached, ok := cache.Get(currentBlock.Hash); ok {
			for h := range cached {
				allAuthenticated[h] = true
			}
			receiptCacheHits++
		} else {
			// Cache miss: fetch receipts, extract events, cache the result
			_, receipts, err := fetcher.FetchReceipts(ctx, currentBlock.Hash)
			if err != nil {
				return nil, NewTemporaryError(fmt.Errorf("batch auth: failed to fetch receipts for block %d: %w", currentBlock.Number, err))
			}
			events := collectAuthEventsFromReceipts(receipts, authenticatorAddr)
			cache.Add(currentBlock.Hash, events)
			for h := range events {
				allAuthenticated[h] = true
			}
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
