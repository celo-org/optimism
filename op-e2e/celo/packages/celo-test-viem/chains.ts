import { chainConfig as l2ChainConfig } from "viem/op-stack";
import { defineChain } from "viem";
import type { Address, Chain, ChainContract, Prettify } from "viem";

export type L2Chain = Chain<typeof l2ChainConfig.formatters>;
export type Chains = {
  l1: Chain;
  l2: L2Chain;
};

export type ChainContractsCeloL2 = Prettify<
  {
    [key: string]: ChainContract | undefined;
  } & {
    goldToken?: ChainContract | undefined;
    registry?: ChainContract | undefined;
    feeCurrencyDirectory?: ChainContract | undefined;
  }
>;

export interface ContractAddressesL1 {
  AddressManager: Address;
  AnchorStateRegistry: Address;
  AnchorStateRegistryProxy: Address;
  CustomGasTokenProxy: Address;
  DelayedWETH: Address;
  DelayedWETHProxy: Address;
  DisputeGameFactory: Address;
  DisputeGameFactoryProxy: Address;
  FastPreimageOracle: Address;
  FaultDisputeGame_0: Address;
  FaultDisputeGame_254: Address;
  FaultDisputeGame_255: Address;
  L1CrossDomainMessenger: Address;
  L1CrossDomainMessengerProxy: Address;
  L1ERC721Bridge: Address;
  L1ERC721BridgeProxy: Address;
  L1StandardBridge: Address;
  L1StandardBridgeProxy: Address;
  L2OutputOracleProxy: Address;
  Mips: Address;
  Multicall3: Address;
  OPContractsManager: Address;
  OPContractsManagerProxy: Address;
  OptimismMintableERC20Factory: Address;
  OptimismMintableERC20FactoryProxy: Address;
  OptimismPortal2: Address;
  OptimismPortalProxy: Address;
  PermissionedDelayedWETHProxy: Address;
  PermissionedDisputeGame: Address;
  PreimageOracle: Address;
  ProtocolVersions: Address;
  ProtocolVersionsProxy: Address;
  ProxyAdmin: Address;
  SuperchainConfig: Address;
  SuperchainConfigProxy: Address;
  SuperchainProxyAdmin: Address;
  SystemConfig: Address;
  SystemConfigProxy: Address;
}

export function makeChainConfigs(
  l1ChainID: number,
  l2ChainID: number,
  rpcUrlL1: string,
  rpcUrlL2: string,
  contractAddressesL1: ContractAddressesL1,
  contractAddressesL2Celo: ChainContractsCeloL2,
): Chains {
  return {
    l2: defineChain({
      formatters: {
        ...l2ChainConfig.formatters,
      },
      serializers: {
        ...l2ChainConfig.serializers,
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
          http: [rpcUrlL2],
        },
      },
      contracts: {
        ...l2ChainConfig.contracts,
        ...contractAddressesL2Celo,
        optimismMintableERC20Factory: {
          address: "0x4200000000000000000000000000000000000012",
          [l1ChainID]: {
            address: contractAddressesL1.OptimismMintableERC20FactoryProxy,
          },
        },
        customGasToken: {
          [l1ChainID]: {
            address: contractAddressesL1.CustomGasTokenProxy,
          },
        },
        l2OutputOracle: {
          [l1ChainID]: {
            address: contractAddressesL1.L2OutputOracleProxy,
          },
        },
        disputeGameFactory: {
          [l1ChainID]: {
            address: contractAddressesL1.DisputeGameFactoryProxy,
          },
        },
        portal: {
          [l1ChainID]: {
            address: contractAddressesL1.OptimismPortalProxy,
          },
        },
        l1StandardBridge: {
          [l1ChainID]: {
            address: contractAddressesL1.L1StandardBridgeProxy,
          },
        },
      },
      sourceId: l1ChainID,
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
          http: [rpcUrlL1],
        },
      },
      contracts: {
        customGasToken: {
          address: contractAddressesL1.CustomGasTokenProxy,
        },
        multicall3: {
          address: contractAddressesL1.Multicall3,
        },
      },
    }),
  };
}
