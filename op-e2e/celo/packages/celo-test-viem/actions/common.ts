import type { Address, Chain, ChainContract } from "viem";
import type { PublicClients } from "../clients/clients.ts";
import type { ERC20 } from "reverse-mirage";

export type StandardBridgeAddresses = {
  l1: Address;
  l2: Address;
};

export type BridgedERC20TokenPair = {
  bridgedToken: ERC20;
  nativeToken: ERC20;
  nativeOnL1: boolean;
};

export function getContractAddress(
  chain: Chain,
  sourceChain: Chain,
  contract: string,
): Address | undefined {
  if (chain.id === sourceChain.id) {
    const c = chain?.contracts?.[contract] as ChainContract | undefined;
    return c?.address;
  }
  const c = chain?.contracts?.[contract] as {
    [sourceId: number]: ChainContract | undefined;
  };
  return c[sourceChain.id]?.address;
}

export function getStandardBridgeAddresses(
  publicClients: PublicClients,
): StandardBridgeAddresses {
  const l1StandardBridgeAddress = getContractAddress(
    publicClients.l2.chain,
    publicClients.l1.chain,
    "l1StandardBridge",
  );
  if (l1StandardBridgeAddress === undefined) {
    throw Error("l1 standard bridge not found");
  }
  const l2standardBridgeAddress = (
    publicClients.l2.chain.contracts?.l2StandardBridge as ChainContract
  ).address;
  if (l2standardBridgeAddress === undefined) {
    throw Error("l2standard bridge not found");
  }
  return { l1: l1StandardBridgeAddress, l2: l2standardBridgeAddress };
}
