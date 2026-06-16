// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IFeeCurrencyAdapter {
    function getAdaptedToken() external view returns (address);
}
