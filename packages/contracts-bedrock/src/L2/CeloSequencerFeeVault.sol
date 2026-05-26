// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Contracts
import { FeeVault } from "src/L2/FeeVault.sol";

// Libraries
import { Types } from "src/libraries/Types.sol";
import { CeloPredeploys } from "src/celo/CeloPredeploys.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

// Interfaces
import { ISemver } from "interfaces/universal/ISemver.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ICeloRegistry } from "src/celo/interfaces/ICeloRegistry.sol";
import { IFeeCurrencyDirectory } from "src/celo/interfaces/IFeeCurrencyDirectory.sol";
import { IFeeCurrencyAdapter } from "src/celo/interfaces/IFeeCurrencyAdapter.sol";

/// @custom:proxied true
/// @custom:predeploy 0x4200000000000000000000000000000000000011
/// @title CeloSequencerFeeVault
/// @notice The CeloSequencerFeeVault is the contract that holds any fees paid to the Sequencer during
///         transaction processing and block production.
///         This is meant to be deployed on the Celo L2 network instead of the default
///         SequencerFeeVault. This version of the fee vault additionally handles fees paid in
///         ERC-20 tokens.
/// @dev    Some assumptions are made to keep the implementation simple:
///         1. None of the other FeeVaults will be useful in the foreseeable future, so the changes
///            are introduced only for the SequencerFeeVault.
///         2. Withdrawals will only be made to an L2 address.
///         3. The minimum withdrawal amount is not important, and is not enforced for ERC-20
///            tokens.
contract CeloSequencerFeeVault is FeeVault, ISemver {
    using SafeERC20 for IERC20;

    /// @custom:semver 1.0.0
    string public constant version = "1.0.0";

    /// @notice Registry identifier hash used to look up the FeeCurrencyDirectory address.
    bytes32 internal constant FEE_CURRENCY_DIRECTORY_REGISTRY_ID = keccak256(abi.encodePacked("FeeCurrencyDirectory"));

    /// @notice Total amount of each ERC-20 token transferred out of the vault.
    ///         Keyed by the actually-transferred token address (the underlying for adapter
    ///         fee currencies, or the registered address itself for native fee currencies).
    mapping(address => uint256) public totalProcessedToken;

    /// @notice Emitted when an ERC-20 fee-currency balance is withdrawn.
    /// @param registered The address registered in the FeeCurrencyDirectory.
    /// @param actual     The ERC-20 token address that was actually transferred.
    /// @param value      Amount transferred (in `actual` token's native units).
    /// @param to         Recipient of the transfer.
    /// @param from       Address that triggered the withdrawal.
    event TokenWithdrawal(
        address indexed registered, address indexed actual, uint256 value, address indexed to, address from
    );

    /// @notice Thrown when the supplied token is not a registered fee currency.
    error CeloSequencerFeeVault_NotRegisteredFeeCurrency();

    /// @notice Thrown when the underlying token of an adapter does not match the address passed as
    ///         `_actual`.
    error CeloSequencerFeeVault_AdapterMismatch();

    /// @notice Thrown when the vault holds no balance of the requested token.
    error CeloSequencerFeeVault_NoTokenBalance();

    /// @notice Constructs the CeloSequencerFeeVault contract.
    /// @param _recipient           Wallet that will receive the fees.
    /// @param _minWithdrawalAmount Minimum balance for withdrawals.
    /// @param _withdrawalNetwork   Network which the recipient will receive fees on.
    constructor(
        address _recipient,
        uint256 _minWithdrawalAmount,
        Types.WithdrawalNetwork _withdrawalNetwork
    )
        FeeVault(_recipient, _minWithdrawalAmount, _withdrawalNetwork)
    { }

    /// @custom:legacy
    /// @notice Legacy getter for the recipient address.
    /// @return The recipient address.
    function l1FeeWallet() public view returns (address) {
        return RECIPIENT;
    }

    /// @notice Withdraws the vault's full balance of a native fee-currency ERC-20 to RECIPIENT.
    ///         Convenience overload for native fee currencies, where the directory-registered
    ///         address is itself the transferable ERC-20.
    /// @param _token The fee-currency token registered in the FeeCurrencyDirectory.
    function withdrawToken(address _token) external {
        withdrawToken(_token, _token);
    }

    /// @notice Withdraws the vault's full balance of an ERC-20 to RECIPIENT, validated against
    ///         the FeeCurrencyDirectory. For adapter-wrapped fee currencies (e.g. USDC), pass the
    ///         registered adapter as `_registered` and the underlying ERC-20 as `_actual`.
    /// @param _registered Address registered in the FeeCurrencyDirectory (native token or adapter).
    /// @param _actual     ERC-20 token to actually transfer. Must equal `_registered` (native case)
    ///                    or `IFeeCurrencyAdapter(_registered).getAdaptedToken()` (adapter case).
    function withdrawToken(address _registered, address _actual) public {
        address directory =
            ICeloRegistry(CeloPredeploys.CELO_REGISTRY).getAddressForOrDie(FEE_CURRENCY_DIRECTORY_REGISTRY_ID);
        IFeeCurrencyDirectory.CurrencyConfig memory cfg =
            IFeeCurrencyDirectory(directory).getCurrencyConfig(_registered);
        if (cfg.oracle == address(0)) revert CeloSequencerFeeVault_NotRegisteredFeeCurrency();

        if (_actual != _registered && IFeeCurrencyAdapter(_registered).getAdaptedToken() != _actual) {
            revert CeloSequencerFeeVault_AdapterMismatch();
        }

        uint256 value = IERC20(_actual).balanceOf(address(this));
        if (value == 0) revert CeloSequencerFeeVault_NoTokenBalance();

        totalProcessedToken[_actual] += value;

        emit TokenWithdrawal(_registered, _actual, value, RECIPIENT, msg.sender);

        IERC20(_actual).safeTransfer(RECIPIENT, value);
    }
}
