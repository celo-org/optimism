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

// FeeCurrencyWhitelistMetaData contains all meta data concerning the FeeCurrencyWhitelist contract.
var FeeCurrencyWhitelistMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"test\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"addToken\",\"inputs\":[{\"name\":\"tokenAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getVersionNumber\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"getWhitelist\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialized\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"removeToken\",\"inputs\":[{\"name\":\"tokenAddress\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"whitelist\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"FeeCurrencyWhitelistRemoved\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"FeeCurrencyWhitelisted\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false}]",
	Bin: "0x608060405234801561001057600080fd5b50604051610a4b380380610a4b83398101604081905261002f916100a9565b8061003933610059565b80610052576000805460ff60a01b1916600160a01b1790555b50506100d2565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6000602082840312156100bb57600080fd5b815180151581146100cb57600080fd5b9392505050565b61096a806100e16000396000f3fe608060405234801561001057600080fd5b50600436106100be5760003560e01c80638129fc1c11610076578063d01f63f51161005b578063d01f63f51461019e578063d48bfca7146101b3578063f2fde38b146101c657600080fd5b80638129fc1c146101785780638da5cb5b1461018057600080fd5b806354255be0116100a757806354255be014610112578063715018a6146101385780637ebd1b301461014057600080fd5b806313baf1e6146100c3578063158ef93e146100d8575b600080fd5b6100d66100d1366004610800565b6101d9565b005b6000546100fd9074010000000000000000000000000000000000000000900460ff1681565b60405190151581526020015b60405180910390f35b600180806000604080519485526020850193909352918301526060820152608001610109565b6100d66103f3565b61015361014e36600461082a565b610407565b60405173ffffffffffffffffffffffffffffffffffffffff9091168152602001610109565b6100d661043e565b60005473ffffffffffffffffffffffffffffffffffffffff16610153565b6101a661050b565b6040516101099190610843565b6100d66101c136600461089d565b61057a565b6100d66101d436600461089d565b61062a565b6101e16106e1565b8173ffffffffffffffffffffffffffffffffffffffff166001828154811061020b5761020b6108bf565b60009182526020909120015473ffffffffffffffffffffffffffffffffffffffff1614610299576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601460248201527f496e64657820646f6573206e6f74206d6174636800000000000000000000000060448201526064015b60405180910390fd5b60018054906102a881836108ee565b815481106102b8576102b86108bf565b6000918252602090912001546001805473ffffffffffffffffffffffffffffffffffffffff90921691849081106102f1576102f16108bf565b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550600180548061034a5761034a61092e565b60008281526020908190207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff908301810180547fffffffffffffffffffffffff000000000000000000000000000000000000000016905590910190915560405173ffffffffffffffffffffffffffffffffffffffff851681527fc1f06ffbe5c19d22daa982fd4b3886ced05d83e2bfe7de3b67222728f5234e28910160405180910390a1505050565b6103fb6106e1565b6104056000610762565b565b6001818154811061041757600080fd5b60009182526020909120015473ffffffffffffffffffffffffffffffffffffffff16905081565b60005474010000000000000000000000000000000000000000900460ff16156104c3576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601c60248201527f636f6e747261637420616c726561647920696e697469616c697a6564000000006044820152606401610290565b600080547fffffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffff167401000000000000000000000000000000000000000017905561040533610762565b6060600180548060200260200160405190810160405280929190818152602001828054801561057057602002820191906000526020600020905b815473ffffffffffffffffffffffffffffffffffffffff168152600190910190602001808311610545575b5050505050905090565b6105826106e1565b6001805480820182556000919091527fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf60180547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff83169081179091556040519081527fcf4fe1d1989a39011040b0c33ba1165fdeeca971a1ab2b0340b23550f93727e09060200160405180910390a150565b6106326106e1565b73ffffffffffffffffffffffffffffffffffffffff81166106d5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f64647265737300000000000000000000000000000000000000000000000000006064820152608401610290565b6106de81610762565b50565b60005473ffffffffffffffffffffffffffffffffffffffff163314610405576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e65726044820152606401610290565b6000805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b803573ffffffffffffffffffffffffffffffffffffffff811681146107fb57600080fd5b919050565b6000806040838503121561081357600080fd5b61081c836107d7565b946020939093013593505050565b60006020828403121561083c57600080fd5b5035919050565b6020808252825182820181905260009190848201906040850190845b8181101561089157835173ffffffffffffffffffffffffffffffffffffffff168352928401929184019160010161085f565b50909695505050505050565b6000602082840312156108af57600080fd5b6108b8826107d7565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b81810381811115610928577f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603160045260246000fdfea164736f6c6343000813000a",
}

