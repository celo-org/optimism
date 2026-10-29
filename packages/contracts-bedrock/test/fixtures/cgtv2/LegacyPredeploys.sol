// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
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

contract CgtFeeTokenFixture is ERC20 {
    constructor() ERC20("Fixture fee token", "FIX") { }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract RejectCgtInitializerFixture {
    error RejectCgtInitializerFixture_InitializationRejected();

    function initialize() external pure {
        revert RejectCgtInitializerFixture_InitializationRejected();
    }
}
