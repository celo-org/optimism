// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing
import { Test } from "forge-std/Test.sol";
import { console2 as console } from "forge-std/console2.sol";
import { Vm } from "forge-std/Vm.sol";

// Scripts
import { DeployImplementations } from "scripts/deploy/DeployImplementations.s.sol";
import { StandardConstants } from "scripts/deploy/StandardConstants.sol";
import { MigrateV1V2, MigrateV1V2Input } from "scripts/cgt/MigrateV1V2.s.sol";

// Contracts
import { CeloGasBridgeL1 } from "src/celo/CeloGasBridgeL1.sol";
import { CeloPredeploys } from "src/celo/CeloPredeploys.sol";
import { PortalMigrator } from "src/celo/PortalMigrator.sol";
import { Proxy } from "src/universal/Proxy.sol";
import { StorageSetter } from "src/universal/StorageSetter.sol";

// Libraries
import { Constants } from "src/libraries/Constants.sol";
import { GameTypes } from "src/dispute/lib/Types.sol";
import { Claim } from "src/dispute/lib/LibUDT.sol";

// Interfaces
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ICeloGasBridgeL1 } from "interfaces/celo/ICeloGasBridgeL1.sol";
import { ICrossDomainMessenger } from "interfaces/universal/ICrossDomainMessenger.sol";
import { IDisputeGameFactory } from "interfaces/dispute/IDisputeGameFactory.sol";
import { IFaultDisputeGame } from "interfaces/dispute/IFaultDisputeGame.sol";
import { IL1CrossDomainMessenger } from "interfaces/L1/IL1CrossDomainMessenger.sol";
import { IOPContractsManager } from "interfaces/L1/IOPContractsManager.sol";
import { IOptimismPortal2 as IOptimismPortal } from "interfaces/L1/IOptimismPortal2.sol";
import { IProtocolVersions } from "interfaces/L1/IProtocolVersions.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { IStandardBridge } from "interfaces/universal/IStandardBridge.sol";
import { ISuperchainConfig } from "interfaces/L1/ISuperchainConfig.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IL1BlockCGT } from "interfaces/L2/IL1BlockCGT.sol";

// Safe
import { Safe as GnosisSafe } from "safe-contracts/Safe.sol";

// Helpers
import { CeloForkSafeExec } from "test/celo/CeloForkSafeExec.sol";
import { EIP1967Helper } from "test/mocks/EIP1967Helper.sol";

