package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"runtime/debug"
	"time"

	"github.com/ethereum-optimism/optimism/op-chain-ops/foundry"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

var (
	deployConfigFlag = &cli.PathFlag{
		Name:     "deploy-config",
		Usage:    "Path to the JSON file that was used for the l1 contracts deployment. A test example can be found here 'op-chain-ops/genesis/testdata/test-deploy-config-full.json' and documentation for the fields is at https://docs.optimism.io/builders/chain-operators/management/configuration",
		Required: true,
	}
	l1DeploymentsFlag = &cli.PathFlag{
		Name:     "l1-deployments",
		Usage:    "Path to L1 deployments JSON file, the output of running the bedrock contracts deployment for the given 'deploy-config'",
		Required: true,
	}
	l1RPCFlag = &cli.StringFlag{
		Name:     "l1-rpc",
		Usage:    "RPC URL for a node of the L1 defined in the 'deploy-config'",
		Required: true,
	}
	l2AllocsFlag = &cli.PathFlag{
		Name:     "l2-allocs",
		Usage:    "Path to L2 genesis allocs file. You can find instructions on how to generate this file in the README",
		Required: true,
	}
	outfileRollupConfigFlag = &cli.PathFlag{
		Name:     "outfile.rollup-config",
		Usage:    "Path to write the rollup config JSON file, to be provided to op-node with the 'rollup.config' flag",
		Required: true,
	}
	migrationBlockTimeFlag = &cli.Uint64Flag{
		Name:  "migration-block-time",
		Usage: "Specifies a unix timestamp to use for the migration block. If not provided, the current time will be used.",
	}
	oldDBPathFlag = &cli.PathFlag{
		Name:     "old-db",
		Usage:    "Path to the old Celo chaindata dir, can be found at '<datadir>/celo/chaindata'",
		Required: true,
	}
	newDBPathFlag = &cli.PathFlag{
		Name:     "new-db",
		Usage:    "Path to write migrated Celo chaindata, note the new node implementation expects to find this chaindata at the following path '<datadir>/geth/chaindata",
		Required: true,
	}
	batchSizeFlag = &cli.Uint64Flag{
		Name:  "batch-size",
		Usage: "Batch size to use for block migration, larger batch sizes can speed up migration but require more memory. If increasing the batch size consider also increasing the memory-limit",
		Value: 50000, // TODO(Alec) optimize default parameters
	}
	bufferSizeFlag = &cli.Uint64Flag{
		Name:  "buffer-size",
		Usage: "Buffer size to use for ancient block migration channels. Defaults to 0. Included to facilitate testing for performance improvements.",
		Value: 0,
	}
	memoryLimitFlag = &cli.Int64Flag{
		Name:  "memory-limit",
		Usage: "Memory limit in MiB, should be set lower than the available amount of memory in your system to prevent out of memory errors",
		Value: 7500,
	}
	clearNonAncientsFlag = &cli.BoolFlag{
		Name:  "clear-non-ancients",
		Usage: "Use this to reset all data except ancients. This flag should be used if a full migration has already been performed on the new db.",
	}
	measureTimeFlag = &cli.BoolFlag{
		Name:  "measure-time",
		Usage: "Use this to log how long each section of the script takes to run",
	}

	preMigrationFlags = []cli.Flag{
		oldDBPathFlag,
		newDBPathFlag,
		batchSizeFlag,
		bufferSizeFlag,
		memoryLimitFlag,
		clearNonAncientsFlag,
		measureTimeFlag,
	}
	fullMigrationFlags = append(
		preMigrationFlags,
		deployConfigFlag,
		l1DeploymentsFlag,
		l1RPCFlag,
		l2AllocsFlag,
		outfileRollupConfigFlag,
		migrationBlockTimeFlag,
	)
)

type preMigrationOptions struct {
	oldDBPath        string
	newDBPath        string
	batchSize        uint64
	bufferSize       uint64
	memoryLimit      int64
	clearNonAncients bool
	measureTime      bool
}

type stateMigrationOptions struct {
	deployConfig        string
	l1Deployments       string
	l1RPC               string
	l2AllocsPath        string
	outfileRollupConfig string
	migrationBlockTime  uint64
}

type fullMigrationOptions struct {
	preMigrationOptions
	stateMigrationOptions
}

func parsePreMigrationOptions(ctx *cli.Context) preMigrationOptions {
	return preMigrationOptions{
		oldDBPath:        ctx.String(oldDBPathFlag.Name),
		newDBPath:        ctx.String(newDBPathFlag.Name),
		batchSize:        ctx.Uint64(batchSizeFlag.Name),
		bufferSize:       ctx.Uint64(bufferSizeFlag.Name),
		memoryLimit:      ctx.Int64(memoryLimitFlag.Name),
		clearNonAncients: ctx.Bool(clearNonAncientsFlag.Name),
		measureTime:      ctx.Bool(measureTimeFlag.Name),
	}
}

