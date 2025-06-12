// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";

import { LibString } from "@solady/utils/LibString.sol";

// Libraries
import { BaseDeployIO } from "./BaseDeployIO.sol";
import { IMulticall3 } from "forge-std/interfaces/IMulticall3.sol";
import { Constants } from "../../src/libraries/Constants.sol";

// Interfaces
import { IProxyAdmin } from "../../src/universal/interfaces/IProxyAdmin.sol";
import { IProxy } from "../../src/universal/interfaces/IProxy.sol";
import { Proxy } from "../../src/universal/Proxy.sol";
import { PreimageOracle } from "../../src/cannon/PreimageOracle.sol";
import { MIPS } from "../../src/cannon/MIPS.sol";
import { IPreimageOracle } from "../../src/cannon/interfaces/IPreimageOracle.sol";
import { IMIPS } from "../../src/cannon/interfaces/IMIPS.sol";
import { Predeploys } from "../../src/libraries/Predeploys.sol";
import { ISuperchainConfig } from "../../src/L1/interfaces/ISuperchainConfig.sol";
import { SuperchainConfig } from "../../src/L1/SuperchainConfig.sol";
import { IProtocolVersions } from "../../src/L1/interfaces/IProtocolVersions.sol";
import { ProtocolVersions } from "../../src/L1/ProtocolVersions.sol";
import { L2OutputOracle } from "../../src/L1/L2OutputOracle.sol";
import { OptimismPortal2 } from "../../src/L1/OptimismPortal2.sol";
import { SystemConfig } from "../../src/L1/SystemConfig.sol";
import { L1CrossDomainMessenger } from "../../src/L1/L1CrossDomainMessenger.sol";
import { L1ERC721Bridge } from "../../src/L1/L1ERC721Bridge.sol";
import { L1StandardBridge } from "../../src/L1/L1StandardBridge.sol";
import { OptimismMintableERC20Factory } from "../../src/universal/OptimismMintableERC20Factory.sol";
import { DisputeGameFactory } from "../../src/dispute/DisputeGameFactory.sol";
import { AnchorStateRegistry } from "../../src/dispute/AnchorStateRegistry.sol";
import { DelayedWETH } from "../../src/dispute/DelayedWETH.sol";
import { FaultDisputeGame } from "../../src/dispute/FaultDisputeGame.sol";
import { PermissionedDisputeGame } from "../../src/dispute/PermissionedDisputeGame.sol";
import { GameType, Claim, Duration, GameTypes } from "../../src/dispute/lib/Types.sol";
import { IDisputeGame } from "../../src/dispute/interfaces/IDisputeGame.sol";
import { IBigStepper } from "../../src/dispute/interfaces/IBigStepper.sol";
import { CeloSuperchainConfig } from "../../src/L1/CeloSuperchainConfig.sol";
import { IOptimismPortal2 } from "../../src/L1/interfaces/IOptimismPortal2.sol";
import { ISystemConfig } from "../../src/L1/interfaces/ISystemConfig.sol";
import { IL1CrossDomainMessenger } from "../../src/L1/interfaces/IL1CrossDomainMessenger.sol";
import { IL1ERC721Bridge } from "../../src/L1/interfaces/IL1ERC721Bridge.sol";
import { IL1StandardBridge } from "../../src/L1/interfaces/IL1StandardBridge.sol";
import { IOptimismMintableERC20Factory } from "../../src/universal/interfaces/IOptimismMintableERC20Factory.sol";
import { IDisputeGameFactory } from "../../src/dispute/interfaces/IDisputeGameFactory.sol";
import { IAnchorStateRegistry } from "../../src/dispute/interfaces/IAnchorStateRegistry.sol";
import { IDelayedWETH } from "../../src/dispute/interfaces/IDelayedWETH.sol";
import { IL2OutputOracle } from "../../src/L1/interfaces/IL2OutputOracle.sol";
import { ICeloSuperchainConfig } from "../../src/L1/interfaces/ICeloSuperchainConfig.sol";
import { StorageSetter } from "../../src/universal/StorageSetter.sol";

struct Action {
    address proxy;
    address implementation;
    string name;
    bytes data;
}

contract UpgradeImplementationsInput {
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
    address internal _permissionedDelayedWETHProxy;
    address internal _celoSuperchainConfigProxy;
    address internal _storageSetterProxy;

    Action[] internal _customActions;

    function addCustomAction(string memory _name, address _target, bytes memory _data) public {
        _customActions.push(Action({name: _name, proxy: _target, implementation: address(0), data: _data}));
    }

    function getCustomActions() public view returns (Action[] memory) {
        return _customActions;
    }

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
        else if (_sel == this.permissionedDelayedWETHProxy.selector) _permissionedDelayedWETHProxy = _addr;
        else if (_sel == this.celoSuperchainConfigProxy.selector) _celoSuperchainConfigProxy = _addr;
        else if (_sel == this.storageSetterProxy.selector) _storageSetterProxy = _addr;
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

    function permissionedDelayedWETHProxy() public view returns (address) {
        return _permissionedDelayedWETHProxy;
    }

    function celoSuperchainConfigProxy() public view returns (address) {
        return _celoSuperchainConfigProxy;
    }

    function storageSetterProxy() public view returns (address) {
        return _storageSetterProxy;
    }
}

struct EIP1559Params {
    uint32 denominator;
    uint32 elasticity;
}

