// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Test } from "forge-std/Test.sol";

import { Deploy } from "scripts/deploy/Deploy.s.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";

/// @notice Tests for the externalSuperchainConfig deploy-config field.
contract ExternalSuperchainConfig_Test is Test {
    /// @dev Mirrors the deterministic address used in test/setup/Setup.sol.
    Deploy internal constant deploy =
        Deploy(address(uint160(uint256(keccak256(abi.encode("optimism.deploy"))))));

    address anAddress;

    function setUp() public {
        anAddress = address(0xdeadbeef);

        // Etch Deploy so deploy.cfg() / deploy.setUp() are callable.
        DeployUtils.etchLabelAndAllowCheatcodes({ _etchTo: address(deploy), _cname: "Deploy" });
        deploy.setUp();
    }

    function test_setterPropagatesToCfg() public {
        deploy.cfg().setExternalSuperchainConfig(anAddress);
        assertEq(deploy.cfg().externalSuperchainConfig(), anAddress);
    }

    function test_defaultIsZero() public view {
        assertEq(deploy.cfg().externalSuperchainConfig(), address(0));
    }
}

/// @notice Negative-path placeholder; kept skipped to match v1.8.
contract ExternalSuperchainConfig_TestFail is Test {
    function test_address0_fails() public {
        vm.skip(true);
    }
}
