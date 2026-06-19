package environment

import (
	"math/big"
	"time"

	bss "github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

// L2TxWithAmount is a helper.TxOptsFn that sets the Amount of the transaction.
func L2TxWithAmount(amount *big.Int) helpers.TxOptsFn {
	return func(opts *helpers.TxOpts) {
		opts.Value = amount
	}
}

// L2TxWithNonce is a helper.TxOptsFn that sets the Nonce of the transaction.
func L2TxWithNonce(nonce uint64) helpers.TxOptsFn {
	return func(opts *helpers.TxOpts) {
		opts.Nonce = nonce
	}
}

// L2WithToAddress is a helper.TxOptsFn that sets the To address of the
// transaction.
func L2TxWithToAddress(toAddr *common.Address) helpers.TxOptsFn {
	return func(opts *helpers.TxOpts) {
		opts.ToAddr = toAddr
	}
}

// L2TxWithVerifyOnClients is a helper.TxOptsFn that sets the list of
// verification clients to verify the transaction on.
func L2TxWithVerifyOnClients(clients ...*ethclient.Client) helpers.TxOptsFn {
	return func(opts *helpers.TxOpts) {
		opts.VerifyOnClients(clients...)
	}
}

// L2TxWithOptions is a helper.TxOptsFn that sets multiple options for the
// transaction. By default the L2 transaction helper function is only able
// to accept a single helpers.TxOptsFn, so this function allows multiple
// to be passed as a single option, allowing for more granular configuration
// options.
func L2TxWithOptions(options ...helpers.TxOptsFn) helpers.TxOptsFn {
	return func(opts *helpers.TxOpts) {
		for _, option := range options {
			option(opts)
		}
	}
}

// WithSequencerUseFinalized is a E2eDevnetLauncherOption that configures the sequencer's
// `SequencerUseFinalized` option to the provided value.
func WithSequencerUseFinalized(useFinalized bool) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			SystemConfigOption: func(cfg *e2esys.SystemConfig) {
				seqConfig := cfg.Nodes[e2esys.RoleSeq]
				seqConfig.Driver.SequencerUseFinalized = useFinalized
			},
		}
	}
}

// WithNonFinalizedProposals is a E2eDevnetLauncherOption that configures the system's
// `NonFinalizedProposals` option to the provided value.
func WithNonFinalizedProposals(useNonFinalized bool) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			SystemConfigOption: func(cfg *e2esys.SystemConfig) {
				cfg.NonFinalizedProposals = useNonFinalized
			},
		}
	}
}

// WithL1FinalizedDistance is a E2eDevnetLauncherOption that configures the system's
// `L1FinalizedDistance` option to the provided value.
func WithL1FinalizedDistance(distance uint64) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			SystemConfigOption: func(cfg *e2esys.SystemConfig) {
				cfg.L1FinalizedDistance = distance
			},
		}
	}
}

// WithSeqWindowSize is a E2eDevnetLauncherOption that configures the deployment's
// `SequencerWindowSize` option to the provided value.
func WithSequencerWindowSize(size uint64) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			SystemConfigOption: func(cfg *e2esys.SystemConfig) {
				cfg.DeployConfig.SequencerWindowSize = size
			},
		}
	}
}

// WithL1BlockTime is a E2eDevnetLauncherOption that configures the system's
// `L1BlockTime` option to the provided value.
//
// The passed block time should be on the order of seconds.  Any sub-second
// resolution will be lost.  The value **MUST** be at least 1 second or greater.
func WithL1BlockTime(blockTime time.Duration) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			SystemConfigOption: func(cfg *e2esys.SystemConfig) {
				cfg.DeployConfig.L1BlockTime = uint64(blockTime / time.Second)
			},
		}
	}
}

