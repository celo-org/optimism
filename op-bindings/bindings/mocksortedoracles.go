// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// MockSortedOraclesMetaData contains all meta data concerning the MockSortedOracles contract.
var MockSortedOraclesMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"DENOMINATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"expired\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOldestReportExpired\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"medianRate\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"medianTimestamp\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numRates\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numerators\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setMedianRate\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"numerator\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMedianTimestamp\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMedianTimestampToNow\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setNumRates\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"rate\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOldestReportExpired\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"}]",
	Bin: "0x608060405234801561001057600080fd5b5061045f806100206000396000f3fe608060405234801561001057600080fd5b50600436106100d45760003560e01c8063b325dd1f11610081578063f06a10e01161005b578063f06a10e0146102c6578063f7ca6963146102e9578063ffe736bf1461030957600080fd5b8063b325dd1f14610235578063bbc66a941461027e578063ef90e1b01461029e57600080fd5b806376ea210e116100b257806376ea210e1461017c5780638d24fc51146101c7578063918f86741461022457600080fd5b8063071b48fc146100d95780631b7aa2db1461010c578063495182a614610145575b600080fd5b6100f96100e7366004610406565b60016020526000908152604090205481565b6040519081526020015b60405180910390f35b61014361011a366004610428565b73ffffffffffffffffffffffffffffffffffffffff909116600090815260016020526040902055565b005b610143610153366004610428565b73ffffffffffffffffffffffffffffffffffffffff909116600090815260026020526040902055565b6101b761018a366004610428565b73ffffffffffffffffffffffffffffffffffffffff91909116600090815260208190526040902055600190565b6040519015158152602001610103565b6101436101d5366004610406565b73ffffffffffffffffffffffffffffffffffffffff16600090815260036020526040902080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00166001179055565b6100f969d3c21bcecceda100000081565b610143610243366004610406565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604090206fffffffffffffffffffffffffffffffff42169055565b6100f961028c366004610406565b60026020526000908152604090205481565b6102b16102ac366004610406565b61036f565b60408051928352602083019190915201610103565b6101b76102d4366004610406565b60036020526000908152604090205460ff1681565b6100f96102f7366004610406565b60006020819052908152604090205481565b610343610317366004610406565b73ffffffffffffffffffffffffffffffffffffffff811660009081526003602052604090205460ff1691565b60408051921515835273ffffffffffffffffffffffffffffffffffffffff909116602083015201610103565b73ffffffffffffffffffffffffffffffffffffffff81166000908152602081905260408120548190156103d257505073ffffffffffffffffffffffffffffffffffffffff166000908152602081905260409020549069d3c21bcecceda100000090565b506000928392509050565b803573ffffffffffffffffffffffffffffffffffffffff8116811461040157600080fd5b919050565b60006020828403121561041857600080fd5b610421826103dd565b9392505050565b6000806040838503121561043b57600080fd5b610444836103dd565b94602093909301359350505056fea164736f6c6343000813000a",
}

// MockSortedOraclesABI is the input ABI used to generate the binding from.
// Deprecated: Use MockSortedOraclesMetaData.ABI instead.
var MockSortedOraclesABI = MockSortedOraclesMetaData.ABI

// MockSortedOraclesBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use MockSortedOraclesMetaData.Bin instead.
var MockSortedOraclesBin = MockSortedOraclesMetaData.Bin

