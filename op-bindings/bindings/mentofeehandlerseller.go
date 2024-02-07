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

// MentoFeeHandlerSellerMetaData contains all meta data concerning the MentoFeeHandlerSeller contract.
var MentoFeeHandlerSellerMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"test\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"receive\",\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"calculateMinAmount\",\"inputs\":[{\"name\":\"midPriceNumerator\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"midPriceDenominator\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxSlippage\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"getVersionNumber\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_registryAddress\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"tokenAddresses\",\"type\":\"address[]\",\"internalType\":\"address[]\"},{\"name\":\"newMininumReports\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialized\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"minimumReports\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registry\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractICeloRegistry\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"sell\",\"inputs\":[{\"name\":\"sellTokenAddress\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"buyTokenAddress\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxSlippage\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMinimumReports\",\"inputs\":[{\"name\":\"tokenAddress\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"newMininumReports\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setRegistry\",\"inputs\":[{\"name\":\"registryAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transfer\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"MinimumReportsSet\",\"inputs\":[{\"name\":\"tokenAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"minimumReports\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RegistrySet\",\"inputs\":[{\"name\":\"registryAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TokenSold\",\"inputs\":[{\"name\":\"soldTokenAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"boughtTokenAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false}]",
	Bin: "0x60806040523480156200001157600080fd5b5060405162001f1c38038062001f1c8339810160408190526200003491620000b4565b8080620000413362000064565b806200005b576000805460ff60a01b1916600160a01b1790555b505050620000df565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b600060208284031215620000c757600080fd5b81518015158114620000d857600080fd5b9392505050565b611e2d80620000ef6000396000f3fe6080604052600436106100d65760003560e01c80637b1039991161007f578063dbba0f0111610059578063dbba0f011461028b578063e4187b13146102ab578063f2fde38b146102cb578063ff1d5752146102eb57600080fd5b80637b103999146101ee5780638da5cb5b14610240578063a91ee0dc1461026b57600080fd5b80634e008cdb116100b05780634e008cdb1461017957806354255be0146101a6578063715018a6146101d957600080fd5b8063158ef93e146100e25780632f257aa01461012957806331de7d151461014b57600080fd5b366100dd57005b600080fd5b3480156100ee57600080fd5b506000546101149074010000000000000000000000000000000000000000900460ff1681565b60405190151581526020015b60405180910390f35b34801561013557600080fd5b50610149610144366004611abd565b61030b565b005b34801561015757600080fd5b5061016b610166366004611ae9565b610321565b604051908152602001610120565b34801561018557600080fd5b5061016b610194366004611b2f565b60026020526000908152604090205481565b3480156101b257600080fd5b50600180600080604080519485526020850193909352918301526060820152608001610120565b3480156101e557600080fd5b50610149610b81565b3480156101fa57600080fd5b5060015461021b9073ffffffffffffffffffffffffffffffffffffffff1681565b60405173ffffffffffffffffffffffffffffffffffffffff9091168152602001610120565b34801561024c57600080fd5b5060005473ffffffffffffffffffffffffffffffffffffffff1661021b565b34801561027757600080fd5b50610149610286366004611b2f565b610b95565b34801561029757600080fd5b506101146102a6366004611b53565b610c89565b3480156102b757600080fd5b5061016b6102c6366004611b95565b610d34565b3480156102d757600080fd5b506101496102e6366004611b2f565b610db7565b3480156102f757600080fd5b50610149610306366004611c13565b610e6e565b610313610fb1565b61031d8282611032565b5050565b6001546040517f476f6c64546f6b656e0000000000000000000000000000000000000000000000602082015260009173ffffffffffffffffffffffffffffffffffffffff169063dcf0aaed90602901604051602081830303815290604052805190602001206040518263ffffffff1660e01b81526004016103a491815260200190565b602060405180830381865afa1580156103c1573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906103e59190611c96565b73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff161461047e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f42757920746f6b656e2063616e206f6e6c7920626520676f6c6420746f6b656e60448201526064015b60405180910390fd5b6040517f70a08231000000000000000000000000000000000000000000000000000000008152306004820152859073ffffffffffffffffffffffffffffffffffffffff8216906370a0823190602401602060405180830381865afa1580156104ea573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061050e9190611cb3565b84111561059d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f42616c616e6365206f6620746f6b656e20746f206275726e206e6f7420656e6f60448201527f75676800000000000000000000000000000000000000000000000000000000006064820152608401610475565b6000600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663dcf0aaed8373ffffffffffffffffffffffffffffffffffffffff166340a12f646040518163ffffffff1660e01b8152600401602060405180830381865afa158015610628573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061064c9190611cb3565b6040518263ffffffff1660e01b815260040161066a91815260200190565b602060405180830381865afa158015610687573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106ab9190611c96565b9050806000806106b9611097565b73ffffffffffffffffffffffffffffffffffffffff8b8116600081815260026020526040908190205490517fbbc66a9400000000000000000000000000000000000000000000000000000000815260048101929092529293509083169063bbc66a9490602401602060405180830381865afa15801561073c573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906107609190611cb3565b10156107ee576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f4e756d626572206f66207265706f72747320666f7220746f6b656e206e6f742060448201527f656e6f75676800000000000000000000000000000000000000000000000000006064820152608401610475565b6040517fef90e1b000000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8b81166004830152600091829184169063ef90e1b0906024016040805180830381865afa15801561085e573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906108829190611ccc565b9150915061089282828c8c610d34565b6040517f095ea7b300000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8881166004830152602482018d90529195509088169063095ea7b3906044016020604051808303816000875af115801561090b573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061092f9190611cf0565b506040517f8ab1a5d4000000000000000000000000000000000000000000000000000000008152600481018b9052602481018590526000604482015273ffffffffffffffffffffffffffffffffffffffff861690638ab1a5d4906064016020604051808303816000875af11580156109ab573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109cf9190611cb3565b5060006109da611161565b6040517f70a0823100000000000000000000000000000000000000000000000000000000815230600482015290915060009073ffffffffffffffffffffffffffffffffffffffff8316906370a0823190602401602060405180830381865afa158015610a4a573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610a6e9190611cb3565b6040517fa9059cbb0000000000000000000000000000000000000000000000000000000081523360048201526024810182905290915073ffffffffffffffffffffffffffffffffffffffff83169063a9059cbb906044016020604051808303816000875af1158015610ae4573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610b089190611cf0565b507fd4cffd6979677853b45a7a17f455188a434e975ba474c5a2613c94beacea537a8e8e8e604051610b689392919073ffffffffffffffffffffffffffffffffffffffff9384168152919092166020820152604081019190915260600190565b60405180910390a19d9c50505050505050505050505050565b610b89610fb1565b610b9360006111b5565b565b610b9d610fb1565b73ffffffffffffffffffffffffffffffffffffffff8116610c1a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f43616e6e6f7420726567697374657220746865206e756c6c20616464726573736044820152606401610475565b600180547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff83169081179091556040517f27fe5f0c1c3b1ed427cc63d0f05759ffdecf9aec9e18d31ef366fc8a6cb5dc3b90600090a250565b6000610c93610fb1565b6040517fa9059cbb00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff83811660048301526024820185905285169063a9059cbb906044016020604051808303816000875af1158015610d08573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610d2c9190611cf0565b949350505050565b600080610d598360408051602080820183526000909152815190810190915290815290565b90506000610d67878761122a565b90506000610d7486611268565b90506000610d828383611347565b9050610daa610da5610d9e84610d988789611347565b90611347565b839061177e565b61181c565b9998505050505050505050565b610dbf610fb1565b73ffffffffffffffffffffffffffffffffffffffff8116610e62576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f64647265737300000000000000000000000000000000000000000000000000006064820152608401610475565b610e6b816111b5565b50565b60005474010000000000000000000000000000000000000000900460ff1615610ef3576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601c60248201527f636f6e747261637420616c726561647920696e697469616c697a6564000000006044820152606401610475565b600080547fffffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffff1674010000000000000000000000000000000000000000179055610f3b336111b5565b610f4485610b95565b60005b83811015610fa957610f97858583818110610f6457610f64611d12565b9050602002016020810190610f799190611b2f565b848484818110610f8b57610f8b611d12565b90506020020135611032565b80610fa181611d70565b915050610f47565b505050505050565b60005473ffffffffffffffffffffffffffffffffffffffff163314610b93576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e65726044820152606401610475565b73ffffffffffffffffffffffffffffffffffffffff8216600081815260026020908152604091829020849055815192835282018390527f03cc7dddcb89dd90027bd8fa62d09d1b5c49ce5d20f8c9bb6bdeaaa62ea1718b910160405180910390a15050565b6001546040517f536f727465644f7261636c657300000000000000000000000000000000000000602082015260009173ffffffffffffffffffffffffffffffffffffffff169063dcf0aaed90602d015b604051602081830303815290604052805190602001206040518263ffffffff1660e01b815260040161111b91815260200190565b602060405180830381865afa158015611138573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061115c9190611c96565b905090565b6001546040517f476f6c64546f6b656e0000000000000000000000000000000000000000000000602082015260009173ffffffffffffffffffffffffffffffffffffffff169063dcf0aaed906029016110e7565b6000805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b604080516020810190915260008152600061124484611268565b9050600061125184611268565b905061125d8282611836565b925050505b92915050565b6040805160208101909152600081527601357c299a88ea76a58924d52ce4f26a85af186c2b9e7482111561131e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603660248201527f63616e277420637265617465206669786964697479206e756d626572206c617260448201527f676572207468616e206d61784e657746697865642829000000000000000000006064820152608401610475565b604051806020016040528069d3c21bcecceda10000008461133f9190611da8565b905292915050565b6040805160208101909152600081528251158061136357508151155b1561137d5750604080516020810190915260008152611262565b81517fffffffffffffffffffffffffffffffffffffffffffff2c3de43133125f000000016113ac575081611262565b82517fffffffffffffffffffffffffffffffffffffffffffff2c3de43133125f000000016113db575080611262565b600069d3c21bcecceda10000006113f18561196f565b516113fc9190611dbf565b90506000611409856119ae565b519050600069d3c21bcecceda10000006114228661196f565b5161142d9190611dbf565b9050600061143a866119ae565b51905060006114498386611da8565b905084156114c3578261145c8683611dbf565b146114c3576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601660248201527f6f766572666c6f772078317931206465746563746564000000000000000000006044820152606401610475565b60006114d969d3c21bcecceda100000083611da8565b9050811561155d5769d3c21bcecceda10000006114f68383611dbf565b1461155d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f6f766572666c6f772078317931202a20666978656431206465746563746564006044820152606401610475565b905080600061156c8587611da8565b905085156115e6578461157f8783611dbf565b146115e6576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601660248201527f6f766572666c6f772078327931206465746563746564000000000000000000006044820152606401610475565b60006115f28589611da8565b9050871561166c57846116058983611dbf565b1461166c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601660248201527f6f766572666c6f772078317932206465746563746564000000000000000000006044820152606401610475565b61167b64e8d4a5100088611dbf565b965061168c64e8d4a5100086611dbf565b9450600061169a8689611da8565b9050871561171457856116ad8983611dbf565b14611714576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601660248201527f6f766572666c6f772078327932206465746563746564000000000000000000006044820152606401610475565b60408051602080820183528782528251908101909252848252906117399082906119f9565b9050611753816040518060200160405280868152506119f9565b905061176d816040518060200160405280858152506119f9565b9d9c50505050505050505050505050565b6040805160208101909152600081528151835110156117f9576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f737562737472616374696f6e20756e646572666c6f77206465746563746564006044820152606401610475565b604080516020810190915282518451829161181391611dfa565b90529392505050565b80516000906112629069d3c21bcecceda100000090611dbf565b60408051602081019091526000815281516000036118b0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601160248201527f63616e27742064697669646520627920300000000000000000000000000000006044820152606401610475565b82516000906118ca9069d3c21bcecceda100000090611da8565b84519091506118e369d3c21bcecceda100000083611dbf565b1461194a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601260248201527f6f766572666c6f772061742064697669646500000000000000000000000000006044820152606401610475565b60405180602001604052808460000151836119659190611dbf565b9052949350505050565b604080516020810190915260008152604051806020016040528069d3c21bcecceda10000008085600001516119a49190611dbf565b61133f9190611da8565b604080516020810190915260008152604051806020016040528069d3c21bcecceda10000008085600001516119e39190611dbf565b6119ed9190611da8565b845161133f9190611dfa565b60408051602081019091526000815281518351600091611a1891611e0d565b8451909150811015611a86576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601560248201527f616464206f766572666c6f7720646574656374656400000000000000000000006044820152606401610475565b60408051602081019091529081529392505050565b73ffffffffffffffffffffffffffffffffffffffff81168114610e6b57600080fd5b60008060408385031215611ad057600080fd5b8235611adb81611a9b565b946020939093013593505050565b60008060008060808587031215611aff57600080fd5b8435611b0a81611a9b565b93506020850135611b1a81611a9b565b93969395505050506040820135916060013590565b600060208284031215611b4157600080fd5b8135611b4c81611a9b565b9392505050565b600080600060608486031215611b6857600080fd5b8335611b7381611a9b565b9250602084013591506040840135611b8a81611a9b565b809150509250925092565b60008060008060808587031215611bab57600080fd5b5050823594602084013594506040840135936060013592509050565b60008083601f840112611bd957600080fd5b50813567ffffffffffffffff811115611bf157600080fd5b6020830191508360208260051b8501011115611c0c57600080fd5b9250929050565b600080600080600060608688031215611c2b57600080fd5b8535611c3681611a9b565b9450602086013567ffffffffffffffff80821115611c5357600080fd5b611c5f89838a01611bc7565b90965094506040880135915080821115611c7857600080fd5b50611c8588828901611bc7565b969995985093965092949392505050565b600060208284031215611ca857600080fd5b8151611b4c81611a9b565b600060208284031215611cc557600080fd5b5051919050565b60008060408385031215611cdf57600080fd5b505080516020909101519092909150565b600060208284031215611d0257600080fd5b81518015158114611b4c57600080fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203611da157611da1611d41565b5060010190565b808202811582820484141761126257611262611d41565b600082611df5577f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b500490565b8181038181111561126257611262611d41565b8082018082111561126257611262611d4156fea164736f6c6343000813000a",
}

