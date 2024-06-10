// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {ERC20} from '@openzeppelin/contracts/token/ERC20/ERC20.sol';

string constant NAME = 'Celo native asset';
string constant SYMBOL = 'CELO';

contract CeloToken is ERC20Upgradeable {
  function initialize(
    uint256 totalSupply,
    address portalProxyAddress
  ) ERC20(NAME, SYMBOL) {
    _mint(portalProxyAddress, totalSupply);
  }
}
