import { setupClients } from '../src/config.js'
import { makeChainConfigs } from '../src/chain.js'
import { privateKeyToAccount } from 'viem/accounts'
import { readFileSync } from 'fs'

const privKey =
  '0x92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e'

async function waitForNoError(func, timeout) {
  const start = Date.now()
  while (Date.now() - start < timeout) {
    try {
      await func()
      return true
    } catch (error) { }
    await new Promise((r) => setTimeout(r, 1000))
  }
  return false
}

async function waitReachable(client, timeout) {
  const f = async () => client.getChainId()
  return waitForNoError(f, timeout)
}

async function waitForNextL2Output(client, l2ChainConfig, timeout) {
  const f = async () =>
    client.waitForNextL2Output({
      pollingInterval: 500,
      l2BlockNumber: 0,
      targetChain: l2ChainConfig,
    })
  return waitForNoError(f, timeout)
}

export async function setup() {
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
    waitForNextL2Output(config.client.l1.public, chainConfig.l2, 60_000),
  ])
  if (success.every((v) => v == true)) {
    return config
  }
  throw new Error('l1 and l2 clients not reachable within the deadline')
}
