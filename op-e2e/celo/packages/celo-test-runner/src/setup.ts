import { join } from "jsr:@std/path";
import { ClientAccountManager } from "@celo-test/viem";
import type { Config } from "./types.ts";
import { Context } from "./context.ts";
import {
  waitClientNotSyncing,
  waitClientReturnsBlockNum,
  waitInitialGame,
  waitInitialL2OracleOutput,
} from "@celo-test/util";
import { setupDevnet, teardownDevnet } from "./devnet.ts";
import { type ContractAddresses, makeChainConfigs } from "@celo-test/viem";

const addressesFilePath = ".devnet/addresses.json";
const celoTestSeedPhrase =
  "test test test test test test test test test test test celery";

export async function setup(
  _: Deno.TestContext,
  config: Config,
): Promise<Context> {
  await setupDevnet(config);
  const addressesJson = Deno.readTextFileSync(
    join(config.MonorepoPath, addressesFilePath),
  );
  const addresses: ContractAddresses = JSON.parse(addressesJson);
  const chains = makeChainConfigs(
    config.L1.ChainID,
    config.L2.ChainID,
    config.L1.RPCURL.toString(),
    config.L2.RPCURL.toString(),
    addresses,
  );

  // the num accounts will be reset later, when we know
  // how many concurrently sending accounts we need.
  // For now, we don't even use the wallet clients.
  const clients = new ClientAccountManager(chains, celoTestSeedPhrase, 1);
  const publicClient = clients.public();
  //TODO: this should check the configuration
  // against what's running on the devnet.
  // Because when the devnet is not spawned by the
  // setup(), we can run into a misconfiguration
  // between test and network...
  if (!config.UseFaultproofSystem) {
    console.log("fault-proofs disabled, using L2OO");
  }
  console.log("waiting for devnet to stabilize");
  const success = await Promise.all([
    waitClientReturnsBlockNum(publicClient.l1, 20),
    waitClientReturnsBlockNum(publicClient.l2, 20),
    waitClientNotSyncing(publicClient.l1, 15),
    waitClientNotSyncing(publicClient.l2, 15),
    config.UseFaultproofSystem
      ? waitInitialGame(publicClient, 120)
      : waitInitialL2OracleOutput(publicClient, 120),
  ]);
  console.log("devnet stable");
  const ctx = new Context(clients, config, undefined, false, addresses);
  if (success.every((v) => v === null)) {
    return ctx;
  }
  throw new Error("l1 and l2 clients not reachable within the deadline");
}

export async function teardown(_: Deno.TestContext, config: Config) {
  await teardownDevnet(config);
}
