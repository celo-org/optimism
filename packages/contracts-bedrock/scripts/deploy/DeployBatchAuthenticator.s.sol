// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Script, console } from "forge-std/Script.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IEspressoTEEVerifier } from "@espresso-tee-contracts/interface/IEspressoTEEVerifier.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { BatchAuthenticator } from "src/L1/BatchAuthenticator.sol";

/// @notice Deploys only the BatchAuthenticator (proxy + impl) against an existing TEEVerifier.
///
/// Usage:
///   forge script scripts/deploy/DeployBatchAuthenticator.s.sol:DeployBatchAuthenticator \
///     --rpc-url <L1_RPC_URL> \
///     --broadcast \
///     --private-key <DEPLOYER_KEY> \
///     --verify \
///     --etherscan-api-key <API_KEY> \
///     --sig "run(address,address,address,address)" \
///     <ESPRESSO_BATCHER_ADDRESS> \
///     <SYSTEM_CONFIG_ADDRESS> \
///     <TEE_VERIFIER_ADDRESS> \
///     <PROXY_ADMIN_OWNER>
contract DeployBatchAuthenticator is Script {
    function run(
        address _espressoBatcher,
        address _systemConfig,
        address _teeVerifier,
        address _proxyAdminOwner
    )
        public
    {
        require(_espressoBatcher != address(0), "DeployBatchAuthenticator: espressoBatcher required");
        require(_systemConfig != address(0), "DeployBatchAuthenticator: systemConfig required");
        require(_teeVerifier != address(0), "DeployBatchAuthenticator: teeVerifier required");

        if (_proxyAdminOwner == address(0)) {
            _proxyAdminOwner = msg.sender;
            console.log("WARN: proxyAdminOwner not set, defaulting to msg.sender");
        }

        vm.startBroadcast(msg.sender);

        // Deploy ProxyAdmin via vm.getCode to avoid importing src/universal/ProxyAdmin.sol or
        // scripts/libraries/DeployUtils.sol, which would merge into the 0.8.28 compilation group
        // alongside files that import src/universal/Proxy.sol, creating duplicate Proxy artifacts.
        IProxyAdmin proxyAdmin;
        {
            bytes memory _initCode =
                abi.encodePacked(vm.getCode("forge-artifacts/ProxyAdmin.sol/ProxyAdmin.json"), abi.encode(msg.sender));
            address payable _addr;
            assembly {
                _addr := create(0, add(_initCode, 0x20), mload(_initCode))
            }
            require(_addr != address(0), "DeployBatchAuthenticator: ProxyAdmin deployment failed");
            proxyAdmin = IProxyAdmin(_addr);
        }
        vm.label(address(proxyAdmin), "BatchAuthenticatorProxyAdmin");
        // Deploy Proxy without importing Proxy.sol to avoid duplicate compilation artifacts.
        // Use the path-qualified form to disambiguate from OZ v5's proxy/Proxy.sol artifact.
        IProxy proxy;
        {
            bytes memory initCode =
                abi.encodePacked(vm.getCode("src/universal/Proxy.sol:Proxy"), abi.encode(address(proxyAdmin)));
            address payable proxyAddr;
            assembly {
                proxyAddr := create(0, add(initCode, 0x20), mload(initCode))
            }
            require(proxyAddr != address(0), "DeployBatchAuthenticator: proxy deployment failed");
            proxy = IProxy(proxyAddr);
        }
        vm.label(address(proxy), "BatchAuthenticatorProxy");
        proxyAdmin.setProxyType(address(proxy), IProxyAdmin.ProxyType.ERC1967);
        BatchAuthenticator impl = new BatchAuthenticator();
        vm.label(address(impl), "BatchAuthenticatorImpl");

        bytes memory initData = abi.encodeCall(
            BatchAuthenticator.initialize,
            (
                IEspressoTEEVerifier(_teeVerifier),
                _espressoBatcher,
                ISystemConfig(_systemConfig),
                _proxyAdminOwner,
                // First deployment: start with the Espresso batcher active.
                true
            )
        );
        proxyAdmin.upgradeAndCall(payable(address(proxy)), address(impl), initData);

        if (_proxyAdminOwner != msg.sender) {
            proxyAdmin.transferOwnership(_proxyAdminOwner);
        }

        vm.stopBroadcast();

        console.log("BatchAuthenticator (proxy):", address(proxy));
        console.log("BatchAuthenticator (impl): ", address(impl));
        console.log("ProxyAdmin:                ", address(proxyAdmin));
    }
}
