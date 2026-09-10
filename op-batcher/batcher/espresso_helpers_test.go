package batcher_test

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"sync"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	tagged_base64 "github.com/EspressoSystems/espresso-network/sdks/go/tagged-base64"
	"github.com/EspressoSystems/espresso-network/sdks/go/types"
	common "github.com/EspressoSystems/espresso-network/sdks/go/types/common"
)

// ErrNotImplemented is a sentinel error used to indicate that a method
// was not implemented.
var ErrNotImplemented = errors.New("not implemented")

// AlwaysFailingEspressoClient is a mock implementation of the EspressoClient
// interface that always returns an error for every method call. This is
// useful for testing error handling in the batcher without relying on a
// real Espresso client.
type AlwaysFailingEspressoClient struct{}

// Compile time check to ensure adherence to EspressoClient interface
var _ espressoClient.EspressoClient = (*AlwaysFailingEspressoClient)(nil)

func (*AlwaysFailingEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	return 0, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) FetchHeaderByHeight(ctx context.Context, height uint64) (types.HeaderImpl, error) {
	return types.HeaderImpl{}, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) FetchRawHeaderByHeight(ctx context.Context, height uint64) (json.RawMessage, error) {
	return json.RawMessage{}, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) FetchHeadersByRange(ctx context.Context, from uint64, until uint64) ([]types.HeaderImpl, error) {
	return []types.HeaderImpl{}, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) FetchTransactionsInBlock(ctx context.Context, blockHeight uint64, namespace uint64) (espressoClient.TransactionsInBlock, error) {
	return espressoClient.TransactionsInBlock{}, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) FetchTransactionByHash(ctx context.Context, hash *types.TaggedBase64) (types.TransactionQueryData, error) {
	return types.TransactionQueryData{}, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) FetchVidCommonByHeight(ctx context.Context, blockHeight uint64) (types.VidCommon, error) {
	return types.VidCommon{}, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) FetchExplorerTransactionByHash(ctx context.Context, hash *types.TaggedBase64) (types.ExplorerTransactionQueryData, error) {
	return types.ExplorerTransactionQueryData{}, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) StreamPayloads(ctx context.Context, height uint64) (espressoClient.Stream[types.PayloadQueryData], error) {
	return nil, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) StreamTransactions(ctx context.Context, height uint64) (espressoClient.Stream[types.TransactionQueryData], error) {
	return nil, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) StreamTransactionsInNamespace(ctx context.Context, height uint64, namespace uint64) (espressoClient.Stream[types.TransactionQueryData], error) {
	return nil, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) FetchNamespaceTransactionsInRange(ctx context.Context, fromHeight uint64, toHeight uint64, namespace uint64) ([]types.NamespaceTransactionsRangeData, error) {
	return []types.NamespaceTransactionsRangeData{}, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) SubmitTransaction(ctx context.Context, tx common.Transaction) (*common.TaggedBase64, error) {
	return nil, ErrNotImplemented
}

// EspressoClientSwappableImplementation is as implementation of EspressoClient
// that is just a proxy.
//
// This allows it to be created and swapped easily as needed for testing.
type EspressoClientSwappableImplementation struct {
	sync.RWMutex
	espClient espressoClient.EspressoClient
}

// Compile time check to ensure adherence to EspressoClient interface
var _ espressoClient.EspressoClient = (*EspressoClientSwappableImplementation)(nil)

func (c *EspressoClientSwappableImplementation) SetEspressoClient(client espressoClient.EspressoClient) {
	c.Lock()
	defer c.Unlock()

	c.espClient = client
}

func (c *EspressoClientSwappableImplementation) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchLatestBlockHeight(ctx)
}

func (c *EspressoClientSwappableImplementation) FetchHeaderByHeight(ctx context.Context, height uint64) (types.HeaderImpl, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchHeaderByHeight(ctx, height)
}

func (c *EspressoClientSwappableImplementation) FetchRawHeaderByHeight(ctx context.Context, height uint64) (json.RawMessage, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchRawHeaderByHeight(ctx, height)
}

func (c *EspressoClientSwappableImplementation) FetchHeadersByRange(ctx context.Context, from uint64, until uint64) ([]types.HeaderImpl, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchHeadersByRange(ctx, from, until)
}

func (c *EspressoClientSwappableImplementation) FetchTransactionsInBlock(ctx context.Context, blockHeight uint64, namespace uint64) (espressoClient.TransactionsInBlock, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchTransactionsInBlock(ctx, blockHeight, namespace)
}

