package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-service/ioutil"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/ethereum-optimism/optimism/op-service/predeploys"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/contracts/addresses"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/holiman/uint256"
)

const (
	MainnetNetworkID   = uint64(42220)
	BaklavaNetworkID   = uint64(62320)
	AlfajoresNetworkID = uint64(44787)

	OutFilePerm = os.FileMode(0o440)
)

var (
	Big10 = uint256.NewInt(10)
	Big9  = uint256.NewInt(9)
	Big18 = uint256.NewInt(18)

	// Allowlist of accounts that are allowed to be overwritten
	// If the value for an account is set to true, the nonce and storage will be overwritten
	// This must be checked for each account, as this might create issues with contracts
	// calling `CREATE` or `CREATE2`
	accountOverwriteAllowlist = map[uint64]map[common.Address]bool{
		// Add any addresses that should be allowed to overwrite existing accounts here.
		AlfajoresNetworkID: {
			// Create2Deployer
			// OP uses a version without an owner who can pause the contract,
			// so we overwrite the existing contract during migration
			common.HexToAddress("0x13b0D85CcB8bf860b6b79AF3029fCA081AE9beF2"): true,

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
			// CreateX
			common.HexToAddress("0xba5Ed099633D3B313e4D5F7bdc1305d3c28ba5Ed"): false,
		},
		BaklavaNetworkID: {
			// DeterministicDeploymentProxy
			common.HexToAddress("0x4e59b44847b379578588920cA78FbF26c0B4956C"): false,
		},
	}
	unreleasedTreasuryAddressMap = map[uint64]common.Address{
		AlfajoresNetworkID: common.HexToAddress("0x07bf0b2461A0cb608D5CF9a82ba97dAbA850F79F"),
		BaklavaNetworkID:   common.HexToAddress("0x022c5d5837E177B6d145761feb4C5574e5b48F5e"),
	}
	celoTokenAddressMap = map[uint64]common.Address{
		AlfajoresNetworkID: addresses.CeloTokenAlfajoresAddress,
		BaklavaNetworkID:   addresses.CeloTokenBaklavaAddress,
		MainnetNetworkID:   addresses.CeloTokenAddress,
	}
)

