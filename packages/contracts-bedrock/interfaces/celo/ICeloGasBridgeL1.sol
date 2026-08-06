// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Interfaces
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IProxyAdminOwnedBase } from "interfaces/L1/IProxyAdminOwnedBase.sol";
import { ICrossDomainMessenger } from "interfaces/universal/ICrossDomainMessenger.sol";
import { IReinitializableBase } from "interfaces/universal/IReinitializableBase.sol";
import { ISemver } from "interfaces/universal/ISemver.sol";
import { IStandardBridge } from "interfaces/universal/IStandardBridge.sol";

interface ICeloGasBridgeL1 is IStandardBridge, IProxyAdminOwnedBase, IReinitializableBase, ISemver {
    error CeloGasBridgeL1_Disabled();
    error CeloGasBridgeL1_NotCgtMode();
    error CeloGasBridgeL1_ZeroRecipient();
    error CeloGasBridgeL1_InvalidRecipient();
    error CeloGasBridgeL1_ZeroAmount();
    error CeloGasBridgeL1_Paused();
    error CeloGasBridgeL1_UnauthorizedSeeder();
    error CeloGasBridgeL1_EscrowAlreadySeeded();
    error CeloGasBridgeL1_NotActivated();

    event EscrowSeeded(address indexed portal, uint256 amount);

    function celoTokenL1() external view returns (IERC20);
    function systemConfig() external view returns (ISystemConfig);
    function escrowSeeded() external view returns (bool);
    function seedEscrow(uint256 _amount) external;
    function seedEscrowGenesis() external;

    function initialize(
        ICrossDomainMessenger _messenger,
        ISystemConfig _systemConfig,
        IStandardBridge _otherBridge,
        IERC20 _celoTokenL1
    )
        external;

    function deposit(address _to, uint256 _amount, uint32 _minGasLimit, bytes memory _extraData) external;

    function finalizeWithdrawal(address _from, address _to, uint256 _amount, bytes memory _extraData) external;

    function __constructor__() external override(IReinitializableBase, IStandardBridge); // Disables interface inheritance error
}
