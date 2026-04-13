//! Command that initializes the node from a genesis file.

use alloy_consensus::Header;
use clap::Parser;
use reth_chainspec::EthChainSpec;
use reth_cli::chainspec::ChainSpecParser;
use reth_cli_commands::common::{AccessRights, CliNodeTypes, Environment};
use reth_db_common::init::init_from_state_dump;
use reth_optimism_chainspec::OpChainSpec;
use reth_optimism_primitives::{
    OpPrimitives,
    bedrock::{BEDROCK_HEADER, BEDROCK_HEADER_HASH},
    celo::{CEL2_HEADER, CEL2_HEADER_HASH},
};
use reth_primitives_traits::{SealedHeader, header::HeaderMut};
use reth_provider::{
    BlockNumReader, DBProvider, DatabaseProviderFactory, StaticFileProviderFactory,
    StaticFileWriter, StorageSettingsCache,
};
use reth_static_file_types::StaticFileSegment;
use std::{io::BufReader, sync::Arc};
use tracing::info;

/// Celo mainnet chain ID.
const CELO_MAINNET_CHAIN_ID: u64 = 42220;

/// Initializes the database with the genesis block.
#[derive(Debug, Parser)]
pub struct InitStateCommandOp<C: ChainSpecParser> {
    #[command(flatten)]
    init_state: reth_cli_commands::init_state::InitStateCommand<C>,

    /// Specifies whether to initialize the state without relying on OVM or EVM historical data.
    ///
    /// When enabled, and before inserting the state, it creates a dummy chain up to the
    /// migration block, then appends the migration header. Hardcoded migration headers are
    /// available for OP mainnet (Bedrock, block #105235063) and Celo mainnet (Cel2, block
    /// #31056500). For other OP chains, a header must be passed in.
    ///
    /// - **Note**: **Do not** import receipts and blocks beforehand, or this will fail or be
    ///   ignored.
    #[arg(long, default_value = "false")]
    without_ovm: bool,
}

impl<C: ChainSpecParser<ChainSpec = OpChainSpec>> InitStateCommandOp<C> {
    /// Execute the `init` command
    pub async fn execute<N: CliNodeTypes<ChainSpec = C::ChainSpec, Primitives = OpPrimitives>>(
        mut self,
    ) -> eyre::Result<()> {
        if self.without_ovm {
            if self.init_state.env.chain.is_optimism_mainnet() {
                return self.execute_with_migration_header::<N>(
                    "OP mainnet",
                    SealedHeader::new(BEDROCK_HEADER, BEDROCK_HEADER_HASH),
                );
            }

            if self.init_state.env.chain.chain_id() == CELO_MAINNET_CHAIN_ID {
                return self.execute_with_migration_header::<N>(
                    "Celo mainnet",
                    SealedHeader::new(CEL2_HEADER, CEL2_HEADER_HASH),
                );
            }

            // For other OP chains with --without-ovm, use the base implementation
            // by setting the without_evm flag
            self.init_state.without_evm = true;
        }

        self.init_state.execute::<N>().await
    }

    /// Execute init-state with a hardcoded migration header.
    ///
    /// Creates a dummy chain up to the block before `migration_header`, appends the migration
    /// header, then imports the state dump.
    fn execute_with_migration_header<
        N: CliNodeTypes<ChainSpec = C::ChainSpec, Primitives = OpPrimitives>,
    >(
        self,
        chain_name: &str,
        migration_header: SealedHeader<Header>,
    ) -> eyre::Result<()> {
        info!(target: "reth::cli", chain_name, "Reth init-state starting");
        let env = self.init_state.env.init::<N>(AccessRights::RW)?;

        let Environment { config, provider_factory, .. } = env;
        let static_file_provider = provider_factory.static_file_provider();
        let provider_rw = provider_factory.database_provider_rw()?;

        let last_block_number = provider_rw.last_block_number()?;
        let migration_block_number = migration_header.header().number;

        if last_block_number == 0 {
            reth_cli_commands::init_state::without_evm::setup_without_evm(
                &provider_rw,
                migration_header,
                |number| {
                    let mut header = Header::default();
                    header.set_number(number);
                    header
                },
            )?;

            // With storage v2, genesis init created changeset segments at block 0.
            // Advance them through the dummy blocks so the next expected block is
            // the migration block. Each increment_block call is cheap (header
            // counter + small offset), and new segment files are created as needed.
            if provider_rw.cached_storage_settings().storage_v2 {
                info!(target: "reth::cli", "Advancing changeset segments to migration block");
                for segment in [
                    StaticFileSegment::AccountChangeSets,
                    StaticFileSegment::StorageChangeSets,
                ] {
                    let mut writer = static_file_provider.latest_writer(segment)?;
                    for block in 1..migration_block_number {
                        writer.increment_block(block)?;
                    }
                }
            }

            // SAFETY: it's safe to commit static files, since in the event of a crash, they
            // will be unwound according to database checkpoints.
            //
            // Necessary to commit, so the migration header is accessible to provider_rw and
            // init_from_state_dump.
            static_file_provider.commit()?;
        } else if last_block_number > 0 && last_block_number < migration_block_number {
            return Err(eyre::eyre!(
                "Data directory should be empty when calling init-state with --without-ovm."
            ));
        }

        info!(target: "reth::cli", "Initiating state dump");

        let reader = BufReader::new(reth_fs_util::open(self.init_state.state)?);
        let hash = init_from_state_dump(reader, &provider_rw, config.stages.etl)?;

        provider_rw.commit()?;

        info!(target: "reth::cli", hash = ?hash, "Genesis block written");
        Ok(())
    }
}

impl<C: ChainSpecParser> InitStateCommandOp<C> {
    /// Returns the underlying chain being used to run this command
    pub fn chain_spec(&self) -> Option<&Arc<C::ChainSpec>> {
        self.init_state.chain_spec()
    }
}
