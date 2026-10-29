// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Test } from "forge-std/Test.sol";
import {
    LegacyCeloVaultFixture,
    LegacyCeloL1BlockFixture,
    CgtFeeTokenFixture,
    RejectCgtInitializerFixture
} from "test/fixtures/cgtv2/LegacyPredeploys.sol";
import { GenerateCgtV2NUTBundle } from "scripts/cgt/GenerateCgtV2NUTBundle.s.sol";
import { Proxy } from "src/universal/Proxy.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";
import { Types } from "src/libraries/Types.sol";
import { IL1Block } from "interfaces/L2/IL1Block.sol";
import { INativeAssetLiquidity } from "interfaces/L2/INativeAssetLiquidity.sol";
import { ILiquidityController } from "interfaces/L2/ILiquidityController.sol";
import { IFeeVault } from "interfaces/L2/IFeeVault.sol";
import { ICeloGasBridgeL2 } from "interfaces/celo/ICeloGasBridgeL2.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";

contract CgtV2Upgrade_Upgrade_Test is Test {
    address internal constant BRIDGE = 0x42000000000000000000000000000000000007fE;
    bytes32 internal constant ADMIN_SLOT = bytes32(uint256(keccak256("eip1967.proxy.admin")) - 1);
    bytes32 internal constant IMPL_SLOT = bytes32(uint256(keccak256("eip1967.proxy.implementation")) - 1);
    bytes32 internal constant INITIALIZABLE_SLOT = 0xf0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a00;
    address internal constant OWNER = address(0xBEEF);
    address internal constant RECIPIENT = address(0xCAFE);
    address internal constant L1_BRIDGE = address(0x1234);
    address internal constant L1_TOKEN = address(0x5678);

    CgtFeeTokenFixture internal feeToken;
    GenerateCgtV2NUTBundle internal generator;

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
        bytes memory code = address(new Proxy(Predeploys.PROXY_ADMIN)).code;
        for (uint256 i; i < proxies.length; i++) {
            vm.etch(proxies[i], code);
            vm.store(proxies[i], ADMIN_SLOT, bytes32(uint256(uint160(Predeploys.PROXY_ADMIN))));
        }
        _install(Predeploys.L1_BLOCK_ATTRIBUTES, address(new LegacyCeloL1BlockFixture()));
        _install(Predeploys.SEQUENCER_FEE_WALLET, address(new LegacyCeloVaultFixture(RECIPIENT)));
        _install(Predeploys.PROXY_ADMIN, deployCode("ProxyAdmin.sol:ProxyAdmin", abi.encode(OWNER)));
        vm.store(Predeploys.PROXY_ADMIN, bytes32(0), bytes32(uint256(uint160(OWNER))));
        vm.store(Predeploys.SEQUENCER_FEE_WALLET, bytes32(0), bytes32(uint256(73)));
        vm.deal(Predeploys.SEQUENCER_FEE_WALLET, 41);
        feeToken = new CgtFeeTokenFixture();
        feeToken.mint(Predeploys.SEQUENCER_FEE_WALLET, 19);
        generator = GenerateCgtV2NUTBundle(deployCode("GenerateCgtV2NUTBundle.s.sol:GenerateCgtV2NUTBundle"));
    }

    function _config() internal pure returns (GenerateCgtV2NUTBundle.NetworkConfig memory) {
        return GenerateCgtV2NUTBundle.NetworkConfig({
            bridgeProxy: BRIDGE,
            l1Bridge: L1_BRIDGE,
            l1Token: L1_TOKEN,
            owner: OWNER,
            vaultRecipient: RECIPIENT,
            vaultMinWithdrawalAmount: 23,
            vaultWithdrawalNetwork: Types.WithdrawalNetwork.L2,
            gasPayingTokenName: "Celo",
            gasPayingTokenSymbol: "CELO"
        });
    }

    function _install(address proxy, address implementation) internal {
        vm.prank(Predeploys.PROXY_ADMIN);
        Proxy(payable(proxy)).upgradeTo(implementation);
    }

    function _run() internal {
        GenerateCgtV2NUTBundle.NutTx[] memory txs = generator.transactions(_config());
        assertEq(txs.length, 15);
        for (uint256 i; i < txs.length; i++) {
            assertTrue(txs[i].to != Predeploys.PROXY_ADMIN, "must not upgrade ProxyAdmin");
            if (txs[i].isCreation) {
                address expected = vm.computeCreateAddress(txs[i].from, 0);
                vm.broadcast(txs[i].from);
                bytes memory code = txs[i].data;
                address deployed;
                uint256 beforeCreate = gasleft();
                assembly {
                    deployed := create(0, add(code, 32), mload(code))
                }
                uint256 bound = beforeCreate - gasleft() + 21_000 + 40 * code.length;
                assertLe(bound, txs[i].gasLimit, txs[i].intent);
                emit log_named_uint(txs[i].intent, bound);
                assertEq(deployed, expected);
                assertGt(deployed.code.length, 0);
            } else {
                vm.prank(txs[i].from);
                uint256 beforeCall = gasleft();
                (bool success,) = txs[i].to.call{ gas: txs[i].gasLimit }(txs[i].data);
                uint256 bound = beforeCall - gasleft() + 21_000 + 40 * txs[i].data.length;
                assertLe(bound, txs[i].gasLimit, txs[i].intent);
                assertTrue(success, txs[i].intent);
            }
        }
    }

    function test_upgrade_generatedNUTGroup_succeeds() external {
        bytes32 oldImplementation = vm.load(Predeploys.PROXY_ADMIN, IMPL_SLOT);
        bytes32 oldCodeHash = Predeploys.PROXY_ADMIN.codehash;
        _run();
        assertEq(vm.load(Predeploys.PROXY_ADMIN, IMPL_SLOT), oldImplementation);
        assertEq(Predeploys.PROXY_ADMIN.codehash, oldCodeHash);
        assertEq(vm.load(Predeploys.PROXY_ADMIN, ADMIN_SLOT), bytes32(uint256(uint160(Predeploys.PROXY_ADMIN))));
        assertEq(IProxyAdmin(Predeploys.PROXY_ADMIN).owner(), OWNER);
        assertTrue(IL1Block(Predeploys.L1_BLOCK_ATTRIBUTES).isCustomGasToken());
        ICeloGasBridgeL2 bridge = ICeloGasBridgeL2(payable(BRIDGE));
        assertEq(address(bridge.otherBridge()), L1_BRIDGE);
        assertEq(bridge.celoTokenL1(), L1_TOKEN);
        ILiquidityController controller = ILiquidityController(Predeploys.LIQUIDITY_CONTROLLER);
        assertEq(controller.owner(), OWNER);
        assertTrue(controller.minters(BRIDGE));
        assertEq(controller.gasPayingTokenName(), "Celo");
        assertEq(controller.gasPayingTokenSymbol(), "CELO");
        IFeeVault vault = IFeeVault(payable(Predeploys.SEQUENCER_FEE_WALLET));
        assertEq(vault.recipient(), RECIPIENT);
        assertEq(vault.minWithdrawalAmount(), 23);
        assertEq(uint256(vault.withdrawalNetwork()), uint256(Types.WithdrawalNetwork.L2));
        assertEq(vault.totalProcessed(), 73);
        assertEq(Predeploys.SEQUENCER_FEE_WALLET.balance, 41);
        assertEq(feeToken.balanceOf(Predeploys.SEQUENCER_FEE_WALLET), 19);
        vm.expectRevert(bytes4(keccak256("InvalidInitialization()")));
        vault.initialize(RECIPIENT, 23, Types.WithdrawalNetwork.L2);
        vm.expectRevert("Ownable: caller is not the owner");
        controller.authorizeMinter(address(1));
        vm.deal(Predeploys.LIQUIDITY_CONTROLLER, 400);
        vm.prank(Predeploys.LIQUIDITY_CONTROLLER);
        (bool funded,) =
            Predeploys.NATIVE_ASSET_LIQUIDITY.call{ value: 400 }(abi.encodeCall(INativeAssetLiquidity.deposit, ()));
        assertTrue(funded);
        assertEq(Predeploys.NATIVE_ASSET_LIQUIDITY.balance, 400);
    }

    function runUpgrade() external {
        _run();
    }

    function test_upgrade_initializedPredeploy_reverts() external {
        address[3] memory proxies = [Predeploys.LIQUIDITY_CONTROLLER, BRIDGE, Predeploys.SEQUENCER_FEE_WALLET];
        bytes32[3] memory slots = [bytes32(0), bytes32(0), INITIALIZABLE_SLOT];
        GenerateCgtV2NUTBundle.NutTx[] memory txs = generator.transactions(_config());
        for (uint256 i; i < proxies.length; i++) {
            vm.store(proxies[i], slots[i], bytes32(uint256(1)));
            vm.expectRevert(bytes(txs[9 + i].intent));
            this.runUpgrade();
            vm.store(proxies[i], slots[i], bytes32(0));
        }
    }

    function test_upgrade_unauthorizedProxyCaller_reverts() external {
        vm.expectRevert(bytes(""));
        Proxy(payable(Predeploys.L1_BLOCK_ATTRIBUTES)).upgradeTo(address(1));
    }

    function test_upgrade_failedInitializerRestoresPreviousImplementation_succeeds() external {
        bytes32 oldImplementation = vm.load(Predeploys.SEQUENCER_FEE_WALLET, IMPL_SLOT);
        address bad = address(new RejectCgtInitializerFixture());
        vm.expectRevert("Proxy: delegatecall to new implementation contract failed");
        vm.prank(address(0));
        Proxy(payable(Predeploys.SEQUENCER_FEE_WALLET)).upgradeToAndCall(
            bad, abi.encodeCall(RejectCgtInitializerFixture.initialize, ())
        );
        assertEq(vm.load(Predeploys.SEQUENCER_FEE_WALLET, IMPL_SLOT), oldImplementation);
        assertEq(Predeploys.SEQUENCER_FEE_WALLET.balance, 41);
    }

    function test_upgrade_zeroReserveFunding_succeeds() external {
        _run();
        vm.prank(Predeploys.LIQUIDITY_CONTROLLER);
        (bool funded,) = Predeploys.NATIVE_ASSET_LIQUIDITY.call(abi.encodeCall(INativeAssetLiquidity.deposit, ()));
        assertTrue(funded);
        assertEq(Predeploys.NATIVE_ASSET_LIQUIDITY.balance, 0);
    }

    function test_runProduction_exportsConfiguredNUTBundle_succeeds() external {
        string memory path = ".testdata/cgtv2-production-entrypoint.json";
        vm.createDir(".testdata", true);
        generator.runProduction(_config(), path);
        string memory json = vm.readFile(path);
        assertEq(vm.parseJsonString(json, ".metadata.version"), "1.0.0");
        assertEq(vm.parseJsonString(json, ".transactions[14].intent"), "CGTv2 flag enable");
        assertFalse(vm.keyExistsJson(json, ".transactions[15]"));
        assertEq(vm.parseJsonAddress(json, ".transactions[9].to"), Predeploys.LIQUIDITY_CONTROLLER);
        assertEq(vm.parseJsonBytes(json, ".transactions[9].data"), generator.transactions(_config())[9].data);
        assertEq(vm.parseJsonBytes(json, ".transactions[10].data"), generator.transactions(_config())[10].data);
        assertEq(vm.parseJsonBytes(json, ".transactions[11].data"), generator.transactions(_config())[11].data);
        vm.removeFile(path);
    }

    function test_runProduction_invalidConfig_reverts() external {
        for (uint256 i; i < 5; i++) {
            GenerateCgtV2NUTBundle.NetworkConfig memory config = _config();
            if (i == 0) config.bridgeProxy = address(0);
            if (i == 1) config.l1Bridge = address(0);
            if (i == 2) config.l1Token = address(0);
            if (i == 3) config.owner = address(0);
            if (i == 4) config.bridgeProxy = Predeploys.L1_BLOCK_ATTRIBUTES;
            vm.expectRevert("GenerateCgtV2NUTBundle: invalid configuration");
            generator.runProduction(config, ".testdata/cgtv2-invalid-entrypoint.json");
        }
    }

    function test_runProduction_emptyOutputPath_reverts() external {
        vm.expectRevert("GenerateCgtV2NUTBundle: empty output path");
        generator.runProduction(_config(), "");
    }
}
