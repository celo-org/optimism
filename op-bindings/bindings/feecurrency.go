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

// FeeCurrencyMetaData contains all meta data concerning the FeeCurrency contract.
var FeeCurrencyMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"name_\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbol_\",\"type\":\"string\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"feeRecipient\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"communityFund\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"refund\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"tipTxFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"baseTxFee\",\"type\":\"uint256\"}],\"name\":\"creditGasFees\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"debitGasFees\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040523480156200001157600080fd5b50604051620011813803806200118183398101604081905262000034916200011f565b600362000042838262000218565b50600462000051828262000218565b505050620002e4565b634e487b7160e01b600052604160045260246000fd5b600082601f8301126200008257600080fd5b81516001600160401b03808211156200009f576200009f6200005a565b604051601f8301601f19908116603f01168101908282118183101715620000ca57620000ca6200005a565b81604052838152602092508683858801011115620000e757600080fd5b600091505b838210156200010b5785820183015181830184015290820190620000ec565b600093810190920192909252949350505050565b600080604083850312156200013357600080fd5b82516001600160401b03808211156200014b57600080fd5b620001598683870162000070565b935060208501519150808211156200017057600080fd5b506200017f8582860162000070565b9150509250929050565b600181811c908216806200019e57607f821691505b602082108103620001bf57634e487b7160e01b600052602260045260246000fd5b50919050565b601f8211156200021357600081815260208120601f850160051c81016020861015620001ee5750805b601f850160051c820191505b818110156200020f57828155600101620001fa565b5050505b505050565b81516001600160401b038111156200023457620002346200005a565b6200024c8162000245845462000189565b84620001c5565b602080601f8311600181146200028457600084156200026b5750858301515b600019600386901b1c1916600185901b1785556200020f565b600085815260208120601f198616915b82811015620002b55788860151825594840194600190910190840162000294565b5085821015620002d45787850151600019600388901b60f8161c191681555b5050505050600190811b01905550565b610e8d80620002f46000396000f3fe608060405234801561001057600080fd5b50600436106100df5760003560e01c806358cf96721161008c57806395d89b411161006657806395d89b41146101ca578063a457c2d7146101d2578063a9059cbb146101e5578063dd62ed3e146101f857600080fd5b806358cf96721461016c5780636a30b2531461018157806370a082311461019457600080fd5b806323b872dd116100bd57806323b872dd14610137578063313ce5671461014a578063395093511461015957600080fd5b806306fdde03146100e4578063095ea7b31461010257806318160ddd14610125575b600080fd5b6100ec61023e565b6040516100f99190610c17565b60405180910390f35b610115610110366004610cac565b6102d0565b60405190151581526020016100f9565b6002545b6040519081526020016100f9565b610115610145366004610cd6565b6102ea565b604051601281526020016100f9565b610115610167366004610cac565b610310565b61017f61017a366004610cac565b61035c565b005b61017f61018f366004610d12565b610420565b6101296101a2366004610d8a565b73ffffffffffffffffffffffffffffffffffffffff1660009081526020819052604090205490565b6100ec610512565b6101156101e0366004610cac565b610521565b6101156101f3366004610cac565b6105fd565b610129610206366004610da5565b73ffffffffffffffffffffffffffffffffffffffff918216600090815260016020908152604080832093909416825291909152205490565b60606003805461024d90610dd8565b80601f016020809104026020016040519081016040528092919081815260200182805461027990610dd8565b80156102c65780601f1061029b576101008083540402835291602001916102c6565b820191906000526020600020905b8154815290600101906020018083116102a957829003601f168201915b5050505050905090565b6000336102de81858561060b565b60019150505b92915050565b6000336102f88582856107be565b610303858585610895565b60019150505b9392505050565b33600081815260016020908152604080832073ffffffffffffffffffffffffffffffffffffffff871684529091528120549091906102de9082908690610357908790610e5a565b61060b565b33156103c9576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c0000000000000000000000000000000060448201526064015b60405180910390fd5b73ffffffffffffffffffffffffffffffffffffffff8216600090815260208190526040812080548392906103fe908490610e6d565b9250508190555080600260008282546104179190610e6d565b90915550505050565b3315610488576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c0000000000000000000000000000000060448201526064016103c0565b73ffffffffffffffffffffffffffffffffffffffff8816600090815260208190526040812080548692906104bd908490610e5a565b909155506104ce9050888683610b48565b6104d89085610e5a565b93506104e5888885610b48565b6104ef9085610e5a565b935083600260008282546105039190610e5a565b90915550505050505050505050565b60606004805461024d90610dd8565b33600081815260016020908152604080832073ffffffffffffffffffffffffffffffffffffffff87168452909152812054909190838110156105e5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602560248201527f45524332303a2064656372656173656420616c6c6f77616e63652062656c6f7760448201527f207a65726f00000000000000000000000000000000000000000000000000000060648201526084016103c0565b6105f2828686840361060b565b506001949350505050565b6000336102de818585610895565b73ffffffffffffffffffffffffffffffffffffffff83166106ad576040517f08c379a0000000000000000000000000000000000000000000000000000000008152602060048201526024808201527f45524332303a20617070726f76652066726f6d20746865207a65726f2061646460448201527f726573730000000000000000000000000000000000000000000000000000000060648201526084016103c0565b73ffffffffffffffffffffffffffffffffffffffff8216610750576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f45524332303a20617070726f766520746f20746865207a65726f20616464726560448201527f737300000000000000000000000000000000000000000000000000000000000060648201526084016103c0565b73ffffffffffffffffffffffffffffffffffffffff83811660008181526001602090815260408083209487168084529482529182902085905590518481527f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925910160405180910390a3505050565b73ffffffffffffffffffffffffffffffffffffffff8381166000908152600160209081526040808320938616835292905220547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff811461088f5781811015610882576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f45524332303a20696e73756666696369656e7420616c6c6f77616e636500000060448201526064016103c0565b61088f848484840361060b565b50505050565b73ffffffffffffffffffffffffffffffffffffffff8316610938576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602560248201527f45524332303a207472616e736665722066726f6d20746865207a65726f20616460448201527f647265737300000000000000000000000000000000000000000000000000000060648201526084016103c0565b73ffffffffffffffffffffffffffffffffffffffff82166109db576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f45524332303a207472616e7366657220746f20746865207a65726f206164647260448201527f657373000000000000000000000000000000000000000000000000000000000060648201526084016103c0565b73ffffffffffffffffffffffffffffffffffffffff831660009081526020819052604090205481811015610a91576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f45524332303a207472616e7366657220616d6f756e742065786365656473206260448201527f616c616e6365000000000000000000000000000000000000000000000000000060648201526084016103c0565b73ffffffffffffffffffffffffffffffffffffffff808516600090815260208190526040808220858503905591851681529081208054849290610ad5908490610e5a565b925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef84604051610b3b91815260200190565b60405180910390a361088f565b600073ffffffffffffffffffffffffffffffffffffffff8316610b6d57506000610309565b73ffffffffffffffffffffffffffffffffffffffff831660009081526020819052604081208054849290610ba2908490610e5a565b925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef84604051610c0891815260200190565b60405180910390a35092915050565b600060208083528351808285015260005b81811015610c4457858101830151858201604001528201610c28565b5060006040828601015260407fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f8301168501019250505092915050565b803573ffffffffffffffffffffffffffffffffffffffff81168114610ca757600080fd5b919050565b60008060408385031215610cbf57600080fd5b610cc883610c83565b946020939093013593505050565b600080600060608486031215610ceb57600080fd5b610cf484610c83565b9250610d0260208501610c83565b9150604084013590509250925092565b600080600080600080600080610100898b031215610d2f57600080fd5b610d3889610c83565b9750610d4660208a01610c83565b9650610d5460408a01610c83565b9550610d6260608a01610c83565b979a969950949760808101359660a0820135965060c0820135955060e0909101359350915050565b600060208284031215610d9c57600080fd5b61030982610c83565b60008060408385031215610db857600080fd5b610dc183610c83565b9150610dcf60208401610c83565b90509250929050565b600181811c90821680610dec57607f821691505b602082108103610e25577f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b808201808211156102e4576102e4610e2b565b818103818111156102e4576102e4610e2b56fea164736f6c6343000813000a",
}

