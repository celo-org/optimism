package main

import (
	"github.com/ethereum/go-ethereum/common"
)

type ChainContractMigrations map[uint64]map[common.Address]common.Address

var (
	alfajoresChainID   uint64 = 123
	alfajoresContracts        = map[common.Address]common.Address{
		common.HexToAddress("0x123"): common.HexToAddress("0x456"),
	}

	contractMigrations = ChainContractMigrations{
		alfajoresChainID: alfajoresContracts,
	}
)
