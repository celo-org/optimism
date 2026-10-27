// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Script } from "forge-std/Script.sol";
import { LegacyCeloVaultFixture, LegacyCeloL1BlockFixture } from "test/fixtures/cgtv2/LegacyPredeploys.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";
import { Constants } from "src/libraries/Constants.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { IL2ProxyAdmin } from "interfaces/L2/IL2ProxyAdmin.sol";
import { CeloCGTV2ContractsManager } from "src/celo/CeloCGTV2ContractsManager.sol";

/// @title GenerateCgtV2NUTBundle
/// @notice Generates static CGTv2 upgrade deposits. Reserve funding is a dynamic wrapper.
/// @dev Does not broadcast. runFixture() uses synthetic configuration, never a production allocation.
contract GenerateCgtV2NUTBundle is Script {
    struct NetworkConfig {
        address bridgeProxy;
        address l1Bridge;
        address l1Token;
        address owner;
    }

    struct NutTx {
        string intent;
        address from;
        address to;
        bytes data;
        uint64 gasLimit;
        bool isCreation;
    }

    function transactions(NetworkConfig memory _config) public view returns (NutTx[] memory txs) {
        txs = new NutTx[](12);
        txs[0] = _deploy("StorageSetter deployment", "StorageSetter.sol:StorageSetter", "");
        txs[1] = _deploy("L2ProxyAdmin deployment", "L2ProxyAdmin.sol:L2ProxyAdmin", "");
        txs[2] = NutTx(
            "L2ProxyAdmin upgrade",
            address(0),
            Predeploys.PROXY_ADMIN,
            abi.encodeCall(IProxy.upgradeTo, (_created(txs[1]))),
            100_000,
            false
        );
        txs[3] = _deploy("L1BlockCGT deployment", "L1BlockCGT.sol:L1BlockCGT", "");
        txs[4] = _deploy("L2ToL1MessagePasserCGT deployment", "L2ToL1MessagePasserCGT.sol:L2ToL1MessagePasserCGT", "");
        txs[5] = _deploy("NativeAssetLiquidity deployment", "NativeAssetLiquidity.sol:NativeAssetLiquidity", "");
        txs[6] = _deploy("LiquidityController deployment", "LiquidityController.sol:LiquidityController", "");
        txs[7] = _deploy("CeloGasBridgeL2 deployment", "CeloGasBridgeL2.sol:CeloGasBridgeL2", "");
        txs[8] = _deploy("CeloSequencerFeeVault deployment", "CeloSequencerFeeVault.sol:CeloSequencerFeeVault", "");
        CeloCGTV2ContractsManager.Implementations memory impls = CeloCGTV2ContractsManager.Implementations({
            storageSetter: _created(txs[0]),
            l1Block: _created(txs[3]),
            messagePasser: _created(txs[4]),
            nativeLiquidity: _created(txs[5]),
            controller: _created(txs[6]),
            bridge: _created(txs[7]),
            vault: _created(txs[8])
        });
        txs[9] = _deploy(
            "CeloCGTV2ContractsManager deployment",
            "CeloCGTV2ContractsManager.sol:CeloCGTV2ContractsManager",
            abi.encode(impls, _config)
        );
        txs[10] = NutTx(
            "CGTv2 predeploy upgrade",
            Constants.DEPOSITOR_ACCOUNT,
            Predeploys.PROXY_ADMIN,
            abi.encodeCall(IL2ProxyAdmin.upgradePredeploys, (_created(txs[9]))),
            2_000_000,
            false
        );
        txs[11] = NutTx(
            "CGTv2 flag enable",
            Constants.DEPOSITOR_ACCOUNT,
            Predeploys.L1_BLOCK_ATTRIBUTES,
            abi.encodeWithSignature("setCustomGasToken()"),
            100_000,
            false
        );
    }

    function _deploy(
        string memory _intent,
        string memory _artifact,
        bytes memory _args
    )
        internal
        view
        returns (NutTx memory)
    {
        address from = address(uint160(uint256(keccak256(abi.encodePacked("Celo CGTv2: ", _intent)))));
        return NutTx(_intent, from, address(0), bytes.concat(vm.getCode(_artifact), _args), 6_000_000, true);
    }

    function _created(NutTx memory _tx) internal pure returns (address) {
        return address(uint160(uint256(keccak256(abi.encodePacked(hex"d694", _tx.from, hex"80")))));
    }

    /// @notice Writes a reproducible synthetic-network bundle. All mint/value fields are implicit zero.
    function runFixture() external {
        NetworkConfig memory config = NetworkConfig({
            bridgeProxy: 0x42000000000000000000000000000000000007fE,
            l1Bridge: address(0xcc),
            l1Token: address(0xbb),
            owner: address(0xaa)
        });
        _writeBundle(config, "snapshots/upgrades/cgtv2-fixture-nut-bundle.json");
        _writePrestate(config);
    }

    /// @notice Generates a bundle from finalized network addresses without broadcasting transactions.
    /// @dev This only writes an artifact. Go and Rust must explicitly register the reviewed bundle
    ///      before that network can activate CGTv2. Gas limits still require production calibration.
    /// @param _config Reserved L2 bridge proxy, L1 bridge, L1 CELO token and controller owner.
    /// @param _outputPath Network-specific JSON path allowed by Foundry's filesystem permissions.
    function runProduction(NetworkConfig memory _config, string memory _outputPath) external {
        require(
            Predeploys.isUpgradeable(_config.bridgeProxy)
                && !Predeploys.isSupportedPredeploy(_config.bridgeProxy, type(uint256).max, true, true)
                && _config.l1Bridge != address(0) && _config.l1Token != address(0) && _config.owner != address(0),
            "GenerateCgtV2NUTBundle: invalid configuration"
        );
        require(bytes(_outputPath).length != 0, "GenerateCgtV2NUTBundle: empty output path");
        _writeBundle(_config, _outputPath);
    }

    function _writeBundle(NetworkConfig memory _config, string memory _outputPath) internal {
        NutTx[] memory txs = transactions(_config);
        string memory json = '{"metadata":{"version":"1.0.0"},"transactions":[';
        for (uint256 i; i < txs.length; i++) {
            string memory to = txs[i].isCreation ? "null" : string.concat('"', vm.toString(txs[i].to), '"');
            json = string.concat(
                json,
                i == 0 ? "" : ",",
                '{"intent":"',
                txs[i].intent,
                '","from":"',
                vm.toString(txs[i].from),
                '","to":',
                to,
                ',"data":"',
                vm.toString(txs[i].data),
                '","gasLimit":',
                vm.toString(txs[i].gasLimit),
                "}"
            );
        }
        vm.writeFile(_outputPath, string.concat(json, "]}"));
    }

    function _account(
        bytes memory _code,
        uint256 _balance,
        string memory _storage
    )
        internal
        pure
        returns (string memory)
    {
        return string.concat(
            '{"code":"', vm.toString(_code), '","balance":"', vm.toString(_balance), '","storage":{', _storage, "}}"
        );
    }

    function _slot(bytes32 _key, bytes32 _value) internal pure returns (string memory) {
        return string.concat('"', vm.toString(_key), '":"', vm.toString(_value), '"');
    }

    function _writePrestate(NetworkConfig memory _config) internal {
        bytes32 adminSlot = bytes32(uint256(keccak256("eip1967.proxy.admin")) - 1);
        bytes32 implSlot = bytes32(uint256(keccak256("eip1967.proxy.implementation")) - 1);
        bytes memory proxyCode = vm.getDeployedCode("Proxy.sol:Proxy");
        bytes memory legacyAdmin = vm.getDeployedCode("ProxyAdmin.sol:ProxyAdmin");
        bytes memory legacyBlock = address(new LegacyCeloL1BlockFixture()).code;
        bytes memory legacyVault = address(new LegacyCeloVaultFixture(address(0xdd))).code;
        address[7] memory proxies = [
            Predeploys.PROXY_ADMIN,
            Predeploys.L1_BLOCK_ATTRIBUTES,
            Predeploys.L2_TO_L1_MESSAGE_PASSER,
            Predeploys.NATIVE_ASSET_LIQUIDITY,
            Predeploys.LIQUIDITY_CONTROLLER,
            Predeploys.SEQUENCER_FEE_WALLET,
            _config.bridgeProxy
        ];
        string memory json = "{";
        for (uint256 i; i < proxies.length; i++) {
            string memory storage_ = _slot(adminSlot, bytes32(uint256(uint160(Predeploys.PROXY_ADMIN))));
            uint256 balance;
            if (i == 0) {
                storage_ = string.concat(
                    storage_,
                    ",",
                    _slot(implSlot, bytes32(uint256(0x1001))),
                    ",",
                    _slot(bytes32(0), bytes32(uint256(uint160(_config.owner))))
                );
            }
            if (i == 1) storage_ = string.concat(storage_, ",", _slot(implSlot, bytes32(uint256(0x1002))));
            if (i == 5) {
                storage_ = string.concat(
                    storage_,
                    ",",
                    _slot(implSlot, bytes32(uint256(0x1003))),
                    ",",
                    _slot(bytes32(0), bytes32(uint256(73)))
                );
                balance = 41;
            }
            json = string.concat(
                json, i == 0 ? "" : ",", '"', vm.toString(proxies[i]), '":', _account(proxyCode, balance, storage_)
            );
        }
        json = string.concat(
            json,
            ',"0x0000000000000000000000000000000000001001":',
            _account(legacyAdmin, 0, ""),
            ',"0x0000000000000000000000000000000000001002":',
            _account(legacyBlock, 0, ""),
            ',"0x0000000000000000000000000000000000001003":',
            _account(legacyVault, 0, ""),
            "}"
        );
        vm.writeFile("snapshots/upgrades/cgtv2-fixture-prestate.json", json);
    }
}
