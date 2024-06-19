import { setupClients } from '../src/config.js'
import { makeChainConfigs } from '../src/chain.js'
import { privateKeyToAccount } from 'viem/accounts'
import { readFileSync } from 'fs'

const privKey =
  '0x92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e'

async function waitReachable(client, timeout) {
  const start = Date.now()
  while (Date.now() - start < timeout) {
    try {
      await client.getChainId()
      return true
    } catch (error) {
      console.error(error)
    }
    await new Promise((r) => setTimeout(r, 500))
  }
  return false
}

export async function setup() {
  // NOTE: `DEVNET_CELO=true make devnet-allocs` has to be
  // called at least once before that.
  // The tests will be run after the devnet has started anyways,
  // so this is given anyways.
  const contractAddrs = JSON.parse(
    readFileSync('../../.devnet/addresses.json', 'utf8')
  )
  const config = { account: privateKeyToAccount(privKey) }
  const chainConfig = makeChainConfigs(900, 901, contractAddrs)

  config.client = setupClients(
    chainConfig.l1,
    chainConfig.l2,
    config.account,
    contractAddrs
  )
  config.addresses = contractAddrs

  const success = await Promise.all([
    waitReachable(config.client.l1.public, 10_000),
    waitReachable(config.client.l2.public, 10_000),
  ])
  if (success.every((v) => v == true)) {
    return config
  }
  throw new Error('l1 and l2 clients not reachable within the deadline')
}
