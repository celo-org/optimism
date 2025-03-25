import {
  initiateBridgeERC20To,
  settleWithdraw,
  waitForDepositReceiptL2,
} from "@celo-test/viem";
import { sleep } from "@celo-test/util";
import { setupERC20BridgeToken } from "./util/bridge.ts";
import type {
  BridgedERC20TokenPair,
  WithdrawReturnType,
} from "@celo-test/viem";
import { addTestOptions, Context } from "@celo-test/runner";
import { parseEther } from "viem";
import type { BaseERC20, ERC20Amount } from "reverse-mirage";
import { expect } from "jsr:@std/expect";

export const withdrawDepositERC20L1Native = addTestOptions({
  Concurrent: true,
  Name: "test-deposit-and-withdraw-back-erc20-l1",
  OnlyRunOnL2ChainIDs: undefined,
})(async function (t: Deno.TestContext, ctx: Context): Promise<boolean> {
  // NOTE: important for mainnet test-runs:
  // the initial L1 balance should cover the gas fee for
  // the bridge contract interactions.
  let initialBalanceNative: ERC20Amount<BaseERC20>;
  let initialBalanceBridged: ERC20Amount<BaseERC20>;
  let initialBalanceL2Eth: bigint;
  const bridgingAmount: bigint = parseEther("10");

  let bridgeTokenPair: BridgedERC20TokenPair;
  let withdrawResult: WithdrawReturnType;

  if (
    !(await t.step(
      "deploy ERC20 contract on l1 and create mintable representation on l2",
      async () => {
        bridgeTokenPair = await setupERC20BridgeToken(
          ctx,
          ctx.public().l1.chain,
          bridgingAmount,
        );
        ctx.storeArtifact("erc20 bridge token metadata", bridgeTokenPair);
      },
    ))
  ) {
    return false;
  }
  if (
    !(await t.step("setup test and query balances", async () => {
      await sleep(10000);
      initialBalanceL2Eth = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });
      ctx.storeArtifact("balance l2 initial", initialBalanceL2Eth);

      initialBalanceNative = await ctx.public().l1.getERC20BalanceOf({
        erc20: bridgeTokenPair.nativeToken,
        address: ctx.wallet().l1.account!.address,
      });
      ctx.storeArtifact("balance native initial", initialBalanceNative.amount);
      initialBalanceBridged = await ctx.public().l2.getERC20BalanceOf({
        erc20: bridgeTokenPair.bridgedToken,
        address: ctx.wallet().l2.account!.address,
      });
      ctx.storeArtifact(
        "balance bridged initial",
        initialBalanceBridged.amount,
      );

      expect(initialBalanceBridged.amount == BigInt(0)).toBe(true);
      expect(initialBalanceNative.amount >= bridgingAmount).toBe(true);
    }))
  ) {
    return false;
  }
  if (
    !(await t.step("deposit", async () => {
      const deposit = await initiateBridgeERC20To(
        bridgingAmount,
        ctx.wallet().l2.account!.address,
        ctx.public().l1.chain, // bridge FROM l1
        bridgeTokenPair,
        ctx.public(),
        ctx.wallet(),
      );
      ctx.storeArtifact("deposit result", deposit);
      expect(deposit.bridge.receipt).toBeDefined();
      expect(deposit.bridge.receipt?.status === "success").toBe(true);

      const depositReceiptL2 = await waitForDepositReceiptL2(
        deposit.bridge.receipt!,
        ctx.public(),
      );
      ctx.storeArtifact("deposit receipt on l2", depositReceiptL2);
      expect(depositReceiptL2.status === "success").toBe(true);
      await sleep(10000);

      const balanceNativeAfterDeposit = await ctx
        .public()
        .l1.getERC20BalanceOf({
          erc20: bridgeTokenPair.nativeToken,
          address: ctx.wallet().l1.account!.address,
        });
      ctx.storeArtifact(
        "balance native after deposit",
        balanceNativeAfterDeposit.amount,
      );

      const balanceBridgedAfterDeposit = await ctx
        .public()
        .l2.getERC20BalanceOf({
          erc20: bridgeTokenPair.bridgedToken,
          address: ctx.wallet().l2.account!.address,
        });
      ctx.storeArtifact(
        "balance bridged after deposit",
        balanceBridgedAfterDeposit.amount,
      );
      const balanceL2EthAfterDeposit = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });
      ctx.storeArtifact("balance l2 eth initial", balanceL2EthAfterDeposit);

      expect(balanceNativeAfterDeposit.amount).toBe(
        initialBalanceNative.amount - bridgingAmount,
      );
      //FIXME: this fails, as the amount is 0
      // maybe this is not deposited yet??
      expect(balanceBridgedAfterDeposit.amount).toBe(bridgingAmount);
      // TODO: do we know the exact expected amount?
      // FIXME: this is not true in the test, is this correct?
      // expect(balanceL2EthAfterDeposit < initialBalanceL2Eth).toBe(true);
    }))
  ) {
    return false;
  }
  if (
    !(await t.step("withdraw", async () => {
      const withdraw = await initiateBridgeERC20To(
        bridgingAmount,
        ctx.wallet().l1.account!.address,
        ctx.public().l2.chain, // bridge FROM L2
        bridgeTokenPair,
        ctx.public(),
        ctx.wallet(),
      );
      ctx.storeArtifact("withdraw result", withdraw);
      expect(withdraw.bridge.receipt).toBeDefined();
      expect(withdraw.bridge.receipt?.status === "success").toBe(true);

      await sleep(10000);
      withdrawResult = await settleWithdraw(
        withdraw.bridge.receipt!,
        ctx.public(),
        ctx.wallet(),
      );
      ctx.storeArtifact("settle withdraw result", withdrawResult);
      expect(withdrawResult.finalize.success).toBe(true);
      expect(withdrawResult.prove.success).toBe(true);

      await sleep(10000);
      const balanceNativeAfterWithdraw = await ctx
        .public()
        .l1.getERC20BalanceOf({
          erc20: bridgeTokenPair.nativeToken,
          address: ctx.wallet().l1.account!.address,
        });
      ctx.storeArtifact(
        "balance native after withdraw",
        balanceNativeAfterWithdraw.amount,
      );
      const balanceBridgedAfterWithdraw = await ctx
        .public()
        .l2.getERC20BalanceOf({
          erc20: bridgeTokenPair.bridgedToken,
          address: ctx.wallet().l2.account!.address,
        });
      ctx.storeArtifact(
        "balance bridged after withdraw",
        balanceBridgedAfterWithdraw.amount,
      );

      const balanceL2EthAfterWithdraw = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });
      ctx.storeArtifact(
        "balance eth l2 after withdraw",
        balanceL2EthAfterWithdraw,
      );

      // XXX: is this true?
      // we should have spent some eth for the tx gas
      expect(balanceL2EthAfterWithdraw < initialBalanceL2Eth).toBe(true);

      // since we don't pay gas on either network in the
      // bridged token, we should be at the same balances as before
      expect(balanceNativeAfterWithdraw.amount).toBe(
        initialBalanceNative.amount,
      );
      expect(balanceBridgedAfterWithdraw.amount).toBe(0n);

      // TODO: also check gas currency balance,
    }))
  ) {
    return false;
  }

  return true;
});
