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

// L1StandardBridgeMetaData contains all meta data concerning the L1StandardBridge contract.
var L1StandardBridgeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"localToken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"remoteToken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"}],\"name\":\"ERC20BridgeFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"localToken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"remoteToken\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"}],\"name\":\"ERC20BridgeInitiated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"l1Token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"l2Token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"}],\"name\":\"ERC20DepositInitiated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"l1Token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"l2Token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"}],\"name\":\"ERC20WithdrawalFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"}],\"name\":\"ETHBridgeFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"}],\"name\":\"ETHBridgeInitiated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"}],\"name\":\"ETHDepositInitiated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"extraData\",\"type\":\"bytes\"}],\"name\":\"ETHWithdrawalFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MESSENGER\",\"outputs\":[{\"internalType\":\"contractCrossDomainMessenger\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"OTHER_BRIDGE\",\"outputs\":[{\"internalType\":\"contractStandardBridge\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_localToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_remoteToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"bridgeERC20\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_localToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_remoteToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"bridgeERC20To\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"bridgeETH\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"bridgeETHTo\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_l1Token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_l2Token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"depositERC20\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_l1Token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_l2Token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"depositERC20To\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"depositETH\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"depositETHTo\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_localToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_remoteToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"finalizeBridgeERC20\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"finalizeBridgeETH\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_l1Token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_l2Token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"finalizeERC20Withdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"finalizeETHWithdrawal\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractCrossDomainMessenger\",\"name\":\"_messenger\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l2TokenBridge\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"messenger\",\"outputs\":[{\"internalType\":\"contractCrossDomainMessenger\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"otherBridge\",\"outputs\":[{\"internalType\":\"contractStandardBridge\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60a06040523480156200001157600080fd5b507342000000000000000000000000000000000000106080526200003660006200003c565b620001bb565b600054600390610100900460ff161580156200005f575060005460ff8083169116105b620000c85760405162461bcd60e51b815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201526d191e481a5b9a5d1a585b1a5e995960921b60648201526084015b60405180910390fd5b6000805461ffff191660ff831617610100179055620000e7826200012c565b6000805461ff001916905560405160ff821681527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a15050565b600054610100900460ff16620001995760405162461bcd60e51b815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201526a6e697469616c697a696e6760a81b6064820152608401620000bf565b600380546001600160a01b0319166001600160a01b0392909216919091179055565b6080516129ff620001fa60003960008181610379015281816104300152818161058501528181610a4c015281816113860152611a3f01526129ff6000f3fe6080604052600436106101635760003560e01c806387087623116100c0578063a9f9e67511610074578063c4d66de811610059578063c4d66de8146104c5578063c89701a214610421578063e11013dd146104e557600080fd5b8063a9f9e67514610492578063b1a1a882146104b257600080fd5b806391c49bf8116100a557806391c49bf814610421578063927ede2d146104545780639a2ac6d51461047f57600080fd5b806387087623146103bb5780638f601f66146103db57600080fd5b8063540abf731161011757806358a997f6116100fc57806358a997f6146103475780637f46ddb214610367578063838b25201461039b57600080fd5b8063540abf73146102d157806354fd4d50146102f157600080fd5b80631532ec34116101485780631532ec34146102545780631635f5fd146102675780633cb747bf1461027a57600080fd5b80630166a07a1461022157806309fc88431461024157600080fd5b3661021c57333b156101fc576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20616e20454f4100000000000000000060648201526084015b60405180910390fd5b61021a333362030d40604051806020016040528060008152506104f8565b005b600080fd5b34801561022d57600080fd5b5061021a61023c36600461243f565b610520565b61021a61024f3660046124f0565b6108e7565b61021a610262366004612543565b6109d3565b61021a610275366004612543565b6109e7565b34801561028657600080fd5b506003546102a79073ffffffffffffffffffffffffffffffffffffffff1681565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020015b60405180910390f35b3480156102dd57600080fd5b5061021a6102ec3660046125b6565b610dce565b3480156102fd57600080fd5b5061033a6040518060400160405280600581526020017f312e342e3100000000000000000000000000000000000000000000000000000081525081565b6040516102c891906126a3565b34801561035357600080fd5b5061021a6103623660046126b6565b610e13565b34801561037357600080fd5b506102a77f000000000000000000000000000000000000000000000000000000000000000081565b3480156103a757600080fd5b5061021a6103b63660046125b6565b610ee7565b3480156103c757600080fd5b5061021a6103d63660046126b6565b610f2c565b3480156103e757600080fd5b506104136103f6366004612739565b600260209081526000928352604080842090915290825290205481565b6040519081526020016102c8565b34801561042d57600080fd5b507f00000000000000000000000000000000000000000000000000000000000000006102a7565b34801561046057600080fd5b5060035473ffffffffffffffffffffffffffffffffffffffff166102a7565b61021a61048d366004612772565b611000565b34801561049e57600080fd5b5061021a6104ad36600461243f565b611042565b61021a6104c03660046124f0565b611051565b3480156104d157600080fd5b5061021a6104e03660046127d5565b611122565b61021a6104f3366004612772565b61126c565b61051a73420000000000000000000000000000000000102385853486866112c4565b50505050565b60035473ffffffffffffffffffffffffffffffffffffffff163314801561060f5750600354604080517f6e296e45000000000000000000000000000000000000000000000000000000008152905173ffffffffffffffffffffffffffffffffffffffff7f00000000000000000000000000000000000000000000000000000000000000008116931691636e296e459160048083019260209291908290030181865afa1580156105d3573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105f791906127f2565b73ffffffffffffffffffffffffffffffffffffffff16145b6106c1576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604160248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20746865206f7468657220627269646760648201527f6500000000000000000000000000000000000000000000000000000000000000608482015260a4016101f3565b6106ca876114ae565b15610818576106d98787611510565b61078b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604a60248201527f5374616e646172644272696467653a2077726f6e672072656d6f746520746f6b60448201527f656e20666f72204f7074696d69736d204d696e7461626c65204552433230206c60648201527f6f63616c20746f6b656e00000000000000000000000000000000000000000000608482015260a4016101f3565b6040517f40c10f1900000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8581166004830152602482018590528816906340c10f1990604401600060405180830381600087803b1580156107fb57600080fd5b505af115801561080f573d6000803e3d6000fd5b5050505061089a565b73ffffffffffffffffffffffffffffffffffffffff8088166000908152600260209081526040808320938a168352929052205461085690849061283e565b73ffffffffffffffffffffffffffffffffffffffff8089166000818152600260209081526040808320948c168352939052919091209190915561089a908585611630565b6108de878787878787878080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061170492505050565b50505050505050565b333b15610976576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20616e20454f4100000000000000000060648201526084016101f3565b6109ce7342000000000000000000000000000000000010233333348787878080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152506112c492505050565b505050565b6109e085858585856109e7565b5050505050565b60035473ffffffffffffffffffffffffffffffffffffffff1633148015610ad65750600354604080517f6e296e45000000000000000000000000000000000000000000000000000000008152905173ffffffffffffffffffffffffffffffffffffffff7f00000000000000000000000000000000000000000000000000000000000000008116931691636e296e459160048083019260209291908290030181865afa158015610a9a573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610abe91906127f2565b73ffffffffffffffffffffffffffffffffffffffff16145b610b88576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604160248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20746865206f7468657220627269646760648201527f6500000000000000000000000000000000000000000000000000000000000000608482015260a4016101f3565b3073ffffffffffffffffffffffffffffffffffffffff851603610c2d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f5374616e646172644272696467653a2063616e6e6f742073656e6420746f207360448201527f656c66000000000000000000000000000000000000000000000000000000000060648201526084016101f3565b60035473ffffffffffffffffffffffffffffffffffffffff90811690851603610cd8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602860248201527f5374616e646172644272696467653a2063616e6e6f742073656e6420746f206d60448201527f657373656e67657200000000000000000000000000000000000000000000000060648201526084016101f3565b610d1a85858585858080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061179292505050565b6000610d37855a8660405180602001604052806000815250611805565b905080610dc6576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f5374616e646172644272696467653a20455448207472616e736665722066616960448201527f6c6564000000000000000000000000000000000000000000000000000000000060648201526084016101f3565b505050505050565b6108de87873388888888888080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061181f92505050565b333b15610ea2576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20616e20454f4100000000000000000060648201526084016101f3565b610dc686863333888888888080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250611b6792505050565b6108de87873388888888888080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250611b6792505050565b333b15610fbb576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20616e20454f4100000000000000000060648201526084016101f3565b610dc686863333888888888080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061181f92505050565b61051a33858585858080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152506104f892505050565b6108de87878787878787610520565b333b156110e0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f5374616e646172644272696467653a2066756e6374696f6e2063616e206f6e6c60448201527f792062652063616c6c65642066726f6d20616e20454f4100000000000000000060648201526084016101f3565b6109ce33338585858080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152506104f892505050565b600054600390610100900460ff16158015611144575060005460ff8083169116105b6111d0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201527f647920696e697469616c697a656400000000000000000000000000000000000060648201526084016101f3565b600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00001660ff83161761010017905561120a82611b76565b600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff16905560405160ff821681527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a15050565b61051a7342000000000000000000000000000000000010233386348787878080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152506112c492505050565b823414611353576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603e60248201527f5374616e646172644272696467653a206272696467696e6720455448206d757360448201527f7420696e636c7564652073756666696369656e74204554482076616c7565000060648201526084016101f3565b61135f85858584611c54565b60035460405173ffffffffffffffffffffffffffffffffffffffff90911690633dbb202b907f0000000000000000000000000000000000000000000000000000000000000000907f0166a07a00000000000000000000000000000000000000000000000000000000906113e1908b906000908c908c908c908b90602401612855565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529181526020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fffffffff000000000000000000000000000000000000000000000000000000009485161790525160e085901b9092168252611474929187906004016128b1565b600060405180830381600087803b15801561148e57600080fd5b505af11580156114a2573d6000803e3d6000fd5b50505050505050505050565b60006114da827f1d1d8b6300000000000000000000000000000000000000000000000000000000611cc7565b8061150a575061150a827fec4fc8e300000000000000000000000000000000000000000000000000000000611cc7565b92915050565b600061153c837f1d1d8b6300000000000000000000000000000000000000000000000000000000611cc7565b156115e5578273ffffffffffffffffffffffffffffffffffffffff1663c01e1bd66040518163ffffffff1660e01b8152600401602060405180830381865afa15801561158c573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906115b091906127f2565b73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1614905061150a565b8273ffffffffffffffffffffffffffffffffffffffff1663d6c0b2c46040518163ffffffff1660e01b8152600401602060405180830381865afa15801561158c573d6000803e3d6000fd5b60405173ffffffffffffffffffffffffffffffffffffffff83166024820152604481018290526109ce9084907fa9059cbb00000000000000000000000000000000000000000000000000000000906064015b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529190526020810180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fffffffff0000000000000000000000000000000000000000000000000000000090931692909217909152611cea565b8373ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff167f3ceee06c1e37648fcbb6ed52e17b3e1f275a1f8c7b22a84b2b84732431e046b386868660405161177c939291906128f6565b60405180910390a4610dc6868686868686611df6565b8273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f2ac69ee804d9a7a0984249f508dfab7cb2534b465b6ce1580f99a38ba9c5e63184846040516117f1929190612934565b60405180910390a361051a84848484611e7e565b600080600080845160208601878a8af19695505050505050565b611828876114ae565b15611976576118378787611510565b6118e9576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604a60248201527f5374616e646172644272696467653a2077726f6e672072656d6f746520746f6b60448201527f656e20666f72204f7074696d69736d204d696e7461626c65204552433230206c60648201527f6f63616c20746f6b656e00000000000000000000000000000000000000000000608482015260a4016101f3565b6040517f9dc29fac00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff868116600483015260248201859052881690639dc29fac90604401600060405180830381600087803b15801561195957600080fd5b505af115801561196d573d6000803e3d6000fd5b50505050611a0a565b61199873ffffffffffffffffffffffffffffffffffffffff8816863086611eeb565b73ffffffffffffffffffffffffffffffffffffffff8088166000908152600260209081526040808320938a16835292905220546119d690849061294d565b73ffffffffffffffffffffffffffffffffffffffff8089166000908152600260209081526040808320938b16835292905220555b611a18878787878786611f49565b60035460405173ffffffffffffffffffffffffffffffffffffffff90911690633dbb202b907f0000000000000000000000000000000000000000000000000000000000000000907f0166a07a0000000000000000000000000000000000000000000000000000000090611a99908b908d908c908c908c908b90602401612965565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529181526020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fffffffff000000000000000000000000000000000000000000000000000000009485161790525160e085901b9092168252611b2c929187906004016128b1565b600060405180830381600087803b158015611b4657600080fd5b505af1158015611b5a573d6000803e3d6000fd5b5050505050505050505050565b6108de8787878787878761181f565b600054610100900460ff16611c0d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201527f6e697469616c697a696e6700000000000000000000000000000000000000000060648201526084016101f3565b600380547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff92909216919091179055565b8273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f35d79ab81f2b2017e19afb5c5571778877782d7a8786f5907f93b0f4702f4f238484604051611cb3929190612934565b60405180910390a361051a84848484611fd7565b6000611cd283612036565b8015611ce35750611ce3838361209a565b9392505050565b6000611d4c826040518060400160405280602081526020017f5361666545524332303a206c6f772d6c6576656c2063616c6c206661696c65648152508573ffffffffffffffffffffffffffffffffffffffff166121699092919063ffffffff16565b8051909150156109ce5780806020019051810190611d6a91906129b4565b6109ce576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602a60248201527f5361666545524332303a204552433230206f7065726174696f6e20646964206e60448201527f6f7420737563636565640000000000000000000000000000000000000000000060648201526084016101f3565b8373ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff167fd59c65b35445225835c83f50b6ede06a7be047d22e357073e250d9af537518cd868686604051611e6e939291906128f6565b60405180910390a4505050505050565b8273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f31b2166ff604fc5672ea5df08a78081d2bc6d746cadce880747f3643d819e83d8484604051611edd929190612934565b60405180910390a350505050565b60405173ffffffffffffffffffffffffffffffffffffffff8085166024830152831660448201526064810182905261051a9085907f23b872dd0000000000000000000000000000000000000000000000000000000090608401611682565b8373ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff167f718594027abd4eaed59f95162563e0cc6d0e8d5b86b1c7be8b1b0ac3343d0396868686604051611fc1939291906128f6565b60405180910390a4610dc6868686868686612180565b8273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f2849b43074093a05396b6f2a937dee8565b15a48a7b3d4bffb732a5017380af58484604051611edd929190612934565b6000612062827f01ffc9a70000000000000000000000000000000000000000000000000000000061209a565b801561150a5750612093827fffffffff0000000000000000000000000000000000000000000000000000000061209a565b1592915050565b604080517fffffffff000000000000000000000000000000000000000000000000000000008316602480830191909152825180830390910181526044909101909152602080820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167f01ffc9a700000000000000000000000000000000000000000000000000000000178152825160009392849283928392918391908a617530fa92503d91506000519050828015612152575060208210155b801561215e5750600081115b979650505050505050565b606061217884846000856121f8565b949350505050565b8373ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff167f7ff126db8024424bbfd9826e8ab82ff59136289ea440b04b39a0df1b03b9cabf868686604051611e6e939291906128f6565b60608247101561228a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f416464726573733a20696e73756666696369656e742062616c616e636520666f60448201527f722063616c6c000000000000000000000000000000000000000000000000000060648201526084016101f3565b73ffffffffffffffffffffffffffffffffffffffff85163b612308576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f416464726573733a2063616c6c20746f206e6f6e2d636f6e747261637400000060448201526064016101f3565b6000808673ffffffffffffffffffffffffffffffffffffffff16858760405161233191906129d6565b60006040518083038185875af1925050503d806000811461236e576040519150601f19603f3d011682016040523d82523d6000602084013e612373565b606091505b509150915061215e8282866060831561238d575081611ce3565b82511561239d5782518084602001fd5b816040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101f391906126a3565b73ffffffffffffffffffffffffffffffffffffffff811681146123f357600080fd5b50565b60008083601f84011261240857600080fd5b50813567ffffffffffffffff81111561242057600080fd5b60208301915083602082850101111561243857600080fd5b9250929050565b600080600080600080600060c0888a03121561245a57600080fd5b8735612465816123d1565b96506020880135612475816123d1565b95506040880135612485816123d1565b94506060880135612495816123d1565b93506080880135925060a088013567ffffffffffffffff8111156124b857600080fd5b6124c48a828b016123f6565b989b979a50959850939692959293505050565b803563ffffffff811681146124eb57600080fd5b919050565b60008060006040848603121561250557600080fd5b61250e846124d7565b9250602084013567ffffffffffffffff81111561252a57600080fd5b612536868287016123f6565b9497909650939450505050565b60008060008060006080868803121561255b57600080fd5b8535612566816123d1565b94506020860135612576816123d1565b935060408601359250606086013567ffffffffffffffff81111561259957600080fd5b6125a5888289016123f6565b969995985093965092949392505050565b600080600080600080600060c0888a0312156125d157600080fd5b87356125dc816123d1565b965060208801356125ec816123d1565b955060408801356125fc816123d1565b945060608801359350612611608089016124d7565b925060a088013567ffffffffffffffff8111156124b857600080fd5b60005b83811015612648578181015183820152602001612630565b8381111561051a5750506000910152565b6000815180845261267181602086016020860161262d565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169290920160200192915050565b602081526000611ce36020830184612659565b60008060008060008060a087890312156126cf57600080fd5b86356126da816123d1565b955060208701356126ea816123d1565b9450604087013593506126ff606088016124d7565b9250608087013567ffffffffffffffff81111561271b57600080fd5b61272789828a016123f6565b979a9699509497509295939492505050565b6000806040838503121561274c57600080fd5b8235612757816123d1565b91506020830135612767816123d1565b809150509250929050565b6000806000806060858703121561278857600080fd5b8435612793816123d1565b93506127a1602086016124d7565b9250604085013567ffffffffffffffff8111156127bd57600080fd5b6127c9878288016123f6565b95989497509550505050565b6000602082840312156127e757600080fd5b8135611ce3816123d1565b60006020828403121561280457600080fd5b8151611ce3816123d1565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000828210156128505761285061280f565b500390565b600073ffffffffffffffffffffffffffffffffffffffff808916835260ff88166020840152808716604084015280861660608401525083608083015260c060a08301526128a560c0830184612659565b98975050505050505050565b73ffffffffffffffffffffffffffffffffffffffff841681526060602082015260006128e06060830185612659565b905063ffffffff83166040830152949350505050565b73ffffffffffffffffffffffffffffffffffffffff8416815282602082015260606040820152600061292b6060830184612659565b95945050505050565b8281526040602082015260006121786040830184612659565b600082198211156129605761296061280f565b500190565b600073ffffffffffffffffffffffffffffffffffffffff80891683528088166020840152808716604084015280861660608401525083608083015260c060a08301526128a560c0830184612659565b6000602082840312156129c657600080fd5b81518015158114611ce357600080fd5b600082516129e881846020870161262d565b919091019291505056fea164736f6c634300080f000a",
}