// MentoFeeHandlerSellerABI is the input ABI used to generate the binding from.
// Deprecated: Use MentoFeeHandlerSellerMetaData.ABI instead.
var MentoFeeHandlerSellerABI = MentoFeeHandlerSellerMetaData.ABI

// MentoFeeHandlerSellerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use MentoFeeHandlerSellerMetaData.Bin instead.
var MentoFeeHandlerSellerBin = MentoFeeHandlerSellerMetaData.Bin

// DeployMentoFeeHandlerSeller deploys a new Ethereum contract, binding an instance of MentoFeeHandlerSeller to it.
func DeployMentoFeeHandlerSeller(auth *bind.TransactOpts, backend bind.ContractBackend, test bool) (common.Address, *types.Transaction, *MentoFeeHandlerSeller, error) {
	parsed, err := MentoFeeHandlerSellerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(MentoFeeHandlerSellerBin), backend, test)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MentoFeeHandlerSeller{MentoFeeHandlerSellerCaller: MentoFeeHandlerSellerCaller{contract: contract}, MentoFeeHandlerSellerTransactor: MentoFeeHandlerSellerTransactor{contract: contract}, MentoFeeHandlerSellerFilterer: MentoFeeHandlerSellerFilterer{contract: contract}}, nil
}

// MentoFeeHandlerSeller is an auto generated Go binding around an Ethereum contract.
type MentoFeeHandlerSeller struct {
	MentoFeeHandlerSellerCaller     // Read-only binding to the contract
	MentoFeeHandlerSellerTransactor // Write-only binding to the contract
	MentoFeeHandlerSellerFilterer   // Log filterer for contract events
}

