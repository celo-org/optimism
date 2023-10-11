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

// UniswapFeeHandlerSellerMetaData contains all meta data concerning the UniswapFeeHandlerSeller contract.
var UniswapFeeHandlerSellerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"test\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"minimumReports\",\"type\":\"uint256\"}],\"name\":\"MinimumReportsSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"tokneAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"router\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"quote\",\"type\":\"uint256\"}],\"name\":\"ReceivedQuote\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"registryAddress\",\"type\":\"address\"}],\"name\":\"RegistrySet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"router\",\"type\":\"address\"}],\"name\":\"RouterAddressRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"router\",\"type\":\"address\"}],\"name\":\"RouterAddressSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"router\",\"type\":\"address\"}],\"name\":\"RouterUsed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"soldTokenAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"boughtTokenAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"TokenSold\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"midPriceNumerator\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"midPriceDenominator\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxSlippage\",\"type\":\"uint256\"}],\"name\":\"calculateMinAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getRoutersForToken\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getVersionNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_registryAddress\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"tokenAddresses\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"newMininumReports\",\"type\":\"uint256[]\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialized\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"minimumReports\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"registry\",\"outputs\":[{\"internalType\":\"contractICeloRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"router\",\"type\":\"address\"}],\"name\":\"removeRouter\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sellTokenAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"buyTokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxSlippage\",\"type\":\"uint256\"}],\"name\":\"sell\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"newMininumReports\",\"type\":\"uint256\"}],\"name\":\"setMinimumReports\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"registryAddress\",\"type\":\"address\"}],\"name\":\"setRegistry\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"router\",\"type\":\"address\"}],\"name\":\"setRouter\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60806040523480156200001157600080fd5b5060405162002ad538038062002ad58339810160408190526200003491620000b4565b8080620000413362000064565b806200005b576000805460ff60a01b1916600160a01b1790555b505050620000df565b600080546001600160a01b038381166001600160a01b0319831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b600060208284031215620000c757600080fd5b81518015158114620000d857600080fd5b9392505050565b6129e680620000ef6000396000f3fe6080604052600436106100f75760003560e01c806373f0a0701161008a578063dbba0f0111610059578063dbba0f011461031d578063e4187b131461033d578063f2fde38b1461035d578063ff1d57521461037d57600080fd5b806373f0a070146102605780637b103999146102805780638da5cb5b146102d2578063a91ee0dc146102fd57600080fd5b806341d68b8f116100c657806341d68b8f146101cb5780634e008cdb146101eb57806354255be014610218578063715018a61461024b57600080fd5b80630c2fef1414610103578063158ef93e146101395780632f257aa01461017b57806331de7d151461019d57600080fd5b366100fe57005b600080fd5b34801561010f57600080fd5b5061012361011e366004612444565b61039d565b60405161013091906124b2565b60405180910390f35b34801561014557600080fd5b5060005461016b9074010000000000000000000000000000000000000000900460ff1681565b6040519015158152602001610130565b34801561018757600080fd5b5061019b6101963660046124c5565b6103d4565b005b3480156101a957600080fd5b506101bd6101b83660046124f1565b6103ea565b604051908152602001610130565b3480156101d757600080fd5b5061019b6101e6366004612537565b610c45565b3480156101f757600080fd5b506101bd610206366004612444565b60026020526000908152604090205481565b34801561022457600080fd5b50600180600080604080519485526020850193909352918301526060820152608001610130565b34801561025757600080fd5b5061019b610c57565b34801561026c57600080fd5b5061019b61027b366004612537565b610c6b565b34801561028c57600080fd5b506001546102ad9073ffffffffffffffffffffffffffffffffffffffff1681565b60405173ffffffffffffffffffffffffffffffffffffffff9091168152602001610130565b3480156102de57600080fd5b5060005473ffffffffffffffffffffffffffffffffffffffff166102ad565b34801561030957600080fd5b5061019b610318366004612444565b610cf8565b34801561032957600080fd5b5061016b610338366004612570565b610dec565b34801561034957600080fd5b506101bd6103583660046125b2565b610e97565b34801561036957600080fd5b5061019b610378366004612444565b610f1a565b34801561038957600080fd5b5061019b610398366004612630565b610fd1565b73ffffffffffffffffffffffffffffffffffffffff811660009081526003602052604090206060906103ce90611114565b92915050565b6103dc611128565b6103e682826111a9565b5050565b6001546040517f476f6c64546f6b656e0000000000000000000000000000000000000000000000602082015260009173ffffffffffffffffffffffffffffffffffffffff169063dcf0aaed90602901604051602081830303815290604052805190602001206040518263ffffffff1660e01b815260040161046d91815260200190565b602060405180830381865afa15801561048a573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104ae91906126b3565b73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1614610547576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f42757920746f6b656e2063616e206f6e6c7920626520676f6c6420746f6b656e60448201526064015b60405180910390fd5b73ffffffffffffffffffffffffffffffffffffffff8516600090815260036020526040812061057590611114565b5111610603576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f726f757465724164647265737365732073686f756c64206265206e6f6e20656d60448201527f7074790000000000000000000000000000000000000000000000000000000000606482015260840161053e565b600061060d611207565b604080516002808252606082018352929350600092839283929190602083019080368337019050509050888160008151811061064b5761064b6126ff565b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508381600181518110610699576106996126ff565b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505060005b73ffffffffffffffffffffffffffffffffffffffff8a16600090815260036020526040902061070490611114565b518110156108a25773ffffffffffffffffffffffffffffffffffffffff8a16600090815260036020526040812061073b90836112d1565b6040517fd06ca61f000000000000000000000000000000000000000000000000000000008152909150819060009073ffffffffffffffffffffffffffffffffffffffff83169063d06ca61f90610797908e90899060040161272e565b600060405180830381865afa1580156107b4573d6000803e3d6000fd5b505050506040513d6000823e601f3d9081017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe01682016040526107fa9190810190612747565b60018151811061080c5761080c6126ff565b602002602001015190508273ffffffffffffffffffffffffffffffffffffffff168d73ffffffffffffffffffffffffffffffffffffffff167fba55c28acee19777ec6c603117b386d3e3b39886c0d3d53bc244be24ee6e7c848360405161087591815260200190565b60405180910390a38581111561088c578095508196505b505050808061089a90612852565b9150506106d6565b508160000361090d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601e60248201527f43616e27742065786368616e67652077697468207a65726f2071756f74650000604482015260640161053e565b600061091b8a888a876112dd565b6040517f095ea7b300000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8681166004830152602482018b9052919250908b169063095ea7b3906044016020604051808303816000875af1158015610994573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109b8919061288a565b5073ffffffffffffffffffffffffffffffffffffffff84166338ed1739898385306109e46014426128ac565b6040518663ffffffff1660e01b8152600401610a049594939291906128bf565b6000604051808303816000875af1158015610a23573d6000803e3d6000fd5b505050506040513d6000823e601f3d9081017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0168201604052610a699190810190612747565b506040517f70a0823100000000000000000000000000000000000000000000000000000000815230600482015260009073ffffffffffffffffffffffffffffffffffffffff8716906370a0823190602401602060405180830381865afa158015610ad7573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610afb9190612908565b6040517fa9059cbb0000000000000000000000000000000000000000000000000000000081523360048201526024810182905290915073ffffffffffffffffffffffffffffffffffffffff87169063a9059cbb906044016020604051808303816000875af1158015610b71573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610b95919061288a565b5060405173ffffffffffffffffffffffffffffffffffffffff861681527f59afd9b3bf745a06d8739721397432b9f161243cee13868b9d6d6fca05b6e5519060200160405180910390a16040805173ffffffffffffffffffffffffffffffffffffffff808e1682528c1660208201529081018a90527fd4cffd6979677853b45a7a17f455188a434e975ba474c5a2613c94beacea537a9060600160405180910390a19a9950505050505050505050565b610c4d611128565b6103e68282611733565b610c5f611128565b610c6960006118c7565b565b610c73611128565b73ffffffffffffffffffffffffffffffffffffffff82166000908152600360205260409020610ca2908261193c565b506040805173ffffffffffffffffffffffffffffffffffffffff8085168252831660208201527f044c4b00bcc14b6c00430f73b8bc07f33aecb2387c7b188142d6d497342de89a91015b60405180910390a15050565b610d00611128565b73ffffffffffffffffffffffffffffffffffffffff8116610d7d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f43616e6e6f7420726567697374657220746865206e756c6c2061646472657373604482015260640161053e565b600180547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff83169081179091556040517f27fe5f0c1c3b1ed427cc63d0f05759ffdecf9aec9e18d31ef366fc8a6cb5dc3b90600090a250565b6000610df6611128565b6040517fa9059cbb00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff83811660048301526024820185905285169063a9059cbb906044016020604051808303816000875af1158015610e6b573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610e8f919061288a565b949350505050565b600080610ebc8360408051602080820183526000909152815190810190915290815290565b90506000610eca878761195e565b90506000610ed78661199a565b90506000610ee58383611a79565b9050610f0d610f08610f0184610efb8789611a79565b90611a79565b8390611eb0565b611f4e565b9998505050505050505050565b610f22611128565b73ffffffffffffffffffffffffffffffffffffffff8116610fc5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f6464726573730000000000000000000000000000000000000000000000000000606482015260840161053e565b610fce816118c7565b50565b60005474010000000000000000000000000000000000000000900460ff1615611056576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601c60248201527f636f6e747261637420616c726561647920696e697469616c697a656400000000604482015260640161053e565b600080547fffffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffff167401000000000000000000000000000000000000000017905561109e336118c7565b6110a785610cf8565b60005b8381101561110c576110fa8585838181106110c7576110c76126ff565b90506020020160208101906110dc9190612444565b8484848181106110ee576110ee6126ff565b905060200201356111a9565b8061110481612852565b9150506110aa565b505050505050565b6060600061112183611f68565b9392505050565b60005473ffffffffffffffffffffffffffffffffffffffff163314610c69576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015260640161053e565b73ffffffffffffffffffffffffffffffffffffffff8216600081815260026020908152604091829020849055815192835282018390527f03cc7dddcb89dd90027bd8fa62d09d1b5c49ce5d20f8c9bb6bdeaaa62ea1718b9101610cec565b6001546040517f476f6c64546f6b656e0000000000000000000000000000000000000000000000602082015260009173ffffffffffffffffffffffffffffffffffffffff169063dcf0aaed906029015b604051602081830303815290604052805190602001206040518263ffffffff1660e01b815260040161128b91815260200190565b602060405180830381865afa1580156112a8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906112cc91906126b3565b905090565b60006111218383611fc4565b6000806112e8611fee565b73ffffffffffffffffffffffffffffffffffffffff878116600081815260026020526040908190205490517fbbc66a940000000000000000000000000000000000000000000000000000000081526004810192909252929350829184169063bbc66a9490602401602060405180830381865afa15801561136c573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906113909190612908565b101561141e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f4e756d626572206f66207265706f72747320666f7220746f6b656e206e6f742060448201527f656e6f7567680000000000000000000000000000000000000000000000000000606482015260840161053e565b600081156114cf576040517fef90e1b000000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8981166004830152600091829186169063ef90e1b0906024016040805180830381865afa158015611496573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906114ba9190612921565b915091506114ca82828a8c610e97565b925050505b60006114d9611207565b905060008673ffffffffffffffffffffffffffffffffffffffff1663c45a01556040518163ffffffff1660e01b8152600401602060405180830381865afa158015611528573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061154c91906126b3565b6040517fe6a4390500000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8c811660048301528481166024830152919091169063e6a4390590604401602060405180830381865afa1580156115c2573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906115e691906126b3565b6040517f70a0823100000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff808316600483015291925060009161171891908d16906370a0823190602401602060405180830381865afa15801561165d573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906116819190612908565b6040517f70a0823100000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff85811660048301528616906370a0823190602401602060405180830381865afa1580156116ed573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906117119190612908565b8b8d610e97565b90506117248185612042565b9b9a5050505050505050505050565b73ffffffffffffffffffffffffffffffffffffffff81166117b0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601c60248201527f526f757465722063616e27742062652061646472657373207a65726f00000000604482015260640161053e565b73ffffffffffffffffffffffffffffffffffffffff821660009081526003602052604090206117df9082612059565b5073ffffffffffffffffffffffffffffffffffffffff8216600090815260036020819052604090912061181190611114565b51111561187a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f4d6178206e756d626572206f6620726f75746572732072656163686564000000604482015260640161053e565b6040805173ffffffffffffffffffffffffffffffffffffffff8085168252831660208201527fb3cbb74e835466bdbf8838b1acb70fa4a8b73e1a00cd5bacb9f68cf4dfc79cf39101610cec565b6000805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681178455604051919092169283917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e09190a35050565b60006111218373ffffffffffffffffffffffffffffffffffffffff841661207b565b60408051602081019091526000815260006119788461199a565b905060006119858461199a565b9050611991828261216e565b95945050505050565b6040805160208101909152600081527601357c299a88ea76a58924d52ce4f26a85af186c2b9e74821115611a50576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603660248201527f63616e277420637265617465206669786964697479206e756d626572206c617260448201527f676572207468616e206d61784e65774669786564282900000000000000000000606482015260840161053e565b604051806020016040528069d3c21bcecceda100000084611a719190612945565b905292915050565b60408051602081019091526000815282511580611a9557508151155b15611aaf57506040805160208101909152600081526103ce565b81517fffffffffffffffffffffffffffffffffffffffffffff2c3de43133125f00000001611ade5750816103ce565b82517fffffffffffffffffffffffffffffffffffffffffffff2c3de43133125f00000001611b0d5750806103ce565b600069d3c21bcecceda1000000611b23856122a7565b51611b2e919061295c565b90506000611b3b856122e6565b519050600069d3c21bcecceda1000000611b54866122a7565b51611b5f919061295c565b90506000611b6c866122e6565b5190506000611b7b8386612945565b90508415611bf55782611b8e868361295c565b14611bf5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601660248201527f6f766572666c6f77207831793120646574656374656400000000000000000000604482015260640161053e565b6000611c0b69d3c21bcecceda100000083612945565b90508115611c8f5769d3c21bcecceda1000000611c28838361295c565b14611c8f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f6f766572666c6f772078317931202a2066697865643120646574656374656400604482015260640161053e565b9050806000611c9e8587612945565b90508515611d185784611cb1878361295c565b14611d18576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601660248201527f6f766572666c6f77207832793120646574656374656400000000000000000000604482015260640161053e565b6000611d248589612945565b90508715611d9e5784611d37898361295c565b14611d9e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601660248201527f6f766572666c6f77207831793220646574656374656400000000000000000000604482015260640161053e565b611dad64e8d4a510008861295c565b9650611dbe64e8d4a510008661295c565b94506000611dcc8689612945565b90508715611e465785611ddf898361295c565b14611e46576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601660248201527f6f766572666c6f77207832793220646574656374656400000000000000000000604482015260640161053e565b6040805160208082018352878252825190810190925284825290611e6b908290612331565b9050611e8581604051806020016040528086815250612331565b9050611e9f81604051806020016040528085815250612331565b9d9c50505050505050505050505050565b604080516020810190915260008152815183511015611f2b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f737562737472616374696f6e20756e646572666c6f7720646574656374656400604482015260640161053e565b6040805160208101909152825184518291611f4591612997565b90529392505050565b80516000906103ce9069d3c21bcecceda10000009061295c565b606081600001805480602002602001604051908101604052809291908181526020018280548015611fb857602002820191906000526020600020905b815481526020019060010190808311611fa4575b50505050509050919050565b6000826000018281548110611fdb57611fdb6126ff565b9060005260206000200154905092915050565b6001546040517f536f727465644f7261636c657300000000000000000000000000000000000000602082015260009173ffffffffffffffffffffffffffffffffffffffff169063dcf0aaed90602d01611257565b6000818310156120525781611121565b5090919050565b60006111218373ffffffffffffffffffffffffffffffffffffffff84166123d3565b6000818152600183016020526040812054801561216457600061209f600183612997565b85549091506000906120b390600190612997565b90508181146121185760008660000182815481106120d3576120d36126ff565b90600052602060002001549050808760000184815481106120f6576120f66126ff565b6000918252602080832090910192909255918252600188019052604090208390555b8554869080612129576121296129aa565b6001900381819060005260206000200160009055905585600101600086815260200190815260200160002060009055600193505050506103ce565b60009150506103ce565b60408051602081019091526000815281516000036121e8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601160248201527f63616e2774206469766964652062792030000000000000000000000000000000604482015260640161053e565b82516000906122029069d3c21bcecceda100000090612945565b845190915061221b69d3c21bcecceda10000008361295c565b14612282576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601260248201527f6f766572666c6f77206174206469766964650000000000000000000000000000604482015260640161053e565b604051806020016040528084600001518361229d919061295c565b9052949350505050565b604080516020810190915260008152604051806020016040528069d3c21bcecceda10000008085600001516122dc919061295c565b611a719190612945565b604080516020810190915260008152604051806020016040528069d3c21bcecceda100000080856000015161231b919061295c565b6123259190612945565b8451611a719190612997565b60408051602081019091526000815281518351600091612350916128ac565b84519091508110156123be576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601560248201527f616464206f766572666c6f772064657465637465640000000000000000000000604482015260640161053e565b60408051602081019091529081529392505050565b600081815260018301602052604081205461241a575081546001818101845560008481526020808220909301849055845484825282860190935260409020919091556103ce565b5060006103ce565b73ffffffffffffffffffffffffffffffffffffffff81168114610fce57600080fd5b60006020828403121561245657600080fd5b813561112181612422565b600081518084526020808501945080840160005b838110156124a757815173ffffffffffffffffffffffffffffffffffffffff1687529582019590820190600101612475565b509495945050505050565b6020815260006111216020830184612461565b600080604083850312156124d857600080fd5b82356124e381612422565b946020939093013593505050565b6000806000806080858703121561250757600080fd5b843561251281612422565b9350602085013561252281612422565b93969395505050506040820135916060013590565b6000806040838503121561254a57600080fd5b823561255581612422565b9150602083013561256581612422565b809150509250929050565b60008060006060848603121561258557600080fd5b833561259081612422565b92506020840135915060408401356125a781612422565b809150509250925092565b600080600080608085870312156125c857600080fd5b5050823594602084013594506040840135936060013592509050565b60008083601f8401126125f657600080fd5b50813567ffffffffffffffff81111561260e57600080fd5b6020830191508360208260051b850101111561262957600080fd5b9250929050565b60008060008060006060868803121561264857600080fd5b853561265381612422565b9450602086013567ffffffffffffffff8082111561267057600080fd5b61267c89838a016125e4565b9096509450604088013591508082111561269557600080fd5b506126a2888289016125e4565b969995985093965092949392505050565b6000602082840312156126c557600080fd5b815161112181612422565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b828152604060208201526000610e8f6040830184612461565b6000602080838503121561275a57600080fd5b825167ffffffffffffffff8082111561277257600080fd5b818501915085601f83011261278657600080fd5b815181811115612798576127986126d0565b8060051b6040517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0603f830116810181811085821117156127db576127db6126d0565b6040529182528482019250838101850191888311156127f957600080fd5b938501935b82851015612817578451845293850193928501926127fe565b98975050505050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff820361288357612883612823565b5060010190565b60006020828403121561289c57600080fd5b8151801515811461112157600080fd5b808201808211156103ce576103ce612823565b85815284602082015260a0604082015260006128de60a0830186612461565b73ffffffffffffffffffffffffffffffffffffffff94909416606083015250608001529392505050565b60006020828403121561291a57600080fd5b5051919050565b6000806040838503121561293457600080fd5b505080516020909101519092909150565b80820281158282048414176103ce576103ce612823565b600082612992577f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b500490565b818103818111156103ce576103ce612823565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603160045260246000fdfea164736f6c6343000813000a",
}

