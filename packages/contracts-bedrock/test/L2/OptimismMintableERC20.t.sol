// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing utilities
import { CommonTest } from "test/setup/CommonTest.sol";

// Target contract
import { OptimismMintableERC20 } from "src/universal/OptimismMintableERC20.sol";

// run this benchmark with:
// `forge test --match-contract="OptimismMintableERC20_Beenchmark_Test" --gas-report`

contract OptimismMintableERC20_Beenchmark_Test is CommonTest {
    // Test token
    OptimismMintableERC20 token;

    // Test addresses
    address bridge = address(0x1);
    address remoteToken = address(0x2);
    address recipient1 = address(0x3);
    address recipient2 = address(0x4);
    address communityFund = address(0x5);
    address from;
    address feeRecipient;

    // Fee amounts
    uint256 refund;
    uint256 tipTxFee;
    uint256 baseTxFee;

    // Standard token amounts
    uint256 initialAmount1 = 100;
    uint256 initialAmount2 = 200;
    uint256 fromAmount = 25;
    uint256 feeRecipientAmount = 30;
    uint256 communityFundAmount = 10;

    function setUp() public override {
        // Deploy OptimismMintableERC20 token for testing
        token = new OptimismMintableERC20(
            bridge,
            remoteToken,
            "Test Token",
            "TEST",
            18
        );

        // Initialize common test values
        from = address(0x6);
        feeRecipient = address(0x7);
        refund = 75;
        tipTxFee = 25;
        baseTxFee = 30;

        // Mint initial token balances for all addresses
        vm.startPrank(bridge);
        token.mint(recipient1, initialAmount1);
        token.mint(recipient2, initialAmount2);
        token.mint(from, fromAmount);
        vm.stopPrank();
    }

    // Helper function for minting tokens using bridge
    function _mintTokens(address recipient, uint256 amount) internal {
        vm.prank(bridge);
        token.mint(recipient, amount);
    }

    // Helper function for debitGasFees
    function _debitGasFees(address from, uint256 amount) internal {
        vm.prank(address(0));
        token.debitGasFees(from, amount);
    }

    // Helper function for legacy creditGasFees
    function _legacyCreditGasFees(
        address from,
        address feeRecipient,
        address communityFundAddr,
        uint256 refund,
        uint256 tipTxFee,
        uint256 baseTxFee
    ) internal {
        address gatewayFeeRecipient = address(0x8); // unused
        uint256 gatewayFee = 10; // unused

        vm.prank(address(0));
        token.creditGasFees(
            from,
            feeRecipient,
            gatewayFeeRecipient,
            communityFundAddr,
            refund,
            tipTxFee,
            gatewayFee,
            baseTxFee
        );
    }

    function test_creditAndDebitGasFees() public {
        // Verify initial balances
        // assertEq(token.balanceOf(recipient1), initialAmount1);
        // assertEq(token.balanceOf(recipient2), initialAmount2);
        // assertEq(token.balanceOf(from), fromAmount);

        // Step 1: Test debitGasFees
        _debitGasFees(recipient1, 50);

        // Verify balance after debit
        assertEq(token.balanceOf(recipient1), initialAmount1 - 50);

        // Step 2: Test legacy version of creditGasFees
        _legacyCreditGasFees(
            from,
            feeRecipient,
            communityFund,
            refund,
            tipTxFee,
            baseTxFee
        );
    }

    function test_creditAndDebitGasFeesWithExistingBalances() public {
        // Initial balances were already set in setUp
        vm.startPrank(bridge);
        token.mint(feeRecipient, feeRecipientAmount);
        token.mint(communityFund, communityFundAmount);
        vm.stopPrank();

        // Step 1: Test debitGasFees with existing balance
        _debitGasFees(recipient1, 50);

        // Verify balance after debit
        assertEq(token.balanceOf(recipient1), initialAmount1 - 50);

        // Step 2: Test legacy creditGasFees with existing balances
        _legacyCreditGasFees(
            from,
            feeRecipient,
            communityFund,
            refund,
            tipTxFee,
            baseTxFee
        );

    }
}