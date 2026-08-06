// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing utilities
import { Test } from "forge-std/Test.sol";

// Contracts
import { CeloGasBridgeL1 } from "src/celo/CeloGasBridgeL1.sol";
import { Proxy } from "src/universal/Proxy.sol";
import { StandardBridge } from "src/universal/StandardBridge.sol";

// Libraries
import { Constants } from "src/libraries/Constants.sol";
import { CeloPredeploys } from "src/celo/CeloPredeploys.sol";

// Mocks
import {
    MockERC20,
    MockCrossDomainMessenger,
    MockOptimismMintableERC20,
    MockSystemConfig,
    MockProxyAdmin
} from "test/celo/CeloBridgeHelpers.sol";

// Interfaces
import { ICeloGasBridgeL1 } from "interfaces/celo/ICeloGasBridgeL1.sol";
import { ICeloGasBridgeL2Finalizer } from "interfaces/celo/IBridgeFinalizers.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IStandardBridge } from "interfaces/universal/IStandardBridge.sol";
import { ICrossDomainMessenger } from "interfaces/universal/ICrossDomainMessenger.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { IProxyAdminOwnedBase } from "interfaces/L1/IProxyAdminOwnedBase.sol";

/// @title CeloGasBridgeL1_TestInit
/// @notice Reusable test initialization for `CeloGasBridgeL1` tests.
abstract contract CeloGasBridgeL1_TestInit is Test {
    event EscrowSeeded(address indexed portal, uint256 amount);

    address internal alice;
    address internal bob;
    address payable internal otherBridge;
    address internal optimismPortal;
    address internal proxyAdminOwnerAddr;

    MockERC20 internal celoTokenL1;
    MockSystemConfig internal systemConfig;
    MockCrossDomainMessenger internal messenger;
    MockProxyAdmin internal proxyAdminContract;

    CeloGasBridgeL1 internal implementation;
    ICeloGasBridgeL1 internal bridge;

    function setUp() public virtual {
        alice = makeAddr("alice");
        bob = makeAddr("bob");
        otherBridge = payable(makeAddr("otherBridge"));
        optimismPortal = makeAddr("optimismPortal");
        proxyAdminOwnerAddr = makeAddr("proxyAdminOwner");

        celoTokenL1 = new MockERC20();
        systemConfig = new MockSystemConfig();
        systemConfig.setIsCustomGasToken(true);
        systemConfig.setOptimismPortal(optimismPortal);
        messenger = new MockCrossDomainMessenger();

        proxyAdminContract = new MockProxyAdmin(proxyAdminOwnerAddr);

        implementation = new CeloGasBridgeL1();

        bridge = _deployBridge();
    }

    /// @notice Deploys a fresh initialized CeloGasBridgeL1 proxy. Escrow is seeded later via `seedEscrow`.
    function _deployBridge() internal returns (ICeloGasBridgeL1 bridge_) {
        Proxy proxy = new Proxy(address(proxyAdminContract));
        bridge_ = ICeloGasBridgeL1(payable(address(proxy)));

        // Stash ProxyAdmin in PROXY_OWNER_ADDRESS so ProxyAdminOwnedBase reads it during initialize().
        vm.store(
            address(bridge_), Constants.PROXY_OWNER_ADDRESS, bytes32(uint256(uint160(address(proxyAdminContract))))
        );

        vm.prank(address(proxyAdminContract));
        proxy.upgradeToAndCall(
            address(implementation),
            abi.encodeCall(
                ICeloGasBridgeL1.initialize,
                (
                    ICrossDomainMessenger(address(messenger)),
                    ISystemConfig(address(systemConfig)),
                    IStandardBridge(otherBridge),
                    IERC20(address(celoTokenL1))
                )
            )
        );
    }

    function _asOtherBridge() internal {
        messenger.setXDomainMessageSender(otherBridge);
        vm.prank(address(messenger));
    }

    function _seedEscrow(ICeloGasBridgeL1 _bridge, uint256 _amount) internal {
        vm.prank(systemConfig.optimismPortal());
        _bridge.seedEscrow(_amount);
    }
}

/// @title CeloGasBridgeL1_Version_Test
/// @notice Tests the `version` function of the `CeloGasBridgeL1` contract.
contract CeloGasBridgeL1_Version_Test is CeloGasBridgeL1_TestInit {
    function test_version_succeeds() external view {
        assertEq(bridge.version(), "1.0.0");
    }
}

