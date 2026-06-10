// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { BaseDeployIO } from "scripts/deploy/BaseDeployIO.sol";
import { Script } from "forge-std/Script.sol";
import { Solarray } from "scripts/libraries/Solarray.sol";
import { IBatchAuthenticator } from "interfaces/L1/IBatchAuthenticator.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IEspressoNitroTEEVerifier } from "@espresso-tee-contracts/interface/IEspressoNitroTEEVerifier.sol";
import { IEspressoTEEVerifier } from "@espresso-tee-contracts/interface/IEspressoTEEVerifier.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { BatchAuthenticator } from "src/L1/BatchAuthenticator.sol";
import { MockEspressoTEEVerifier } from "test/mocks/MockEspressoTEEVerifiers.sol";
import { MockEspressoNitroTEEVerifier } from "test/mocks/MockEspressoTEEVerifiers.sol";

contract DeployEspressoInput is BaseDeployIO {
    address internal _nitroEnclaveVerifier;
    address internal _espressoBatcher;
    address internal _systemConfig;
    address internal _espressoOwner;
    address internal _sharedProxyAdmin;

    function set(bytes4 _sel, address _val) public {
        if (_sel == this.nitroEnclaveVerifier.selector) {
            _nitroEnclaveVerifier = _val;
        } else if (_sel == this.espressoBatcher.selector) {
            _espressoBatcher = _val;
        } else if (_sel == this.systemConfig.selector) {
            _systemConfig = _val;
        } else if (_sel == this.espressoOwner.selector) {
            _espressoOwner = _val;
        } else if (_sel == this.sharedProxyAdmin.selector) {
            _sharedProxyAdmin = _val;
        } else {
            revert("DeployEspressoInput: unknown selector");
        }
    }

    /// @notice Address of the underlying AWS NitroEnclaveVerifier (from Automata).
    ///         Set to address(0) to deploy mock verifiers (dev/test only).
    function nitroEnclaveVerifier() public view returns (address) {
        return _nitroEnclaveVerifier;
    }

    function espressoBatcher() public view returns (address) {
        return _espressoBatcher;
    }

    function systemConfig() public view returns (address) {
        return _systemConfig;
    }

    /// @notice The application-level (OZ Ownable / OwnableWithGuardians) owner for the Espresso
    ///         contracts — gates operational setters (setEspressoBatcher, setActiveIsEspresso,
    ///         setEnclaveHash, etc.). This is NOT the shared ProxyAdmin owner, which controls upgrades
    ///         and `initialize`. Defaults to the deployer if not set.
    function espressoOwner() public view returns (address) {
        return _espressoOwner;
    }

    /// @notice Address of the existing (shared) OP Stack ProxyAdmin that the BatchAuthenticator and
    ///         TEEVerifier proxies are handed over to. Required.
    function sharedProxyAdmin() public view returns (address) {
        require(_sharedProxyAdmin != address(0), "DeployEspressoInput: sharedProxyAdmin not set");
        return _sharedProxyAdmin;
    }
}

contract DeployEspressoOutput is BaseDeployIO {
    address internal _batchAuthenticatorAddress;
    address internal _teeVerifierProxy;
    address internal _nitroTEEVerifier;

    function set(bytes4 _sel, address _addr) public {
        require(_addr != address(0), "DeployEspressoOutput: cannot set zero address");
        if (_sel == this.batchAuthenticatorAddress.selector) {
            _batchAuthenticatorAddress = _addr;
        } else if (_sel == this.teeVerifierProxy.selector) {
            _teeVerifierProxy = _addr;
        } else if (_sel == this.nitroTEEVerifier.selector) {
            _nitroTEEVerifier = _addr;
        } else {
            revert("DeployEspressoOutput: unknown selector");
        }
    }

    function batchAuthenticatorAddress() public view returns (address) {
        require(_batchAuthenticatorAddress != address(0), "DeployEspressoOutput: batch authenticator address not set");
        return _batchAuthenticatorAddress;
    }

    function teeVerifierProxy() public view returns (address) {
        require(_teeVerifierProxy != address(0), "DeployEspressoOutput: tee verifier proxy not set");
        return _teeVerifierProxy;
    }

    function nitroTEEVerifier() public view returns (address) {
        require(_nitroTEEVerifier != address(0), "DeployEspressoOutput: nitro tee verifier proxy not set");
        return _nitroTEEVerifier;
    }

    /// @notice Alias for teeVerifierProxy for convenience
    function teeVerifierAddress() public view returns (address) {
        return teeVerifierProxy();
    }
}

