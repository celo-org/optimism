// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Test } from "forge-std/Test.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { LegacyCeloVaultFixture, LegacyCeloL1BlockFixture } from "test/fixtures/cgtv2/LegacyPredeploys.sol";
import { GenerateCgtV2NUTBundle } from "scripts/cgt/GenerateCgtV2NUTBundle.s.sol";
import { Proxy } from "src/universal/Proxy.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";
import { Constants } from "src/libraries/Constants.sol";
import { Types } from "src/libraries/Types.sol";
import { IL1Block } from "interfaces/L2/IL1Block.sol";
import { ILiquidityController } from "interfaces/L2/ILiquidityController.sol";
import { IFeeVault } from "interfaces/L2/IFeeVault.sol";
import { ICeloGasBridgeL2 } from "interfaces/celo/ICeloGasBridgeL2.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";

contract CgtFeeTokenFixture is ERC20 {
    constructor() ERC20("Fixture fee token", "FIX") { }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

interface ICgtUpgrade {
    function upgrade() external;
}

interface IUpgradeProxyAdmin {
    function upgradePredeploys(address manager) external;
}

contract RejectCgtInitializerFixture {
    fallback() external {
        revert("fixture initializer failure");
    }
}

contract CgtV2Upgrade_Test is Test {
    // Fixture only. Production allocation remains unresolved.
    address internal constant BRIDGE = 0x42000000000000000000000000000000000007fE;
    bytes32 internal constant ADMIN_SLOT = bytes32(uint256(keccak256("eip1967.proxy.admin")) - 1);
    bytes32 internal constant IMPL_SLOT = bytes32(uint256(keccak256("eip1967.proxy.implementation")) - 1);
    bytes32 internal constant CGT_SLOT = 0xd2ff82c9b477ff6a09f530b1c627ffb4b0b81e2ae2ba427f824162e8dad020aa;
    address internal constant OWNER = address(0xBEEF);
    address internal constant RECIPIENT = address(0xCAFE);
    address internal constant L1_BRIDGE = address(0x1234);
    address internal constant L1_TOKEN = address(0x5678);

    struct Implementations {
        address storageSetter;
        address l1Block;
        address messagePasser;
        address nativeLiquidity;
        address controller;
        address bridge;
        address vault;
    }

    struct NetworkConfig {
        address bridgeProxy;
        address l1Bridge;
        address l1Token;
        address owner;
    }

    Implementations internal implementations;
    address internal manager;
    CgtFeeTokenFixture internal feeToken;

    function setUp() public {
        address[7] memory proxies = [
            Predeploys.PROXY_ADMIN,
            Predeploys.L1_BLOCK_ATTRIBUTES,
            Predeploys.L2_TO_L1_MESSAGE_PASSER,
            Predeploys.NATIVE_ASSET_LIQUIDITY,
            Predeploys.LIQUIDITY_CONTROLLER,
            Predeploys.SEQUENCER_FEE_WALLET,
            BRIDGE
        ];
        bytes memory proxyCode = address(new Proxy(Predeploys.PROXY_ADMIN)).code;
        for (uint256 i; i < proxies.length; i++) {
            vm.etch(proxies[i], proxyCode);
            vm.store(proxies[i], ADMIN_SLOT, bytes32(uint256(uint160(Predeploys.PROXY_ADMIN))));
        }
        _install(Predeploys.L1_BLOCK_ATTRIBUTES, address(new LegacyCeloL1BlockFixture()));
        _install(Predeploys.SEQUENCER_FEE_WALLET, address(new LegacyCeloVaultFixture(RECIPIENT)));
        // Counter belongs to the proxy, not the fixture implementation.
        vm.store(Predeploys.SEQUENCER_FEE_WALLET, bytes32(0), bytes32(uint256(73)));
        vm.deal(Predeploys.SEQUENCER_FEE_WALLET, 41);
        _install(Predeploys.PROXY_ADMIN, deployCode("L2ProxyAdmin.sol:L2ProxyAdmin"));
        implementations = Implementations({
            storageSetter: deployCode("StorageSetter.sol:StorageSetter"),
            l1Block: deployCode("L1BlockCGT.sol:L1BlockCGT"),
            messagePasser: deployCode("L2ToL1MessagePasserCGT.sol:L2ToL1MessagePasserCGT"),
            nativeLiquidity: deployCode("NativeAssetLiquidity.sol:NativeAssetLiquidity"),
            controller: deployCode("LiquidityController.sol:LiquidityController"),
            bridge: deployCode("CeloGasBridgeL2.sol:CeloGasBridgeL2"),
            vault: deployCode("CeloSequencerFeeVault.sol:CeloSequencerFeeVault")
        });
        manager = _manager(implementations, BRIDGE);
        feeToken = new CgtFeeTokenFixture();
        feeToken.mint(Predeploys.SEQUENCER_FEE_WALLET, 19);
    }

    function _install(address proxy, address implementation) internal {
        vm.prank(Predeploys.PROXY_ADMIN);
        Proxy(payable(proxy)).upgradeTo(implementation);
    }

    function _manager(Implementations memory impls, address bridge) internal returns (address) {
        return deployCode(
            "CeloCGTV2ContractsManager.sol:CeloCGTV2ContractsManager",
            abi.encode(impls, NetworkConfig(bridge, L1_BRIDGE, L1_TOKEN, OWNER))
        );
    }

    function _upgrade(address target) internal {
        vm.prank(Constants.DEPOSITOR_ACCOUNT);
        IUpgradeProxyAdmin(Predeploys.PROXY_ADMIN).upgradePredeploys(target);
    }

    function test_upgrade_preservesVaultAndInitializesReservedBridge_succeeds() external {
        assertGt(BRIDGE.code.length, 0);
        assertEq(vm.load(Predeploys.LIQUIDITY_CONTROLLER, IMPL_SLOT), bytes32(0));
        _upgrade(manager);
        assertEq(vm.load(BRIDGE, IMPL_SLOT), bytes32(uint256(uint160(implementations.bridge))));
        IFeeVault vault = IFeeVault(payable(Predeploys.SEQUENCER_FEE_WALLET));
        assertEq(vault.recipient(), RECIPIENT);
        assertEq(vault.minWithdrawalAmount(), 23);
        assertEq(uint256(vault.withdrawalNetwork()), uint256(Types.WithdrawalNetwork.L2));
        assertEq(vault.totalProcessed(), 73);
        assertEq(Predeploys.SEQUENCER_FEE_WALLET.balance, 41);
        assertEq(feeToken.balanceOf(Predeploys.SEQUENCER_FEE_WALLET), 19);
        ICeloGasBridgeL2 bridge = ICeloGasBridgeL2(payable(BRIDGE));
        assertEq(address(bridge.otherBridge()), L1_BRIDGE);
        assertEq(bridge.celoTokenL1(), L1_TOKEN);
        ILiquidityController controller = ILiquidityController(Predeploys.LIQUIDITY_CONTROLLER);
        assertEq(controller.owner(), OWNER);
        assertTrue(controller.minters(BRIDGE));
        assertEq(controller.gasPayingTokenName(), "Celo");
        assertEq(controller.gasPayingTokenSymbol(), "CELO");
        assertEq(vm.load(Predeploys.L1_BLOCK_ATTRIBUTES, CGT_SLOT), bytes32(0));
        vm.prank(Constants.DEPOSITOR_ACCOUNT);
        (bool success,) = Predeploys.L1_BLOCK_ATTRIBUTES.call(abi.encodeWithSignature("setCustomGasToken()"));
        assertTrue(success);
        assertTrue(IL1Block(Predeploys.L1_BLOCK_ATTRIBUTES).isCustomGasToken());
        vm.expectRevert();
        vault.initialize(RECIPIENT, 23, Types.WithdrawalNetwork.L2);
    }

    struct NutTx {
        string intent;
        address from;
        address to;
        bytes data;
        uint64 gasLimit;
        bool isCreation;
    }

    function test_upgrade_generatedNUTGroup_executes_succeeds() external {
        address generator = deployCode("GenerateCgtV2NUTBundle.s.sol:GenerateCgtV2NUTBundle");
        (bool generated, bytes memory data) = generator.call(
            abi.encodeWithSignature(
                "transactions((address,address,address,address))", NetworkConfig(BRIDGE, L1_BRIDGE, L1_TOKEN, OWNER)
            )
        );
        assertTrue(generated);
        NutTx[] memory txs = abi.decode(data, (NutTx[]));
        assertEq(txs.length, 12);
        for (uint256 i; i < txs.length; i++) {
            if (txs[i].isCreation) {
                address expected = vm.computeCreateAddress(txs[i].from, 0);
                vm.broadcast(txs[i].from);
                bytes memory code = txs[i].data;
                address deployed;
                uint256 beforeCreate = gasleft();
                assembly { deployed := create(0, add(code, 32), mload(code)) }
                uint256 executionGas = beforeCreate - gasleft();
                // Conservative bound: CREATE already pays its base cost. Add a transaction
                // base and the maximum calldata floor, including zero bytes, for deposits.
                // Real deposit execution is also covered by the shared-client fixture tests.
                uint256 gasUpperBound = executionGas + 21_000 + 40 * code.length;
                assertLe(gasUpperBound, txs[i].gasLimit, txs[i].intent);
                emit log_named_uint(txs[i].intent, gasUpperBound);
                assertEq(deployed, expected);
                assertGt(deployed.code.length, 0);
            } else {
                vm.prank(txs[i].from);
                (bool success,) = txs[i].to.call{ gas: txs[i].gasLimit }(txs[i].data);
                assertTrue(success, txs[i].intent);
            }
        }
        assertEq(ICeloGasBridgeL2(payable(BRIDGE)).celoTokenL1(), L1_TOKEN);
        assertTrue(IL1Block(Predeploys.L1_BLOCK_ATTRIBUTES).isCustomGasToken());
        assertEq(IFeeVault(payable(Predeploys.SEQUENCER_FEE_WALLET)).totalProcessed(), 73);
        vm.deal(Predeploys.LIQUIDITY_CONTROLLER, 400);
        vm.prank(Predeploys.LIQUIDITY_CONTROLLER);
        (bool funded,) = Predeploys.NATIVE_ASSET_LIQUIDITY.call{ value: 400 }(abi.encodeWithSignature("deposit()"));
        assertTrue(funded);
        assertEq(Predeploys.NATIVE_ASSET_LIQUIDITY.balance, 400);
    }

    function test_upgrade_zeroReserveFunding_succeeds() external {
        _upgrade(manager);
        vm.prank(Predeploys.LIQUIDITY_CONTROLLER);
        (bool funded,) = Predeploys.NATIVE_ASSET_LIQUIDITY.call(abi.encodeWithSignature("deposit()"));
        assertTrue(funded);
        assertEq(Predeploys.NATIVE_ASSET_LIQUIDITY.balance, 0);
    }

    function test_upgrade_directCall_reverts() external {
        vm.expectRevert();
        ICgtUpgrade(manager).upgrade();
    }

    function test_upgrade_unauthorizedCaller_reverts() external {
        vm.expectRevert();
        IUpgradeProxyAdmin(Predeploys.PROXY_ADMIN).upgradePredeploys(manager);
    }

    function test_upgrade_repeatedActivation_reverts() external {
        _upgrade(manager);
        vm.expectRevert();
        _upgrade(manager);
    }

    function test_upgrade_failedInitializer_rollsBack_succeeds() external {
        address oldVault = address(uint160(uint256(vm.load(Predeploys.SEQUENCER_FEE_WALLET, IMPL_SLOT))));
        implementations.bridge = address(new RejectCgtInitializerFixture());
        address badManager = _manager(implementations, BRIDGE);
        vm.expectRevert();
        _upgrade(badManager);
        assertEq(vm.load(Predeploys.LIQUIDITY_CONTROLLER, IMPL_SLOT), bytes32(0));
        assertEq(vm.load(BRIDGE, IMPL_SLOT), bytes32(0));
        assertEq(vm.load(Predeploys.SEQUENCER_FEE_WALLET, IMPL_SLOT), bytes32(uint256(uint160(oldVault))));
        assertEq(Predeploys.SEQUENCER_FEE_WALLET.balance, 41);
        assertEq(feeToken.balanceOf(Predeploys.SEQUENCER_FEE_WALLET), 19);
    }
}
