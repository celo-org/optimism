package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/mattn/go-isatty"

	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	deployConfigFlag = &cli.PathFlag{
		Name:     "deploy-config",
		Usage:    "Path to deploy config file",
		Required: true,
	}
	l1DeploymentsFlag = &cli.PathFlag{
		Name:  "l1-deployments",
		Usage: "Path to L1 deployments JSON file. Cannot be used with --deployment-dir",
	}
	l1RPCFlag = &cli.StringFlag{
		Name:  "l1-rpc",
		Usage: "RPC URL for an Ethereum L1 node. Cannot be used with --l1-starting-block",
	}
	outfileL2Flag = &cli.PathFlag{
		Name:  "outfile.l2",
		Usage: "Path to L2 genesis output file",
	}
	outfileRollupFlag = &cli.PathFlag{
		Name:  "outfile.rollup",
		Usage: "Path to rollup output file",
	}

	dbPathFlag = &cli.StringFlag{
		Name:     "db-path",
		Usage:    "Path to database",
		Required: true,
	}
	dbCacheFlag = &cli.IntFlag{
		Name:  "db-cache",
		Usage: "LevelDB cache size in mb",
		Value: 1024,
	}
	dbHandlesFlag = &cli.IntFlag{
		Name:  "db-handles",
		Usage: "LevelDB number of handles",
		Value: 60,
	}
	dryRunFlag = &cli.BoolFlag{
		Name:  "dry-run",
		Usage: "Dry run the upgrade by not committing the database",
	}

	flags = []cli.Flag{
		deployConfigFlag,
		l1DeploymentsFlag,
		l1RPCFlag,
		outfileL2Flag,
		outfileRollupFlag,
		dbPathFlag,
		dbCacheFlag,
		dbHandlesFlag,
		dryRunFlag,
	}

	// from `packages/contracts-bedrock/deploy-config/internal-devnet.json`
	EIP1559Denominator = uint64(50) // TODO(pl): select values
	EIP1559Elasticity  = uint64(10)
)