func applyStateMigrationChanges(config *genesis.DeployConfig, l2Allocs types.GenesisAlloc, dbPath, genesisOutPath string, migrationBlockTime uint64, l1StartBlock *types.Block) (*types.Header, error) {
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
	log.Info("Reading chain tip num from database", "number", *num)

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
	err = applyAllocsToState(db, l2Allocs, accountOverwriteAllowlist[cfg.ChainID.Uint64()])
	if err != nil {
		return nil, fmt.Errorf("cannot apply allocations to state: %w", err)
	}

	// Initialize the unreleased treasury contract
	// This uses the original config which won't enable recent hardforks (and things like the PUSH0 opcode)
	// This is fine, as the token uses solc 0.5.x and therefore compatible bytecode
	err = setupUnreleasedTreasury(db, cfg)
	if err != nil {
		// An error here shouldn't stop the migration, just log it
		log.Warn("Error setting up unreleased treasury", "error", err)
	}

	migrationBlockNumber := new(big.Int).Add(header.Number, common.Big1)

	// We're done messing around with the database, so we can now commit the changes to the DB.
	// Note that this doesn't actually write the changes to disk.
	log.Info("Committing state DB")
	newRoot, err := db.Commit(migrationBlockNumber.Uint64(), true)
	if err != nil {
		return nil, err
	}

	// Set the standard options.
	cfg.LondonBlock = migrationBlockNumber
	cfg.BerlinBlock = migrationBlockNumber
	cfg.ArrowGlacierBlock = migrationBlockNumber
	cfg.GrayGlacierBlock = migrationBlockNumber
	cfg.MergeNetsplitBlock = migrationBlockNumber
	cfg.TerminalTotalDifficulty = big.NewInt(0)
	cfg.TerminalTotalDifficultyPassed = true
	cfg.ShanghaiTime = &migrationBlockTime
	cfg.CancunTime = &migrationBlockTime

	// Set the Optimism options.
	cfg.Optimism = &params.OptimismConfig{
		EIP1559Denominator:       config.EIP1559Denominator,
		EIP1559DenominatorCanyon: &config.EIP1559DenominatorCanyon,
		EIP1559Elasticity:        config.EIP1559Elasticity,
	}

	// Set the Celo options.
	cfg.Celo = &params.CeloConfig{
		EIP1559BaseFeeFloor: config.EIP1559BaseFeeFloor,
	}

	// Set Optimism hardforks
	cfg.BedrockBlock = migrationBlockNumber
	cfg.RegolithTime = &migrationBlockTime
	cfg.CanyonTime = &migrationBlockTime
	cfg.EcotoneTime = &migrationBlockTime
	cfg.FjordTime = &migrationBlockTime
	cfg.GraniteTime = &migrationBlockTime
	cfg.Cel2Time = &migrationBlockTime

	// Calculate the base fee for the migration block.
	baseFee := new(big.Int).SetUint64(params.InitialBaseFee)
	if header.BaseFee != nil {
		baseFee = eip1559.CalcBaseFee(cfg, header, migrationBlockTime)
	}

	// If gas limit was zero at the transition point use a default of 30M.
	// Note that in op-geth we use gasLimit==0 to indicate a pre-gingerbread
	// block and adjust encoding appropriately, so we must make sure that
	// gasLimit is non-zero, because L2 blocks are all post gingerbread.
	gasLimit := header.GasLimit
	if gasLimit == 0 {
		gasLimit = 30e6
	}
	// Create the header for the Cel2 transition block.
	cel2Header := &types.Header{
		ParentHash:  header.Hash(),
		UncleHash:   types.EmptyUncleHash,
		Coinbase:    predeploys.SequencerFeeVaultAddr,
		Root:        newRoot,
		TxHash:      types.EmptyTxsHash,
		ReceiptHash: types.EmptyReceiptsHash,
		Bloom:       types.Bloom{},
		Difficulty:  new(big.Int).Set(common.Big0),
		Number:      migrationBlockNumber,
		GasLimit:    gasLimit,
		GasUsed:     0,
		Time:        migrationBlockTime,
		Extra:       []byte("Celo L2 migration"),
		MixDigest:   common.Hash{},
		Nonce:       types.BlockNonce{},
		BaseFee:     baseFee,
		// Added during Shanghai hardfork
		// As there're no withdrawals in L2, we set it to the empty hash
		WithdrawalsHash: &types.EmptyWithdrawalsHash,
		// Blobs are disabled in L2
		BlobGasUsed:   new(uint64),
		ExcessBlobGas: new(uint64),
		// This is set to the ParentBeaconRoot of the L1 origin (see `PreparePayloadAttributes`)
		// Use the L1 start block's ParentBeaconRoot
		ParentBeaconRoot: l1StartBlock.Header().ParentBeaconRoot,
	}
	log.Info("Build Cel2 migration header", "header", cel2Header)

	// We need to set empty withdrawals in the body, otherwise types.NewBlock will nullify the withdrawals hash in the given header.
	b := &types.Body{
		Withdrawals: []*types.Withdrawal{},
	}
	// Create the Cel2 transition block from the header. Note that there are no transactions,
	// uncle blocks, or receipts in the Cel2 transition block.
	cel2Block := types.NewBlock(cel2Header, b, nil, trie.NewStackTrie(nil))

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

	// Next we write the Cel2 migration block to the database.
	rawdb.WriteTd(ldb, cel2Block.Hash(), cel2Block.NumberU64(), cel2Block.Difficulty())
	rawdb.WriteBlock(ldb, cel2Block)
	rawdb.WriteReceipts(ldb, cel2Block.Hash(), cel2Block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(ldb, cel2Block.Hash(), cel2Block.NumberU64())
	rawdb.WriteHeadBlockHash(ldb, cel2Block.Hash())
	rawdb.WriteHeadFastBlockHash(ldb, cel2Block.Hash())
	rawdb.WriteHeadHeaderHash(ldb, cel2Block.Hash())

	// Mark the first CeL2 block as finalized
	rawdb.WriteFinalizedBlockHash(ldb, cel2Block.Hash())

	// Write the chain config to disk.
	rawdb.WriteChainConfig(ldb, genesisHash, cfg)
	marshalledConfig, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chain config to JSON: %w", err)
	}
	log.Info("Wrote updated chain config", "config", string(marshalledConfig))

	// Write genesis JSON to outfile and store genesis state spec in the database.
	if err = writeGenesis(cfg, ldb, genesisOutPath, genesisHash); err != nil {
		return nil, err
	}

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
func applyAllocsToState(db vm.StateDB, allocs types.GenesisAlloc, allowlist map[common.Address]bool) error {
	log.Info("Starting to migrate OP contracts into state DB")

	copyCounter := 0
	overwriteCounter := 0

	for k, v := range allocs {
		// Check that the balance of the account to written is zero,
		// as we must not create new CELO tokens
		if v.Balance != nil && v.Balance.Cmp(big.NewInt(0)) != 0 {
			return fmt.Errorf("account balance is not zero, would change celo supply: %s", k.Hex())
		}

		if db.Exist(k) {
			writeNonceAndStorage := false
			writeCode, allowed := allowlist[k]

			// If the account is not allowed and has a non zero nonce or code size, bail out we will need to manually investigate how to handle this.
			if !allowed && (db.GetCodeSize(k) > 0 || db.GetNonce(k) > 0) {
				return fmt.Errorf("account exists and is not allowed, account: %s, nonce: %d, code: %d", k.Hex(), db.GetNonce(k), db.GetCode(k))
			}

			// This means that the account just has balance, in that case we wan to copy over the account
			if db.GetCodeSize(k) == 0 && db.GetNonce(k) == 0 {
				writeCode = true
				writeNonceAndStorage = true
			}

			if writeCode {
				overwriteCounter++

				db.SetCode(k, v.Code)

				if writeNonceAndStorage {
					db.SetNonce(k, v.Nonce)
					for key, value := range v.Storage {
						db.SetState(k, key, value)
					}
				}
				log.Info("Overwrote account", "address", k.Hex(), "writeNonceAndStorage", writeNonceAndStorage)
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

	log.Info("Migrated OP contracts into state DB", "totalAllocs", len(allocs), "copiedAccounts", copyCounter, "overwrittenAccounts", overwriteCounter)
	return nil
}

// setupUnreleasedTreasury sets up the unreleased treasury contract with the correct balance
// The balance is set to the difference between the ceiling and the total supply of the token
func setupUnreleasedTreasury(db *state.StateDB, config *params.ChainConfig) error {
	log.Info("Setting up CeloUnreleasedTreasury balance")

	celoUnreleasedTreasuryAddress, exists := unreleasedTreasuryAddressMap[config.ChainID.Uint64()]
	if !exists {
		return errors.New("CeloUnreleasedTreasury address not configured for this chain, skipping migration step")
	}

	if !db.Exist(celoUnreleasedTreasuryAddress) {
		return errors.New("CeloUnreleasedTreasury account does not exist, skipping migration step")
	}

	tokenAddress, exists := celoTokenAddressMap[config.ChainID.Uint64()]
	if !exists {
		return errors.New("celo token address not configured for this chain, skipping migration step")
	}
	log.Info("Read contract addresses", "tokenAddress", tokenAddress, "celoUnreleasedTreasuryAddress", celoUnreleasedTreasuryAddress)

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
	balance = new(uint256.Int).Add(balance, db.GetBalance(celoUnreleasedTreasuryAddress))
	db.SetBalance(celoUnreleasedTreasuryAddress, balance, tracing.BalanceChangeUnspecified)

	log.Info("Set up CeloUnreleasedTreasury balance", "celoUnreleasedTreasuryAddress", celoUnreleasedTreasuryAddress, "balance", balance, "total_supply", supplyU256, "ceiling", ceiling)
	return nil
}

// writeGenesis writes the genesis json to --outfile.genesis and stores the genesis state spec (alloc) in the database.
// Note that this is different than the cel2Block / migration block. Rather, this is the migrated genesis block of Celo from before the L2 transition.
// Nodes will need the genesis json file in order to snap sync on the L2 chain.
func writeGenesis(config *params.ChainConfig, db ethdb.Database, genesisOutPath string, genesisHash common.Hash) error {
	// Derive the genesis object using hardcoded legacy alloc and the transformed extra data stored in the new db.
	legacyGenesisAlloc, err := GetCeloL1GenesisAlloc(config)
	if err != nil {
		return err
	}
	genesisHeader := rawdb.ReadHeader(db, genesisHash, 0)
	genesis, err := BuildGenesis(config, legacyGenesisAlloc, genesisHeader.Extra, genesisHeader.Time)
	if err != nil {
		return err
	}

	// Convert genesis to JSON byte slice
	genesisBytes, err := json.Marshal(genesis)
	if err != nil {
		return fmt.Errorf("failed to marshal genesis to JSON: %w", err)
	}

	// Unmarshal JSON byte slice to map
	var genesisMap map[string]interface{}
	if err := json.Unmarshal(genesisBytes, &genesisMap); err != nil {
		return fmt.Errorf("failed to unmarshal genesis JSON to map: %w", err)
	}

	// Delete fields that are not in Celo Legacy Genesis, otherwise genesis hashes won't match when syncing
	delete(genesisMap, "difficulty")
	delete(genesisMap, "gasLimit")
	delete(genesisMap, "excessBlobGas")
	delete(genesisMap, "blobGasUsed")
	delete(genesisMap, "baseFeePerGas")
	delete(genesisMap, "mixHash")
	delete(genesisMap, "nonce")

	// Write the modified JSON to the file
	if err := jsonutil.WriteJSON(genesisMap, ioutil.ToStdOutOrFileOrNoop(genesisOutPath, OutFilePerm)); err != nil {
		return fmt.Errorf("failed to write genesis JSON to file: %w", err)
	}
	log.Info("Wrote genesis file for syncing new nodes", "path", genesisOutPath)

	// Legacy Celo did not store the genesis state spec (alloc) in the database.
	// Write it now for forward compatibility.
	rawdb.WriteGenesisStateSpec(db, genesisHash, legacyGenesisAlloc)
	log.Info("Wrote genesis state spec (alloc) to database")

	return nil
}