// L1StandardBridgeABI is the input ABI used to generate the binding from.
// Deprecated: Use L1StandardBridgeMetaData.ABI instead.
var L1StandardBridgeABI = L1StandardBridgeMetaData.ABI

// L1StandardBridgeBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use L1StandardBridgeMetaData.Bin instead.
var L1StandardBridgeBin = L1StandardBridgeMetaData.Bin

// DeployL1StandardBridge deploys a new Ethereum contract, binding an instance of L1StandardBridge to it.
func DeployL1StandardBridge(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *L1StandardBridge, error) {
	parsed, err := L1StandardBridgeMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(L1StandardBridgeBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &L1StandardBridge{L1StandardBridgeCaller: L1StandardBridgeCaller{contract: contract}, L1StandardBridgeTransactor: L1StandardBridgeTransactor{contract: contract}, L1StandardBridgeFilterer: L1StandardBridgeFilterer{contract: contract}}, nil
}

// L1StandardBridge is an auto generated Go binding around an Ethereum contract.
type L1StandardBridge struct {
	L1StandardBridgeCaller     // Read-only binding to the contract
	L1StandardBridgeTransactor // Write-only binding to the contract
	L1StandardBridgeFilterer   // Log filterer for contract events
}

// L1StandardBridgeCaller is an auto generated read-only Go binding around an Ethereum contract.
type L1StandardBridgeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L1StandardBridgeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type L1StandardBridgeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L1StandardBridgeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type L1StandardBridgeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L1StandardBridgeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type L1StandardBridgeSession struct {
	Contract     *L1StandardBridge // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// L1StandardBridgeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type L1StandardBridgeCallerSession struct {
	Contract *L1StandardBridgeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// L1StandardBridgeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type L1StandardBridgeTransactorSession struct {
	Contract     *L1StandardBridgeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// L1StandardBridgeRaw is an auto generated low-level Go binding around an Ethereum contract.
type L1StandardBridgeRaw struct {
	Contract *L1StandardBridge // Generic contract binding to access the raw methods on
}

// L1StandardBridgeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type L1StandardBridgeCallerRaw struct {
	Contract *L1StandardBridgeCaller // Generic read-only contract binding to access the raw methods on
}

// L1StandardBridgeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type L1StandardBridgeTransactorRaw struct {
	Contract *L1StandardBridgeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewL1StandardBridge creates a new instance of L1StandardBridge, bound to a specific deployed contract.
func NewL1StandardBridge(address common.Address, backend bind.ContractBackend) (*L1StandardBridge, error) {
	contract, err := bindL1StandardBridge(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridge{L1StandardBridgeCaller: L1StandardBridgeCaller{contract: contract}, L1StandardBridgeTransactor: L1StandardBridgeTransactor{contract: contract}, L1StandardBridgeFilterer: L1StandardBridgeFilterer{contract: contract}}, nil
}

// NewL1StandardBridgeCaller creates a new read-only instance of L1StandardBridge, bound to a specific deployed contract.
func NewL1StandardBridgeCaller(address common.Address, caller bind.ContractCaller) (*L1StandardBridgeCaller, error) {
	contract, err := bindL1StandardBridge(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeCaller{contract: contract}, nil
}

// NewL1StandardBridgeTransactor creates a new write-only instance of L1StandardBridge, bound to a specific deployed contract.
func NewL1StandardBridgeTransactor(address common.Address, transactor bind.ContractTransactor) (*L1StandardBridgeTransactor, error) {
	contract, err := bindL1StandardBridge(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeTransactor{contract: contract}, nil
}

// NewL1StandardBridgeFilterer creates a new log filterer instance of L1StandardBridge, bound to a specific deployed contract.
func NewL1StandardBridgeFilterer(address common.Address, filterer bind.ContractFilterer) (*L1StandardBridgeFilterer, error) {
	contract, err := bindL1StandardBridge(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeFilterer{contract: contract}, nil
}

// bindL1StandardBridge binds a generic wrapper to an already deployed contract.
func bindL1StandardBridge(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(L1StandardBridgeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_L1StandardBridge *L1StandardBridgeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _L1StandardBridge.Contract.L1StandardBridgeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_L1StandardBridge *L1StandardBridgeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.L1StandardBridgeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_L1StandardBridge *L1StandardBridgeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.L1StandardBridgeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_L1StandardBridge *L1StandardBridgeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _L1StandardBridge.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_L1StandardBridge *L1StandardBridgeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_L1StandardBridge *L1StandardBridgeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.contract.Transact(opts, method, params...)
}

// MESSENGER is a free data retrieval call binding the contract method 0x927ede2d.
//
// Solidity: function MESSENGER() view returns(address)
func (_L1StandardBridge *L1StandardBridgeCaller) MESSENGER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _L1StandardBridge.contract.Call(opts, &out, "MESSENGER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MESSENGER is a free data retrieval call binding the contract method 0x927ede2d.
//
// Solidity: function MESSENGER() view returns(address)
func (_L1StandardBridge *L1StandardBridgeSession) MESSENGER() (common.Address, error) {
	return _L1StandardBridge.Contract.MESSENGER(&_L1StandardBridge.CallOpts)
}

// MESSENGER is a free data retrieval call binding the contract method 0x927ede2d.
//
// Solidity: function MESSENGER() view returns(address)
func (_L1StandardBridge *L1StandardBridgeCallerSession) MESSENGER() (common.Address, error) {
	return _L1StandardBridge.Contract.MESSENGER(&_L1StandardBridge.CallOpts)
}

// OTHERBRIDGE is a free data retrieval call binding the contract method 0x7f46ddb2.
//
// Solidity: function OTHER_BRIDGE() view returns(address)
func (_L1StandardBridge *L1StandardBridgeCaller) OTHERBRIDGE(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _L1StandardBridge.contract.Call(opts, &out, "OTHER_BRIDGE")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OTHERBRIDGE is a free data retrieval call binding the contract method 0x7f46ddb2.
//
// Solidity: function OTHER_BRIDGE() view returns(address)
func (_L1StandardBridge *L1StandardBridgeSession) OTHERBRIDGE() (common.Address, error) {
	return _L1StandardBridge.Contract.OTHERBRIDGE(&_L1StandardBridge.CallOpts)
}

// OTHERBRIDGE is a free data retrieval call binding the contract method 0x7f46ddb2.
//
// Solidity: function OTHER_BRIDGE() view returns(address)
func (_L1StandardBridge *L1StandardBridgeCallerSession) OTHERBRIDGE() (common.Address, error) {
	return _L1StandardBridge.Contract.OTHERBRIDGE(&_L1StandardBridge.CallOpts)
}

// Deposits is a free data retrieval call binding the contract method 0x8f601f66.
//
// Solidity: function deposits(address , address ) view returns(uint256)
func (_L1StandardBridge *L1StandardBridgeCaller) Deposits(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _L1StandardBridge.contract.Call(opts, &out, "deposits", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Deposits is a free data retrieval call binding the contract method 0x8f601f66.
//
// Solidity: function deposits(address , address ) view returns(uint256)
func (_L1StandardBridge *L1StandardBridgeSession) Deposits(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _L1StandardBridge.Contract.Deposits(&_L1StandardBridge.CallOpts, arg0, arg1)
}

// Deposits is a free data retrieval call binding the contract method 0x8f601f66.
//
// Solidity: function deposits(address , address ) view returns(uint256)
func (_L1StandardBridge *L1StandardBridgeCallerSession) Deposits(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _L1StandardBridge.Contract.Deposits(&_L1StandardBridge.CallOpts, arg0, arg1)
}

// L2TokenBridge is a free data retrieval call binding the contract method 0x91c49bf8.
//
// Solidity: function l2TokenBridge() view returns(address)
func (_L1StandardBridge *L1StandardBridgeCaller) L2TokenBridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _L1StandardBridge.contract.Call(opts, &out, "l2TokenBridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L2TokenBridge is a free data retrieval call binding the contract method 0x91c49bf8.
//
// Solidity: function l2TokenBridge() view returns(address)
func (_L1StandardBridge *L1StandardBridgeSession) L2TokenBridge() (common.Address, error) {
	return _L1StandardBridge.Contract.L2TokenBridge(&_L1StandardBridge.CallOpts)
}

// L2TokenBridge is a free data retrieval call binding the contract method 0x91c49bf8.
//
// Solidity: function l2TokenBridge() view returns(address)
func (_L1StandardBridge *L1StandardBridgeCallerSession) L2TokenBridge() (common.Address, error) {
	return _L1StandardBridge.Contract.L2TokenBridge(&_L1StandardBridge.CallOpts)
}

// Messenger is a free data retrieval call binding the contract method 0x3cb747bf.
//
// Solidity: function messenger() view returns(address)
func (_L1StandardBridge *L1StandardBridgeCaller) Messenger(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _L1StandardBridge.contract.Call(opts, &out, "messenger")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Messenger is a free data retrieval call binding the contract method 0x3cb747bf.
//
// Solidity: function messenger() view returns(address)
func (_L1StandardBridge *L1StandardBridgeSession) Messenger() (common.Address, error) {
	return _L1StandardBridge.Contract.Messenger(&_L1StandardBridge.CallOpts)
}

// Messenger is a free data retrieval call binding the contract method 0x3cb747bf.
//
// Solidity: function messenger() view returns(address)
func (_L1StandardBridge *L1StandardBridgeCallerSession) Messenger() (common.Address, error) {
	return _L1StandardBridge.Contract.Messenger(&_L1StandardBridge.CallOpts)
}

// OtherBridge is a free data retrieval call binding the contract method 0xc89701a2.
//
// Solidity: function otherBridge() view returns(address)
func (_L1StandardBridge *L1StandardBridgeCaller) OtherBridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _L1StandardBridge.contract.Call(opts, &out, "otherBridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OtherBridge is a free data retrieval call binding the contract method 0xc89701a2.
//
// Solidity: function otherBridge() view returns(address)
func (_L1StandardBridge *L1StandardBridgeSession) OtherBridge() (common.Address, error) {
	return _L1StandardBridge.Contract.OtherBridge(&_L1StandardBridge.CallOpts)
}

// OtherBridge is a free data retrieval call binding the contract method 0xc89701a2.
//
// Solidity: function otherBridge() view returns(address)
func (_L1StandardBridge *L1StandardBridgeCallerSession) OtherBridge() (common.Address, error) {
	return _L1StandardBridge.Contract.OtherBridge(&_L1StandardBridge.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_L1StandardBridge *L1StandardBridgeCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _L1StandardBridge.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_L1StandardBridge *L1StandardBridgeSession) Version() (string, error) {
	return _L1StandardBridge.Contract.Version(&_L1StandardBridge.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_L1StandardBridge *L1StandardBridgeCallerSession) Version() (string, error) {
	return _L1StandardBridge.Contract.Version(&_L1StandardBridge.CallOpts)
}

// BridgeERC20 is a paid mutator transaction binding the contract method 0x87087623.
//
// Solidity: function bridgeERC20(address _localToken, address _remoteToken, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) BridgeERC20(opts *bind.TransactOpts, _localToken common.Address, _remoteToken common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "bridgeERC20", _localToken, _remoteToken, _amount, _minGasLimit, _extraData)
}

// BridgeERC20 is a paid mutator transaction binding the contract method 0x87087623.
//
// Solidity: function bridgeERC20(address _localToken, address _remoteToken, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeSession) BridgeERC20(_localToken common.Address, _remoteToken common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.BridgeERC20(&_L1StandardBridge.TransactOpts, _localToken, _remoteToken, _amount, _minGasLimit, _extraData)
}

// BridgeERC20 is a paid mutator transaction binding the contract method 0x87087623.
//
// Solidity: function bridgeERC20(address _localToken, address _remoteToken, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) BridgeERC20(_localToken common.Address, _remoteToken common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.BridgeERC20(&_L1StandardBridge.TransactOpts, _localToken, _remoteToken, _amount, _minGasLimit, _extraData)
}

// BridgeERC20To is a paid mutator transaction binding the contract method 0x540abf73.
//
// Solidity: function bridgeERC20To(address _localToken, address _remoteToken, address _to, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) BridgeERC20To(opts *bind.TransactOpts, _localToken common.Address, _remoteToken common.Address, _to common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "bridgeERC20To", _localToken, _remoteToken, _to, _amount, _minGasLimit, _extraData)
}

// BridgeERC20To is a paid mutator transaction binding the contract method 0x540abf73.
//
// Solidity: function bridgeERC20To(address _localToken, address _remoteToken, address _to, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeSession) BridgeERC20To(_localToken common.Address, _remoteToken common.Address, _to common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.BridgeERC20To(&_L1StandardBridge.TransactOpts, _localToken, _remoteToken, _to, _amount, _minGasLimit, _extraData)
}

// BridgeERC20To is a paid mutator transaction binding the contract method 0x540abf73.
//
// Solidity: function bridgeERC20To(address _localToken, address _remoteToken, address _to, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) BridgeERC20To(_localToken common.Address, _remoteToken common.Address, _to common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.BridgeERC20To(&_L1StandardBridge.TransactOpts, _localToken, _remoteToken, _to, _amount, _minGasLimit, _extraData)
}

// BridgeETH is a paid mutator transaction binding the contract method 0x09fc8843.
//
// Solidity: function bridgeETH(uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) BridgeETH(opts *bind.TransactOpts, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "bridgeETH", _minGasLimit, _extraData)
}

// BridgeETH is a paid mutator transaction binding the contract method 0x09fc8843.
//
// Solidity: function bridgeETH(uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeSession) BridgeETH(_minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.BridgeETH(&_L1StandardBridge.TransactOpts, _minGasLimit, _extraData)
}

// BridgeETH is a paid mutator transaction binding the contract method 0x09fc8843.
//
// Solidity: function bridgeETH(uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) BridgeETH(_minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.BridgeETH(&_L1StandardBridge.TransactOpts, _minGasLimit, _extraData)
}

// BridgeETHTo is a paid mutator transaction binding the contract method 0xe11013dd.
//
// Solidity: function bridgeETHTo(address _to, uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) BridgeETHTo(opts *bind.TransactOpts, _to common.Address, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "bridgeETHTo", _to, _minGasLimit, _extraData)
}

// BridgeETHTo is a paid mutator transaction binding the contract method 0xe11013dd.
//
// Solidity: function bridgeETHTo(address _to, uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeSession) BridgeETHTo(_to common.Address, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.BridgeETHTo(&_L1StandardBridge.TransactOpts, _to, _minGasLimit, _extraData)
}

// BridgeETHTo is a paid mutator transaction binding the contract method 0xe11013dd.
//
// Solidity: function bridgeETHTo(address _to, uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) BridgeETHTo(_to common.Address, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.BridgeETHTo(&_L1StandardBridge.TransactOpts, _to, _minGasLimit, _extraData)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x58a997f6.
//
// Solidity: function depositERC20(address _l1Token, address _l2Token, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) DepositERC20(opts *bind.TransactOpts, _l1Token common.Address, _l2Token common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "depositERC20", _l1Token, _l2Token, _amount, _minGasLimit, _extraData)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x58a997f6.
//
// Solidity: function depositERC20(address _l1Token, address _l2Token, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeSession) DepositERC20(_l1Token common.Address, _l2Token common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.DepositERC20(&_L1StandardBridge.TransactOpts, _l1Token, _l2Token, _amount, _minGasLimit, _extraData)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x58a997f6.
//
// Solidity: function depositERC20(address _l1Token, address _l2Token, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) DepositERC20(_l1Token common.Address, _l2Token common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.DepositERC20(&_L1StandardBridge.TransactOpts, _l1Token, _l2Token, _amount, _minGasLimit, _extraData)
}

// DepositERC20To is a paid mutator transaction binding the contract method 0x838b2520.
//
// Solidity: function depositERC20To(address _l1Token, address _l2Token, address _to, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) DepositERC20To(opts *bind.TransactOpts, _l1Token common.Address, _l2Token common.Address, _to common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "depositERC20To", _l1Token, _l2Token, _to, _amount, _minGasLimit, _extraData)
}

// DepositERC20To is a paid mutator transaction binding the contract method 0x838b2520.
//
// Solidity: function depositERC20To(address _l1Token, address _l2Token, address _to, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeSession) DepositERC20To(_l1Token common.Address, _l2Token common.Address, _to common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.DepositERC20To(&_L1StandardBridge.TransactOpts, _l1Token, _l2Token, _to, _amount, _minGasLimit, _extraData)
}

// DepositERC20To is a paid mutator transaction binding the contract method 0x838b2520.
//
// Solidity: function depositERC20To(address _l1Token, address _l2Token, address _to, uint256 _amount, uint32 _minGasLimit, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) DepositERC20To(_l1Token common.Address, _l2Token common.Address, _to common.Address, _amount *big.Int, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.DepositERC20To(&_L1StandardBridge.TransactOpts, _l1Token, _l2Token, _to, _amount, _minGasLimit, _extraData)
}

