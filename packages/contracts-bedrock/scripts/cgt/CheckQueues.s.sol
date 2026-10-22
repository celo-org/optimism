// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import { Script } from "forge-std/Script.sol";
import { Vm } from "forge-std/Vm.sol";
import { console2 as console } from "forge-std/console2.sol";

import { IOptimismPortal2 } from "interfaces/L1/IOptimismPortal2.sol";
import { IL1Block } from "interfaces/L2/IL1Block.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";

/// @title CheckQueues
/// @notice Asserts both bridging queues are empty before the CGT v1 -> v2 migration.
/// @notice A withdrawal proven under v1 but not finalized can never be paid out afterwards: v6
///         reverts OptimismPortal_NotAllowedOnCGTMode on any withdrawal carrying value. A deposit
///         still being derived mints CELO on L2 after the L1 snapshot is taken, which leaves
///         NativeAssetLiquidity funded from a stale number.
/// @notice Lives outside MigrateV1V2.preflight() because it needs L1 logs and L2 state, and an L1
///         view function can read neither.
contract CheckQueues is Script {
    bytes32 internal constant WITHDRAWAL_PROVEN = keccak256("WithdrawalProven(bytes32,address,address)");
    bytes32 internal constant TRANSACTION_DEPOSITED = keccak256("TransactionDeposited(address,address,uint256,bytes)");

    function run() external {
        address portal = vm.envAddress("PORTAL_PROXY");
        uint256 fromBlock = vm.envUint("FROM_BLOCK");
        string memory l1Rpc = vm.envString("L1_RPC_URL");
        string memory l2Rpc = vm.envString("L2_RPC_URL");

        console.log("--- CheckQueues ---");
        console.log("portal    :", portal);
        console.log("fromBlock :", fromBlock);

        vm.createSelectFork(l1Rpc);
        uint256 unfinalized = _checkWithdrawals(portal, fromBlock);
        uint256 lastDepositBlock = _lastDepositBlock(portal, fromBlock);

        vm.createSelectFork(l2Rpc);
        uint256 l1Origin = IL1Block(Predeploys.L1_BLOCK_ATTRIBUTES).number();
        bool depositsDerived = l1Origin >= lastDepositBlock;
        console.log("L2 L1Block.number      :", l1Origin);
        console.log("last deposit L1 block  :", lastDepositBlock);

        if (depositsDerived) {
            console.log("OK  deposit queue empty, L2 has derived past the last deposit");
        } else {
            console.log("FAIL deposit queue not empty. L1 blocks still undrained:", lastDepositBlock - l1Origin);
        }

        require(unfinalized == 0, "CheckQueues: withdrawal queue not empty");
        require(depositsDerived, "CheckQueues: deposit queue not empty");
        console.log("Both queues empty, safe to migrate");
    }

    /// @notice Counts withdrawals proven on L1 that have not been finalized.
    // TODO: chunk the scan. Providers cap eth_getLogs ranges, so a mainnet portal spanning millions
    // of blocks needs a fromBlock..head loop rather than one call. Note a short FROM_BLOCK fails
    // silently with an empty result, unlike a capped range which errors.
    function _checkWithdrawals(address _portal, uint256 _fromBlock) internal returns (uint256 unfinalized_) {
        bytes32[] memory topics = new bytes32[](1);
        topics[0] = WITHDRAWAL_PROVEN;
        Vm.EthGetLogs[] memory proven = vm.eth_getLogs(_fromBlock, block.number, _portal, topics);

        for (uint256 i = 0; i < proven.length; i++) {
            bytes32 hash = proven[i].topics[1];
            if (!IOptimismPortal2(payable(_portal)).finalizedWithdrawals(hash)) {
                unfinalized_++;
                console.log("PENDING withdrawal:");
                console.logBytes32(hash);
            }
        }

        console.log("withdrawals proven     :", proven.length);
        console.log("withdrawals unfinalized:", unfinalized_);
        if (unfinalized_ == 0) console.log("OK  withdrawal queue empty");
    }

    /// @notice L1 block of the most recent deposit, or 0 if there are none.
    function _lastDepositBlock(address _portal, uint256 _fromBlock) internal returns (uint256) {
        bytes32[] memory topics = new bytes32[](1);
        topics[0] = TRANSACTION_DEPOSITED;
        Vm.EthGetLogs[] memory deposits = vm.eth_getLogs(_fromBlock, block.number, _portal, topics);
        console.log("deposits found         :", deposits.length);
        return deposits.length == 0 ? 0 : deposits[deposits.length - 1].blockNumber;
    }
}
