package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/sync/errgroup"
)

// RLPBlockRange is a range of blocks in RLP format
type RLPBlockRange struct {
	start    uint64
	hashes   [][]byte
	headers  [][]byte
	bodies   [][]byte
	receipts [][]byte
	tds      [][]byte
}

type RLPBlockElement struct {
	decodedHeader *types.Header
	hash          []byte
	header        []byte
	body          []byte
	receipts      []byte
	td            []byte
}

func (r *RLPBlockRange) Element(i uint64) (*RLPBlockElement, error) {
	header := types.Header{}
	err := rlp.DecodeBytes(r.headers[i], &header)
	if err != nil {
		return nil, fmt.Errorf("can't decode header: %w", err)
	}
	return &RLPBlockElement{
		decodedHeader: &header,
		hash:          r.hashes[i],
		header:        r.headers[i],
		body:          r.bodies[i],
		receipts:      r.receipts[i],
		td:            r.tds[i],
	}, nil
}

func (r *RLPBlockRange) DropFirst() {
	r.start = r.start + 1
	r.hashes = r.hashes[1:]
	r.headers = r.headers[1:]
	r.bodies = r.bodies[1:]
	r.receipts = r.receipts[1:]
	r.tds = r.tds[1:]
}

func (e *RLPBlockElement) Header() *types.Header {

	return e.decodedHeader
}

func (e *RLPBlockElement) Follows(prev *RLPBlockElement) error {
	if e.Header().Number.Uint64() != prev.Header().Number.Uint64()+1 {
		return fmt.Errorf("header number mismatch: expected %d, actual %d", prev.Header().Number.Uint64()+1, e.Header().Number.Uint64())
	}
	// We compare the parent hash with the stored hash of the previous block because
	// at this point the header object will not calculate the correct hash since it
	// first needs to be transformed.
	if e.Header().ParentHash != common.Hash(prev.hash) {
		return fmt.Errorf("parent hash mismatch between blocks %d and %d", e.Header().Number.Uint64(), prev.Header().Number.Uint64())
	}
	return nil
}

// NewChainFreezer is a small utility method around NewFreezer that sets the
// default parameters for the chain storage.
func NewChainFreezer(datadir string, namespace string, readonly bool) (*rawdb.Freezer, error) {
	const freezerTableSize = 2 * 1000 * 1000 * 1000
	// chainFreezerNoSnappy configures whether compression is disabled for the ancient-tables.
	// Hashes and difficulties don't compress well.
	var chainFreezerNoSnappy = map[string]bool{
		rawdb.ChainFreezerHeaderTable:     false,
		rawdb.ChainFreezerHashTable:       true,
		rawdb.ChainFreezerBodiesTable:     false,
		rawdb.ChainFreezerReceiptTable:    false,
		rawdb.ChainFreezerDifficultyTable: true,
	}
	return rawdb.NewFreezer(datadir, namespace, readonly, freezerTableSize, chainFreezerNoSnappy)
}

