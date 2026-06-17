// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";

// Contracts
import { CeloGasBridgeL1 } from "src/celo/CeloGasBridgeL1.sol";
import { PortalMigrator } from "src/celo/PortalMigrator.sol";
import { Proxy } from "src/universal/Proxy.sol";
import { StandardBridge } from "src/universal/StandardBridge.sol";
import { StorageSetter } from "src/universal/StorageSetter.sol";

// Libraries
import { BaseDeployIO } from "scripts/deploy/BaseDeployIO.sol";

// Interfaces
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ICeloGasBridgeL1 } from "interfaces/celo/ICeloGasBridgeL1.sol";
import { ICrossDomainMessenger } from "interfaces/universal/ICrossDomainMessenger.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";

/// @title DeployV2Input
/// @notice Inputs for the `DeployV2` pre-deployment script.
contract DeployV2Input is BaseDeployIO {
    IERC20 internal _celoTokenL1;
    address internal _l1CrossDomainMessenger;
    address internal _systemConfig;
    address internal _l1BridgeProxyAdmin;
    address internal _otherBridge;
    uint256 internal _proofMaturityDelaySeconds;
    uint256 internal _legacyPortalBalance;

    function set(bytes4 _sel, address _value) public {
        require(_value != address(0), "DeployV2Input: cannot set zero address");

        if (_sel == this.celoTokenL1.selector) _celoTokenL1 = IERC20(_value);
        else if (_sel == this.l1CrossDomainMessenger.selector) _l1CrossDomainMessenger = _value;
        else if (_sel == this.systemConfig.selector) _systemConfig = _value;
        else if (_sel == this.l1BridgeProxyAdmin.selector) _l1BridgeProxyAdmin = _value;
        else if (_sel == this.otherBridge.selector) _otherBridge = _value;
        else revert("DeployV2Input: unknown selector");
    }

    function set(bytes4 _sel, uint256 _value) public {
        if (_sel == this.proofMaturityDelaySeconds.selector) _proofMaturityDelaySeconds = _value;
        else if (_sel == this.legacyPortalBalance.selector) _legacyPortalBalance = _value;
        else revert("DeployV2Input: unknown selector");
    }

    function celoTokenL1() public view returns (IERC20) {
        require(address(_celoTokenL1) != address(0), "DeployV2Input: celoTokenL1 not set");
        return _celoTokenL1;
    }

    function l1CrossDomainMessenger() public view returns (address) {
        require(_l1CrossDomainMessenger != address(0), "DeployV2Input: l1CrossDomainMessenger not set");
        return _l1CrossDomainMessenger;
    }

    function systemConfig() public view returns (address) {
        require(_systemConfig != address(0), "DeployV2Input: systemConfig not set");
        return _systemConfig;
    }

    function l1BridgeProxyAdmin() public view returns (address) {
        require(_l1BridgeProxyAdmin != address(0), "DeployV2Input: l1BridgeProxyAdmin not set");
        return _l1BridgeProxyAdmin;
    }

    function otherBridge() public view returns (address) {
        require(_otherBridge != address(0), "DeployV2Input: otherBridge not set");
        return _otherBridge;
    }

    function proofMaturityDelaySeconds() public view returns (uint256) {
        require(_proofMaturityDelaySeconds != 0, "DeployV2Input: proofMaturityDelaySeconds not set");
        return _proofMaturityDelaySeconds;
    }

    function legacyPortalBalance() public view returns (uint256) {
        return _legacyPortalBalance;
    }
}

