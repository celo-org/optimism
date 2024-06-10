
import { parseEther } from 'viem'
import { getWithdrawals } from 'viem/op-stack'
import { account, publicClientL1, publicClientL2, walletClientL1, walletClientL2 } from './config'


const main = async function() {
  const txhash = process.argv[2]
  console.log('txhash:', txhash)

  const initiateHash = await walletClientL2.initiateWithdrawal({
    account,
    request: {
      gas: 21_000n,
      to: '0x70997970c51812dc3a010c7d01b50e0d17dc79c8',
      value: parseEther('1')
    },
  })

  const receipt = await publicClientL2.getTransactionConfirmations({
    hash: initiateHash,
  })

  const [withdrawal] = getWithdrawals(receipt)
  const output = await publicClientL1.getL2Output({
    l2BlockNumber: receipt.blockNumber,
    targetChain: publicClientL2.chain,
  })

  const args = await publicClientL2.buildProveWithdrawal({
    account,
    output,
    withdrawal,
  })

  const proveHash = await walletClientL1.proveWithdrawal(args)
}
main()