func (c *EspressoClientSwappableImplementation) FetchTransactionByHash(ctx context.Context, hash *types.TaggedBase64) (types.TransactionQueryData, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchTransactionByHash(ctx, hash)
}

func (c *EspressoClientSwappableImplementation) FetchVidCommonByHeight(ctx context.Context, blockHeight uint64) (types.VidCommon, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchVidCommonByHeight(ctx, blockHeight)
}

func (c *EspressoClientSwappableImplementation) FetchExplorerTransactionByHash(ctx context.Context, hash *types.TaggedBase64) (types.ExplorerTransactionQueryData, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchExplorerTransactionByHash(ctx, hash)
}

func (c *EspressoClientSwappableImplementation) StreamPayloads(ctx context.Context, height uint64) (espressoClient.Stream[types.PayloadQueryData], error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.StreamPayloads(ctx, height)
}

func (c *EspressoClientSwappableImplementation) StreamTransactions(ctx context.Context, height uint64) (espressoClient.Stream[types.TransactionQueryData], error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.StreamTransactions(ctx, height)
}

func (c *EspressoClientSwappableImplementation) StreamTransactionsInNamespace(ctx context.Context, height uint64, namespace uint64) (espressoClient.Stream[types.TransactionQueryData], error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.StreamTransactionsInNamespace(ctx, height, namespace)
}

func (c *EspressoClientSwappableImplementation) FetchNamespaceTransactionsInRange(ctx context.Context, fromHeight uint64, toHeight uint64, namespace uint64) ([]types.NamespaceTransactionsRangeData, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchNamespaceTransactionsInRange(ctx, fromHeight, toHeight, namespace)
}

func (c *EspressoClientSwappableImplementation) SubmitTransaction(ctx context.Context, tx common.Transaction) (*common.TaggedBase64, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.SubmitTransaction(ctx, tx)
}

// FakeSubmissionSucceedingEspressoClient is a mock implementation of the
// EspressoClient for the explicit purposes of submitting transactions, and
// seeing their response.
type FakeSubmissionSucceedingEspressoClient struct {
	sync.RWMutex
	espressoClient.EspressoClient
	txns map[string]common.Transaction
}

// Compile time check to ensure adherence to EspressoClient interface
var _ espressoClient.EspressoClient = (*FakeSubmissionSucceedingEspressoClient)(nil)

var (
	ErrNotInitialized      = errors.New("not initialized")
	ErrHashCannotBeNil     = errors.New("hash cannot be nil")
	ErrTransactionNotFound = errors.New("transaction not found")
)

func (c *FakeSubmissionSucceedingEspressoClient) Init() {
	c.Lock()
	defer c.Unlock()
	c.txns = make(map[string]common.Transaction)
}

// FetchTransactionByHash simulates fetching a transaction by its hash. it
// looks up a transaction for the given hash, and returns the transaction
// if it is found.
func (c *FakeSubmissionSucceedingEspressoClient) FetchTransactionByHash(ctx context.Context, hash *types.TaggedBase64) (types.TransactionQueryData, error) {
	c.RLock()
	defer c.RUnlock()
	if c.txns == nil {
		return types.TransactionQueryData{}, ErrNotInitialized
	}

	if hash == nil {
		return types.TransactionQueryData{}, ErrHashCannotBeNil
	}

	txn, found := c.txns[hash.String()]
	if !found {
		return types.TransactionQueryData{}, ErrTransactionNotFound
	}

	// Just to simulate some processing on the transaction
	height := binary.LittleEndian.Uint64(txn.Payload)

	return types.TransactionQueryData{
		Transaction: txn,
		Hash:        hash,
		Index:       0,
		Proof:       json.RawMessage{},
		BlockHash:   nil,
		BlockHeight: height,
	}, nil
}

// SubmitTransaction simulates a successful transaction submission and stores
// it for future retrieval
func (c *FakeSubmissionSucceedingEspressoClient) SubmitTransaction(ctx context.Context, tx common.Transaction) (*common.TaggedBase64, error) {
	c.Lock()
	defer c.Unlock()
	if c.txns == nil {
		return nil, ErrNotInitialized
	}

	hash, err := tagged_base64.New("TX", tx.Payload)
	if err != nil {
		return nil, err
	}

	c.txns[hash.String()] = tx
	return hash, nil
}
