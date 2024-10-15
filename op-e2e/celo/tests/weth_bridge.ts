import { Context } from "@celo-test/runner";
import { join } from "jsr:@std/path";

export async function wrappedETHBridgeConcurrent(
  _: Deno.TestContext,
  ctx: Context,
) {
  const contractsPath = join(
    ctx.config.MonorepoPath,
    "packages/contracts-bedrock",
  );

  const env = {
    ETH_RPC_URL: String(ctx.config.L2.RPCURL),
    ETH_RPC_URL_L1: String(ctx.config.L1.RPCURL),
    REGISTRY_ADDR: "0x000000000000000000000000000000000000ce10",
    TOKEN_ADDR: "0x471ece3750da237f93b8e339c536989b8978a438",
    FEE_CURRENCY_DIRECTORY_ADDR: "0x9212Fb72ae65367A7c887eC4Ad9bE310BAC611BF",
    // NOTE: the FeeCurrencyDirectory owner is hardcoded to be the first devnet account
    // when the l2-genesis is generated with 'deployCeloContracts=true'.
    // so don't use the provided funded account in the context
    // but hardcode the privkey here as well
    ACC_PRIVKEY:
      "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
    CONTRACTS_DIR: contractsPath,
  };

  const envAsString: Record<string, string> = Object.fromEntries(
    Object.entries(env).map(([key, value]) => [key, String(value)]),
  );

  // TODO: port this to viem. For now just spawn a subprocess and call
  // the shell script from this deno test to benefit from the common setup/teardown
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
}