/// @title DeployV2Output
/// @notice Outputs of the `DeployV2` pre-deployment script.
contract DeployV2Output is BaseDeployIO {
    CeloGasBridgeL1 internal _celoGasBridgeL1Impl;
    ICeloGasBridgeL1 internal _celoGasBridgeL1Proxy;
    PortalMigrator internal _portalMigratorImpl;
    StorageSetter internal _storageSetter;

    function set(bytes4 _sel, address _value) public {
        require(_value != address(0), "DeployV2Output: cannot set zero address");

        if (_sel == this.celoGasBridgeL1Impl.selector) _celoGasBridgeL1Impl = CeloGasBridgeL1(payable(_value));
        else if (_sel == this.celoGasBridgeL1Proxy.selector) _celoGasBridgeL1Proxy = ICeloGasBridgeL1(payable(_value));
        else if (_sel == this.portalMigratorImpl.selector) _portalMigratorImpl = PortalMigrator(payable(_value));
        else if (_sel == this.storageSetter.selector) _storageSetter = StorageSetter(_value);
        else revert("DeployV2Output: unknown selector");
    }

    function celoGasBridgeL1Impl() public view returns (CeloGasBridgeL1) {
        require(address(_celoGasBridgeL1Impl) != address(0), "DeployV2Output: celoGasBridgeL1Impl not set");
        return _celoGasBridgeL1Impl;
    }

    function celoGasBridgeL1Proxy() public view returns (ICeloGasBridgeL1) {
        require(address(_celoGasBridgeL1Proxy) != address(0), "DeployV2Output: celoGasBridgeL1Proxy not set");
        return _celoGasBridgeL1Proxy;
    }

    function portalMigratorImpl() public view returns (PortalMigrator) {
        require(address(_portalMigratorImpl) != address(0), "DeployV2Output: portalMigratorImpl not set");
        return _portalMigratorImpl;
    }

    function storageSetter() public view returns (StorageSetter) {
        require(address(_storageSetter) != address(0), "DeployV2Output: storageSetter not set");
        return _storageSetter;
    }
}

/// @title DeployV2
/// @notice Pre-deployment script for the CGT v1 → v2 migration.
/// @dev Deploys the CeloGasBridgeL1 proxy and PortalMigrator implementation.
contract DeployV2 is Script {
    function run(DeployV2Input _input, DeployV2Output _output) external {
        // ---- Start broadcasting transactions ----
        vm.startBroadcast(msg.sender);

        // ---- Deploy CeloGasBridgeL1 implementation ----
        CeloGasBridgeL1 celoGasBridgeL1Impl = new CeloGasBridgeL1();
        _output.set(_output.celoGasBridgeL1Impl.selector, address(celoGasBridgeL1Impl));
        console.log("CeloGasBridgeL1 implementation deployed at:", address(celoGasBridgeL1Impl));

        // ---- Deploy CeloGasBridgeL1 proxy with the deployer as temporary admin ----
        Proxy celoGasBridgeL1Proxy = new Proxy(msg.sender);
        _output.set(_output.celoGasBridgeL1Proxy.selector, address(celoGasBridgeL1Proxy));
        console.log("CeloGasBridgeL1 proxy deployed at:", address(celoGasBridgeL1Proxy));

        // ---- Set implementation and initialize (deployer acts as admin) ----
        celoGasBridgeL1Proxy.upgradeToAndCall(
            address(celoGasBridgeL1Impl),
            abi.encodeCall(
                CeloGasBridgeL1.initialize,
                (
                    ICrossDomainMessenger(_input.l1CrossDomainMessenger()),
                    ISystemConfig(_input.systemConfig()),
                    StandardBridge(payable(_input.otherBridge())),
                    _input.celoTokenL1()
                )
            )
        );
        console.log("CeloGasBridgeL1 proxy initialized");

        // ---- Transfer ownership to ProxyAdmin ----
        celoGasBridgeL1Proxy.changeAdmin(_input.l1BridgeProxyAdmin());
        console.log("CeloGasBridgeL1 proxy admin set to:", _input.l1BridgeProxyAdmin());

        // ---- Deploy PortalMigrator implementation ----
        PortalMigrator portalMigratorImpl = new PortalMigrator({
            _proofMaturityDelaySeconds: _input.proofMaturityDelaySeconds(),
            _celoTokenL1: _input.celoTokenL1(),
            _gasBridge: address(celoGasBridgeL1Proxy),
            _legacyPortalBalance: _input.legacyPortalBalance()
        });
        _output.set(_output.portalMigratorImpl.selector, address(portalMigratorImpl));
        console.log("PortalMigrator implementation deployed at:", address(portalMigratorImpl));

        // ---- Deploy StorageSetter (used to change SystemConfig.superchainConfig during migration) ----
        StorageSetter storageSetter = new StorageSetter();
        _output.set(_output.storageSetter.selector, address(storageSetter));
        console.log("StorageSetter deployed at:", address(storageSetter));

        // ---- Stop broadcasting transactions ----
        vm.stopBroadcast();
    }
}
