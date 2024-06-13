import { withdraw } from '../src/withdraw.js'
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

const opts = {
  privKey: '0x92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e',
}

// beforeAll(() => {
//   console.log(execSync('"make" devnet-up', execOpts))
// })
// afterAll(() => {
//   console.log(execSync('"make" devnet-down', execOpts))
// })

test(
  'tries a simple withdraw',
  async () => {
    const success = await withdraw('1', opts)
    expect(success).toBe(true)
  },
  5 * 60 * 1000
)
