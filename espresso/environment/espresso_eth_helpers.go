package environment

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ErrBalanceDidNotIncrease is a sentinel error that indicates that the balance
// did not increase before the request was cancelled.
var ErrBalanceDidNotIncrease = errors.New("balance did not increase")

// WaitForIncreasedBalance waits for the balance of the given account to
// increase from the given initial balance. It will return nil if the balance
// increases, or an error if the context is cancelled before the balance
// increases.
func WaitForIncreasedBalance(ctx context.Context, client *ethclient.Client, account common.Address, initialBalance *big.Int) error {
	for {
		// Check context to see if we should stop
		select {
		case <-ctx.Done():
			return ErrBalanceDidNotIncrease

		default:
		}

		nextBalance, err := client.BalanceAt(ctx, account, nil)
		if err != nil {
			return err
		}

		if nextBalance.Cmp(initialBalance) > 0 {
			// Our balance has increased
			return nil
		}

		// Sleep for a bit
		time.Sleep(time.Millisecond * 100)
	}
}

func SignTransaction(txData gethTypes.TxData, privateKey *ecdsa.PrivateKey, chainID *big.Int) (*gethTypes.Transaction, error) {
	tx := gethTypes.NewTx(txData)
	signer := opcrypto.PrivateKeySignerFn(privateKey, chainID)
	return signer(crypto.PubkeyToAddress(privateKey.PublicKey), tx)
}
