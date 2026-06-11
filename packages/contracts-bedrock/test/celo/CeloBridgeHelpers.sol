// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Interfaces
import { ILiquidityController } from "interfaces/L2/ILiquidityController.sol";
import { IOptimismMintableERC20 } from "interfaces/universal/IOptimismMintableERC20.sol";

/// @notice Shared mock ERC20 with mint/approve/transfer for Celo bridge tests.
contract MockERC20 {
    string public name = "Celo";
    string public symbol = "CELO";
    uint8 public decimals = 18;
    uint256 public totalSupply;
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;

    function mint(address _to, uint256 _amount) external {
        balanceOf[_to] += _amount;
        totalSupply += _amount;
    }

    function approve(address _spender, uint256 _amount) external returns (bool) {
        allowance[msg.sender][_spender] = _amount;
        return true;
    }

    function transfer(address _to, uint256 _amount) external returns (bool) {
        balanceOf[msg.sender] -= _amount;
        balanceOf[_to] += _amount;
        return true;
    }

    function transferFrom(address _from, address _to, uint256 _amount) external returns (bool) {
        if (allowance[_from][msg.sender] != type(uint256).max) {
            allowance[_from][msg.sender] -= _amount;
        }
        balanceOf[_from] -= _amount;
        balanceOf[_to] += _amount;
        return true;
    }
}

/// @notice Shared mock cross-domain messenger for Celo bridge tests.
contract MockCrossDomainMessenger {
    address public xDomainMessageSender;
    address public lastTarget;
    bytes public lastMessage;
    uint32 public lastMinGasLimit;
    uint256 public lastValue;

    function setXDomainMessageSender(address _xDomainMessageSender) external {
        xDomainMessageSender = _xDomainMessageSender;
    }

    function sendMessage(address _target, bytes calldata _message, uint32 _minGasLimit) external payable {
        lastTarget = _target;
        lastMessage = _message;
        lastMinGasLimit = _minGasLimit;
        lastValue = msg.value;
    }
}

/// @notice Mock OptimismMintableERC20 recording the last mint, used to drive the inherited bridge path.
contract MockOptimismMintableERC20 {
    address public remoteToken;
    address public lastMintTo;
    uint256 public lastMintAmount;

    constructor(address _remoteToken) {
        remoteToken = _remoteToken;
    }

    function mint(address _to, uint256 _amount) external {
        lastMintTo = _to;
        lastMintAmount = _amount;
    }

    function supportsInterface(bytes4 _interfaceId) external pure returns (bool) {
        return _interfaceId == type(IOptimismMintableERC20).interfaceId || _interfaceId == 0x01ffc9a7;
    }
}

/// @notice Mock SystemConfig exposing pause, CGT-mode, and the OptimismPortal getter (L1 bridge tests).
contract MockSystemConfig {
    bool public paused;
    bool public isCustomGasToken;
    address public optimismPortal;

    function setPaused(bool _paused) external {
        paused = _paused;
    }

    function setIsCustomGasToken(bool _isCustomGasToken) external {
        isCustomGasToken = _isCustomGasToken;
    }

    function setOptimismPortal(address _optimismPortal) external {
        optimismPortal = _optimismPortal;
    }
}

/// @notice Mock ProxyAdmin exposing a fixed owner.
contract MockProxyAdmin {
    address public owner;

    constructor(address _owner) {
        owner = _owner;
    }
}

/// @notice Mock L1Block exposing the CGT-mode flag (L2 bridge tests).
contract MockL1Block {
    bool public isCustomGasToken;

    function setIsCustomGasToken(bool _isCustomGasToken) external {
        isCustomGasToken = _isCustomGasToken;
    }
}

/// @notice Mock LiquidityController recording burn/mint and optionally reverting unauthorized.
contract MockLiquidityController {
    address public lastBurnCaller;
    uint256 public lastBurnAmount;
    address public lastMintCaller;
    address public lastMintTo;
    uint256 public lastMintAmount;
    bool public burnUnauthorized;
    bool public mintUnauthorized;

    function setBurnUnauthorized(bool _burnUnauthorized) external {
        burnUnauthorized = _burnUnauthorized;
    }

    function setMintUnauthorized(bool _mintUnauthorized) external {
        mintUnauthorized = _mintUnauthorized;
    }

    function burn() external payable {
        if (burnUnauthorized) revert ILiquidityController.LiquidityController_Unauthorized();

        lastBurnCaller = msg.sender;
        lastBurnAmount = msg.value;
    }

    function mint(address _to, uint256 _amount) external {
        if (mintUnauthorized) revert ILiquidityController.LiquidityController_Unauthorized();

        lastMintCaller = msg.sender;
        lastMintTo = _to;
        lastMintAmount = _amount;
    }
}

/// @notice Mock SystemConfig exposing only the OptimismPortal getter (PortalMigrator tests).
contract MockPortalSystemConfig {
    address public optimismPortal;

    function setOptimismPortal(address _optimismPortal) external {
        optimismPortal = _optimismPortal;
    }
}

/// @notice Mock CeloGasBridgeL1 capturing the migration-time seedEscrow call (PortalMigrator tests).
contract MockSeedBridge {
    MockPortalSystemConfig public systemConfig;
    uint256 public seededAmount;
    bool public escrowSeeded;

    error UnauthorizedSeeder();

    constructor() {
        systemConfig = new MockPortalSystemConfig();
    }

    function setOptimismPortal(address _optimismPortal) external {
        systemConfig.setOptimismPortal(_optimismPortal);
    }

    function seedEscrow(uint256 _amount) external {
        if (msg.sender != systemConfig.optimismPortal()) revert UnauthorizedSeeder();
        escrowSeeded = true;
        seededAmount = _amount;
    }
}
