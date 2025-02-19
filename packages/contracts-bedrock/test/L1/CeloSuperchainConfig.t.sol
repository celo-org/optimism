// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { CommonTest } from "test/setup/CommonTest.sol";

// Target contract dependencies
import { IProxy } from "src/universal/interfaces/IProxy.sol";

// Target contract
import { ICeloSuperchainConfig } from "src/L1/interfaces/ICeloSuperchainConfig.sol";

import { DeployUtils } from "scripts/libraries/DeployUtils.sol";

contract CeloSuperchainConfig_Test_Setup is CommonTest {
    address internal celoGuardian;

    function setUp() public override {
        super.setUp();

        celoGuardian = deploy.cfg().superchainConfigGuardian();
    }
}

contract CeloSuperchainConfig_Init_Test is CeloSuperchainConfig_Test_Setup {
    /// @dev Tests that initialization sets the correct values.
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
            abi.encodeWithSignature("initialize(address,bool,address)", celoGuardian, true, address(superchainConfig))
        );

        assertTrue(ICeloSuperchainConfig(address(newProxy)).paused());
        assertEq(ICeloSuperchainConfig(address(newProxy)).guardian(), celoGuardian);
        assertEq(celoSuperchainConfig.superchainConfig(), address(superchainConfig));
    }

    /// @dev Tests that it will be intialized as paused if Superchain was paused
    ///      at initialization time.
    function test_initialize_whenSuperchainPaused_succeeds() external {
        vm.prank(superchainConfig.guardian());
        superchainConfig.pause("identifier");
        assertTrue(superchainConfig.paused());

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
            abi.encodeWithSignature("initialize(address,bool,address)", celoGuardian, false, address(superchainConfig))
        );
        vm.stopPrank();

        vm.prank(superchainConfig.guardian());
        superchainConfig.unpause();
        assertFalse(superchainConfig.paused());

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
        vm.expectRevert("CeloSuperchainConfig: only guardian can pause");
        vm.prank(alice);
        celoSuperchainConfig.pause("identifier");

        assertFalse(celoSuperchainConfig.paused());
    }

    /// @dev Tests that `pause` reverts when called by the Superchain Guardian.
    ///      Currently this test is skipped because the dev Deploy setup ends up
    ///      using the same address for both Guardian roles.
    function test_pause_superchainGuardian_reverts() external {
        assertFalse(celoSuperchainConfig.paused());

        address superchainConfigGuardian = deploy.cfg().superchainConfigGuardian();
        if (superchainConfigGuardian == celoSuperchainConfig.guardian()) {
            vm.skip(true);
        }

        vm.expectRevert("CeloSuperchainConfig: only guardian can pause");
        vm.prank(superchainConfigGuardian);
        celoSuperchainConfig.pause("identifier");

        assertFalse(celoSuperchainConfig.paused());
    }
}

contract CeloSuperchainConfig_Pause_Test is CeloSuperchainConfig_Test_Setup {
    /// @dev Tests that `pause` successfully pauses when called by the guardian.
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
        vm.expectRevert("CeloSuperchainConfig: only guardian can unpause");
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

    /// @dev Tests that after `unpause`, if the Superchain is still paused, Celo
    ///      remains paused anyways.
    function test_unpause_whenSuperchainPaused_succeeds() external {
        vm.prank(celoSuperchainConfig.guardian());
        celoSuperchainConfig.pause("identifier");
        assertEq(celoSuperchainConfig.paused(), true);

        vm.prank(superchainConfig.guardian());
        superchainConfig.pause("identifier");
        assertTrue(superchainConfig.paused());

        vm.prank(celoSuperchainConfig.guardian());
        celoSuperchainConfig.unpause();

        assertEq(celoSuperchainConfig.paused(), true);
    }
}

