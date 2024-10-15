import { deposit, withdraw, type WithdrawReturnType } from "@celo-test/viem";
import { Context } from "@celo-test/runner";
import { parseEther } from "viem";
import type { BaseERC20, ERC20, ERC20Amount } from "reverse-mirage";
import { expect } from "jsr:@std/expect";

export async function withdrawDepositConcurrent(
  t: Deno.TestContext,
  ctx: Context,
) {
  const bridgingAmount = parseEther("1");

  let initialBalanceL1: ERC20Amount<BaseERC20>;
  let initialBalanceL2: bigint;
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
    }))
  ) {
    return;
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
    return;
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
    return;
  }
}