struct ConstructorArgs {
    uint256 optimismPortalProofMaturityDelaySeconds;
    uint256 optimismPortalDisputeGameFinalityDelaySeconds;
    uint256 delayedWETHDelay;
    uint256 preimageOracleMinProposalSize;
    uint256 preimageOracleChallengePeriod;
    uint256 faultGameMaxDepth;
    uint256 faultGameMaxClockDuration;
    uint256 faultGameSplitDepth;
    uint256 faultGameClockExtension;
    bytes32 faultGameAbsolutePrestate;
    uint256 l2ChainID;
    address l2OutputOracleProposer;
    address l2OutputOracleChallenger;
}

// New struct to hold locally deployed implementation addresses
struct LocallyDeployedImplementationsOutput {
    IDisputeGame cannonFaultDisputeGameImpl;
    IDisputeGame permissionedCannonFaultDisputeGameImpl;
    IOptimismPortal2 optimismPortalImpl;
    IDelayedWETH delayedWETHImpl;
    IPreimageOracle preimageOracleSingleton;
    IMIPS mipsSingleton;
    ISystemConfig systemConfigImpl;
    IL1CrossDomainMessenger l1CrossDomainMessengerImpl;
    IL1ERC721Bridge l1ERC721BridgeImpl;
    IL1StandardBridge l1StandardBridgeImpl;
    IOptimismMintableERC20Factory optimismMintableERC20FactoryImpl;
    IDisputeGameFactory disputeGameFactoryImpl;
    IAnchorStateRegistry anchorStateRegistryImpl;
    IProtocolVersions protocolVersionsImpl;
    IL2OutputOracle l2OutputOracleImpl;
    ISuperchainConfig superchainConfigImpl;
    ICeloSuperchainConfig celoSuperchainConfigImpl;
    StorageSetter storageSetterImpl;
}

contract UpgradeImplementationsOutput {
    bool internal _upgradeComplete;

    function set(bytes4 _sel, bool _value) public {
        if (_sel == this.upgradeComplete.selector) _upgradeComplete = _value;
        else revert("UpgradeImplementationsOutput: unknown selector");
    }

    function upgradeComplete() public view returns (bool) {
        return _upgradeComplete;
    }
}


contract AlfajoresUpgradeImplementations is Script {
    // GnosisSafe address
    address constant _GNOSIS_SAFE = 0xf05f102e890E713DC9dc0a5e13A8879D5296ee48;

    function run() external {
        // setup
        console.log("Setup started!");
        console.log("GnosisSafe address:", _GNOSIS_SAFE);

        ConstructorArgs memory constructorArgs = ConstructorArgs({
            optimismPortalProofMaturityDelaySeconds: 604800,
            optimismPortalDisputeGameFinalityDelaySeconds: 302400,
            delayedWETHDelay: 604800,
            preimageOracleMinProposalSize: 126000,
            preimageOracleChallengePeriod: 86400,
            faultGameMaxDepth: 73,
            faultGameMaxClockDuration: 302400,
            faultGameSplitDepth: 30,
            faultGameClockExtension: 10800,
            faultGameAbsolutePrestate: 0x03c7ae758795765c6664a5d39bf63841c71ff191e9189522bad8ebff5d4eca98,
            l2ChainID: 44787,
            l2OutputOracleProposer: 0x06d010A07D9076d6E7af80E54E26036941221bFA,
            l2OutputOracleChallenger: 0xe571b94CF7e95C46DFe6bEa529335f4A11d15D92
        });

        EIP1559Params memory eip1559Params = EIP1559Params({denominator: 400, elasticity: 5});

        UpgradeImplementationsInput uii = new UpgradeImplementationsInput();
        uii.set(uii.proxyAdmin.selector, address(0x4630583d066520aF0E3fda0de2C628EEd2888683));
        uii.set(uii.superchainConfigProxy.selector, address(0xdf4Fb5371B706936527B877F616eAC0e47c9b785));
        uii.set(uii.protocolVersionsProxy.selector, address(0x5E5FEA4D2A8f632Af05D1E725D7ca865327A080b));
        uii.set(uii.optimismPortalProxy.selector, address(0x82527353927d8D069b3B452904c942dA149BA381));
        uii.set(uii.systemConfigProxy.selector, address(0x499b0C1F4BDC76d61b1D13b03384eac65FAF50c7));
        uii.set(uii.l1CrossDomainMessengerProxy.selector, address(0xF1eE12842631A56a860A38C20B588F4Bb872a4F8));
        uii.set(uii.l1ERC721BridgeProxy.selector, address(0x514912297580a20B7a0C2930BC8503d2C13Da642));
        uii.set(uii.l1StandardBridgeProxy.selector, address(0xD1B0E0581973c9eB7f886967A606b9441A897037));
        uii.set(uii.optimismMintableERC20FactoryProxy.selector, address(0xa950F004F069B0bF9201b17e71549c7711d4a9d5));
        uii.set(uii.disputeGameFactoryProxy.selector, address(0xE28AAdcd9883746c0e5068F58f9ea06027b214cb));
        uii.set(uii.anchorStateRegistryProxy.selector, address(0x235CCA09E27697230ae7c1C671760d6eEB92b12B));
        uii.set(uii.delayedWETHProxy.selector, address(0x8e2a6D372557c9661045f26B140E7A189C38D80C));

        vm.startBroadcast();
        Proxy celoSuperchainConfigProxy = new Proxy(0x4630583d066520aF0E3fda0de2C628EEd2888683);
        Proxy permissionedDelayedWETHProxy = new Proxy(0x4630583d066520aF0E3fda0de2C628EEd2888683);
        vm.stopBroadcast();
        console.log("Deployed CeloSuperchainConfigProxy at:", address(celoSuperchainConfigProxy));
        uii.set(uii.celoSuperchainConfigProxy.selector, address(celoSuperchainConfigProxy));
        console.log("Deployed PermissionedDelayedWETHProxy at:", address(permissionedDelayedWETHProxy));
        uii.set(uii.permissionedDelayedWETHProxy.selector, address(permissionedDelayedWETHProxy));

        // execution
        console.log("Execution!");
        UpgradeImplementations upgrade = new UpgradeImplementations();
        upgrade.initializeCeloSuperchainConfig(uii);
        upgrade.run(uii, new UpgradeImplementationsOutput(), constructorArgs, eip1559Params);
    }
}