// MentoFeeHandlerSellerCaller is an auto generated read-only Go binding around an Ethereum contract.
type MentoFeeHandlerSellerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MentoFeeHandlerSellerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MentoFeeHandlerSellerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MentoFeeHandlerSellerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MentoFeeHandlerSellerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MentoFeeHandlerSellerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MentoFeeHandlerSellerSession struct {
	Contract     *MentoFeeHandlerSeller // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// MentoFeeHandlerSellerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MentoFeeHandlerSellerCallerSession struct {
	Contract *MentoFeeHandlerSellerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// MentoFeeHandlerSellerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MentoFeeHandlerSellerTransactorSession struct {
	Contract     *MentoFeeHandlerSellerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// MentoFeeHandlerSellerRaw is an auto generated low-level Go binding around an Ethereum contract.
type MentoFeeHandlerSellerRaw struct {
	Contract *MentoFeeHandlerSeller // Generic contract binding to access the raw methods on
}

// MentoFeeHandlerSellerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MentoFeeHandlerSellerCallerRaw struct {
	Contract *MentoFeeHandlerSellerCaller // Generic read-only contract binding to access the raw methods on
}

// MentoFeeHandlerSellerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MentoFeeHandlerSellerTransactorRaw struct {
	Contract *MentoFeeHandlerSellerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMentoFeeHandlerSeller creates a new instance of MentoFeeHandlerSeller, bound to a specific deployed contract.