// UniswapFeeHandlerSellerABI is the input ABI used to generate the binding from.
// Deprecated: Use UniswapFeeHandlerSellerMetaData.ABI instead.
var UniswapFeeHandlerSellerABI = UniswapFeeHandlerSellerMetaData.ABI

// UniswapFeeHandlerSellerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use UniswapFeeHandlerSellerMetaData.Bin instead.
var UniswapFeeHandlerSellerBin = UniswapFeeHandlerSellerMetaData.Bin

// DeployUniswapFeeHandlerSeller deploys a new Ethereum contract, binding an instance of UniswapFeeHandlerSeller to it.
func DeployUniswapFeeHandlerSeller(auth *bind.TransactOpts, backend bind.ContractBackend, test bool) (common.Address, *types.Transaction, *UniswapFeeHandlerSeller, error) {
	parsed, err := UniswapFeeHandlerSellerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(UniswapFeeHandlerSellerBin), backend, test)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &UniswapFeeHandlerSeller{UniswapFeeHandlerSellerCaller: UniswapFeeHandlerSellerCaller{contract: contract}, UniswapFeeHandlerSellerTransactor: UniswapFeeHandlerSellerTransactor{contract: contract}, UniswapFeeHandlerSellerFilterer: UniswapFeeHandlerSellerFilterer{contract: contract}}, nil
}

