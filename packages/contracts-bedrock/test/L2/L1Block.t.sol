// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing
import { CommonTest } from "test/setup/CommonTest.sol";

// Libraries
import { Encoding } from "src/libraries/Encoding.sol";
import { Constants } from "src/libraries/Constants.sol";
import "src/libraries/L1BlockErrors.sol";

/// @title L1Block_ TestInit
/// @notice Reusable test initialization for `L1Block` tests.
contract L1Block_TestInit is CommonTest {
    address depositor;

    event GasPayingTokenSet(address indexed token, uint8 indexed decimals, bytes32 name, bytes32 symbol);

    /// @notice Sets up the test suite.
    function setUp() public virtual override {
        super.setUp();
        depositor = l1Block.DEPOSITOR_ACCOUNT();
    }

    /// @notice Asserts that legacy high 128-bit ranges in key storage slots remain zeroed.
    function assertEmptyLegacySlotRanges() internal view {
        // 128 high bits mask for 32-byte word
        bytes32 mask128 = hex"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000000000000000000000000000";
        // Check scalars and sequenceNumber slot (slot 3)
        bytes32 scalarsSlot = vm.load(address(l1Block), bytes32(uint256(3)));
        assertEq(0, scalarsSlot & mask128);
        // Check number and timestamp slot (slot 0)
        bytes32 numberTimestampSlot = vm.load(address(l1Block), bytes32(uint256(0)));
        assertEq(0, numberTimestampSlot & mask128);
    }
}

/// @title L1Block_GasPayingToken_Test
/// @notice Tests the `gasPayingToken` function of the `L1Block` contract.
contract L1Block_GasPayingToken_Test is L1Block_TestInit {
    /// @notice Tests that the `gasPayingToken` function returns the correct token address and
    ///         decimals.
    function test_gasPayingToken_succeeds() external view {
        (address token, uint8 decimals) = l1Block.gasPayingToken();
        assertEq(token, Constants.ETHER);
        assertEq(uint256(decimals), uint256(18));
    }
}

/// @title L1Block_GasPayingTokenName_Test
/// @notice Tests the `gasPayingTokenName` function of the `L1Block` contract.
contract L1Block_GasPayingTokenName_Test is L1Block_TestInit {
    /// @notice Tests that the `gasPayingTokenName` function returns the correct token name.
    function test_gasPayingTokenName_succeeds() external view {
        assertEq("Ether", l1Block.gasPayingTokenName());
    }
}

/// @title L1Block_GasPayingTokenSymbol_Test
/// @notice Tests the `gasPayingTokenSymbol` function of the `L1Block` contract.
contract L1Block_GasPayingTokenSymbol_Test is L1Block_TestInit {
    /// @notice Tests that the `gasPayingTokenSymbol` function returns the correct token symbol.
    function test_gasPayingTokenSymbol_succeeds() external view {
        assertEq("ETH", l1Block.gasPayingTokenSymbol());
    }
}

/// @title L1Block_IsCustomGasToken_Test
/// @notice Tests the `isCustomGasToken` function of the `L1Block` contract.
contract L1Block_IsCustomGasToken_Test is L1Block_TestInit {
    /// @notice Tests that the `isCustomGasToken` function returns false when no custom gas token
    ///         is used.
    function test_isCustomGasToken_succeeds() external view {
        assertFalse(l1Block.isCustomGasToken());
    }
}

/// @title L1Block_SetL1BlockValues_Test
/// @notice Tests the `setL1BlockValues` function of the `L1Block` contract.
contract L1Block_SetL1BlockValues_Test is L1Block_TestInit {
    /// @notice Tests that `setL1BlockValues` updates the values correctly.
    function testFuzz_setL1BlockValues_succeeds(
        uint64 n,
        uint64 t,
        uint256 b,
        bytes32 h,
        uint64 s,
        bytes32 bt,
        uint256 fo,
        uint256 fs
    )
        external
    {
        vm.prank(depositor);
        l1Block.setL1BlockValues(n, t, b, h, s, bt, fo, fs);
        assertEq(l1Block.number(), n);
        assertEq(l1Block.timestamp(), t);
        assertEq(l1Block.basefee(), b);
        assertEq(l1Block.hash(), h);
        assertEq(l1Block.sequenceNumber(), s);
        assertEq(l1Block.batcherHash(), bt);
        assertEq(l1Block.l1FeeOverhead(), fo);
        assertEq(l1Block.l1FeeScalar(), fs);
    }

    /// @notice Tests that `setL1BlockValues` succeeds with max values
    function test_setL1BlockValuesMax_succeeds() external {
        vm.prank(depositor);
        l1Block.setL1BlockValues({
            _number: type(uint64).max,
            _timestamp: type(uint64).max,
            _basefee: type(uint256).max,
            _hash: keccak256(abi.encode(1)),
            _sequenceNumber: type(uint64).max,
            _batcherHash: bytes32(type(uint256).max),
            _l1FeeOverhead: type(uint256).max,
            _l1FeeScalar: type(uint256).max
        });
    }

    /// @notice Tests that `setL1BlockValues` reverts if sender address is not the depositor
    function test_setL1BlockValues_notDepositor_reverts() external {
        vm.expectRevert("L1Block: only the depositor account can set L1 block values");
        l1Block.setL1BlockValues({
            _number: type(uint64).max,
            _timestamp: type(uint64).max,
            _basefee: type(uint256).max,
            _hash: keccak256(abi.encode(1)),
            _sequenceNumber: type(uint64).max,
            _batcherHash: bytes32(type(uint256).max),
            _l1FeeOverhead: type(uint256).max,
            _l1FeeScalar: type(uint256).max
        });
    }
}

