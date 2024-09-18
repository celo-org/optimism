package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
)

func copyDbExceptAncients(oldDbPath, newDbPath string) error {
	defer timer("copyDbExceptAncients")()

	log.Info("Copying files from old database (excluding ancients)", "process", "non-ancients")

	// Get rsync help output
	cmdHelp := exec.Command("rsync", "--help")
	output, _ := cmdHelp.CombinedOutput()

	// Convert output to string
	outputStr := string(output)

	// Check for supported options
	var cmd *exec.Cmd
	// Prefer --info=progress2 over --progress
	if strings.Contains(outputStr, "--info") {
		cmd = exec.Command("rsync", "-v", "-a", "--info=progress2", "--exclude=ancient", "--delete", oldDbPath+"/", newDbPath)
	} else if strings.Contains(outputStr, "--progress") {
		cmd = exec.Command("rsync", "-v", "-a", "--progress", "--exclude=ancient", "--delete", oldDbPath+"/", newDbPath)
	} else {
		cmd = exec.Command("rsync", "-v", "-a", "--exclude=ancient", "--delete", oldDbPath+"/", newDbPath)
	}

	// rsync copies any file with a different timestamp or size.
	//
	// '--exclude=ancient' excludes the ancient directory from the copy
	//
	// '--delete' Tells rsync to delete extraneous files from the receiving side (ones that aren’t on the sending side)
	//
	// '-a' archive mode; equals -rlptgoD. It is a quick way of saying you want recursion and want to preserve almost everything, including timestamps, ownerships, permissions, etc.
	// Timestamps are important here because they are used to determine which files are newer and should be copied over.
	//
	// '--whole-file' This is the default when both the source and destination are specified as local paths, which they are here (oldDbPath and newDbPath).
	// This option disables rsync’s delta-transfer algorithm, which causes all transferred files to be sent whole. The delta-transfer algorithm is normally used when the destination is a remote system.

	log.Info("Running rsync command", "command", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy old database to new database: %w", err)
	}
	return nil
}

func migrateNonAncientsDb(newDB ethdb.Database, lastBlock, numAncients, batchSize uint64) (uint64, error) {
	defer timer("migrateNonAncientsDb")()

	// The genesis block is also migrated in the ancient db migration as it is stored in both places.
	// The genesis block is the only block that should remain stored in the non-ancient db even after it is frozen.
	if numAncients > 0 {
		log.Info("Migrating genesis block in non-ancient db", "process", "non-ancients")
		if err := migrateNonAncientBlock(0, rawdb.ReadCanonicalHash(newDB, 0), newDB); err != nil {
			return 0, err
		}
	}

	for i := numAncients; i <= lastBlock; i += batchSize {
		numbersHash := rawdb.ReadAllHashesInRange(newDB, i, i+batchSize-1)

		log.Info("Processing Block Range", "process", "non-ancients", "from", i, "to(inclusve)", i+batchSize-1, "count", len(numbersHash))
		for _, numberHash := range numbersHash {
			if err := migrateNonAncientBlock(numberHash.Number, numberHash.Hash, newDB); err != nil {
				return 0, err
			}
		}
	}

	migratedCount := lastBlock - numAncients + 1
	return migratedCount, nil
}

func migrateNonAncientBlock(number uint64, hash common.Hash, newDB ethdb.Database) error {
	// read header and body
	header := rawdb.ReadHeaderRLP(newDB, hash, number)
	body := rawdb.ReadBodyRLP(newDB, hash, number)

	// transform header and body
	newHeader, err := transformHeader(header)
	if err != nil {
		return fmt.Errorf("failed to transform header: block %d - %x: %w", number, hash, err)
	}
	newBody, err := transformBlockBody(body)
	if err != nil {
		return fmt.Errorf("failed to transform body: block %d - %x: %w", number, hash, err)
	}

	if yes, newHash := hasSameHash(newHeader, hash[:]); !yes {
		log.Error("Hash mismatch", "block", number, "oldHash", hash, "newHash", newHash)
		return fmt.Errorf("hash mismatch at block %d - %x", number, hash)
	}

	// write header and body
	batch := newDB.NewBatch()
	rawdb.WriteBodyRLP(batch, hash, number, newBody)
	_ = batch.Put(headerKey(number, hash), newHeader)
	if err := batch.Write(); err != nil {
		return fmt.Errorf("failed to write header and body: block %d - %x: %w", number, hash, err)
	}

	return nil
}
