import { addTestOptions, Context } from "@celo-test/runner";
import type { ChainContractsCeloL2 } from "@celo-test/viem";
import { join } from "jsr:@std/path";

export const wethBridge = addTestOptions({
  Concurrent: true,
  Name: "test-weth-bridge",
  // only run on local devnet, since this test
  // works with hardcoded owner accounts on the devnet
  // deployment
  OnlyRunOnL2ChainIDs: [999],
})(async function (_: Deno.TestContext, ctx: Context): Promise<boolean> {
  const contractsPath = join(
    ctx.config.MonorepoPath,
    "packages/contracts-bedrock",
  );

  const l2Contracts: ChainContractsCeloL2 = ctx.public().l2.chain!
    .contracts as ChainContractsCeloL2;

  const env = {
    ETH_RPC_URL: String(ctx.config.L2.RPCURL),
    ETH_RPC_URL_L1: String(ctx.config.L1.RPCURL),
    REGISTRY_ADDR: l2Contracts.registry!.address,
    TOKEN_ADDR: l2Contracts.goldToken!.address,
    FEE_CURRENCY_DIRECTORY_ADDR: l2Contracts.feeCurrencyDirectory!.address,
    // NOTE: the FeeCurrencyDirectory owner is hardcoded to be the first devnet account
    // when the l2-genesis is generated with 'deployCeloContracts=true'.
    // so don't use the provided funded account from the Context
    // but hardcode the privkey here as well.
    // This only works for the local devnet
    ACC_PRIVKEY:
      "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
    CONTRACTS_DIR: contractsPath,
  };

  const envAsString: Record<string, string> = Object.fromEntries(
    Object.entries(env).map(([key, value]) => [key, String(value)]),
  );

  const devnetUp = new Deno.Command("bash", {
    args: ["weth_bridge.sh"],
    stdout: "piped",
    stderr: "piped",
    cwd: ctx.config.TestDirPath,
    env: envAsString,
  });

  const process = devnetUp.spawn();
  const { code, stderr } = await process.output();

  if (code !== 0) {
    const errorOutput = new TextDecoder().decode(stderr);
    throw Error(`Failed to execute test script: ${errorOutput}`);
  }
  return true;
});
