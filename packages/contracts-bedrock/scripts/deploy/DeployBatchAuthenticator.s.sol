// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Script, console } from "forge-std/Script.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IEspressoTEEVerifier } from "@espresso-tee-contracts/interface/IEspressoTEEVerifier.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { BatchAuthenticator } from "src/L1/BatchAuthenticator.sol";

/// @notice Deploys only the BatchAuthenticator (proxy + impl) against an existing TEEVerifier and
///         wires the proxy to an existing (shared) OP Stack ProxyAdmin.
///
/// @dev The proxy is deployed with the deployer as its transient admin so the deployer can call
///      `upgradeToAndCall` to initialize it directly, then `changeAdmin` hands the proxy over to the
///      shared ProxyAdmin (same pattern as DeployAltDA / DeployFeesDepositor).
///
/// @dev `_batchAuthenticatorOwner` is the application-level (OZ Ownable / OwnableWithGuardians) owner,
///      which gates operational setters (setEspressoBatcher, setActiveIsEspresso). It is distinct from
///      `_proxyAdmin`'s owner, which controls upgrades and `initialize` after the `changeAdmin`.
///
/// Usage:
///   forge script scripts/deploy/DeployBatchAuthenticator.s.sol:DeployBatchAuthenticator \
///     --rpc-url <L1_RPC_URL> \
///     --broadcast \
///     --private-key <DEPLOYER_KEY> \
///     --verify \
///     --etherscan-api-key <API_KEY> \
///     --sig "run(address,address,address,address,address)" \
///     <ESPRESSO_BATCHER_ADDRESS> \
///     <SYSTEM_CONFIG_ADDRESS> \
///     <TEE_VERIFIER_ADDRESS> \
///     <PROXY_ADMIN_ADDRESS> \
///     <BATCH_AUTHENTICATOR_OWNER>
contract DeployBatchAuthenticator is Script {
    function run(
        address _espressoBatcher,
        address _systemConfig,
        address _teeVerifier,
        address _proxyAdmin,
        address _batchAuthenticatorOwner
    )
        public
    {
        require(_espressoBatcher != address(0), "DeployBatchAuthenticator: espressoBatcher required");
        require(_systemConfig != address(0), "DeployBatchAuthenticator: systemConfig required");
        require(_teeVerifier != address(0), "DeployBatchAuthenticator: teeVerifier required");
        require(_proxyAdmin != address(0), "DeployBatchAuthenticator: proxyAdmin required");

        if (_batchAuthenticatorOwner == address(0)) {
            _batchAuthenticatorOwner = msg.sender;
            console.log("WARN: batchAuthenticatorOwner not set, defaulting to msg.sender");
        }

        vm.startBroadcast(msg.sender);

        // Deploy the Proxy with the deployer as its transient admin so the deployer can initialize it
        // directly below. Deploy without importing Proxy.sol to avoid duplicate compilation artifacts;
        // use the path-qualified form to disambiguate from OZ v5's proxy/Proxy.sol artifact.
        IProxy proxy;
        {
            bytes memory initCode =
                abi.encodePacked(vm.getCode("src/universal/Proxy.sol:Proxy"), abi.encode(msg.sender));
            address payable proxyAddr;
            assembly {
                proxyAddr := create(0, add(initCode, 0x20), mload(initCode))
            }
            require(proxyAddr != address(0), "DeployBatchAuthenticator: proxy deployment failed");
            proxy = IProxy(proxyAddr);
        }
        vm.label(address(proxy), "BatchAuthenticatorProxy");
        BatchAuthenticator impl = new BatchAuthenticator();
        vm.label(address(impl), "BatchAuthenticatorImpl");

        bytes memory initData = abi.encodeCall(
            BatchAuthenticator.initialize,
            (
                IEspressoTEEVerifier(_teeVerifier),
                _espressoBatcher,
                ISystemConfig(_systemConfig),
                _batchAuthenticatorOwner,
                // First deployment: start with the Espresso batcher active.
                true
            )
        );
        // Initialize directly via the proxy. The deployer is still the proxy admin at this point, so
        // BatchAuthenticator.initialize's `_assertOnlyProxyAdminOrProxyAdminOwner` check passes.
        proxy.upgradeToAndCall(address(impl), initData);

        // Hand the proxy over to the shared OP Stack ProxyAdmin. No setProxyType call is needed: the
        // ProxyAdmin treats unregistered proxies as ProxyType.ERC1967 (enum value 0), which matches
        // src/universal/Proxy.sol.
        proxy.changeAdmin(_proxyAdmin);

        vm.stopBroadcast();

        console.log("BatchAuthenticator (proxy):", address(proxy));
        console.log("BatchAuthenticator (impl): ", address(impl));
        console.log("ProxyAdmin (shared):       ", _proxyAdmin);
    }
}
