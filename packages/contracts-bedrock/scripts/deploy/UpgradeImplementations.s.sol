// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";

import { LibString } from "@solady/utils/LibString.sol";

// Libraries
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";
import { Solarray } from "scripts/libraries/Solarray.sol";
import { BaseDeployIO } from "scripts/deploy/BaseDeployIO.sol";
import { IMulticall3 } from "forge-std/interfaces/IMulticall3.sol";

// Interfaces
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { ISuperchainConfig } from "interfaces/L1/ISuperchainConfig.sol";
import { IProtocolVersions } from "interfaces/L1/IProtocolVersions.sol";
import { IOptimismPortal2 } from "interfaces/L1/IOptimismPortal2.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IL1CrossDomainMessenger } from "interfaces/L1/IL1CrossDomainMessenger.sol";
import { IL1ERC721Bridge } from "interfaces/L1/IL1ERC721Bridge.sol";
import { IL1StandardBridge } from "interfaces/L1/IL1StandardBridge.sol";
import { IOptimismMintableERC20Factory } from "interfaces/universal/IOptimismMintableERC20Factory.sol";
import { IDisputeGameFactory } from "interfaces/dispute/IDisputeGameFactory.sol";
import { IAnchorStateRegistry } from "interfaces/dispute/IAnchorStateRegistry.sol";
import { IDelayedWETH } from "interfaces/dispute/IDelayedWETH.sol";

// Import the DeployImplementations script and its input/output contracts
import { DeployImplementations, DeployImplementationsInput, DeployImplementationsOutput } from "scripts/deploy/DeployImplementations.s.sol";

contract UpgradeImplementationsInput is BaseDeployIO {
    IProxyAdmin internal _proxyAdmin;

    // Proxy addresses that need to be upgraded
    address internal _superchainConfigProxy;
    address internal _protocolVersionsProxy;
    address internal _optimismPortalProxy;
    address internal _systemConfigProxy;
    address internal _l1CrossDomainMessengerProxy;
    address internal _l1ERC721BridgeProxy;
    address internal _l1StandardBridgeProxy;
    address internal _optimismMintableERC20FactoryProxy;
    address internal _disputeGameFactoryProxy;
    address internal _anchorStateRegistryProxy;
    address internal _delayedWETHProxy;

    function set(bytes4 _sel, address _addr) public {
        require(_addr != address(0), "UpgradeImplementationsInput: cannot set zero address");

        if (_sel == this.proxyAdmin.selector) _proxyAdmin = IProxyAdmin(_addr);
        else if (_sel == this.superchainConfigProxy.selector) _superchainConfigProxy = _addr;
        else if (_sel == this.protocolVersionsProxy.selector) _protocolVersionsProxy = _addr;
        else if (_sel == this.optimismPortalProxy.selector) _optimismPortalProxy = _addr;
        else if (_sel == this.systemConfigProxy.selector) _systemConfigProxy = _addr;
        else if (_sel == this.l1CrossDomainMessengerProxy.selector) _l1CrossDomainMessengerProxy = _addr;
        else if (_sel == this.l1ERC721BridgeProxy.selector) _l1ERC721BridgeProxy = _addr;
        else if (_sel == this.l1StandardBridgeProxy.selector) _l1StandardBridgeProxy = _addr;
        else if (_sel == this.optimismMintableERC20FactoryProxy.selector) _optimismMintableERC20FactoryProxy = _addr;
        else if (_sel == this.disputeGameFactoryProxy.selector) _disputeGameFactoryProxy = _addr;
        else if (_sel == this.anchorStateRegistryProxy.selector) _anchorStateRegistryProxy = _addr;
        else if (_sel == this.delayedWETHProxy.selector) _delayedWETHProxy = _addr;
        else revert("UpgradeImplementationsInput: unknown selector");
    }

    function proxyAdmin() public view returns (IProxyAdmin) {
        require(address(_proxyAdmin) != address(0), "UpgradeImplementationsInput: not set");
        return _proxyAdmin;
    }

    function superchainConfigProxy() public view returns (address) {
        return _superchainConfigProxy;
    }

    function protocolVersionsProxy() public view returns (address) {
        return _protocolVersionsProxy;
    }

    function optimismPortalProxy() public view returns (address) {
        require(_optimismPortalProxy != address(0), "UpgradeImplementationsInput: not set");
        return _optimismPortalProxy;
    }

    function systemConfigProxy() public view returns (address) {
        require(_systemConfigProxy != address(0), "UpgradeImplementationsInput: not set");
        return _systemConfigProxy;
    }

    function l1CrossDomainMessengerProxy() public view returns (address) {
        require(_l1CrossDomainMessengerProxy != address(0), "UpgradeImplementationsInput: not set");
        return _l1CrossDomainMessengerProxy;
    }

    function l1ERC721BridgeProxy() public view returns (address) {
        require(_l1ERC721BridgeProxy != address(0), "UpgradeImplementationsInput: not set");
        return _l1ERC721BridgeProxy;
    }

    function l1StandardBridgeProxy() public view returns (address) {
        require(_l1StandardBridgeProxy != address(0), "UpgradeImplementationsInput: not set");
        return _l1StandardBridgeProxy;
    }

    function optimismMintableERC20FactoryProxy() public view returns (address) {
        require(_optimismMintableERC20FactoryProxy != address(0), "UpgradeImplementationsInput: not set");
        return _optimismMintableERC20FactoryProxy;
    }

    function disputeGameFactoryProxy() public view returns (address) {
        require(_disputeGameFactoryProxy != address(0), "UpgradeImplementationsInput: not set");
        return _disputeGameFactoryProxy;
    }

    function anchorStateRegistryProxy() public view returns (address) {
        require(_anchorStateRegistryProxy != address(0), "UpgradeImplementationsInput: not set");
        return _anchorStateRegistryProxy;
    }

    function delayedWETHProxy() public view returns (address) {
        return _delayedWETHProxy;
    }
}

