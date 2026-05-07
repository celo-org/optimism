// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {IEspressoTEEVerifier} from "@espresso-tee-contracts/interface/IEspressoTEEVerifier.sol";
import {ISystemConfig} from "interfaces/L1/ISystemConfig.sol";

interface IBatchAuthenticator {
    /// @notice Error thrown when an invalid address (zero address) is provided.
    error InvalidAddress(address contract_);

    /// @notice Error thrown when the contract is paused.
    error BatchAuthenticator_Paused();

    /// @notice Error thrown when the fallback batcher caller does not match the expected address.
    error UnauthorizedFallbackBatcher(address sender, address expected);

    /// @notice Emitted when a batch info is authenticated.
    event BatchInfoAuthenticated(bytes32 indexed commitment);

    /// @notice Emitted when a signer registration is initiated through this contract.
    event SignerRegistrationInitiated(address indexed caller);

    /// @notice Emitted when the Espresso batcher address is updated.
    event EspressoBatcherUpdated(
        address indexed oldEspressoBatcher,
        address indexed newEspressoBatcher
    );

    /// @notice Emitted when the active batcher is switched.
    event BatcherSwitched(bool indexed activeIsEspresso);

    function authenticateBatchInfo(bytes32 commitment, bytes memory _signature) external;

    function espressoTEEVerifier() external view returns (IEspressoTEEVerifier);

    function nitroValidator() external view returns (address);

    function owner() external view returns (address);

    function espressoBatcher() external view returns (address);

    function registerSigner(bytes memory verificationData, bytes memory data) external;

    function activeIsEspresso() external view returns (bool);

    function systemConfig() external view returns (ISystemConfig);

    function paused() external view returns (bool);

    function switchBatcher() external;

    function setEspressoBatcher(address _newEspressoBatcher) external;
}
