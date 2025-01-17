import { getL2TransactionHashes } from 'viem/op-stack';
import { Account, Hex, PublicClient, TransactionReceipt } from 'viem';
import { OptimismPortalABI } from './OptimismPortal';
import { Config } from './setup';
import { portalAbi } from 'viem/_types/op-stack/abis';

interface ConstructDepositCustomGasRequest {
  data?: string;
  gas?: bigint;
  isCreation?: boolean;
  mint?: bigint;
  to?: string;
  value?: bigint;
}

interface TargetChain {
  contracts: {
    portal: Record<number, { address: string }>;
  };
}

export interface ConstructDepositCustomGasParameters {
  account: Account; // Use appropriate Account type if available
  chain?: any;     // Replace 'any' with specific Chain type if known
  gas?: bigint | null;
  maxFeePerGas?: bigint;
  maxPriorityFeePerGas?: bigint;
  nonce?: number;
  request: ConstructDepositCustomGasRequest;
  targetChain: TargetChain;
  portalAddress?: string;
}

const zeroAddress = "0x0000000000000000000000000000000000000000"

export async function constructDepositCustomGas(
  client: PublicClient,
  parameters: ConstructDepositCustomGasParameters
) {
  const {
    account,
    chain = client.chain,
    gas,
    maxFeePerGas,
    maxPriorityFeePerGas,
    nonce,
    request: {
      data = '0x',
      gas: l2Gas,
      isCreation = false,
      mint,
      to = '0x',
      value,
    },
    targetChain,
  } = parameters;

  const portalAddress = (() => {
    if (parameters.portalAddress) return parameters.portalAddress;
    if (chain) return targetChain.contracts.portal[chain.id].address;
    return Object.values(targetChain.contracts.portal)[0].address;
  })();

  const callArgs: Parameters<typeof client.simulateContract>[0] = {
    account: account.address,
    abi: OptimismPortalABI,
    address: portalAddress as Hex,
    chain,
    functionName: 'depositERC20Transaction',
    args: [
      isCreation ? zeroAddress : to,
      mint ?? value ?? 0n,
      value ?? mint ?? 0n,
      l2Gas,
      isCreation,
      data,
    ],
    maxFeePerGas,
    maxPriorityFeePerGas,
    nonce,
    gas: 200_000n // default gas limit * 2
  };

  // const gas_ =
  //   typeof gas !== 'number' && gas !== null
  //     ? await client.estimateContractGas(callArgs)
  //     : undefined;
  // callArgs.gas = gas_!;
  const safeStringify = (obj: unknown): string =>
    JSON.stringify(obj, (_key: string, value: unknown): unknown =>
      typeof value === 'bigint' ? value.toString() : value
    );

  const result = await client.simulateContract(callArgs);

  callArgs.account = null

  return { result, args: callArgs };
}

interface DepositArgs {
  mint: bigint;
  to: Hex;
}


export async function deposit(
  args: DepositArgs,
  config: Config
): Promise<{ success: boolean; l1GasPayment: bigint }> {
  let spentGas = BigInt(0);

  const depositArgs = await config.client.l2.public.buildDepositTransaction({
    mint: args.mint,
    to: args.to,
    account: config.account,
  });

  const celoToken = await config.client.l1.public.getERC20({
    erc20: {
      address: config.addresses.CustomGasTokenProxy as Hex,
      chainID: config.client.l1.public.chain.id,
    },
  });

  const portalAddress =
    config.client.l2.public.chain.contracts.portal[
      config.client.l1.public.chain.id
    ].address;

  const approve = await config.client.l1.wallet.simulateERC20Approve({
    amount: { amount: args.mint, token: celoToken },
    spender: portalAddress,
  });
  if (!approve.result) {
    return {
      success: false,
      l1GasPayment: spentGas,
    };
  }

  const approveHash = await config.client.l1.wallet.writeContract(
    approve.request
  );
  // Wait for the L1 transaction to be processed.
  const approveReceipt = await config.client.l1.public.waitForTransactionReceipt({
    hash: approveHash,
  }) as TransactionReceipt;

  spentGas += approveReceipt.gasUsed * approveReceipt.effectiveGasPrice;

  console.log("wallet l1 address", config.client.l1.wallet.account.address);

  const dep = await config.client.l1.public.prepareDepositGasPayingTokenERC20(depositArgs as any)
  dep.args.account = config.account;
  const depositHash = await config.client.l1.wallet.writeContract(dep.args);

  // Wait for the L1 transaction to be processed.
  const depositReceipt = await config.client.l1.public.waitForTransactionReceipt({
    hash: depositHash,
  }) as TransactionReceipt;

  spentGas += depositReceipt.gasUsed * depositReceipt.effectiveGasPrice;

  if (depositReceipt.status !== 'success') {
    return {
      success: false,
      l1GasPayment: spentGas
    }
  }

  // Get the L2 transaction hash from the L1 transaction receipt.
  const [l2Hash] = getL2TransactionHashes(depositReceipt);

  // Wait for the L2 transaction to be processed.
  const l2Receipt = await config.client.l2.public.waitForTransactionReceipt({
    hash: l2Hash,
  });

  return {
    success: l2Receipt.status === 'success',
    l1GasPayment: spentGas,
  };
}