// DeployMockSortedOracles deploys a new Ethereum contract, binding an instance of MockSortedOracles to it.
func DeployMockSortedOracles(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *MockSortedOracles, error) {
	parsed, err := MockSortedOraclesMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(MockSortedOraclesBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MockSortedOracles{MockSortedOraclesCaller: MockSortedOraclesCaller{contract: contract}, MockSortedOraclesTransactor: MockSortedOraclesTransactor{contract: contract}, MockSortedOraclesFilterer: MockSortedOraclesFilterer{contract: contract}}, nil
}

// MockSortedOracles is an auto generated Go binding around an Ethereum contract.
type MockSortedOracles struct {
	MockSortedOraclesCaller     // Read-only binding to the contract
	MockSortedOraclesTransactor // Write-only binding to the contract
	MockSortedOraclesFilterer   // Log filterer for contract events
}

// MockSortedOraclesCaller is an auto generated read-only Go binding around an Ethereum contract.
type MockSortedOraclesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockSortedOraclesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MockSortedOraclesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockSortedOraclesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MockSortedOraclesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockSortedOraclesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MockSortedOraclesSession struct {
	Contract     *MockSortedOracles // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// MockSortedOraclesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MockSortedOraclesCallerSession struct {
	Contract *MockSortedOraclesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// MockSortedOraclesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MockSortedOraclesTransactorSession struct {
	Contract     *MockSortedOraclesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// MockSortedOraclesRaw is an auto generated low-level Go binding around an Ethereum contract.
type MockSortedOraclesRaw struct {
	Contract *MockSortedOracles // Generic contract binding to access the raw methods on
}

// MockSortedOraclesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MockSortedOraclesCallerRaw struct {
	Contract *MockSortedOraclesCaller // Generic read-only contract binding to access the raw methods on
}

// MockSortedOraclesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MockSortedOraclesTransactorRaw struct {
	Contract *MockSortedOraclesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMockSortedOracles creates a new instance of MockSortedOracles, bound to a specific deployed contract.
func NewMockSortedOracles(address common.Address, backend bind.ContractBackend) (*MockSortedOracles, error) {
	contract, err := bindMockSortedOracles(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MockSortedOracles{MockSortedOraclesCaller: MockSortedOraclesCaller{contract: contract}, MockSortedOraclesTransactor: MockSortedOraclesTransactor{contract: contract}, MockSortedOraclesFilterer: MockSortedOraclesFilterer{contract: contract}}, nil
}

// NewMockSortedOraclesCaller creates a new read-only instance of MockSortedOracles, bound to a specific deployed contract.
func NewMockSortedOraclesCaller(address common.Address, caller bind.ContractCaller) (*MockSortedOraclesCaller, error) {
	contract, err := bindMockSortedOracles(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MockSortedOraclesCaller{contract: contract}, nil
}

// NewMockSortedOraclesTransactor creates a new write-only instance of MockSortedOracles, bound to a specific deployed contract.
func NewMockSortedOraclesTransactor(address common.Address, transactor bind.ContractTransactor) (*MockSortedOraclesTransactor, error) {
	contract, err := bindMockSortedOracles(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MockSortedOraclesTransactor{contract: contract}, nil
}

// NewMockSortedOraclesFilterer creates a new log filterer instance of MockSortedOracles, bound to a specific deployed contract.
func NewMockSortedOraclesFilterer(address common.Address, filterer bind.ContractFilterer) (*MockSortedOraclesFilterer, error) {
	contract, err := bindMockSortedOracles(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MockSortedOraclesFilterer{contract: contract}, nil
}

// bindMockSortedOracles binds a generic wrapper to an already deployed contract.
func bindMockSortedOracles(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MockSortedOraclesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MockSortedOracles *MockSortedOraclesRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MockSortedOracles.Contract.MockSortedOraclesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MockSortedOracles *MockSortedOraclesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.MockSortedOraclesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MockSortedOracles *MockSortedOraclesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.MockSortedOraclesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MockSortedOracles *MockSortedOraclesCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MockSortedOracles.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MockSortedOracles *MockSortedOraclesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MockSortedOracles *MockSortedOraclesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.contract.Transact(opts, method, params...)
}

// DENOMINATOR is a free data retrieval call binding the contract method 0x918f8674.
//
// Solidity: function DENOMINATOR() view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesCaller) DENOMINATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MockSortedOracles.contract.Call(opts, &out, "DENOMINATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DENOMINATOR is a free data retrieval call binding the contract method 0x918f8674.
//
// Solidity: function DENOMINATOR() view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesSession) DENOMINATOR() (*big.Int, error) {
	return _MockSortedOracles.Contract.DENOMINATOR(&_MockSortedOracles.CallOpts)
}

// DENOMINATOR is a free data retrieval call binding the contract method 0x918f8674.
//
// Solidity: function DENOMINATOR() view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesCallerSession) DENOMINATOR() (*big.Int, error) {
	return _MockSortedOracles.Contract.DENOMINATOR(&_MockSortedOracles.CallOpts)
}

// Expired is a free data retrieval call binding the contract method 0xf06a10e0.
//
// Solidity: function expired(address ) view returns(bool)
func (_MockSortedOracles *MockSortedOraclesCaller) Expired(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _MockSortedOracles.contract.Call(opts, &out, "expired", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Expired is a free data retrieval call binding the contract method 0xf06a10e0.
//
// Solidity: function expired(address ) view returns(bool)
func (_MockSortedOracles *MockSortedOraclesSession) Expired(arg0 common.Address) (bool, error) {
	return _MockSortedOracles.Contract.Expired(&_MockSortedOracles.CallOpts, arg0)
}

// Expired is a free data retrieval call binding the contract method 0xf06a10e0.
//
// Solidity: function expired(address ) view returns(bool)
func (_MockSortedOracles *MockSortedOraclesCallerSession) Expired(arg0 common.Address) (bool, error) {
	return _MockSortedOracles.Contract.Expired(&_MockSortedOracles.CallOpts, arg0)
}

// IsOldestReportExpired is a free data retrieval call binding the contract method 0xffe736bf.
//
// Solidity: function isOldestReportExpired(address token) view returns(bool, address)
func (_MockSortedOracles *MockSortedOraclesCaller) IsOldestReportExpired(opts *bind.CallOpts, token common.Address) (bool, common.Address, error) {
	var out []interface{}
	err := _MockSortedOracles.contract.Call(opts, &out, "isOldestReportExpired", token)

	if err != nil {
		return *new(bool), *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)
	out1 := *abi.ConvertType(out[1], new(common.Address)).(*common.Address)

	return out0, out1, err

}

// IsOldestReportExpired is a free data retrieval call binding the contract method 0xffe736bf.
//
// Solidity: function isOldestReportExpired(address token) view returns(bool, address)
func (_MockSortedOracles *MockSortedOraclesSession) IsOldestReportExpired(token common.Address) (bool, common.Address, error) {
	return _MockSortedOracles.Contract.IsOldestReportExpired(&_MockSortedOracles.CallOpts, token)
}

// IsOldestReportExpired is a free data retrieval call binding the contract method 0xffe736bf.
//
// Solidity: function isOldestReportExpired(address token) view returns(bool, address)
func (_MockSortedOracles *MockSortedOraclesCallerSession) IsOldestReportExpired(token common.Address) (bool, common.Address, error) {
	return _MockSortedOracles.Contract.IsOldestReportExpired(&_MockSortedOracles.CallOpts, token)
}

// MedianRate is a free data retrieval call binding the contract method 0xef90e1b0.
//
// Solidity: function medianRate(address token) view returns(uint256, uint256)
func (_MockSortedOracles *MockSortedOraclesCaller) MedianRate(opts *bind.CallOpts, token common.Address) (*big.Int, *big.Int, error) {
	var out []interface{}
	err := _MockSortedOracles.contract.Call(opts, &out, "medianRate", token)

	if err != nil {
		return *new(*big.Int), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// MedianRate is a free data retrieval call binding the contract method 0xef90e1b0.
//
// Solidity: function medianRate(address token) view returns(uint256, uint256)
func (_MockSortedOracles *MockSortedOraclesSession) MedianRate(token common.Address) (*big.Int, *big.Int, error) {
	return _MockSortedOracles.Contract.MedianRate(&_MockSortedOracles.CallOpts, token)
}

// MedianRate is a free data retrieval call binding the contract method 0xef90e1b0.
//
// Solidity: function medianRate(address token) view returns(uint256, uint256)
func (_MockSortedOracles *MockSortedOraclesCallerSession) MedianRate(token common.Address) (*big.Int, *big.Int, error) {
	return _MockSortedOracles.Contract.MedianRate(&_MockSortedOracles.CallOpts, token)
}

// MedianTimestamp is a free data retrieval call binding the contract method 0x071b48fc.
//
// Solidity: function medianTimestamp(address ) view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesCaller) MedianTimestamp(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _MockSortedOracles.contract.Call(opts, &out, "medianTimestamp", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MedianTimestamp is a free data retrieval call binding the contract method 0x071b48fc.
//
// Solidity: function medianTimestamp(address ) view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesSession) MedianTimestamp(arg0 common.Address) (*big.Int, error) {
	return _MockSortedOracles.Contract.MedianTimestamp(&_MockSortedOracles.CallOpts, arg0)
}

// MedianTimestamp is a free data retrieval call binding the contract method 0x071b48fc.
//
// Solidity: function medianTimestamp(address ) view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesCallerSession) MedianTimestamp(arg0 common.Address) (*big.Int, error) {
	return _MockSortedOracles.Contract.MedianTimestamp(&_MockSortedOracles.CallOpts, arg0)
}

// NumRates is a free data retrieval call binding the contract method 0xbbc66a94.
//
// Solidity: function numRates(address ) view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesCaller) NumRates(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _MockSortedOracles.contract.Call(opts, &out, "numRates", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumRates is a free data retrieval call binding the contract method 0xbbc66a94.
//
// Solidity: function numRates(address ) view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesSession) NumRates(arg0 common.Address) (*big.Int, error) {
	return _MockSortedOracles.Contract.NumRates(&_MockSortedOracles.CallOpts, arg0)
}

// NumRates is a free data retrieval call binding the contract method 0xbbc66a94.
//
// Solidity: function numRates(address ) view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesCallerSession) NumRates(arg0 common.Address) (*big.Int, error) {
	return _MockSortedOracles.Contract.NumRates(&_MockSortedOracles.CallOpts, arg0)
}

// Numerators is a free data retrieval call binding the contract method 0xf7ca6963.
//
// Solidity: function numerators(address ) view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesCaller) Numerators(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _MockSortedOracles.contract.Call(opts, &out, "numerators", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Numerators is a free data retrieval call binding the contract method 0xf7ca6963.
//
// Solidity: function numerators(address ) view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesSession) Numerators(arg0 common.Address) (*big.Int, error) {
	return _MockSortedOracles.Contract.Numerators(&_MockSortedOracles.CallOpts, arg0)
}

// Numerators is a free data retrieval call binding the contract method 0xf7ca6963.
//
// Solidity: function numerators(address ) view returns(uint256)
func (_MockSortedOracles *MockSortedOraclesCallerSession) Numerators(arg0 common.Address) (*big.Int, error) {
	return _MockSortedOracles.Contract.Numerators(&_MockSortedOracles.CallOpts, arg0)
}

// SetMedianRate is a paid mutator transaction binding the contract method 0x76ea210e.
//
// Solidity: function setMedianRate(address token, uint256 numerator) returns(bool)
func (_MockSortedOracles *MockSortedOraclesTransactor) SetMedianRate(opts *bind.TransactOpts, token common.Address, numerator *big.Int) (*types.Transaction, error) {
	return _MockSortedOracles.contract.Transact(opts, "setMedianRate", token, numerator)
}

// SetMedianRate is a paid mutator transaction binding the contract method 0x76ea210e.
//
// Solidity: function setMedianRate(address token, uint256 numerator) returns(bool)
func (_MockSortedOracles *MockSortedOraclesSession) SetMedianRate(token common.Address, numerator *big.Int) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.SetMedianRate(&_MockSortedOracles.TransactOpts, token, numerator)
}

// SetMedianRate is a paid mutator transaction binding the contract method 0x76ea210e.
//
// Solidity: function setMedianRate(address token, uint256 numerator) returns(bool)
func (_MockSortedOracles *MockSortedOraclesTransactorSession) SetMedianRate(token common.Address, numerator *big.Int) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.SetMedianRate(&_MockSortedOracles.TransactOpts, token, numerator)
}