func parseStateMigrationOptions(ctx *cli.Context) stateMigrationOptions {
	return stateMigrationOptions{
		deployConfig:        ctx.Path(deployConfigFlag.Name),
		l1Deployments:       ctx.Path(l1DeploymentsFlag.Name),
		l1RPC:               ctx.String(l1RPCFlag.Name),
		l2AllocsPath:        ctx.Path(l2AllocsFlag.Name),
		outfileRollupConfig: ctx.Path(outfileRollupConfigFlag.Name),
		migrationBlockTime:  ctx.Uint64(migrationBlockTimeFlag.Name),
	}
}

func parseFullMigrationOptions(ctx *cli.Context) fullMigrationOptions {
	return fullMigrationOptions{
		preMigrationOptions:   parsePreMigrationOptions(ctx),
		stateMigrationOptions: parseStateMigrationOptions(ctx),
	}
}

func main() {

	color := isatty.IsTerminal(os.Stderr.Fd())
	handler := log.NewTerminalHandlerWithLevel(os.Stderr, slog.LevelDebug, color)
	oplog.SetGlobalLogHandler(handler)

	log.Info("Beginning Cel2 Migration")

	app := &cli.App{
		Name:  "celo-migrate",
		Usage: "Migrate Celo block and state data to a CeL2 DB",
		Commands: []*cli.Command{
			{
				Name:    "pre",
				Aliases: []string{"pre-migration", "p"},
				Usage:   "Perform a  pre-migration of ancient blocks and copy over all other data without transforming it. This should be run a day before the full migration command is run to minimize downtime.",
				Flags:   preMigrationFlags,
				Action: func(ctx *cli.Context) error {
					if _, err := runPreMigration(parsePreMigrationOptions(ctx)); err != nil {
						return fmt.Errorf("failed to run pre-migration: %w", err)
					}
					return nil
				},
			},
			{
				Name:    "final",
				Aliases: []string{"full", "f"},
				Usage:   "Perform a full migration of both block and state data to a CeL2 DB",
				Flags:   fullMigrationFlags,
				Action: func(ctx *cli.Context) error {
					if err := runFullMigration(parseFullMigrationOptions(ctx)); err != nil {
						return fmt.Errorf("failed to run full migration: %w", err)
					}
					return nil
				},
			},
		},
		OnUsageError: func(ctx *cli.Context, err error, isSubcommand bool) error {
			if isSubcommand {
				return err
			}
			_ = cli.ShowAppHelp(ctx)
			return fmt.Errorf("please provide a valid command")
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Crit("error in migration", "err", err)
	}
	log.Info("Finished migration successfully!")
}

func runFullMigration(opts fullMigrationOptions) error {
	if opts.measureTime {
		defer timer("full migration")()
	}

	log.Info("Full Migration Started", "oldDBPath", opts.oldDBPath, "newDBPath", opts.newDBPath)

	var err error

	var numAncients uint64
	if numAncients, err = runPreMigration(opts.preMigrationOptions); err != nil {
		return fmt.Errorf("failed to run pre-migration: %w", err)
	}

	// TODO(Alec) can these be parallelized?
	if err = runNonAncientMigration(opts.newDBPath, opts.measureTime, opts.batchSize, numAncients); err != nil {
		return fmt.Errorf("failed to run non-ancient migration: %w", err)
	}
	if err := runStateMigration(opts.newDBPath, opts.measureTime, opts.stateMigrationOptions); err != nil {
		return fmt.Errorf("failed to run state migration: %w", err)
	}

	log.Info("Full Migration Finished", "oldDBPath", opts.oldDBPath, "newDBPath", opts.newDBPath)

	return nil
}

func runPreMigration(opts preMigrationOptions) (uint64, error) {
	if opts.measureTime {
		defer timer("pre-migration")()
	}

	log.Info("Pre-Migration Started", "oldDBPath", opts.oldDBPath, "newDBPath", opts.newDBPath, "batchSize", opts.batchSize, "memoryLimit", opts.memoryLimit, "clearNonAncients", opts.clearNonAncients)

	// Check that `rsync` command is available. We use this to copy the db excluding ancients, which we will copy separately
	if _, err := exec.LookPath("rsync"); err != nil {
		return 0, fmt.Errorf("please install `rsync` to run block migration")
	}

	debug.SetMemoryLimit(opts.memoryLimit * 1 << 20) // Set memory limit, converting from MiB to bytes

	var err error

	if err = createNewDbPathIfNotExists(opts.newDBPath); err != nil {
		return 0, fmt.Errorf("failed to create new db path: %w", err)
	}

	var numAncientsNewBefore uint64
	var numAncientsNewAfter uint64

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		if numAncientsNewBefore, numAncientsNewAfter, err = migrateAncientsDb(ctx, opts.oldDBPath, opts.newDBPath, opts.batchSize, opts.bufferSize, opts.measureTime); err != nil {
			return fmt.Errorf("failed to migrate ancients database: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		// By doing this once during the premigration, we get a speedup when we run it again in a full migration.
		return copyDbExceptAncients(opts.oldDBPath, opts.newDBPath, opts.measureTime)
	})

	if err = g.Wait(); err != nil {
		return numAncientsNewAfter, fmt.Errorf("failed to migrate blocks: %w", err)
	}

	log.Info("Pre-Migration Finished", "oldDBPath", opts.oldDBPath, "newDBPath", opts.newDBPath, "migratedAncients", numAncientsNewAfter-numAncientsNewBefore)

	return numAncientsNewAfter, nil
}

func runNonAncientMigration(newDBPath string, measureTime bool, batchSize, numAncients uint64) error {
	if measureTime {
		defer timer("non-ancient migration")()
	}

	newDB, err := openDBWithoutFreezer(newDBPath, false)
	if err != nil {
		return fmt.Errorf("failed to open new database: %w", err)
	}
	defer newDB.Close()

	// The non-ancient block migration makes transformations to the database that would be corrupted by subsequent migrations.
	// We need to mark that a full migration has been attempted so that we can reset the database if the script is run again.
	if err = writeFullMigrationMarker(newDB); err != nil {
		return fmt.Errorf("failed to write full migration attempt marker: %w", err)
	}

	// get the last block number
	hash := rawdb.ReadHeadHeaderHash(newDB)
	lastBlock := *rawdb.ReadHeaderNumber(newDB, hash)
	lastAncient := numAncients - 1

	log.Info("Non-Ancient Block Migration Started", "process", "non-ancients", "newDBPath", newDBPath, "batchSize", batchSize, "startBlock", numAncients, "endBlock", lastBlock, "count", lastBlock-lastAncient, "lastAncientBlock", lastAncient)

	var numNonAncients uint64
	if numNonAncients, err = migrateNonAncientsDb(newDB, lastBlock, numAncients, batchSize); err != nil {
		return fmt.Errorf("failed to migrate non-ancients database: %w", err)
	}

	log.Info("Non-Ancient Block Migration Completed", "process", "non-ancients", "migratedNonAncients", numNonAncients)

	return nil
}

func runStateMigration(newDBPath string, measureTime bool, opts stateMigrationOptions) error {
	if measureTime {
		defer timer("state migration")()
	}

	log.Info("State Migration Started", "newDBPath", newDBPath, "deployConfig", opts.deployConfig, "l1Deployments", opts.l1Deployments, "l1RPC", opts.l1RPC, "l2AllocsPath", opts.l2AllocsPath, "outfileRollupConfig", opts.outfileRollupConfig)

	// Read deployment configuration
	config, err := genesis.NewDeployConfig(opts.deployConfig)
	if err != nil {
		return err
	}

	if config.DeployCeloContracts {
		return errors.New("DeployCeloContracts is not supported in migration")
	}
	if config.FundDevAccounts {
		return errors.New("FundDevAccounts is not supported in migration")
	}

	// Try reading the L1 deployment information
	deployments, err := genesis.NewL1Deployments(opts.l1Deployments)
	if err != nil {
		return fmt.Errorf("cannot read L1 deployments at %s: %w", opts.l1Deployments, err)
	}
	config.SetDeployments(deployments)

	// Get latest block information from L1
	var l1StartBlock *types.Block
	client, err := ethclient.Dial(opts.l1RPC)
	if err != nil {
		return fmt.Errorf("cannot dial %s: %w", opts.l1RPC, err)
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
	l2Allocs, err := foundry.LoadForgeAllocs(opts.l2AllocsPath)
	if err != nil {
		return err
	}

	l2Genesis, err := genesis.BuildL2Genesis(config, l2Allocs, l1StartBlock)
	if err != nil {
		return fmt.Errorf("error creating l2 genesis: %w", err)
	}

	// Write changes to state to actual state database
	cel2Header, err := applyStateMigrationChanges(config, l2Genesis, newDBPath, opts.migrationBlockTime)
	if err != nil {
		return err
	}
	log.Info("Updated Cel2 state")

	rollupConfig, err := config.RollupConfig(l1StartBlock, cel2Header.Hash(), cel2Header.Number.Uint64())
	if err != nil {
		return err
	}
	if err := rollupConfig.Check(); err != nil {
		return fmt.Errorf("generated rollup config does not pass validation: %w", err)
	}

	log.Info("Writing rollup config", "file", opts.outfileRollupConfig)
	if err := jsonutil.WriteJSON(opts.outfileRollupConfig, rollupConfig, OutFilePerm); err != nil {
		return err
	}

	log.Info("State Migration Completed")

	return nil
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		log.Info("TIMER", "process", name, "duration", time.Since(start))
	}
}
