import type { PublicClients, WalletClients } from "../clients/clients.d.ts";
import { getContractAddress } from "./common.ts";
import { simulateBridgeERC20To } from "./bridge.ts";
import { parseAbiItem } from "viem";
import type { Account, Address, Chain, ChainContract } from "viem";
import { getL2TransactionHashes } from "viem/op-stack";
import type { BridgedERC20TokenPair } from "./common.ts";
import type { ERC20 } from "reverse-mirage";
import type { BuildDepositTransactionReturnType } from "viem/op-stack";

export type DepositReturnType = {
  success: boolean;
  l1GasPayment: bigint;
};

// FIXME: we also have to change the call signature
// to get the chain and bridgedTokenPair instead?
export async function bridgeERC20To(
  chain: any,
  amount: bigint,
  to: Address,
  publicClients: PublicClients,
  walletClients: WalletClients<Account>,
): Promise<DepositReturnType> {
  let spentGas = BigInt(0);

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

  const approve = await walletClients.l1.simulateERC20Approve({
    args: {
      amount: { amount: amount, token: tokenL1, type: "Amount" },
      spender: l1StandardBridgeAddress,
    },
  });
  if (!approve.result) {
    return {
      success: false,
      l1GasPayment: spentGas,
    };
  }
  const approveHash = await walletClients.l1.writeContract(approve.request);
  const approveReceipt = await publicClients.l1.waitForTransactionReceipt({
    hash: approveHash,
  });
  spentGas += approveReceipt.gasUsed * approveReceipt.effectiveGasPrice;

  // for now, this is only l1->l2, but this could be written universally
  const depositERC20 = await simulateBridgeERC20To(
    {
      account: walletClients.l1.account,
      chain: publicClients.l1.chain,
      request: {
        //TODO: calculate gas for the l2 execution, so this would be a ERC20 transfer?
        gas: 200000,
        to: to,
        value: amount,
        data: "0x",
      },
      localToken: tokenL1.address,
      remoteToken: tokenL2.address,
    },
    publicClients,
  );
  const hash = await walletClients.l1.writeContract(depositERC20.args);
  const receipt = await publicClients.l1.waitForTransactionReceipt({
    hash: hash,
  });
  console.log("deposit erc20 tx-hash (l1)", receipt.transactionHash);
  //
  spentGas += receipt.gasUsed * receipt.effectiveGasPrice;
  const [l2Hash] = getL2TransactionHashes(receipt);
  console.log("waiting for erc20 deposit tx-hash (l1)", l2Hash);
  const l2Receipt = await publicClients.l2.waitForTransactionReceipt({
    hash: l2Hash,
    // 30 minutes timeout, we need the l1 block to finalise in order for the
    // deposit tx to appear on the l2
    timeout: 30 * 60000,
  });
  console.log("deposit erc20 tx-hash (l2)", l2Receipt.transactionHash);
  return {
    success: l2Receipt.status == "success",
    l1GasPayment: spentGas,
  };
}

export async function deposit(
  amount: bigint,
  to: Address,
  publicClients: PublicClients,
  walletClients: WalletClients<Account>,
): Promise<DepositReturnType> {
  let spentGas = BigInt(0);

  const depositArgs = await publicClients.l2.buildDepositTransaction({
    account: walletClients.l1.account,
    to: to,
    mint: amount,
    value: BigInt(0),
  });

  const customGasTokenAddress = getContractAddress(
    publicClients.l2.chain,
    publicClients.l1.chain,
    "customGasToken",
  );
  if (customGasTokenAddress === undefined) {
    throw Error("customGasToken contract is not defined on l2 chain config");
  }
  const portalAddress = getContractAddress(
    publicClients.l2.chain,
    publicClients.l1.chain,
    "portal",
  );
  if (portalAddress === undefined) {
    throw Error("portal contract is not defined on l2 chain config");
  }
  const celoToken = await publicClients.l1.getERC20({
    erc20: {
      address: customGasTokenAddress,
      chainID: publicClients.l1.chain!.id,
    },
  });
  const approve = await walletClients.l1.simulateERC20Approve({
    args: {
      amount: { amount: amount, token: celoToken, type: "Amount" },
      spender: portalAddress,
    },
  });
  if (!approve.result) {
    return {
      success: false,
      l1GasPayment: spentGas,
    };
  }

  const approveHash = await walletClients.l1.writeContract(approve.request);
  const approveReceipt = await publicClients.l1.waitForTransactionReceipt({
    hash: approveHash,
  });

  spentGas += approveReceipt.gasUsed * approveReceipt.effectiveGasPrice;

  const depositCustomGas = await simulateDepositCustomGas(
    depositArgs,
    publicClients,
  );
  const hash = await walletClients.l1.writeContract(depositCustomGas.args);
  const receipt = await publicClients.l1.waitForTransactionReceipt({
    hash: hash,
  });
  console.log("deposit custom-gas tx-hash (l1)", receipt.transactionHash);
  //
  spentGas += receipt.gasUsed * receipt.effectiveGasPrice;
  const [l2Hash] = getL2TransactionHashes(receipt);
  console.log("waiting for custom-gas deposit tx-hash (l1)", l2Hash);
  const l2Receipt = await publicClients.l2.waitForTransactionReceipt({
    hash: l2Hash,
    // 30 minutes timeout, we need the l1 block to finalise in order for the
    // deposit tx to appear on the l2
    timeout: 30 * 60000,
  });
  console.log("deposit custom-gas tx-hash (l2)", l2Receipt.transactionHash);
  return {
    success: l2Receipt.status == "success",
    l1GasPayment: spentGas,
  };
}

async function simulateDepositCustomGas(
  parameters: BuildDepositTransactionReturnType<Account, Account>,
  publicClients: PublicClients,
) {
  const client = publicClients.l1;
  const {
    account,
    chain = client.chain,
    gas,
    maxFeePerGas,
    maxPriorityFeePerGas,
    nonce,
    request: {
      data = "0x",
      gas: l2Gas,
      isCreation = false,
      mint,
      to = "0x",
      value,
    },
    targetChain,
  } = parameters;

  const depositERC20Transaction = parseAbiItem(
    "function depositERC20Transaction(address _to, uint256 _mint, uint256 _value, uint64 _gasLimit, bool _isCreation, bytes memory _data) public",
  );

  const portalAddress = (() => {
    if (parameters.portalAddress) return parameters.portalAddress;
    if (chain) return targetChain!.contracts.portal[chain.id].address;
    return Object.values(targetChain!.contracts.portal)[0].address;
  })();
  const args = {
    address: portalAddress,
    abi: [depositERC20Transaction],
    account: account,
    chain: chain as Chain,
    functionName: depositERC20Transaction.name,
    args: [
      to,
      mint ?? value ?? 0n,
      value ?? mint ?? 0n,
      l2Gas,
      isCreation,
      data,
    ] as any, //TODO: use the proper type
    maxFeePerGas,
    maxPriorityFeePerGas,
    nonce,
    gas: gas === null ? undefined : gas,
  };
  if (typeof args.gas !== "number") {
    const gas_ = await client.estimateContractGas(args);
    args.gas = gas_;
  }
  const result = client.simulateContract(args);
  return { result: result, args: args };
}
