// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { console } from "forge-std/console.sol";

import { DeployConfig } from "scripts/deploy/DeployConfig.s.sol";
import { SuperchainConfig } from "src/L1/SuperchainConfig.sol";

// Testing utilities
import { CommonTest } from "test/setup/CommonTest.sol";

contract ExternalSuperchainConfig_Test is CommonTest {
    address anAddress;

    function setUp() public override {
        anAddress = address(new SuperchainConfig());
        super.enableExternalSuperchainConfig(anAddress);
        super.setUp();
    }

    function test_setsSuperchainConfigCorrectly() public {
        assertEq(address(superchainConfig), anAddress);
    }
}

contract ExternalSuperchainConfig_TestFail is CommonTest {
    function setUp() public override { }

    function test_address0_fails() public {
        // With the below `expectRevert` uncommented, the test fails with "next call did not revert
        // as expected", even though, when commented, instead the test fails with the expected error
        // message.
        vm.skip(true);
        super.enableExternalSuperchainConfig(address(0));
        //vm.expectRevert("Need to provide the external SuperchainConfig address!");
        super.setUp();
    }
}
