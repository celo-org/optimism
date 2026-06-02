package derive

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	altda "github.com/ethereum-optimism/optimism/op-alt-da"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type DataIter interface {
	Next(ctx context.Context) (eth.Data, error)
}

type L1TransactionFetcher interface {
	InfoAndTxsByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, types.Transactions, error)
}

type L1BlobsFetcher interface {
	// GetBlobsByHash fetches blobs that were confirmed at the given timestamp with the given versioned hashes.
	GetBlobsByHash(ctx context.Context, time uint64, hashes []common.Hash) ([]*eth.Blob, error)
}

type AltDAInputFetcher interface {
	// GetInput fetches the input for the given commitment at the given block number from the DA storage service.
	GetInput(ctx context.Context, l1 altda.L1Fetcher, c altda.CommitmentData, blockId eth.L1BlockRef) (eth.Data, error)
	// AdvanceL1Origin advances the L1 origin to the given block number, syncing the DA challenge events.
	AdvanceL1Origin(ctx context.Context, l1 altda.L1Fetcher, blockId eth.BlockID) error
	// Reset the challenge origin in case of L1 reorg
	Reset(ctx context.Context, base eth.L1BlockRef, baseCfg eth.SystemConfig) error
}

// DataSourceFactory reads raw transactions from a given block & then filters for
// batch submitter transactions.
// This is not a stage in the pipeline, but a wrapper for another stage in the pipeline
type DataSourceFactory struct {
	log          log.Logger
	dsCfg        DataSourceConfig
	fetcher      L1Fetcher
	blobsFetcher L1BlobsFetcher
	altDAFetcher AltDAInputFetcher
	ecotoneTime  *uint64
}

func NewDataSourceFactory(log log.Logger, cfg *rollup.Config, fetcher L1Fetcher, blobsFetcher L1BlobsFetcher, altDAFetcher AltDAInputFetcher) *DataSourceFactory {
	lookbackWindow := cfg.BatchAuthLookbackWindowOrDefault()
	var caches *BatchAuthCaches
	if cfg.EspressoTime != nil {
		caches = NewBatchAuthCaches(lookbackWindow)
	}
	config := DataSourceConfig{
		l1Signer:                  cfg.L1Signer(),
		batchInboxAddress:         cfg.BatchInboxAddress,
		altDAEnabled:              cfg.AltDAEnabled(),
		batchAuthenticatorAddress: cfg.BatchAuthenticatorAddress,
		batchAuthLookbackWindow:   lookbackWindow,
		batchAuthCaches:           caches,
		espressoTime:              cfg.EspressoTime,
	}
	return &DataSourceFactory{
		log:          log,
		dsCfg:        config,
		fetcher:      fetcher,
		blobsFetcher: blobsFetcher,
		altDAFetcher: altDAFetcher,
		ecotoneTime:  cfg.EcotoneTime,
	}
}

// OpenData returns the appropriate data source for the L1 block `ref`.
//
// The Espresso gate is evaluated against the L1 origin time (ref.Time),
// mirroring the upstream pattern used for ecotoneTime: the data-source layer
// is per-L1-block, so it gates on L1 time. The fork timestamp itself is
// conceptually an L2 timestamp but the per-L1-block decision is stable as
// long as L1 origin time and L2 block time are within MaxSequencerDrift of
// each other (always true on a healthy chain).
func (ds *DataSourceFactory) OpenData(ctx context.Context, ref eth.L1BlockRef, batcherAddr common.Address) (DataIter, error) {
	// Creates a data iterator from blob or calldata source so we can forward it to the altDA source
	// if enabled as it still requires an L1 data source for fetching input commmitments.
	var src DataIter
	if ds.ecotoneTime != nil && ref.Time >= *ds.ecotoneTime {
		if ds.blobsFetcher == nil {
			return nil, NewCriticalError(fmt.Errorf("ecotone upgrade active but beacon endpoint not configured"))
		}
		src = NewBlobDataSource(ctx, ds.log, ds.dsCfg, ds.fetcher, ds.blobsFetcher, ref, batcherAddr)
	} else {
		src = NewCalldataSource(ctx, ds.log, ds.dsCfg, ds.fetcher, ref, batcherAddr)
	}
	if ds.dsCfg.altDAEnabled {
		// altDA([calldata | blobdata](l1Ref)) -> data
		return NewAltDADataSource(ds.log, src, ds.fetcher, ds.altDAFetcher, ref), nil
	}
	return src, nil
}

