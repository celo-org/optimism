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

// LegacyERC20ETHMetaData contains all meta data concerning the LegacyERC20ETH contract.
var LegacyERC20ETHMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"BRIDGE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"REMOTE_TOKEN\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"allowance\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"approve\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"balanceOf\",\"inputs\":[{\"name\":\"_who\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"bridge\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"burn\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"creditGasFees\",\"inputs\":[{\"name\":\"recipients\",\"type\":\"address[]\",\"internalType\":\"address[]\"},{\"name\":\"amounts\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"creditGasFees\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"communityFund\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"refund\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"tipTxFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"baseTxFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"debitGasFees\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"decimals\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"decreaseAllowance\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"increaseAllowance\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"l1Token\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"l2Bridge\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"mint\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"name\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"remoteToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"_interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"symbol\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalSupply\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transfer\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferFrom\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"version\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"Approval\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"spender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Burn\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Mint\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Transfer\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false}]",
	Bin: "0x60e06040523480156200001157600080fd5b5073420000000000000000000000000000000000001060006040518060400160405280600581526020016422ba3432b960d91b8152506040518060400160405280600381526020016208aa8960eb1b81525060128282816003908162000078919062000152565b50600462000087828262000152565b5050506001600160a01b039384166080529390921660a052505060ff1660c0526200021e565b634e487b7160e01b600052604160045260246000fd5b600181811c90821680620000d857607f821691505b602082108103620000f957634e487b7160e01b600052602260045260246000fd5b50919050565b601f8211156200014d57600081815260208120601f850160051c81016020861015620001285750805b601f850160051c820191505b81811015620001495782815560010162000134565b5050505b505050565b81516001600160401b038111156200016e576200016e620000ad565b62000186816200017f8454620000c3565b84620000ff565b602080601f831160018114620001be5760008415620001a55750858301515b600019600386901b1c1916600185901b17855562000149565b600085815260208120601f198616915b82811015620001ef57888601518255948401946001909101908401620001ce565b50858210156200020e5787850151600019600388901b60f8161c191681555b5050505050600190811b01905550565b60805160a05160c0516112c36200025c600039600061027a015260008181610397015261042c0152600081816101ca01526103bd01526112c36000f3fe608060405234801561001057600080fd5b50600436106101985760003560e01c80636a30b253116100e3578063ae1f6aaf1161008c578063dd62ed3e11610066578063dd62ed3e146103e1578063e78cea9214610395578063ee9a31a21461042757600080fd5b8063ae1f6aaf14610395578063c01e1bd6146103bb578063d6c0b2c4146103bb57600080fd5b80639dc29fac116100bd5780639dc29fac1461035c578063a457c2d71461036f578063a9059cbb1461038257600080fd5b80636a30b2531461031957806370a082311461032c57806395d89b411461035457600080fd5b80632e0f98ad1161014557806340c10f191161011f57806340c10f19146102b757806354fd4d50146102ca57806358cf96721461030657600080fd5b80632e0f98ad1461025e578063313ce5671461027357806339509351146102a457600080fd5b8063095ea7b311610176578063095ea7b31461022657806318160ddd1461023957806323b872dd1461024b57600080fd5b806301ffc9a71461019d578063033964be146101c557806306fdde0314610211575b600080fd5b6101b06101ab366004610ed5565b61044e565b60405190151581526020015b60405180910390f35b6101ec7f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016101bc565b61021961053f565b6040516101bc9190610f1e565b6101b0610234366004610fba565b6105d1565b6002545b6040519081526020016101bc565b6101b0610259366004610fe4565b610661565b61027161026c36600461106c565b6106ec565b005b60405160ff7f00000000000000000000000000000000000000000000000000000000000000001681526020016101bc565b6101b06102b2366004610fba565b61084f565b6102716102c5366004610fba565b6108da565b6102196040518060400160405280600581526020017f312e332e3000000000000000000000000000000000000000000000000000000081525081565b610271610314366004610fba565b61093c565b6102716103273660046110d8565b6109b2565b61023d61033a366004611150565b73ffffffffffffffffffffffffffffffffffffffff163190565b610219610a42565b61027161036a366004610fba565b610a51565b6101b061037d366004610fba565b610ab3565b6101b0610390366004610fba565b610b3e565b7f00000000000000000000000000000000000000000000000000000000000000006101ec565b7f00000000000000000000000000000000000000000000000000000000000000006101ec565b61023d6103ef36600461116b565b73ffffffffffffffffffffffffffffffffffffffff918216600090815260016020908152604080832093909416825291909152205490565b6101ec7f000000000000000000000000000000000000000000000000000000000000000081565b60007f01ffc9a7000000000000000000000000000000000000000000000000000000007f1d1d8b63000000000000000000000000000000000000000000000000000000007fec4fc8e3000000000000000000000000000000000000000000000000000000007fffffffff00000000000000000000000000000000000000000000000000000000851683148061050757507fffffffff00000000000000000000000000000000000000000000000000000000858116908316145b8061053657507fffffffff00000000000000000000000000000000000000000000000000000000858116908216145b95945050505050565b60606003805461054e9061119e565b80601f016020809104026020016040519081016040528092919081815260200182805461057a9061119e565b80156105c75780601f1061059c576101008083540402835291602001916105c7565b820191906000526020600020905b8154815290600101906020018083116105aa57829003601f168201915b5050505050905090565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f4c656761637945524332304554483a20617070726f766520697320646973616260448201527f6c6564000000000000000000000000000000000000000000000000000000000060648201526000906084015b60405180910390fd5b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602860248201527f4c656761637945524332304554483a207472616e7366657246726f6d2069732060448201527f64697361626c65640000000000000000000000000000000000000000000000006064820152600090608401610658565b3315610754576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c000000000000000000000000000000006044820152606401610658565b8281146107e3576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602f60248201527f526563697069656e747320616e6420616d6f756e7473206d757374206265207460448201527f68652073616d65206c656e6774682e00000000000000000000000000000000006064820152608401610658565b60005b8381101561084857610836858583818110610803576108036111f1565b90506020020160208101906108189190611150565b84848481811061082a5761082a6111f1565b90506020020135610bc8565b806108408161124f565b9150506107e6565b5050505050565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f4c656761637945524332304554483a20696e637265617365416c6c6f77616e6360448201527f652069732064697361626c6564000000000000000000000000000000000000006064820152600090608401610658565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4c656761637945524332304554483a206d696e742069732064697361626c65646044820152606401610658565b33156109a4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c000000000000000000000000000000006044820152606401610658565b6109ae8282610ce8565b5050565b3315610a1a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c000000000000000000000000000000006044820152606401610658565b610a248885610bc8565b610a2e8784610bc8565b610a388582610bc8565b5050505050505050565b60606004805461054e9061119e565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4c656761637945524332304554483a206275726e2069732064697361626c65646044820152606401610658565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f4c656761637945524332304554483a206465637265617365416c6c6f77616e6360448201527f652069732064697361626c6564000000000000000000000000000000000000006064820152600090608401610658565b6040517f08c379a0000000000000000000000000000000000000000000000000000000008152602060048201526024808201527f4c656761637945524332304554483a207472616e73666572206973206469736160448201527f626c6564000000000000000000000000000000000000000000000000000000006064820152600090608401610658565b73ffffffffffffffffffffffffffffffffffffffff8216610c45576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f45524332303a206d696e7420746f20746865207a65726f2061646472657373006044820152606401610658565b8060026000828254610c579190611287565b909155505073ffffffffffffffffffffffffffffffffffffffff821660009081526020819052604081208054839290610c91908490611287565b909155505060405181815273ffffffffffffffffffffffffffffffffffffffff8316906000907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9060200160405180910390a35050565b73ffffffffffffffffffffffffffffffffffffffff8216610d8b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602160248201527f45524332303a206275726e2066726f6d20746865207a65726f2061646472657360448201527f73000000000000000000000000000000000000000000000000000000000000006064820152608401610658565b73ffffffffffffffffffffffffffffffffffffffff821660009081526020819052604090205481811015610e41576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f45524332303a206275726e20616d6f756e7420657863656564732062616c616e60448201527f63650000000000000000000000000000000000000000000000000000000000006064820152608401610658565b73ffffffffffffffffffffffffffffffffffffffff83166000908152602081905260408120838303905560028054849290610e7d90849061129f565b909155505060405182815260009073ffffffffffffffffffffffffffffffffffffffff8516907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9060200160405180910390a3505050565b600060208284031215610ee757600080fd5b81357fffffffff0000000000000000000000000000000000000000000000000000000081168114610f1757600080fd5b9392505050565b600060208083528351808285015260005b81811015610f4b57858101830151858201604001528201610f2f565b81811115610f5d576000604083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016929092016040019392505050565b803573ffffffffffffffffffffffffffffffffffffffff81168114610fb557600080fd5b919050565b60008060408385031215610fcd57600080fd5b610fd683610f91565b946020939093013593505050565b600080600060608486031215610ff957600080fd5b61100284610f91565b925061101060208501610f91565b9150604084013590509250925092565b60008083601f84011261103257600080fd5b50813567ffffffffffffffff81111561104a57600080fd5b6020830191508360208260051b850101111561106557600080fd5b9250929050565b6000806000806040858703121561108257600080fd5b843567ffffffffffffffff8082111561109a57600080fd5b6110a688838901611020565b909650945060208701359150808211156110bf57600080fd5b506110cc87828801611020565b95989497509550505050565b600080600080600080600080610100898b0312156110f557600080fd5b6110fe89610f91565b975061110c60208a01610f91565b965061111a60408a01610f91565b955061112860608a01610f91565b979a969950949760808101359660a0820135965060c0820135955060e0909101359350915050565b60006020828403121561116257600080fd5b610f1782610f91565b6000806040838503121561117e57600080fd5b61118783610f91565b915061119560208401610f91565b90509250929050565b600181811c908216806111b257607f821691505b6020821081036111eb577f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff820361128057611280611220565b5060010190565b6000821982111561129a5761129a611220565b500190565b6000828210156112b1576112b1611220565b50039056fea164736f6c634300080f000a",
}

