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
	"github.com/syndtr/goleveldb/leveldb"
)

// Constants for the database
const (
	DBCache                = 1024                  // size of the cache in MB
	DBHandles              = 60                    // number of handles
	CeloMigrationMarkerKey = "celoMigrationMarker" // Marker to indicate whether a previous full migration attempt has modified the database
)

var (
	headerPrefix             = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header
	celoMigrationMarkerValue = []byte{1}   // Marker to indicate whether a previous full migration attempt has modified the database
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
func openDB(chaindataPath string, readOnly bool) (ethdb.Database, error) {
	if _, err := os.Stat(chaindataPath); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	ldb, err := rawdb.Open(rawdb.OpenOptions{
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
	return ldb, nil
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

// readFullMigrationMarker reads a marker indicating whether a previous full migration attempt has modified the database
func readFullMigrationMarker(db ethdb.KeyValueReader) ([]byte, error) {
	return db.Get([]byte(CeloMigrationMarkerKey))
}

// writeFullMigrationMarker writes a marker indicating that a full migration has been attempted on the database
func writeFullMigrationMarker(db ethdb.KeyValueWriter) error {
	return db.Put([]byte(CeloMigrationMarkerKey), celoMigrationMarkerValue)
}

// checkForPrevFullMigration checks if a previous full migration has already been attempted on the database
func checkForPrevFullMigration(newDBPath string, measureTime bool) (bool, error) {
	if measureTime {
		// TODO(Alec) this function is taking ~31s on alfajores
		defer timer("checkForPrevFullMigration")()
	}

	newDB, err := openDBWithoutFreezer(newDBPath, true)
	if err != nil {
		return false, fmt.Errorf("failed to open new database: %w", err)
	}
	defer newDB.Close()

	marker, err := readFullMigrationMarker(newDB)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read celo migration marker: %w", err)
	}

	return len(marker) > 0, nil
}

// TODO(Alec) delete these
// func checkForPrevFullMigration(newDBPath string) (bool, error) {
// 	newDB, err := openDB(newDBPath)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to open new database: %w", err)
// 	}
// 	defer newDB.Close()

// 	// Get body of first non-ancient block and see if it's already been transformed
// 	numAncients, err := newDB.Ancients()
// 	if err != nil {
// 		return false, fmt.Errorf("failed to read numAncients from new database: %w", err)
// 	}
// 	hash := rawdb.ReadCanonicalHash(newDB, numAncients)
// 	bodyData := rawdb.ReadBodyRLP(newDB, hash, numAncients)

// 	log.Info("bodyData: ", bodyData, "numAncients: ", numAncients, "hash: ", hash)

// 	// Try decoding body into celo-blockchain Body structure
// 	var celoBody struct {
// 		Transactions   types.Transactions
// 		Randomness     rlp.RawValue
// 		EpochSnarkData rlp.RawValue
// 	}
// 	if err := rlp.DecodeBytes(bodyData, &celoBody); err != nil {
// 		// body didn't decode, may have already been transformed in a previous migration
// 		body := types.Body{}
// 		if err := rlp.DecodeBytes(bodyData, &body); err == nil {
// 			// body decodes into our new type, so it's already been transformed in a previous migration
// 			return true, nil
// 		}
// 		return false, fmt.Errorf("failed to RLP decode body: %w", err)
// 	}

// 	return false, nil
// }

// func checkForPrevFullMigration(oldDBPath, newDBPath string) (bool, error) {
// 	// Compare the last modified time of files in both directories
// 	newFiles, err := os.ReadDir(newDBPath)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to read directory: %w", err)
// 	}

// 	for _, newFile := range newFiles {
// 		if newFile.Name() != "ancient" {
// 			newPath := filepath.Join(newDBPath, newFile.Name())
// 			oldPath := filepath.Join(oldDBPath, newFile.Name())

// 			newFileInfo, err := os.Stat(newPath)
// 			if err != nil {
// 				return false, fmt.Errorf("failed to get file info for %s: %w", newPath, err)
// 			}
// 			oldFileInfo, err := os.Stat(oldPath)
// 			if err != nil {
// 				return false, fmt.Errorf("failed to get file info for %s: %w", oldPath, err)
// 			}

// 			if newFileInfo.ModTime().After(oldFileInfo.ModTime()) {
// 				// New file has been modified more recently
// 				return true, nil
// 			}
// 		}
// 	}

// 	return false, nil
// }

func resetDbIfNeeded(newDbPath string, measureTime bool) error {
	if measureTime {
		defer timer("resetDbIfNeeded")()
	}

	var err error
	var prevFullMigration bool
	if prevFullMigration, err = checkForPrevFullMigration(newDbPath, measureTime); err != nil {
		return fmt.Errorf("failed to check if a full migration has previously been run on the new db: %w", err)
	}
	if prevFullMigration {
		log.Info("Previous full migration attempt detected, resetting all non-ancient data")
		if err = cleanupNonAncientDb(newDbPath); err != nil {
			return fmt.Errorf("failed to reset non-ancient database: %w", err)
		}
		return nil
	}
	log.Info("No previous full migration attempt detected, leaving non-ancient data intact")
	return nil
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
