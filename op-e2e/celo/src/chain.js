import { chainConfig } from 'viem/op-stack'
import { optimismGoerli, goerli } from 'viem/chains'
import { defineChain } from 'viem'

// const l1ChainID = 900
// const l2ChainID = 901

// const contract_addresses = {
//   AddressManager: contract_addrs.AddressManager,
//   CustomGasToken: contract_addrs.CustomGasToken,
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

export function makeChainConfigs(l1ChainID, l2ChainID, contractAddresses) {
  return {
    l2: defineChain({
      ...chainConfig.formatters,
      ...chainConfig.serializers,
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
        // l2OutputOracle: {
        //   [l1ChainID]: {
        //     address: contractAddresses.L2OutputOracleProxy,
        //   },
        // },
        // portal: {
        //   [l1ChainID]: {
        //     address: contractAddresses.OptimismPortalProxy,
        //   },
        // },
        // l1StandardBridge: {
        //   [l1ChainID]: {
        //     address: contractAddresses.L1StandardBridgeProxy,
        //   },
        // },
      },
    }),
  }
}
