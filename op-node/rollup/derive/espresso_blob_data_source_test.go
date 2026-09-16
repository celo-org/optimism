package derive

import (
	"context"
	"crypto/ecdsa"
	"io"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
	"github.com/ethereum/go-ethereum/log"
)

// mockAuthEvents sets up L1 mock expectations for CollectAuthenticatedBatches to find auth events
// for the given batch hashes at the given ref's block number. Auth events for batch hashes in
// `authenticated` are placed in the ref block's receipts; all other blocks in the lookback
// window have empty receipts.
//
// CollectAuthenticatedBatches traverses backward from ref via parent hashes, so this helper
// builds a chain of L1BlockRef values with proper parent-hash linkage, sets up FetchReceipts
// for each block, and L1BlockRefByHash for each parent.
//
// The auth events are emitted with `caller` as the indexed caller, which the
// pipeline matches against the batch transaction's L1 sender. Tests pass the
// expected batcher address here.
//
// Returns the updated ref with its ParentHash properly set to the chain. Callers must use
// the returned ref when calling functions that invoke CollectAuthenticatedBatches.
func mockAuthEvents(l1F *testutils.MockL1Source, rng *rand.Rand, ref eth.L1BlockRef, authenticatorAddr, caller common.Address, authenticated []common.Hash) eth.L1BlockRef {
	startBlock := ref.Number
	if startBlock > BatchAuthLookbackWindow {
		startBlock = ref.Number - BatchAuthLookbackWindow
	} else {
		startBlock = 0
	}
	windowSize := ref.Number - startBlock + 1

	// Build the auth receipts for the ref block. The commitment is the unindexed
	// data argument; only the caller is indexed (Topics[1]).
	var authLogs []*types.Log
	for _, bh := range authenticated {
		authLogs = append(authLogs, &types.Log{
			Address: authenticatorAddr,
			Topics: []common.Hash{
				batchInfoAuthenticatedTopic,
				common.BytesToHash(caller.Bytes()),
			},
			Data: bh.Bytes(),
		})
	}
	authReceipts := types.Receipts{}
	if len(authLogs) > 0 {
		authReceipts = types.Receipts{{Status: types.ReceiptStatusSuccessful, Logs: authLogs}}
	}

	// Build parent-hash-linked chain from startBlock to ref.Number.
	// chain[i] corresponds to block number startBlock + i.
	chain := make([]eth.L1BlockRef, windowSize)
	for i := uint64(0); i < windowSize; i++ {
		blockNum := startBlock + i
		if blockNum == ref.Number {
			chain[i] = ref
		} else {
			chain[i] = eth.L1BlockRef{Number: blockNum, Hash: testutils.RandomHash(rng)}
		}
		if i > 0 {
			chain[i].ParentHash = chain[i-1].Hash
		}
	}

	// Update the ref at the end of the chain with the correct ParentHash
	updatedRef := chain[windowSize-1]

	// Set up expectations for backward traversal: ref -> ref-1 -> ... -> startBlock
	for i := int(windowSize) - 1; i >= 0; i-- {
		blockRef := chain[i]
		if blockRef.Number == ref.Number {
			l1F.ExpectFetchReceipts(blockRef.Hash, nil, authReceipts, nil)
		} else {
			l1F.ExpectFetchReceipts(blockRef.Hash, nil, types.Receipts{}, nil)
		}
		// L1BlockRefByHash is called for every parent except when we've reached the end of the window
		if i > 0 {
			l1F.ExpectL1BlockRefByHash(chain[i-1].Hash, chain[i-1], nil)
		}
	}

	return updatedRef
}