func main() {
	log.Root().SetHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(isatty.IsTerminal(os.Stderr.Fd()))))

	app := &cli.App{
		Name:  "migrate",
		Usage: "Migrate Celo state to a CeL2 DB",
		Flags: flags,
		Action: func(ctx *cli.Context) error {
			deployConfig := ctx.Path("deploy-config")
			if deployConfig == "" {
				return fmt.Errorf("must specify --deploy-config")
			}
			log.Info("Deploy config", "path", deployConfig)
			config, err := genesis.NewDeployConfig(deployConfig)
			if err != nil {
				return err
			}

			// Try reading the L1 deployment information
			l1Deployments := ctx.Path("l1-deployments")
			if l1Deployments == "" {
				return fmt.Errorf("must specify --l1-deployments")
			}
			deployments, err := genesis.NewL1Deployments(l1Deployments)
			if err != nil {
				return fmt.Errorf("cannot read L1 deployments at %s: %w", l1Deployments, err)
			}
			config.SetDeployments(deployments)

			// Get latest block information from L1
			l1RPC := ctx.String("l1-rpc")
			if l1RPC == "" {
				return fmt.Errorf("must specify --l1-rpc")
			}

			outfileL2 := ctx.Path("outfile.l2")
			if outfileL2 == "" {
				return fmt.Errorf("must specify --outfile.l2")
			}

			outfileRollup := ctx.Path("outfile.rollup")
			if outfileRollup == "" {
				return fmt.Errorf("must specify --outfile.rollup")
			}

			var l1StartBlock *types.Block
			client, err := ethclient.Dial(l1RPC)
			if err != nil {
				return fmt.Errorf("cannot dial %s: %w", l1RPC, err)
			}

			if config.L1StartingBlockTag == nil {
				l1StartBlock, err = client.BlockByNumber(context.Background(), nil)
				if err != nil {
					return fmt.Errorf("cannot fetch latest block: %w", err)
				}
				tag := rpc.BlockNumberOrHashWithHash(l1StartBlock.Hash(), true)
				config.L1StartingBlockTag = (*genesis.MarshalableRPCBlockNumberOrHash)(&tag)
			} else if config.L1StartingBlockTag.BlockHash != nil {
				l1StartBlock, err = client.BlockByHash(context.Background(), *config.L1StartingBlockTag.BlockHash)
				if err != nil {
					return fmt.Errorf("cannot fetch block by hash: %w", err)
				}
			} else if config.L1StartingBlockTag.BlockNumber != nil {
				l1StartBlock, err = client.BlockByNumber(context.Background(), big.NewInt(config.L1StartingBlockTag.BlockNumber.Int64()))
				if err != nil {
					return fmt.Errorf("cannot fetch block by number: %w", err)
				}
			}

			// Ensure that there is a starting L1 block
			if l1StartBlock == nil {
				return fmt.Errorf("no starting L1 block")
			}

			// Sanity check the config. Do this after filling in the L1StartingBlockTag
			// if it is not defined.
			if err := config.Check(); err != nil {
				return err
			}

			log.Info("Using L1 Start Block", "number", l1StartBlock.Number(), "hash", l1StartBlock.Hash().Hex())

			// Build the L2 genesis block
			l2Genesis, err := genesis.BuildL2Genesis(config, l1StartBlock)
			if err != nil {
				return fmt.Errorf("error creating l2 genesis: %w", err)
			}

			// So far we applied changes in the memory VM and collected changes in the genesis struct
			// No we iterate through all accounts that have been written there and set them inside the statedb.
			// This will change the state root
			// TODO(pl): We should have some checks that check we don't write accounts that aren't necessary, e.g. dev account
			// Another property is that the total balance changes must be 0

			// Write changes to state to actual state database
			dbPath := ctx.String("db-path")
			if dbPath == "" {
				return fmt.Errorf("must specify --db-path")
			}
			dbCache := ctx.Int("db-cache")
			dbHandles := ctx.Int("db-handles")
			dryRun := ctx.Bool("dry-run")
			// TODO(pl): Move this into the function
			log.Info("Opening database", "dbCache", dbCache, "dbHandles", dbHandles, "dbPath", dbPath)
			ldb, err := Open(dbPath, dbCache, dbHandles)
			if err != nil {
				return fmt.Errorf("cannot open DB: %w", err)
			}
			log.Info("Loaded Celo L1 DB", "db", ldb)

			cel2Header, err := ApplyMigrationChangesToDB(ldb, l2Genesis, !dryRun)
			if err != nil {
				return err
			}

			// Close the database handle
			if err := ldb.Close(); err != nil {
				return err
			}

			log.Info("Updated Cel2 state")

			log.Info("Writing state diff", "file", outfileL2)
			// Write genesis file to check created state
			if err := jsonutil.WriteJSON(outfileL2, l2Genesis); err != nil {
				return err
			}

			rollupConfig, err := config.RollupConfig(l1StartBlock, cel2Header.Hash(), cel2Header.Number.Uint64())
			if err != nil {
				return err
			}
			if err := rollupConfig.Check(); err != nil {
				return fmt.Errorf("generated rollup config does not pass validation: %w", err)
			}

			log.Info("Writing rollup config", "file", outfileRollup)
			if err := jsonutil.WriteJSON(outfileRollup, rollupConfig); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Crit("error in migration", "err", err)
	}
	log.Info("Finished migration successfully!")
}

