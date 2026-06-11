// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";

// Libraries
import { BaseDeployIO } from "scripts/deploy/BaseDeployIO.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";
import { CeloPredeploys } from "src/celo/CeloPredeploys.sol";

// Interfaces
import { IL1Block } from "interfaces/L2/IL1Block.sol";
import { ILiquidityController } from "interfaces/L2/ILiquidityController.sol";
import { INativeAssetLiquidity } from "interfaces/L2/INativeAssetLiquidity.sol";
import { ICeloGasBridgeL2 } from "interfaces/celo/ICeloGasBridgeL2.sol";

/// @title VerifyL2MigrationInput
/// @notice Inputs for the `VerifyL2Migration` post-migration verification script.
contract VerifyL2MigrationInput is BaseDeployIO {
    address internal _celoGasBridgeL1Proxy;

    function set(bytes4 _sel, address _value) public {
        require(_value != address(0), "VerifyL2MigrationInput: cannot set zero address");

        if (_sel == this.celoGasBridgeL1Proxy.selector) _celoGasBridgeL1Proxy = _value;
        else revert("VerifyL2MigrationInput: unknown selector");
    }

    function celoGasBridgeL1Proxy() public view returns (address) {
        require(_celoGasBridgeL1Proxy != address(0), "VerifyL2MigrationInput: celoGasBridgeL1Proxy not set");
        return _celoGasBridgeL1Proxy;
    }
}

/// @title VerifyL2Migration
/// @notice Post-migration L2 verification script for the CGT v1 -> v2 migration.
contract VerifyL2Migration is Script {
    /// @notice Verifies that the expected L2-side CGT migration wiring is present.
    function verify(VerifyL2MigrationInput _input) public view {
        console.log("--- VerifyL2Migration: L2 verification ---");

        require(
            CeloPredeploys.CELO_GAS_BRIDGE_L2.code.length > 0,
            "VerifyL2Migration: CeloGasBridgeL2 code missing"
        );
        console.log("OK  CeloGasBridgeL2 code present");

        require(
            ILiquidityController(Predeploys.LIQUIDITY_CONTROLLER).minters(CeloPredeploys.CELO_GAS_BRIDGE_L2),
            "VerifyL2Migration: bridge is not an authorized LiquidityController minter"
        );
        console.log("OK  LiquidityController minter authorization present");

        INativeAssetLiquidity liquidity = INativeAssetLiquidity(Predeploys.NATIVE_ASSET_LIQUIDITY);
        require(address(liquidity).balance > 0, "VerifyL2Migration: NativeAssetLiquidity is not funded");
        console.log("OK  NativeAssetLiquidity funded:", address(liquidity).balance);

        require(
            IL1Block(Predeploys.L1_BLOCK_ATTRIBUTES).isCustomGasToken(),
            "VerifyL2Migration: L1Block.isCustomGasToken() != true"
        );
        console.log("OK  L1Block.isCustomGasToken() == true");

        ICeloGasBridgeL2 bridge = ICeloGasBridgeL2(payable(CeloPredeploys.CELO_GAS_BRIDGE_L2));
        require(
            address(bridge.otherBridge()) == _input.celoGasBridgeL1Proxy(),
            "VerifyL2Migration: bridge otherBridge mismatch"
        );
        console.log("OK  CeloGasBridgeL2.otherBridge matches expected L1 bridge");

        console.log("L2 migration verified.");
    }
}
