package espresso

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoTaggedBase64 "github.com/EspressoSystems/espresso-network/sdks/go/tagged-base64"
	espressoTypes "github.com/EspressoSystems/espresso-network/sdks/go/types"
)

// MockEspressoClient is an in-memory implementation of the Espresso SDK's
// client.EspressoClient interface used in the e2e tests in place of a real
// dockerized espresso-dev-node.
//
// It models a HotShot chain as an append-only sequence of blocks. Submitted
// transactions are appended to the pending block; a background ticker seals the
// pending block (whether or not it contains transactions) so the block height
// advances continuously, matching how the batcher's verification logic measures
// elapsed blocks. The read path (FetchNamespaceTransactionsInRange /
// FetchTransactionByHash) round-trips the submitted payloads back to the
// streamer, which performs no cryptographic verification of HotShot data.
//
// Only the methods the batcher and streamer use are functional; the remaining
// QueryService methods are stubs.
type MockEspressoClient struct {
	mu sync.Mutex
	// sealed blocks, index = block height. Each block is the list of txs it contains.
	blocks [][]espressoTypes.Transaction
	// transactions still accumulating into the next (unsealed) block.
	pending []espressoTypes.Transaction
	// txByHash maps a submitted transaction's hash string to its stored copy.
	txByHash map[string]storedTx

	blockTime time.Duration
	stop      chan struct{}
	stopped   bool
}

type storedTx struct {
	tx          espressoTypes.Transaction
	hash        *espressoTypes.TaggedBase64
	blockHeight uint64
	index       uint64
}

var _ espressoClient.EspressoClient = (*MockEspressoClient)(nil)

// NewMockEspressoClient creates a mock Espresso client and starts its block
// production ticker. Call Close to stop it. blockTime controls how often an
// (empty or non-empty) block is sealed.
func NewMockEspressoClient(blockTime time.Duration) *MockEspressoClient {
	if blockTime <= 0 {
		blockTime = 250 * time.Millisecond
	}
	c := &MockEspressoClient{
		txByHash:  make(map[string]storedTx),
		blockTime: blockTime,
		stop:      make(chan struct{}),
	}
	// Seal an initial genesis block so the height starts at 1, mirroring a live
	// node that always has at least the genesis block available.
	c.blocks = append(c.blocks, nil)
	go c.produceBlocks()
	return c
}

// Close stops the block-production ticker.
func (c *MockEspressoClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopped {
		return
	}
	c.stopped = true
	close(c.stop)
}

func (c *MockEspressoClient) produceBlocks() {
	ticker := time.NewTicker(c.blockTime)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.sealBlock()
		}
	}
}

// sealBlock moves the pending transactions into a new sealed block, advancing
// the height. Empty blocks are sealed too so the height keeps advancing.
func (c *MockEspressoClient) sealBlock() {
	c.mu.Lock()
	defer c.mu.Unlock()
	height := uint64(len(c.blocks))
	block := c.pending
	c.pending = nil
	c.blocks = append(c.blocks, block)
	for i := range block {
		// index of this tx within the just-sealed block
		st := storedTx{
			tx:          block[i],
			blockHeight: height,
			index:       uint64(i),
		}
		hashStr, err := transactionHashString(block[i])
		if err != nil {
			continue
		}
		if existing, ok := c.txByHash[hashStr]; ok {
			existing.blockHeight = height
			existing.index = uint64(i)
			c.txByHash[hashStr] = existing
			st.hash = existing.hash
		}
	}
}

// transactionHashString derives a stable, opaque hash string for a transaction.
// The exact value is not consensus-meaningful for the mock; it only needs to be
// stable for a given transaction so submit and FetchTransactionByHash agree.
func transactionHashString(tx espressoTypes.Transaction) (string, error) {
	commit := tx.Commit()
	tb, err := espressoTaggedBase64.New("TX", commit[:])
	if err != nil {
		return "", err
	}
	return tb.String(), nil
}

func transactionHash(tx espressoTypes.Transaction) (*espressoTypes.TaggedBase64, error) {
	commit := tx.Commit()
	return espressoTaggedBase64.New("TX", commit[:])
}

