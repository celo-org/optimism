package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-service/predeploys"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/addresses"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/holiman/uint256"
)

var (
	Big10 = uint256.NewInt(10)
	Big9  = uint256.NewInt(9)
	Big18 = uint256.NewInt(18)

	OutFilePerm = os.FileMode(0o440)

	alfajoresChainId uint64 = 44787
	mainnetChainId   uint64 = 42220

	// Allowlist of accounts that are allowed to be overwritten
	// If the value for an account is set to true, the nonce and storage will be overwritten
	// This must be checked for each account, as this might create issues with contracts
	// calling `CREATE` or `CREATE2`
	accountOverwriteAllowlist = map[uint64]map[common.Address]bool{
		// Add any addresses that should be allowed to overwrite existing accounts here.
		alfajoresChainId: {
			// Create2Deployer
			common.HexToAddress("0x13b0D85CcB8bf860b6b79AF3029fCA081AE9beF2"): false,

			// Same code as in allocs file
			// EntryPoint_v070
			common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"): false,
			// Permit2
			common.HexToAddress("0x000000000022D473030F116dDEE9F6B43aC78BA3"): false,
			// EntryPoint_v060
			common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789"): false,
			// DeterministicDeploymentProxy
			common.HexToAddress("0x4e59b44847b379578588920cA78FbF26c0B4956C"): false,
			// SafeL2_v130
			common.HexToAddress("0xfb1bffC9d739B8D520DaF37dF666da4C687191EA"): false,
			// MultiSend_v130
			common.HexToAddress("0x998739BFdAAdde7C933B942a68053933098f9EDa"): false,
			// SenderCreator_v070
			common.HexToAddress("0xEFC2c1444eBCC4Db75e7613d20C6a62fF67A167C"): false,
			// SenderCreator_v060
			common.HexToAddress("0x7fc98430eAEdbb6070B35B39D798725049088348"): false,
			// MultiCall3
			common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"): false,
			// Safe_v130
			common.HexToAddress("0x69f4D1788e39c87893C980c06EdF4b7f686e2938"): false,
			// MultiSendCallOnly_v130
			common.HexToAddress("0xA1dabEF33b3B82c7814B6D82A79e50F4AC44102B"): false,
			// SafeSingletonFactory
			common.HexToAddress("0x914d7Fec6aaC8cd542e72Bca78B30650d45643d7"): false,
		},
	}
	distributionScheduleAddressMap = map[uint64]common.Address{
		alfajoresChainId: common.HexToAddress("0x78af211ad79bce6bf636640ce8c2c2b29e02365a"),
	}
	celoTokenAddressMap = map[uint64]common.Address{
		alfajoresChainId: addresses.CeloTokenAlfajoresAddress,
		mainnetChainId:   addresses.CeloTokenAddress,
	}
)

