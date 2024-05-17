Rebasing on top of `9047beb54`


* 3636885a1 Implement Cel2 migration tool (#110) (HEAD -> celo4, origin/celo4)
--^ TODO: will require changes, @palango to redo after rebase
* e5dd62e90 Update semver-lock
--^ Note: Skipped
* 3057080cd Update snapshots
--^ Note: Skipped
* a69c1a24e Update op-bindings
--^ Note: Skipped
* 7e243f576 Update soldity pragma for interfaces
* 1a9a2daf3 Add fee currency functions to MintableERC20
* 70d151fa6 Update snapshots
--^ Note: Skipped
* 3f892a23b Update op-bindings
--^ Note: Skipped
* 98670b5f5 Add more powerful MockSortedOracles
* 063933de7 Initial version of state migration tool (origin/karlb/start-from-upstream-6)
--^ Note: Skipped
* 21564a5ef Update gas-snapshot
--^ Note: Skipped
* 30dde72e4 Update semver-lock
--^ Note: Skipped
* 160d7f76f Update contract snapshots
--^ Note: Skipped
* fa649f77c Update op-bindings after bridging changes
--^ Note: Skipped
* c3621943c L2->L1 ETH bridging unit test
* 50d96ec27 Fix tests after bridging changes
* 9533dd8e4 Test bridging
* aa86a4112 Change bridgeETH(To) calls to ETHToERC20 bridging
* d9cf404a6 Configure L2->L1 ETH bridging
--^ TODO: Check TODO
* 0794b3e25 Configure L1->L2 ETH bridging
--^ TODO: Test, just did a naive merge
* 946be0b6d Add cel2 testnet setup code
* e2ed217fc Run new e2e tests as part of devnet CI job
* 07b5082ec Add token duality e2e test
* ca25d49e0 Enable Cel2 in e2e tests
* 61982e7c7 Change base fee recipient in TestFees
* 2d6a00063 Use celo-org/op-geth container for devnet
* a185157b6 Enable cel2 fork by default
* 3424eab24 Update op-bindings
--^ TODO: Skipped for now
* dd0715736 Remove common/libraries/ReentrancyGuard.sol (duplicate)
--^ Note: Squashed in last commit
* 4a6e4c427 Remove common/interfaces/IExchange.sol (duplicate)
--^ Note: Squashed in last commit
* 9988f02e6 Add Celo contracts
--^ TODO: Did not reimplement contract deployment in L2 again.
* 1fd85bdfb Add forkdiff comparison to optimism (#32)
* 3bedc9674 New parameter due to FC exchange rate
--^ Note: not required any more
* ba6b969af Add fee currency parameter to IntrinsicGas
* df923fe4f Update op-geth
* f0f74bd01 adding trivy scanning to the Docker files (#41)
* 8888ba337 dependabot: no PRs for version updates
* df5a8fefe Simplify CI

### Notes
- Bindings are not global anymore. See:
	- https://github.com/ethereum-optimism/optimism/pull/10466
	- https://github.com/ethereum-optimism/optimism/pull/10338
	- Should we add a file just like `op-chain-ops/justfile`?
- Fix this line: ` // TODO: Check if this can be uncommented`
- Fix shellcheck