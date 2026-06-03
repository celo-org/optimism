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

type blobOrCalldata struct {
	// union type. exactly one of calldata or blob should be non-nil
	blob     *eth.Blob
	calldata *eth.Data
}

// BlobDataSource fetches blobs or calldata as appropriate and transforms them into usable rollup
// data.
type BlobDataSource struct {
	data         []blobOrCalldata
	ref          eth.L1BlockRef
	batcherAddr  common.Address
	dsCfg        DataSourceConfig
	fetcher      L1Fetcher
	blobsFetcher L1BlobsFetcher
	log          log.Logger
}

// NewBlobDataSource creates a new blob data source.
func NewBlobDataSource(ctx context.Context, log log.Logger, dsCfg DataSourceConfig, fetcher L1Fetcher, blobsFetcher L1BlobsFetcher, ref eth.L1BlockRef, batcherAddr common.Address) DataIter {
	return &BlobDataSource{
		ref:          ref,
		dsCfg:        dsCfg,
		fetcher:      fetcher,
		log:          log.New("origin", ref),
		batcherAddr:  batcherAddr,
		blobsFetcher: blobsFetcher,
	}
}

// Next returns the next piece of batcher data, or an io.EOF error if no data remains. It returns
// ResetError if it cannot find the referenced block or a referenced blob, or TemporaryError for
// any other failure to fetch a block or blob.
func (ds *BlobDataSource) Next(ctx context.Context) (eth.Data, error) {
	if ds.data == nil {
		var err error
		if ds.data, err = ds.open(ctx); err != nil {
			return nil, err
		}
	}

	if len(ds.data) == 0 {
		return nil, io.EOF
	}

	next := ds.data[0]
	ds.data = ds.data[1:]
	if next.calldata != nil {
		return *next.calldata, nil
	}

	data, err := next.blob.ToData()
	if err != nil {
		ds.log.Error("ignoring blob due to parse failure", "err", err)
		return ds.Next(ctx)
	}
	return data, nil
}

// open fetches and returns the blob or calldata (as appropriate) from all valid batcher
// transactions in the referenced block. Returns an empty (non-nil) array if no batcher
// transactions are found. It returns ResetError if it cannot find the referenced block or a
// referenced blob, or TemporaryError for any other failure to fetch a block or blob.
func (ds *BlobDataSource) open(ctx context.Context) ([]blobOrCalldata, error) {
	_, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.ref.Hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to open blob data source: %w", err))
		}
		return nil, NewTemporaryError(fmt.Errorf("failed to open blob data source: %w", err))
	}

	data, hashes, err := dataAndHashesFromTxs(ctx, txs, &ds.dsCfg, ds.batcherAddr, ds.fetcher, ds.ref, ds.log)
	if err != nil {
		return nil, err
	}

	if len(hashes) == 0 {
		// there are no blobs to fetch so we can return immediately
		return data, nil
	}

	// download the actual blob bodies corresponding to the versioned hashes
	blobs, err := ds.blobsFetcher.GetBlobsByHash(ctx, ds.ref.Time, hashes)
	if errors.Is(err, ethereum.NotFound) {
		// If the L1 block was available, then the blobs should be available too. The only
		// exception is if the blob retention window has expired, which we will ultimately handle
		// by failing over to a blob archival service.
		return nil, NewResetError(fmt.Errorf("failed to fetch blobs: %w", err))
	} else if err != nil {
		return nil, NewTemporaryError(fmt.Errorf("failed to fetch blobs: %w", err))
	}

	// go back over the data array and populate the blob pointers
	if err := fillBlobPointers(data, blobs); err != nil {
		// this shouldn't happen unless there is a bug in the blobs fetcher
		return nil, NewResetError(fmt.Errorf("failed to fill blob pointers: %w", err))
	}
	return data, nil
}

// dataAndHashesFromTxs extracts calldata and datahashes from the input transactions and returns them. It
// creates a placeholder blobOrCalldata element for each returned blob hash that must be populated
// by fillBlobPointers after blob bodies are retrieved.
//
// Pre-Espresso (the L1 origin time of `ref` is < *EspressoTime, or unset),
// this runs upstream Optimism semantics: filter by batch inbox + sender ==
// batcher.
//
// Post-Espresso, it collects all authenticated batch hashes from a lookback
// window once and rejects any batch whose commitment hash is not in the
// authenticated set. For blob transactions, the batch hash is computed from
// the concatenated blob versioned hashes.
func dataAndHashesFromTxs(ctx context.Context, txs types.Transactions, config *DataSourceConfig, batcherAddr common.Address, fetcher L1Fetcher, ref eth.L1BlockRef, logger log.Logger) ([]blobOrCalldata, []common.Hash, error) {
	// Only collect authenticated batch commitments when the Espresso fork is
	// active at the L1 origin time of the block we're scanning. Pre-fork, the
	// upstream sender-based authorization path is used and authenticatedHashes
	// is unused.
	var authenticatedHashes map[common.Hash]common.Address
	if config.rollupCfg.IsEspresso(ref.Time) {
		var err error
		authenticatedHashes, err = CollectAuthenticatedBatches(
			ctx, fetcher, ref, config.rollupCfg.BatchAuthenticatorAddress, config.rollupCfg.BatchAuthLookbackWindowOrDefault(), config.batchAuthCaches, logger,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	data := []blobOrCalldata{}
	var hashes []common.Hash
	for _, tx := range txs {
		// skip any non-batcher transactions (wrong type or wrong To address)
		if !isValidBatchTx(tx, config.batchInboxAddress, logger) {
			continue
		}

		// Compute batch hash depending on tx type
		var batchHash common.Hash
		if tx.Type() == types.BlobTxType {
			batchHash = ComputeBlobBatchHash(tx.BlobHashes())
		} else {
			batchHash = ComputeCalldataBatchHash(tx.Data())
		}

		// Check authorization (sender-based pre-fork; event-based post-fork).
		if !isBatchTxAuthorized(tx, *config, batcherAddr, batchHash, authenticatedHashes, ref.Time, logger) {
			continue
		}

		// handle non-blob batcher transactions by extracting their calldata
		if tx.Type() != types.BlobTxType {
			calldata := eth.Data(tx.Data())
			data = append(data, blobOrCalldata{nil, &calldata})
			continue
		}
		// handle blob batcher transactions by extracting their blob hashes, ignoring any calldata.
		if len(tx.Data()) > 0 {
			log.Warn("blob tx has calldata, which will be ignored", "txhash", tx.Hash())
		}
		for _, h := range tx.BlobHashes() {
			hashes = append(hashes, h)
			data = append(data, blobOrCalldata{nil, nil}) // will fill in blob pointers after we download them below
		}
	}
	return data, hashes, nil
}

// fillBlobPointers goes back through the data array and fills in the pointers to the fetched blob
// bodies. There should be exactly one placeholder blobOrCalldata element for each blob, otherwise
// error is returned.
func fillBlobPointers(data []blobOrCalldata, blobs []*eth.Blob) error {
	blobIndex := 0
	for i := range data {
		if data[i].calldata != nil {
			continue
		}
		if blobIndex >= len(blobs) {
			return fmt.Errorf("didn't get enough blobs")
		}
		if blobs[blobIndex] == nil {
			return fmt.Errorf("found a nil blob")
		}
		data[i].blob = blobs[blobIndex]
		blobIndex++
	}
	if blobIndex != len(blobs) {
		return fmt.Errorf("got too many blobs")
	}
	return nil
}
