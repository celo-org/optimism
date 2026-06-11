// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable-v5/access/OwnableUpgradeable.sol";
import { Checkpoints } from "@openzeppelin/contracts-v5/utils/structs/Checkpoints.sol";
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
    using Checkpoints for Checkpoints.Trace160;

    /// @notice Semantic version.
    /// @custom:semver 1.2.0
    string public constant version = "1.2.0";

    /// @notice Address of the Espresso TEE Verifier contract.
    IEspressoTEEVerifier public espressoTEEVerifier;

    /// @notice Flag indicating which batcher is currently active.
    /// @dev When true the Espresso batcher is active; when false the fallback batcher is active.
    bool public activeIsEspresso;

    /// @notice The SystemConfig contract, used to resolve the fallback batcher address.
    ISystemConfig public systemConfig;

    /// @notice Append-only history of authorized Espresso batcher addresses keyed by the L1 block
    ///         at which each became active.
    /// @dev    `Trace160` is OZ's `(uint96 key, uint160 value)` checkpoint variant — `uint160`
    ///         exactly fits an address with no waste, and `uint96` easily covers L1 block numbers.
    ///         An entry remains the authorized batcher until the next entry's key, or — for the
    ///         last entry — indefinitely.
    Checkpoints.Trace160 internal _espressoBatcherHistory;

    /// @notice Constructor disables initializers on implementation
    constructor() ReinitializableBase(1) {
        _disableInitializers();
    }

    /// @notice Initializes the contract.
    function initialize(
        IEspressoTEEVerifier _espressoTEEVerifier,
        address _espressoBatcher,
        ISystemConfig _systemConfig,
        address _owner,
        bool _activeIsEspresso
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
        systemConfig = _systemConfig;
        activeIsEspresso = _activeIsEspresso;

        // Seed the history with the initial Espresso batcher. Skip the append
        // on re-initialization (e.g., a future `initVersion()` bump) so the
        // initializer stays idempotent — appending here would create duplicate
        // history entries and emit a misleading `EspressoBatcherUpdated` event.
        // To update the batcher after deployment, callers must use
        // `setEspressoBatcher`.
        if (_espressoBatcherHistory.length() == 0) {
            uint96 fromBlock = uint96(block.number);
            _espressoBatcherHistory.push(fromBlock, uint160(_espressoBatcher));
            emit EspressoBatcherUpdated(address(0), _espressoBatcher, uint64(fromBlock));
        }
    }

    /// @notice Returns the owner of the contract.
    function owner() public view override(IBatchAuthenticator, OwnableUpgradeable) returns (address) {
        return super.owner();
    }

    /// @notice Sets which batcher is active. Pass `true` to activate the Espresso batcher, or
    ///         `false` to activate the fallback batcher. This is intentionally a setter rather
    ///         than a toggle so that guardian/owner intent is explicit at the call site — the
    ///         caller must name the target mode rather than rely on the contract's current state.
    ///         No-ops (and skips the `BatcherSwitched` event) when `_desired` already matches
    ///         the current state, so off-chain indexers only ever see real transitions.
    function setActiveIsEspresso(bool _desired) external onlyGuardianOrOwner {
        if (activeIsEspresso == _desired) return;
        activeIsEspresso = _desired;
        emit BatcherSwitched(_desired);
    }

    /// @notice Updates the Espresso batcher address.
    function setEspressoBatcher(address _newEspressoBatcher) external onlyOwner {
        address oldEspressoBatcher = espressoBatcher();
        if (_newEspressoBatcher == oldEspressoBatcher) revert NoChange(_newEspressoBatcher);

        uint96 fromBlock = uint96(block.number);
        _espressoBatcherHistory.push(fromBlock, uint160(_newEspressoBatcher));
        emit EspressoBatcherUpdated(oldEspressoBatcher, _newEspressoBatcher, uint64(fromBlock));
    }

    /// @notice Returns the currently-active Espresso batcher address (the value of the most
    ///         recent history entry).
    function espressoBatcher() public view returns (address) {
        return address(_espressoBatcherHistory.latest());
    }

    /// @notice Number of entries in the Espresso batcher history.
    function espressoBatcherHistoryLength() external view returns (uint256) {
        return _espressoBatcherHistory.length();
    }

    /// @notice Returns the Espresso batcher history entry at `_index` (oldest first).
    ///         Reverts on out-of-bounds index.
    function espressoBatcherAt(uint32 _index) external view returns (address batcher_, uint64 fromBlock_) {
        Checkpoints.Checkpoint160 memory ckpt = _espressoBatcherHistory.at(_index);
        return (address(ckpt._value), uint64(ckpt._key));
    }

    /// @notice Returns the Espresso batcher address that was authorized at
    ///         L1 block `_l1Block`. Returns `address(0)` if `_l1Block` precedes
    ///         the first entry.
    function espressoBatcherAtBlock(uint64 _l1Block) external view returns (address) {
        return address(_espressoBatcherHistory.upperLookupRecent(uint96(_l1Block)));
    }

    function authenticateBatchInfo(bytes32 _commitment, bytes calldata _signature) external {
        if (activeIsEspresso) {
            // Espresso batcher path: caller must be the configured espressoBatcher.
            address activeEspressoBatcher = espressoBatcher();
            if (msg.sender != activeEspressoBatcher) {
                revert UnauthorizedEspressoBatcher(msg.sender, activeEspressoBatcher);
            }
            // TEE batcher path: verify via registered TEE signer.
            // Setting TEEType as Nitro because OP integration only supports AWS Nitro currently.
            // `verify` is expected to revert on failure, but we still check the return value as a
            // defensive measure just in case.
            if (!espressoTEEVerifier.verify(_signature, _commitment, IEspressoTEEVerifier.TeeType.NITRO)) {
                revert IEspressoTEEVerifier.InvalidSignature();
            }
        } else {
            // Fallback batcher path: the caller must be the SystemConfig batcher address.
            // No signature verification needed — the transaction itself is already signed by msg.sender.
            address fallbackBatcher = address(uint160(uint256(systemConfig.batcherHash())));
            if (msg.sender != fallbackBatcher) revert UnauthorizedFallbackBatcher(msg.sender, fallbackBatcher);
        }

        emit BatchInfoAuthenticated(_commitment, msg.sender);
    }

    /// @notice Permissionless registration of a TEE-generated signer.
    ///         Anyone may call this; safety is enforced by the verifier:
    ///           1. `verificationData` must contain a valid AWS Nitro attestation, verified via Succinct ZK proof.
    ///           2. The attestation's PCR0 measurement must match an enclave hash pre-approved by the TEE
    ///              verifier's owner/guardian.
    ///           3. The registered signer address is derived from the public key inside the attestation
    ///              — the caller cannot choose it.
    ///         An attacker would need to compromise governance (to whitelist a malicious enclave hash), forge
    ///         an AWS Nitro signature, or break the Succinct ZK proof — all outside the contract's threat model.
    function registerSigner(bytes calldata _verificationData, bytes calldata _data) external {
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
