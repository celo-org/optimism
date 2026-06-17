// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Interfaces
import { IStandardBridge } from "interfaces/universal/IStandardBridge.sol";
import { ISemver } from "interfaces/universal/ISemver.sol";

interface ICeloGasBridgeL2 is IStandardBridge, ISemver {
    error CeloGasBridgeL2_Disabled();
    error CeloGasBridgeL2_NotCgtMode();
    error CeloGasBridgeL2_ZeroRecipient();
    error CeloGasBridgeL2_InvalidRecipient();
    error CeloGasBridgeL2_ValueMismatch();
    error CeloGasBridgeL2_ZeroAmount();

    function initialize(IStandardBridge _otherBridge, address _celoTokenL1) external;

    function celoTokenL1() external view returns (address);

    function withdraw(address _to, uint256 _amount, uint32 _minGasLimit, bytes memory _extraData) external payable;

    function finalizeDeposit(address _from, address _to, uint256 _amount, bytes memory _extraData) external;
}