// DepositETH is a paid mutator transaction binding the contract method 0xb1a1a882.
//
// Solidity: function depositETH(uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) DepositETH(opts *bind.TransactOpts, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "depositETH", _minGasLimit, _extraData)
}

// DepositETH is a paid mutator transaction binding the contract method 0xb1a1a882.
//
// Solidity: function depositETH(uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeSession) DepositETH(_minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.DepositETH(&_L1StandardBridge.TransactOpts, _minGasLimit, _extraData)
}

// DepositETH is a paid mutator transaction binding the contract method 0xb1a1a882.
//
// Solidity: function depositETH(uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) DepositETH(_minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.DepositETH(&_L1StandardBridge.TransactOpts, _minGasLimit, _extraData)
}

// DepositETHTo is a paid mutator transaction binding the contract method 0x9a2ac6d5.
//
// Solidity: function depositETHTo(address _to, uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) DepositETHTo(opts *bind.TransactOpts, _to common.Address, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "depositETHTo", _to, _minGasLimit, _extraData)
}

// DepositETHTo is a paid mutator transaction binding the contract method 0x9a2ac6d5.
//
// Solidity: function depositETHTo(address _to, uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeSession) DepositETHTo(_to common.Address, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.DepositETHTo(&_L1StandardBridge.TransactOpts, _to, _minGasLimit, _extraData)
}

