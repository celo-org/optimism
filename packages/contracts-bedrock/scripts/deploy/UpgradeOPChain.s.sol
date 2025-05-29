// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";
import { OPContractsManager, ISystemConfig, IProxyAdmin, Claim } from "src/L1/OPContractsManager.sol";
import { BaseDeployIO } from "scripts/deploy/BaseDeployIO.sol";

contract UpgradeOPChainInput is BaseDeployIO {
    address internal _prank;
    OPContractsManager internal _opcm;
    bytes _opChainConfigs;

    // Setter for OPContractsManager type
    function set(bytes4 _sel, address _value) public {
        require(address(_value) != address(0), "UpgradeOPCMInput: cannot set zero address");

        if (_sel == this.prank.selector) _prank = _value;
        else if (_sel == this.opcm.selector) _opcm = OPContractsManager(_value);
        else revert("UpgradeOPCMInput: unknown selector");
    }

    function set(bytes4 _sel, OPContractsManager.OpChainConfig[] memory _value) public {
        require(_value.length > 0, "UpgradeOPCMInput: cannot set empty array");

        if (_sel == this.opChainConfigs.selector) _opChainConfigs = abi.encode(_value);
        else revert("UpgradeOPCMInput: unknown selector");
    }

    function prank() public view returns (address) {
        require(address(_prank) != address(0), "UpgradeOPCMInput: prank not set");
        return _prank;
    }

    function opcm() public view returns (OPContractsManager) {
        require(address(_opcm) != address(0), "UpgradeOPCMInput: not set");
        return _opcm;
    }

    function opChainConfigs() public view returns (bytes memory) {
        require(_opChainConfigs.length > 0, "UpgradeOPCMInput: not set");
        return _opChainConfigs;
    }
}

contract CeloUpgradeOPChain is Script {
    function convert(bytes32 _claim) internal pure returns (Claim claim_) {
        assembly {
            claim_ := _claim
        }
    }
}

contract CeloUpgradeAlfajores is CeloUpgradeOPChain {
    function run() external {
        // setup
        console.log("Setup started!");
        UpgradeOPChainInput uoci = new UpgradeOPChainInput();
        OPContractsManager.OpChainConfig[] memory config = new OPContractsManager.OpChainConfig[](1);
        config[0] = OPContractsManager.OpChainConfig(
            ISystemConfig(0x499b0C1F4BDC76d61b1D13b03384eac65FAF50c7),
            IProxyAdmin(0x4630583d066520aF0E3fda0de2C628EEd2888683),
            convert(bytes32(hex"03b357b30095022ecbb44ef00d1de19df39cf69ee92a60683a6be2c6f8fe6a3e"))
        );
        uoci.set(UpgradeOPChainInput.prank.selector, address(0xf05f102e890E713DC9dc0a5e13A8879D5296ee48));
        uoci.set(UpgradeOPChainInput.opcm.selector, address(0x83cccf6d865EA06d103dcc9CF10D56B17Dd4e74E));
        uoci.set(UpgradeOPChainInput.opChainConfigs.selector, config);

        // execution
        console.log("Execution!");
        UpgradeOPChain upgrade = new UpgradeOPChain();
        upgrade.run(uoci);
    }
}

contract CeloUpgradeBaklava is CeloUpgradeOPChain {
    function run() external {
        // setup
        console.log("Setup started!");
        UpgradeOPChainInput uoci = new UpgradeOPChainInput();
        OPContractsManager.OpChainConfig[] memory config = new OPContractsManager.OpChainConfig[](1);
        config[0] = OPContractsManager.OpChainConfig(
            ISystemConfig(0x3ee24bF404e4a5D27A437d910F56E1eD999B1De8),
            IProxyAdmin(0xBF101Bd81fb69aB00ab261465454dF1a171726Bf),
            convert(bytes32(hex"03b357b30095022ecbb44ef00d1de19df39cf69ee92a60683a6be2c6f8fe6a3e"))
        );
        uoci.set(UpgradeOPChainInput.prank.selector, address(0xd542f3328ff2516443FE4db1c89E427F67169D94));
        // uoci.set(UpgradeOPChainInput.opcm.selector, address(0xAF66Cb99Fb7f632394269B7d746CD4c37D736678));
        uoci.set(UpgradeOPChainInput.opcm.selector, address(0x216e0Bd7F565A0bbb1d09F04bE39650Da2697339));
        uoci.set(UpgradeOPChainInput.opChainConfigs.selector, config);

        // execution
        console.log("Execution!");
        UpgradeOPChain upgrade = new UpgradeOPChain();
        upgrade.run(uoci);
    }
}

contract UpgradeOPChain is Script {
    function run(UpgradeOPChainInput _uoci) external {
        OPContractsManager opcm = _uoci.opcm();
        OPContractsManager.OpChainConfig[] memory opChainConfigs =
            abi.decode(_uoci.opChainConfigs(), (OPContractsManager.OpChainConfig[]));

        // Etch DummyCaller contract. This contract is used to mimic the contract that is used
        // as the source of the delegatecall to the OPCM. In practice this will be the governance
        // 2/2 or similar.
        address prank = _uoci.prank();
        bytes memory code = vm.getDeployedCode("UpgradeOPChain.s.sol:DummyCaller");
        vm.etch(prank, code);
        vm.store(prank, bytes32(0), bytes32(uint256(uint160(address(opcm)))));
        vm.label(prank, "DummyCaller");

        // Call into the DummyCaller. This will perform the delegatecall under the hood and
        // return the result.
        vm.broadcast(msg.sender);
        (bool success,) = DummyCaller(prank).upgrade(opChainConfigs);
        require(success, "UpgradeChain: upgrade failed");
    }
}

contract DummyCaller {
    address internal _opcmAddr;

    function upgrade(OPContractsManager.OpChainConfig[] memory _opChainConfigs) external returns (bool, bytes memory) {
        bytes memory data = abi.encodeCall(DummyCaller.upgrade, _opChainConfigs);
        (bool success, bytes memory result) = _opcmAddr.delegatecall(data);
        return (success, result);
    }
}