// SetMedianTimestamp is a paid mutator transaction binding the contract method 0x1b7aa2db.
//
// Solidity: function setMedianTimestamp(address token, uint256 timestamp) returns()
func (_MockSortedOracles *MockSortedOraclesTransactor) SetMedianTimestamp(opts *bind.TransactOpts, token common.Address, timestamp *big.Int) (*types.Transaction, error) {
	return _MockSortedOracles.contract.Transact(opts, "setMedianTimestamp", token, timestamp)
}

// SetMedianTimestamp is a paid mutator transaction binding the contract method 0x1b7aa2db.
//
// Solidity: function setMedianTimestamp(address token, uint256 timestamp) returns()
func (_MockSortedOracles *MockSortedOraclesSession) SetMedianTimestamp(token common.Address, timestamp *big.Int) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.SetMedianTimestamp(&_MockSortedOracles.TransactOpts, token, timestamp)
}

// SetMedianTimestamp is a paid mutator transaction binding the contract method 0x1b7aa2db.
//
// Solidity: function setMedianTimestamp(address token, uint256 timestamp) returns()
func (_MockSortedOracles *MockSortedOraclesTransactorSession) SetMedianTimestamp(token common.Address, timestamp *big.Int) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.SetMedianTimestamp(&_MockSortedOracles.TransactOpts, token, timestamp)
}