// UniswapFeeHandlerSeller is an auto generated Go binding around an Ethereum contract.
type UniswapFeeHandlerSeller struct {
	UniswapFeeHandlerSellerCaller     // Read-only binding to the contract
	UniswapFeeHandlerSellerTransactor // Write-only binding to the contract
	UniswapFeeHandlerSellerFilterer   // Log filterer for contract events
}

// UniswapFeeHandlerSellerCaller is an auto generated read-only Go binding around an Ethereum contract.
type UniswapFeeHandlerSellerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapFeeHandlerSellerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UniswapFeeHandlerSellerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapFeeHandlerSellerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UniswapFeeHandlerSellerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UniswapFeeHandlerSellerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UniswapFeeHandlerSellerSession struct {
	Contract     *UniswapFeeHandlerSeller // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// UniswapFeeHandlerSellerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UniswapFeeHandlerSellerCallerSession struct {
	Contract *UniswapFeeHandlerSellerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// UniswapFeeHandlerSellerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UniswapFeeHandlerSellerTransactorSession struct {
	Contract     *UniswapFeeHandlerSellerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// UniswapFeeHandlerSellerRaw is an auto generated low-level Go binding around an Ethereum contract.
type UniswapFeeHandlerSellerRaw struct {
	Contract *UniswapFeeHandlerSeller // Generic contract binding to access the raw methods on
}

// UniswapFeeHandlerSellerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UniswapFeeHandlerSellerCallerRaw struct {
	Contract *UniswapFeeHandlerSellerCaller // Generic read-only contract binding to access the raw methods on
}

// UniswapFeeHandlerSellerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UniswapFeeHandlerSellerTransactorRaw struct {
	Contract *UniswapFeeHandlerSellerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUniswapFeeHandlerSeller creates a new instance of UniswapFeeHandlerSeller, bound to a specific deployed contract.
func NewUniswapFeeHandlerSeller(address common.Address, backend bind.ContractBackend) (*UniswapFeeHandlerSeller, error) {
	contract, err := bindUniswapFeeHandlerSeller(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSeller{UniswapFeeHandlerSellerCaller: UniswapFeeHandlerSellerCaller{contract: contract}, UniswapFeeHandlerSellerTransactor: UniswapFeeHandlerSellerTransactor{contract: contract}, UniswapFeeHandlerSellerFilterer: UniswapFeeHandlerSellerFilterer{contract: contract}}, nil
}

// NewUniswapFeeHandlerSellerCaller creates a new read-only instance of UniswapFeeHandlerSeller, bound to a specific deployed contract.
func NewUniswapFeeHandlerSellerCaller(address common.Address, caller bind.ContractCaller) (*UniswapFeeHandlerSellerCaller, error) {
	contract, err := bindUniswapFeeHandlerSeller(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerCaller{contract: contract}, nil
}

// NewUniswapFeeHandlerSellerTransactor creates a new write-only instance of UniswapFeeHandlerSeller, bound to a specific deployed contract.
func NewUniswapFeeHandlerSellerTransactor(address common.Address, transactor bind.ContractTransactor) (*UniswapFeeHandlerSellerTransactor, error) {
	contract, err := bindUniswapFeeHandlerSeller(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerTransactor{contract: contract}, nil
}

// NewUniswapFeeHandlerSellerFilterer creates a new log filterer instance of UniswapFeeHandlerSeller, bound to a specific deployed contract.
func NewUniswapFeeHandlerSellerFilterer(address common.Address, filterer bind.ContractFilterer) (*UniswapFeeHandlerSellerFilterer, error) {
	contract, err := bindUniswapFeeHandlerSeller(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerFilterer{contract: contract}, nil
}

// bindUniswapFeeHandlerSeller binds a generic wrapper to an already deployed contract.
func bindUniswapFeeHandlerSeller(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(UniswapFeeHandlerSellerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _UniswapFeeHandlerSeller.Contract.UniswapFeeHandlerSellerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.UniswapFeeHandlerSellerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.UniswapFeeHandlerSellerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _UniswapFeeHandlerSeller.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.contract.Transact(opts, method, params...)
}

// CalculateMinAmount is a free data retrieval call binding the contract method 0xe4187b13.
//
// Solidity: function calculateMinAmount(uint256 midPriceNumerator, uint256 midPriceDenominator, uint256 amount, uint256 maxSlippage) pure returns(uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCaller) CalculateMinAmount(opts *bind.CallOpts, midPriceNumerator *big.Int, midPriceDenominator *big.Int, amount *big.Int, maxSlippage *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _UniswapFeeHandlerSeller.contract.Call(opts, &out, "calculateMinAmount", midPriceNumerator, midPriceDenominator, amount, maxSlippage)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateMinAmount is a free data retrieval call binding the contract method 0xe4187b13.
//
// Solidity: function calculateMinAmount(uint256 midPriceNumerator, uint256 midPriceDenominator, uint256 amount, uint256 maxSlippage) pure returns(uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) CalculateMinAmount(midPriceNumerator *big.Int, midPriceDenominator *big.Int, amount *big.Int, maxSlippage *big.Int) (*big.Int, error) {
	return _UniswapFeeHandlerSeller.Contract.CalculateMinAmount(&_UniswapFeeHandlerSeller.CallOpts, midPriceNumerator, midPriceDenominator, amount, maxSlippage)
}

// CalculateMinAmount is a free data retrieval call binding the contract method 0xe4187b13.
//
// Solidity: function calculateMinAmount(uint256 midPriceNumerator, uint256 midPriceDenominator, uint256 amount, uint256 maxSlippage) pure returns(uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCallerSession) CalculateMinAmount(midPriceNumerator *big.Int, midPriceDenominator *big.Int, amount *big.Int, maxSlippage *big.Int) (*big.Int, error) {
	return _UniswapFeeHandlerSeller.Contract.CalculateMinAmount(&_UniswapFeeHandlerSeller.CallOpts, midPriceNumerator, midPriceDenominator, amount, maxSlippage)
}

// GetRoutersForToken is a free data retrieval call binding the contract method 0x0c2fef14.
//
// Solidity: function getRoutersForToken(address token) view returns(address[])
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCaller) GetRoutersForToken(opts *bind.CallOpts, token common.Address) ([]common.Address, error) {
	var out []interface{}
	err := _UniswapFeeHandlerSeller.contract.Call(opts, &out, "getRoutersForToken", token)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetRoutersForToken is a free data retrieval call binding the contract method 0x0c2fef14.
//
// Solidity: function getRoutersForToken(address token) view returns(address[])
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) GetRoutersForToken(token common.Address) ([]common.Address, error) {
	return _UniswapFeeHandlerSeller.Contract.GetRoutersForToken(&_UniswapFeeHandlerSeller.CallOpts, token)
}

// GetRoutersForToken is a free data retrieval call binding the contract method 0x0c2fef14.
//
// Solidity: function getRoutersForToken(address token) view returns(address[])
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCallerSession) GetRoutersForToken(token common.Address) ([]common.Address, error) {
	return _UniswapFeeHandlerSeller.Contract.GetRoutersForToken(&_UniswapFeeHandlerSeller.CallOpts, token)
}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCaller) GetVersionNumber(opts *bind.CallOpts) (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	var out []interface{}
	err := _UniswapFeeHandlerSeller.contract.Call(opts, &out, "getVersionNumber")

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
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) GetVersionNumber() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _UniswapFeeHandlerSeller.Contract.GetVersionNumber(&_UniswapFeeHandlerSeller.CallOpts)
}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCallerSession) GetVersionNumber() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _UniswapFeeHandlerSeller.Contract.GetVersionNumber(&_UniswapFeeHandlerSeller.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCaller) Initialized(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _UniswapFeeHandlerSeller.contract.Call(opts, &out, "initialized")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) Initialized() (bool, error) {
	return _UniswapFeeHandlerSeller.Contract.Initialized(&_UniswapFeeHandlerSeller.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCallerSession) Initialized() (bool, error) {
	return _UniswapFeeHandlerSeller.Contract.Initialized(&_UniswapFeeHandlerSeller.CallOpts)
}

// MinimumReports is a free data retrieval call binding the contract method 0x4e008cdb.
//
// Solidity: function minimumReports(address ) view returns(uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCaller) MinimumReports(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _UniswapFeeHandlerSeller.contract.Call(opts, &out, "minimumReports", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinimumReports is a free data retrieval call binding the contract method 0x4e008cdb.
//
// Solidity: function minimumReports(address ) view returns(uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) MinimumReports(arg0 common.Address) (*big.Int, error) {
	return _UniswapFeeHandlerSeller.Contract.MinimumReports(&_UniswapFeeHandlerSeller.CallOpts, arg0)
}

