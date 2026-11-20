// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { CommonTest } from "test/setup/CommonTest.sol";

// Libraries
import "src/libraries/GasConstants.sol" as GasConstants;

/// @title GasConstants_Wiring_Test
/// @notice Pins the direction of the per-chain relay budgets. A messenger must budget `baseGas`
///         with the OTHER chain's values, because the relay it is paying for executes there.
///         Getting this backwards silently under-provisions one direction of the bridge.
contract GasConstants_Wiring_Test is CommonTest {
    /// @notice L1's public budgets describe the L2 relay that `baseGas` pays for.
    function test_l1Messenger_publicBudgetsAreRemote_works() public view {
        assertEq(l1CrossDomainMessenger.RELAY_RESERVED_GAS(), GasConstants.L2_RELAY_RESERVED_GAS);
        assertEq(l1CrossDomainMessenger.RELAY_GAS_CHECK_BUFFER(), GasConstants.L2_RELAY_GAS_CHECK_BUFFER);
    }

    /// @notice L2's public budgets describe the L1 relay that `baseGas` pays for.
    function test_l2Messenger_publicBudgetsAreRemote_works() public view {
        assertEq(l2CrossDomainMessenger.RELAY_RESERVED_GAS(), GasConstants.L1_RELAY_RESERVED_GAS);
        assertEq(l2CrossDomainMessenger.RELAY_GAS_CHECK_BUFFER(), GasConstants.L1_RELAY_GAS_CHECK_BUFFER);
    }

    /// @notice The L1 budgets must exceed the L2 ones: Glamsterdam charges EIP-8037 state gas for
    ///         the fresh `successfulMessages` slot, and that state gas comes out of `gas_left` for
    ///         any transaction with `tx.gas <= 2**24`.
    function test_l1BudgetsCoverStateGas_works() public pure {
        // GAS_STORAGE_SET = 64 * CPSB = 64 * 1530.
        uint64 freshSlotStateGas = 64 * 1530;
        assertGt(GasConstants.L1_RELAY_RESERVED_GAS, freshSlotStateGas);
        assertGt(GasConstants.L1_RELAY_RESERVED_GAS, GasConstants.L2_RELAY_RESERVED_GAS);
        assertGt(GasConstants.L1_RELAY_GAS_CHECK_BUFFER, GasConstants.L2_RELAY_GAS_CHECK_BUFFER);
    }

    /// @notice `baseGas` must move with the remote budgets, proving the plumbing is connected.
    function test_baseGas_tracksRemoteBudgets_works() public view {
        uint64 l1Base = l1CrossDomainMessenger.baseGas(hex"ff", 100_000);
        uint64 l2Base = l2CrossDomainMessenger.baseGas(hex"ff", 100_000);

        // L2 pays for an L1 relay, so its budget is larger by exactly the L1/L2 difference.
        uint64 expectedDelta = (GasConstants.L1_RELAY_RESERVED_GAS - GasConstants.L2_RELAY_RESERVED_GAS)
            + (GasConstants.L1_RELAY_GAS_CHECK_BUFFER - GasConstants.L2_RELAY_GAS_CHECK_BUFFER);
        assertEq(l2Base - l1Base, expectedDelta);
    }
}
