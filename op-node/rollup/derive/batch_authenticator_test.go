package derive

import (
	"context"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/espresso"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

func TestComputeCalldataBatchHash(t *testing.T) {
	data := []byte("hello world")
	hash := ComputeCalldataBatchHash(data)
	expected := crypto.Keccak256Hash(data)
	require.Equal(t, expected, hash)
}

func TestComputeCalldataBatchHashEmpty(t *testing.T) {
	hash := ComputeCalldataBatchHash([]byte{})
	expected := crypto.Keccak256Hash([]byte{})
	require.Equal(t, expected, hash)
}

func TestComputeBlobBatchHash(t *testing.T) {
	h1 := common.HexToHash("0x0100000000000000000000000000000000000000000000000000000000000001")
	h2 := common.HexToHash("0x0100000000000000000000000000000000000000000000000000000000000002")

	hash := ComputeBlobBatchHash([]common.Hash{h1, h2})

	// Manually compute expected: keccak256(h1 ++ h2)
	concatenated := make([]byte, 64)
	copy(concatenated[0:32], h1[:])
	copy(concatenated[32:64], h2[:])
	expected := crypto.Keccak256Hash(concatenated)
	require.Equal(t, expected, hash)
}

func TestComputeBlobBatchHashSingle(t *testing.T) {
	h := common.HexToHash("0xabcdef")
	hash := ComputeBlobBatchHash([]common.Hash{h})
	expected := crypto.Keccak256Hash(h[:])
	require.Equal(t, expected, hash)
}

func TestFindBatchAuthEvent(t *testing.T) {
	authenticatorAddr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	batchHash := crypto.Keccak256Hash([]byte("test batch data"))

	t.Run("event found", func(t *testing.T) {
		receipts := types.Receipts{
			{
				Status: types.ReceiptStatusSuccessful,
				Logs: []*types.Log{
					{
						Address: authenticatorAddr,
						Topics: []common.Hash{
							BatchInfoAuthenticatedABIHash,
							batchHash,
						},
					},
				},
			},
		}
		require.True(t, FindBatchAuthEvent(receipts, authenticatorAddr, batchHash))
	})

	t.Run("event not found - wrong hash", func(t *testing.T) {
		wrongHash := crypto.Keccak256Hash([]byte("wrong data"))
		receipts := types.Receipts{
			{
				Status: types.ReceiptStatusSuccessful,
				Logs: []*types.Log{
					{
						Address: authenticatorAddr,
						Topics: []common.Hash{
							BatchInfoAuthenticatedABIHash,
							wrongHash,
						},
					},
				},
			},
		}
		require.False(t, FindBatchAuthEvent(receipts, authenticatorAddr, batchHash))
	})

	t.Run("event not found - wrong address", func(t *testing.T) {
		wrongAddr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
		receipts := types.Receipts{
			{
				Status: types.ReceiptStatusSuccessful,
				Logs: []*types.Log{
					{
						Address: wrongAddr,
						Topics: []common.Hash{
							BatchInfoAuthenticatedABIHash,
							batchHash,
						},
					},
				},
			},
		}
		require.False(t, FindBatchAuthEvent(receipts, authenticatorAddr, batchHash))
	})

	t.Run("event not found - reverted receipt", func(t *testing.T) {
		receipts := types.Receipts{
			{
				Status: types.ReceiptStatusFailed,
				Logs: []*types.Log{
					{
						Address: authenticatorAddr,
						Topics: []common.Hash{
							BatchInfoAuthenticatedABIHash,
							batchHash,
						},
					},
				},
			},
		}
		require.False(t, FindBatchAuthEvent(receipts, authenticatorAddr, batchHash))
	})

	t.Run("event not found - empty receipts", func(t *testing.T) {
		require.False(t, FindBatchAuthEvent(types.Receipts{}, authenticatorAddr, batchHash))
	})

	t.Run("event found among multiple receipts", func(t *testing.T) {
		receipts := types.Receipts{
			{
				Status: types.ReceiptStatusSuccessful,
				Logs: []*types.Log{
					{
						Address: common.HexToAddress("0x1111"),
						Topics:  []common.Hash{common.HexToHash("0xdead")},
					},
				},
			},
			{
				Status: types.ReceiptStatusSuccessful,
				Logs: []*types.Log{
					{
						Address: authenticatorAddr,
						Topics: []common.Hash{
							BatchInfoAuthenticatedABIHash,
							batchHash,
						},
					},
				},
			},
		}
		require.True(t, FindBatchAuthEvent(receipts, authenticatorAddr, batchHash))
	})
}

// buildL1Chain creates a chain of L1BlockRef values with proper parent-hash linkage.
// The chain goes from block number `start` to `end` (inclusive).
// Returns a slice indexed by block number (relative to start), and the full map by number.
func buildL1Chain(rng *rand.Rand, start, end uint64) map[uint64]eth.L1BlockRef {
	chain := make(map[uint64]eth.L1BlockRef)
	for num := start; num <= end; num++ {
		ref := eth.L1BlockRef{
			Number: num,
			Hash:   testutils.RandomHash(rng),
		}
		if num > start {
			ref.ParentHash = chain[num-1].Hash
		}
		chain[num] = ref
	}
	return chain
}

func TestCollectAuthenticatedBatches(t *testing.T) {
	resetBatchAuthCaches()
	logger := testlog.Logger(t, log.LevelDebug)
	ctx := context.Background()
	rng := rand.New(rand.NewSource(1234))

	authenticatorAddr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	batchHash := crypto.Keccak256Hash([]byte("test batch data"))

	// Build a matching receipt
	matchingReceipts := types.Receipts{
		{
			Status: types.ReceiptStatusSuccessful,
			Logs: []*types.Log{
				{
					Address: authenticatorAddr,
					Topics: []common.Hash{
						BatchInfoAuthenticatedABIHash,
						batchHash,
					},
				},
			},
		},
	}
	emptyReceipts := types.Receipts{}

	// expectChainTraversal sets up mock expectations for a backward parent-hash
	// traversal from chain[end] down to chain[start]. For each block it expects
	// FetchReceipts (by hash), and for all blocks except the first (end) it
	// expects L1BlockRefByHash to resolve the parent hash.
	// receiptsByBlock allows overriding receipts for specific block numbers.
	expectChainTraversal := func(l1F *testutils.MockL1Source, chain map[uint64]eth.L1BlockRef, start, end uint64, receiptsByBlock map[uint64]types.Receipts) {
		for num := end; num >= start; num-- {
			ref := chain[num]
			receipts := emptyReceipts
			if r, ok := receiptsByBlock[num]; ok {
				receipts = r
			}
			l1F.ExpectFetchReceipts(ref.Hash, nil, receipts, nil)
			// L1BlockRefByHash is called for every block except the first one (ref itself)
			if num > start {
				l1F.ExpectL1BlockRefByHash(chain[num-1].Hash, chain[num-1], nil)
			}
			if num == 0 {
				break // avoid underflow
			}
		}
	}

	t.Run("found in same block", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		chain := buildL1Chain(rng, 100, 200)
		ref := chain[200]

		// Auth event is in block 200 (same block as ref). Traversal goes 200 -> 100.
		expectChainTraversal(l1F, chain, 100, 200, map[uint64]types.Receipts{
			200: matchingReceipts,
		})

		result, err := CollectAuthenticatedBatches(ctx, l1F, ref, authenticatorAddr, espresso.DefaultBatchAuthLookbackWindow, logger)
		require.NoError(t, err)
		require.True(t, result[batchHash])
		require.Len(t, result, 1)
		l1F.AssertExpectations(t)
	})

	t.Run("found in earliest block of window", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		chain := buildL1Chain(rng, 100, 200)
		ref := chain[200]

		// Auth event is in block 100 (last block of the lookback window).
		expectChainTraversal(l1F, chain, 100, 200, map[uint64]types.Receipts{
			100: matchingReceipts,
		})

		result, err := CollectAuthenticatedBatches(ctx, l1F, ref, authenticatorAddr, espresso.DefaultBatchAuthLookbackWindow, logger)
		require.NoError(t, err)
		require.True(t, result[batchHash])
		require.Len(t, result, 1)
		l1F.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		chain := buildL1Chain(rng, 100, 200)
		ref := chain[200]

		// No auth event in any block in the window
		expectChainTraversal(l1F, chain, 100, 200, nil)

		result, err := CollectAuthenticatedBatches(ctx, l1F, ref, authenticatorAddr, espresso.DefaultBatchAuthLookbackWindow, logger)
		require.NoError(t, err)
		require.Len(t, result, 0)
		l1F.AssertExpectations(t)
	})

	t.Run("low block number - window clamps to 0", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		chain := buildL1Chain(rng, 0, 10)
		ref := chain[10]

		// Window should clamp to [0, 10]. Auth event is in block 10.
		expectChainTraversal(l1F, chain, 0, 10, map[uint64]types.Receipts{
			10: matchingReceipts,
		})

		result, err := CollectAuthenticatedBatches(ctx, l1F, ref, authenticatorAddr, espresso.DefaultBatchAuthLookbackWindow, logger)
		require.NoError(t, err)
		require.True(t, result[batchHash])
		require.Len(t, result, 1)
		l1F.AssertExpectations(t)
	})

	t.Run("multiple hashes collected", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		chain := buildL1Chain(rng, 0, 10)
		ref := chain[10]

		batchHash2 := crypto.Keccak256Hash([]byte("second batch"))
		multiReceipts := types.Receipts{
			{
				Status: types.ReceiptStatusSuccessful,
				Logs: []*types.Log{
					{
						Address: authenticatorAddr,
						Topics: []common.Hash{
							BatchInfoAuthenticatedABIHash,
							batchHash,
						},
					},
					{
						Address: authenticatorAddr,
						Topics: []common.Hash{
							BatchInfoAuthenticatedABIHash,
							batchHash2,
						},
					},
				},
			},
		}

		// Both auth events are in block 10
		expectChainTraversal(l1F, chain, 0, 10, map[uint64]types.Receipts{
			10: multiReceipts,
		})

		result, err := CollectAuthenticatedBatches(ctx, l1F, ref, authenticatorAddr, espresso.DefaultBatchAuthLookbackWindow, logger)
		require.NoError(t, err)
		require.Len(t, result, 2)
		require.True(t, result[batchHash])
		require.True(t, result[batchHash2])
		l1F.AssertExpectations(t)
	})
}

