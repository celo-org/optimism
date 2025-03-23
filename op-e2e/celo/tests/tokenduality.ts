import { addTestOptions, Context } from "@celo-test/runner";
import type { ChainContractsCeloL2 } from "@celo-test/viem";
import { createAmountFromString } from "reverse-mirage";
import { expect } from "jsr:@std/expect";

export const tokenduality = addTestOptions({
  Concurrent: true,
  Name: "test-tokenduality",
  OnlyRunOnL2ChainIDs: [999],
})(async function (_: Deno.TestContext, ctx: Context): Promise<boolean> {
  const l2Contracts: ChainContractsCeloL2 = ctx.public().l2.chain!
    .contracts as ChainContractsCeloL2;
  const goldTokenAddress = l2Contracts?.goldToken?.address;
  if (goldTokenAddress === undefined) {
    throw Error("`GoldToken` address is not known");
  }

  const dualityToken = await ctx.public().l2.getERC20({
    erc20: {
      address: goldTokenAddress,
      chainID: ctx.public().l2.chain!.id,
    },
  });

  const receiverAddr = "0x000000000000000000000000000000000000dEaD";
  const balanceBefore = await ctx.public().l2.getBalance({
    address: receiverAddr,
  });

  //FIXME: only send less than the balance before, and don't specify
  // an absolute amount
  const sendAmount = createAmountFromString(dualityToken, "0.00001");
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
  console.log("token-duality tx-hash (l2)", receipt.transactionHash);

  expect(receipt.status).toBe("success");
  const balanceAfter = await ctx.public().l2.getBalance({
    address: receiverAddr,
  });

  expect(balanceAfter).toBe(balanceBefore + sendAmount.amount);
  return true;
});
