import { join } from "jsr:@std/path";
import {
  depositERC,
  initiateERC20Withdraw,
  settleWithdraw,
} from "@celo-test/viem";

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
  OnlyRunOnL2ChainIDs: [901],
})(async function (t: Deno.TestContext, ctx: Context): Promise<boolean> {
  // NOTE: important for mainnet test-runs:
  // the initial L1 balance should cover the gas fee for
  // the bridge contract interactions.
  let initialBalanceNative: ERC20Amount<BaseERC20>;
  let initialBalanceBridged: ERC20Amount<BaseERC20>;
  let bridgingAmount: bigint = parseEther("100");

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
      },
    ))
  ) {
    return false;
  }
  if (
    !(await t.step("setup test and query balances", async () => {
      initialBalanceNative = await ctx.public().l1.getERC20BalanceOf({
        erc20: bridgeTokenPair.nativeToken,
        address: ctx.wallet().l1.account!.address,
      });
      initialBalanceBridged = await ctx.public().l2.getERC20BalanceOf({
        erc20: bridgeTokenPair.bridgedToken,
        address: ctx.wallet().l2.account!.address,
      });

      expect(initialBalanceBridged.amount == BigInt(0)).toBe(true);
      expect(initialBalanceNative.amount >= bridgingAmount).toBe(true);
    }))
  ) {
    return false;
  }
  if (
    !(await t.step("deposit", async () => {
      // FIXME: approve not on token, maybe wrong token passed
      const depositResult = await depositERC(
        bridgeTokenPair.nativeToken,
        bridgeTokenPair.bridgedToken,
        bridgingAmount,
        ctx.wallet().l2.account!.address,
        ctx.public(),
        ctx.wallet(),
      );

      expect(depositResult.success).toBe(true);

      const balanceNativeAfterDeposit = await ctx
        .public()
        .l1.getERC20BalanceOf({
          erc20: bridgeTokenPair.nativeToken,
          address: ctx.wallet().l1.account!.address,
        });

      const balanceBridgedAfterDeposit = await ctx
        .public()
        .l2.getERC20BalanceOf({
          erc20: bridgeTokenPair.bridgedToken,
          address: ctx.wallet().l2.account!.address,
        });

      expect(balanceNativeAfterDeposit.amount).toBe(
        initialBalanceNative.amount - bridgingAmount,
      );
      expect(balanceBridgedAfterDeposit.amount).toBe(bridgingAmount);
      // TODO: also check gas amounts in "eth", especially on l2
    }))
  ) {
    return false;
  }
  if (
    !(await t.step("withdraw", async () => {
      const withdraw = await initiateERC20Withdraw(
        bridgingAmount,
        ctx.wallet().l1.account!.address,
        {
          nativeToken: bridgeTokenPair.bridgedToken,
          bridgedToken: bridgeTokenPair.nativeToken,
          nativeOnL1: true,
        },
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

      const balanceNativeAfterWithdraw = await ctx
        .public()
        .l1.getERC20BalanceOf({
          erc20: bridgeTokenPair.nativeToken,
          address: ctx.wallet().l1.account!.address,
        });
      const balanceBridgedAfterWithdraw = await ctx
        .public()
        .l2.getERC20BalanceOf({
          erc20: bridgeTokenPair.bridgedToken,
          address: ctx.wallet().l2.account!.address,
        });

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
