// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { VmSafe } from "forge-std/Vm.sol";
import { console2 as console } from "forge-std/console2.sol";
import { stdJson } from "forge-std/StdJson.sol";
import { EIP1967Helper } from "test/mocks/EIP1967Helper.sol";

// Scripts
import { Deployer } from "scripts/deploy/Deployer.sol";
import { Chains } from "scripts/libraries/Chains.sol";
import { Config } from "scripts/libraries/Config.sol";
import { StateDiff } from "scripts/libraries/StateDiff.sol";
import { ChainAssertions } from "scripts/deploy/ChainAssertions.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";
import { DeploySuperchain } from "scripts/deploy/DeploySuperchain.s.sol";
import { DeployImplementations } from "scripts/deploy/DeployImplementations.s.sol";
import { DeployAltDA } from "scripts/deploy/DeployAltDA.s.sol";
import { StandardConstants } from "scripts/deploy/StandardConstants.sol";

// Libraries
import { Constants } from "src/libraries/Constants.sol";
import { Types } from "scripts/libraries/Types.sol";
import { Duration } from "src/dispute/lib/LibUDT.sol";
import { StorageSlot, ForgeArtifacts } from "scripts/libraries/ForgeArtifacts.sol";
import { GameType, Claim, GameTypes, Proposal, Hash } from "src/dispute/lib/Types.sol";

// Interfaces
import { IOPContractsManager } from "interfaces/L1/IOPContractsManager.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { ICeloSuperchainConfig } from "interfaces/L1/ICeloSuperchainConfig.sol";
import { ISuperchainConfig } from "interfaces/L1/ISuperchainConfig.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IDisputeGameFactory } from "interfaces/dispute/IDisputeGameFactory.sol";
import { IDelayedWETH } from "interfaces/dispute/IDelayedWETH.sol";
import { IAnchorStateRegistry } from "interfaces/dispute/IAnchorStateRegistry.sol";
import { IMIPS64 } from "interfaces/cannon/IMIPS64.sol";
import { IPreimageOracle } from "interfaces/cannon/IPreimageOracle.sol";
import { IProtocolVersions } from "interfaces/L1/IProtocolVersions.sol";
import { IL1CrossDomainMessenger } from "interfaces/L1/IL1CrossDomainMessenger.sol";
import { IETHLockbox } from "interfaces/L1/IETHLockbox.sol";
import { IOptimismPortal2 } from "interfaces/L1/IOptimismPortal2.sol";
import { IL1StandardBridge } from "interfaces/L1/IL1StandardBridge.sol";
import { IL1ERC721Bridge } from "interfaces/L1/IL1ERC721Bridge.sol";
import { IOptimismMintableERC20Factory } from "interfaces/universal/IOptimismMintableERC20Factory.sol";