// LegacyERC20ETHABI is the input ABI used to generate the binding from.
// Deprecated: Use LegacyERC20ETHMetaData.ABI instead.
var LegacyERC20ETHABI = LegacyERC20ETHMetaData.ABI

// LegacyERC20ETHBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use LegacyERC20ETHMetaData.Bin instead.
var LegacyERC20ETHBin = LegacyERC20ETHMetaData.Bin

// DeployLegacyERC20ETH deploys a new Ethereum contract, binding an instance of LegacyERC20ETH to it.
func DeployLegacyERC20ETH(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *LegacyERC20ETH, error) {
	parsed, err := LegacyERC20ETHMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(LegacyERC20ETHBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &LegacyERC20ETH{LegacyERC20ETHCaller: LegacyERC20ETHCaller{contract: contract}, LegacyERC20ETHTransactor: LegacyERC20ETHTransactor{contract: contract}, LegacyERC20ETHFilterer: LegacyERC20ETHFilterer{contract: contract}}, nil
}

// LegacyERC20ETH is an auto generated Go binding around an Ethereum contract.
type LegacyERC20ETH struct {
	LegacyERC20ETHCaller     // Read-only binding to the contract
	LegacyERC20ETHTransactor // Write-only binding to the contract
	LegacyERC20ETHFilterer   // Log filterer for contract events
}

// LegacyERC20ETHCaller is an auto generated read-only Go binding around an Ethereum contract.
type LegacyERC20ETHCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LegacyERC20ETHTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LegacyERC20ETHTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LegacyERC20ETHFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LegacyERC20ETHFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LegacyERC20ETHSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LegacyERC20ETHSession struct {
	Contract     *LegacyERC20ETH   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LegacyERC20ETHCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LegacyERC20ETHCallerSession struct {
	Contract *LegacyERC20ETHCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// LegacyERC20ETHTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LegacyERC20ETHTransactorSession struct {
	Contract     *LegacyERC20ETHTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// LegacyERC20ETHRaw is an auto generated low-level Go binding around an Ethereum contract.
type LegacyERC20ETHRaw struct {
	Contract *LegacyERC20ETH // Generic contract binding to access the raw methods on
}

// LegacyERC20ETHCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LegacyERC20ETHCallerRaw struct {
	Contract *LegacyERC20ETHCaller // Generic read-only contract binding to access the raw methods on
}

// LegacyERC20ETHTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LegacyERC20ETHTransactorRaw struct {
	Contract *LegacyERC20ETHTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLegacyERC20ETH creates a new instance of LegacyERC20ETH, bound to a specific deployed contract.
func NewLegacyERC20ETH(address common.Address, backend bind.ContractBackend) (*LegacyERC20ETH, error) {
	contract, err := bindLegacyERC20ETH(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20ETH{LegacyERC20ETHCaller: LegacyERC20ETHCaller{contract: contract}, LegacyERC20ETHTransactor: LegacyERC20ETHTransactor{contract: contract}, LegacyERC20ETHFilterer: LegacyERC20ETHFilterer{contract: contract}}, nil
}

// NewLegacyERC20ETHCaller creates a new read-only instance of LegacyERC20ETH, bound to a specific deployed contract.
func NewLegacyERC20ETHCaller(address common.Address, caller bind.ContractCaller) (*LegacyERC20ETHCaller, error) {
	contract, err := bindLegacyERC20ETH(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20ETHCaller{contract: contract}, nil
}

// NewLegacyERC20ETHTransactor creates a new write-only instance of LegacyERC20ETH, bound to a specific deployed contract.
func NewLegacyERC20ETHTransactor(address common.Address, transactor bind.ContractTransactor) (*LegacyERC20ETHTransactor, error) {
	contract, err := bindLegacyERC20ETH(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20ETHTransactor{contract: contract}, nil
}

// NewLegacyERC20ETHFilterer creates a new log filterer instance of LegacyERC20ETH, bound to a specific deployed contract.
func NewLegacyERC20ETHFilterer(address common.Address, filterer bind.ContractFilterer) (*LegacyERC20ETHFilterer, error) {
	contract, err := bindLegacyERC20ETH(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20ETHFilterer{contract: contract}, nil
}

// bindLegacyERC20ETH binds a generic wrapper to an already deployed contract.
func bindLegacyERC20ETH(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(LegacyERC20ETHABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LegacyERC20ETH *LegacyERC20ETHRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _LegacyERC20ETH.Contract.LegacyERC20ETHCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LegacyERC20ETH *LegacyERC20ETHRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.LegacyERC20ETHTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LegacyERC20ETH *LegacyERC20ETHRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.LegacyERC20ETHTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LegacyERC20ETH *LegacyERC20ETHCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _LegacyERC20ETH.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LegacyERC20ETH *LegacyERC20ETHTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LegacyERC20ETH *LegacyERC20ETHTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.contract.Transact(opts, method, params...)
}

// BRIDGE is a free data retrieval call binding the contract method 0xee9a31a2.
//
// Solidity: function BRIDGE() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) BRIDGE(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "BRIDGE")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BRIDGE is a free data retrieval call binding the contract method 0xee9a31a2.
//
// Solidity: function BRIDGE() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHSession) BRIDGE() (common.Address, error) {
	return _LegacyERC20ETH.Contract.BRIDGE(&_LegacyERC20ETH.CallOpts)
}

// BRIDGE is a free data retrieval call binding the contract method 0xee9a31a2.
//
// Solidity: function BRIDGE() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) BRIDGE() (common.Address, error) {
	return _LegacyERC20ETH.Contract.BRIDGE(&_LegacyERC20ETH.CallOpts)
}

// REMOTETOKEN is a free data retrieval call binding the contract method 0x033964be.
//
// Solidity: function REMOTE_TOKEN() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) REMOTETOKEN(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "REMOTE_TOKEN")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// REMOTETOKEN is a free data retrieval call binding the contract method 0x033964be.
//
// Solidity: function REMOTE_TOKEN() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHSession) REMOTETOKEN() (common.Address, error) {
	return _LegacyERC20ETH.Contract.REMOTETOKEN(&_LegacyERC20ETH.CallOpts)
}

// REMOTETOKEN is a free data retrieval call binding the contract method 0x033964be.
//
// Solidity: function REMOTE_TOKEN() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) REMOTETOKEN() (common.Address, error) {
	return _LegacyERC20ETH.Contract.REMOTETOKEN(&_LegacyERC20ETH.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_LegacyERC20ETH *LegacyERC20ETHSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _LegacyERC20ETH.Contract.Allowance(&_LegacyERC20ETH.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _LegacyERC20ETH.Contract.Allowance(&_LegacyERC20ETH.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _who) view returns(uint256)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) BalanceOf(opts *bind.CallOpts, _who common.Address) (*big.Int, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "balanceOf", _who)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _who) view returns(uint256)
func (_LegacyERC20ETH *LegacyERC20ETHSession) BalanceOf(_who common.Address) (*big.Int, error) {
	return _LegacyERC20ETH.Contract.BalanceOf(&_LegacyERC20ETH.CallOpts, _who)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _who) view returns(uint256)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) BalanceOf(_who common.Address) (*big.Int, error) {
	return _LegacyERC20ETH.Contract.BalanceOf(&_LegacyERC20ETH.CallOpts, _who)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHSession) Bridge() (common.Address, error) {
	return _LegacyERC20ETH.Contract.Bridge(&_LegacyERC20ETH.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) Bridge() (common.Address, error) {
	return _LegacyERC20ETH.Contract.Bridge(&_LegacyERC20ETH.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_LegacyERC20ETH *LegacyERC20ETHSession) Decimals() (uint8, error) {
	return _LegacyERC20ETH.Contract.Decimals(&_LegacyERC20ETH.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) Decimals() (uint8, error) {
	return _LegacyERC20ETH.Contract.Decimals(&_LegacyERC20ETH.CallOpts)
}

// L1Token is a free data retrieval call binding the contract method 0xc01e1bd6.
//
// Solidity: function l1Token() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) L1Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "l1Token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L1Token is a free data retrieval call binding the contract method 0xc01e1bd6.
//
// Solidity: function l1Token() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHSession) L1Token() (common.Address, error) {
	return _LegacyERC20ETH.Contract.L1Token(&_LegacyERC20ETH.CallOpts)
}

// L1Token is a free data retrieval call binding the contract method 0xc01e1bd6.
//
// Solidity: function l1Token() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) L1Token() (common.Address, error) {
	return _LegacyERC20ETH.Contract.L1Token(&_LegacyERC20ETH.CallOpts)
}

// L2Bridge is a free data retrieval call binding the contract method 0xae1f6aaf.
//
// Solidity: function l2Bridge() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) L2Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "l2Bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L2Bridge is a free data retrieval call binding the contract method 0xae1f6aaf.
//
// Solidity: function l2Bridge() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHSession) L2Bridge() (common.Address, error) {
	return _LegacyERC20ETH.Contract.L2Bridge(&_LegacyERC20ETH.CallOpts)
}

// L2Bridge is a free data retrieval call binding the contract method 0xae1f6aaf.
//
// Solidity: function l2Bridge() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) L2Bridge() (common.Address, error) {
	return _LegacyERC20ETH.Contract.L2Bridge(&_LegacyERC20ETH.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_LegacyERC20ETH *LegacyERC20ETHSession) Name() (string, error) {
	return _LegacyERC20ETH.Contract.Name(&_LegacyERC20ETH.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) Name() (string, error) {
	return _LegacyERC20ETH.Contract.Name(&_LegacyERC20ETH.CallOpts)
}

// RemoteToken is a free data retrieval call binding the contract method 0xd6c0b2c4.
//
// Solidity: function remoteToken() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) RemoteToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "remoteToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RemoteToken is a free data retrieval call binding the contract method 0xd6c0b2c4.
//
// Solidity: function remoteToken() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHSession) RemoteToken() (common.Address, error) {
	return _LegacyERC20ETH.Contract.RemoteToken(&_LegacyERC20ETH.CallOpts)
}

// RemoteToken is a free data retrieval call binding the contract method 0xd6c0b2c4.
//
// Solidity: function remoteToken() view returns(address)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) RemoteToken() (common.Address, error) {
	return _LegacyERC20ETH.Contract.RemoteToken(&_LegacyERC20ETH.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 _interfaceId) pure returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) SupportsInterface(opts *bind.CallOpts, _interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "supportsInterface", _interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 _interfaceId) pure returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHSession) SupportsInterface(_interfaceId [4]byte) (bool, error) {
	return _LegacyERC20ETH.Contract.SupportsInterface(&_LegacyERC20ETH.CallOpts, _interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 _interfaceId) pure returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) SupportsInterface(_interfaceId [4]byte) (bool, error) {
	return _LegacyERC20ETH.Contract.SupportsInterface(&_LegacyERC20ETH.CallOpts, _interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_LegacyERC20ETH *LegacyERC20ETHSession) Symbol() (string, error) {
	return _LegacyERC20ETH.Contract.Symbol(&_LegacyERC20ETH.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) Symbol() (string, error) {
	return _LegacyERC20ETH.Contract.Symbol(&_LegacyERC20ETH.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_LegacyERC20ETH *LegacyERC20ETHSession) TotalSupply() (*big.Int, error) {
	return _LegacyERC20ETH.Contract.TotalSupply(&_LegacyERC20ETH.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) TotalSupply() (*big.Int, error) {
	return _LegacyERC20ETH.Contract.TotalSupply(&_LegacyERC20ETH.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_LegacyERC20ETH *LegacyERC20ETHCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _LegacyERC20ETH.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_LegacyERC20ETH *LegacyERC20ETHSession) Version() (string, error) {
	return _LegacyERC20ETH.Contract.Version(&_LegacyERC20ETH.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_LegacyERC20ETH *LegacyERC20ETHCallerSession) Version() (string, error) {
	return _LegacyERC20ETH.Contract.Version(&_LegacyERC20ETH.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHTransactor) Approve(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.contract.Transact(opts, "approve", arg0, arg1)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHSession) Approve(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.Approve(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHTransactorSession) Approve(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.Approve(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address , uint256 ) returns()
func (_LegacyERC20ETH *LegacyERC20ETHTransactor) Burn(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.contract.Transact(opts, "burn", arg0, arg1)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address , uint256 ) returns()
func (_LegacyERC20ETH *LegacyERC20ETHSession) Burn(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.Burn(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address , uint256 ) returns()
func (_LegacyERC20ETH *LegacyERC20ETHTransactorSession) Burn(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.Burn(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// CreditGasFees is a paid mutator transaction binding the contract method 0x2e0f98ad.
//
// Solidity: function creditGasFees(address[] recipients, uint256[] amounts) returns()
func (_LegacyERC20ETH *LegacyERC20ETHTransactor) CreditGasFees(opts *bind.TransactOpts, recipients []common.Address, amounts []*big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.contract.Transact(opts, "creditGasFees", recipients, amounts)
}

// CreditGasFees is a paid mutator transaction binding the contract method 0x2e0f98ad.
//
// Solidity: function creditGasFees(address[] recipients, uint256[] amounts) returns()
func (_LegacyERC20ETH *LegacyERC20ETHSession) CreditGasFees(recipients []common.Address, amounts []*big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.CreditGasFees(&_LegacyERC20ETH.TransactOpts, recipients, amounts)
}

// CreditGasFees is a paid mutator transaction binding the contract method 0x2e0f98ad.
//
// Solidity: function creditGasFees(address[] recipients, uint256[] amounts) returns()
func (_LegacyERC20ETH *LegacyERC20ETHTransactorSession) CreditGasFees(recipients []common.Address, amounts []*big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.CreditGasFees(&_LegacyERC20ETH.TransactOpts, recipients, amounts)
}

// CreditGasFees0 is a paid mutator transaction binding the contract method 0x6a30b253.
//
// Solidity: function creditGasFees(address from, address feeRecipient, address , address communityFund, uint256 refund, uint256 tipTxFee, uint256 , uint256 baseTxFee) returns()
func (_LegacyERC20ETH *LegacyERC20ETHTransactor) CreditGasFees0(opts *bind.TransactOpts, from common.Address, feeRecipient common.Address, arg2 common.Address, communityFund common.Address, refund *big.Int, tipTxFee *big.Int, arg6 *big.Int, baseTxFee *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.contract.Transact(opts, "creditGasFees0", from, feeRecipient, arg2, communityFund, refund, tipTxFee, arg6, baseTxFee)
}

// CreditGasFees0 is a paid mutator transaction binding the contract method 0x6a30b253.
//
// Solidity: function creditGasFees(address from, address feeRecipient, address , address communityFund, uint256 refund, uint256 tipTxFee, uint256 , uint256 baseTxFee) returns()
func (_LegacyERC20ETH *LegacyERC20ETHSession) CreditGasFees0(from common.Address, feeRecipient common.Address, arg2 common.Address, communityFund common.Address, refund *big.Int, tipTxFee *big.Int, arg6 *big.Int, baseTxFee *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.CreditGasFees0(&_LegacyERC20ETH.TransactOpts, from, feeRecipient, arg2, communityFund, refund, tipTxFee, arg6, baseTxFee)
}

// CreditGasFees0 is a paid mutator transaction binding the contract method 0x6a30b253.
//
// Solidity: function creditGasFees(address from, address feeRecipient, address , address communityFund, uint256 refund, uint256 tipTxFee, uint256 , uint256 baseTxFee) returns()
func (_LegacyERC20ETH *LegacyERC20ETHTransactorSession) CreditGasFees0(from common.Address, feeRecipient common.Address, arg2 common.Address, communityFund common.Address, refund *big.Int, tipTxFee *big.Int, arg6 *big.Int, baseTxFee *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.CreditGasFees0(&_LegacyERC20ETH.TransactOpts, from, feeRecipient, arg2, communityFund, refund, tipTxFee, arg6, baseTxFee)
}

// DebitGasFees is a paid mutator transaction binding the contract method 0x58cf9672.
//
// Solidity: function debitGasFees(address from, uint256 value) returns()
func (_LegacyERC20ETH *LegacyERC20ETHTransactor) DebitGasFees(opts *bind.TransactOpts, from common.Address, value *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.contract.Transact(opts, "debitGasFees", from, value)
}

// DebitGasFees is a paid mutator transaction binding the contract method 0x58cf9672.
//
// Solidity: function debitGasFees(address from, uint256 value) returns()
func (_LegacyERC20ETH *LegacyERC20ETHSession) DebitGasFees(from common.Address, value *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.DebitGasFees(&_LegacyERC20ETH.TransactOpts, from, value)
}

// DebitGasFees is a paid mutator transaction binding the contract method 0x58cf9672.
//
// Solidity: function debitGasFees(address from, uint256 value) returns()
func (_LegacyERC20ETH *LegacyERC20ETHTransactorSession) DebitGasFees(from common.Address, value *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.DebitGasFees(&_LegacyERC20ETH.TransactOpts, from, value)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHTransactor) DecreaseAllowance(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.contract.Transact(opts, "decreaseAllowance", arg0, arg1)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHSession) DecreaseAllowance(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.DecreaseAllowance(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHTransactorSession) DecreaseAllowance(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.DecreaseAllowance(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHTransactor) IncreaseAllowance(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.contract.Transact(opts, "increaseAllowance", arg0, arg1)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHSession) IncreaseAllowance(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.IncreaseAllowance(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHTransactorSession) IncreaseAllowance(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.IncreaseAllowance(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address , uint256 ) returns()
func (_LegacyERC20ETH *LegacyERC20ETHTransactor) Mint(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.contract.Transact(opts, "mint", arg0, arg1)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address , uint256 ) returns()
func (_LegacyERC20ETH *LegacyERC20ETHSession) Mint(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.Mint(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address , uint256 ) returns()
func (_LegacyERC20ETH *LegacyERC20ETHTransactorSession) Mint(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.Mint(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHTransactor) Transfer(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.contract.Transact(opts, "transfer", arg0, arg1)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHSession) Transfer(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.Transfer(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHTransactorSession) Transfer(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.Transfer(&_LegacyERC20ETH.TransactOpts, arg0, arg1)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address , address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHTransactor) TransferFrom(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.contract.Transact(opts, "transferFrom", arg0, arg1, arg2)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address , address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHSession) TransferFrom(arg0 common.Address, arg1 common.Address, arg2 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.TransferFrom(&_LegacyERC20ETH.TransactOpts, arg0, arg1, arg2)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address , address , uint256 ) returns(bool)
func (_LegacyERC20ETH *LegacyERC20ETHTransactorSession) TransferFrom(arg0 common.Address, arg1 common.Address, arg2 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20ETH.Contract.TransferFrom(&_LegacyERC20ETH.TransactOpts, arg0, arg1, arg2)
}

// LegacyERC20ETHApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the LegacyERC20ETH contract.
type LegacyERC20ETHApprovalIterator struct {
	Event *LegacyERC20ETHApproval // Event containing the contract specifics and raw log

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
func (it *LegacyERC20ETHApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LegacyERC20ETHApproval)
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
		it.Event = new(LegacyERC20ETHApproval)
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
func (it *LegacyERC20ETHApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LegacyERC20ETHApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LegacyERC20ETHApproval represents a Approval event raised by the LegacyERC20ETH contract.
type LegacyERC20ETHApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*LegacyERC20ETHApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _LegacyERC20ETH.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20ETHApprovalIterator{contract: _LegacyERC20ETH.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *LegacyERC20ETHApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _LegacyERC20ETH.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LegacyERC20ETHApproval)
				if err := _LegacyERC20ETH.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) ParseApproval(log types.Log) (*LegacyERC20ETHApproval, error) {
	event := new(LegacyERC20ETHApproval)
	if err := _LegacyERC20ETH.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LegacyERC20ETHBurnIterator is returned from FilterBurn and is used to iterate over the raw logs and unpacked data for Burn events raised by the LegacyERC20ETH contract.
type LegacyERC20ETHBurnIterator struct {
	Event *LegacyERC20ETHBurn // Event containing the contract specifics and raw log

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
func (it *LegacyERC20ETHBurnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LegacyERC20ETHBurn)
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
		it.Event = new(LegacyERC20ETHBurn)
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
func (it *LegacyERC20ETHBurnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LegacyERC20ETHBurnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LegacyERC20ETHBurn represents a Burn event raised by the LegacyERC20ETH contract.
type LegacyERC20ETHBurn struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5.
//
// Solidity: event Burn(address indexed account, uint256 amount)
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) FilterBurn(opts *bind.FilterOpts, account []common.Address) (*LegacyERC20ETHBurnIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _LegacyERC20ETH.contract.FilterLogs(opts, "Burn", accountRule)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20ETHBurnIterator{contract: _LegacyERC20ETH.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5.
//
// Solidity: event Burn(address indexed account, uint256 amount)
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *LegacyERC20ETHBurn, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _LegacyERC20ETH.contract.WatchLogs(opts, "Burn", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LegacyERC20ETHBurn)
				if err := _LegacyERC20ETH.contract.UnpackLog(event, "Burn", log); err != nil {
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

// ParseBurn is a log parse operation binding the contract event 0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5.
//
// Solidity: event Burn(address indexed account, uint256 amount)
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) ParseBurn(log types.Log) (*LegacyERC20ETHBurn, error) {
	event := new(LegacyERC20ETHBurn)
	if err := _LegacyERC20ETH.contract.UnpackLog(event, "Burn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LegacyERC20ETHMintIterator is returned from FilterMint and is used to iterate over the raw logs and unpacked data for Mint events raised by the LegacyERC20ETH contract.
type LegacyERC20ETHMintIterator struct {
	Event *LegacyERC20ETHMint // Event containing the contract specifics and raw log

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
func (it *LegacyERC20ETHMintIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LegacyERC20ETHMint)
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
		it.Event = new(LegacyERC20ETHMint)
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
func (it *LegacyERC20ETHMintIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LegacyERC20ETHMintIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LegacyERC20ETHMint represents a Mint event raised by the LegacyERC20ETH contract.
type LegacyERC20ETHMint struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMint is a free log retrieval operation binding the contract event 0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885.
//
// Solidity: event Mint(address indexed account, uint256 amount)
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) FilterMint(opts *bind.FilterOpts, account []common.Address) (*LegacyERC20ETHMintIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _LegacyERC20ETH.contract.FilterLogs(opts, "Mint", accountRule)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20ETHMintIterator{contract: _LegacyERC20ETH.contract, event: "Mint", logs: logs, sub: sub}, nil
}

// WatchMint is a free log subscription operation binding the contract event 0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885.
//
// Solidity: event Mint(address indexed account, uint256 amount)
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) WatchMint(opts *bind.WatchOpts, sink chan<- *LegacyERC20ETHMint, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _LegacyERC20ETH.contract.WatchLogs(opts, "Mint", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LegacyERC20ETHMint)
				if err := _LegacyERC20ETH.contract.UnpackLog(event, "Mint", log); err != nil {
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

// ParseMint is a log parse operation binding the contract event 0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885.
//
// Solidity: event Mint(address indexed account, uint256 amount)
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) ParseMint(log types.Log) (*LegacyERC20ETHMint, error) {
	event := new(LegacyERC20ETHMint)
	if err := _LegacyERC20ETH.contract.UnpackLog(event, "Mint", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LegacyERC20ETHTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the LegacyERC20ETH contract.
type LegacyERC20ETHTransferIterator struct {
	Event *LegacyERC20ETHTransfer // Event containing the contract specifics and raw log

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
func (it *LegacyERC20ETHTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LegacyERC20ETHTransfer)
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
		it.Event = new(LegacyERC20ETHTransfer)
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
func (it *LegacyERC20ETHTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LegacyERC20ETHTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LegacyERC20ETHTransfer represents a Transfer event raised by the LegacyERC20ETH contract.
type LegacyERC20ETHTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*LegacyERC20ETHTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _LegacyERC20ETH.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20ETHTransferIterator{contract: _LegacyERC20ETH.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *LegacyERC20ETHTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _LegacyERC20ETH.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LegacyERC20ETHTransfer)
				if err := _LegacyERC20ETH.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_LegacyERC20ETH *LegacyERC20ETHFilterer) ParseTransfer(log types.Log) (*LegacyERC20ETHTransfer, error) {
	event := new(LegacyERC20ETHTransfer)
	if err := _LegacyERC20ETH.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
