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

            address superchainConfigImpl = address(_dio.superchainConfigImpl());
            address superchainConfigProxy = _uii.superchainConfigProxy();

            sendFirstUpgradeToGnosisSafe(address(proxyAdmin), superchainConfigProxy, superchainConfigImpl);
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
    ) public {
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

        console.log("Instructions:");
        console.log("1. Copy each transaction data above for individual transactions");
        console.log("3. Submit to Gnosis Safe with the corresponding 'To' address");
        console.log("4. Set value to 0 for all transactions");
        console.log("5. Execute transactions in order after Safe approval");
    }


    /// @notice Send the first upgrade transaction to Gnosis Safe with proper signature
    /// @dev This function submits the first upgrade transaction directly to the Gnosis Safe with proper signature
    /// @param proxyAdminAddress The address of the ProxyAdmin contract
    /// @param proxyAddress The address of the proxy to upgrade
    /// @param implementationAddress The address of the new implementation
    function sendFirstUpgradeToGnosisSafe(
        address proxyAdminAddress,
        address proxyAddress,
        address implementationAddress
    ) public {
        _logUpgradeInfo(proxyAdminAddress, proxyAddress, implementationAddress);

        bytes memory upgradeData = _encodeUpgradeData(proxyAddress, implementationAddress);
        bytes32 safeTxHash = _generateSafeTxHash(proxyAdminAddress, upgradeData);
        bytes memory signature = _signTransaction(safeTxHash);
        _executeTransaction(proxyAdminAddress, upgradeData, signature);
    }

    /// @notice Log upgrade information
    function _logUpgradeInfo(
        address proxyAdminAddress,
        address proxyAddress,
        address implementationAddress
    ) internal {
        address gnosisSafe = 0xd542f3328ff2516443FE4db1c89E427F67169D94;

        console.log("=== SENDING FIRST UPGRADE TO GNOSIS SAFE ===");
        console.log("GnosisSafe address:", gnosisSafe);
        console.log("ProxyAdmin address:", proxyAdminAddress);
        console.log("Proxy address:", proxyAddress);
        console.log("Implementation address:", implementationAddress);
        console.log("Signer (tx.origin):", tx.origin);
    }

    /// @notice Encode upgrade transaction data
    function _encodeUpgradeData(
        address proxyAddress,
        address implementationAddress
    ) internal pure returns (bytes memory) {
        return abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            proxyAddress,
            implementationAddress
        );
    }

     /// @notice Generate Safe transaction hash
    function _generateSafeTxHash(
        address proxyAdminAddress,
        bytes memory upgradeData
    ) internal returns (bytes32) {
        address gnosisSafe = 0xd542f3328ff2516443FE4db1c89E427F67169D94;

        console.log("Transaction data length:", upgradeData.length);

        (bool success, bytes memory result) = gnosisSafe.staticcall(
            abi.encodeWithSignature("nonce()")
        );
        uint256 nonce = success ? abi.decode(result, (uint256)) : 0;
        console.log("Safe nonce:", nonce);

        (success, result) = gnosisSafe.staticcall(
            abi.encodeWithSignature("domainSeparator()")
        );
        bytes32 domainSeparator = success ? abi.decode(result, (bytes32)) : bytes32(0);
        console.log("Domain separator:", LibString.toHexString(abi.encodePacked(domainSeparator)));

        bytes32 safeTxHash = keccak256(
            abi.encodePacked(
                bytes1(0x19),
                bytes1(0x01),
                domainSeparator,
                keccak256(
                    abi.encode(
                        keccak256("SafeTx(address to,uint256 value,bytes data,uint8 operation,uint256 safeTxGas,uint256 baseGas,uint256 gasPrice,address gasToken,address refundReceiver,uint256 nonce)"),
                        proxyAdminAddress,
                        0,
                        keccak256(upgradeData),
                        0,
                        0,
                        0,
                        0,
                        address(0),
                        address(0),
                        nonce
                    )
                )
            )
        );

        console.log("Safe transaction hash:", LibString.toHexString(abi.encodePacked(safeTxHash)));
        return safeTxHash;
    }

    /// @notice Sign the transaction hash
    function _signTransaction(bytes32 safeTxHash) internal returns (bytes memory) {
        uint256 privateKey = 0xa76702cf707f31b7a7b0eaebf228bcc92f22b70b7a8db278ddf0372de0a0531d;//vm.envUint("PRIVATE_KEY");
        console.log("private key", privateKey);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, safeTxHash);

        bytes memory signature = abi.encodePacked(r, s, v);
        console.log("Signature:", LibString.toHexString(signature));

        return signature;
    }

    /// @notice Execute the transaction through Gnosis Safe
    /// @dev Uses call() which is correct for executing contract functions with data.
    /// send() would be inappropriate here as it only transfers ETH with limited gas.
    function _executeTransaction(
        address proxyAdminAddress,
        bytes memory upgradeData,
        bytes memory signature
    ) internal {
        address gnosisSafe = 0xd542f3328ff2516443FE4db1c89E427F67169D94;

        // Verify the Safe contract exists at the expected address
        require(gnosisSafe.code.length > 0, "Gnosis Safe contract not found at expected address");

        vm.startBroadcast();

        // Use low-level call to execute transaction on Gnosis Safe
        // call() is the correct method for executing contract functions with custom data
        // send() would be inappropriate as it only transfers ETH with 2300 gas stipend
        (bool execSuccess, bytes memory returnData) = gnosisSafe.call(
            abi.encodeWithSignature(
                "execTransaction(address,uint256,bytes,uint8,uint256,uint256,uint256,address,address,bytes)",
                proxyAdminAddress,
                0,
                upgradeData,
                0,  // operation: 0 = CALL (actual execution)
                0,  // safeTxGas
                0,  // baseGas
                0,  // gasPrice
                address(0),  // gasToken
                address(0),  // refundReceiver
                signature
            )
        );

        vm.stopBroadcast();

        if (execSuccess) {
            console.log("Successfully executed upgrade transaction through Gnosis Safe on-chain!");
            console.log("Transaction has been broadcast and will be mined in the next block.");
        } else {
            console.log("Failed to execute transaction through Gnosis Safe");
            console.log("Error data:", LibString.toHexString(returnData));
            console.log("This may require proper multisig threshold signatures");

            // Attempt to decode common error reasons
            if (returnData.length >= 4) {
                bytes4 errorSelector = bytes4(returnData);
                if (errorSelector == bytes4(keccak256("GS013"))) {
                    console.log("Error: Invalid signature provided");
                } else if (errorSelector == bytes4(keccak256("GS020"))) {
                    console.log("Error: Signatures data too short");
                }
            }
        }

        console.log("=== UPGRADE TRANSACTION EXECUTION COMPLETE ===");
    }
}
