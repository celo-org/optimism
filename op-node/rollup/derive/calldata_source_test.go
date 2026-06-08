package derive

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

type testTx struct {
	to      *common.Address
	dataLen int
	author  *ecdsa.PrivateKey
	good    bool
	value   int
}

func (tx *testTx) Create(t *testing.T, signer types.Signer, rng *rand.Rand) *types.Transaction {
	t.Helper()
	out, err := types.SignNewTx(tx.author, signer, &types.DynamicFeeTx{
		ChainID:   signer.ChainID(),
		Nonce:     0,
		GasTipCap: big.NewInt(2 * params.GWei),
		GasFeeCap: big.NewInt(30 * params.GWei),
		Gas:       100_000,
		To:        tx.to,
		Value:     big.NewInt(int64(tx.value)),
		Data:      testutils.RandomData(rng, tx.dataLen),
	})
	require.NoError(t, err)
	return out
}

type calldataTest struct {
	name string
	txs  []testTx
}

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
				BatchInfoAuthenticatedABIHash,
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

// TestDataFromEVMTransactionsEventAuth tests event-based batch authentication
// where a BatchInfoAuthenticated event in the lookback window authorizes a batch.
//
// Event-based authentication is only active post-Espresso; the fixture
// activates the fork at L1 origin time 0 (genesis) so all test refs satisfy
// ref.Time >= *EspressoTime.
func TestDataFromEVMTransactionsEventAuth(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	batcherPriv := testutils.RandomKey()
	altAuthor := testutils.RandomKey()
	batchInboxAddr := testutils.RandomAddress(rng)
	authenticatorAddr := testutils.RandomAddress(rng)
	batcherAddr := crypto.PubkeyToAddress(batcherPriv.PublicKey)
	altAuthorAddr := crypto.PubkeyToAddress(altAuthor.PublicKey)
	signer := types.NewCancunSigner(big.NewInt(100))

	espressoTime := uint64(0)
	dsCfg := DataSourceConfig{
		l1Signer:          signer,
		batchInboxAddress: batchInboxAddr,
		rollupCfg: &rollup.Config{
			EspressoTime:              &espressoTime,
			BatchAuthenticatorAddress: authenticatorAddr,
		},
		batchAuthCaches: NewBatchAuthCaches(),
	}

	ctx := context.Background()
	logger := testlog.Logger(t, log.LevelDebug)

	t.Run("authenticated tx accepted", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		txData := testutils.RandomData(rng, 100)
		tx, err := types.SignNewTx(batcherPriv, signer, &types.DynamicFeeTx{
			ChainID: big.NewInt(100), Nonce: 0, Gas: 100_000,
			GasTipCap: big.NewInt(2 * params.GWei), GasFeeCap: big.NewInt(30 * params.GWei),
			To: &batchInboxAddr, Data: txData,
		})
		require.NoError(t, err)

		// Use block number 1 so lookback window is [0, 1] — only 2 blocks to mock
		ref := eth.L1BlockRef{Number: 1, Hash: testutils.RandomHash(rng)}
		batchHash := ComputeCalldataBatchHash(txData)
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, []common.Hash{batchHash})

		out, err := DataFromEVMTransactions(ctx, dsCfg, batcherAddr, types.Transactions{tx}, l1F, ref, logger)
		require.NoError(t, err)
		require.Len(t, out, 1)
		require.Equal(t, eth.Data(txData), out[0])
		l1F.AssertExpectations(t)
	})

	t.Run("unauthenticated tx from unknown sender rejected", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		txData := testutils.RandomData(rng, 100)
		tx, err := types.SignNewTx(altAuthor, signer, &types.DynamicFeeTx{
			ChainID: big.NewInt(100), Nonce: 0, Gas: 100_000,
			GasTipCap: big.NewInt(2 * params.GWei), GasFeeCap: big.NewInt(30 * params.GWei),
			To: &batchInboxAddr, Data: txData,
		})
		require.NoError(t, err)

		ref := eth.L1BlockRef{Number: 1, Hash: testutils.RandomHash(rng)}
		// No auth events — empty authenticated list
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, nil)

		out, err := DataFromEVMTransactions(ctx, dsCfg, batcherAddr, types.Transactions{tx}, l1F, ref, logger)
		require.NoError(t, err)
		require.Len(t, out, 0)
		l1F.AssertExpectations(t)
	})

	t.Run("fallback batcher without auth event rejected", func(t *testing.T) {
		// The fallback batcher now also authenticates via BatchAuthenticator events.
		// Without an auth event, even the SystemConfig batcher address is rejected.
		l1F := &testutils.MockL1Source{}
		txData := testutils.RandomData(rng, 100)
		tx, err := types.SignNewTx(batcherPriv, signer, &types.DynamicFeeTx{
			ChainID: big.NewInt(100), Nonce: 0, Gas: 100_000,
			GasTipCap: big.NewInt(2 * params.GWei), GasFeeCap: big.NewInt(30 * params.GWei),
			To: &batchInboxAddr, Data: txData,
		})
		require.NoError(t, err)

		ref := eth.L1BlockRef{Number: 1, Hash: testutils.RandomHash(rng)}
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, nil)

		out, err := DataFromEVMTransactions(ctx, dsCfg, batcherAddr, types.Transactions{tx}, l1F, ref, logger)
		require.NoError(t, err)
		require.Len(t, out, 0, "fallback batcher without auth event should be rejected")
		l1F.AssertExpectations(t)
	})

	t.Run("wrong inbox address rejected without auth check", func(t *testing.T) {
		// Tx to wrong address should be filtered by isValidBatchTx.
		// CollectAuthenticatedBatches still runs (it's a block-level operation),
		// but no tx passes the inbox address check.
		l1F := &testutils.MockL1Source{}
		wrongAddr := testutils.RandomAddress(rng)
		txData := testutils.RandomData(rng, 100)
		tx, err := types.SignNewTx(batcherPriv, signer, &types.DynamicFeeTx{
			ChainID: big.NewInt(100), Nonce: 0, Gas: 100_000,
			GasTipCap: big.NewInt(2 * params.GWei), GasFeeCap: big.NewInt(30 * params.GWei),
			To: &wrongAddr, Data: txData,
		})
		require.NoError(t, err)

		ref := eth.L1BlockRef{Number: 1, Hash: testutils.RandomHash(rng)}
		// Mock the lookback window scan (returns no authenticated hashes)
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, nil)

		out, err := DataFromEVMTransactions(ctx, dsCfg, batcherAddr, types.Transactions{tx}, l1F, ref, logger)
		require.NoError(t, err)
		require.Len(t, out, 0)
		l1F.AssertExpectations(t)
	})

	t.Run("mixed: only event-authenticated txs accepted", func(t *testing.T) {
		l1F := &testutils.MockL1Source{}
		// tx1: has auth event — should be accepted
		txData1 := testutils.RandomData(rng, 100)
		tx1, err := types.SignNewTx(batcherPriv, signer, &types.DynamicFeeTx{
			ChainID: big.NewInt(100), Nonce: 0, Gas: 100_000,
			GasTipCap: big.NewInt(2 * params.GWei), GasFeeCap: big.NewInt(30 * params.GWei),
			To: &batchInboxAddr, Data: txData1,
		})
		require.NoError(t, err)

		// tx2: no auth event — should be rejected even though sender is batcherAddr
		txData2 := testutils.RandomData(rng, 100)
		tx2, err := types.SignNewTx(batcherPriv, signer, &types.DynamicFeeTx{
			ChainID: big.NewInt(100), Nonce: 1, Gas: 100_000,
			GasTipCap: big.NewInt(2 * params.GWei), GasFeeCap: big.NewInt(30 * params.GWei),
			To: &batchInboxAddr, Data: txData2,
		})
		require.NoError(t, err)

		// tx3: unknown sender without auth event — should be rejected
		txData3 := testutils.RandomData(rng, 100)
		tx3, err := types.SignNewTx(altAuthor, signer, &types.DynamicFeeTx{
			ChainID: big.NewInt(100), Nonce: 2, Gas: 100_000,
			GasTipCap: big.NewInt(2 * params.GWei), GasFeeCap: big.NewInt(30 * params.GWei),
			To: &batchInboxAddr, Data: txData3,
		})
		require.NoError(t, err)

		ref := eth.L1BlockRef{Number: 1, Hash: testutils.RandomHash(rng)}
		batchHash1 := ComputeCalldataBatchHash(txData1)
		// Only tx1 has an auth event (caller = batcherAddr, matching tx1's sender).
		// tx2 and tx3 do not — both should be rejected.
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, []common.Hash{batchHash1})

		out, err := DataFromEVMTransactions(ctx, dsCfg, batcherAddr, types.Transactions{tx1, tx2, tx3}, l1F, ref, logger)
		require.NoError(t, err)
		require.Len(t, out, 1, "only event-authenticated tx should pass")
		require.Equal(t, eth.Data(txData1), out[0])
		l1F.AssertExpectations(t)
	})

	t.Run("auth event accepts a non-batcher sender that matches its caller", func(t *testing.T) {
		// Event-based mode does not require the SystemConfig batcher: any sender is
		// accepted as long as it matches the caller that emitted the auth event.
		// Here altAuthor both submits the batch and is the auth event caller.
		l1F := &testutils.MockL1Source{}
		txData := testutils.RandomData(rng, 100)
		tx, err := types.SignNewTx(altAuthor, signer, &types.DynamicFeeTx{
			ChainID: big.NewInt(100), Nonce: 0, Gas: 100_000,
			GasTipCap: big.NewInt(2 * params.GWei), GasFeeCap: big.NewInt(30 * params.GWei),
			To: &batchInboxAddr, Data: txData,
		})
		require.NoError(t, err)

		ref := eth.L1BlockRef{Number: 1, Hash: testutils.RandomHash(rng)}
		batchHash := ComputeCalldataBatchHash(txData)
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, altAuthorAddr, []common.Hash{batchHash})

		out, err := DataFromEVMTransactions(ctx, dsCfg, batcherAddr, types.Transactions{tx}, l1F, ref, logger)
		require.NoError(t, err)
		require.Len(t, out, 1)
		require.Equal(t, eth.Data(txData), out[0])
		l1F.AssertExpectations(t)
	})

	t.Run("authenticated batch from a different sender than the caller is rejected", func(t *testing.T) {
		// The batch commitment is authenticated, but by batcherAddr; the batch tx is
		// submitted by altAuthor. The sender must match the auth event caller, so the
		// batch is rejected even though the commitment was authenticated.
		l1F := &testutils.MockL1Source{}
		txData := testutils.RandomData(rng, 100)
		tx, err := types.SignNewTx(altAuthor, signer, &types.DynamicFeeTx{
			ChainID: big.NewInt(100), Nonce: 0, Gas: 100_000,
			GasTipCap: big.NewInt(2 * params.GWei), GasFeeCap: big.NewInt(30 * params.GWei),
			To: &batchInboxAddr, Data: txData,
		})
		require.NoError(t, err)

		ref := eth.L1BlockRef{Number: 1, Hash: testutils.RandomHash(rng)}
		batchHash := ComputeCalldataBatchHash(txData)
		ref = mockAuthEvents(l1F, rng, ref, authenticatorAddr, batcherAddr, []common.Hash{batchHash})

		out, err := DataFromEVMTransactions(ctx, dsCfg, batcherAddr, types.Transactions{tx}, l1F, ref, logger)
		require.NoError(t, err)
		require.Len(t, out, 0, "batch authenticated by a different address than the submitter must be rejected")
		l1F.AssertExpectations(t)
	})
}

