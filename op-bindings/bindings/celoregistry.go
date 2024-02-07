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

// CeloRegistryMetaData contains all meta data concerning the CeloRegistry contract.
var CeloRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"test\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getAddressFor\",\"inputs\":[{\"name\":\"identifierHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAddressForOrDie\",\"inputs\":[{\"name\":\"identifierHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAddressForString\",\"inputs\":[{\"name\":\"identifier\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAddressForStringOrDie\",\"inputs\":[{\"name\":\"identifier\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialized\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOneOf\",\"inputs\":[{\"name\":\"identifierHashes\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registry\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAddressFor\",\"inputs\":[{\"name\":\"identifier\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RegistryUpdated\",\"inputs\":[{\"name\":\"identifier\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"identifierHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"addr\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false}]",
	Bin: "0x608060405234801561001057600080fd5b50604051610c13380380610c1383398101604081905261002f916100a9565b8061003933610059565b80610052576000805460ff60a01b1916600160a01b1790555b50506100d2565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b6000602082840312156100bb57600080fd5b815180151581146100cb57600080fd5b9392505050565b610b32806100e16000396000f3fe608060405234801561001057600080fd5b50600436106100d45760003560e01c80638932cbf411610081578063dcf0aaed1161005b578063dcf0aaed146101ea578063dd927233146101fd578063f2fde38b1461023357600080fd5b80638932cbf4146101a65780638da5cb5b146101b9578063c5865793146101d757600080fd5b80637ef50298116100b25780637ef50298146101305780638129fc1c1461018b578063853db3231461019357600080fd5b8063158ef93e146100d957806317c5081814610113578063715018a614610126575b600080fd5b6000546100fe9074010000000000000000000000000000000000000000900460ff1681565b60405190151581526020015b60405180910390f35b6100fe6101213660046108ac565b610246565b61012e6102dd565b005b61016661013e366004610930565b60016020526000908152604090205473ffffffffffffffffffffffffffffffffffffffff1681565b60405173ffffffffffffffffffffffffffffffffffffffff909116815260200161010a565b61012e6102f1565b6101666101a1366004610992565b6103c3565b6101666101b4366004610992565b61043a565b60005473ffffffffffffffffffffffffffffffffffffffff16610166565b61012e6101e53660046109d4565b61053e565b6101666101f8366004610930565b610622565b61016661020b366004610930565b60009081526001602052604090205473ffffffffffffffffffffffffffffffffffffffff1690565b61012e610241366004610a1f565b6106d6565b6000805b838110156102d0578273ffffffffffffffffffffffffffffffffffffffff166001600087878581811061027f5761027f610a3a565b602090810292909201358352508101919091526040016000205473ffffffffffffffffffffffffffffffffffffffff16036102be5760019150506102d6565b806102c881610a69565b91505061024a565b50600090505b9392505050565b6102e561078d565b6102ef600061080e565b565b60005474010000000000000000000000000000000000000000900460ff161561037b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601c60248201527f636f6e747261637420616c726561647920696e697469616c697a65640000000060448201526064015b60405180910390fd5b600080547fffffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffff16740100000000000000000000000000000000000000001790556102ef3361080e565b60008083836040516020016103d9929190610ac8565b604080518083037fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe001815291815281516020928301206000908152600190925290205473ffffffffffffffffffffffffffffffffffffffff16949350505050565b6000808383604051602001610450929190610ac8565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe081840301815291815281516020928301206000818152600190935291205490915073ffffffffffffffffffffffffffffffffffffffff16610512576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f6964656e74696669657220686173206e6f20726567697374727920656e7472796044820152606401610372565b60009081526001602052604090205473ffffffffffffffffffffffffffffffffffffffff169392505050565b61054661078d565b6000838360405160200161055b929190610ac8565b604080518083037fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe001815282825280516020918201206000818152600190925291902080547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff86169081179091559092509082907f4166d073a7a5e704ce0db7113320f88da2457f872d46dc020c805c562c1582a0906106149088908890610ad8565b60405180910390a350505050565b60008181526001602052604081205473ffffffffffffffffffffffffffffffffffffffff166106ad576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f6964656e74696669657220686173206e6f20726567697374727920656e7472796044820152606401610372565b5060009081526001602052604090205473ffffffffffffffffffffffffffffffffffffffff1690565b6106de61078d565b73ffffffffffffffffffffffffffffffffffffffff8116610781576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f64647265737300000000000000000000000000000000000000000000000000006064820152608401610372565b61078a8161080e565b50565b60005473ffffffffffffffffffffffffffffffffffffffff1633146102ef576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e65726044820152606401610372565b6000805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b803573ffffffffffffffffffffffffffffffffffffffff811681146108a757600080fd5b919050565b6000806000604084860312156108c157600080fd5b833567ffffffffffffffff808211156108d957600080fd5b818601915086601f8301126108ed57600080fd5b8135818111156108fc57600080fd5b8760208260051b850101111561091157600080fd5b6020928301955093506109279186019050610883565b90509250925092565b60006020828403121561094257600080fd5b5035919050565b60008083601f84011261095b57600080fd5b50813567ffffffffffffffff81111561097357600080fd5b60208301915083602082850101111561098b57600080fd5b9250929050565b600080602083850312156109a557600080fd5b823567ffffffffffffffff8111156109bc57600080fd5b6109c885828601610949565b90969095509350505050565b6000806000604084860312156109e957600080fd5b833567ffffffffffffffff811115610a0057600080fd5b610a0c86828701610949565b9094509250610927905060208501610883565b600060208284031215610a3157600080fd5b6102d682610883565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203610ac1577f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b5060010190565b8183823760009101908152919050565b60208152816020820152818360408301376000818301604090810191909152601f9092017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016010191905056fea164736f6c6343000813000a",
}

// CeloRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use CeloRegistryMetaData.ABI instead.
var CeloRegistryABI = CeloRegistryMetaData.ABI

// CeloRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use CeloRegistryMetaData.Bin instead.
var CeloRegistryBin = CeloRegistryMetaData.Bin

// DeployCeloRegistry deploys a new Ethereum contract, binding an instance of CeloRegistry to it.
func DeployCeloRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, test bool) (common.Address, *types.Transaction, *CeloRegistry, error) {
	parsed, err := CeloRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(CeloRegistryBin), backend, test)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CeloRegistry{CeloRegistryCaller: CeloRegistryCaller{contract: contract}, CeloRegistryTransactor: CeloRegistryTransactor{contract: contract}, CeloRegistryFilterer: CeloRegistryFilterer{contract: contract}}, nil
}

// CeloRegistry is an auto generated Go binding around an Ethereum contract.
type CeloRegistry struct {
	CeloRegistryCaller     // Read-only binding to the contract
	CeloRegistryTransactor // Write-only binding to the contract
	CeloRegistryFilterer   // Log filterer for contract events
}

// CeloRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type CeloRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CeloRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CeloRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CeloRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CeloRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CeloRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CeloRegistrySession struct {
	Contract     *CeloRegistry     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CeloRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CeloRegistryCallerSession struct {
	Contract *CeloRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// CeloRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CeloRegistryTransactorSession struct {
	Contract     *CeloRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// CeloRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type CeloRegistryRaw struct {
	Contract *CeloRegistry // Generic contract binding to access the raw methods on
}

// CeloRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CeloRegistryCallerRaw struct {
	Contract *CeloRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// CeloRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CeloRegistryTransactorRaw struct {
	Contract *CeloRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCeloRegistry creates a new instance of CeloRegistry, bound to a specific deployed contract.
func NewCeloRegistry(address common.Address, backend bind.ContractBackend) (*CeloRegistry, error) {
	contract, err := bindCeloRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CeloRegistry{CeloRegistryCaller: CeloRegistryCaller{contract: contract}, CeloRegistryTransactor: CeloRegistryTransactor{contract: contract}, CeloRegistryFilterer: CeloRegistryFilterer{contract: contract}}, nil
}

