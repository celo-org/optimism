import { chainConfig } from 'viem/op-stack'
import { defineChain } from 'viem'

export function makeChainConfigs(l1ChainID, l2ChainID, contractAddresses) {
  return {
    l2: defineChain({
      formatters: {
        ...chainConfig.formatters,
      },
      serializers: {
        ...chainConfig.serializers,
      },
      id: l2ChainID,
      name: 'Celo',
      nativeCurrency: {
        decimals: 18,
        name: 'Celo - native currency',
        symbol: 'CELO',
      },
      rpcUrls: {
        default: {
          http: ['http://localhost:9545'],
        },
      },
      contracts: {
        ...chainConfig.contracts,
        l2OutputOracle: {
          [l1ChainID]: {
            address: contractAddresses.L2OutputOracleProxy,
          },
        },
        portal: {
          [l1ChainID]: {
            address: contractAddresses.OptimismPortalProxy,
          },
        },
        l1StandardBridge: {
          [l1ChainID]: {
            address: contractAddresses.L1StandardBridgeProxy,
          },
        },
      },
    }),
    l1: defineChain({
      id: l1ChainID,
      testnet: true,
      name: 'Ethereum L1',
      nativeCurrency: {
        decimals: 18,
        name: 'Ether',
        symbol: 'ETH',
      },
      rpcUrls: {
        default: {
          http: ['http://localhost:8545'],
        },
      },
      contracts: {
        multicall3: {
          address: contractAddresses.Multicall3,
        },
      },
    }),
  }
}
