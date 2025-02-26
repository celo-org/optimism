// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { ERC20Upgradeable } from "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";

contract CeloTokenL1 is ERC20Upgradeable {
    string constant NAME = "Celo native asset";
    string constant SYMBOL = "CELO";
    uint256 constant TOTAL_MARKET_CAP = 1000000000e18; // 1 billion CELO

    constructor() {
        _disableInitializers();
    }

    function initialize(address portalProxyAddress) external initializer {
        __ERC20_init(NAME, SYMBOL);
        _mint(portalProxyAddress, TOTAL_MARKET_CAP);
    }
}
