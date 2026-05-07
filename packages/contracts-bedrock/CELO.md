# Celo notes

L1-side deltas between Celo's fork and upstream OP at v5.0.0.

## Dual guardian via CeloSuperchainConfig

`src/celo/CeloSuperchainConfig.sol` wraps a `SuperchainConfig` and adds a Celo-side guardian. Either guardian can pause.

ABI:

- `initialize(address guardian, bool paused, address superchainConfig)`
- `paused()` / `paused(address identifier)` — true if Celo paused locally OR the wrapped SuperchainConfig is paused
- `pause(string identifier)` / `unpause()` — Celo guardian only
- `checkAndPauseIfSuperchainPaused()` — copies the wrapped SuperchainConfig's paused state into local

State lives in unstructured (`keccak256`-derived) slots, so it doesn't show in `forge inspect storage-layout`.

## How L1 contracts reach the CeloSuperchainConfig

L1 consumers (L1CrossDomainMessenger, L1StandardBridge, L1ERC721Bridge, OptimismPortal2, AnchorStateRegistry, DelayedWETH) don't hold the CeloSuperchainConfig as a state var. They hold `SystemConfig`, and `SystemConfig.superchainConfig` points at the CeloSuperchainConfig.

```
consumer.paused() → systemConfig.paused() → celoSuperchainConfig.paused(id)
                                          → celoSuperchainConfig.celoPaused()
                                            || externalSuperchainConfig.paused(id)
```

Going through `SystemConfig` keeps our diff against upstream small. Behaviour is the same as wiring the CeloSuperchainConfig directly.

## External SuperchainConfig

On Celo Mainnet and Celo Sepolia the CeloSuperchainConfig wraps the upstream Superchain Registry's SuperchainConfig.

`DeployConfig.externalSuperchainConfig` selects the path:

| Value | Deploy entrypoint | Use case |
|---|---|---|
| `0x0` (or unset) | `Deploy.run()` | Devnet |
| Set to an address | `Deploy.runCelo(protocolVersionsProxy)` | Celo Mainnet, Celo Sepolia |

`runCelo` deploys the CeloSuperchainConfig proxy and passes its address to `OPContractsManager.deploy()` through the optional `DeployInput.superchainConfigOverride` field, so `SystemConfig.superchainConfig` is initialized to the CeloSuperchainConfig instead of OPCM's own SuperchainConfig. End-to-end coverage against the real Mainnet and Sepolia SuperchainConfig addresses lives in `test/L1/RunCelo.t.sol`.

For tests, call `CommonTest.enableExternalSuperchainConfig(addr)` before `setUp()`. See `test/setup/ExternalSuperchainConfig.t.sol`.

## CeloSuperchainConfig is not part of OPCM upgrades

`OPCM.upgrade(OpChainConfig[], bool)` does not touch the CeloSuperchainConfig implementation. The CeloSuperchainConfig is stable across OP releases. If we ever need to upgrade it, it's a one-shot operation (like the v3 → v4.1 migration was).

## Deploy

1. L1 contracts:
   - Devnet: `Deploy.run()`
   - Celo Mainnet, Celo Sepolia: `Deploy.runCelo(<protocolVersionsProxy>)`
2. Generate L2 genesis: `make generate_l2_allocs` or run `L2Genesis.s.sol` directly.
3. Feed allocs into the migration tool.
4. Start sequencer + nodes.

## CustomGasToken

We kept CustomGasToken after upstream removed it.

- `CeloTokenL1` is deployed with a 1B CELO balance and used as the gas token.
- We use `OptimismPortal2` (FaultProofs).
- Migrating an existing balance into the Portal uses a storage-setter trick: a temporary impl is deployed to the Portal proxy to write a single slot, then the real impl takes over.

## L2 contracts

- Devnet: required Celo core contracts are added in `L2Genesis.s.sol`. Celo Mainnet and Celo Sepolia skip this — the migration tool brings the existing state across.
- `OptimismMintableERC20` mixes in `AbstractFeeCurrency` so any natively-bridged ERC20 can be whitelisted as a fee currency.