func NewMentoFeeHandlerSeller(address common.Address, backend bind.ContractBackend) (*MentoFeeHandlerSeller, error) {
	contract, err := bindMentoFeeHandlerSeller(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MentoFeeHandlerSeller{MentoFeeHandlerSellerCaller: MentoFeeHandlerSellerCaller{contract: contract}, MentoFeeHandlerSellerTransactor: MentoFeeHandlerSellerTransactor{contract: contract}, MentoFeeHandlerSellerFilterer: MentoFeeHandlerSellerFilterer{contract: contract}}, nil
}

// NewMentoFeeHandlerSellerCaller creates a new read-only instance of MentoFeeHandlerSeller, bound to a specific deployed contract.
func NewMentoFeeHandlerSellerCaller(address common.Address, caller bind.ContractCaller) (*MentoFeeHandlerSellerCaller, error) {
	contract, err := bindMentoFeeHandlerSeller(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MentoFeeHandlerSellerCaller{contract: contract}, nil
}

// NewMentoFeeHandlerSellerTransactor creates a new write-only instance of MentoFeeHandlerSeller, bound to a specific deployed contract.
func NewMentoFeeHandlerSellerTransactor(address common.Address, transactor bind.ContractTransactor) (*MentoFeeHandlerSellerTransactor, error) {
	contract, err := bindMentoFeeHandlerSeller(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MentoFeeHandlerSellerTransactor{contract: contract}, nil
}

// NewMentoFeeHandlerSellerFilterer creates a new log filterer instance of MentoFeeHandlerSeller, bound to a specific deployed contract.
func NewMentoFeeHandlerSellerFilterer(address common.Address, filterer bind.ContractFilterer) (*MentoFeeHandlerSellerFilterer, error) {
	contract, err := bindMentoFeeHandlerSeller(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MentoFeeHandlerSellerFilterer{contract: contract}, nil
}

// bindMentoFeeHandlerSeller binds a generic wrapper to an already deployed contract.
func bindMentoFeeHandlerSeller(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MentoFeeHandlerSellerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MentoFeeHandlerSeller.Contract.MentoFeeHandlerSellerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.MentoFeeHandlerSellerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.MentoFeeHandlerSellerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MentoFeeHandlerSeller.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.contract.Transact(opts, method, params...)
}

// CalculateMinAmount is a free data retrieval call binding the contract method 0xe4187b13.
//
// Solidity: function calculateMinAmount(uint256 midPriceNumerator, uint256 midPriceDenominator, uint256 amount, uint256 maxSlippage) pure returns(uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCaller) CalculateMinAmount(opts *bind.CallOpts, midPriceNumerator *big.Int, midPriceDenominator *big.Int, amount *big.Int, maxSlippage *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _MentoFeeHandlerSeller.contract.Call(opts, &out, "calculateMinAmount", midPriceNumerator, midPriceDenominator, amount, maxSlippage)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateMinAmount is a free data retrieval call binding the contract method 0xe4187b13.
//
// Solidity: function calculateMinAmount(uint256 midPriceNumerator, uint256 midPriceDenominator, uint256 amount, uint256 maxSlippage) pure returns(uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) CalculateMinAmount(midPriceNumerator *big.Int, midPriceDenominator *big.Int, amount *big.Int, maxSlippage *big.Int) (*big.Int, error) {
	return _MentoFeeHandlerSeller.Contract.CalculateMinAmount(&_MentoFeeHandlerSeller.CallOpts, midPriceNumerator, midPriceDenominator, amount, maxSlippage)
}

// CalculateMinAmount is a free data retrieval call binding the contract method 0xe4187b13.
//
// Solidity: function calculateMinAmount(uint256 midPriceNumerator, uint256 midPriceDenominator, uint256 amount, uint256 maxSlippage) pure returns(uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCallerSession) CalculateMinAmount(midPriceNumerator *big.Int, midPriceDenominator *big.Int, amount *big.Int, maxSlippage *big.Int) (*big.Int, error) {
	return _MentoFeeHandlerSeller.Contract.CalculateMinAmount(&_MentoFeeHandlerSeller.CallOpts, midPriceNumerator, midPriceDenominator, amount, maxSlippage)
}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCaller) GetVersionNumber(opts *bind.CallOpts) (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	var out []interface{}
	err := _MentoFeeHandlerSeller.contract.Call(opts, &out, "getVersionNumber")

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
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) GetVersionNumber() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _MentoFeeHandlerSeller.Contract.GetVersionNumber(&_MentoFeeHandlerSeller.CallOpts)
}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCallerSession) GetVersionNumber() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _MentoFeeHandlerSeller.Contract.GetVersionNumber(&_MentoFeeHandlerSeller.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCaller) Initialized(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _MentoFeeHandlerSeller.contract.Call(opts, &out, "initialized")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) Initialized() (bool, error) {
	return _MentoFeeHandlerSeller.Contract.Initialized(&_MentoFeeHandlerSeller.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCallerSession) Initialized() (bool, error) {
	return _MentoFeeHandlerSeller.Contract.Initialized(&_MentoFeeHandlerSeller.CallOpts)
}

// MinimumReports is a free data retrieval call binding the contract method 0x4e008cdb.
//
// Solidity: function minimumReports(address ) view returns(uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCaller) MinimumReports(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _MentoFeeHandlerSeller.contract.Call(opts, &out, "minimumReports", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinimumReports is a free data retrieval call binding the contract method 0x4e008cdb.
//
// Solidity: function minimumReports(address ) view returns(uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) MinimumReports(arg0 common.Address) (*big.Int, error) {
	return _MentoFeeHandlerSeller.Contract.MinimumReports(&_MentoFeeHandlerSeller.CallOpts, arg0)
}

