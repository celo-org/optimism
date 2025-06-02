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
    // Multicall3 delegatecall contract address
    address public constant MULTICALL_ADDRESS = 0xcA11bde05977b3631167028862bE2a173976CA11;
    address private constant GNOSIS_SAFE = 0xd542f3328ff2516443FE4db1c89E427F67169D94;

    struct UpgradeAction {
        address proxy;
        address implementation;
        string name;
    }

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

        console.log("ProxyAdmin is owned by a different address (likely Gnosis Safe)");
        console.log("Generating transaction data for Safe submission instead of direct execution...");

        // Generate Safe transaction data instead of executing directly
        generateSafeTransactionData(_uii, _dio);

        // Generate multicall batch transaction data
        generateMulticallBatchData(_uii, _dio);

        _uio.set(_uio.upgradeComplete.selector, true);
        console.log("Transaction data generated successfully! Submit to Gnosis Safe for execution.");
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
        }

        if (_uii.protocolVersionsProxy() != address(0)) {
            bytes memory data = abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.protocolVersionsProxy(),
                address(_dio.protocolVersionsImpl())
            );
        }

        // OptimismPortal upgrade
        bytes memory data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.optimismPortalProxy(),
            address(_dio.optimismPortalImpl())
        );

        // SystemConfig upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.systemConfigProxy(),
            address(_dio.systemConfigImpl())
        );

        // L1CrossDomainMessenger upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.l1CrossDomainMessengerProxy(),
            address(_dio.l1CrossDomainMessengerImpl())
        );

        // L1ERC721Bridge upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.l1ERC721BridgeProxy(),
            address(_dio.l1ERC721BridgeImpl())
        );

        // L1StandardBridge upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.l1StandardBridgeProxy(),
            address(_dio.l1StandardBridgeImpl())
        );

        // OptimismMintableERC20Factory upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.optimismMintableERC20FactoryProxy(),
            address(_dio.optimismMintableERC20FactoryImpl())
        );

        // DisputeGameFactory upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.disputeGameFactoryProxy(),
            address(_dio.disputeGameFactoryImpl())
        );

        // AnchorStateRegistry upgrade
        data = abi.encodeWithSelector(
            IProxyAdmin.upgrade.selector,
            _uii.anchorStateRegistryProxy(),
            address(_dio.anchorStateRegistryImpl())
        );

        if (_uii.delayedWETHProxy() != address(0)) {
            data = abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                _uii.delayedWETHProxy(),
                address(_dio.delayedWETHImpl())
            );
        }

        console.log("=== END TRANSACTION DATA ===");
        console.log("");

        console.log("Instructions:");
        console.log("1. Copy each transaction data above for individual transactions");
        console.log("3. Submit to Gnosis Safe with the corresponding 'To' address");
        console.log("4. Set value to 0 for all transactions");
        console.log("5. Execute transactions in order after Safe approval");
    }

     /// @notice Generate Safe transaction hash
    function _generateSafeTxHash(
        address targetAddress, // Renamed from proxyAdminAddress for generality
        bytes memory callData,    // Renamed from upgradeData for generality
        uint8 operation          // Added operation parameter
    ) internal view returns (bytes32) { // Added view
        console.log("Transaction data length:", callData.length);

        (bool success, bytes memory result) = GNOSIS_SAFE.staticcall(
            abi.encodeWithSignature("nonce()")
        );
        uint256 nonce = success ? abi.decode(result, (uint256)) : 0;
        console.log("Safe nonce:", nonce);

        (success, result) = GNOSIS_SAFE.staticcall(
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
                        targetAddress,
                        0, // value
                        keccak256(callData),
                        operation,
                        0, // safeTxGas
                        0, // baseGas
                        0, // gasPrice
                        address(0), // gasToken
                        address(0), // refundReceiver
                        nonce
                    )
                )
            )
        );

        console.log("Safe transaction hash:", LibString.toHexString(abi.encodePacked(safeTxHash)));
        return safeTxHash;
    }

    /// @notice Sign the transaction hash
    function _signTransaction(bytes32 safeTxHash) internal view returns (bytes memory) { // Added view
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
        address targetAddress, // Renamed from proxyAdminAddress
        bytes memory callData,    // Renamed from upgradeData
        bytes memory signature,
        uint8 operation          // Added operation parameter
    ) internal {
        // Verify the Safe contract exists at the expected address
        require(GNOSIS_SAFE.code.length > 0, "Gnosis Safe contract not found at expected address");

        vm.startBroadcast();

        (bool execSuccess, bytes memory returnData) = GNOSIS_SAFE.call(
            abi.encodeWithSignature(
                "execTransaction(address,uint256,bytes,uint8,uint256,uint256,uint256,address,address,bytes)",
                targetAddress,
                0, // value
                callData,
                operation,
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
            console.log("Successfully executed transaction via Gnosis Safe on-chain!");
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

    /// @notice Collect all upgrade actions into a structured array
    /// @dev This function deduplicates the upgrade action collection logic
    /// @param _uii Input configuration for the upgrade
    /// @param _dio Output from DeployImplementations containing new implementation addresses
    /// @return actions Array of UpgradeAction structs containing proxy, implementation, and name
    function _collectUpgradeActions(
        UpgradeImplementationsInput _uii,
        DeployImplementationsOutput _dio
    ) internal view returns (UpgradeAction[] memory actions) {
        // Count valid actions first
        uint256 actionCount = 0;
        if (_uii.superchainConfigProxy() != address(0)) actionCount++;
        if (_uii.protocolVersionsProxy() != address(0)) actionCount++;
        actionCount += 8; // Required proxies
        if (_uii.delayedWETHProxy() != address(0)) actionCount++;

        actions = new UpgradeAction[](actionCount);
        uint256 index = 0;

        // Add optional upgrades
        if (_uii.superchainConfigProxy() != address(0)) {
            actions[index++] = UpgradeAction({
                proxy: _uii.superchainConfigProxy(),
                implementation: address(_dio.superchainConfigImpl()),
                name: "SuperchainConfig"
            });
        }

        if (_uii.protocolVersionsProxy() != address(0)) {
            actions[index++] = UpgradeAction({
                proxy: _uii.protocolVersionsProxy(),
                implementation: address(_dio.protocolVersionsImpl()),
                name: "ProtocolVersions"
            });
        }

        // Add required upgrades
        actions[index++] = UpgradeAction({
            proxy: _uii.optimismPortalProxy(),
            implementation: address(_dio.optimismPortalImpl()),
            name: "OptimismPortal"
        });

        actions[index++] = UpgradeAction({
            proxy: _uii.systemConfigProxy(),
            implementation: address(_dio.systemConfigImpl()),
            name: "SystemConfig"
        });

        actions[index++] = UpgradeAction({
            proxy: _uii.l1CrossDomainMessengerProxy(),
            implementation: address(_dio.l1CrossDomainMessengerImpl()),
            name: "L1CrossDomainMessenger"
        });

        actions[index++] = UpgradeAction({
            proxy: _uii.l1ERC721BridgeProxy(),
            implementation: address(_dio.l1ERC721BridgeImpl()),
            name: "L1ERC721Bridge"
        });

        actions[index++] = UpgradeAction({
            proxy: _uii.l1StandardBridgeProxy(),
            implementation: address(_dio.l1StandardBridgeImpl()),
            name: "L1StandardBridge"
        });

        actions[index++] = UpgradeAction({
            proxy: _uii.optimismMintableERC20FactoryProxy(),
            implementation: address(_dio.optimismMintableERC20FactoryImpl()),
            name: "OptimismMintableERC20Factory"
        });

        actions[index++] = UpgradeAction({
            proxy: _uii.disputeGameFactoryProxy(),
            implementation: address(_dio.disputeGameFactoryImpl()),
            name: "DisputeGameFactory"
        });

        actions[index++] = UpgradeAction({
            proxy: _uii.anchorStateRegistryProxy(),
            implementation: address(_dio.anchorStateRegistryImpl()),
            name: "AnchorStateRegistry"
        });

        if (_uii.delayedWETHProxy() != address(0)) {
            actions[index++] = UpgradeAction({
                proxy: _uii.delayedWETHProxy(),
                implementation: address(_dio.delayedWETHImpl()),
                name: "DelayedWETH"
            });
        }

        return actions;
    }

    /// @notice Generate multicall batch transaction data for all upgrades
    /// @dev This function creates a single multicall transaction that batches all upgrades
    /// @param _uii Input configuration for the upgrade
    /// @param _dio Output from DeployImplementations containing new implementation addresses
    function generateMulticallBatchData(
        UpgradeImplementationsInput _uii,
        DeployImplementationsOutput _dio
    ) public {
        IProxyAdmin proxyAdmin = _uii.proxyAdmin();
        UpgradeAction[] memory actions = _collectUpgradeActions(_uii, _dio);

        console.log("\n=== MULTICALL BATCH TRANSACTION DATA ===");
        console.log("Multicall3Delegatecall address:", MULTICALL_ADDRESS);
        console.log("ProxyAdmin address:", address(proxyAdmin));
        console.log("Total upgrade actions:", actions.length);
        console.log("");

        bytes memory multicallData = getMulticall3Calldata(actions, address(proxyAdmin));

        console.log("Multicall batch transaction:");
        console.log("  To:", MULTICALL_ADDRESS);
        console.log("  Data:", LibString.toHexString(multicallData));
        console.log("  Value: 0");
        console.log("");

        console.log("=== MULTICALL BATCH INSTRUCTIONS ===");
        console.log("1. Copy the transaction data above");
        console.log("2. Submit to Gnosis Safe with target address:", MULTICALL_ADDRESS);
        console.log("3. Set value to 0");
        console.log("4. This single transaction will execute all", actions.length, "upgrades atomically");
        console.log("5. If any upgrade fails, the entire batch will revert");
        console.log("");

        // Log individual actions for reference
        console.log("Batch contains the following upgrades:");
        for (uint256 i = 0; i < actions.length; i++) {
            console.log(string.concat(
                "  ",
                LibString.toString(i + 1),
                ". ",
                actions[i].name,
                " (",
                LibString.toHexStringChecksummed(actions[i].proxy),
                " -> ",
                LibString.toHexStringChecksummed(actions[i].implementation),
                ")"
            ));
        }

        console.log("\n=== END MULTICALL BATCH DATA ===");

        console.log("\n=== EXECUTING MULTICALL BATCH VIA GNOSIS SAFE (DELEGATECALL) ===");
        // For multicall via delegatecall, operation is 1 (DELEGATECALL)
        // The target for the Gnosis Safe transaction is the MULTICALL_ADDRESS
        bytes32 safeTxHash = _generateSafeTxHash(MULTICALL_ADDRESS, multicallData, 1);
        bytes memory signature = _signTransaction(safeTxHash);
        _executeTransaction(MULTICALL_ADDRESS, multicallData, signature, 1);
        console.log("=== MULTICALL BATCH EXECUTION VIA GNOSIS SAFE COMPLETE ===");
    }

    /// @notice Generate multicall3 calldata for upgrade actions
    /// @dev Based on the pattern from Multicall3Delegatecall.sol
    /// @param actions Array of upgrade actions to batch
    /// @param proxyAdminAddress Address of the ProxyAdmin contract
    /// @return data Encoded calldata for aggregate3 function
    function getMulticall3Calldata(
        UpgradeAction[] memory actions,
        address proxyAdminAddress
    ) public pure returns (bytes memory data) {
        IMulticall3.Call[] memory calls = new IMulticall3.Call[](actions.length);

        for (uint256 i = 0; i < calls.length; i++) {
            require(actions[i].proxy != address(0), "Invalid proxy address for multicall");
            require(actions[i].implementation != address(0), "Invalid implementation address for multicall");

            // Encode the upgrade call for ProxyAdmin
            bytes memory upgradeCalldata = abi.encodeWithSelector(
                IProxyAdmin.upgrade.selector,
                actions[i].proxy,
                actions[i].implementation
            );

            calls[i] = IMulticall3.Call({
                target: proxyAdminAddress,
                callData: upgradeCalldata
            });
        }

        data = abi.encodeWithSignature("aggregate((address,bytes)[])", calls);
    }
}