/// @title CeloGasBridgeL1_Constructor_Test
/// @notice Tests the constructor behavior of the `CeloGasBridgeL1` contract.
contract CeloGasBridgeL1_Constructor_Test is CeloGasBridgeL1_TestInit {
    function test_constructor_initializeImplementation_reverts() external {
        vm.expectRevert("Initializable: contract is already initialized");
        implementation.initialize(
            ICrossDomainMessenger(address(messenger)),
            ISystemConfig(address(systemConfig)),
            StandardBridge(otherBridge),
            IERC20(address(celoTokenL1))
        );
    }
}

/// @title CeloGasBridgeL1_Initialize_Test
/// @notice Tests the `initialize` function of the `CeloGasBridgeL1` contract.
contract CeloGasBridgeL1_Initialize_Test is CeloGasBridgeL1_TestInit {
    function test_initialize_succeeds() external view {
        assertEq(address(bridge.celoTokenL1()), address(celoTokenL1));
        assertEq(address(bridge.systemConfig()), address(systemConfig));
        assertEq(address(bridge.MESSENGER()), address(messenger));
        assertEq(address(bridge.messenger()), address(messenger));
        assertEq(address(bridge.OTHER_BRIDGE()), otherBridge);
        assertEq(address(bridge.otherBridge()), otherBridge);
        assertEq(bridge.deposits(address(celoTokenL1), address(0)), 0);
        assertFalse(bridge.escrowSeeded());
    }

    function test_initialize_alreadyInitialized_reverts() external {
        vm.prank(address(proxyAdminContract));
        vm.expectRevert("Initializable: contract is already initialized");
        bridge.initialize(
            ICrossDomainMessenger(address(messenger)),
            ISystemConfig(address(systemConfig)),
            IStandardBridge(otherBridge),
            IERC20(address(celoTokenL1))
        );
    }
}

/// @title CeloGasBridgeL1_SeedEscrow_Test
/// @notice Tests the `seedEscrow` function of the `CeloGasBridgeL1` contract.
contract CeloGasBridgeL1_SeedEscrow_Test is CeloGasBridgeL1_TestInit {
    function test_seedEscrow_succeeds() external {
        uint256 amount = 500;

        vm.expectEmit(address(bridge));
        emit EscrowSeeded(optimismPortal, amount);

        _seedEscrow(bridge, amount);

        assertEq(bridge.deposits(address(celoTokenL1), address(0)), amount);
        assertTrue(bridge.escrowSeeded());
    }

    function test_seedEscrow_solvent_succeeds() external {
        uint256 amount = 700;
        celoTokenL1.mint(address(bridge), amount);

        _seedEscrow(bridge, amount);

        assertEq(bridge.deposits(address(celoTokenL1), address(0)), celoTokenL1.balanceOf(address(bridge)));
        assertEq(celoTokenL1.balanceOf(address(bridge)), amount);
    }
}

/// @title CeloGasBridgeL1_SeedEscrow_TestFail
/// @notice Tests revert cases for the `seedEscrow` function of the `CeloGasBridgeL1` contract.
contract CeloGasBridgeL1_SeedEscrow_TestFail is CeloGasBridgeL1_TestInit {
    function test_seedEscrow_unauthorizedSeeder_reverts() external {
        vm.prank(alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_UnauthorizedSeeder.selector);
        bridge.seedEscrow(1);
    }

    function test_seedEscrow_alreadySeeded_reverts() external {
        _seedEscrow(bridge, 1);

        vm.prank(optimismPortal);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_EscrowAlreadySeeded.selector);
        bridge.seedEscrow(2);
    }
}