// MinimumReports is a free data retrieval call binding the contract method 0x4e008cdb.
//
// Solidity: function minimumReports(address ) view returns(uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCallerSession) MinimumReports(arg0 common.Address) (*big.Int, error) {
	return _MentoFeeHandlerSeller.Contract.MinimumReports(&_MentoFeeHandlerSeller.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MentoFeeHandlerSeller.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) Owner() (common.Address, error) {
	return _MentoFeeHandlerSeller.Contract.Owner(&_MentoFeeHandlerSeller.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCallerSession) Owner() (common.Address, error) {
	return _MentoFeeHandlerSeller.Contract.Owner(&_MentoFeeHandlerSeller.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCaller) Registry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MentoFeeHandlerSeller.contract.Call(opts, &out, "registry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) Registry() (common.Address, error) {
	return _MentoFeeHandlerSeller.Contract.Registry(&_MentoFeeHandlerSeller.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerCallerSession) Registry() (common.Address, error) {
	return _MentoFeeHandlerSeller.Contract.Registry(&_MentoFeeHandlerSeller.CallOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0xff1d5752.
//
// Solidity: function initialize(address _registryAddress, address[] tokenAddresses, uint256[] newMininumReports) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactor) Initialize(opts *bind.TransactOpts, _registryAddress common.Address, tokenAddresses []common.Address, newMininumReports []*big.Int) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.contract.Transact(opts, "initialize", _registryAddress, tokenAddresses, newMininumReports)
}

// Initialize is a paid mutator transaction binding the contract method 0xff1d5752.
//
// Solidity: function initialize(address _registryAddress, address[] tokenAddresses, uint256[] newMininumReports) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) Initialize(_registryAddress common.Address, tokenAddresses []common.Address, newMininumReports []*big.Int) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.Initialize(&_MentoFeeHandlerSeller.TransactOpts, _registryAddress, tokenAddresses, newMininumReports)
}

// Initialize is a paid mutator transaction binding the contract method 0xff1d5752.
//
// Solidity: function initialize(address _registryAddress, address[] tokenAddresses, uint256[] newMininumReports) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactorSession) Initialize(_registryAddress common.Address, tokenAddresses []common.Address, newMininumReports []*big.Int) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.Initialize(&_MentoFeeHandlerSeller.TransactOpts, _registryAddress, tokenAddresses, newMininumReports)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) RenounceOwnership() (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.RenounceOwnership(&_MentoFeeHandlerSeller.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.RenounceOwnership(&_MentoFeeHandlerSeller.TransactOpts)
}

// Sell is a paid mutator transaction binding the contract method 0x31de7d15.
//
// Solidity: function sell(address sellTokenAddress, address buyTokenAddress, uint256 amount, uint256 maxSlippage) returns(uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactor) Sell(opts *bind.TransactOpts, sellTokenAddress common.Address, buyTokenAddress common.Address, amount *big.Int, maxSlippage *big.Int) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.contract.Transact(opts, "sell", sellTokenAddress, buyTokenAddress, amount, maxSlippage)
}

// Sell is a paid mutator transaction binding the contract method 0x31de7d15.
//
// Solidity: function sell(address sellTokenAddress, address buyTokenAddress, uint256 amount, uint256 maxSlippage) returns(uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) Sell(sellTokenAddress common.Address, buyTokenAddress common.Address, amount *big.Int, maxSlippage *big.Int) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.Sell(&_MentoFeeHandlerSeller.TransactOpts, sellTokenAddress, buyTokenAddress, amount, maxSlippage)
}

