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

    function test_aliased_succeeds() public {
        cfg.setProxyAdminOwnerSettings(anOwner, aliasedOwner, true);
        cfg.verifyProxyAdminOwners();
    }

    function test_equal_succeeds() public {
        cfg.setProxyAdminOwnerSettings(anOwner, anOwner, false);
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
        cfg.setProxyAdminOwnerSettings(anOwner, anOwner, true);
        vm.expectRevert("Expected proxyAdminOwner to be aliased finalSystemOwner");
        cfg.verifyProxyAdminOwners();
    }

    function test_aliased_whenIncorrectAlias_fails() public {
        cfg.setProxyAdminOwnerSettings(anOwner, offByOneOwner, true);
        vm.expectRevert("Expected proxyAdminOwner to be aliased finalSystemOwner");
        cfg.verifyProxyAdminOwners();
    }

    function test_equal_whenAliased_fails() public {
        cfg.setProxyAdminOwnerSettings(anOwner, aliasedOwner, false);
        vm.expectRevert("Expected finalSystemOwner and proxyAdminOwner to be equal");
        cfg.verifyProxyAdminOwners();
    }

    function test_equal_whenOffBy1_fails() public {
        cfg.setProxyAdminOwnerSettings(anOwner, offByOneOwner, false);
        vm.expectRevert("Expected finalSystemOwner and proxyAdminOwner to be equal");
        cfg.verifyProxyAdminOwners();
    }
}
