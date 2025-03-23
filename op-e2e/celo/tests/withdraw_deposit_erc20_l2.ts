import {
  bridgeERC20To,
  initiateERC20Withdraw,
  settleWithdraw,
} from "@celo-test/viem";
import type {
  BridgedERC20TokenPair,
  WithdrawReturnType,
} from "@celo-test/viem";
import { addTestOptions, Context } from "@celo-test/runner";
import { parseEther } from "viem";
import { setupERC20BridgeToken } from "./util/bridge.ts";
import type { BaseERC20, ERC20Amount } from "reverse-mirage";
import { expect } from "jsr:@std/expect";

export const withdrawDepositERC20L2Native = addTestOptions({
  Concurrent: true,
  Name: "test-withdraw-and-deposit-back-erc20-l2",
  OnlyRunOnL2ChainIDs: [901],
})(async function (t: Deno.TestContext, ctx: Context): Promise<boolean> {
  // NOTE: important for mainnet test-runs:
  // the initial L1 balance should cover the gas fee for
  // the bridge contract interactions.

  let initialBalanceBridged: ERC20Amount<BaseERC20>;
  let initialBalanceNative: ERC20Amount<BaseERC20>;

  let bridgeTokenPair: BridgedERC20TokenPair;
  let bridgingAmount: bigint = parseEther("100");
  let withdrawResult: WithdrawReturnType;

  if (
    !(await t.step(
      "deploy ERC20 contract on l2 and create mintable representation on l1",
      async () => {
        bridgeTokenPair = await setupERC20BridgeToken(
          ctx,
          ctx.public().l2.chain,
          bridgingAmount,
        );
      },
    ))
  ) {
    return false;
  }
  if (
    !(await t.step("setup test and query balances", async () => {
      initialBalanceBridged = await ctx.public().l1.getERC20BalanceOf({
        erc20: bridgeTokenPair.bridgedToken,
        address: ctx.wallet().l1.account!.address,
      });
      initialBalanceNative = await ctx.public().l2.getERC20BalanceOf({
        erc20: bridgeTokenPair.nativeToken,
        address: ctx.wallet().l2.account!.address,
      });
      expect(initialBalanceNative.amount >= BigInt(bridgingAmount)).toBe(true);
      // we just created the bridged token, nothin is there yet
      expect(initialBalanceBridged.amount == BigInt(0)).toBe(true);
    }))
  ) {
    return false;
  }
  if (
    !(await t.step("withdraw", async () => {
      const withdraw = await initiateERC20Withdraw(
        bridgingAmount,
        ctx.wallet().l1.account!.address,
        bridgeTokenPair,
        ctx.public(),
        ctx.wallet(),
      );
      expect(withdraw.receipt.status === "success").toBe(true);
      withdrawResult = await settleWithdraw(
        withdraw.receipt,
        ctx.public(),
        ctx.wallet(),
      );
      expect(withdrawResult.success).toBe(true);

      const balanceBridgedAfterWithdraw = await ctx
        .public()
        .l1.getERC20BalanceOf({
          erc20: bridgeTokenPair.bridgedToken,
          address: ctx.wallet().l1.account!.address,
        });
      const balanceNativeAfterWithdraw = await ctx
        .public()
        .l2.getERC20BalanceOf({
          erc20: bridgeTokenPair.nativeToken,
          address: ctx.wallet().l2.account!.address,
        });

      // the full bridged amount should be on the bridged
      // token representation (on l1) now
      expect(balanceBridgedAfterWithdraw.amount).toBe(bridgingAmount);
      expect(balanceNativeAfterWithdraw.amount).toBe(
        initialBalanceNative.amount - bridgingAmount,
      );

      // TODO: also check gas currency balance,
    }))
  ) {
    return false;
  }
  if (
    !(await t.step("deposit", async () => {
      const depositResult = await bridgeERC20To(
        bridgeTokenPair.bridgedToken, // l1 token
        bridgeTokenPair.nativeToken, // l2 token
        bridgingAmount,
        ctx.wallet().l2.account!.address,
        ctx.public(),
        ctx.wallet(),
      );

      expect(depositResult.success).toBe(true);

      const balanceBridgedAfterDeposit = await ctx
        .public()
        .l1.getERC20BalanceOf({
          erc20: bridgeTokenPair.bridgedToken,
          address: ctx.wallet().l1.account!.address,
        });

      const balanceNativeAfterDeposit = await ctx
        .public()
        .l2.getERC20BalanceOf({
          erc20: bridgeTokenPair.nativeToken,
          address: ctx.wallet().l2.account!.address,
        });

      expect(balanceBridgedAfterDeposit.amount).toBe(
        initialBalanceBridged.amount - bridgingAmount,
      );
      expect(balanceNativeAfterDeposit.amount).toBe(bridgingAmount);
      // TODO: also check gas amounts in "eth", especially on l2
    }))
  ) {
    return false;
  }

  return true;
});