/// @title L1Block_SetL1BlockValuesEcotone_Test
/// @notice Tests the `setL1BlockValuesEcotone` function of the `L1Block` contract.
contract L1Block_SetL1BlockValuesEcotone_Test is L1Block_TestInit {
    /// @notice Tests that setL1BlockValuesEcotone updates the values appropriately.
    function testFuzz_setL1BlockValuesEcotone_succeeds(
        uint32 baseFeeScalar,
        uint32 blobBaseFeeScalar,
        uint64 sequenceNumber,
        uint64 timestamp,
        uint64 number,
        uint256 baseFee,
        uint256 blobBaseFee,
        bytes32 hash,
        bytes32 batcherHash
    )
        external
    {
        bytes memory functionCallDataPacked = Encoding.encodeSetL1BlockValuesEcotone(
            baseFeeScalar, blobBaseFeeScalar, sequenceNumber, timestamp, number, baseFee, blobBaseFee, hash, batcherHash
        );

        vm.prank(depositor);
        (bool success,) = address(l1Block).call(functionCallDataPacked);
        assertTrue(success, "Function call failed");

        assertEq(l1Block.baseFeeScalar(), baseFeeScalar);
        assertEq(l1Block.blobBaseFeeScalar(), blobBaseFeeScalar);
        assertEq(l1Block.sequenceNumber(), sequenceNumber);
        assertEq(l1Block.timestamp(), timestamp);
        assertEq(l1Block.number(), number);
        assertEq(l1Block.basefee(), baseFee);
        assertEq(l1Block.blobBaseFee(), blobBaseFee);
        assertEq(l1Block.hash(), hash);
        assertEq(l1Block.batcherHash(), batcherHash);

        assertEmptyLegacySlotRanges();
    }

    /// @notice Tests that `setL1BlockValuesEcotone` succeeds with max values
    function test_setL1BlockValuesEcotone_isDepositorMax_succeeds() external {
        bytes memory functionCallDataPacked = Encoding.encodeSetL1BlockValuesEcotone(
            type(uint32).max,
            type(uint32).max,
            type(uint64).max,
            type(uint64).max,
            type(uint64).max,
            type(uint256).max,
            type(uint256).max,
            bytes32(type(uint256).max),
            bytes32(type(uint256).max)
        );

        vm.prank(depositor);
        (bool success,) = address(l1Block).call(functionCallDataPacked);
        assertTrue(success, "function call failed");
    }

    /// @notice Tests that `setL1BlockValuesEcotone` reverts if sender address is not the depositor
    function test_setL1BlockValuesEcotone_notDepositor_reverts() external {
        bytes memory functionCallDataPacked = Encoding.encodeSetL1BlockValuesEcotone(
            type(uint32).max,
            type(uint32).max,
            type(uint64).max,
            type(uint64).max,
            type(uint64).max,
            type(uint256).max,
            type(uint256).max,
            bytes32(type(uint256).max),
            bytes32(type(uint256).max)
        );

        (bool success, bytes memory data) = address(l1Block).call(functionCallDataPacked);
        assertTrue(!success, "function call should have failed");
        // make sure return value is the expected function selector for "NotDepositor()"
        bytes memory expReturn = hex"3cc50b45";
        assertEq(data, expReturn);
    }
}

