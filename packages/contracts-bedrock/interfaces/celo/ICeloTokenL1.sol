// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface ICeloTokenL1 {
    function initialize(address escrowRecipient) external;
    function balanceOf(address account) external view returns (uint256);

    function __constructor__() external;
}