contract DeployEspresso is Script {
    function run(DeployEspressoInput _input, DeployEspressoOutput _output, address _deployerAddress) public {
        IEspressoTEEVerifier teeVerifier = deployTEEContracts(_input, _output, _deployerAddress);
        deployBatchAuthenticator(_input, _output, _deployerAddress, teeVerifier);
        checkOutput(_output);
    }

    function deployBatchAuthenticator(
        DeployEspressoInput _input,
        DeployEspressoOutput _output,
        address _deployerAddress,
        IEspressoTEEVerifier _teeVerifier
    )
        public
        returns (IBatchAuthenticator)
    {
        // The BatchAuthenticator app-level owner (OwnableWithGuardians). Distinct from the shared
        // ProxyAdmin owner.
        address batchAuthenticatorOwner = _input.espressoOwner();
        if (batchAuthenticatorOwner == address(0)) batchAuthenticatorOwner = _deployerAddress;

        // Deploy the proxy with the deployer as its transient admin so the deployer can initialize it
        // directly, then `changeAdmin` hands the proxy over to the shared ProxyAdmin (same pattern as
        // DeployAltDA / DeployFeesDepositor).
        address sharedProxyAdmin = _input.sharedProxyAdmin();

        // Deploy Proxy without importing Proxy.sol to avoid duplicate compilation artifacts.
        IProxy proxy;
        {
            bytes memory initCode =
                abi.encodePacked(vm.getCode("src/universal/Proxy.sol:Proxy"), abi.encode(msg.sender));
            address payable proxyAddr;
            vm.broadcast(msg.sender);
            assembly {
                proxyAddr := create(0, add(initCode, 0x20), mload(initCode))
            }
            require(proxyAddr != address(0), "DeployEspresso: proxy deployment failed");
            proxy = IProxy(proxyAddr);
        }
        vm.label(address(proxy), "BatchAuthenticatorProxy");
        vm.broadcast(msg.sender);
        BatchAuthenticator impl = new BatchAuthenticator();
        vm.label(address(impl), "BatchAuthenticatorImpl");

        bytes memory initData = abi.encodeCall(
            BatchAuthenticator.initialize,
            (
                _teeVerifier,
                _input.espressoBatcher(),
                ISystemConfig(_input.systemConfig()),
                batchAuthenticatorOwner,
                // First deployment: start with the Espresso batcher active.
                true
            )
        );
        // Initialize directly via the proxy. The deployer is still the proxy admin at this point, so
        // BatchAuthenticator.initialize's `_assertOnlyProxyAdminOrProxyAdminOwner` check passes.
        vm.broadcast(msg.sender);
        proxy.upgradeToAndCall(address(impl), initData);

        // Hand the proxy over to the shared OP Stack ProxyAdmin. No setProxyType call is needed: the
        // ProxyAdmin treats unregistered proxies as ProxyType.ERC1967 (enum value 0), which matches
        // src/universal/Proxy.sol.
        vm.broadcast(msg.sender);
        proxy.changeAdmin(sharedProxyAdmin);

        _output.set(_output.batchAuthenticatorAddress.selector, address(proxy));
        return IBatchAuthenticator(address(proxy));
    }

    /// @notice Deploys NitroTEEVerifier and TEEVerifier (production path).
    ///         Deployment order:
    ///         1. Deploy TEEVerifier (impl + OP-style ERC-1967 Proxy + ProxyAdmin) with placeholder nitro address
    ///         2. Deploy NitroTEEVerifier pointing to the TEEVerifier proxy
    ///         3. Update TEEVerifier with the actual NitroTEEVerifier address
    ///
    ///         The TEEVerifier is deployed behind src/universal/Proxy.sol rather than the
    ///         upstream's OZ v5 TransparentUpgradeableProxy. This avoids pulling OZ's TUP +
    ///         ProxyAdmin into the OP artifact tree (which would shadow src/universal/ProxyAdmin.sol).
    ///
    ///         If nitroEnclaveVerifier is address(0), deploys our local mocks (dev/test only).
    function deployTEEContracts(
        DeployEspressoInput _input,
        DeployEspressoOutput _output,
        address _deployerAddress
    )
        public
        returns (IEspressoTEEVerifier)
    {
        address nitroEnclaveVerifier = _input.nitroEnclaveVerifier();
        if (nitroEnclaveVerifier == address(0)) {
            return _deployMockTEEContracts(_output);
        }
        return _deployProductionTEEContracts(_input, _output, _deployerAddress, nitroEnclaveVerifier);
    }

    function _deployMockTEEContracts(DeployEspressoOutput _output) internal returns (IEspressoTEEVerifier) {
        // Use our local mocks — they carry OP-specific test behavior (permissive isSignerValid,
        // test helper overrides, special address exceptions) that the submodule mocks don't have.
        // The mocks are unproxied, so there is no ProxyAdmin to wire here.
        vm.broadcast(msg.sender);
        MockEspressoNitroTEEVerifier nitroMock = new MockEspressoNitroTEEVerifier();
        vm.label(address(nitroMock), "MockEspressoNitroTEEVerifier");

        vm.broadcast(msg.sender);
        MockEspressoTEEVerifier teeMock = new MockEspressoTEEVerifier(IEspressoNitroTEEVerifier(address(nitroMock)));
        vm.label(address(teeMock), "MockEspressoTEEVerifier");

        _output.set(_output.nitroTEEVerifier.selector, address(nitroMock));
        _output.set(_output.teeVerifierProxy.selector, address(teeMock));
        return IEspressoTEEVerifier(address(teeMock));
    }

    function _deployProductionTEEContracts(
        DeployEspressoInput _input,
        DeployEspressoOutput _output,
        address _deployerAddress,
        address _nitroEnclaveVerifier
    )
        internal
        returns (IEspressoTEEVerifier)
    {
        address teeVerifierOwner = _input.espressoOwner();
        if (teeVerifierOwner == address(0)) teeVerifierOwner = _deployerAddress;

        address sharedProxyAdmin = _input.sharedProxyAdmin();

        // Deploy OP's ERC-1967 Proxy with the deployer as its transient admin so the deployer can
        // initialize it directly below, then `changeAdmin` hands it to the shared ProxyAdmin (same
        // pattern as deployBatchAuthenticator / DeployAltDA / DeployFeesDepositor).
        address payable teeProxyAddr;
        {
            bytes memory initCode =
                abi.encodePacked(vm.getCode("src/universal/Proxy.sol:Proxy"), abi.encode(msg.sender));
            vm.broadcast(msg.sender);
            assembly {
                teeProxyAddr := create(0, add(initCode, 0x20), mload(initCode))
            }
            require(teeProxyAddr != address(0), "DeployEspresso: tee proxy deployment failed");
        }
        vm.label(teeProxyAddr, "TEEVerifierProxy");

        // Deploy the implementation.
        // Use vm.getCode against the submodule's own out/ to avoid pulling the impl closure
        // (TEEHelper, JournalValidation, aws-nitro-enclave-attestation) into OP's compile group.
        address payable teeImplAddr;
        {
            bytes memory teeImplCode =
                vm.getCode("lib/espresso-tee-contracts/out/EspressoTEEVerifier.sol/EspressoTEEVerifier.json");
            vm.broadcast(msg.sender);
            assembly {
                teeImplAddr := create(0, add(teeImplCode, 0x20), mload(teeImplCode))
            }
            require(teeImplAddr != address(0), "DeployEspresso: EspressoTEEVerifier impl deployment failed");
        }
        IEspressoTEEVerifier teeImpl = IEspressoTEEVerifier(teeImplAddr);
        vm.label(teeImplAddr, "TEEVerifierImpl");

        // Deploy NitroTEEVerifier first (no proxy; its constructor only stores the TEE proxy address
        // for access control). Deploying it before init lets us wire it directly via `initialize`,
        // avoiding a separate onlyOwner call and the Ownable2Step ownership-transfer dance.
        // Use vm.getCode against the submodule's own out/ to avoid pulling the impl closure
        // into OP's compile group.
        address nitroVerifierAddr;
        {
            bytes memory nitroImplCode = abi.encodePacked(
                vm.getCode("lib/espresso-tee-contracts/out/EspressoNitroTEEVerifier.sol/EspressoNitroTEEVerifier.json"),
                abi.encode(teeProxyAddr, _nitroEnclaveVerifier)
            );
            vm.broadcast(msg.sender);
            assembly {
                nitroVerifierAddr := create(0, add(nitroImplCode, 0x20), mload(nitroImplCode))
            }
            require(nitroVerifierAddr != address(0), "DeployEspresso: EspressoNitroTEEVerifier deployment failed");
        }
        IEspressoNitroTEEVerifier nitroVerifier = IEspressoNitroTEEVerifier(nitroVerifierAddr);
        vm.label(nitroVerifierAddr, "NitroTEEVerifier");

        // initialize(address _owner, address _espressoNitroTEEVerifier). Sets the final contract owner
        // and wires the Nitro verifier in one shot, so no post-init onlyOwner call is needed. The
        // deployer is still the proxy admin at this point, so it can call upgradeToAndCall directly.
        bytes memory initData =
            abi.encodeWithSignature("initialize(address,address)", teeVerifierOwner, nitroVerifierAddr);
        vm.broadcast(msg.sender);
        IProxy(teeProxyAddr).upgradeToAndCall(address(teeImpl), initData);

        // Hand the proxy over to the shared OP Stack ProxyAdmin. No setProxyType call is needed: the
        // ProxyAdmin treats unregistered proxies as ProxyType.ERC1967 (enum value 0), which matches
        // src/universal/Proxy.sol.
        vm.broadcast(msg.sender);
        IProxy(teeProxyAddr).changeAdmin(sharedProxyAdmin);

        _output.set(_output.teeVerifierProxy.selector, teeProxyAddr);
        _output.set(_output.nitroTEEVerifier.selector, address(nitroVerifier));

        return IEspressoTEEVerifier(teeProxyAddr);
    }

    function checkOutput(DeployEspressoOutput _output) public view {
        address[] memory addresses = Solarray.addresses(
            _output.batchAuthenticatorAddress(), _output.teeVerifierProxy(), _output.nitroTEEVerifier()
        );
        for (uint256 i = 0; i < addresses.length; i++) {
            require(
                addresses[i] != address(0) && addresses[i].code.length > 0, "DeployEspresso: invalid contract address"
            );
        }
    }
}