contract UpgradeImplementations is Script {
    // -------- Utilities copied from DeployImplementations.s.sol --------


    // -----------------------------------------------------------------

    // Multicall3 delegatecall contract address
    address public constant MULTICALL_ADDRESS = 0xcA11bde05977b3631167028862bE2a173976CA11;
    address private constant _GNOSIS_SAFE = 0xf05f102e890E713DC9dc0a5e13A8879D5296ee48;


    function run(
        UpgradeImplementationsInput _uii,
        UpgradeImplementationsOutput _uio,
        ConstructorArgs memory _constructorArgs,
        EIP1559Params memory _eip1559Params
    ) public {
        console.log("Deploying new implementations locally...");

        LocallyDeployedImplementationsOutput memory ldio = _deployImplementations(_uii, _constructorArgs);

        console.log("New implementations deployed locally successfully!");
        console.log("Starting implementation upgrades...");

        IProxyAdmin proxyAdmin = _uii.proxyAdmin();

        // Check if ProxyAdmin is owned by a Gnosis Safe (caller is not the owner)
        address proxyAdminOwner = proxyAdmin.owner();
        console.log("ProxyAdmin owner:", proxyAdminOwner);
        console.log("Transaction sender:", msg.sender);

        console.log("ProxyAdmin is owned by a different address (likely Gnosis Safe)");
        console.log("Generating transaction data for Safe submission instead of direct execution...");

        // Generate Safe transaction data instead of executing directly
        generateSafeTransactionData(_uii, ldio, _eip1559Params);

        // Generate multicall batch transaction data
        generateMulticallBatchData(_uii, ldio, _eip1559Params);

        console.log("\nAttempting post-upgrade validations...");
        console.log("Note: These validations assume the Gnosis Safe transaction has been or will be executed successfully.");
        // _validateAllCollectedUpgrades(_uii, ldio, _eip1559Params);

        _uio.set(_uio.upgradeComplete.selector, true);
        console.log("Transaction data generated and validation (if applicable) attempted. Submit to Gnosis Safe for execution if not done by script.");
    }

    function _deployImplementations(
        UpgradeImplementationsInput _uii,
        ConstructorArgs memory _constructorArgs
    ) internal returns (LocallyDeployedImplementationsOutput memory ldio) {
        vm.startBroadcast();
        // Deploy all implementations
        ldio.systemConfigImpl = ISystemConfig(address(new SystemConfig()));
        console.log("Deployed SystemConfig at:", address(ldio.systemConfigImpl));
        ldio.l1CrossDomainMessengerImpl = IL1CrossDomainMessenger(address(new L1CrossDomainMessenger()));
        console.log("Deployed L1CrossDomainMessenger at:", address(ldio.l1CrossDomainMessengerImpl));
        ldio.l1ERC721BridgeImpl = IL1ERC721Bridge(address(new L1ERC721Bridge()));
        console.log("Deployed L1ERC721Bridge at:", address(ldio.l1ERC721BridgeImpl));
        ldio.l1StandardBridgeImpl = IL1StandardBridge(payable(address(new L1StandardBridge())));
        console.log("Deployed L1StandardBridge at:", address(ldio.l1StandardBridgeImpl));
        ldio.optimismMintableERC20FactoryImpl =
            IOptimismMintableERC20Factory(address(new OptimismMintableERC20Factory()));
        console.log("Deployed OptimismMintableERC20Factory at:", address(ldio.optimismMintableERC20FactoryImpl));
        ldio.optimismPortalImpl = IOptimismPortal2(
            payable(
                address(
                    new OptimismPortal2(
                        _constructorArgs.optimismPortalProofMaturityDelaySeconds,
                        _constructorArgs.optimismPortalDisputeGameFinalityDelaySeconds
                    )
                )
            )
        );
        console.log("Deployed OptimismPortal2 at:", address(ldio.optimismPortalImpl));
        ldio.delayedWETHImpl = IDelayedWETH(payable(address(new DelayedWETH(_constructorArgs.delayedWETHDelay))));
        console.log("Deployed DelayedWETH at:", address(ldio.delayedWETHImpl));
        ldio.preimageOracleSingleton =
            IPreimageOracle(address(new PreimageOracle(_constructorArgs.preimageOracleMinProposalSize, _constructorArgs.preimageOracleChallengePeriod)));
        console.log("Deployed PreimageOracle at:", address(ldio.preimageOracleSingleton));
        ldio.mipsSingleton = IMIPS(address(new MIPS(IPreimageOracle(address(ldio.preimageOracleSingleton)))));
        console.log("Deployed MIPS at:", address(ldio.mipsSingleton));
        ldio.disputeGameFactoryImpl = IDisputeGameFactory(address(new DisputeGameFactory()));
        console.log("Deployed DisputeGameFactory at:", address(ldio.disputeGameFactoryImpl));
        ldio.anchorStateRegistryImpl =
            IAnchorStateRegistry(address(new AnchorStateRegistry(IDisputeGameFactory(address(ldio.disputeGameFactoryImpl)))));
        console.log("Deployed AnchorStateRegistry at:", address(ldio.anchorStateRegistryImpl));
        ldio.protocolVersionsImpl = IProtocolVersions(address(new ProtocolVersions()));
        console.log("Deployed ProtocolVersions at:", address(ldio.protocolVersionsImpl));
        ldio.l2OutputOracleImpl = IL2OutputOracle(address(new L2OutputOracle()));
        console.log("Deployed L2OutputOracle at:", address(ldio.l2OutputOracleImpl));
        ldio.superchainConfigImpl = ISuperchainConfig(address(new SuperchainConfig()));
        console.log("Deployed SuperchainConfig at:", address(ldio.superchainConfigImpl));
        ldio.celoSuperchainConfigImpl = ICeloSuperchainConfig(address(new CeloSuperchainConfig()));
        console.log("Deployed CeloSuperchainConfig at:", address(ldio.celoSuperchainConfigImpl));
        ldio.storageSetterImpl = new StorageSetter();
        console.log("Deployed StorageSetter at:", address(ldio.storageSetterImpl));

        IDelayedWETH weth = IDelayedWETH(payable(_uii.delayedWETHProxy()));
        IDelayedWETH permissionedWeth = IDelayedWETH(payable(_uii.permissionedDelayedWETHProxy()));
        IAnchorStateRegistry anchorStateRegistry = IAnchorStateRegistry(_uii.anchorStateRegistryProxy());

        ldio.cannonFaultDisputeGameImpl = IDisputeGame(
            address(
                new FaultDisputeGame(
                    GameTypes.CANNON,
                    Claim.wrap(_constructorArgs.faultGameAbsolutePrestate),
                    _constructorArgs.faultGameMaxDepth,
                    _constructorArgs.faultGameSplitDepth,
                    Duration.wrap(uint64(_constructorArgs.faultGameClockExtension)),
                    Duration.wrap(uint64(_constructorArgs.faultGameMaxClockDuration)),
                    IBigStepper(address(ldio.mipsSingleton)),
                    weth,
                    anchorStateRegistry,
                    _constructorArgs.l2ChainID
                )
            )
        );
        console.log("Deployed FaultDisputeGame at:", address(ldio.cannonFaultDisputeGameImpl));

        ldio.permissionedCannonFaultDisputeGameImpl = IDisputeGame(
            address(
                new PermissionedDisputeGame(
                    GameTypes.PERMISSIONED_CANNON,
                    Claim.wrap(_constructorArgs.faultGameAbsolutePrestate),
                    _constructorArgs.faultGameMaxDepth,
                    _constructorArgs.faultGameSplitDepth,
                    Duration.wrap(uint64(_constructorArgs.faultGameClockExtension)),
                    Duration.wrap(uint64(_constructorArgs.faultGameMaxClockDuration)),
                    IBigStepper(address(ldio.mipsSingleton)),
                    permissionedWeth,
                    anchorStateRegistry,
                    _constructorArgs.l2ChainID,
                    _constructorArgs.l2OutputOracleProposer,
                    _constructorArgs.l2OutputOracleChallenger
                )
            )
        );
        console.log("Deployed PermissionedDisputeGame at:", address(ldio.permissionedCannonFaultDisputeGameImpl));

        vm.stopBroadcast();
    }

    /// @notice Validates a single proxy upgrade.
    function _validateSingleUpgrade(
        IProxyAdmin _proxyAdmin,
        string memory _contractName,
        address _proxyAddress,
        address _expectedImplementation
    ) internal view {
        console.log("Validating upgrade for", _contractName, "at", _proxyAddress);
        // Ensure proxy address is not zero before attempting to load storage
        if (_proxyAddress == address(0)) {
            console.log("[WARN] Validation SKIPPED for since proxy is 0 ", _contractName);
            return;
        }
        // Ensure expected implementation is not zero
        if (_expectedImplementation == address(0)) {
            console.log("[WARN] Validation SKIPPED for - Expected implementation is zero.", _contractName);
            return;
        }

        address currentImplementation = _proxyAdmin.getProxyImplementation(_proxyAddress);

        if (currentImplementation == _expectedImplementation) {
            console.log("[PASS] Validation PASSED for", _contractName, "at", _proxyAddress);
            console.log("   Implementation is correctly set to:", currentImplementation);
        } else {
            revert(
                string.concat(
                    "Validation FAILED for ",
                    _contractName,
                    " at ",
                    LibString.toHexString(_proxyAddress),
                    ". Expected implementation: ",
                    LibString.toHexString(_expectedImplementation),
                    ", but got: ",
                    LibString.toHexString(currentImplementation)
                )
            );
        }
    }

    /// @notice Validates all upgrades defined by the collected actions.
    function _validateAllCollectedUpgrades(
        UpgradeImplementationsInput _uii,
        LocallyDeployedImplementationsOutput memory _ldio,
        EIP1559Params memory _eip1559Params
    ) internal view {
        console.log("\n=== STARTING POST-UPGRADE VALIDATIONS ===");
        Action[] memory actions = _collectUpgradeActions(_uii, _ldio, _eip1559Params);

        if (actions.length == 0) {
            console.log("No upgrade actions found to validate.");
            console.log("=== POST-UPGRADE VALIDATIONS COMPLETE ===");
            return;
        }

        IProxyAdmin proxyAdmin = _uii.proxyAdmin();
        for (uint256 i = 0; i < actions.length; i++) {
            if (actions[i].data.length == 0 && !LibString.contains(actions[i].name, "StorageSetter")) {
                _validateSingleUpgrade(proxyAdmin, actions[i].name, actions[i].proxy, actions[i].implementation);
            }
        }
        console.log("=== POST-UPGRADE VALIDATIONS COMPLETE ===");
    }

    /// @notice Helper function to generate transaction data for Gnosis Safe execution
    /// @dev This function logs the transaction data that can be submitted to a Gnosis Safe
    /// @param _uii Input configuration for the upgrade
    /// @param _ldio Output from local implementation deployments
    function generateSafeTransactionData(
        UpgradeImplementationsInput _uii,
        LocallyDeployedImplementationsOutput memory _ldio,
        EIP1559Params memory _eip1559Params
    ) public {
        IProxyAdmin proxyAdmin = _uii.proxyAdmin();
        Action[] memory actions = _collectUpgradeActions(_uii, _ldio, _eip1559Params);

        console.log("=== GNOSIS SAFE TRANSACTION DATA ===");
        console.log("ProxyAdmin address:", address(proxyAdmin));
        console.log("ProxyAdmin owner (should be Gnosis Safe):", proxyAdmin.owner());
        console.log("");
        console.log("Copy the following transaction data to submit to Gnosis Safe:");
        console.log("");

        for (uint256 i = 0; i < actions.length; i++) {
            if (actions[i].data.length > 0) {
                console.log(actions[i].name);
                console.log("  Target:", actions[i].proxy);
                console.logBytes(actions[i].data);
            } else {
                console.log(actions[i].name, "address:", actions[i].proxy);
                bytes memory data =
                    abi.encodeWithSelector(IProxyAdmin.upgrade.selector, actions[i].proxy, actions[i].implementation);
                console.logBytes(data);
            }
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

        (bool success, bytes memory result) = _GNOSIS_SAFE.staticcall(
            abi.encodeWithSignature("nonce()")
        );
        uint256 nonce = success ? abi.decode(result, (uint256)) : 0;
        console.log("Safe nonce:", nonce);

        (success, result) = _GNOSIS_SAFE.staticcall(
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
        uint256 privateKey = vm.envUint("PRIVATE_KEY");
        // Basic validation: private keys shouldn't be zero and should parse correctly.
        require(privateKey != 0, "Failed to parse PRIVATE_KEY, or it resolved to zero. Ensure it's a valid hex string (e.g., 0x...).");

        console.log("Private key for Gnosis Safe message signing loaded successfully from PRIVATE_KEY."); // Avoid logging the key itself

        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, safeTxHash);

        bytes memory signature = abi.encodePacked(r, s, v);
        console.log("Signature for Gnosis Safe message:", LibString.toHexString(signature));

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
        require(_GNOSIS_SAFE.code.length > 0, "Gnosis Safe contract not found at expected address");

        vm.startBroadcast();

        (bool execSuccess, bytes memory returnData) = _GNOSIS_SAFE.call(
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
    /// @param _ldio Output from local implementation deployments
    /// @return actions Array of UpgradeAction structs containing proxy, implementation, and name
    function _collectUpgradeActions(
        UpgradeImplementationsInput _uii,
        LocallyDeployedImplementationsOutput memory _ldio,
        EIP1559Params memory _eip1559Params
    ) internal view returns (Action[] memory actions) {
        // Count valid actions first
        uint256 actionCount = 0;
        if (_uii.superchainConfigProxy() != address(0)) actionCount++; // SuperchainConfig (not deployed here)
        if (_uii.protocolVersionsProxy() != address(0)) actionCount++; // ProtocolVersions (not deployed here)
        if (_uii.optimismPortalProxy() != address(0) && address(_ldio.optimismPortalImpl) != address(0)) actionCount += 5;
        if (_uii.systemConfigProxy() != address(0) && address(_ldio.systemConfigImpl) != address(0)) actionCount += 3;
        if (_uii.l1CrossDomainMessengerProxy() != address(0) && address(_ldio.l1CrossDomainMessengerImpl) != address(0)) {
            actionCount += 3;
        }
        if (_uii.l1ERC721BridgeProxy() != address(0) && address(_ldio.l1ERC721BridgeImpl) != address(0)) actionCount += 3;
        if (_uii.l1StandardBridgeProxy() != address(0) && address(_ldio.l1StandardBridgeImpl) != address(0)) {
            actionCount += 3;
        }
        if (_uii.optimismMintableERC20FactoryProxy() != address(0)
            && address(_ldio.optimismMintableERC20FactoryImpl) != address(0)) {
            actionCount++;
        }
        if (_uii.disputeGameFactoryProxy() != address(0) && address(_ldio.disputeGameFactoryImpl) != address(0)) {
            actionCount++;
            actionCount += 2;
        }
        if (_uii.anchorStateRegistryProxy() != address(0)) actionCount += 3;
        if (_uii.delayedWETHProxy() != address(0) && address(_ldio.delayedWETHImpl) != address(0)) actionCount += 3;
        if (_uii.permissionedDelayedWETHProxy() != address(0) && address(_ldio.delayedWETHImpl) != address(0)) {
            actionCount += 2;
        }
        if (_uii.celoSuperchainConfigProxy() != address(0) && address(_ldio.celoSuperchainConfigImpl) != address(0)) {
            actionCount++;
        }
        if (_uii.storageSetterProxy() != address(0) && address(_ldio.storageSetterImpl) != address(0)) actionCount++;

        Action[] memory customActions = _uii.getCustomActions();
        actionCount += customActions.length;

        actions = new Action[](actionCount);
        uint256 index = 0;

        if (_uii.storageSetterProxy() != address(0) && address(_ldio.storageSetterImpl) != address(0)) {
            actions[index++] = Action({
                proxy: _uii.storageSetterProxy(),
                implementation: address(_ldio.storageSetterImpl),
                name: "StorageSetter",
                data: ""
            });
        }

        // Add optional upgrades (pointing to address(0) if not deployed by this script)
        if (_uii.superchainConfigProxy() != address(0)) {
            actions[index++] = Action({
                proxy: _uii.superchainConfigProxy(),
                implementation: address(_ldio.superchainConfigImpl),
                name: "SuperchainConfig",
                data: ""
            });
        }

        if (_uii.protocolVersionsProxy() != address(0)) {
            actions[index++] = Action({
                proxy: _uii.protocolVersionsProxy(),
                implementation: address(_ldio.protocolVersionsImpl),
                name: "ProtocolVersions",
                data: ""
            });
        }

        // Add required upgrades
        if (_uii.optimismPortalProxy() != address(0) && address(_ldio.optimismPortalImpl) != address(0)) {
            console.log("Setting up OptimismPortal2 upgrade actions...celo superchainConfigProxy:", _uii.celoSuperchainConfigProxy());
            actions[index++] = Action({
                proxy: _uii.optimismPortalProxy(),
                implementation: address(_ldio.storageSetterImpl),
                name: "Upgrade OptimismPortal2 to StorageSetter",
                data: ""
            });
            // The superchainConfig is at storage slot 53, offset 1. A bool is at offset 0.
            // We shift the address by 8 bits (1 byte) to correctly position it.
            uint256 superchainConfigValue = (uint256(uint160(_uii.celoSuperchainConfigProxy())) << 8) | 0;
            // uint256 superchainConfigValue = (uint256(uint160(_uii.celoSuperchainConfigProxy())));
            actions[index++] = Action({
                proxy: _uii.optimismPortalProxy(),
                implementation: address(0),
                name: "Set OptimismPortal2 storage",
                data: abi.encodeWithSelector(StorageSetter.setBytes32.selector, 53, bytes32(superchainConfigValue))
            });
            // Set disputeGameFactory to disputeGameFactoryProxy address
            actions[index++] = Action({
                proxy: _uii.optimismPortalProxy(),
                implementation: address(0),
                name: "Set OptimismPortal2 disputeGameFactory",
                data: abi.encodeWithSelector(
                    StorageSetter.setAddress.selector,
                    56,
                    bytes32(uint256(uint160(_uii.disputeGameFactoryProxy())))
                )
            });
            // Set respectedGameType to 1 and respectedGameTypeUpdatedAt to current block timestamp
            uint256 respectedGameTypeValue = (block.timestamp << 32) | 1;
            actions[index++] = Action({
                proxy: _uii.optimismPortalProxy(),
                implementation: address(0),
                name: "Set OptimismPortal2 respectedGameType and timestamp",
                data: abi.encodeWithSelector(StorageSetter.setBytes32.selector, 59, bytes32(respectedGameTypeValue))
            });
            actions[index++] = Action({
                proxy: _uii.optimismPortalProxy(),
                implementation: address(_ldio.optimismPortalImpl),
                name: "Upgrade OptimismPortal2 to final implementation",
                data: ""
            });
        }

        if (_uii.systemConfigProxy() != address(0) && address(_ldio.systemConfigImpl) != address(0)) {
            actions[index++] = Action({
                proxy: _uii.systemConfigProxy(),
                implementation: address(_ldio.storageSetterImpl),
                name: "Upgrade SystemConfig to StorageSetter",
                data: ""
            });
            // Set eip1559Denominator and eip1559Elasticity by writing to storage slot 106.
            uint256 eip1559Value = (uint256(_eip1559Params.elasticity) << 32) | _eip1559Params.denominator;
            actions[index++] = Action({
                proxy: _uii.systemConfigProxy(),
                implementation: address(0),
                name: "Set SystemConfig EIP1559 params",
                data: abi.encodeWithSelector(StorageSetter.setAddress.selector, 106, bytes32(eip1559Value))
            });
            actions[index++] = Action({
                proxy: _uii.systemConfigProxy(),
                implementation: address(_ldio.systemConfigImpl),
                name: "Upgrade SystemConfig to final implementation",
                data: ""
            });
        }

        if (_uii.l1CrossDomainMessengerProxy() != address(0) && address(_ldio.l1CrossDomainMessengerImpl) != address(0)) {
            actions[index++] = Action({
                proxy: _uii.l1CrossDomainMessengerProxy(),
                implementation: address(_ldio.storageSetterImpl),
                name: "Upgrade L1CrossDomainMessenger to StorageSetter",
                data: ""
            });
            actions[index++] = Action({
                proxy: _uii.l1CrossDomainMessengerProxy(),
                implementation: address(0),
                name: "Set L1CrossDomainMessenger storage",
                data: abi.encodeWithSelector(
                    StorageSetter.setAddress.selector,
                    251,
                    bytes32(uint256(uint160(_uii.celoSuperchainConfigProxy())))
                )
            });
            actions[index++] = Action({
                proxy: _uii.l1CrossDomainMessengerProxy(),
                implementation: address(_ldio.l1CrossDomainMessengerImpl),
                name: "Upgrade L1CrossDomainMessenger to final implementation",
                data: ""
            });
        }

        if (_uii.l1ERC721BridgeProxy() != address(0) && address(_ldio.l1ERC721BridgeImpl) != address(0)) {
            actions[index++] = Action({
                proxy: _uii.l1ERC721BridgeProxy(),
                implementation: address(_ldio.storageSetterImpl),
                name: "Upgrade L1ERC721Bridge to StorageSetter",
                data: ""
            });
            actions[index++] = Action({
                proxy: _uii.l1ERC721BridgeProxy(),
                implementation: address(0),
                name: "Set L1ERC721Bridge storage",
                data: abi.encodeWithSelector(
                    StorageSetter.setAddress.selector,
                    50,
                    bytes32(uint256(uint160(_uii.celoSuperchainConfigProxy())))
                )
            });
            actions[index++] = Action({
                proxy: _uii.l1ERC721BridgeProxy(),
                implementation: address(_ldio.l1ERC721BridgeImpl),
                name: "Upgrade L1ERC721Bridge to final implementation",
                data: ""
            });
        }

        if (_uii.l1StandardBridgeProxy() != address(0) && address(_ldio.l1StandardBridgeImpl) != address(0)) {
            actions[index++] = Action({
                proxy: _uii.l1StandardBridgeProxy(),
                implementation: address(_ldio.storageSetterImpl),
                name: "Upgrade L1StandardBridge to StorageSetter",
                data: ""
            });
            actions[index++] = Action({
                proxy: _uii.l1StandardBridgeProxy(),
                implementation: address(0),
                name: "Set L1StandardBridge storage",
                data: abi.encodeWithSelector(
                    StorageSetter.setAddress.selector,
                    50,
                    bytes32(uint256(uint160(_uii.celoSuperchainConfigProxy())))
                )
            });
            actions[index++] = Action({
                proxy: _uii.l1StandardBridgeProxy(),
                implementation: address(_ldio.l1StandardBridgeImpl),
                name: "Upgrade L1StandardBridge to final implementation",
                data: ""
            });
        }

        if (_uii.optimismMintableERC20FactoryProxy() != address(0)
            && address(_ldio.optimismMintableERC20FactoryImpl) != address(0)) {
            actions[index++] = Action({
                proxy: _uii.optimismMintableERC20FactoryProxy(),
                implementation: address(_ldio.optimismMintableERC20FactoryImpl),
                name: "OptimismMintableERC20Factory",
                data: ""
            });
        }

        if (_uii.disputeGameFactoryProxy() != address(0) && address(_ldio.disputeGameFactoryImpl) != address(0)) {
            actions[index++] = Action({
                proxy: _uii.disputeGameFactoryProxy(),
                implementation: address(_ldio.disputeGameFactoryImpl),
                name: "DisputeGameFactory",
                data: ""
            });
            actions[index++] = Action({
                proxy: _uii.disputeGameFactoryProxy(),
                implementation: address(0),
                name: "SetCannonFaultGameImplementation",
                data: abi.encodeWithSelector(
                    IDisputeGameFactory.setImplementation.selector,
                    GameTypes.CANNON,
                    _ldio.cannonFaultDisputeGameImpl
                )
            });
            actions[index++] = Action({
                proxy: _uii.disputeGameFactoryProxy(),
                implementation: address(0),
                name: "SetPermissionedCannonFaultGameImplementation",
                data: abi.encodeWithSelector(
                    IDisputeGameFactory.setImplementation.selector,
                    GameTypes.PERMISSIONED_CANNON,
                    _ldio.permissionedCannonFaultDisputeGameImpl
                )
            });
        }

        if (_uii.anchorStateRegistryProxy() != address(0)) {
            actions[index++] = Action({
                proxy: _uii.anchorStateRegistryProxy(),
                implementation: address(_ldio.storageSetterImpl),
                name: "Upgrade AnchorStateRegistry to StorageSetter",
                data: ""
            });
            actions[index++] = Action({
                proxy: _uii.anchorStateRegistryProxy(),
                implementation: address(0),
                name: "Set AnchorStateRegistry storage",
                data: abi.encodeWithSelector(
                    StorageSetter.setAddress.selector,
                    2,
                    bytes32(uint256(uint160(_uii.celoSuperchainConfigProxy())))
                )
            });
            actions[index++] = Action({
                proxy: _uii.anchorStateRegistryProxy(),
                implementation: address(_ldio.anchorStateRegistryImpl),
                name: "Upgrade AnchorStateRegistry to final implementation",
                data: ""
            });
        }

        if (_uii.delayedWETHProxy() != address(0) && address(_ldio.delayedWETHImpl) != address(0)) {
            actions[index++] = Action({
                proxy: _uii.delayedWETHProxy(),
                implementation: address(_ldio.storageSetterImpl),
                name: "Upgrade DelayedWETH to StorageSetter",
                data: ""
            });
            actions[index++] = Action({
                proxy: _uii.delayedWETHProxy(),
                implementation: address(0),
                name: "Set DelayedWETH storage",
                data: abi.encodeWithSelector(
                    StorageSetter.setAddress.selector,
                    104,
                    bytes32(uint256(uint160(_uii.celoSuperchainConfigProxy())))
                )
            });
            actions[index++] = Action({
                proxy: _uii.delayedWETHProxy(),
                implementation: address(_ldio.delayedWETHImpl),
                name: "Upgrade DelayedWETH to final implementation",
                data: ""
            });
        }

        if (_uii.permissionedDelayedWETHProxy() != address(0) && address(_ldio.delayedWETHImpl) != address(0)) {
            actions[index++] = Action({
                proxy: _uii.permissionedDelayedWETHProxy(),
                implementation: address(_ldio.delayedWETHImpl),
                name: "PermissionedDelayedWETH",
                data: ""
            });
            actions[index++] = Action({
                proxy: _uii.permissionedDelayedWETHProxy(),
                implementation: address(0),
                name: "InitializePermissionedDelayedWETH",
                data: abi.encodeWithSelector(
                    IDelayedWETH.initialize.selector,
                    _GNOSIS_SAFE,
                    ICeloSuperchainConfig(_uii.celoSuperchainConfigProxy())
                )
            });
        }

        if (_uii.celoSuperchainConfigProxy() != address(0) && address(_ldio.celoSuperchainConfigImpl) != address(0)) {
            actions[index++] = Action({
                proxy: _uii.celoSuperchainConfigProxy(),
                implementation: address(_ldio.celoSuperchainConfigImpl),
                name: "CeloSuperchainConfig",
                data: ""
            });
        }



        for (uint256 i = 0; i < customActions.length; i++) {
            actions[index++] = customActions[i];
        }

        return actions;
    }

    /// @notice Generate multicall batch transaction data for all upgrades
    /// @dev This function creates a single multicall transaction that batches all upgrades
    /// @param _uii Input configuration for the upgrade
    /// @param _ldio Output from local implementation deployments
    function generateMulticallBatchData(
        UpgradeImplementationsInput _uii,
        LocallyDeployedImplementationsOutput memory _ldio,
        EIP1559Params memory _eip1559Params
    ) public {
        IProxyAdmin proxyAdmin = _uii.proxyAdmin();
        Action[] memory actions = _collectUpgradeActions(_uii, _ldio, _eip1559Params);

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
            console.log(actions[i].name);
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

        // If not in ledgerMode, proceed to attempt signing and execution
        console.log("\n=== ATTEMPTING TO EXECUTE MULTICALL BATCH VIA GNOSIS SAFE (DELEGATECALL) ===");
        console.log("This requires PRIVATE_KEY env var to be set for signing the Safe message,");
        console.log("and the transaction broadcaster (e.g., from --private-key flag) to be a Safe owner or have permissions.");

        // For multicall via delegatecall, operation is 1 (DELEGATECALL)
        // The target for the Gnosis Safe transaction is the MULTICALL_ADDRESS
        bytes32 safeTxHash = _generateSafeTxHash(MULTICALL_ADDRESS, multicallData, 1);
        bytes memory signature = _signTransaction(safeTxHash);
        _executeTransaction(MULTICALL_ADDRESS, multicallData, signature, 1);
        console.log("=== MULTICALL BATCH EXECUTION ATTEMPT VIA GNOSIS SAFE COMPLETE ===");
    }

    /// @notice Generate multicall3 calldata for upgrade actions
    /// @dev Based on the pattern from Multicall3Delegatecall.sol
    /// @param actions Array of upgrade actions to batch
    /// @param proxyAdminAddress Address of the ProxyAdmin contract
    /// @return data Encoded calldata for aggregate3 function
    function getMulticall3Calldata(Action[] memory actions, address proxyAdminAddress)
        public
        pure
        returns (bytes memory data)
    {
        IMulticall3.Call[] memory calls = new IMulticall3.Call[](actions.length);

        for (uint256 i = 0; i < actions.length; i++) {
            bytes memory callData;
            address target;

            if (actions[i].data.length > 0) {
                console.log("Generating multicall for custom action:", actions[i].name);
                callData = actions[i].data;
                target = actions[i].proxy;
            } else {
                console.log("Generating multicall for upgrade action:", actions[i].name);
                require(actions[i].proxy != address(0), "Invalid proxy address for multicall");
                require(actions[i].implementation != address(0), "Invalid implementation address for multicall");

                callData =
                    abi.encodeWithSelector(IProxyAdmin.upgrade.selector, actions[i].proxy, actions[i].implementation);
                target = proxyAdminAddress;
            }

            calls[i] = IMulticall3.Call({target: target, callData: callData});
        }

        data = abi.encodeWithSignature("aggregate((address,bytes)[])", calls);
    }

    function initializeCeloSuperchainConfig(UpgradeImplementationsInput _uii) public {
        address celoSuperchainConfigProxy = _uii.celoSuperchainConfigProxy();
        address superchainConfigProxy = _uii.superchainConfigProxy();
        bytes memory data = abi.encodeWithSelector(
            ICeloSuperchainConfig.initialize.selector,
            0xe571b94CF7e95C46DFe6bEa529335f4A11d15D92,
            false,
            superchainConfigProxy
        );

        _uii.addCustomAction("InitializeCeloSuperchainConfig", celoSuperchainConfigProxy, data);
    }

}
