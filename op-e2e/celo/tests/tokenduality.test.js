import { createAmountFromString } from 'reverse-mirage'
import { setup } from './setup.js'

const minute = 60 * 1000
const config = {}

beforeAll(async () => {
  config = await setup()
}, 30_000)

test(
  'test token duality',
  async () => {
    const receiverAddr = '0x000000000000000000000000000000000000dEaD'
    const dualityToken = await config.client.l2.public.getERC20({
      erc20: {
        address: '0x471ece3750da237f93b8e339c536989b8978a438',
        chainID: config.client.l2.public.chain.id,
      },
    })
    console.log(dualityToken)
    const balanceBefore = await config.client.l2.public.getBalance({
      address: receiverAddr,
    })

    const sendAmount = createAmountFromString(dualityToken, '100')
    console.log(sendAmount)
    const { request } = await config.client.l2.wallet.simulateERC20Transfer({
      to: receiverAddr,
      amount: sendAmount,
    })
    console.log(request)
    const transferHash = await config.client.l2.wallet.writeContract(request)
    const receipt = await config.client.l2.public.waitForTransactionReceipt({
      hash: transferHash,
    })
    console.log(receipt)
    const balanceAfter = await config.client.l2.public.getBalance({
      address: receiverAddr,
    })

    expect(balanceAfter.amount).toBe(balanceBefore.amount + sendAmount.amount)
  },
  1 * minute
)