// DepositETHTo is a paid mutator transaction binding the contract method 0x9a2ac6d5.
//
// Solidity: function depositETHTo(address _to, uint32 _minGasLimit, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) DepositETHTo(_to common.Address, _minGasLimit uint32, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.DepositETHTo(&_L1StandardBridge.TransactOpts, _to, _minGasLimit, _extraData)
}

// FinalizeBridgeERC20 is a paid mutator transaction binding the contract method 0x0166a07a.
//
// Solidity: function finalizeBridgeERC20(address _localToken, address _remoteToken, address _from, address _to, uint256 _amount, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) FinalizeBridgeERC20(opts *bind.TransactOpts, _localToken common.Address, _remoteToken common.Address, _from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "finalizeBridgeERC20", _localToken, _remoteToken, _from, _to, _amount, _extraData)
}

// FinalizeBridgeERC20 is a paid mutator transaction binding the contract method 0x0166a07a.
//
// Solidity: function finalizeBridgeERC20(address _localToken, address _remoteToken, address _from, address _to, uint256 _amount, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeSession) FinalizeBridgeERC20(_localToken common.Address, _remoteToken common.Address, _from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.FinalizeBridgeERC20(&_L1StandardBridge.TransactOpts, _localToken, _remoteToken, _from, _to, _amount, _extraData)
}

