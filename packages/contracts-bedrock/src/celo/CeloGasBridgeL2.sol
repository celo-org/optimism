// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Contracts
import { StandardBridge } from "src/universal/StandardBridge.sol";

// Interfaces
import { ICeloGasBridgeL1Finalizer } from "interfaces/celo/IBridgeFinalizers.sol";
import { IL1Block } from "interfaces/L2/IL1Block.sol";
import { ILiquidityController } from "interfaces/L2/ILiquidityController.sol";
import { ICrossDomainMessenger } from "interfaces/universal/ICrossDomainMessenger.sol";
import { ISemver } from "interfaces/universal/ISemver.sol";

// Libraries
import { Predeploys } from "src/libraries/Predeploys.sol";

/// @custom:proxied true
/// @custom:predeploy 0x4200000000000000000000000000000000001023
/// @title CeloGasBridgeL2
/// @notice Single-asset L2 bridge for native CELO under CGT v2.
/// @dev    Inherited `bridgeETH`, `bridgeETHTo`, `finalizeBridgeETH`, and `finalizeBridgeERC20`
///         are not marked `virtual` in upstream OP v6.0.0. They are disabled via the virtual
///         `_emit*` hook overrides below, which always revert in practice.
/// @dev    TODO: op-node MUST install this predeploy before relaying any user deposit in the
///         first post-flag L2 block. Otherwise, funds can be stuck.
contract CeloGasBridgeL2 is StandardBridge, ISemver {
    // ================================================================
    //                            ERRORS
    // ================================================================

    /// @notice Thrown when a disabled StandardBridge entrypoint is called.
    error CeloGasBridgeL2_Disabled();

    /// @notice Thrown when the system is not running in CGT mode.
    error CeloGasBridgeL2_NotCgtMode();

    /// @notice Thrown when the withdrawal or deposit recipient is the zero address.
    error CeloGasBridgeL2_ZeroRecipient();

    /// @notice Thrown when the recipient is the bridge itself or the cross-domain messenger.
    error CeloGasBridgeL2_InvalidRecipient();

    /// @notice Thrown when the ETH value sent does not match the bridged amount.
    error CeloGasBridgeL2_ValueMismatch();

    /// @notice Thrown when a zero-value withdrawal is attempted.
    error CeloGasBridgeL2_ZeroAmount();

    // ================================================================
    //                            STORAGE
    // ================================================================

    /// @notice Semantic version.
    /// @custom:semver 1.0.0
    string public constant version = "1.0.0";

    // ================================================================
    //                           EXTERNAL
    // ================================================================

    constructor() StandardBridge() {
        _disableInitializers();
    }

    /// @notice Initializer.
    /// @param _otherBridge Contract for the bridge on the other network.
    function initialize(StandardBridge _otherBridge) external initializer {
        __StandardBridge_init({
            _messenger: ICrossDomainMessenger(Predeploys.L2_CROSS_DOMAIN_MESSENGER),
            _otherBridge: _otherBridge
        });
    }

    /// @notice L2 bridge is not pausable; always returns false.
    function paused() public pure override returns (bool) {
        return false;
    }

    /// @notice Initiates a CELO withdrawal from L2 to L1.
    /// @param _to Recipient on L1.
    /// @param _amount Amount of CELO to bridge.
    /// @param _minGasLimit Minimum gas limit for the L1 finalization transaction.
    /// @param _extraData Additional data attached to the withdrawal.
    function withdraw(address _to, uint256 _amount, uint32 _minGasLimit, bytes calldata _extraData)
        external
        payable
    {
        if (!IL1Block(Predeploys.L1_BLOCK_ATTRIBUTES).isCustomGasToken()) {
            revert CeloGasBridgeL2_NotCgtMode();
        }
        if (_to == address(0)) revert CeloGasBridgeL2_ZeroRecipient();
        if (_to == address(this) || _to == address(messenger)) revert CeloGasBridgeL2_InvalidRecipient();
        if (_amount == 0) revert CeloGasBridgeL2_ZeroAmount();
        if (msg.value != _amount) revert CeloGasBridgeL2_ValueMismatch();

        ILiquidityController(Predeploys.LIQUIDITY_CONTROLLER).burn{ value: _amount }();

        // Both token fields are zero: on L2, CELO is the native asset, not an ERC20.
        // TODO: address(CELO_TOKEN) for indexer correlation — TBD.
        emit ERC20BridgeInitiated(address(0), address(0), msg.sender, _to, _amount, _extraData);

        messenger.sendMessage({
            _target: address(otherBridge),
            _message: abi.encodeCall(ICeloGasBridgeL1Finalizer.finalizeWithdrawal, (msg.sender, _to, _amount, _extraData)),
            _minGasLimit: _minGasLimit
        });
    }

    /// @notice Finalizes a CELO deposit from L1 by minting native CELO on L2.
    /// @param _from Address of the depositor on L1.
    /// @param _to Recipient on L2.
    /// @param _amount Amount of CELO to mint.
    /// @param _extraData Additional data attached to the deposit.
    function finalizeDeposit(address _from, address _to, uint256 _amount, bytes calldata _extraData)
        external
        onlyOtherBridge
    {
        if (_to == address(0)) revert CeloGasBridgeL2_ZeroRecipient();
        if (_to == address(this) || _to == address(messenger)) revert CeloGasBridgeL2_InvalidRecipient();

        ILiquidityController(Predeploys.LIQUIDITY_CONTROLLER).mint(_to, _amount);

        // NOTE: Both localToken and remoteToken are address(0)
        // TODO: address(CELO_TOKEN) for indexer correlation — TBD.
        emit ERC20BridgeFinalized(address(0), address(0), _from, _to, _amount, _extraData);
    }

    /// @dev Disabled — inherited from StandardBridge.
    ///      This bridge supports CGT (native CELO on L2) only — not arbitrary ERC20s.
    ///      Use `withdraw` for the CELO path.
    function bridgeERC20(address, address, uint256, uint32, bytes calldata) public pure override {
        revert CeloGasBridgeL2_Disabled();
    }

    /// @dev Disabled — inherited from StandardBridge.
    ///      This bridge supports CGT (native CELO on L2) only — not arbitrary ERC20s.
    ///      Use `withdraw` for the CELO path.
    function bridgeERC20To(address, address, address, uint256, uint32, bytes calldata) public pure override {
        revert CeloGasBridgeL2_Disabled();
    }

    /// @dev Disabled — inherited from StandardBridge.
    ///      No raw native value via `receive`. Use `withdraw` for the CELO path.
    receive() external payable override {
        revert CeloGasBridgeL2_Disabled();
    }

    // ================================================================
    //                           INTERNAL
    // ================================================================

    /// @dev Disabled — inherited from StandardBridge.
    ///      Blocks inherited `bridgeETH` / `bridgeETHTo`. CELO bridging goes through `withdraw`,
    ///      which burns native CELO via `LiquidityController`.
    function _emitETHBridgeInitiated(address, address, uint256, bytes memory) internal view override {
        if (address(messenger) == address(0)) return; // necessary to compile; prevents silencing unreachable errors
        revert CeloGasBridgeL2_Disabled();
    }

    /// @dev Disabled — inherited from StandardBridge.
    ///      Blocks inherited `finalizeBridgeETH`. CELO finalization goes through `finalizeDeposit`,
    ///      which mints native CELO via `LiquidityController`.
    function _emitETHBridgeFinalized(address, address, uint256, bytes memory) internal view override {
        if (address(messenger) == address(0)) return; // necessary to compile; prevents silencing unreachable errors
        revert CeloGasBridgeL2_Disabled();
    }

    /// @dev Disabled — inherited from StandardBridge.
    ///      Blocks inherited `finalizeBridgeERC20`. This bridge does not handle arbitrary ERC20s;
    ///      CELO finalization uses `finalizeDeposit`.
    function _emitERC20BridgeFinalized(address, address, address, address, uint256, bytes memory)
        internal
        view
        override
    {
        if (address(messenger) == address(0)) return; // necessary to compile; prevents silencing unreachable errors
        revert CeloGasBridgeL2_Disabled();
    }
}
