use alloy_consensus::BlockHeader;
use alloy_genesis::Genesis;
use reth_e2e_test_utils::setup_engine;
use reth_node_api::TreeConfig;
use reth_node_metrics::recorder::install_prometheus_recorder;
use reth_optimism_chainspec::{OP_SEPOLIA, OpChainSpecBuilder};
use reth_optimism_node::{
    OpNode,
    utils::{advance_chain, optimism_payload_attributes},
};
use reth_provider::{BlockNumReader, HeaderProvider, StateProviderFactory};
use std::sync::Arc;
use tokio::sync::Mutex;

/// Tests that the sparse trie pipeline can be shared with the payload builder.
#[tokio::test]
async fn test_share_sparse_trie_with_payload_builder() -> eyre::Result<()> {
    reth_tracing::init_test_tracing();

    let tree_config = TreeConfig::default()
        .with_share_execution_cache_with_payload_builder(true)
        .with_share_sparse_trie_with_payload_builder(true);

    let genesis: Genesis = serde_json::from_str(include_str!("../assets/genesis.json")).unwrap();
    let (mut nodes, wallet) = setup_engine::<OpNode>(
        1,
        Arc::new(
            OpChainSpecBuilder::default()
                .chain(OP_SEPOLIA.chain)
                .genesis(genesis)
                .ecotone_activated()
                .build(),
        ),
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

    let from_shared_trie = counter("state_root_from_shared_trie_total");
    let fallbacks = counter("state_root_fallback_total");
    assert!(
        from_shared_trie >= num_blocks as u64,
        "expected at least {num_blocks} roots from the shared trie, got {from_shared_trie} \
         ({fallbacks} fallbacks)"
    );

    // Independently walk the canonical database trie and compare its root with the payload header.
    // This catches a shared-trie wiring bug even if the engine accepts the same incorrect result.
    let header =
        node.inner.provider.header_by_number(best_block)?.expect("best block header should exist");
    let synchronous_root = node.inner.provider.latest()?.state_root(Default::default())?;
    assert_eq!(synchronous_root, header.state_root());

    Ok(())
}

/// Reads an `optimism_payload_builder` counter from the process-wide Prometheus recorder.
fn counter(name: &str) -> u64 {
    let rendered = install_prometheus_recorder().handle().render();
    let key = format!("reth_optimism_payload_builder_{name} ");
    rendered
        .lines()
        .find_map(|line| line.strip_prefix(&key))
        .and_then(|value| value.trim().parse().ok())
        .unwrap_or(0)
}
