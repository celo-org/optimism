import { deposit, withdraw, type WithdrawReturnType } from "@celo-test/viem";
import { addTestOptions, Context } from "@celo-test/runner";
import { parseEther } from "viem";
import type { BaseERC20, ERC20, ERC20Amount } from "reverse-mirage";
import { expect } from "jsr:@std/expect";

export const withdrawDeposit = addTestOptions({
  Concurrent: true,
  Name: "test-withdraw-and-deposit-back",
  OnlyRunOnL2ChainIDs: undefined,
})(async function (t: Deno.TestContext, ctx: Context): Promise<boolean> {
  // NOTE: important for mainnet test-runs:
  // the initial L1 balance should cover the gas fee for
  // the bridge contract interactions.
  // Last time I checked locally this was around 441745 gas
  let initialBalanceL1: ERC20Amount<BaseERC20>;
  let initialBalanceL2: bigint;
  let bridgingAmount: bigint;
  let celoToken: ERC20;
  let withdrawResult: WithdrawReturnType;

  if (
    !(await t.step("setup test and query balances", async () => {
      celoToken = await ctx.public().l1.getERC20({
        erc20: {
          address: ctx.contracts.CustomGasTokenProxy,
          chainID: ctx.public().l1.chain!.id,
        },
      });
      initialBalanceL1 = await ctx.public().l1.getERC20BalanceOf({
        erc20: celoToken,
        address: ctx.wallet().l1.account!.address,
      });

      initialBalanceL2 = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });
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
    }))
  ) {
    return false;
  }
  if (
    !(await t.step("withdraw", async () => {
      withdrawResult = await withdraw(
        bridgingAmount,
        ctx.wallet().l1.account!.address,
        21_000n,
        ctx.public(),
        ctx.wallet(),
      );
      expect(withdrawResult.success).toBe(true);

      const balanceL1AfterWithdraw = await ctx.public().l1.getERC20BalanceOf({
        erc20: celoToken,
        address: ctx.wallet().l1.account!.address,
      });
      const balanceL2AfterWithdraw = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });

      expect(balanceL1AfterWithdraw.amount).toBe(
        initialBalanceL1.amount + BigInt(bridgingAmount),
      );
      expect(balanceL2AfterWithdraw).toBe(
        initialBalanceL2 - BigInt(bridgingAmount) - withdrawResult.l2GasPayment,
      );
    }))
  ) {
    return false;
  }

  if (
    !(await t.step("deposit", async () => {
      const depositResult = await deposit(
        bridgingAmount,
        ctx.wallet().l2.account!.address,
        ctx.public(),
        ctx.wallet(),
      );

      expect(depositResult.success).toBe(true);

      const balanceL1AfterDeposit = await ctx.public().l1.getERC20BalanceOf({
        erc20: celoToken,
        address: ctx.wallet().l1.account!.address,
      });

      const balanceL2AfterDeposit = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });

      expect(balanceL1AfterDeposit.amount).toBe(initialBalanceL1.amount);
      expect(balanceL2AfterDeposit).toBe(
        initialBalanceL2 - withdrawResult.l2GasPayment,
      );
    }))
  ) {
    return false;
  }
  return true;
});
