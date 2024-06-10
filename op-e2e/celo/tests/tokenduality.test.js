import { createAmountFromString } from 'reverse-mirage'
import { setup } from './setup.js'

const minute = 60 * 1000
let config = {}

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
    const balanceBefore = await config.client.l2.public.getBalance({
      address: receiverAddr,
    })

    const sendAmount = createAmountFromString(dualityToken, '100')
    const { request } = await config.client.l2.wallet.simulateERC20Transfer({
      to: receiverAddr,
      amount: sendAmount,
    })
    const transferHash = await config.client.l2.wallet.writeContract(request)
    const receipt = await config.client.l2.public.waitForTransactionReceipt({
      hash: transferHash,
    })
    expect(receipt.status).toBe('success')
    const balanceAfter = await config.client.l2.public.getBalance({
      address: receiverAddr,
    })

    expect(balanceAfter).toBe(balanceBefore + sendAmount.amount)
  },
  1 * minute
)