// Sell is a paid mutator transaction binding the contract method 0x31de7d15.
//
// Solidity: function sell(address sellTokenAddress, address buyTokenAddress, uint256 amount, uint256 maxSlippage) returns(uint256)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactorSession) Sell(sellTokenAddress common.Address, buyTokenAddress common.Address, amount *big.Int, maxSlippage *big.Int) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.Sell(&_MentoFeeHandlerSeller.TransactOpts, sellTokenAddress, buyTokenAddress, amount, maxSlippage)
}

// SetMinimumReports is a paid mutator transaction binding the contract method 0x2f257aa0.
//
// Solidity: function setMinimumReports(address tokenAddress, uint256 newMininumReports) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactor) SetMinimumReports(opts *bind.TransactOpts, tokenAddress common.Address, newMininumReports *big.Int) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.contract.Transact(opts, "setMinimumReports", tokenAddress, newMininumReports)
}

// SetMinimumReports is a paid mutator transaction binding the contract method 0x2f257aa0.
//
// Solidity: function setMinimumReports(address tokenAddress, uint256 newMininumReports) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) SetMinimumReports(tokenAddress common.Address, newMininumReports *big.Int) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.SetMinimumReports(&_MentoFeeHandlerSeller.TransactOpts, tokenAddress, newMininumReports)
}

// SetMinimumReports is a paid mutator transaction binding the contract method 0x2f257aa0.
//
// Solidity: function setMinimumReports(address tokenAddress, uint256 newMininumReports) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactorSession) SetMinimumReports(tokenAddress common.Address, newMininumReports *big.Int) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.SetMinimumReports(&_MentoFeeHandlerSeller.TransactOpts, tokenAddress, newMininumReports)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registryAddress) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactor) SetRegistry(opts *bind.TransactOpts, registryAddress common.Address) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.contract.Transact(opts, "setRegistry", registryAddress)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registryAddress) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) SetRegistry(registryAddress common.Address) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.SetRegistry(&_MentoFeeHandlerSeller.TransactOpts, registryAddress)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registryAddress) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactorSession) SetRegistry(registryAddress common.Address) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.SetRegistry(&_MentoFeeHandlerSeller.TransactOpts, registryAddress)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address token, uint256 amount, address to) returns(bool)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactor) Transfer(opts *bind.TransactOpts, token common.Address, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.contract.Transact(opts, "transfer", token, amount, to)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address token, uint256 amount, address to) returns(bool)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) Transfer(token common.Address, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.Transfer(&_MentoFeeHandlerSeller.TransactOpts, token, amount, to)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address token, uint256 amount, address to) returns(bool)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactorSession) Transfer(token common.Address, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.Transfer(&_MentoFeeHandlerSeller.TransactOpts, token, amount, to)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.TransferOwnership(&_MentoFeeHandlerSeller.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.TransferOwnership(&_MentoFeeHandlerSeller.TransactOpts, newOwner)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerSession) Receive() (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.Receive(&_MentoFeeHandlerSeller.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerTransactorSession) Receive() (*types.Transaction, error) {
	return _MentoFeeHandlerSeller.Contract.Receive(&_MentoFeeHandlerSeller.TransactOpts)
}