contract CGTMigrationFork_Test is Test, CeloForkSafeExec {
    address internal constant SYSTEM_CONFIG_PROXY = 0x89E31965D844a309231B1f17759Ccaf1b7c09861;
    address internal constant PROXY_ADMIN = 0x783A434532Ee94667979213af1711505E8bFE374;
    address internal constant PARENT_SAFE = 0x4092A77bAF58fef0309452cEaCb09221e556E112;
    address internal constant COUNCIL_SAFE = 0xC03172263409584f7860C25B6eB4985f0f6F4636;
    address internal constant CLABS_SAFE = 0x9Eb44Da23433b5cAA1c87e35594D15FcEb08D34d;
    address internal constant MULTICALL3 = 0xcA11bde05977b3631167028862bE2a173976CA11;
    address internal constant PORTAL_PROXY = 0xc5c5D157928BDBD2ACf6d0777626b6C75a9EAEDC;
    address internal constant ORIGINAL_PORTAL_IMPL = 0x2c431080Fc733E259654f3b91E39468d9A85Ac9b;
    address internal constant CELO_TOKEN_PROXY = 0x057898f3C43F129a17517B9056D23851F124b19f;
    address internal constant PROTOCOL_VERSIONS_PROXY = 0x1b6dEB2197418075AB314ac4D52Ca1D104a8F663;
    address internal constant EXTERNAL_SUPERCHAIN_CONFIG = 0x95703e0982140D16f8ebA6d158FccEde42f04a4C;
    address internal constant SYSTEM_CONFIG_OWNER = 0x9Eb44Da23433b5cAA1c87e35594D15FcEb08D34d;

    uint256 internal constant DEFAULT_FORK_BLOCK = 25_680_000;
    uint256 internal constant PER_TX_GAS_CAP = 16_000_000; // EIP-7825 (Fusaka) per-tx gas cap.
    bytes32 internal constant EIP1967_IMPL_SLOT = bytes32(uint256(keccak256("eip1967.proxy.implementation")) - 1);
    uint32 internal constant DEPOSIT_GAS_LIMIT = 250_000;
    uint256 internal constant DEPOSIT_AMOUNT = 1e18;
    bytes32 internal constant MIGRATED_SLOT = bytes32(uint256(keccak256("celo.op.portal.migrated")) - 1);
    bytes32 internal constant CANNON64_FALLBACK_PRESTATE =
        0x03b7eaa4e3cbce90381921a4b48008f4769871d64f93d113fcadca08ecee503b;
    bytes32 internal constant SENT_MESSAGE_SIG = keccak256("SentMessage(address,address,bytes,uint256,uint256)");

    IERC20 internal celoTokenL1;
    GnosisSafe internal parentSafe;
    IL1CrossDomainMessenger internal messenger;
    IOPContractsManager internal opcm;
    IOptimismPortal internal portal;
    IProxyAdmin internal proxyAdmin;
    ISystemConfig internal systemConfig;
    ICeloGasBridgeL1 internal bridge;
    PortalMigrator internal portalMigrator;
    MigrateV1V2 internal migrator;
    MigrateV1V2Input internal input;

    address internal safeOwner;
    uint256 internal legacyPortalBalance;
    bytes32 internal cannonPrestate;
    bytes32 internal cannonKonaPrestate;

    function setUp() public {
        string memory rpcUrl = vm.envOr("CGT_FORK_RPC_URL", string(""));
        if (bytes(rpcUrl).length == 0) {
            console.log("CGTMigration fork test skipped: CGT_FORK_RPC_URL unset");
            vm.skip(true);
        }

        // Pin a block for reproducibility; CGT_FORK_BLOCK=0 forks at latest (works on non-archival RPCs).
        uint256 forkBlock = vm.envOr("CGT_FORK_BLOCK", DEFAULT_FORK_BLOCK);
        if (forkBlock == 0) vm.createSelectFork(rpcUrl);
        else vm.createSelectFork(rpcUrl, forkBlock);

        systemConfig = ISystemConfig(SYSTEM_CONFIG_PROXY);
        proxyAdmin = IProxyAdmin(PROXY_ADMIN);
        portal = IOptimismPortal(payable(PORTAL_PROXY));
        parentSafe = GnosisSafe(payable(PARENT_SAFE));
        messenger = IL1CrossDomainMessenger(systemConfig.l1CrossDomainMessenger());
        celoTokenL1 = IERC20(_resolveCeloToken());
        safeOwner = _resolveSafeOwner();

        require(address(systemConfig.optimismPortal()) == PORTAL_PROXY, "portal mismatch");
        require(address(proxyAdmin.owner()) == PARENT_SAFE, "safe mismatch");
        require(
            proxyAdmin.getProxyImplementation(payable(PORTAL_PROXY)) == ORIGINAL_PORTAL_IMPL, "portal impl mismatch"
        );

        legacyPortalBalance = celoTokenL1.balanceOf(PORTAL_PROXY);
        (cannonPrestate, cannonKonaPrestate) = _resolvePrestates();
        opcm = IOPContractsManager(_bootstrapOpcm());

        bridge = _deployBridge();
        portalMigrator =
            new PortalMigrator(portal.proofMaturityDelaySeconds(), celoTokenL1, address(bridge), legacyPortalBalance);

        migrator = new MigrateV1V2();
        input = new MigrateV1V2Input();
        input.set(input.portalProxy.selector, PORTAL_PROXY);
        input.set(input.systemConfig.selector, address(systemConfig));
        input.set(input.proxyAdmin.selector, address(proxyAdmin));
        input.set(input.opcm.selector, address(opcm));
        input.set(input.safe.selector, PARENT_SAFE);
        input.set(input.portalMigratorImpl.selector, address(portalMigrator));
        input.set(input.originalPortalImpl.selector, ORIGINAL_PORTAL_IMPL);
        input.set(input.bridgeL1Proxy.selector, address(bridge));
        input.set(input.storageSetter.selector, address(new StorageSetter()));
        input.set(input.externalSuperchainConfig.selector, EXTERNAL_SUPERCHAIN_CONFIG);
        input.set(input.celoTokenL1.selector, address(celoTokenL1));
        input.set(input.legacyPortalBalance.selector, legacyPortalBalance);
        input.set(input.cannonPrestate.selector, cannonPrestate);
        input.set(input.cannonKonaPrestate.selector, cannonKonaPrestate);
    }

    /// @notice Drives the full 4-tx migration bundle through the real Safe and asserts the end-to-end
    ///         post-state: drain, escrow seed, OPCM v6 upgrade, CGT v2 flag, then a live deposit.
    function testFork_mainnetBundle_executesEndToEnd() external {
        migrator.preflight(input);

        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_NotActivated.selector);
        bridge.deposit(makeAddr("preMigrationRecipient"), 1, 1, hex"");

        uint64 minBaseFeeBeforeTx2 = systemConfig.minBaseFee();
        string memory versionBeforeTx2 = systemConfig.version();
        MigrateV1V2.Task[4] memory tasks = migrator.buildBundle(input);

        assertEq(tasks[0].target, MULTICALL3);
        assertEq(tasks[1].target, MULTICALL3);
        assertEq(tasks[2].target, address(opcm));
        assertEq(tasks[3].target, MULTICALL3);

        _execSafeDelegateCall(vm, PARENT_SAFE, safeOwner, tasks[0].target, tasks[0].calldata_);

        assertEq(celoTokenL1.balanceOf(PORTAL_PROXY), 0);
        assertEq(celoTokenL1.balanceOf(address(bridge)), legacyPortalBalance);
        assertEq(bridge.deposits(address(celoTokenL1), address(0)), legacyPortalBalance);
        assertTrue(bridge.escrowSeeded());
        assertEq(vm.load(PORTAL_PROXY, MIGRATED_SLOT), bytes32(uint256(1)));
        assertEq(proxyAdmin.getProxyImplementation(payable(PORTAL_PROXY)), ORIGINAL_PORTAL_IMPL);

        _execSafeDelegateCall(vm, PARENT_SAFE, safeOwner, tasks[1].target, tasks[1].calldata_);

        assertEq(address(systemConfig.superchainConfig()), EXTERNAL_SUPERCHAIN_CONFIG);
        assertEq(systemConfig.minBaseFee(), minBaseFeeBeforeTx2);
        assertEq(_hash(systemConfig.version()), _hash(versionBeforeTx2));

        _execSafeDelegateCall(vm, PARENT_SAFE, safeOwner, tasks[2].target, tasks[2].calldata_);

        assertTrue(_hash(systemConfig.version()) != _hash(versionBeforeTx2));
        assertTrue(proxyAdmin.getProxyImplementation(payable(PORTAL_PROXY)) != ORIGINAL_PORTAL_IMPL);

        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_NotCgtMode.selector);
        bridge.deposit(makeAddr("betweenRecipient"), 1, 1, hex"");

        _execSafeDelegateCall(vm, PARENT_SAFE, safeOwner, tasks[3].target, tasks[3].calldata_);

        assertTrue(systemConfig.isCustomGasToken());

        migrator.verify(input);

        uint256 trackedBefore = bridge.deposits(address(celoTokenL1), address(0));
        uint256 bridgeBalanceBefore = celoTokenL1.balanceOf(address(bridge));
        uint256 nonceBefore = messenger.messageNonce();
        address recipient = makeAddr("postMigrationRecipient");
        bytes memory extraData = hex"1234";

        deal(address(celoTokenL1), address(this), DEPOSIT_AMOUNT, true);
        celoTokenL1.approve(address(bridge), DEPOSIT_AMOUNT);

        vm.recordLogs();
        bridge.deposit(recipient, DEPOSIT_AMOUNT, DEPOSIT_GAS_LIMIT, extraData);
        Vm.Log[] memory logs = vm.getRecordedLogs();

        assertEq(bridge.deposits(address(celoTokenL1), address(0)), trackedBefore + DEPOSIT_AMOUNT);
        assertEq(celoTokenL1.balanceOf(address(bridge)), bridgeBalanceBefore + DEPOSIT_AMOUNT);
        assertEq(messenger.messageNonce(), nonceBefore + 1);
        assertTrue(_sawSentMessage(logs, address(messenger), address(bridge.otherBridge())));
    }

    /// @notice Asserts each of the 4 Safe txs stays under the EIP-7825 (Fusaka) 16M per-tx gas cap on live state.
    function testFork_bundleGasUnderPerTxCap() external {
        MigrateV1V2.Task[4] memory tasks = migrator.buildBundle(input);
        for (uint256 i; i < tasks.length; i++) {
            uint256 before = gasleft();
            _execSafeDelegateCall(vm, PARENT_SAFE, safeOwner, tasks[i].target, tasks[i].calldata_);
            assertLt(before - gasleft(), PER_TX_GAS_CAP);
        }
    }

    /// @notice After a full migration, re-running Tx1 fails and preflight rejects the already-migrated chain.
    function testFork_migrationIsNotReplayable() external {
        MigrateV1V2.Task[4] memory tasks = migrator.buildBundle(input);
        for (uint256 i; i < tasks.length; i++) {
            _execSafeDelegateCall(vm, PARENT_SAFE, safeOwner, tasks[i].target, tasks[i].calldata_);
        }

        // Re-running Tx1 fails: migrate() reverts AlreadyMigrated, so the Safe tx does not succeed.
        assertFalse(_execSafeDelegateCallAllowFail(vm, PARENT_SAFE, safeOwner, tasks[0].target, tasks[0].calldata_));

        // Preflight now rejects the migrated chain (CGT v2 already active).
        vm.expectRevert();
        migrator.preflight(input);
    }

    /// @notice Preflight reverts when the live portal impl is not the expected original.
    function testFork_preflightRejectsWrongPortalImpl() external {
        // Point the portal proxy at a different impl via its EIP-1967 slot (bypasses ProxyAdmin auth).
        vm.store(PORTAL_PROXY, EIP1967_IMPL_SLOT, bytes32(uint256(uint160(input.portalMigratorImpl()))));

        vm.expectRevert(bytes("MigrateV1V2: current portal impl != originalPortalImpl"));
        migrator.preflight(input);
    }

    /// @notice Post-migration owner pause/unpause and guardian pause gate consumers.
    function testFork_celoOwnerPause_endToEnd() external {
        _runFullMigration();

        ISuperchainConfig externalSc = ISuperchainConfig(EXTERNAL_SUPERCHAIN_CONFIG);
        IOptimismPortal v6Portal = IOptimismPortal(payable(PORTAL_PROXY));

        // Post-migration baseline: nothing paused via either authority.
        assertFalse(systemConfig.celoPaused());
        assertFalse(systemConfig.paused());
        assertFalse(v6Portal.paused());

        // Non-owner cannot pause.
        vm.expectRevert("Ownable: caller is not the owner");
        vm.prank(makeAddr("stranger"));
        systemConfig.pause();

        // Owner pauses: celoPaused + paused.
        assertEq(systemConfig.owner(), SYSTEM_CONFIG_OWNER);
        vm.prank(SYSTEM_CONFIG_OWNER);
        systemConfig.pause();
        assertTrue(systemConfig.celoPaused());
        assertTrue(systemConfig.paused());
        assertTrue(v6Portal.paused());

        // Bridge deposit is gated by the owner-pause.
        deal(address(celoTokenL1), address(this), DEPOSIT_AMOUNT, true);
        celoTokenL1.approve(address(bridge), DEPOSIT_AMOUNT);
        vm.expectRevert(ICeloGasBridgeL1.CeloGasBridgeL1_Paused.selector);
        bridge.deposit(makeAddr("recipient"), DEPOSIT_AMOUNT, DEPOSIT_GAS_LIMIT, hex"");

        // Owner unpauses: everything clears and deposits work again.
        vm.prank(SYSTEM_CONFIG_OWNER);
        systemConfig.unpause();
        assertFalse(systemConfig.celoPaused());
        assertFalse(systemConfig.paused());
        assertFalse(v6Portal.paused());
        bridge.deposit(makeAddr("recipient"), DEPOSIT_AMOUNT, DEPOSIT_GAS_LIMIT, hex"");

        // The external guardian pause is independent of the owner pause.
        assertFalse(systemConfig.celoPaused());
        vm.prank(externalSc.guardian());
        externalSc.pause(address(0));
        assertFalse(systemConfig.celoPaused());
        assertTrue(systemConfig.paused());
        assertTrue(v6Portal.paused());
    }

    /// @notice A failing sub-call in Tx1 rolls the whole bundle back: no drain, portal impl and flags untouched.
    function testFork_tx1AtomicRollbackOnFailure() external {
        // Pre-set the migration flag so migrate() reverts AlreadyMigrated mid-bundle.
        vm.store(PORTAL_PROXY, MIGRATED_SLOT, bytes32(uint256(1)));

        MigrateV1V2.Task[4] memory tasks = migrator.buildBundle(input);
        assertFalse(_execSafeDelegateCallAllowFail(vm, PARENT_SAFE, safeOwner, tasks[0].target, tasks[0].calldata_));

        // Whole tx reverted: portal not drained, impl unchanged, escrow never seeded.
        assertEq(celoTokenL1.balanceOf(PORTAL_PROXY), legacyPortalBalance);
        assertEq(celoTokenL1.balanceOf(address(bridge)), 0);
        assertEq(proxyAdmin.getProxyImplementation(payable(PORTAL_PROXY)), ORIGINAL_PORTAL_IMPL);
        assertFalse(bridge.escrowSeeded());
    }

    /// @notice Executes the full 4-tx migration bundle through the real Safe.
    function _runFullMigration() internal {
        MigrateV1V2.Task[4] memory tasks = migrator.buildBundle(input);
        for (uint256 i; i < tasks.length; i++) {
            _execSafeDelegateCall(vm, PARENT_SAFE, safeOwner, tasks[i].target, tasks[i].calldata_);
        }
    }

    function _deployBridge() internal returns (ICeloGasBridgeL1 bridge_) {
        CeloGasBridgeL1 implementation = new CeloGasBridgeL1();
        Proxy proxy = new Proxy(address(this));
        bridge_ = ICeloGasBridgeL1(payable(address(proxy)));

        vm.store(address(bridge_), Constants.PROXY_OWNER_ADDRESS, bytes32(uint256(uint160(address(this)))));
        proxy.upgradeToAndCall(
            address(implementation),
            abi.encodeCall(
                ICeloGasBridgeL1.initialize,
                (
                    ICrossDomainMessenger(address(messenger)),
                    systemConfig,
                    IStandardBridge(payable(CeloPredeploys.CELO_GAS_BRIDGE_L2)),
                    celoTokenL1
                )
            )
        );
        proxy.changeAdmin(address(proxyAdmin));
    }

    function _bootstrapOpcm() internal returns (address opcm_) {
        ISuperchainConfig externalSuperchainConfig = ISuperchainConfig(EXTERNAL_SUPERCHAIN_CONFIG);
        IProtocolVersions protocolVersions = IProtocolVersions(PROTOCOL_VERSIONS_PROXY);
        IProxyAdmin superchainProxyAdmin = IProxyAdmin(EIP1967Helper.getAdmin(address(externalSuperchainConfig)));

        require(address(externalSuperchainConfig).code.length != 0, "missing external SC");
        require(address(protocolVersions).code.length != 0, "missing ProtocolVersions");
        require(address(superchainProxyAdmin).code.length != 0, "missing SC ProxyAdmin");

        DeployImplementations.Input memory bootstrapInput = DeployImplementations.Input({
            withdrawalDelaySeconds: 100,
            minProposalSizeBytes: 200,
            challengePeriodSeconds: 300,
            proofMaturityDelaySeconds: 400,
            disputeGameFinalityDelaySeconds: 500,
            mipsVersion: StandardConstants.MIPS_VERSION,
            devFeatureBitmap: bytes32(0),
            faultGameV2MaxGameDepth: 73,
            faultGameV2SplitDepth: 30,
            faultGameV2ClockExtension: 10_800,
            faultGameV2MaxClockDuration: 302_400,
            superchainConfigProxy: externalSuperchainConfig,
            protocolVersionsProxy: protocolVersions,
            superchainProxyAdmin: superchainProxyAdmin,
            l1ProxyAdminOwner: PARENT_SAFE,
            challenger: makeAddr("challenger")
        });

        DeployImplementations.Output memory output = new DeployImplementations().run(bootstrapInput);
        opcm_ = address(output.opcm);
    }

    function _resolveCeloToken() internal view returns (address celoTokenL1_) {
        celoTokenL1_ = CELO_TOKEN_PROXY;
        if (celoTokenL1_ == address(0)) {
            (celoTokenL1_,) = IL1BlockCGT(CeloPredeploys.CELO_GAS_BRIDGE_L2).gasPayingToken();
        }
    }

    function _resolvePrestates() internal view returns (bytes32 cannon_, bytes32 cannonKona_) {
        IDisputeGameFactory disputeGameFactory = IDisputeGameFactory(systemConfig.disputeGameFactory());

        cannon_ = Claim.unwrap(IFaultDisputeGame(address(disputeGameFactory.gameImpls(GameTypes.CANNON))).absolutePrestate());

        // No live Kona game on this chain yet: reuse the live cannon prestate (read on-chain, not zero).
        address cannonKonaGame = address(disputeGameFactory.gameImpls(GameTypes.CANNON_KONA));
        cannonKona_ =
            cannonKonaGame != address(0) ? Claim.unwrap(IFaultDisputeGame(cannonKonaGame).absolutePrestate()) : cannon_;
    }

    function _resolveSafeOwner() internal view returns (address owner_) {
        address[] memory owners = parentSafe.getOwners();
        for (uint256 i; i < owners.length; i++) {
            if (owners[i] == COUNCIL_SAFE || owners[i] == CLABS_SAFE) {
                return owners[i];
            }
        }
        revert("missing safe owner");
    }

    function _sawSentMessage(Vm.Log[] memory _logs, address _messenger, address _target) internal pure returns (bool) {
        for (uint256 i; i < _logs.length; i++) {
            if (_logs[i].emitter != _messenger) continue;
            if (_logs[i].topics.length < 2) continue;
            if (_logs[i].topics[0] != SENT_MESSAGE_SIG) continue;
            if (_logs[i].topics[1] != bytes32(uint256(uint160(_target)))) continue;
            return true;
        }
        return false;
    }

    function _hash(string memory _value) internal pure returns (bytes32 hash_) {
        hash_ = keccak256(bytes(_value));
    }
}
