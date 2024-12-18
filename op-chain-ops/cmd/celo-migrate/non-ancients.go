package main

import (
	"bytes"
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
	output, err := cmdHelp.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get rsync help output: %w", err)
	}

	// Convert output to string
	outputStr := string(output)

	opts := []string{"-v", "-a", "--exclude=ancient", "--checksum", "--delete"}

	// Check for supported options
	// Prefer --info=progress2 over --progress
	if strings.Contains(outputStr, "--info") {
		opts = append(opts, "--info=progress2")
	} else if strings.Contains(outputStr, "--progress") {
		opts = append(opts, "--progress")
	}

	cmd := exec.Command("rsync", append(opts, oldDbPath+"/", newDbPath)...)

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
	//
	// '--checksum' This forces rsync to compare the checksums of all files to determine if they are the same. This is slows down the transfer but ensures that source and destination directories end up with the same contents (excluding /ancients).

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

	// Delete bad blocks, we could migrate them, but we have no need of the historical bad blocks. AFAICS bad blocks
	// are stored solely so that they can be retrieved or traced via the debug API, but we are no longer interested
	// in these old bad blocks.
	rawdb.DeleteBadBlocks(newDB)

	// The genesis block is the only block that should remain stored in the non-ancient db even after it is frozen.
	if numAncients > 0 {
		log.Info("Migrating genesis block in non-ancient db", "process", "non-ancients")
		if err := migrateNonAncientBlock(0, rawdb.ReadCanonicalHash(newDB, 0), newDB); err != nil {
			return 0, err
		}
	}

	prevBlockNumber := uint64(numAncients - 1) // Will underflow if numAncients is 0

	for i := numAncients; i <= lastBlock; i += batchSize {
		numbersHash := rawdb.ReadAllHashesInRange(newDB, i, i+batchSize-1)

		log.Info("Processing Block Range", "process", "non-ancients", "from", i, "to(inclusve)", i+batchSize-1, "count", len(numbersHash))
		for _, numberHash := range numbersHash {
			if numberHash.Number != prevBlockNumber+1 { // prevBlocNumber will overflow back to 0 here if numAncients is 0
				return 0, fmt.Errorf("gap found between non-ancient blocks numbered %d and %d. Please delete the target directory and repeat the migration with an uncorrupted source directory", prevBlockNumber, numberHash.Number)
			}
			prevBlockNumber = numberHash.Number

			if err := migrateNonAncientBlock(numberHash.Number, numberHash.Hash, newDB); err != nil {
				return 0, fmt.Errorf("failed to migrate non-ancient block %d - %x: %w", numberHash.Number, numberHash.Hash, err)
			}

			if err := checkOtherDataForNonAncientBlock(numberHash.Number, numberHash.Hash, newDB); err != nil {
				return 0, fmt.Errorf("failed to ensure all non-transformed data is present for non-ancient block %d - %x: %w. Please delete the target directory and repeat the migration with an uncorrupted source directory", numberHash.Number, numberHash.Hash, err)
			}
		}
	}

	migratedCount := lastBlock - numAncients + 1
	return migratedCount, nil
}

func migrateNonAncientBlock(number uint64, hash common.Hash, newDB ethdb.Database) error {
	// read header and body
	header, err := newDB.Get(headerKey(number, hash))
	if err != nil {
		return fmt.Errorf("failed to read header: block %d - %x: %w", number, hash, err)
	}
	body, err := newDB.Get(blockBodyKey(number, hash))
	if err != nil {
		return fmt.Errorf("failed to read body: block %d - %x: %w", number, hash, err)
	}

	// transform header and body
	newHeader, err := transformHeader(header)
	if err != nil {
		return fmt.Errorf("failed to transform header: block %d - %x: %w", number, hash, err)
	}
	newBody, err := transformBlockBody(body)
	if err != nil {
		return fmt.Errorf("failed to transform body: block %d - %x: %w", number, hash, err)
	}

	if err := checkTransformedHeader(newHeader, hash[:], number); err != nil {
		return err
	}

	// write header and body
	batch := newDB.NewBatch()
	rawdb.WriteBodyRLP(batch, hash, number, newBody)
	if err := batch.Put(headerKey(number, hash), newHeader); err != nil {
		return fmt.Errorf("failed to write header: block %d - %x: %w", number, hash, err)
	}
	if err := batch.Write(); err != nil {
		return fmt.Errorf("failed to write header and body: block %d - %x: %w", number, hash, err)
	}

	return nil
}

// checkOtherDataForNonAncientBlock checks that all the data that is not transformed is successfully copied for non-ancient blocks.
// I.e. receipts, total difficulty, canonical hash, and block number.
// If an error is returned, it is likely the source directory is corrupted and the migration should be restarted with a clean source directory.
func checkOtherDataForNonAncientBlock(number uint64, hash common.Hash, newDB ethdb.Database) error {
	// Ensure receipts and total difficulty are present in non-ancient db
	if has, err := newDB.Has(blockReceiptsKey(number, hash)); !has || err != nil {
		return fmt.Errorf("failed to find receipts in newDB leveldb: block %d - %x: %w", number, hash, err)
	}
	if has, err := newDB.Has(headerTDKey(number, hash)); !has || err != nil {
		return fmt.Errorf("failed to find total difficulty in newDB leveldb: block %d - %x: %w", number, hash, err)
	}
	// Ensure canonical hash and number are present in non-ancient db and that they match expected values
	hashFromDB, err := newDB.Get(headerHashKey(number))
	if err != nil {
		return fmt.Errorf("failed to find canonical hash in newDB leveldb: block %d - %x: %w", number, hash, err)
	}
	if !bytes.Equal(hashFromDB, hash[:]) {
		return fmt.Errorf("canonical hash mismatch in newDB leveldb: block %d - %x: %w", number, hash, err)
	}
	numberFromDB, err := newDB.Get(headerNumberKey(hash))
	if err != nil {
		return fmt.Errorf("failed to find number for hash in newDB leveldb: block %d - %x: %w", number, hash, err)
	}
	if !bytes.Equal(numberFromDB, encodeBlockNumber(number)) {
		log.Error("Number for hash mismatch", "block", number, "numberFromDB", numberFromDB, "hash", hash)
		return fmt.Errorf("number for hash mismatch in newDB leveldb: block %d - %x: %w", number, hash, err)
	}

	return nil
}
