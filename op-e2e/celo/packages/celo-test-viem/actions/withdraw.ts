import type {
  PublicClient,
  PublicClients,
  WalletClients,
} from "../clients/clients.ts";
import { simulateBridgeERC20To } from "./bridge.ts";
import type { BridgedERC20TokenPair } from "./common.ts";
import type {
  Account,
  Address,
  Chain,
  GetBlockReturnType,
  HttpTransport,
  TransactionReceipt,
} from "viem";
import type {
  BuildProveWithdrawalParameters,
  WaitToFinalizeParameters,
  WaitToFinalizeReturnType,
} from "viem/op-stack";
import { getPortalVersion } from "viem/op-stack";
import { ERC20 } from "reverse-mirage";
import { parseAbi } from "viem";
import { pollFunction, sleepSeconds } from "@celo-test/util";

// partial ABI, only the used functions and errors are defined here
const portal2Abi = parseAbi([
  "function disputeGameFinalityDelaySeconds() view returns (uint256)",
  "function numProofSubmitters(bytes32 _withdrawalHash) view returns (uint256)",
  "function proofSubmitters(bytes32 _withdrawalHash, uint256 index) view returns (address)",
  "function checkWithdrawal(bytes32 _withdrawalHash, address _proofSubmitter) view",
  "error AlreadyFinalized()",
  "error GasEstimation()",
  "error TransferFailed()",
  "error BadTarget()",
  "error Blacklisted()",
  "error Unproven()",
]);

export type WithdrawReturnType = {
  success: boolean;
  l2GasPayment: bigint;
};

export type InitiateWithdrawReturnType = {
  receipt: TransactionReceipt;
  gasPaid: bigint;
};

export async function initiateERC20Withdraw(
  value: bigint,
  to: Address,
  // l1Gas: bigint, // TODO: do we need this here, or should we simulate?
  tokenPair: BridgedERC20TokenPair,
  publicClients: PublicClients,
  walletClients: WalletClients<Account>,
): Promise<InitiateWithdrawReturnType> {
  let localToken: ERC20;
  let remoteToken: ERC20;
  // XXX: why did this work before we correct the
  // if statement?
  if (tokenPair.nativeOnL1 === true) {
    localToken = tokenPair.bridgedToken;
    remoteToken = tokenPair.nativeToken;
  } else {
    localToken = tokenPair.nativeToken;
    remoteToken = tokenPair.bridgedToken;
  }
  const bridgeERC20 = await simulateBridgeERC20To(
    {
      account: walletClients.l2.account,
      chain: publicClients.l2.chain,
      request: {
        //TODO: calculate gas for the l1 execution, so this would be a ERC20 transfer with gas prices on l1?
        gas: 200000,
        to: to,
        value: value,
        data: "0x",
      },
      localToken: localToken.address,
      remoteToken: remoteToken.address,
    },
    publicClients,
  );
  const hash = await walletClients.l2.writeContract(bridgeERC20.args); // TODO: fix type
  const receipt = await publicClients.l2.waitForTransactionReceipt({
    hash: hash,
  });
  console.log("initiateWithdrawal receipt (l2)", receipt);

  return {
    receipt: receipt,
    gasPaid:
      // TODO: when we implement the other direction,
      // l1Fee doesn't exist on l1 receipt,
      receipt.gasUsed * receipt.effectiveGasPrice + (receipt.l1Fee ?? 0n),
  };
}

export async function initiateNativeWithdraw(
  value: bigint,
  to: Address,
  l1Gas: bigint,
  publicClients: PublicClients,
  walletClients: WalletClients<Account>,
): Promise<InitiateWithdrawReturnType> {
  const initiateHash = await walletClients.l2.initiateWithdrawal({
    request: {
      gas: l1Gas,
      to: to,
      value: value,
    },
  });
  const receipt = await publicClients.l2.waitForTransactionReceipt({
    hash: initiateHash,
  });
  console.log("initiateWithdrawal receipt (l2)", receipt);

  return {
    receipt: receipt,
    gasPaid:
      receipt.gasUsed * receipt.effectiveGasPrice + (receipt.l1Fee ?? 0n),
  };
}

