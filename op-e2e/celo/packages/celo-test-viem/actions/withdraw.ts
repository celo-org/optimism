import type {
  PublicClient,
  PublicClients,
  WalletClients,
} from "../clients/clients.ts";
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
  GetGameReturnType,
  WaitToFinalizeParameters,
  WaitToFinalizeReturnType,
} from "viem/op-stack";
import { getPortalVersion } from "viem/op-stack";
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
  prove: {
    receipt: TransactionReceipt | undefined;
    success: boolean;
    chainId: number;
  };
  finalize: {
    receipt: TransactionReceipt | undefined;
    success: boolean;
    chainId: number;
  };
};

export type InitiateWithdrawReturnType = {
  receipt: TransactionReceipt;
  gasPaid: bigint;
};

export type InitiateBridgeERC20ToReturnType = {
  receipt: TransactionReceipt;
  gasPaid: bigint;
};

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
  // the getTimeToProve broadly speaking takes the average delta of the last 10 games
  // so this overestimates when we recently stopped the propeser / challenger.
  // that's why we rather poll the get games for the l2Blocknumber,
  // so that the waitToProve immediately finds the game and calculates a wait time of 0

  console.log(
    "waiting for dispute game that includes l2-block: ",
    withdrawReceipt.blockNumber,
  );
  await pollFunction(
    async (): Promise<GetGameReturnType> =>
      await publicClients.l1.getGame({
        l2BlockNumber: withdrawReceipt.blockNumber,
        targetChain: publicClients.l2.chain as any,
      }),
    (val: GetGameReturnType | null, _err: Error | null) => {
      if (_err != null) {
        return false;
      }
      if (val !== null) {
        // we found a game for the blocknumber
        return true;
      } else {
        return false;
      }
    },
    60_000,
    undefined,
    false,
  );
  // now check again that the time to prove is actually 0
  // // XXX: (this is mainly for testing that the poll time calculation in waitToProve now returns 0)
  // first wait with our wait function, because it is target-time based
  // and considers CPU suspend and hibernate:
  // NOTE: viem only returns a time here when more than 2 fault-games
  // can be found from the DisputeGameFactory
  const timeToProve = await publicClients.l1.getTimeToProve({
    receipt: withdrawReceipt,
    // deno-lint-ignore no-explicit-any
    targetChain: publicClients.l2.chain as any,
  });
  console.log("timeToProve", timeToProve);
  if (timeToProve.seconds === undefined) {
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
  const proveHash = await walletClients.l1.proveWithdrawal(
    // deno-lint-ignore no-explicit-any
    proveWithdrawalArgs as any,
  );

  const proveReceipt = await publicClients.l1.waitForTransactionReceipt({
    hash: proveHash,
  });
  const proveSuccess = proveReceipt.status === "success";
  if (!proveSuccess) {
    return {
      prove: {
        receipt: proveReceipt,
        success: proveSuccess,
        chainId: publicClients.l1.chain.id,
      },
      finalize: {
        receipt: undefined,
        success: false,
        chainId: publicClients.l1.chain.id,
      },
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

  return {
    prove: {
      receipt: proveReceipt,
      success: proveSuccess,
      chainId: publicClients.l1.chain.id,
    },
    finalize: {
      receipt: finalizeReceipt,
      success: finalizeReceipt.status === "success",
      chainId: publicClients.l1.chain.id,
    },
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
