import { Account, createPublicClient, createWalletClient, Hex, http, PublicClient, ReadContractParameters, WalletClient } from 'viem'
import { readContract } from 'viem/actions'
import { constructDepositCustomGas, ConstructDepositCustomGasParameters } from './deposit'
import {
  getERC20,
  simulateERC20Transfer,
  getERC20BalanceOf,
  getERC20Symbol,
  getERC20Decimals,
  simulateERC20Approve,
  getERC20Allowance,
} from 'reverse-mirage'
import {
  publicActionsL1,
  publicActionsL2,
  walletActionsL1,
  walletActionsL2,
} from 'viem/op-stack'
import { ChainConfigL1L2 } from './chain'

interface ReadContractArgs {
  functionName: string
  args: any[]
}

export function makeReadContract(contractAddress: Hex, contractABI: any) {
  return (client: any) => ({
    readContract: (args: ReadContractArgs) => {
      const rcArgs: ReadContractParameters = {
        address: contractAddress,
        abi: contractABI,
        functionName: args.functionName,
        args: args.args,
      }
      return readContract(client, rcArgs)
    },
  })
}

export function erc20PublicActions(client: any) {
  return {
    getERC20: (args: Parameters<typeof getERC20>[1]) => getERC20(client, args),
    getERC20Symbol: (args: Parameters<typeof getERC20Symbol>[1]) => getERC20Symbol(client, args),
    getERC20BalanceOf: (args: Parameters<typeof getERC20BalanceOf>[1]) => getERC20BalanceOf(client, args),
    getERC20Decimals: (args: Parameters<typeof getERC20Decimals>[1]) => getERC20Decimals(client, args),
    getERC20Allowance: (args: Parameters<typeof getERC20Allowance>[1]) => getERC20Allowance(client, args),
  }
}

export function erc20WalletActions(client: any) {
  return {
    simulateERC20Transfer: (args: any) =>
      simulateERC20Transfer(client, { args }),
    simulateERC20Approve: (args: any) =>
      simulateERC20Approve(client, { args }),
  }
}

export function celoL1PublicActions(client: any) {
  return {
    prepareDepositGasPayingTokenERC20: (args: ConstructDepositCustomGasParameters) =>
      constructDepositCustomGas(client, args),
  }
}

export type SetupClientsReturn = ReturnType<typeof setupClients>;

export function setupClients(
  chainConfigsL1L2: ChainConfigL1L2,
  account: Account
) {
  const { l1: l1ChainConfig, l2: l2ChainConfig } = chainConfigsL1L2

  return {
    l1: {
      public: createPublicClient({
        transport: http(),
        chain: l1ChainConfig,
      })
        .extend(publicActionsL1())
        .extend(celoL1PublicActions)
        .extend(erc20PublicActions),
      wallet: createWalletClient({
        account,
        chain: l1ChainConfig,
        transport: http(),
      })
        .extend(erc20WalletActions)
        .extend(walletActionsL1()),
    },
    l2: {
      public: createPublicClient({
        chain: l2ChainConfig,
        transport: http(),
      })
        .extend(publicActionsL2())
        .extend(erc20PublicActions),
      wallet: createWalletClient({
        account,
        chain: l2ChainConfig,
        transport: http(),
      })
        .extend(erc20WalletActions)
        .extend(walletActionsL2()),
    },
  }
}