/// @title CeloGasBridgeL1_SeedEscrowGenesis_Test
/// @notice Tests the `seedEscrowGenesis` function of the `CeloGasBridgeL1` contract.
contract CeloGasBridgeL1_SeedEscrowGenesis_Test is CeloGasBridgeL1_TestInit {
    function test_seedEscrowGenesis_succeeds() external {
        uint256 amount = 1_000;
        celoTokenL1.mint(address(bridge), amount);

        vm.expectEmit(address(bridge));
        emit EscrowSeeded(address(proxyAdminContract), amount);

        vm.prank(address(proxyAdminContract));
        bridge.seedEscrowGenesis();

        assertTrue(bridge.escrowSeeded());
        assertEq(bridge.deposits(address(celoTokenL1), address(0)), amount);
        assertEq(bridge.deposits(address(celoTokenL1), address(0)), celoTokenL1.balanceOf(address(bridge)));
    }

    function test_seedEscrowGenesis_asProxyAdminOwner_succeeds() external {
        uint256 amount = 2_500;
        celoTokenL1.mint(address(bridge), amount);

        vm.expectEmit(address(bridge));
        emit EscrowSeeded(proxyAdminOwnerAddr, amount);

        vm.prank(proxyAdminOwnerAddr);
        bridge.seedEscrowGenesis();

        assertTrue(bridge.escrowSeeded());
        assertEq(bridge.deposits(address(celoTokenL1), address(0)), amount);
    }

    function test_seedEscrowGenesis_thenDeposit_succeeds() external {
        // A deposit reverts before genesis seeding; genesis-seed activates the bridge so deposits work.
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_NotActivated.selector);
        bridge.deposit(bob, 1, 1, hex"");

        vm.prank(address(proxyAdminContract));
        bridge.seedEscrowGenesis();

        uint256 amount = 100;
        celoTokenL1.mint(alice, amount);
        vm.prank(alice, alice);
        celoTokenL1.approve(address(bridge), amount);

        vm.prank(alice, alice);
        bridge.deposit(bob, amount, 250_000, hex"1234");

        assertEq(celoTokenL1.balanceOf(alice), 0);
        assertEq(bridge.deposits(address(celoTokenL1), address(0)), amount);
        assertEq(messenger.lastTarget(), otherBridge);
    }
}

/// @title CeloGasBridgeL1_SeedEscrowGenesis_TestFail
/// @notice Tests revert cases for the `seedEscrowGenesis` function of the `CeloGasBridgeL1` contract.
contract CeloGasBridgeL1_SeedEscrowGenesis_TestFail is CeloGasBridgeL1_TestInit {
    function test_seedEscrowGenesis_unauthorized_reverts() external {
        vm.prank(alice);
        vm.expectRevert(IProxyAdminOwnedBase.ProxyAdminOwnedBase_NotProxyAdminOrProxyAdminOwner.selector);
        bridge.seedEscrowGenesis();
    }

    function test_seedEscrowGenesis_alreadySeeded_reverts() external {
        vm.prank(address(proxyAdminContract));
        bridge.seedEscrowGenesis();

        vm.prank(address(proxyAdminContract));
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_EscrowAlreadySeeded.selector);
        bridge.seedEscrowGenesis();
    }

    function test_seedEscrowGenesis_afterSeedEscrow_reverts() external {
        // Mutual exclusion: once the migration path seeds escrow, genesis seeding is blocked.
        _seedEscrow(bridge, 1);

        vm.prank(address(proxyAdminContract));
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_EscrowAlreadySeeded.selector);
        bridge.seedEscrowGenesis();
    }

    function test_seedEscrow_afterSeedEscrowGenesis_reverts() external {
        // Mutual exclusion: once genesis-seeded, the migration seedEscrow path is blocked.
        vm.prank(address(proxyAdminContract));
        bridge.seedEscrowGenesis();

        vm.prank(optimismPortal);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_EscrowAlreadySeeded.selector);
        bridge.seedEscrow(1);
    }
}

/// @title CeloGasBridgeL1_Paused_Test
/// @notice Tests the `paused` getter of the `CeloGasBridgeL1` contract.
contract CeloGasBridgeL1_Paused_Test is CeloGasBridgeL1_TestInit {
    function test_paused_succeeds() external {
        assertFalse(bridge.paused());

        systemConfig.setPaused(true);
        assertTrue(bridge.paused());
    }
}

