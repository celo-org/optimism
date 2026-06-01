// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { BaseDeployIO } from "scripts/deploy/BaseDeployIO.sol";
import { Script } from "forge-std/Script.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";
import { Solarray } from "scripts/libraries/Solarray.sol";
import { IBatchAuthenticator } from "interfaces/L1/IBatchAuthenticator.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IEspressoNitroTEEVerifier } from "@espresso-tee-contracts/interface/IEspressoNitroTEEVerifier.sol";
import { IEspressoTEEVerifier } from "@espresso-tee-contracts/interface/IEspressoTEEVerifier.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { BatchAuthenticator } from "src/L1/BatchAuthenticator.sol";
import { MockEspressoTEEVerifier } from "test/mocks/MockEspressoTEEVerifiers.sol";
import { MockEspressoNitroTEEVerifier } from "test/mocks/MockEspressoTEEVerifiers.sol";

contract DeployEspressoInput is BaseDeployIO {
    address internal _nitroEnclaveVerifier;
    address internal _espressoBatcher;
    address internal _systemConfig;
    address internal _proxyAdminOwner;

    function set(bytes4 _sel, address _val) public {
        if (_sel == this.nitroEnclaveVerifier.selector) {
            _nitroEnclaveVerifier = _val;
        } else if (_sel == this.espressoBatcher.selector) {
            _espressoBatcher = _val;
        } else if (_sel == this.systemConfig.selector) {
            _systemConfig = _val;
        } else if (_sel == this.proxyAdminOwner.selector) {
            _proxyAdminOwner = _val;
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

    /// @notice The address that will own the ProxyAdmin contracts. Defaults to msg.sender if not set.
    function proxyAdminOwner() public view returns (address) {
        return _proxyAdminOwner;
    }
}

contract DeployEspressoOutput is BaseDeployIO {
    address internal _batchAuthenticatorAddress;
    address internal _teeVerifierProxy;
    address internal _teeVerifierProxyAdmin;
    address internal _nitroTEEVerifier;

    function set(bytes4 _sel, address _addr) public {
        require(_addr != address(0), "DeployEspressoOutput: cannot set zero address");
        if (_sel == this.batchAuthenticatorAddress.selector) {
            _batchAuthenticatorAddress = _addr;
        } else if (_sel == this.teeVerifierProxy.selector) {
            _teeVerifierProxy = _addr;
        } else if (_sel == this.teeVerifierProxyAdmin.selector) {
            _teeVerifierProxyAdmin = _addr;
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

    function teeVerifierProxyAdmin() public view returns (address) {
        require(_teeVerifierProxyAdmin != address(0), "DeployEspressoOutput: tee verifier proxy admin not set");
        return _teeVerifierProxyAdmin;
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
        address proxyAdminOwner = _input.proxyAdminOwner();
        if (proxyAdminOwner == address(0)) proxyAdminOwner = _deployerAddress;

        vm.broadcast(msg.sender);
        IProxyAdmin proxyAdmin = _deployProxyAdmin(msg.sender);
        vm.label(address(proxyAdmin), "BatchAuthenticatorProxyAdmin");
        // Deploy Proxy without importing Proxy.sol to avoid duplicate compilation artifacts.
        IProxy proxy;
        {
            bytes memory initCode =
                abi.encodePacked(vm.getCode("src/universal/Proxy.sol:Proxy"), abi.encode(address(proxyAdmin)));
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
        proxyAdmin.setProxyType(address(proxy), IProxyAdmin.ProxyType.ERC1967);
        vm.broadcast(msg.sender);
        BatchAuthenticator impl = new BatchAuthenticator();
        vm.label(address(impl), "BatchAuthenticatorImpl");

        bytes memory initData = abi.encodeCall(
            BatchAuthenticator.initialize,
            (
                _teeVerifier,
                _input.espressoBatcher(),
                ISystemConfig(_input.systemConfig()),
                proxyAdminOwner,
                // First deployment: start with the Espresso batcher active.
                true
            )
        );
        vm.broadcast(msg.sender);
        proxyAdmin.upgradeAndCall(payable(address(proxy)), address(impl), initData);

        if (proxyAdminOwner != msg.sender) {
            vm.broadcast(msg.sender);
            proxyAdmin.transferOwnership(proxyAdminOwner);
        }

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
            return _deployMockTEEContracts(_input, _output);
        }
        return _deployProductionTEEContracts(_input, _output, _deployerAddress, nitroEnclaveVerifier);
    }

    function _deployMockTEEContracts(
        DeployEspressoInput _input,
        DeployEspressoOutput _output
    )
        internal
        returns (IEspressoTEEVerifier)
    {
        address proxyAdminOwner = _input.proxyAdminOwner();
        if (proxyAdminOwner == address(0)) proxyAdminOwner = msg.sender;

        // Use our local mocks — they carry OP-specific test behavior (permissive isSignerValid,
        // test helper overrides, special address exceptions) that the submodule mocks don't have.
        vm.broadcast(msg.sender);
        MockEspressoNitroTEEVerifier nitroMock = new MockEspressoNitroTEEVerifier();
        vm.label(address(nitroMock), "MockEspressoNitroTEEVerifier");

        vm.broadcast(msg.sender);
        MockEspressoTEEVerifier teeMock = new MockEspressoTEEVerifier(IEspressoNitroTEEVerifier(address(nitroMock)));
        vm.label(address(teeMock), "MockEspressoTEEVerifier");

        // Deploy a dummy ProxyAdmin so the output proxy-admin field is a valid distinct address.
        vm.broadcast(msg.sender);
        IProxyAdmin dummyAdmin = _deployProxyAdmin(proxyAdminOwner);
        vm.label(address(dummyAdmin), "MockTEEVerifierDummyProxyAdmin");

        _output.set(_output.nitroTEEVerifier.selector, address(nitroMock));
        _output.set(_output.teeVerifierProxy.selector, address(teeMock));
        _output.set(_output.teeVerifierProxyAdmin.selector, address(dummyAdmin));
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
        address proxyAdminOwner = _input.proxyAdminOwner();
        if (proxyAdminOwner == address(0)) proxyAdminOwner = _deployerAddress;

        // Deploy OP's ProxyAdmin (owned by msg.sender for now so we can upgradeAndCall).
        vm.broadcast(msg.sender);
        IProxyAdmin proxyAdmin = _deployProxyAdmin(msg.sender);
        vm.label(address(proxyAdmin), "TEEVerifierProxyAdmin");

        // Deploy OP's ERC-1967 Proxy pointing at the ProxyAdmin.
        address payable teeProxyAddr;
        {
            bytes memory initCode =
                abi.encodePacked(vm.getCode("src/universal/Proxy.sol:Proxy"), abi.encode(address(proxyAdmin)));
            vm.broadcast(msg.sender);
            assembly {
                teeProxyAddr := create(0, add(initCode, 0x20), mload(initCode))
            }
            require(teeProxyAddr != address(0), "DeployEspresso: tee proxy deployment failed");
        }
        vm.label(teeProxyAddr, "TEEVerifierProxy");

        vm.broadcast(msg.sender);
        proxyAdmin.setProxyType(teeProxyAddr, IProxyAdmin.ProxyType.ERC1967);

        // Deploy the implementation and initialize with the configured owner. The contract uses
        // OZ Ownable2Step under the hood, so setting the final owner via `initialize` avoids
        // the two-step transfer dance.
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

        // initialize(address _owner, address _espressoNitroTEEVerifier)
        bytes memory initData = abi.encodeWithSignature(
            "initialize(address,address)", proxyAdminOwner, address(0)
        );
        vm.broadcast(msg.sender);
        proxyAdmin.upgradeAndCall(teeProxyAddr, address(teeImpl), initData);

        if (proxyAdminOwner != msg.sender) {
            vm.broadcast(msg.sender);
            proxyAdmin.transferOwnership(proxyAdminOwner);
        }

        // Deploy NitroTEEVerifier (no proxy; it stores teeProxy for access control).
        // Use vm.getCode against the submodule's own out/ to avoid pulling the impl closure
        // into OP's compile group.
        address nitroVerifierAddr;
        {
            bytes memory nitroImplCode = abi.encodePacked(
                vm.getCode(
                    "lib/espresso-tee-contracts/out/EspressoNitroTEEVerifier.sol/EspressoNitroTEEVerifier.json"
                ),
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

        // Wire the verifier into the TEE verifier. `setEspressoNitroTEEVerifier` is onlyOwner,
        // so this implicitly requires msg.sender == proxyAdminOwner (same constraint the
        // previous implementation had).
        vm.broadcast(msg.sender);
        IEspressoTEEVerifier(teeProxyAddr).setEspressoNitroTEEVerifier(
            IEspressoNitroTEEVerifier(address(nitroVerifier))
        );

        _output.set(_output.teeVerifierProxy.selector, teeProxyAddr);
        _output.set(_output.teeVerifierProxyAdmin.selector, address(proxyAdmin));
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
        require(
            _output.teeVerifierProxy() != _output.teeVerifierProxyAdmin(),
            "DeployEspresso: tee proxy and proxy admin should be different"
        );
    }

    /// @notice Deploys a ProxyAdmin via vm.getCode to avoid importing src/universal/ProxyAdmin.sol or
    ///         scripts/libraries/DeployUtils.sol, which would merge into the 0.8.28 compilation group
    ///         alongside files that import src/universal/Proxy.sol, creating duplicate Proxy artifacts.
    ///         The explicit artifact path is used to deterministically resolve to the default
    ///         compilation profile's bytecode (a plain `vm.getCode("ProxyAdmin")` is ambiguous when
    ///         ProxyAdmin is also compiled in the dispute profile via transitive imports).
    function _deployProxyAdmin(address _owner) internal returns (IProxyAdmin proxyAdmin_) {
        bytes memory _initCode =
            abi.encodePacked(vm.getCode("forge-artifacts/ProxyAdmin.sol/ProxyAdmin.json"), abi.encode(_owner));
        address payable _addr;
        assembly {
            _addr := create(0, add(_initCode, 0x20), mload(_initCode))
        }
        require(_addr != address(0), "DeployEspresso: ProxyAdmin deployment failed");
        proxyAdmin_ = IProxyAdmin(_addr);
    }
}
