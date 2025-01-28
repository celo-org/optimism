// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { CommonTest } from "test/setup/CommonTest.sol";

// Target contract dependencies
import { IProxy } from "src/universal/interfaces/IProxy.sol";

// Target contract
import { ICeloSuperchainConfig } from "src/L1/interfaces/ICeloSuperchainConfig.sol";

import { DeployUtils } from "scripts/libraries/DeployUtils.sol";

/// @dev For now, testing this using an individual setup contract here, rather
///      than including this contract in the global CommonTest setup. TBD how
///      this should be handled.
contract CeloSuperchainConfig_Test_Setup is CommonTest {
    ICeloSuperchainConfig internal celoSuperchainConfig;
    address internal celoGuardian;

    function setUp() public override {
        super.setUp();

        celoGuardian = makeAddr("celoGuardian");

        IProxy newProxy = IProxy(
            DeployUtils.create1({
                _name: "Proxy",
                _args: DeployUtils.encodeConstructor(abi.encodeCall(IProxy.__constructor__, (alice)))
            })
        );
        ICeloSuperchainConfig newImpl = ICeloSuperchainConfig(
            DeployUtils.create1({
                _name: "CeloSuperchainConfig",
                _args: DeployUtils.encodeConstructor(abi.encodeCall(ICeloSuperchainConfig.__constructor__, ()))
            })
        );

        vm.startPrank(alice);
        newProxy.upgradeToAndCall(
            address(newImpl),
            abi.encodeWithSignature(
                "initialize(address,bool,address)",
                celoGuardian,
                false,
                address(superchainConfig)
            )
        );
        vm.stopPrank();

        celoSuperchainConfig = ICeloSuperchainConfig(address(newProxy));
    }
}

contract CeloSuperchainConfig_Init_Test is CeloSuperchainConfig_Test_Setup {
    /// @dev Tests that initialization sets the correct values. These are defined in CommonTest.sol.
    function test_initialize_unpaused_succeeds() external view {
        assertFalse(celoSuperchainConfig.paused());
        assertEq(celoSuperchainConfig.guardian(), celoGuardian);
        assertEq(celoSuperchainConfig.superchainConfig(), address(superchainConfig));
    }

    /// @dev Tests that it can be intialized as paused.
    function test_initialize_paused_succeeds() external {
        IProxy newProxy = IProxy(
            DeployUtils.create1({
                _name: "Proxy",
                _args: DeployUtils.encodeConstructor(abi.encodeCall(IProxy.__constructor__, (alice)))
            })
        );
        ICeloSuperchainConfig newImpl = ICeloSuperchainConfig(
            DeployUtils.create1({
                _name: "CeloSuperchainConfig",
                _args: DeployUtils.encodeConstructor(abi.encodeCall(ICeloSuperchainConfig.__constructor__, ()))
            })
        );

        vm.startPrank(alice);
        newProxy.upgradeToAndCall(
            address(newImpl),
            abi.encodeWithSignature(
                "initialize(address,bool,address)",
                celoGuardian,
                true,
                address(superchainConfig)
            )
        );

        assertTrue(ICeloSuperchainConfig(address(newProxy)).paused());
        assertEq(ICeloSuperchainConfig(address(newProxy)).guardian(), celoGuardian);
        assertEq(celoSuperchainConfig.superchainConfig(), address(superchainConfig));
    }
}

contract CeloSuperchainConfig_Pause_TestFail is CeloSuperchainConfig_Test_Setup {
    /// @dev Tests that `pause` reverts when called by a non-guardian.
    function test_pause_notGuardian_reverts() external {
        assertFalse(celoSuperchainConfig.paused());

        assertTrue(celoSuperchainConfig.guardian() != alice);
        vm.expectRevert("SuperchainConfig: only guardian can pause");
        vm.prank(alice);
        celoSuperchainConfig.pause("identifier");

        assertFalse(celoSuperchainConfig.paused());
    }

    /// @dev Tests that `pause` reverts when called by the Superchain Guardian.
    function test_pause_superchainGuardian_reverts() external {
        assertFalse(celoSuperchainConfig.paused());

        address superchainConfigGuardian = deploy.cfg().superchainConfigGuardian();
        assertTrue(celoSuperchainConfig.guardian() != superchainConfigGuardian);

        vm.expectRevert("SuperchainConfig: only guardian can pause");
        vm.prank(superchainConfigGuardian);
        celoSuperchainConfig.pause("identifier");

        assertFalse(celoSuperchainConfig.paused());
    }
}

contract CeloSuperchainConfig_Pause_Test is CeloSuperchainConfig_Test_Setup {
    /// @dev Tests that `pause` successfully pauses
    ///      when called by the guardian.
    function test_pause_succeeds() external {
        assertFalse(celoSuperchainConfig.paused());

        vm.expectEmit(address(celoSuperchainConfig));
        emit Paused("identifier");

        vm.prank(celoSuperchainConfig.guardian());
        celoSuperchainConfig.pause("identifier");

        assertTrue(celoSuperchainConfig.paused());
    }
}

contract CeloSuperchainConfig_Unpause_TestFail is CeloSuperchainConfig_Test_Setup {
    /// @dev Tests that `unpause` reverts when called by a non-guardian.
    function test_unpause_notGuardian_reverts() external {
        vm.prank(celoSuperchainConfig.guardian());
        celoSuperchainConfig.pause("identifier");
        assertEq(celoSuperchainConfig.paused(), true);

        assertTrue(celoSuperchainConfig.guardian() != alice);
        vm.expectRevert("SuperchainConfig: only guardian can unpause");
        vm.prank(alice);
        celoSuperchainConfig.unpause();

        assertTrue(celoSuperchainConfig.paused());
    }
}

contract CeloSuperchainConfig_Unpause_Test is CeloSuperchainConfig_Test_Setup {
    /// @dev Tests that `unpause` successfully unpauses
    ///      when called by the guardian.
    function test_unpause_succeeds() external {
        vm.startPrank(celoSuperchainConfig.guardian());
        celoSuperchainConfig.pause("identifier");
        assertEq(celoSuperchainConfig.paused(), true);

        vm.expectEmit(address(celoSuperchainConfig));
        emit Unpaused();
        celoSuperchainConfig.unpause();

        assertFalse(celoSuperchainConfig.paused());
    }
}
