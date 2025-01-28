// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { ISuperchainConfig } from "./ISuperchainConfig.sol";

interface ICeloSuperchainConfig is ISuperchainConfig {
    function SUPERCHAIN_CONFIG_SLOT() external view returns (bytes32);
    function initialize(address _guardian, bool _paused, address _superchainConfig) external;
    function superchainConfig() external view returns (address superchainConfig_);
    function pauseIfSuperchainPaused() external;

    function __constructor__() external;
}
