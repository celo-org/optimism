// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

// Contracts
import { FeeVault } from "src/L2/FeeVault.sol";

// Libraries
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
/// @notice Celo L2 sequencer fee vault; replaces SequencerFeeVault and also withdraws ERC-20 fees.
/// @dev    ERC-20 path is sequencer-only and skips the minimum withdrawal check.
contract CeloSequencerFeeVault is FeeVault, ISemver {
    using SafeERC20 for IERC20;

    /// @custom:semver 1.6.0-celo
    string public constant version = "1.6.0-celo";

    /// @notice Registry id used to resolve the FeeCurrencyDirectory address.
    bytes32 internal constant FEE_CURRENCY_DIRECTORY_REGISTRY_ID = keccak256(abi.encodePacked("FeeCurrencyDirectory"));

    /// @notice Total of each ERC-20 withdrawn, keyed by the transferred token.
    mapping(address => uint256) public totalProcessedToken;

    /// @notice Emitted when an ERC-20 fee-currency balance is withdrawn.
    /// @param registered Address registered in the FeeCurrencyDirectory.
    /// @param actual     ERC-20 token actually transferred.
    /// @param value      Amount transferred.
    /// @param to         Recipient of the transfer.
    /// @param from       Caller that triggered the withdrawal.
    event TokenWithdrawal(
        address indexed registered, address indexed actual, uint256 value, address indexed to, address from
    );

    /// @notice Thrown when the token is not a registered fee currency.
    error CeloSequencerFeeVault_NotRegisteredFeeCurrency();

    /// @notice Thrown when the adapter's underlying token does not match `_actual`.
    error CeloSequencerFeeVault_AdapterMismatch();

    /// @notice Thrown when the vault holds no balance of the requested token.
    error CeloSequencerFeeVault_NoTokenBalance();

    /// @custom:legacy
    /// @notice Legacy getter for the recipient address.
    /// @return The recipient address.
    function l1FeeWallet() public view returns (address) {
        return recipient;
    }

    /// @notice Withdraws the full balance of a native fee-currency ERC-20 to the recipient.
    /// @param _token Fee-currency token registered in the FeeCurrencyDirectory.
    function withdrawToken(address _token) external {
        withdrawToken(_token, _token);
    }

    /// @notice Withdraws the full ERC-20 balance to the recipient, validated via the FeeCurrencyDirectory.
    /// @param _registered Address registered in the FeeCurrencyDirectory (native token or adapter).
    /// @param _actual     Token to transfer: equals `_registered`, or its adapter's `getAdaptedToken()`.
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

        emit TokenWithdrawal(_registered, _actual, value, recipient, msg.sender);

        IERC20(_actual).safeTransfer(recipient, value);
    }
}