/// @title L1Block_SetL1BlockValuesIsthmus_Test
/// @notice Tests the `setL1BlockValuesIsthmus` function of the `L1Block` contract.
contract L1Block_SetL1BlockValuesIsthmus_Test is L1Block_TestInit {
    /// @notice Tests that setL1BlockValuesIsthmus updates the values appropriately.
    function testFuzz_setL1BlockValuesIsthmus_succeeds(
        uint32 baseFeeScalar,
        uint32 blobBaseFeeScalar,
        uint64 sequenceNumber,
        uint64 timestamp,
        uint64 number,
        uint256 baseFee,
        uint256 blobBaseFee,
        bytes32 hash,
        bytes32 batcherHash,
        uint32 operatorFeeScalar,
        uint64 operatorFeeConstant
    )
        external
    {
        bytes memory functionCallDataPacked = Encoding.encodeSetL1BlockValuesIsthmus(
            baseFeeScalar,
            blobBaseFeeScalar,
            sequenceNumber,
            timestamp,
            number,
            baseFee,
            blobBaseFee,
            hash,
            batcherHash,
            operatorFeeScalar,
            operatorFeeConstant
        );

        vm.prank(depositor);
        (bool success,) = address(l1Block).call(functionCallDataPacked);
        assertTrue(success, "Function call failed");

        assertEq(l1Block.baseFeeScalar(), baseFeeScalar);
        assertEq(l1Block.blobBaseFeeScalar(), blobBaseFeeScalar);
        assertEq(l1Block.sequenceNumber(), sequenceNumber);
        assertEq(l1Block.timestamp(), timestamp);
        assertEq(l1Block.number(), number);
        assertEq(l1Block.basefee(), baseFee);
        assertEq(l1Block.blobBaseFee(), blobBaseFee);
        assertEq(l1Block.hash(), hash);
        assertEq(l1Block.batcherHash(), batcherHash);
        assertEq(l1Block.operatorFeeScalar(), operatorFeeScalar);
        assertEq(l1Block.operatorFeeConstant(), operatorFeeConstant);

        assertEmptyLegacySlotRanges();
    }

    /// @notice Tests that `setL1BlockValuesIsthmus` succeeds with max values
    function test_setL1BlockValuesIsthmus_isDepositorMax_succeeds() external {
        bytes memory functionCallDataPacked = Encoding.encodeSetL1BlockValuesIsthmus(
            type(uint32).max,
            type(uint32).max,
            type(uint64).max,
            type(uint64).max,
            type(uint64).max,
            type(uint256).max,
            type(uint256).max,
            bytes32(type(uint256).max),
            bytes32(type(uint256).max),
            type(uint32).max,
            type(uint64).max
        );

        vm.prank(depositor);
        (bool success,) = address(l1Block).call(functionCallDataPacked);
        assertTrue(success, "function call failed");
    }

    /// @notice Tests that `setL1BlockValuesIsthmus` reverts if sender address is not the depositor
    function test_setL1BlockValuesIsthmus_notDepositor_reverts() external {
        bytes memory functionCallDataPacked = Encoding.encodeSetL1BlockValuesIsthmus(
            type(uint32).max,
            type(uint32).max,
            type(uint64).max,
            type(uint64).max,
            type(uint64).max,
            type(uint256).max,
            type(uint256).max,
            bytes32(type(uint256).max),
            bytes32(type(uint256).max),
            type(uint32).max,
            type(uint64).max
        );

        (bool success, bytes memory data) = address(l1Block).call(functionCallDataPacked);
        assertTrue(!success, "function call should have failed");
        // make sure return value is the expected function selector for "NotDepositor()"
        bytes memory expReturn = hex"3cc50b45";
        assertEq(data, expReturn);
    }
}

