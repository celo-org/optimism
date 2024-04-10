package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/mattn/go-isatty"

	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
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

			l2GenesisBlock := l2Genesis.ToBlock()
			rollupConfig, err := config.RollupConfig(l1StartBlock, l2GenesisBlock.Hash(), l2GenesisBlock.Number().Uint64())
			if err != nil {
				return err
			}
			if err := rollupConfig.Check(); err != nil {
				return fmt.Errorf("generated rollup config does not pass validation: %w", err)
			}

			outfileL2 := ctx.Path("outfile.l2")
			if outfileL2 == "" {
				return fmt.Errorf("must specify --outfile.l2")
			}

			outfileRollup := ctx.Path("outfile.rollup")
			if outfileRollup == "" {
				return fmt.Errorf("must specify --outfile.rollup")
			}

			if err := jsonutil.WriteJSON(outfileL2, l2Genesis); err != nil {
				return err
			}
			if err := jsonutil.WriteJSON(outfileRollup, rollupConfig); err != nil {
				return err
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
			// dryRun := ctx.Bool("dry-run")
			// TODO(pl): Move this into the function
			log.Info("Opening database", "dbCache", dbCache, "dbHandles", dbHandles, "dbPath", dbPath)
			ldb, err := Open(dbPath, dbCache, dbHandles)
			if err != nil {
				return fmt.Errorf("cannot open DB: %w", err)
			}

			if err := ApplyMigrationChangesToDB(ldb, l2Genesis); err != nil {
				return err
			}

			// Close the database handle
			if err := ldb.Close(); err != nil {
				return err
			}

			log.Info("Loaded Celo L1 DB", "db", ldb)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Crit("error in migration", "err", err)
	}
	log.Info("Finished migration successfully!")
}

func ApplyMigrationChangesToDB(ldb ethdb.Database, genesis *core.Genesis) error {
	panic("unimplemented")
}

func Open(path string, cache int, handles int) (ethdb.Database, error) {
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
