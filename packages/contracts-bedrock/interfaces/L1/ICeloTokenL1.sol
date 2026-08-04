// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface ICeloTokenL1 {
    function initialize(address _portalProxyAddress) external;
    function name() external view returns (string memory);
    function symbol() external view returns (string memory);
    function decimals() external view returns (uint8);
    function totalSupply() external view returns (uint256);
    function balanceOf(address _account) external view returns (uint256);

    function __constructor__() external;
}