// FeeCurrencyWhitelistABI is the input ABI used to generate the binding from.
// Deprecated: Use FeeCurrencyWhitelistMetaData.ABI instead.
var FeeCurrencyWhitelistABI = FeeCurrencyWhitelistMetaData.ABI

// FeeCurrencyWhitelistBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use FeeCurrencyWhitelistMetaData.Bin instead.
var FeeCurrencyWhitelistBin = FeeCurrencyWhitelistMetaData.Bin

// DeployFeeCurrencyWhitelist deploys a new Ethereum contract, binding an instance of FeeCurrencyWhitelist to it.
func DeployFeeCurrencyWhitelist(auth *bind.TransactOpts, backend bind.ContractBackend, test bool) (common.Address, *types.Transaction, *FeeCurrencyWhitelist, error) {
	parsed, err := FeeCurrencyWhitelistMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(FeeCurrencyWhitelistBin), backend, test)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &FeeCurrencyWhitelist{FeeCurrencyWhitelistCaller: FeeCurrencyWhitelistCaller{contract: contract}, FeeCurrencyWhitelistTransactor: FeeCurrencyWhitelistTransactor{contract: contract}, FeeCurrencyWhitelistFilterer: FeeCurrencyWhitelistFilterer{contract: contract}}, nil
}

// FeeCurrencyWhitelist is an auto generated Go binding around an Ethereum contract.
type FeeCurrencyWhitelist struct {
	FeeCurrencyWhitelistCaller     // Read-only binding to the contract
	FeeCurrencyWhitelistTransactor // Write-only binding to the contract
	FeeCurrencyWhitelistFilterer   // Log filterer for contract events
}

// FeeCurrencyWhitelistCaller is an auto generated read-only Go binding around an Ethereum contract.
type FeeCurrencyWhitelistCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FeeCurrencyWhitelistTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FeeCurrencyWhitelistTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FeeCurrencyWhitelistFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FeeCurrencyWhitelistFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FeeCurrencyWhitelistSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FeeCurrencyWhitelistSession struct {
	Contract     *FeeCurrencyWhitelist // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// FeeCurrencyWhitelistCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FeeCurrencyWhitelistCallerSession struct {
	Contract *FeeCurrencyWhitelistCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// FeeCurrencyWhitelistTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FeeCurrencyWhitelistTransactorSession struct {
	Contract     *FeeCurrencyWhitelistTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// FeeCurrencyWhitelistRaw is an auto generated low-level Go binding around an Ethereum contract.
type FeeCurrencyWhitelistRaw struct {
	Contract *FeeCurrencyWhitelist // Generic contract binding to access the raw methods on
}

// FeeCurrencyWhitelistCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FeeCurrencyWhitelistCallerRaw struct {
	Contract *FeeCurrencyWhitelistCaller // Generic read-only contract binding to access the raw methods on
}

// FeeCurrencyWhitelistTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FeeCurrencyWhitelistTransactorRaw struct {
	Contract *FeeCurrencyWhitelistTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFeeCurrencyWhitelist creates a new instance of FeeCurrencyWhitelist, bound to a specific deployed contract.