contract CeloSuperchainConfig_Paused_Test is CeloSuperchainConfig_Test_Setup {
    /// @dev Tests that the `paused` getter returns `false` whenever both Celo
    ///      and Superchain are unpaused.
    function test_paused_whenSuperchainUnpaused_succeeds() external view {
        assertFalse(superchainConfig.paused());
        bool paused = celoSuperchainConfig.paused();
        assertFalse(paused);
    }

    /// @dev Tests that the `paused` getter returns `true` whenever Superchain
    ///      is paused, even when Celo is unpaused.
    function test_paused_whenSuperchainPaused_succeeds() external {
        vm.prank(superchainConfig.guardian());
        superchainConfig.pause("identifier");
        assertTrue(superchainConfig.paused());

        bool paused = celoSuperchainConfig.paused();
        assertTrue(paused);
    }

    /// @dev Tests that the `paused` getter returns `true` whenever Celo is paused,
    ///      even when the Superchain is unpaused.
    function test_paused_whenCeloPaused_succeeds() external {
        vm.prank(celoGuardian);
        celoSuperchainConfig.pause("identifier");

        assertFalse(superchainConfig.paused());

        bool paused = celoSuperchainConfig.paused();
        assertTrue(paused);
    }

    /// @dev Tests that the `paused` getter returns `true` whenever both Celo
    ///      and the Superchain are paused.
    function test_paused_whenBothPaused_succeeds() external {
        vm.prank(superchainConfig.guardian());
        superchainConfig.pause("identifier");
        assertTrue(superchainConfig.paused());

        vm.prank(celoGuardian);
        celoSuperchainConfig.pause("identifier");

        bool paused = celoSuperchainConfig.paused();
        assertTrue(paused);
    }
}

contract CeloSuperchainConfig_CheckAndPauseIfSuperchainPaused_Test is CeloSuperchainConfig_Test_Setup {
    /// @dev Tests that `checkAndPauseIfSuperchainPaused` is a no-op when
    ///      Superchain is unpaused.
    function test_checkAndPauseIfSuperchainPaused_whenSuperchainUnpaused_succeeds() external {
        assertFalse(superchainConfig.paused());

        bool paused = celoSuperchainConfig.checkAndPauseIfSuperchainPaused();
        assertFalse(paused);

        paused = celoSuperchainConfig.paused();
        assertFalse(paused);
    }

    /// @dev Tests that `checkAndPauseIfSuperchainPaused` propagatates
    ///      Superchain's paused status to Celo when the Superchain is paused.
    function test_checkAndPauseIfSuperchainPaused_whenSuperchainPaused_succeeds() external {
        vm.prank(superchainConfig.guardian());
        superchainConfig.pause("identifier");
        assertTrue(superchainConfig.paused());

        bool paused = celoSuperchainConfig.checkAndPauseIfSuperchainPaused();
        assertTrue(paused);

        paused = celoSuperchainConfig.paused();
        assertTrue(paused);

        vm.prank(superchainConfig.guardian());
        superchainConfig.unpause();
        assertFalse(superchainConfig.paused());

        paused = celoSuperchainConfig.paused();
        assertTrue(paused);
    }

    /// @dev Tests that `checkAndPauseIfSuperchainPaused` works even if Superchain is not set.
    function test_checkAndPauseIfSuperchainPaused_whenSuperchainNotSet() external {
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
            abi.encodeWithSignature("initialize(address,bool,address)", celoGuardian, false, address(0))
        );
        vm.stopPrank();

        ICeloSuperchainConfig newCeloSuperchainConfig = ICeloSuperchainConfig(address(newProxy));

        bool paused = newCeloSuperchainConfig.checkAndPauseIfSuperchainPaused();
        assertFalse(paused);

        vm.prank(newCeloSuperchainConfig.guardian());
        newCeloSuperchainConfig.pause("identifier");
        assertTrue(newCeloSuperchainConfig.paused());

        paused = newCeloSuperchainConfig.checkAndPauseIfSuperchainPaused();
        assertTrue(paused);

        vm.prank(newCeloSuperchainConfig.guardian());
        newCeloSuperchainConfig.unpause();
        assertFalse(newCeloSuperchainConfig.paused());

        paused = newCeloSuperchainConfig.paused();
        assertFalse(paused);
    }
}
