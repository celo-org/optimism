// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Contracts
import { FeeVault } from "src/L2/FeeVault.sol";

// Libraries
import { Types } from "src/libraries/Types.sol";

// Interfaces
import { ISemver } from "interfaces/universal/ISemver.sol";

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
    /// @custom:semver 1.0.0
    string public constant version = "1.0.0";

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
}
