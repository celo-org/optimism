// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";
import { IMulticall3 } from "forge-std/interfaces/IMulticall3.sol";

// Contracts
import { PortalMigrator } from "src/celo/PortalMigrator.sol";

// Libraries
import { Claim } from "src/dispute/lib/LibUDT.sol";
import { GameType, GameTypes } from "src/dispute/lib/Types.sol";
import { IDisputeGameFactory } from "interfaces/dispute/IDisputeGameFactory.sol";

// Interfaces
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ICeloGasBridgeL1 } from "interfaces/celo/ICeloGasBridgeL1.sol";
import { IOPContractsManager } from "interfaces/L1/IOPContractsManager.sol";
import { IOptimismPortal2 } from "interfaces/L1/IOptimismPortal2.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { ISemver } from "interfaces/universal/ISemver.sol";
import { SemverComp } from "src/libraries/SemverComp.sol";

/// @notice Inputs for the `MigrateV1V2` migration script. Each phase calls `_validate` first, so
///         the rest of the code can read fields without guarding them.
struct MigrateV1V2Input {
    address payable portalProxy;
    address systemConfig;
    address proxyAdmin;
    address opcm;
    address safe;
    address portalMigratorImpl;
    address originalPortalImpl;
    address bridgeL1Proxy;
    address storageSetter;
    address externalSuperchainConfig;
    address succinctGameImpl;
    IERC20 celoTokenL1;
    uint256 legacyPortalBalance;
    bytes32 cannonPrestate;
    bytes32 cannonKonaPrestate;
}