func migrateAncientsDb(ctx context.Context, oldDBPath, newDBPath string, batchSize, bufferSize uint64) (numAncientsNewBefore uint64, numAncientsNewAfter uint64, err error) {
	defer timer("ancients")()

	oldFreezer, err := NewChainFreezer(filepath.Join(oldDBPath, "ancient"), "", false) // Can't be readonly because we need the .meta files to be created
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open old freezer: %w", err)
	}
	defer func() {
		err = errors.Join(err, oldFreezer.Close())
	}()

	newFreezer, err := NewChainFreezer(filepath.Join(newDBPath, "ancient"), "", false)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open new freezer: %w", err)
	}
	defer func() {
		err = errors.Join(err, newFreezer.Close())
	}()

	numAncientsOld, err := oldFreezer.Ancients()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get number of ancients in old freezer: %w", err)
	}

	numAncientsNewBefore, err = newFreezer.Ancients()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get number of ancients in new freezer: %w", err)
	}

	if numAncientsNewBefore >= numAncientsOld {
		log.Info("Ancient Block Migration Skipped", "process", "ancients", "ancientsInOldDB", numAncientsOld, "ancientsInNewDB", numAncientsNewBefore)
		return numAncientsNewBefore, numAncientsNewBefore, nil
	}

	log.Info("Ancient Block Migration Started", "process", "ancients", "startBlock", numAncientsNewBefore, "endBlock", numAncientsOld-1, "count", numAncientsOld-numAncientsNewBefore, "step", batchSize)

	g, ctx := errgroup.WithContext(ctx)
	readChan := make(chan RLPBlockRange, bufferSize)
	transformChan := make(chan RLPBlockRange, bufferSize)

	g.Go(func() error {
		return readAncientBlocks(ctx, oldFreezer, numAncientsNewBefore, numAncientsOld, batchSize, readChan)
	})
	g.Go(func() error { return transformBlocks(ctx, readChan, transformChan, numAncientsNewBefore) })
	g.Go(func() error { return writeAncientBlocks(ctx, newFreezer, transformChan, numAncientsOld) })

	if err = g.Wait(); err != nil {
		return 0, 0, fmt.Errorf("failed to migrate ancients: %w", err)
	}

	numAncientsNewAfter, err = newFreezer.Ancients()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get number of ancients in new freezer: %w", err)
	}

	if numAncientsNewAfter != numAncientsOld {
		return 0, 0, fmt.Errorf("failed to migrate all ancients from old to new db. Expected %d, got %d", numAncientsOld, numAncientsNewAfter)
	}

	log.Info("Ancient Block Migration Ended", "process", "ancients", "ancientsInOldDB", numAncientsOld, "ancientsInNewDB", numAncientsNewAfter, "migrated", numAncientsNewAfter-numAncientsNewBefore)
	return numAncientsNewBefore, numAncientsNewAfter, nil
}

func readAncientBlocks(ctx context.Context, freezer *rawdb.Freezer, startBlock, endBlock, batchSize uint64, out chan<- RLPBlockRange) error {
	defer close(out)
	for i := startBlock; i < endBlock; i += batchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			count := min(batchSize, endBlock-i)
			start := i
			// If we are not at genesis include the last block of
			// the previous range so we can check for continuity between ranges.
			if start > 0 {
				start = start - 1
				count = count + 1
			}

			blockRange, err := loadRange(freezer, start, count)
			if err != nil {
				return fmt.Errorf("failed to load ancient block range: %w", err)
			}

			// Check continuity between blocks
			var prevElement *RLPBlockElement
			for i := uint64(0); i < count; i++ {
				currElement, err := blockRange.Element(i)
				if err != nil {
					return err
				}
				if prevElement != nil {
					if err := currElement.Follows(prevElement); err != nil {
						return err
					}
				}
				prevElement = currElement
			}

			if start > 0 {
				blockRange.DropFirst()
			}
			out <- *blockRange
		}
	}
	return nil
}

func loadRange(freezer *rawdb.Freezer, start, count uint64) (*RLPBlockRange, error) {
	blockRange := &RLPBlockRange{
		start:    start,
		hashes:   make([][]byte, count),
		headers:  make([][]byte, count),
		bodies:   make([][]byte, count),
		receipts: make([][]byte, count),
		tds:      make([][]byte, count),
	}

	var err error
	blockRange.hashes, err = freezer.AncientRange(rawdb.ChainFreezerHashTable, start, count, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read hashes from old freezer: %w", err)
	}
	blockRange.headers, err = freezer.AncientRange(rawdb.ChainFreezerHeaderTable, start, count, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read headers from old freezer: %w", err)
	}
	blockRange.bodies, err = freezer.AncientRange(rawdb.ChainFreezerBodiesTable, start, count, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read bodies from old freezer: %w", err)
	}
	blockRange.receipts, err = freezer.AncientRange(rawdb.ChainFreezerReceiptTable, start, count, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read receipts from old freezer: %w", err)
	}
	blockRange.tds, err = freezer.AncientRange(rawdb.ChainFreezerDifficultyTable, start, count, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read tds from old freezer: %w", err)
	}

	// Make sure the number of elements retrieved from each table matches the expected length
	if uint64(len(blockRange.hashes)) != count {
		err = fmt.Errorf("Expected count mismatch in block range hashes: expected %d, actual %d", count, len(blockRange.hashes))
	}
	if uint64(len(blockRange.bodies)) != count {
		err = errors.Join(err, fmt.Errorf("Expected count mismatch in block range bodies: expected %d, actual %d", count, len(blockRange.bodies)))
	}
	if uint64(len(blockRange.headers)) != count {
		err = errors.Join(err, fmt.Errorf("Expected count mismatch in block range headers: expected %d, actual %d", count, len(blockRange.headers)))
	}
	if uint64(len(blockRange.receipts)) != count {
		err = errors.Join(err, fmt.Errorf("Expected count mismatch in block range receipts: expected %d, actual %d", count, len(blockRange.receipts)))
	}
	if uint64(len(blockRange.tds)) != count {
		err = errors.Join(err, fmt.Errorf("Expected count mismatch in block range total difficulties: expected %d, actual %d", count, len(blockRange.tds)))
	}
	return blockRange, err
}

