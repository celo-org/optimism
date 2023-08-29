// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import "../universal/OptimismMintableERC20.sol";

contract BridgedETH is OptimismMintableERC20 {
    constructor(address _bridge) OptimismMintableERC20(_bridge, address(0), "Ether", "ETH", 18) { }
}