func applyStateMigrationChanges(config *genesis.DeployConfig, genesis *core.Genesis, dbPath string, migrationBlockTime uint64) (*types.Header, error) {
	log.Info("Opening Celo database", "dbPath", dbPath)

	ldb, err := openDBWithoutFreezer(dbPath, false)
	if err != nil {
		return nil, fmt.Errorf("cannot open DB: %w", err)
	}
	log.Info("Loaded Celo L1 DB", "db", ldb)

	// Grab the hash of the tip of the legacy chain.
	hash := rawdb.ReadHeadHeaderHash(ldb)
	log.Info("Reading chain tip from database", "hash", hash)

	// Grab the header number.
	num := rawdb.ReadHeaderNumber(ldb, hash)
	if num == nil {
		return nil, fmt.Errorf("cannot find header number for %s", hash)
	}
	log.Info("Reading chain tip num from database", "number", num)

	// Grab the full header.
	header := rawdb.ReadHeader(ldb, hash, *num)
	log.Info("Read header from database", "header", header)

	// We need to update the chain config to set the correct hardforks.
	genesisHash := rawdb.ReadCanonicalHash(ldb, 0)
	cfg := rawdb.ReadChainConfig(ldb, genesisHash)
	if cfg == nil {
		log.Crit("chain config not found")
	}
	log.Info("Read chain config from database", "config", cfg)

	// Set up the backing store.
	underlyingDB := state.NewDatabase(ldb)

	// Open up the state database.
	db, err := state.New(header.Root, underlyingDB, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot open StateDB: %w", err)
	}

	// Apply the changes to the state DB.
	err = applyAllocsToState(db, genesis, accountOverwriteAllowlist[cfg.ChainID.Uint64()])
	if err != nil {
		return nil, fmt.Errorf("cannot apply allocations to state: %w", err)
	}

	// Initialize the distribution schedule contract
	// This uses the original config which won't enable recent hardforks (and things like the PUSH0 opcode)
	// This is fine, as the token uses solc 0.5.x and therefore compatible bytecode
	err = setupDistributionSchedule(db, cfg)
	if err != nil {
		// An error here shouldn't stop the migration, just log it
		log.Warn("Error setting up distribution schedule", "error", err)
	}

	migrationBlock := new(big.Int).Add(header.Number, common.Big1)

	// We're done messing around with the database, so we can now commit the changes to the DB.
	// Note that this doesn't actually write the changes to disk.
	log.Info("Committing state DB")
	newRoot, err := db.Commit(migrationBlock.Uint64(), true)
	if err != nil {
		return nil, err
	}

	baseFee := new(big.Int).SetUint64(params.InitialBaseFee)
	if header.BaseFee != nil {
		baseFee = header.BaseFee
	}

	if migrationBlockTime == 0 {
		migrationBlockTime = uint64(time.Now().Unix())
	}

	// If gas limit was zero at the transition point use a default of 30M.
	// Note that in op-geth we use gasLimit==0 to indicate a pre-gingerbread
	// block and adjust encoding appropriately, so we must make sure that
	// gasLimit is non-zero, bacause L2 blocks are all post gingerbread.
	gasLimit := header.GasLimit
	if gasLimit == 0 {
		gasLimit = 30e6
	}
	// Create the header for the Cel2 transition block.
	cel2Header := &types.Header{
		ParentHash:       header.Hash(),
		UncleHash:        types.EmptyUncleHash,
		Coinbase:         predeploys.SequencerFeeVaultAddr,
		Root:             newRoot,
		TxHash:           types.EmptyTxsHash,
		ReceiptHash:      types.EmptyReceiptsHash,
		Bloom:            types.Bloom{},
		Difficulty:       new(big.Int).Set(common.Big0),
		Number:           migrationBlock,
		GasLimit:         gasLimit,
		GasUsed:          0,
		Time:             migrationBlockTime,
		Extra:            []byte("CeL2 migration"),
		MixDigest:        common.Hash{},
		Nonce:            types.BlockNonce{},
		BaseFee:          baseFee,
		WithdrawalsHash:  &types.EmptyWithdrawalsHash,
		BlobGasUsed:      new(uint64),
		ExcessBlobGas:    new(uint64),
		ParentBeaconRoot: &common.Hash{},
	}
	log.Info("Build Cel2 migration header", "header", cel2Header)

	// Create the Cel2 transition block from the header. Note that there are no transactions,
	// uncle blocks, or receipts in the Cel2 transition block.
	cel2Block := types.NewBlock(cel2Header, nil, nil, trie.NewStackTrie(nil))

	// We did it!
	log.Info(
		"Built Cel2 migration block",
		"hash", cel2Block.Hash(),
		"root", cel2Block.Root(),
		"number", cel2Block.NumberU64(),
	)

	log.Info("Committing trie DB")
	if err := db.Database().TrieDB().Commit(newRoot, true); err != nil {
		return nil, err
	}

	// Next we write the Cel2 genesis block to the database.
	rawdb.WriteTd(ldb, cel2Block.Hash(), cel2Block.NumberU64(), cel2Block.Difficulty())
	rawdb.WriteBlock(ldb, cel2Block)
	rawdb.WriteReceipts(ldb, cel2Block.Hash(), cel2Block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(ldb, cel2Block.Hash(), cel2Block.NumberU64())
	rawdb.WriteHeadBlockHash(ldb, cel2Block.Hash())
	rawdb.WriteHeadFastBlockHash(ldb, cel2Block.Hash())
	rawdb.WriteHeadHeaderHash(ldb, cel2Block.Hash())

	// Mark the first CeL2 block as finalized
	rawdb.WriteFinalizedBlockHash(ldb, cel2Block.Hash())

	// Set the standard options.
	cfg.LondonBlock = cel2Block.Number()
	cfg.BerlinBlock = cel2Block.Number()
	cfg.ArrowGlacierBlock = cel2Block.Number()
	cfg.GrayGlacierBlock = cel2Block.Number()
	cfg.MergeNetsplitBlock = cel2Block.Number()
	cfg.TerminalTotalDifficulty = big.NewInt(0)
	cfg.TerminalTotalDifficultyPassed = true
	cfg.ShanghaiTime = &cel2Header.Time
	cfg.CancunTime = &cel2Header.Time

	// Set the Optimism options.
	cfg.BedrockBlock = cel2Block.Number()
	// Enable Regolith from the start of Bedrock
	cfg.RegolithTime = new(uint64) // what are those? do we need those?
	cfg.Optimism = &params.OptimismConfig{
		EIP1559Denominator:       config.EIP1559Denominator,
		EIP1559DenominatorCanyon: &config.EIP1559DenominatorCanyon,
		EIP1559Elasticity:        config.EIP1559Elasticity,
	}
	cfg.CanyonTime = &cel2Header.Time
	cfg.EcotoneTime = &cel2Header.Time
	cfg.FjordTime = &cel2Header.Time
	cfg.Cel2Time = &cel2Header.Time

	// Write the chain config to disk.
	rawdb.WriteChainConfig(ldb, genesisHash, cfg)
	marhslledConfig, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chain config to JSON: %w", err)
	}
	log.Info("Wrote updated chain config", "config", string(marhslledConfig))

	// We're done!
	log.Info(
		"Wrote CeL2 migration block",
		"height", cel2Header.Number,
		"root", cel2Header.Root.String(),
		"hash", cel2Header.Hash().String(),
		"timestamp", cel2Header.Time,
	)

	// Close the database handle
	if err := ldb.Close(); err != nil {
		return nil, err
	}

	return cel2Header, nil
}