// TestDataFromEVMTransactions creates some transactions from a specified template and asserts
// that DataFromEVMTransactions properly filters and returns the data from the authorized transactions
// inside the transaction set.
func TestDataFromEVMTransactions(t *testing.T) {
	inboxPriv := testutils.RandomKey()
	batcherPriv := testutils.RandomKey()
	cfg := &rollup.Config{
		L1ChainID:         big.NewInt(100),
		BatchInboxAddress: crypto.PubkeyToAddress(inboxPriv.PublicKey),
	}
	batcherAddr := crypto.PubkeyToAddress(batcherPriv.PublicKey)

	altInbox := testutils.RandomAddress(rand.New(rand.NewSource(1234)))
	altAuthor := testutils.RandomKey()

	testCases := []calldataTest{
		{
			name: "simple",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, author: batcherPriv, good: true}},
		},
		{
			name: "other inbox",
			txs:  []testTx{{to: &altInbox, dataLen: 1234, author: batcherPriv, good: false}}},
		{
			name: "other author",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, author: altAuthor, good: false}}},
		{
			name: "inbox is author",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, author: inboxPriv, good: false}}},
		{
			name: "author is inbox",
			txs:  []testTx{{to: &batcherAddr, dataLen: 1234, author: batcherPriv, good: false}}},
		{
			name: "unrelated",
			txs:  []testTx{{to: &altInbox, dataLen: 1234, author: altAuthor, good: false}}},
		{
			name: "contract creation",
			txs:  []testTx{{to: nil, dataLen: 1234, author: batcherPriv, good: false}}},
		{
			name: "empty tx",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 0, author: batcherPriv, good: true}}},
		{
			name: "value tx",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, value: 42, author: batcherPriv, good: true}}},
		{
			name: "empty block", txs: []testTx{},
		},
		{
			name: "mixed txs",
			txs: []testTx{
				{to: &cfg.BatchInboxAddress, dataLen: 1234, value: 42, author: batcherPriv, good: true},
				{to: &cfg.BatchInboxAddress, dataLen: 3333, value: 32, author: altAuthor, good: false},
				{to: &cfg.BatchInboxAddress, dataLen: 2000, value: 22, author: batcherPriv, good: true},
				{to: &altInbox, dataLen: 2020, value: 12, author: batcherPriv, good: false},
			},
		},
		// TODO: test with different batcher key, i.e. when it's changed from initial config value by L1 contract
	}

	for i, tc := range testCases {
		rng := rand.New(rand.NewSource(int64(i)))
		signer := cfg.L1Signer()

		var expectedData []eth.Data
		var txs []*types.Transaction
		for i, tx := range tc.txs {
			transaction := tx.Create(t, signer, rng)
			txs = append(txs, transaction)

			if tx.good {
				expectedData = append(expectedData, txs[i].Data())
			}
		}

		// Legacy mode (no batch authenticator, Espresso inactive) — uses sender-based auth
		dsCfg := DataSourceConfig{
			l1Signer:          cfg.L1Signer(),
			batchInboxAddress: cfg.BatchInboxAddress,
			rollupCfg:         cfg,
		}
		ref := eth.L1BlockRef{Number: 1}
		// In legacy mode, no L1Fetcher calls are needed for auth (sender check is local)
		out, err := DataFromEVMTransactions(context.Background(), dsCfg, batcherAddr, txs, nil, ref, testlog.Logger(t, log.LevelCrit))
		require.NoError(t, err)
		require.ElementsMatch(t, expectedData, out)
	}
}
