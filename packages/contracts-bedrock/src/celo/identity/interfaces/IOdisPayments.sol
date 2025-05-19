// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IOdisPayments {
    function payInCUSD(address account, uint256 value) external;
    function totalPaidCUSD(address) external view returns (uint256);
}
