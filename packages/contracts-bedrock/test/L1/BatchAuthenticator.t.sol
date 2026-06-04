// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Test } from "test/setup/Test.sol";
import { console2 as console } from "forge-std/console2.sol";
import { Vm } from "forge-std/Vm.sol";

import { BatchAuthenticator } from "src/L1/BatchAuthenticator.sol";
import { IBatchAuthenticator } from "interfaces/L1/IBatchAuthenticator.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { IEspressoTEEVerifier } from "@espresso-tee-contracts/interface/IEspressoTEEVerifier.sol";
import { IEspressoNitroTEEVerifier } from "@espresso-tee-contracts/interface/IEspressoNitroTEEVerifier.sol";
import { EIP1967Helper } from "test/mocks/EIP1967Helper.sol";
import { EspressoTEEVerifierMock } from "@espresso-tee-contracts/mocks/EspressoTEEVerifier.sol";
import { EspressoNitroTEEVerifierMock } from "@espresso-tee-contracts/mocks/EspressoNitroTEEVerifierMock.sol";
import {
    VerifierJournal,
    VerificationResult,
    Pcr
} from "aws-nitro-enclave-attestation/interfaces/INitroEnclaveVerifier.sol";

import { Config } from "scripts/libraries/Config.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IBatchAuthenticator } from "interfaces/L1/IBatchAuthenticator.sol";

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable-v5/access/OwnableUpgradeable.sol";
import { OwnableWithGuardiansUpgradeable } from "lib/espresso-tee-contracts/src/OwnableWithGuardiansUpgradeable.sol";
import { ECDSA } from "@openzeppelin/contracts-v5/utils/cryptography/ECDSA.sol";

/// @notice Minimal mock of SystemConfig that exposes a settable paused() flag
///         and a configurable batcherHash() used by the fallback batcher path.
contract MockSystemConfig {
    bool private _paused;
    bytes32 private _batcherHash;

    function setPaused(bool val) external {
        _paused = val;
    }

    function paused() external view returns (bool) {
        return _paused;
    }

    function setBatcherHash(bytes32 val) external {
        _batcherHash = val;
    }

    function batcherHash() external view returns (bytes32) {
        return _batcherHash;
    }
}