// FinalizeBridgeERC20 is a paid mutator transaction binding the contract method 0x0166a07a.
//
// Solidity: function finalizeBridgeERC20(address _localToken, address _remoteToken, address _from, address _to, uint256 _amount, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) FinalizeBridgeERC20(_localToken common.Address, _remoteToken common.Address, _from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.FinalizeBridgeERC20(&_L1StandardBridge.TransactOpts, _localToken, _remoteToken, _from, _to, _amount, _extraData)
}

// FinalizeBridgeETH is a paid mutator transaction binding the contract method 0x1635f5fd.
//
// Solidity: function finalizeBridgeETH(address _from, address _to, uint256 _amount, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) FinalizeBridgeETH(opts *bind.TransactOpts, _from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "finalizeBridgeETH", _from, _to, _amount, _extraData)
}

// FinalizeBridgeETH is a paid mutator transaction binding the contract method 0x1635f5fd.
//
// Solidity: function finalizeBridgeETH(address _from, address _to, uint256 _amount, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeSession) FinalizeBridgeETH(_from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.FinalizeBridgeETH(&_L1StandardBridge.TransactOpts, _from, _to, _amount, _extraData)
}

// FinalizeBridgeETH is a paid mutator transaction binding the contract method 0x1635f5fd.
//
// Solidity: function finalizeBridgeETH(address _from, address _to, uint256 _amount, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) FinalizeBridgeETH(_from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.FinalizeBridgeETH(&_L1StandardBridge.TransactOpts, _from, _to, _amount, _extraData)
}

// FinalizeERC20Withdrawal is a paid mutator transaction binding the contract method 0xa9f9e675.
//
// Solidity: function finalizeERC20Withdrawal(address _l1Token, address _l2Token, address _from, address _to, uint256 _amount, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) FinalizeERC20Withdrawal(opts *bind.TransactOpts, _l1Token common.Address, _l2Token common.Address, _from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "finalizeERC20Withdrawal", _l1Token, _l2Token, _from, _to, _amount, _extraData)
}

// FinalizeERC20Withdrawal is a paid mutator transaction binding the contract method 0xa9f9e675.
//
// Solidity: function finalizeERC20Withdrawal(address _l1Token, address _l2Token, address _from, address _to, uint256 _amount, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeSession) FinalizeERC20Withdrawal(_l1Token common.Address, _l2Token common.Address, _from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.FinalizeERC20Withdrawal(&_L1StandardBridge.TransactOpts, _l1Token, _l2Token, _from, _to, _amount, _extraData)
}

// FinalizeERC20Withdrawal is a paid mutator transaction binding the contract method 0xa9f9e675.
//
// Solidity: function finalizeERC20Withdrawal(address _l1Token, address _l2Token, address _from, address _to, uint256 _amount, bytes _extraData) returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) FinalizeERC20Withdrawal(_l1Token common.Address, _l2Token common.Address, _from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.FinalizeERC20Withdrawal(&_L1StandardBridge.TransactOpts, _l1Token, _l2Token, _from, _to, _amount, _extraData)
}

// FinalizeETHWithdrawal is a paid mutator transaction binding the contract method 0x1532ec34.
//
// Solidity: function finalizeETHWithdrawal(address _from, address _to, uint256 _amount, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) FinalizeETHWithdrawal(opts *bind.TransactOpts, _from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "finalizeETHWithdrawal", _from, _to, _amount, _extraData)
}

// FinalizeETHWithdrawal is a paid mutator transaction binding the contract method 0x1532ec34.
//
// Solidity: function finalizeETHWithdrawal(address _from, address _to, uint256 _amount, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeSession) FinalizeETHWithdrawal(_from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.FinalizeETHWithdrawal(&_L1StandardBridge.TransactOpts, _from, _to, _amount, _extraData)
}

// FinalizeETHWithdrawal is a paid mutator transaction binding the contract method 0x1532ec34.
//
// Solidity: function finalizeETHWithdrawal(address _from, address _to, uint256 _amount, bytes _extraData) payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) FinalizeETHWithdrawal(_from common.Address, _to common.Address, _amount *big.Int, _extraData []byte) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.FinalizeETHWithdrawal(&_L1StandardBridge.TransactOpts, _from, _to, _amount, _extraData)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _messenger) returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) Initialize(opts *bind.TransactOpts, _messenger common.Address) (*types.Transaction, error) {
	return _L1StandardBridge.contract.Transact(opts, "initialize", _messenger)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _messenger) returns()
