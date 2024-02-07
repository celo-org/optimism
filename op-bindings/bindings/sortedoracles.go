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

// SortedOraclesMetaData contains all meta data concerning the SortedOracles contract.
var SortedOraclesMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"test\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"addOracle\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"oracleAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getOracles\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRates\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"},{\"name\":\"\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"},{\"name\":\"\",\"type\":\"uint8[]\",\"internalType\":\"enumSortedLinkedListWithMedian.MedianRelation[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTimestamps\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"},{\"name\":\"\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"},{\"name\":\"\",\"type\":\"uint8[]\",\"internalType\":\"enumSortedLinkedListWithMedian.MedianRelation[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTokenReportExpirySeconds\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVersionNumber\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_reportExpirySeconds\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialized\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOldestReportExpired\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOracle\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"medianRate\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"medianTimestamp\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numRates\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"numTimestamps\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"oracles\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"removeExpiredReports\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"n\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"removeOracle\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"oracleAddress\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"report\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"lesserKey\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"greaterKey\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"reportExpirySeconds\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setReportExpiry\",\"inputs\":[{\"name\":\"_reportExpirySeconds\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTokenReportExpiry\",\"inputs\":[{\"name\":\"_token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_reportExpirySeconds\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"tokenReportExpirySeconds\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"MedianUpdated\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OracleAdded\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oracleAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OracleRemoved\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oracleAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OracleReportRemoved\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oracle\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OracleReported\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oracle\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReportExpirySet\",\"inputs\":[{\"name\":\"reportExpiry\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TokenReportExpirySet\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"reportExpiry\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false}]",
	Bin: "0x60806040523480156200001157600080fd5b5060405162002c3d38038062002c3d8339810160408190526200003491620000b2565b80620000403362000062565b806200005a576000805460ff60a01b1916600160a01b1790555b5050620000dd565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b600060208284031215620000c557600080fd5b81518015158114620000d657600080fd5b9392505050565b612b5080620000ed6000396000f3fe608060405234801561001057600080fd5b50600436106101a35760003560e01c80638da5cb5b116100ee578063ebc1d6bb11610097578063f2fde38b11610071578063f2fde38b146103f5578063fc20935d14610408578063fe4b84df1461041b578063ffe736bf1461042e57600080fd5b8063ebc1d6bb146103a7578063ef90e1b0146103ba578063f0ca4adb146103e257600080fd5b8063b9292158116100c8578063b92921581461036e578063bbc66a9414610381578063dd34ca3b1461039457600080fd5b80638da5cb5b146102fc5780638e7492811461033b578063a00a8b2c1461035b57600080fd5b806353a57297116101505780636deb67991161012a5780636deb6799146102ce578063715018a6146102e157806380e50744146102e957600080fd5b806353a572971461028057806354255be0146102955780636dd6ef0c146102bb57600080fd5b80632e86bc01116101815780632e86bc0114610229578063370c998e14610249578063493a353c1461027757600080fd5b806302f55b61146101a8578063071b48fc146101d3578063158ef93e146101f4575b600080fd5b6101bb6101b636600461255a565b61046d565b6040516101ca939291906125c8565b60405180910390f35b6101e66101e136600461255a565b61055c565b6040519081526020016101ca565b6000546102199074010000000000000000000000000000000000000000900460ff1681565b60405190151581526020016101ca565b6101e661023736600461255a565b60066020526000908152604090205481565b61021961025736600461268f565b600360209081526000928352604080842090915290825290205460ff1681565b6101e660055481565b61029361028e3660046126c8565b610616565b005b6001806002806040805194855260208501939093529183015260608201526080016101ca565b6101e66102c936600461255a565b6109ab565b6101e66102dc36600461255a565b610a22565b610293610a7f565b6102936102f7366004612709565b610a93565b60005473ffffffffffffffffffffffffffffffffffffffff165b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016101ca565b61034e61034936600461255a565b6111bc565b6040516101ca919061275c565b61031661036936600461276f565b61124c565b6101bb61037c36600461255a565b611291565b6101e661038f36600461255a565b611310565b6102936103a236600461276f565b611387565b6102936103b536600461279b565b61153b565b6103cd6103c836600461255a565b61169f565b604080519283526020830191909152016101ca565b6102936103f036600461268f565b611783565b61029361040336600461255a565b611981565b61029361041636600461276f565b611a38565b61029361042936600461279b565b611be9565b61044161043c36600461255a565b611cbf565b60408051921515835273ffffffffffffffffffffffffffffffffffffffff9091166020830152016101ca565b73ffffffffffffffffffffffffffffffffffffffff81166000908152600160205260409081902090517f6cfa38730000000000000000000000000000000000000000000000000000000081526060918291829173ed477a99035d0c1e11369f1d7a4e587893cc002b91636cfa3873916104ec9160040190815260200190565b600060405180830381865af4158015610509573d6000803e3d6000fd5b505050506040513d6000823e601f3d9081017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016820160405261054f919081019061292a565b9250925092509193909250565b73ffffffffffffffffffffffffffffffffffffffff811660009081526002602052604080822090517f59d556a8000000000000000000000000000000000000000000000000000000008152600481019190915273ed477a99035d0c1e11369f1d7a4e587893cc002b906359d556a8906024015b602060405180830381865af41580156105ec573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106109190612a13565b92915050565b61061e611ee3565b73ffffffffffffffffffffffffffffffffffffffff831615801590610658575073ffffffffffffffffffffffffffffffffffffffff821615155b8015610688575073ffffffffffffffffffffffffffffffffffffffff831660009081526004602052604090205481105b80156106ef575073ffffffffffffffffffffffffffffffffffffffff8381166000908152600460205260409020805491841691839081106106cb576106cb612a2c565b60009182526020909120015473ffffffffffffffffffffffffffffffffffffffff16145b6107a6576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152605660248201527f746f6b656e2061646472206e756c6c206f72206f7261636c652061646472206e60448201527f756c6c206f7220696e646578206f6620746f6b656e206f7261636c65206e6f7460648201527f206d617070656420746f206f7261636c65206164647200000000000000000000608482015260a4015b60405180910390fd5b73ffffffffffffffffffffffffffffffffffffffff8084166000818152600360209081526040808320948716835293815283822080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00169055918152600490915220805461081790600190612a8a565b8154811061082757610827612a2c565b600091825260208083209091015473ffffffffffffffffffffffffffffffffffffffff86811684526004909252604090922080549190921691908390811061087157610871612a2c565b600091825260208083209190910180547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff94851617905591851681526004909152604090208054806108db576108db612a9d565b60008281526020902081017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff90810180547fffffffffffffffffffffffff000000000000000000000000000000000000000016905501905561093d8383611f64565b1561094c5761094c8383612105565b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f6dc84b66cc948d847632b9d829f7cb1cb904fbf2c084554a9bc22ad9d845334060405160405180910390a3505050565b73ffffffffffffffffffffffffffffffffffffffff811660009081526002602052604080822090517f6eafa6c3000000000000000000000000000000000000000000000000000000008152600481019190915273ed477a99035d0c1e11369f1d7a4e587893cc002b90636eafa6c3906024016105cf565b73ffffffffffffffffffffffffffffffffffffffff81166000908152600660205260408120548103610a5657505060055490565b5073ffffffffffffffffffffffffffffffffffffffff1660009081526006602052604090205490565b610a87611ee3565b610a9160006124c3565b565b73ffffffffffffffffffffffffffffffffffffffff84166000908152600360209081526040808320338452909152902054849060ff16610b55576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602760248201527f73656e64657220776173206e6f7420616e206f7261636c6520666f7220746f6b60448201527f656e206164647200000000000000000000000000000000000000000000000000606482015260840161079d565b73ffffffffffffffffffffffffffffffffffffffff851660009081526001602052604080822090517f59d556a8000000000000000000000000000000000000000000000000000000008152600481019190915273ed477a99035d0c1e11369f1d7a4e587893cc002b906359d556a890602401602060405180830381865af4158015610be4573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610c089190612a13565b73ffffffffffffffffffffffffffffffffffffffff87166000908152600160205260409081902090517f95073a79000000000000000000000000000000000000000000000000000000008152600481019190915233602482015290915073ed477a99035d0c1e11369f1d7a4e587893cc002b906395073a7990604401602060405180830381865af4158015610ca1573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610cc59190612acc565b15610e395773ffffffffffffffffffffffffffffffffffffffff8681166000908152600160205260409081902090517f832a21470000000000000000000000000000000000000000000000000000000081526004810191909152336024820152604481018790528582166064820152908416608482015273ed477a99035d0c1e11369f1d7a4e587893cc002b9063832a21479060a40160006040518083038186803b158015610d7357600080fd5b505af4158015610d87573d6000803e3d6000fd5b5050505073ffffffffffffffffffffffffffffffffffffffff86166000908152600260205260409081902090517fc1e728e9000000000000000000000000000000000000000000000000000000008152600481019190915233602482015273ed477a99035d0c1e11369f1d7a4e587893cc002b9063c1e728e99060440160006040518083038186803b158015610e1c57600080fd5b505af4158015610e30573d6000803e3d6000fd5b50505050610efb565b73ffffffffffffffffffffffffffffffffffffffff8681166000908152600160205260409081902090517fd4a092720000000000000000000000000000000000000000000000000000000081526004810191909152336024820152604481018790528582166064820152908416608482015273ed477a99035d0c1e11369f1d7a4e587893cc002b9063d4a092729060a40160006040518083038186803b158015610ee257600080fd5b505af4158015610ef6573d6000803e3d6000fd5b505050505b73ffffffffffffffffffffffffffffffffffffffff86166000908152600260205260409081902090517f0944c5940000000000000000000000000000000000000000000000000000000081526004810182905273ed477a99035d0c1e11369f1d7a4e587893cc002b9163d4a0927291339042908590630944c59490602401602060405180830381865af4158015610f96573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610fba9190612aee565b6040517fffffffff0000000000000000000000000000000000000000000000000000000060e087901b168152600481019490945273ffffffffffffffffffffffffffffffffffffffff928316602485015260448401919091521660648201526000608482015260a40160006040518083038186803b15801561103b57600080fd5b505af415801561104f573d6000803e3d6000fd5b5050604080514281526020810189905233935073ffffffffffffffffffffffffffffffffffffffff8a1692507f7cebb17173a9ed273d2b7538f64395c0ebf352ff743f1cf8ce66b437a6144213910160405180910390a373ffffffffffffffffffffffffffffffffffffffff861660009081526001602052604080822090517f59d556a8000000000000000000000000000000000000000000000000000000008152600481019190915273ed477a99035d0c1e11369f1d7a4e587893cc002b906359d556a890602401602060405180830381865af4158015611135573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906111599190612a13565b90508181146111b3578673ffffffffffffffffffffffffffffffffffffffff167fa9981ebfc3b766a742486e898f54959b050a66006dbce1a4155c1f84a08bcf41826040516111aa91815260200190565b60405180910390a25b50505050505050565b73ffffffffffffffffffffffffffffffffffffffff811660009081526004602090815260409182902080548351818402810184019094528084526060939283018282801561124057602002820191906000526020600020905b815473ffffffffffffffffffffffffffffffffffffffff168152600190910190602001808311611215575b50505050509050919050565b6004602052816000526040600020818154811061126857600080fd5b60009182526020909120015473ffffffffffffffffffffffffffffffffffffffff169150829050565b73ffffffffffffffffffffffffffffffffffffffff81166000908152600260205260409081902090517f6cfa38730000000000000000000000000000000000000000000000000000000081526060918291829173ed477a99035d0c1e11369f1d7a4e587893cc002b91636cfa3873916104ec9160040190815260200190565b73ffffffffffffffffffffffffffffffffffffffff811660009081526001602052604080822090517f6eafa6c3000000000000000000000000000000000000000000000000000000008152600481019190915273ed477a99035d0c1e11369f1d7a4e587893cc002b90636eafa6c3906024016105cf565b73ffffffffffffffffffffffffffffffffffffffff82161580159061145e575073ffffffffffffffffffffffffffffffffffffffff82166000908152600260205260409081902090517f6eafa6c3000000000000000000000000000000000000000000000000000000008152600481019190915273ed477a99035d0c1e11369f1d7a4e587893cc002b90636eafa6c390602401602060405180830381865af4158015611437573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061145b9190612a13565b81105b6114ea576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603460248201527f746f6b656e2061646472206e756c6c206f7220747279696e6720746f2072656d60448201527f6f766520746f6f206d616e79207265706f727473000000000000000000000000606482015260840161079d565b60005b818110156115365760008061150185611cbf565b91509150811561151a576115158582612105565b611521565b5050505050565b5050808061152e90612b0b565b9150506114ed565b505050565b611543611ee3565b600081116115d3576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602160248201527f7265706f727420657870697279207365636f6e6473206d757374206265203e2060448201527f3000000000000000000000000000000000000000000000000000000000000000606482015260840161079d565b6005548103611664576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f7265706f72744578706972795365636f6e6473206861736e2774206368616e6760448201527f6564000000000000000000000000000000000000000000000000000000000000606482015260840161079d565b60058190556040518181527fc68a9b88effd8a11611ff410efbc83569f0031b7bc70dd455b61344c7f0a042f9060200160405180910390a150565b73ffffffffffffffffffffffffffffffffffffffff811660009081526001602052604080822090517f59d556a8000000000000000000000000000000000000000000000000000000008152829173ed477a99035d0c1e11369f1d7a4e587893cc002b916359d556a8916117189160040190815260200190565b602060405180830381865af4158015611735573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906117599190612a13565b61176284611310565b156117775769d3c21bcecceda100000061177a565b60005b91509150915091565b61178b611ee3565b73ffffffffffffffffffffffffffffffffffffffff8216158015906117c5575073ffffffffffffffffffffffffffffffffffffffff811615155b8015611804575073ffffffffffffffffffffffffffffffffffffffff80831660009081526003602090815260408083209385168352929052205460ff16155b6118b6576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152605a60248201527f746f6b656e206164647220776173206e756c6c206f72206f7261636c6520616460448201527f647220776173206e756c6c206f72206f7261636c652061646472206973206e6f60648201527f7420616e206f7261636c6520666f7220746f6b656e2061646472000000000000608482015260a40161079d565b73ffffffffffffffffffffffffffffffffffffffff808316600081815260036020908152604080832094861680845294825280832080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff001660019081179091558484526004835281842080549182018155845291832090910180547fffffffffffffffffffffffff00000000000000000000000000000000000000001685179055517f828d2be040dede7698182e08dfa8bfbd663c879aee772509c4a2bd961d0ed43f9190a35050565b611989611ee3565b73ffffffffffffffffffffffffffffffffffffffff8116611a2c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f6464726573730000000000000000000000000000000000000000000000000000606482015260840161079d565b611a35816124c3565b50565b611a40611ee3565b60008111611ad0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602160248201527f7265706f727420657870697279207365636f6e6473206d757374206265203e2060448201527f3000000000000000000000000000000000000000000000000000000000000000606482015260840161079d565b73ffffffffffffffffffffffffffffffffffffffff82166000908152600660205260409020548103611b84576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602860248201527f746f6b656e207265706f72744578706972795365636f6e6473206861736e277460448201527f206368616e676564000000000000000000000000000000000000000000000000606482015260840161079d565b73ffffffffffffffffffffffffffffffffffffffff8216600081815260066020908152604091829020849055815192835282018390527ff8324c8592dfd9991ee3e717351afe0a964605257959e3d99b0eb3d45bff9422910160405180910390a15050565b60005474010000000000000000000000000000000000000000900460ff1615611c6e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601c60248201527f636f6e747261637420616c726561647920696e697469616c697a656400000000604482015260640161079d565b600080547fffffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffff1674010000000000000000000000000000000000000000179055611cb6336124c3565b611a358161153b565b60008073ffffffffffffffffffffffffffffffffffffffff8316611d3f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601c60248201527f746f6b656e20616464726573732063616e6e6f74206265206e756c6c00000000604482015260640161079d565b73ffffffffffffffffffffffffffffffffffffffff831660009081526002602052604080822090517fd938ec7b000000000000000000000000000000000000000000000000000000008152600481019190915273ed477a99035d0c1e11369f1d7a4e587893cc002b9063d938ec7b90602401602060405180830381865af4158015611dce573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190611df29190612aee565b73ffffffffffffffffffffffffffffffffffffffff85811660009081526002602052604080822090517f7c6bb8620000000000000000000000000000000000000000000000000000000081526004810191909152918316602483015291925073ed477a99035d0c1e11369f1d7a4e587893cc002b90637c6bb86290604401602060405180830381865af4158015611e8d573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190611eb19190612a13565b9050611ebc85610a22565b611ec68242612a8a565b10611ed75750600194909350915050565b50600094909350915050565b60005473ffffffffffffffffffffffffffffffffffffffff163314610a91576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015260640161079d565b73ffffffffffffffffffffffffffffffffffffffff821660009081526001602052604080822090517f95073a7900000000000000000000000000000000000000000000000000000000815273ed477a99035d0c1e11369f1d7a4e587893cc002b916395073a7991611ff99190869060040191825273ffffffffffffffffffffffffffffffffffffffff16602082015260400190565b602060405180830381865af4158015612016573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061203a9190612acc565b80156120fe575073ffffffffffffffffffffffffffffffffffffffff8381166000908152600260205260409081902090517f95073a790000000000000000000000000000000000000000000000000000000081526004810191909152908316602482015273ed477a99035d0c1e11369f1d7a4e587893cc002b906395073a7990604401602060405180830381865af41580156120da573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906120fe9190612acc565b9392505050565b61210e826109ab565b600114801561212257506121228282611f64565b1561212b575050565b73ffffffffffffffffffffffffffffffffffffffff821660009081526001602052604080822090517f59d556a8000000000000000000000000000000000000000000000000000000008152600481019190915273ed477a99035d0c1e11369f1d7a4e587893cc002b906359d556a890602401602060405180830381865af41580156121ba573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906121de9190612a13565b73ffffffffffffffffffffffffffffffffffffffff84166000908152600160205260409081902090517fc1e728e900000000000000000000000000000000000000000000000000000000815291925073ed477a99035d0c1e11369f1d7a4e587893cc002b9163c1e728e99161227691869060040191825273ffffffffffffffffffffffffffffffffffffffff16602082015260400190565b60006040518083038186803b15801561228e57600080fd5b505af41580156122a2573d6000803e3d6000fd5b50505073ffffffffffffffffffffffffffffffffffffffff84166000908152600260205260409081902090517fc1e728e900000000000000000000000000000000000000000000000000000000815273ed477a99035d0c1e11369f1d7a4e587893cc002b925063c1e728e99161233b91869060040191825273ffffffffffffffffffffffffffffffffffffffff16602082015260400190565b60006040518083038186803b15801561235357600080fd5b505af4158015612367573d6000803e3d6000fd5b505060405173ffffffffffffffffffffffffffffffffffffffff8086169350861691507fe21a44017b6fa1658d84e937d56ff408501facdb4ff7427c479ac460d76f789390600090a373ffffffffffffffffffffffffffffffffffffffff831660009081526001602052604080822090517f59d556a8000000000000000000000000000000000000000000000000000000008152600481019190915273ed477a99035d0c1e11369f1d7a4e587893cc002b906359d556a890602401602060405180830381865af415801561243f573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906124639190612a13565b90508181146124bd578373ffffffffffffffffffffffffffffffffffffffff167fa9981ebfc3b766a742486e898f54959b050a66006dbce1a4155c1f84a08bcf41826040516124b491815260200190565b60405180910390a25b50505050565b6000805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b73ffffffffffffffffffffffffffffffffffffffff81168114611a3557600080fd5b60006020828403121561256c57600080fd5b81356120fe81612538565b600081518084526020808501945080840160005b838110156125bd57815173ffffffffffffffffffffffffffffffffffffffff168752958201959082019060010161258b565b509495945050505050565b6060815260006125db6060830186612577565b82810360208481019190915285518083528682019282019060005b81811015612612578451835293830193918301916001016125f6565b50508481036040860152855180825290820192508186019060005b81811015612681578251600480821061266e577f4e487b71000000000000000000000000000000000000000000000000000000006000526021815260246000fd5b508552938301939183019160010161262d565b509298975050505050505050565b600080604083850312156126a257600080fd5b82356126ad81612538565b915060208301356126bd81612538565b809150509250929050565b6000806000606084860312156126dd57600080fd5b83356126e881612538565b925060208401356126f881612538565b929592945050506040919091013590565b6000806000806080858703121561271f57600080fd5b843561272a81612538565b935060208501359250604085013561274181612538565b9150606085013561275181612538565b939692955090935050565b6020815260006120fe6020830184612577565b6000806040838503121561278257600080fd5b823561278d81612538565b946020939093013593505050565b6000602082840312156127ad57600080fd5b5035919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b604051601f82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016810167ffffffffffffffff8111828210171561282a5761282a6127b4565b604052919050565b600067ffffffffffffffff82111561284c5761284c6127b4565b5060051b60200190565b600082601f83011261286757600080fd5b8151602061287c61287783612832565b6127e3565b82815260059290921b8401810191818101908684111561289b57600080fd5b8286015b848110156128b6578051835291830191830161289f565b509695505050505050565b600082601f8301126128d257600080fd5b815160206128e261287783612832565b82815260059290921b8401810191818101908684111561290157600080fd5b8286015b848110156128b65780516004811061291d5760008081fd5b8352918301918301612905565b60008060006060848603121561293f57600080fd5b835167ffffffffffffffff8082111561295757600080fd5b818601915086601f83011261296b57600080fd5b8151602061297b61287783612832565b82815260059290921b8401810191818101908a84111561299a57600080fd5b948201945b838610156129c15785516129b281612538565b8252948201949082019061299f565b918901519197509093505050808211156129da57600080fd5b6129e687838801612856565b935060408601519150808211156129fc57600080fd5b50612a09868287016128c1565b9150509250925092565b600060208284031215612a2557600080fd5b5051919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b8181038181111561061057610610612a5b565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603160045260246000fd5b600060208284031215612ade57600080fd5b815180151581146120fe57600080fd5b600060208284031215612b0057600080fd5b81516120fe81612538565b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203612b3c57612b3c612a5b565b506001019056fea164736f6c6343000813000a",
}

