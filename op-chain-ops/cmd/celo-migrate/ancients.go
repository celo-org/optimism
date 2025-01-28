package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"golang.org/x/sync/errgroup"
)

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
	g.Go(func() error { return transformBlocks(ctx, readChan, transformChan) })
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
		count := min(batchSize, endBlock-i)
		start := i

		blockRange, err := loadAncientRange(freezer, start, count)
		if err != nil {
			return fmt.Errorf("Failed to load ancient block range: %w", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- *blockRange:
		}

		log.Info("Read ancient blocks", "start", start, "end", start+count-1, "count", count)
	}
	return nil
}

func loadAncientRange(freezer *rawdb.Freezer, start, count uint64) (*RLPBlockRange, error) {
	log.Info("Loading ancient block range", "start", start, "end", start+count-1, "count", count)

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
		return nil, fmt.Errorf("failed to read hashes from freezer: %w", err)
	}
	blockRange.headers, err = freezer.AncientRange(rawdb.ChainFreezerHeaderTable, start, count, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read headers from freezer: %w", err)
	}
	blockRange.bodies, err = freezer.AncientRange(rawdb.ChainFreezerBodiesTable, start, count, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read bodies from freezer: %w", err)
	}
	blockRange.receipts, err = freezer.AncientRange(rawdb.ChainFreezerReceiptTable, start, count, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read receipts from freezer: %w", err)
	}
	blockRange.tds, err = freezer.AncientRange(rawdb.ChainFreezerDifficultyTable, start, count, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read tds from freezer: %w", err)
	}

	return blockRange, nil
}

func transformBlocks(ctx context.Context, in <-chan RLPBlockRange, out chan<- RLPBlockRange) error {
	// Transform blocks from the in channel and send them to the out channel
	defer close(out)

	for blockRange := range in {
		for i := range blockRange.hashes {
			blockNumber := blockRange.start + uint64(i)

			newHeader, err := transformHeader(blockRange.headers[i])
			if err != nil {
				return fmt.Errorf("can't transform header: %w", err)
			}
			newBody, err := transformBlockBody(blockRange.bodies[i])
			if err != nil {
				return fmt.Errorf("can't transform body: %w", err)
			}

			if yes, newHash := hasSameHash(newHeader, blockRange.hashes[i]); !yes {
				log.Error("Hash mismatch", "block", blockNumber, "oldHash", common.BytesToHash(blockRange.hashes[i]), "newHash", newHash)
				return fmt.Errorf("hash mismatch at block %d", blockNumber)
			}

			blockRange.headers[i] = newHeader
			blockRange.bodies[i] = newBody
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- blockRange:
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

// Get the last ancient block data so we can check for continuity between ancients and non-ancients
func loadLastAncient(freezer *rawdb.Freezer) (*RLPBlockElement, error) {
	numAncients, err := freezer.Ancients()
	if err != nil {
		return nil, fmt.Errorf("failed to get number of ancients in freezer: %w", err)
	}
	blockRange, err := loadAncientRange(freezer, numAncients-1, 1)
	if err != nil {
		return nil, err
	}
	return blockRange.Element(0)
}
