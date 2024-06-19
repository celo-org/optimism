import { createPublicClient, createWalletClient, http } from 'viem'
import { readFileSync } from 'fs'
import { readContract } from 'viem/actions'
import {
  getERC20,
  simulateERC20Transfer,
  getERC20BalanceOf,
  getERC20Symbol,
  getERC20Decimals,
} from 'reverse-mirage'
import {
  publicActionsL1,
  publicActionsL2,
  walletActionsL1,
  walletActionsL2,
} from 'viem/op-stack'

export function makeReadContract(contractAddress, contractABI) {
  return (client) => {
    return {
      readContract: (args) => {
        const rcArgs = {
          address: contractAddress,
          abi: contractABI,
          functionName: args.functionName,
          args: args.args,
        }
        return readContract(client, rcArgs)
      },
    }
  }
}

export function erc20PublicActions(client) {
  return {
    getERC20: (args) => getERC20(client, args),
    getERC20Symbol: (args) => getERC20Symbol(client, args),
    getERC20BalanceOf: (args) => getERC20BalanceOf(client, args),
    getERC20Decimals: (args) => getERC20Decimals(client, args),
  }
}
export function erc20WalletActions(client) {
  return {
    simulateERC20Transfer: (args) => {
      return simulateERC20Transfer(client, { args: args })
    },
  }
}

export function setupClients(l1ChainConfig, l2ChainConfig, account, addresses) {
  return {
    l1: {
      public: createPublicClient({
        account,
        chain: l1ChainConfig,
        transport: http(),
      })
        .extend(publicActionsL1())
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
        account,
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
