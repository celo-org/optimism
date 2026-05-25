// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Test } from "forge-std/Test.sol";
import { ICeloSuperchainConfig } from "interfaces/L1/ICeloSuperchainConfig.sol";
import { ISuperchainConfig } from "interfaces/L1/ISuperchainConfig.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { Deploy } from "scripts/deploy/Deploy.s.sol";
import { Artifacts } from "scripts/Artifacts.s.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";

/// @notice Forks an L1 (Mainnet or Sepolia) and runs Deploy.runCelo against the upstream OP
///         SuperchainConfig and ProtocolVersions on that L1.
///         Run with: ETH_RPC_URL=<mainnet|sepolia> forge test --match-contract RunCelo_ForkTest --fork-url $ETH_RPC_URL
abstract contract RunCelo_ForkTest_Base is Test {
    Deploy internal deploy;

    function _expectedChainId() internal pure virtual returns (uint256);
    function _externalSuperchainConfig() internal pure virtual returns (address);
    function _protocolVersionsProxy() internal pure virtual returns (address payable);

    function setUp() public {
        // Skip unless running on the matching L1 fork.
        if (block.chainid != _expectedChainId()) {
            vm.skip(true);
            return;
        }

        // Etch the Deploy script at a deterministic address (matches the test/setup/Setup.sol pattern).
        deploy = Deploy(address(uint160(uint256(keccak256(abi.encode("optimism.deploy"))))));
        DeployUtils.etchLabelAndAllowCheatcodes({ _etchTo: address(deploy), _cname: "Deploy" });
        deploy.setUp();

        deploy.cfg().setExternalSuperchainConfig(_externalSuperchainConfig());
    }

    function test_runCelo_wiresSystemConfigToCelo() public {
        address externalSuperchainConfig = _externalSuperchainConfig();
        address payable protocolVersionsProxy = _protocolVersionsProxy();
        assertGt(externalSuperchainConfig.code.length, 0, "SuperchainConfig has no code");
        assertGt(protocolVersionsProxy.code.length, 0, "ProtocolVersions has no code");

        deploy.runCelo(protocolVersionsProxy);

        Artifacts artifacts = deploy.artifacts();
        address celoSuperchainConfig = artifacts.mustGetAddress("CeloSuperchainConfigProxy");
        address systemConfig = artifacts.mustGetAddress("SystemConfigProxy");

        // SystemConfig points at the CeloSuperchainConfig (not the external SuperchainConfig).
        assertEq(
            address(ISystemConfig(systemConfig).superchainConfig()),
            celoSuperchainConfig,
            "SystemConfig should point at CeloSuperchainConfig"
        );

        // CeloSuperchainConfig wraps the external SuperchainConfig supplied via cfg.
        assertEq(
            ICeloSuperchainConfig(celoSuperchainConfig).superchainConfig(),
            externalSuperchainConfig,
            "CeloSuperchainConfig should wrap the external SuperchainConfig"
        );

        // CeloSuperchainConfig's guardian came from cfg, not from the external SuperchainConfig.
        assertEq(
            ICeloSuperchainConfig(celoSuperchainConfig).guardian(),
            deploy.cfg().superchainConfigGuardian(),
            "CeloSuperchainConfig guardian mismatch"
        );

        // Paused state propagates from the external SuperchainConfig into CeloSuperchainConfig's view.
        assertEq(
            ICeloSuperchainConfig(celoSuperchainConfig).paused(),
            ISuperchainConfig(externalSuperchainConfig).paused(),
            "CeloSuperchainConfig.paused should mirror external SuperchainConfig"
        );
    }
}

/// @notice Mainnet fork variant. External SuperchainConfig: 0x95703e0982140D16f8ebA6d158FccEde42f04a4C.
contract RunCelo_ForkTest_Mainnet is RunCelo_ForkTest_Base {
    function _expectedChainId() internal pure override returns (uint256) {
        return 1;
    }

    function _externalSuperchainConfig() internal pure override returns (address) {
        return 0x95703e0982140D16f8ebA6d158FccEde42f04a4C;
    }

    function _protocolVersionsProxy() internal pure override returns (address payable) {
        return payable(0x8062AbC286f5e7D9428a0Ccb9AbD71e50d93b935);
    }
}

/// @notice Sepolia fork variant. External SuperchainConfig: 0xC2Be75506d5724086DEB7245bd260Cc9753911Be.
contract RunCelo_ForkTest_Sepolia is RunCelo_ForkTest_Base {
    function _expectedChainId() internal pure override returns (uint256) {
        return 11155111;
    }

    function _externalSuperchainConfig() internal pure override returns (address) {
        return 0xC2Be75506d5724086DEB7245bd260Cc9753911Be;
    }

    function _protocolVersionsProxy() internal pure override returns (address payable) {
        return payable(0x79ADD5713B383DAa0a138d3C4780C7A1804a8090);
    }
}