// MentoFeeHandlerSellerMinimumReportsSetIterator is returned from FilterMinimumReportsSet and is used to iterate over the raw logs and unpacked data for MinimumReportsSet events raised by the MentoFeeHandlerSeller contract.
type MentoFeeHandlerSellerMinimumReportsSetIterator struct {
	Event *MentoFeeHandlerSellerMinimumReportsSet // Event containing the contract specifics and raw log

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
func (it *MentoFeeHandlerSellerMinimumReportsSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MentoFeeHandlerSellerMinimumReportsSet)
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
		it.Event = new(MentoFeeHandlerSellerMinimumReportsSet)
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
func (it *MentoFeeHandlerSellerMinimumReportsSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MentoFeeHandlerSellerMinimumReportsSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MentoFeeHandlerSellerMinimumReportsSet represents a MinimumReportsSet event raised by the MentoFeeHandlerSeller contract.
type MentoFeeHandlerSellerMinimumReportsSet struct {
	TokenAddress   common.Address
	MinimumReports *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterMinimumReportsSet is a free log retrieval operation binding the contract event 0x03cc7dddcb89dd90027bd8fa62d09d1b5c49ce5d20f8c9bb6bdeaaa62ea1718b.
//
// Solidity: event MinimumReportsSet(address tokenAddress, uint256 minimumReports)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) FilterMinimumReportsSet(opts *bind.FilterOpts) (*MentoFeeHandlerSellerMinimumReportsSetIterator, error) {

	logs, sub, err := _MentoFeeHandlerSeller.contract.FilterLogs(opts, "MinimumReportsSet")
	if err != nil {
		return nil, err
	}
	return &MentoFeeHandlerSellerMinimumReportsSetIterator{contract: _MentoFeeHandlerSeller.contract, event: "MinimumReportsSet", logs: logs, sub: sub}, nil
}

// WatchMinimumReportsSet is a free log subscription operation binding the contract event 0x03cc7dddcb89dd90027bd8fa62d09d1b5c49ce5d20f8c9bb6bdeaaa62ea1718b.
//
// Solidity: event MinimumReportsSet(address tokenAddress, uint256 minimumReports)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) WatchMinimumReportsSet(opts *bind.WatchOpts, sink chan<- *MentoFeeHandlerSellerMinimumReportsSet) (event.Subscription, error) {

	logs, sub, err := _MentoFeeHandlerSeller.contract.WatchLogs(opts, "MinimumReportsSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MentoFeeHandlerSellerMinimumReportsSet)
				if err := _MentoFeeHandlerSeller.contract.UnpackLog(event, "MinimumReportsSet", log); err != nil {
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

// ParseMinimumReportsSet is a log parse operation binding the contract event 0x03cc7dddcb89dd90027bd8fa62d09d1b5c49ce5d20f8c9bb6bdeaaa62ea1718b.
//
// Solidity: event MinimumReportsSet(address tokenAddress, uint256 minimumReports)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) ParseMinimumReportsSet(log types.Log) (*MentoFeeHandlerSellerMinimumReportsSet, error) {
	event := new(MentoFeeHandlerSellerMinimumReportsSet)
	if err := _MentoFeeHandlerSeller.contract.UnpackLog(event, "MinimumReportsSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MentoFeeHandlerSellerOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the MentoFeeHandlerSeller contract.
type MentoFeeHandlerSellerOwnershipTransferredIterator struct {
	Event *MentoFeeHandlerSellerOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *MentoFeeHandlerSellerOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MentoFeeHandlerSellerOwnershipTransferred)
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
		it.Event = new(MentoFeeHandlerSellerOwnershipTransferred)
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
func (it *MentoFeeHandlerSellerOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MentoFeeHandlerSellerOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MentoFeeHandlerSellerOwnershipTransferred represents a OwnershipTransferred event raised by the MentoFeeHandlerSeller contract.
type MentoFeeHandlerSellerOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*MentoFeeHandlerSellerOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MentoFeeHandlerSeller.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &MentoFeeHandlerSellerOwnershipTransferredIterator{contract: _MentoFeeHandlerSeller.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *MentoFeeHandlerSellerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MentoFeeHandlerSeller.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MentoFeeHandlerSellerOwnershipTransferred)
				if err := _MentoFeeHandlerSeller.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) ParseOwnershipTransferred(log types.Log) (*MentoFeeHandlerSellerOwnershipTransferred, error) {
	event := new(MentoFeeHandlerSellerOwnershipTransferred)
	if err := _MentoFeeHandlerSeller.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MentoFeeHandlerSellerRegistrySetIterator is returned from FilterRegistrySet and is used to iterate over the raw logs and unpacked data for RegistrySet events raised by the MentoFeeHandlerSeller contract.
type MentoFeeHandlerSellerRegistrySetIterator struct {
	Event *MentoFeeHandlerSellerRegistrySet // Event containing the contract specifics and raw log

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
func (it *MentoFeeHandlerSellerRegistrySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MentoFeeHandlerSellerRegistrySet)
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
		it.Event = new(MentoFeeHandlerSellerRegistrySet)
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
func (it *MentoFeeHandlerSellerRegistrySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MentoFeeHandlerSellerRegistrySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MentoFeeHandlerSellerRegistrySet represents a RegistrySet event raised by the MentoFeeHandlerSeller contract.
type MentoFeeHandlerSellerRegistrySet struct {
	RegistryAddress common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterRegistrySet is a free log retrieval operation binding the contract event 0x27fe5f0c1c3b1ed427cc63d0f05759ffdecf9aec9e18d31ef366fc8a6cb5dc3b.
//
// Solidity: event RegistrySet(address indexed registryAddress)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) FilterRegistrySet(opts *bind.FilterOpts, registryAddress []common.Address) (*MentoFeeHandlerSellerRegistrySetIterator, error) {

	var registryAddressRule []interface{}
	for _, registryAddressItem := range registryAddress {
		registryAddressRule = append(registryAddressRule, registryAddressItem)
	}

	logs, sub, err := _MentoFeeHandlerSeller.contract.FilterLogs(opts, "RegistrySet", registryAddressRule)
	if err != nil {
		return nil, err
	}
	return &MentoFeeHandlerSellerRegistrySetIterator{contract: _MentoFeeHandlerSeller.contract, event: "RegistrySet", logs: logs, sub: sub}, nil
}

// WatchRegistrySet is a free log subscription operation binding the contract event 0x27fe5f0c1c3b1ed427cc63d0f05759ffdecf9aec9e18d31ef366fc8a6cb5dc3b.
//
// Solidity: event RegistrySet(address indexed registryAddress)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) WatchRegistrySet(opts *bind.WatchOpts, sink chan<- *MentoFeeHandlerSellerRegistrySet, registryAddress []common.Address) (event.Subscription, error) {

	var registryAddressRule []interface{}
	for _, registryAddressItem := range registryAddress {
		registryAddressRule = append(registryAddressRule, registryAddressItem)
	}

	logs, sub, err := _MentoFeeHandlerSeller.contract.WatchLogs(opts, "RegistrySet", registryAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MentoFeeHandlerSellerRegistrySet)
				if err := _MentoFeeHandlerSeller.contract.UnpackLog(event, "RegistrySet", log); err != nil {
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

// ParseRegistrySet is a log parse operation binding the contract event 0x27fe5f0c1c3b1ed427cc63d0f05759ffdecf9aec9e18d31ef366fc8a6cb5dc3b.
//
// Solidity: event RegistrySet(address indexed registryAddress)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) ParseRegistrySet(log types.Log) (*MentoFeeHandlerSellerRegistrySet, error) {
	event := new(MentoFeeHandlerSellerRegistrySet)
	if err := _MentoFeeHandlerSeller.contract.UnpackLog(event, "RegistrySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MentoFeeHandlerSellerTokenSoldIterator is returned from FilterTokenSold and is used to iterate over the raw logs and unpacked data for TokenSold events raised by the MentoFeeHandlerSeller contract.
type MentoFeeHandlerSellerTokenSoldIterator struct {
	Event *MentoFeeHandlerSellerTokenSold // Event containing the contract specifics and raw log

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
func (it *MentoFeeHandlerSellerTokenSoldIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MentoFeeHandlerSellerTokenSold)
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
		it.Event = new(MentoFeeHandlerSellerTokenSold)
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
func (it *MentoFeeHandlerSellerTokenSoldIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MentoFeeHandlerSellerTokenSoldIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MentoFeeHandlerSellerTokenSold represents a TokenSold event raised by the MentoFeeHandlerSeller contract.
type MentoFeeHandlerSellerTokenSold struct {
	SoldTokenAddress   common.Address
	BoughtTokenAddress common.Address
	Amount             *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterTokenSold is a free log retrieval operation binding the contract event 0xd4cffd6979677853b45a7a17f455188a434e975ba474c5a2613c94beacea537a.
//
// Solidity: event TokenSold(address soldTokenAddress, address boughtTokenAddress, uint256 amount)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) FilterTokenSold(opts *bind.FilterOpts) (*MentoFeeHandlerSellerTokenSoldIterator, error) {

	logs, sub, err := _MentoFeeHandlerSeller.contract.FilterLogs(opts, "TokenSold")
	if err != nil {
		return nil, err
	}
	return &MentoFeeHandlerSellerTokenSoldIterator{contract: _MentoFeeHandlerSeller.contract, event: "TokenSold", logs: logs, sub: sub}, nil
}

// WatchTokenSold is a free log subscription operation binding the contract event 0xd4cffd6979677853b45a7a17f455188a434e975ba474c5a2613c94beacea537a.
//
// Solidity: event TokenSold(address soldTokenAddress, address boughtTokenAddress, uint256 amount)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) WatchTokenSold(opts *bind.WatchOpts, sink chan<- *MentoFeeHandlerSellerTokenSold) (event.Subscription, error) {

	logs, sub, err := _MentoFeeHandlerSeller.contract.WatchLogs(opts, "TokenSold")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MentoFeeHandlerSellerTokenSold)
				if err := _MentoFeeHandlerSeller.contract.UnpackLog(event, "TokenSold", log); err != nil {
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

// ParseTokenSold is a log parse operation binding the contract event 0xd4cffd6979677853b45a7a17f455188a434e975ba474c5a2613c94beacea537a.
//
// Solidity: event TokenSold(address soldTokenAddress, address boughtTokenAddress, uint256 amount)
func (_MentoFeeHandlerSeller *MentoFeeHandlerSellerFilterer) ParseTokenSold(log types.Log) (*MentoFeeHandlerSellerTokenSold, error) {
	event := new(MentoFeeHandlerSellerTokenSold)
	if err := _MentoFeeHandlerSeller.contract.UnpackLog(event, "TokenSold", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
