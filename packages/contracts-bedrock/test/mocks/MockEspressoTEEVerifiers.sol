// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { IEspressoTEEVerifier } from "@espresso-tee-contracts/interface/IEspressoTEEVerifier.sol";
import { IEspressoNitroTEEVerifier } from "@espresso-tee-contracts/interface/IEspressoNitroTEEVerifier.sol";
import { INitroEnclaveVerifier } from "aws-nitro-enclave-attestation/interfaces/INitroEnclaveVerifier.sol";
import { ECDSA } from "@openzeppelin/contracts-v5/utils/cryptography/ECDSA.sol";
import { EIP712 } from "@openzeppelin/contracts-v5/utils/cryptography/EIP712.sol";

/// @notice Mock implementation of IEspressoNitroTEEVerifier for testing without real attestation verification.
///         Used by deployment scripts and tests.
contract MockEspressoNitroTEEVerifier is IEspressoNitroTEEVerifier {
    address internal _teeVerifier;
    mapping(address => bool) private _registeredServices;

    constructor() { }

    function isSignerValid(address signer) external view override returns (bool) {
        // Special condition for test TestE2eDevnetWithUnattestedBatcherKey
        if (signer == address(0xe16d5c4080C0faD6D2Ef4eb07C657674a217271C)) {
            return false;
        }
        // If signer was explicitly registered, return true
        if (_registeredServices[signer]) {
            return true;
        }
        // Default permissive behavior for deployment scripts (when no signers registered)
        // This allows the mock to work in both test (explicit registration) and deploy (permissive) modes
        return true;
    }

    function registeredEnclaveHash(bytes32) external pure override returns (bool) {
        return true;
    }

    function registerService(bytes calldata attestation, bytes calldata) external override {
        if (attestation.length >= 20) {
            address signer = address(uint160(bytes20(attestation[:20])));
            _registeredServices[signer] = true;
        }
    }

    function setEnclaveHash(bytes32, bool) external override { }

    function deleteEnclaveHashes(bytes32[] memory) external override { }

    function setNitroEnclaveVerifier(address) external override { }

    function nitroEnclaveVerifier() external pure override returns (INitroEnclaveVerifier) {
        return INitroEnclaveVerifier(address(0));
    }

    function teeVerifier() external view override returns (address) {
        return _teeVerifier;
    }

    /// @notice Test helper to directly set registered signer status.
    function setRegisteredSigner(address signer, bool value) external {
        _registeredServices[signer] = value;
    }
}

/// @notice Mock implementation of IEspressoTEEVerifier for testing.
///         Can optionally wrap a MockEspressoNitroTEEVerifier or act as its own Nitro verifier.
///         Inherits EIP712 to match the real EspressoTEEVerifier's signature verification.
contract MockEspressoTEEVerifier is IEspressoTEEVerifier, IEspressoNitroTEEVerifier, EIP712 {
    IEspressoNitroTEEVerifier private _nitroVerifier;
    mapping(address => bool) private _registeredServices;
    bool private _useExternalNitroVerifier;

    bytes32 private constant ESPRESSO_TEE_VERIFIER_TYPE_HASH = keccak256("EspressoTEEVerifier(bytes32 commitment)");

    /// @notice Constructor that optionally takes an external Nitro verifier.
    /// @param nitroVerifier_ The external Nitro verifier to use. If address(0), acts as standalone.
    constructor(IEspressoNitroTEEVerifier nitroVerifier_) EIP712("EspressoTEEVerifier", "1") {
        if (address(nitroVerifier_) != address(0)) {
            _nitroVerifier = nitroVerifier_;
            _useExternalNitroVerifier = true;
        } else {
            _useExternalNitroVerifier = false;
        }
    }

    // ============ IEspressoTEEVerifier Implementation ============

    function espressoNitroTEEVerifier() external view override returns (IEspressoNitroTEEVerifier) {
        if (_useExternalNitroVerifier) {
            return _nitroVerifier;
        }
        return this;
    }

    function isSignerValid(address signer, TeeType) external view returns (bool) {
        IEspressoNitroTEEVerifier nitroVerifier_ =
            _useExternalNitroVerifier ? _nitroVerifier : IEspressoNitroTEEVerifier(address(this));
        return nitroVerifier_.isSignerValid(signer);
    }

    function verify(
        bytes memory signature,
        bytes32 userDataHash,
        TeeType teeType
    )
        external
        view
        override
        returns (bool)
    {
        if (teeType != TeeType.NITRO) {
            revert InvalidSignature();
        }
        bytes32 structHash = keccak256(abi.encode(ESPRESSO_TEE_VERIFIER_TYPE_HASH, userDataHash));
        bytes32 digest = _hashTypedDataV4(structHash);
        address signer = ECDSA.recover(digest, signature);
        IEspressoNitroTEEVerifier nitroVerifier_ =
            _useExternalNitroVerifier ? _nitroVerifier : IEspressoNitroTEEVerifier(address(this));
        if (!nitroVerifier_.isSignerValid(signer)) {
            revert InvalidSignature();
        }
        return true;
    }

    function registerService(bytes calldata attestation, bytes calldata, TeeType teeType) external override {
        require(teeType == TeeType.NITRO, "MockEspressoTEEVerifier: only NITRO supported");
        if (attestation.length >= 20) {
            address signer = address(uint160(bytes20(attestation[:20])));
            _registeredServices[signer] = true;
        }
    }

    function registeredEnclaveHashes(bytes32, TeeType) external pure override returns (bool) {
        return false;
    }

    function setEspressoNitroTEEVerifier(IEspressoNitroTEEVerifier verifier) external override {
        _nitroVerifier = verifier;
        _useExternalNitroVerifier = address(verifier) != address(0);
    }

    function setEnclaveHash(bytes32, bool, TeeType) external override { }

    function deleteEnclaveHashes(bytes32[] memory, TeeType) external override { }

    function setNitroEnclaveVerifier(address) external override(IEspressoNitroTEEVerifier, IEspressoTEEVerifier) { }

    // ============ IEspressoNitroTEEVerifier Implementation (for standalone mode) ============

    function isSignerValid(address signer) external view override returns (bool) {
        return _registeredServices[signer];
    }

    function registeredEnclaveHash(bytes32) external pure override returns (bool) {
        return false;
    }

    function registerService(bytes calldata attestation, bytes calldata) external override {
        if (attestation.length >= 20) {
            address signer = address(uint160(bytes20(attestation[:20])));
            _registeredServices[signer] = true;
        }
    }

    function setEnclaveHash(bytes32, bool) external pure override { }

    function deleteEnclaveHashes(bytes32[] memory) external pure override { }

    function nitroEnclaveVerifier() external pure override returns (INitroEnclaveVerifier) {
        return INitroEnclaveVerifier(address(0));
    }

    function teeVerifier() external view override returns (address) {
        return address(this);
    }

    // ============ Test Helpers ============

    /// @notice Test helper to directly set registered signer status.
    function setRegisteredSigner(address signer, bool value) external {
        if (value) {
            _registeredServices[signer] = true;
        } else {
            revert("MockEspressoTEEVerifier: unregistering not supported");
        }
    }
}