// SortedOraclesABI is the input ABI used to generate the binding from.
// Deprecated: Use SortedOraclesMetaData.ABI instead.
var SortedOraclesABI = SortedOraclesMetaData.ABI

// SortedOraclesBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SortedOraclesMetaData.Bin instead.
var SortedOraclesBin = SortedOraclesMetaData.Bin

// DeploySortedOracles deploys a new Ethereum contract, binding an instance of SortedOracles to it.
func DeploySortedOracles(auth *bind.TransactOpts, backend bind.ContractBackend, test bool) (common.Address, *types.Transaction, *SortedOracles, error) {
	parsed, err := SortedOraclesMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SortedOraclesBin), backend, test)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SortedOracles{SortedOraclesCaller: SortedOraclesCaller{contract: contract}, SortedOraclesTransactor: SortedOraclesTransactor{contract: contract}, SortedOraclesFilterer: SortedOraclesFilterer{contract: contract}}, nil
}

// SortedOracles is an auto generated Go binding around an Ethereum contract.
type SortedOracles struct {
	SortedOraclesCaller     // Read-only binding to the contract
	SortedOraclesTransactor // Write-only binding to the contract
	SortedOraclesFilterer   // Log filterer for contract events
}

// SortedOraclesCaller is an auto generated read-only Go binding around an Ethereum contract.
type SortedOraclesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SortedOraclesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SortedOraclesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SortedOraclesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SortedOraclesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SortedOraclesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SortedOraclesSession struct {
	Contract     *SortedOracles    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SortedOraclesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SortedOraclesCallerSession struct {
	Contract *SortedOraclesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// SortedOraclesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SortedOraclesTransactorSession struct {
	Contract     *SortedOraclesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// SortedOraclesRaw is an auto generated low-level Go binding around an Ethereum contract.
type SortedOraclesRaw struct {
	Contract *SortedOracles // Generic contract binding to access the raw methods on
}

// SortedOraclesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SortedOraclesCallerRaw struct {
	Contract *SortedOraclesCaller // Generic read-only contract binding to access the raw methods on
}

// SortedOraclesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SortedOraclesTransactorRaw struct {
	Contract *SortedOraclesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSortedOracles creates a new instance of SortedOracles, bound to a specific deployed contract.
func NewSortedOracles(address common.Address, backend bind.ContractBackend) (*SortedOracles, error) {
	contract, err := bindSortedOracles(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SortedOracles{SortedOraclesCaller: SortedOraclesCaller{contract: contract}, SortedOraclesTransactor: SortedOraclesTransactor{contract: contract}, SortedOraclesFilterer: SortedOraclesFilterer{contract: contract}}, nil
}

// NewSortedOraclesCaller creates a new read-only instance of SortedOracles, bound to a specific deployed contract.
func NewSortedOraclesCaller(address common.Address, caller bind.ContractCaller) (*SortedOraclesCaller, error) {
	contract, err := bindSortedOracles(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SortedOraclesCaller{contract: contract}, nil
}

// NewSortedOraclesTransactor creates a new write-only instance of SortedOracles, bound to a specific deployed contract.
func NewSortedOraclesTransactor(address common.Address, transactor bind.ContractTransactor) (*SortedOraclesTransactor, error) {
	contract, err := bindSortedOracles(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SortedOraclesTransactor{contract: contract}, nil
}

// NewSortedOraclesFilterer creates a new log filterer instance of SortedOracles, bound to a specific deployed contract.
func NewSortedOraclesFilterer(address common.Address, filterer bind.ContractFilterer) (*SortedOraclesFilterer, error) {
	contract, err := bindSortedOracles(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SortedOraclesFilterer{contract: contract}, nil
}

// bindSortedOracles binds a generic wrapper to an already deployed contract.
func bindSortedOracles(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SortedOraclesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SortedOracles *SortedOraclesRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SortedOracles.Contract.SortedOraclesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SortedOracles *SortedOraclesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SortedOracles.Contract.SortedOraclesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SortedOracles *SortedOraclesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SortedOracles.Contract.SortedOraclesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SortedOracles *SortedOraclesCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SortedOracles.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SortedOracles *SortedOraclesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SortedOracles.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SortedOracles *SortedOraclesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SortedOracles.Contract.contract.Transact(opts, method, params...)
}

// GetOracles is a free data retrieval call binding the contract method 0x8e749281.
//
// Solidity: function getOracles(address token) view returns(address[])
func (_SortedOracles *SortedOraclesCaller) GetOracles(opts *bind.CallOpts, token common.Address) ([]common.Address, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "getOracles", token)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetOracles is a free data retrieval call binding the contract method 0x8e749281.
//
// Solidity: function getOracles(address token) view returns(address[])
func (_SortedOracles *SortedOraclesSession) GetOracles(token common.Address) ([]common.Address, error) {
	return _SortedOracles.Contract.GetOracles(&_SortedOracles.CallOpts, token)
}

// GetOracles is a free data retrieval call binding the contract method 0x8e749281.
//
// Solidity: function getOracles(address token) view returns(address[])
func (_SortedOracles *SortedOraclesCallerSession) GetOracles(token common.Address) ([]common.Address, error) {
	return _SortedOracles.Contract.GetOracles(&_SortedOracles.CallOpts, token)
}

// GetRates is a free data retrieval call binding the contract method 0x02f55b61.
//
// Solidity: function getRates(address token) view returns(address[], uint256[], uint8[])
func (_SortedOracles *SortedOraclesCaller) GetRates(opts *bind.CallOpts, token common.Address) ([]common.Address, []*big.Int, []uint8, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "getRates", token)

	if err != nil {
		return *new([]common.Address), *new([]*big.Int), *new([]uint8), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	out1 := *abi.ConvertType(out[1], new([]*big.Int)).(*[]*big.Int)
	out2 := *abi.ConvertType(out[2], new([]uint8)).(*[]uint8)

	return out0, out1, out2, err

}

// GetRates is a free data retrieval call binding the contract method 0x02f55b61.
//
// Solidity: function getRates(address token) view returns(address[], uint256[], uint8[])
func (_SortedOracles *SortedOraclesSession) GetRates(token common.Address) ([]common.Address, []*big.Int, []uint8, error) {
	return _SortedOracles.Contract.GetRates(&_SortedOracles.CallOpts, token)
}

// GetRates is a free data retrieval call binding the contract method 0x02f55b61.
//
// Solidity: function getRates(address token) view returns(address[], uint256[], uint8[])
func (_SortedOracles *SortedOraclesCallerSession) GetRates(token common.Address) ([]common.Address, []*big.Int, []uint8, error) {
	return _SortedOracles.Contract.GetRates(&_SortedOracles.CallOpts, token)
}

// GetTimestamps is a free data retrieval call binding the contract method 0xb9292158.
//
// Solidity: function getTimestamps(address token) view returns(address[], uint256[], uint8[])
func (_SortedOracles *SortedOraclesCaller) GetTimestamps(opts *bind.CallOpts, token common.Address) ([]common.Address, []*big.Int, []uint8, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "getTimestamps", token)

	if err != nil {
		return *new([]common.Address), *new([]*big.Int), *new([]uint8), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	out1 := *abi.ConvertType(out[1], new([]*big.Int)).(*[]*big.Int)
	out2 := *abi.ConvertType(out[2], new([]uint8)).(*[]uint8)

	return out0, out1, out2, err

}

// GetTimestamps is a free data retrieval call binding the contract method 0xb9292158.
//
// Solidity: function getTimestamps(address token) view returns(address[], uint256[], uint8[])
func (_SortedOracles *SortedOraclesSession) GetTimestamps(token common.Address) ([]common.Address, []*big.Int, []uint8, error) {
	return _SortedOracles.Contract.GetTimestamps(&_SortedOracles.CallOpts, token)
}

// GetTimestamps is a free data retrieval call binding the contract method 0xb9292158.
//
// Solidity: function getTimestamps(address token) view returns(address[], uint256[], uint8[])
func (_SortedOracles *SortedOraclesCallerSession) GetTimestamps(token common.Address) ([]common.Address, []*big.Int, []uint8, error) {
	return _SortedOracles.Contract.GetTimestamps(&_SortedOracles.CallOpts, token)
}

// GetTokenReportExpirySeconds is a free data retrieval call binding the contract method 0x6deb6799.
//
// Solidity: function getTokenReportExpirySeconds(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesCaller) GetTokenReportExpirySeconds(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "getTokenReportExpirySeconds", token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTokenReportExpirySeconds is a free data retrieval call binding the contract method 0x6deb6799.
//
// Solidity: function getTokenReportExpirySeconds(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesSession) GetTokenReportExpirySeconds(token common.Address) (*big.Int, error) {
	return _SortedOracles.Contract.GetTokenReportExpirySeconds(&_SortedOracles.CallOpts, token)
}

// GetTokenReportExpirySeconds is a free data retrieval call binding the contract method 0x6deb6799.
//
// Solidity: function getTokenReportExpirySeconds(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesCallerSession) GetTokenReportExpirySeconds(token common.Address) (*big.Int, error) {
	return _SortedOracles.Contract.GetTokenReportExpirySeconds(&_SortedOracles.CallOpts, token)
}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_SortedOracles *SortedOraclesCaller) GetVersionNumber(opts *bind.CallOpts) (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "getVersionNumber")

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
func (_SortedOracles *SortedOraclesSession) GetVersionNumber() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _SortedOracles.Contract.GetVersionNumber(&_SortedOracles.CallOpts)
}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_SortedOracles *SortedOraclesCallerSession) GetVersionNumber() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _SortedOracles.Contract.GetVersionNumber(&_SortedOracles.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_SortedOracles *SortedOraclesCaller) Initialized(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "initialized")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_SortedOracles *SortedOraclesSession) Initialized() (bool, error) {
	return _SortedOracles.Contract.Initialized(&_SortedOracles.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_SortedOracles *SortedOraclesCallerSession) Initialized() (bool, error) {
	return _SortedOracles.Contract.Initialized(&_SortedOracles.CallOpts)
}

// IsOldestReportExpired is a free data retrieval call binding the contract method 0xffe736bf.
//
// Solidity: function isOldestReportExpired(address token) view returns(bool, address)
func (_SortedOracles *SortedOraclesCaller) IsOldestReportExpired(opts *bind.CallOpts, token common.Address) (bool, common.Address, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "isOldestReportExpired", token)

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
func (_SortedOracles *SortedOraclesSession) IsOldestReportExpired(token common.Address) (bool, common.Address, error) {
	return _SortedOracles.Contract.IsOldestReportExpired(&_SortedOracles.CallOpts, token)
}

// IsOldestReportExpired is a free data retrieval call binding the contract method 0xffe736bf.
//
// Solidity: function isOldestReportExpired(address token) view returns(bool, address)
func (_SortedOracles *SortedOraclesCallerSession) IsOldestReportExpired(token common.Address) (bool, common.Address, error) {
	return _SortedOracles.Contract.IsOldestReportExpired(&_SortedOracles.CallOpts, token)
}

// IsOracle is a free data retrieval call binding the contract method 0x370c998e.
//
// Solidity: function isOracle(address , address ) view returns(bool)
func (_SortedOracles *SortedOraclesCaller) IsOracle(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (bool, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "isOracle", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOracle is a free data retrieval call binding the contract method 0x370c998e.
//
// Solidity: function isOracle(address , address ) view returns(bool)
func (_SortedOracles *SortedOraclesSession) IsOracle(arg0 common.Address, arg1 common.Address) (bool, error) {
	return _SortedOracles.Contract.IsOracle(&_SortedOracles.CallOpts, arg0, arg1)
}

// IsOracle is a free data retrieval call binding the contract method 0x370c998e.
//
// Solidity: function isOracle(address , address ) view returns(bool)
func (_SortedOracles *SortedOraclesCallerSession) IsOracle(arg0 common.Address, arg1 common.Address) (bool, error) {
	return _SortedOracles.Contract.IsOracle(&_SortedOracles.CallOpts, arg0, arg1)
}

// MedianRate is a free data retrieval call binding the contract method 0xef90e1b0.
//
// Solidity: function medianRate(address token) view returns(uint256, uint256)
func (_SortedOracles *SortedOraclesCaller) MedianRate(opts *bind.CallOpts, token common.Address) (*big.Int, *big.Int, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "medianRate", token)

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
func (_SortedOracles *SortedOraclesSession) MedianRate(token common.Address) (*big.Int, *big.Int, error) {
	return _SortedOracles.Contract.MedianRate(&_SortedOracles.CallOpts, token)
}

// MedianRate is a free data retrieval call binding the contract method 0xef90e1b0.
//
// Solidity: function medianRate(address token) view returns(uint256, uint256)
func (_SortedOracles *SortedOraclesCallerSession) MedianRate(token common.Address) (*big.Int, *big.Int, error) {
	return _SortedOracles.Contract.MedianRate(&_SortedOracles.CallOpts, token)
}

// MedianTimestamp is a free data retrieval call binding the contract method 0x071b48fc.
//
// Solidity: function medianTimestamp(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesCaller) MedianTimestamp(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "medianTimestamp", token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MedianTimestamp is a free data retrieval call binding the contract method 0x071b48fc.
//
// Solidity: function medianTimestamp(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesSession) MedianTimestamp(token common.Address) (*big.Int, error) {
	return _SortedOracles.Contract.MedianTimestamp(&_SortedOracles.CallOpts, token)
}

// MedianTimestamp is a free data retrieval call binding the contract method 0x071b48fc.
//
// Solidity: function medianTimestamp(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesCallerSession) MedianTimestamp(token common.Address) (*big.Int, error) {
	return _SortedOracles.Contract.MedianTimestamp(&_SortedOracles.CallOpts, token)
}

// NumRates is a free data retrieval call binding the contract method 0xbbc66a94.
//
// Solidity: function numRates(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesCaller) NumRates(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "numRates", token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumRates is a free data retrieval call binding the contract method 0xbbc66a94.
//
// Solidity: function numRates(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesSession) NumRates(token common.Address) (*big.Int, error) {
	return _SortedOracles.Contract.NumRates(&_SortedOracles.CallOpts, token)
}