// TestDataAndHashesFromTxsEventAuth tests event-based batch authentication in the blob
// data source path, and that blob-carrying inbox transactions are dropped post-fork
// regardless of authentication (calldata-only DA).
//
// Event-based authentication is only enforced once the fork has been active for
// BatchAuthEnforcementDelaySecs; the fixture activates the fork at L1 origin time 0
// (genesis) and all test refs use Time = BatchAuthEnforcementDelaySecs — the first
// enforced timestamp.
func TestDataAndHashesFromTxsEventAuth(t *testing.T) {
	rng := rand.New(rand.NewSource(9999))
	privateKey := testutils.InsecureRandomKey(rng)
	altKey := testutils.InsecureRandomKey(rng)
	batcherAddr := crypto.PubkeyToAddress(*privateKey.Public().(*ecdsa.PublicKey))
	altAddr := crypto.PubkeyToAddress(*altKey.Public().(*ecdsa.PublicKey))
	batchInboxAddr := testutils.RandomAddress(rng)
	authenticatorAddr := testutils.RandomAddress(rng)
	logger := testlog.Logger(t, log.LvlInfo)

	chainId := new(big.Int).SetUint64(rng.Uint64())
	signer := types.NewPragueSigner(chainId)
	espressoTime := uint64(0)
	config := DataSourceConfig{
		l1Signer:          signer,
		batchInboxAddress: batchInboxAddr,
		rollupCfg: &rollup.Config{
			EspressoTime:              &espressoTime,
			BatchAuthenticatorAddress: authenticatorAddr,
		},
		batchAuthCaches: NewBatchAuthCaches(),
	}

	ctx := context.Background()

	t.Run("authenticated calldata tx accepted", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		txData := &types.LegacyTx{
			Nonce:    rng.Uint64(),
			GasPrice: new(big.Int).SetUint64(rng.Uint64()),
			Gas:      2_000_000,
			To:       &batchInboxAddr,
			Value:    big.NewInt(10),
			Data:     testutils.RandomData(rng, 200),
		}
		calldataTx, _ := types.SignNewTx(privateKey, signer, txData)

		ref := eth.L1BlockRef{Number: 1, Time: BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
		batchHash := ComputeCalldataBatchHash(calldataTx.Data())
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, []common.Hash{batchHash})

		data, blobHashes, err := dataAndHashesFromTxs(ctx, types.Transactions{calldataTx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 1, len(data))
		require.Equal(t, 0, len(blobHashes))
		require.NotNil(t, data[0].calldata)
		require.Equal(t, eth.Data(calldataTx.Data()), *data[0].calldata)
		l1F.AssertExpectations(t)
	})

	t.Run("authenticated blob tx rejected: blob DA unsupported post-fork", func(t *testing.T) {
		// Even a fully event-authenticated blob batch must be dropped post-fork:
		// batch data is calldata-only because the Celo fault-proof
		// host cannot supply blob preimages.
		l1F := &testutils.MockL1Source{}
		blobHash := testutils.RandomHash(rng)
		blobTxData := &types.BlobTx{
			Nonce:      rng.Uint64(),
			Gas:        2_000_000,
			To:         batchInboxAddr,
			Data:       testutils.RandomData(rng, 100),
			BlobHashes: []common.Hash{blobHash},
		}
		blobTx, _ := types.SignNewTx(privateKey, signer, blobTxData)

		ref := eth.L1BlockRef{Number: 1, Time: BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
		batchHash := ComputeBlobBatchHash([]common.Hash{blobHash})
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, []common.Hash{batchHash})

		data, blobHashes, err := dataAndHashesFromTxs(ctx, types.Transactions{blobTx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 0, len(data), "authenticated blob batch must be dropped post-fork")
		require.Equal(t, 0, len(blobHashes), "no blob preimages may be requested post-fork")
		l1F.AssertExpectations(t)
	})

	t.Run("unknown sender rejected without auth event", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		txData := &types.LegacyTx{
			Nonce:    rng.Uint64(),
			GasPrice: new(big.Int).SetUint64(rng.Uint64()),
			Gas:      2_000_000,
			To:       &batchInboxAddr,
			Value:    big.NewInt(10),
			Data:     testutils.RandomData(rng, 200),
		}
		// Signed by an unknown key (not batcherAddr), no auth event — should be rejected
		calldataTx, _ := types.SignNewTx(altKey, signer, txData)

		ref := eth.L1BlockRef{Number: 1, Time: BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, nil) // no auth events

		data, blobHashes, err := dataAndHashesFromTxs(ctx, types.Transactions{calldataTx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 0, len(data), "unknown sender tx without auth event should be rejected")
		require.Equal(t, 0, len(blobHashes))
		l1F.AssertExpectations(t)
	})

	t.Run("fallback batcher without auth event rejected", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		txData := &types.LegacyTx{
			Nonce:    rng.Uint64(),
			GasPrice: new(big.Int).SetUint64(rng.Uint64()),
			Gas:      2_000_000,
			To:       &batchInboxAddr,
			Value:    big.NewInt(10),
			Data:     testutils.RandomData(rng, 200),
		}
		// Signed by batcher key (SystemConfig batcherAddr), no auth event — should be rejected
		// because all batchers now require event-based authentication
		calldataTx, _ := types.SignNewTx(privateKey, signer, txData)

		ref := eth.L1BlockRef{Number: 1, Time: BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, nil) // no auth events

		data, blobHashes, err := dataAndHashesFromTxs(ctx, types.Transactions{calldataTx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 0, len(data), "fallback batcher without auth event should be rejected")
		require.Equal(t, 0, len(blobHashes))
		l1F.AssertExpectations(t)
	})

	t.Run("non-batcher sender accepted when it matches the auth caller", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		txData := &types.LegacyTx{
			Nonce:    rng.Uint64(),
			GasPrice: new(big.Int).SetUint64(rng.Uint64()),
			Gas:      2_000_000,
			To:       &batchInboxAddr,
			Value:    big.NewInt(10),
			Data:     testutils.RandomData(rng, 200),
		}
		// Signed by alt key (not the SystemConfig batcher), and the auth event was
		// emitted by that same alt address — should be accepted.
		calldataTx, _ := types.SignNewTx(altKey, signer, txData)

		ref := eth.L1BlockRef{Number: 1, Time: BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
		batchHash := ComputeCalldataBatchHash(calldataTx.Data())
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, altAddr, []common.Hash{batchHash})

		data, blobHashes, err := dataAndHashesFromTxs(ctx, types.Transactions{calldataTx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 1, len(data))
		require.Equal(t, 0, len(blobHashes))
		require.NotNil(t, data[0].calldata)
		require.Equal(t, eth.Data(calldataTx.Data()), *data[0].calldata) // the authenticated tx, not just any
		l1F.AssertExpectations(t)
	})

	t.Run("authenticated tx rejected when sender differs from auth caller", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		txData := &types.LegacyTx{
			Nonce:    rng.Uint64(),
			GasPrice: new(big.Int).SetUint64(rng.Uint64()),
			Gas:      2_000_000,
			To:       &batchInboxAddr,
			Value:    big.NewInt(10),
			Data:     testutils.RandomData(rng, 200),
		}
		// Signed by alt key, but the commitment was authenticated by batcherAddr.
		// The submitter must match the auth caller — should be rejected.
		calldataTx, _ := types.SignNewTx(altKey, signer, txData)

		ref := eth.L1BlockRef{Number: 1, Time: BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
		batchHash := ComputeCalldataBatchHash(calldataTx.Data())
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, []common.Hash{batchHash})

		data, blobHashes, err := dataAndHashesFromTxs(ctx, types.Transactions{calldataTx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 0, len(data), "batch authenticated by a different address than the submitter must be rejected")
		require.Equal(t, 0, len(blobHashes))
		l1F.AssertExpectations(t)
	})

	t.Run("mixed block: only the event-authenticated tx accepted", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		newInboxTx := func(key *ecdsa.PrivateKey) *types.Transaction {
			tx, err := types.SignNewTx(key, signer, &types.LegacyTx{
				Nonce:    rng.Uint64(),
				GasPrice: new(big.Int).SetUint64(rng.Uint64()),
				Gas:      2_000_000,
				To:       &batchInboxAddr,
				Value:    big.NewInt(10),
				Data:     testutils.RandomData(rng, 200),
			})
			require.NoError(t, err)
			return tx
		}
		// tx1: batcher-signed, commitment authenticated — accepted.
		// tx2: batcher-signed, commitment NOT authenticated — rejected even though its
		// sender authenticated tx1's commitment (each tx must be matched against its
		// own commitment, not just any commitment from an authenticated caller).
		// tx3: unknown sender, not authenticated — rejected.
		tx1 := newInboxTx(privateKey)
		tx2 := newInboxTx(privateKey)
		tx3 := newInboxTx(altKey)

		ref := eth.L1BlockRef{Number: 1, Time: BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr,
			[]common.Hash{ComputeCalldataBatchHash(tx1.Data())})

		data, blobHashes, err := dataAndHashesFromTxs(ctx, types.Transactions{tx1, tx2, tx3}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 1, len(data), "only the event-authenticated tx should pass")
		require.Equal(t, 0, len(blobHashes))
		require.NotNil(t, data[0].calldata)
		require.Equal(t, eth.Data(tx1.Data()), *data[0].calldata)
		require.NotEqual(t, eth.Data(tx2.Data()), *data[0].calldata)
		l1F.AssertExpectations(t)
	})

	t.Run("mixed calldata+blob block: only the calldata batch accepted", func(t *testing.T) {
		// One calldata batch and one blob batch in the same block, each authenticated
		// via its own commitment (calldata hash vs blob-hash concatenation). Post-fork
		// only the calldata batch may pass; the blob batch is dropped despite its
		// valid authentication (calldata-only DA).
		//
		// The blob tx goes first, ahead of the calldata batch. That ordering is what makes
		// the drop's per-transaction scope observable: a whole-block short-circuit (break
		// where the gate has continue) would swallow the calldata batch behind it. With the
		// calldata batch first, this test passes either way.
		l1F := &testutils.MockL1Source{}
		calldataTx, _ := types.SignNewTx(privateKey, signer, &types.LegacyTx{
			Nonce:    rng.Uint64(),
			GasPrice: new(big.Int).SetUint64(rng.Uint64()),
			Gas:      2_000_000,
			To:       &batchInboxAddr,
			Value:    big.NewInt(10),
			Data:     testutils.RandomData(rng, 200),
		})
		blobHash := testutils.RandomHash(rng)
		blobTx, _ := types.SignNewTx(privateKey, signer, &types.BlobTx{
			Nonce:      rng.Uint64(),
			Gas:        2_000_000,
			To:         batchInboxAddr,
			BlobHashes: []common.Hash{blobHash},
		})

		ref := eth.L1BlockRef{Number: 1, Time: BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, []common.Hash{
			ComputeCalldataBatchHash(calldataTx.Data()),
			ComputeBlobBatchHash([]common.Hash{blobHash}),
		})

		data, blobHashes, err := dataAndHashesFromTxs(ctx, types.Transactions{blobTx, calldataTx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 1, len(data), "only the calldata batch may pass post-fork")
		require.Equal(t, 0, len(blobHashes))
		require.NotNil(t, data[0].calldata)
		require.Equal(t, eth.Data(calldataTx.Data()), *data[0].calldata)
		l1F.AssertExpectations(t)
	})
}

// TestDataAndHashesFromTxsForkBoundary exercises the Espresso fork gate flipping in the
// blob data source path (dataAndHashesFromTxs) across a single fixed DataSourceConfig.
//
// This is the path a chain with Ecotone active actually runs: OpenData always selects the
// blob source, and calldata (type-2) batches flow through its non-blob branch. It walks the
// gate across the fork boundary: sender-based authorization (no event scanning) both pre-fork
// AND through the grace window [EspressoTime, EspressoTime+BatchAuthEnforcementDelaySecs), then
// event-based authentication from EspressoTime+BatchAuthEnforcementDelaySecs onward. See
// BatchAuthEnforcementDelaySecs (params.go) for the full grace-period mechanism.
func TestDataAndHashesFromTxsForkBoundary(t *testing.T) {
	rng := rand.New(rand.NewSource(7777))
	privateKey := testutils.InsecureRandomKey(rng)
	altKey := testutils.InsecureRandomKey(rng)
	batcherAddr := crypto.PubkeyToAddress(*privateKey.Public().(*ecdsa.PublicKey))
	batchInboxAddr := testutils.RandomAddress(rng)
	authenticatorAddr := testutils.RandomAddress(rng)
	logger := testlog.Logger(t, log.LvlInfo)

	chainId := new(big.Int).SetUint64(rng.Uint64())
	signer := types.NewPragueSigner(chainId)

	// Fork activates at L1 origin time 1000. A single config is reused across all
	// sub-tests; only ref.Time changes to cross the boundary.
	espressoTime := uint64(1000)
	config := DataSourceConfig{
		l1Signer:          signer,
		batchInboxAddress: batchInboxAddr,
		rollupCfg: &rollup.Config{
			EspressoTime:              &espressoTime,
			BatchAuthenticatorAddress: authenticatorAddr,
		},
		batchAuthCaches: NewBatchAuthCaches(),
	}

	ctx := context.Background()

	// newCalldataBatchTx builds a type-2 calldata batch tx to the inbox (the tx shape an
	// Ecotone-active, calldata-batching chain submits through the blob source).
	newCalldataBatchTx := func(t *testing.T, author *ecdsa.PrivateKey, data []byte) *types.Transaction {
		t.Helper()
		tx, err := types.SignNewTx(author, signer, &types.DynamicFeeTx{
			ChainID: chainId, Nonce: rng.Uint64(), Gas: 2_000_000,
			GasTipCap: big.NewInt(2 * params.GWei), GasFeeCap: big.NewInt(30 * params.GWei),
			To: &batchInboxAddr, Data: data,
		})
		require.NoError(t, err)
		return tx
	}

	t.Run("pre-fork: batcher accepted via sender auth, no event scan", func(t *testing.T) {
		// The empty mock asserts pre-fork derivation performs zero L1 receipt scanning:
		// any FetchReceipts/L1BlockRefByHash call would be an unexpected call and panic.
		l1F := &testutils.MockL1Source{}
		txData := testutils.RandomData(rng, 200)
		tx := newCalldataBatchTx(t, privateKey, txData)

		ref := eth.L1BlockRef{Number: 1, Time: espressoTime - 1, Hash: testutils.RandomHash(rng)}
		data, hashes, err := dataAndHashesFromTxs(ctx, types.Transactions{tx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 1, len(data), "pre-fork batcher tx should be accepted via sender-based auth")
		require.Equal(t, 0, len(hashes))
		require.NotNil(t, data[0].calldata)
		require.Equal(t, eth.Data(txData), *data[0].calldata)
		l1F.AssertExpectations(t)
	})

	t.Run("pre-fork: non-batcher sender rejected", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		tx := newCalldataBatchTx(t, altKey, testutils.RandomData(rng, 200))

		ref := eth.L1BlockRef{Number: 1, Time: espressoTime - 1, Hash: testutils.RandomHash(rng)}
		data, hashes, err := dataAndHashesFromTxs(ctx, types.Transactions{tx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 0, len(data), "pre-fork tx from a non-batcher sender should be rejected")
		require.Equal(t, 0, len(hashes))
		l1F.AssertExpectations(t)
	})

	t.Run("grace window: batcher accepted via sender auth, no event scan", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		txData := testutils.RandomData(rng, 200)
		tx := newCalldataBatchTx(t, privateKey, txData)

		for _, refTime := range []uint64{espressoTime, espressoTime + BatchAuthEnforcementDelaySecs - 1} {
			ref := eth.L1BlockRef{Number: 1, Time: refTime, Hash: testutils.RandomHash(rng)}
			data, hashes, err := dataAndHashesFromTxs(ctx, types.Transactions{tx}, &config, batcherAddr, l1F, ref, logger)
			require.NoError(t, err)
			require.Equal(t, 1, len(data), "grace-window batcher tx should be accepted via sender-based auth")
			require.Equal(t, 0, len(hashes))
			require.Equal(t, eth.Data(txData), *data[0].calldata)
		}
		l1F.AssertExpectations(t)
	})

	t.Run("enforcement time: same batcher tx rejected without auth event", func(t *testing.T) {
		// At ref.Time == EspressoTime+BatchAuthEnforcementDelaySecs the event-based path is
		// enforced, so a sender-only batcher tx is no longer sufficient.
		l1F := &testutils.MockL1Source{}
		txData := testutils.RandomData(rng, 200)
		tx := newCalldataBatchTx(t, privateKey, txData)

		ref := eth.L1BlockRef{Number: 1, Time: espressoTime + BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, nil)

		data, hashes, err := dataAndHashesFromTxs(ctx, types.Transactions{tx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 0, len(data), "enforced batcher tx without an auth event must be rejected")
		require.Equal(t, 0, len(hashes))
		l1F.AssertExpectations(t)
	})

	t.Run("enforcement time: same batcher tx accepted with auth event", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		txData := testutils.RandomData(rng, 200)
		tx := newCalldataBatchTx(t, privateKey, txData)

		ref := eth.L1BlockRef{Number: 1, Time: espressoTime + BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
		batchHash := ComputeCalldataBatchHash(tx.Data())
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, []common.Hash{batchHash})

		data, hashes, err := dataAndHashesFromTxs(ctx, types.Transactions{tx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 1, len(data), "post-fork batcher tx with a matching auth event must be accepted")
		require.Equal(t, 0, len(hashes))
		require.NotNil(t, data[0].calldata)
		require.Equal(t, eth.Data(txData), *data[0].calldata)
		l1F.AssertExpectations(t)
	})

	// newBlobBatchTx builds a blob batch tx to the inbox, signed by the batcher key.
	newBlobBatchTx := func(t *testing.T, blobHash common.Hash) *types.Transaction {
		t.Helper()
		tx, err := types.SignNewTx(privateKey, signer, &types.BlobTx{
			Nonce:      rng.Uint64(),
			Gas:        2_000_000,
			To:         batchInboxAddr,
			BlobHashes: []common.Hash{blobHash},
		})
		require.NoError(t, err)
		return tx
	}

	t.Run("pre-fork: blob batcher tx accepted via sender auth", func(t *testing.T) {
		// Pre-fork the pipeline keeps upstream semantics: blob batches from the
		// batcher are accepted and their versioned hashes requested.
		l1F := &testutils.MockL1Source{}
		blobHash := testutils.RandomHash(rng)
		tx := newBlobBatchTx(t, blobHash)

		ref := eth.L1BlockRef{Number: 1, Time: espressoTime - 1, Hash: testutils.RandomHash(rng)}
		data, hashes, err := dataAndHashesFromTxs(ctx, types.Transactions{tx}, &config, batcherAddr, l1F, ref, logger)
		require.NoError(t, err)
		require.Equal(t, 1, len(data), "pre-fork blob batcher tx should be accepted via sender-based auth")
		require.Equal(t, 1, len(hashes))
		require.Equal(t, blobHash, hashes[0])
		l1F.AssertExpectations(t)
	})

	t.Run("post-fork: blob batcher tx dropped from activation onward", func(t *testing.T) {
		// From the moment the fork activates — including the grace window, where
		// calldata batches still pass on sender auth — blob batches are dropped
		// (calldata-only DA). The empty mock also asserts no receipt
		// scanning happens pre-enforcement.
		l1F := &testutils.MockL1Source{}
		tx := newBlobBatchTx(t, testutils.RandomHash(rng))

		for _, refTime := range []uint64{espressoTime, espressoTime + BatchAuthEnforcementDelaySecs - 1} {
			ref := eth.L1BlockRef{Number: 1, Time: refTime, Hash: testutils.RandomHash(rng)}
			data, hashes, err := dataAndHashesFromTxs(ctx, types.Transactions{tx}, &config, batcherAddr, l1F, ref, logger)
			require.NoError(t, err)
			require.Equal(t, 0, len(data), "post-fork blob batcher tx must be dropped")
			require.Equal(t, 0, len(hashes), "no blob preimages may be requested post-fork")
		}
		l1F.AssertExpectations(t)
	})
}

// TestOpenDataDropsBlobsPostEspresso drives the whole data-source path — NewDataSourceFactory,
// OpenData, Next — rather than calling dataAndHashesFromTxs directly, and asserts that a blob
// inbox tx costs no blob fetch once Espresso is active. The blobs fetcher has no expectations
// set, so any GetBlobsByHash call is unexpected and fails the test; that is what pins the
// no-blobs-to-fetch short-circuit in BlobDataSource.open.
//
// The ref sits past the enforcement delay with a non-zero espresso_time, so the auth event
// scan runs and the drop is exercised on the enforced path rather than the grace window.
func TestOpenDataDropsBlobsPostEspresso(t *testing.T) {
	rng := rand.New(rand.NewSource(5555))
	privateKey := testutils.InsecureRandomKey(rng)
	batcherAddr := crypto.PubkeyToAddress(*privateKey.Public().(*ecdsa.PublicKey))
	batchInboxAddr := testutils.RandomAddress(rng)
	authenticatorAddr := testutils.RandomAddress(rng)
	logger := testlog.Logger(t, log.LvlInfo)

	chainId := new(big.Int).SetUint64(rng.Uint64())
	signer := types.NewPragueSigner(chainId)

	// Ecotone at genesis so OpenData selects the blob source at all; Espresso at 1000.
	ecotoneTime := uint64(0)
	espressoTime := uint64(1000)
	cfg := &rollup.Config{
		L1ChainID:                 chainId,
		BatchInboxAddress:         batchInboxAddr,
		EcotoneTime:               &ecotoneTime,
		EspressoTime:              &espressoTime,
		BatchAuthenticatorAddress: authenticatorAddr,
	}

	blobTx, err := types.SignNewTx(privateKey, signer, &types.BlobTx{
		Nonce:      rng.Uint64(),
		Gas:        2_000_000,
		To:         batchInboxAddr,
		BlobHashes: []common.Hash{testutils.RandomHash(rng)},
	})
	require.NoError(t, err)

	l1F := &testutils.MockL1Source{}
	blobsFetcher := &testutils.MockBlobsFetcher{}

	ref := eth.L1BlockRef{Number: 1, Time: espressoTime + BatchAuthEnforcementDelaySecs, Hash: testutils.RandomHash(rng)}
	ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, nil)
	l1F.ExpectInfoAndTxsByHash(ref.Hash, testutils.RandomBlockInfo(rng), types.Transactions{blobTx}, nil)

	ctx := context.Background()
	src, err := NewDataSourceFactory(logger, cfg, l1F, blobsFetcher, nil).OpenData(ctx, ref, batcherAddr)
	require.NoError(t, err)

	_, err = src.Next(ctx)
	require.ErrorIs(t, err, io.EOF, "the blob batch must be dropped, leaving no data to return")

	l1F.AssertExpectations(t)
	blobsFetcher.AssertExpectations(t)
}
