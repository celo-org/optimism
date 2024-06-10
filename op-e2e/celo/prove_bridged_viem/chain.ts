import { chainConfig } from 'viem/op-stack'
import { optimismGoerli } from 'viem/chains'
import { readFileSync } from 'fs'
import { defineChain } from 'viem'

const contract_addrs = JSON.parse(readFileSync('../../.devnet/addresses.json', 'utf8'))

// const contract_addresses = {
//   AddressManager: contract_addrs.AddressManager,
//   CustomTokenAddress: contract_addrs.CustomTokenAddress,
//   L1CrossDomainMessenger: contract_addrs.L1CrossDomainMessengerProxy,
//   L1StandardBridge: contract_addrs.L1StandardBridgeProxy,
//   OptimismPortal: contract_addrs.OptimismPortalProxy,
//   L2OutputOracle: contract_addrs.L2OutputOracleProxy,
//
//   // These are only needed for backwards compatibility and are not used
//   StateCommitmentChain: "0x".padEnd(42, "0"),
//   CanonicalTransactionChain: "0x".padEnd(42, "0"),
//   BondManager: "0x".padEnd(42, "0"),
// }

//TODO: do we need this, e.g. for the contracts? what do we need?
export const cel2 = defineChain({
  ...chainConfig.contracts,
  ...chainConfig.formatters,
  ...chainConfig.serializers,
  //TODO: check the id
  id: 7777777,
  name: 'Celo',
  nativeCurrency: {
    decimals: 18,
    name: 'Celo - native currency',
    symbol: 'CELO',
  },
  rpcUrls: {
    default: {
      http: [process.env.ETH_RPC_URL],
    },
  },
})

//TODO: check the id
const sourceId = 42;
export const testnetl1 = defineChain({
  id: sourceId,
  testnet: true,
  name: 'CeloL1',
  nativeCurrency: {
    decimals: 18,
    name: 'Ether',
    symbol: 'ETH',
  },
  rpcUrls: {
    default: {
      http: [process.env.ETH_RPC_URL_L1],
    },
  },
  ...chainConfig.contracts,
  l2OutputOracle: {
    [sourceId]: {
      address: contract_addrs.L2OutputOracleProxy,
    },
  },
  portal: {
    [sourceId]: {
      address: contract_addrs.OptimismPortalProxy,
    },
  },
  l1StandardBridge: {
    [sourceId]: {
      address: contract_addrs.L1StandardBridgeProxy,
    },
  },
})