// NumRates is a free data retrieval call binding the contract method 0xbbc66a94.
//
// Solidity: function numRates(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesCallerSession) NumRates(token common.Address) (*big.Int, error) {
	return _SortedOracles.Contract.NumRates(&_SortedOracles.CallOpts, token)
}

// NumTimestamps is a free data retrieval call binding the contract method 0x6dd6ef0c.
//
// Solidity: function numTimestamps(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesCaller) NumTimestamps(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "numTimestamps", token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumTimestamps is a free data retrieval call binding the contract method 0x6dd6ef0c.
//
// Solidity: function numTimestamps(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesSession) NumTimestamps(token common.Address) (*big.Int, error) {
	return _SortedOracles.Contract.NumTimestamps(&_SortedOracles.CallOpts, token)
}

// NumTimestamps is a free data retrieval call binding the contract method 0x6dd6ef0c.
//
// Solidity: function numTimestamps(address token) view returns(uint256)
func (_SortedOracles *SortedOraclesCallerSession) NumTimestamps(token common.Address) (*big.Int, error) {
	return _SortedOracles.Contract.NumTimestamps(&_SortedOracles.CallOpts, token)
}

// Oracles is a free data retrieval call binding the contract method 0xa00a8b2c.
//
// Solidity: function oracles(address , uint256 ) view returns(address)
func (_SortedOracles *SortedOraclesCaller) Oracles(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "oracles", arg0, arg1)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Oracles is a free data retrieval call binding the contract method 0xa00a8b2c.
//
// Solidity: function oracles(address , uint256 ) view returns(address)
func (_SortedOracles *SortedOraclesSession) Oracles(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _SortedOracles.Contract.Oracles(&_SortedOracles.CallOpts, arg0, arg1)
}

