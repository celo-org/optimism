// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing
import { CommonTest } from "test/setup/CommonTest.sol";
import { Reverter } from "test/mocks/Callers.sol";
import { TestERC20 } from "test/mocks/TestERC20.sol";

// Interfaces
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { ICeloSequencerFeeVault } from "interfaces/L2/ICeloSequencerFeeVault.sol";

// Libraries
import { Hashing } from "src/libraries/Hashing.sol";
import { Types } from "src/libraries/Types.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";
import { CeloPredeploys } from "src/celo/CeloPredeploys.sol";

// Celo interfaces
import { ICeloRegistry } from "src/celo/interfaces/ICeloRegistry.sol";
import { IFeeCurrencyDirectory } from "src/celo/interfaces/IFeeCurrencyDirectory.sol";
import { IFeeCurrencyAdapter } from "src/celo/interfaces/IFeeCurrencyAdapter.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/// @title CeloSequencerFeeVault_TestInit
/// @notice Reusable test initialization for `CeloSequencerFeeVault` tests.
contract CeloSequencerFeeVault_TestInit is CommonTest {
    address recipient;

    /// @dev Sets up the test suite.
    function setUp() public virtual override {
        super.setUp();
        recipient = deploy.cfg().sequencerFeeVaultRecipient();
    }
}

/// @title CeloSequencerFeeVault_Constructor_Test
/// @notice Tests the initialized configuration of the `CeloSequencerFeeVault` contract.
contract CeloSequencerFeeVault_Constructor_Test is CeloSequencerFeeVault_TestInit {
    /// @notice Tests that the l1 fee wallet and initialized values are correct.
    function test_constructor_succeeds() external view {
        assertEq(sequencerFeeVault.l1FeeWallet(), recipient);
        assertEq(sequencerFeeVault.RECIPIENT(), recipient);
        assertEq(sequencerFeeVault.recipient(), recipient);
        assertEq(sequencerFeeVault.MIN_WITHDRAWAL_AMOUNT(), deploy.cfg().sequencerFeeVaultMinimumWithdrawalAmount());
        assertEq(sequencerFeeVault.minWithdrawalAmount(), deploy.cfg().sequencerFeeVaultMinimumWithdrawalAmount());
        assertEq(uint8(sequencerFeeVault.WITHDRAWAL_NETWORK()), uint8(Types.WithdrawalNetwork.L1));
        assertEq(uint8(sequencerFeeVault.withdrawalNetwork()), uint8(Types.WithdrawalNetwork.L1));
    }
}

/// @title CeloSequencerFeeVault_Receive_Test
/// @notice Tests the `receive` function of the `CeloSequencerFeeVault` contract.
contract CeloSequencerFeeVault_Receive_Test is CeloSequencerFeeVault_TestInit {
    /// @notice Tests that the fee vault is able to receive ETH.
    function test_receive_succeeds() external {
        uint256 balance = address(sequencerFeeVault).balance;

        vm.prank(alice);
        (bool success,) = address(sequencerFeeVault).call{ value: 100 }(hex"");

        assertEq(success, true);
        assertEq(address(sequencerFeeVault).balance, balance + 100);
    }
}