func ApplyMigrationChangesToDB(ldb ethdb.Database, genesis *core.Genesis, commit bool) (*types.Header, error) {
	log.Info("Migrating DB")

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
	// trieRoot := header.Root
	log.Info("Read header from database", "number", header)

	// We need to update the chain config to set the correct hardforks.
	genesisHash := rawdb.ReadCanonicalHash(ldb, 0)
	cfg := rawdb.ReadChainConfig(ldb, genesisHash)
	if cfg == nil {
		log.Crit("chain config not found")
	}
	log.Info("Read config from database", "config", cfg)

	dbFactory := func() (*state.StateDB, error) {
		// Set up the backing store.
		underlyingDB := state.NewDatabaseWithConfig(ldb, &trie.Config{
			Preimages: true,
		})

		// Open up the state database.
		db, err := state.New(header.Root, underlyingDB, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot open StateDB: %w", err)
		}

		return db, nil
	}

	db, err := dbFactory()
	if err != nil {
		return nil, fmt.Errorf("cannot create StateDB: %w", err)
	}

	for k, v := range genesis.Alloc {
		if db.Exist(k) {
			log.Warn("Operating on existing state", "account", k)
		}
		// TODO(pl): decide what to do with existing accounts.
		db.CreateAccount(k)

		db.SetNonce(k, v.Nonce)
		db.SetBalance(k, v.Balance)
		db.SetCode(k, v.Code)
		db.SetStorage(k, v.Storage)
		log.Info("Moved account", "address", k)
	}

	// We're done messing around with the database, so we can now commit the changes to the DB.
	// Note that this doesn't actually write the changes to disk.
	log.Info("Committing state DB")
	// TODO(pl): What block info to put here?
	newRoot, err := db.Commit(1234, true)
	if err != nil {
		return nil, err
	}

	log.Info("Creating new Genesis block")
	// Create the header for the Bedrock transition block.
	cel2Header := &types.Header{
		ParentHash:  header.Hash(),
		UncleHash:   types.EmptyUncleHash,
		Coinbase:    predeploys.SequencerFeeVaultAddr,
		Root:        newRoot,
		TxHash:      types.EmptyRootHash,
		ReceiptHash: types.EmptyRootHash,
		Bloom:       types.Bloom{},
		Difficulty:  new(big.Int).Set(common.Big0),
		Number:      new(big.Int).Add(header.Number, common.Big1),
		GasLimit:    header.GasLimit,
		GasUsed:     0,
		Time:        header.Time,
		Extra:       []byte("CeL2 migration"),
		MixDigest:   common.Hash{},
		Nonce:       types.BlockNonce{},
		BaseFee:     new(big.Int).Set(header.BaseFee),
	}

	// Create the Bedrock transition block from the header. Note that there are no transactions,
	// uncle blocks, or receipts in the Bedrock transition block.
	cel2Block := types.NewBlock(cel2Header, nil, nil, nil, trie.NewStackTrie(nil))

	// We did it!
	log.Info(
		"Built Celo migration block",
		"hash", cel2Block.Hash(),
		"root", cel2Block.Root(),
		"number", cel2Block.NumberU64(),
		"gas-used", cel2Block.GasUsed(),
		"gas-limit", cel2Block.GasLimit(),
	)

	log.Info("Header", "header", cel2Header)

	// If we're not actually writing this to disk, then we're done.
	if !commit {
		log.Info("Dry run complete")
		return nil, nil
	}

	// Otherwise we need to write the changes to disk. First we commit the state changes.
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

	// TODO(pl): What does this mean?
	// Make the first CeL2 block a finalized block.
	rawdb.WriteFinalizedBlockHash(ldb, cel2Block.Hash())

	// Set the standard options.
	// TODO: What about earlier hardforks, e.g. does berlin have to be enabled as it never was on Celo?
	cfg.LondonBlock = cel2Block.Number()
	cfg.ArrowGlacierBlock = cel2Block.Number()
	cfg.GrayGlacierBlock = cel2Block.Number()
	cfg.MergeNetsplitBlock = cel2Block.Number()
	cfg.TerminalTotalDifficulty = big.NewInt(0)
	cfg.TerminalTotalDifficultyPassed = true

	// Set the Optimism options.
	cfg.BedrockBlock = cel2Block.Number()
	// Enable Regolith from the start of Bedrock
	cfg.RegolithTime = new(uint64) // what are those? do we need those?
	cfg.Optimism = &params.OptimismConfig{
		EIP1559Denominator: EIP1559Denominator,
		EIP1559Elasticity:  EIP1559Elasticity,
	}
	cfg.CanyonTime = &cel2Header.Time
	cfg.EcotoneTime = &cel2Header.Time

	// TODO(pl): What about Ethereum hardforks

	log.Info("Write new config to database", "config", cfg)

	// Write the chain config to disk.
	rawdb.WriteChainConfig(ldb, cel2Block.Hash(), cfg)

	// Yay!
	log.Info(
		"Wrote chain config",
		"1559-denominator", EIP1559Denominator,
		"1559-elasticity", EIP1559Elasticity,
	)

	// We're done!
	log.Info(
		"Wrote CeL2 transition block",
		"height", cel2Header.Number,
		"root", cel2Header.Root.String(),
		"hash", cel2Header.Hash().String(),
		"timestamp", cel2Header.Time,
	)

	return cel2Header, nil
}

func Open(path string, cache int, handles int) (ethdb.Database, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	chaindataPath := filepath.Join(path, "celo", "chaindata")
	ancientPath := filepath.Join(chaindataPath, "ancient")
	ldb, err := rawdb.Open(rawdb.OpenOptions{
		Type:              "leveldb",
		Directory:         chaindataPath,
		AncientsDirectory: ancientPath,
		Namespace:         "",
		Cache:             cache,
		Handles:           handles,
		ReadOnly:          false,
	})
	if err != nil {
		return nil, err
	}
	return ldb, nil
}