// MinimumReports is a free data retrieval call binding the contract method 0x4e008cdb.
//
// Solidity: function minimumReports(address ) view returns(uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCallerSession) MinimumReports(arg0 common.Address) (*big.Int, error) {
	return _UniswapFeeHandlerSeller.Contract.MinimumReports(&_UniswapFeeHandlerSeller.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _UniswapFeeHandlerSeller.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) Owner() (common.Address, error) {
	return _UniswapFeeHandlerSeller.Contract.Owner(&_UniswapFeeHandlerSeller.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCallerSession) Owner() (common.Address, error) {
	return _UniswapFeeHandlerSeller.Contract.Owner(&_UniswapFeeHandlerSeller.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCaller) Registry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _UniswapFeeHandlerSeller.contract.Call(opts, &out, "registry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) Registry() (common.Address, error) {
	return _UniswapFeeHandlerSeller.Contract.Registry(&_UniswapFeeHandlerSeller.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerCallerSession) Registry() (common.Address, error) {
	return _UniswapFeeHandlerSeller.Contract.Registry(&_UniswapFeeHandlerSeller.CallOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0xff1d5752.
//
// Solidity: function initialize(address _registryAddress, address[] tokenAddresses, uint256[] newMininumReports) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactor) Initialize(opts *bind.TransactOpts, _registryAddress common.Address, tokenAddresses []common.Address, newMininumReports []*big.Int) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.contract.Transact(opts, "initialize", _registryAddress, tokenAddresses, newMininumReports)
}

// Initialize is a paid mutator transaction binding the contract method 0xff1d5752.
//
// Solidity: function initialize(address _registryAddress, address[] tokenAddresses, uint256[] newMininumReports) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) Initialize(_registryAddress common.Address, tokenAddresses []common.Address, newMininumReports []*big.Int) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.Initialize(&_UniswapFeeHandlerSeller.TransactOpts, _registryAddress, tokenAddresses, newMininumReports)
}

// Initialize is a paid mutator transaction binding the contract method 0xff1d5752.
//
// Solidity: function initialize(address _registryAddress, address[] tokenAddresses, uint256[] newMininumReports) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorSession) Initialize(_registryAddress common.Address, tokenAddresses []common.Address, newMininumReports []*big.Int) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.Initialize(&_UniswapFeeHandlerSeller.TransactOpts, _registryAddress, tokenAddresses, newMininumReports)
}

// RemoveRouter is a paid mutator transaction binding the contract method 0x73f0a070.
//
// Solidity: function removeRouter(address token, address router) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactor) RemoveRouter(opts *bind.TransactOpts, token common.Address, router common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.contract.Transact(opts, "removeRouter", token, router)
}

// RemoveRouter is a paid mutator transaction binding the contract method 0x73f0a070.
//
// Solidity: function removeRouter(address token, address router) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) RemoveRouter(token common.Address, router common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.RemoveRouter(&_UniswapFeeHandlerSeller.TransactOpts, token, router)
}

// RemoveRouter is a paid mutator transaction binding the contract method 0x73f0a070.
//
// Solidity: function removeRouter(address token, address router) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorSession) RemoveRouter(token common.Address, router common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.RemoveRouter(&_UniswapFeeHandlerSeller.TransactOpts, token, router)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) RenounceOwnership() (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.RenounceOwnership(&_UniswapFeeHandlerSeller.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.RenounceOwnership(&_UniswapFeeHandlerSeller.TransactOpts)
}

// Sell is a paid mutator transaction binding the contract method 0x31de7d15.
//
// Solidity: function sell(address sellTokenAddress, address buyTokenAddress, uint256 amount, uint256 maxSlippage) returns(uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactor) Sell(opts *bind.TransactOpts, sellTokenAddress common.Address, buyTokenAddress common.Address, amount *big.Int, maxSlippage *big.Int) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.contract.Transact(opts, "sell", sellTokenAddress, buyTokenAddress, amount, maxSlippage)
}

// Sell is a paid mutator transaction binding the contract method 0x31de7d15.
//
// Solidity: function sell(address sellTokenAddress, address buyTokenAddress, uint256 amount, uint256 maxSlippage) returns(uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) Sell(sellTokenAddress common.Address, buyTokenAddress common.Address, amount *big.Int, maxSlippage *big.Int) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.Sell(&_UniswapFeeHandlerSeller.TransactOpts, sellTokenAddress, buyTokenAddress, amount, maxSlippage)
}

// Sell is a paid mutator transaction binding the contract method 0x31de7d15.
//
// Solidity: function sell(address sellTokenAddress, address buyTokenAddress, uint256 amount, uint256 maxSlippage) returns(uint256)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorSession) Sell(sellTokenAddress common.Address, buyTokenAddress common.Address, amount *big.Int, maxSlippage *big.Int) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.Sell(&_UniswapFeeHandlerSeller.TransactOpts, sellTokenAddress, buyTokenAddress, amount, maxSlippage)
}

// SetMinimumReports is a paid mutator transaction binding the contract method 0x2f257aa0.
//
// Solidity: function setMinimumReports(address tokenAddress, uint256 newMininumReports) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactor) SetMinimumReports(opts *bind.TransactOpts, tokenAddress common.Address, newMininumReports *big.Int) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.contract.Transact(opts, "setMinimumReports", tokenAddress, newMininumReports)
}

// SetMinimumReports is a paid mutator transaction binding the contract method 0x2f257aa0.
//
// Solidity: function setMinimumReports(address tokenAddress, uint256 newMininumReports) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) SetMinimumReports(tokenAddress common.Address, newMininumReports *big.Int) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.SetMinimumReports(&_UniswapFeeHandlerSeller.TransactOpts, tokenAddress, newMininumReports)
}

// SetMinimumReports is a paid mutator transaction binding the contract method 0x2f257aa0.
//
// Solidity: function setMinimumReports(address tokenAddress, uint256 newMininumReports) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorSession) SetMinimumReports(tokenAddress common.Address, newMininumReports *big.Int) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.SetMinimumReports(&_UniswapFeeHandlerSeller.TransactOpts, tokenAddress, newMininumReports)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registryAddress) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactor) SetRegistry(opts *bind.TransactOpts, registryAddress common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.contract.Transact(opts, "setRegistry", registryAddress)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registryAddress) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) SetRegistry(registryAddress common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.SetRegistry(&_UniswapFeeHandlerSeller.TransactOpts, registryAddress)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registryAddress) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorSession) SetRegistry(registryAddress common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.SetRegistry(&_UniswapFeeHandlerSeller.TransactOpts, registryAddress)
}

