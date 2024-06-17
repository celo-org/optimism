import { getWithdrawals } from 'viem/op-stack'
import { privateKeyToAccount } from 'viem/accounts'
import {
  publicClientL1,
  publicClientL2,
  walletClientL1,
  walletClientL2,
} from './config.js'

export const withdraw = async function(args, config) {
  account = config.account
  const initiateHash = await walletClientL2.initiateWithdrawal({
    account,
    request: {
      gas: 21_000n,
      to: args.to,
      value: args.amount,
    },
  })

  console.log('withdraw hash', initiateHash)

  const receipt = await publicClientL2.waitForTransactionReceipt({
    hash: initiateHash,
  })
  console.log('receipt', receipt)

  // this function requires the mulitcall contract to be deployed
  // on the L1 chain
  const { output, withdrawal } = await publicClientL1.waitToProve({
    receipt,
    targetChain: publicClientL2.chain,
  })
  console.log('output', output)
  console.log('withdrawal', withdrawal)

  const args = await publicClientL2.buildProveWithdrawal({
    account,
    output,
    withdrawal,
  })
  const proveHash = await walletClientL1.proveWithdrawal(args)
  const proveReceipt = await publicClientL1.waitForTransactionReceipt({
    hash: proveHash,
  })
  return proveReceipt
}
