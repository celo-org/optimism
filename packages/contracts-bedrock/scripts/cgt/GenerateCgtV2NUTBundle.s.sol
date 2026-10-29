// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Script } from "forge-std/Script.sol";
import { LegacyCeloVaultFixture, LegacyCeloL1BlockFixture } from "test/fixtures/cgtv2/LegacyPredeploys.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";
import { Constants } from "src/libraries/Constants.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { Types } from "src/libraries/Types.sol";
import { IL1BlockCGT } from "interfaces/L2/IL1BlockCGT.sol";
import { ILiquidityController } from "interfaces/L2/ILiquidityController.sol";
import { ICeloGasBridgeL2 } from "interfaces/celo/ICeloGasBridgeL2.sol";
import { IStandardBridge } from "interfaces/universal/IStandardBridge.sol";
import { IFeeVault } from "interfaces/L2/IFeeVault.sol";

/// @title GenerateCgtV2NUTBundle
/// @notice Generates static CGTv2 upgrade deposits. Reserve funding is a dynamic wrapper.
/// @dev Does not broadcast. runFixture() uses synthetic configuration, never a production allocation.
contract GenerateCgtV2NUTBundle is Script {
    struct NetworkConfig {
        address bridgeProxy;
        address l1Bridge;
        address l1Token;
        address owner;
        address vaultRecipient;
        uint256 vaultMinWithdrawalAmount;
        Types.WithdrawalNetwork vaultWithdrawalNetwork;
        string gasPayingTokenName;
        string gasPayingTokenSymbol;
    }

    struct NutTx {
        string intent;
        address from;
        address to;
        bytes data;
        uint64 gasLimit;
        bool isCreation;
    }

    /// @notice Pins the legacy getter values in network-specific upgrade calldata.
    /// @dev Verify legacy settings and zero initializer slots before freezing a production bundle.
    function transactions(NetworkConfig memory _config) public view returns (NutTx[] memory txs_) {
        _validateConfig(_config);
        txs_ = new NutTx[](15);
        txs_[0] = _deploy("L1BlockCGT deployment", "L1BlockCGT.sol:L1BlockCGT", "");
        txs_[1] = _deploy("L2ToL1MessagePasserCGT deployment", "L2ToL1MessagePasserCGT.sol:L2ToL1MessagePasserCGT", "");
        txs_[2] = _deploy("NativeAssetLiquidity deployment", "NativeAssetLiquidity.sol:NativeAssetLiquidity", "");
        txs_[3] = _deploy("LiquidityController deployment", "LiquidityController.sol:LiquidityController", "");
        txs_[4] = _deploy("CeloGasBridgeL2 deployment", "CeloGasBridgeL2.sol:CeloGasBridgeL2", "");
        txs_[5] = _deploy("CeloSequencerFeeVault deployment", "CeloSequencerFeeVault.sol:CeloSequencerFeeVault", "");
        txs_[6] = _upgrade("L1BlockCGT upgrade", Predeploys.L1_BLOCK_ATTRIBUTES, _created(txs_[0]), "");
        txs_[7] = _upgrade("L2ToL1MessagePasserCGT upgrade", Predeploys.L2_TO_L1_MESSAGE_PASSER, _created(txs_[1]), "");
        txs_[8] = _upgrade("NativeAssetLiquidity upgrade", Predeploys.NATIVE_ASSET_LIQUIDITY, _created(txs_[2]), "");
        txs_[9] = _upgrade(
            "LiquidityController upgrade and initialize",
            Predeploys.LIQUIDITY_CONTROLLER,
            _created(txs_[3]),
            abi.encodeCall(
                ILiquidityController.initialize,
                (Constants.DEPOSITOR_ACCOUNT, _config.gasPayingTokenName, _config.gasPayingTokenSymbol)
            )
        );
        txs_[10] = _upgrade(
            "CeloGasBridgeL2 upgrade and initialize",
            _config.bridgeProxy,
            _created(txs_[4]),
            abi.encodeCall(ICeloGasBridgeL2.initialize, (IStandardBridge(payable(_config.l1Bridge)), _config.l1Token))
        );
        txs_[11] = _upgrade(
            "CeloSequencerFeeVault upgrade and initialize",
            Predeploys.SEQUENCER_FEE_WALLET,
            _created(txs_[5]),
            abi.encodeCall(
                IFeeVault.initialize,
                (_config.vaultRecipient, _config.vaultMinWithdrawalAmount, _config.vaultWithdrawalNetwork)
            )
        );
        txs_[12] = _call(
            "CeloGasBridgeL2 minter authorization",
            Constants.DEPOSITOR_ACCOUNT,
            Predeploys.LIQUIDITY_CONTROLLER,
            abi.encodeCall(ILiquidityController.authorizeMinter, (_config.bridgeProxy))
        );
        txs_[13] = _call(
            "LiquidityController ownership transfer",
            Constants.DEPOSITOR_ACCOUNT,
            Predeploys.LIQUIDITY_CONTROLLER,
            abi.encodeCall(ILiquidityController.transferOwnership, (_config.owner))
        );
        txs_[14] = _call(
            "CGTv2 flag enable",
            Constants.DEPOSITOR_ACCOUNT,
            Predeploys.L1_BLOCK_ATTRIBUTES,
            abi.encodeCall(IL1BlockCGT.setCustomGasToken, ())
        );
    }

    function _call(
        string memory _intent,
        address _from,
        address _to,
        bytes memory _data
    )
        internal
        pure
        returns (NutTx memory)
    {
        return NutTx(_intent, _from, _to, _data, 500_000, false);
    }

    function _upgrade(
        string memory _intent,
        address _proxy,
        address _implementation,
        bytes memory _data
    )
        internal
        pure
        returns (NutTx memory)
    {
        bytes memory callData = _data.length == 0
            ? abi.encodeCall(IProxy.upgradeTo, (_implementation))
            : abi.encodeCall(IProxy.upgradeToAndCall, (_implementation, _data));
        // Existing Proxy permits protocol upgrade deposits from address(0).
        return _call(_intent, address(0), _proxy, callData);
    }

    function _validateConfig(NetworkConfig memory _config) internal pure {
        require(
            Predeploys.isPredeployNamespace(_config.bridgeProxy) && !Predeploys.notProxied(_config.bridgeProxy)
                && !Predeploys.isSupportedPredeploy(_config.bridgeProxy, type(uint256).max, true, true)
                && _config.l1Bridge != address(0) && _config.l1Token != address(0) && _config.owner != address(0),
            "GenerateCgtV2NUTBundle: invalid configuration"
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
            owner: address(0xaa),
            vaultRecipient: address(0xdd),
            vaultMinWithdrawalAmount: 23,
            vaultWithdrawalNetwork: Types.WithdrawalNetwork.L2,
            gasPayingTokenName: "Celo",
            gasPayingTokenSymbol: "CELO"
        });
        _writeBundle(config, "snapshots/upgrades/cgtv2-fixture-nut-bundle.json");
        _writePrestate(config);
    }

    /// @notice Low-level bundle writer. Use scripts/cgt/cgtv2_bundle.py to capture L2 getters.
    /// @dev This only writes an artifact. Go and Rust must explicitly register the reviewed bundle
    ///      before that network can activate CGTv2. Gas limits still require production calibration.
    /// @param _config Deployment addresses, explicit metadata and legacy vault settings at a fixed L2 block hash.
    /// @param _outputPath Network-specific JSON path allowed by Foundry's filesystem permissions.
    function runProduction(NetworkConfig memory _config, string memory _outputPath) public {
        _validateConfig(_config);
        require(bytes(_outputPath).length != 0, "GenerateCgtV2NUTBundle: empty output path");
        _writeBundle(_config, _outputPath);
    }

    /// @notice Accepts settings captured by the host tool without CLI tuple string escaping.
    function runProductionFromJson(string memory _configPath, string memory _outputPath) external {
        string memory json = vm.readFile(_configPath);
        NetworkConfig memory config = NetworkConfig({
            bridgeProxy: vm.parseJsonAddress(json, ".bridgeProxy"),
            l1Bridge: vm.parseJsonAddress(json, ".l1Bridge"),
            l1Token: vm.parseJsonAddress(json, ".l1Token"),
            owner: vm.parseJsonAddress(json, ".owner"),
            vaultRecipient: vm.parseJsonAddress(json, ".vaultRecipient"),
            vaultMinWithdrawalAmount: vm.parseJsonUint(json, ".vaultMinWithdrawalAmount"),
            vaultWithdrawalNetwork: Types.WithdrawalNetwork(vm.parseJsonUint(json, ".vaultWithdrawalNetwork")),
            gasPayingTokenName: vm.parseJsonString(json, ".gasPayingTokenName"),
            gasPayingTokenSymbol: vm.parseJsonString(json, ".gasPayingTokenSymbol")
        });
        runProduction(config, _outputPath);
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