func NewFeeCurrencyWhitelist(address common.Address, backend bind.ContractBackend) (*FeeCurrencyWhitelist, error) {
	contract, err := bindFeeCurrencyWhitelist(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyWhitelist{FeeCurrencyWhitelistCaller: FeeCurrencyWhitelistCaller{contract: contract}, FeeCurrencyWhitelistTransactor: FeeCurrencyWhitelistTransactor{contract: contract}, FeeCurrencyWhitelistFilterer: FeeCurrencyWhitelistFilterer{contract: contract}}, nil
}

// NewFeeCurrencyWhitelistCaller creates a new read-only instance of FeeCurrencyWhitelist, bound to a specific deployed contract.
func NewFeeCurrencyWhitelistCaller(address common.Address, caller bind.ContractCaller) (*FeeCurrencyWhitelistCaller, error) {
	contract, err := bindFeeCurrencyWhitelist(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyWhitelistCaller{contract: contract}, nil
}

// NewFeeCurrencyWhitelistTransactor creates a new write-only instance of FeeCurrencyWhitelist, bound to a specific deployed contract.
func NewFeeCurrencyWhitelistTransactor(address common.Address, transactor bind.ContractTransactor) (*FeeCurrencyWhitelistTransactor, error) {
	contract, err := bindFeeCurrencyWhitelist(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyWhitelistTransactor{contract: contract}, nil
}

// NewFeeCurrencyWhitelistFilterer creates a new log filterer instance of FeeCurrencyWhitelist, bound to a specific deployed contract.
func NewFeeCurrencyWhitelistFilterer(address common.Address, filterer bind.ContractFilterer) (*FeeCurrencyWhitelistFilterer, error) {
	contract, err := bindFeeCurrencyWhitelist(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyWhitelistFilterer{contract: contract}, nil
}

// bindFeeCurrencyWhitelist binds a generic wrapper to an already deployed contract.
func bindFeeCurrencyWhitelist(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(FeeCurrencyWhitelistABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FeeCurrencyWhitelist.Contract.FeeCurrencyWhitelistCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.FeeCurrencyWhitelistTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.FeeCurrencyWhitelistTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FeeCurrencyWhitelist.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.contract.Transact(opts, method, params...)
}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCaller) GetVersionNumber(opts *bind.CallOpts) (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	var out []interface{}
	err := _FeeCurrencyWhitelist.contract.Call(opts, &out, "getVersionNumber")

	if err != nil {
		return *new(*big.Int), *new(*big.Int), *new(*big.Int), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	out2 := *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	out3 := *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return out0, out1, out2, out3, err

}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistSession) GetVersionNumber() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _FeeCurrencyWhitelist.Contract.GetVersionNumber(&_FeeCurrencyWhitelist.CallOpts)
}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCallerSession) GetVersionNumber() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _FeeCurrencyWhitelist.Contract.GetVersionNumber(&_FeeCurrencyWhitelist.CallOpts)
}

// GetWhitelist is a free data retrieval call binding the contract method 0xd01f63f5.
//
// Solidity: function getWhitelist() view returns(address[])
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCaller) GetWhitelist(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _FeeCurrencyWhitelist.contract.Call(opts, &out, "getWhitelist")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetWhitelist is a free data retrieval call binding the contract method 0xd01f63f5.
//
// Solidity: function getWhitelist() view returns(address[])
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistSession) GetWhitelist() ([]common.Address, error) {
	return _FeeCurrencyWhitelist.Contract.GetWhitelist(&_FeeCurrencyWhitelist.CallOpts)
}

