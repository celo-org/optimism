import { join } from "jsr:@std/path";
import { ClientAccountManager } from "@celo-test/viem";
import type { Config } from "./types.ts";
import { Context } from "./context.ts";
import { TestLogger } from "./logger.ts";
import {
  waitClientNotSyncing,
  waitClientReturnsBlockNum,
  waitInitialL2OracleOutput,
  waitUntilTwoGames,
} from "@celo-test/util";
import { setupDevnet, teardownDevnet } from "./devnet.ts";
import { makeChainConfigs } from "@celo-test/viem";
import { isHex } from "viem";
import { getPortalVersion } from "viem/op-stack";
import type {
  ChainContractsCeloL2,
  ContractAddressesL1,
} from "@celo-test/viem";

type CeloCLIContractInfo = {
  contract: string;
  proxy: string;
  implementation: string;
  version: string;
};

type CeloCLIContractsList = CeloCLIContractInfo[];

// uses the celocli "network:contracts" --json formatted
// contract spec and converts it into something
// that can be injected into viem's "Chain" config.
function readCeloCLIAddresses(config: Config): ChainContractsCeloL2 {
  if (config.ContractAddressesL2FilePath === undefined) {
    return {};
  }
  const addressesJson = Deno.readTextFileSync(
    join(config.ContractAddressesL2FilePath),
  );

  const addresses: CeloCLIContractsList = JSON.parse(addressesJson);
  const contractAddressesL2Celo = addresses.reduce(
    (acc: ChainContractsCeloL2, contract: CeloCLIContractInfo) => {
      let contractName = contract.contract;
      if (!contractName) {
        return acc;
      }
      // lower-case the first char of the contract name,
      // since this is how it is standard in viem
      contractName =
        contractName.charAt(0).toLowerCase() + contractName.slice(1);
      if (!isHex(contract.proxy)) {
        throw Error(
          `provided Celo contract proxy address of contract '${contract.contract}' is not Hex formatted: ${contract.proxy}`,
        );
      }
      acc[contractName] = { address: contract.proxy };
      return acc;
    },
    {},
  );
  return contractAddressesL2Celo;
}

export async function setup(
  _: Deno.TestContext,
  config: Config,
): Promise<Context> {
  await setupDevnet(config);
  const addressesJson = Deno.readTextFileSync(
    join(config.ContractAddressesL1FilePath),
  );
  const addressesL1: ContractAddressesL1 = JSON.parse(addressesJson);
  const addressesL2Celo = readCeloCLIAddresses(config);

  const chains = makeChainConfigs(
    config.L1.ChainID,
    config.L2.ChainID,
    config.L1.RPCURL.toString(),
    config.L2.RPCURL.toString(),
    addressesL1,
    addressesL2Celo,
  );

  // the num accounts will be reset later, when we know
  // how many concurrently sending accounts we need.
  // For now, we don't even use the wallet clients.
  const clients = new ClientAccountManager(
    chains,
    config.AccountsSeedPhrase,
    1,
  );
  const publicClient = clients.public();
  console.log("waiting for chain networks to stabilize");
  {
    const success = await Promise.all([
      waitClientReturnsBlockNum(publicClient.l1, 20),
      waitClientReturnsBlockNum(publicClient.l2, 20),
      waitClientNotSyncing(publicClient.l1, 15),
      waitClientNotSyncing(publicClient.l2, 15),
    ]);
    if (!success.every((v) => v === null)) {
      throw new Error("l1 and l2 clients not reachable within the deadline");
    }
  }
  const portalVersion = await getPortalVersion(clients.public().l1, {
    //TODO: use proper types
    // deno-lint-ignore no-explicit-any
    targetChain: clients.public().l2.chain as any,
  });
  const chainUsesFaultProofs = portalVersion.major >= 3;
  if (chainUsesFaultProofs !== config.UseFaultproofSystem) {
    console.log(
      `'UseFaultproofSystem' is set to ${config.UseFaultproofSystem}, ` +
        `but the chain contracts do not reflect that`,
    );
  }
  if (chainUsesFaultProofs) {
    console.log("L2 chain uses fault-proofs, wait until two games available");
    // NOTE: viem needs at least two games to infer the
    // time to next game, otherwise the function
    // returns NaN
    await waitUntilTwoGames(publicClient, 120);
  } else {
    console.log("L2 chain uses output-oracle, waiting for initial oracle");
    await waitInitialL2OracleOutput(publicClient, 120);
  }
  console.log("chain networks stable");

  const logger = new TestLogger(config.ArtifactsDirPath);
  return new Context(clients, config, undefined, false, addressesL1, logger);
}

export async function teardown(_: Deno.TestContext, config: Config) {
  await teardownDevnet(config);
}
