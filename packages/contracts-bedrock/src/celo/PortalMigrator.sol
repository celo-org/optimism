// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Contracts
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import { StorageSlot } from "@openzeppelin/contracts/utils/StorageSlot.sol";

// Interfaces
import { ICeloGasBridgeL1 } from "interfaces/celo/ICeloGasBridgeL1.sol";
import { IPortalMigrator } from "interfaces/celo/IPortalMigrator.sol";
import { OptimismPortal2 } from "src/L1/OptimismPortal2.sol";

/// @custom:proxied true
/// @title PortalMigrator
/// @notice Temporary OptimismPortal2 implementation that transfers the full CELO balance held by the
///         portal into the L1 gas bridge during the CGT migration window.
contract PortalMigrator is OptimismPortal2, IPortalMigrator {
    using SafeERC20 for IERC20;

    /// @notice CELO token held by the portal before the migration.
    IERC20 public immutable override CELO_TOKEN_L1;

    /// @notice L1 gas bridge that receives the full CELO balance during migration.
    address public immutable override CELO_GAS_BRIDGE_L1;

    /// @notice Exact CELO balance expected to be held by the portal when migration executes.
    uint256 public immutable override LEGACY_PORTAL_BALANCE;

    /// @notice EIP-1967 storage slot for migration flag.
    /// @dev keccak256("celo.op.portal.migrated") - 1
    bytes32 internal constant MIGRATED_SLOT = bytes32(uint256(keccak256("celo.op.portal.migrated")) - 1);

    /// @param _proofMaturityDelaySeconds Proof maturity delay passed through to OptimismPortal2.
    /// @param _celoTokenL1                 CELO ERC-20 held by the portal.
    /// @param _gasBridge                 Destination L1 gas bridge.
    /// @param _legacyPortalBalance       Exact CELO balance expected during migration.
    constructor(
        uint256 _proofMaturityDelaySeconds,
        IERC20 _celoTokenL1,
        address _gasBridge,
        uint256 _legacyPortalBalance
    )
        OptimismPortal2(_proofMaturityDelaySeconds)
    {
        CELO_TOKEN_L1 = _celoTokenL1;
        CELO_GAS_BRIDGE_L1 = _gasBridge;
        LEGACY_PORTAL_BALANCE = _legacyPortalBalance;
    }

    /// @notice One time drain of the portal's full CELO balance to the L1 gas bridge. Callable only
    ///         by the proxy admin owner. Drains whatever balance is held at execution time: the
    ///         migration may be signed days before it executes, during which the balance can drift.
    ///         A drift from LEGACY_PORTAL_BALANCE does NOT revert — it emits a BalanceMismatch
    ///         warning to be reconciled manually post-migration.
    function migrate() external {
        // Check ProxyAdmin ownership.
        if (msg.sender != proxyAdminOwner()) {
            revert PortalMigrator_NotProxyAdminOwner();
        }

        // Check migration flag.
        if (StorageSlot.getBooleanSlot(MIGRATED_SLOT).value) {
            revert PortalMigrator_AlreadyMigrated();
        }

        // Drain the full balance.
        uint256 balance = CELO_TOKEN_L1.balanceOf(address(this));
        if (balance != LEGACY_PORTAL_BALANCE) {
            emit BalanceMismatch(LEGACY_PORTAL_BALANCE, balance);
        }

        // Migrate funds & seed the bridge escrow with the drained amount, set migration flag.
        StorageSlot.getBooleanSlot(MIGRATED_SLOT).value = true;
        CELO_TOKEN_L1.safeTransfer(CELO_GAS_BRIDGE_L1, balance);
        ICeloGasBridgeL1(payable(CELO_GAS_BRIDGE_L1)).seedEscrow(balance);
        emit CeloMigrated(CELO_GAS_BRIDGE_L1, balance);
    }

    /// @notice Returns whether the migration has executed.
    function migrated() external view returns (bool migrated_) {
        migrated_ = StorageSlot.getBooleanSlot(MIGRATED_SLOT).value;
    }
}