// FeeCurrencyABI is the input ABI used to generate the binding from.
// Deprecated: Use FeeCurrencyMetaData.ABI instead.
var FeeCurrencyABI = FeeCurrencyMetaData.ABI

// FeeCurrencyBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use FeeCurrencyMetaData.Bin instead.
var FeeCurrencyBin = FeeCurrencyMetaData.Bin

// DeployFeeCurrency deploys a new Ethereum contract, binding an instance of FeeCurrency to it.
func DeployFeeCurrency(auth *bind.TransactOpts, backend bind.ContractBackend, name_ string, symbol_ string) (common.Address, *types.Transaction, *FeeCurrency, error) {
	parsed, err := FeeCurrencyMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(FeeCurrencyBin), backend, name_, symbol_)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &FeeCurrency{FeeCurrencyCaller: FeeCurrencyCaller{contract: contract}, FeeCurrencyTransactor: FeeCurrencyTransactor{contract: contract}, FeeCurrencyFilterer: FeeCurrencyFilterer{contract: contract}}, nil
}

// FeeCurrency is an auto generated Go binding around an Ethereum contract.
type FeeCurrency struct {
	FeeCurrencyCaller     // Read-only binding to the contract
	FeeCurrencyTransactor // Write-only binding to the contract
	FeeCurrencyFilterer   // Log filterer for contract events
}

