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
import { ProxyAdminOwnedBase } from "src/L1/ProxyAdminOwnedBase.sol";
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
    /// @notice One epoch in the Espresso-batcher history. The address is the
    ///         authorized Espresso batcher signer starting at L1 block
    ///         `fromBlock`. It remains the authorized batcher until the next
    ///         entry's `fromBlock`, or — for the last entry — indefinitely.
    /// @dev    `address` (20 bytes) + `uint64` (8 bytes) packs into a single
    ///         storage slot.
    struct EspressoBatcherEntry {
        address batcher;
        uint64 fromBlock;
    }

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

    /// @notice Append-only history of authorized Espresso batcher addresses
    ///         and the L1 block at which each became active.
    EspressoBatcherEntry[] internal _espressoBatcherHistory;

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
        if (_espressoBatcherHistory.length == 0) {
            uint64 fromBlock = uint64(block.number);
            _espressoBatcherHistory.push(EspressoBatcherEntry({ batcher: _espressoBatcher, fromBlock: fromBlock }));
            emit EspressoBatcherUpdated(address(0), _espressoBatcher, fromBlock);
        }
    }

    /// @notice Returns the owner of the contract.
    function owner() public view override(IBatchAuthenticator, OwnableUpgradeable) returns (address) {
        return super.owner();
    }

    /// @notice Toggles the active batcher between the Espresso and fallback batcher.
    function switchBatcher() external onlyGuardianOrOwner {
        activeIsEspresso = !activeIsEspresso;
        emit BatcherSwitched(activeIsEspresso);
    }

    /// @notice Updates the Espresso batcher address.
    function setEspressoBatcher(address _newEspressoBatcher) external onlyOwner {
        EspressoBatcherEntry storage last = _espressoBatcherHistory[_espressoBatcherHistory.length - 1];
        address oldEspressoBatcher = last.batcher;
        if (_newEspressoBatcher == oldEspressoBatcher) revert NoChange(_newEspressoBatcher);

        uint64 fromBlock = uint64(block.number);
        _espressoBatcherHistory.push(EspressoBatcherEntry({ batcher: _newEspressoBatcher, fromBlock: fromBlock }));
        emit EspressoBatcherUpdated(oldEspressoBatcher, _newEspressoBatcher, fromBlock);
    }

    /// @notice Returns the currently-active Espresso batcher address.
    function espressoBatcher() public view returns (address) {
        return _espressoBatcherHistory[_espressoBatcherHistory.length - 1].batcher;
    }

    /// @notice Number of entries in the Espresso batcher history.
    function espressoBatcherHistoryLength() external view returns (uint256) {
        return _espressoBatcherHistory.length;
    }

    /// @notice Returns the Espresso batcher history entry at `index`
    ///         (oldest first). Reverts on out-of-bounds index (default
    ///         Solidity array bounds check).
    function espressoBatcherAt(uint256 index) external view returns (address batcher, uint64 fromBlock) {
        EspressoBatcherEntry storage entry = _espressoBatcherHistory[index];
        return (entry.batcher, entry.fromBlock);
    }

    /// @notice Returns the Espresso batcher address that was authorized at
    ///         L1 block `l1Block`. Returns `address(0)` if `l1Block` precedes
    ///         the first entry. Uses binary search; history is monotonically
    ///         non-decreasing by `fromBlock`.
    function espressoBatcherAtBlock(uint64 l1Block) external view returns (address) {
        uint256 len = _espressoBatcherHistory.length;

        if (len == 0) return address(0);
        if (l1Block < _espressoBatcherHistory[0].fromBlock) return address(0);

        // Binary search for the greatest entry with `fromBlock <= l1Block`.
        uint256 lo = 0;
        uint256 hi = len; // exclusive upper bound
        while (lo + 1 < hi) {
            uint256 mid = (lo + hi) >> 1;
            if (_espressoBatcherHistory[mid].fromBlock <= l1Block) {
                lo = mid;
            } else {
                hi = mid;
            }
        }
        return _espressoBatcherHistory[lo].batcher;
    }

    function authenticateBatchInfo(bytes32 _commitment, bytes calldata _signature) external {
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
