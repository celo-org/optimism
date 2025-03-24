import {
  initiateBridgeERC20To,
  settleWithdraw,
  waitForDepositReceiptL2,
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
  let initialBalanceL2Eth: bigint;

  let bridgeTokenPair: BridgedERC20TokenPair;
  const bridgingAmount: bigint = parseEther("100");
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
        ctx.storeArtifact("erc20 bridge token metadata", bridgeTokenPair);
      },
    ))
  ) {
    return false;
  }
  if (
    !(await t.step("setup test and query balances", async () => {
      initialBalanceL2Eth = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });
      ctx.storeArtifact("balance l2 eth inital", initialBalanceL2Eth);

      initialBalanceBridged = await ctx.public().l1.getERC20BalanceOf({
        erc20: bridgeTokenPair.bridgedToken,
        address: ctx.wallet().l1.account!.address,
      });
      ctx.storeArtifact(
        "balance bridged initial",
        initialBalanceBridged.amount,
      );

      initialBalanceNative = await ctx.public().l2.getERC20BalanceOf({
        erc20: bridgeTokenPair.nativeToken,
        address: ctx.wallet().l2.account!.address,
      });
      ctx.storeArtifact("balance native initial", initialBalanceNative.amount);

      expect(initialBalanceNative.amount >= BigInt(bridgingAmount)).toBe(true);
      // we just created the bridged token, nothin is there yet
      expect(initialBalanceBridged.amount == BigInt(0)).toBe(true);
    }))
  ) {
    return false;
  }
  if (
    !(await t.step("withdraw", async () => {
      const withdraw = await initiateBridgeERC20To(
        bridgingAmount,
        ctx.wallet().l1.account!.address,
        ctx.public().l2.chain, // bridge FROM l2
        bridgeTokenPair,
        ctx.public(),
        ctx.wallet(),
      );
      ctx.storeArtifact("settle withdraw result", withdraw);
      expect(withdraw.bridge.receipt).toBeDefined();
      expect(withdraw.bridge.receipt?.status === "success").toBe(true);
      withdrawResult = await settleWithdraw(
        withdraw.bridge.receipt!,
        ctx.public(),
        ctx.wallet(),
      );
      ctx.storeArtifact("settle withdraw result", withdrawResult);
      expect(withdrawResult.success).toBe(true);

      const balanceBridgedAfterWithdraw = await ctx
        .public()
        .l1.getERC20BalanceOf({
          erc20: bridgeTokenPair.bridgedToken,
          address: ctx.wallet().l1.account!.address,
        });
      ctx.storeArtifact(
        "balance bridged after withdraw",
        initialBalanceNative.amount,
      );
      const balanceNativeAfterWithdraw = await ctx
        .public()
        .l2.getERC20BalanceOf({
          erc20: bridgeTokenPair.nativeToken,
          address: ctx.wallet().l2.account!.address,
        });
      ctx.storeArtifact(
        "balance native after withdraw",
        initialBalanceNative.amount,
      );

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
    //l2 native, "deposit", so bridge back l1->l2
    !(await t.step("deposit", async () => {
      const depositResult = await initiateBridgeERC20To(
        bridgingAmount,
        ctx.wallet().l2.account!.address,
        ctx.public().l1.chain, // bridge FROM l1
        bridgeTokenPair,
        ctx.public(),
        ctx.wallet(),
      );
      ctx.storeArtifact("deposit result", depositResult);
      expect(depositResult.bridge.receipt).toBeDefined();
      // now wait for the deposit transaction to be included
      // in the l2 unsafe head
      const depositReceiptL2 = await waitForDepositReceiptL2(
        depositResult.bridge.receipt!,
        ctx.public(),
      );
      ctx.storeArtifact("deposit receipt on l2", depositReceiptL2);
      expect(depositReceiptL2.status).toBe("success");

      const balanceBridgedAfterDeposit = await ctx
        .public()
        .l1.getERC20BalanceOf({
          erc20: bridgeTokenPair.bridgedToken,
          address: ctx.wallet().l1.account!.address,
        });
      ctx.storeArtifact(
        "balance bridged after deposit",
        balanceBridgedAfterDeposit.amount,
      );

      const balanceNativeAfterDeposit = await ctx
        .public()
        .l2.getERC20BalanceOf({
          erc20: bridgeTokenPair.nativeToken,
          address: ctx.wallet().l2.account!.address,
        });
      ctx.storeArtifact(
        "balance native after deposit",
        balanceNativeAfterDeposit.amount,
      );

      const balanceL2EthAfterDeposit = await ctx.public().l2.getBalance({
        address: ctx.wallet().l2.account!.address,
      });
      ctx.storeArtifact(
        "balance eth l2 after deposit",
        balanceL2EthAfterDeposit,
      );

      // XXX: is this true?
      // we should have spent some eth for the tx gas
      // TODO: specify the exact amount expected?
      expect(balanceL2EthAfterDeposit < initialBalanceL2Eth).toBe(true);

      expect(balanceBridgedAfterDeposit.amount).toBe(0n);
      expect(balanceNativeAfterDeposit.amount).toBe(bridgingAmount);
    }))
  ) {
    return false;
  }

  return true;
});
