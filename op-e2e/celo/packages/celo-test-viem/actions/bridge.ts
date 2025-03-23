import type { PublicClients } from "../clients/clients.d.ts";
import { standardBridge } from "../abis/abis.ts";
import type { Address, Chain } from "viem";
import { getStandardBridgeAddresses } from "./common.ts";

export async function simulateBridgeERC20To(
  parameters: any,
  publicClients: PublicClients,
) {
  const {
    account,
    chain,
    request: { data = "0x", gas: depositGas, to = "0x", value },
    localToken,
    remoteToken,
  } = parameters;

  let client;
  let localStandardBridgeAddress;
  const standardBridges = getStandardBridgeAddresses(publicClients);
  if (chain.id === publicClients.l1.chain.id) {
    client = publicClients.l1;
    localStandardBridgeAddress = standardBridges.l1;
  } else if (chain.id === publicClients.l2.chain.id) {
    client = publicClients.l2;
    localStandardBridgeAddress = standardBridges.l2;
  } else {
    throw Error("provided chain does not match any client");
  }

  const { maxFeePerGas, maxPriorityFeePerGas } =
    await client.estimateFeesPerGas();
  // TODO: estimate the erc20 contract call on the other side for the depositGas
  const args = {
    address: localStandardBridgeAddress as Address,
    abi: standardBridge.abi,
    account: account,
    chain: chain as Chain,
    functionName: "bridgeERC20To",
    args: [localToken, remoteToken, to, value ?? 0n, depositGas, data] as any, //TODO: use the proper type
    maxFeePerGas,
    maxPriorityFeePerGas,
    gas: BigInt(0), //estimate later
  };
  if (typeof args.gas !== "number") {
    const gas_ = await client.estimateContractGas(args);
    args.gas = gas_;
  }
  const result = client.simulateContract(args);
  return { result: result, args: args };
}
