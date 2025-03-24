import { createPublicClient, createWalletClient, http } from "viem";
import { privateKeyToAccount } from "viem/accounts";
import { sleep } from "@celo-test/util";
import { publicActionsERC20, walletActionsERC20 } from "./erc20.ts";
import type {
  Account,
  Address,
  Chain,
  HDAccount,
  Hex,
  HttpTransport,
  PublicClient as ViemPublicClient,
  TransactionReceipt,
  WalletClient as ViemWalletClient,
} from "viem";
import type {
  PublicClient,
  PublicClients,
  WalletClients,
} from "./clients.d.ts";
import type { Chains } from "../chains.ts";
import {
  publicActionsL1,
  publicActionsL2,
  walletActionsL1,
  walletActionsL2,
} from "viem/op-stack";
import { mnemonicToAccount } from "viem/accounts";
export type { PublicClient };

export function createPublicClients(chains: Chains): PublicClients {
  return {
    l1: createPublicClient({
      chain: chains.l1,
      transport: http(),
    })
      .extend(publicActionsL1())
      .extend(publicActionsERC20()),
    l2: createPublicClient({
      chain: chains.l2,
      transport: http(),
    })
      .extend(publicActionsL2())
      .extend(publicActionsERC20()),
  };
}

export function createWalletClients<
  accountOrAddress extends Account | Address | undefined = undefined,
>(
  chains: Chains,
  account: accountOrAddress,
  accountL2: accountOrAddress | undefined,
): WalletClients<accountOrAddress> {
  return {
    l1: createWalletClient({
      account: account,
      chain: chains.l1,
      transport: http(),
    })
      .extend(walletActionsERC20())
      .extend(walletActionsL1()),
    l2: createWalletClient({
      account: accountL2 ?? account,
      chain: chains.l2,
      transport: http(),
    })
      .extend(walletActionsERC20())
      .extend(walletActionsL2()),
  };
}

export class ClientAccountManager {
  chains: Chains;
  seedPhrase: string;
  l1Iterator: Generator<Account>;
  l2Iterator: Generator<Account>;
  numAccounts: number;

  constructor(chains: Chains, seedPhrase: string, numAccounts: number) {
    this.seedPhrase = seedPhrase;
    this.chains = chains;
    this.numAccounts = numAccounts;
    this.l1Iterator = this.iterFundedAccounts(this.chains.l1, this.numAccounts);
    this.l2Iterator = this.iterFundedAccounts(this.chains.l2, this.numAccounts);
  }

  public(): PublicClients {
    return createPublicClients(this.chains);
  }
  wallet(l1Account: Account, l2Account: Account): WalletClients<Account> {
    return createWalletClients(this.chains, l1Account, l2Account);
  }

  private deriveAccount(chain: Chain, index: number): HDAccount {
    return mnemonicToAccount(this.seedPhrase, {
      // changeIndex: chain.id, //NOTE: for now use same accounts on different chains
      // maps to the last number in the path:
      addressIndex: index,
    });
  }
  async fundAccountsFrom(privkey: Hex): Promise<void> {
    const funderAccount = privateKeyToAccount(privkey);
    const funder = createWalletClients(
      this.chains,
      funderAccount,
      funderAccount,
    );
    const pub = this.public();
    const res = await Promise.allSettled([
      this._fundAccountsForChainFrom(
        funder.l1 as ViemWalletClient<HttpTransport, Chain, Account>,
        pub.l1 as ViemPublicClient<HttpTransport, Chain>,
      ),
      this._fundAccountsForChainFrom(
        funder.l2 as ViemWalletClient<HttpTransport, Chain, Account>,
        pub.l2 as ViemPublicClient<HttpTransport, Chain>,
      ),
    ]);
    // Flatten the results and check for success
    const successes = res
      .filter((result) => result.status === "fulfilled")
      .map(
        (result) =>
          (result as PromiseFulfilledResult<Array<TransactionReceipt>>).value,
      );

    const errors = res
      .filter((result) => result.status === "rejected")
      .map((result) => (result as PromiseRejectedResult).reason);

    if (errors.length > 0) {
      throw new Error(`funding accounts failed: ${JSON.stringify(errors)}`);
    }
    successes.forEach((receipts) => {
      receipts.forEach((receipt) => {
        if (receipt.status !== "success") {
          throw new Error(
            `funding accounts failed with 'reverted' transaction, ` +
              `tx-hash: ${receipt.transactionHash}`,
          );
        }
      });
    });
  }
  private async _fundAccountsForChainFrom(
    leader: ViemWalletClient<HttpTransport, Chain, Account>,
    publicClient: ViemPublicClient<HttpTransport, Chain>,
  ): Promise<Array<TransactionReceipt>> {
    const it = this.iterFundedAccounts(leader.chain, this.numAccounts);
    const balance = await publicClient.getBalance({
      address: leader.account.address,
    });

    const gasPrice = await publicClient.getGasPrice();
    // overshoot the current gas-price for fluctuation
    const maxFeePerGas = (gasPrice * BigInt(15)) / BigInt(10);
    // We need some funds for gas to distribute to the test accounts.
    const feePerTx = maxFeePerGas * BigInt(21000);
    const sendBalance =
      (balance as bigint) / BigInt(this.numAccounts) - feePerTx;
    if (sendBalance <= BigInt(0)) {
      throw Error("leader account insufficient funds");
    }

    const receipts: Array<Promise<TransactionReceipt>> = [];
    let transactionCount = await publicClient.getTransactionCount({
      address: leader.account.address,
    });
    for (const acc of it) {
      if (acc.address === leader.account.address) {
        console.log("skipping funding leader account");
        continue;
      }
      await sleep(500);
      const hash = await leader.sendTransaction({
        type: "eip1559",
        maxFeePerGas: maxFeePerGas,
        value: sendBalance,
        to: acc.address,
        nonce: transactionCount,
      });
      transactionCount++;
      receipts.push(
        publicClient.waitForTransactionReceipt({
          hash: hash,
        }),
      );
    }
    return await Promise.all(receipts);
  }

  reset(newNumAccounts: number | null) {
    this.numAccounts = newNumAccounts ?? this.numAccounts;
    this.l1Iterator = this.iterFundedAccounts(this.chains.l1, this.numAccounts);
    this.l2Iterator = this.iterFundedAccounts(this.chains.l2, this.numAccounts);
  }

  //TODO: in all 3 methods below:
  // what if iterator exhausted
  // error or return undefined?
  nextFundedL1Account(): HDAccount {
    return this.l1Iterator.next().value;
  }
  nextFundedL2Account(): HDAccount {
    return this.l2Iterator.next().value;
  }
  private *iterFundedAccounts(chain: Chain, num: number): Generator<HDAccount> {
    for (let i = 0; i < num; i++) {
      yield this.deriveAccount(chain, i);
    }
  }
}

export type { PublicClients, WalletClients };
