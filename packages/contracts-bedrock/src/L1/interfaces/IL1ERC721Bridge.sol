// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { IERC721Bridge } from "src/universal/interfaces/IERC721Bridge.sol";
import { ICrossDomainMessenger } from "src/universal/interfaces/ICrossDomainMessenger.sol";
import { ICeloSuperchainConfig } from "src/L1/interfaces/ICeloSuperchainConfig.sol";

interface IL1ERC721Bridge is IERC721Bridge {
    function bridgeERC721(
        address _localToken,
        address _remoteToken,
        uint256 _tokenId,
        uint32 _minGasLimit,
        bytes memory _extraData
    )
        external;
    function bridgeERC721To(
        address _localToken,
        address _remoteToken,
        address _to,
        uint256 _tokenId,
        uint32 _minGasLimit,
        bytes memory _extraData
    )
        external;
    function deposits(address, address, uint256) external view returns (bool);
    function finalizeBridgeERC721(
        address _localToken,
        address _remoteToken,
        address _from,
        address _to,
        uint256 _tokenId,
        bytes memory _extraData
    )
        external;
    function initialize(ICrossDomainMessenger _messenger, ICeloSuperchainConfig _superchainConfig) external;
    function paused() external view returns (bool);
    function superchainConfig() external view returns (ICeloSuperchainConfig);
    function version() external view returns (string memory);

    function __constructor__() external;
}
