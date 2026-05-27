package espresso

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum-optimism/optimism/espresso/bindings"
)

// AdaptL1BlockRefClient is a wrapper around eth.L1BlockRef that implements the espresso.L1Client interface
type AdaptL1BlockRefClient struct {
	L1Client *ethclient.Client
}

// NewAdaptL1BlockRefClient creates a new L1BlockRefClient
func NewAdaptL1BlockRefClient(L1Client *ethclient.Client) *AdaptL1BlockRefClient {
	return &AdaptL1BlockRefClient{
		L1Client: L1Client,
	}
}

// HeaderHashByNumber implements the espresso.L1Client interface
func (c *AdaptL1BlockRefClient) HeaderHashByNumber(ctx context.Context, number *big.Int) (common.Hash, error) {
	expectedL1BlockRef, err := c.L1Client.HeaderByNumber(ctx, number)
	if err != nil {
		return common.Hash{}, err
	}

	return expectedL1BlockRef.Hash(), nil
}

func (c *AdaptL1BlockRefClient) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return c.L1Client.CodeAt(ctx, contract, blockNumber)
}

func (c *AdaptL1BlockRefClient) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return c.L1Client.CallContract(ctx, call, blockNumber)
}

// FetchEspressoBatcherAddress reads the Espresso batcher address from the BatchAuthenticator
// contract on L1. This is used by the caff node to determine which address signed
// Espresso batches, since the Espresso batcher may use a different key than the
// SystemConfig batcher (fallback batcher).
func FetchEspressoBatcherAddress(ctx context.Context, l1Client *ethclient.Client, batchAuthenticatorAddr common.Address) (common.Address, error) {
	caller, err := bindings.NewBatchAuthenticatorCaller(batchAuthenticatorAddr, l1Client)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to bind BatchAuthenticator at %s: %w", batchAuthenticatorAddr, err)
	}
	addr, err := caller.EspressoBatcher(&bind.CallOpts{Context: ctx})
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to call BatchAuthenticator.espressoBatcher(): %w", err)
	}
	return addr, nil
}
