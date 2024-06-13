package main

import (
	"math/big"
	"os"
	"testing"

	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	"github.com/holiman/uint256"
	"github.com/mattn/go-isatty"
	"github.com/stretchr/testify/require"
)

var (
	account1 = common.HexToAddress("0x1111")
	account2 = common.HexToAddress("0x2222")
	code     = []byte{5, 7, 8, 3, 4}
	balance  = big.NewInt(100)
)

func createGenesis() *core.Genesis {
	l2Genesis := &core.Genesis{
		Config: &params.ChainConfig{
			ChainID: big.NewInt(123),
		},
		Difficulty: common.Big0,
		ParentHash: common.Hash{},
		BaseFee:    big.NewInt(7),
		Alloc: map[common.Address]types.Account{
			account1: {
				Balance: balance,
				Code:    code,
				Storage: map[common.Hash]common.Hash{
					common.HexToHash("0x01"): common.HexToHash("0x11"),
					common.HexToHash("0x02"): common.HexToHash("0x12"),
					common.HexToHash("0x03"): common.HexToHash("0x13"),
				},
			},
		},
	}
	return l2Genesis
}

func createMigrations() ChainContractMigrations {
	testMigrations := map[common.Address]common.Address{
		account1: account2,
	}

	return ChainContractMigrations{
		123: testMigrations,
	}
}

func TestUpdateState(t *testing.T) {
	color := isatty.IsTerminal(os.Stderr.Fd())
	handler := log.NewTerminalHandler(os.Stderr, color)
	oplog.SetGlobalLogHandler(handler)

	l2Genesis := createGenesis()
	migrations := createMigrations()

	db := rawdb.NewDatabase(memorydb.New())
	trieDB := triedb.NewDatabase(db, &triedb.Config{HashDB: hashdb.Defaults})
	genesisBlock := l2Genesis.MustCommit(db, trieDB)

	statedb, err := state.New(genesisBlock.Root(), state.NewDatabase(rawdb.NewDatabase(db)), nil)
	require.NoError(t, err)

	require.NotEqual(t, code, statedb.GetCode(account2))
	require.NotEqual(t, balance, statedb.GetBalance(account2))
	require.NotEqual(t, common.HexToHash("0x11"), statedb.GetState(account2, common.HexToHash("0x01")))

	err = migrateTestnetAccounts(statedb, l2Genesis.Config, migrations)
	require.NoError(t, err)

	require.Equal(t, code, statedb.GetCode(account2))
	require.Equal(t, uint256.MustFromBig(balance), statedb.GetBalance(account2))
	require.Equal(t, common.HexToHash("0x11"), statedb.GetState(account2, common.HexToHash("0x01")))
}
