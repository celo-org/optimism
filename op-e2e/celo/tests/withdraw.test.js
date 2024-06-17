import { withdraw } from '../src/withdraw.js'
import { parseEther } from 'viem'
import { exec, spawn, execSync } from 'node:child_process'
import { promisify } from 'node:util'

// const execP = promisify(exec)

const execOpts = {
  env: {
    ...process.env,
    DEVNET_CELO: true,
  },
  //FIXME: don't hardcode
  cwd: '/Users/maximilian/code/optimism',
}

// beforeAll(() => {
//   console.log(execSync('"make" devnet-up', execOpts))
// })
// afterAll(() => {
//   console.log(execSync('"make" devnet-down', execOpts))
// })
//
//
const minute = 60 * 1000
const privKey =
  '0x92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e'
const config = {
  account: privateKeyToAccount(privKey),
}

test(
  'execute a withdraw (L2 to L1)',
  async () => {
    // TODO: balance before
    expect(
      withdraw(
        {
          amount: parseEther('1'),
          to: '0x70997970c51812dc3a010c7d01b50e0d17dc79c8',
        },
        config
      )
    ).resolves.toBe(true)
    // TODO: balance after == +amount
    // TODO: check the currency
  },
  2 * minute
)
