// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing
import { Test } from "forge-std/Test.sol";

// Contracts
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ProxyAdmin } from "src/universal/ProxyAdmin.sol";
import { PortalMigrator } from "src/celo/PortalMigrator.sol";

// Interfaces
import { IPortalMigrator } from "interfaces/celo/IPortalMigrator.sol";

// Libraries
import { Constants } from "src/libraries/Constants.sol";

// Mocks
import { TestERC20 } from "test/mocks/TestERC20.sol";
import { MockSeedBridge } from "test/celo/CeloBridgeHelpers.sol";

/// @title PortalMigrator_TestInit
/// @notice Reusable test initialization for `PortalMigrator` tests.
abstract contract PortalMigrator_TestInit is Test {
    /// @notice Emitted when the actual portal balance differs from the expected legacy balance.
    event BalanceMismatch(uint256 expected, uint256 actual);

    /// @notice Emitted when CELO is migrated to the bridge.
    event CeloMigrated(address indexed bridge, uint256 amount);

    uint256 internal constant PROOF_MATURITY_DELAY_SECONDS = 12;
    uint256 internal constant LEGACY_PORTAL_BALANCE = 100;

    address internal adminOwner;
    address internal bridge;

    ProxyAdmin internal proxyAdmin;
    TestERC20 internal celoToken;
    MockSeedBridge internal mockBridge;
    PortalMigrator internal portalMigrator;

    function setUp() public virtual {
        adminOwner = makeAddr("adminOwner");

        proxyAdmin = new ProxyAdmin(adminOwner);
        celoToken = new TestERC20();
        mockBridge = new MockSeedBridge();
        bridge = address(mockBridge);
        portalMigrator = _deployPortalMigrator(IERC20(address(celoToken)), LEGACY_PORTAL_BALANCE);
        mockBridge.setOptimismPortal(address(portalMigrator));
    }

    function _deployPortalMigrator(IERC20 _celoToken, uint256 _legacyPortalBalance)
        internal
        returns (PortalMigrator portalMigrator_)
    {
        portalMigrator_ = new PortalMigrator(PROOF_MATURITY_DELAY_SECONDS, _celoToken, bridge, _legacyPortalBalance);
        vm.store(
            address(portalMigrator_),
            Constants.PROXY_OWNER_ADDRESS,
            bytes32(uint256(uint160(address(proxyAdmin))))
        );
    }

    function _mintExpectedBalance() internal {
        celoToken.mint(address(portalMigrator), LEGACY_PORTAL_BALANCE);
    }

    function _assertMigrated(address _token, uint256 _amount) internal view {
        assertEq(IERC20(_token).balanceOf(address(portalMigrator)), 0);
        assertEq(IERC20(_token).balanceOf(bridge), _amount);
        assertEq(mockBridge.seededAmount(), _amount);
        assertTrue(mockBridge.escrowSeeded());
        assertTrue(portalMigrator.migrated());
    }
}

/// @title PortalMigrator_Constructor_Test
/// @notice Tests the constructor behavior of `PortalMigrator`.
contract PortalMigrator_Constructor_Test is PortalMigrator_TestInit {
    function test_constructor_succeeds() external view {
        assertEq(portalMigrator.proofMaturityDelaySeconds(), PROOF_MATURITY_DELAY_SECONDS);
        assertEq(address(portalMigrator.CELO_TOKEN()), address(celoToken));
        assertEq(portalMigrator.CELO_GAS_BRIDGE_L1(), bridge);
        assertEq(portalMigrator.LEGACY_PORTAL_BALANCE(), LEGACY_PORTAL_BALANCE);
        assertFalse(portalMigrator.migrated());
    }
}

/// @title PortalMigrator_Migrate_Test
/// @notice Tests the `migrate` function of `PortalMigrator`.
contract PortalMigrator_Migrate_Test is PortalMigrator_TestInit {
    function test_migrate_succeeds() external {
        _mintExpectedBalance();

        vm.expectEmit(address(portalMigrator));
        emit CeloMigrated(bridge, LEGACY_PORTAL_BALANCE);

        vm.prank(adminOwner);
        portalMigrator.migrate();

        _assertMigrated(address(celoToken), LEGACY_PORTAL_BALANCE);
    }

    function test_storageLayout_migratedInUnstructuredSlot() external {
        _mintExpectedBalance();

        vm.prank(adminOwner);
        portalMigrator.migrate();

        // Unstructured slot: keccak256("celo.op.portal.migrated") - 1.
        bytes32 slot = bytes32(uint256(keccak256("celo.op.portal.migrated")) - 1);
        assertEq(vm.load(address(portalMigrator), slot), bytes32(uint256(1)));

        // Inherited sequential range is left untouched (no residue at slot 64).
        assertEq(vm.load(address(portalMigrator), bytes32(uint256(64))), bytes32(0));
    }

    function test_migrate_balanceTooHigh_drainsAndEmits() external {
        uint256 actual = LEGACY_PORTAL_BALANCE + 1;
        celoToken.mint(address(portalMigrator), actual);

        vm.expectEmit(address(portalMigrator));
        emit BalanceMismatch(LEGACY_PORTAL_BALANCE, actual);

        vm.prank(adminOwner);
        portalMigrator.migrate();

        _assertMigrated(address(celoToken), actual);
    }

    function test_migrate_balanceTooLow_drainsAndEmits() external {
        uint256 actual = LEGACY_PORTAL_BALANCE - 1;
        celoToken.mint(address(portalMigrator), actual);

        vm.expectEmit(address(portalMigrator));
        emit BalanceMismatch(LEGACY_PORTAL_BALANCE, actual);

        vm.prank(adminOwner);
        portalMigrator.migrate();

        _assertMigrated(address(celoToken), actual);
    }
}

/// @title PortalMigrator_Migrate_TestFail
/// @notice Tests revert cases for `PortalMigrator.migrate`.
contract PortalMigrator_Migrate_TestFail is PortalMigrator_TestInit {
    function test_migrate_alreadyMigrated_reverts() external {
        _mintExpectedBalance();

        vm.prank(adminOwner);
        portalMigrator.migrate();

        vm.prank(adminOwner);
        vm.expectRevert(IPortalMigrator.PortalMigrator_AlreadyMigrated.selector);
        portalMigrator.migrate();
    }

    function test_migrate_notProxyAdminOwner_reverts() external {
        vm.expectRevert(IPortalMigrator.PortalMigrator_NotProxyAdminOwner.selector);
        portalMigrator.migrate();
    }
}
