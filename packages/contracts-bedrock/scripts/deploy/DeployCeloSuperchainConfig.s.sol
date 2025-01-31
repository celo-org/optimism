// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Script } from "forge-std/Script.sol";
import { stdToml } from "forge-std/StdToml.sol";

import { ISuperchainConfig } from "src/L1/interfaces/ISuperchainConfig.sol";
import { ICeloSuperchainConfig } from "src/L1/interfaces/ICeloSuperchainConfig.sol";
import { IProxyAdmin } from "src/universal/interfaces/IProxyAdmin.sol";
import { IProxy } from "src/universal/interfaces/IProxy.sol";

import { DeployUtils } from "scripts/libraries/DeployUtils.sol";
import { Solarray } from "scripts/libraries/Solarray.sol";
import { BaseDeployIO } from "scripts/deploy/BaseDeployIO.sol";

contract DeployCeloSuperchainConfigInput is BaseDeployIO {
    using stdToml for string;

    // Role inputs.
    address internal _superchainConfig;
    address internal _celoGuardian;

    // Other inputs.
    bool internal _paused;

    function set(bytes4 _sel, address _address) public {
        require(_address != address(0), "DeployCeloSuperchainConfigInput: cannot set zero address");
        if (_sel == this.superchainConfig.selector) _superchainConfig = _address;
        else if (_sel == this.celoGuardian.selector) _celoGuardian = _address;
        else revert("DeployCeloSuperchainConfigInput: unknown selector");
    }

    function set(bytes4 _sel, bool _value) public {
        if (_sel == this.paused.selector) _paused = _value;
        else revert("DeployCeloSuperchainConfigInput: unknown selector");
    }

    function superchainConfig() public view returns (address) {
        require(_superchainConfig != address(0), "DeployCeloSuperchainConfigInput: superchainConfig not set");
        return _superchainConfig;
    }

    function celoGuardian() public view returns (address) {
        require(_celoGuardian != address(0), "DeployCeloSuperchainConfigInput: celoGuardian not set");
        return _celoGuardian;
    }

    function paused() public view returns (bool) {
        return _paused;
    }
}

contract DeployCeloSuperchainConfigOutput is BaseDeployIO {
    ICeloSuperchainConfig internal _celoSuperchainConfigImpl;
    ICeloSuperchainConfig internal _celoSuperchainConfigProxy;
    IProxyAdmin internal _tmpProxyAdmin;

    function set(bytes4 _sel, address _address) public {
        require(_address != address(0), "DeployCeloSuperchainConfigOutput: cannot set zero address");
        if (_sel == this.celoSuperchainConfigImpl.selector) _celoSuperchainConfigImpl = ICeloSuperchainConfig(_address);
        else if (_sel == this.celoSuperchainConfigProxy.selector) _celoSuperchainConfigProxy = ICeloSuperchainConfig(_address);
        else if (_sel == this.tmpProxyAdmin.selector) _tmpProxyAdmin = IProxyAdmin(_address);
        else revert("DeployCeloSuperchainConfigOutput: unknown selector");
    }

    function checkOutput(DeployCeloSuperchainConfigInput _dsi) public {
        address[] memory addrs = Solarray.addresses(
            address(this.celoSuperchainConfigImpl()),
            address(this.celoSuperchainConfigProxy()),
            address(this.tmpProxyAdmin())
        );
        DeployUtils.assertValidContractAddresses(addrs);

        // To read the implementations we prank as the zero address due to the proxyCallIfNotAdmin modifier.
        vm.startPrank(address(0));
        address actualCeloSuperchainConfigImpl = IProxy(payable(address(_celoSuperchainConfigProxy))).implementation();
        vm.stopPrank();

        require(actualCeloSuperchainConfigImpl == address(_celoSuperchainConfigImpl), "100"); // nosemgrep:
            // sol-style-malformed-require

        assertValidDeploy(_dsi);

        // TODO(m-chrzan): consider if any checks necessary for tmpProxyAdmin??
    }

    function celoSuperchainConfigImpl() public view returns (ICeloSuperchainConfig) {
        DeployUtils.assertValidContractAddress(address(_celoSuperchainConfigImpl));
        return _celoSuperchainConfigImpl;
    }

    function celoSuperchainConfigProxy() public view returns (ICeloSuperchainConfig) {
        DeployUtils.assertValidContractAddress(address(_celoSuperchainConfigProxy));
        return _celoSuperchainConfigProxy;
    }

    function tmpProxyAdmin() public view returns (IProxyAdmin) {
        DeployUtils.assertValidContractAddress(address(_tmpProxyAdmin));
        return _tmpProxyAdmin;
    }

    // -------- Deployment Assertions --------
    function assertValidDeploy(DeployCeloSuperchainConfigInput _dsi) public {
        assertValidCeloSuperchainConfig(_dsi);
    }

    function assertValidCeloSuperchainConfig(DeployCeloSuperchainConfigInput _dsi) internal {
        // Proxy checks.
        ICeloSuperchainConfig celoSuperchainConfig = celoSuperchainConfigProxy();
        DeployUtils.assertInitialized({ _contractAddress: address(celoSuperchainConfig), _slot: 0, _offset: 0 });
        require(celoSuperchainConfig.guardian() == _dsi.celoGuardian(), "SUPCON-10");
        require(celoSuperchainConfig.paused() == _dsi.paused(), "SUPCON-20");
        require(celoSuperchainConfig.superchainConfig() == _dsi.superchainConfig(), "SUPCON-30");

        vm.startPrank(address(0));
        require(
            IProxy(payable(address(celoSuperchainConfig))).implementation() == address(celoSuperchainConfigImpl()), "SUPCON-30"
        );
        require(IProxy(payable(address(celoSuperchainConfig))).admin() == address(tmpProxyAdmin()), "SUPCON-40");
        vm.stopPrank();

        // Implementation checks
        celoSuperchainConfig = celoSuperchainConfigImpl();
        require(celoSuperchainConfig.guardian() == address(0), "SUPCON-50");
        require(celoSuperchainConfig.superchainConfig() == address(0), "SUPCON-50");
        require(celoSuperchainConfig.paused() == false, "SUPCON-60");
    }
}

