//! Celo L2 migration related data.

use alloy_consensus::{EMPTY_OMMER_ROOT_HASH, EMPTY_ROOT_HASH, Header};
use alloy_primitives::{B64, B256, U256, address, b256, bloom, bytes};

/// Cel2 migration header hash on Celo Mainnet.
pub const CEL2_HEADER_HASH: B256 =
    b256!("0x7586014e20c69b3fa7c9070baf1a7edd95833db57853126f32593b455da2e5c5");

/// Cel2 migration header on Celo Mainnet. (`31_056_500`)
pub const CEL2_HEADER: Header = Header {
    difficulty: U256::ZERO,
    extra_data: bytes!("43656c6f204c32206d6967726174696f6e"),
    gas_limit: 50_000_000,
    gas_used: 0,
    logs_bloom: bloom!(
        "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
    ),
    nonce: B64::ZERO,
    number: 31_056_500,
    parent_hash: b256!("0x4ecb0660a3b5e8bba3b3851e8926e9a44d0a61fe141e04c6e3e1c01644ce7c20"),
    receipts_root: EMPTY_ROOT_HASH,
    state_root: b256!("0xed980641a4bd4d2e84c6c8db980b7f05e95733c92be2e0045db3735efeb1d807"),
    timestamp: 1_742_957_258,
    transactions_root: EMPTY_ROOT_HASH,
    ommers_hash: EMPTY_OMMER_ROOT_HASH,
    beneficiary: address!("0x4200000000000000000000000000000000000011"),
    withdrawals_root: Some(EMPTY_ROOT_HASH),
    mix_hash: B256::ZERO,
    base_fee_per_gas: Some(0x5d240390e),
    blob_gas_used: Some(0),
    excess_blob_gas: Some(0),
    parent_beacon_block_root: Some(b256!(
        "0x6cb2e365f9d78b9071b90e8a1f4675d378cd0867b858571dc1b172ef1d3e085c"
    )),
    requests_hash: None,
};

/// Cel2 migration total difficulty on Celo Mainnet.
pub const CEL2_HEADER_TTD: U256 = U256::ZERO;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cel2_header() {
        assert_eq!(CEL2_HEADER.hash_slow(), CEL2_HEADER_HASH);
    }
}