// WithL2BlockTime is a E2eDevnetLauncherOption that configures the system's
// `L2BlockTime` option to the provided value.
//
// The passed block time should be on the order of seconds.  Any sub-second
// resolution will be lost.  The value **MUST** be at least 1 second or greater.
func WithL2BlockTime(blockTime time.Duration) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			SystemConfigOption: func(cfg *e2esys.SystemConfig) {
				cfg.DeployConfig.L2BlockTime = uint64(blockTime / time.Second)
			},
		}
	}
}

// WithEspressoEnforcementOffset is an E2eDevnetLauncherOption that activates the
// EspressoEnforcement hardfork at the given offset (in seconds; sub-second
// resolution is dropped) after L1 genesis. By default the Espresso devnet
// activates the hardfork at genesis (offset = 0); use this helper to test
// pre-fork behavior of the chain.
func WithEspressoEnforcementOffset(offset time.Duration) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			SystemConfigOption: func(cfg *e2esys.SystemConfig) {
				seconds := hexutil.Uint64(offset / time.Second)
				cfg.DeployConfig.L2GenesisEspressoTimeOffset = &seconds
			},
		}
	}
}

// WithFallbackAuthLeadTime is an E2eDevnetLauncherOption that sets the
// Espresso CLIConfig.FallbackAuthLeadTime on the batcher CLI config. The
// fallback batcher inherits this value from the TEE batcher's config struct
// (it is copied from the TEE batcher's CLIConfig in the e2e setup, before
// Espresso.Enabled is flipped to false), so a single launcher option suffices
// to configure both. Useful for tests that exercise the EspressoEnforcement
// hardfork boundary with a tighter lead time than the production default.
func WithFallbackAuthLeadTime(leadTime time.Duration) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				{
					Key:  "fallbackAuthLeadTime",
					Role: e2esys.RoleSeq,
					BatcherMod: func(batchConfig *bss.CLIConfig, sys *e2esys.System) {
						batchConfig.FallbackAuthLeadTime = leadTime
					},
				},
			},
		}
	}
}

// WithBatcherTargetNumFrames is a E2eDevnetLauncherOption that configures the
// batcher's `TargetNumFrames` option to the provided value.
//
// This governs how many frames the batcher will attempt to utilize when
// submitting a channel to the L1.
func WithBatcherTargetNumFrames(size int) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				{
					Key:  "maxL1NumFrames",
					Role: e2esys.RoleSeq,
					BatcherMod: func(batchConfig *bss.CLIConfig, sys *e2esys.System) {
						batchConfig.TargetNumFrames = size
					},
				},
			},
		}
	}
}

// WithBatcherMaxPendingTransactions is a E2eDevnetLauncherOption that
// configures the batcher's `MaxPendingTransactions` option to the provided
// value.
//
// This governs how many pending L1 transactions the batcher will allow
// before pausing new submissions.
func WithBatcherMaxPendingTransactions(pendingTransactions uint64) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				{
					Key:  "maxPendingTransactions",
					Role: e2esys.RoleSeq,
					BatcherMod: func(batchConfig *bss.CLIConfig, sys *e2esys.System) {
						batchConfig.MaxPendingTransactions = pendingTransactions
					},
				},
			},
		}
	}
}

// WithBatcherMaxL1TxSize is a E2eDevnetLauncherOption that configures the
// batcher's `MaxL1TxSize` option to the provided value.
//
// This governs the maximum L1 transaction size that the batcher will attempt
// to submit when submitting a channel to L1.
func WithBatcherMaxL1TxSize(maxL1TxSize uint64) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				{
					Key:  "maxL1TxSize",
					Role: e2esys.RoleSeq,
					BatcherMod: func(batchConfig *bss.CLIConfig, sys *e2esys.System) {
						batchConfig.MaxL1TxSize = maxL1TxSize

						if batchConfig.DataAvailabilityType == flags.BlobsType {
							// If we're setting the max data size for blobs,
							// we need to also inform the batcher to use that
							// setting when calculating blob sizes.
							//
							// Otherwise it will use the max blob size constant.
							batchConfig.TestUseMaxTxSizeForBlobs = true
						}
					},
				},
			},
		}
	}
}

