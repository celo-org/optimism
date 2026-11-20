// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Per-chain gas budgets for `CrossDomainMessenger.relayMessage`, named for the chain the relay
// executes on. A messenger uses its own chain's values in `relayMessage` and the other chain's
// values in `baseGas`, so each value must cover the worst case on the chain it is named for.
//
// Only the two budgets that differ per chain live here. `RELAY_CONSTANT_OVERHEAD` (200,000) and
// `RELAY_CALL_OVERHEAD` (40,000) are identical on both chains and stay in `CrossDomainMessenger`.
//
// Ported from op-contracts/v9.0.0 (`src/libraries/GasConstants.sol`, commit 8fdb2775), trimmed to
// the constants Celo needs. Upstream's full version also splits the two shared budgets and is what
// a later v7-v10 port should replace this file with.
//
// File-level rather than a library so that contract constants can be initialized from them. Import
// as `import "src/libraries/GasConstants.sol" as GasConstants;`. NatSpec on file-level constants
// accepts only `@dev`.

////////////////////////////////////////////////////////////////
//                        Reserved gas                        //
////////////////////////////////////////////////////////////////

// `relayMessage` forwards `gasleft()` minus this to the target, so it covers everything after the
// call: resetting `xDomainMsgSender`, writing `successfulMessages` or `failedMessages` (a fresh
// slot) and emitting the event. Too low, and a target that uses all its gas makes the relay revert
// after the call, so the message is recorded in neither mapping and cannot be replayed.

/// @dev Includes the EIP-8037 state gas for the fresh slot (GAS_STORAGE_SET = 64 * CPSB = 97,920).
///      State gas is drawn from `gas_left` whenever the transaction has no state-gas reservoir,
///      i.e. for every transaction with `tx.gas <= 2**24`. Upstream measured minimum under
///      Amsterdam: 127,089.
uint64 constant L1_RELAY_RESERVED_GAS = 135_000;

/// @dev Upstream measured minimum under Prague: 35,969.
uint64 constant L2_RELAY_RESERVED_GAS = 40_000;

////////////////////////////////////////////////////////////////
//                      Gas check buffer                      //
////////////////////////////////////////////////////////////////

// Covers the work between the `hasMinGas` check and the CALL: setting `xDomainMsgSender` and
// preparing the call's arguments. Too low, and the target can receive less than `minGasLimit`.

/// @dev The `xDomainMsgSender` write alone costs 10,100 under Amsterdam.
uint64 constant L1_RELAY_GAS_CHECK_BUFFER = 15_000;

/// @dev The `xDomainMsgSender` write alone costs 2,900 under Prague.
uint64 constant L2_RELAY_GAS_CHECK_BUFFER = 5_000;
