// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Initializable } from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import { ISemver } from "src/universal/interfaces/ISemver.sol";
import { ISuperchainConfig } from "./interfaces/ISuperchainConfig.sol";
import { Storage } from "src/libraries/Storage.sol";

/// @custom:proxied true
/// @custom:audit none This contracts is not yet audited.
/// @title CeloSuperchainConfig
/// @notice The CeloSuperchainConfig contract is used to manage values that are
/// typically part of the global superchain configuration, but potentially need to
/// be handled differently by Celo.
contract CeloSuperchainConfig is Initializable, ISemver {
    /// @notice Enum representing different types of updates.
    /// @custom:value GUARDIAN           Represents an update to the guardian.
    /// @custom:value SUPERCHAIN_CONFIG  Represents an update to the SuperchainConfig address.
    enum UpdateType {
        GUARDIAN,
        SUPERCHAIN_CONFIG
    }

    /// @notice Whether or not the Celo system is paused.
    bytes32 public constant PAUSED_SLOT = bytes32(uint256(keccak256("celoSuperchainConfig.paused")) - 1);

    /// @notice The address of the guardian, which can pause withdrawals from the System.
    ///         It can only be modified by an upgrade.
    bytes32 public constant GUARDIAN_SLOT = bytes32(uint256(keccak256("celoSuperchainConfig.guardian")) - 1);

    /// @notice The address of the global OP Superchain SuperchainConfig contract.
    ///         It can only be modified by an upgrade.
    bytes32 public constant SUPERCHAIN_CONFIG_SLOT =
        bytes32(uint256(keccak256("celoSuperchainConfig.superchainConfig")) - 1);
    /// @notice Emitted when the pause is triggered.
    /// @param identifier A string helping to identify provenance of the pause transaction.

    event Paused(string identifier);

    /// @notice Emitted when the pause is lifted.
    event Unpaused();

    /// @notice Emitted when configuration of the Celo-specific portion of the
    ///         config is updated.
    /// @param updateType Type of update.
    /// @param data       Encoded update data.
    event ConfigUpdate(UpdateType indexed updateType, bytes data);

    /// @notice Semantic version.
    /// @custom:semver 1.0.0-beta.1
    string public constant version = "1.0.0-beta.1";

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
        address superchainConfig_ = superchainConfig();
        if (superchainConfig_ != address(0) && ISuperchainConfig(superchainConfig_).paused()) {
            _pause("Superchain paused");
            return true;
        }

        return paused();
    }

    /// @notice Getter for the address of the global SuperchainConfig.
    function superchainConfig() public view returns (address superchainConfig_) {
        superchainConfig_ = Storage.getAddress(SUPERCHAIN_CONFIG_SLOT);
    }

    /// @notice Getter for the guardian address.
    function guardian() public view returns (address guardian_) {
        guardian_ = Storage.getAddress(GUARDIAN_SLOT);
    }

    /// @notice Getter for the current paused status, which depends both on the
    ///         local paused value, and the paused status of Superchain.
    function paused() public view returns (bool paused_) {
        paused_ = celoPaused();
        if (paused_) {
            return paused_;
        }

        if (superchainConfig() != address(0)) {
            paused_ = ISuperchainConfig(superchainConfig()).paused();
        }

        return paused_;
    }

    /// @notice Pauses withdrawals.
    /// @param _identifier (Optional) A string to identify provenance of the pause transaction.
    function pause(string memory _identifier) external {
        require(msg.sender == guardian(), "CeloSuperchainConfig: only guardian can pause");
        _pause(_identifier);
    }

    /// @notice Pauses withdrawals.
    /// @param _identifier (Optional) A string to identify provenance of the pause transaction.
    function _pause(string memory _identifier) internal {
        Storage.setBool(PAUSED_SLOT, true);
        emit Paused(_identifier);
    }

    /// @notice Unpauses withdrawals.
    function unpause() external {
        require(msg.sender == guardian(), "CeloSuperchainConfig: only guardian can unpause");
        Storage.setBool(PAUSED_SLOT, false);
        emit Unpaused();
    }

    /// @notice Getter for the current local paused value. Note that the actual paused status of the
    ///         Celo system also depends on the paused status of the Superchain. Use `paused()` to
    ///         get that.
    function celoPaused() public view returns (bool paused_) {
        paused_ = Storage.getBool(PAUSED_SLOT);
    }

    /// @notice Sets the guardian address. This is only callable during initialization, so an upgrade
    ///         will be required to change the guardian.
    /// @param _guardian The new guardian address.
    function _setGuardian(address _guardian) internal {
        Storage.setAddress(GUARDIAN_SLOT, _guardian);
        emit ConfigUpdate(UpdateType.GUARDIAN, abi.encode(_guardian));
    }

    /// @notice Sets the global SuperchainConfig address. This is only callable
    ///         during initialization, so an upgrade will be required to change this address.
    /// @param _superchainConfig The new SuperchainConfig address.
    function _setSuperchainConfig(address _superchainConfig) internal {
        Storage.setAddress(SUPERCHAIN_CONFIG_SLOT, _superchainConfig);
        emit ConfigUpdate(UpdateType.SUPERCHAIN_CONFIG, abi.encode(_superchainConfig));
    }
}