/// @title L1Block_SetL1BlockValuesJovian_Test
/// @notice Tests the `setL1BlockValuesJovian` function of the `L1Block` contract.
contract L1Block_SetL1BlockValuesJovian_Test is L1Block_TestInit {
    /// @notice Struct to group parameters for L1BlockValuesJovian to avoid stack too deep.
    struct L1BlockValuesJovianParams {
        uint32 baseFeeScalar;
        uint32 blobBaseFeeScalar;
        uint64 sequenceNumber;
        uint64 timestamp;
        uint64 number;
        uint256 baseFee;
        uint256 blobBaseFee;
        bytes32 hash;
        bytes32 batcherHash;
        uint32 operatorFeeScalar;
        uint64 operatorFeeConstant;
        uint16 daFootprintGasScalar;
    }

    /// @notice Tests that setL1BlockValuesJovian updates the values appropriately.
    function testFuzz_setL1BlockValuesJovian_succeeds(L1BlockValuesJovianParams memory params) external {
        bytes memory functionCallDataPacked = Encoding.encodeSetL1BlockValuesJovian(
            params.baseFeeScalar,
            params.blobBaseFeeScalar,
            params.sequenceNumber,
            params.timestamp,
            params.number,
            params.baseFee,
            params.blobBaseFee,
            params.hash,
            params.batcherHash,
            params.operatorFeeScalar,
            params.operatorFeeConstant,
            params.daFootprintGasScalar
        );

        vm.prank(depositor);
        (bool success,) = address(l1Block).call(functionCallDataPacked);
        assertTrue(success, "Function call failed");

        assertEq(l1Block.baseFeeScalar(), params.baseFeeScalar);
        assertEq(l1Block.blobBaseFeeScalar(), params.blobBaseFeeScalar);
        assertEq(l1Block.sequenceNumber(), params.sequenceNumber);
        assertEq(l1Block.timestamp(), params.timestamp);
        assertEq(l1Block.number(), params.number);
        assertEq(l1Block.basefee(), params.baseFee);
        assertEq(l1Block.blobBaseFee(), params.blobBaseFee);
        assertEq(l1Block.hash(), params.hash);
        assertEq(l1Block.batcherHash(), params.batcherHash);
        assertEq(l1Block.operatorFeeScalar(), params.operatorFeeScalar);
        assertEq(l1Block.operatorFeeConstant(), params.operatorFeeConstant);
        assertEq(l1Block.daFootprintGasScalar(), params.daFootprintGasScalar);

        assertEmptyLegacySlotRanges();
    }

    /// @notice Tests that `setL1BlockValuesJovian` succeeds with max values
    function test_setL1BlockValuesJovian_isDepositorMax_succeeds() external {
        bytes memory functionCallDataPacked = Encoding.encodeSetL1BlockValuesJovian(
            type(uint32).max,
            type(uint32).max,
            type(uint64).max,
            type(uint64).max,
            type(uint64).max,
            type(uint256).max,
            type(uint256).max,
            bytes32(type(uint256).max),
            bytes32(type(uint256).max),
            type(uint32).max,
            type(uint64).max,
            type(uint16).max
        );

        vm.prank(depositor);
        (bool success,) = address(l1Block).call(functionCallDataPacked);
        assertTrue(success, "function call failed");
    }

    /// @notice Tests that `setL1BlockValuesJovian` reverts if sender address is not the depositor
    function test_setL1BlockValuesJovian_notDepositor_reverts() external {
        bytes memory functionCallDataPacked = Encoding.encodeSetL1BlockValuesJovian(
            type(uint32).max,
            type(uint32).max,
            type(uint64).max,
            type(uint64).max,
            type(uint64).max,
            type(uint256).max,
            type(uint256).max,
            bytes32(type(uint256).max),
            bytes32(type(uint256).max),
            type(uint32).max,
            type(uint64).max,
            type(uint16).max
        );

        (bool success, bytes memory data) = address(l1Block).call(functionCallDataPacked);
        assertTrue(!success, "function call should have failed");
        // make sure return value is the expected function selector for "NotDepositor()"
        bytes memory expReturn = hex"3cc50b45";
        assertEq(data, expReturn);
    }
}

contract L1BlockCustomGasToken_Test is L1Block_TestInit {
    function testFuzz_setGasPayingToken_succeeds(
        address _token,
        uint8 _decimals,
        string calldata _name,
        string calldata _symbol
    )
        external
    {
        vm.assume(_token != address(0));
        vm.assume(_token != Constants.ETHER);

        // Using vm.assume() would cause too many test rejections.
        string memory name = _name;
        if (bytes(_name).length > 32) {
            name = _name[:32];
        }
        bytes32 b32name = bytes32(abi.encodePacked(name));

        // Using vm.assume() would cause too many test rejections.
        string memory symbol = _symbol;
        if (bytes(_symbol).length > 32) {
            symbol = _symbol[:32];
        }
        bytes32 b32symbol = bytes32(abi.encodePacked(symbol));

        vm.expectEmit(address(l1Block));
        emit GasPayingTokenSet({ token: _token, decimals: _decimals, name: b32name, symbol: b32symbol });

        vm.prank(depositor);
        l1Block.setGasPayingToken({ _token: _token, _decimals: _decimals, _name: b32name, _symbol: b32symbol });

        (address token, uint8 decimals) = l1Block.gasPayingToken();
        assertEq(token, _token);
        assertEq(decimals, _decimals);

        assertEq(name, l1Block.gasPayingTokenName());
        assertEq(symbol, l1Block.gasPayingTokenSymbol());
        assertTrue(l1Block.isCustomGasToken());
    }

    function test_setGasPayingToken_isDepositor_reverts() external {
        vm.expectRevert(NotDepositor.selector);
        l1Block.setGasPayingToken(address(this), 18, "Test", "TST");
    }
}
