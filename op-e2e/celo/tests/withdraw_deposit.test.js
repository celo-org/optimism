import { withdraw } from '../src/withdraw.js'
import { deposit } from '../src/deposit.js'
import { parseEther } from 'viem'
import { setup } from './setup.js'

const minute = 60 * 1000
var config = {}

beforeAll(async () => {
  config = await setup()
}, minute)

test(
  'execute a withdraw and a deposit in succession',
  async () => {
    const celoToken = await config.client.l1.public.getERC20({
      erc20: {
        address: config.addresses.CustomGasTokenProxy,
        chainID: config.client.l1.public.chain.id,
      },
    })
    const balanceL1Before = await config.client.l1.public.getERC20BalanceOf({
      erc20: celoToken,
      address: config.account.address,
    })
    const balanceL2Before = await config.client.l2.public.getBalance({
      address: config.account.address,
    })
    const withdrawAmount = parseEther('1')
    const withdrawResult = await withdraw(
      {
        amount: withdrawAmount,
        to: config.account.address,
        gas: 21_000n,
      },
      config
    )
    expect(withdrawResult.success).toBe(true)
    const balanceL1AfterWithdraw =
      await config.client.l1.public.getERC20BalanceOf({
        erc20: celoToken,
        address: config.account.address,
      })
    const balanceL2AfterWithdraw = await config.client.l2.public.getBalance({
      address: config.account.address,
    })
    expect(balanceL1AfterWithdraw.amount).toBe(
      balanceL1Before.amount + BigInt(withdrawAmount)
    )
    expect(balanceL2AfterWithdraw).toBe(
      balanceL2Before - BigInt(withdrawAmount) - withdrawResult.l2GasPayment
    )
    const depositResult = await deposit(
      {
        mint: withdrawAmount,
        to: config.account.address,
      },
      config
    )
    expect(depositResult.success).toBe(true)

    const balanceL1AfterDeposit =
      await config.client.l1.public.getERC20BalanceOf({
        erc20: celoToken,
        address: config.account.address,
      })
    const balanceL2AfterDeposit = await config.client.l2.public.getBalance({
      address: config.account.address,
    })

    expect(balanceL1AfterDeposit.amount).toBe(balanceL1Before.amount)
    expect(balanceL2AfterDeposit).toBe(
      balanceL2Before - withdrawResult.l2GasPayment
    )
  },
  15 * minute
)
