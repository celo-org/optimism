// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable-v5/access/OwnableUpgradeable.sol";
import { ECDSA } from "@openzeppelin/contracts-v5/utils/cryptography/ECDSA.sol";
import { ISemver } from "interfaces/universal/ISemver.sol";
// espresso: use direct paths (not @espresso-tee-contracts/ remapping) so that Foundry's
// context-specific remappings correctly apply to files within lib/espresso-tee-contracts/.
import { IEspressoTEEVerifier } from "lib/espresso-tee-contracts/src/interface/IEspressoTEEVerifier.sol";
import { IBatchAuthenticator } from "interfaces/L1/IBatchAuthenticator.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { OwnableWithGuardiansUpgradeable } from "lib/espresso-tee-contracts/src/OwnableWithGuardiansUpgradeable.sol";
import { ProxyAdminOwnedBase } from "src/universal/ProxyAdminOwnedBase.sol";
import { ReinitializableBase } from "src/universal/ReinitializableBase.sol";

/// @notice Upgradeable contract that authenticates batch information using the Transparent Proxy
///         pattern.
///         Supports switching between Espresso and fallback batchers.
contract BatchAuthenticator is
    IBatchAuthenticator,
    ISemver,
    OwnableWithGuardiansUpgradeable,
    ProxyAdminOwnedBase,
    ReinitializableBase
{
    /// @notice Semantic version.
    /// @custom:semver 1.2.0
    string public constant version = "1.2.0";

    /// @notice Address of the Espresso batcher whose signatures may authenticate batches.
    address public espressoBatcher;

    /// @notice Address of the Espresso TEE Verifier contract.
    IEspressoTEEVerifier public espressoTEEVerifier;

    /// @notice Flag indicating which batcher is currently active.
    /// @dev When true the Espresso batcher is active; when false the fallback batcher is active.
    bool public activeIsEspresso;

    /// @notice The SystemConfig contract, used to check the paused status.
    ISystemConfig public systemConfig;

    /// @notice Constructor disables initializers on implementation
    constructor() ReinitializableBase(1) {
        _disableInitializers();
    }

    function initialize(
        IEspressoTEEVerifier _espressoTEEVerifier,
        address _espressoBatcher,
        ISystemConfig _systemConfig,
        address _owner
    )
        external
        reinitializer(initVersion())
    {
        // Initialization transactions must come from the ProxyAdmin or its owner.
        _assertOnlyProxyAdminOrProxyAdminOwner();

        // Initialize OwnableWithGuardians with the provided owner address
        __OwnableWithGuardians_init(_owner);

        if (_espressoBatcher == address(0)) revert InvalidAddress(_espressoBatcher);
        if (address(_systemConfig) == address(0)) revert InvalidAddress(address(_systemConfig));
        if (address(_espressoTEEVerifier) == address(0)) {
            revert InvalidAddress(address(_espressoTEEVerifier));
        }

        espressoTEEVerifier = _espressoTEEVerifier;
        espressoBatcher = _espressoBatcher;
        systemConfig = _systemConfig;
        // By default, start with the Espresso batcher active.
        activeIsEspresso = true;
    }

    /// @notice Returns the owner of the contract.
    function owner() public view override(IBatchAuthenticator, OwnableUpgradeable) returns (address) {
        return super.owner();
    }

    /// @notice Getter for the current paused status.
    function paused() public view returns (bool) {
        return systemConfig.paused();
    }

    /// @notice Toggles the active batcher between the Espresso and fallback batcher.
    function switchBatcher() external onlyGuardianOrOwner {
        activeIsEspresso = !activeIsEspresso;
        emit BatcherSwitched(activeIsEspresso);
    }

    /// @notice Updates the Espresso batcher address.
    function setEspressoBatcher(address _newEspressoBatcher) external onlyOwner {
        if (_newEspressoBatcher == address(0)) revert InvalidAddress(_newEspressoBatcher);
        address oldEspressoBatcher = espressoBatcher;
        espressoBatcher = _newEspressoBatcher;
        emit EspressoBatcherUpdated(oldEspressoBatcher, _newEspressoBatcher);
    }

    function authenticateBatchInfo(bytes32 _commitment, bytes calldata _signature) external {
        if (paused()) revert BatchAuthenticator_Paused();

        if (activeIsEspresso) {
            // TEE batcher path: verify via registered TEE signer.
            // Setting TEEType as Nitro because OP integration only supports AWS Nitro currently.
            espressoTEEVerifier.verify(_signature, _commitment, IEspressoTEEVerifier.TeeType.NITRO);
        } else {
            // Fallback batcher path: the caller must be the SystemConfig batcher address.
            // No signature verification needed — the transaction itself is already signed by msg.sender.
            address fallbackBatcher = address(uint160(uint256(systemConfig.batcherHash())));
            if (msg.sender != fallbackBatcher) revert UnauthorizedFallbackBatcher(msg.sender, fallbackBatcher);
        }

        emit BatchInfoAuthenticated(_commitment);
    }

    function registerSigner(bytes calldata _verificationData, bytes calldata _data) external {
        if (paused()) revert BatchAuthenticator_Paused();

        espressoTEEVerifier.registerService(_verificationData, _data, IEspressoTEEVerifier.TeeType.NITRO);
        emit SignerRegistrationInitiated(msg.sender);
    }

    /// @notice Returns the address of the Nitro TEE validator.
    function nitroValidator() external view returns (address) {
        return address(espressoTEEVerifier.espressoNitroTEEVerifier());
    }

    // NOTE: This contract only provides authenticateBatchInfo (which emits BatchInfoAuthenticated events)
    // and signer management. Batch authentication is performed off-chain by the derivation pipeline,
    // which scans L1 receipts for BatchInfoAuthenticated events in a lookback window.
    // Batch data is sent as plain transactions to the BatchInbox EOA address.
}
