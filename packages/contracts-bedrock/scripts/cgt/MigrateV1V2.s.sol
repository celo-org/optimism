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
import { BaseDeployIO } from "scripts/deploy/BaseDeployIO.sol";

// Interfaces
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ICeloGasBridgeL1 } from "interfaces/celo/ICeloGasBridgeL1.sol";
import { IOPContractsManager } from "interfaces/L1/IOPContractsManager.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";

/// @title MigrateV1V2Input
/// @notice Inputs for the `MigrateV1V2` migration script.
contract MigrateV1V2Input is BaseDeployIO {
    // Addresses
    address payable internal _portalProxy;
    address internal _systemConfig;
    address internal _proxyAdmin;
    address internal _opcm;
    address internal _safe;

    // Implementation targets
    address internal _portalMigratorImpl;
    address internal _originalPortalImpl;
    address internal _bridgeL1Proxy;

    // CELO ERC-20 and expected pre-migration balance
    IERC20 internal _celoToken;
    uint256 internal _legacyPortalBalance;

    // OPCM upgrade prestates
    bytes32 internal _cannonPrestate;
    bytes32 internal _cannonKonaPrestate;

    function set(bytes4 _sel, address _value) public {
        require(_value != address(0), "MigrateV1V2Input: cannot set zero address");

        if (_sel == this.portalProxy.selector) _portalProxy = payable(_value);
        else if (_sel == this.systemConfig.selector) _systemConfig = _value;
        else if (_sel == this.proxyAdmin.selector) _proxyAdmin = _value;
        else if (_sel == this.opcm.selector) _opcm = _value;
        else if (_sel == this.safe.selector) _safe = _value;
        else if (_sel == this.portalMigratorImpl.selector) _portalMigratorImpl = _value;
        else if (_sel == this.originalPortalImpl.selector) _originalPortalImpl = _value;
        else if (_sel == this.bridgeL1Proxy.selector) _bridgeL1Proxy = _value;
        else if (_sel == this.celoToken.selector) _celoToken = IERC20(_value);
        else revert("MigrateV1V2Input: unknown selector");
    }

    function set(bytes4 _sel, uint256 _value) public {
        if (_sel == this.legacyPortalBalance.selector) _legacyPortalBalance = _value;
        else revert("MigrateV1V2Input: unknown selector");
    }

    function set(bytes4 _sel, bytes32 _value) public {
        if (_sel == this.cannonPrestate.selector) _cannonPrestate = _value;
        else if (_sel == this.cannonKonaPrestate.selector) _cannonKonaPrestate = _value;
        else revert("MigrateV1V2Input: unknown selector");
    }

    function portalProxy() public view returns (address payable) {
        require(_portalProxy != address(0), "MigrateV1V2Input: portalProxy not set");
        return _portalProxy;
    }

    function systemConfig() public view returns (address) {
        require(_systemConfig != address(0), "MigrateV1V2Input: systemConfig not set");
        return _systemConfig;
    }

    function proxyAdmin() public view returns (address) {
        require(_proxyAdmin != address(0), "MigrateV1V2Input: proxyAdmin not set");
        return _proxyAdmin;
    }

    function opcm() public view returns (address) {
        require(_opcm != address(0), "MigrateV1V2Input: opcm not set");
        return _opcm;
    }

    function safe() public view returns (address) {
        require(_safe != address(0), "MigrateV1V2Input: safe not set");
        return _safe;
    }

    function portalMigratorImpl() public view returns (address) {
        require(_portalMigratorImpl != address(0), "MigrateV1V2Input: portalMigratorImpl not set");
        return _portalMigratorImpl;
    }

    function originalPortalImpl() public view returns (address) {
        require(_originalPortalImpl != address(0), "MigrateV1V2Input: originalPortalImpl not set");
        return _originalPortalImpl;
    }

    function bridgeL1Proxy() public view returns (address) {
        require(_bridgeL1Proxy != address(0), "MigrateV1V2Input: bridgeL1Proxy not set");
        return _bridgeL1Proxy;
    }

    function celoToken() public view returns (IERC20) {
        require(address(_celoToken) != address(0), "MigrateV1V2Input: celoToken not set");
        return _celoToken;
    }

    function legacyPortalBalance() public view returns (uint256) {
        return _legacyPortalBalance;
    }

    function cannonPrestate() public view returns (bytes32) {
        return _cannonPrestate;
    }

    function cannonKonaPrestate() public view returns (bytes32) {
        return _cannonKonaPrestate;
    }
}

