// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Contracts
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

interface IPortalMigrator {
    error PortalMigrator_AlreadyMigrated();
    error PortalMigrator_NotProxyAdminOwner();

    event BalanceMismatch(uint256 expected, uint256 actual);
    event CeloMigrated(address indexed bridge, uint256 amount);

    function CELO_TOKEN() external view returns (IERC20);
    function CELO_GAS_BRIDGE_L1() external view returns (address);
    function LEGACY_PORTAL_BALANCE() external view returns (uint256);
    function migrate() external;
    function migrated() external view returns (bool);
}
