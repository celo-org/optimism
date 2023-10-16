package derive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/Layr-Labs/eigenda/api/grpc/retriever"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/protobuf/proto"

	"github.com/ethereum-optimism/optimism/op-node/da"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/proto/gen/op_service/v1"
)

type DataIter interface {
	Next(ctx context.Context) (eth.Data, error)
}

type L1TransactionFetcher interface {
	InfoAndTxsByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, types.Transactions, error)
}

// DataSourceFactory readers raw transactions from a given block & then filters for
// batch submitter transactions.
// This is not a stage in the pipeline, but a wrapper for another stage in the pipeline
type DataSourceFactory struct {
	log     log.Logger
	cfg     *rollup.Config
	fetcher L1TransactionFetcher
	daCfg   *da.DAConfig
}

func NewDataSourceFactory(log log.Logger, cfg *rollup.Config, daCfg *da.DAConfig, fetcher L1TransactionFetcher) *DataSourceFactory {
	return &DataSourceFactory{log: log, cfg: cfg, daCfg: daCfg, fetcher: fetcher}
}

// OpenData returns a DataIter. This struct implements the `Next` function.
func (ds *DataSourceFactory) OpenData(ctx context.Context, id eth.BlockID, batcherAddr common.Address) DataIter {
	return NewDataSource(ctx, ds.log, ds.cfg, ds.daCfg, ds.fetcher, id, batcherAddr)
}

// DataSource is a fault tolerant approach to fetching data.
// The constructor will never fail & it will instead re-attempt the fetcher
// at a later point.
type DataSource struct {
	// Internal state + data
	open bool
	data []eth.Data
	// Required to re-attempt fetching
	id      eth.BlockID
	cfg     *rollup.Config // TODO: `DataFromEVMTransactions` should probably not take the full config
	fetcher L1TransactionFetcher
	log     log.Logger

	batcherAddr common.Address
	daCfg       *da.DAConfig
}

// NewDataSource creates a new calldata source. It suppresses errors in fetching the L1 block if they occur.
// If there is an error, it will attempt to fetch the result on the next call to `Next`.
func NewDataSource(ctx context.Context, log log.Logger, cfg *rollup.Config, daCfg *da.DAConfig, fetcher L1TransactionFetcher, block eth.BlockID, batcherAddr common.Address) DataIter {
	_, txs, err := fetcher.InfoAndTxsByHash(ctx, block.Hash)
	if err != nil {
		return &DataSource{
			open:        false,
			id:          block,
			cfg:         cfg,
			fetcher:     fetcher,
			log:         log,
			batcherAddr: batcherAddr,
			daCfg:       daCfg,
		}
	} else {
		return &DataSource{
			open:  true,
			data:  DataFromEVMTransactions(cfg, daCfg, batcherAddr, txs, log.New("origin", block)),
			daCfg: daCfg,
		}
	}
}

// Next returns the next piece of data if it has it. If the constructor failed, this
// will attempt to reinitialize itself. If it cannot find the block it returns a ResetError
// otherwise it returns a temporary error if fetching the block returns an error.
func (ds *DataSource) Next(ctx context.Context) (eth.Data, error) {
	if !ds.open {
		if _, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.id.Hash); err == nil {
			ds.open = true
			ds.data = DataFromEVMTransactions(ds.cfg, ds.daCfg, ds.batcherAddr, txs, log.New("origin", ds.id))
		} else if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to open calldata source: %w", err))
		} else {
			return nil, NewTemporaryError(fmt.Errorf("failed to open calldata source: %w", err))
		}
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
func DataFromEVMTransactions(config *rollup.Config, daCfg *da.DAConfig, batcherAddr common.Address, txs types.Transactions, log log.Logger) []eth.Data {
	var out []eth.Data
	l1Signer := config.L1Signer()
	for j, tx := range txs {
		if to := tx.To(); to != nil && *to == config.BatchInboxAddress {
			seqDataSubmitter, err := l1Signer.Sender(tx) // optimization: only derive sender if To is correct
			if err != nil {
				log.Warn("tx in inbox with invalid signature", "index", j, "err", err)
				continue // bad signature, ignore
			}
			// some random L1 user might have sent a transaction to our batch inbox, ignore them
			if seqDataSubmitter != batcherAddr {
				log.Warn("tx in inbox with unauthorized submitter", "index", j, "err", err)
				continue // not an authorized batch submitter, ignore
			}

			calldataFrame := &op_service.CalldataFrame{}
			err = proto.Unmarshal(tx.Data(), calldataFrame)
			if err != nil {
				log.Warn("unable to decode calldata frame", "index", j, "err", err)
				return nil
			}

			switch calldataFrame.Value.(type) {
			case *op_service.CalldataFrame_FrameRef:
				frameRef := calldataFrame.GetFrameRef()
				if len(frameRef.QuorumIds) == 0 {
					log.Warn("decoded frame ref contains no quorum IDs", "index", j, "err", err)
					return nil
				}

				log.Info("requesting data from EigenDA", "quorum id", frameRef.QuorumIds[0], "confirmation block number", frameRef.ReferenceBlockNumber)
				blobRequest := &retriever.BlobRequest{
					BatchHeaderHash:      frameRef.BatchHeaderHash,
					BlobIndex:            frameRef.BlobIndex,
					ReferenceBlockNumber: frameRef.ReferenceBlockNumber,
					QuorumId:             frameRef.QuorumIds[0],
				}
				blobRes, err := daCfg.Client.RetrieveBlob(context.Background(), blobRequest)
				if err != nil {
					retrieveReqJSON, _ := json.Marshal(struct {
						batch_header_hash      string
						blob_index             uint32
						reference_block_number uint32
						quorum_id              uint32
					}{
						batch_header_hash:      base64.StdEncoding.EncodeToString(frameRef.BatchHeaderHash),
						blob_index:             frameRef.BlobIndex,
						reference_block_number: frameRef.ReferenceBlockNumber,
						quorum_id:              frameRef.QuorumIds[0],
					})
					log.Warn("could not retrieve data from EigenDA", "request", string(retrieveReqJSON), "err", err)
					return nil
				}
				log.Info("Successfully retrieved data from EigenDA", "quorum id", frameRef.QuorumIds[0], "confirmation block number", frameRef.ReferenceBlockNumber)
				data := blobRes.Data[:frameRef.BlobLength]
				out = append(out, data)
			case *op_service.CalldataFrame_Frame:
				log.Info("Successfully read data from calldata (not EigenDA)")
				frame := calldataFrame.GetFrame()
				out = append(out, frame)
			}

			out = append(out, tx.Data())
		}
	}
	return out
}
