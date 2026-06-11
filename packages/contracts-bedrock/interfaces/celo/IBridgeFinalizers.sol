// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @notice Minimal interface used by `CeloGasBridgeL2` to encode cross-domain messages targeting
///         `CeloGasBridgeL1.finalizeWithdrawal`.
interface ICeloGasBridgeL1Finalizer {
    function finalizeWithdrawal(address _from, address _to, uint256 _amount, bytes calldata _extraData) external;
}

/// @notice Minimal interface used by `CeloGasBridgeL1` to encode cross-domain messages targeting
///         `CeloGasBridgeL2.finalizeDeposit`.
interface ICeloGasBridgeL2Finalizer {
    function finalizeDeposit(address _from, address _to, uint256 _amount, bytes calldata _extraData) external;
}