// WithBatcherMaxBlocksPerSpanBatch is a E2eDevnetLauncherOption that
// configures the batcher's `MaxBlocksPerSpanBatch` option to the provided
// value.
//
// This governs how many blocks the batcher will include in a single span
// when creating batches to submit to L1.
func WithBatcherMaxBlocksPerSpanBatch(maxBlocksPerSpanBatch int) E2eDevnetLauncherOption {
	return func(c *E2eDevnetLauncherContext) E2eSystemOption {
		return E2eSystemOption{
			StartOptions: []e2esys.StartOption{
				{
					Key:  "maxBlocksPerSpanBatch",
					Role: e2esys.RoleSeq,
					BatcherMod: func(batchConfig *bss.CLIConfig, sys *e2esys.System) {
						batchConfig.MaxBlocksPerSpanBatch = maxBlocksPerSpanBatch
					},
				},
			},
		}
	}
}

// WithBatcherDataAvailabilityType is a E2eDevnetLauncherOption that configures
// the batcher's `DataAvailabilityType` option to the provided value.
//
// This governs which data availability method the batcher will use when
// submitting frames to L1.
func WithBatcherDataAvailabilityType(daAvailabilityType flags.DataAvailabilityType) E2eDevnetLauncherOption {
	{
		return func(c *E2eDevnetLauncherContext) E2eSystemOption {
			return E2eSystemOption{
				StartOptions: []e2esys.StartOption{
					{
						Key:  "dataAvailabilityType",
						Role: e2esys.RoleSeq,
						BatcherMod: func(batchConfig *bss.CLIConfig, sys *e2esys.System) {
							batchConfig.DataAvailabilityType = daAvailabilityType
						},
					},
				},
			}
		}
	}
}

// WithBatcherMaxChannelDuration is a configuration option that modifies the
// MaxChannelDuration for the Batcher Config.  This value will then be
// utilized by the Channels created by the batcher.
func WithBatcherMaxChannelDuration(maxChannelDuration uint64) E2eDevnetLauncherOption {
	{
		return func(c *E2eDevnetLauncherContext) E2eSystemOption {
			return E2eSystemOption{
				StartOptions: []e2esys.StartOption{
					{
						Key:  "maxChannelDuration",
						Role: e2esys.RoleSeq,
						BatcherMod: func(batchConfig *bss.CLIConfig, sys *e2esys.System) {
							batchConfig.MaxChannelDuration = maxChannelDuration
						},
					},
				},
			}
		}
	}
}

// WithBatcherMaxFrameSize is a configuration option that modifies the
// MaxChannelDuration for the Batcher Config.  This value will then be
// utilized by the channels created by the batcher.
func WithBatcherMaxFrameSize(maxFrameSize uint64) E2eDevnetLauncherOption {
	{
		return func(c *E2eDevnetLauncherContext) E2eSystemOption {
			return E2eSystemOption{
				StartOptions: []e2esys.StartOption{
					{
						Key:  "maxFrameSize",
						Role: e2esys.RoleSeq,
						BatcherMod: func(batchConfig *bss.CLIConfig, sys *e2esys.System) {
							batchConfig.MaxChannelDuration = maxFrameSize
						},
					},
				},
			}
		}
	}
}

// WithBatcherCompressor is a configuration option that modifies the Compressor
// setting of the Batcher Config.  This value will be utilized to determine
// compression options for the channels created by the batcher.
func WithBatcherCompressor(compressor string) E2eDevnetLauncherOption {
	{
		return func(c *E2eDevnetLauncherContext) E2eSystemOption {
			return E2eSystemOption{
				StartOptions: []e2esys.StartOption{
					{
						Key:  "compressor",
						Role: e2esys.RoleSeq,
						BatcherMod: func(batchConfig *bss.CLIConfig, sys *e2esys.System) {
							batchConfig.Compressor = compressor
						},
					},
				},
			}
		}
	}
}
