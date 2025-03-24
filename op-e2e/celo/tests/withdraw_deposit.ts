import {
  deposit,
  initiateNativeWithdraw,
  settleWithdraw,
} from "@celo-test/viem";
import { addTestOptions, Context } from "@celo-test/runner";
import { parseEther } from "viem";
import type { BaseERC20, ERC20, ERC20Amount } from "reverse-mirage";
import { expect } from "jsr:@std/expect";

export const withdrawDeposit = addTestOptions({
  Concurrent: true,
  Name: "test-withdraw-and-deposit-back",
  OnlyRunOnL2ChainIDs: [901],
})(async function (t: Deno.TestContext, ctx: Context): Promise<boolean> {
  // NOTE: important for mainnet test-runs:
  // the initial L1 balance should cover the gas fee for
  // the bridge contract interactions.
  // Last time I checked locally this was around 441745 gas
  let initialBalanceL1: ERC20Amount<BaseERC20>;
  let initialBalanceL2: bigint;
  let bridgingAmount: bigint;
  let celoToken: ERC20;
  let l2GasPaid: bigint = BigInt(0);

  if (
    !(await t.step("setup test and query balances", async () => {
      celoToken = await ctx.public().l1.getERC20({
        erc20: {
          address: ctx.contracts.CustomGasTokenProxy,
          chainID: ctx.public().l1.chain!.id,
        },
      });
      ctx.storeArtifact("l1 custom gas token metadata", celoToken);
      initialBalanceL1 = await ctx.public().l1.getERC20BalanceOf({
        erc20: celoToken,
        address: ctx.wallet().l1.account!.address,
      });
      ctx.storeArtifact("balance l1 initial", initialBalanceL1.amount);

      initialBalanceL2 = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });
      ctx.storeArtifact("balance l2 initial", initialBalanceL2);
      // minimum withdraw amount
      expect(initialBalanceL2 >= parseEther("0.02")).toBe(true);
      // use half of the initial balance to account for gas cost.
      // this isn't fool proof when amounts get small,
      // but right now we don't want to calculate
      // withdraw gascost.
      bridgingAmount = initialBalanceL2 / BigInt(2);
      // maximum brdiging amount
      if (bridgingAmount > parseEther("1")) {
        bridgingAmount = parseEther("1");
      }
      ctx.storeArtifact("bridgingAmount", bridgingAmount);
    }))
  ) {
    return false;
  }
  if (
    !(await t.step("withdraw from l2", async () => {
      const withdraw = await initiateNativeWithdraw(
        bridgingAmount,
        ctx.wallet().l1.account!.address,
        21_000n,
        ctx.public(),
        ctx.wallet(),
      );
      ctx.storeArtifact("withdraw transaction from l2", withdraw);
      expect(withdraw.receipt.status === "success").toBe(true);
      const withdrawSettleResult = await settleWithdraw(
        withdraw.receipt,
        ctx.public(),
        ctx.wallet(),
      );
      ctx.storeArtifact("withdraw settle result l1", withdrawSettleResult);

      const balanceL1AfterWithdraw = await ctx.public().l1.getERC20BalanceOf({
        erc20: celoToken,
        address: ctx.wallet().l1.account!.address,
      });
      ctx.storeArtifact(
        "balance l1 after withdraw",
        balanceL1AfterWithdraw.amount,
      );
      const balanceL2AfterWithdraw = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });
      ctx.storeArtifact("balance l2 after withdraw", balanceL2AfterWithdraw);

      //FIXME: l1 gas?
      expect(balanceL1AfterWithdraw.amount).toBe(
        initialBalanceL1.amount + BigInt(bridgingAmount),
      );

      expect(balanceL2AfterWithdraw).toBe(
        initialBalanceL2 - BigInt(bridgingAmount) - withdraw.gasPaid,
      );
      l2GasPaid += withdraw.gasPaid;
    }))
  ) {
    return false;
  }

  if (
    !(await t.step("deposit back to l2", async () => {
      const depositResult = await deposit(
        bridgingAmount,
        ctx.wallet().l2.account!.address,
        ctx.public(),
        ctx.wallet(),
      );

      ctx.storeArtifact("deposit back to l2 result", depositResult);
      expect(depositResult.l1Approve.success).toBe(true);
      expect(depositResult.l1Deposit.success).toBe(true);
      expect(depositResult.l2Deposit.success).toBe(true);

      const balanceL1AfterDeposit = await ctx.public().l1.getERC20BalanceOf({
        erc20: celoToken,
        address: ctx.wallet().l1.account!.address,
      });
      ctx.storeArtifact(
        "balance l1 after deposit",
        balanceL1AfterDeposit.amount,
      );

      const balanceL2AfterDeposit = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });
      ctx.storeArtifact("balance l2 after deposit", balanceL2AfterDeposit);

      expect(balanceL1AfterDeposit.amount).toBe(initialBalanceL1.amount);
      expect(balanceL2AfterDeposit).toBe(initialBalanceL2 - l2GasPaid);
    }))
  ) {
    return false;
  }
  return true;
});
