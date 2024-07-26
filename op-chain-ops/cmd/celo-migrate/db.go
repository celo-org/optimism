package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
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
