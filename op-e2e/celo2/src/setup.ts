import { setupClients, SetupClientsReturn } from "../src/config";
import { makeChainConfigs } from "../src/chain";
import { privateKeyToAccount, Account } from "viem/accounts";
import type { Hex, PublicClient } from "viem"; // Example: adjust based on actual types from viem
import * as addresses from "../../../.devnet/addresses.json";

// Default Anvil dev account that has a pre-allocation on the op-devnet:
// "test test test test test test test test test test test junk" mnemonic account,
// on path "m/44'/60'/0'/0/6".
// Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266.
const privKey: Hex =
  "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80";

async function waitForNoError(
  func: () => Promise<any>,
  timeout: number,
): Promise<boolean> {
  const start = Date.now();
  while (Date.now() - start < timeout) {
    try {
      await func();
      return true;
    } catch (error) {
      // Optionally handle/log error here
    }
    await new Promise((r) => setTimeout(r, 1000));
  }
  return false;
}

async function waitReachable(
  client: PublicClient | any,
  timeout: number,
): Promise<boolean> {
  const f = async () => client.getChainId();
  return waitForNoError(f, timeout);
}

async function waitForNextGame(
  client: any,
  l2ChainConfig: any,
  timeout: number,
): Promise<boolean> {
  const f = async () =>
    client.waitForNextGame({
      pollingInterval: 500,
      l2BlockNumber: 0,
      targetChain: l2ChainConfig,
    });
  return waitForNoError(f, timeout);
}

export type AddressesType = typeof addresses;

export interface Config {
  account: Account;
  client: SetupClientsReturn;
  addresses: AddressesType;
}

export async function setup() {
  const chainConfig = makeChainConfigs(900, 901, addresses);
  const account = privateKeyToAccount(privKey);
  const client = setupClients(chainConfig, account);

  const config: Config = {
    account,
    client,
    addresses,
  };

  const success = await Promise.all([
    waitReachable(config.client.l1.public, 10_000),
    waitReachable(config.client.l2.public, 10_000),
    waitForNextGame(config.client.l1.public, chainConfig.l2, 60_000),
  ]);

  if (success.every((v) => v === true)) {
    return config;
  }
  throw new Error("l1 and l2 clients not reachable within the deadline");
}

export interface Addresses {
  AddressManager: string;
  AnchorStateRegistry: string;
  AnchorStateRegistryProxy: string;
  DelayedWETH: string;
  DelayedWETHProxy: string;
  DisputeGameFactory: string;
  DisputeGameFactoryProxy: string;
  FastPreimageOracle: string;
  FaultDisputeGame_0: string;
  FaultDisputeGame_254: string;
  FaultDisputeGame_255: string;
  L1CrossDomainMessenger: string;
  L1CrossDomainMessengerProxy: string;
  L1ERC721Bridge: string;
  L1ERC721BridgeProxy: string;
  L1StandardBridge: string;
  L1StandardBridgeProxy: string;
  Mips: string;
  OPContractsManager: string;
  OPContractsManagerProxy: string;
  OptimismMintableERC20Factory: string;
  OptimismMintableERC20FactoryProxy: string;
  OptimismPortal2: string;
  OptimismPortalProxy: string;
  PermissionedDelayedWETHProxy: string;
  PermissionedDisputeGame: string;
  PreimageOracle: string;
  ProtocolVersions: string;
  ProtocolVersionsProxy: string;
  ProxyAdmin: string;
  SuperchainConfig: string;
  SuperchainConfigProxy: string;
  SuperchainProxyAdmin: string;
  SystemConfig: string;
  SystemConfigProxy: string;
}