/// @title CeloGasBridgeL1_Deposit_Test
/// @notice Tests the `deposit` function of the `CeloGasBridgeL1` contract.
contract CeloGasBridgeL1_Deposit_Test is CeloGasBridgeL1_TestInit {
    event ERC20BridgeInitiated(
        address indexed localToken,
        address indexed remoteToken,
        address indexed from,
        address to,
        uint256 amount,
        bytes extraData
    );

    function setUp() public override {
        super.setUp();
        // Deposits are gated on escrow seeding (CGT v2 activation).
        _seedEscrow(bridge, 0);
    }

    function test_deposit_notActivated_reverts() external {
        ICeloGasBridgeL1 unseededBridge = _deployBridge();

        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_NotActivated.selector);
        unseededBridge.deposit(bob, 1, 1, hex"");
    }

    function test_deposit_succeeds() external {
        bytes memory extraData = hex"1234";
        uint256 amount = 100;
        uint32 minGasLimit = 250_000;

        celoTokenL1.mint(alice, amount);

        vm.prank(alice, alice);
        celoTokenL1.approve(address(bridge), amount);

        vm.expectEmit(address(bridge));
        emit ERC20BridgeInitiated(address(celoTokenL1), CeloPredeploys.GOLD_TOKEN, alice, bob, amount, extraData);

        vm.prank(alice, alice);
        bridge.deposit(bob, amount, minGasLimit, extraData);

        assertEq(celoTokenL1.balanceOf(alice), 0);
        assertEq(celoTokenL1.balanceOf(address(bridge)), amount);
        assertEq(bridge.deposits(address(celoTokenL1), address(0)), amount);

        assertEq(messenger.lastTarget(), otherBridge);
        assertEq(
            messenger.lastMessage(),
            abi.encodeCall(ICeloGasBridgeL2Finalizer.finalizeDeposit, (alice, bob, amount, extraData))
        );
        assertEq(messenger.lastMinGasLimit(), minGasLimit);
        assertEq(messenger.lastValue(), 0);
    }

    function test_deposit_invalidRecipient_self_reverts() external {
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_InvalidRecipient.selector);
        bridge.deposit(address(bridge), 1, 1, hex"");
    }

    function test_deposit_invalidRecipient_messenger_reverts() external {
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_InvalidRecipient.selector);
        bridge.deposit(address(messenger), 1, 1, hex"");
    }

    function test_deposit_paused_reverts() external {
        systemConfig.setPaused(true);

        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_Paused.selector);
        bridge.deposit(bob, 1, 1, hex"");
    }

    function test_deposit_notCgtMode_reverts() external {
        systemConfig.setIsCustomGasToken(false);

        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_NotCgtMode.selector);
        bridge.deposit(bob, 1, 1, hex"");
    }

    function test_deposit_zeroRecipient_reverts() external {
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_ZeroRecipient.selector);
        bridge.deposit(address(0), 1, 1, hex"");
    }

    function test_deposit_zeroAmount_reverts() external {
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_ZeroAmount.selector);
        bridge.deposit(bob, 0, 1, hex"");
    }

    function test_deposit_fromContract_succeeds() external {
        // Contracts may bridge now that `onlyEOA` is removed; `_to` is explicit.
        uint256 amount = 100;
        celoTokenL1.mint(address(this), amount);
        celoTokenL1.approve(address(bridge), amount);

        bridge.deposit(bob, amount, 250_000, hex"");

        assertEq(bridge.deposits(address(celoTokenL1), address(0)), amount);
        assertEq(messenger.lastTarget(), otherBridge);
    }
}

