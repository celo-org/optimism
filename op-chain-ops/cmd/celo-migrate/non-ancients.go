package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/log"
)

func migrateNonAncientsDb(oldDbPath, newDbPath string, lastAncientBlock, batchSize uint64) (uint64, error) {
	// First copy files from old database to new database
	log.Info("Copy files from old database (excluding ancients)", "process", "non-ancients")
	cmd := exec.Command("rsync", "-v", "-a", "--exclude=ancient", oldDbPath+"/", newDbPath)
	cmd.Stdout = os.Stdout // TODO(Alec) debug what happens when rsync is run multiple times
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to copy old database to new database: %w", err)
	}

	// Open the new database without access to AncientsDb
	newDB, err := rawdb.NewLevelDBDatabase(newDbPath, DBCache, DBHandles, "", false)
	if err != nil {
		return 0, fmt.Errorf("failed to open new database: %w", err)
	}
	defer newDB.Close()

	// get the last block number
	hash := rawdb.ReadHeadHeaderHash(newDB)
	lastBlock := *rawdb.ReadHeaderNumber(newDB, hash)
	lastMigratedNonAncientBlock := readLastMigratedNonAncientBlock(newDB) // returns 0 if not found

	// if migration was interrupted, start from the last migrated block
	fromBlock := max(lastAncientBlock, lastMigratedNonAncientBlock) + 1

	// TODO(Alec) keep logging of lastMigratedNonAncientBlock here if always 0?

	if fromBlock >= lastBlock {
		log.Info("Non-Ancient Block Migration Skipped", "process", "non-ancients", "lastMigratedNonAncientBlock", lastMigratedNonAncientBlock, "lastAncientBlock", lastAncientBlock, "endBlock", lastBlock)
		if lastMigratedNonAncientBlock != lastBlock {
			return 0, fmt.Errorf("Migration range empty but last migrated block is not the last block in the database")
		}
		return 0, nil
	}

	log.Info("Non-Ancient Block Migration Started", "process", "non-ancients", "startBlock", fromBlock, "endBlock", lastBlock, "count", lastBlock-fromBlock, "lastMigratedNonAncientBlock", lastMigratedNonAncientBlock, "lastAncientBlock", lastAncientBlock)

	for i := fromBlock; i <= lastBlock; i += batchSize {
		numbersHash := rawdb.ReadAllHashesInRange(newDB, i, i+batchSize-1)

		log.Info("Processing Block Range", "process", "non-ancients", "from", i, "to(inclusve)", i+batchSize-1, "count", len(numbersHash))
		for _, numberHash := range numbersHash {
			// read header and body
			header := rawdb.ReadHeaderRLP(newDB, numberHash.Hash, numberHash.Number)
			body := rawdb.ReadBodyRLP(newDB, numberHash.Hash, numberHash.Number)

			// transform header and body
			newHeader, err := transformHeader(header)
			if err != nil {
				return 0, fmt.Errorf("failed to transform header: block %d - %x: %w", numberHash.Number, numberHash.Hash, err)
			}
			newBody, err := transformBlockBody(body)
			if err != nil {
				return 0, fmt.Errorf("failed to transform body: block %d - %x: %w", numberHash.Number, numberHash.Hash, err)
			}

			if yes, newHash := hasSameHash(newHeader, numberHash.Hash[:]); !yes {
				log.Error("Hash mismatch", "block", numberHash.Number, "oldHash", numberHash.Hash, "newHash", newHash)
				return 0, fmt.Errorf("hash mismatch at block %d - %x", numberHash.Number, numberHash.Hash)
			}

			// write header and body
			batch := newDB.NewBatch()
			rawdb.WriteBodyRLP(batch, numberHash.Hash, numberHash.Number, newBody)
			_ = batch.Put(headerKey(numberHash.Number, numberHash.Hash), newHeader)
			_ = writeLastMigratedNonAncientBlock(batch, numberHash.Number)
			if err := batch.Write(); err != nil {
				return 0, fmt.Errorf("failed to write header and body: block %d - %x: %w", numberHash.Number, numberHash.Hash, err)
			}
		}
	}

	toBeRemoved := rawdb.ReadAllHashesInRange(newDB, 1, lastAncientBlock)
	log.Info("Removing frozen blocks", "process", "non-ancients", "count", len(toBeRemoved))
	batch := newDB.NewBatch()
	for _, numberHash := range toBeRemoved {
		rawdb.DeleteBlockWithoutNumber(batch, numberHash.Hash, numberHash.Number)
		rawdb.DeleteCanonicalHash(batch, numberHash.Number)
	}
	if err := batch.Write(); err != nil {
		return 0, fmt.Errorf("failed to delete frozen blocks: %w", err)
	}

	// TODO(Alec) decide whether to keep this or not
	// if migration finished, remove the last migration number
	if err := deleteLastMigratedNonAncientBlock(newDB); err != nil {
		return 0, fmt.Errorf("failed to delete last migration number: %w", err)
	}

	log.Info("Non-Ancient Block Migration Ended", "process", "non-ancients", "migratedBlocks", lastBlock-fromBlock+1, "removedBlocks", len(toBeRemoved))

	return lastBlock - fromBlock + 1, nil
}