// Oracles is a free data retrieval call binding the contract method 0xa00a8b2c.
//
// Solidity: function oracles(address , uint256 ) view returns(address)
func (_SortedOracles *SortedOraclesCallerSession) Oracles(arg0 common.Address, arg1 *big.Int) (common.Address, error) {
	return _SortedOracles.Contract.Oracles(&_SortedOracles.CallOpts, arg0, arg1)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SortedOracles *SortedOraclesCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SortedOracles *SortedOraclesSession) Owner() (common.Address, error) {
	return _SortedOracles.Contract.Owner(&_SortedOracles.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SortedOracles *SortedOraclesCallerSession) Owner() (common.Address, error) {
	return _SortedOracles.Contract.Owner(&_SortedOracles.CallOpts)
}

// ReportExpirySeconds is a free data retrieval call binding the contract method 0x493a353c.
//
// Solidity: function reportExpirySeconds() view returns(uint256)
func (_SortedOracles *SortedOraclesCaller) ReportExpirySeconds(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "reportExpirySeconds")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ReportExpirySeconds is a free data retrieval call binding the contract method 0x493a353c.
//
// Solidity: function reportExpirySeconds() view returns(uint256)
func (_SortedOracles *SortedOraclesSession) ReportExpirySeconds() (*big.Int, error) {
	return _SortedOracles.Contract.ReportExpirySeconds(&_SortedOracles.CallOpts)
}

// ReportExpirySeconds is a free data retrieval call binding the contract method 0x493a353c.
//
// Solidity: function reportExpirySeconds() view returns(uint256)
func (_SortedOracles *SortedOraclesCallerSession) ReportExpirySeconds() (*big.Int, error) {
	return _SortedOracles.Contract.ReportExpirySeconds(&_SortedOracles.CallOpts)
}

// TokenReportExpirySeconds is a free data retrieval call binding the contract method 0x2e86bc01.
//
// Solidity: function tokenReportExpirySeconds(address ) view returns(uint256)
func (_SortedOracles *SortedOraclesCaller) TokenReportExpirySeconds(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SortedOracles.contract.Call(opts, &out, "tokenReportExpirySeconds", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenReportExpirySeconds is a free data retrieval call binding the contract method 0x2e86bc01.
//
// Solidity: function tokenReportExpirySeconds(address ) view returns(uint256)
func (_SortedOracles *SortedOraclesSession) TokenReportExpirySeconds(arg0 common.Address) (*big.Int, error) {
	return _SortedOracles.Contract.TokenReportExpirySeconds(&_SortedOracles.CallOpts, arg0)
}

// TokenReportExpirySeconds is a free data retrieval call binding the contract method 0x2e86bc01.
//
// Solidity: function tokenReportExpirySeconds(address ) view returns(uint256)
func (_SortedOracles *SortedOraclesCallerSession) TokenReportExpirySeconds(arg0 common.Address) (*big.Int, error) {
	return _SortedOracles.Contract.TokenReportExpirySeconds(&_SortedOracles.CallOpts, arg0)
}

// AddOracle is a paid mutator transaction binding the contract method 0xf0ca4adb.
//
// Solidity: function addOracle(address token, address oracleAddress) returns()
func (_SortedOracles *SortedOraclesTransactor) AddOracle(opts *bind.TransactOpts, token common.Address, oracleAddress common.Address) (*types.Transaction, error) {
	return _SortedOracles.contract.Transact(opts, "addOracle", token, oracleAddress)
}

// AddOracle is a paid mutator transaction binding the contract method 0xf0ca4adb.
//
// Solidity: function addOracle(address token, address oracleAddress) returns()
func (_SortedOracles *SortedOraclesSession) AddOracle(token common.Address, oracleAddress common.Address) (*types.Transaction, error) {
	return _SortedOracles.Contract.AddOracle(&_SortedOracles.TransactOpts, token, oracleAddress)
}

// AddOracle is a paid mutator transaction binding the contract method 0xf0ca4adb.
//
// Solidity: function addOracle(address token, address oracleAddress) returns()
func (_SortedOracles *SortedOraclesTransactorSession) AddOracle(token common.Address, oracleAddress common.Address) (*types.Transaction, error) {
	return _SortedOracles.Contract.AddOracle(&_SortedOracles.TransactOpts, token, oracleAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0xfe4b84df.
//
// Solidity: function initialize(uint256 _reportExpirySeconds) returns()
func (_SortedOracles *SortedOraclesTransactor) Initialize(opts *bind.TransactOpts, _reportExpirySeconds *big.Int) (*types.Transaction, error) {
	return _SortedOracles.contract.Transact(opts, "initialize", _reportExpirySeconds)
}

// Initialize is a paid mutator transaction binding the contract method 0xfe4b84df.
//
// Solidity: function initialize(uint256 _reportExpirySeconds) returns()
func (_SortedOracles *SortedOraclesSession) Initialize(_reportExpirySeconds *big.Int) (*types.Transaction, error) {
	return _SortedOracles.Contract.Initialize(&_SortedOracles.TransactOpts, _reportExpirySeconds)
}

// Initialize is a paid mutator transaction binding the contract method 0xfe4b84df.
//
// Solidity: function initialize(uint256 _reportExpirySeconds) returns()
func (_SortedOracles *SortedOraclesTransactorSession) Initialize(_reportExpirySeconds *big.Int) (*types.Transaction, error) {
	return _SortedOracles.Contract.Initialize(&_SortedOracles.TransactOpts, _reportExpirySeconds)
}

// RemoveExpiredReports is a paid mutator transaction binding the contract method 0xdd34ca3b.
//
// Solidity: function removeExpiredReports(address token, uint256 n) returns()
func (_SortedOracles *SortedOraclesTransactor) RemoveExpiredReports(opts *bind.TransactOpts, token common.Address, n *big.Int) (*types.Transaction, error) {
	return _SortedOracles.contract.Transact(opts, "removeExpiredReports", token, n)
}

// RemoveExpiredReports is a paid mutator transaction binding the contract method 0xdd34ca3b.
//
// Solidity: function removeExpiredReports(address token, uint256 n) returns()
func (_SortedOracles *SortedOraclesSession) RemoveExpiredReports(token common.Address, n *big.Int) (*types.Transaction, error) {
	return _SortedOracles.Contract.RemoveExpiredReports(&_SortedOracles.TransactOpts, token, n)
}

// RemoveExpiredReports is a paid mutator transaction binding the contract method 0xdd34ca3b.
//
// Solidity: function removeExpiredReports(address token, uint256 n) returns()
func (_SortedOracles *SortedOraclesTransactorSession) RemoveExpiredReports(token common.Address, n *big.Int) (*types.Transaction, error) {
	return _SortedOracles.Contract.RemoveExpiredReports(&_SortedOracles.TransactOpts, token, n)
}

// RemoveOracle is a paid mutator transaction binding the contract method 0x53a57297.
//
// Solidity: function removeOracle(address token, address oracleAddress, uint256 index) returns()
func (_SortedOracles *SortedOraclesTransactor) RemoveOracle(opts *bind.TransactOpts, token common.Address, oracleAddress common.Address, index *big.Int) (*types.Transaction, error) {
	return _SortedOracles.contract.Transact(opts, "removeOracle", token, oracleAddress, index)
}

// RemoveOracle is a paid mutator transaction binding the contract method 0x53a57297.
//
// Solidity: function removeOracle(address token, address oracleAddress, uint256 index) returns()
func (_SortedOracles *SortedOraclesSession) RemoveOracle(token common.Address, oracleAddress common.Address, index *big.Int) (*types.Transaction, error) {
	return _SortedOracles.Contract.RemoveOracle(&_SortedOracles.TransactOpts, token, oracleAddress, index)
}

// RemoveOracle is a paid mutator transaction binding the contract method 0x53a57297.
//
// Solidity: function removeOracle(address token, address oracleAddress, uint256 index) returns()
func (_SortedOracles *SortedOraclesTransactorSession) RemoveOracle(token common.Address, oracleAddress common.Address, index *big.Int) (*types.Transaction, error) {
	return _SortedOracles.Contract.RemoveOracle(&_SortedOracles.TransactOpts, token, oracleAddress, index)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SortedOracles *SortedOraclesTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SortedOracles.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SortedOracles *SortedOraclesSession) RenounceOwnership() (*types.Transaction, error) {
	return _SortedOracles.Contract.RenounceOwnership(&_SortedOracles.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SortedOracles *SortedOraclesTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SortedOracles.Contract.RenounceOwnership(&_SortedOracles.TransactOpts)
}

// Report is a paid mutator transaction binding the contract method 0x80e50744.
//
// Solidity: function report(address token, uint256 value, address lesserKey, address greaterKey) returns()
func (_SortedOracles *SortedOraclesTransactor) Report(opts *bind.TransactOpts, token common.Address, value *big.Int, lesserKey common.Address, greaterKey common.Address) (*types.Transaction, error) {
	return _SortedOracles.contract.Transact(opts, "report", token, value, lesserKey, greaterKey)
}

// Report is a paid mutator transaction binding the contract method 0x80e50744.
//
// Solidity: function report(address token, uint256 value, address lesserKey, address greaterKey) returns()
func (_SortedOracles *SortedOraclesSession) Report(token common.Address, value *big.Int, lesserKey common.Address, greaterKey common.Address) (*types.Transaction, error) {
	return _SortedOracles.Contract.Report(&_SortedOracles.TransactOpts, token, value, lesserKey, greaterKey)
}

// Report is a paid mutator transaction binding the contract method 0x80e50744.
//
// Solidity: function report(address token, uint256 value, address lesserKey, address greaterKey) returns()
func (_SortedOracles *SortedOraclesTransactorSession) Report(token common.Address, value *big.Int, lesserKey common.Address, greaterKey common.Address) (*types.Transaction, error) {
	return _SortedOracles.Contract.Report(&_SortedOracles.TransactOpts, token, value, lesserKey, greaterKey)
}

// SetReportExpiry is a paid mutator transaction binding the contract method 0xebc1d6bb.
//
// Solidity: function setReportExpiry(uint256 _reportExpirySeconds) returns()
func (_SortedOracles *SortedOraclesTransactor) SetReportExpiry(opts *bind.TransactOpts, _reportExpirySeconds *big.Int) (*types.Transaction, error) {
	return _SortedOracles.contract.Transact(opts, "setReportExpiry", _reportExpirySeconds)
}

// SetReportExpiry is a paid mutator transaction binding the contract method 0xebc1d6bb.
//
// Solidity: function setReportExpiry(uint256 _reportExpirySeconds) returns()
func (_SortedOracles *SortedOraclesSession) SetReportExpiry(_reportExpirySeconds *big.Int) (*types.Transaction, error) {
	return _SortedOracles.Contract.SetReportExpiry(&_SortedOracles.TransactOpts, _reportExpirySeconds)
}

// SetReportExpiry is a paid mutator transaction binding the contract method 0xebc1d6bb.
//
// Solidity: function setReportExpiry(uint256 _reportExpirySeconds) returns()
func (_SortedOracles *SortedOraclesTransactorSession) SetReportExpiry(_reportExpirySeconds *big.Int) (*types.Transaction, error) {
	return _SortedOracles.Contract.SetReportExpiry(&_SortedOracles.TransactOpts, _reportExpirySeconds)
}

// SetTokenReportExpiry is a paid mutator transaction binding the contract method 0xfc20935d.
//
// Solidity: function setTokenReportExpiry(address _token, uint256 _reportExpirySeconds) returns()
func (_SortedOracles *SortedOraclesTransactor) SetTokenReportExpiry(opts *bind.TransactOpts, _token common.Address, _reportExpirySeconds *big.Int) (*types.Transaction, error) {
	return _SortedOracles.contract.Transact(opts, "setTokenReportExpiry", _token, _reportExpirySeconds)
}

// SetTokenReportExpiry is a paid mutator transaction binding the contract method 0xfc20935d.
//
// Solidity: function setTokenReportExpiry(address _token, uint256 _reportExpirySeconds) returns()
func (_SortedOracles *SortedOraclesSession) SetTokenReportExpiry(_token common.Address, _reportExpirySeconds *big.Int) (*types.Transaction, error) {
	return _SortedOracles.Contract.SetTokenReportExpiry(&_SortedOracles.TransactOpts, _token, _reportExpirySeconds)
}

// SetTokenReportExpiry is a paid mutator transaction binding the contract method 0xfc20935d.
//
// Solidity: function setTokenReportExpiry(address _token, uint256 _reportExpirySeconds) returns()
func (_SortedOracles *SortedOraclesTransactorSession) SetTokenReportExpiry(_token common.Address, _reportExpirySeconds *big.Int) (*types.Transaction, error) {
	return _SortedOracles.Contract.SetTokenReportExpiry(&_SortedOracles.TransactOpts, _token, _reportExpirySeconds)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SortedOracles *SortedOraclesTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SortedOracles.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SortedOracles *SortedOraclesSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SortedOracles.Contract.TransferOwnership(&_SortedOracles.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SortedOracles *SortedOraclesTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SortedOracles.Contract.TransferOwnership(&_SortedOracles.TransactOpts, newOwner)
}

// SortedOraclesMedianUpdatedIterator is returned from FilterMedianUpdated and is used to iterate over the raw logs and unpacked data for MedianUpdated events raised by the SortedOracles contract.
type SortedOraclesMedianUpdatedIterator struct {
	Event *SortedOraclesMedianUpdated // Event containing the contract specifics and raw log

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
func (it *SortedOraclesMedianUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SortedOraclesMedianUpdated)
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
		it.Event = new(SortedOraclesMedianUpdated)
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
func (it *SortedOraclesMedianUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SortedOraclesMedianUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SortedOraclesMedianUpdated represents a MedianUpdated event raised by the SortedOracles contract.
type SortedOraclesMedianUpdated struct {
	Token common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterMedianUpdated is a free log retrieval operation binding the contract event 0xa9981ebfc3b766a742486e898f54959b050a66006dbce1a4155c1f84a08bcf41.
//
// Solidity: event MedianUpdated(address indexed token, uint256 value)
func (_SortedOracles *SortedOraclesFilterer) FilterMedianUpdated(opts *bind.FilterOpts, token []common.Address) (*SortedOraclesMedianUpdatedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _SortedOracles.contract.FilterLogs(opts, "MedianUpdated", tokenRule)
	if err != nil {
		return nil, err
	}
	return &SortedOraclesMedianUpdatedIterator{contract: _SortedOracles.contract, event: "MedianUpdated", logs: logs, sub: sub}, nil
}

// WatchMedianUpdated is a free log subscription operation binding the contract event 0xa9981ebfc3b766a742486e898f54959b050a66006dbce1a4155c1f84a08bcf41.
//
// Solidity: event MedianUpdated(address indexed token, uint256 value)
func (_SortedOracles *SortedOraclesFilterer) WatchMedianUpdated(opts *bind.WatchOpts, sink chan<- *SortedOraclesMedianUpdated, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _SortedOracles.contract.WatchLogs(opts, "MedianUpdated", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SortedOraclesMedianUpdated)
				if err := _SortedOracles.contract.UnpackLog(event, "MedianUpdated", log); err != nil {
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

// ParseMedianUpdated is a log parse operation binding the contract event 0xa9981ebfc3b766a742486e898f54959b050a66006dbce1a4155c1f84a08bcf41.
//
// Solidity: event MedianUpdated(address indexed token, uint256 value)
func (_SortedOracles *SortedOraclesFilterer) ParseMedianUpdated(log types.Log) (*SortedOraclesMedianUpdated, error) {
	event := new(SortedOraclesMedianUpdated)
	if err := _SortedOracles.contract.UnpackLog(event, "MedianUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SortedOraclesOracleAddedIterator is returned from FilterOracleAdded and is used to iterate over the raw logs and unpacked data for OracleAdded events raised by the SortedOracles contract.
type SortedOraclesOracleAddedIterator struct {
	Event *SortedOraclesOracleAdded // Event containing the contract specifics and raw log

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
func (it *SortedOraclesOracleAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SortedOraclesOracleAdded)
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
		it.Event = new(SortedOraclesOracleAdded)
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
func (it *SortedOraclesOracleAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SortedOraclesOracleAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SortedOraclesOracleAdded represents a OracleAdded event raised by the SortedOracles contract.
type SortedOraclesOracleAdded struct {
	Token         common.Address
	OracleAddress common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOracleAdded is a free log retrieval operation binding the contract event 0x828d2be040dede7698182e08dfa8bfbd663c879aee772509c4a2bd961d0ed43f.
//
// Solidity: event OracleAdded(address indexed token, address indexed oracleAddress)
func (_SortedOracles *SortedOraclesFilterer) FilterOracleAdded(opts *bind.FilterOpts, token []common.Address, oracleAddress []common.Address) (*SortedOraclesOracleAddedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var oracleAddressRule []interface{}
	for _, oracleAddressItem := range oracleAddress {
		oracleAddressRule = append(oracleAddressRule, oracleAddressItem)
	}

	logs, sub, err := _SortedOracles.contract.FilterLogs(opts, "OracleAdded", tokenRule, oracleAddressRule)
	if err != nil {
		return nil, err
	}
	return &SortedOraclesOracleAddedIterator{contract: _SortedOracles.contract, event: "OracleAdded", logs: logs, sub: sub}, nil
}

// WatchOracleAdded is a free log subscription operation binding the contract event 0x828d2be040dede7698182e08dfa8bfbd663c879aee772509c4a2bd961d0ed43f.
//
// Solidity: event OracleAdded(address indexed token, address indexed oracleAddress)
func (_SortedOracles *SortedOraclesFilterer) WatchOracleAdded(opts *bind.WatchOpts, sink chan<- *SortedOraclesOracleAdded, token []common.Address, oracleAddress []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var oracleAddressRule []interface{}
	for _, oracleAddressItem := range oracleAddress {
		oracleAddressRule = append(oracleAddressRule, oracleAddressItem)
	}

	logs, sub, err := _SortedOracles.contract.WatchLogs(opts, "OracleAdded", tokenRule, oracleAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SortedOraclesOracleAdded)
				if err := _SortedOracles.contract.UnpackLog(event, "OracleAdded", log); err != nil {
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

// ParseOracleAdded is a log parse operation binding the contract event 0x828d2be040dede7698182e08dfa8bfbd663c879aee772509c4a2bd961d0ed43f.
//
// Solidity: event OracleAdded(address indexed token, address indexed oracleAddress)
func (_SortedOracles *SortedOraclesFilterer) ParseOracleAdded(log types.Log) (*SortedOraclesOracleAdded, error) {
	event := new(SortedOraclesOracleAdded)
	if err := _SortedOracles.contract.UnpackLog(event, "OracleAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SortedOraclesOracleRemovedIterator is returned from FilterOracleRemoved and is used to iterate over the raw logs and unpacked data for OracleRemoved events raised by the SortedOracles contract.
type SortedOraclesOracleRemovedIterator struct {
	Event *SortedOraclesOracleRemoved // Event containing the contract specifics and raw log

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
func (it *SortedOraclesOracleRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SortedOraclesOracleRemoved)
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
		it.Event = new(SortedOraclesOracleRemoved)
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
func (it *SortedOraclesOracleRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SortedOraclesOracleRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SortedOraclesOracleRemoved represents a OracleRemoved event raised by the SortedOracles contract.
type SortedOraclesOracleRemoved struct {
	Token         common.Address
	OracleAddress common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOracleRemoved is a free log retrieval operation binding the contract event 0x6dc84b66cc948d847632b9d829f7cb1cb904fbf2c084554a9bc22ad9d8453340.
//
// Solidity: event OracleRemoved(address indexed token, address indexed oracleAddress)
func (_SortedOracles *SortedOraclesFilterer) FilterOracleRemoved(opts *bind.FilterOpts, token []common.Address, oracleAddress []common.Address) (*SortedOraclesOracleRemovedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var oracleAddressRule []interface{}
	for _, oracleAddressItem := range oracleAddress {
		oracleAddressRule = append(oracleAddressRule, oracleAddressItem)
	}

	logs, sub, err := _SortedOracles.contract.FilterLogs(opts, "OracleRemoved", tokenRule, oracleAddressRule)
	if err != nil {
		return nil, err
	}
	return &SortedOraclesOracleRemovedIterator{contract: _SortedOracles.contract, event: "OracleRemoved", logs: logs, sub: sub}, nil
}

// WatchOracleRemoved is a free log subscription operation binding the contract event 0x6dc84b66cc948d847632b9d829f7cb1cb904fbf2c084554a9bc22ad9d8453340.
//
// Solidity: event OracleRemoved(address indexed token, address indexed oracleAddress)
func (_SortedOracles *SortedOraclesFilterer) WatchOracleRemoved(opts *bind.WatchOpts, sink chan<- *SortedOraclesOracleRemoved, token []common.Address, oracleAddress []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var oracleAddressRule []interface{}
	for _, oracleAddressItem := range oracleAddress {
		oracleAddressRule = append(oracleAddressRule, oracleAddressItem)
	}

	logs, sub, err := _SortedOracles.contract.WatchLogs(opts, "OracleRemoved", tokenRule, oracleAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SortedOraclesOracleRemoved)
				if err := _SortedOracles.contract.UnpackLog(event, "OracleRemoved", log); err != nil {
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

// ParseOracleRemoved is a log parse operation binding the contract event 0x6dc84b66cc948d847632b9d829f7cb1cb904fbf2c084554a9bc22ad9d8453340.
//
// Solidity: event OracleRemoved(address indexed token, address indexed oracleAddress)
func (_SortedOracles *SortedOraclesFilterer) ParseOracleRemoved(log types.Log) (*SortedOraclesOracleRemoved, error) {
	event := new(SortedOraclesOracleRemoved)
	if err := _SortedOracles.contract.UnpackLog(event, "OracleRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SortedOraclesOracleReportRemovedIterator is returned from FilterOracleReportRemoved and is used to iterate over the raw logs and unpacked data for OracleReportRemoved events raised by the SortedOracles contract.
type SortedOraclesOracleReportRemovedIterator struct {
	Event *SortedOraclesOracleReportRemoved // Event containing the contract specifics and raw log

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
func (it *SortedOraclesOracleReportRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SortedOraclesOracleReportRemoved)
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
		it.Event = new(SortedOraclesOracleReportRemoved)
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
func (it *SortedOraclesOracleReportRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SortedOraclesOracleReportRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SortedOraclesOracleReportRemoved represents a OracleReportRemoved event raised by the SortedOracles contract.
type SortedOraclesOracleReportRemoved struct {
	Token  common.Address
	Oracle common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterOracleReportRemoved is a free log retrieval operation binding the contract event 0xe21a44017b6fa1658d84e937d56ff408501facdb4ff7427c479ac460d76f7893.
//
// Solidity: event OracleReportRemoved(address indexed token, address indexed oracle)
func (_SortedOracles *SortedOraclesFilterer) FilterOracleReportRemoved(opts *bind.FilterOpts, token []common.Address, oracle []common.Address) (*SortedOraclesOracleReportRemovedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var oracleRule []interface{}
	for _, oracleItem := range oracle {
		oracleRule = append(oracleRule, oracleItem)
	}

	logs, sub, err := _SortedOracles.contract.FilterLogs(opts, "OracleReportRemoved", tokenRule, oracleRule)
	if err != nil {
		return nil, err
	}
	return &SortedOraclesOracleReportRemovedIterator{contract: _SortedOracles.contract, event: "OracleReportRemoved", logs: logs, sub: sub}, nil
}

// WatchOracleReportRemoved is a free log subscription operation binding the contract event 0xe21a44017b6fa1658d84e937d56ff408501facdb4ff7427c479ac460d76f7893.
//
// Solidity: event OracleReportRemoved(address indexed token, address indexed oracle)
func (_SortedOracles *SortedOraclesFilterer) WatchOracleReportRemoved(opts *bind.WatchOpts, sink chan<- *SortedOraclesOracleReportRemoved, token []common.Address, oracle []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var oracleRule []interface{}
	for _, oracleItem := range oracle {
		oracleRule = append(oracleRule, oracleItem)
	}

	logs, sub, err := _SortedOracles.contract.WatchLogs(opts, "OracleReportRemoved", tokenRule, oracleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SortedOraclesOracleReportRemoved)
				if err := _SortedOracles.contract.UnpackLog(event, "OracleReportRemoved", log); err != nil {
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

// ParseOracleReportRemoved is a log parse operation binding the contract event 0xe21a44017b6fa1658d84e937d56ff408501facdb4ff7427c479ac460d76f7893.
//
// Solidity: event OracleReportRemoved(address indexed token, address indexed oracle)
func (_SortedOracles *SortedOraclesFilterer) ParseOracleReportRemoved(log types.Log) (*SortedOraclesOracleReportRemoved, error) {
	event := new(SortedOraclesOracleReportRemoved)
	if err := _SortedOracles.contract.UnpackLog(event, "OracleReportRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SortedOraclesOracleReportedIterator is returned from FilterOracleReported and is used to iterate over the raw logs and unpacked data for OracleReported events raised by the SortedOracles contract.
type SortedOraclesOracleReportedIterator struct {
	Event *SortedOraclesOracleReported // Event containing the contract specifics and raw log

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
func (it *SortedOraclesOracleReportedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SortedOraclesOracleReported)
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
		it.Event = new(SortedOraclesOracleReported)
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
func (it *SortedOraclesOracleReportedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SortedOraclesOracleReportedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SortedOraclesOracleReported represents a OracleReported event raised by the SortedOracles contract.
type SortedOraclesOracleReported struct {
	Token     common.Address
	Oracle    common.Address
	Timestamp *big.Int
	Value     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterOracleReported is a free log retrieval operation binding the contract event 0x7cebb17173a9ed273d2b7538f64395c0ebf352ff743f1cf8ce66b437a6144213.
//
// Solidity: event OracleReported(address indexed token, address indexed oracle, uint256 timestamp, uint256 value)
func (_SortedOracles *SortedOraclesFilterer) FilterOracleReported(opts *bind.FilterOpts, token []common.Address, oracle []common.Address) (*SortedOraclesOracleReportedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var oracleRule []interface{}
	for _, oracleItem := range oracle {
		oracleRule = append(oracleRule, oracleItem)
	}

	logs, sub, err := _SortedOracles.contract.FilterLogs(opts, "OracleReported", tokenRule, oracleRule)
	if err != nil {
		return nil, err
	}
	return &SortedOraclesOracleReportedIterator{contract: _SortedOracles.contract, event: "OracleReported", logs: logs, sub: sub}, nil
}

// WatchOracleReported is a free log subscription operation binding the contract event 0x7cebb17173a9ed273d2b7538f64395c0ebf352ff743f1cf8ce66b437a6144213.
//
// Solidity: event OracleReported(address indexed token, address indexed oracle, uint256 timestamp, uint256 value)
func (_SortedOracles *SortedOraclesFilterer) WatchOracleReported(opts *bind.WatchOpts, sink chan<- *SortedOraclesOracleReported, token []common.Address, oracle []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var oracleRule []interface{}
	for _, oracleItem := range oracle {
		oracleRule = append(oracleRule, oracleItem)
	}

	logs, sub, err := _SortedOracles.contract.WatchLogs(opts, "OracleReported", tokenRule, oracleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SortedOraclesOracleReported)
				if err := _SortedOracles.contract.UnpackLog(event, "OracleReported", log); err != nil {
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

// ParseOracleReported is a log parse operation binding the contract event 0x7cebb17173a9ed273d2b7538f64395c0ebf352ff743f1cf8ce66b437a6144213.
//
// Solidity: event OracleReported(address indexed token, address indexed oracle, uint256 timestamp, uint256 value)
func (_SortedOracles *SortedOraclesFilterer) ParseOracleReported(log types.Log) (*SortedOraclesOracleReported, error) {
	event := new(SortedOraclesOracleReported)
	if err := _SortedOracles.contract.UnpackLog(event, "OracleReported", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SortedOraclesOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SortedOracles contract.
type SortedOraclesOwnershipTransferredIterator struct {
	Event *SortedOraclesOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SortedOraclesOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SortedOraclesOwnershipTransferred)
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
		it.Event = new(SortedOraclesOwnershipTransferred)
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
func (it *SortedOraclesOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SortedOraclesOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SortedOraclesOwnershipTransferred represents a OwnershipTransferred event raised by the SortedOracles contract.
type SortedOraclesOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SortedOracles *SortedOraclesFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SortedOraclesOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SortedOracles.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SortedOraclesOwnershipTransferredIterator{contract: _SortedOracles.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SortedOracles *SortedOraclesFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SortedOraclesOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SortedOracles.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SortedOraclesOwnershipTransferred)
				if err := _SortedOracles.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_SortedOracles *SortedOraclesFilterer) ParseOwnershipTransferred(log types.Log) (*SortedOraclesOwnershipTransferred, error) {
	event := new(SortedOraclesOwnershipTransferred)
	if err := _SortedOracles.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SortedOraclesReportExpirySetIterator is returned from FilterReportExpirySet and is used to iterate over the raw logs and unpacked data for ReportExpirySet events raised by the SortedOracles contract.
type SortedOraclesReportExpirySetIterator struct {
	Event *SortedOraclesReportExpirySet // Event containing the contract specifics and raw log

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
func (it *SortedOraclesReportExpirySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SortedOraclesReportExpirySet)
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
		it.Event = new(SortedOraclesReportExpirySet)
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
func (it *SortedOraclesReportExpirySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SortedOraclesReportExpirySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SortedOraclesReportExpirySet represents a ReportExpirySet event raised by the SortedOracles contract.
type SortedOraclesReportExpirySet struct {
	ReportExpiry *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterReportExpirySet is a free log retrieval operation binding the contract event 0xc68a9b88effd8a11611ff410efbc83569f0031b7bc70dd455b61344c7f0a042f.
//
// Solidity: event ReportExpirySet(uint256 reportExpiry)
func (_SortedOracles *SortedOraclesFilterer) FilterReportExpirySet(opts *bind.FilterOpts) (*SortedOraclesReportExpirySetIterator, error) {

	logs, sub, err := _SortedOracles.contract.FilterLogs(opts, "ReportExpirySet")
	if err != nil {
		return nil, err
	}
	return &SortedOraclesReportExpirySetIterator{contract: _SortedOracles.contract, event: "ReportExpirySet", logs: logs, sub: sub}, nil
}

// WatchReportExpirySet is a free log subscription operation binding the contract event 0xc68a9b88effd8a11611ff410efbc83569f0031b7bc70dd455b61344c7f0a042f.
//
// Solidity: event ReportExpirySet(uint256 reportExpiry)
func (_SortedOracles *SortedOraclesFilterer) WatchReportExpirySet(opts *bind.WatchOpts, sink chan<- *SortedOraclesReportExpirySet) (event.Subscription, error) {

	logs, sub, err := _SortedOracles.contract.WatchLogs(opts, "ReportExpirySet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SortedOraclesReportExpirySet)
				if err := _SortedOracles.contract.UnpackLog(event, "ReportExpirySet", log); err != nil {
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

// ParseReportExpirySet is a log parse operation binding the contract event 0xc68a9b88effd8a11611ff410efbc83569f0031b7bc70dd455b61344c7f0a042f.
//
// Solidity: event ReportExpirySet(uint256 reportExpiry)
func (_SortedOracles *SortedOraclesFilterer) ParseReportExpirySet(log types.Log) (*SortedOraclesReportExpirySet, error) {
	event := new(SortedOraclesReportExpirySet)
	if err := _SortedOracles.contract.UnpackLog(event, "ReportExpirySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SortedOraclesTokenReportExpirySetIterator is returned from FilterTokenReportExpirySet and is used to iterate over the raw logs and unpacked data for TokenReportExpirySet events raised by the SortedOracles contract.
type SortedOraclesTokenReportExpirySetIterator struct {
	Event *SortedOraclesTokenReportExpirySet // Event containing the contract specifics and raw log

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
func (it *SortedOraclesTokenReportExpirySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SortedOraclesTokenReportExpirySet)
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
		it.Event = new(SortedOraclesTokenReportExpirySet)
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
func (it *SortedOraclesTokenReportExpirySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SortedOraclesTokenReportExpirySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SortedOraclesTokenReportExpirySet represents a TokenReportExpirySet event raised by the SortedOracles contract.
type SortedOraclesTokenReportExpirySet struct {
	Token        common.Address
	ReportExpiry *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterTokenReportExpirySet is a free log retrieval operation binding the contract event 0xf8324c8592dfd9991ee3e717351afe0a964605257959e3d99b0eb3d45bff9422.
//
// Solidity: event TokenReportExpirySet(address token, uint256 reportExpiry)
func (_SortedOracles *SortedOraclesFilterer) FilterTokenReportExpirySet(opts *bind.FilterOpts) (*SortedOraclesTokenReportExpirySetIterator, error) {

	logs, sub, err := _SortedOracles.contract.FilterLogs(opts, "TokenReportExpirySet")
	if err != nil {
		return nil, err
	}
	return &SortedOraclesTokenReportExpirySetIterator{contract: _SortedOracles.contract, event: "TokenReportExpirySet", logs: logs, sub: sub}, nil
}

// WatchTokenReportExpirySet is a free log subscription operation binding the contract event 0xf8324c8592dfd9991ee3e717351afe0a964605257959e3d99b0eb3d45bff9422.
//
// Solidity: event TokenReportExpirySet(address token, uint256 reportExpiry)
func (_SortedOracles *SortedOraclesFilterer) WatchTokenReportExpirySet(opts *bind.WatchOpts, sink chan<- *SortedOraclesTokenReportExpirySet) (event.Subscription, error) {

	logs, sub, err := _SortedOracles.contract.WatchLogs(opts, "TokenReportExpirySet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SortedOraclesTokenReportExpirySet)
				if err := _SortedOracles.contract.UnpackLog(event, "TokenReportExpirySet", log); err != nil {
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

// ParseTokenReportExpirySet is a log parse operation binding the contract event 0xf8324c8592dfd9991ee3e717351afe0a964605257959e3d99b0eb3d45bff9422.
//
// Solidity: event TokenReportExpirySet(address token, uint256 reportExpiry)
func (_SortedOracles *SortedOraclesFilterer) ParseTokenReportExpirySet(log types.Log) (*SortedOraclesTokenReportExpirySet, error) {
	event := new(SortedOraclesTokenReportExpirySet)
	if err := _SortedOracles.contract.UnpackLog(event, "TokenReportExpirySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
