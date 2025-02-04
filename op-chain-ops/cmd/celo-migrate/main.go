package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"log/slog"

	"github.com/ethereum-optimism/optimism/op-chain-ops/foundry"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-service/ioutil"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/hashicorp/go-multierror"
	"github.com/mattn/go-isatty"

	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"golang.org/x/sync/errgroup"
)

var (
	deployConfigFlag = &cli.PathFlag{
		Name:     "deploy-config",
		Usage:    "Path to the JSON file that was used for the bedrock contracts deployment. A test example can be found here 'op-chain-ops/genesis/testdata/test-deploy-config-full.json' and documentation for the fields is at https://docs.optimism.io/builders/chain-operators/management/configuration",
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
	outfileGenesisFlag = &cli.PathFlag{
		Name:     "outfile.genesis",
		Usage:    "Path to write the genesis JSON file, to be used to sync new nodes",
		Required: true,
	}
	migrationBlockNumberFlag = &cli.Uint64Flag{
		Name:     "migration-block-number",
		Usage:    "Specifies the migration block number. If the source db is not synced exactly to the block immediately before this number (i.e. migration-block-number - 1), the migration will fail.",
		Required: true,
	}
	migrationBlockTimeFlag = &cli.Uint64Flag{
		Name:     "migration-block-time",
		Usage:    "Specifies a unix timestamp to use for the migration block. This should be set to the same timestamp as was used for the sequencer migration. If performing the sequencer migration, this should set to a time in the future around when the migration script is expected to complete.",
		Required: true,
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
		Value: 50000,
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
	reset = &cli.BoolFlag{
		Name:  "reset",
		Usage: "Delete everything in the destination directory aside from /ancients. This is useful if you need to re-run the full migration but do not want to repeat the lengthy ancients migration. If you'd like to reset the entire destination directory, you can delete it manually.",
		Value: false,
	}
	dbCheckPathFlag = &cli.PathFlag{
		Name:     "db-path",
		Usage:    "Path to the db to perform a continuity check on",
		Required: true,
	}
	dbCheckFailFastFlag = &cli.BoolFlag{
		Name:  "fail-fast",
		Usage: "Fail fast on the first error encountered. If set, the db check will stop on the first error encountered, otherwise it will continue to check all blocks and print out all errors at the end.",
		Value: false,
	}

	preMigrationFlags = []cli.Flag{
		oldDBPathFlag,
		newDBPathFlag,
		batchSizeFlag,
		bufferSizeFlag,
		memoryLimitFlag,
		reset,
	}
	fullMigrationFlags = append(
		preMigrationFlags,
		deployConfigFlag,
		l1DeploymentsFlag,
		l1RPCFlag,
		l2AllocsFlag,
		outfileRollupConfigFlag,
		outfileGenesisFlag,
		migrationBlockTimeFlag,
		migrationBlockNumberFlag,
	)
	dbCheckFlags = []cli.Flag{
		dbCheckPathFlag,
		batchSizeFlag,
		dbCheckFailFastFlag,
	}
)

type preMigrationOptions struct {
	oldDBPath        string
	newDBPath        string
	batchSize        uint64
	bufferSize       uint64
	memoryLimit      int64
	resetNonAncients bool
}

type stateMigrationOptions struct {
	deployConfig        string
	l1Deployments       string
	l1RPC               string
	l2AllocsPath        string
	outfileRollupConfig string
	outfileGenesis      string
	migrationBlockTime  uint64
}

type fullMigrationOptions struct {
	preMigrationOptions
	stateMigrationOptions
	migrationBlockNumber uint64
}

type dbCheckOptions struct {
	dbPath    string
	batchSize uint64
	failFast  bool
}

func parsePreMigrationOptions(ctx *cli.Context) preMigrationOptions {
	return preMigrationOptions{
		oldDBPath:        ctx.String(oldDBPathFlag.Name),
		newDBPath:        ctx.String(newDBPathFlag.Name),
		batchSize:        ctx.Uint64(batchSizeFlag.Name),
		bufferSize:       ctx.Uint64(bufferSizeFlag.Name),
		memoryLimit:      ctx.Int64(memoryLimitFlag.Name),
		resetNonAncients: ctx.Bool(reset.Name),
	}
}

func parseStateMigrationOptions(ctx *cli.Context) stateMigrationOptions {
	return stateMigrationOptions{
		deployConfig:        ctx.Path(deployConfigFlag.Name),
		l1Deployments:       ctx.Path(l1DeploymentsFlag.Name),
		l1RPC:               ctx.String(l1RPCFlag.Name),
		l2AllocsPath:        ctx.Path(l2AllocsFlag.Name),
		outfileRollupConfig: ctx.Path(outfileRollupConfigFlag.Name),
		outfileGenesis:      ctx.Path(outfileGenesisFlag.Name),
		migrationBlockTime:  ctx.Uint64(migrationBlockTimeFlag.Name),
	}
}

func parseFullMigrationOptions(ctx *cli.Context) fullMigrationOptions {
	return fullMigrationOptions{
		preMigrationOptions:   parsePreMigrationOptions(ctx),
		stateMigrationOptions: parseStateMigrationOptions(ctx),
		migrationBlockNumber:  ctx.Uint64(migrationBlockNumberFlag.Name),
	}
}

func parseDBCheckOptions(ctx *cli.Context) dbCheckOptions {
	return dbCheckOptions{
		dbPath:    ctx.String(dbCheckPathFlag.Name),
		batchSize: ctx.Uint64(batchSizeFlag.Name),
		failFast:  ctx.Bool(dbCheckFailFastFlag.Name),
	}
}

func main() {

	color := isatty.IsTerminal(os.Stderr.Fd())
	handler := log.NewTerminalHandlerWithLevel(os.Stderr, slog.LevelInfo, color)
	oplog.SetGlobalLogHandler(handler)

	app := &cli.App{
		Name:  "celo-migrate",
		Usage: "Migrate Celo block and state data to a CeL2 DB",
		Commands: []*cli.Command{
			{
				Name:  "pre",
				Usage: "Perform a  pre-migration of ancient blocks and copy over all other data without transforming it. This should be run a day before the full migration command is run to minimize downtime.",
				Flags: preMigrationFlags,
				Action: func(ctx *cli.Context) error {
					if _, _, err := runPreMigration(parsePreMigrationOptions(ctx)); err != nil {
						return fmt.Errorf("failed to run pre-migration: %w", err)
					}
					log.Info("Finished pre migration successfully!")
					return nil
				},
			},
			{
				Name:  "full",
				Usage: "Perform a full migration of both block and state data to a CeL2 DB",
				Flags: fullMigrationFlags,
				Action: func(ctx *cli.Context) error {
					if err := runFullMigration(parseFullMigrationOptions(ctx)); err != nil {
						return fmt.Errorf("failed to run full migration: %w", err)
					}
					log.Info("Finished full migration successfully!")
					return nil
				},
			},
			{
				Name:  "check-db",
				Usage: "Perform a continuity check on the db, ensuring that all blocks are present and in order",
				Flags: dbCheckFlags,
				Action: func(ctx *cli.Context) error {
					if err := runDBCheck(parseDBCheckOptions(ctx)); err != nil {
						return fmt.Errorf("DB continuity check failed: %w", err)
					}
					log.Info("Finished db continuity check successfully!")
					return nil
				},
			},
		},
		OnUsageError: func(ctx *cli.Context, err error, isSubcommand bool) error {
			if isSubcommand {
				return err
			}
			if err := cli.ShowAppHelp(ctx); err != nil {
				log.Error("failed to show cli help", "err", err)
			}
			return fmt.Errorf("please provide a valid command")
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Crit("error in celo-migrate", "err", err)
	}
}

func runFullMigration(opts fullMigrationOptions) error {
	defer timer("full migration")()

	log.Info("Full Migration Started", "oldDBPath", opts.oldDBPath, "newDBPath", opts.newDBPath)

	head, err := getHeadHeader(opts.oldDBPath)
	if err != nil {
		return fmt.Errorf("failed to get head header: %w", err)
	}
	if head.Number.Uint64() != opts.migrationBlockNumber-1 {
		return fmt.Errorf("old-db head block number not synced to the block immediately before the migration block number: %d != %d", head.Number.Uint64(), opts.migrationBlockNumber-1)
	}

	log.Info("Source db is synced to correct height", "head", head.Number.Uint64(), "migrationBlock", opts.migrationBlockNumber)

	var numAncients uint64
	var strayAncientBlocks []*rawdb.NumberHash

	if strayAncientBlocks, numAncients, err = runPreMigration(opts.preMigrationOptions); err != nil {
		return fmt.Errorf("failed to run pre-migration: %w", err)
	}

	if err = runNonAncientMigration(opts.newDBPath, strayAncientBlocks, opts.batchSize, numAncients); err != nil {
		return fmt.Errorf("failed to run non-ancient migration: %w", err)
	}
	if err = runStateMigration(opts.newDBPath, opts.stateMigrationOptions); err != nil {
		return fmt.Errorf("failed to run state migration: %w", err)
	}

	log.Info("Full Migration Finished", "oldDBPath", opts.oldDBPath, "newDBPath", opts.newDBPath)

	return nil
}

func runPreMigration(opts preMigrationOptions) ([]*rawdb.NumberHash, uint64, error) {
	defer timer("pre-migration")()

	log.Info("Pre-Migration Started", "oldDBPath", opts.oldDBPath, "newDBPath", opts.newDBPath, "batchSize", opts.batchSize, "memoryLimit", opts.memoryLimit)

	// Check that `rsync` command is available. We use this to copy the db excluding ancients, which we will copy separately
	if _, err := exec.LookPath("rsync"); err != nil {
		return nil, 0, fmt.Errorf("please install `rsync` to run block migration")
	}

	debug.SetMemoryLimit(opts.memoryLimit * 1 << 20) // Set memory limit, converting from MiB to bytes

	var err error

	if err = createNewDbPathIfNotExists(opts.newDBPath); err != nil {
		return nil, 0, fmt.Errorf("failed to create new db path: %w", err)
	}

	if opts.resetNonAncients {
		if err = cleanupNonAncientDb(opts.newDBPath); err != nil {
			return nil, 0, fmt.Errorf("failed to cleanup non-ancient db: %w", err)
		}
	}

	var numAncientsNewBefore uint64
	var numAncientsNewAfter uint64
	var strayAncientBlocks []*rawdb.NumberHash
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		if numAncientsNewBefore, numAncientsNewAfter, err = migrateAncientsDb(ctx, opts.oldDBPath, opts.newDBPath, opts.batchSize, opts.bufferSize); err != nil {
			return fmt.Errorf("failed to migrate ancients database: %w", err)
		}
		// Scanning for stray ancient blocks is slow, so we do it as soon as we can after the lock on oldDB is released by migrateAncientsDb
		// Doing this in parallel with copyDbExceptAncients still saves time if ancients have already been pre-migrated
		if strayAncientBlocks, err = getStrayAncientBlocks(opts.oldDBPath); err != nil {
			return fmt.Errorf("failed to get stray ancient blocks: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		// By doing this once during the premigration, we get a speedup when we run it again in a full migration.
		return copyDbExceptAncients(opts.oldDBPath, opts.newDBPath)
	})

	if err = g.Wait(); err != nil {
		return nil, 0, fmt.Errorf("failed to migrate blocks: %w", err)
	}

	log.Info("Pre-Migration Finished", "oldDBPath", opts.oldDBPath, "newDBPath", opts.newDBPath, "migratedAncients", numAncientsNewAfter-numAncientsNewBefore, "strayAncientBlocks", len(strayAncientBlocks))

	return strayAncientBlocks, numAncientsNewAfter, nil
}

func runNonAncientMigration(newDBPath string, strayAncientBlocks []*rawdb.NumberHash, batchSize, numAncients uint64) (err error) {
	defer timer("non-ancient migration")()

	newDB, err := openDBWithoutFreezer(newDBPath, false)
	if err != nil {
		return fmt.Errorf("failed to open new database: %w", err)
	}
	defer func() {
		err = errors.Join(err, newDB.Close())
	}()

	// get the last block number
	hash := rawdb.ReadHeadHeaderHash(newDB)
	lastBlock := *rawdb.ReadHeaderNumber(newDB, hash)
	lastAncient := numAncients - 1

	log.Info("Non-Ancient Block Migration Started", "process", "non-ancients", "newDBPath", newDBPath, "batchSize", batchSize, "startBlock", numAncients, "endBlock", lastBlock, "count", lastBlock-lastAncient, "lastAncientBlock", lastAncient)

	var numNonAncients uint64
	if numNonAncients, err = migrateNonAncientsDb(newDB, lastBlock, numAncients, batchSize); err != nil {
		return fmt.Errorf("failed to migrate non-ancients database: %w", err)
	}

	err = removeBlocks(newDB, strayAncientBlocks)
	if err != nil {
		return fmt.Errorf("failed to remove stray ancient blocks: %w", err)
	}
	log.Info("Removed stray ancient blocks still in leveldb", "process", "non-ancients", "removedBlocks", len(strayAncientBlocks))

	log.Info("Non-Ancient Block Migration Completed", "process", "non-ancients", "migratedNonAncients", numNonAncients)

	return nil
}

func runStateMigration(newDBPath string, opts stateMigrationOptions) error {
	defer timer("state migration")()

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
	if err := config.Check(log.New()); err != nil {
		return err
	}

	log.Info("Using L1 Start Block", "number", l1StartBlock.Number(), "hash", l1StartBlock.Hash().Hex())

	// Build the L2 genesis block
	l2Allocs, err := foundry.LoadForgeAllocs(opts.l2AllocsPath)
	if err != nil {
		return err
	}

	l2Genesis, err := genesis.BuildL2Genesis(config, l2Allocs, l1StartBlock.Header())
	if err != nil {
		return fmt.Errorf("error creating l2 genesis: %w", err)
	}

	// Write changes to state to actual state database
	cel2Header, err := applyStateMigrationChanges(config, l2Genesis.Alloc, newDBPath, opts.outfileGenesis, opts.migrationBlockTime, l1StartBlock)
	if err != nil {
		return err
	}
	log.Info("Updated Cel2 state")

	rollupConfig, err := config.RollupConfig(l1StartBlock.Header(), cel2Header.Hash(), cel2Header.Number.Uint64())
	if err != nil {
		return err
	}
	if err := rollupConfig.Check(); err != nil {
		return fmt.Errorf("generated rollup config does not pass validation: %w", err)
	}

	log.Info("Writing rollup config", "file", opts.outfileRollupConfig)
	if err := jsonutil.WriteJSON(rollupConfig, ioutil.ToStdOutOrFileOrNoop(opts.outfileRollupConfig, OutFilePerm)); err != nil {
		return err
	}

	log.Info("State Migration Completed")

	return nil
}

func runDBCheck(opts dbCheckOptions) (err error) {
	defer timer("db continuity check")()

	log.Info("DB Continuity Check Started", "dbPath", opts.dbPath)

	ancientDB, err := NewChainFreezer(filepath.Join(opts.dbPath, "ancient"), "", true)
	if err != nil {
		return fmt.Errorf("failed to open ancient db: %w", err)
	}
	defer func() {
		err = errors.Join(err, ancientDB.Close())
	}()
	nonAncientDB, err := openDBWithoutFreezer(opts.dbPath, true)
	if err != nil {
		return fmt.Errorf("failed to open non-ancient db: %w", err)
	}
	defer func() {
		err = errors.Join(err, nonAncientDB.Close())
	}()

	lastAncient, err := loadLastAncient(ancientDB)
	if err != nil {
		return fmt.Errorf("failed to load last ancient block: %w", err)
	}
	lastAncientNumber := lastAncient.Number()
	lastBlockNumber := *rawdb.ReadHeaderNumber(nonAncientDB, rawdb.ReadHeadHeaderHash(nonAncientDB))

	var errResult *multierror.Error

	// First, check continuity between ancients and non-ancients.
	// Gaps in data will often halt the freezing process, so attempting to load the first non-ancient block
	// will most likely fail if there is a gap.
	firstNonAncientRange, err := loadNonAncientRange(nonAncientDB, lastAncientNumber+1, 1)
	if err != nil {
		if opts.failFast {
			return fmt.Errorf("failed to load first non-ancient block: %w", err)
		}
		// We don't need to add the error to errResult here because it will be added below when we call checkContinuity on non-ancients
	} else {
		if _, err := firstNonAncientRange.CheckContinuity(lastAncient, 1); err != nil {
			err = fmt.Errorf("failed continuity check between ancients and non-ancients: %w", err)
			if opts.failFast {
				return err
			}
			errResult = multierror.Append(errResult, err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	// Use double the CPUs to account for the fact that each routine will block while reading from the db.
	g.SetLimit(runtime.NumCPU() * 2)

	var mu sync.Mutex

	checkRange := func(start, count uint64, loadRangeFunc func(uint64, uint64) (*RLPBlockRange, error)) {
		// If we are not at genesis or the first non-ancient block, include the last block of
		// the previous range so we can check for continuity between ranges.
		if start != 0 && start != lastAncientNumber+1 {
			start = start - 1
			count = count + 1
		}
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				blockRange, err := loadRangeFunc(start, count)
				if err != nil {
					err = fmt.Errorf("failed to load block range: %w", err)
					if opts.failFast {
						return err
					}
					log.Error(err.Error())
					mu.Lock()
					errResult = multierror.Append(errResult, err)
					mu.Unlock()
					return nil
				}
				if _, err := blockRange.CheckContinuity(nil, count); err != nil {
					err = fmt.Errorf("failed continuity check: %w", err)
					if opts.failFast {
						return err
					}
					log.Error(err.Error())
					mu.Lock()
					errResult = multierror.Append(errResult, err)
					mu.Unlock()
					return nil
				}
				log.Info("Successfully checked block range continuity", "start", start, "end", start+count-1, "count", count)
				return nil
			}
		})
	}
	checkContinuity := func(start, end uint64, loadRangeFunc func(uint64, uint64) (*RLPBlockRange, error)) error {
		if (start <= lastAncientNumber && end > lastAncientNumber) || (end > lastBlockNumber) || (end < start) {
			return fmt.Errorf("invalid range for continuity check: start=%d, end=%d, lastAncientNumber=%d, lastBlockNumber=%d", start, end, lastAncientNumber, lastBlockNumber)
		}
		for i := start; i <= end; i += opts.batchSize {
			count := min(opts.batchSize, end-i+1)
			checkRange(i, count, loadRangeFunc)
		}
		return nil
	}

	log.Info("Checking continuity of ancient blocks", "start", 0, "end", lastAncientNumber, "count", lastAncientNumber+1)
	if err := checkContinuity(0, lastAncientNumber, func(start, count uint64) (*RLPBlockRange, error) {
		return loadAncientRange(ancientDB, start, count)
	}); err != nil {
		return err
	}
	log.Info("Checking continuity of non-ancient blocks", "start", lastAncientNumber+1, "end", lastBlockNumber, "count", lastBlockNumber-lastAncientNumber)
	if err := checkContinuity(lastAncientNumber+1, lastBlockNumber, func(start, count uint64) (*RLPBlockRange, error) {
		return loadNonAncientRange(nonAncientDB, start, count)
	}); err != nil {
		return err
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if errResult.ErrorOrNil() != nil {
		return errResult
	}

	log.Info("DB Continuity Check Finished", "dbPath", opts.dbPath)

	return nil
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		log.Info("TIMER", "process", name, "duration", time.Since(start))
	}
}
