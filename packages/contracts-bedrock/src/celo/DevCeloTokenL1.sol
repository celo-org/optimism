// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {ERC20Upgradeable} from "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";

contract DevCeloTokenL1 is ERC20Upgradeable {
    string constant NAME = "Celo";
    string constant SYMBOL = "CELO";
    uint256 constant L2_INITIAL_MARKET_CAP = 1000000000e18;
    uint256 constant L1_INITIAL_AMOUNT_PER_ACOUNT = 1000e18;

    constructor() {
        _disableInitializers();
    }

    function initialize(address _portalProxyAddress) external initializer {
        // m/44'/60'/0'/0/i
        address[21] memory devAccountsCelo = [
            0x8943545177806ED17B9F23F0a21ee5948eCaa776,
            0xE25583099BA105D9ec0A67f5Ae86D90e50036425,
            0x3e95dFbBaF6B348396E6674C7871546dCC568e56,
            0x5918b2e647464d4743601a865753e64C8059Dc4F,
            0x589A698b7b7dA0Bec545177D3963A2741105C7C9,
            0x4d1CB4eB7969f8806E2CaAc0cbbB71f88C8ec413,
            0xF5504cE2BcC52614F121aff9b93b2001d92715CA,
            0xF61E98E7D47aB884C244E39E031978E33162ff4b,
            0xf1424826861ffbbD25405F5145B5E50d0F1bFc90,
            0xfDCe42116f541fc8f7b0776e2B30832bD5621C85,
            0xD9211042f35968820A3407ac3d80C725f8F75c14,
            0xD8F3183DEF51A987222D845be228e0Bbb932C222,
            0x614561D2d143621E126e87831AEF287678B442b8,
            0xafF0CA253b97e54440965855cec0A8a2E2399896,
            0xf93Ee4Cf8c6c40b329b0c0626F28333c132CF241,
            0x802dCbE1B1A97554B4F50DB5119E37E8e7336417,
            0xAe95d8DA9244C37CaC0a3e16BA966a8e852Bb6D6,
            0x2c57d1CFC6d5f8E4182a56b4cf75421472eBAEa4,
            0x741bFE4802cE1C4b5b00F9Df2F5f179A1C89171A,
            0xc3913d4D8bAb4914328651C2EAE817C8b78E1f4c,
            0x65D08a056c17Ae13370565B04cF77D2AfA1cB9FA
        ];

        __ERC20_init(NAME, SYMBOL);
        _mint(_portalProxyAddress, L2_INITIAL_MARKET_CAP);
        for (uint256 i = 0; i < devAccountsCelo.length; i++) {
            _mint(devAccountsCelo[i], L1_INITIAL_AMOUNT_PER_ACOUNT);
        }
    }
}