contract UpgradeImplementationsOutput is BaseDeployIO {
    bool internal _upgradeComplete;

    function set(bytes4 _sel, bool _value) public {
        if (_sel == this.upgradeComplete.selector) _upgradeComplete = _value;
        else revert("UpgradeImplementationsOutput: unknown selector");
    }

    function upgradeComplete() public view returns (bool) {
        return _upgradeComplete;
    }
}

contract BaklavaUpgradeImplementations is Script {
    // GnosisSafe address
    address constant GNOSIS_SAFE = 0xd542f3328ff2516443FE4db1c89E427F67169D94;

    function run() external {
        // setup
        console.log("Setup started!");
        console.log("GnosisSafe address:", GNOSIS_SAFE);
        DeployImplementationsInput dii = new DeployImplementationsInput();
        dii.set(DeployImplementationsInput.withdrawalDelaySeconds.selector, 302400);
        dii.set(DeployImplementationsInput.minProposalSizeBytes.selector, 126000);
        dii.set(DeployImplementationsInput.challengePeriodSeconds.selector, 86400);
        dii.set(DeployImplementationsInput.proofMaturityDelaySeconds.selector, 604800);
        dii.set(DeployImplementationsInput.disputeGameFinalityDelaySeconds.selector, 302400);
        dii.set(DeployImplementationsInput.mipsVersion.selector, 1);
        dii.set(DeployImplementationsInput.l1ContractsRelease.selector, "op-contracts/v2.0.0");
        dii.set(DeployImplementationsInput.superchainConfigProxy.selector, address(0xf07502A4a950d870c43b12660fB1Dd18c170D344));
        dii.set(DeployImplementationsInput.protocolVersionsProxy.selector, address(0x3d438C63e0431DA844d3F60E6c712d10FC75c529));
        dii.set(DeployImplementationsInput.superchainProxyAdmin.selector, address(0xBF101Bd81fb69aB00ab261465454dF1a171726Bf));
        dii.set(DeployImplementationsInput.upgradeController.selector, address(0xd542f3328ff2516443FE4db1c89E427F67169D94));

        UpgradeImplementationsInput uii = new UpgradeImplementationsInput();
        // Set proxy addresses from deployment
        uii.set(uii.proxyAdmin.selector, address(0xBF101Bd81fb69aB00ab261465454dF1a171726Bf));
        uii.set(uii.superchainConfigProxy.selector, address(0xf07502A4a950d870c43b12660fB1Dd18c170D344));
        uii.set(uii.protocolVersionsProxy.selector, address(0x3d438C63e0431DA844d3F60E6c712d10FC75c529));
        uii.set(uii.optimismPortalProxy.selector, address(0x87e9cB54f185a32266689138fbA56F0C994CF50c));
        uii.set(uii.systemConfigProxy.selector, address(0x3ee24bF404e4a5D27A437d910F56E1eD999B1De8));
        uii.set(uii.l1CrossDomainMessengerProxy.selector, address(0x418F16753b868adDc5f0C2860b05D8b921fCee75));
        uii.set(uii.l1ERC721BridgeProxy.selector, address(0xe65F0FE051FCe8a8De3A5c8DcFa474025F6F2D73));
        uii.set(uii.l1StandardBridgeProxy.selector, address(0x6fd3fF186975aD8B66Ab40b705EC016b36da0486));
        uii.set(uii.optimismMintableERC20FactoryProxy.selector, address(0x0e8e9173e163C593AcF823BA0880f25bCF40fEa1));
        uii.set(uii.disputeGameFactoryProxy.selector, address(0x788ef5850c3a51d41f59Dc4327017EF8D754eD80));
        uii.set(uii.anchorStateRegistryProxy.selector, address(0x9C4B83AED7e5103dC6F38DfA265933Fe418Ed090));
        uii.set(uii.delayedWETHProxy.selector, address(0x9B014033014c8C9026Dbc5151049ffFA8581942E));

        // execution
        console.log("Execution!");
        UpgradeImplementations upgrade = new UpgradeImplementations();
        upgrade.run(uii, new UpgradeImplementationsOutput(), dii);
    }
}

