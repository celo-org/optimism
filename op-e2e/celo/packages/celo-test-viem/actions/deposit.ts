import type { PublicClients, WalletClients } from "../clients/clients.d.ts";
import type { L2Chain } from "../chains.ts";
import { getContractAddress } from "./common.ts";
import { simulateBridgeERC20To } from "./bridge.ts";
import { parseAbiItem } from "viem";
import type {
  Account,
  Address,
  Chain,
  TransactionReceipt,
  WaitForTransactionReceiptReturnType,
} from "viem";
import { getL2TransactionHashes } from "viem/op-stack";
import type { BuildDepositTransactionReturnType } from "viem/op-stack";

export type DepositReturnType = {
  l1Approve: {
    success: boolean;
    receipt: TransactionReceipt | undefined;
  };
  l1Deposit: {
    success: boolean;
    receipt: TransactionReceipt | undefined;
  };
  l2Deposit: {
    success: boolean;
    receipt: TransactionReceipt | undefined;
  };
};

export async function waitForDepositReceiptL2(
  l1Receipt: TransactionReceipt,
  publicClients: PublicClients,
): Promise<WaitForTransactionReceiptReturnType<L2Chain>> {
  const [l2Hash] = getL2TransactionHashes(l1Receipt);
  const l2Receipt = await publicClients.l2.waitForTransactionReceipt({
    hash: l2Hash,
    // 30 minutes timeout, we need the l1 block to finalise in order for the
    // deposit tx to appear on the l2
    timeout: 30 * 60000,
  });
  return l2Receipt as WaitForTransactionReceiptReturnType<L2Chain>;
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
      l1Approve: {
        success: false,
        receipt: undefined,
      },
      l1Deposit: {
        success: false,
        receipt: undefined,
      },
      l2Deposit: {
        success: false,
        receipt: undefined,
      },
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
  const l1Receipt = await publicClients.l1.waitForTransactionReceipt({
    hash: hash,
  });
  spentGas += l1Receipt.gasUsed * l1Receipt.effectiveGasPrice;
  const [l2Hash] = getL2TransactionHashes(l1Receipt);
  const l2Receipt = await publicClients.l2.waitForTransactionReceipt({
    hash: l2Hash,
    // 30 minutes timeout, we need the l1 block to finalise in order for the
    // deposit tx to appear on the l2
    timeout: 30 * 60000,
  });
  return {
    l1Approve: {
      success: true,
      receipt: approveReceipt,
    },
    l1Deposit: {
      success: true,
      receipt: l1Receipt,
    },
    l2Deposit: {
      success: l2Receipt.status === "success",
      receipt: l2Receipt,
    },
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