func transformBlocks(ctx context.Context, in <-chan RLPBlockRange, out chan<- RLPBlockRange, startBlock uint64) error {
	// Transform blocks from the in channel and send them to the out channel
	defer close(out)

	prevBlockNumber := uint64(startBlock - 1) // Will underflow when startBlock is 0, but then overflow back to 0

	for blockRange := range in {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			for i := range blockRange.hashes {
				blockNumber := blockRange.start + uint64(i)

				if blockNumber != prevBlockNumber+1 { // Overflows back to 0 when startBlock is 0
					return fmt.Errorf("gap found between ancient blocks numbered %d and %d. Please delete the target directory and repeat the migration with an uncorrupted source directory", prevBlockNumber, blockNumber)
				}
				// Block ranges are in order because they are read sequentially from the freezer
				prevBlockNumber = blockNumber

				newHeader, err := transformHeader(blockRange.headers[i])
				if err != nil {
					return fmt.Errorf("can't transform header: %w", err)
				}
				newBody, err := transformBlockBody(blockRange.bodies[i])
				if err != nil {
					return fmt.Errorf("can't transform body: %w", err)
				}

				if err := checkTransformedHeader(newHeader, blockRange.hashes[i], blockNumber); err != nil {
					return err
				}

				blockRange.headers[i] = newHeader
				blockRange.bodies[i] = newBody
			}
			out <- blockRange
		}
	}
	return nil
}

func writeAncientBlocks(ctx context.Context, freezer *rawdb.Freezer, in <-chan RLPBlockRange, totalAncientBlocks uint64) error {
	// Write blocks from the in channel to the newDb
	for blockRange := range in {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err := freezer.ModifyAncients(func(aWriter ethdb.AncientWriteOp) error {
				for i := range blockRange.hashes {
					blockNumber := blockRange.start + uint64(i)
					if err := aWriter.AppendRaw(rawdb.ChainFreezerHashTable, blockNumber, blockRange.hashes[i]); err != nil {
						return fmt.Errorf("can't write hash to Freezer: %w", err)
					}
					if err := aWriter.AppendRaw(rawdb.ChainFreezerHeaderTable, blockNumber, blockRange.headers[i]); err != nil {
						return fmt.Errorf("can't write header to Freezer: %w", err)
					}
					if err := aWriter.AppendRaw(rawdb.ChainFreezerBodiesTable, blockNumber, blockRange.bodies[i]); err != nil {
						return fmt.Errorf("can't write body to Freezer: %w", err)
					}
					if err := aWriter.AppendRaw(rawdb.ChainFreezerReceiptTable, blockNumber, blockRange.receipts[i]); err != nil {
						return fmt.Errorf("can't write receipts to Freezer: %w", err)
					}
					if err := aWriter.AppendRaw(rawdb.ChainFreezerDifficultyTable, blockNumber, blockRange.tds[i]); err != nil {
						return fmt.Errorf("can't write td to Freezer: %w", err)
					}
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to write block range: %w", err)
			}
			blockRangeEnd := blockRange.start + uint64(len(blockRange.hashes)) - 1
			log.Info("Wrote ancient blocks", "start", blockRange.start, "end", blockRangeEnd, "count", len(blockRange.hashes), "remaining", totalAncientBlocks-(blockRangeEnd+1))
		}
	}
	return nil
}

// getStrayAncientBlocks returns a list of ancient block numbers / hashes that somehow were not removed from leveldb
func getStrayAncientBlocks(dbPath string) (blocks []*rawdb.NumberHash, err error) {
	defer timer("getStrayAncientBlocks")()

	db, err := openDB(dbPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()

	numAncients, err := db.Ancients()
	if err != nil {
		return nil, fmt.Errorf("failed to get number of ancients in database: %w", err)
	}

	return rawdb.ReadAllHashesInRange(db, 1, numAncients-1), nil
}