/// @title CeloGasBridgeL1_FinalizeWithdrawal_Test
/// @notice Tests the `finalizeWithdrawal` function of the `CeloGasBridgeL1` contract.
contract CeloGasBridgeL1_FinalizeWithdrawal_Test is CeloGasBridgeL1_TestInit {
    event ERC20BridgeFinalized(
        address indexed localToken,
        address indexed remoteToken,
        address indexed from,
        address to,
        uint256 amount,
        bytes extraData
    );

    function test_finalizeWithdrawal_succeeds() external {
        uint256 amount = 100;
        bytes memory extraData = hex"1234";
        celoTokenL1.mint(address(bridge), amount);

        // Set the deposits mapping to reflect that this amount was previously bridged.
        vm.store(
            address(bridge),
            keccak256(abi.encode(address(0), keccak256(abi.encode(address(celoTokenL1), uint256(2))))),
            bytes32(amount)
        );

        _asOtherBridge();

        vm.expectEmit(address(bridge));
        emit ERC20BridgeFinalized(address(celoTokenL1), CeloPredeploys.GOLD_TOKEN, alice, bob, amount, extraData);

        bridge.finalizeWithdrawal(alice, bob, amount, extraData);

        assertEq(celoTokenL1.balanceOf(bob), amount);
        assertEq(celoTokenL1.balanceOf(address(bridge)), 0);
    }

    function test_finalizeWithdrawal_paused_reverts() external {
        systemConfig.setPaused(true);

        _asOtherBridge();
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_Paused.selector);
        bridge.finalizeWithdrawal(alice, bob, 1, hex"");
    }

    function test_finalizeWithdrawal_zeroRecipient_reverts() external {
        _asOtherBridge();
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_ZeroRecipient.selector);
        bridge.finalizeWithdrawal(alice, address(0), 1, hex"");
    }

    function test_finalizeWithdrawal_invalidRecipient_self_reverts() external {
        _asOtherBridge();
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_InvalidRecipient.selector);
        bridge.finalizeWithdrawal(alice, address(bridge), 1, hex"");
    }

    function test_finalizeWithdrawal_invalidRecipient_messenger_reverts() external {
        _asOtherBridge();
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_InvalidRecipient.selector);
        bridge.finalizeWithdrawal(alice, address(messenger), 1, hex"");
    }

    function test_finalizeWithdrawal_notOtherBridge_reverts() external {
        vm.expectRevert("StandardBridge: function can only be called from the other bridge");
        bridge.finalizeWithdrawal(alice, bob, 1, hex"");
    }

    function test_finalizeWithdrawal_wrongXDomainSender_reverts() external {
        messenger.setXDomainMessageSender(makeAddr("wrongBridge"));

        vm.prank(address(messenger));
        vm.expectRevert("StandardBridge: function can only be called from the other bridge");
        bridge.finalizeWithdrawal(alice, bob, 1, hex"");
    }

    function test_finalizeWithdrawal_consumesSeededDeposits_succeeds() external {
        // First withdrawal on a freshly-migrated bridge (escrow seeded, CELO in) must work without a prior deposit.
        uint256 seed = 500;
        ICeloGasBridgeL1 seeded = _deployBridge();
        celoTokenL1.mint(address(seeded), seed);
        _seedEscrow(seeded, seed);

        messenger.setXDomainMessageSender(otherBridge);
        vm.prank(address(messenger));
        seeded.finalizeWithdrawal(alice, bob, seed, hex"");

        assertEq(celoTokenL1.balanceOf(bob), seed);
        assertEq(seeded.deposits(address(celoTokenL1), address(0)), 0);
    }
}

/// @title CeloGasBridgeL1_DisabledMethods_Test
/// @notice Tests the disabled inherited `StandardBridge` entrypoints.
/// @dev The non-`virtual` upstream entrypoints are blocked via the virtual `_emit*` hook overrides,
///      which revert before any state change or messenger interaction.
contract CeloGasBridgeL1_DisabledMethods_Test is CeloGasBridgeL1_TestInit {
    function test_bridgeERC20_disabled_reverts() external {
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_Disabled.selector);
        bridge.bridgeERC20(address(1), address(2), 3, 4, hex"");
    }

    function test_bridgeERC20To_disabled_reverts() external {
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_Disabled.selector);
        bridge.bridgeERC20To(address(1), address(2), address(3), 4, 5, hex"");
    }

    function test_bridgeETH_disabled_reverts() external {
        vm.deal(alice, 1);
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_Disabled.selector);
        bridge.bridgeETH{ value: 1 }(1, hex"");
    }

    function test_bridgeETHTo_disabled_reverts() external {
        vm.deal(alice, 1);
        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_Disabled.selector);
        bridge.bridgeETHTo{ value: 1 }(bob, 1, hex"");
    }

    function test_finalizeBridgeETH_disabled_reverts() external {
        vm.deal(address(messenger), 1);
        _asOtherBridge();
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_Disabled.selector);
        bridge.finalizeBridgeETH{ value: 1 }(alice, bob, 1, hex"");
    }

    function test_finalizeBridgeERC20_disabled_reverts() external {
        // Mintable mock so execution reaches `_emitERC20BridgeFinalized` (our defense); the non-mintable
        // branch would underflow on `deposits -= _amount` before getting there.
        address localToken = address(new MockOptimismMintableERC20(address(0xBEEF)));
        _asOtherBridge();
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_Disabled.selector);
        bridge.finalizeBridgeERC20(localToken, address(0xBEEF), alice, bob, 1, hex"");
    }

    function test_receive_disabled_reverts() external {
        vm.deal(alice, 1);

        vm.prank(alice, alice);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_Disabled.selector);
        (bool revertsAsExpected,) = address(bridge).call{ value: 1 }(hex"");
        assertTrue(revertsAsExpected, "expectRevert: call did not revert");
    }
}
