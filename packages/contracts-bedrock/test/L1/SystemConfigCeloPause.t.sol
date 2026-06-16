// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing
import { CommonTest } from "test/setup/CommonTest.sol";

// Libraries
import { Features } from "src/libraries/Features.sol";

// Interfaces
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";

/// @title SystemConfigCeloPause_TestInit
/// @notice Reusable init for the Celo owner-pause tests.
abstract contract SystemConfigCeloPause_TestInit is CommonTest {
    address owner;

    function setUp() public virtual override {
        super.setUp();
        owner = systemConfig.owner();
    }
}

/// @title SystemConfig_Pause_Test
/// @notice Tests the owner-controlled `pause` / `unpause` / `celoPaused`.
contract SystemConfig_Pause_Test is SystemConfigCeloPause_TestInit {
    /// @notice `celoPaused()` starts false.
    function test_celoPaused_default_succeeds() external view {
        assertFalse(systemConfig.celoPaused());
    }

    /// @notice Owner can pause; `celoPaused()` flips true and `Paused` is emitted.
    function test_pause_owner_succeeds() external {
        vm.expectEmit(address(systemConfig));
        emit Paused(address(0));
        vm.prank(owner);
        systemConfig.pause();
        assertTrue(systemConfig.celoPaused());
    }

    /// @notice Owner can unpause after pausing; `celoPaused()` flips back false.
    function test_unpause_owner_succeeds() external {
        vm.startPrank(owner);
        systemConfig.pause();
        assertTrue(systemConfig.celoPaused());

        vm.expectEmit(address(systemConfig));
        emit Unpaused(address(0));
        systemConfig.unpause();
        vm.stopPrank();
        assertFalse(systemConfig.celoPaused());
    }

    /// @notice Non-owner cannot pause.
    function test_pause_notOwner_reverts() external {
        vm.expectRevert("Ownable: caller is not the owner");
        vm.prank(makeAddr("stranger"));
        systemConfig.pause();
    }

    /// @notice Non-owner cannot unpause.
    function test_unpause_notOwner_reverts() external {
        vm.prank(owner);
        systemConfig.pause();

        vm.expectRevert("Ownable: caller is not the owner");
        vm.prank(makeAddr("stranger"));
        systemConfig.unpause();
    }
}

/// @title SystemConfig_PausedCelo_Test
/// @notice Tests that the owner-pause is OR-ed into `paused()` alongside the SuperchainConfig pause.
contract SystemConfig_PausedCelo_Test is SystemConfigCeloPause_TestInit {
    /// @notice An owner-pause alone makes `paused()` return true.
    function test_paused_celoPause_succeeds() external {
        assertFalse(systemConfig.paused());

        vm.prank(owner);
        systemConfig.pause();
        assertTrue(systemConfig.paused());
    }

    /// @notice `paused()` returns true if either the owner-pause or the SuperchainConfig pause is active.
    function test_paused_celoOrSuperchain_succeeds() external {
        vm.prank(owner);
        systemConfig.pause();
        assertTrue(systemConfig.paused());

        vm.prank(owner);
        systemConfig.unpause();
        assertFalse(systemConfig.paused());

        vm.prank(superchainConfig.guardian());
        superchainConfig.pause(address(0));
        assertTrue(systemConfig.paused());
    }

    /// @notice An owner-pause blocks ETH_LOCKBOX feature toggles (setFeature reads paused()).
    function test_setFeature_whileCeloPaused_reverts() external {
        skipIfSysFeatureEnabled(Features.ETH_LOCKBOX);

        address proxyAdminOwner = systemConfig.proxyAdminOwner();
        vm.prank(owner);
        systemConfig.pause();
        assertTrue(systemConfig.paused());

        vm.prank(proxyAdminOwner);
        vm.expectRevert(ISystemConfig.SystemConfig_InvalidFeatureState.selector);
        systemConfig.setFeature(Features.ETH_LOCKBOX, true);
    }
}