// SetRouter is a paid mutator transaction binding the contract method 0x41d68b8f.
//
// Solidity: function setRouter(address token, address router) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactor) SetRouter(opts *bind.TransactOpts, token common.Address, router common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.contract.Transact(opts, "setRouter", token, router)
}

// SetRouter is a paid mutator transaction binding the contract method 0x41d68b8f.
//
// Solidity: function setRouter(address token, address router) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) SetRouter(token common.Address, router common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.SetRouter(&_UniswapFeeHandlerSeller.TransactOpts, token, router)
}

// SetRouter is a paid mutator transaction binding the contract method 0x41d68b8f.
//
// Solidity: function setRouter(address token, address router) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorSession) SetRouter(token common.Address, router common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.SetRouter(&_UniswapFeeHandlerSeller.TransactOpts, token, router)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address token, uint256 amount, address to) returns(bool)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactor) Transfer(opts *bind.TransactOpts, token common.Address, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.contract.Transact(opts, "transfer", token, amount, to)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address token, uint256 amount, address to) returns(bool)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) Transfer(token common.Address, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.Transfer(&_UniswapFeeHandlerSeller.TransactOpts, token, amount, to)
}

// Transfer is a paid mutator transaction binding the contract method 0xdbba0f01.
//
// Solidity: function transfer(address token, uint256 amount, address to) returns(bool)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorSession) Transfer(token common.Address, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.Transfer(&_UniswapFeeHandlerSeller.TransactOpts, token, amount, to)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.TransferOwnership(&_UniswapFeeHandlerSeller.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.TransferOwnership(&_UniswapFeeHandlerSeller.TransactOpts, newOwner)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerSession) Receive() (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.Receive(&_UniswapFeeHandlerSeller.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerTransactorSession) Receive() (*types.Transaction, error) {
	return _UniswapFeeHandlerSeller.Contract.Receive(&_UniswapFeeHandlerSeller.TransactOpts)
}

