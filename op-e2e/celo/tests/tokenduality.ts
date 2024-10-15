import { Context } from "@celo-test/runner";
import { createAmountFromString } from "reverse-mirage";
import { expect } from "jsr:@std/expect";

export async function tokendualityConcurrent(
  _: Deno.TestContext,
  ctx: Context,
) {
  const receiverAddr = "0x000000000000000000000000000000000000dEaD";
  const dualityToken = await ctx.public().l2.getERC20({
    erc20: {
      address: "0x471ece3750da237f93b8e339c536989b8978a438",
      chainID: ctx.public().l2.chain!.id,
    },
  });

  const balanceBefore = await ctx.public().l2.getBalance({
    address: receiverAddr,
  });

  const sendAmount = createAmountFromString(dualityToken, "100");
  const { request } = await ctx.wallet().l2.simulateERC20Transfer({
    args: {
      to: receiverAddr,
      amount: sendAmount,
    },
  });
  const transferHash = await ctx.wallet().l2.writeContract(request);
  const receipt = await ctx.public().l2.waitForTransactionReceipt({
    hash: transferHash,
    timeout: 30_000,
  });

  expect(receipt.status).toBe("success");
  const balanceAfter = await ctx.public().l2.getBalance({
    address: receiverAddr,
  });

  expect(balanceAfter).toBe(balanceBefore + sendAmount.amount);
}
