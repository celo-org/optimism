// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing utilities
import { Test } from "forge-std/Test.sol";

// Contracts
import { CeloGasBridgeL2 } from "src/celo/CeloGasBridgeL2.sol";
import { Proxy } from "src/universal/Proxy.sol";
import { StandardBridge } from "src/universal/StandardBridge.sol";

// Libraries
import { CeloPredeploys } from "src/celo/CeloPredeploys.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";

// Mocks
import {
    MockCrossDomainMessenger,
    MockOptimismMintableERC20,
    MockL1Block,
    MockLiquidityController
} from "test/celo/CeloBridgeHelpers.sol";

// Interfaces
import { ICeloGasBridgeL2 } from "interfaces/celo/ICeloGasBridgeL2.sol";
import { ICeloGasBridgeL1Finalizer } from "interfaces/celo/IBridgeFinalizers.sol";
import { ILiquidityController } from "interfaces/L2/ILiquidityController.sol";
import { IStandardBridge } from "interfaces/universal/IStandardBridge.sol";

/// @title CeloGasBridgeL2_TestInit
/// @notice Reusable test initialization for `CeloGasBridgeL2` tests.
abstract contract CeloGasBridgeL2_TestInit is Test {
    address internal alice;
    address internal bob;
    address payable internal otherBridge;
    address internal proxyAdmin;
    address internal celoTokenL1;

    CeloGasBridgeL2 internal implementation;
    ICeloGasBridgeL2 internal bridge;
    MockL1Block internal l1Block;
    MockLiquidityController internal liquidityController;
    MockCrossDomainMessenger internal messenger;

    function setUp() public virtual {
        alice = makeAddr("alice");
        bob = makeAddr("bob");
        otherBridge = payable(makeAddr("otherBridge"));
        proxyAdmin = makeAddr("proxyAdmin");
        celoTokenL1 = makeAddr("l1CeloToken");

        vm.etch(Predeploys.L1_BLOCK_ATTRIBUTES, address(new MockL1Block()).code);
        vm.etch(Predeploys.LIQUIDITY_CONTROLLER, address(new MockLiquidityController()).code);
        vm.etch(Predeploys.L2_CROSS_DOMAIN_MESSENGER, address(new MockCrossDomainMessenger()).code);

        l1Block = MockL1Block(Predeploys.L1_BLOCK_ATTRIBUTES);
        liquidityController = MockLiquidityController(Predeploys.LIQUIDITY_CONTROLLER);
        messenger = MockCrossDomainMessenger(Predeploys.L2_CROSS_DOMAIN_MESSENGER);

        l1Block.setIsCustomGasToken(true);

        implementation = new CeloGasBridgeL2();

        Proxy proxy = new Proxy(proxyAdmin);
        bridge = ICeloGasBridgeL2(payable(address(proxy)));

        vm.prank(proxyAdmin);
        proxy.upgradeToAndCall(
            address(implementation), abi.encodeCall(ICeloGasBridgeL2.initialize, (IStandardBridge(otherBridge), celoTokenL1))
        );
    }

    function _authorizeFinalizeDeposit() internal {
        messenger.setXDomainMessageSender(otherBridge);
        vm.prank(Predeploys.L2_CROSS_DOMAIN_MESSENGER);
    }
}

/// @title CeloGasBridgeL2_Version_Test
/// @notice Tests the `version` function of the `CeloGasBridgeL2` contract.
contract CeloGasBridgeL2_Version_Test is CeloGasBridgeL2_TestInit {
    function test_version_succeeds() external view {
        assertEq(bridge.version(), "1.0.0");
    }
}

/// @title CeloGasBridgeL2_Constructor_Test
/// @notice Tests the constructor behavior of the `CeloGasBridgeL2` contract.
contract CeloGasBridgeL2_Constructor_Test is CeloGasBridgeL2_TestInit {
    function test_constructor_initializeImplementation_reverts() external {
        vm.expectRevert("Initializable: contract is already initialized");
        implementation.initialize(StandardBridge(otherBridge), celoTokenL1);
    }
}

/// @title CeloGasBridgeL2_Initialize_Test
/// @notice Tests the `initialize` function of the `CeloGasBridgeL2` contract.
contract CeloGasBridgeL2_Initialize_Test is CeloGasBridgeL2_TestInit {
    function test_initialize_succeeds() external view {
        assertEq(address(bridge.MESSENGER()), Predeploys.L2_CROSS_DOMAIN_MESSENGER);
        assertEq(address(bridge.messenger()), Predeploys.L2_CROSS_DOMAIN_MESSENGER);
        assertEq(address(bridge.OTHER_BRIDGE()), otherBridge);
        assertEq(address(bridge.otherBridge()), otherBridge);
        assertEq(bridge.celoTokenL1(), celoTokenL1);
    }

    function test_initialize_alreadyInitialized_reverts() external {
        vm.expectRevert("Initializable: contract is already initialized");
        bridge.initialize(IStandardBridge(payable(makeAddr("newOtherBridge"))), celoTokenL1);
    }
}

