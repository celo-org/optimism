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
            bytes memory _code = vm.getCode("ProxyAdmin");
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
                proxyAdminOwner
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
                proxyAdminOwner
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
                proxyAdminOwner
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
                proxyAdminOwner
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

    /// @notice Test that switchBatcher can be called by owner or guardian.
    function test_switchBatcher_ownerOrGuardian_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // ProxyAdmin owner (now contract owner) can switch.
        vm.expectEmit(true, false, false, false);
        emit BatcherSwitched(false);
        vm.prank(proxyAdminOwner);
        authenticator.switchBatcher();
        assertFalse(authenticator.activeIsEspresso());

        // Switch back.
        vm.expectEmit(true, false, false, false);
        emit BatcherSwitched(true);
        vm.prank(proxyAdminOwner);
        authenticator.switchBatcher();
        assertTrue(authenticator.activeIsEspresso());

        // Add a guardian.
        vm.prank(proxyAdminOwner);
        authenticator.addGuardian(guardian);
        assertTrue(authenticator.isGuardian(guardian));

        // Guardian can switch.
        vm.expectEmit(true, false, false, false);
        emit BatcherSwitched(false);
        vm.prank(guardian);
        authenticator.switchBatcher();
        assertFalse(authenticator.activeIsEspresso());

        // Guardian can switch back.
        vm.expectEmit(true, false, false, false);
        emit BatcherSwitched(true);
        vm.prank(guardian);
        authenticator.switchBatcher();
        assertTrue(authenticator.activeIsEspresso());

        // Unauthorized cannot switch.
        vm.prank(unauthorized);
        vm.expectRevert(
            abi.encodeWithSelector(OwnableWithGuardiansUpgradeable.NotGuardianOrOwner.selector, unauthorized)
        );
        authenticator.switchBatcher();

        // ProxyAdmin cannot switch.
        vm.prank(address(proxyAdmin));
        vm.expectRevert(
            abi.encodeWithSelector(OwnableWithGuardiansUpgradeable.NotGuardianOrOwner.selector, address(proxyAdmin))
        );
        authenticator.switchBatcher();
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
        vm.expectEmit(true, false, false, false);
        emit BatchInfoAuthenticated(commitment);

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

        // ProxyAdmin owner can set.
        vm.expectEmit(true, true, false, false);
        emit EspressoBatcherUpdated(espressoBatcher, newEspressoBatcher);
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

    /// @notice Test that setEspressoBatcher reverts when zero address is provided.
    function test_setEspressoBatcher_whenZeroAddress_reverts() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        vm.prank(proxyAdminOwner);
        vm.expectRevert(abi.encodeWithSelector(IBatchAuthenticator.InvalidAddress.selector, address(0)));
        authenticator.setEspressoBatcher(address(0));
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
        authenticator.authenticateBatchInfo(commitment, signature);

        // Switch batcher to test boolean flag preservation.
        vm.prank(proxyAdminOwner);
        authenticator.switchBatcher();
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
        authenticator.switchBatcher();
        assertFalse(authenticator.activeIsEspresso());

        // Configure the SystemConfig batcher to a known address.
        address fallbackBatcher = address(0xCAFE);
        mockSystemConfig.setBatcherHash(bytes32(uint256(uint160(fallbackBatcher))));

        bytes32 commitment = keccak256("fallback commitment");

        // The fallback batcher path ignores the signature; pass empty bytes.
        vm.expectEmit(true, false, false, false);
        emit BatchInfoAuthenticated(commitment);

        vm.prank(fallbackBatcher);
        authenticator.authenticateBatchInfo(commitment, "");
    }

    /// @notice Test that authenticateBatchInfo reverts in fallback mode when called by
    ///         a sender that is not the SystemConfig batcher address.
    function test_authenticateBatchInfo_fallback_revertsOnWrongSender() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // Switch to fallback mode.
        vm.prank(proxyAdminOwner);
        authenticator.switchBatcher();
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

    /// @notice Test that in Espresso (default) mode, the TEE path is taken — calling with
    ///         the fallback-batcher address but no valid TEE signature must revert.
    function test_authenticateBatchInfo_espresso_revertsOnFallbackSender() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();
        // Sanity: still in Espresso mode.
        assertTrue(authenticator.activeIsEspresso());

        // Configure a fallback batcher; Espresso mode must NOT use it.
        address fallbackBatcher = address(0xCAFE);
        mockSystemConfig.setBatcherHash(bytes32(uint256(uint160(fallbackBatcher))));

        bytes32 commitment = keccak256("espresso commitment");

        // Calling with empty signature — TEE path runs ECDSA.recover, which rejects the
        // zero-length input as ECDSAInvalidSignatureLength(0).
        vm.prank(fallbackBatcher);
        vm.expectRevert(abi.encodeWithSelector(ECDSA.ECDSAInvalidSignatureLength.selector, uint256(0)));
        authenticator.authenticateBatchInfo(commitment, "");
    }

    /// @notice Test that paused() delegates to SystemConfig.
    function test_paused_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // Initially not paused.
        assertFalse(authenticator.paused());

        // Pause the mock SystemConfig.
        mockSystemConfig.setPaused(true);
        assertTrue(authenticator.paused());

        // Unpause.
        mockSystemConfig.setPaused(false);
        assertFalse(authenticator.paused());
    }

    /// @notice Test that authenticateBatchInfo reverts when paused.
    function test_authenticateBatchInfo_whenPaused_reverts() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        uint256 privateKey = 1;
        bytes32 commitment = keccak256("test commitment");

        // Register signer and create valid signature.
        _registerNitroSigner(privateKey);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, _computeEIP712Digest(commitment));
        bytes memory signature = abi.encodePacked(r, s, v);

        // Pause the system.
        mockSystemConfig.setPaused(true);

        // Should revert with BatchAuthenticator_Paused.
        vm.expectRevert(abi.encodeWithSelector(IBatchAuthenticator.BatchAuthenticator_Paused.selector));
        authenticator.authenticateBatchInfo(commitment, signature);
    }

    /// @notice Test that authenticateBatchInfo succeeds when not paused.
    function test_authenticateBatchInfo_whenNotPaused_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        uint256 privateKey = 1;
        bytes32 commitment = keccak256("test commitment");

        // Register signer and create valid signature.
        _registerNitroSigner(privateKey);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, _computeEIP712Digest(commitment));
        bytes memory signature = abi.encodePacked(r, s, v);

        // Ensure not paused.
        mockSystemConfig.setPaused(false);

        // Should succeed.
        vm.expectEmit(true, false, false, false);
        emit BatchInfoAuthenticated(commitment);
        authenticator.authenticateBatchInfo(commitment, signature);
    }

    /// @notice Test that registerSigner reverts when paused.
    function test_registerSigner_whenPaused_reverts() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        uint256 privateKey = 1;
        bytes memory signerData = _nitroRegistrationOutputForPrivateKey(privateKey);
        bytes memory proofBytes = "";

        // Pause the system.
        mockSystemConfig.setPaused(true);

        // Should revert with BatchAuthenticator_Paused.
        vm.expectRevert(abi.encodeWithSelector(IBatchAuthenticator.BatchAuthenticator_Paused.selector));
        authenticator.registerSigner(signerData, proofBytes);
    }

    /// @notice Test that switchBatcher still works when paused (emergency recovery).
    function test_switchBatcher_whenPaused_succeeds() external {
        BatchAuthenticator authenticator = _deployAndInitializeProxy();

        // Pause the system.
        mockSystemConfig.setPaused(true);

        // Owner can still switch batcher while paused.
        vm.prank(proxyAdminOwner);
        authenticator.switchBatcher();
        assertFalse(authenticator.activeIsEspresso());
    }

    // Event declarations for expectEmit.
    event BatchInfoAuthenticated(bytes32 indexed commitment);
    event SignerRegistrationInitiated(address indexed caller);
    event EspressoBatcherUpdated(address indexed oldEspressoBatcher, address indexed newEspressoBatcher);
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
            bytes memory _code = vm.getCode("ProxyAdmin");
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
                proxyAdminOwner
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

    /// @notice Test switchBatcher on the fork.
    function test_switchBatcher_succeeds() external {
        assertTrue(authenticator.activeIsEspresso());

        vm.prank(proxyAdminOwner);
        authenticator.switchBatcher();

        assertFalse(authenticator.activeIsEspresso());

        vm.prank(proxyAdminOwner);
        authenticator.switchBatcher();

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
        vm.expectEmit(true, false, false, false);
        emit BatchInfoAuthenticated(commitment);
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
        authenticator.authenticateBatchInfo(commitment, signature);

        // Switch batcher
        vm.prank(proxyAdminOwner);
        authenticator.switchBatcher();
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
    event BatchInfoAuthenticated(bytes32 indexed commitment);
    event SignerRegistrationInitiated(address indexed caller);
    event EspressoBatcherUpdated(address indexed oldEspressoBatcher, address indexed newEspressoBatcher);
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
