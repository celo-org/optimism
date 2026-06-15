// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing
import { Test } from "forge-std/Test.sol";
import { StdStorage, stdStorage } from "forge-std/StdStorage.sol";
import { Vm } from "forge-std/Vm.sol";

// Safe
import { Safe as GnosisSafe } from "safe-contracts/Safe.sol";
import { Enum } from "safe-contracts/common/Enum.sol";

/// @title CeloForkSafeExec
/// @notice Minimal Safe delegatecall helper for fork tests.
abstract contract CeloForkSafeExec is Test {
    using stdStorage for StdStorage;

    /// @notice Executes a Safe delegatecall; reverts if the inner call fails.
    function _execSafeDelegateCall(Vm _vm, address _safe, address _owner, address _target, bytes memory _data)
        internal
        returns (bool success_)
    {
        success_ = _execSafeDelegateCallAllowFail(_vm, _safe, _owner, _target, _data);
        require(success_, "CeloForkSafeExec: execTransaction failed");
    }

    /// @notice Executes a Safe delegatecall and returns success without reverting on inner failure.
    /// @dev    With safeTxGas == 0 the Safe reverts (GS013) on inner failure rather than returning false,
    ///         so the execTransaction call is wrapped to surface failure as a boolean.
    function _execSafeDelegateCallAllowFail(Vm _vm, address _safe, address _owner, address _target, bytes memory _data)
        internal
        returns (bool success_)
    {
        GnosisSafe safe = GnosisSafe(payable(_safe));
        uint256 thresholdSlot = stdstore.target(_safe).sig("getThreshold()").find();

        _vm.store(_safe, bytes32(thresholdSlot), bytes32(uint256(1)));
        require(safe.getThreshold() == 1, "CeloForkSafeExec: threshold override failed");

        bytes32 txHash = safe.getTransactionHash(
            _target, 0, _data, Enum.Operation.DelegateCall, 0, 0, 0, address(0), address(0), safe.nonce()
        );

        _vm.prank(_owner);
        safe.approveHash(txHash);

        bytes memory signature = abi.encodePacked(bytes32(uint256(uint160(_owner))), bytes32(0), uint8(1));
        try safe.execTransaction(
            _target, 0, _data, Enum.Operation.DelegateCall, 0, 0, 0, address(0), payable(address(0)), signature
        ) returns (bool ok) {
            success_ = ok;
        } catch {
            success_ = false;
        }
    }
}