/// @title CeloSequencerFeeVault_Withdraw_Test
/// @notice Tests the `withdraw` function of the `CeloSequencerFeeVault` contract.
contract CeloSequencerFeeVault_Withdraw_Test is CeloSequencerFeeVault_TestInit {
    /// @notice Helper function to set up L2 withdrawal configuration using the v6 setter pattern.
    function _setupL2Withdrawal() internal {
        vm.prank(IProxyAdmin(Predeploys.PROXY_ADMIN).owner());
        sequencerFeeVault.setWithdrawalNetwork(Types.WithdrawalNetwork.L2);
        recipient = deploy.cfg().sequencerFeeVaultRecipient();
    }

    /// @notice Tests that `withdraw` reverts when the balance is below the minimum.
    function test_withdraw_notEnough_reverts() external {
        assert(address(sequencerFeeVault).balance < sequencerFeeVault.MIN_WITHDRAWAL_AMOUNT());

        vm.expectRevert("FeeVault: withdrawal amount must be greater than minimum withdrawal amount");
        sequencerFeeVault.withdraw();
    }

    /// @notice Tests that `withdraw` successfully initiates a withdrawal to L1.
    function test_withdraw_toL1_succeeds() external {
        uint256 amount = sequencerFeeVault.MIN_WITHDRAWAL_AMOUNT() + 1;
        vm.deal(address(sequencerFeeVault), amount);

        // No ether has been withdrawn yet
        assertEq(sequencerFeeVault.totalProcessed(), 0);

        vm.expectEmit(address(Predeploys.SEQUENCER_FEE_WALLET));
        emit Withdrawal(address(sequencerFeeVault).balance, recipient, address(this));
        vm.expectEmit(address(Predeploys.SEQUENCER_FEE_WALLET));
        emit Withdrawal(address(sequencerFeeVault).balance, recipient, address(this), Types.WithdrawalNetwork.L1);

        // The entire vault's balance is withdrawn
        vm.expectCall(Predeploys.L2_TO_L1_MESSAGE_PASSER, address(sequencerFeeVault).balance, hex"");

        // The message is passed to the correct recipient
        vm.expectEmit(Predeploys.L2_TO_L1_MESSAGE_PASSER);
        emit MessagePassed(
            l2ToL1MessagePasser.messageNonce(),
            address(sequencerFeeVault),
            recipient,
            amount,
            400_000,
            hex"",
            Hashing.hashWithdrawal(
                Types.WithdrawalTransaction({
                    nonce: l2ToL1MessagePasser.messageNonce(),
                    sender: address(sequencerFeeVault),
                    target: recipient,
                    value: amount,
                    gasLimit: 400_000,
                    data: hex""
                })
            )
        );

        sequencerFeeVault.withdraw();

        // The withdrawal was successful
        assertEq(sequencerFeeVault.totalProcessed(), amount);
        assertEq(address(sequencerFeeVault).balance, 0);
        assertEq(Predeploys.L2_TO_L1_MESSAGE_PASSER.balance, amount);
    }

    /// @notice Tests that `withdraw` successfully initiates a withdrawal to L2.
    function test_withdraw_toL2_succeeds() external {
        _setupL2Withdrawal();

        uint256 amount = sequencerFeeVault.MIN_WITHDRAWAL_AMOUNT() + 1;
        vm.deal(address(sequencerFeeVault), amount);

        // No ether has been withdrawn yet
        assertEq(sequencerFeeVault.totalProcessed(), 0);

        vm.expectEmit(address(Predeploys.SEQUENCER_FEE_WALLET));
        emit Withdrawal(address(sequencerFeeVault).balance, sequencerFeeVault.RECIPIENT(), address(this));
        vm.expectEmit(address(Predeploys.SEQUENCER_FEE_WALLET));
        emit Withdrawal(
            address(sequencerFeeVault).balance, sequencerFeeVault.RECIPIENT(), address(this), Types.WithdrawalNetwork.L2
        );

        // The entire vault's balance is withdrawn
        vm.expectCall(recipient, address(sequencerFeeVault).balance, bytes(""));

        sequencerFeeVault.withdraw();

        // The withdrawal was successful
        assertEq(sequencerFeeVault.totalProcessed(), amount);
        assertEq(address(sequencerFeeVault).balance, 0);
        assertEq(recipient.balance, amount);
    }

    /// @notice Tests that `withdraw` fails when the recipient reverts (e.g. insufficient gas).
    function test_withdraw_toL2recipientReverts_fails() external {
        _setupL2Withdrawal();

        uint256 amount = sequencerFeeVault.MIN_WITHDRAWAL_AMOUNT();

        vm.deal(address(sequencerFeeVault), amount);
        // No ether has been withdrawn yet
        assertEq(sequencerFeeVault.totalProcessed(), 0);

        // Ensure the RECIPIENT reverts
        vm.etch(sequencerFeeVault.RECIPIENT(), type(Reverter).runtimeCode);

        // The entire vault's balance is withdrawn
        vm.expectCall(recipient, address(sequencerFeeVault).balance, bytes(""));
        vm.expectRevert("FeeVault: failed to send ETH to L2 fee recipient");
        sequencerFeeVault.withdraw();
        assertEq(sequencerFeeVault.totalProcessed(), 0);
    }
}

