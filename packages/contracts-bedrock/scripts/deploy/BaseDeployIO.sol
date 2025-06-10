// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import { Script } from "forge-std/Script.sol";
import { Vm } from "forge-std/Vm.sol";

/// @title BaseDeployer
/// @notice This contract is responsible for deploying all of the contracts in the system.
/// It is meant to be used as a library, and is not meant to be deployed itself.
contract BaseDeployIO is Script {
    /// @notice The address of the `ProxyAdmin` contract.
    address internal proxyAdmin;

    /// @notice The address of the `L1CrossDomainMessenger` proxy contract.
    address internal l1CrossDomainMessengerProxy;

    /// @notice The address of the `L1StandardBridge` proxy contract.
    address internal l1StandardBridgeProxy;

    /// @notice The address of the `L2OutputOracle` proxy contract.
    address internal l2OutputOracleProxy;

    /// @notice The address of the `OptimismPortal` proxy contract.
    address internal optimismPortalProxy;

    /// @notice The address of the `SystemConfig` proxy contract.
    address internal systemConfigProxy;

    /// @notice The address of the `OptimismMintableERC20Factory` proxy contract.
    address internal optimismMintableERC20FactoryProxy;

    /// @notice The address of the `L1ERC721Bridge` proxy contract.
    address internal l1ERC721BridgeProxy;

    /// @notice The address of the `DisputeGameFactory` proxy contract.
    address internal disputeGameFactoryProxy;

    /// @notice The address of the `AnchorStateRegistry` proxy contract.
    address internal anchorStateRegistryProxy;

    /// @notice The address of the `DelayedWETH` proxy contract.
    address internal delayedWETHProxy;

    /// @notice The address of the permissioned `DelayedWETH` proxy contract.
    address internal permissionedDelayedWETHProxy;

    /// @notice The address of the `ProtocolVersions` proxy contract.
    address internal protocolVersionsProxy;

    /// @notice The address of the `SuperchainConfig` proxy contract.
    address internal superchainConfigProxy;

    /// @notice The address of the `CeloSuperchainConfig` proxy contract.
    address internal celoSuperchainConfigProxy;
}
