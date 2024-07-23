package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
)

func copyDbExceptAncients(oldDbPath, newDbPath string) error {
	log.Info("Copy files from old database (excluding ancients)", "process", "non-ancients")

	// Get rsync help output
	cmdHelp := exec.Command("rsync", "--help")
	output, _ := cmdHelp.CombinedOutput()

	// Convert output to string
	outputStr := string(output)

	// Check for supported options
	var cmd *exec.Cmd
	// Prefer --info=progress2 over --progress
	if strings.Contains(outputStr, "--info") {
		cmd = exec.Command("rsync", "-v", "-a", "--info=progress2", "--exclude=ancient", "--update", "--delete", oldDbPath+"/", newDbPath)
	} else if strings.Contains(outputStr, "--progress") {
		cmd = exec.Command("rsync", "-v", "-a", "--progress", "--exclude=ancient", "--update", "--delete", oldDbPath+"/", newDbPath)
	} else {
		cmd = exec.Command("rsync", "-v", "-a", "--exclude=ancient", "--update", "--delete", oldDbPath+"/", newDbPath)
	}

	// The '--update' and '--delete' flags together instruct rsync to only copy over new data if this command is re-run

	log.Info("Running rsync command", "command", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy old database to new database: %w", err)
	}
	return nil
}

func migrateNonAncientsDb(newDB ethdb.Database, lastBlock, numAncients, batchSize uint64) (uint64, error) {
	for i := numAncients; i <= lastBlock; i += batchSize {
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
			if err := batch.Write(); err != nil {
				return 0, fmt.Errorf("failed to write header and body: block %d - %x: %w", numberHash.Number, numberHash.Hash, err)
			}
		}
	}

	var lastAncient = numAncients - 1
	if lastAncient > 0 {
		toBeRemoved := rawdb.ReadAllHashesInRange(newDB, 1, lastAncient)
		log.Info("Removing frozen blocks", "process", "non-ancients", "count", len(toBeRemoved))
		batch := newDB.NewBatch()
		for _, numberHash := range toBeRemoved {
			rawdb.DeleteBlockWithoutNumber(batch, numberHash.Hash, numberHash.Number)
			rawdb.DeleteCanonicalHash(batch, numberHash.Number)
		}
		if err := batch.Write(); err != nil {
			return 0, fmt.Errorf("failed to delete frozen blocks: %w", err)
		}
		log.Info("Removed frozen blocks, still in leveldb", "process", "non-ancients", "removedBlocks", len(toBeRemoved))
	}
	migratedCount := lastBlock - numAncients + 1
	return migratedCount, nil
}
