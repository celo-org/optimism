// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";

import { Multicall3Delegatecall } from "scripts/cgt/Multicall3Delegatecall.sol";

/// @title DeployMulticall3Delegatecall
/// @notice Fallback predeploy of Multicall3Delegatecall for networks that lack it, so the CGT v2
///         migration can be signed as one atomic Safe transaction there too.
/// @notice Not needed for any currently planned network: the canonical address already has code on
///         both Ethereum mainnet and Sepolia, covering Celo mainnet, Celo Sepolia and Havoc. Where
///         it is genuinely absent, pass the deployed address as MULTICALL3_DELEGATECALL when
///         building the bundle, since a fresh deployment will not land on the canonical address.
/// @notice scripts/cgt/Multicall3Delegatecall.sol is a byte-identical copy of superchain-ops'
///         src/Multicall3Delegatecall.sol, kept unmodified so it stays trivially diffable against
///         upstream. Do not edit it; put any local notes here instead.
/// @notice Building that copy here does NOT reproduce the deployed bytecode: same solc 0.8.15, but
///         this repo sets bytecode_hash = "none" with its own optimizer settings, giving 1222 bytes
///         against the deployed 1753. Harmless, since both target L1s already have a real
///         deployment and this script refuses to deploy over the canonical address.
contract DeployMulticall3Delegatecall is Script {
    /// @notice Canonical deployment address, reused as the default by MigrateV1V2.
    address internal constant CANONICAL = 0x93dc480940585D9961bfcEab58124fFD3d60f76a;

    function run() external {
        if (CANONICAL.code.length != 0) {
            console.log("Multicall3Delegatecall already present at canonical address:", CANONICAL);
            console.log("Nothing to deploy. Omit MULTICALL3_DELEGATECALL to use it.");
            return;
        }

        vm.broadcast();
        Multicall3Delegatecall deployed = new Multicall3Delegatecall();

        require(address(deployed).code.length != 0, "DeployMulticall3Delegatecall: no code at deployment");
        console.log("Multicall3Delegatecall deployed at:", address(deployed));
        console.log("Export this for the atomic bundle:");
        console.log("  MULTICALL3_DELEGATECALL=", address(deployed));
    }
}