// UniswapFeeHandlerSellerMinimumReportsSetIterator is returned from FilterMinimumReportsSet and is used to iterate over the raw logs and unpacked data for MinimumReportsSet events raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerMinimumReportsSetIterator struct {
	Event *UniswapFeeHandlerSellerMinimumReportsSet // Event containing the contract specifics and raw log

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
func (it *UniswapFeeHandlerSellerMinimumReportsSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UniswapFeeHandlerSellerMinimumReportsSet)
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
		it.Event = new(UniswapFeeHandlerSellerMinimumReportsSet)
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
func (it *UniswapFeeHandlerSellerMinimumReportsSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UniswapFeeHandlerSellerMinimumReportsSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UniswapFeeHandlerSellerMinimumReportsSet represents a MinimumReportsSet event raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerMinimumReportsSet struct {
	TokenAddress   common.Address
	MinimumReports *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterMinimumReportsSet is a free log retrieval operation binding the contract event 0x03cc7dddcb89dd90027bd8fa62d09d1b5c49ce5d20f8c9bb6bdeaaa62ea1718b.
//
// Solidity: event MinimumReportsSet(address tokenAddress, uint256 minimumReports)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) FilterMinimumReportsSet(opts *bind.FilterOpts) (*UniswapFeeHandlerSellerMinimumReportsSetIterator, error) {

	logs, sub, err := _UniswapFeeHandlerSeller.contract.FilterLogs(opts, "MinimumReportsSet")
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerMinimumReportsSetIterator{contract: _UniswapFeeHandlerSeller.contract, event: "MinimumReportsSet", logs: logs, sub: sub}, nil
}

// WatchMinimumReportsSet is a free log subscription operation binding the contract event 0x03cc7dddcb89dd90027bd8fa62d09d1b5c49ce5d20f8c9bb6bdeaaa62ea1718b.
//
// Solidity: event MinimumReportsSet(address tokenAddress, uint256 minimumReports)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) WatchMinimumReportsSet(opts *bind.WatchOpts, sink chan<- *UniswapFeeHandlerSellerMinimumReportsSet) (event.Subscription, error) {

	logs, sub, err := _UniswapFeeHandlerSeller.contract.WatchLogs(opts, "MinimumReportsSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UniswapFeeHandlerSellerMinimumReportsSet)
				if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "MinimumReportsSet", log); err != nil {
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
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) ParseMinimumReportsSet(log types.Log) (*UniswapFeeHandlerSellerMinimumReportsSet, error) {
	event := new(UniswapFeeHandlerSellerMinimumReportsSet)
	if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "MinimumReportsSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// UniswapFeeHandlerSellerOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerOwnershipTransferredIterator struct {
	Event *UniswapFeeHandlerSellerOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *UniswapFeeHandlerSellerOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UniswapFeeHandlerSellerOwnershipTransferred)
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
		it.Event = new(UniswapFeeHandlerSellerOwnershipTransferred)
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
func (it *UniswapFeeHandlerSellerOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UniswapFeeHandlerSellerOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UniswapFeeHandlerSellerOwnershipTransferred represents a OwnershipTransferred event raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*UniswapFeeHandlerSellerOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _UniswapFeeHandlerSeller.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerOwnershipTransferredIterator{contract: _UniswapFeeHandlerSeller.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *UniswapFeeHandlerSellerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _UniswapFeeHandlerSeller.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UniswapFeeHandlerSellerOwnershipTransferred)
				if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) ParseOwnershipTransferred(log types.Log) (*UniswapFeeHandlerSellerOwnershipTransferred, error) {
	event := new(UniswapFeeHandlerSellerOwnershipTransferred)
	if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// UniswapFeeHandlerSellerReceivedQuoteIterator is returned from FilterReceivedQuote and is used to iterate over the raw logs and unpacked data for ReceivedQuote events raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerReceivedQuoteIterator struct {
	Event *UniswapFeeHandlerSellerReceivedQuote // Event containing the contract specifics and raw log

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
func (it *UniswapFeeHandlerSellerReceivedQuoteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UniswapFeeHandlerSellerReceivedQuote)
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
		it.Event = new(UniswapFeeHandlerSellerReceivedQuote)
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
func (it *UniswapFeeHandlerSellerReceivedQuoteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UniswapFeeHandlerSellerReceivedQuoteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UniswapFeeHandlerSellerReceivedQuote represents a ReceivedQuote event raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerReceivedQuote struct {
	TokneAddress common.Address
	Router       common.Address
	Quote        *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterReceivedQuote is a free log retrieval operation binding the contract event 0xba55c28acee19777ec6c603117b386d3e3b39886c0d3d53bc244be24ee6e7c84.
//
// Solidity: event ReceivedQuote(address indexed tokneAddress, address indexed router, uint256 quote)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) FilterReceivedQuote(opts *bind.FilterOpts, tokneAddress []common.Address, router []common.Address) (*UniswapFeeHandlerSellerReceivedQuoteIterator, error) {

	var tokneAddressRule []interface{}
	for _, tokneAddressItem := range tokneAddress {
		tokneAddressRule = append(tokneAddressRule, tokneAddressItem)
	}
	var routerRule []interface{}
	for _, routerItem := range router {
		routerRule = append(routerRule, routerItem)
	}

	logs, sub, err := _UniswapFeeHandlerSeller.contract.FilterLogs(opts, "ReceivedQuote", tokneAddressRule, routerRule)
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerReceivedQuoteIterator{contract: _UniswapFeeHandlerSeller.contract, event: "ReceivedQuote", logs: logs, sub: sub}, nil
}

// WatchReceivedQuote is a free log subscription operation binding the contract event 0xba55c28acee19777ec6c603117b386d3e3b39886c0d3d53bc244be24ee6e7c84.
//
// Solidity: event ReceivedQuote(address indexed tokneAddress, address indexed router, uint256 quote)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) WatchReceivedQuote(opts *bind.WatchOpts, sink chan<- *UniswapFeeHandlerSellerReceivedQuote, tokneAddress []common.Address, router []common.Address) (event.Subscription, error) {

	var tokneAddressRule []interface{}
	for _, tokneAddressItem := range tokneAddress {
		tokneAddressRule = append(tokneAddressRule, tokneAddressItem)
	}
	var routerRule []interface{}
	for _, routerItem := range router {
		routerRule = append(routerRule, routerItem)
	}

	logs, sub, err := _UniswapFeeHandlerSeller.contract.WatchLogs(opts, "ReceivedQuote", tokneAddressRule, routerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UniswapFeeHandlerSellerReceivedQuote)
				if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "ReceivedQuote", log); err != nil {
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

// ParseReceivedQuote is a log parse operation binding the contract event 0xba55c28acee19777ec6c603117b386d3e3b39886c0d3d53bc244be24ee6e7c84.
//
// Solidity: event ReceivedQuote(address indexed tokneAddress, address indexed router, uint256 quote)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) ParseReceivedQuote(log types.Log) (*UniswapFeeHandlerSellerReceivedQuote, error) {
	event := new(UniswapFeeHandlerSellerReceivedQuote)
	if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "ReceivedQuote", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// UniswapFeeHandlerSellerRegistrySetIterator is returned from FilterRegistrySet and is used to iterate over the raw logs and unpacked data for RegistrySet events raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerRegistrySetIterator struct {
	Event *UniswapFeeHandlerSellerRegistrySet // Event containing the contract specifics and raw log

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
func (it *UniswapFeeHandlerSellerRegistrySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UniswapFeeHandlerSellerRegistrySet)
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
		it.Event = new(UniswapFeeHandlerSellerRegistrySet)
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
func (it *UniswapFeeHandlerSellerRegistrySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UniswapFeeHandlerSellerRegistrySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UniswapFeeHandlerSellerRegistrySet represents a RegistrySet event raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerRegistrySet struct {
	RegistryAddress common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterRegistrySet is a free log retrieval operation binding the contract event 0x27fe5f0c1c3b1ed427cc63d0f05759ffdecf9aec9e18d31ef366fc8a6cb5dc3b.
//
// Solidity: event RegistrySet(address indexed registryAddress)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) FilterRegistrySet(opts *bind.FilterOpts, registryAddress []common.Address) (*UniswapFeeHandlerSellerRegistrySetIterator, error) {

	var registryAddressRule []interface{}
	for _, registryAddressItem := range registryAddress {
		registryAddressRule = append(registryAddressRule, registryAddressItem)
	}

	logs, sub, err := _UniswapFeeHandlerSeller.contract.FilterLogs(opts, "RegistrySet", registryAddressRule)
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerRegistrySetIterator{contract: _UniswapFeeHandlerSeller.contract, event: "RegistrySet", logs: logs, sub: sub}, nil
}

// WatchRegistrySet is a free log subscription operation binding the contract event 0x27fe5f0c1c3b1ed427cc63d0f05759ffdecf9aec9e18d31ef366fc8a6cb5dc3b.
//
// Solidity: event RegistrySet(address indexed registryAddress)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) WatchRegistrySet(opts *bind.WatchOpts, sink chan<- *UniswapFeeHandlerSellerRegistrySet, registryAddress []common.Address) (event.Subscription, error) {

	var registryAddressRule []interface{}
	for _, registryAddressItem := range registryAddress {
		registryAddressRule = append(registryAddressRule, registryAddressItem)
	}

	logs, sub, err := _UniswapFeeHandlerSeller.contract.WatchLogs(opts, "RegistrySet", registryAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UniswapFeeHandlerSellerRegistrySet)
				if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "RegistrySet", log); err != nil {
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
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) ParseRegistrySet(log types.Log) (*UniswapFeeHandlerSellerRegistrySet, error) {
	event := new(UniswapFeeHandlerSellerRegistrySet)
	if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "RegistrySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// UniswapFeeHandlerSellerRouterAddressRemovedIterator is returned from FilterRouterAddressRemoved and is used to iterate over the raw logs and unpacked data for RouterAddressRemoved events raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerRouterAddressRemovedIterator struct {
	Event *UniswapFeeHandlerSellerRouterAddressRemoved // Event containing the contract specifics and raw log

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
func (it *UniswapFeeHandlerSellerRouterAddressRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UniswapFeeHandlerSellerRouterAddressRemoved)
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
		it.Event = new(UniswapFeeHandlerSellerRouterAddressRemoved)
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
func (it *UniswapFeeHandlerSellerRouterAddressRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UniswapFeeHandlerSellerRouterAddressRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UniswapFeeHandlerSellerRouterAddressRemoved represents a RouterAddressRemoved event raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerRouterAddressRemoved struct {
	Token  common.Address
	Router common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRouterAddressRemoved is a free log retrieval operation binding the contract event 0x044c4b00bcc14b6c00430f73b8bc07f33aecb2387c7b188142d6d497342de89a.
//
// Solidity: event RouterAddressRemoved(address token, address router)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) FilterRouterAddressRemoved(opts *bind.FilterOpts) (*UniswapFeeHandlerSellerRouterAddressRemovedIterator, error) {

	logs, sub, err := _UniswapFeeHandlerSeller.contract.FilterLogs(opts, "RouterAddressRemoved")
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerRouterAddressRemovedIterator{contract: _UniswapFeeHandlerSeller.contract, event: "RouterAddressRemoved", logs: logs, sub: sub}, nil
}

// WatchRouterAddressRemoved is a free log subscription operation binding the contract event 0x044c4b00bcc14b6c00430f73b8bc07f33aecb2387c7b188142d6d497342de89a.
//
// Solidity: event RouterAddressRemoved(address token, address router)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) WatchRouterAddressRemoved(opts *bind.WatchOpts, sink chan<- *UniswapFeeHandlerSellerRouterAddressRemoved) (event.Subscription, error) {

	logs, sub, err := _UniswapFeeHandlerSeller.contract.WatchLogs(opts, "RouterAddressRemoved")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UniswapFeeHandlerSellerRouterAddressRemoved)
				if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "RouterAddressRemoved", log); err != nil {
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

// ParseRouterAddressRemoved is a log parse operation binding the contract event 0x044c4b00bcc14b6c00430f73b8bc07f33aecb2387c7b188142d6d497342de89a.
//
// Solidity: event RouterAddressRemoved(address token, address router)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) ParseRouterAddressRemoved(log types.Log) (*UniswapFeeHandlerSellerRouterAddressRemoved, error) {
	event := new(UniswapFeeHandlerSellerRouterAddressRemoved)
	if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "RouterAddressRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// UniswapFeeHandlerSellerRouterAddressSetIterator is returned from FilterRouterAddressSet and is used to iterate over the raw logs and unpacked data for RouterAddressSet events raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerRouterAddressSetIterator struct {
	Event *UniswapFeeHandlerSellerRouterAddressSet // Event containing the contract specifics and raw log

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
func (it *UniswapFeeHandlerSellerRouterAddressSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UniswapFeeHandlerSellerRouterAddressSet)
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
		it.Event = new(UniswapFeeHandlerSellerRouterAddressSet)
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
func (it *UniswapFeeHandlerSellerRouterAddressSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UniswapFeeHandlerSellerRouterAddressSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UniswapFeeHandlerSellerRouterAddressSet represents a RouterAddressSet event raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerRouterAddressSet struct {
	Token  common.Address
	Router common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRouterAddressSet is a free log retrieval operation binding the contract event 0xb3cbb74e835466bdbf8838b1acb70fa4a8b73e1a00cd5bacb9f68cf4dfc79cf3.
//
// Solidity: event RouterAddressSet(address token, address router)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) FilterRouterAddressSet(opts *bind.FilterOpts) (*UniswapFeeHandlerSellerRouterAddressSetIterator, error) {

	logs, sub, err := _UniswapFeeHandlerSeller.contract.FilterLogs(opts, "RouterAddressSet")
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerRouterAddressSetIterator{contract: _UniswapFeeHandlerSeller.contract, event: "RouterAddressSet", logs: logs, sub: sub}, nil
}

// WatchRouterAddressSet is a free log subscription operation binding the contract event 0xb3cbb74e835466bdbf8838b1acb70fa4a8b73e1a00cd5bacb9f68cf4dfc79cf3.
//
// Solidity: event RouterAddressSet(address token, address router)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) WatchRouterAddressSet(opts *bind.WatchOpts, sink chan<- *UniswapFeeHandlerSellerRouterAddressSet) (event.Subscription, error) {

	logs, sub, err := _UniswapFeeHandlerSeller.contract.WatchLogs(opts, "RouterAddressSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UniswapFeeHandlerSellerRouterAddressSet)
				if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "RouterAddressSet", log); err != nil {
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

// ParseRouterAddressSet is a log parse operation binding the contract event 0xb3cbb74e835466bdbf8838b1acb70fa4a8b73e1a00cd5bacb9f68cf4dfc79cf3.
//
// Solidity: event RouterAddressSet(address token, address router)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) ParseRouterAddressSet(log types.Log) (*UniswapFeeHandlerSellerRouterAddressSet, error) {
	event := new(UniswapFeeHandlerSellerRouterAddressSet)
	if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "RouterAddressSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// UniswapFeeHandlerSellerRouterUsedIterator is returned from FilterRouterUsed and is used to iterate over the raw logs and unpacked data for RouterUsed events raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerRouterUsedIterator struct {
	Event *UniswapFeeHandlerSellerRouterUsed // Event containing the contract specifics and raw log

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
func (it *UniswapFeeHandlerSellerRouterUsedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UniswapFeeHandlerSellerRouterUsed)
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
		it.Event = new(UniswapFeeHandlerSellerRouterUsed)
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
func (it *UniswapFeeHandlerSellerRouterUsedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UniswapFeeHandlerSellerRouterUsedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UniswapFeeHandlerSellerRouterUsed represents a RouterUsed event raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerRouterUsed struct {
	Router common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRouterUsed is a free log retrieval operation binding the contract event 0x59afd9b3bf745a06d8739721397432b9f161243cee13868b9d6d6fca05b6e551.
//
// Solidity: event RouterUsed(address router)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) FilterRouterUsed(opts *bind.FilterOpts) (*UniswapFeeHandlerSellerRouterUsedIterator, error) {

	logs, sub, err := _UniswapFeeHandlerSeller.contract.FilterLogs(opts, "RouterUsed")
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerRouterUsedIterator{contract: _UniswapFeeHandlerSeller.contract, event: "RouterUsed", logs: logs, sub: sub}, nil
}

// WatchRouterUsed is a free log subscription operation binding the contract event 0x59afd9b3bf745a06d8739721397432b9f161243cee13868b9d6d6fca05b6e551.
//
// Solidity: event RouterUsed(address router)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) WatchRouterUsed(opts *bind.WatchOpts, sink chan<- *UniswapFeeHandlerSellerRouterUsed) (event.Subscription, error) {

	logs, sub, err := _UniswapFeeHandlerSeller.contract.WatchLogs(opts, "RouterUsed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UniswapFeeHandlerSellerRouterUsed)
				if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "RouterUsed", log); err != nil {
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

// ParseRouterUsed is a log parse operation binding the contract event 0x59afd9b3bf745a06d8739721397432b9f161243cee13868b9d6d6fca05b6e551.
//
// Solidity: event RouterUsed(address router)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) ParseRouterUsed(log types.Log) (*UniswapFeeHandlerSellerRouterUsed, error) {
	event := new(UniswapFeeHandlerSellerRouterUsed)
	if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "RouterUsed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// UniswapFeeHandlerSellerTokenSoldIterator is returned from FilterTokenSold and is used to iterate over the raw logs and unpacked data for TokenSold events raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerTokenSoldIterator struct {
	Event *UniswapFeeHandlerSellerTokenSold // Event containing the contract specifics and raw log

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
func (it *UniswapFeeHandlerSellerTokenSoldIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UniswapFeeHandlerSellerTokenSold)
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
		it.Event = new(UniswapFeeHandlerSellerTokenSold)
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
func (it *UniswapFeeHandlerSellerTokenSoldIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UniswapFeeHandlerSellerTokenSoldIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UniswapFeeHandlerSellerTokenSold represents a TokenSold event raised by the UniswapFeeHandlerSeller contract.
type UniswapFeeHandlerSellerTokenSold struct {
	SoldTokenAddress   common.Address
	BoughtTokenAddress common.Address
	Amount             *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterTokenSold is a free log retrieval operation binding the contract event 0xd4cffd6979677853b45a7a17f455188a434e975ba474c5a2613c94beacea537a.
//
// Solidity: event TokenSold(address soldTokenAddress, address boughtTokenAddress, uint256 amount)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) FilterTokenSold(opts *bind.FilterOpts) (*UniswapFeeHandlerSellerTokenSoldIterator, error) {

	logs, sub, err := _UniswapFeeHandlerSeller.contract.FilterLogs(opts, "TokenSold")
	if err != nil {
		return nil, err
	}
	return &UniswapFeeHandlerSellerTokenSoldIterator{contract: _UniswapFeeHandlerSeller.contract, event: "TokenSold", logs: logs, sub: sub}, nil
}

// WatchTokenSold is a free log subscription operation binding the contract event 0xd4cffd6979677853b45a7a17f455188a434e975ba474c5a2613c94beacea537a.
//
// Solidity: event TokenSold(address soldTokenAddress, address boughtTokenAddress, uint256 amount)
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) WatchTokenSold(opts *bind.WatchOpts, sink chan<- *UniswapFeeHandlerSellerTokenSold) (event.Subscription, error) {

	logs, sub, err := _UniswapFeeHandlerSeller.contract.WatchLogs(opts, "TokenSold")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UniswapFeeHandlerSellerTokenSold)
				if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "TokenSold", log); err != nil {
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
func (_UniswapFeeHandlerSeller *UniswapFeeHandlerSellerFilterer) ParseTokenSold(log types.Log) (*UniswapFeeHandlerSellerTokenSold, error) {
	event := new(UniswapFeeHandlerSellerTokenSold)
	if err := _UniswapFeeHandlerSeller.contract.UnpackLog(event, "TokenSold", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
