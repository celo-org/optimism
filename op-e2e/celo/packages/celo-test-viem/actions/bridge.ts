import { standardBridge } from "../abis/abis.ts";
import { getStandardBridgeAddresses } from "./common.ts";
import type { PublicClients, WalletClients } from "../clients/clients.ts";
import type { BridgedERC20TokenPair } from "./common.ts";
import type { Account, Address, Chain, TransactionReceipt } from "viem";
import type { ERC20 } from "reverse-mirage";

export type InitiateBridgeERC20ToReturnType = {
  approve: {
    receipt: TransactionReceipt | undefined;
    chainId: number;
  };
  bridge: {
    receipt: TransactionReceipt | undefined;
    chainId: number;
  };
};

// works in both directions
export async function initiateBridgeERC20To(
  value: bigint,
  to: Address,
  chain: any, // this is the source chain, where the tx is sent on
  // l1Gas: bigint, // TODO: do we need this here, or should we simulate?
  tokenPair: BridgedERC20TokenPair,
  publicClients: PublicClients,
  walletClients: WalletClients<Account>,
): Promise<InitiateBridgeERC20ToReturnType> {
  let localToken: ERC20;
  let remoteToken: ERC20;

  let pubClient: any;
  let walletClient: any;
  let localStandardBridge: Address;
  let approveReceipt: TransactionReceipt | undefined;
  let gasPaid = BigInt(0);

  const standardBridges = getStandardBridgeAddresses(publicClients);
  let chainHasNativeToken = false;
  if (chain.id === publicClients.l1.chain.id) {
    chainHasNativeToken = tokenPair.nativeOnL1;
    pubClient = publicClients.l1;
    walletClient = walletClients.l1;
    localStandardBridge = standardBridges.l1;
  } else if (chain.id === publicClients.l2.chain.id) {
    chainHasNativeToken = !tokenPair.nativeOnL1;
    pubClient = publicClients.l2;
    walletClient = walletClients.l2;
    localStandardBridge = standardBridges.l2;
  } else {
    throw Error("native chain does not match any client chain ids");
  }
  if (chainHasNativeToken === true) {
    localToken = tokenPair.nativeToken;
    remoteToken = tokenPair.bridgedToken;
  } else {
    localToken = tokenPair.bridgedToken;
    remoteToken = tokenPair.nativeToken;
  }

  //TODO: return approve receipt
  // approve the allowance for the bridge,
  // but only if we are on the "native" side
  if (chainHasNativeToken === true) {
    const approve = await walletClient.simulateERC20Approve({
      args: {
        amount: { amount: value, token: localToken, type: "Amount" },
        spender: localStandardBridge,
      },
    });
    if (!approve.result) {
      throw Error("couldn't approve to bridge");
    }
    const approveHash = await walletClient.writeContract(approve.request);
    approveReceipt = await pubClient.waitForTransactionReceipt({
      hash: approveHash,
    });
  }

  const bridgeERC20 = await simulateBridgeERC20To(
    {
      account: walletClient.account,
      chain: chain,
      request: {
        //TODO: calculate gas for the other chain execution,
        //so this would be a ERC20 transfer simulated on the other chain rpc?
        // gas: 200000,
        gas: 200000, // FIXME: this seems to work, but this is only what's added to some base-gas calculations within the messenger.
        to: to,
        value: value,
        data: "0x",
      },
      localToken: localToken.address,
      remoteToken: remoteToken.address,
    },
    publicClients,
  );
  const hash = await walletClient.writeContract(bridgeERC20.args); // TODO: fix type
  const receipt = await pubClient.waitForTransactionReceipt({
    hash: hash,
  });

  return {
    approve: {
      receipt: approveReceipt,
      chainId: chain.id,
    },
    bridge: {
      receipt: receipt,
      chainId: chain.id,
    },
  };
}

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
  };
  const gas_ = await client.estimateContractGas(args);
  (args as any).gas = gas_;
  // XXX: maybe have to add the depositGas gas here as well?

  const result = client.simulateContract(args);
  return { result: result, args: args };
}
