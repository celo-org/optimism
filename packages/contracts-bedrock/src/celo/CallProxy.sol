// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.15;

contract CallProxy {
    address public implementation;

    constructor(address _implementation) {
        implementation = _implementation;
    }

    function _delegate(address _implementation) internal virtual {
        uint256 val = msg.value;
        assembly {
            // Copy msg.data. We take full control of memory in this inline assembly
            // block because it will not return to Solidity code. We overwrite the
            // Solidity scratch pad at memory position 0.
            calldatacopy(0, 0, calldatasize())

            // Call the implementation.
            // out and outsize are 0 because we don't know the size yet.
            let result := call(gas(), _implementation, val, 0, calldatasize(), 0, 0)

            // Copy the returned data.
            returndatacopy(0, 0, returndatasize())

            switch result
            // delegatecall returns 0 on error.
            case 0 {
                revert(0, returndatasize())
            }
            default {
                return(0, returndatasize())
            }
        }
    }

    function _fallback() internal virtual {
        _delegate(implementation);
    }

    fallback() external payable virtual {
        _fallback();
    }
}
