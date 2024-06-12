package main

import (
	"github.com/ethereum/go-ethereum/common"
)

type ChainContractMigrations map[uint64]map[common.Address]common.Address

var (
	alfajoresChainID   uint64 = 44787
	alfajoresContracts        = map[common.Address]common.Address{
		// GoldToken
		common.HexToAddress("0xF194afDf50B03e69Bd7D057c1Aa9e10c9954E4C9"): common.HexToAddress("0x471EcE3750Da237f93B8E339c536989b8978a438"),
		// FeeHandler
		common.HexToAddress("0xEAaFf71AB67B5d0eF34ba62Ea06Ac3d3E2dAAA38"): common.HexToAddress("0xcD437749E43A154C07F3553504c68fBfD56B8778"),
		// FeeCurrencyDirectory
		// TODO: Add once the address is known
	}

	contractMigrations = ChainContractMigrations{
		alfajoresChainID: alfajoresContracts,
	}
)