/// @title CeloGasBridgeL2_Predeploy_Test
/// @notice Tests the non-genesis absence of the `CeloGasBridgeL2` predeploy.
contract CeloGasBridgeL2_Predeploy_Test is CeloGasBridgeL2_TestInit {
    function test_predeploy_nonCgtGenesisAbsent_succeeds() external view {
        assertEq(CeloPredeploys.CELO_GAS_BRIDGE_L2.code.length, 0);
    }
}

/// @title CeloGasBridgeL2_Withdraw_Test
/// @notice Tests the `withdraw` function of the `CeloGasBridgeL2` contract.
contract CeloGasBridgeL2_Withdraw_Test is CeloGasBridgeL2_TestInit {
    event ERC20BridgeInitiated(
        address indexed localToken,
        address indexed remoteToken,
        address indexed from,
        address to,
        uint256 amount,
        bytes extraData
    );

    function test_withdraw_succeeds() external {
        bytes memory extraData = hex"1234";
        uint256 amount = 100;
        uint32 minGasLimit = 250_000;

        vm.deal(alice, amount);

        vm.expectEmit(true, true, true, true, address(bridge));
        emit ERC20BridgeInitiated(CeloPredeploys.GOLD_TOKEN, celoTokenL1, alice, bob, amount, extraData);

        vm.prank(alice, alice);
        bridge.withdraw{ value: amount }(bob, amount, minGasLimit, extraData);

        assertEq(liquidityController.lastBurnCaller(), address(bridge));
        assertEq(liquidityController.lastBurnAmount(), amount);
        assertEq(address(liquidityController).balance, amount);

        assertEq(messenger.lastTarget(), otherBridge);
        assertEq(
            messenger.lastMessage(),
            abi.encodeCall(ICeloGasBridgeL1Finalizer.finalizeWithdrawal, (alice, bob, amount, extraData))
        );
        assertEq(messenger.lastMinGasLimit(), minGasLimit);
        assertEq(messenger.lastValue(), 0);
        assertEq(address(bridge).balance, 0);
    }

    function test_withdraw_invalidRecipient_self_reverts() external {
        vm.deal(alice, 1);
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_InvalidRecipient.selector);
        bridge.withdraw{ value: 1 }(address(bridge), 1, 1, hex"");
    }

    function test_withdraw_invalidRecipient_messenger_reverts() external {
        vm.deal(alice, 1);
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_InvalidRecipient.selector);
        bridge.withdraw{ value: 1 }(address(messenger), 1, 1, hex"");
    }

    function test_withdraw_notCgtMode_reverts() external {
        l1Block.setIsCustomGasToken(false);

        vm.deal(alice, 1);
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_NotCgtMode.selector);
        bridge.withdraw{ value: 1 }(bob, 1, 1, hex"");
    }

    function test_withdraw_zeroRecipient_reverts() external {
        vm.deal(alice, 1);
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_ZeroRecipient.selector);
        bridge.withdraw{ value: 1 }(address(0), 1, 1, hex"");
    }

    function test_withdraw_zeroAmount_reverts() external {
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_ZeroAmount.selector);
        bridge.withdraw(address(1), 0, 1, hex"");
    }

    function test_withdraw_valueMismatch_reverts() external {
        vm.deal(alice, 1);
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_ValueMismatch.selector);
        bridge.withdraw{ value: 1 }(bob, 2, 1, hex"");
    }

    function test_withdraw_fromContract_succeeds() external {
        // Contracts may bridge now that `onlyEOA` is removed; `_to` is explicit.
        uint256 amount = 100;
        vm.deal(address(this), amount);

        bridge.withdraw{ value: amount }(bob, amount, 250_000, hex"");

        assertEq(liquidityController.lastBurnAmount(), amount);
        assertEq(messenger.lastTarget(), otherBridge);
    }

    function test_withdraw_liquidityControllerUnauthorized_reverts() external {
        liquidityController.setBurnUnauthorized(true);

        vm.deal(alice, 1);
        vm.prank(alice, alice);
        vm.expectRevert(ILiquidityController.LiquidityController_Unauthorized.selector);
        bridge.withdraw{ value: 1 }(bob, 1, 1, hex"");
    }
}

