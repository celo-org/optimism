
pragma solidity ^0.8.15;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";
import { LibString } from "../../lib/solady/src/utils/LibString.sol";
import { Constants } from "../../src/libraries/Constants.sol";
import { IProxyAdmin } from "../../interfaces/universal/IProxyAdmin.sol";
import { ISystemConfig } from "../../interfaces/L1/ISystemConfig.sol";
import { IOptimismPortal2 } from "../../interfaces/L1/IOptimismPortal2.sol";
import { IL1CrossDomainMessenger } from "../../interfaces/L1/IL1CrossDomainMessenger.sol";
import { IL1ERC721Bridge } from "../../interfaces/L1/IL1ERC721Bridge.sol";
import { IL1StandardBridge } from "../../interfaces/L1/IL1StandardBridge.sol";
import { CeloSuperchainConfig } from "../../src/celo/CeloSuperchainConfig.sol";
import { UpgradeImplementationsInput } from "./UpgradeImplementations.s.sol";
import { DeployImplementationsOutput, DeployImplementationsInput } from "./DeployImplementations.s.sol";

contract ValidateImplementations is Script {
    function run(
        UpgradeImplementationsInput _uii,
        DeployImplementationsOutput _dio,
        address _expectedCustomTokenAddr,
        string memory _expectedTokenName,
        string memory _expectedTokenSymbol
    ) public view {
        console.log("\nValidating custom gas token OptimismPortal...");
        _validateCustomGasTokenOptimismPortal(_uii, _dio, _expectedCustomTokenAddr);
        console.log("\nValidating custom gas token SystemConfig...");
        _validateCustomGasTokenSystemConfig(_uii, _dio, _expectedCustomTokenAddr, _expectedTokenName, _expectedTokenSymbol);
        console.log("\nValidating Celo SuperchainConfig references...");
        _validateCeloSuperchainConfigReferences(_uii);
    }

    function validateBaklava() public {
        console.log("Setting up Baklava validation parameters...");

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

        DeployImplementationsOutput dio = new DeployImplementationsOutput();
        IProxyAdmin proxyAdmin = uii.proxyAdmin();


        if (uii.superchainConfigProxy() != address(0)) {
            dio.set(dio.superchainConfigImpl.selector, proxyAdmin.getProxyImplementation(uii.superchainConfigProxy()));
        }
        if (uii.protocolVersionsProxy() != address(0)) {
            dio.set(dio.protocolVersionsImpl.selector, proxyAdmin.getProxyImplementation(uii.protocolVersionsProxy()));
        }
        dio.set(dio.optimismPortalImpl.selector, proxyAdmin.getProxyImplementation(uii.optimismPortalProxy()));
        dio.set(dio.systemConfigImpl.selector, proxyAdmin.getProxyImplementation(uii.systemConfigProxy()));
        dio.set(dio.l1CrossDomainMessengerImpl.selector, proxyAdmin.getProxyImplementation(uii.l1CrossDomainMessengerProxy()));
        dio.set(dio.l1ERC721BridgeImpl.selector, proxyAdmin.getProxyImplementation(uii.l1ERC721BridgeProxy()));
        dio.set(dio.l1StandardBridgeImpl.selector, proxyAdmin.getProxyImplementation(uii.l1StandardBridgeProxy()));
        dio.set(dio.optimismMintableERC20FactoryImpl.selector, proxyAdmin.getProxyImplementation(uii.optimismMintableERC20FactoryProxy()));
        dio.set(dio.disputeGameFactoryImpl.selector, proxyAdmin.getProxyImplementation(uii.disputeGameFactoryProxy()));
        dio.set(dio.anchorStateRegistryImpl.selector, proxyAdmin.getProxyImplementation(uii.anchorStateRegistryProxy()));
        if (uii.delayedWETHProxy() != address(0)) {
            dio.set(dio.delayedWETHImpl.selector, proxyAdmin.getProxyImplementation(uii.delayedWETHProxy()));
        }

        address baklavaExpectedCustomTokenAddr = 0xE692fD8305e097b0e73f1b61aCA8b74Cd921443B;
        string memory baklavaExpectedTokenName = "Celo native asset";
        string memory baklavaExpectedTokenSymbol = "CELO";

        console.log("Running Baklava validations...");
        run(uii, dio, baklavaExpectedCustomTokenAddr, baklavaExpectedTokenName, baklavaExpectedTokenSymbol);
        console.log("Baklava validations complete.");
    }

    function _validateCeloSuperchainConfigReferences(UpgradeImplementationsInput _uii) internal view {
        address globalSuperchainConfigProxy = _uii.superchainConfigProxy();
        require(globalSuperchainConfigProxy != address(0), "Global SuperchainConfig proxy is 0");

        _validateCeloScSlot("OptimismPortal", address(IOptimismPortal2(payable(_uii.optimismPortalProxy())).superchainConfig()), globalSuperchainConfigProxy);
        _validateCeloScSlot("L1CrossDomainMessenger", address(IL1CrossDomainMessenger(_uii.l1CrossDomainMessengerProxy()).superchainConfig()), globalSuperchainConfigProxy);
        _validateCeloScSlot("L1StandardBridge", address(IL1StandardBridge(payable(_uii.l1StandardBridgeProxy())).superchainConfig()), globalSuperchainConfigProxy);
        _validateCeloScSlot("L1ERC721Bridge", address(IL1ERC721Bridge(_uii.l1ERC721BridgeProxy()).superchainConfig()), globalSuperchainConfigProxy);
    }

    function _validateCeloScSlot(string memory _contractName, address _celoScProxyAddr, address _expectedGlobalScAddr) internal view {
        if (_celoScProxyAddr == address(0)) {
            revert(string.concat(_contractName, ": L1 contract's superchainConfig() is address(0)"));
        }
        address globalSuperchainConfigAddrFromCeloSCGetter = CeloSuperchainConfig(_celoScProxyAddr).superchainConfig();
        if (globalSuperchainConfigAddrFromCeloSCGetter == _expectedGlobalScAddr) {
            console.log("[PASS]", _contractName, "references global SC via CeloSC");
        } else {
            revert(string.concat(_contractName, ": CeloSC's superchainConfig() mismatch. L1SC:", LibString.toHexStringChecksummed(_celoScProxyAddr), ", CeloSC.SC():", LibString.toHexStringChecksummed(globalSuperchainConfigAddrFromCeloSCGetter), ", ExpectedGlobalSC:", LibString.toHexStringChecksummed(_expectedGlobalScAddr)));
        }
    }

    function _validateCustomGasTokenOptimismPortal(
        UpgradeImplementationsInput _uii,
        DeployImplementationsOutput _dio,
        address _expectedCustomTokenAddr
    ) internal view {
        address portalProxy = _uii.optimismPortalProxy();
        address systemConfigProxy = _uii.systemConfigProxy();
        require(portalProxy != address(0), "OptimismPortal proxy is 0");
        require(systemConfigProxy != address(0), "SystemConfig proxy is 0");

        ISystemConfig systemConfig = ISystemConfig(systemConfigProxy);
        IOptimismPortal2 portal = IOptimismPortal2(payable(portalProxy));
        address expectedPortalImpl = address(_dio.optimismPortalImpl());
        address currentPortalImpl = address(uint160(uint256(vm.load(portalProxy, Constants.PROXY_IMPLEMENTATION_ADDRESS))));

        require(currentPortalImpl == expectedPortalImpl, "OptimismPortal impl mismatch");
        (address configuredGasToken, ) = systemConfig.gasPayingToken();
        require(configuredGasToken == _expectedCustomTokenAddr, "SystemConfig not configured with expected custom gas token");
        require(!portal.paused(), "OptimismPortal: Portal is PAUSED");
        console.log("  [PASS] OptimismPortal custom gas token validation.");
    }

    function _validateCustomGasTokenSystemConfig(
        UpgradeImplementationsInput _uii,
        DeployImplementationsOutput _dio,
        address _expectedCustomTokenAddr,
        string memory _expectedTokenName,
        string memory _expectedTokenSymbol
    ) internal view {
        address systemConfigProxy = _uii.systemConfigProxy();
        require(systemConfigProxy != address(0), "SystemConfig proxy is 0");
        ISystemConfig systemConfig = ISystemConfig(systemConfigProxy);
        console.log("  [INFO] SystemConfig reports a custom gas token is active.");

        (address tokenAddress, uint8 decimals) = systemConfig.gasPayingToken();
        console.log("     Token Address:", tokenAddress);
        console.log("     Decimals:", decimals);

        require(tokenAddress == _expectedCustomTokenAddr, "SystemConfig: Unexpected custom gas token address");
        require(tokenAddress != address(0), "SystemConfig: Custom gas token address is address(0)");
        require(tokenAddress != Constants.ETHER, "SystemConfig: Custom gas token address is Constants.ETHER, but isCustomGasToken is true");

        require(decimals == 18, "SystemConfig: Unexpected custom gas token decimals");

        string memory name = systemConfig.gasPayingTokenName();
        string memory symbol = systemConfig.gasPayingTokenSymbol();
        console.log("     Name from SystemConfig:", name);
        console.log("     Symbol from SystemConfig:", symbol);

        require(keccak256(abi.encodePacked(name)) == keccak256(abi.encodePacked(_expectedTokenName)), "SystemConfig: Unexpected token name");
        require(keccak256(abi.encodePacked(symbol)) == keccak256(abi.encodePacked(_expectedTokenSymbol)), "SystemConfig: Unexpected token symbol");

        console.log("  [PASS] Custom gas token address is not address(0).");
        console.log("  [PASS] Custom gas token address is not Constants.ETHER.");
        console.log("  [PASS] Custom gas token decimals are 18.");
        console.log("  [PASS] Custom gas token name is not empty.");
        console.log("  [PASS] Custom gas token symbol is not empty.");

        console.log("  === CUSTOM GAS TOKEN VALIDATION COMPLETE: SystemConfig ===");
    }
}
