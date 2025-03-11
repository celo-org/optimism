package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/ethereum/go-ethereum/log"
)

// Constants for the database
const (
	DBCache                 = 1024                      // size of the cache in MB
	DBHandles               = 60                        // number of handles
	CeloFullMigrationMarker = "celoFullMigrationMarker" // Marker to indicate whether a previous full migration attempt has modified the database
)

var (
	headerPrefix       = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header
	headerTDSuffix     = []byte("t") // headerPrefix + num (uint64 big endian) + hash + headerTDSuffix -> td
	headerHashSuffix   = []byte("n") // headerPrefix + num (uint64 big endian) + headerHashSuffix -> hash
	headerNumberPrefix = []byte("H") // headerNumberPrefix + hash -> num (uint64 big endian)

	blockBodyPrefix     = []byte("b") // blockBodyPrefix + num (uint64 big endian) + hash -> block body
	blockReceiptsPrefix = []byte("r") // blockReceiptsPrefix + num (uint64 big endian) + hash -> block receipts
)

// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

// headerKey = headerPrefix + num (uint64 big endian) + hash
func headerKey(number uint64, hash common.Hash) []byte {
	return append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

// headerTDKey = headerPrefix + num (uint64 big endian) + hash + headerTDSuffix
func headerTDKey(number uint64, hash common.Hash) []byte {
	return append(headerKey(number, hash), headerTDSuffix...)
}

// headerHashKey = headerPrefix + num (uint64 big endian) + headerHashSuffix
func headerHashKey(number uint64) []byte {
	return append(append(headerPrefix, encodeBlockNumber(number)...), headerHashSuffix...)
}

// headerNumberKey = headerNumberPrefix + hash
func headerNumberKey(hash common.Hash) []byte {
	return append(headerNumberPrefix, hash.Bytes()...)
}

// blockBodyKey = blockBodyPrefix + num (uint64 big endian) + hash
func blockBodyKey(number uint64, hash common.Hash) []byte {
	return append(append(blockBodyPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

// blockReceiptsKey = blockReceiptsPrefix + num (uint64 big endian) + hash
func blockReceiptsKey(number uint64, hash common.Hash) []byte {
	return append(append(blockReceiptsPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

// readFullMigrationMarker reads a marker indicating whether a previous full migration attempt has modified the database
func readFullMigrationMarker(db ethdb.KeyValueReader) ([]byte, error) {
	return db.Get([]byte(CeloFullMigrationMarker))
}

// writeFullMigrationMarker writes a marker indicating that a full migration has been attempted on the database
func writeFullMigrationMarker(db ethdb.KeyValueWriter) error {
	return db.Put([]byte(CeloFullMigrationMarker), []byte{1})
}

// checkForPrevFullMigration checks if a previous full migration has already been attempted on the database
func checkForPrevFullMigration(newDBPath string) (bool, error) {
	defer timer("checkForPrevFullMigration")()

	newDB, err := openDBWithoutFreezer(newDBPath, true)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("failed to open new database: %w", err)
	}
	defer newDB.Close()

	marker, _ := readFullMigrationMarker(newDB)
	return len(marker) > 0, nil
}

// Opens a database without access to AncientsDb
func openDBWithoutFreezer(chaindataPath string, readOnly bool) (ethdb.Database, error) {
	if _, err := os.Stat(chaindataPath); err != nil {
		return nil, err
	}

	kvs, err := leveldb.New(chaindataPath, DBCache, DBHandles, "", readOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %w", err)
	}
	newDB := rawdb.NewDatabase(kvs)

	return newDB, nil
}

func createNewDbPathIfNotExists(newDBPath string) error {
	if err := os.MkdirAll(newDBPath, 0755); err != nil {
		return fmt.Errorf("failed to create new database directory: %w", err)
	}
	return nil
}

func removeBlocks(ldb ethdb.Database, numberHashes []*rawdb.NumberHash) error {
	defer timer("removeBlocks")()

	if len(numberHashes) == 0 {
		return nil
	}

	batch := ldb.NewBatch()

	for _, numberHash := range numberHashes {
		log.Debug("Removing block", "block", numberHash.Number)
		rawdb.DeleteBlockWithoutNumber(batch, numberHash.Hash, numberHash.Number)
		rawdb.DeleteCanonicalHash(batch, numberHash.Number)
	}
	if err := batch.Write(); err != nil {
		log.Error("Failed to write batch", "error", err)
	}

	return nil
}

func getHeadHeader(dbpath string) (headHeader *types.Header, err error) {
	db, err := openDBWithoutFreezer(dbpath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to open database at %q err: %w", dbpath, err)
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()

	headHeader = rawdb.ReadHeadHeader(db)
	if headHeader == nil {
		return nil, fmt.Errorf("head header not in database at: %s", dbpath)
	}
	return headHeader, nil
}

func cleanupNonAncientDb(dir string) error {
	log.Info("Cleaning up non-ancient data in new db")

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}
	for _, file := range files {
		if file.Name() != "ancient" {
			err := os.RemoveAll(filepath.Join(dir, file.Name()))
			if err != nil {
				return fmt.Errorf("failed to remove file: %w", err)
			}
		}
	}
	return nil
}
