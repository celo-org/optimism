package derive

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// CalldataSource is a fault tolerant approach to fetching data.
// The constructor will never fail & it will instead re-attempt the fetcher
// at a later point.
type CalldataSource struct {
	// Internal state + data
	open bool
	data []eth.Data
	// Required to re-attempt fetching
	ref     eth.L1BlockRef
	dsCfg   DataSourceConfig
	fetcher L1Fetcher
	log     log.Logger

	batcherAddr common.Address
}

// NewCalldataSource creates a new calldata source. It suppresses errors in fetching the L1 block if they occur.
// If there is an error, it will attempt to fetch the result on the next call to `Next`.
func NewCalldataSource(ctx context.Context, log log.Logger, dsCfg DataSourceConfig, fetcher L1Fetcher, ref eth.L1BlockRef, batcherAddr common.Address) DataIter {
	closedSource := &CalldataSource{
		open:        false,
		ref:         ref,
		dsCfg:       dsCfg,
		fetcher:     fetcher,
		log:         log,
		batcherAddr: batcherAddr,
	}

	_, txs, err := fetcher.InfoAndTxsByHash(ctx, ref.Hash)
	if err != nil {
		return closedSource
	}
	data, err := DataFromEVMTransactions(ctx, dsCfg, batcherAddr, txs, fetcher, ref, log.New("origin", ref))
	if err != nil {
		return closedSource
	}
	return &CalldataSource{
		open: true,
		data: data,
	}
}

// Next returns the next piece of data if it has it. If the constructor failed, this
// will attempt to reinitialize itself. If it cannot find the block it returns a ResetError
// otherwise it returns a temporary error if fetching the block returns an error.
func (ds *CalldataSource) Next(ctx context.Context) (eth.Data, error) {
	if !ds.open {
		_, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.ref.Hash)
		if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to open calldata source: %w", err))
		} else if err != nil {
			return nil, NewTemporaryError(fmt.Errorf("failed to open calldata source: %w", err))
		}
		ds.data, err = DataFromEVMTransactions(ctx, ds.dsCfg, ds.batcherAddr, txs, ds.fetcher, ds.ref, ds.log)
		if err != nil {
			return nil, err
		}
		ds.open = true
	}
	if len(ds.data) == 0 {
		return nil, io.EOF
	} else {
		data := ds.data[0]
		ds.data = ds.data[1:]
		return data, nil
	}
}

// DataFromEVMTransactions filters all of the transactions and returns the calldata from transactions
// that are sent to the batch inbox address from the batch sender address.
// This will return an empty array if no valid transactions are found.
//
// Pre-Espresso (the L1 origin time of `ref` is < *EspressoTime, or unset),
// this runs upstream Optimism semantics: filter by batch inbox + sender ==
// batcher.
//
// Post-Espresso, it collects all authenticated batch hashes from a lookback
// window once and rejects any batch whose commitment hash is not in the
// authenticated set.
func DataFromEVMTransactions(ctx context.Context, dsCfg DataSourceConfig, batcherAddr common.Address, txs types.Transactions, fetcher L1Fetcher, ref eth.L1BlockRef, log log.Logger) ([]eth.Data, error) {
	// Only collect authenticated batch commitments when the Espresso fork is
	// active at the L1 origin time of the block we're scanning. Pre-fork, the
	// upstream sender-based authorization path inside isBatchTxAuthorized is used
	// and the authenticatedHashes map is unused.
	var authenticatedHashes map[common.Hash]common.Address
	if dsCfg.isEspresso(ref.Time) {
		var err error
		authenticatedHashes, err = CollectAuthenticatedBatches(
			ctx, fetcher, ref, dsCfg.batchAuthenticatorAddress, dsCfg.batchAuthLookbackWindow, log,
		)
		if err != nil {
			return nil, err
		}
	}

	out := []eth.Data{}
	for _, tx := range txs {
		if !isValidBatchTx(tx, dsCfg.batchInboxAddress, log) {
			continue
		}

		batchHash := ComputeCalldataBatchHash(tx.Data())
		if !isBatchTxAuthorized(tx, dsCfg, batcherAddr, batchHash, authenticatedHashes, ref.Time, log) {
			continue
		}

		out = append(out, tx.Data())
	}
	return out, nil
}