/// @title CeloSequencerFeeVault_WithdrawToken_Test
/// @notice Tests the `withdrawToken` overloads of the `CeloSequencerFeeVault` contract.
contract CeloSequencerFeeVault_WithdrawToken_Test is CeloSequencerFeeVault_TestInit {
    /// @dev Stand-in address used to mock the FeeCurrencyDirectory.
    address internal constant DIRECTORY = address(0xD12EC707);

    /// @dev Stand-in address used to mock an oracle (only its non-zeroness matters).
    address internal constant ORACLE = address(0x07AC1E);

    /// @dev keccak256("FeeCurrencyDirectory") — must match the constant in the contract.
    bytes32 internal constant FCD_ID = keccak256(abi.encodePacked("FeeCurrencyDirectory"));

    /// @dev Cached event signature, mirrored here so `vm.expectEmit` can match.
    event TokenWithdrawal(
        address indexed registered, address indexed actual, uint256 value, address indexed to, address from
    );

    function setUp() public override {
        super.setUp();

        // Always mock the registry → directory lookup.
        vm.mockCall(
            CeloPredeploys.CELO_REGISTRY,
            abi.encodeCall(ICeloRegistry.getAddressForOrDie, (FCD_ID)),
            abi.encode(DIRECTORY)
        );
    }

    function _mockCurrencyConfig(address token, address oracle) internal {
        vm.mockCall(
            DIRECTORY,
            abi.encodeCall(IFeeCurrencyDirectory.getCurrencyConfig, (token)),
            abi.encode(IFeeCurrencyDirectory.CurrencyConfig({ oracle: oracle, intrinsicGas: 0 }))
        );
    }

    function _mockAdaptedToken(address adapter, address underlying) internal {
        vm.mockCall(adapter, abi.encodeCall(IFeeCurrencyAdapter.getAdaptedToken, ()), abi.encode(underlying));
    }

    /// @notice `withdrawToken(token, token)` succeeds for a directly-registered native fee currency.
    function test_withdrawToken_native_succeeds() external {
        TestERC20 token = new TestERC20();
        _mockCurrencyConfig(address(token), ORACLE);

        uint256 amount = 1_000 ether;
        token.mint(address(sequencerFeeVault), amount);

        vm.expectEmit(address(sequencerFeeVault));
        emit TokenWithdrawal(address(token), address(token), amount, recipient, address(this));

        sequencerFeeVault.withdrawToken(address(token), address(token));

        assertEq(token.balanceOf(address(sequencerFeeVault)), 0);
        assertEq(token.balanceOf(recipient), amount);
        assertEq(sequencerFeeVault.totalProcessedToken(address(token)), amount);
    }

    /// @notice One-arg overload delegates to the two-arg version with the address repeated.
    function test_withdrawToken_singleArgOverload_succeeds() external {
        TestERC20 token = new TestERC20();
        _mockCurrencyConfig(address(token), ORACLE);

        uint256 amount = 42;
        token.mint(address(sequencerFeeVault), amount);

        vm.expectEmit(address(sequencerFeeVault));
        emit TokenWithdrawal(address(token), address(token), amount, recipient, address(this));

        sequencerFeeVault.withdrawToken(address(token));

        assertEq(token.balanceOf(address(sequencerFeeVault)), 0);
        assertEq(token.balanceOf(recipient), amount);
        assertEq(sequencerFeeVault.totalProcessedToken(address(token)), amount);
    }

    /// @notice For an adapter-registered fee currency, the underlying ERC-20 is transferred.
    function test_withdrawToken_adapter_succeeds() external {
        // Adapter is a mocked address; no deployed contract needed.
        address adapter = address(0xADAB7E5);
        TestERC20 underlying = new TestERC20();
        _mockCurrencyConfig(adapter, ORACLE);
        _mockAdaptedToken(adapter, address(underlying));

        uint256 amount = 26_660_316; // shaped like a USDC balance (6 decimals)
        underlying.mint(address(sequencerFeeVault), amount);

        vm.expectEmit(address(sequencerFeeVault));
        emit TokenWithdrawal(adapter, address(underlying), amount, recipient, address(this));

        sequencerFeeVault.withdrawToken(adapter, address(underlying));

        assertEq(underlying.balanceOf(address(sequencerFeeVault)), 0);
        assertEq(underlying.balanceOf(recipient), amount);
        assertEq(sequencerFeeVault.totalProcessedToken(address(underlying)), amount);
        // Adapter mapping key is NOT used.
        assertEq(sequencerFeeVault.totalProcessedToken(adapter), 0);
    }

    /// @notice Reverts when the registered token has no oracle in the directory (i.e. unregistered).
    function test_withdrawToken_invalidRegistered_reverts() external {
        TestERC20 token = new TestERC20();
        _mockCurrencyConfig(address(token), address(0));

        vm.expectRevert(ICeloSequencerFeeVault.CeloSequencerFeeVault_NotRegisteredFeeCurrency.selector);
        sequencerFeeVault.withdrawToken(address(token), address(token));
    }

    /// @notice Reverts when adapter routing is requested but `getAdaptedToken` returns the wrong address.
    function test_withdrawToken_adapterMismatch_reverts() external {
        address adapter = address(0xADAB7E5);
        TestERC20 wrongUnderlying = new TestERC20();
        TestERC20 actualPassed = new TestERC20();
        _mockCurrencyConfig(adapter, ORACLE);
        _mockAdaptedToken(adapter, address(wrongUnderlying));

        vm.expectRevert(ICeloSequencerFeeVault.CeloSequencerFeeVault_AdapterMismatch.selector);
        sequencerFeeVault.withdrawToken(adapter, address(actualPassed));
    }

    /// @notice Reverts when the vault has no balance of a valid registered token.
    function test_withdrawToken_zeroBalance_reverts() external {
        TestERC20 token = new TestERC20();
        _mockCurrencyConfig(address(token), ORACLE);

        vm.expectRevert(ICeloSequencerFeeVault.CeloSequencerFeeVault_NoTokenBalance.selector);
        sequencerFeeVault.withdrawToken(address(token), address(token));
    }

    /// @notice Sequential withdrawals accumulate `totalProcessedToken`.
    function test_withdrawToken_cumulative_succeeds() external {
        TestERC20 token = new TestERC20();
        _mockCurrencyConfig(address(token), ORACLE);

        token.mint(address(sequencerFeeVault), 100);
        sequencerFeeVault.withdrawToken(address(token));
        assertEq(sequencerFeeVault.totalProcessedToken(address(token)), 100);

        token.mint(address(sequencerFeeVault), 50);
        sequencerFeeVault.withdrawToken(address(token));
        assertEq(sequencerFeeVault.totalProcessedToken(address(token)), 150);
        assertEq(token.balanceOf(recipient), 150);
    }

    /// @notice Token withdrawal does not affect the ETH `totalProcessed` counter.
    function test_withdrawToken_doesNotAffectETHTotalProcessed() external {
        TestERC20 token = new TestERC20();
        _mockCurrencyConfig(address(token), ORACLE);
        token.mint(address(sequencerFeeVault), 1);

        uint256 ethProcessedBefore = sequencerFeeVault.totalProcessed();
        sequencerFeeVault.withdrawToken(address(token));
        assertEq(sequencerFeeVault.totalProcessed(), ethProcessedBefore);
    }

    /// @notice The one-arg overload reverts on an adapter token; use the two-arg overload instead.
    function test_withdrawToken_singleArgOnAdapter_reverts() external {
        address adapter = address(0xADAB7E5);
        vm.etch(adapter, hex"00"); // non-empty code so vm.mockCall is reachable
        _mockCurrencyConfig(adapter, ORACLE);
        vm.mockCall(adapter, abi.encodeCall(IERC20.balanceOf, (address(sequencerFeeVault))), abi.encode(uint256(1)));
        // Adapters have no `transfer`; the dispatcher reverts on the unknown selector.
        vm.mockCallRevert(adapter, abi.encodeWithSelector(IERC20.transfer.selector), "");

        vm.expectRevert();
        sequencerFeeVault.withdrawToken(adapter);
    }
}