// applyAllocsToState applies the account allocations from the allocation file to the state database.
// It creates new accounts, sets their nonce, balance, code, and storage values.
// If an account already exists, it adds the balance of the new account to the existing balance.
// If the code of an existing account is different from the code in the genesis block, it logs a warning.
// This changes the state root, so `Commit` needs to be called after this function.
func applyAllocsToState(db vm.StateDB, genesis *core.Genesis, allowlist map[common.Address]bool) error {
	log.Info("Starting to migrate OP contracts into state DB")

	copyCounter := 0
	overwriteCounter := 0

	for k, v := range genesis.Alloc {
		// Check that the balance of the account to written is zero,
		// as we must not create new CELO tokens
		if v.Balance != nil && v.Balance.Cmp(big.NewInt(0)) != 0 {
			return fmt.Errorf("account balance is not zero, would change celo supply: %s", k.Hex())
		}

		if db.Exist(k) {
			var allowed bool
			overwrite, allowed := allowlist[k]

			// If the account is not allowed and has a non zero nonce or code size, bail out we will need to manually investigate how to handle this.
			if !allowed && (db.GetCodeSize(k) > 0 || db.GetNonce(k) > 0) {
				return fmt.Errorf("account exists and is not allowed, account: %s, nonce: %d, code: %d", k.Hex(), db.GetNonce(k), db.GetCode(k))
			}

			// This means that the account just has balance, in that case we wan to copy over the account
			if db.GetCodeSize(k) == 0 && db.GetNonce(k) == 0 {
				overwrite = true
			}

			if overwrite {
				overwriteCounter++

				db.SetCode(k, v.Code)
				db.SetNonce(k, v.Nonce)
				for key, value := range v.Storage {
					db.SetState(k, key, value)
				}
				log.Info("Overwrote account", "address", k.Hex())
			}
			continue
		}

		// Account does not exist, create it
		db.CreateAccount(k)
		db.SetCode(k, v.Code)
		db.SetNonce(k, v.Nonce)
		for key, value := range v.Storage {
			db.SetState(k, key, value)
		}

		copyCounter++
		log.Info("Copied account", "address", k.Hex())
	}

	log.Info("Migrated OP contracts into state DB", "totalAllocs", len(genesis.Alloc), "copiedAccounts", copyCounter, "overwrittenAccounts", overwriteCounter)
	return nil
}

// setupDistributionSchedule sets up the distribution schedule contract with the correct balance
// The balance is set to the difference between the ceiling and the total supply of the token
func setupDistributionSchedule(db *state.StateDB, config *params.ChainConfig) error {
	log.Info("Setting up CeloDistributionSchedule balance")

	celoDistributionScheduleAddress, exists := distributionScheduleAddressMap[config.ChainID.Uint64()]
	if !exists {
		return errors.New("DistributionSchedule address not configured for this chain, skipping migration step")
	}

	if !db.Exist(celoDistributionScheduleAddress) {
		return errors.New("DistributionSchedule account does not exist, skipping migration step")
	}

	tokenAddress, exists := celoTokenAddressMap[config.ChainID.Uint64()]
	if !exists {
		return errors.New("celo token address not configured for this chain, skipping migration step")
	}
	log.Info("Read contract addresses", "tokenAddress", tokenAddress, "distributionScheduleAddress", celoDistributionScheduleAddress)

	// totalSupply is stored in the third slot
	totalSupply := db.GetState(tokenAddress, common.HexToHash("0x02")).Big()

	// Get total supply of celo token
	billion := new(uint256.Int).Exp(Big10, Big9)
	ethInWei := new(uint256.Int).Exp(Big10, Big18)

	ceiling := new(uint256.Int).Mul(billion, ethInWei)

	supplyU256 := uint256.MustFromBig(totalSupply)
	if supplyU256.Cmp(ceiling) > 0 {
		return fmt.Errorf("supply %s is greater than ceiling %s", totalSupply, ceiling)
	}

	balance := new(uint256.Int).Sub(ceiling, supplyU256)
	// Don't discard existing balance of the account
	balance = new(uint256.Int).Add(balance, db.GetBalance(celoDistributionScheduleAddress))
	db.SetBalance(celoDistributionScheduleAddress, balance, tracing.BalanceChangeUnspecified)

	log.Info("Set up CeloDistributionSchedule balance", "distributionScheduleAddress", celoDistributionScheduleAddress, "balance", balance, "total_supply", supplyU256, "ceiling", ceiling)
	return nil
}
