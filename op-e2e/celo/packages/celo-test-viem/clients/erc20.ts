import type { Account, Chain, Client, Transport } from "viem";
import * as rmir from "reverse-mirage";
import type { PublicActionsERC20, WalletActionsERC20 } from "./erc20.d.ts";

export function publicActionsERC20() {
  return <
    transport extends Transport,
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<transport, chain, account>,
  ): PublicActionsERC20<chain, account> => {
    return {
      getERC20: (args) => rmir.getERC20(client, args),
      getERC20Allowance: (args) => rmir.getERC20Allowance(client, args),
      getERC20BalanceOf: (args) => rmir.getERC20BalanceOf(client, args),
      getERC20Decimals: (args) => rmir.getERC20Decimals(client, args),
      getERC20DomainSeparator: (args) =>
        rmir.getERC20DomainSeparator(client, args),
      getERC20Name: (args) => rmir.getERC20Name(client, args),
      getERC20Permit: (args) => rmir.getERC20Permit(client, args),
      getERC20PermitData: (args) => rmir.getERC20PermitData(client, args),
      getERC20PermitNonce: (args) => rmir.getERC20PermitNonce(client, args),
      getERC20Symbol: (args) => rmir.getERC20Symbol(client, args),
      getERC20TotalSupply: (args) => rmir.getERC20TotalSupply(client, args),
      getIsERC20Permit: (args) => rmir.getIsERC20Permit(client, args),
    };
  };
}

export function walletActionsERC20() {
  return <
    transport extends Transport,
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<transport, chain, account>,
  ): WalletActionsERC20<chain, account> => {
    return {
      signERC20Permit: (args) => rmir.signERC20Permit(client, args),
      simulateERC20Approve: (args) => rmir.simulateERC20Approve(client, args),
      simulateERC20Permit: (args) => rmir.simulateERC20Permit(client, args),
      simulateERC20Transfer: (args) => rmir.simulateERC20Transfer(client, args),
      simulateERC20TransferFrom: (args) =>
        rmir.simulateERC20TransferFrom(client, args),
    };
  };
}