/// @title Deploy
/// @notice Script used to deploy a bedrock system. The entire system is deployed within the `run` function.
///         To add a new contract to the system, add a public function that deploys that individual contract.
///         Then add a call to that function inside of `run`. Be sure to call the `save` function after each
///         deployment so that hardhat-deploy style artifacts can be generated using a call to `sync()`.
///         This contract must not have constructor logic because it is set into state using `etch`.
contract Deploy is Deployer {
    using stdJson for string;

    ////////////////////////////////////////////////////////////////
    //                        Modifiers                           //
    ////////////////////////////////////////////////////////////////

    /// @notice Modifier that wraps a function in broadcasting.
    modifier broadcast() {
        vm.startBroadcast(msg.sender);
        _;
        vm.stopBroadcast();
    }

    /// @notice Modifier that will only allow a function to be called on devnet.
    modifier onlyDevnet() {
        uint256 chainid = block.chainid;
        if (chainid == Chains.LocalDevnet || chainid == Chains.GethDevnet) {
            _;
        }
    }

    /// @notice Modifier that wraps a function with statediff recording.
    ///         The returned AccountAccess[] array is then written to
    ///         the `snapshots/state-diff/<name>.json` output file.
    modifier stateDiff() {
        vm.startStateDiffRecording();
        _;
        VmSafe.AccountAccess[] memory accesses = vm.stopAndReturnStateDiff();
        console.log(
            "Writing %d state diff account accesses to snapshots/state-diff/%s.json",
            accesses.length,
            vm.toString(block.chainid)
        );
        string memory json = StateDiff.encodeAccountAccesses(accesses);
        string memory statediffPath =
            string.concat(vm.projectRoot(), "/snapshots/state-diff/", vm.toString(block.chainid), ".json");
        vm.writeJson({ json: json, path: statediffPath });
    }

    ////////////////////////////////////////////////////////////////
    //                        Accessors                           //
    ////////////////////////////////////////////////////////////////

    /// @notice The create2 salt used for deployment of the contract implementations.
    ///         Using this helps to reduce config across networks as the implementation
    ///         addresses will be the same across networks when deployed with create2.
    function _implSalt() internal view returns (bytes32) {
        return keccak256(bytes(Config.implSalt()));
    }

    /// @notice Returns the proxy addresses, not reverting if any are unset.
    function _proxies() internal view returns (Types.ContractSet memory proxies_) {
        proxies_ = Types.ContractSet({
            L1CrossDomainMessenger: artifacts.getAddress("L1CrossDomainMessengerProxy"),
            L1StandardBridge: artifacts.getAddress("L1StandardBridgeProxy"),
            L2OutputOracle: artifacts.getAddress("L2OutputOracleProxy"),
            DisputeGameFactory: artifacts.getAddress("DisputeGameFactoryProxy"),
            DelayedWETH: artifacts.getAddress("DelayedWETHProxy"),
            PermissionedDelayedWETH: artifacts.getAddress("PermissionedDelayedWETHProxy"),
            AnchorStateRegistry: artifacts.getAddress("AnchorStateRegistryProxy"),
            OptimismMintableERC20Factory: artifacts.getAddress("OptimismMintableERC20FactoryProxy"),
            OptimismPortal: artifacts.getAddress("OptimismPortalProxy"),
            ETHLockbox: artifacts.getAddress("ETHLockboxProxy"),
            SystemConfig: artifacts.getAddress("SystemConfigProxy"),
            L1ERC721Bridge: artifacts.getAddress("L1ERC721BridgeProxy"),
            ProtocolVersions: artifacts.getAddress("ProtocolVersionsProxy"),
            CeloSuperchainConfig: artifacts.getAddress("CeloSuperchainConfigProxy"),
            SuperchainConfig: artifacts.getAddress("SuperchainConfigProxy")
        });
    }

    ////////////////////////////////////////////////////////////////
    //                    SetUp and Run                           //
    ////////////////////////////////////////////////////////////////

    /// @notice Deploy all of the L1 contracts necessary for a full Superchain with a single Op Chain.
    function run() public {
        console.log("Deploying a fresh OP Stack including SuperchainConfig");
        _run({ _needsSuperchain: true });
    }

    /// @notice Test-only entrypoint: skip the fresh SuperchainConfig deploy when false.
    function run(bool _needsSuperchain) public {
        _run({ _needsSuperchain: _needsSuperchain });
    }

    /// @notice Deploy a fresh OP Stack for Celo, wrapping cfg.externalSuperchainConfig() as the
    ///         SuperchainConfig. ProtocolVersions is passed in; SuperchainProxyAdmin is deployed
    ///         fresh so Celo owns it.
    /// @param _protocolVersionsProxy Address of the existing ProtocolVersions proxy.
    function runCelo(address payable _protocolVersionsProxy) public {
        address externalSC = cfg.externalSuperchainConfig();
        require(externalSC != address(0), "Deploy: must provide externalSuperchainConfig in deploy config");
        require(_protocolVersionsProxy != address(0), "Deploy: must specify address for protocol versions proxy");

        vm.chainId(cfg.l1ChainID());

        console.log("Deploying OP Stack for Celo, wrapping external SuperchainConfig at %s", externalSC);

        artifacts.save("SuperchainConfigImpl", EIP1967Helper.getImplementation(externalSC));
        artifacts.save("SuperchainConfigProxy", externalSC);

        artifacts.save("ProtocolVersionsImpl", EIP1967Helper.getImplementation(_protocolVersionsProxy));
        artifacts.save("ProtocolVersionsProxy", _protocolVersionsProxy);

        _deploySuperchainProxyAdmin();

        _run({ _needsSuperchain: false });

        // Celo: transfer last so the CeloSuperchainConfig upgrade inside _run broadcasts as the
        //       deployer, not the final multisig owner.
        _transferSuperchainProxyAdminOwnership();
    }

    /// @notice Deploy a fresh SuperchainProxyAdmin owned by the deployer.
    function _deploySuperchainProxyAdmin() internal {
        vm.broadcast(msg.sender);
        IProxyAdmin superchainProxyAdmin = IProxyAdmin(
            DeployUtils.create1({
                _name: "ProxyAdmin",
                _args: DeployUtils.encodeConstructor(abi.encodeCall(IProxyAdmin.__constructor__, (msg.sender)))
            })
        );

        vm.label(address(superchainProxyAdmin), "SuperchainProxyAdmin");
        artifacts.save("SuperchainProxyAdmin", address(superchainProxyAdmin));
    }

    /// @notice Hand SuperchainProxyAdmin ownership to `cfg.finalSystemOwner()`.
    function _transferSuperchainProxyAdminOwnership() internal {
        IProxyAdmin superchainProxyAdmin = IProxyAdmin(artifacts.mustGetAddress("SuperchainProxyAdmin"));
        vm.startBroadcast(msg.sender);
        superchainProxyAdmin.transferOwnership(cfg.finalSystemOwner());
        vm.stopBroadcast();
    }

    /// @notice Deploy a new OP Chain using an existing SuperchainConfig and ProtocolVersions
    /// @param _superchainConfigProxy Address of the existing SuperchainConfig proxy
    /// @param _protocolVersionsProxy Address of the existing ProtocolVersions proxy
    function runWithSuperchain(address payable _superchainConfigProxy, address payable _protocolVersionsProxy) public {
        require(_superchainConfigProxy != address(0), "Deploy: must specify address for superchain config proxy");
        require(_protocolVersionsProxy != address(0), "Deploy: must specify address for protocol versions proxy");

        vm.chainId(cfg.l1ChainID());

        console.log("Deploying a fresh OP Stack with existing SuperchainConfig and ProtocolVersions");

        IProxy scProxy = IProxy(_superchainConfigProxy);
        artifacts.save("SuperchainConfigImpl", scProxy.implementation());
        artifacts.save("SuperchainConfigProxy", _superchainConfigProxy);

        IProxy pvProxy = IProxy(_protocolVersionsProxy);
        artifacts.save("ProtocolVersionsImpl", pvProxy.implementation());
        artifacts.save("ProtocolVersionsProxy", _protocolVersionsProxy);

        // setupCeloSuperchainConfig() (called inside _run) needs SuperchainProxyAdmin.
        // Use the existing proxy admin of the supplied SuperchainConfig.
        artifacts.save("SuperchainProxyAdmin", EIP1967Helper.getAdmin(_superchainConfigProxy));

        _run({ _needsSuperchain: false });
    }

    /// @notice Deploy all L1 contracts and write the state diff to a file.
    ///         Used to generate kontrol tests.
    function runWithStateDiff() public stateDiff {
        _run({ _needsSuperchain: true });
    }

    /// @notice Internal function containing the deploy logic.
    function _run(bool _needsSuperchain) internal virtual {
        console.log("start of L1 Deploy!");

        // Set up the Superchain if needed.
        if (_needsSuperchain) {
            deploySuperchain();
        } else {
            // No fresh superchain. If SuperchainConfigProxy isn't already pre-saved (e.g. by
            // runCelo / runWithSuperchain), fall back to cfg.externalSuperchainConfig().
            if (artifacts.getAddress("SuperchainConfigProxy") == address(0)) {
                address externalSC = cfg.externalSuperchainConfig();
                require(externalSC != address(0), "Deploy: externalSuperchainConfig is zero");
                console.log("Using external SuperchainConfig at %s", externalSC);
                artifacts.save("SuperchainConfigProxy", externalSC);
            }
            // PV + SuperchainProxyAdmin are still required downstream; fail fast with a useful
            // message instead of letting DeploymentDoesNotExist surface deep in the deploy.
            require(
                artifacts.getAddress("ProtocolVersionsProxy") != address(0),
                "Deploy: ProtocolVersionsProxy not seeded (use runCelo / runWithSuperchain)"
            );
            require(
                artifacts.getAddress("SuperchainProxyAdmin") != address(0),
                "Deploy: SuperchainProxyAdmin not seeded (use runCelo / runWithSuperchain)"
            );
        }

        deployImplementations({ _isInterop: cfg.useInterop() });

        // Must run before deployOpChain so CeloSuperchainConfigProxy resolves in _proxies().
        setupCeloSuperchainConfig();

        // Deploy Current OPChain Contracts
        deployOpChain();

        // Set the respected game type according to the deploy config
        address guardianSuperchainConfig = artifacts.getAddress("CeloSuperchainConfigProxy");
        if (guardianSuperchainConfig == address(0)) {
            guardianSuperchainConfig = artifacts.mustGetAddress("SuperchainConfigProxy");
        }
        vm.startPrank(ISuperchainConfig(guardianSuperchainConfig).guardian());
        IAnchorStateRegistry(artifacts.mustGetAddress("AnchorStateRegistryProxy")).setRespectedGameType(
            GameType.wrap(uint32(cfg.respectedGameType()))
        );
        vm.stopPrank();

        if (cfg.useCustomGasToken()) {
            // Reset the systemconfig then reinitialize it with the custom gas token
            resetInitializedProxy("SystemConfig");
            initializeSystemConfig();
        }

        if (cfg.useAltDA()) {
            bytes32 typeHash = keccak256(bytes(cfg.daCommitmentType()));
            bytes32 keccakHash = keccak256(bytes("KeccakCommitment"));
            if (typeHash == keccakHash) {
                console.log("Deploying OP AltDA");

                DeployAltDA da = new DeployAltDA();
                DeployAltDA.Input memory dii = DeployAltDA.Input({
                    salt: _implSalt(),
                    proxyAdmin: IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin")),
                    challengeContractOwner: cfg.finalSystemOwner(),
                    challengeWindow: cfg.daChallengeWindow(),
                    resolveWindow: cfg.daResolveWindow(),
                    bondSize: cfg.daBondSize(),
                    resolverRefundPercentage: cfg.daResolverRefundPercentage()
                });

                DeployAltDA.Output memory dio = da.run(dii);

                artifacts.save("DataAvailabilityChallengeProxy", address(dio.dataAvailabilityChallengeProxy));
                artifacts.save("DataAvailabilityChallengeImpl", address(dio.dataAvailabilityChallengeImpl));
            }
        }

        console.log("set up op chain!");
    }

    ////////////////////////////////////////////////////////////////
    //           High Level Deployment Functions                  //
    ////////////////////////////////////////////////////////////////

    /// @notice Deploy a full system with a new SuperchainConfig
    ///         The Superchain system has 2 singleton contracts which lie outside of an OP Chain:
    ///         1. The SuperchainConfig contract
    ///         2. The ProtocolVersions contract
    function deploySuperchain() public {
        console.log("Setting up Superchain");
        DeploySuperchain ds = new DeploySuperchain();

        // Run the deployment script.
        DeploySuperchain.Output memory dso = ds.run(
            DeploySuperchain.Input({
                guardian: cfg.superchainConfigGuardian(),
                // TODO: when DeployAuthSystem is done, finalSystemOwner should be replaced with the Foundation Upgrades
                // Safe
                protocolVersionsOwner: cfg.finalSystemOwner(),
                superchainProxyAdminOwner: cfg.finalSystemOwner(),
                paused: false,
                recommendedProtocolVersion: bytes32(cfg.recommendedProtocolVersion()),
                requiredProtocolVersion: bytes32(cfg.requiredProtocolVersion())
            })
        );

        // Store the artifacts
        artifacts.save("SuperchainProxyAdmin", address(dso.superchainProxyAdmin));
        artifacts.save("SuperchainConfigProxy", address(dso.superchainConfigProxy));
        artifacts.save("SuperchainConfigImpl", address(dso.superchainConfigImpl));
        artifacts.save("ProtocolVersionsProxy", address(dso.protocolVersionsProxy));
        artifacts.save("ProtocolVersionsImpl", address(dso.protocolVersionsImpl));

        // First run assertions for the ProtocolVersions and SuperchainConfig proxy contracts.
        Types.ContractSet memory contracts = _proxies();
        ChainAssertions.checkProtocolVersions({ _contracts: contracts, _cfg: cfg, _isProxy: true });
        ChainAssertions.checkSuperchainConfig({ _contracts: contracts, _cfg: cfg, _isProxy: true });

        // Then replace the ProtocolVersions proxy with the implementation address and run assertions on it.
        contracts.ProtocolVersions = artifacts.mustGetAddress("ProtocolVersionsImpl");
        ChainAssertions.checkProtocolVersions({ _contracts: contracts, _cfg: cfg, _isProxy: false });

        // Finally replace the SuperchainConfig proxy with the implementation address and run assertions on it.
        contracts.SuperchainConfig = artifacts.mustGetAddress("SuperchainConfigImpl");
        ChainAssertions.checkSuperchainConfig({ _contracts: contracts, _cfg: cfg, _isProxy: false });
    }

    /// @notice Deploy all of the implementations
    /// @param _isInterop Whether to use interop
    function deployImplementations(bool _isInterop) public {
        // TODO _isInterop is no longer being used in DeployImplementations, this might no longer be necessary
        require(_isInterop == cfg.useInterop(), "Deploy: Interop setting mismatch.");

        console.log("Deploying implementations");

        DeployImplementations di = new DeployImplementations();

        ISuperchainConfig superchainConfigProxy = ISuperchainConfig(artifacts.mustGetAddress("SuperchainConfigProxy"));
        IProxyAdmin superchainProxyAdmin = IProxyAdmin(EIP1967Helper.getAdmin(address(superchainConfigProxy)));

        DeployImplementations.Output memory dio = di.run(
            DeployImplementations.Input({
                withdrawalDelaySeconds: cfg.faultGameWithdrawalDelay(),
                minProposalSizeBytes: cfg.preimageOracleMinProposalSize(),
                challengePeriodSeconds: cfg.preimageOracleChallengePeriod(),
                proofMaturityDelaySeconds: cfg.proofMaturityDelaySeconds(),
                disputeGameFinalityDelaySeconds: cfg.disputeGameFinalityDelaySeconds(),
                mipsVersion: StandardConstants.MIPS_VERSION,
                devFeatureBitmap: cfg.devFeatureBitmap(),
                faultGameV2MaxGameDepth: cfg.faultGameV2MaxGameDepth(),
                faultGameV2SplitDepth: cfg.faultGameV2SplitDepth(),
                faultGameV2ClockExtension: cfg.faultGameV2ClockExtension(),
                faultGameV2MaxClockDuration: cfg.faultGameV2MaxClockDuration(),
                protocolVersionsProxy: IProtocolVersions(artifacts.mustGetAddress("ProtocolVersionsProxy")),
                superchainConfigProxy: superchainConfigProxy,
                superchainProxyAdmin: superchainProxyAdmin,
                l1ProxyAdminOwner: superchainProxyAdmin.owner(),
                challenger: cfg.l2OutputOracleChallenger()
            })
        );

        // Save the implementation addresses which are needed outside of this function or script.
        // When called in a fork test, this will overwrite the existing implementations.
        artifacts.save("MipsSingleton", address(dio.mipsSingleton));
        artifacts.save("OPContractsManager", address(dio.opcm));
        artifacts.save("DelayedWETHImpl", address(dio.delayedWETHImpl));
        artifacts.save("PreimageOracle", address(dio.preimageOracleSingleton));
        artifacts.save("SystemConfigImpl", address(dio.systemConfigImpl));
        artifacts.save("CeloSuperchainConfigImpl", address(dio.celoSuperchainConfigImpl));

        // Get a contract set from the implementation addresses which were just deployed.
        Types.ContractSet memory impls = ChainAssertions.dioToContractSet(dio);

        ChainAssertions.checkL1CrossDomainMessenger(IL1CrossDomainMessenger(impls.L1CrossDomainMessenger), vm, false);
        ChainAssertions.checkL1StandardBridgeImpl(IL1StandardBridge(payable(impls.L1StandardBridge)));
        ChainAssertions.checkL1ERC721BridgeImpl(IL1ERC721Bridge(impls.L1ERC721Bridge));
        ChainAssertions.checkOptimismPortal2({
            _contracts: impls,
            _superchainConfig: superchainConfigProxy,
            _opChainProxyAdminOwner: cfg.finalSystemOwner(),
            _isProxy: false
        });
        ChainAssertions.checkETHLockboxImpl(
            IETHLockbox(impls.ETHLockbox), IOptimismPortal2(payable(impls.OptimismPortal))
        );
        ChainAssertions.checkOptimismMintableERC20FactoryImpl(
            IOptimismMintableERC20Factory(impls.OptimismMintableERC20Factory)
        );
        ChainAssertions.checkDisputeGameFactory(
            IDisputeGameFactory(impls.DisputeGameFactory), address(0), address(0), false
        );
        ChainAssertions.checkDelayedWETHImpl(IDelayedWETH(payable(impls.DelayedWETH)), cfg.faultGameWithdrawalDelay());
        ChainAssertions.checkMIPS({
            _mips: IMIPS64(address(dio.mipsSingleton)),
            _oracle: IPreimageOracle(address(dio.preimageOracleSingleton))
        });
        ChainAssertions.checkOPContractsManager({
            _impls: impls,
            _proxies: _proxies(),
            _opcm: IOPContractsManager(address(dio.opcm)),
            _mips: IMIPS64(address(dio.mipsSingleton))
        });
        ChainAssertions.checkSystemConfigImpls(impls);
        ChainAssertions.checkAnchorStateRegistryProxy(IAnchorStateRegistry(impls.AnchorStateRegistry), false);
    }

    /// @notice Deploy all of the OP Chain specific contracts
    function deployOpChain() public {
        console.log("Deploying OP Chain");

        // Ensure that the requisite contracts are deployed
        IOPContractsManager opcm = IOPContractsManager(artifacts.mustGetAddress("OPContractsManager"));

        IOPContractsManager.DeployInput memory deployInput = getDeployInput();
        IOPContractsManager.DeployOutput memory deployOutput = opcm.deploy(deployInput);

        // Store code in the Final system owner address so that it can be used for prank delegatecalls
        // Store "fe" opcode so that accidental calls to this address revert
        vm.etch(cfg.finalSystemOwner(), hex"fe");

        // Save all deploy outputs from the OPCM, in the order they are declared in the DeployOutput struct
        artifacts.save("ProxyAdmin", address(deployOutput.opChainProxyAdmin));
        artifacts.save("AddressManager", address(deployOutput.addressManager));
        artifacts.save("L1ERC721BridgeProxy", address(deployOutput.l1ERC721BridgeProxy));
        artifacts.save("SystemConfigProxy", address(deployOutput.systemConfigProxy));
        artifacts.save("OptimismMintableERC20FactoryProxy", address(deployOutput.optimismMintableERC20FactoryProxy));
        artifacts.save("L1StandardBridgeProxy", address(deployOutput.l1StandardBridgeProxy));
        artifacts.save("L1CrossDomainMessengerProxy", address(deployOutput.l1CrossDomainMessengerProxy));
        artifacts.save("ETHLockboxProxy", address(deployOutput.ethLockboxProxy));

        // Fault Proof contracts
        artifacts.save("DisputeGameFactoryProxy", address(deployOutput.disputeGameFactoryProxy));
        artifacts.save("PermissionedDelayedWETHProxy", address(deployOutput.delayedWETHPermissionedGameProxy));
        artifacts.save("AnchorStateRegistryProxy", address(deployOutput.anchorStateRegistryProxy));
        artifacts.save("PermissionedDisputeGame", address(deployOutput.permissionedDisputeGame));
        artifacts.save("OptimismPortalProxy", address(deployOutput.optimismPortalProxy));
        artifacts.save("OptimismPortal2Proxy", address(deployOutput.optimismPortalProxy));

        // Check if the permissionless game implementation is already set
        IDisputeGameFactory factory = IDisputeGameFactory(artifacts.mustGetAddress("DisputeGameFactoryProxy"));
        address permissionlessGameImpl = address(factory.gameImpls(GameTypes.CANNON));

        // Deploy and setup the PermissionlessDelayedWeth not provided by the OPCM.
        // If the following require statement is hit, you can delete the block of code after it.
        require(
            permissionlessGameImpl == address(0),
            "Deploy: The PermissionlessDelayedWETH is already set by the OPCM, it is no longer necessary to deploy it separately."
        );
        address delayedWETHImpl = artifacts.mustGetAddress("DelayedWETHImpl");
        address delayedWETHPermissionlessGameProxy =
            deployERC1967ProxyWithOwner("DelayedWETHProxy", address(deployOutput.opChainProxyAdmin));
        vm.broadcast(address(deployOutput.opChainProxyAdmin));
        IProxy(payable(delayedWETHPermissionlessGameProxy)).upgradeToAndCall({
            _implementation: delayedWETHImpl,
            _data: abi.encodeCall(IDelayedWETH.initialize, (deployOutput.systemConfigProxy))
        });
    }

    /// @notice Deploy the CeloSuperchainConfig proxy under SuperchainProxyAdmin and initialize
    ///         it. The implementation is deployed earlier by DeployImplementations.
    function setupCeloSuperchainConfig() public {
        console.log("Setting up CeloSuperchainConfig");
        IProxyAdmin superchainProxyAdmin = IProxyAdmin(artifacts.mustGetAddress("SuperchainProxyAdmin"));
        deployERC1967ProxyWithOwner("CeloSuperchainConfigProxy", address(superchainProxyAdmin));
        initializeCeloSuperchainConfig();
    }

    /// @notice Upgrade the CeloSuperchainConfig proxy to its impl and initialize it with the
    ///         Celo guardian and the SuperchainConfig pointer.
    function initializeCeloSuperchainConfig() public {
        address payable superchainConfigProxy = artifacts.mustGetAddress("SuperchainConfigProxy");
        address payable celoSuperchainConfigProxy = artifacts.mustGetAddress("CeloSuperchainConfigProxy");
        address celoSuperchainConfigImpl = artifacts.mustGetAddress("CeloSuperchainConfigImpl");
        IProxyAdmin superchainProxyAdmin = IProxyAdmin(artifacts.mustGetAddress("SuperchainProxyAdmin"));

        vm.startBroadcast(superchainProxyAdmin.owner());
        superchainProxyAdmin.upgradeAndCall({
            _proxy: celoSuperchainConfigProxy,
            _implementation: celoSuperchainConfigImpl,
            _data: abi.encodeCall(
                ICeloSuperchainConfig.initialize,
                (cfg.superchainConfigGuardian(), false, superchainConfigProxy)
            )
        });
        vm.stopBroadcast();

        // Assumes the wrapped (external) SuperchainConfig isn't paused at deploy time.
        ChainAssertions.checkCeloSuperchainConfig({ _contracts: _proxies(), _cfg: cfg, _isPaused: false });
    }

    ////////////////////////////////////////////////////////////////
    //                Proxy Deployment Functions                  //
    ////////////////////////////////////////////////////////////////

    /// @notice Deploys an ERC1967Proxy contract with a specified owner.
    /// @param _name The name of the proxy contract to be deployed.
    /// @param _proxyOwner The address of the owner of the proxy contract.
    /// @return addr_ The address of the deployed proxy contract.
    function deployERC1967ProxyWithOwner(
        string memory _name,
        address _proxyOwner
    )
        public
        broadcast
        returns (address addr_)
    {
        IProxy proxy = IProxy(
            DeployUtils.create2AndSave({
                _save: artifacts,
                _salt: keccak256(abi.encode(_implSalt(), _name)),
                _name: "src/universal/Proxy.sol:Proxy", // Espresso: disambiguate from OZ v5 proxy/Proxy.sol artifact
                _nick: _name,
                _args: DeployUtils.encodeConstructor(abi.encodeCall(IProxy.__constructor__, (_proxyOwner)))
            })
        );
        require(EIP1967Helper.getAdmin(address(proxy)) == _proxyOwner, "Deploy: EIP1967Proxy admin not set");
        addr_ = address(proxy);
    }

    /// @notice Initialize the SystemConfig
    function initializeSystemConfig() public {
        vm.startBroadcast(msg.sender);

        console.log("Upgrading and initializing SystemConfig proxy");
        address systemConfigProxy = artifacts.mustGetAddress("SystemConfigProxy");
        address systemConfig = artifacts.mustGetAddress("SystemConfigImpl");

        bytes32 batcherHash = bytes32(uint256(uint160(cfg.batchSenderAddress())));

        address customGasTokenAddress = Constants.ETHER;
        if (cfg.useCustomGasToken()) {
            customGasTokenAddress = cfg.customGasTokenAddress();
        }

        IProxyAdmin proxyAdmin = IProxyAdmin(payable(artifacts.mustGetAddress("ProxyAdmin")));
        vm.stopBroadcast();

        vm.startBroadcast(proxyAdmin.owner());
        proxyAdmin.upgradeAndCall({
            _proxy: payable(systemConfigProxy),
            _implementation: systemConfig,
            _data: abi.encodeCall(
                ISystemConfig.initialize,
                (
                    cfg.finalSystemOwner(),
                    cfg.basefeeScalar(),
                    cfg.blobbasefeeScalar(),
                    batcherHash,
                    uint64(cfg.l2GenesisBlockGasLimit()),
                    cfg.p2pSequencerAddress(),
                    Constants.DEFAULT_RESOURCE_CONFIG(),
                    cfg.batchInboxAddress(),
                    ISystemConfig.Addresses({
                        l1CrossDomainMessenger: artifacts.mustGetAddress("L1CrossDomainMessengerProxy"),
                        l1ERC721Bridge: artifacts.mustGetAddress("L1ERC721BridgeProxy"),
                        l1StandardBridge: artifacts.mustGetAddress("L1StandardBridgeProxy"),
                        optimismPortal: artifacts.mustGetAddress("OptimismPortalProxy"),
                        optimismMintableERC20Factory: artifacts.mustGetAddress("OptimismMintableERC20FactoryProxy"),
                        gasPayingToken: customGasTokenAddress
                    }),
                    cfg.l2ChainID(),
                    ISuperchainConfig(artifacts.mustGetAddress("CeloSuperchainConfigProxy"))
                )
            )
        });
        vm.stopBroadcast();

        vm.startBroadcast(msg.sender);
        ISystemConfig config = ISystemConfig(systemConfigProxy);
        string memory version = config.version();
        console.log("SystemConfig version: %s", version);

        IOPContractsManager.DeployInput memory di = getDeployInput();
        Types.DeployOPChainInput memory doi = Types.DeployOPChainInput({
            opChainProxyAdminOwner: di.roles.opChainProxyAdminOwner,
            systemConfigOwner: di.roles.systemConfigOwner,
            batcher: di.roles.batcher,
            unsafeBlockSigner: di.roles.unsafeBlockSigner,
            proposer: di.roles.proposer,
            challenger: di.roles.challenger,
            basefeeScalar: di.basefeeScalar,
            blobBaseFeeScalar: di.blobBasefeeScalar,
            l2ChainId: di.l2ChainId,
            opcm: artifacts.mustGetAddress("OPContractsManager"),
            saltMixer: di.saltMixer,
            gasLimit: di.gasLimit,
            disputeGameType: di.disputeGameType,
            disputeAbsolutePrestate: di.disputeAbsolutePrestate,
            disputeMaxGameDepth: di.disputeMaxGameDepth,
            disputeSplitDepth: di.disputeSplitDepth,
            disputeClockExtension: di.disputeClockExtension,
            disputeMaxClockDuration: di.disputeMaxClockDuration,
            allowCustomDisputeParameters: false,
            operatorFeeScalar: 0,
            operatorFeeConstant: 0
        });
        ChainAssertions.checkSystemConfigProxies({ _contracts: _proxies(), _doi: doi });
        vm.stopBroadcast();
    }

    /// @notice Get the DeployInput struct to use for testing
    function getDeployInput() public view returns (IOPContractsManager.DeployInput memory) {
        string memory saltMixer = "salt mixer";
        return IOPContractsManager.DeployInput({
            roles: IOPContractsManager.Roles({
                opChainProxyAdminOwner: cfg.finalSystemOwner(),
                systemConfigOwner: cfg.finalSystemOwner(),
                batcher: cfg.batchSenderAddress(),
                unsafeBlockSigner: cfg.p2pSequencerAddress(),
                proposer: cfg.l2OutputOracleProposer(),
                challenger: cfg.l2OutputOracleChallenger()
            }),
            basefeeScalar: cfg.basefeeScalar(),
            blobBasefeeScalar: cfg.blobbasefeeScalar(),
            l2ChainId: cfg.l2ChainID(),
            startingAnchorRoot: abi.encode(
                Proposal({ root: Hash.wrap(cfg.faultGameGenesisOutputRoot()), l2SequenceNumber: cfg.faultGameGenesisBlock() })
            ),
            saltMixer: saltMixer,
            gasLimit: uint64(cfg.l2GenesisBlockGasLimit()),
            disputeGameType: GameTypes.PERMISSIONED_CANNON,
            disputeAbsolutePrestate: Claim.wrap(bytes32(cfg.faultGameAbsolutePrestate())),
            disputeMaxGameDepth: cfg.faultGameMaxDepth(),
            disputeSplitDepth: cfg.faultGameSplitDepth(),
            disputeClockExtension: Duration.wrap(uint64(cfg.faultGameClockExtension())),
            disputeMaxClockDuration: Duration.wrap(uint64(cfg.faultGameMaxClockDuration())),
            superchainConfigOverride: artifacts.getAddress("CeloSuperchainConfigProxy")
        });
    }

    /// @notice Reset the initialized value on a proxy contract so that it can be initialized again
    function resetInitializedProxy(string memory _contractName) internal {
        console.log("resetting initialized value on %s Proxy", _contractName);
        address proxy = artifacts.mustGetAddress(string.concat(_contractName, "Proxy"));
        StorageSlot memory slot = ForgeArtifacts.getInitializedSlot(_contractName);
        bytes32 slotVal = vm.load(proxy, bytes32(slot.slot));
        uint256 value = uint256(slotVal);
        value = value & ~(0xFF << (slot.offset * 8));
        slotVal = bytes32(value);
        vm.store(proxy, bytes32(slot.slot), slotVal);
    }
}
