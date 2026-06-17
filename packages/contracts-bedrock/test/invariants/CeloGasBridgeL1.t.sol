// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing
import { Test } from "forge-std/Test.sol";
import { StdInvariant } from "forge-std/StdInvariant.sol";
import { StdUtils } from "forge-std/StdUtils.sol";
import { Vm } from "forge-std/Vm.sol";

// Contracts
import { CeloGasBridgeL1 } from "src/celo/CeloGasBridgeL1.sol";
import { Proxy } from "src/universal/Proxy.sol";
import { StandardBridge } from "src/universal/StandardBridge.sol";

// Libraries
import { Constants } from "src/libraries/Constants.sol";

// Mocks
import { MockERC20, MockCrossDomainMessenger, MockSystemConfig, MockProxyAdmin } from "test/celo/CeloBridgeHelpers.sol";

// Interfaces
import { ICeloGasBridgeL1 } from "interfaces/celo/ICeloGasBridgeL1.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IStandardBridge } from "interfaces/universal/IStandardBridge.sol";
import { ICrossDomainMessenger } from "interfaces/universal/ICrossDomainMessenger.sol";

/// @title CeloGasBridgeL1_Handler
/// @notice Drives randomized deposit / withdrawal / seed / donation against the bridge, tracking ghost
///         sums so the invariant can assert exact escrow accounting.
contract CeloGasBridgeL1_Handler is StdUtils {
    Vm internal immutable vm;
    ICeloGasBridgeL1 internal immutable bridge;
    MockERC20 internal immutable celoTokenL1;
    MockCrossDomainMessenger internal immutable messenger;
    address internal immutable portal;
    address internal immutable otherBridge;

    // Ghost accounting.
    uint256 public seedAmount;
    uint256 public totalDeposited;
    uint256 public totalWithdrawn;
    bool public seeded;

    constructor(
        Vm _vm,
        ICeloGasBridgeL1 _bridge,
        MockERC20 _celoTokenL1,
        MockCrossDomainMessenger _messenger,
        address _portal,
        address _otherBridge
    ) {
        vm = _vm;
        bridge = _bridge;
        celoTokenL1 = _celoTokenL1;
        messenger = _messenger;
        portal = _portal;
        otherBridge = _otherBridge;
    }

    /// @notice One-shot escrow seed by the portal; mints the matching CELO into the bridge.
    function seedEscrow(uint256 _amount) external {
        if (seeded) return;
        _amount = bound(_amount, 0, type(uint96).max);
        celoTokenL1.mint(address(bridge), _amount);
        vm.prank(portal);
        bridge.seedEscrow(_amount);
        seedAmount = _amount;
        seeded = true;
    }

    /// @notice Deposit CELO from a fresh actor (requires escrow seeded).
    function deposit(uint256 _amount) external {
        if (!seeded) return;
        _amount = bound(_amount, 1, type(uint96).max);
        celoTokenL1.mint(address(this), _amount);
        celoTokenL1.approve(address(bridge), _amount);
        bridge.deposit(address(0xBEEF), _amount, 200_000, hex"");
        totalDeposited += _amount;
    }

    /// @notice Finalize a withdrawal via the messenger; bounded to tracked escrow to avoid trivial reverts.
    function finalizeWithdrawal(uint256 _amount) external {
        uint256 tracked = bridge.deposits(address(celoTokenL1), address(0));
        if (tracked == 0) return;
        _amount = bound(_amount, 1, tracked);
        messenger.setXDomainMessageSender(otherBridge);
        vm.prank(address(messenger));
        bridge.finalizeWithdrawal(address(0xBEEF), address(0xCAFE), _amount, hex"");
        totalWithdrawn += _amount;
    }

    /// @notice Direct CELO donation: must never break solvency (proves the `<=` invariant choice).
    function donate(uint256 _amount) external {
        _amount = bound(_amount, 0, type(uint96).max);
        celoTokenL1.mint(address(bridge), _amount);
    }
}

/// @title CeloGasBridgeL1_Solvency_Invariant
/// @notice Escrow accounting stays solvent and exact across any action sequence.
contract CeloGasBridgeL1_Solvency_Invariant is StdInvariant, Test {
    MockERC20 internal celoTokenL1;
    MockSystemConfig internal systemConfig;
    MockCrossDomainMessenger internal messenger;
    MockProxyAdmin internal proxyAdminContract;
    ICeloGasBridgeL1 internal bridge;
    CeloGasBridgeL1_Handler internal handler;

    address internal portal = makeAddr("optimismPortal");
    address payable internal otherBridge = payable(makeAddr("otherBridge"));

    function setUp() public {
        celoTokenL1 = new MockERC20();
        systemConfig = new MockSystemConfig();
        systemConfig.setIsCustomGasToken(true);
        systemConfig.setOptimismPortal(portal);
        messenger = new MockCrossDomainMessenger();
        proxyAdminContract = new MockProxyAdmin(makeAddr("proxyAdminOwner"));

        CeloGasBridgeL1 implementation = new CeloGasBridgeL1();
        Proxy proxy = new Proxy(address(proxyAdminContract));
        bridge = ICeloGasBridgeL1(payable(address(proxy)));
        vm.store(address(bridge), Constants.PROXY_OWNER_ADDRESS, bytes32(uint256(uint160(address(proxyAdminContract)))));
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

        handler = new CeloGasBridgeL1_Handler(vm, bridge, celoTokenL1, messenger, portal, otherBridge);
        targetContract(address(handler));
    }

    /// @notice Tracked escrow never exceeds the bridge's actual CELO balance (no theft; donations safe).
    function invariant_escrowCoveredByBalance() public view {
        assertLe(
            bridge.deposits(address(celoTokenL1), address(0)),
            celoTokenL1.balanceOf(address(bridge)),
            "escrow exceeds CELO balance"
        );
    }

    /// @notice Tracked escrow exactly equals seed + deposits - withdrawals (no leaks either direction).
    function invariant_escrowMatchesGhostAccounting() public view {
        assertEq(
            bridge.deposits(address(celoTokenL1), address(0)),
            handler.seedAmount() + handler.totalDeposited() - handler.totalWithdrawn(),
            "escrow != seed + deposits - withdrawals"
        );
    }
}
