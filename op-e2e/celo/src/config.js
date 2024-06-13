import {
  createPublicClient,
  createWalletClient,
  custom,
  http,
  publicActions,
} from 'viem'
import { mainnet, optimism, optimismGoerli } from 'viem/chains'
import { cel2, testnetl1 } from './chain.js'
import {
  publicActionsL1,
  publicActionsL2,
  walletActionsL1,
  walletActionsL2,
} from 'viem/op-stack'

// TODO: check what contracts have to be defined on the l1 chain
// and implement them in the chain
export const publicClientL1 = createPublicClient({
  batch: {
    multicall: false,
  },
  chain: testnetl1,
  transport: http(),
}).extend((client) => {
  var i = publicActionsL1()(client)
  i.multicall = (args) => multicallBackport(client, args)
  return i
})

const multicallBackport = function(client, args) {
  console.log('called multicall itnercept')
  return false
}

export const walletClientL1 = createWalletClient({
  batch: {
    multicall: false,
  },
  chain: testnetl1,
  transport: http(),
}).extend(walletActionsL1())

export const walletClientL2 = createWalletClient({
  chain: cel2,
  transport: http(),
}).extend(walletActionsL2())

export const publicClientL2 = createPublicClient({
  chain: cel2,
  transport: http(),
}).extend(publicActionsL2())