// TestCollectAuthenticatedBatchesBlockRefCache verifies that the block ref LRU cache
// eliminates redundant L1BlockRefByHash RPC calls when processing consecutive L1 blocks.
// On the first call (block N), all ~100 L1BlockRefByHash calls are made. On the second
// call (block N+1), the overlapping window means ~99 block refs are already cached,
// so only 1 new L1BlockRefByHash call is needed.
func TestCollectAuthenticatedBatchesBlockRefCache(t *testing.T) {
	resetBatchAuthCaches()
	logger := testlog.Logger(t, log.LevelDebug)
	ctx := context.Background()
	rng := rand.New(rand.NewSource(5678))

	authenticatorAddr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	emptyReceipts := types.Receipts{}

	// Build a chain long enough for two consecutive lookback windows:
	// Block 200's window is [100, 200], block 201's window is [101, 201].
	chain := buildL1Chain(rng, 100, 201)

	// --- First call: block 200, window [100, 200] ---
	// Expects all 101 FetchReceipts calls and 100 L1BlockRefByHash calls (full traversal).
	l1F := &testutils.MockL1Source{}
	for num := uint64(200); num >= 100; num-- {
		ref := chain[num]
		l1F.ExpectFetchReceipts(ref.Hash, nil, emptyReceipts, nil)
		if num > 100 {
			l1F.ExpectL1BlockRefByHash(chain[num-1].Hash, chain[num-1], nil)
		}
	}

	result, err := CollectAuthenticatedBatches(ctx, l1F, chain[200], authenticatorAddr, espresso.DefaultBatchAuthLookbackWindow, logger)
	require.NoError(t, err)
	require.Len(t, result, 0)
	l1F.AssertExpectations(t)

	// --- Second call: block 201, window [101, 201] ---
	// Both receipt and block ref caches are warm for blocks [100, 200].
	// Only block 201 needs FetchReceipts (new block, not in receipt cache).
	// Only block 200 needs L1BlockRefByHash resolution — but it was cached as the
	// `ref` of the previous call (we cache ref.Hash -> ref at the top of the function).
	// So NO L1BlockRefByHash calls should be needed at all.
	l1F2 := &testutils.MockL1Source{}
	// Only block 201's receipts are uncached
	l1F2.ExpectFetchReceipts(chain[201].Hash, nil, emptyReceipts, nil)
	// All block refs in [101, 200] are cached from the first call, and block 200
	// was cached as the ref argument. No L1BlockRefByHash calls expected.

	result2, err := CollectAuthenticatedBatches(ctx, l1F2, chain[201], authenticatorAddr, espresso.DefaultBatchAuthLookbackWindow, logger)
	require.NoError(t, err)
	require.Len(t, result2, 0)
	l1F2.AssertExpectations(t)
}

func TestBatchInfoAuthenticatedABIHash(t *testing.T) {
	// Verify the ABI hash matches what Solidity would compute
	expected := crypto.Keccak256Hash([]byte("BatchInfoAuthenticated(bytes32)"))
	require.Equal(t, expected, BatchInfoAuthenticatedABIHash)
}
