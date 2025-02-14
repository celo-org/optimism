// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { SuperchainConfig } from "./SuperchainConfig.sol";
import { ISuperchainConfig } from "./interfaces/ISuperchainConfig.sol";
import { Storage } from "src/libraries/Storage.sol";

/// @custom:proxied true
/// @custom:audit none This contracts is not yet audited.
/// @title CeloSuperchainConfig
/// @notice The CeloSuperchainConfig contract is used to manage values that are
/// typically part of the global superchain configuration, but potentially need to
/// be handled differently by Celo.
contract CeloSuperchainConfig is SuperchainConfig {
    /// @notice Enum representing different types of updates for the Celo
    //          extension of SuperchainConfig.
    /// @custom:value SUPERCHAIN_CONFIG  Represents an update to the SuperchainConfig address.
    enum CeloUpdateType {
        SUPERCHAIN_CONFIG
    }

    /// @notice The address of the global OP Superchain SuperchainConfig contract.
    ///         It can only be modified by an upgrade.
    bytes32 public constant SUPERCHAIN_CONFIG_SLOT =
        bytes32(uint256(keccak256("celoSuperchainConfig.superchainConfig")) - 1);

    /// @notice Emitted when configuration of the Celo-specific portion of the
    ///         config is updated.
    /// @param updateType Type of update.
    /// @param data       Encoded update data.
    event CeloConfigUpdate(CeloUpdateType indexed updateType, bytes data);

    /// @notice Constructs the CeloSuperchainConfig contract.
    constructor() {
        initialize({ _guardian: address(0), _paused: false, _superchainConfig: address(0) });
    }

    /// @notice Initializer.
    /// @param _guardian           Address of the guardian, can pause the OptimismPortal.
    /// @param _paused             Initial paused status.
    /// @param _superchainConfig   Address of the global SuperchainConfig.
    function initialize(address _guardian, bool _paused, address _superchainConfig) public initializer {
        _setGuardian(_guardian);
        _setSuperchainConfig(_superchainConfig);
        if (_paused) {
            _pause("Initializer paused");
        } else if (_superchainConfig != address(0)) {
            checkAndPauseIfSuperchainPaused();
        }
    }

    /// @notice Checks whether the Celo system should be paused, while also
    ///         propagating the paused value from Superchain to Celo if
    ///         necessary.
    function checkAndPauseIfSuperchainPaused() public returns (bool paused_) {
        if (ISuperchainConfig(superchainConfig()).paused()) {
            _pause("Superchain paused");
            return true;
        }

        return paused();
    }

    /// @notice Getter for the address of the global SuperchainConfig.
    function superchainConfig() public view returns (address superchainConfig_) {
        superchainConfig_ = Storage.getAddress(SUPERCHAIN_CONFIG_SLOT);
    }

    /// @notice Getter for the current paused status, which depends both on the
    ///         local paused value, and the paused status of Superchain.
    function paused() public view override returns (bool paused_) {
        paused_ = super.paused();
        if (paused_) {
            return paused_;
        }

        if (superchainConfig() != address(0)) {
            paused_ = ISuperchainConfig(superchainConfig()).paused();
        }

        return paused_;
    }

    /// @notice Sets the global SuperchainConfig address. This is only callable
    ///         during initialization, so an upgrade will be required to change this address.
    /// @param _superchainConfig The new SuperchainConfig address.
    function _setSuperchainConfig(address _superchainConfig) internal {
        Storage.setAddress(SUPERCHAIN_CONFIG_SLOT, _superchainConfig);
        emit CeloConfigUpdate(CeloUpdateType.SUPERCHAIN_CONFIG, abi.encode(_superchainConfig));
    }
}
