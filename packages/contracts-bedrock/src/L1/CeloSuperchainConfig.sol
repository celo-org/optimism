// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { SuperchainConfig } from "./SuperchainConfig.sol";
import { Storage } from "src/libraries/Storage.sol";

/// @custom:proxied true
/// @custom:audit none This contracts is not yet audited.
/// @title CeloSuperchainConfig
/// @notice The CeloSuperchainConfig contract is used to manage values that are
/// typically part of the global superchain configuration, but potentially need to
/// be handled differently by Celo.
contract CeloSuperchainConfig is SuperchainConfig {
    /// @notice The address of the global OP Superchain SuperchainConfig contract.
    ///         It can only be modified by an upgrade.
    bytes32 public constant SUPERCHAIN_CONFIG_SLOT =
        bytes32(uint256(keccak256("superchainConfig.superchainConfig")) - 1);

    constructor() {
        initialize({ _guardian: address(0), _paused: false, _superchainConfig: address(0) });
    }

    /// @notice Initializer.
    /// @param _guardian    Address of the guardian, can pause the OptimismPortal.
    /// @param _paused      Initial paused status.
    /// @param _superchainConfig      Address of the global SuperchainConfig.
    function initialize(address _guardian, bool _paused, address _superchainConfig) public initializer {
        _setGuardian(_guardian);
        _setSuperchainConfig(_superchainConfig);
        if (_paused) {
            _pause("Initializer paused");
        }
    }

    function _setSuperchainConfig(address _superchainConfig) internal {
        Storage.setAddress(SUPERCHAIN_CONFIG_SLOT, _superchainConfig);
    }
}
