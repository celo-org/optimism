// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { ERC20Upgradeable } from "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";

contract CeloTokenL1 is ERC20Upgradeable {
    string constant NAME = "Celo";
    string constant SYMBOL = "CELO";
    uint256 constant TOTAL_MARKET_CAP = 1000000000e18; // 1 billion CELO

    constructor() {
        _disableInitializers();
    }

    /// @param escrowRecipient Address that receives the full CELO supply and escrows it. A fresh CGT v2
    ///                        chain passes the CeloGasBridgeL1 proxy, whereas the legacy v1 deployment
    ///                        passed the OptimismPortal.
    function initialize(address escrowRecipient) external initializer {
        __ERC20_init(NAME, SYMBOL);
        _mint(escrowRecipient, TOTAL_MARKET_CAP);
    }
}