// DataSourceConfig regroups the mandatory rollup.Config fields needed for DataFromEVMTransactions.
type DataSourceConfig struct {
	l1Signer          types.Signer
	batchInboxAddress common.Address
	altDAEnabled      bool
	// batchAuthenticatorAddress is the L1 address of the BatchAuthenticator contract.
	// Event-based authentication via this contract is required only post-Espresso
	// activation; pre-fork the data source uses upstream sender-based authorization.
	batchAuthenticatorAddress common.Address
	// batchAuthLookbackWindow is the number of L1 blocks to scan for BatchInfoAuthenticated events.
	batchAuthLookbackWindow uint64
	// batchAuthCaches holds the LRU caches for batch authentication lookback
	// window traversal. Nil when Espresso is not configured.
	batchAuthCaches *BatchAuthCaches
	// espressoTime is the activation timestamp of the Espresso hardfork. When the
	// L1 origin time of the block being scanned is >= *espressoTime (and this
	// pointer is non-nil), batches must be authenticated by emitted
	// BatchInfoAuthenticated events. Otherwise upstream sender-based
	// authorization applies.
	espressoTime *uint64
}

// isEspresso returns true if the Espresso hardfork is active for the given L1
// origin time. The fork is conceptually an L2-timestamp hardfork but the
// per-L1-block data-source decision is gated on L1 origin time, mirroring
// upstream's ecotoneTime treatment.
func (c DataSourceConfig) isEspresso(l1OriginTime uint64) bool {
	return c.espressoTime != nil && l1OriginTime >= *c.espressoTime
}

// isValidBatchTx checks basic transaction validity for batch submission:
//  1. the transaction type is any of Legacy, ACL, DynamicFee, Blob, or Deposit (for L3s).
//  2. the transaction has a To() address that matches the batch inbox address
//
// It does NOT check authentication (sender or event-based) — that is handled separately
// by isBatchTxAuthorized.
func isValidBatchTx(tx *types.Transaction, batchInboxAddr common.Address, logger log.Logger) bool {
	// For now, we want to disallow the SetCodeTx type or any future types.
	if tx.Type() > types.BlobTxType && tx.Type() != types.DepositTxType {
		return false
	}

	to := tx.To()
	if to == nil || *to != batchInboxAddr {
		return false
	}

	return true
}

// isAuthorizedBatchSender performs upstream-style sender-based authorization: it
// recovers the L1 sender of the transaction and checks it matches the configured
// batcher address. This is the pre-Espresso authorization path.
func isAuthorizedBatchSender(tx *types.Transaction, l1Signer types.Signer, batcherAddr common.Address, logger log.Logger) bool {
	sender, err := l1Signer.Sender(tx)
	if err != nil {
		logger.Warn("tx in inbox with invalid signature", "hash", tx.Hash(), "err", err)
		return false
	}
	if sender != batcherAddr {
		logger.Warn("tx in inbox with unauthorized submitter", "addr", sender, "hash", tx.Hash())
		return false
	}
	return true
}

// isBatchTxAuthorized determines whether a batch transaction is authorized for inclusion.
//
// The fork gate is evaluated against the L1 origin time of the enclosing L1
// block (passed as l1OriginTime), mirroring the data-source layer's ecotoneTime
// treatment.
//
// Pre-Espresso (l1OriginTime < *EspressoTime, or unset):
//
//	upstream behavior — the L1 sender of the transaction must match the configured
//	batcher address. The authenticatedHashes map is unused.
//
// Post-Espresso:
//
//	the batch's commitment hash must appear in authenticatedHashes (i.e. a
//	BatchInfoAuthenticated event was emitted for this commitment within the
//	derivation pipeline's lookback window) AND the L1 sender of the batch
//	transaction must equal the caller that emitted that event. This binds each
//	batch to the address that authenticated it, so a batch authenticated by one
//	batcher cannot be submitted by another. Sender-based-only authorization is
//	rejected.
func isBatchTxAuthorized(
	tx *types.Transaction,
	dsCfg DataSourceConfig,
	batcherAddr common.Address,
	batchHash common.Hash,
	authenticatedHashes map[common.Hash]common.Address,
	l1OriginTime uint64,
	logger log.Logger,
) bool {
	if !dsCfg.isEspresso(l1OriginTime) {
		// Pre-fork: upstream sender-based authorization.
		return isAuthorizedBatchSender(tx, dsCfg.l1Signer, batcherAddr, logger)
	}
	// Post-fork: the commitment must have been authenticated within the lookback window.
	authCaller, ok := authenticatedHashes[batchHash]
	if !ok {
		logger.Warn("batch not authenticated",
			"txHash", tx.Hash(), "batchHash", batchHash)
		return false
	}
	// The batch tx must be submitted by the same address that authenticated it.
	sender, err := dsCfg.l1Signer.Sender(tx)
	if err != nil {
		logger.Warn("authenticated batch tx with invalid signature",
			"txHash", tx.Hash(), "batchHash", batchHash, "err", err)
		return false
	}
	if sender != authCaller {
		logger.Warn("authenticated batch submitted by a different sender than the authenticating caller",
			"txHash", tx.Hash(), "batchHash", batchHash, "sender", sender, "authCaller", authCaller)
		return false
	}
	return true
}
