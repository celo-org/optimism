// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Contracts
import { ProxyAdminOwnedBase } from "src/L1/ProxyAdminOwnedBase.sol";
import { StandardBridge } from "src/universal/StandardBridge.sol";
import { ReinitializableBase } from "src/universal/ReinitializableBase.sol";

// Interfaces
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import { ICeloGasBridgeL2Finalizer } from "interfaces/celo/IBridgeFinalizers.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { ICrossDomainMessenger } from "interfaces/universal/ICrossDomainMessenger.sol";
import { ISemver } from "interfaces/universal/ISemver.sol";

// Libraries
import { CeloPredeploys } from "src/celo/CeloPredeploys.sol";

/// @custom:proxied true
/// @title CeloGasBridgeL1
/// @notice Single-asset L1 bridge for CELO under CGT v2.
/// @dev    Inherited `bridgeETH`, `bridgeETHTo`, `finalizeBridgeETH`, and `finalizeBridgeERC20`
///         are not marked `virtual` in upstream OP v6.0.0. They are disabled via the virtual
///         `_emit*` hook overrides below, which always revert in practice.
contract CeloGasBridgeL1 is StandardBridge, ProxyAdminOwnedBase, ReinitializableBase, ISemver {
    using SafeERC20 for IERC20;

    // ================================================================
    //                            ERRORS
    // ================================================================

    /// @notice Thrown when a disabled StandardBridge entrypoint is called.
    error CeloGasBridgeL1_Disabled();

    /// @notice Thrown when the system is not running in CGT mode.
    error CeloGasBridgeL1_NotCgtMode();

    /// @notice Thrown when the withdrawal or deposit recipient is the zero address.
    error CeloGasBridgeL1_ZeroRecipient();

    /// @notice Thrown when the recipient is the bridge itself or the cross-domain messenger.
    error CeloGasBridgeL1_InvalidRecipient();

    /// @notice Thrown when a zero-value deposit is attempted.
    error CeloGasBridgeL1_ZeroAmount();

    /// @notice Thrown when the bridge is paused.
    error CeloGasBridgeL1_Paused();

    /// @notice Thrown when escrow seeding is not called by the portal proxy.
    error CeloGasBridgeL1_UnauthorizedSeeder();

    /// @notice Thrown when escrow seeding is attempted more than once.
    error CeloGasBridgeL1_EscrowAlreadySeeded();

    /// @notice Thrown when depositing before the CGT v2 migration has seeded the escrow.
    error CeloGasBridgeL1_NotActivated();

    // ================================================================
    //                            EVENTS
    // ================================================================

    /// @notice Emitted when escrow is seeded from the drained portal amount.
    event EscrowSeeded(address indexed portal, uint256 amount);

    // ================================================================
    //                            STORAGE
    // ================================================================

    /// @notice CELO token bridged by this contract.
    /// @custom:network-specific
    IERC20 public celoTokenL1;

    /// @notice Address of the SystemConfig contract.
    ISystemConfig public systemConfig;

    /// @notice True once escrow was seeded.
    bool public escrowSeeded;

    /// @notice Reserved storage slots for future upgrades. Sized to leave the parent's storage
    ///         contiguous and to allow new state in subsequent versions without layout shifts.
    uint256[48] private __gap;

    /// @notice Semantic version.
    /// @custom:semver 1.0.0
    string public constant version = "1.0.0";

    // ================================================================
    //                           EXTERNAL
    // ================================================================

    constructor() StandardBridge() ReinitializableBase(1) {
        _disableInitializers();
    }

    /// @notice Initializer.
    /// @param _messenger Contract for the CrossDomainMessenger on this network.
    /// @param _systemConfig Contract for the SystemConfig on this network.
    /// @param _otherBridge Contract for the bridge on the other network.
    /// @param _celoTokenL1 CELO token on L1.
    function initialize(
        ICrossDomainMessenger _messenger,
        ISystemConfig _systemConfig,
        StandardBridge _otherBridge,
        IERC20 _celoTokenL1
    )
        external
        reinitializer(initVersion())
    {
        _assertOnlyProxyAdminOrProxyAdminOwner();

        celoTokenL1 = _celoTokenL1;
        systemConfig = _systemConfig;
        __StandardBridge_init({ _messenger: _messenger, _otherBridge: _otherBridge });
    }

    /// @notice One-shot escrow seeding by the OptimismPortal (through PortalMigrator) with the drained CELO balance.
    /// @param _amount The actual CELO balance drained from the portal into this bridge.
    function seedEscrow(uint256 _amount) external {
        address portal = systemConfig.optimismPortal();
        if (msg.sender != portal) revert CeloGasBridgeL1_UnauthorizedSeeder();
        if (escrowSeeded) revert CeloGasBridgeL1_EscrowAlreadySeeded();

        escrowSeeded = true;
        deposits[address(celoTokenL1)][address(0)] = _amount;

        emit EscrowSeeded(portal, _amount);
    }

    /// @notice One-shot escrow seeding for a new chain deployed at genesis on v6.
    ///         Called by the ProxyAdmin through OPCM. Chains migrating from v5 use `seedEscrow`.
    function seedEscrowGenesis() external {
        _assertOnlyProxyAdminOrProxyAdminOwner();
        if (escrowSeeded) revert CeloGasBridgeL1_EscrowAlreadySeeded();

        uint256 amount = celoTokenL1.balanceOf(address(this));

        escrowSeeded = true;
        deposits[address(celoTokenL1)][address(0)] = amount;

        emit EscrowSeeded(msg.sender, amount);
    }

    /// @notice Pause state is delegated to SystemConfig.
    function paused() public view override returns (bool) {
        return systemConfig.paused();
    }

    /// @notice Deposits CELO into escrow for later L2 finalization.
    /// @param _to Recipient on L2.
    /// @param _amount Amount of CELO to deposit.
    /// @param _minGasLimit Minimum gas limit for the L2 deposit transaction.
    /// @param _extraData Additional data attached to the deposit.
    function deposit(address _to, uint256 _amount, uint32 _minGasLimit, bytes calldata _extraData) external {
        if (!escrowSeeded) revert CeloGasBridgeL1_NotActivated();
        if (paused()) revert CeloGasBridgeL1_Paused();
        if (!systemConfig.isCustomGasToken()) revert CeloGasBridgeL1_NotCgtMode();
        if (_to == address(0)) revert CeloGasBridgeL1_ZeroRecipient();
        if (_to == address(this) || _to == address(messenger)) revert CeloGasBridgeL1_InvalidRecipient();
        if (_amount == 0) revert CeloGasBridgeL1_ZeroAmount();

        celoTokenL1.safeTransferFrom(msg.sender, address(this), _amount);
        deposits[address(celoTokenL1)][address(0)] += _amount;
        // localToken = L1 CELO (celoTokenL1); remoteToken = L2 CELO (GOLD_TOKEN).
        _emitERC20BridgeInitiated(address(celoTokenL1), CeloPredeploys.GOLD_TOKEN, msg.sender, _to, _amount, _extraData);

        messenger.sendMessage({
            _target: address(otherBridge),
            _message: abi.encodeCall(ICeloGasBridgeL2Finalizer.finalizeDeposit, (msg.sender, _to, _amount, _extraData)),
            _minGasLimit: _minGasLimit
        });
    }

    /// @notice Finalizes a CELO withdrawal from L2.
    /// @param _from Address of the withdrawer on L2.
    /// @param _to Recipient on L1.
    /// @param _amount Amount of CELO to release.
    /// @param _extraData Additional data attached to the withdrawal.
    function finalizeWithdrawal(
        address _from,
        address _to,
        uint256 _amount,
        bytes calldata _extraData
    )
        external
        onlyOtherBridge
    {
        if (paused()) revert CeloGasBridgeL1_Paused();
        if (_to == address(0)) revert CeloGasBridgeL1_ZeroRecipient();
        if (_to == address(this) || _to == address(messenger)) revert CeloGasBridgeL1_InvalidRecipient();

        deposits[address(celoTokenL1)][address(0)] -= _amount;
        celoTokenL1.safeTransfer(_to, _amount);

        // Emit event directly. `_emitERC20BridgeFinalized` reverts to disable the
        // inherited `finalizeBridgeERC20` path. remoteToken = GOLD_TOKEN (L2 CELO id).
        emit ERC20BridgeFinalized(address(celoTokenL1), CeloPredeploys.GOLD_TOKEN, _from, _to, _amount, _extraData);
    }

    /// @dev Disabled — inherited from StandardBridge.
    ///      This bridge supports CGT (CELO ERC20) only — not arbitrary ERC20s.
    ///      Use `deposit` for the CELO path.
    function bridgeERC20(address, address, uint256, uint32, bytes calldata) public pure override {
        revert CeloGasBridgeL1_Disabled();
    }

    /// @dev Disabled — inherited from StandardBridge.
    ///      This bridge supports CGT (CELO ERC20) only — not arbitrary ERC20s.
    ///      Use `deposit` for the CELO path.
    function bridgeERC20To(address, address, address, uint256, uint32, bytes calldata) public pure override {
        revert CeloGasBridgeL1_Disabled();
    }

    /// @dev Disabled — inherited from StandardBridge.
    ///      No raw native value accepted. CELO is an L1 ERC20; use `deposit`.
    receive() external payable override {
        revert CeloGasBridgeL1_Disabled();
    }

    // ================================================================
    //                           INTERNAL
    // ================================================================

    /// @dev Disabled — inherited from StandardBridge.
    ///      Blocks inherited `bridgeETH` / `bridgeETHTo`. CELO is bridged as an L1 ERC20, not ETH.
    function _emitETHBridgeInitiated(address, address, uint256, bytes memory) internal view override {
        if (address(messenger) == address(0)) return; // necessary to compile; prevents silencing unreachable errors
        revert CeloGasBridgeL1_Disabled();
    }

    /// @dev Disabled — inherited from StandardBridge.
    ///      Blocks inherited `finalizeBridgeETH`. CELO is bridged as an L1 ERC20, not ETH.
    function _emitETHBridgeFinalized(address, address, uint256, bytes memory) internal view override {
        if (address(messenger) == address(0)) return; // necessary to compile; prevents silencing unreachable errors
        revert CeloGasBridgeL1_Disabled();
    }

    /// @dev Disabled — inherited from StandardBridge.
    ///      Blocks inherited `finalizeBridgeERC20`. CELO finalization uses `finalizeWithdrawal`,
    ///      which emits `ERC20BridgeFinalized` directly.
    function _emitERC20BridgeFinalized(address, address, address, address, uint256, bytes memory)
        internal
        view
        override
    {
        if (address(messenger) == address(0)) return; // necessary to compile; prevents silencing unreachable errors
        revert CeloGasBridgeL1_Disabled();
    }
}
