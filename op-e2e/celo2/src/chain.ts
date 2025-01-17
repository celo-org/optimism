import { chainConfig } from "viem/op-stack";
import { Address, defineChain } from "viem";
import { AddressesType } from "./setup";

export type ChainConfigL1L2 = ReturnType<typeof makeChainConfigs>;

export function makeChainConfigs(
  l1ChainID: number,
  l2ChainID: number,
  contractAddresses: AddressesType,
) {
  console.log("chainConfigchainConfig", chainConfig);
  console.log(process.env);
  return {
    l2: defineChain({
      formatters: {
        ...chainConfig.formatters,
      },
      serializers: {
        ...chainConfig.serializers,
      },
      id: l2ChainID,
      name: "Celo",
      nativeCurrency: {
        decimals: 18,
        name: "Celo - native currency",
        symbol: "CELO",
      },
      rpcUrls: {
        default: {
          http: [/*process.env.ETH_RPC_URL*/ "http://localhost:9545"],
        },
      },
      contracts: {
        ...chainConfig.contracts,
        l2OutputOracle: {
          [l1ChainID]: {
            address: contractAddresses.L2OutputOracleProxy as Address,
          },
        },
        disputeGameFactory: {
          [l1ChainID]: {
            address: contractAddresses.DisputeGameFactoryProxy as Address,
          },
        },
        portal: {
          [l1ChainID]: {
            address: contractAddresses.OptimismPortalProxy as Address,
          },
        },
        l1StandardBridge: {
          [l1ChainID]: {
            address: contractAddresses.L1StandardBridgeProxy as Address,
          },
        },
      },
    }),
    l1: defineChain({
      id: l1ChainID,
      testnet: true,
      name: "Ethereum L1",
      nativeCurrency: {
        decimals: 18,
        name: "Ether",
        symbol: "ETH",
      },
      rpcUrls: {
        default: {
          http: [/*process.env.ETH_RPC_URL_L1*/ "http://localhost:8545"],
        },
      },
      contracts: {
        multicall3: {
          address: contractAddresses.Multicall3 as Address,
        },
        l2OutputOracle: {
          [l1ChainID]: {
            address: contractAddresses.L2OutputOracleProxy as Address,
          },
        },
        disputeGameFactory: {
          [l1ChainID]: {
            address: contractAddresses.DisputeGameFactoryProxy as Address,
          },
        },
        portal: {
          [l1ChainID]: {
            address: contractAddresses.OptimismPortalProxy as Address,
          },
        },
        l1StandardBridge: {
          [l1ChainID]: {
            address: contractAddresses.L1StandardBridgeProxy as Address,
          },
        },
      },
    }),
  };
}
