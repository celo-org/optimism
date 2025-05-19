// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { stdJson } from "forge-std/StdJson.sol";
import { console2 as console } from "forge-std/console2.sol";
import { Deployer } from "scripts/deploy/Deployer.sol";

import { GoldToken } from "src/celo/GoldToken.sol";
import { CeloPredeploys } from "src/celo/CeloPredeploys.sol";
import { CeloRegistry } from "src/celo/CeloRegistry.sol";
import { FeeHandler } from "src/celo/FeeHandler.sol";
import { MentoFeeHandlerSeller } from "src/celo/MentoFeeHandlerSeller.sol";
import { UniswapFeeHandlerSeller } from "src/celo/UniswapFeeHandlerSeller.sol";
import { SortedOracles } from "src/celo/stability/SortedOracles.sol";
import { FeeCurrencyDirectory } from "src/celo/FeeCurrencyDirectory.sol";
import { FeeCurrency } from "src/celo/testing/FeeCurrency.sol";
import { StableTokenV2 } from "src/celo/StableTokenV2.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";
import { EIP1967Helper } from "test/mocks/EIP1967Helper.sol";

contract L2GenesisCelo is Deployer {
    mapping(string => address) public deployedContractNamesToAddresses;
    string internal _celoL2Outfile;

    address internal deployerCelo = makeAddr("deployer");

    address[30] internal devAccountsCelo = [
        0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266, // 0
        0x70997970C51812dc3A010C7d01b50e0d17dc79C8, // 1
        0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC, // 2
        0x90F79bf6EB2c4f870365E785982E1f101E93b906, // 3
        0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65, // 4
        0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc, // 5
        0x976EA74026E726554dB657fA54763abd0C3a0aa9, // 6
        0x14dC79964da2C08b23698B3D3cc7Ca32193d9955, // 7
        0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f, // 8
        0xa0Ee7A142d267C1f36714E4a8F75612F20a79720, // 9
        0xBcd4042DE499D14e55001CcbB24a551F3b954096, // 10
        0x71bE63f3384f5fb98995898A86B02Fb2426c5788, // 11
        0xFABB0ac9d68B0B445fB7357272Ff202C5651694a, // 12
        0x1CBd3b2770909D4e10f157cABC84C7264073C9Ec, // 13
        0xdF3e18d64BC6A983f673Ab319CCaE4f1a57C7097, // 14
        0xcd3B766CCDd6AE721141F452C550Ca635964ce71, // 15
        0x2546BcD3c84621e976D8185a91A922aE77ECEc30, // 16
        0xbDA5747bFD65F08deb54cb465eB87D40e51B197E, // 17
        0xdD2FD4581271e230360230F9337D5c0430Bf44C0, // 18
        0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199, // 19
        0x09DB0a93B389bEF724429898f539AEB7ac2Dd55f, // 20
        0x02484cb50AAC86Eae85610D6f4Bf026f30f6627D, // 21
        0x08135Da0A343E492FA2d4282F2AE34c6c5CC1BbE, // 22
        0x5E661B79FE2D3F6cE70F5AAC07d8Cd9abb2743F1, // 23
        0x61097BA76cD906d2ba4FD106E757f7Eb455fc295, // 24
        0xDf37F81dAAD2b0327A0A50003740e1C935C70913, // 25
        0x553BC17A05702530097c3677091C5BB47a3a7931, // 26
        0x87BdCE72c06C21cd96219BD8521bDF1F42C78b5e, // 27
        0x40Fc963A729c542424cD800349a7E4Ecc4896624, // 28
        0x9DCCe783B6464611f38631e6C851bf441907c710 // 29
    ];

    function celoL2Outfile() internal view returns (string memory _env) {
        _env = vm.envOr(
            "L2_OUTFILE",
            string.concat(vm.projectRoot(), "/deployments/", vm.toString(block.chainid), "-l2-deploy.json")
        );
    }

    function celoSave(string memory _name, address _impl, address _proxy) public {
        if (deployedContractNamesToAddresses[_name] == address(0)) {
            deployedContractNamesToAddresses[_name] = _impl;

            _celoWrite(_name, _impl);
        }

        if (_proxy != address(0)) {
            string memory _proxyName = string.concat(_name, "Proxy");
            deployedContractNamesToAddresses[_proxyName] = _proxy;

            _celoWrite(_proxyName, _proxy);
        }
    }

    function _celoWrite(string memory _name, address _deployed) internal {
        console.log("Writing l2 deploy %s: %s", _name, _deployed);

        vm.writeJson({ json: stdJson.serialize("celo_l2_deploys", _name, _deployed), path: _celoL2Outfile });
    }

    ///@notice Sets all proxies and implementations for Celo contracts
    function setCeloPredeploys() internal {
        console.log("Deploying Celo contracts");

        setCeloRegistry();
        setCeloGoldToken();
        setCeloFeeHandler();
        setCeloMentoFeeHandlerSeller();
        setCeloUniswapFeeHandlerSeller();
        // setCeloSortedOracles();
        // setCeloAddressSortedLinkedListWithMedian();
        setCeloFeeCurrency();
        setFeeCurrencyDirectory();

        address[] memory initialBalanceAddresses = new address[](1);
        initialBalanceAddresses[0] = devAccountsCelo[0];

        uint256[] memory initialBalances = new uint256[](1);
        initialBalances[0] = 100_000 ether;
        //deploycUSD(initialBalanceAddresses, initialBalances, 2);
    }

    /// @notice Sets up a proxy for the given impl address
    function _setupProxy(address addr, address impl) internal returns (address) {
        bytes memory code = vm.getDeployedCode("Proxy.sol:Proxy");
        vm.etch(addr, code);
        EIP1967Helper.setAdmin(addr, Predeploys.PROXY_ADMIN);

        console.log("Setting proxy %s with implementation: %s", addr, impl);
        EIP1967Helper.setImplementation(addr, impl);

        return addr;
    }

    function setCeloRegistry() internal {
        CeloRegistry kontract = new CeloRegistry({ test: false });

        address precompile = CeloPredeploys.CELO_REGISTRY;
        string memory cname = CeloPredeploys.getName(precompile);
        console.log("Deploying %s implementation at: %s", cname, address(kontract));

        vm.resetNonce(address(kontract));
        _setupProxy(precompile, address(kontract));
    }

    function setCeloGoldToken() internal {
        GoldToken kontract = new GoldToken({ test: false });

        address precompile = CeloPredeploys.GOLD_TOKEN;
        string memory cname = CeloPredeploys.getName(precompile);
        console.log("Deploying %s implementation at: %s", cname, address(kontract));

        vm.resetNonce(address(kontract));
        _setupProxy(precompile, address(kontract));
    }

    function setCeloFeeHandler() internal {
        FeeHandler kontract = new FeeHandler({ test: false });

        address precompile = CeloPredeploys.FEE_HANDLER;
        string memory cname = CeloPredeploys.getName(precompile);
        console.log("Deploying %s implementation at: %s", cname, address(kontract));

        vm.resetNonce(address(kontract));
        _setupProxy(precompile, address(kontract));
    }

    function setCeloMentoFeeHandlerSeller() internal {
        MentoFeeHandlerSeller kontract = new MentoFeeHandlerSeller({ test: false });

        address precompile = CeloPredeploys.MENTO_FEE_HANDLER_SELLER;
        string memory cname = CeloPredeploys.getName(precompile);
        console.log("Deploying %s implementation at: %s", cname, address(kontract));

        vm.resetNonce(address(kontract));
        _setupProxy(precompile, address(kontract));
    }

    function setCeloUniswapFeeHandlerSeller() internal {
        UniswapFeeHandlerSeller kontract = new UniswapFeeHandlerSeller({ test: false });

        address precompile = CeloPredeploys.UNISWAP_FEE_HANDLER_SELLER;
        string memory cname = CeloPredeploys.getName(precompile);
        console.log("Deploying %s implementation at: %s", cname, address(kontract));

        vm.resetNonce(address(kontract));
        _setupProxy(precompile, address(kontract));
    }

    function setCeloSortedOracles() internal {
        SortedOracles kontract = new SortedOracles({ test: false });

        address precompile = CeloPredeploys.SORTED_ORACLES;
        string memory cname = CeloPredeploys.getName(precompile);
        console.log("Deploying %s implementation at: %s", cname, address(kontract));

        vm.resetNonce(address(kontract));
        _setupProxy(precompile, address(kontract));
    }

    function setFeeCurrencyDirectory() internal {
        FeeCurrencyDirectory feeCurrencyDirectory = new FeeCurrencyDirectory({ test: false });

        address precompile = CeloPredeploys.FEE_CURRENCY_DIRECTORY;
        string memory cname = CeloPredeploys.getName(precompile);
        console.log("Deploying %s implementation at: %s", cname, address(feeCurrencyDirectory));

        vm.resetNonce(address(feeCurrencyDirectory));
        _setupProxy(precompile, address(feeCurrencyDirectory));

        vm.startPrank(devAccountsCelo[0]);
        FeeCurrencyDirectory(precompile).initialize();
        vm.stopPrank();
    }

    // function setCeloAddressSortedLinkedListWithMedian() internal {
    //     AddressSortedLinkedListWithMedian kontract = new AddressSortedLinkedListWithMedian({
    //     });
    //     address precompile = CeloPredeploys.ADDRESS_SORTED_LINKED_LIST_WITH_MEDIAN;
    //     string memory cname = CeloPredeploys.getName(precompile);
    //     console.log("Deploying %s implementation at: %s", cname, address(kontract ));
    //     vm.resetNonce(address(kontract ));
    //     _setupProxy(precompile, address(kontract));
    // }

    function setCeloFeeCurrency() internal {
        FeeCurrency kontract = new FeeCurrency({ name_: "Test", symbol_: "TST" });
        address precompile = CeloPredeploys.FEE_CURRENCY;
        string memory cname = CeloPredeploys.getName(precompile);
        console.log("Deploying %s implementation at: %s", cname, address(kontract));
        vm.resetNonce(address(kontract));
        _setupProxy(precompile, address(kontract));
    }

    function deploycUSD(
        address[] memory initialBalanceAddresses,
        uint256[] memory initialBalanceValues,
        uint256 celoPrice
    )
        public
    {
        StableTokenV2 kontract = new StableTokenV2({ disable: false });
        address cusdProxyAddress = CeloPredeploys.cUSD;
        string memory cname = CeloPredeploys.getName(cusdProxyAddress);
        console.log("Deploying %s implementation at: %s", cname, address(kontract));
        vm.resetNonce(address(kontract));

        _setupProxy(cusdProxyAddress, address(kontract));

        kontract.initialize("Celo Dollar", "cUSD", initialBalanceAddresses, initialBalanceValues);

        SortedOracles sortedOracles = SortedOracles(CeloPredeploys.SORTED_ORACLES);

        console.log("beofre add oracle");

        vm.startPrank(sortedOracles.owner());
        sortedOracles.addOracle(cusdProxyAddress, deployerCelo);
        vm.stopPrank();
        vm.startPrank(deployerCelo);

        if (celoPrice != 0) {
            sortedOracles.report(cusdProxyAddress, celoPrice * 1e24, address(0), address(0)); // TODO use fixidity
        }

        /*
    Arbitrary intrinsic gas number take from existing `FeeCurrencyDirectory.t.sol` tests
        Source:
        https://github.com/celo-org/celo-monorepo/blob/2cec07d43328cf4216c62491a35eacc4960fffb6/packages/protocol/test-sol/common/FeeCurrencyDirectory.t.sol#L27
        */
        uint256 mockIntrinsicGas = 21000;

        FeeCurrencyDirectory feeCurrencyDirectory = FeeCurrencyDirectory(CeloPredeploys.FEE_CURRENCY_DIRECTORY);
        vm.startPrank(feeCurrencyDirectory.owner());
        feeCurrencyDirectory.setCurrencyConfig(cusdProxyAddress, address(sortedOracles), mockIntrinsicGas);
        vm.stopPrank();
        vm.startPrank(deployerCelo);
    }
}
