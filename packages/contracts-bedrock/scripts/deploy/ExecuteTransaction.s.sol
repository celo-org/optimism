// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";
import { LibString } from "@solady/utils/LibString.sol";

/**
 * @title ExecuteTransaction
 * @notice Script to execute transactions through a Gnosis Safe
 * @dev This script extracts the _executeTransaction functionality from UpgradeImplementations.s.sol
 *      and executes a specific transaction with predefined calldata.
 *
 * Usage:
 *   forge script scripts/deploy/ExecuteTransaction.s.sol --rpc-url <RPC_URL> --private-key <PRIVATE_KEY> --broadcast
 *
 * Requirements:
 *   - PRIVATE_KEY environment variable must be set for signing
 *   - The private key must correspond to a Gnosis Safe owner or have execution permissions
 */
contract ExecuteTransaction is Script {
    // GnosisSafe address
    address private constant _GNOSIS_SAFE = 0xd542f3328ff2516443FE4db1c89E427F67169D94;

    function run() external {
        console.log("=== EXECUTING TRANSACTION VIA GNOSIS SAFE ===");
        console.log("GnosisSafe address:", _GNOSIS_SAFE);

        // Target address and calldata provided by user
        address targetAddress = 0x93dc480940585D9961bfcEab58124fFD3d60f76a;
        bytes memory callData = hex"82ad56cb000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000d29841fbcff24eb5157f2abe7ed0b9819340159a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a4ff2dd5a1000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000010000000000000000000000003ee24bf404e4a5d27a437d910f56e1ed999b1de8000000000000000000000000bf101bd81fb69ab00ab261465454df1a171726bf03b357b30095022ecbb44ef00d1de19df39cf69ee92a60683a6be2c6f8fe6a3e00000000000000000000000000000000000000000000000000000000";

        console.log("Target address:", targetAddress);
        console.log("Calldata:", LibString.toHexString(callData));
        console.log("Calldata length:", callData.length);

        uint8 operation = 1; // Delegate call operation

        // Generate Safe transaction hash (operation 0 = CALL)
        bytes32 safeTxHash = _generateSafeTxHash(targetAddress, callData, operation);

        // Sign the transaction
        bytes memory signature = _signTransaction(safeTxHash);

        // Execute the transaction through Gnosis Safe
        _executeTransaction(targetAddress, callData, signature, operation);

        console.log("=== TRANSACTION EXECUTION COMPLETE ===");
    }

    /// @notice Generate Safe transaction hash
    function _generateSafeTxHash(
        address targetAddress,
        bytes memory callData,
        uint8 operation
    ) internal view returns (bytes32) {
        console.log("Generating Safe transaction hash...");
        console.log("Transaction data length:", callData.length);

        (bool success, bytes memory result) = _GNOSIS_SAFE.staticcall(
            abi.encodeWithSignature("nonce()")
        );
        uint256 nonce = success ? abi.decode(result, (uint256)) : 0;
        console.log("Safe nonce:", nonce);

        (success, result) = _GNOSIS_SAFE.staticcall(
            abi.encodeWithSignature("domainSeparator()")
        );
        bytes32 domainSeparator = success ? abi.decode(result, (bytes32)) : bytes32(0);
        console.log("Domain separator:", LibString.toHexString(abi.encodePacked(domainSeparator)));

        bytes32 safeTxHash = keccak256(
            abi.encodePacked(
                bytes1(0x19),
                bytes1(0x01),
                domainSeparator,
                keccak256(
                    abi.encode(
                        keccak256("SafeTx(address to,uint256 value,bytes data,uint8 operation,uint256 safeTxGas,uint256 baseGas,uint256 gasPrice,address gasToken,address refundReceiver,uint256 nonce)"),
                        targetAddress,
                        0, // value
                        keccak256(callData),
                        operation,
                        0, // safeTxGas
                        0, // baseGas
                        0, // gasPrice
                        address(0), // gasToken
                        address(0), // refundReceiver
                        nonce
                    )
                )
            )
        );

        console.log("Safe transaction hash:", LibString.toHexString(abi.encodePacked(safeTxHash)));
        return safeTxHash;
    }

    /// @notice Sign the transaction hash
    function _signTransaction(bytes32 safeTxHash) internal view returns (bytes memory) {
        console.log("Signing transaction...");
        uint256 privateKey = vm.envUint("PRIVATE_KEY");
        require(privateKey != 0, "Failed to parse PRIVATE_KEY, or it resolved to zero. Ensure it's a valid hex string (e.g., 0x...).");

        console.log("address of privateKey:", vm.addr(privateKey));

        console.log("Private key for Gnosis Safe message signing loaded successfully from PRIVATE_KEY.");

        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, safeTxHash);

        bytes memory signature = abi.encodePacked(r, s, v);
        console.log("Signature for Gnosis Safe message:", LibString.toHexString(signature));

        return signature;
    }

    /// @notice Execute the transaction through Gnosis Safe
    function _executeTransaction(
        address targetAddress,
        bytes memory callData,
        bytes memory signature,
        uint8 operation
    ) internal {
        console.log("Executing transaction through Gnosis Safe...");

        // Verify the Safe contract exists at the expected address
        require(_GNOSIS_SAFE.code.length > 0, "Gnosis Safe contract not found at expected address");

        vm.startBroadcast();

        (bool execSuccess, bytes memory returnData) = _GNOSIS_SAFE.call(
            abi.encodeWithSignature(
                "execTransaction(address,uint256,bytes,uint8,uint256,uint256,uint256,address,address,bytes)",
                targetAddress,
                0, // value
                callData,
                operation,
                0,  // safeTxGas
                0,  // baseGas
                0,  // gasPrice
                address(0),  // gasToken
                address(0),  // refundReceiver
                signature
            )
        );

        vm.stopBroadcast();

        if (execSuccess) {
            console.log("Successfully executed transaction via Gnosis Safe on-chain!");
            console.log("Transaction has been broadcast and will be mined in the next block.");
        } else {
            console.log("Failed to execute transaction through Gnosis Safe");
            console.log("Error data:", LibString.toHexString(returnData));
            console.log("This may require proper multisig threshold signatures");

            // Attempt to decode common error reasons
            if (returnData.length >= 4) {
                bytes4 errorSelector = bytes4(returnData);
                if (errorSelector == bytes4(keccak256("GS013"))) {
                    console.log("Error: Invalid signature provided");
                } else if (errorSelector == bytes4(keccak256("GS020"))) {
                    console.log("Error: Signatures data too short");
                }
            }
        }

        console.log("=== TRANSACTION EXECUTION COMPLETE ===");
    }
}