/// @title MigrateV1V2
/// @notice CGT v1 → v2 migration, in three phases. Never broadcasts.
///         preflight()   reverts if any precondition fails.
///         buildBundle() returns the 5 to 7 Safe tasks for CeloSuperchainOps to sign.
///         verify()      asserts the post-state, after the Safe has executed.
/// @dev Each phase takes an input struct, which is how the tests drive it. Operators use the run*
///      wrappers, which read the same values from env; runBuildBundle batches every task into one
///      atomic Safe transaction and writes it to BUNDLE_PATH.
///        forge script scripts/cgt/MigrateV1V2.s.sol --rpc-url <url> \
///          --sig "runPreflight()" | "runBuildBundle()" | "runVerify()"
contract MigrateV1V2 is Script {
    // ---------- Types ----------

    /// @notice One Safe transaction to sign: target plus the calldata the Safe delegatecalls.
    struct Task {
        address target;
        bytes calldata_;
    }

    // ---------- Constants ----------

    /// @notice CGT feature flag identifier on the v2 SystemConfig.
    bytes32 internal constant CUSTOM_GAS_TOKEN_FEATURE = "CUSTOM_GAS_TOKEN";

    /// @notice Canonical Multicall3 address (same on every chain).
    address internal constant MULTICALL3 = 0xcA11bde05977b3631167028862bE2a173976CA11;

    /// @notice Holds the CELO amount migrate() drained, written to PROXY storage so it outlives the
    ///         migrator impl. Zero until migration. Must match PortalMigrator.MIGRATED_SLOT.
    bytes32 internal constant MIGRATED_SLOT = bytes32(uint256(keccak256("celo.op.portal.migrated")) - 1);

    /// @notice Upstream storage slot of SystemConfig.superchainConfig (packed with minBaseFee at offset 20).
    bytes32 internal constant SUPERCHAIN_CONFIG_SLOT = bytes32(uint256(108));

    /// @notice Backs SystemConfig.celoPaused(); the only pause state the repoint cannot orphan.
    bytes32 internal constant CELO_PAUSED_SLOT = bytes32(uint256(keccak256("celo.op.systemconfig.paused")) - 1);

    /// @notice EIP-1967 implementation slot.
    bytes32 internal constant EIP1967_IMPL_SLOT = bytes32(uint256(keccak256("eip1967.proxy.implementation")) - 1);

    /// @notice superchain-ops' OPCM batching multicall
    address internal constant MULTICALL3_DELEGATECALL = 0x93dc480940585D9961bfcEab58124fFD3d60f76a;

    /// @notice OP-Succinct fault dispute game type, the respected type on Celo mainnet.
    uint32 internal constant OP_SUCCINCT_GAME_TYPE = 42;

    // ---------- Env entry points ----------

    /// @notice Env entry: preflight checks.
    function runPreflight() external view {
        preflight(_inputFromEnv());
    }

    /// @notice Env entry: write the migration as one atomic task to BUNDLE_PATH.
    function runBuildBundle() external {
        require(
            MULTICALL3_DELEGATECALL.code.length > 0, "MigrateV1V2: Multicall3Delegatecall has no code on this network"
        );

        Task[] memory tasks = buildBundle(_inputFromEnv());
        IMulticall3.Call3[] memory calls = new IMulticall3.Call3[](tasks.length);
        for (uint256 i; i < tasks.length; i++) {
            calls[i] = IMulticall3.Call3({ target: tasks[i].target, allowFailure: false, callData: tasks[i].calldata_ });
        }

        Task[] memory atomic = new Task[](1);
        atomic[0] = Task({ target: MULTICALL3_DELEGATECALL, calldata_: _encodeAggregate3(calls) });
        console.log("Atomic bundle via Multicall3Delegatecall:", MULTICALL3_DELEGATECALL);
        _logTask("atomic", atomic[0]);
        _writeBundle(atomic, vm.envOr("BUNDLE_PATH", string("./deployments/cgt-bundle.json")));
    }

    /// @notice Env entry: post-execution verification.
    function runVerify() external view {
        verify(_inputFromEnv());
    }

    // ---------- Phase 1: Preflight ----------

    function preflight(MigrateV1V2Input memory _input) public view {
        console.log("--- MigrateV1V2: Phase 1 preflight ---");
        _validate(_input);

        // Warn-only on balance drift
        uint256 portalBalance = _input.celoTokenL1.balanceOf(_input.portalProxy);
        if (portalBalance != _input.legacyPortalBalance) {
            console.log("WARN portal CELO balance drifted from expected. expected:", _input.legacyPortalBalance);
            console.log("WARN actual portal CELO balance:", portalBalance);
        } else {
            console.log("OK  portal CELO balance matches expected:", portalBalance);
        }

        // Pre-upgrade `isCustomGasToken()` is the legacy (pre-v6) getter: must be true (chain on CGT v1).
        require(
            ISystemConfig(_input.systemConfig).isCustomGasToken(),
            "MigrateV1V2: legacy CGT v1 not active (isCustomGasToken() != true pre-upgrade)"
        );
        console.log("OK  legacy CGT v1 active (isCustomGasToken() == true pre-upgrade)");

        // ProxyAdmin must be owned by the governance Safe.
        require(
            IProxyAdmin(_input.proxyAdmin).owner() == _input.safe,
            "MigrateV1V2: ProxyAdmin owner is not the governance Safe"
        );
        console.log("OK  ProxyAdmin.owner == governance Safe");

        // PortalMigrator must be deployed and configured for this balance + bridge.
        PortalMigrator migrator = PortalMigrator(payable(_input.portalMigratorImpl));
        require(
            address(migrator.CELO_TOKEN_L1()) == address(_input.celoTokenL1),
            "MigrateV1V2: migrator CELO_TOKEN_L1 mismatch"
        );
        require(migrator.CELO_GAS_BRIDGE_L1() == _input.bridgeL1Proxy, "MigrateV1V2: migrator gas bridge mismatch");
        require(
            migrator.LEGACY_PORTAL_BALANCE() == _input.legacyPortalBalance,
            "MigrateV1V2: migrator LEGACY_PORTAL_BALANCE mismatch"
        );
        require(vm.load(_input.portalProxy, MIGRATED_SLOT) == bytes32(0), "MigrateV1V2: portal already migrated");
        console.log("OK  PortalMigrator configuration matches");

        // Current portal impl must match _originalPortalImpl.
        address currentImpl = IProxyAdmin(_input.proxyAdmin).getProxyImplementation(_input.portalProxy);
        require(currentImpl == _input.originalPortalImpl, "MigrateV1V2: current portal impl != originalPortalImpl");
        console.log("OK  current portal impl matches originalPortalImpl");

        // L1 bridge must be initialized, wired, and not yet escrow-seeded.
        ICeloGasBridgeL1 bridge = ICeloGasBridgeL1(payable(_input.bridgeL1Proxy));
        require(
            address(bridge.celoTokenL1()) == address(_input.celoTokenL1), "MigrateV1V2: bridge celoTokenL1 mismatch"
        );
        require(address(bridge.systemConfig()) == _input.systemConfig, "MigrateV1V2: bridge systemConfig not wired");
        require(address(bridge.messenger()) != address(0), "MigrateV1V2: bridge messenger not wired");
        require(address(bridge.otherBridge()) != address(0), "MigrateV1V2: bridge otherBridge not wired");
        require(!bridge.escrowSeeded(), "MigrateV1V2: bridge escrow already seeded");
        console.log("OK  CeloGasBridgeL1 initialized, wired, and not yet seeded");

        // StorageSetter and the external SuperchainConfig must be deployed for the repoint.
        require(_input.storageSetter.code.length > 0, "MigrateV1V2: storageSetter has no code");
        address externalSc = _input.externalSuperchainConfig;
        require(externalSc.code.length > 0, "MigrateV1V2: externalSuperchainConfig has no code");
        // Upstream patches this contract in place, so accept any version at or above the baseline.
        require(
            SemverComp.gte(ISemver(externalSc).version(), "2.4.0"),
            "MigrateV1V2: externalSuperchainConfig version < 2.4.0"
        );
        console.log("OK  StorageSetter + external SuperchainConfig deployed");

        // Queue draining lives in CheckQueues.s.sol; it needs L1 logs and L2 state.
        IOptimismPortal2 portal = IOptimismPortal2(_input.portalProxy);
        if (!portal.paused()) {
            console.log("WARN portal is not paused; withdrawal queue can still change before execution");
        } else {
            console.log("OK  portal paused, withdrawal queue is frozen");
        }

        // The pause task writes celoPaused = true, so the guardian must pause the CSC beforehand.
        if (!ISystemConfig(_input.systemConfig).paused()) {
            console.log("WARN system is NOT paused. Have the guardian pause before executing the bundle.");
        } else {
            console.log("OK  system is paused; the pause task will carry that state onto SystemConfig");
        }

        console.log("Preflight passed");
    }

    // ---------- Phase 2: Build Safe tasks ----------

    /// @notice Builds the Safe tasks in execution order; does NOT broadcast or write to disk.
    /// @notice Every task is a DELEGATECALL (operation = 1) and the order matters. Two tasks are
    ///         conditional, so the bundle holds 5 to 7 and the numbering shifts with them:
    ///           1. drain the portal into CeloGasBridgeL1
    ///           2. repoint SystemConfig.superchainConfig to the external SuperchainConfig
    ///           3. retire the legacy CANNON game type      (only where the chain runs it)
    ///           4. OPCM.upgrade, installing the v6 impls   (requires step 2)
    ///           5. restore the pause by writing celoPaused (requires step 4's impl to read it)
    ///           6. enable the CUSTOM_GAS_TOKEN feature     (requires step 4's SystemConfig impl)
    ///           7. point game type 42 at a new impl        (only where the chain runs it)
    function buildBundle(MigrateV1V2Input memory _input) public view returns (Task[] memory tasks_) {
        console.log("--- MigrateV1V2: Phase 2 build tasks ---");
        _validate(_input);

        // Both game-type tasks are driven by what the chain actually runs, never by operator
        // choice: Havoc has neither CANNON nor game type 42, mainnet has both.
        bool clearLegacyCannon = _gameImpl(_input, GameTypes.CANNON) != address(0);
        bool swapSuccinctGame = _gameImpl(_input, GameType.wrap(OP_SUCCINCT_GAME_TYPE)) != address(0);
        if (swapSuccinctGame) {
            require(
                _input.succinctGameImpl != address(0), "MigrateV1V2: succinctGameImpl required, chain runs game type 42"
            );
        } else {
            require(
                _input.succinctGameImpl == address(0),
                "MigrateV1V2: succinctGameImpl set, but chain has no game type 42"
            );
        }
        tasks_ = new Task[](5 + (clearLegacyCannon ? 1 : 0) + (swapSuccinctGame ? 1 : 0));
        uint256 n;

        tasks_[n++] = Task({ target: MULTICALL3, calldata_: _buildDrainTask(_input) });
        _logStep(n, "aggregate3[install migrator, migrate(), restore original portal]");

        tasks_[n++] = Task({ target: MULTICALL3, calldata_: _buildRepointTask(_input) });
        _logStep(n, "aggregate3[repoint superchainConfig CSC -> external SC via StorageSetter]");

        // Must precede OPCM: a zero CANNON impl makes OPCM skip its permissionless branch.
        if (clearLegacyCannon) {
            tasks_[n++] = Task({ target: MULTICALL3, calldata_: _buildClearLegacyCannonTask(_input) });
            _logStep(n, "aggregate3[DGF.setImplementation(CANNON, address(0))]");
        }

        tasks_[n++] = Task({ target: _input.opcm, calldata_: _buildOpcmTask(_input) });
        _logStep(n, "OPCM.upgrade(opChainConfigs) [direct delegatecall]");

        // Immediately after OPCM: that impl is the first that reads celoPaused, so this is the
        // earliest the pre-upgrade pause can be carried across. Everything below runs paused.
        tasks_[n++] = Task({ target: MULTICALL3, calldata_: _buildPauseTask(_input) });
        _logStep(n, "aggregate3[pause via celoPaused slot, restore v6 SystemConfig impl]");

        tasks_[n++] = Task({ target: MULTICALL3, calldata_: _buildCgtFlagTask(_input) });
        _logStep(n, "aggregate3[SystemConfig.setFeature(CUSTOM_GAS_TOKEN, true)]");

        if (swapSuccinctGame) {
            tasks_[n++] = Task({ target: MULTICALL3, calldata_: _buildSuccinctGameSwapTask(_input) });
            _logStep(n, "aggregate3[DGF.setImplementation(42, redeployed impl)]");
        }

        require(n == tasks_.length, "MigrateV1V2: task count mismatch");
        for (uint256 i; i < tasks_.length; i++) {
            _logTask(string.concat("tx", vm.toString(i + 1)), tasks_[i]);
        }

        console.log("Tasks built. Pass each target + calldata to CeloSuperchainOps for signing.");
    }

    // ---------- Phase 3: Post-execution verification ----------

    function verify(MigrateV1V2Input memory _input) public view {
        console.log("--- MigrateV1V2: Phase 3 verification ---");
        _validate(_input);

        // Portal CELO balance should be zero after migration.
        uint256 portalBalance = _input.celoTokenL1.balanceOf(_input.portalProxy);
        require(portalBalance == 0, "MigrateV1V2: portal still holds CELO");
        console.log("OK  portal CELO balance == 0");

        // The slot is the expected amount; legacyPortalBalance's env default reads the now-drained portal.
        uint256 migratedAmount = uint256(vm.load(_input.portalProxy, MIGRATED_SLOT));
        require(migratedAmount != 0, "MigrateV1V2: portal not migrated");

        // The bridge may hold more than it was seeded with once deposits resume, never less.
        uint256 bridgeBalance = _input.celoTokenL1.balanceOf(_input.bridgeL1Proxy);
        require(bridgeBalance >= migratedAmount, "MigrateV1V2: bridge holds less CELO than was migrated");
        console.log("OK  CeloGasBridgeL1 covers the migrated amount:", migratedAmount);
        if (bridgeBalance != migratedAmount) {
            console.log("     bridge balance now (post-migration deposits):", bridgeBalance);
        }

        // CGT feature flag should be flipped on.
        require(
            ISystemConfig(_input.systemConfig).isCustomGasToken(),
            "MigrateV1V2: SystemConfig.isCustomGasToken() != true after migration"
        );
        console.log("OK  SystemConfig.isCustomGasToken() == true");

        // The bundle must leave the system paused. Run this before unpausing, or it will fail.
        require(
            ISystemConfig(_input.systemConfig).paused(),
            "MigrateV1V2: SystemConfig.paused() != true after migration (already unpaused?)"
        );
        console.log("OK  SystemConfig.paused() == true (system sealed until unpause)");

        // Portal impl must be the v6 impl OPCM installed: neither the migrator nor the restored original.
        address currentImpl = IProxyAdmin(_input.proxyAdmin).getProxyImplementation(_input.portalProxy);
        require(currentImpl != _input.portalMigratorImpl, "MigrateV1V2: portal still backed by migrator");
        require(currentImpl != _input.originalPortalImpl, "MigrateV1V2: portal impl not upgraded by OPCM");
        console.log("OK  portal implementation upgraded by OPCM (v6)");

        // Escrow must be seeded and covered by the bridge's actual CELO balance.
        ICeloGasBridgeL1 bridge = ICeloGasBridgeL1(payable(_input.bridgeL1Proxy));
        require(bridge.escrowSeeded(), "MigrateV1V2: bridge escrow not seeded");
        uint256 trackedEscrow = bridge.deposits(address(_input.celoTokenL1), address(0));
        require(trackedEscrow <= bridgeBalance, "MigrateV1V2: bridge escrow > token balance (insolvent)");
        console.log("OK  tracked escrow covered by CELO balance:", trackedEscrow);

        require(migratedAmount == trackedEscrow, "MigrateV1V2: migrated amount != tracked escrow");
        console.log("OK  migrated amount recorded in proxy storage:", migratedAmount);

        // SystemConfig.superchainConfig must now resolve to the external SuperchainConfig.
        require(
            address(ISystemConfig(_input.systemConfig).superchainConfig()) == _input.externalSuperchainConfig,
            "MigrateV1V2: superchainConfig not repointed to external SuperchainConfig"
        );
        console.log("OK  SystemConfig.superchainConfig == external SuperchainConfig");

        console.log("Migration verified.");
    }

    // ---------- Internal helpers ----------

    /// @notice Builds MigrateV1V2Input from env. Prestates are required, never defaulted.
    function _inputFromEnv() internal view returns (MigrateV1V2Input memory input_) {
        address payable portalProxy = payable(vm.envAddress("PORTAL_PROXY"));
        IERC20 celoTokenL1 = IERC20(vm.envAddress("CELO_TOKEN_L1"));

        input_ = MigrateV1V2Input({
            portalProxy: portalProxy,
            systemConfig: vm.envAddress("SYSTEM_CONFIG"),
            proxyAdmin: vm.envAddress("PROXY_ADMIN"),
            opcm: vm.envAddress("OPCM"),
            safe: vm.envAddress("SAFE"),
            portalMigratorImpl: vm.envAddress("PORTAL_MIGRATOR_IMPL"),
            originalPortalImpl: vm.envAddress("ORIGINAL_PORTAL_IMPL"),
            bridgeL1Proxy: vm.envAddress("BRIDGE_L1_PROXY"),
            storageSetter: vm.envAddress("STORAGE_SETTER"),
            externalSuperchainConfig: vm.envAddress("EXTERNAL_SUPERCHAIN_CONFIG"),
            succinctGameImpl: vm.envOr("SUCCINCT_GAME_IMPL", address(0)),
            celoTokenL1: celoTokenL1,
            // Live read is only correct pre-drain; set explicitly when verifying after the drain.
            legacyPortalBalance: vm.envOr("LEGACY_PORTAL_BALANCE", celoTokenL1.balanceOf(portalProxy)),
            cannonPrestate: vm.envBytes32("CANNON_PRESTATE"),
            cannonKonaPrestate: vm.envBytes32("CANNON_KONA_PRESTATE")
        });
    }

    /// @notice Reverts unless every address the phases dereference is set.
    function _validate(MigrateV1V2Input memory _input) internal pure {
        require(_input.portalProxy != address(0), "MigrateV1V2Input: portalProxy not set");
        require(_input.systemConfig != address(0), "MigrateV1V2Input: systemConfig not set");
        require(_input.proxyAdmin != address(0), "MigrateV1V2Input: proxyAdmin not set");
        require(_input.opcm != address(0), "MigrateV1V2Input: opcm not set");
        require(_input.safe != address(0), "MigrateV1V2Input: safe not set");
        require(_input.portalMigratorImpl != address(0), "MigrateV1V2Input: portalMigratorImpl not set");
        require(_input.originalPortalImpl != address(0), "MigrateV1V2Input: originalPortalImpl not set");
        require(_input.bridgeL1Proxy != address(0), "MigrateV1V2Input: bridgeL1Proxy not set");
        require(_input.storageSetter != address(0), "MigrateV1V2Input: storageSetter not set");
        require(_input.externalSuperchainConfig != address(0), "MigrateV1V2Input: externalSuperchainConfig not set");
        require(address(_input.celoTokenL1) != address(0), "MigrateV1V2Input: celoTokenL1 not set");
    }

    /// @notice Writes the bundle as [{target, calldata, operation}] for exec-upgrade.sh to consume.
    /// @notice operation is always 1: the Safe delegatecalls every target, Multicall3 and OPCM alike.
    function _writeBundle(Task[] memory _tasks, string memory _path) internal {
        string memory json = "[";
        for (uint256 i = 0; i < _tasks.length; i++) {
            json = string.concat(
                json,
                i == 0 ? "" : ",",
                '{"target":"',
                vm.toString(_tasks[i].target),
                '","calldata":"',
                vm.toString(_tasks[i].calldata_),
                '","operation":1}'
            );
        }
        vm.writeFile(_path, string.concat(json, "]"));
        console.log("Bundle written to:", _path);
    }

    /// @notice Live implementation for a game type, or address(0) when the chain does not run it.
    function _gameImpl(MigrateV1V2Input memory _input, GameType _gameType) internal view returns (address) {
        IDisputeGameFactory dgf = IDisputeGameFactory(ISystemConfig(_input.systemConfig).disputeGameFactory());
        return address(dgf.gameImpls(_gameType));
    }

    /// @notice Logs a task's one-based position and what it does.
    function _logStep(uint256 _position, string memory _what) internal pure {
        console.log(string.concat("Tx", vm.toString(_position), ": ", _what));
    }

    /// @notice Installs the migrator, drains the portal, restores the original portal impl.
    function _buildDrainTask(MigrateV1V2Input memory _input) internal pure returns (bytes memory) {
        IMulticall3.Call3[] memory calls = new IMulticall3.Call3[](3);
        calls[0] = IMulticall3.Call3({
            target: _input.proxyAdmin,
            allowFailure: false,
            callData: abi.encodeWithSignature("upgrade(address,address)", _input.portalProxy, _input.portalMigratorImpl)
        });
        calls[1] = IMulticall3.Call3({
            target: _input.portalProxy,
            allowFailure: false,
            callData: abi.encodeWithSignature("migrate()")
        });
        calls[2] = IMulticall3.Call3({
            target: _input.proxyAdmin,
            allowFailure: false,
            callData: abi.encodeWithSignature("upgrade(address,address)", _input.portalProxy, _input.originalPortalImpl)
        });
        return _encodeAggregate3(calls);
    }

    /// @notice Repoints SystemConfig.superchainConfig to the external SuperchainConfig.
    /// @notice Slot 108 packs minBaseFee, so the live word is overlaid rather than replaced.
    function _buildRepointTask(MigrateV1V2Input memory _input) internal view returns (bytes memory) {
        bytes32 slotData = _overwriteSuperchainConfig(
            vm.load(_input.systemConfig, SUPERCHAIN_CONFIG_SLOT), _input.externalSuperchainConfig
        );
        IMulticall3.Call3[] memory calls = new IMulticall3.Call3[](3);
        calls[0] = IMulticall3.Call3({
            target: _input.proxyAdmin,
            allowFailure: false,
            callData: abi.encodeWithSignature("upgrade(address,address)", _input.systemConfig, _input.storageSetter)
        });
        calls[1] = IMulticall3.Call3({
            target: _input.systemConfig,
            allowFailure: false,
            callData: abi.encodeWithSignature("setBytes32(bytes32,bytes32)", SUPERCHAIN_CONFIG_SLOT, slotData)
        });
        calls[2] = IMulticall3.Call3({
            target: _input.proxyAdmin,
            allowFailure: false,
            callData: abi.encodeWithSignature(
                "upgrade(address,address)",
                _input.systemConfig,
                address(uint160(uint256(vm.load(_input.systemConfig, EIP1967_IMPL_SLOT))))
            )
        });
        return _encodeAggregate3(calls);
    }

    /// @notice Retires the unused legacy CANNON game type by zeroing its implementation.
    /// @notice Only CANNON is safe to clear; OPCM reads PERMISSIONED_CANNON unconditionally.
    function _buildClearLegacyCannonTask(MigrateV1V2Input memory _input) internal view returns (bytes memory) {
        IMulticall3.Call3[] memory calls = new IMulticall3.Call3[](1);
        calls[0] = IMulticall3.Call3({
            target: ISystemConfig(_input.systemConfig).disputeGameFactory(),
            allowFailure: false,
            callData: abi.encodeWithSignature("setImplementation(uint32,address)", GameTypes.CANNON.raw(), address(0))
        });
        return _encodeAggregate3(calls);
    }

    /// @notice Repoints game type 42 at the redeployed OPSuccinctFaultDisputeGame.
    /// @notice Required when the AccessManager is redeployed; existing games keep their old impl.
    function _buildSuccinctGameSwapTask(MigrateV1V2Input memory _input) internal view returns (bytes memory) {
        IMulticall3.Call3[] memory calls = new IMulticall3.Call3[](1);
        calls[0] = IMulticall3.Call3({
            target: ISystemConfig(_input.systemConfig).disputeGameFactory(),
            allowFailure: false,
            callData: abi.encodeWithSignature(
                "setImplementation(uint32,address)", OP_SUCCINCT_GAME_TYPE, _input.succinctGameImpl
            )
        });
        return _encodeAggregate3(calls);
    }

    /// @notice OPCM.upgrade, delegatecalled directly by the Safe.
    function _buildOpcmTask(MigrateV1V2Input memory _input) internal pure returns (bytes memory) {
        IOPContractsManager.OpChainConfig[] memory configs = new IOPContractsManager.OpChainConfig[](1);
        configs[0] = IOPContractsManager.OpChainConfig({
            systemConfigProxy: ISystemConfig(_input.systemConfig),
            cannonPrestate: Claim.wrap(_input.cannonPrestate),
            cannonKonaPrestate: Claim.wrap(_input.cannonKonaPrestate)
        });
        return abi.encodeCall(IOPContractsManager.upgrade, (configs));
    }

    /// @notice Flips the CUSTOM_GAS_TOKEN feature flag on the v6 SystemConfig.
    /// @notice Runs paused, which is fine: setFeature only rejects a paused system for ETH_LOCKBOX.
    function _buildCgtFlagTask(MigrateV1V2Input memory _input) internal pure returns (bytes memory) {
        IMulticall3.Call3[] memory calls = new IMulticall3.Call3[](1);
        calls[0] = IMulticall3.Call3({
            target: _input.systemConfig,
            allowFailure: false,
            callData: abi.encodeWithSignature("setFeature(bytes32,bool)", CUSTOM_GAS_TOKEN_FEATURE, true)
        });
        return _encodeAggregate3(calls);
    }

    /// @notice Writes the celoPaused slot via StorageSetter, then restores the v6 impl.
    /// @notice pause() is onlyOwner on a different Safe, so the flag is written as storage.
    function _buildPauseTask(MigrateV1V2Input memory _input) internal view returns (bytes memory) {
        IMulticall3.Call3[] memory calls = new IMulticall3.Call3[](3);
        calls[0] = IMulticall3.Call3({
            target: _input.proxyAdmin,
            allowFailure: false,
            callData: abi.encodeWithSignature("upgrade(address,address)", _input.systemConfig, _input.storageSetter)
        });
        calls[1] = IMulticall3.Call3({
            target: _input.systemConfig,
            allowFailure: false,
            callData: abi.encodeWithSignature("setBool(bytes32,bool)", CELO_PAUSED_SLOT, true)
        });
        calls[2] = IMulticall3.Call3({
            target: _input.proxyAdmin,
            allowFailure: false,
            callData: abi.encodeWithSignature(
                "upgrade(address,address)",
                _input.systemConfig,
                IOPContractsManager(_input.opcm).implementations().systemConfigImpl
            )
        });
        return _encodeAggregate3(calls);
    }

    /// @notice ABI-encodes Multicall3.aggregate3((address,bool,bytes)[]) for the given calls.
    function _encodeAggregate3(IMulticall3.Call3[] memory _calls) internal pure returns (bytes memory) {
        return abi.encodeCall(IMulticall3.aggregate3, (_calls));
    }

    /// @notice Overwrites slot 108's low 160 bits (superchainConfig), preserving minBaseFee at offset 20.
    function _overwriteSuperchainConfig(bytes32 _slot, address _superchainConfig) internal pure returns (bytes32) {
        uint256 mask160 = type(uint160).max;
        return bytes32((uint256(_slot) & ~mask160) | uint256(uint160(_superchainConfig)));
    }

    /// @notice Logs one task's target and calldata to stdout.
    function _logTask(string memory _name, Task memory _task) internal pure {
        console.log(_name, "target:", _task.target);
        console.log(_name, "calldata:");
        console.logBytes(_task.calldata_);
    }
}
