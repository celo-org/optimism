// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {IEspressoTEEVerifier} from "@espresso-tee-contracts/interface/IEspressoTEEVerifier.sol";
import {ISystemConfig} from "interfaces/L1/ISystemConfig.sol";

interface IBatchAuthenticator {
    /// @notice Error thrown when an invalid address (zero address) is provided.
    error InvalidAddress(address contract_);

    /// @notice Error thrown when the fallback batcher caller does not match the expected address.
    error UnauthorizedFallbackBatcher(address sender, address expected);

    /// @notice Error thrown when `setEspressoBatcher` is called with the address
    ///         that is already the currently-active batcher.
    error NoChange(address batcher);

    /// @notice Emitted when a batch info is authenticated.
    event BatchInfoAuthenticated(bytes32 indexed commitment);

    /// @notice Emitted when a signer registration is initiated through this contract.
    event SignerRegistrationInitiated(address indexed caller);

    /// @notice Emitted when the Espresso batcher address is updated. `fromBlock`
    ///         is the L1 block number at which `newEspressoBatcher` becomes the
    ///         authorized batcher.
    event EspressoBatcherUpdated(
        address indexed oldEspressoBatcher,
        address indexed newEspressoBatcher,
        uint64 indexed fromBlock
    );

    /// @notice Emitted when the active batcher is switched.
    event BatcherSwitched(bool indexed activeIsEspresso);

    function authenticateBatchInfo(bytes32 commitment, bytes memory _signature) external;

    function espressoTEEVerifier() external view returns (IEspressoTEEVerifier);

    function nitroValidator() external view returns (address);

    function owner() external view returns (address);

    /// @notice Returns the currently-active Espresso batcher address (the
    ///         `batcher` field of the latest history entry).
    function espressoBatcher() external view returns (address);

    /// @notice Number of entries in the Espresso batcher history.
    function espressoBatcherHistoryLength() external view returns (uint256);

    /// @notice Returns the Espresso batcher history entry at `index`
    ///         (oldest first). Reverts on out-of-bounds index.
    function espressoBatcherAt(uint256 index) external view returns (address batcher, uint64 fromBlock);

    /// @notice Returns the Espresso batcher address that was authorized at
    ///         L1 block `l1Block`. Returns `address(0)` if `l1Block` precedes
    ///         the first entry.
    function espressoBatcherAtBlock(uint64 l1Block) external view returns (address);

    function registerSigner(bytes memory verificationData, bytes memory data) external;

    function activeIsEspresso() external view returns (bool);

    function systemConfig() external view returns (ISystemConfig);

    function setActiveIsEspresso(bool _desired) external;

    function setEspressoBatcher(address _newEspressoBatcher) external;
}
