import { ClientAccountManager } from "@celo-test/viem";
import type {
  Chains,
  ContractAddressesL1,
  PublicClients,
  WalletClients,
} from "@celo-test/viem";
import type { Config } from "./types.ts";
import type { Account, HDAccount, Hex } from "viem";
import { toHex } from "viem";

export class Context {
  concurrent: boolean;
  config: Config;
  contracts: ContractAddressesL1;
  clientManager: ClientAccountManager;
  parent: Context | null;
  chains: Chains;

  private _pubClients: PublicClients | undefined;
  private _walletClients: WalletClients<Account> | undefined;

  constructor(
    clientManager: ClientAccountManager,
    config: Config,
    parent: Context | undefined,
    concurrent: boolean,
    contracts: ContractAddressesL1,
  ) {
    this.contracts = contracts;
    this.clientManager = clientManager;
    this.chains = clientManager.chains;
    this.config = config;
    this.concurrent = concurrent;
    if (parent === undefined) {
      this.parent = null;
    } else {
      this.parent = parent;
      this._pubClients = parent.public();
    }
  }
  public(): PublicClients {
    if (this._pubClients !== undefined) {
      return this._pubClients;
    }
    this._pubClients = this.clientManager.public();
    return this._pubClients;
  }

  l1PrivateKey(): Hex {
    const acc = this.wallet().l1.account as HDAccount;
    return toHex(acc.getHdKey().privateKey!);
  }
  l2PrivateKey(): Hex {
    const acc = this.wallet().l2.account as HDAccount;
    return toHex(acc.getHdKey().privateKey!);
  }

  resetClients(numFundedAccounts: number) {
    this._pubClients = undefined;
    this._walletClients = undefined;
    this.clientManager.reset(numFundedAccounts);
  }

  wallet(): WalletClients<Account> {
    if (this._walletClients !== undefined) {
      return this._walletClients;
    }

    if (this.concurrent || this.parent === null) {
      const account = this.clientManager.nextFundedL2Account();
      // reuse the l2 private-key for l1 as well.
      // TODO: check if they are funded?
      this._walletClients = this.clientManager.wallet(account, account);
    } else {
      this._walletClients = this.parent.wallet();
    }
    return this._walletClients;
  }

  child(concurrent: boolean): Context {
    return new Context(
      this.clientManager,
      this.config,
      this,
      concurrent,
      this.contracts,
    );
  }
}