/// @notice Tests for the upgradeable BatchAuthenticator contract using the Transparent Proxy pattern.
contract BatchAuthenticator_Uncategorized_Test is Test {
    address public deployer = address(0xABCD);
    address public proxyAdminOwner = address(0xBEEF);
    address public unauthorized = address(0xDEAD);
    address public guardian = address(0xFACE);

    address public espressoBatcher = address(0x1234);

    MockSystemConfig public mockSystemConfig;
    EspressoTEEVerifierMock public teeVerifier;
    EspressoNitroTEEVerifierMock public nitroVerifier;
    BatchAuthenticator public implementation;
    IProxyAdmin public proxyAdmin;

    bytes32 private constant _ESPRESSO_TEE_VERIFIER_TYPE_HASH = keccak256("EspressoTEEVerifier(bytes32 commitment)");

    bytes32 private constant _EIP712_DOMAIN_TYPE_HASH =
        keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)");

    /// @notice Compute the EIP-712 digest that the TEE verifier mock expects.
    function _computeEIP712Digest(bytes32 commitment) internal view returns (bytes32) {
        bytes32 structHash = keccak256(abi.encode(_ESPRESSO_TEE_VERIFIER_TYPE_HASH, commitment));
        bytes32 domainSeparator = keccak256(
            abi.encode(
                _EIP712_DOMAIN_TYPE_HASH,
                keccak256("EspressoTEEVerifier"),
                keccak256("1"),
                block.chainid,
                address(teeVerifier)
            )
        );
        return keccak256(abi.encodePacked("\x19\x01", domainSeparator, structHash));
    }

    function setUp() public {
        // Deploy the mock SystemConfig.
        mockSystemConfig = new MockSystemConfig();

        // Deploy the mock TEE verifier with a mock Nitro verifier.
        // and the authenticator implementation.
        nitroVerifier = new EspressoNitroTEEVerifierMock();
        teeVerifier = new EspressoTEEVerifierMock(IEspressoNitroTEEVerifier(address(nitroVerifier)));
        implementation = new BatchAuthenticator();

        // Deploy the proxy admin via vm.getCode to avoid duplicate ProxyAdmin artifacts.
        {
            bytes memory _code = vm.getCode("forge-artifacts/ProxyAdmin.sol/ProxyAdmin.json");
            bytes memory _args = abi.encode(proxyAdminOwner);
            bytes memory _initCode = abi.encodePacked(_code, _args);
            address _addr;
            assembly {
                _addr := create(0, add(_initCode, 0x20), mload(_initCode))
            }
            proxyAdmin = IProxyAdmin(_addr);
        }
    }

    function _nitroRegistrationOutputForPrivateKey(uint256 privateKey) internal returns (bytes memory) {
        Vm.Wallet memory wallet = vm.createWallet(privateKey);
        bytes memory publicKey = abi.encodePacked(bytes1(0x04), bytes32(wallet.publicKeyX), bytes32(wallet.publicKeyY));

        VerifierJournal memory journal = VerifierJournal({
            result: VerificationResult.Success,
            trustedCertsPrefixLen: 0,
            timestamp: 0,
            certs: new bytes32[](0),
            userData: new bytes(0),
            nonce: new bytes(0),
            publicKey: publicKey,
            pcrs: new Pcr[](0),
            moduleId: ""
        });

        return abi.encode(journal);
    }

    function _registerNitroSigner(uint256 privateKey) internal {
        nitroVerifier.registerService(_nitroRegistrationOutputForPrivateKey(privateKey), "");
    }

    /// @notice Create and initialize a proxy.
    function _deployAndInitializeProxy() internal returns (BatchAuthenticator) {
        IProxy proxy = _newProxy(address(proxyAdmin));
        vm.prank(proxyAdminOwner);
        proxyAdmin.setProxyType(address(proxy), IProxyAdmin.ProxyType.ERC1967);

        bytes memory initData = abi.encodeCall(
            BatchAuthenticator.initialize,
            (
                IEspressoTEEVerifier(address(teeVerifier)),
                espressoBatcher,
                ISystemConfig(address(mockSystemConfig)),
                proxyAdminOwner,
                // First deployment: start with the Espresso batcher active.
                true
            )
        );
        vm.prank(proxyAdminOwner);
        proxyAdmin.upgradeAndCall(payable(address(proxy)), address(implementation), initData);

        return BatchAuthenticator(address(proxy));
    }

    /// @notice Test that the initialization can only be called once.
    function test_constructor_whenAlreadyInitialized_reverts() external {
        IProxy proxy = _newProxy(address(proxyAdmin));
        vm.prank(proxyAdminOwner);
        proxyAdmin.setProxyType(address(proxy), IProxyAdmin.ProxyType.ERC1967);

        bytes memory initData = abi.encodeCall(
            BatchAuthenticator.initialize,
            (
                IEspressoTEEVerifier(address(teeVerifier)),
                espressoBatcher,
                ISystemConfig(address(mockSystemConfig)),
                proxyAdminOwner,
                true
            )
        );

        // First initialization succeeds.
        vm.prank(proxyAdminOwner);
        proxyAdmin.upgradeAndCall(payable(address(proxy)), address(implementation), initData);

        // Second initialization should revert.
        // Our custom Proxy.upgradeToAndCall wraps delegatecall failures with a fixed string,
        // rather than bubbling up the inner revert (InvalidInitialization). This is a known
        // limitation of src/universal/Proxy.sol vs OZ's TransparentUpgradeableProxy.
        vm.prank(proxyAdminOwner);
        vm.expectRevert("Proxy: delegatecall to new implementation contract failed");
        proxyAdmin.upgradeAndCall(payable(address(proxy)), address(implementation), initData);
    }

    /// @notice Test that initialize reverts when espressoBatcher is zero.
    function test_constructor_whenEspressoBatcherIsZero_reverts() external {
        IProxy proxy = _newProxy(address(proxyAdmin));
        vm.prank(proxyAdminOwner);
        proxyAdmin.setProxyType(address(proxy), IProxyAdmin.ProxyType.ERC1967);

        bytes memory initData = abi.encodeCall(
            BatchAuthenticator.initialize,
            (
                IEspressoTEEVerifier(address(teeVerifier)),
                address(0),
                ISystemConfig(address(mockSystemConfig)),
                proxyAdminOwner,
                true
            )
        );

        vm.prank(proxyAdminOwner);
        vm.expectRevert("Proxy: delegatecall to new implementation contract failed");
        proxyAdmin.upgradeAndCall(payable(address(proxy)), address(implementation), initData);
    }

    /// @notice Test that initialize reverts when verifier is zero.
    function test_constructor_whenVerifierIsZero_reverts() external {
        IProxy proxy = _newProxy(address(proxyAdmin));
        vm.prank(proxyAdminOwner);
        proxyAdmin.setProxyType(address(proxy), IProxyAdmin.ProxyType.ERC1967);

        bytes memory initData = abi.encodeCall(
            BatchAuthenticator.initialize,
            (
                IEspressoTEEVerifier(address(0)),
                espressoBatcher,
                ISystemConfig(address(mockSystemConfig)),
                proxyAdminOwner,
                true
            )
        );

        vm.prank(proxyAdminOwner);
        vm.expectRevert("Proxy: delegatecall to new implementation contract failed");
        proxyAdmin.upgradeAndCall(payable(address(proxy)), address(implementation), initData);
    }

    /// @notice Test that initialize succeeds with valid addresses.
    function test_constructor_withValidAddresses_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        assertEq(address(authenticator.espressoTEEVerifier()), address(teeVerifier));
        assertEq(authenticator.espressoBatcher(), espressoBatcher);
        assertTrue(authenticator.activeIsEspresso());
    }

    /// @notice Test that initialize honors the explicit `_activeIsEspresso` parameter.
    ///         Guards against the non-idempotent-init footgun: if a future `initVersion()` bump
    ///         re-runs `initialize` with `_activeIsEspresso = false`, the contract must reflect
    ///         that — not silently revert to a hardcoded default.
    function test_constructor_respectsActiveIsEspressoFalse() external {
        IProxy proxy = _newProxy(address(proxyAdmin));
        vm.prank(proxyAdminOwner);
        proxyAdmin.setProxyType(address(proxy), IProxyAdmin.ProxyType.ERC1967);

        bytes memory initData = abi.encodeCall(
            BatchAuthenticator.initialize,
            (
                IEspressoTEEVerifier(address(teeVerifier)),
                espressoBatcher,
                ISystemConfig(address(mockSystemConfig)),
                proxyAdminOwner,
                false
            )
        );
        vm.prank(proxyAdminOwner);
        proxyAdmin.upgradeAndCall(payable(address(proxy)), address(implementation), initData);

        assertFalse(BatchAuthenticator(address(proxy)).activeIsEspresso());
    }

    /// @notice Test that setActiveIsEspresso can be called by owner or guardian.
    function test_setActiveIsEspresso_ownerOrGuardian_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // ProxyAdmin owner (now contract owner) can set.
        vm.expectEmit(true, false, false, false);
        emit BatcherSwitched(false);
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(false);
        assertFalse(authenticator.activeIsEspresso());

        // Set back.
        vm.expectEmit(true, false, false, false);
        emit BatcherSwitched(true);
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(true);
        assertTrue(authenticator.activeIsEspresso());

        // Add a guardian.
        vm.prank(proxyAdminOwner);
        authenticator.addGuardian(guardian);
        assertTrue(authenticator.isGuardian(guardian));

        // Guardian can set.
        vm.expectEmit(true, false, false, false);
        emit BatcherSwitched(false);
        vm.prank(guardian);
        authenticator.setActiveIsEspresso(false);
        assertFalse(authenticator.activeIsEspresso());

        // Guardian can set back.
        vm.expectEmit(true, false, false, false);
        emit BatcherSwitched(true);
        vm.prank(guardian);
        authenticator.setActiveIsEspresso(true);
        assertTrue(authenticator.activeIsEspresso());

        // Unauthorized cannot set.
        vm.prank(unauthorized);
        vm.expectRevert(
            abi.encodeWithSelector(OwnableWithGuardiansUpgradeable.NotGuardianOrOwner.selector, unauthorized)
        );
        authenticator.setActiveIsEspresso(false);

        // ProxyAdmin cannot set.
        vm.prank(address(proxyAdmin));
        vm.expectRevert(
            abi.encodeWithSelector(OwnableWithGuardiansUpgradeable.NotGuardianOrOwner.selector, address(proxyAdmin))
        );
        authenticator.setActiveIsEspresso(false);
    }

    /// @notice `setActiveIsEspresso` is a no-op (and emits no event) when the
    ///         desired value already matches the current state.
    function test_setActiveIsEspresso_noChange_noOps() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // Initial state is `activeIsEspresso == true`.
        assertTrue(authenticator.activeIsEspresso());

        // Re-setting to `true` must NOT emit `BatcherSwitched`. `vm.recordLogs`
        // captures every emitted log; asserting zero entries proves no event
        // fired (a narrower `expectEmit(false)` doesn't exist).
        vm.recordLogs();
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(true);
        assertEq(vm.getRecordedLogs().length, 0);
        assertTrue(authenticator.activeIsEspresso());

        // Flip to `false` so we can re-test the no-op from the other state.
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(false);
        assertFalse(authenticator.activeIsEspresso());

        // Re-setting to `false` is also a no-op.
        vm.recordLogs();
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(false);
        assertEq(vm.getRecordedLogs().length, 0);
        assertFalse(authenticator.activeIsEspresso());
    }

    /// @notice Test that authenticateBatchInfo works correctly.
    function test_authenticateBatchInfo_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        uint256 privateKey = 1;
        bytes32 commitment = keccak256("test commitment");

        // Register signer.
        _registerNitroSigner(privateKey);

        // Create signature.
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, _computeEIP712Digest(commitment));
        bytes memory signature = abi.encodePacked(r, s, v);

        // Authenticate.
        vm.expectEmit(true, false, false, true);
        emit BatchInfoAuthenticated(commitment, espressoBatcher);

        vm.prank(espressoBatcher);
        authenticator.authenticateBatchInfo(commitment, signature);
    }

    /// @notice Test that authenticateBatchInfo reverts for unregistered signers.
    function test_authenticateBatchInfo_forUnregisteredSigner_reverts() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        uint256 privateKey = 1;
        bytes32 commitment = keccak256("test commitment");

        // DO NOT register signer - signer is not registered in the TEE verifier

        // Create valid signature from unregistered signer.
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, _computeEIP712Digest(commitment));
        bytes memory signature = abi.encodePacked(r, s, v);

        // Should revert because signer is not registered.
        vm.expectRevert(abi.encodeWithSelector(IEspressoTEEVerifier.InvalidSignature.selector));
        vm.prank(espressoBatcher);
        authenticator.authenticateBatchInfo(commitment, signature);
    }

    /// @notice Test that authenticateBatchInfo reverts for invalid signature (zero address recovery).
    function test_authenticateBatchInfo_forInvalidSignature_reverts() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        bytes32 commitment = keccak256("test commitment");

        // Create an invalid signature that will recover to address(0)
        // 65 bytes: v=0, r=0, s=0 — passes length check, but ecrecover returns address(0)
        bytes memory invalidSignature = new bytes(65);

        // OZ v5 ECDSA.recover reverts with ECDSAInvalidSignature() when ecrecover returns address(0)
        // (not ECDSAInvalidSignatureLength, which only fires when length != 65)
        vm.expectRevert(abi.encodeWithSelector(ECDSA.ECDSAInvalidSignature.selector));
        vm.prank(espressoBatcher);
        authenticator.authenticateBatchInfo(commitment, invalidSignature);
    }

    /// @notice Test that registerSigner works correctly.
    function test_registerSigner_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        uint256 privateKey = 1;
        bytes memory signerData = _nitroRegistrationOutputForPrivateKey(privateKey);
        bytes memory proofBytes = "";

        vm.expectEmit(true, false, false, false);
        emit SignerRegistrationInitiated(address(this));

        authenticator.registerSigner(signerData, proofBytes);
    }

    /// @notice Test that setEspressoBatcher can only be called by ProxyAdmin owner.
    function test_setEspressoBatcher_ownerOnly_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();
        address newEspressoBatcher = address(0x9999);

        // Roll forward so the new entry lands in a new block (avoid same-block overwrite).
        vm.roll(block.number + 1);

        // ProxyAdmin owner can set.
        vm.expectEmit(true, true, true, false);
        emit EspressoBatcherUpdated(espressoBatcher, newEspressoBatcher, uint64(block.number));
        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(newEspressoBatcher);
        assertEq(authenticator.espressoBatcher(), newEspressoBatcher);

        // Unauthorized cannot set.
        vm.prank(unauthorized);
        vm.expectRevert(abi.encodeWithSelector(OwnableUpgradeable.OwnableUnauthorizedAccount.selector, unauthorized));
        authenticator.setEspressoBatcher(address(0x7777));

        // ProxyAdmin cannot set.
        vm.prank(address(proxyAdmin));
        vm.expectRevert(
            abi.encodeWithSelector(OwnableUpgradeable.OwnableUnauthorizedAccount.selector, address(proxyAdmin))
        );
        authenticator.setEspressoBatcher(address(0x8888));
    }

    /// @notice `setEspressoBatcher(address(0))` is allowed and represents an
    ///         explicit revocation without replacement.
    function test_setEspressoBatcher_zeroAddress_revokes() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        vm.roll(block.number + 1);
        uint64 revokeBlock = uint64(block.number);

        vm.expectEmit(true, true, true, false);
        emit EspressoBatcherUpdated(espressoBatcher, address(0), revokeBlock);
        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(address(0));

        assertEq(authenticator.espressoBatcher(), address(0));
        assertEq(authenticator.espressoBatcherHistoryLength(), 2);
    }

    /// @notice `setEspressoBatcher` reverts with `NoChange` when called with
    ///         the value that is already the active batcher.
    function test_setEspressoBatcher_noChange_reverts() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // Replacing with the same non-zero address reverts.
        vm.roll(block.number + 1);
        vm.prank(proxyAdminOwner);
        vm.expectRevert(abi.encodeWithSelector(IBatchAuthenticator.NoChange.selector, espressoBatcher));
        authenticator.setEspressoBatcher(espressoBatcher);

        // Revoking-when-already-revoked also reverts.
        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(address(0));

        vm.roll(block.number + 1);
        vm.prank(proxyAdminOwner);
        vm.expectRevert(abi.encodeWithSelector(IBatchAuthenticator.NoChange.selector, address(0)));
        authenticator.setEspressoBatcher(address(0));
    }

    /// @notice History length is 1 immediately after initialize, with the seed
    ///         entry's `fromBlock` equal to the deployment block.
    function test_history_seededByInitialize() external {
        uint256 deployBlock = block.number;
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        assertEq(authenticator.espressoBatcherHistoryLength(), 1);
        (address b0, uint64 f0) = authenticator.espressoBatcherAt(0);
        assertEq(b0, espressoBatcher);
        assertEq(uint256(f0), deployBlock);
        assertEq(authenticator.espressoBatcher(), espressoBatcher);
    }

    /// @notice Two `setEspressoBatcher` calls in different blocks append two
    ///         new history entries.
    function test_setEspressoBatcher_appendsAcrossBlocks() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        address b1 = address(0x1111);
        address b2 = address(0x2222);

        vm.roll(block.number + 5);
        uint64 f1 = uint64(block.number);
        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(b1);

        vm.roll(block.number + 7);
        uint64 f2 = uint64(block.number);
        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(b2);

        assertEq(authenticator.espressoBatcherHistoryLength(), 3);
        (address a0,) = authenticator.espressoBatcherAt(0);
        (address a1, uint64 ff1) = authenticator.espressoBatcherAt(1);
        (address a2, uint64 ff2) = authenticator.espressoBatcherAt(2);
        assertEq(a0, espressoBatcher);
        assertEq(a1, b1);
        assertEq(uint256(ff1), uint256(f1));
        assertEq(a2, b2);
        assertEq(uint256(ff2), uint256(f2));
        assertEq(authenticator.espressoBatcher(), b2);
    }

    /// @notice Two `setEspressoBatcher` calls in the same L1 block overwrite
    ///         the last entry rather than appending a new one.
    function test_setEspressoBatcher_sameBlockOverwrites() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        address b1 = address(0x1111);
        address b2 = address(0x2222);

        vm.roll(block.number + 1);
        uint64 fBlock = uint64(block.number);

        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(b1);
        // After first call: length=2.
        assertEq(authenticator.espressoBatcherHistoryLength(), 2);

        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(b2);
        // After second call in the same block: still length=2 (overwrite).
        assertEq(authenticator.espressoBatcherHistoryLength(), 2);

        (address a1, uint64 f1) = authenticator.espressoBatcherAt(1);
        assertEq(a1, b2);
        assertEq(uint256(f1), uint256(fBlock));
    }

    /// @notice Revoking then setting a new non-zero address succeeds and
    ///         appends both entries.
    function test_setEspressoBatcher_revokeThenReplace() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        vm.roll(block.number + 1);
        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(address(0));
        assertEq(authenticator.espressoBatcher(), address(0));
        assertEq(authenticator.espressoBatcherHistoryLength(), 2);

        address b1 = address(0x1111);
        vm.roll(block.number + 1);
        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(b1);

        assertEq(authenticator.espressoBatcher(), b1);
        assertEq(authenticator.espressoBatcherHistoryLength(), 3);
    }

    /// @notice `espressoBatcherAtBlock` returns the correct historical address
    ///         across the whole timeline.
    function test_espressoBatcherAtBlock_lookup() external {
        // Move forward a bit so f0 > 0 (lets us test "before first entry").
        vm.roll(block.number + 10);
        uint64 f0 = uint64(block.number);
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // Append b1.
        vm.roll(block.number + 5);
        uint64 f1 = uint64(block.number);
        address b1 = address(0x1111);
        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(b1);

        // Revoke.
        vm.roll(block.number + 4);
        uint64 f2 = uint64(block.number);
        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(address(0));

        // Append b3.
        vm.roll(block.number + 3);
        uint64 f3 = uint64(block.number);
        address b3 = address(0x3333);
        vm.prank(proxyAdminOwner);
        authenticator.setEspressoBatcher(b3);

        // Before the first entry → address(0).
        assertEq(authenticator.espressoBatcherAtBlock(f0 - 1), address(0));

        // At exactly f0 → seed batcher.
        assertEq(authenticator.espressoBatcherAtBlock(f0), espressoBatcher);

        // In [f0, f1) → seed batcher.
        assertEq(authenticator.espressoBatcherAtBlock(f1 - 1), espressoBatcher);

        // In [f1, f2) → b1.
        assertEq(authenticator.espressoBatcherAtBlock(f1), b1);
        assertEq(authenticator.espressoBatcherAtBlock(f2 - 1), b1);

        // In [f2, f3) → address(0) (revoked).
        assertEq(authenticator.espressoBatcherAtBlock(f2), address(0));
        assertEq(authenticator.espressoBatcherAtBlock(f3 - 1), address(0));

        // At and after f3 → b3.
        assertEq(authenticator.espressoBatcherAtBlock(f3), b3);
        assertEq(authenticator.espressoBatcherAtBlock(f3 + 100), b3);
    }

    /// @notice `espressoBatcherAt` reverts on out-of-bounds index. The revert is the
    ///         default Solidity array-out-of-bounds panic (0x32) from `Checkpoints.at`.
    function test_espressoBatcherAt_outOfBounds_reverts() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();
        // length == 1, so index 1 is out of bounds.
        vm.expectRevert(abi.encodeWithSelector(bytes4(0x4e487b71), uint256(0x32)));
        authenticator.espressoBatcherAt(1);
    }

    /// @notice Test upgrade to new implementation with comprehensive state preservation.
    function test_upgrade_preservesState_succeeds() external {
        // Create and initialize a proxy.
        BatchAuthenticator authenticator = _deployAndInitializeProxy();
        IProxy proxy = IProxy(payable(address(authenticator)));

        // Set up initial state.
        bytes32 commitment = keccak256("test commitment");
        uint256 privateKey = 1;
        _registerNitroSigner(privateKey);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, _computeEIP712Digest(commitment));
        bytes memory signature = abi.encodePacked(r, s, v);
        vm.prank(espressoBatcher);
        authenticator.authenticateBatchInfo(commitment, signature);

        // Switch batcher to test boolean flag preservation.
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(false);
        assertFalse(authenticator.activeIsEspresso());

        // Deploy new implementation and upgrade.
        BatchAuthenticator newImpl = new BatchAuthenticator();
        vm.prank(proxyAdminOwner);
        proxyAdmin.upgrade(payable(address(proxy)), address(newImpl));

        // Verify implementation changed.
        address newImplementation = EIP1967Helper.getImplementation(address(proxy));
        assertEq(newImplementation, address(newImpl));

        // Verify state is preserved.
        assertEq(address(authenticator.espressoTEEVerifier()), address(teeVerifier));
        assertEq(authenticator.espressoBatcher(), espressoBatcher);
        assertFalse(authenticator.activeIsEspresso());
    }

    /// @notice Test that authenticateBatchInfo succeeds in fallback mode when called by
    ///         the SystemConfig batcher address.
    function test_authenticateBatchInfo_fallback_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // Switch to fallback mode.
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(false);
        assertFalse(authenticator.activeIsEspresso());

        // Configure the SystemConfig batcher to a known address.
        address fallbackBatcher = address(0xCAFE);
        mockSystemConfig.setBatcherHash(bytes32(uint256(uint160(fallbackBatcher))));

        bytes32 commitment = keccak256("fallback commitment");

        // The fallback batcher path ignores the signature; pass empty bytes.
        vm.expectEmit(true, false, false, true);
        emit BatchInfoAuthenticated(commitment, fallbackBatcher);

        vm.prank(fallbackBatcher);
        authenticator.authenticateBatchInfo(commitment, "");
    }

    /// @notice Test that authenticateBatchInfo reverts in fallback mode when called by
    ///         a sender that is not the SystemConfig batcher address.
    function test_authenticateBatchInfo_fallback_revertsOnWrongSender() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // Switch to fallback mode.
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(false);
        assertFalse(authenticator.activeIsEspresso());

        address fallbackBatcher = address(0xCAFE);
        mockSystemConfig.setBatcherHash(bytes32(uint256(uint160(fallbackBatcher))));

        bytes32 commitment = keccak256("fallback commitment");

        // An unauthorized sender must be rejected.
        vm.prank(unauthorized);
        vm.expectRevert(
            abi.encodeWithSelector(
                IBatchAuthenticator.UnauthorizedFallbackBatcher.selector, unauthorized, fallbackBatcher
            )
        );
        authenticator.authenticateBatchInfo(commitment, "");
    }

    /// @notice Test that in Espresso (default) mode, any sender (including the fallback batcher)
    ///         other than espressoBatcher is rejected before signature verification.
    function test_authenticateBatchInfo_espresso_revertsOnUnauthorizedSender() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();
        // Sanity: still in Espresso mode.
        assertTrue(authenticator.activeIsEspresso());

        // Configure a fallback batcher; Espresso mode must NOT use it.
        address fallbackBatcher = address(0xCAFE);
        mockSystemConfig.setBatcherHash(bytes32(uint256(uint160(fallbackBatcher))));

        bytes32 commitment = keccak256("espresso commitment");

        // Any non-espressoBatcher sender must revert with UnauthorizedEspressoBatcher.
        vm.prank(fallbackBatcher);
        vm.expectRevert(
            abi.encodeWithSelector(IBatchAuthenticator.UnauthorizedEspressoBatcher.selector, fallbackBatcher, espressoBatcher)
        );
        authenticator.authenticateBatchInfo(commitment, "");
    }

    /// @notice Test that authenticateBatchInfo ignores the SystemConfig paused flag.
    ///         The pause domain of the optimism stack must not gate batch authentication.
    function test_authenticateBatchInfo_ignoresPause_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        uint256 privateKey = 1;
        bytes32 commitment = keccak256("test commitment");

        // Register signer and create valid signature.
        _registerNitroSigner(privateKey);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, _computeEIP712Digest(commitment));
        bytes memory signature = abi.encodePacked(r, s, v);

        // Pause the SystemConfig — authentication must still succeed.
        mockSystemConfig.setPaused(true);

        vm.expectEmit(true, false, false, true);
        emit BatchInfoAuthenticated(commitment, espressoBatcher);
        vm.prank(espressoBatcher);
        authenticator.authenticateBatchInfo(commitment, signature);
    }

    /// @notice Test that registerSigner ignores the SystemConfig paused flag.
    function test_registerSigner_ignoresPause_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        uint256 privateKey = 1;
        bytes memory signerData = _nitroRegistrationOutputForPrivateKey(privateKey);
        bytes memory proofBytes = "";

        // Pause the SystemConfig — registration must still succeed.
        mockSystemConfig.setPaused(true);

        vm.expectEmit(true, false, false, false);
        emit SignerRegistrationInitiated(address(this));
        authenticator.registerSigner(signerData, proofBytes);
    }

    /// @notice End-to-end coverage of the dual-batcher flow: authenticate via Espresso, switch
    ///         to fallback, authenticate via the SystemConfig batcher, switch back, authenticate
    ///         via Espresso again. Verifies that switching doesn't corrupt either path and that
    ///         each mode rejects the other mode's caller.
    function test_switchAndAuthenticate_endToEnd_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // 1. Espresso path: register signer and authenticate one commitment.
        uint256 privateKey = 1;
        _registerNitroSigner(privateKey);

        bytes32 espressoCommitment1 = keccak256("espresso-1");
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, _computeEIP712Digest(espressoCommitment1));
        bytes memory espressoSig1 = abi.encodePacked(r, s, v);

        vm.expectEmit(true, false, false, true);
        emit BatchInfoAuthenticated(espressoCommitment1, espressoBatcher);
        vm.prank(espressoBatcher);
        authenticator.authenticateBatchInfo(espressoCommitment1, espressoSig1);

        // 2. Switch to fallback and configure the SystemConfig batcher.
        address fallbackBatcher = address(0xCAFE);
        mockSystemConfig.setBatcherHash(bytes32(uint256(uint160(fallbackBatcher))));

        vm.expectEmit(true, false, false, false);
        emit BatcherSwitched(false);
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(false);
        assertFalse(authenticator.activeIsEspresso());

        // 3. Fallback path: only the configured batcher may authenticate; signature is ignored.
        bytes32 fallbackCommitment = keccak256("fallback");
        vm.expectEmit(true, false, false, true);
        emit BatchInfoAuthenticated(fallbackCommitment, fallbackBatcher);
        vm.prank(fallbackBatcher);
        authenticator.authenticateBatchInfo(fallbackCommitment, "");

        // Re-issue the exact same call that succeeded in step 1 — same sender, same commitment,
        // same signature — and assert it now reverts. Demonstrates that the mode switch alone
        // is sufficient to change the outcome; the previously-valid Espresso signature is no
        // longer consulted at all.
        vm.expectRevert(
            abi.encodeWithSelector(
                IBatchAuthenticator.UnauthorizedFallbackBatcher.selector, address(this), fallbackBatcher
            )
        );
        authenticator.authenticateBatchInfo(espressoCommitment1, espressoSig1);

        // 4. Switch back to Espresso.
        vm.expectEmit(true, false, false, false);
        emit BatcherSwitched(true);
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(true);
        assertTrue(authenticator.activeIsEspresso());

        // 5. Espresso path again with a new commitment — registration must have survived
        //    the switch round-trip.
        bytes32 espressoCommitment2 = keccak256("espresso-2");
        (v, r, s) = vm.sign(privateKey, _computeEIP712Digest(espressoCommitment2));
        bytes memory espressoSig2 = abi.encodePacked(r, s, v);

        vm.expectEmit(true, false, false, true);
        emit BatchInfoAuthenticated(espressoCommitment2, espressoBatcher);
        vm.prank(espressoBatcher);
        authenticator.authenticateBatchInfo(espressoCommitment2, espressoSig2);
    }

    // Event declarations for expectEmit.
    event BatchInfoAuthenticated(bytes32 commitment, address indexed caller);
    event SignerRegistrationInitiated(address indexed caller);
    event EspressoBatcherUpdated(
        address indexed oldEspressoBatcher, address indexed newEspressoBatcher, uint64 indexed fromBlock
    );
    event BatcherSwitched(bool indexed activeIsEspresso);

    /// @notice Deploy a Proxy without importing Proxy.sol to avoid duplicate compilation artifacts
    ///         that break vm.getCode("Proxy") disambiguation in tests.
    function _newProxy(address _admin) internal returns (IProxy) {
        bytes memory initCode = abi.encodePacked(vm.getCode("src/universal/Proxy.sol:Proxy"), abi.encode(_admin));
        address payable proxyAddr;
        assembly {
            proxyAddr := create(0, add(initCode, 0x20), mload(initCode))
        }
        require(proxyAddr != address(0), "BatchAuthenticator_Uncategorized_Test: proxy deployment failed");
        return IProxy(proxyAddr);
    }
}