/// @title CeloGasBridgeL2_FinalizeDeposit_Test
/// @notice Tests the `finalizeDeposit` function of the `CeloGasBridgeL2` contract.
contract CeloGasBridgeL2_FinalizeDeposit_Test is CeloGasBridgeL2_TestInit {
    event ERC20BridgeFinalized(
        address indexed localToken,
        address indexed remoteToken,
        address indexed from,
        address to,
        uint256 amount,
        bytes extraData
    );

    function test_finalizeDeposit_succeeds() external {
        bytes memory extraData = hex"1234";

        _authorizeFinalizeDeposit();
        vm.expectEmit(true, true, true, true, address(bridge));
        emit ERC20BridgeFinalized(CeloPredeploys.GOLD_TOKEN, celoTokenL1, alice, bob, 100, extraData);
        bridge.finalizeDeposit(alice, bob, 100, extraData);

        assertEq(liquidityController.lastMintCaller(), address(bridge));
        assertEq(liquidityController.lastMintTo(), bob);
        assertEq(liquidityController.lastMintAmount(), 100);
    }

    function test_finalizeDeposit_zeroRecipient_reverts() external {
        _authorizeFinalizeDeposit();

        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_ZeroRecipient.selector);
        bridge.finalizeDeposit(alice, address(0), 100, hex"");
    }

    function test_finalizeDeposit_invalidRecipient_self_reverts() external {
        _authorizeFinalizeDeposit();
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_InvalidRecipient.selector);
        bridge.finalizeDeposit(alice, address(bridge), 100, hex"");
    }

    function test_finalizeDeposit_invalidRecipient_messenger_reverts() external {
        _authorizeFinalizeDeposit();
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_InvalidRecipient.selector);
        bridge.finalizeDeposit(alice, address(messenger), 100, hex"");
    }

    function test_finalizeDeposit_notOtherBridge_reverts() external {
        vm.expectRevert("StandardBridge: function can only be called from the other bridge");
        bridge.finalizeDeposit(alice, bob, 100, hex"");
    }

    function test_finalizeDeposit_wrongXDomainSender_reverts() external {
        messenger.setXDomainMessageSender(makeAddr("wrongBridge"));

        vm.prank(Predeploys.L2_CROSS_DOMAIN_MESSENGER);
        vm.expectRevert("StandardBridge: function can only be called from the other bridge");
        bridge.finalizeDeposit(alice, bob, 100, hex"");
    }

    function test_finalizeDeposit_liquidityControllerUnauthorized_reverts() external {
        liquidityController.setMintUnauthorized(true);

        _authorizeFinalizeDeposit();
        vm.expectRevert(ILiquidityController.LiquidityController_Unauthorized.selector);
        bridge.finalizeDeposit(alice, bob, 100, hex"");
    }
}

/// @title CeloGasBridgeL2_DisabledMethods_Test
/// @notice Tests the disabled inherited `StandardBridge` entrypoints.
contract CeloGasBridgeL2_DisabledMethods_Test is CeloGasBridgeL2_TestInit {
    function test_bridgeETH_disabled_reverts() external {
        vm.etch(alice, hex"ffff");
        vm.deal(alice, 1);

        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_Disabled.selector);
        bridge.bridgeETH{ value: 1 }(1, hex"");
    }

    function test_bridgeETHTo_disabled_reverts() external {
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_Disabled.selector);
        bridge.bridgeETHTo(bob, 1, hex"");
    }

    function test_bridgeERC20_disabled_reverts() external {
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_Disabled.selector);
        bridge.bridgeERC20(address(1), address(2), 3, 4, hex"");
    }

    function test_bridgeERC20To_disabled_reverts() external {
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_Disabled.selector);
        bridge.bridgeERC20To(address(1), address(2), address(3), 4, 5, hex"");
    }

    function test_finalizeBridgeETH_disabled_reverts() external {
        _authorizeFinalizeDeposit();
        vm.deal(Predeploys.L2_CROSS_DOMAIN_MESSENGER, 1);

        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_Disabled.selector);
        bridge.finalizeBridgeETH{ value: 1 }(alice, bob, 1, hex"");
    }

    function test_finalizeBridgeERC20_disabled_reverts() external {
        address localToken = address(new MockOptimismMintableERC20(address(2)));

        _authorizeFinalizeDeposit();

        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_Disabled.selector);
        bridge.finalizeBridgeERC20(localToken, address(2), alice, bob, 3, hex"");
    }

    function test_receive_disabled_reverts() external {
        vm.deal(alice, 1);

        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL2.CeloGasBridgeL2_Disabled.selector);
        (bool revertsAsExpected,) = address(bridge).call{ value: 1 }(hex"");
        assertTrue(revertsAsExpected, "expectRevert: call did not revert");
    }
}
