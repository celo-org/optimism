import type {
  Account,
  Address,
  Chain,
  Client,
  HttpTransport,
  ParseAccount,
  PublicActions,
  RpcSchema,
  Transport,
  WalletActions,
} from "viem";
import type { PublicActionsERC20, WalletActionsERC20 } from "./erc20.d.ts";
import type { L2Chain } from "../chains.ts";
import type {
  PublicActionsL1,
  PublicActionsL2,
  WalletActionsL1,
  WalletActionsL2,
} from "viem/op-stack";

export type PublicClients = publicClients<HttpTransport, Chain>;

type publicClients<
  transport extends Transport = Transport,
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = {
  l1: PublicL1Client<transport, chain, account>;
  l2: PublicL2Client<transport, L2Chain, account>;
};

export type WalletClients<
  accountOrAddress extends Account | Address | undefined = Account | undefined,
> = walletClients<HttpTransport, Chain, ParseAccount<accountOrAddress>>;

type walletClients<
  transport extends Transport = Transport,
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = {
  l1: WalletL1Client<transport, chain, account>;
  l2: WalletL2Client<transport, chain, account>;
};

export type PublicClient<
  transport extends Transport = Transport,
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = Client<
  transport,
  chain,
  account,
  RpcSchema,
  PublicActions<transport, chain> & PublicActionsERC20<chain, account>
>;

export type PublicL1Client<
  transport extends Transport = Transport,
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = Client<
  transport,
  chain,
  account,
  RpcSchema,
  PublicActions<transport, chain> &
    PublicActionsL1<chain, account> &
    PublicActionsERC20<chain, account>
>;

export type PublicL2Client<
  transport extends Transport = Transport,
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = Client<
  transport,
  chain,
  account,
  RpcSchema,
  PublicActions<transport, chain> &
    PublicActionsL2<chain, account> &
    PublicActionsERC20<chain, account>
>;

export type WalletL1Client<
  transport extends Transport = Transport,
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = Client<
  transport,
  chain,
  account,
  RpcSchema,
  WalletActions<chain, account> &
    WalletActionsL1<chain, account> &
    WalletActionsERC20<chain, account>
>;

export type WalletL2Client<
  transport extends Transport = Transport,
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = Client<
  transport,
  chain,
  account,
  RpcSchema,
  WalletActions<chain, account> &
    WalletActionsL2<chain, account> &
    WalletActionsERC20<chain, account>
>;