// NewCeloRegistryCaller creates a new read-only instance of CeloRegistry, bound to a specific deployed contract.
func NewCeloRegistryCaller(address common.Address, caller bind.ContractCaller) (*CeloRegistryCaller, error) {
	contract, err := bindCeloRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CeloRegistryCaller{contract: contract}, nil
}

// NewCeloRegistryTransactor creates a new write-only instance of CeloRegistry, bound to a specific deployed contract.
func NewCeloRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*CeloRegistryTransactor, error) {
	contract, err := bindCeloRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CeloRegistryTransactor{contract: contract}, nil
}

// NewCeloRegistryFilterer creates a new log filterer instance of CeloRegistry, bound to a specific deployed contract.
func NewCeloRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*CeloRegistryFilterer, error) {
	contract, err := bindCeloRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CeloRegistryFilterer{contract: contract}, nil
}

// bindCeloRegistry binds a generic wrapper to an already deployed contract.
func bindCeloRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(CeloRegistryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CeloRegistry *CeloRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CeloRegistry.Contract.CeloRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CeloRegistry *CeloRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CeloRegistry.Contract.CeloRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CeloRegistry *CeloRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CeloRegistry.Contract.CeloRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CeloRegistry *CeloRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CeloRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CeloRegistry *CeloRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CeloRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CeloRegistry *CeloRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CeloRegistry.Contract.contract.Transact(opts, method, params...)
}