// SetMedianTimestampToNow is a paid mutator transaction binding the contract method 0xb325dd1f.
//
// Solidity: function setMedianTimestampToNow(address token) returns()
func (_MockSortedOracles *MockSortedOraclesTransactor) SetMedianTimestampToNow(opts *bind.TransactOpts, token common.Address) (*types.Transaction, error) {
	return _MockSortedOracles.contract.Transact(opts, "setMedianTimestampToNow", token)
}

// SetMedianTimestampToNow is a paid mutator transaction binding the contract method 0xb325dd1f.
//
// Solidity: function setMedianTimestampToNow(address token) returns()
func (_MockSortedOracles *MockSortedOraclesSession) SetMedianTimestampToNow(token common.Address) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.SetMedianTimestampToNow(&_MockSortedOracles.TransactOpts, token)
}

// SetMedianTimestampToNow is a paid mutator transaction binding the contract method 0xb325dd1f.
//
// Solidity: function setMedianTimestampToNow(address token) returns()
func (_MockSortedOracles *MockSortedOraclesTransactorSession) SetMedianTimestampToNow(token common.Address) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.SetMedianTimestampToNow(&_MockSortedOracles.TransactOpts, token)
}

// SetNumRates is a paid mutator transaction binding the contract method 0x495182a6.
//
// Solidity: function setNumRates(address token, uint256 rate) returns()
func (_MockSortedOracles *MockSortedOraclesTransactor) SetNumRates(opts *bind.TransactOpts, token common.Address, rate *big.Int) (*types.Transaction, error) {
	return _MockSortedOracles.contract.Transact(opts, "setNumRates", token, rate)
}

