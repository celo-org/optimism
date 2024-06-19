import { withdraw } from '../src/withdraw.js'
import { parseEther } from 'viem'
import { setup } from './setup.js'

const minute = 60 * 1000
const config = {}

beforeAll(async () => {
  config = await setup()
}, 30_000)

test.skip(
  'execute a withdraw (L2 native to L1 erc20)',
  async () => {
    const celoToken = await config.client.l1.public.getERC20({
      erc20: {
        address: config.addresses.CustomGasToken,
        chainID: config.client.l1.public.chain.id,
      },
    })
    const balanceBefore = await config.client.l1.public.getERC20BalanceOf({
      erc20: celoToken,
      address: config.account.address,
    })
    const withdrawAmount = parseEther('1')
    const success = await withdraw(
      {
        amount: withdrawAmount,
        to: config.account.address,
        gas: 21_000n,
      },
      config
    )
    expect(success).toBe(true)
    const balanceAfter = await config.client.l1.public.getERC20BalanceOf({
      erc20: celoToken,
      address: config.account.address,
    })
    expect(balanceAfter.amount).toBe(
      balanceBefore.amount + BigInt(withdrawAmount)
    )
  },
  3 * minute
)
