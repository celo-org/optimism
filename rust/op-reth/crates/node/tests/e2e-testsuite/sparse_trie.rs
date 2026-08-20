use alloy_genesis::Genesis;
use reth_e2e_test_utils::setup_engine;
use reth_node_api::TreeConfig;
use reth_optimism_chainspec::OpChainSpecBuilder;
use reth_optimism_node::{
    OpNode,
    utils::{advance_chain, optimism_payload_attributes},
};
use reth_provider::BlockNumReader;
use std::sync::Arc;
use tokio::sync::Mutex;

/// Tests that the sparse trie pipeline can be shared with the payload builder.
///
/// Enables both `share_execution_cache_with_payload_builder` and
/// `share_sparse_trie_with_payload_builder`, then advances multiple blocks. Each FCU spawns a
/// `StateRootHandle` that the payload builder uses for incremental state root computation instead
/// of the blocking trie walk in `BlockBuilder::finish`.
///
/// The test validates that all blocks are successfully built and their state roots are accepted by
/// the engine (`newPayload` returns VALID). Without the wiring in `OpBuilder`, `trie_handle` is
/// discarded and this exercises nothing; with it, a root that disagreed with the engine's own
/// computation would fail the payload submission inside `advance_chain`.
#[tokio::test]
async fn test_share_sparse_trie_with_payload_builder() -> eyre::Result<()> {
    reth_tracing::init_test_tracing();

    let tree_config = TreeConfig::default()
        .with_legacy_state_root(false)
        .with_share_execution_cache_with_payload_builder(true)
        .with_share_sparse_trie_with_payload_builder(true);

    let genesis: Genesis = serde_json::from_str(include_str!("../assets/genesis.json")).unwrap();
    let (mut nodes, wallet) = setup_engine::<OpNode>(
        1,
        Arc::new(OpChainSpecBuilder::base_mainnet().genesis(genesis).ecotone_activated().build()),
        false,
        tree_config,
        optimism_payload_attributes,
    )
    .await?;

    let mut node = nodes.pop().unwrap();
    let wallet = Arc::new(Mutex::new(wallet));

    let num_blocks = 5;
    advance_chain(num_blocks, &mut node, wallet).await?;

    let best_block = node.inner.provider.best_block_number()?;
    assert_eq!(best_block, num_blocks as u64, "Expected {num_blocks} blocks, got {best_block}");

    Ok(())
}