contract UpgradeImplementations is Script {

    function run(
        UpgradeImplementationsInput _uii,
        UpgradeImplementationsOutput _uio,
        DeployImplementationsInput _dii
    ) public {
        console.log("Deploying new implementations...");

        // Deploy new implementations using DeployImplementations following BaklavaDeployImplementations pattern
        DeployImplementations deployImplementations = new DeployImplementations();
        DeployImplementationsOutput _dio = new DeployImplementationsOutput();

        // Call DeployImplementations to get fresh implementation addresses
        deployImplementations.run(_dii, _dio);

        console.log("New implementations deployed successfully!");
        console.log("Starting implementation upgrades...");

        IProxyAdmin proxyAdmin = _uii.proxyAdmin();

        // Check if ProxyAdmin is owned by a Gnosis Safe (caller is not the owner)
        address proxyAdminOwner = proxyAdmin.owner();
        console.log("ProxyAdmin owner:", proxyAdminOwner);
        console.log("Transaction sender:", msg.sender);

        if (proxyAdminOwner != msg.sender) {
            console.log("ProxyAdmin is owned by a different address (likely Gnosis Safe)");
            console.log("Generating transaction data for Safe submission instead of direct execution...");

            // Generate Safe transaction data instead of executing directly
            generateSafeTransactionData(_uii, _dio);

            _uio.set(_uio.upgradeComplete.selector, true);
            console.log("Transaction data generated successfully! Submit to Gnosis Safe for execution.");
            return;
        }

        console.log("ProxyAdmin is owned by caller - proceeding with direct execution...");

        // Cache critical addresses to avoid static calls after broadcast
        // Split into smaller groups to avoid stack too deep error

        // Cache superchain proxy addresses
        address superchainConfigProxy = _uii.superchainConfigProxy();
        address protocolVersionsProxy = _uii.protocolVersionsProxy();
        address delayedWETHProxy = _uii.delayedWETHProxy();

        // Cache superchain implementation addresses
        address superchainConfigImpl = address(_dio.superchainConfigImpl());
        address protocolVersionsImpl = address(_dio.protocolVersionsImpl());
        address delayedWETHImpl = address(_dio.delayedWETHImpl());

        vm.startBroadcast();

        // Upgrade SuperchainConfig if proxy address is provided
        if (superchainConfigProxy != address(0)) {
            console.log("Upgrading SuperchainConfig proxy...");
            proxyAdmin.upgrade(
                payable(superchainConfigProxy),
                superchainConfigImpl
            );
            console.log("SuperchainConfig upgraded to:", superchainConfigImpl);
        }

        // Upgrade ProtocolVersions if proxy address is provided
        if (protocolVersionsProxy != address(0)) {
            console.log("Upgrading ProtocolVersions proxy...");
            proxyAdmin.upgrade(
                payable(protocolVersionsProxy),
                protocolVersionsImpl
            );
            console.log("ProtocolVersions upgraded to:", protocolVersionsImpl);
        }

        // Upgrade OptimismPortal
        console.log("Upgrading OptimismPortal proxy...");
        proxyAdmin.upgrade(
            payable(_uii.optimismPortalProxy()),
            address(_dio.optimismPortalImpl())
        );
        console.log("OptimismPortal upgraded to:", address(_dio.optimismPortalImpl()));

        // Upgrade SystemConfig
        console.log("Upgrading SystemConfig proxy...");
        proxyAdmin.upgrade(
            payable(_uii.systemConfigProxy()),
            address(_dio.systemConfigImpl())
        );
        console.log("SystemConfig upgraded to:", address(_dio.systemConfigImpl()));

        // Upgrade L1CrossDomainMessenger
        console.log("Upgrading L1CrossDomainMessenger proxy...");
        proxyAdmin.upgrade(
            payable(_uii.l1CrossDomainMessengerProxy()),
            address(_dio.l1CrossDomainMessengerImpl())
        );
        console.log("L1CrossDomainMessenger upgraded to:", address(_dio.l1CrossDomainMessengerImpl()));

        // Upgrade L1ERC721Bridge
        console.log("Upgrading L1ERC721Bridge proxy...");
        proxyAdmin.upgrade(
            payable(_uii.l1ERC721BridgeProxy()),
            address(_dio.l1ERC721BridgeImpl())
        );
        console.log("L1ERC721Bridge upgraded to:", address(_dio.l1ERC721BridgeImpl()));

        // Upgrade L1StandardBridge
        console.log("Upgrading L1StandardBridge proxy...");
        proxyAdmin.upgrade(
            payable(_uii.l1StandardBridgeProxy()),
            address(_dio.l1StandardBridgeImpl())
        );
        console.log("L1StandardBridge upgraded to:", address(_dio.l1StandardBridgeImpl()));

        // Upgrade OptimismMintableERC20Factory
        console.log("Upgrading OptimismMintableERC20Factory proxy...");
        proxyAdmin.upgrade(
            payable(_uii.optimismMintableERC20FactoryProxy()),
            address(_dio.optimismMintableERC20FactoryImpl())
        );
        console.log("OptimismMintableERC20Factory upgraded to:", address(_dio.optimismMintableERC20FactoryImpl()));

        // Upgrade DisputeGameFactory
        console.log("Upgrading DisputeGameFactory proxy...");
        proxyAdmin.upgrade(
            payable(_uii.disputeGameFactoryProxy()),
            address(_dio.disputeGameFactoryImpl())
        );
        console.log("DisputeGameFactory upgraded to:", address(_dio.disputeGameFactoryImpl()));

        // Upgrade AnchorStateRegistry
        console.log("Upgrading AnchorStateRegistry proxy...");
        proxyAdmin.upgrade(
            payable(_uii.anchorStateRegistryProxy()),
            address(_dio.anchorStateRegistryImpl())
        );
        console.log("AnchorStateRegistry upgraded to:", address(_dio.anchorStateRegistryImpl()));

        // Upgrade DelayedWETH if proxy address is provided
        if (delayedWETHProxy != address(0)) {
            console.log("Upgrading DelayedWETH proxy...");
            proxyAdmin.upgrade(
                payable(delayedWETHProxy),
                delayedWETHImpl
            );
            console.log("DelayedWETH upgraded to:", delayedWETHImpl);
        }

        vm.stopBroadcast();

        _uio.set(_uio.upgradeComplete.selector, true);
        console.log("All implementation upgrades completed successfully!");
    }

    function etchIOContracts() public returns (UpgradeImplementationsInput uii_, UpgradeImplementationsOutput uio_) {
        (uii_, uio_) = getIOContracts();

        DeployUtils.etchLabelAndAllowCheatcodes({
            _etchTo: address(uii_),
            _cname: "UpgradeImplementationsInput",
            _artifactPath: "UpgradeImplementations.s.sol:UpgradeImplementationsInput"
        });

        DeployUtils.etchLabelAndAllowCheatcodes({
            _etchTo: address(uio_),
            _cname: "UpgradeImplementationsOutput",
            _artifactPath: "UpgradeImplementations.s.sol:UpgradeImplementationsOutput"
        });
    }

    function getIOContracts() public view returns (UpgradeImplementationsInput uii_, UpgradeImplementationsOutput uio_) {
        // Use tx.origin instead of msg.sender to avoid wallet mismatch issues
        uii_ = UpgradeImplementationsInput(DeployUtils.toIOAddress(tx.origin, "optimism.UpgradeImplementationsInput"));
        uio_ = UpgradeImplementationsOutput(DeployUtils.toIOAddress(tx.origin, "optimism.UpgradeImplementationsOutput"));
    }

    /// @notice Helper function to generate transaction data for Gnosis Safe execution
    /// @dev This function logs the transaction data that can be submitted to a Gnosis Safe
    /// @param _uii Input configuration for the upgrade
    /// @param _dio Output from DeployImplementations containing new implementation addresses
    function generateSafeTransactionData(
        UpgradeImplementationsInput _uii,
        DeployImplementationsOutput _dio
    ) public view {
        IProxyAdmin proxyAdmin = _uii.proxyAdmin();
        // Reference the GnosisSafe address from BaklavaUpgradeImplementations
        address gnosisSafe = 0xd542f3328ff2516443FE4db1c89E427F67169D94;

        console.log("=== GNOSIS SAFE TRANSACTION DATA ===");
        console.log("GnosisSafe address:", gnosisSafe);
        console.log("ProxyAdmin address:", address(proxyAdmin));
        console.log("ProxyAdmin owner (should be Gnosis Safe):", proxyAdmin.owner());
        console.log("");
        console.log("Copy the following transaction data to submit to Gnosis Safe:");
        console.log("");

        // Generate transaction data for each upgrade
        if (_uii.superchainConfigProxy() != address(0)) {
            bytes memory data = abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.superchainConfigProxy(),
                address(_dio.superchainConfigImpl())
            );
            console.log("SuperchainConfig upgrade:");
            console.log("  To:", address(proxyAdmin));
            console.log("  Data:", LibString.toHexString(data));
            console.log("");
        }

        if (_uii.protocolVersionsProxy() != address(0)) {
            bytes memory data = abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.protocolVersionsProxy(),
                address(_dio.protocolVersionsImpl())
            );
            console.log("ProtocolVersions upgrade:");
            console.log("  To:", address(proxyAdmin));
            console.log("  Data:", LibString.toHexString(data));
            console.log("");
        }

        // OptimismPortal upgrade
        bytes memory data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.optimismPortalProxy(),
            address(_dio.optimismPortalImpl())
        );
        console.log("OptimismPortal upgrade:");
        console.log("  To:", address(proxyAdmin));
        console.log("  Data:", LibString.toHexString(data));
        console.log("");

        // SystemConfig upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.systemConfigProxy(),
            address(_dio.systemConfigImpl())
        );
        console.log("SystemConfig upgrade:");
        console.log("  To:", address(proxyAdmin));
        console.log("  Data:", LibString.toHexString(data));
        console.log("");

        // L1CrossDomainMessenger upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.l1CrossDomainMessengerProxy(),
            address(_dio.l1CrossDomainMessengerImpl())
        );
        console.log("L1CrossDomainMessenger upgrade:");
        console.log("  To:", address(proxyAdmin));
        console.log("  Data:", LibString.toHexString(data));
        console.log("");

        // L1ERC721Bridge upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.l1ERC721BridgeProxy(),
            address(_dio.l1ERC721BridgeImpl())
        );
        console.log("L1ERC721Bridge upgrade:");
        console.log("  To:", address(proxyAdmin));
        console.log("  Data:", LibString.toHexString(data));
        console.log("");

        // L1StandardBridge upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.l1StandardBridgeProxy(),
            address(_dio.l1StandardBridgeImpl())
        );
        console.log("L1StandardBridge upgrade:");
        console.log("  To:", address(proxyAdmin));
        console.log("  Data:", LibString.toHexString(data));
        console.log("");

        // OptimismMintableERC20Factory upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.optimismMintableERC20FactoryProxy(),
            address(_dio.optimismMintableERC20FactoryImpl())
        );
        console.log("OptimismMintableERC20Factory upgrade:");
        console.log("  To:", address(proxyAdmin));
        console.log("  Data:", LibString.toHexString(data));
        console.log("");

        // DisputeGameFactory upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.disputeGameFactoryProxy(),
            address(_dio.disputeGameFactoryImpl())
        );
        console.log("DisputeGameFactory upgrade:");
        console.log("  To:", address(proxyAdmin));
        console.log("  Data:", LibString.toHexString(data));
        console.log("");

        // AnchorStateRegistry upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.anchorStateRegistryProxy(),
            address(_dio.anchorStateRegistryImpl())
        );
        console.log("AnchorStateRegistry upgrade:");
        console.log("  To:", address(proxyAdmin));
        console.log("  Data:", LibString.toHexString(data));
        console.log("");

        if (_uii.delayedWETHProxy() != address(0)) {
            data = abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.delayedWETHProxy(),
                address(_dio.delayedWETHImpl())
            );
            console.log("DelayedWETH upgrade:");
            console.log("  To:", address(proxyAdmin));
            console.log("  Data:", LibString.toHexString(data));
            console.log("");
        }

        console.log("=== END TRANSACTION DATA ===");
        console.log("");

        // Generate multicall3 bundle
        generateMulticall3Bundle(_uii, _dio);

        console.log("Instructions:");
        console.log("1. Copy each transaction data above for individual transactions");
        console.log("2. OR use the multicall3 bundle below for a single transaction");
        console.log("3. Submit to Gnosis Safe with the corresponding 'To' address");
        console.log("4. Set value to 0 for all transactions");
        console.log("5. Execute transactions in order after Safe approval");
    }

    /// @notice Generate multicall3 bundle for all upgrade transactions
    /// @dev This function creates a single multicall3 transaction that bundles all upgrades
    /// @param _uii Input configuration for the upgrade
    /// @param _dio Output from DeployImplementations containing new implementation addresses
    function generateMulticall3Bundle(
        UpgradeImplementationsInput _uii,
        DeployImplementationsOutput _dio
    ) public view {
        IProxyAdmin proxyAdmin = _uii.proxyAdmin();
        address multicall3Address = 0xcA11bde05977b3631167028862bE2a173976CA11;

        console.log("=== MULTICALL3 BUNDLE ===");
        console.log("Multicall3 address:", multicall3Address);
        console.log("ProxyAdmin address:", address(proxyAdmin));
        console.log("");

        // Build array of calls for multicall3
        IMulticall3.Call[] memory calls = new IMulticall3.Call[](0);
        uint256 callCount = 0;

        // Count total number of calls first
        if (_uii.superchainConfigProxy() != address(0)) callCount++;
        if (_uii.protocolVersionsProxy() != address(0)) callCount++;
        callCount += 9; // Required upgrades: OptimismPortal, SystemConfig, L1CrossDomainMessenger, L1ERC721Bridge, L1StandardBridge, OptimismMintableERC20Factory, DisputeGameFactory, AnchorStateRegistry
        if (_uii.delayedWETHProxy() != address(0)) callCount++;

        // Create properly sized array
        calls = new IMulticall3.Call[](callCount);
        uint256 currentIndex = 0;

        // Add SuperchainConfig upgrade if proxy address is provided
        if (_uii.superchainConfigProxy() != address(0)) {
            calls[currentIndex] = IMulticall3.Call({
                target: address(proxyAdmin),
                callData: abi.encodeWithSelector(
                    IProxyAdmin.upgrade.selector,
                    _uii.superchainConfigProxy(),
                    address(_dio.superchainConfigImpl())
                )
            });
            currentIndex++;
        }

        // Add ProtocolVersions upgrade if proxy address is provided
        if (_uii.protocolVersionsProxy() != address(0)) {
            calls[currentIndex] = IMulticall3.Call({
                target: address(proxyAdmin),
                callData: abi.encodeWithSelector(
                    IProxyAdmin.upgrade.selector,
                    _uii.protocolVersionsProxy(),
                    address(_dio.protocolVersionsImpl())
                )
            });
            currentIndex++;
        }

        // Add OptimismPortal upgrade
        calls[currentIndex] = IMulticall3.Call({
            target: address(proxyAdmin),
            callData: abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.optimismPortalProxy(),
                address(_dio.optimismPortalImpl())
            )
        });
        currentIndex++;

        // Add SystemConfig upgrade
        calls[currentIndex] = IMulticall3.Call({
            target: address(proxyAdmin),
            callData: abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.systemConfigProxy(),
                address(_dio.systemConfigImpl())
            )
        });
        currentIndex++;

        // Add L1CrossDomainMessenger upgrade
        calls[currentIndex] = IMulticall3.Call({
            target: address(proxyAdmin),
            callData: abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.l1CrossDomainMessengerProxy(),
                address(_dio.l1CrossDomainMessengerImpl())
            )
        });
        currentIndex++;

        // Add L1ERC721Bridge upgrade
        calls[currentIndex] = IMulticall3.Call({
            target: address(proxyAdmin),
            callData: abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.l1ERC721BridgeProxy(),
                address(_dio.l1ERC721BridgeImpl())
            )
        });
        currentIndex++;

        // Add L1StandardBridge upgrade
        calls[currentIndex] = IMulticall3.Call({
            target: address(proxyAdmin),
            callData: abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.l1StandardBridgeProxy(),
                address(_dio.l1StandardBridgeImpl())
            )
        });
        currentIndex++;

        // Add OptimismMintableERC20Factory upgrade
        calls[currentIndex] = IMulticall3.Call({
            target: address(proxyAdmin),
            callData: abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.optimismMintableERC20FactoryProxy(),
                address(_dio.optimismMintableERC20FactoryImpl())
            )
        });
        currentIndex++;

        // Add DisputeGameFactory upgrade
        calls[currentIndex] = IMulticall3.Call({
            target: address(proxyAdmin),
            callData: abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.disputeGameFactoryProxy(),
                address(_dio.disputeGameFactoryImpl())
            )
        });
        currentIndex++;

        // Add AnchorStateRegistry upgrade
        calls[currentIndex] = IMulticall3.Call({
            target: address(proxyAdmin),
            callData: abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.anchorStateRegistryProxy(),
                address(_dio.anchorStateRegistryImpl())
            )
        });
        currentIndex++;

        // Add DelayedWETH upgrade if proxy address is provided
        if (_uii.delayedWETHProxy() != address(0)) {
            calls[currentIndex] = IMulticall3.Call({
                target: address(proxyAdmin),
                callData: abi.encodeWithSelector(
                    IProxyAdmin.upgrade.selector,
                    _uii.delayedWETHProxy(),
                    address(_dio.delayedWETHImpl())
                )
            });
            currentIndex++;
        }

        // Encode the multicall3 aggregate call
        bytes memory multicallData = abi.encodeWithSelector(
            IMulticall3.aggregate.selector,
            calls
        );

        console.log("Multicall3 Bundle Transaction:");
        console.log("  To:", multicall3Address);
        console.log("  Value: 0");
        console.log("  Data:", LibString.toHexString(multicallData));
        console.log("");
        console.log("Total calls bundled:", callCount);
        console.log("=== END MULTICALL3 BUNDLE ===");
    }
}

