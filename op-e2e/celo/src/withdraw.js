import { parseEther } from 'viem'
import { getWithdrawals } from 'viem/op-stack'
import { privateKeyToAccount } from 'viem/accounts'
import {
  publicClientL1,
  publicClientL2,
  walletClientL1,
  walletClientL2,
} from './config.js'

export const withdraw = async function (amount, opts) {
  // const txhash = process.argv[2]
  // console.log('txhash:', txhash)
  //
  //process.env.ACC_PRIVKEY
  const account = privateKeyToAccount(opts.privKey)
  const initiateHash = await walletClientL2.initiateWithdrawal({
    account,
    request: {
      gas: 21_000n,
      to: '0x70997970c51812dc3a010c7d01b50e0d17dc79c8',
      value: parseEther(amount),
    },
  })

  console.log('withdraw hash', initiateHash)

  const receipt = await publicClientL2.waitForTransactionReceipt({
    hash: initiateHash,
  })
  console.log('receipt', receipt)

  // FIXME: this only works when the mulitcall3 contract
  // is deployed on the L1 chain...
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
  return true
}