/// @notice Fork tests for BatchAuthenticator. Runs against the FORK_RPC_URL fork when FORK_TEST=true,
///         using the repo's standard fork-test env vars (FORK_TEST, FORK_RPC_URL, FORK_BLOCK_NUMBER)
///         exposed via the Config library.
contract BatchAuthenticator_Fork_Test is Test {
    address public proxyAdminOwner = address(0xBEEF);
    address public espressoBatcher = address(0x1234);

    MockSystemConfig public mockSystemConfig;
    EspressoTEEVerifierMock public teeVerifier;
    EspressoNitroTEEVerifierMock public nitroVerifier;
    BatchAuthenticator public implementation;
    IProxy public proxy;
    IProxyAdmin public proxyAdmin;
    BatchAuthenticator public authenticator;

    bytes32 private constant _ESPRESSO_TEE_VERIFIER_TYPE_HASH = keccak256("EspressoTEEVerifier(bytes32 commitment)");

    bytes32 private constant _EIP712_DOMAIN_TYPE_HASH =
        keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)");

    /// @notice Compute the EIP-712 digest that the TEE verifier mock expects.
    function _computeEIP712Digest(bytes32 commitment) internal view returns (bytes32) {
        bytes32 structHash = keccak256(abi.encode(_ESPRESSO_TEE_VERIFIER_TYPE_HASH, commitment));
        bytes32 domainSeparator = keccak256(
            abi.encode(
                _EIP712_DOMAIN_TYPE_HASH,
                keccak256("EspressoTEEVerifier"),
                keccak256("1"),
                block.chainid,
                address(teeVerifier)
            )
        );
        return keccak256(abi.encodePacked("\x19\x01", domainSeparator, structHash));
    }

    function setUp() public {
        // Skip unless fork tests are explicitly enabled.
        if (!Config.l1ForkTest()) {
            vm.skip(true);
            return;
        }

        vm.createSelectFork(Config.forkRpcUrl(), Config.forkBlockNumber());

        console.log("BatchAuthenticator_Fork_Test: forked at block", block.number);

        mockSystemConfig = new MockSystemConfig();
        nitroVerifier = new EspressoNitroTEEVerifierMock();
        teeVerifier = new EspressoTEEVerifierMock(IEspressoNitroTEEVerifier(address(nitroVerifier)));
        implementation = new BatchAuthenticator();

        // Deploy ProxyAdmin via vm.getCode to avoid duplicate ProxyAdmin artifacts.
        {
            bytes memory _code = vm.getCode("forge-artifacts/ProxyAdmin.sol/ProxyAdmin.json");
            bytes memory _args = abi.encode(proxyAdminOwner);
            bytes memory _initCode = abi.encodePacked(_code, _args);
            address _addr;
            assembly {
                _addr := create(0, add(_initCode, 0x20), mload(_initCode))
            }
            proxyAdmin = IProxyAdmin(_addr);
        }
        proxy = _newProxy(address(proxyAdmin));
        vm.prank(proxyAdminOwner);
        proxyAdmin.setProxyType(address(proxy), IProxyAdmin.ProxyType.ERC1967);

        bytes memory initData = abi.encodeCall(
            BatchAuthenticator.initialize,
            (
                IEspressoTEEVerifier(address(teeVerifier)),
                espressoBatcher,
                ISystemConfig(address(mockSystemConfig)),
                proxyAdminOwner,
                true
            )
        );
        vm.prank(proxyAdminOwner);
        proxyAdmin.upgradeAndCall(payable(address(proxy)), address(implementation), initData);

        authenticator = BatchAuthenticator(address(proxy));
    }

    function _nitroRegistrationOutputForPrivateKey(uint256 privateKey) internal returns (bytes memory) {
        Vm.Wallet memory wallet = vm.createWallet(privateKey);
        // uncompressed secp256k1 public key similar to the key TEE generates
        bytes memory publicKey = abi.encodePacked(
            // uncompressed key prefix
            bytes1(0x04),
            bytes32(wallet.publicKeyX),
            bytes32(wallet.publicKeyY)
        );

        VerifierJournal memory journal = VerifierJournal({
            result: VerificationResult.Success,
            trustedCertsPrefixLen: 0,
            timestamp: 0,
            certs: new bytes32[](0),
            userData: new bytes(0),
            nonce: new bytes(0),
            publicKey: publicKey,
            pcrs: new Pcr[](0),
            moduleId: ""
        });

        return abi.encode(journal);
    }

    function _registerNitroSigner(uint256 privateKey) internal {
        nitroVerifier.registerService(_nitroRegistrationOutputForPrivateKey(privateKey), "");
    }

    /// @notice Test deployment and initialization on the fork.
    function test_deployment_succeeds() external view {
        assertEq(address(authenticator.espressoTEEVerifier()), address(teeVerifier));
        assertEq(authenticator.espressoBatcher(), espressoBatcher);
        assertTrue(authenticator.activeIsEspresso());
        assertEq(authenticator.version(), "1.2.0");

        // Verify proxy admin.
        address admin = EIP1967Helper.getAdmin(address(proxy));
        assertEq(admin, address(proxyAdmin));
    }

    /// @notice Test setActiveIsEspresso on the fork.
    function test_setActiveIsEspresso_succeeds() external {
        assertTrue(authenticator.activeIsEspresso());

        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(false);

        assertFalse(authenticator.activeIsEspresso());

        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(true);

        assertTrue(authenticator.activeIsEspresso());
    }

    /// @notice Test authenticateBatchInfo on the fork.
    function test_authenticateBatchInfo_succeeds() external {
        bytes32 commitment = keccak256("test commitment on fork");

        // Create a signature.
        uint256 privateKey = 1;

        // Register the signer.
        _registerNitroSigner(privateKey);

        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, _computeEIP712Digest(commitment));
        bytes memory signature = abi.encodePacked(r, s, v);

        // Authenticate.
        vm.expectEmit(true, false, false, true);
        emit BatchInfoAuthenticated(commitment, espressoBatcher);
        vm.prank(espressoBatcher);
        authenticator.authenticateBatchInfo(commitment, signature);
    }

    /// @notice Test upgrade on the fork preserves state.
    function test_upgrade_succeeds() external {
        // Initialize the authenticator.
        bytes32 commitment = keccak256("test commitment");
        uint256 privateKey = 1;

        // Register the signer.
        _registerNitroSigner(privateKey);

        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, _computeEIP712Digest(commitment));
        bytes memory signature = abi.encodePacked(r, s, v);
        vm.prank(espressoBatcher);
        authenticator.authenticateBatchInfo(commitment, signature);

        // Switch batcher
        vm.prank(proxyAdminOwner);
        authenticator.setActiveIsEspresso(false);
        assertFalse(authenticator.activeIsEspresso());

        // Deploy new implementation and upgrade.
        BatchAuthenticator newImpl = new BatchAuthenticator();
        vm.prank(proxyAdminOwner);
        proxyAdmin.upgrade(payable(address(proxy)), address(newImpl));

        // Verify state is preserved.
        assertFalse(authenticator.activeIsEspresso());
        assertEq(address(authenticator.espressoTEEVerifier()), address(teeVerifier));
        assertEq(authenticator.espressoBatcher(), espressoBatcher);
    }

    /// @notice Test that the contract works against live forked L1 state.
    function test_integrationWithFork_succeeds() external view {
        assertEq(authenticator.version(), "1.2.0");
        assertTrue(authenticator.activeIsEspresso());

        uint256 blockNum = block.number;
        assertGt(blockNum, 0);
        console.log("Fork block number:", blockNum);
    }

    // Event declarations for expectEmit.
    event BatchInfoAuthenticated(bytes32 commitment, address indexed caller);
    event SignerRegistrationInitiated(address indexed caller);
    event EspressoBatcherUpdated(
        address indexed oldEspressoBatcher, address indexed newEspressoBatcher, uint64 indexed fromBlock
    );
    event BatcherSwitched(bool indexed activeIsEspresso);

    /// @notice Deploy a Proxy without importing Proxy.sol to avoid duplicate compilation artifacts.
    function _newProxy(address _admin) internal returns (IProxy) {
        bytes memory initCode = abi.encodePacked(vm.getCode("src/universal/Proxy.sol:Proxy"), abi.encode(_admin));
        address payable proxyAddr;
        assembly {
            proxyAddr := create(0, add(initCode, 0x20), mload(initCode))
        }
        require(proxyAddr != address(0), "BatchAuthenticator_Fork_Test: proxy deployment failed");
        return IProxy(proxyAddr);
    }
}
