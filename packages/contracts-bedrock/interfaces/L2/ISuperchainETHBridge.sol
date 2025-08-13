// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { ISemver } from "interfaces/universal/ISemver.sol";

interface ISuperchainETHBridge is ISemver {
    error Unauthorized();
    error NotCustomGasToken();
    error InvalidCrossDomainSender();
    error ZeroAddress();

    event SendETH(address indexed from, address indexed to, uint256 amount, uint256 destination);

    event RelayETH(address indexed from, address indexed to, uint256 amount, uint256 source);

    function balanceOf(address src) external view returns (uint256);
    function withdraw(uint256 _amount) external;
    function supportsInterface(bytes4 _interfaceId) external view returns (bool);
    function sendETH(address _to, uint256 _chainId) external payable returns (bytes32 msgHash_);
    function relayETH(address _from, address _to, uint256 _amount) external;

    function __constructor__() external;
}