// SubmitTransaction appends the transaction to the pending block and returns its hash.
func (c *MockEspressoClient) SubmitTransaction(ctx context.Context, tx espressoTypes.Transaction) (*espressoTypes.TaggedBase64, error) {
	hash, err := transactionHash(tx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", espressoClient.ErrPermanent, err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pending = append(c.pending, tx)
	c.txByHash[hash.String()] = storedTx{tx: tx, hash: hash}
	return hash, nil
}

// FetchLatestBlockHeight returns the current sealed block height.
func (c *MockEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return uint64(len(c.blocks)), nil
}

// FetchNamespaceTransactionsInRange returns the transactions in blocks [from, until)
// matching the requested namespace, one NamespaceTransactionsRangeData per block.
func (c *MockEspressoClient) FetchNamespaceTransactionsInRange(ctx context.Context, from uint64, until uint64, namespace uint64) ([]espressoTypes.NamespaceTransactionsRangeData, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if until <= from {
		return nil, nil
	}
	height := uint64(len(c.blocks))
	if until > height {
		until = height
	}
	res := make([]espressoTypes.NamespaceTransactionsRangeData, 0, until-from)
	for h := from; h < until; h++ {
		var nsTxs []espressoTypes.Transaction
		for _, tx := range c.blocks[h] {
			if tx.Namespace == namespace {
				nsTxs = append(nsTxs, tx)
			}
		}
		res = append(res, espressoTypes.NamespaceTransactionsRangeData{
			Transactions: nsTxs,
		})
	}
	return res, nil
}

// FetchTransactionByHash returns the stored transaction for the given hash, erroring
// if it has not been submitted.
func (c *MockEspressoClient) FetchTransactionByHash(ctx context.Context, hash *espressoTypes.TaggedBase64) (espressoTypes.TransactionQueryData, error) {
	if hash == nil {
		return espressoTypes.TransactionQueryData{}, fmt.Errorf("%w: hash is nil", espressoClient.ErrPermanent)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	st, ok := c.txByHash[hash.String()]
	if !ok {
		return espressoTypes.TransactionQueryData{}, fmt.Errorf("%w: transaction not found", espressoClient.ErrEphemeral)
	}
	return espressoTypes.TransactionQueryData{
		Transaction: st.tx,
		Hash:        st.hash,
		Index:       st.index,
		BlockHeight: st.blockHeight,
	}, nil
}

// ---- Stubbed QueryService methods (not used by the batcher or streamer) ----

func (c *MockEspressoClient) FetchHeaderByHeight(ctx context.Context, height uint64) (espressoTypes.HeaderImpl, error) {
	return espressoTypes.HeaderImpl{}, fmt.Errorf("%w: FetchHeaderByHeight not supported by mock", espressoClient.ErrPermanent)
}

func (c *MockEspressoClient) FetchRawHeaderByHeight(ctx context.Context, height uint64) (json.RawMessage, error) {
	return nil, fmt.Errorf("%w: FetchRawHeaderByHeight not supported by mock", espressoClient.ErrPermanent)
}

func (c *MockEspressoClient) FetchHeadersByRange(ctx context.Context, from uint64, until uint64) ([]espressoTypes.HeaderImpl, error) {
	return nil, fmt.Errorf("%w: FetchHeadersByRange not supported by mock", espressoClient.ErrPermanent)
}

func (c *MockEspressoClient) FetchTransactionsInBlock(ctx context.Context, blockHeight uint64, namespace uint64) (espressoClient.TransactionsInBlock, error) {
	return espressoClient.TransactionsInBlock{}, fmt.Errorf("%w: FetchTransactionsInBlock not supported by mock", espressoClient.ErrPermanent)
}

func (c *MockEspressoClient) FetchVidCommonByHeight(ctx context.Context, blockHeight uint64) (espressoTypes.VidCommon, error) {
	return espressoTypes.VidCommon{}, fmt.Errorf("%w: FetchVidCommonByHeight not supported by mock", espressoClient.ErrPermanent)
}

func (c *MockEspressoClient) FetchExplorerTransactionByHash(ctx context.Context, hash *espressoTypes.TaggedBase64) (espressoTypes.ExplorerTransactionQueryData, error) {
	return espressoTypes.ExplorerTransactionQueryData{}, fmt.Errorf("%w: FetchExplorerTransactionByHash not supported by mock", espressoClient.ErrPermanent)
}

func (c *MockEspressoClient) StreamPayloads(ctx context.Context, height uint64) (espressoClient.Stream[espressoTypes.PayloadQueryData], error) {
	return nil, fmt.Errorf("%w: StreamPayloads not supported by mock", espressoClient.ErrPermanent)
}

func (c *MockEspressoClient) StreamTransactions(ctx context.Context, height uint64) (espressoClient.Stream[espressoTypes.TransactionQueryData], error) {
	return nil, fmt.Errorf("%w: StreamTransactions not supported by mock", espressoClient.ErrPermanent)
}

func (c *MockEspressoClient) StreamTransactionsInNamespace(ctx context.Context, height uint64, namespace uint64) (espressoClient.Stream[espressoTypes.TransactionQueryData], error) {
	return nil, fmt.Errorf("%w: StreamTransactionsInNamespace not supported by mock", espressoClient.ErrPermanent)
}
