package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
)

// Constants for the database
const (
	DBCache   = 1024 // size of the cache in MB
	DBHandles = 60   // number of handles
)

var (
	headerPrefix = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header
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

// Opens a database with access to AncientsDb
func openDB(chaindataPath string, readOnly bool) (ethdb.Database, error) {
	if _, err := os.Stat(chaindataPath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	db, err := rawdb.Open(rawdb.OpenOptions{
		Type:              "leveldb",
		Directory:         chaindataPath,
		AncientsDirectory: filepath.Join(chaindataPath, "ancient"),
		Namespace:         "",
		Cache:             DBCache,
		Handles:           DBHandles,
		ReadOnly:          readOnly,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Opens a database without access to AncientsDb
func openDBWithoutFreezer(chaindataPath string, readOnly bool) (ethdb.Database, error) {
	if _, err := os.Stat(chaindataPath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	newDB, err := rawdb.NewLevelDBDatabase(chaindataPath, DBCache, DBHandles, "", readOnly)
	if err != nil {
		return nil, err
	}

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