func (_L1StandardBridge *L1StandardBridgeSession) Initialize(_messenger common.Address) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.Initialize(&_L1StandardBridge.TransactOpts, _messenger)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _messenger) returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) Initialize(_messenger common.Address) (*types.Transaction, error) {
	return _L1StandardBridge.Contract.Initialize(&_L1StandardBridge.TransactOpts, _messenger)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L1StandardBridge.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_L1StandardBridge *L1StandardBridgeSession) Receive() (*types.Transaction, error) {
	return _L1StandardBridge.Contract.Receive(&_L1StandardBridge.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_L1StandardBridge *L1StandardBridgeTransactorSession) Receive() (*types.Transaction, error) {
	return _L1StandardBridge.Contract.Receive(&_L1StandardBridge.TransactOpts)
}

// L1StandardBridgeERC20BridgeFinalizedIterator is returned from FilterERC20BridgeFinalized and is used to iterate over the raw logs and unpacked data for ERC20BridgeFinalized events raised by the L1StandardBridge contract.
type L1StandardBridgeERC20BridgeFinalizedIterator struct {
	Event *L1StandardBridgeERC20BridgeFinalized // Event containing the contract specifics and raw log

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
func (it *L1StandardBridgeERC20BridgeFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1StandardBridgeERC20BridgeFinalized)
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
		it.Event = new(L1StandardBridgeERC20BridgeFinalized)
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
func (it *L1StandardBridgeERC20BridgeFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1StandardBridgeERC20BridgeFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1StandardBridgeERC20BridgeFinalized represents a ERC20BridgeFinalized event raised by the L1StandardBridge contract.
type L1StandardBridgeERC20BridgeFinalized struct {
	LocalToken  common.Address
	RemoteToken common.Address
	From        common.Address
	To          common.Address
	Amount      *big.Int
	ExtraData   []byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterERC20BridgeFinalized is a free log retrieval operation binding the contract event 0xd59c65b35445225835c83f50b6ede06a7be047d22e357073e250d9af537518cd.
//
// Solidity: event ERC20BridgeFinalized(address indexed localToken, address indexed remoteToken, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) FilterERC20BridgeFinalized(opts *bind.FilterOpts, localToken []common.Address, remoteToken []common.Address, from []common.Address) (*L1StandardBridgeERC20BridgeFinalizedIterator, error) {

	var localTokenRule []interface{}
	for _, localTokenItem := range localToken {
		localTokenRule = append(localTokenRule, localTokenItem)
	}
	var remoteTokenRule []interface{}
	for _, remoteTokenItem := range remoteToken {
		remoteTokenRule = append(remoteTokenRule, remoteTokenItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _L1StandardBridge.contract.FilterLogs(opts, "ERC20BridgeFinalized", localTokenRule, remoteTokenRule, fromRule)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeERC20BridgeFinalizedIterator{contract: _L1StandardBridge.contract, event: "ERC20BridgeFinalized", logs: logs, sub: sub}, nil
}

// WatchERC20BridgeFinalized is a free log subscription operation binding the contract event 0xd59c65b35445225835c83f50b6ede06a7be047d22e357073e250d9af537518cd.
//
// Solidity: event ERC20BridgeFinalized(address indexed localToken, address indexed remoteToken, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) WatchERC20BridgeFinalized(opts *bind.WatchOpts, sink chan<- *L1StandardBridgeERC20BridgeFinalized, localToken []common.Address, remoteToken []common.Address, from []common.Address) (event.Subscription, error) {

	var localTokenRule []interface{}
	for _, localTokenItem := range localToken {
		localTokenRule = append(localTokenRule, localTokenItem)
	}
	var remoteTokenRule []interface{}
	for _, remoteTokenItem := range remoteToken {
		remoteTokenRule = append(remoteTokenRule, remoteTokenItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _L1StandardBridge.contract.WatchLogs(opts, "ERC20BridgeFinalized", localTokenRule, remoteTokenRule, fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1StandardBridgeERC20BridgeFinalized)
				if err := _L1StandardBridge.contract.UnpackLog(event, "ERC20BridgeFinalized", log); err != nil {
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

// ParseERC20BridgeFinalized is a log parse operation binding the contract event 0xd59c65b35445225835c83f50b6ede06a7be047d22e357073e250d9af537518cd.
//
// Solidity: event ERC20BridgeFinalized(address indexed localToken, address indexed remoteToken, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) ParseERC20BridgeFinalized(log types.Log) (*L1StandardBridgeERC20BridgeFinalized, error) {
	event := new(L1StandardBridgeERC20BridgeFinalized)
	if err := _L1StandardBridge.contract.UnpackLog(event, "ERC20BridgeFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1StandardBridgeERC20BridgeInitiatedIterator is returned from FilterERC20BridgeInitiated and is used to iterate over the raw logs and unpacked data for ERC20BridgeInitiated events raised by the L1StandardBridge contract.
type L1StandardBridgeERC20BridgeInitiatedIterator struct {
	Event *L1StandardBridgeERC20BridgeInitiated // Event containing the contract specifics and raw log

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
func (it *L1StandardBridgeERC20BridgeInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1StandardBridgeERC20BridgeInitiated)
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
		it.Event = new(L1StandardBridgeERC20BridgeInitiated)
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
func (it *L1StandardBridgeERC20BridgeInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1StandardBridgeERC20BridgeInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1StandardBridgeERC20BridgeInitiated represents a ERC20BridgeInitiated event raised by the L1StandardBridge contract.
type L1StandardBridgeERC20BridgeInitiated struct {
	LocalToken  common.Address
	RemoteToken common.Address
	From        common.Address
	To          common.Address
	Amount      *big.Int
	ExtraData   []byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterERC20BridgeInitiated is a free log retrieval operation binding the contract event 0x7ff126db8024424bbfd9826e8ab82ff59136289ea440b04b39a0df1b03b9cabf.
//
// Solidity: event ERC20BridgeInitiated(address indexed localToken, address indexed remoteToken, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) FilterERC20BridgeInitiated(opts *bind.FilterOpts, localToken []common.Address, remoteToken []common.Address, from []common.Address) (*L1StandardBridgeERC20BridgeInitiatedIterator, error) {

	var localTokenRule []interface{}
	for _, localTokenItem := range localToken {
		localTokenRule = append(localTokenRule, localTokenItem)
	}
	var remoteTokenRule []interface{}
	for _, remoteTokenItem := range remoteToken {
		remoteTokenRule = append(remoteTokenRule, remoteTokenItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _L1StandardBridge.contract.FilterLogs(opts, "ERC20BridgeInitiated", localTokenRule, remoteTokenRule, fromRule)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeERC20BridgeInitiatedIterator{contract: _L1StandardBridge.contract, event: "ERC20BridgeInitiated", logs: logs, sub: sub}, nil
}

// WatchERC20BridgeInitiated is a free log subscription operation binding the contract event 0x7ff126db8024424bbfd9826e8ab82ff59136289ea440b04b39a0df1b03b9cabf.
//
// Solidity: event ERC20BridgeInitiated(address indexed localToken, address indexed remoteToken, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) WatchERC20BridgeInitiated(opts *bind.WatchOpts, sink chan<- *L1StandardBridgeERC20BridgeInitiated, localToken []common.Address, remoteToken []common.Address, from []common.Address) (event.Subscription, error) {

	var localTokenRule []interface{}
	for _, localTokenItem := range localToken {
		localTokenRule = append(localTokenRule, localTokenItem)
	}
	var remoteTokenRule []interface{}
	for _, remoteTokenItem := range remoteToken {
		remoteTokenRule = append(remoteTokenRule, remoteTokenItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _L1StandardBridge.contract.WatchLogs(opts, "ERC20BridgeInitiated", localTokenRule, remoteTokenRule, fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1StandardBridgeERC20BridgeInitiated)
				if err := _L1StandardBridge.contract.UnpackLog(event, "ERC20BridgeInitiated", log); err != nil {
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

// ParseERC20BridgeInitiated is a log parse operation binding the contract event 0x7ff126db8024424bbfd9826e8ab82ff59136289ea440b04b39a0df1b03b9cabf.
//
// Solidity: event ERC20BridgeInitiated(address indexed localToken, address indexed remoteToken, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) ParseERC20BridgeInitiated(log types.Log) (*L1StandardBridgeERC20BridgeInitiated, error) {
	event := new(L1StandardBridgeERC20BridgeInitiated)
	if err := _L1StandardBridge.contract.UnpackLog(event, "ERC20BridgeInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1StandardBridgeERC20DepositInitiatedIterator is returned from FilterERC20DepositInitiated and is used to iterate over the raw logs and unpacked data for ERC20DepositInitiated events raised by the L1StandardBridge contract.
type L1StandardBridgeERC20DepositInitiatedIterator struct {
	Event *L1StandardBridgeERC20DepositInitiated // Event containing the contract specifics and raw log

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
func (it *L1StandardBridgeERC20DepositInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1StandardBridgeERC20DepositInitiated)
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
		it.Event = new(L1StandardBridgeERC20DepositInitiated)
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
func (it *L1StandardBridgeERC20DepositInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1StandardBridgeERC20DepositInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1StandardBridgeERC20DepositInitiated represents a ERC20DepositInitiated event raised by the L1StandardBridge contract.
type L1StandardBridgeERC20DepositInitiated struct {
	L1Token   common.Address
	L2Token   common.Address
	From      common.Address
	To        common.Address
	Amount    *big.Int
	ExtraData []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterERC20DepositInitiated is a free log retrieval operation binding the contract event 0x718594027abd4eaed59f95162563e0cc6d0e8d5b86b1c7be8b1b0ac3343d0396.
//
// Solidity: event ERC20DepositInitiated(address indexed l1Token, address indexed l2Token, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) FilterERC20DepositInitiated(opts *bind.FilterOpts, l1Token []common.Address, l2Token []common.Address, from []common.Address) (*L1StandardBridgeERC20DepositInitiatedIterator, error) {

	var l1TokenRule []interface{}
	for _, l1TokenItem := range l1Token {
		l1TokenRule = append(l1TokenRule, l1TokenItem)
	}
	var l2TokenRule []interface{}
	for _, l2TokenItem := range l2Token {
		l2TokenRule = append(l2TokenRule, l2TokenItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _L1StandardBridge.contract.FilterLogs(opts, "ERC20DepositInitiated", l1TokenRule, l2TokenRule, fromRule)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeERC20DepositInitiatedIterator{contract: _L1StandardBridge.contract, event: "ERC20DepositInitiated", logs: logs, sub: sub}, nil
}

// WatchERC20DepositInitiated is a free log subscription operation binding the contract event 0x718594027abd4eaed59f95162563e0cc6d0e8d5b86b1c7be8b1b0ac3343d0396.
//
// Solidity: event ERC20DepositInitiated(address indexed l1Token, address indexed l2Token, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) WatchERC20DepositInitiated(opts *bind.WatchOpts, sink chan<- *L1StandardBridgeERC20DepositInitiated, l1Token []common.Address, l2Token []common.Address, from []common.Address) (event.Subscription, error) {

	var l1TokenRule []interface{}
	for _, l1TokenItem := range l1Token {
		l1TokenRule = append(l1TokenRule, l1TokenItem)
	}
	var l2TokenRule []interface{}
	for _, l2TokenItem := range l2Token {
		l2TokenRule = append(l2TokenRule, l2TokenItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _L1StandardBridge.contract.WatchLogs(opts, "ERC20DepositInitiated", l1TokenRule, l2TokenRule, fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1StandardBridgeERC20DepositInitiated)
				if err := _L1StandardBridge.contract.UnpackLog(event, "ERC20DepositInitiated", log); err != nil {
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

// ParseERC20DepositInitiated is a log parse operation binding the contract event 0x718594027abd4eaed59f95162563e0cc6d0e8d5b86b1c7be8b1b0ac3343d0396.
//
// Solidity: event ERC20DepositInitiated(address indexed l1Token, address indexed l2Token, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) ParseERC20DepositInitiated(log types.Log) (*L1StandardBridgeERC20DepositInitiated, error) {
	event := new(L1StandardBridgeERC20DepositInitiated)
	if err := _L1StandardBridge.contract.UnpackLog(event, "ERC20DepositInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1StandardBridgeERC20WithdrawalFinalizedIterator is returned from FilterERC20WithdrawalFinalized and is used to iterate over the raw logs and unpacked data for ERC20WithdrawalFinalized events raised by the L1StandardBridge contract.
type L1StandardBridgeERC20WithdrawalFinalizedIterator struct {
	Event *L1StandardBridgeERC20WithdrawalFinalized // Event containing the contract specifics and raw log

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
func (it *L1StandardBridgeERC20WithdrawalFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1StandardBridgeERC20WithdrawalFinalized)
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
		it.Event = new(L1StandardBridgeERC20WithdrawalFinalized)
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
func (it *L1StandardBridgeERC20WithdrawalFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1StandardBridgeERC20WithdrawalFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1StandardBridgeERC20WithdrawalFinalized represents a ERC20WithdrawalFinalized event raised by the L1StandardBridge contract.
type L1StandardBridgeERC20WithdrawalFinalized struct {
	L1Token   common.Address
	L2Token   common.Address
	From      common.Address
	To        common.Address
	Amount    *big.Int
	ExtraData []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterERC20WithdrawalFinalized is a free log retrieval operation binding the contract event 0x3ceee06c1e37648fcbb6ed52e17b3e1f275a1f8c7b22a84b2b84732431e046b3.
//
// Solidity: event ERC20WithdrawalFinalized(address indexed l1Token, address indexed l2Token, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) FilterERC20WithdrawalFinalized(opts *bind.FilterOpts, l1Token []common.Address, l2Token []common.Address, from []common.Address) (*L1StandardBridgeERC20WithdrawalFinalizedIterator, error) {

	var l1TokenRule []interface{}
	for _, l1TokenItem := range l1Token {
		l1TokenRule = append(l1TokenRule, l1TokenItem)
	}
	var l2TokenRule []interface{}
	for _, l2TokenItem := range l2Token {
		l2TokenRule = append(l2TokenRule, l2TokenItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _L1StandardBridge.contract.FilterLogs(opts, "ERC20WithdrawalFinalized", l1TokenRule, l2TokenRule, fromRule)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeERC20WithdrawalFinalizedIterator{contract: _L1StandardBridge.contract, event: "ERC20WithdrawalFinalized", logs: logs, sub: sub}, nil
}

// WatchERC20WithdrawalFinalized is a free log subscription operation binding the contract event 0x3ceee06c1e37648fcbb6ed52e17b3e1f275a1f8c7b22a84b2b84732431e046b3.
//
// Solidity: event ERC20WithdrawalFinalized(address indexed l1Token, address indexed l2Token, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) WatchERC20WithdrawalFinalized(opts *bind.WatchOpts, sink chan<- *L1StandardBridgeERC20WithdrawalFinalized, l1Token []common.Address, l2Token []common.Address, from []common.Address) (event.Subscription, error) {

	var l1TokenRule []interface{}
	for _, l1TokenItem := range l1Token {
		l1TokenRule = append(l1TokenRule, l1TokenItem)
	}
	var l2TokenRule []interface{}
	for _, l2TokenItem := range l2Token {
		l2TokenRule = append(l2TokenRule, l2TokenItem)
	}
	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _L1StandardBridge.contract.WatchLogs(opts, "ERC20WithdrawalFinalized", l1TokenRule, l2TokenRule, fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1StandardBridgeERC20WithdrawalFinalized)
				if err := _L1StandardBridge.contract.UnpackLog(event, "ERC20WithdrawalFinalized", log); err != nil {
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

// ParseERC20WithdrawalFinalized is a log parse operation binding the contract event 0x3ceee06c1e37648fcbb6ed52e17b3e1f275a1f8c7b22a84b2b84732431e046b3.
//
// Solidity: event ERC20WithdrawalFinalized(address indexed l1Token, address indexed l2Token, address indexed from, address to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) ParseERC20WithdrawalFinalized(log types.Log) (*L1StandardBridgeERC20WithdrawalFinalized, error) {
	event := new(L1StandardBridgeERC20WithdrawalFinalized)
	if err := _L1StandardBridge.contract.UnpackLog(event, "ERC20WithdrawalFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1StandardBridgeETHBridgeFinalizedIterator is returned from FilterETHBridgeFinalized and is used to iterate over the raw logs and unpacked data for ETHBridgeFinalized events raised by the L1StandardBridge contract.
type L1StandardBridgeETHBridgeFinalizedIterator struct {
	Event *L1StandardBridgeETHBridgeFinalized // Event containing the contract specifics and raw log

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
func (it *L1StandardBridgeETHBridgeFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1StandardBridgeETHBridgeFinalized)
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
		it.Event = new(L1StandardBridgeETHBridgeFinalized)
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
func (it *L1StandardBridgeETHBridgeFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1StandardBridgeETHBridgeFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1StandardBridgeETHBridgeFinalized represents a ETHBridgeFinalized event raised by the L1StandardBridge contract.
type L1StandardBridgeETHBridgeFinalized struct {
	From      common.Address
	To        common.Address
	Amount    *big.Int
	ExtraData []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterETHBridgeFinalized is a free log retrieval operation binding the contract event 0x31b2166ff604fc5672ea5df08a78081d2bc6d746cadce880747f3643d819e83d.
//
// Solidity: event ETHBridgeFinalized(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) FilterETHBridgeFinalized(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*L1StandardBridgeETHBridgeFinalizedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _L1StandardBridge.contract.FilterLogs(opts, "ETHBridgeFinalized", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeETHBridgeFinalizedIterator{contract: _L1StandardBridge.contract, event: "ETHBridgeFinalized", logs: logs, sub: sub}, nil
}

// WatchETHBridgeFinalized is a free log subscription operation binding the contract event 0x31b2166ff604fc5672ea5df08a78081d2bc6d746cadce880747f3643d819e83d.
//
// Solidity: event ETHBridgeFinalized(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) WatchETHBridgeFinalized(opts *bind.WatchOpts, sink chan<- *L1StandardBridgeETHBridgeFinalized, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _L1StandardBridge.contract.WatchLogs(opts, "ETHBridgeFinalized", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1StandardBridgeETHBridgeFinalized)
				if err := _L1StandardBridge.contract.UnpackLog(event, "ETHBridgeFinalized", log); err != nil {
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

// ParseETHBridgeFinalized is a log parse operation binding the contract event 0x31b2166ff604fc5672ea5df08a78081d2bc6d746cadce880747f3643d819e83d.
//
// Solidity: event ETHBridgeFinalized(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) ParseETHBridgeFinalized(log types.Log) (*L1StandardBridgeETHBridgeFinalized, error) {
	event := new(L1StandardBridgeETHBridgeFinalized)
	if err := _L1StandardBridge.contract.UnpackLog(event, "ETHBridgeFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1StandardBridgeETHBridgeInitiatedIterator is returned from FilterETHBridgeInitiated and is used to iterate over the raw logs and unpacked data for ETHBridgeInitiated events raised by the L1StandardBridge contract.
type L1StandardBridgeETHBridgeInitiatedIterator struct {
	Event *L1StandardBridgeETHBridgeInitiated // Event containing the contract specifics and raw log

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
func (it *L1StandardBridgeETHBridgeInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1StandardBridgeETHBridgeInitiated)
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
		it.Event = new(L1StandardBridgeETHBridgeInitiated)
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
func (it *L1StandardBridgeETHBridgeInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1StandardBridgeETHBridgeInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1StandardBridgeETHBridgeInitiated represents a ETHBridgeInitiated event raised by the L1StandardBridge contract.
type L1StandardBridgeETHBridgeInitiated struct {
	From      common.Address
	To        common.Address
	Amount    *big.Int
	ExtraData []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterETHBridgeInitiated is a free log retrieval operation binding the contract event 0x2849b43074093a05396b6f2a937dee8565b15a48a7b3d4bffb732a5017380af5.
//
// Solidity: event ETHBridgeInitiated(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) FilterETHBridgeInitiated(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*L1StandardBridgeETHBridgeInitiatedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _L1StandardBridge.contract.FilterLogs(opts, "ETHBridgeInitiated", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeETHBridgeInitiatedIterator{contract: _L1StandardBridge.contract, event: "ETHBridgeInitiated", logs: logs, sub: sub}, nil
}

// WatchETHBridgeInitiated is a free log subscription operation binding the contract event 0x2849b43074093a05396b6f2a937dee8565b15a48a7b3d4bffb732a5017380af5.
//
// Solidity: event ETHBridgeInitiated(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) WatchETHBridgeInitiated(opts *bind.WatchOpts, sink chan<- *L1StandardBridgeETHBridgeInitiated, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _L1StandardBridge.contract.WatchLogs(opts, "ETHBridgeInitiated", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1StandardBridgeETHBridgeInitiated)
				if err := _L1StandardBridge.contract.UnpackLog(event, "ETHBridgeInitiated", log); err != nil {
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

// ParseETHBridgeInitiated is a log parse operation binding the contract event 0x2849b43074093a05396b6f2a937dee8565b15a48a7b3d4bffb732a5017380af5.
//
// Solidity: event ETHBridgeInitiated(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) ParseETHBridgeInitiated(log types.Log) (*L1StandardBridgeETHBridgeInitiated, error) {
	event := new(L1StandardBridgeETHBridgeInitiated)
	if err := _L1StandardBridge.contract.UnpackLog(event, "ETHBridgeInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1StandardBridgeETHDepositInitiatedIterator is returned from FilterETHDepositInitiated and is used to iterate over the raw logs and unpacked data for ETHDepositInitiated events raised by the L1StandardBridge contract.
type L1StandardBridgeETHDepositInitiatedIterator struct {
	Event *L1StandardBridgeETHDepositInitiated // Event containing the contract specifics and raw log

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
func (it *L1StandardBridgeETHDepositInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1StandardBridgeETHDepositInitiated)
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
		it.Event = new(L1StandardBridgeETHDepositInitiated)
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
func (it *L1StandardBridgeETHDepositInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1StandardBridgeETHDepositInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1StandardBridgeETHDepositInitiated represents a ETHDepositInitiated event raised by the L1StandardBridge contract.
type L1StandardBridgeETHDepositInitiated struct {
	From      common.Address
	To        common.Address
	Amount    *big.Int
	ExtraData []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterETHDepositInitiated is a free log retrieval operation binding the contract event 0x35d79ab81f2b2017e19afb5c5571778877782d7a8786f5907f93b0f4702f4f23.
//
// Solidity: event ETHDepositInitiated(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) FilterETHDepositInitiated(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*L1StandardBridgeETHDepositInitiatedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _L1StandardBridge.contract.FilterLogs(opts, "ETHDepositInitiated", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeETHDepositInitiatedIterator{contract: _L1StandardBridge.contract, event: "ETHDepositInitiated", logs: logs, sub: sub}, nil
}

// WatchETHDepositInitiated is a free log subscription operation binding the contract event 0x35d79ab81f2b2017e19afb5c5571778877782d7a8786f5907f93b0f4702f4f23.
//
// Solidity: event ETHDepositInitiated(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) WatchETHDepositInitiated(opts *bind.WatchOpts, sink chan<- *L1StandardBridgeETHDepositInitiated, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _L1StandardBridge.contract.WatchLogs(opts, "ETHDepositInitiated", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1StandardBridgeETHDepositInitiated)
				if err := _L1StandardBridge.contract.UnpackLog(event, "ETHDepositInitiated", log); err != nil {
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

// ParseETHDepositInitiated is a log parse operation binding the contract event 0x35d79ab81f2b2017e19afb5c5571778877782d7a8786f5907f93b0f4702f4f23.
//
// Solidity: event ETHDepositInitiated(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) ParseETHDepositInitiated(log types.Log) (*L1StandardBridgeETHDepositInitiated, error) {
	event := new(L1StandardBridgeETHDepositInitiated)
	if err := _L1StandardBridge.contract.UnpackLog(event, "ETHDepositInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1StandardBridgeETHWithdrawalFinalizedIterator is returned from FilterETHWithdrawalFinalized and is used to iterate over the raw logs and unpacked data for ETHWithdrawalFinalized events raised by the L1StandardBridge contract.
type L1StandardBridgeETHWithdrawalFinalizedIterator struct {
	Event *L1StandardBridgeETHWithdrawalFinalized // Event containing the contract specifics and raw log

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
func (it *L1StandardBridgeETHWithdrawalFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1StandardBridgeETHWithdrawalFinalized)
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
		it.Event = new(L1StandardBridgeETHWithdrawalFinalized)
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
func (it *L1StandardBridgeETHWithdrawalFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1StandardBridgeETHWithdrawalFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1StandardBridgeETHWithdrawalFinalized represents a ETHWithdrawalFinalized event raised by the L1StandardBridge contract.
type L1StandardBridgeETHWithdrawalFinalized struct {
	From      common.Address
	To        common.Address
	Amount    *big.Int
	ExtraData []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterETHWithdrawalFinalized is a free log retrieval operation binding the contract event 0x2ac69ee804d9a7a0984249f508dfab7cb2534b465b6ce1580f99a38ba9c5e631.
//
// Solidity: event ETHWithdrawalFinalized(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) FilterETHWithdrawalFinalized(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*L1StandardBridgeETHWithdrawalFinalizedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _L1StandardBridge.contract.FilterLogs(opts, "ETHWithdrawalFinalized", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeETHWithdrawalFinalizedIterator{contract: _L1StandardBridge.contract, event: "ETHWithdrawalFinalized", logs: logs, sub: sub}, nil
}

// WatchETHWithdrawalFinalized is a free log subscription operation binding the contract event 0x2ac69ee804d9a7a0984249f508dfab7cb2534b465b6ce1580f99a38ba9c5e631.
//
// Solidity: event ETHWithdrawalFinalized(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) WatchETHWithdrawalFinalized(opts *bind.WatchOpts, sink chan<- *L1StandardBridgeETHWithdrawalFinalized, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _L1StandardBridge.contract.WatchLogs(opts, "ETHWithdrawalFinalized", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1StandardBridgeETHWithdrawalFinalized)
				if err := _L1StandardBridge.contract.UnpackLog(event, "ETHWithdrawalFinalized", log); err != nil {
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

// ParseETHWithdrawalFinalized is a log parse operation binding the contract event 0x2ac69ee804d9a7a0984249f508dfab7cb2534b465b6ce1580f99a38ba9c5e631.
//
// Solidity: event ETHWithdrawalFinalized(address indexed from, address indexed to, uint256 amount, bytes extraData)
func (_L1StandardBridge *L1StandardBridgeFilterer) ParseETHWithdrawalFinalized(log types.Log) (*L1StandardBridgeETHWithdrawalFinalized, error) {
	event := new(L1StandardBridgeETHWithdrawalFinalized)
	if err := _L1StandardBridge.contract.UnpackLog(event, "ETHWithdrawalFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1StandardBridgeInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the L1StandardBridge contract.
type L1StandardBridgeInitializedIterator struct {
	Event *L1StandardBridgeInitialized // Event containing the contract specifics and raw log

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
func (it *L1StandardBridgeInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1StandardBridgeInitialized)
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
		it.Event = new(L1StandardBridgeInitialized)
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
func (it *L1StandardBridgeInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1StandardBridgeInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1StandardBridgeInitialized represents a Initialized event raised by the L1StandardBridge contract.
type L1StandardBridgeInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_L1StandardBridge *L1StandardBridgeFilterer) FilterInitialized(opts *bind.FilterOpts) (*L1StandardBridgeInitializedIterator, error) {

	logs, sub, err := _L1StandardBridge.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &L1StandardBridgeInitializedIterator{contract: _L1StandardBridge.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_L1StandardBridge *L1StandardBridgeFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *L1StandardBridgeInitialized) (event.Subscription, error) {

	logs, sub, err := _L1StandardBridge.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1StandardBridgeInitialized)
				if err := _L1StandardBridge.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_L1StandardBridge *L1StandardBridgeFilterer) ParseInitialized(log types.Log) (*L1StandardBridgeInitialized, error) {
	event := new(L1StandardBridgeInitialized)
	if err := _L1StandardBridge.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
