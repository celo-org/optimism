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
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
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

// openDB opens the chaindata database at the given path. Note this path is below the datadir
func openDB(chaindataPath string) (ethdb.Database, error) {
	if _, err := os.Stat(chaindataPath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	kvs, err := leveldb.New(chaindataPath, 1024, 60, "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %w", err)
	}
	ldb, err := rawdb.NewDatabaseWithFreezer(kvs, filepath.Join(chaindataPath, "ancient"), "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to open db with freezer: %w", err)
	}

	return ldb, nil
}

func createNewDbIfNotExists(newDBPath string) error {
	if err := os.MkdirAll(newDBPath, 0755); err != nil {
		return fmt.Errorf("failed to create new database directory: %w", err)
	}
	return nil
}

func cleanupNonAncientDb(dir string) error {
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