// GetWhitelist is a free data retrieval call binding the contract method 0xd01f63f5.
//
// Solidity: function getWhitelist() view returns(address[])
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCallerSession) GetWhitelist() ([]common.Address, error) {
	return _FeeCurrencyWhitelist.Contract.GetWhitelist(&_FeeCurrencyWhitelist.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCaller) Initialized(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _FeeCurrencyWhitelist.contract.Call(opts, &out, "initialized")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistSession) Initialized() (bool, error) {
	return _FeeCurrencyWhitelist.Contract.Initialized(&_FeeCurrencyWhitelist.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCallerSession) Initialized() (bool, error) {
	return _FeeCurrencyWhitelist.Contract.Initialized(&_FeeCurrencyWhitelist.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FeeCurrencyWhitelist.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistSession) Owner() (common.Address, error) {
	return _FeeCurrencyWhitelist.Contract.Owner(&_FeeCurrencyWhitelist.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCallerSession) Owner() (common.Address, error) {
	return _FeeCurrencyWhitelist.Contract.Owner(&_FeeCurrencyWhitelist.CallOpts)
}

// Whitelist is a free data retrieval call binding the contract method 0x7ebd1b30.
//
// Solidity: function whitelist(uint256 ) view returns(address)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCaller) Whitelist(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _FeeCurrencyWhitelist.contract.Call(opts, &out, "whitelist", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Whitelist is a free data retrieval call binding the contract method 0x7ebd1b30.
//
// Solidity: function whitelist(uint256 ) view returns(address)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistSession) Whitelist(arg0 *big.Int) (common.Address, error) {
	return _FeeCurrencyWhitelist.Contract.Whitelist(&_FeeCurrencyWhitelist.CallOpts, arg0)
}

// Whitelist is a free data retrieval call binding the contract method 0x7ebd1b30.
//
// Solidity: function whitelist(uint256 ) view returns(address)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistCallerSession) Whitelist(arg0 *big.Int) (common.Address, error) {
	return _FeeCurrencyWhitelist.Contract.Whitelist(&_FeeCurrencyWhitelist.CallOpts, arg0)
}

// AddToken is a paid mutator transaction binding the contract method 0xd48bfca7.
//
// Solidity: function addToken(address tokenAddress) returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactor) AddToken(opts *bind.TransactOpts, tokenAddress common.Address) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.contract.Transact(opts, "addToken", tokenAddress)
}

// AddToken is a paid mutator transaction binding the contract method 0xd48bfca7.
//
// Solidity: function addToken(address tokenAddress) returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistSession) AddToken(tokenAddress common.Address) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.AddToken(&_FeeCurrencyWhitelist.TransactOpts, tokenAddress)
}

// AddToken is a paid mutator transaction binding the contract method 0xd48bfca7.
//
// Solidity: function addToken(address tokenAddress) returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactorSession) AddToken(tokenAddress common.Address) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.AddToken(&_FeeCurrencyWhitelist.TransactOpts, tokenAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistSession) Initialize() (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.Initialize(&_FeeCurrencyWhitelist.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactorSession) Initialize() (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.Initialize(&_FeeCurrencyWhitelist.TransactOpts)
}

// RemoveToken is a paid mutator transaction binding the contract method 0x13baf1e6.
//
// Solidity: function removeToken(address tokenAddress, uint256 index) returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactor) RemoveToken(opts *bind.TransactOpts, tokenAddress common.Address, index *big.Int) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.contract.Transact(opts, "removeToken", tokenAddress, index)
}

// RemoveToken is a paid mutator transaction binding the contract method 0x13baf1e6.
//
// Solidity: function removeToken(address tokenAddress, uint256 index) returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistSession) RemoveToken(tokenAddress common.Address, index *big.Int) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.RemoveToken(&_FeeCurrencyWhitelist.TransactOpts, tokenAddress, index)
}

// RemoveToken is a paid mutator transaction binding the contract method 0x13baf1e6.
//
// Solidity: function removeToken(address tokenAddress, uint256 index) returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactorSession) RemoveToken(tokenAddress common.Address, index *big.Int) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.RemoveToken(&_FeeCurrencyWhitelist.TransactOpts, tokenAddress, index)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistSession) RenounceOwnership() (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.RenounceOwnership(&_FeeCurrencyWhitelist.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.RenounceOwnership(&_FeeCurrencyWhitelist.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.TransferOwnership(&_FeeCurrencyWhitelist.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _FeeCurrencyWhitelist.Contract.TransferOwnership(&_FeeCurrencyWhitelist.TransactOpts, newOwner)
}