// SetNumRates is a paid mutator transaction binding the contract method 0x495182a6.
//
// Solidity: function setNumRates(address token, uint256 rate) returns()
func (_MockSortedOracles *MockSortedOraclesSession) SetNumRates(token common.Address, rate *big.Int) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.SetNumRates(&_MockSortedOracles.TransactOpts, token, rate)
}

// SetNumRates is a paid mutator transaction binding the contract method 0x495182a6.
//
// Solidity: function setNumRates(address token, uint256 rate) returns()
func (_MockSortedOracles *MockSortedOraclesTransactorSession) SetNumRates(token common.Address, rate *big.Int) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.SetNumRates(&_MockSortedOracles.TransactOpts, token, rate)
}

// SetOldestReportExpired is a paid mutator transaction binding the contract method 0x8d24fc51.
//
// Solidity: function setOldestReportExpired(address token) returns()
func (_MockSortedOracles *MockSortedOraclesTransactor) SetOldestReportExpired(opts *bind.TransactOpts, token common.Address) (*types.Transaction, error) {
	return _MockSortedOracles.contract.Transact(opts, "setOldestReportExpired", token)
}

// SetOldestReportExpired is a paid mutator transaction binding the contract method 0x8d24fc51.
//
// Solidity: function setOldestReportExpired(address token) returns()
func (_MockSortedOracles *MockSortedOraclesSession) SetOldestReportExpired(token common.Address) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.SetOldestReportExpired(&_MockSortedOracles.TransactOpts, token)
}

// SetOldestReportExpired is a paid mutator transaction binding the contract method 0x8d24fc51.
//
// Solidity: function setOldestReportExpired(address token) returns()
func (_MockSortedOracles *MockSortedOraclesTransactorSession) SetOldestReportExpired(token common.Address) (*types.Transaction, error) {
	return _MockSortedOracles.Contract.SetOldestReportExpired(&_MockSortedOracles.TransactOpts, token)
}