// FeeCurrencyCaller is an auto generated read-only Go binding around an Ethereum contract.
type FeeCurrencyCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FeeCurrencyTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FeeCurrencyTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FeeCurrencyFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FeeCurrencyFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FeeCurrencySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FeeCurrencySession struct {
	Contract     *FeeCurrency      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FeeCurrencyCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FeeCurrencyCallerSession struct {
	Contract *FeeCurrencyCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// FeeCurrencyTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FeeCurrencyTransactorSession struct {
	Contract     *FeeCurrencyTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// FeeCurrencyRaw is an auto generated low-level Go binding around an Ethereum contract.
type FeeCurrencyRaw struct {
	Contract *FeeCurrency // Generic contract binding to access the raw methods on
}

// FeeCurrencyCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FeeCurrencyCallerRaw struct {
	Contract *FeeCurrencyCaller // Generic read-only contract binding to access the raw methods on
}

// FeeCurrencyTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FeeCurrencyTransactorRaw struct {
	Contract *FeeCurrencyTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFeeCurrency creates a new instance of FeeCurrency, bound to a specific deployed contract.
func NewFeeCurrency(address common.Address, backend bind.ContractBackend) (*FeeCurrency, error) {
	contract, err := bindFeeCurrency(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FeeCurrency{FeeCurrencyCaller: FeeCurrencyCaller{contract: contract}, FeeCurrencyTransactor: FeeCurrencyTransactor{contract: contract}, FeeCurrencyFilterer: FeeCurrencyFilterer{contract: contract}}, nil
}

// NewFeeCurrencyCaller creates a new read-only instance of FeeCurrency, bound to a specific deployed contract.
func NewFeeCurrencyCaller(address common.Address, caller bind.ContractCaller) (*FeeCurrencyCaller, error) {
	contract, err := bindFeeCurrency(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyCaller{contract: contract}, nil
}

// NewFeeCurrencyTransactor creates a new write-only instance of FeeCurrency, bound to a specific deployed contract.
func NewFeeCurrencyTransactor(address common.Address, transactor bind.ContractTransactor) (*FeeCurrencyTransactor, error) {
	contract, err := bindFeeCurrency(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyTransactor{contract: contract}, nil
}

// NewFeeCurrencyFilterer creates a new log filterer instance of FeeCurrency, bound to a specific deployed contract.
func NewFeeCurrencyFilterer(address common.Address, filterer bind.ContractFilterer) (*FeeCurrencyFilterer, error) {
	contract, err := bindFeeCurrency(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyFilterer{contract: contract}, nil
}

// bindFeeCurrency binds a generic wrapper to an already deployed contract.
func bindFeeCurrency(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(FeeCurrencyABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FeeCurrency *FeeCurrencyRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FeeCurrency.Contract.FeeCurrencyCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FeeCurrency *FeeCurrencyRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FeeCurrency.Contract.FeeCurrencyTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FeeCurrency *FeeCurrencyRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FeeCurrency.Contract.FeeCurrencyTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FeeCurrency *FeeCurrencyCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FeeCurrency.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FeeCurrency *FeeCurrencyTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FeeCurrency.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FeeCurrency *FeeCurrencyTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FeeCurrency.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_FeeCurrency *FeeCurrencyCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _FeeCurrency.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_FeeCurrency *FeeCurrencySession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _FeeCurrency.Contract.Allowance(&_FeeCurrency.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_FeeCurrency *FeeCurrencyCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _FeeCurrency.Contract.Allowance(&_FeeCurrency.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_FeeCurrency *FeeCurrencyCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _FeeCurrency.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_FeeCurrency *FeeCurrencySession) BalanceOf(account common.Address) (*big.Int, error) {
	return _FeeCurrency.Contract.BalanceOf(&_FeeCurrency.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_FeeCurrency *FeeCurrencyCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _FeeCurrency.Contract.BalanceOf(&_FeeCurrency.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_FeeCurrency *FeeCurrencyCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _FeeCurrency.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_FeeCurrency *FeeCurrencySession) Decimals() (uint8, error) {
	return _FeeCurrency.Contract.Decimals(&_FeeCurrency.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_FeeCurrency *FeeCurrencyCallerSession) Decimals() (uint8, error) {
	return _FeeCurrency.Contract.Decimals(&_FeeCurrency.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_FeeCurrency *FeeCurrencyCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _FeeCurrency.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_FeeCurrency *FeeCurrencySession) Name() (string, error) {
	return _FeeCurrency.Contract.Name(&_FeeCurrency.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_FeeCurrency *FeeCurrencyCallerSession) Name() (string, error) {
	return _FeeCurrency.Contract.Name(&_FeeCurrency.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_FeeCurrency *FeeCurrencyCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _FeeCurrency.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_FeeCurrency *FeeCurrencySession) Symbol() (string, error) {
	return _FeeCurrency.Contract.Symbol(&_FeeCurrency.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_FeeCurrency *FeeCurrencyCallerSession) Symbol() (string, error) {
	return _FeeCurrency.Contract.Symbol(&_FeeCurrency.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_FeeCurrency *FeeCurrencyCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _FeeCurrency.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_FeeCurrency *FeeCurrencySession) TotalSupply() (*big.Int, error) {
	return _FeeCurrency.Contract.TotalSupply(&_FeeCurrency.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_FeeCurrency *FeeCurrencyCallerSession) TotalSupply() (*big.Int, error) {
	return _FeeCurrency.Contract.TotalSupply(&_FeeCurrency.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_FeeCurrency *FeeCurrencyTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_FeeCurrency *FeeCurrencySession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.Approve(&_FeeCurrency.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_FeeCurrency *FeeCurrencyTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.Approve(&_FeeCurrency.TransactOpts, spender, amount)
}

// CreditGasFees is a paid mutator transaction binding the contract method 0x6a30b253.
//
// Solidity: function creditGasFees(address from, address feeRecipient, address , address communityFund, uint256 refund, uint256 tipTxFee, uint256 , uint256 baseTxFee) returns()
func (_FeeCurrency *FeeCurrencyTransactor) CreditGasFees(opts *bind.TransactOpts, from common.Address, feeRecipient common.Address, arg2 common.Address, communityFund common.Address, refund *big.Int, tipTxFee *big.Int, arg6 *big.Int, baseTxFee *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.contract.Transact(opts, "creditGasFees", from, feeRecipient, arg2, communityFund, refund, tipTxFee, arg6, baseTxFee)
}

// CreditGasFees is a paid mutator transaction binding the contract method 0x6a30b253.
//
// Solidity: function creditGasFees(address from, address feeRecipient, address , address communityFund, uint256 refund, uint256 tipTxFee, uint256 , uint256 baseTxFee) returns()
func (_FeeCurrency *FeeCurrencySession) CreditGasFees(from common.Address, feeRecipient common.Address, arg2 common.Address, communityFund common.Address, refund *big.Int, tipTxFee *big.Int, arg6 *big.Int, baseTxFee *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.CreditGasFees(&_FeeCurrency.TransactOpts, from, feeRecipient, arg2, communityFund, refund, tipTxFee, arg6, baseTxFee)
}

// CreditGasFees is a paid mutator transaction binding the contract method 0x6a30b253.
//
// Solidity: function creditGasFees(address from, address feeRecipient, address , address communityFund, uint256 refund, uint256 tipTxFee, uint256 , uint256 baseTxFee) returns()
func (_FeeCurrency *FeeCurrencyTransactorSession) CreditGasFees(from common.Address, feeRecipient common.Address, arg2 common.Address, communityFund common.Address, refund *big.Int, tipTxFee *big.Int, arg6 *big.Int, baseTxFee *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.CreditGasFees(&_FeeCurrency.TransactOpts, from, feeRecipient, arg2, communityFund, refund, tipTxFee, arg6, baseTxFee)
}

// DebitGasFees is a paid mutator transaction binding the contract method 0x58cf9672.
//
// Solidity: function debitGasFees(address from, uint256 value) returns()
func (_FeeCurrency *FeeCurrencyTransactor) DebitGasFees(opts *bind.TransactOpts, from common.Address, value *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.contract.Transact(opts, "debitGasFees", from, value)
}

// DebitGasFees is a paid mutator transaction binding the contract method 0x58cf9672.
//
// Solidity: function debitGasFees(address from, uint256 value) returns()
func (_FeeCurrency *FeeCurrencySession) DebitGasFees(from common.Address, value *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.DebitGasFees(&_FeeCurrency.TransactOpts, from, value)
}

// DebitGasFees is a paid mutator transaction binding the contract method 0x58cf9672.
//
// Solidity: function debitGasFees(address from, uint256 value) returns()
func (_FeeCurrency *FeeCurrencyTransactorSession) DebitGasFees(from common.Address, value *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.DebitGasFees(&_FeeCurrency.TransactOpts, from, value)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_FeeCurrency *FeeCurrencyTransactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_FeeCurrency *FeeCurrencySession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.DecreaseAllowance(&_FeeCurrency.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_FeeCurrency *FeeCurrencyTransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.DecreaseAllowance(&_FeeCurrency.TransactOpts, spender, subtractedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_FeeCurrency *FeeCurrencyTransactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_FeeCurrency *FeeCurrencySession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.IncreaseAllowance(&_FeeCurrency.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_FeeCurrency *FeeCurrencyTransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.IncreaseAllowance(&_FeeCurrency.TransactOpts, spender, addedValue)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_FeeCurrency *FeeCurrencyTransactor) Transfer(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.contract.Transact(opts, "transfer", to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_FeeCurrency *FeeCurrencySession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.Transfer(&_FeeCurrency.TransactOpts, to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_FeeCurrency *FeeCurrencyTransactorSession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.Transfer(&_FeeCurrency.TransactOpts, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_FeeCurrency *FeeCurrencyTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.contract.Transact(opts, "transferFrom", from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_FeeCurrency *FeeCurrencySession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.TransferFrom(&_FeeCurrency.TransactOpts, from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_FeeCurrency *FeeCurrencyTransactorSession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FeeCurrency.Contract.TransferFrom(&_FeeCurrency.TransactOpts, from, to, amount)
}

// FeeCurrencyApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the FeeCurrency contract.
type FeeCurrencyApprovalIterator struct {
	Event *FeeCurrencyApproval // Event containing the contract specifics and raw log

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
func (it *FeeCurrencyApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FeeCurrencyApproval)
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
		it.Event = new(FeeCurrencyApproval)
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
func (it *FeeCurrencyApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FeeCurrencyApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FeeCurrencyApproval represents a Approval event raised by the FeeCurrency contract.
type FeeCurrencyApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_FeeCurrency *FeeCurrencyFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*FeeCurrencyApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _FeeCurrency.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyApprovalIterator{contract: _FeeCurrency.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_FeeCurrency *FeeCurrencyFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *FeeCurrencyApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _FeeCurrency.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FeeCurrencyApproval)
				if err := _FeeCurrency.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_FeeCurrency *FeeCurrencyFilterer) ParseApproval(log types.Log) (*FeeCurrencyApproval, error) {
	event := new(FeeCurrencyApproval)
	if err := _FeeCurrency.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FeeCurrencyTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the FeeCurrency contract.
type FeeCurrencyTransferIterator struct {
	Event *FeeCurrencyTransfer // Event containing the contract specifics and raw log

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
func (it *FeeCurrencyTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FeeCurrencyTransfer)
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
		it.Event = new(FeeCurrencyTransfer)
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
func (it *FeeCurrencyTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FeeCurrencyTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FeeCurrencyTransfer represents a Transfer event raised by the FeeCurrency contract.
type FeeCurrencyTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_FeeCurrency *FeeCurrencyFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*FeeCurrencyTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _FeeCurrency.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &FeeCurrencyTransferIterator{contract: _FeeCurrency.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_FeeCurrency *FeeCurrencyFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *FeeCurrencyTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _FeeCurrency.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FeeCurrencyTransfer)
				if err := _FeeCurrency.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_FeeCurrency *FeeCurrencyFilterer) ParseTransfer(log types.Log) (*FeeCurrencyTransfer, error) {
	event := new(FeeCurrencyTransfer)
	if err := _FeeCurrency.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
