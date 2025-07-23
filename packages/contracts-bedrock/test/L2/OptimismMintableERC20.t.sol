// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing utilities
import { CommonTest } from "test/setup/CommonTest.sol";

// Target contract
import { OptimismMintableERC20 } from "src/universal/OptimismMintableERC20.sol";

contract OptimismMintableERC20_Beenchmark_Test is CommonTest {
    // Test token
    OptimismMintableERC20 token;

    // Test addresses
    address bridge = address(0x1);
    address remoteToken = address(0x2);
    address recipient1 = address(0x3);
    address recipient2 = address(0x4);
    address communityFund = address(0x5);

    function setUp() public override {
        // Deploy OptimismMintableERC20 token for testing
        token = new OptimismMintableERC20(
            bridge,
            remoteToken,
            "Test Token",
            "TEST",
            18
        );
    }

    function test_hapyCase() public {
        // Existing empty test
    }

    function test_creditAndDebitGasFees() public {
        // Step 1: Test new version of creditGasFees with arrays
        address[] memory recipients = new address[](2);
        recipients[0] = recipient1;
        recipients[1] = recipient2;

        uint256[] memory amounts = new uint256[](2);
        amounts[0] = 100;
        amounts[1] = 200;

        // Call creditGasFees as the VM (address 0)
        vm.prank(address(0));
        token.creditGasFees(recipients, amounts);

        // Verify balances
        assertEq(token.balanceOf(recipient1), 100);
        assertEq(token.balanceOf(recipient2), 200);

        // Step 2: Test debitGasFees
        vm.prank(address(0));
        token.debitGasFees(recipient1, 50);

        // Verify balance after debit
        assertEq(token.balanceOf(recipient1), 50);

        // Step 3: Test legacy version of creditGasFees
        address from = address(0x6);
        address feeRecipient = address(0x7);
        address gatewayFeeRecipient = address(0x8); // unused
        uint256 refund = 75;
        uint256 tipTxFee = 25;
        uint256 gatewayFee = 10; // unused
        uint256 baseTxFee = 30;

        vm.prank(address(0));
        token.creditGasFees(
            from,
            feeRecipient,
            gatewayFeeRecipient,
            communityFund,
            refund,
            tipTxFee,
            gatewayFee,
            baseTxFee
        );

        // Verify balances after legacy creditGasFees
        assertEq(token.balanceOf(from), 75);
        assertEq(token.balanceOf(feeRecipient), 25);
        assertEq(token.balanceOf(communityFund), 30);
    }
}