/// @title MigrateV1V2
/// @notice CGT v1 → v2 migration script. Three phases in one file.
///         Phase 1 — preflight checks (revert if any precondition fails).
///         Phase 2 — build the three Safe tasks (target + calldata) and log/return them for
///                   CeloSuperchainOps to sign. Does NOT broadcast or write files.
///         Phase 3 — post-execution verification (asserts post-state, called after Safe execution).
///
/// @dev Standard usage:
///        forge script scripts/cgt/MigrateV1V2.s.sol --rpc-url <url> --sig "preflight(address)" <inputAddr>
///        forge script scripts/cgt/MigrateV1V2.s.sol --rpc-url <url> --sig "buildBundle(address)" <inputAddr>
///        forge script scripts/cgt/MigrateV1V2.s.sol --rpc-url <url> --sig "verify(address)" <inputAddr>
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

    // ---------- Phase 1: Preflight ----------

    function preflight(MigrateV1V2Input _input) public view {
        console.log("--- MigrateV1V2: Phase 1 preflight ---");

        // Warn-only on balance drift
        uint256 portalBalance = _input.celoToken().balanceOf(_input.portalProxy());
        if (portalBalance != _input.legacyPortalBalance()) {
            console.log("WARN portal CELO balance drifted from expected. expected:", _input.legacyPortalBalance());
            console.log("WARN actual portal CELO balance:", portalBalance);
        } else {
            console.log("OK  portal CELO balance matches expected:", portalBalance);
        }

        // Pre-upgrade `isCustomGasToken()` is the legacy (pre-v6) getter: must be true (chain on CGT v1).
        require(
            ISystemConfig(_input.systemConfig()).isCustomGasToken(),
            "MigrateV1V2: legacy CGT v1 not active (isCustomGasToken() != true pre-upgrade)"
        );
        console.log("OK  legacy CGT v1 active (isCustomGasToken() == true pre-upgrade)");

        // Asserts ProxyAdmin is owned by the governance Safe (so the Safe can drive upgrades).
        require(
            IProxyAdmin(_input.proxyAdmin()).owner() == _input.safe(),
            "MigrateV1V2: ProxyAdmin owner is not the governance Safe"
        );
        console.log("OK  ProxyAdmin.owner == governance Safe");

        // Asserts PortalMigrator is deployed and configured for this exact balance + bridge.
        PortalMigrator migrator = PortalMigrator(payable(_input.portalMigratorImpl()));
        require(address(migrator.CELO_TOKEN()) == address(_input.celoToken()), "MigrateV1V2: migrator CELO_TOKEN mismatch");
        require(migrator.CELO_GAS_BRIDGE_L1() == _input.bridgeL1Proxy(), "MigrateV1V2: migrator gas bridge mismatch");
        require(migrator.LEGACY_PORTAL_BALANCE() == _input.legacyPortalBalance(), "MigrateV1V2: migrator LEGACY_PORTAL_BALANCE mismatch");
        require(!migrator.migrated(), "MigrateV1V2: migrator already used");
        console.log("OK  PortalMigrator configuration matches");

        // Asserts the current portal impl matches _originalPortalImpl.
        address currentImpl = IProxyAdmin(_input.proxyAdmin()).getProxyImplementation(_input.portalProxy());
        require(currentImpl == _input.originalPortalImpl(), "MigrateV1V2: current portal impl != originalPortalImpl");
        console.log("OK  current portal impl matches originalPortalImpl");

        // Asserts the L1 bridge is deployed, initialized and wired, and not yet escrow-seeded.
        ICeloGasBridgeL1 bridge = ICeloGasBridgeL1(payable(_input.bridgeL1Proxy()));
        require(address(bridge.CELO_TOKEN()) == address(_input.celoToken()), "MigrateV1V2: bridge CELO_TOKEN mismatch");
        require(address(bridge.systemConfig()) == _input.systemConfig(), "MigrateV1V2: bridge systemConfig not wired");
        require(address(bridge.messenger()) != address(0), "MigrateV1V2: bridge messenger not wired");
        require(address(bridge.otherBridge()) != address(0), "MigrateV1V2: bridge otherBridge not wired");
        require(!bridge.escrowSeeded(), "MigrateV1V2: bridge escrow already seeded");
        console.log("OK  CeloGasBridgeL1 initialized, wired, and not yet seeded");

        console.log("Preflight passed");
    }

    // ---------- Phase 2: Build Safe tasks ----------

    /// @notice Builds the three Safe tasks, logs each target + calldata, and returns them. Does NOT
    ///         broadcast or write to disk. Three sequential DELEGATECALLs (operation = 1); order matters:
    ///           Tx1  Multicall3.aggregate3([ ProxyAdmin->PortalMigrator, portal.migrate(), ProxyAdmin->originalPortalImpl ])
    ///           Tx2  OPCM.upgrade(OpChainConfig[])                — installs the v6 portal + SystemConfig impls
    ///           Tx3  Multicall3.aggregate3([ SystemConfig.setFeature(CGT, true) ])  — needs the v6 SystemConfig impl
    function buildBundle(MigrateV1V2Input _input) public view returns (Task[3] memory tasks_) {
        console.log("--- MigrateV1V2: Phase 2 build tasks ---");

        // --- Tx1: install migrator, drain, restore original portal impl (Multicall3 batch) ---
        IMulticall3.Call3[] memory tx1Calls = new IMulticall3.Call3[](3);
        tx1Calls[0] = IMulticall3.Call3({
            target: _input.proxyAdmin(),
            allowFailure: false,
            callData: abi.encodeWithSignature(
                "upgrade(address,address)", _input.portalProxy(), _input.portalMigratorImpl()
            )
        });
        tx1Calls[1] = IMulticall3.Call3({
            target: _input.portalProxy(),
            allowFailure: false,
            callData: abi.encodeWithSignature("migrate()")
        });
        tx1Calls[2] = IMulticall3.Call3({
            target: _input.proxyAdmin(),
            allowFailure: false,
            callData: abi.encodeWithSignature(
                "upgrade(address,address)", _input.portalProxy(), _input.originalPortalImpl()
            )
        });
        bytes memory tx1Data = _encodeAggregate3(tx1Calls);
        console.log("Tx1: Multicall3.aggregate3[install migrator, migrate(), restore original portal]");

        // --- Tx2: OPCM.upgrade (direct Safe -> OPCM delegatecall) ---
        IOPContractsManager.OpChainConfig[] memory configs = new IOPContractsManager.OpChainConfig[](1);
        configs[0] = IOPContractsManager.OpChainConfig({
            systemConfigProxy: ISystemConfig(_input.systemConfig()),
            cannonPrestate: Claim.wrap(_input.cannonPrestate()),
            cannonKonaPrestate: Claim.wrap(_input.cannonKonaPrestate())
        });
        bytes memory tx2Data = abi.encodeCall(IOPContractsManager.upgrade, (configs));
        console.log("Tx2: OPCM.upgrade(opChainConfigs) [direct delegatecall]");

        // --- Tx3: flip CGT feature flag (Multicall3 batch of one) ---
        IMulticall3.Call3[] memory tx3Calls = new IMulticall3.Call3[](1);
        tx3Calls[0] = IMulticall3.Call3({
            target: _input.systemConfig(),
            allowFailure: false,
            callData: abi.encodeWithSignature("setFeature(bytes32,bool)", CUSTOM_GAS_TOKEN_FEATURE, true)
        });
        bytes memory tx3Data = _encodeAggregate3(tx3Calls);
        console.log("Tx3: Multicall3.aggregate3[SystemConfig.setFeature(CUSTOM_GAS_TOKEN, true)]");

        // --- Assemble tasks, log target + calldata for each ---
        tasks_[0] = Task({ target: MULTICALL3, calldata_: tx1Data });
        tasks_[1] = Task({ target: _input.opcm(), calldata_: tx2Data });
        tasks_[2] = Task({ target: MULTICALL3, calldata_: tx3Data });

        _logTask("tx1", tasks_[0]);
        _logTask("tx2", tasks_[1]);
        _logTask("tx3", tasks_[2]);

        console.log("Tasks built. Pass each target + calldata to CeloSuperchainOps for signing.");
    }

    // ---------- Phase 3: Post-execution verification ----------

    function verify(MigrateV1V2Input _input) public view {
        console.log("--- MigrateV1V2: Phase 3 verification ---");

        // Portal CELO balance should be zero after migration.
        uint256 portalBalance = _input.celoToken().balanceOf(_input.portalProxy());
        require(portalBalance == 0, "MigrateV1V2: portal still holds CELO");
        console.log("OK  portal CELO balance == 0");

        // CeloGasBridgeL1 holds the migrated CELO; warn if the amount drifted from expected.
        uint256 bridgeBalance = _input.celoToken().balanceOf(_input.bridgeL1Proxy());
        if (bridgeBalance != _input.legacyPortalBalance()) {
            console.log("WARN CeloGasBridgeL1 balance differs from expected. expected:", _input.legacyPortalBalance());
            console.log("WARN actual CeloGasBridgeL1 balance:", bridgeBalance);
        } else {
            console.log("OK  CeloGasBridgeL1 balance matches expected migrated amount:", bridgeBalance);
        }

        // CGT feature flag should be flipped on.
        require(
            ISystemConfig(_input.systemConfig()).isCustomGasToken(),
            "MigrateV1V2: SystemConfig.isCustomGasToken() != true after migration"
        );
        console.log("OK  SystemConfig.isCustomGasToken() == true");

        // Portal impl must be the v6 impl OPCM installed: neither the migrator nor the restored original.
        address currentImpl = IProxyAdmin(_input.proxyAdmin()).getProxyImplementation(_input.portalProxy());
        require(currentImpl != _input.portalMigratorImpl(), "MigrateV1V2: portal still backed by migrator");
        require(currentImpl != _input.originalPortalImpl(), "MigrateV1V2: portal impl not upgraded by OPCM");
        console.log("OK  portal implementation upgraded by OPCM (v6)");

        // Migrator should record the migration as completed.
        require(
            PortalMigrator(payable(_input.portalMigratorImpl())).migrated(),
            "MigrateV1V2: PortalMigrator.migrated() != true"
        );
        console.log("OK  PortalMigrator.migrated() == true");

        // Escrow must be seeded and equal to the bridge's actual CELO balance.
        ICeloGasBridgeL1 bridge = ICeloGasBridgeL1(payable(_input.bridgeL1Proxy()));
        require(bridge.escrowSeeded(), "MigrateV1V2: bridge escrow not seeded");
        uint256 trackedEscrow = bridge.deposits(address(_input.celoToken()), address(0));
        require(trackedEscrow == bridgeBalance, "MigrateV1V2: bridge escrow != token balance (insolvent)");
        console.log("OK  tracked escrow == CELO balance:", trackedEscrow);

        console.log("Migration verified.");
    }

    // ---------- Internal helpers ----------

    /// @notice ABI-encodes Multicall3.aggregate3((address,bool,bytes)[]) for the given calls.
    function _encodeAggregate3(IMulticall3.Call3[] memory _calls) internal pure returns (bytes memory) {
        return abi.encodeCall(IMulticall3.aggregate3, (_calls));
    }

    /// @notice Logs one task's target and calldata to stdout.
    function _logTask(string memory _name, Task memory _task) internal pure {
        console.log(_name, "target:", _task.target);
        console.log(_name, "calldata:");
        console.logBytes(_task.calldata_);
    }
}
