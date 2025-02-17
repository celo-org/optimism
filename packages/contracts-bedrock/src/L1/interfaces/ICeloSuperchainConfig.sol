// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface ICeloSuperchainConfig {
    enum UpdateType {
        GUARDIAN,
        SUPERCHAIN_CONFIG
    }

    event ConfigUpdate(UpdateType indexed updateType, bytes data);
    event Paused(string identifier);
    event Unpaused();

    function GUARDIAN_SLOT() external view returns (bytes32);
    function PAUSED_SLOT() external view returns (bytes32);
    function SUPERCHAIN_CONFIG_SLOT() external view returns (bytes32);
    function guardian() external view returns (address guardian_);
    function initialize(address _guardian, bool _paused, address _superchainConfig) external;
    function superchainConfig() external view returns (address superchainConfig_);
    function pause(string memory _identifier) external;
    function paused() external view returns (bool paused_);
    function unpause() external;
    function checkAndPauseIfSuperchainPaused() external returns (bool paused_);

    function __constructor__() external;
}
