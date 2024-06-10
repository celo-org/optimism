import { createPublicClient, createWalletClient, custom, http } from 'viem'
import { privateKeyToAccount } from 'viem/accounts'
import { mainnet, optimism, optimismGoerli } from 'viem/chains'
import { cel2, testnetl1 } from './chain'
import { publicActionsL1, publicActionsL2, walletActionsL1, walletActionsL2 } from 'viem/op-stack'


// TODO: check what contracts have to be defined on the l1 chain
// and implement them in the chain
export const publicClientL1 = createPublicClient({
  chain: testnetl1,
  transport: http()
}).extend(publicActionsL1())

export const walletClientL1 = createWalletClient({
  chain: testnetl1,
  transport: http()
}).extend(walletActionsL1())

export const walletClientL2 = createWalletClient({
  chain: cel2,
  transport: http()
}).extend(walletActionsL2())

export const publicClientL2 = createPublicClient({
  chain: cel2,
  transport: http()
}).extend(publicActionsL2())

export const account = privateKeyToAccount(process.env.ACC_PROVER_PRIVKEY)