// FeeCurrencyWhitelistFeeCurrencyWhitelistRemovedIterator is returned from FilterFeeCurrencyWhitelistRemoved and is used to iterate over the raw logs and unpacked data for FeeCurrencyWhitelistRemoved events raised by the FeeCurrencyWhitelist contract.
type FeeCurrencyWhitelistFeeCurrencyWhitelistRemovedIterator struct {
	Event *FeeCurrencyWhitelistFeeCurrencyWhitelistRemoved // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FeeCurrencyWhitelistFeeCurrencyWhitelistRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FeeCurrencyWhitelistFeeCurrencyWhitelistRemoved)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FeeCurrencyWhitelistFeeCurrencyWhitelistRemoved)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FeeCurrencyWhitelistFeeCurrencyWhitelistRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FeeCurrencyWhitelistFeeCurrencyWhitelistRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FeeCurrencyWhitelistFeeCurrencyWhitelistRemoved represents a FeeCurrencyWhitelistRemoved event raised by the FeeCurrencyWhitelist contract.
type FeeCurrencyWhitelistFeeCurrencyWhitelistRemoved struct {
	Token common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterFeeCurrencyWhitelistRemoved is a free log retrieval operation binding the contract event 0xc1f06ffbe5c19d22daa982fd4b3886ced05d83e2bfe7de3b67222728f5234e28.
//
// Solidity: event FeeCurrencyWhitelistRemoved(address token)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistFilterer) FilterFeeCurrencyWhitelistRemoved(opts *bind.FilterOpts) (*FeeCurrencyWhitelistFeeCurrencyWhitelistRemovedIterator, error) {

	logs, sub, err := _FeeCurrencyWhitelist.contract.FilterLogs(opts, "FeeCurrencyWhitelistRemoved")
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyWhitelistFeeCurrencyWhitelistRemovedIterator{contract: _FeeCurrencyWhitelist.contract, event: "FeeCurrencyWhitelistRemoved", logs: logs, sub: sub}, nil
}

// WatchFeeCurrencyWhitelistRemoved is a free log subscription operation binding the contract event 0xc1f06ffbe5c19d22daa982fd4b3886ced05d83e2bfe7de3b67222728f5234e28.
//
// Solidity: event FeeCurrencyWhitelistRemoved(address token)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistFilterer) WatchFeeCurrencyWhitelistRemoved(opts *bind.WatchOpts, sink chan<- *FeeCurrencyWhitelistFeeCurrencyWhitelistRemoved) (event.Subscription, error) {

	logs, sub, err := _FeeCurrencyWhitelist.contract.WatchLogs(opts, "FeeCurrencyWhitelistRemoved")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FeeCurrencyWhitelistFeeCurrencyWhitelistRemoved)
				if err := _FeeCurrencyWhitelist.contract.UnpackLog(event, "FeeCurrencyWhitelistRemoved", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFeeCurrencyWhitelistRemoved is a log parse operation binding the contract event 0xc1f06ffbe5c19d22daa982fd4b3886ced05d83e2bfe7de3b67222728f5234e28.
//
// Solidity: event FeeCurrencyWhitelistRemoved(address token)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistFilterer) ParseFeeCurrencyWhitelistRemoved(log types.Log) (*FeeCurrencyWhitelistFeeCurrencyWhitelistRemoved, error) {
	event := new(FeeCurrencyWhitelistFeeCurrencyWhitelistRemoved)
	if err := _FeeCurrencyWhitelist.contract.UnpackLog(event, "FeeCurrencyWhitelistRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FeeCurrencyWhitelistFeeCurrencyWhitelistedIterator is returned from FilterFeeCurrencyWhitelisted and is used to iterate over the raw logs and unpacked data for FeeCurrencyWhitelisted events raised by the FeeCurrencyWhitelist contract.
type FeeCurrencyWhitelistFeeCurrencyWhitelistedIterator struct {
	Event *FeeCurrencyWhitelistFeeCurrencyWhitelisted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FeeCurrencyWhitelistFeeCurrencyWhitelistedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FeeCurrencyWhitelistFeeCurrencyWhitelisted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FeeCurrencyWhitelistFeeCurrencyWhitelisted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FeeCurrencyWhitelistFeeCurrencyWhitelistedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FeeCurrencyWhitelistFeeCurrencyWhitelistedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FeeCurrencyWhitelistFeeCurrencyWhitelisted represents a FeeCurrencyWhitelisted event raised by the FeeCurrencyWhitelist contract.
type FeeCurrencyWhitelistFeeCurrencyWhitelisted struct {
	Token common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterFeeCurrencyWhitelisted is a free log retrieval operation binding the contract event 0xcf4fe1d1989a39011040b0c33ba1165fdeeca971a1ab2b0340b23550f93727e0.
//
// Solidity: event FeeCurrencyWhitelisted(address token)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistFilterer) FilterFeeCurrencyWhitelisted(opts *bind.FilterOpts) (*FeeCurrencyWhitelistFeeCurrencyWhitelistedIterator, error) {

	logs, sub, err := _FeeCurrencyWhitelist.contract.FilterLogs(opts, "FeeCurrencyWhitelisted")
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyWhitelistFeeCurrencyWhitelistedIterator{contract: _FeeCurrencyWhitelist.contract, event: "FeeCurrencyWhitelisted", logs: logs, sub: sub}, nil
}

// WatchFeeCurrencyWhitelisted is a free log subscription operation binding the contract event 0xcf4fe1d1989a39011040b0c33ba1165fdeeca971a1ab2b0340b23550f93727e0.
//
// Solidity: event FeeCurrencyWhitelisted(address token)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistFilterer) WatchFeeCurrencyWhitelisted(opts *bind.WatchOpts, sink chan<- *FeeCurrencyWhitelistFeeCurrencyWhitelisted) (event.Subscription, error) {

	logs, sub, err := _FeeCurrencyWhitelist.contract.WatchLogs(opts, "FeeCurrencyWhitelisted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FeeCurrencyWhitelistFeeCurrencyWhitelisted)
				if err := _FeeCurrencyWhitelist.contract.UnpackLog(event, "FeeCurrencyWhitelisted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFeeCurrencyWhitelisted is a log parse operation binding the contract event 0xcf4fe1d1989a39011040b0c33ba1165fdeeca971a1ab2b0340b23550f93727e0.
//
// Solidity: event FeeCurrencyWhitelisted(address token)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistFilterer) ParseFeeCurrencyWhitelisted(log types.Log) (*FeeCurrencyWhitelistFeeCurrencyWhitelisted, error) {
	event := new(FeeCurrencyWhitelistFeeCurrencyWhitelisted)
	if err := _FeeCurrencyWhitelist.contract.UnpackLog(event, "FeeCurrencyWhitelisted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FeeCurrencyWhitelistOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the FeeCurrencyWhitelist contract.
type FeeCurrencyWhitelistOwnershipTransferredIterator struct {
	Event *FeeCurrencyWhitelistOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FeeCurrencyWhitelistOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FeeCurrencyWhitelistOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FeeCurrencyWhitelistOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FeeCurrencyWhitelistOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FeeCurrencyWhitelistOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FeeCurrencyWhitelistOwnershipTransferred represents a OwnershipTransferred event raised by the FeeCurrencyWhitelist contract.
type FeeCurrencyWhitelistOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*FeeCurrencyWhitelistOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _FeeCurrencyWhitelist.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyWhitelistOwnershipTransferredIterator{contract: _FeeCurrencyWhitelist.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *FeeCurrencyWhitelistOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _FeeCurrencyWhitelist.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FeeCurrencyWhitelistOwnershipTransferred)
				if err := _FeeCurrencyWhitelist.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_FeeCurrencyWhitelist *FeeCurrencyWhitelistFilterer) ParseOwnershipTransferred(log types.Log) (*FeeCurrencyWhitelistOwnershipTransferred, error) {
	event := new(FeeCurrencyWhitelistOwnershipTransferred)
	if err := _FeeCurrencyWhitelist.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
