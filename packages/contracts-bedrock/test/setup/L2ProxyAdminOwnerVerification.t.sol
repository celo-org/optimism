// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { DeployConfig } from "scripts/deploy/DeployConfig.s.sol";

// Testing utilities
import { CommonTest } from "test/setup/CommonTest.sol";

contract L2ProxyAdminOwnerVerification_Test is CommonTest {
    address constant anOwner = 0xCAfEcAfeCAfECaFeCaFecaFecaFECafECafeCaFe;
    address constant anotherOwner = 0xBEeFbeefbEefbeEFbeEfbEEfBEeFbeEfBeEfBeef;
    address constant aliasedOwner = 0xDc0FCafeCAFecAfeCafecAfecaFecAFeCAfEdc0F;

    DeployConfig cfg;

    function setUp() public override {
        super.setUp();
        cfg = deploy.cfg();
    }

    function test_noCheck_succeeds() public {
        cfg.setProxyAdminOwnerSettings(anOwner, anotherOwner, "no-check");
        cfg.verifyProxyAdminOwners();
    }

    function test_aliased_succeeds() public {
        cfg.setProxyAdminOwnerSettings(anOwner, aliasedOwner, "aliased");
        cfg.verifyProxyAdminOwners();
    }

    function test_equal_succeeds() public {
        cfg.setProxyAdminOwnerSettings(anOwner, anOwner, "equal");
        cfg.verifyProxyAdminOwners();
    }
}

contract L2ProxyAdminOwnerVerification_TestFail is CommonTest {
    address constant anOwner = 0xCAfEcAfeCAfECaFeCaFecaFecaFECafECafeCaFe;
    address constant anotherOwner = 0xBEeFbeefbEefbeEFbeEfbEEfBEeFbeEfBeEfBeef;
    address constant aliasedOwner = 0xDc0FCafeCAFecAfeCafecAfecaFecAFeCAfEdc0F;
    address constant offByOneOwner = 0xcAFeCAfEcAFECaFecaFECAFeCAfEcAFecAFECAff;

    DeployConfig cfg;

    function setUp() public override {
        super.setUp();
        cfg = deploy.cfg();
    }

    function test_aliased_whenSameAddress_fails() public {
        cfg.setProxyAdminOwnerSettings(anOwner, anOwner, "aliased");
        vm.expectRevert("Expected aliased address");
        cfg.verifyProxyAdminOwners();
    }

    function test_aliased_whenIncorrectAlias_fails() public {
        cfg.setProxyAdminOwnerSettings(anOwner, offByOneOwner, "aliased");
        vm.expectRevert("Expected aliased address");
        cfg.verifyProxyAdminOwners();
    }

    function test_equal_whenAliased_fails() public {
        cfg.setProxyAdminOwnerSettings(anOwner, aliasedOwner, "equal");
        vm.expectRevert("Expected finalSystemOwner and proxyAdminOwner to be equal");
        cfg.verifyProxyAdminOwners();
    }

    function test_equal_whenOffBy1_fails() public {
        cfg.setProxyAdminOwnerSettings(anOwner, offByOneOwner, "equal");
        vm.expectRevert("Expected finalSystemOwner and proxyAdminOwner to be equal");
        cfg.verifyProxyAdminOwners();
    }

    function test_badConfig_fails() public {
        cfg.setProxyAdminOwnerSettings(anOwner, anOwner, "bad-option");
        vm.expectRevert("Incorrect value for l2ProxyAdminOwnerVerification");
        cfg.verifyProxyAdminOwners();
    }
}
