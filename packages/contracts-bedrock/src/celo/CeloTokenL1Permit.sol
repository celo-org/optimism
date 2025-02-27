// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { ERC20Permit } from "@openzeppelin/contracts-v5/token/ERC20/extensions/ERC20Permit.sol";
import { ERC20 } from "@openzeppelin/contracts-v5/token/ERC20/ERC20.sol";

contract CeloTokenL1Permit is ERC20Permit {
    string constant NAME = "Celo native asset";
    string constant SYMBOL = "CELO";
    uint256 constant TOTAL_MARKET_CAP = 1000000000e18; // 1 billion CELO

    constructor(address portalProxyAddress) ERC20(NAME, SYMBOL) ERC20Permit(NAME) {
        _mint(portalProxyAddress, TOTAL_MARKET_CAP);
    }
}