// Contract specifically for generating Gnosis Safe transaction data
contract GenerateSafeUpgradeData is Script {
    function run() external {
        console.log("Generating Gnosis Safe transaction data for upgrades...");
        console.log("GnosisSafe address:", 0xd542f3328ff2516443FE4db1c89E427F67169D94);

        // Set up the same configuration as BaklavaUpgradeImplementations
        DeployImplementationsInput dii = new DeployImplementationsInput();
        dii.set(DeployImplementationsInput.withdrawalDelaySeconds.selector, 302400);
        dii.set(DeployImplementationsInput.minProposalSizeBytes.selector, 126000);
        dii.set(DeployImplementationsInput.challengePeriodSeconds.selector, 86400);
        dii.set(DeployImplementationsInput.proofMaturityDelaySeconds.selector, 604800);
        dii.set(DeployImplementationsInput.disputeGameFinalityDelaySeconds.selector, 302400);
        dii.set(DeployImplementationsInput.mipsVersion.selector, 1);
        dii.set(DeployImplementationsInput.l1ContractsRelease.selector, "op-contracts/v2.0.0");
        dii.set(DeployImplementationsInput.superchainConfigProxy.selector, address(0xf07502A4a950d870c43b12660fB1Dd18c170D344));
        dii.set(DeployImplementationsInput.protocolVersionsProxy.selector, address(0x3d438C63e0431DA844d3F60E6c712d10FC75c529));
        dii.set(DeployImplementationsInput.superchainProxyAdmin.selector, address(0xBF101Bd81fb69aB00ab261465454dF1a171726Bf));
        dii.set(DeployImplementationsInput.upgradeController.selector, address(0xd542f3328ff2516443FE4db1c89E427F67169D94));

        UpgradeImplementationsInput uii = new UpgradeImplementationsInput();
        uii.set(uii.proxyAdmin.selector, address(0xBF101Bd81fb69aB00ab261465454dF1a171726Bf));
        uii.set(uii.superchainConfigProxy.selector, address(0xf07502A4a950d870c43b12660fB1Dd18c170D344));
        uii.set(uii.protocolVersionsProxy.selector, address(0x3d438C63e0431DA844d3F60E6c712d10FC75c529));
        uii.set(uii.optimismPortalProxy.selector, address(0x87e9cB54f185a32266689138fbA56F0C994CF50c));
        uii.set(uii.systemConfigProxy.selector, address(0x3ee24bF404e4a5D27A437d910F56E1eD999B1De8));
        uii.set(uii.l1CrossDomainMessengerProxy.selector, address(0x418F16753b868adDc5f0C2860b05D8b921fCee75));
        uii.set(uii.l1ERC721BridgeProxy.selector, address(0xe65F0FE051FCe8a8De3A5c8DcFa474025F6F2D73));
        uii.set(uii.l1StandardBridgeProxy.selector, address(0x6fd3fF186975aD8B66Ab40b705EC016b36da0486));
        uii.set(uii.optimismMintableERC20FactoryProxy.selector, address(0x0e8e9173e163C593AcF823BA0880f25bCF40fEa1));
        uii.set(uii.disputeGameFactoryProxy.selector, address(0x788ef5850c3a51d41f59Dc4327017EF8D754eD80));
        uii.set(uii.anchorStateRegistryProxy.selector, address(0x9C4B83AED7e5103dC6F38DfA265933Fe418Ed090));
        uii.set(uii.delayedWETHProxy.selector, address(0x9B014033014c8C9026Dbc5151049ffFA8581942E));

        // Deploy new implementations first
        DeployImplementations deployImplementations = new DeployImplementations();
        DeployImplementationsOutput dio = new DeployImplementationsOutput();
        deployImplementations.run(dii, dio);

        // Generate the Safe transaction data
        UpgradeImplementations upgrade = new UpgradeImplementations();
        upgrade.generateSafeTransactionData(uii, dio);
    }
}