export async function settleWithdraw(
  withdrawReceipt: TransactionReceipt,
  publicClients: PublicClients,
  walletClients: WalletClients<Account>,
): Promise<WithdrawReturnType> {
  // first wait with our wait function, because it is target-time based
  // and considers CPU suspend and hibernate:
  // NOTE: viem only returns a time here when more than 2 fault-games
  // can be found from the DisputeGameFactory
  const timeToProve = await publicClients.l1.getTimeToProve({
    receipt: withdrawReceipt,
    // deno-lint-ignore no-explicit-any
    targetChain: publicClients.l2.chain as any,
  });
  if (!timeToProve.seconds) {
    throw Error("couldn't calculate time to prove");
  }
  console.log("waiting for time to prove (in s):", timeToProve.seconds);
  await sleepSeconds(timeToProve.seconds);

  // Only now call the waitToProve, which now will immediately
  // try to call the L1 contracts
  // (XXX: I think? why is this taking so
  // long on the mainnet after already waiting the timeToProve?).
  // TODO: check again what get-time-toprove would show now.
  // because internally the waitToProve waits this time initially
  // before making any calls..
  //
  //
  // NOTE: for the L2OO system,
  // this function requires the mulitcall3 contract to be deployed
  // on the L1 chain.
  //
  console.log("call wait to prove");
  const { output, game, withdrawal } = await publicClients.l1.waitToProve({
    receipt: withdrawReceipt,
    // deno-lint-ignore no-explicit-any
    targetChain: publicClients.l2.chain as any,
  });

  const portalVersion = await getPortalVersion(publicClients.l1, {
    // deno-lint-ignore no-explicit-any
    targetChain: publicClients.l2.chain as any,
  });
  const usesL2OO = portalVersion.major < 3;
  const proveWithdrawalParams: BuildProveWithdrawalParameters = (() => {
    if (usesL2OO) {
      return {
        chain: publicClients.l2.chain,
        withdrawal: withdrawal,
        output: output,
      };
    } else {
      return {
        chain: publicClients.l2.chain,
        withdrawal: withdrawal,
        game: game,
      };
    }
  })();

  const proveWithdrawalArgs = await publicClients.l2.buildProveWithdrawal(
    proveWithdrawalParams,
  );
  console.log(
    "built prove-withdrawal args, to be posted on l1",
    proveWithdrawalArgs,
  );
  const proveHash = await walletClients.l1.proveWithdrawal(
    // deno-lint-ignore no-explicit-any
    proveWithdrawalArgs as any,
  );

  console.log("wait for prove-withdrawal tx hash (l1):)", proveHash);
  const proveReceipt = await publicClients.l1.waitForTransactionReceipt({
    hash: proveHash,
  });
  console.log("proveWithdrawal tx-hash (l1)", proveReceipt.transactionHash);

  if (proveReceipt.status != "success") {
    return {
      success: false,
      l2GasPayment: 0n, //FIXME: we don't need this here actually because now you have it in the initate function
    };
  }

  const waitToFinalizeParams = {
    withdrawalHash: withdrawal.withdrawalHash,
    // deno-lint-ignore no-explicit-any
    targetChain: publicClients.l2.chain satisfies Chain as any,
  };
  const timeToFinalize =
    await publicClients.l1.getTimeToFinalize(waitToFinalizeParams);

  console.log("waiting for time to finalize (in s):", timeToFinalize.seconds);
  await sleepSeconds(timeToFinalize.seconds);
  if (usesL2OO) {
    // the getTimeToFinalize calculates the expected time to finalization
    // based on some contract parameters in wall-time.
    // This could be imprecise, either due to lagging l1 progress (?),
    // or because the time has not been inferred correctly with an incorrect
    // replication of the contracts businesss logic.
    // That's why we actually wait until the block timestamp has reached
    // the calculated finalization time.
    await pollFunction(
      async (): Promise<GetBlockReturnType> =>
        await publicClients.l1.getBlock(),
      (val: GetBlockReturnType | null, _err: Error | null) => {
        if (val !== null) {
          // block timestamp is given in seconds,
          // finalize-suggestion timestamp in ms
          return Number(val.timestamp) * 1000 >= timeToFinalize.timestamp;
        } else {
          return false;
        }
      },
      10000,
      undefined,
      true,
    );
    // XXX: is this enough to be certain of a successful finalization for the L2OO system
    // or do we also need to poll a contract call similarly to the Fault-Proof system?
    // (see else case below)
  } else {
    // We waited the inferred "time-to-finalize" that viem gave us.
    // Now we actually simulate a call to the contract until it
    // lets us finalize.
    // This is necessary because viem's inferrence of the wait
    // time is not correct and underestimates e.g. the additional air-gap
    // time.
    await pollCheckWithdrawal(publicClients.l1, waitToFinalizeParams);
  }

  const finalizeHash = await walletClients.l1.finalizeWithdrawal({
    withdrawal: withdrawal,
    // deno-lint-ignore no-explicit-any
    targetChain: publicClients.l2.chain satisfies Chain as any,
  });

  const finalizeReceipt = await publicClients.l1.waitForTransactionReceipt({
    hash: finalizeHash,
  });
  console.log(
    "finalizeWithdrawal tx-hash (l1)",
    finalizeReceipt.transactionHash,
  );

  return {
    success: finalizeReceipt.status == "success",
    l2GasPayment: 0n, // FIXME: we don't need this here anymore
  };
}

export async function pollCheckWithdrawal<
  chain extends Chain | undefined,
  account extends Account | undefined,
  chainOverride extends Chain | undefined = undefined,
>(
  client: PublicClient<HttpTransport, chain, account>,
  parameters: WaitToFinalizeParameters<chain, chainOverride>,
): Promise<WaitToFinalizeReturnType> {
  const { chain = client.chain, withdrawalHash, targetChain } = parameters;

  const portalAddress = (() => {
    if (parameters.portalAddress) return parameters.portalAddress;
    if (chain) return targetChain!.contracts.portal[chain.id].address;
    return Object.values(targetChain!.contracts.portal)[0].address;
  })();
  const numProofSubmitters = await client
    .readContract({
      abi: portal2Abi,
      address: portalAddress,
      functionName: "numProofSubmitters",
      args: [withdrawalHash],
    })
    .catch(() => 1n);

  const proofSubmitter = await client
    .readContract({
      abi: portal2Abi,
      address: portalAddress,
      functionName: "proofSubmitters",
      args: [withdrawalHash, numProofSubmitters - 1n],
    })
    .catch(() => undefined);
  if (proofSubmitter === undefined) {
    throw Error("no proof submitter found");
  }

  await pollFunction(
    async (): Promise<void> => {
      await client.readContract({
        abi: portal2Abi,
        address: portalAddress,
        functionName: "checkWithdrawal",
        args: [withdrawalHash, proofSubmitter],
      });
    },
    (_value: void | null | null, err: Error | null): boolean => {
      if (err === null) {
        return true;
      }
      return false;
    },
    10000,
    undefined,
    true,
  );
}
