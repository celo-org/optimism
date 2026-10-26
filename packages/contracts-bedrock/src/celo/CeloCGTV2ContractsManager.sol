// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Predeploys } from "src/libraries/Predeploys.sol";
import { L2ContractsManagerUtils } from "src/libraries/L2ContractsManagerUtils.sol";
import { L2ContractsManagerTypes } from "src/libraries/L2ContractsManagerTypes.sol";
import { IL1Block } from "interfaces/L2/IL1Block.sol";
import { ILiquidityController } from "interfaces/L2/ILiquidityController.sol";
import { IFeeVault } from "interfaces/L2/IFeeVault.sol";
import { ICeloGasBridgeL2 } from "interfaces/celo/ICeloGasBridgeL2.sol";
import { IStandardBridge } from "interfaces/universal/IStandardBridge.sol";
import { IL2ProxyAdmin } from "interfaces/L2/IL2ProxyAdmin.sol";

/// @title CeloCGTV2ContractsManager
/// @notice Installs CGTv2 through the upstream L2ProxyAdmin delegatecall upgrade path.
/// @dev Captures legacy settings before upgrades. A separate depositor transaction sets the CGT flag.
contract CeloCGTV2ContractsManager {
    struct Implementations {
        address storageSetter;
        address l1Block;
        address messagePasser;
        address nativeLiquidity;
        address controller;
        address bridge;
        address vault;
    }

    struct NetworkConfig {
        address bridgeProxy;
        address l1Bridge;
        address l1Token;
        address owner;
    }

    bytes32 internal constant INITIALIZABLE_SLOT_OZ_V5 =
        0xf0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a00;
    address internal immutable STORAGE_SETTER;
    address internal immutable L1_BLOCK;
    address internal immutable MESSAGE_PASSER;
    address internal immutable NATIVE_LIQUIDITY;
    address internal immutable CONTROLLER;
    address internal immutable BRIDGE;
    address internal immutable VAULT;
    address internal immutable BRIDGE_PROXY;
    address internal immutable L1_BRIDGE;
    address internal immutable L1_TOKEN;
    address internal immutable OWNER;

    error CeloCGTV2_OnlyProxyAdminDelegatecall();
    error CeloCGTV2_InvalidConfiguration();
    error CeloCGTV2_AlreadyInstalled();

    constructor(Implementations memory _impls, NetworkConfig memory _config) {
        if (
            !Predeploys.isUpgradeable(_config.bridgeProxy)
                || Predeploys.isSupportedPredeploy(_config.bridgeProxy, type(uint256).max, true, true)
                || _config.l1Bridge == address(0) || _config.l1Token == address(0) || _config.owner == address(0)
        ) {
            revert CeloCGTV2_InvalidConfiguration();
        }
        address[7] memory implementations = [
            _impls.storageSetter,
            _impls.l1Block,
            _impls.messagePasser,
            _impls.nativeLiquidity,
            _impls.controller,
            _impls.bridge,
            _impls.vault
        ];
        for (uint256 i; i < implementations.length; i++) {
            if (implementations[i].code.length == 0) revert CeloCGTV2_InvalidConfiguration();
        }
        STORAGE_SETTER = _impls.storageSetter;
        L1_BLOCK = _impls.l1Block;
        MESSAGE_PASSER = _impls.messagePasser;
        NATIVE_LIQUIDITY = _impls.nativeLiquidity;
        CONTROLLER = _impls.controller;
        BRIDGE = _impls.bridge;
        VAULT = _impls.vault;
        BRIDGE_PROXY = _config.bridgeProxy;
        L1_BRIDGE = _config.l1Bridge;
        L1_TOKEN = _config.l1Token;
        OWNER = _config.owner;
    }

    /// @notice Upgrades the six Celo predeploys, preserving the legacy fee-vault configuration.
    /// @dev Runs via L2ProxyAdmin. The controller is initialized with temporary ProxyAdmin ownership
    ///      to authorize the bridge, then ownership is transferred to the configured governance owner.
    ///      First-install checks prevent this initializer and authorization flow from running twice.
    function upgrade() external {
        if (address(this) != Predeploys.PROXY_ADMIN) revert CeloCGTV2_OnlyProxyAdminDelegatecall();
        IL2ProxyAdmin admin = IL2ProxyAdmin(Predeploys.PROXY_ADMIN);
        if (
            admin.getProxyImplementation(Predeploys.LIQUIDITY_CONTROLLER).code.length != 0
                || admin.getProxyImplementation(BRIDGE_PROXY).code.length != 0
        ) {
            revert CeloCGTV2_AlreadyInstalled();
        }

        L2ContractsManagerTypes.FeeVaultConfig memory vault =
            L2ContractsManagerUtils.readFeeVaultConfig(Predeploys.SEQUENCER_FEE_WALLET);
        string memory name = IL1Block(Predeploys.L1_BLOCK_ATTRIBUTES).gasPayingTokenName();
        string memory symbol = IL1Block(Predeploys.L1_BLOCK_ATTRIBUTES).gasPayingTokenSymbol();

        L2ContractsManagerUtils.upgradeTo(Predeploys.L1_BLOCK_ATTRIBUTES, L1_BLOCK, false);
        L2ContractsManagerUtils.upgradeTo(Predeploys.L2_TO_L1_MESSAGE_PASSER, MESSAGE_PASSER, false);
        L2ContractsManagerUtils.upgradeTo(Predeploys.NATIVE_ASSET_LIQUIDITY, NATIVE_LIQUIDITY, false);
        L2ContractsManagerUtils.upgradeToAndCall(
            Predeploys.LIQUIDITY_CONTROLLER,
            CONTROLLER,
            STORAGE_SETTER,
            abi.encodeCall(ILiquidityController.initialize, (address(this), name, symbol)),
            bytes32(0),
            0,
            false
        );
        L2ContractsManagerUtils.upgradeToAndCall(
            BRIDGE_PROXY,
            BRIDGE,
            STORAGE_SETTER,
            abi.encodeCall(ICeloGasBridgeL2.initialize, (IStandardBridge(payable(L1_BRIDGE)), L1_TOKEN)),
            bytes32(0),
            0,
            false
        );
        L2ContractsManagerUtils.upgradeToAndCall(
            Predeploys.SEQUENCER_FEE_WALLET,
            VAULT,
            STORAGE_SETTER,
            abi.encodeCall(IFeeVault.initialize, (vault.recipient, vault.minWithdrawalAmount, vault.withdrawalNetwork)),
            INITIALIZABLE_SLOT_OZ_V5,
            0,
            false
        );

        ILiquidityController controller = ILiquidityController(Predeploys.LIQUIDITY_CONTROLLER);
        controller.authorizeMinter(BRIDGE_PROXY);
        controller.transferOwnership(OWNER);
    }
}