// For all broadcasts in this script we explicitly specify the deployer as `msg.sender` because for
// testing we deploy this script from a test contract. If we provide no argument, the foundry
// default sender would be the broadcaster during test, but the broadcaster needs to be the deployer
// since they are set to the initial proxy admin owner.
contract DeployCeloSuperchainConfig is Script {
    function run(DeployCeloSuperchainConfigInput _dsi, DeployCeloSuperchainConfigOutput _dso) public {
        deployCeloSuperchainImplementationContracts(_dsi, _dso);
        deployTmpProxyAdmin(_dso);
        deployAndInitializeCeloSuperchainConfig(_dsi, _dso);

        _dso.checkOutput(_dsi);
    }

    // -------- Deployment Steps --------

    function deployCeloSuperchainImplementationContracts(DeployCeloSuperchainConfigInput, DeployCeloSuperchainConfigOutput _dso) public {
        // Deploy implementation contracts.
        vm.startBroadcast(msg.sender);
        ICeloSuperchainConfig celoSuperchainConfigImpl = ICeloSuperchainConfig(
            DeployUtils.create1({
                _name: "CeloSuperchainConfig",
                _args: DeployUtils.encodeConstructor(abi.encodeCall(ISuperchainConfig.__constructor__, ()))
            })
        );
        vm.stopBroadcast();

        vm.label(address(celoSuperchainConfigImpl), "CeloSuperchainConfigImpl");

        _dso.set(_dso.celoSuperchainConfigImpl.selector, address(celoSuperchainConfigImpl));
    }

    function deployTmpProxyAdmin(DeployCeloSuperchainConfigOutput _dso) public {
        vm.startBroadcast(msg.sender);
        IProxyAdmin proxyAdmin = IProxyAdmin(
            DeployUtils.create1({
                _name: "ProxyAdmin",
                _args:
                    DeployUtils.encodeConstructor(abi.encodeCall(IProxyAdmin.__constructor__, (msg.sender)))
            })
        );
        vm.stopBroadcast();

        vm.label(address(proxyAdmin), "TmpProxyAdmin");

        _dso.set(_dso.tmpProxyAdmin.selector, address(proxyAdmin));
    }

    function deployAndInitializeCeloSuperchainConfig(DeployCeloSuperchainConfigInput _dsi, DeployCeloSuperchainConfigOutput _dso) public {
        address guardian = _dsi.celoGuardian();
        bool paused = _dsi.paused();
        address superchainConfig = _dsi.superchainConfig();

        IProxyAdmin celoProxyAdmin = IProxyAdmin(_dso.tmpProxyAdmin());
        ICeloSuperchainConfig celoSuperchainConfigImpl = _dso.celoSuperchainConfigImpl();

        vm.startBroadcast(msg.sender);
        ICeloSuperchainConfig celoSuperchainConfigProxy = ICeloSuperchainConfig(
            DeployUtils.create1({
                _name: "Proxy",
                _args: DeployUtils.encodeConstructor(
                    abi.encodeCall(IProxy.__constructor__, (address(celoProxyAdmin)))
                )
            })
        );
        vm.stopBroadcast();
        // TODO(m-chrzan): should be able to set this to msg.sender now that
        // we're using tmpProxyAdmin?
        vm.startBroadcast(celoProxyAdmin.owner());
        celoProxyAdmin.upgradeAndCall(
            payable(address(celoSuperchainConfigProxy)),
            address(celoSuperchainConfigImpl),
            abi.encodeCall(ICeloSuperchainConfig.initialize, (guardian, paused, superchainConfig))
        );
        vm.stopBroadcast();

        vm.label(address(celoSuperchainConfigProxy), "CeloSuperchainConfigProxy");
        _dso.set(_dso.celoSuperchainConfigProxy.selector, address(celoSuperchainConfigProxy));
    }

    // -------- Utilities --------

    // This etches the IO contracts into memory so that we can use them in tests.
    // When interacting with the script programmatically (e.g. in a Solidity test), this must be called.
    function etchIOContracts() public returns (DeployCeloSuperchainConfigInput dsi_, DeployCeloSuperchainConfigOutput dso_) {
        (dsi_, dso_) = getIOContracts();
        vm.etch(address(dsi_), type(DeployCeloSuperchainConfigInput).runtimeCode);
        vm.etch(address(dso_), type(DeployCeloSuperchainConfigOutput).runtimeCode);
        vm.allowCheatcodes(address(dsi_));
        vm.allowCheatcodes(address(dso_));
    }

    // This returns the addresses of the IO contracts for this script.
    function getIOContracts() public view returns (DeployCeloSuperchainConfigInput dsi_, DeployCeloSuperchainConfigOutput dso_) {
        dsi_ = DeployCeloSuperchainConfigInput(DeployUtils.toIOAddress(msg.sender, "optimism.DeployCeloSuperchainConfigInput"));
        dso_ = DeployCeloSuperchainConfigOutput(DeployUtils.toIOAddress(msg.sender, "optimism.DeployCeloSuperchainConfigOutput"));
    }
}