// GetAddressFor is a free data retrieval call binding the contract method 0xdd927233.
//
// Solidity: function getAddressFor(bytes32 identifierHash) view returns(address)
func (_CeloRegistry *CeloRegistryCaller) GetAddressFor(opts *bind.CallOpts, identifierHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _CeloRegistry.contract.Call(opts, &out, "getAddressFor", identifierHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAddressFor is a free data retrieval call binding the contract method 0xdd927233.
//
// Solidity: function getAddressFor(bytes32 identifierHash) view returns(address)
func (_CeloRegistry *CeloRegistrySession) GetAddressFor(identifierHash [32]byte) (common.Address, error) {
	return _CeloRegistry.Contract.GetAddressFor(&_CeloRegistry.CallOpts, identifierHash)
}

// GetAddressFor is a free data retrieval call binding the contract method 0xdd927233.
//
// Solidity: function getAddressFor(bytes32 identifierHash) view returns(address)
func (_CeloRegistry *CeloRegistryCallerSession) GetAddressFor(identifierHash [32]byte) (common.Address, error) {
	return _CeloRegistry.Contract.GetAddressFor(&_CeloRegistry.CallOpts, identifierHash)
}

// GetAddressForOrDie is a free data retrieval call binding the contract method 0xdcf0aaed.
//
// Solidity: function getAddressForOrDie(bytes32 identifierHash) view returns(address)
func (_CeloRegistry *CeloRegistryCaller) GetAddressForOrDie(opts *bind.CallOpts, identifierHash [32]byte) (common.Address, error) {
	var out []interface{}
	err := _CeloRegistry.contract.Call(opts, &out, "getAddressForOrDie", identifierHash)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAddressForOrDie is a free data retrieval call binding the contract method 0xdcf0aaed.
//
// Solidity: function getAddressForOrDie(bytes32 identifierHash) view returns(address)
func (_CeloRegistry *CeloRegistrySession) GetAddressForOrDie(identifierHash [32]byte) (common.Address, error) {
	return _CeloRegistry.Contract.GetAddressForOrDie(&_CeloRegistry.CallOpts, identifierHash)
}

// GetAddressForOrDie is a free data retrieval call binding the contract method 0xdcf0aaed.
//
// Solidity: function getAddressForOrDie(bytes32 identifierHash) view returns(address)
func (_CeloRegistry *CeloRegistryCallerSession) GetAddressForOrDie(identifierHash [32]byte) (common.Address, error) {
	return _CeloRegistry.Contract.GetAddressForOrDie(&_CeloRegistry.CallOpts, identifierHash)
}

// GetAddressForString is a free data retrieval call binding the contract method 0x853db323.
//
// Solidity: function getAddressForString(string identifier) view returns(address)
func (_CeloRegistry *CeloRegistryCaller) GetAddressForString(opts *bind.CallOpts, identifier string) (common.Address, error) {
	var out []interface{}
	err := _CeloRegistry.contract.Call(opts, &out, "getAddressForString", identifier)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAddressForString is a free data retrieval call binding the contract method 0x853db323.
//
// Solidity: function getAddressForString(string identifier) view returns(address)
func (_CeloRegistry *CeloRegistrySession) GetAddressForString(identifier string) (common.Address, error) {
	return _CeloRegistry.Contract.GetAddressForString(&_CeloRegistry.CallOpts, identifier)
}

// GetAddressForString is a free data retrieval call binding the contract method 0x853db323.
//
// Solidity: function getAddressForString(string identifier) view returns(address)
func (_CeloRegistry *CeloRegistryCallerSession) GetAddressForString(identifier string) (common.Address, error) {
	return _CeloRegistry.Contract.GetAddressForString(&_CeloRegistry.CallOpts, identifier)
}

// GetAddressForStringOrDie is a free data retrieval call binding the contract method 0x8932cbf4.
//
// Solidity: function getAddressForStringOrDie(string identifier) view returns(address)
func (_CeloRegistry *CeloRegistryCaller) GetAddressForStringOrDie(opts *bind.CallOpts, identifier string) (common.Address, error) {
	var out []interface{}
	err := _CeloRegistry.contract.Call(opts, &out, "getAddressForStringOrDie", identifier)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAddressForStringOrDie is a free data retrieval call binding the contract method 0x8932cbf4.
//
// Solidity: function getAddressForStringOrDie(string identifier) view returns(address)
func (_CeloRegistry *CeloRegistrySession) GetAddressForStringOrDie(identifier string) (common.Address, error) {
	return _CeloRegistry.Contract.GetAddressForStringOrDie(&_CeloRegistry.CallOpts, identifier)
}

// GetAddressForStringOrDie is a free data retrieval call binding the contract method 0x8932cbf4.
//
// Solidity: function getAddressForStringOrDie(string identifier) view returns(address)
func (_CeloRegistry *CeloRegistryCallerSession) GetAddressForStringOrDie(identifier string) (common.Address, error) {
	return _CeloRegistry.Contract.GetAddressForStringOrDie(&_CeloRegistry.CallOpts, identifier)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_CeloRegistry *CeloRegistryCaller) Initialized(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _CeloRegistry.contract.Call(opts, &out, "initialized")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_CeloRegistry *CeloRegistrySession) Initialized() (bool, error) {
	return _CeloRegistry.Contract.Initialized(&_CeloRegistry.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_CeloRegistry *CeloRegistryCallerSession) Initialized() (bool, error) {
	return _CeloRegistry.Contract.Initialized(&_CeloRegistry.CallOpts)
}

// IsOneOf is a free data retrieval call binding the contract method 0x17c50818.
//
// Solidity: function isOneOf(bytes32[] identifierHashes, address sender) view returns(bool)
func (_CeloRegistry *CeloRegistryCaller) IsOneOf(opts *bind.CallOpts, identifierHashes [][32]byte, sender common.Address) (bool, error) {
	var out []interface{}
	err := _CeloRegistry.contract.Call(opts, &out, "isOneOf", identifierHashes, sender)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOneOf is a free data retrieval call binding the contract method 0x17c50818.
//
// Solidity: function isOneOf(bytes32[] identifierHashes, address sender) view returns(bool)
func (_CeloRegistry *CeloRegistrySession) IsOneOf(identifierHashes [][32]byte, sender common.Address) (bool, error) {
	return _CeloRegistry.Contract.IsOneOf(&_CeloRegistry.CallOpts, identifierHashes, sender)
}

// IsOneOf is a free data retrieval call binding the contract method 0x17c50818.
//
// Solidity: function isOneOf(bytes32[] identifierHashes, address sender) view returns(bool)
func (_CeloRegistry *CeloRegistryCallerSession) IsOneOf(identifierHashes [][32]byte, sender common.Address) (bool, error) {
	return _CeloRegistry.Contract.IsOneOf(&_CeloRegistry.CallOpts, identifierHashes, sender)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CeloRegistry *CeloRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CeloRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CeloRegistry *CeloRegistrySession) Owner() (common.Address, error) {
	return _CeloRegistry.Contract.Owner(&_CeloRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CeloRegistry *CeloRegistryCallerSession) Owner() (common.Address, error) {
	return _CeloRegistry.Contract.Owner(&_CeloRegistry.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7ef50298.
//
// Solidity: function registry(bytes32 ) view returns(address)
func (_CeloRegistry *CeloRegistryCaller) Registry(opts *bind.CallOpts, arg0 [32]byte) (common.Address, error) {
	var out []interface{}
	err := _CeloRegistry.contract.Call(opts, &out, "registry", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Registry is a free data retrieval call binding the contract method 0x7ef50298.
//
// Solidity: function registry(bytes32 ) view returns(address)
func (_CeloRegistry *CeloRegistrySession) Registry(arg0 [32]byte) (common.Address, error) {
	return _CeloRegistry.Contract.Registry(&_CeloRegistry.CallOpts, arg0)
}

// Registry is a free data retrieval call binding the contract method 0x7ef50298.
//
// Solidity: function registry(bytes32 ) view returns(address)
func (_CeloRegistry *CeloRegistryCallerSession) Registry(arg0 [32]byte) (common.Address, error) {
	return _CeloRegistry.Contract.Registry(&_CeloRegistry.CallOpts, arg0)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_CeloRegistry *CeloRegistryTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CeloRegistry.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_CeloRegistry *CeloRegistrySession) Initialize() (*types.Transaction, error) {
	return _CeloRegistry.Contract.Initialize(&_CeloRegistry.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_CeloRegistry *CeloRegistryTransactorSession) Initialize() (*types.Transaction, error) {
	return _CeloRegistry.Contract.Initialize(&_CeloRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CeloRegistry *CeloRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CeloRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CeloRegistry *CeloRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _CeloRegistry.Contract.RenounceOwnership(&_CeloRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CeloRegistry *CeloRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _CeloRegistry.Contract.RenounceOwnership(&_CeloRegistry.TransactOpts)
}

// SetAddressFor is a paid mutator transaction binding the contract method 0xc5865793.
//
// Solidity: function setAddressFor(string identifier, address addr) returns()
func (_CeloRegistry *CeloRegistryTransactor) SetAddressFor(opts *bind.TransactOpts, identifier string, addr common.Address) (*types.Transaction, error) {
	return _CeloRegistry.contract.Transact(opts, "setAddressFor", identifier, addr)
}

// SetAddressFor is a paid mutator transaction binding the contract method 0xc5865793.
//
// Solidity: function setAddressFor(string identifier, address addr) returns()
func (_CeloRegistry *CeloRegistrySession) SetAddressFor(identifier string, addr common.Address) (*types.Transaction, error) {
	return _CeloRegistry.Contract.SetAddressFor(&_CeloRegistry.TransactOpts, identifier, addr)
}

// SetAddressFor is a paid mutator transaction binding the contract method 0xc5865793.
//
// Solidity: function setAddressFor(string identifier, address addr) returns()
func (_CeloRegistry *CeloRegistryTransactorSession) SetAddressFor(identifier string, addr common.Address) (*types.Transaction, error) {
	return _CeloRegistry.Contract.SetAddressFor(&_CeloRegistry.TransactOpts, identifier, addr)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CeloRegistry *CeloRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _CeloRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CeloRegistry *CeloRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _CeloRegistry.Contract.TransferOwnership(&_CeloRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CeloRegistry *CeloRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _CeloRegistry.Contract.TransferOwnership(&_CeloRegistry.TransactOpts, newOwner)
}

// CeloRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the CeloRegistry contract.
type CeloRegistryOwnershipTransferredIterator struct {
	Event *CeloRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *CeloRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CeloRegistryOwnershipTransferred)
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
		it.Event = new(CeloRegistryOwnershipTransferred)
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
func (it *CeloRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CeloRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CeloRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the CeloRegistry contract.
type CeloRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CeloRegistry *CeloRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*CeloRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CeloRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &CeloRegistryOwnershipTransferredIterator{contract: _CeloRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CeloRegistry *CeloRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *CeloRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CeloRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CeloRegistryOwnershipTransferred)
				if err := _CeloRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_CeloRegistry *CeloRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*CeloRegistryOwnershipTransferred, error) {
	event := new(CeloRegistryOwnershipTransferred)
	if err := _CeloRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CeloRegistryRegistryUpdatedIterator is returned from FilterRegistryUpdated and is used to iterate over the raw logs and unpacked data for RegistryUpdated events raised by the CeloRegistry contract.
type CeloRegistryRegistryUpdatedIterator struct {
	Event *CeloRegistryRegistryUpdated // Event containing the contract specifics and raw log

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
func (it *CeloRegistryRegistryUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CeloRegistryRegistryUpdated)
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
		it.Event = new(CeloRegistryRegistryUpdated)
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
func (it *CeloRegistryRegistryUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CeloRegistryRegistryUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CeloRegistryRegistryUpdated represents a RegistryUpdated event raised by the CeloRegistry contract.
type CeloRegistryRegistryUpdated struct {
	Identifier     string
	IdentifierHash [32]byte
	Addr           common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterRegistryUpdated is a free log retrieval operation binding the contract event 0x4166d073a7a5e704ce0db7113320f88da2457f872d46dc020c805c562c1582a0.
//
// Solidity: event RegistryUpdated(string identifier, bytes32 indexed identifierHash, address indexed addr)
func (_CeloRegistry *CeloRegistryFilterer) FilterRegistryUpdated(opts *bind.FilterOpts, identifierHash [][32]byte, addr []common.Address) (*CeloRegistryRegistryUpdatedIterator, error) {

	var identifierHashRule []interface{}
	for _, identifierHashItem := range identifierHash {
		identifierHashRule = append(identifierHashRule, identifierHashItem)
	}
	var addrRule []interface{}
	for _, addrItem := range addr {
		addrRule = append(addrRule, addrItem)
	}

	logs, sub, err := _CeloRegistry.contract.FilterLogs(opts, "RegistryUpdated", identifierHashRule, addrRule)
	if err != nil {
		return nil, err
	}
	return &CeloRegistryRegistryUpdatedIterator{contract: _CeloRegistry.contract, event: "RegistryUpdated", logs: logs, sub: sub}, nil
}

// WatchRegistryUpdated is a free log subscription operation binding the contract event 0x4166d073a7a5e704ce0db7113320f88da2457f872d46dc020c805c562c1582a0.
//
// Solidity: event RegistryUpdated(string identifier, bytes32 indexed identifierHash, address indexed addr)
func (_CeloRegistry *CeloRegistryFilterer) WatchRegistryUpdated(opts *bind.WatchOpts, sink chan<- *CeloRegistryRegistryUpdated, identifierHash [][32]byte, addr []common.Address) (event.Subscription, error) {

	var identifierHashRule []interface{}
	for _, identifierHashItem := range identifierHash {
		identifierHashRule = append(identifierHashRule, identifierHashItem)
	}
	var addrRule []interface{}
	for _, addrItem := range addr {
		addrRule = append(addrRule, addrItem)
	}

	logs, sub, err := _CeloRegistry.contract.WatchLogs(opts, "RegistryUpdated", identifierHashRule, addrRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CeloRegistryRegistryUpdated)
				if err := _CeloRegistry.contract.UnpackLog(event, "RegistryUpdated", log); err != nil {
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

// ParseRegistryUpdated is a log parse operation binding the contract event 0x4166d073a7a5e704ce0db7113320f88da2457f872d46dc020c805c562c1582a0.
//
// Solidity: event RegistryUpdated(string identifier, bytes32 indexed identifierHash, address indexed addr)
func (_CeloRegistry *CeloRegistryFilterer) ParseRegistryUpdated(log types.Log) (*CeloRegistryRegistryUpdated, error) {
	event := new(CeloRegistryRegistryUpdated)
	if err := _CeloRegistry.contract.UnpackLog(event, "RegistryUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
