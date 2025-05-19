# Celo Diff

- Since we are using the CustomGasFeature (which is currently being deprecated from OP), we cannot use the OP Deployer to deploy contracts. Instead, we need to deploy contracts using the `Deploy.s.sol` script.
- The `Deploy.s.sol` script is not from the original 1.8.0 smart contract branch, as it was likely never used for testnet deployment (due to the use of Anvil cheat codes). We ported the `Deploy.s.sol` script from our Celo client branch, specifically based on version 1.9.3.

## L1 Deploy Script Notes (`Deploy.s.sol`)
- `CeloTokenL1` is deployed with a balance of 1 billion CELO and is used as the Custom Gas token.
- Due to FaultProofs, we are using `OptimismPortal2` rather than `OptimismPortal`.
- Since we are using the Custom Gas Token feature with a preexisting balance, we use a storage setter design pattern (deploying a different implementation to the OptimismPortal proxy during the transition period), which allows us to set a specific storage slot to the desired value.

## L2 Deploy Script Notes (`L2Genesis.s.sol`)
- The required Celo core contracts were added to `L2Genesis.s.sol` for devnet deployment. For mainnet, this step is skipped, and we instead rely on the blockchain migration tool for the entire blockchain state.

## CeloSuperchainConfig
- The Superchain has the power to pause/unpause the Celo network. To enable us to unpause the network when necessary, we introduced `CeloSuperchainConfig`, which adds a Guardian account with the same permissions as the Superchain. This allows us to pause/unpause the network in case of an emergency.

## AbstractFeeCurrency
- We added `AbstractFeeCurrency` to the bridged ERC20 token (specifically, `OptimismMintableERC20.sol`), which implements credit and debit functions. This enhancement allows us to whitelist and use any natively bridged token as the native fee currency.

# Release Mainnet Process

1. L1 smart contracts are deployed using the `Deploy.s.sol` script.
2. The L2 Genesis is created either by running `make generate_l2_allocs` or by manually using the `L2Genesis.s.sol` script (which creates the L2 allocs file).
3. The L2 allocs file is fed into the blockchain migration tool.
4. The sequencer and nodes are started.