// Example usage contract that shows how to set up the upgrade
contract ExampleUpgradeImplementations is Script {
    function run() external {
        console.log("Setting up example upgrade...");

        UpgradeImplementationsInput uii = new UpgradeImplementationsInput();
        UpgradeImplementationsOutput uio = new UpgradeImplementationsOutput();
        DeployImplementationsInput dii = new DeployImplementationsInput();

        // Example: Set the ProxyAdmin address (replace with actual address)
        // uii.set(uii.proxyAdmin.selector, address(0x...));

        // Example: Set proxy addresses that need to be upgraded (replace with actual addresses)
        // uii.set(uii.optimismPortalProxy.selector, address(0x...));
        // uii.set(uii.systemConfigProxy.selector, address(0x...));
        // uii.set(uii.l1CrossDomainMessengerProxy.selector, address(0x...));
        // uii.set(uii.l1ERC721BridgeProxy.selector, address(0x...));
        // uii.set(uii.l1StandardBridgeProxy.selector, address(0x...));
        // uii.set(uii.optimismMintableERC20FactoryProxy.selector, address(0x...));
        // uii.set(uii.disputeGameFactoryProxy.selector, address(0x...));
        // uii.set(uii.anchorStateRegistryProxy.selector, address(0x...));

        // Optional: Set superchain proxy addresses if upgrading superchain contracts
        // uii.set(uii.superchainConfigProxy.selector, address(0x...));
        // uii.set(uii.protocolVersionsProxy.selector, address(0x...));
        // uii.set(uii.delayedWETHProxy.selector, address(0x...));

        // Set up DeployImplementations input parameters
        // dii.set(dii.withdrawalDelaySeconds.selector, 302400);
        // dii.set(dii.minProposalSizeBytes.selector, 126000);
        // dii.set(dii.challengePeriodSeconds.selector, 86400);
        // dii.set(dii.proofMaturityDelaySeconds.selector, 604800);
        // dii.set(dii.disputeGameFinalityDelaySeconds.selector, 302400);
        // dii.set(dii.mipsVersion.selector, 1);
        // dii.set(dii.l1ContractsRelease.selector, "op-contracts/v2.0.0");
        // dii.set(dii.superchainConfigProxy.selector, address(0x...));
        // dii.set(dii.protocolVersionsProxy.selector, address(0x...));
        // dii.set(dii.superchainProxyAdmin.selector, address(0x...));
        // dii.set(dii.upgradeController.selector, address(0x...));

        UpgradeImplementations upgrade = new UpgradeImplementations();
        upgrade.run(uii, uio, dii);
    }
}
