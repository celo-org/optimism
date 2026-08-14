package batcher_test

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"sync"

	tagged_base64 "github.com/EspressoSystems/espresso-network/sdks/go/tagged-base64"
	"github.com/EspressoSystems/espresso-network/sdks/go/types"
	common "github.com/EspressoSystems/espresso-network/sdks/go/types/common"

	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
)

// ErrNotImplemented is a sentinel error used to indicate that a method
// was not implemented.
var ErrNotImplemented = errors.New("not implemented")

// The fakes below implement batcher.EspressoSubmitClient — the narrow slice of
// the SDK the transaction submitter consumes — rather than the SDK's full
// EspressoClient, so SDK interface growth cannot break them.

// AlwaysFailingEspressoClient returns an error from every method call. This is
// useful for testing error handling in the batcher without relying on a real
// Espresso client.
type AlwaysFailingEspressoClient struct{}

var _ batcher.EspressoSubmitClient = (*AlwaysFailingEspressoClient)(nil)

func (*AlwaysFailingEspressoClient) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	return 0, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) FetchTransactionByHash(ctx context.Context, hash *types.TaggedBase64) (types.TransactionQueryData, error) {
	return types.TransactionQueryData{}, ErrNotImplemented
}

func (*AlwaysFailingEspressoClient) SubmitTransaction(ctx context.Context, tx common.Transaction) (*common.TaggedBase64, error) {
	return nil, ErrNotImplemented
}

// EspressoClientSwappableImplementation is just a proxy to an inner client.
//
// This allows it to be created and swapped easily as needed for testing.
type EspressoClientSwappableImplementation struct {
	sync.RWMutex
	espClient batcher.EspressoSubmitClient
}

var _ batcher.EspressoSubmitClient = (*EspressoClientSwappableImplementation)(nil)

func (c *EspressoClientSwappableImplementation) SetEspressoClient(client batcher.EspressoSubmitClient) {
	c.Lock()
	defer c.Unlock()

	c.espClient = client
}

func (c *EspressoClientSwappableImplementation) FetchLatestBlockHeight(ctx context.Context) (uint64, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchLatestBlockHeight(ctx)
}

func (c *EspressoClientSwappableImplementation) FetchTransactionByHash(ctx context.Context, hash *types.TaggedBase64) (types.TransactionQueryData, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.FetchTransactionByHash(ctx, hash)
}

func (c *EspressoClientSwappableImplementation) SubmitTransaction(ctx context.Context, tx common.Transaction) (*common.TaggedBase64, error) {
	c.RLock()
	defer c.RUnlock()

	return c.espClient.SubmitTransaction(ctx, tx)
}

// FakeSubmissionSucceedingEspressoClient is a mock for the explicit purposes
// of submitting transactions and seeing their response. Methods it does not
// implement itself delegate to the embedded client.
type FakeSubmissionSucceedingEspressoClient struct {
	sync.RWMutex
	batcher.EspressoSubmitClient
	txns map[string]common.Transaction
}

var _ batcher.EspressoSubmitClient = (*FakeSubmissionSucceedingEspressoClient)(nil)

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
