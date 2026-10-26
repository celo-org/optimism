// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Types } from "src/libraries/Types.sol";
import { L1Block } from "src/L2/L1Block.sol";

// Models the legacy vault's immutable settings and slot-zero processed counter.
contract LegacyCeloVaultFixture {
    uint256 public totalProcessed = 73;
    address public immutable RECIPIENT;
    uint256 public constant MIN_WITHDRAWAL_AMOUNT = 23;
    Types.WithdrawalNetwork public constant WITHDRAWAL_NETWORK = Types.WithdrawalNetwork.L2;
    string public constant version = "1.5.0-celo";

    constructor(address recipient) {
        RECIPIENT = recipient;
    }
    receive() external payable { }
}

contract LegacyCeloL1BlockFixture is L1Block {
    function gasPayingTokenName() public pure override returns (string memory) {
        return "Celo";
    }

    function gasPayingTokenSymbol() public pure override returns (string memory) {
        return "CELO";
    }

    function isCustomGasToken() public pure override returns (bool) {
        return true;
    }

    function version() public pure override returns (string memory) {
        return "1.0.0";
    }
}

