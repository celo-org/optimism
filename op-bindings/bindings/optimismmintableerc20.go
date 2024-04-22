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

// OptimismMintableERC20MetaData contains all meta data concerning the OptimismMintableERC20 contract.
var OptimismMintableERC20MetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_bridge\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_remoteToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_symbol\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"BRIDGE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"REMOTE_TOKEN\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"allowance\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"approve\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"balanceOf\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"bridge\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"burn\",\"inputs\":[{\"name\":\"_from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"creditGasFees\",\"inputs\":[{\"name\":\"recipients\",\"type\":\"address[]\",\"internalType\":\"address[]\"},{\"name\":\"amounts\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"creditGasFees\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeRecipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"communityFund\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"refund\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"tipTxFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"baseTxFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"debitGasFees\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"decimals\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"decreaseAllowance\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"subtractedValue\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"increaseAllowance\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"addedValue\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"l1Token\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"l2Bridge\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"mint\",\"inputs\":[{\"name\":\"_to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"name\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"remoteToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"_interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"symbol\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalSupply\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transfer\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferFrom\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"version\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"Approval\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"spender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Burn\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Mint\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Transfer\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false}]",
	Bin: "0x60e06040523480156200001157600080fd5b5060405162001be438038062001be4833981016040819052620000349162000163565b828260036200004483826200029e565b5060046200005382826200029e565b5050506001600160a01b039384166080529390921660a052505060ff1660c0526200036a565b80516001600160a01b03811681146200009157600080fd5b919050565b634e487b7160e01b600052604160045260246000fd5b600082601f830112620000be57600080fd5b81516001600160401b0380821115620000db57620000db62000096565b604051601f8301601f19908116603f0116810190828211818310171562000106576200010662000096565b816040528381526020925086838588010111156200012357600080fd5b600091505b8382101562000147578582018301518183018401529082019062000128565b83821115620001595760008385830101525b9695505050505050565b600080600080600060a086880312156200017c57600080fd5b620001878662000079565b9450620001976020870162000079565b60408701519094506001600160401b0380821115620001b557600080fd5b620001c389838a01620000ac565b94506060880151915080821115620001da57600080fd5b50620001e988828901620000ac565b925050608086015160ff811681146200020157600080fd5b809150509295509295909350565b600181811c908216806200022457607f821691505b6020821081036200024557634e487b7160e01b600052602260045260246000fd5b50919050565b601f8211156200029957600081815260208120601f850160051c81016020861015620002745750805b601f850160051c820191505b81811015620002955782815560010162000280565b5050505b505050565b81516001600160401b03811115620002ba57620002ba62000096565b620002d281620002cb84546200020f565b846200024b565b602080601f8311600181146200030a5760008415620002f15750858301515b600019600386901b1c1916600185901b17855562000295565b600085815260208120601f198616915b828110156200033b578886015182559484019460019091019084016200031a565b50858210156200035a5787850151600019600388901b60f8161c191681555b5050505050600190811b01905550565b60805160a05160c05161182e620003b6600039600061027a0152600081816103a50152818161043a015281816107e70152610a1f0152600081816101ca01526103cb015261182e6000f3fe608060405234801561001057600080fd5b50600436106101985760003560e01c80636a30b253116100e3578063ae1f6aaf1161008c578063dd62ed3e11610066578063dd62ed3e146103ef578063e78cea92146103a3578063ee9a31a21461043557600080fd5b8063ae1f6aaf146103a3578063c01e1bd6146103c9578063d6c0b2c4146103c957600080fd5b80639dc29fac116100bd5780639dc29fac1461036a578063a457c2d71461037d578063a9059cbb1461039057600080fd5b80636a30b2531461031957806370a082311461032c57806395d89b411461036257600080fd5b80632e0f98ad1161014557806340c10f191161011f57806340c10f19146102b757806354fd4d50146102ca57806358cf96721461030657600080fd5b80632e0f98ad1461025e578063313ce5671461027357806339509351146102a457600080fd5b8063095ea7b311610176578063095ea7b31461022657806318160ddd1461023957806323b872dd1461024b57600080fd5b806301ffc9a71461019d578063033964be146101c557806306fdde0314610211575b600080fd5b6101b06101ab366004611440565b61045c565b60405190151581526020015b60405180910390f35b6101ec7f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016101bc565b61021961054d565b6040516101bc9190611489565b6101b0610234366004611525565b6105df565b6002545b6040519081526020016101bc565b6101b061025936600461154f565b6105f7565b61027161026c3660046115d7565b61061b565b005b60405160ff7f00000000000000000000000000000000000000000000000000000000000000001681526020016101bc565b6101b06102b2366004611525565b610783565b6102716102c5366004611525565b6107cf565b6102196040518060400160405280600581526020017f312e332e3000000000000000000000000000000000000000000000000000000081525081565b610271610314366004611525565b6108f2565b610271610327366004611643565b610968565b61023d61033a3660046116bb565b73ffffffffffffffffffffffffffffffffffffffff1660009081526020819052604090205490565b6102196109f8565b610271610378366004611525565b610a07565b6101b061038b366004611525565b610b1e565b6101b061039e366004611525565b610bef565b7f00000000000000000000000000000000000000000000000000000000000000006101ec565b7f00000000000000000000000000000000000000000000000000000000000000006101ec565b61023d6103fd3660046116d6565b73ffffffffffffffffffffffffffffffffffffffff918216600090815260016020908152604080832093909416825291909152205490565b6101ec7f000000000000000000000000000000000000000000000000000000000000000081565b60007f01ffc9a7000000000000000000000000000000000000000000000000000000007f1d1d8b63000000000000000000000000000000000000000000000000000000007fec4fc8e3000000000000000000000000000000000000000000000000000000007fffffffff00000000000000000000000000000000000000000000000000000000851683148061051557507fffffffff00000000000000000000000000000000000000000000000000000000858116908316145b8061054457507fffffffff00000000000000000000000000000000000000000000000000000000858116908216145b95945050505050565b60606003805461055c90611709565b80601f016020809104026020016040519081016040528092919081815260200182805461058890611709565b80156105d55780601f106105aa576101008083540402835291602001916105d5565b820191906000526020600020905b8154815290600101906020018083116105b857829003601f168201915b5050505050905090565b6000336105ed818585610bfd565b5060019392505050565b600033610605858285610db1565b610610858585610e88565b506001949350505050565b3315610688576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c0000000000000000000000000000000060448201526064015b60405180910390fd5b828114610717576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602f60248201527f526563697069656e747320616e6420616d6f756e7473206d757374206265207460448201527f68652073616d65206c656e6774682e0000000000000000000000000000000000606482015260840161067f565b60005b8381101561077c5761076a8585838181106107375761073761175c565b905060200201602081019061074c91906116bb565b84848481811061075e5761075e61175c565b9050602002013561113b565b80610774816117ba565b91505061071a565b5050505050565b33600081815260016020908152604080832073ffffffffffffffffffffffffffffffffffffffff871684529091528120549091906105ed90829086906107ca9087906117f2565b610bfd565b3373ffffffffffffffffffffffffffffffffffffffff7f00000000000000000000000000000000000000000000000000000000000000001614610894576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603460248201527f4f7074696d69736d4d696e7461626c6545524332303a206f6e6c79206272696460448201527f67652063616e206d696e7420616e64206275726e000000000000000000000000606482015260840161067f565b61089e828261113b565b8173ffffffffffffffffffffffffffffffffffffffff167f0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885826040516108e691815260200190565b60405180910390a25050565b331561095a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c00000000000000000000000000000000604482015260640161067f565b610964828261125b565b5050565b33156109d0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c00000000000000000000000000000000604482015260640161067f565b6109da888561113b565b6109e4878461113b565b6109ee858261113b565b5050505050505050565b60606004805461055c90611709565b3373ffffffffffffffffffffffffffffffffffffffff7f00000000000000000000000000000000000000000000000000000000000000001614610acc576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603460248201527f4f7074696d69736d4d696e7461626c6545524332303a206f6e6c79206272696460448201527f67652063616e206d696e7420616e64206275726e000000000000000000000000606482015260840161067f565b610ad6828261125b565b8173ffffffffffffffffffffffffffffffffffffffff167fcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5826040516108e691815260200190565b33600081815260016020908152604080832073ffffffffffffffffffffffffffffffffffffffff8716845290915281205490919083811015610be2576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602560248201527f45524332303a2064656372656173656420616c6c6f77616e63652062656c6f7760448201527f207a65726f000000000000000000000000000000000000000000000000000000606482015260840161067f565b6106108286868403610bfd565b6000336105ed818585610e88565b73ffffffffffffffffffffffffffffffffffffffff8316610c9f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152602060048201526024808201527f45524332303a20617070726f76652066726f6d20746865207a65726f2061646460448201527f7265737300000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff8216610d42576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f45524332303a20617070726f766520746f20746865207a65726f20616464726560448201527f7373000000000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff83811660008181526001602090815260408083209487168084529482529182902085905590518481527f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92591015b60405180910390a3505050565b73ffffffffffffffffffffffffffffffffffffffff8381166000908152600160209081526040808320938616835292905220547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8114610e825781811015610e75576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f45524332303a20696e73756666696369656e7420616c6c6f77616e6365000000604482015260640161067f565b610e828484848403610bfd565b50505050565b73ffffffffffffffffffffffffffffffffffffffff8316610f2b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602560248201527f45524332303a207472616e736665722066726f6d20746865207a65726f20616460448201527f6472657373000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff8216610fce576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f45524332303a207472616e7366657220746f20746865207a65726f206164647260448201527f6573730000000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff831660009081526020819052604090205481811015611084576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f45524332303a207472616e7366657220616d6f756e742065786365656473206260448201527f616c616e63650000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff8085166000908152602081905260408082208585039055918516815290812080548492906110c89084906117f2565b925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef8460405161112e91815260200190565b60405180910390a3610e82565b73ffffffffffffffffffffffffffffffffffffffff82166111b8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f45524332303a206d696e7420746f20746865207a65726f206164647265737300604482015260640161067f565b80600260008282546111ca91906117f2565b909155505073ffffffffffffffffffffffffffffffffffffffff8216600090815260208190526040812080548392906112049084906117f2565b909155505060405181815273ffffffffffffffffffffffffffffffffffffffff8316906000907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9060200160405180910390a35050565b73ffffffffffffffffffffffffffffffffffffffff82166112fe576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602160248201527f45524332303a206275726e2066726f6d20746865207a65726f2061646472657360448201527f7300000000000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff8216600090815260208190526040902054818110156113b4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f45524332303a206275726e20616d6f756e7420657863656564732062616c616e60448201527f6365000000000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff831660009081526020819052604081208383039055600280548492906113f090849061180a565b909155505060405182815260009073ffffffffffffffffffffffffffffffffffffffff8516907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef90602001610da4565b60006020828403121561145257600080fd5b81357fffffffff000000000000000000000000000000000000000000000000000000008116811461148257600080fd5b9392505050565b600060208083528351808285015260005b818110156114b65785810183015185820160400152820161149a565b818111156114c8576000604083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016929092016040019392505050565b803573ffffffffffffffffffffffffffffffffffffffff8116811461152057600080fd5b919050565b6000806040838503121561153857600080fd5b611541836114fc565b946020939093013593505050565b60008060006060848603121561156457600080fd5b61156d846114fc565b925061157b602085016114fc565b9150604084013590509250925092565b60008083601f84011261159d57600080fd5b50813567ffffffffffffffff8111156115b557600080fd5b6020830191508360208260051b85010111156115d057600080fd5b9250929050565b600080600080604085870312156115ed57600080fd5b843567ffffffffffffffff8082111561160557600080fd5b6116118883890161158b565b9096509450602087013591508082111561162a57600080fd5b506116378782880161158b565b95989497509550505050565b600080600080600080600080610100898b03121561166057600080fd5b611669896114fc565b975061167760208a016114fc565b965061168560408a016114fc565b955061169360608a016114fc565b979a969950949760808101359660a0820135965060c0820135955060e0909101359350915050565b6000602082840312156116cd57600080fd5b611482826114fc565b600080604083850312156116e957600080fd5b6116f2836114fc565b9150611700602084016114fc565b90509250929050565b600181811c9082168061171d57607f821691505b602082108103611756577f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82036117eb576117eb61178b565b5060010190565b600082198211156118055761180561178b565b500190565b60008282101561181c5761181c61178b565b50039056fea164736f6c634300080f000a",
}

// OptimismMintableERC20ABI is the input ABI used to generate the binding from.
// Deprecated: Use OptimismMintableERC20MetaData.ABI instead.
var OptimismMintableERC20ABI = OptimismMintableERC20MetaData.ABI

// OptimismMintableERC20Bin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use OptimismMintableERC20MetaData.Bin instead.
var OptimismMintableERC20Bin = OptimismMintableERC20MetaData.Bin

// DeployOptimismMintableERC20 deploys a new Ethereum contract, binding an instance of OptimismMintableERC20 to it.
func DeployOptimismMintableERC20(auth *bind.TransactOpts, backend bind.ContractBackend, _bridge common.Address, _remoteToken common.Address, _name string, _symbol string, _decimals uint8) (common.Address, *types.Transaction, *OptimismMintableERC20, error) {
	parsed, err := OptimismMintableERC20MetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(OptimismMintableERC20Bin), backend, _bridge, _remoteToken, _name, _symbol, _decimals)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &OptimismMintableERC20{OptimismMintableERC20Caller: OptimismMintableERC20Caller{contract: contract}, OptimismMintableERC20Transactor: OptimismMintableERC20Transactor{contract: contract}, OptimismMintableERC20Filterer: OptimismMintableERC20Filterer{contract: contract}}, nil
}

// OptimismMintableERC20 is an auto generated Go binding around an Ethereum contract.
type OptimismMintableERC20 struct {
	OptimismMintableERC20Caller     // Read-only binding to the contract
	OptimismMintableERC20Transactor // Write-only binding to the contract
	OptimismMintableERC20Filterer   // Log filterer for contract events
}

// OptimismMintableERC20Caller is an auto generated read-only Go binding around an Ethereum contract.
type OptimismMintableERC20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OptimismMintableERC20Transactor is an auto generated write-only Go binding around an Ethereum contract.
type OptimismMintableERC20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OptimismMintableERC20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OptimismMintableERC20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OptimismMintableERC20Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OptimismMintableERC20Session struct {
	Contract     *OptimismMintableERC20 // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// OptimismMintableERC20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OptimismMintableERC20CallerSession struct {
	Contract *OptimismMintableERC20Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// OptimismMintableERC20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OptimismMintableERC20TransactorSession struct {
	Contract     *OptimismMintableERC20Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// OptimismMintableERC20Raw is an auto generated low-level Go binding around an Ethereum contract.
type OptimismMintableERC20Raw struct {
	Contract *OptimismMintableERC20 // Generic contract binding to access the raw methods on
}

// OptimismMintableERC20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OptimismMintableERC20CallerRaw struct {
	Contract *OptimismMintableERC20Caller // Generic read-only contract binding to access the raw methods on
}

// OptimismMintableERC20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OptimismMintableERC20TransactorRaw struct {
	Contract *OptimismMintableERC20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewOptimismMintableERC20 creates a new instance of OptimismMintableERC20, bound to a specific deployed contract.
func NewOptimismMintableERC20(address common.Address, backend bind.ContractBackend) (*OptimismMintableERC20, error) {
	contract, err := bindOptimismMintableERC20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OptimismMintableERC20{OptimismMintableERC20Caller: OptimismMintableERC20Caller{contract: contract}, OptimismMintableERC20Transactor: OptimismMintableERC20Transactor{contract: contract}, OptimismMintableERC20Filterer: OptimismMintableERC20Filterer{contract: contract}}, nil
}

// NewOptimismMintableERC20Caller creates a new read-only instance of OptimismMintableERC20, bound to a specific deployed contract.
func NewOptimismMintableERC20Caller(address common.Address, caller bind.ContractCaller) (*OptimismMintableERC20Caller, error) {
	contract, err := bindOptimismMintableERC20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OptimismMintableERC20Caller{contract: contract}, nil
}

// NewOptimismMintableERC20Transactor creates a new write-only instance of OptimismMintableERC20, bound to a specific deployed contract.
func NewOptimismMintableERC20Transactor(address common.Address, transactor bind.ContractTransactor) (*OptimismMintableERC20Transactor, error) {
	contract, err := bindOptimismMintableERC20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OptimismMintableERC20Transactor{contract: contract}, nil
}

// NewOptimismMintableERC20Filterer creates a new log filterer instance of OptimismMintableERC20, bound to a specific deployed contract.
func NewOptimismMintableERC20Filterer(address common.Address, filterer bind.ContractFilterer) (*OptimismMintableERC20Filterer, error) {
	contract, err := bindOptimismMintableERC20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OptimismMintableERC20Filterer{contract: contract}, nil
}

// bindOptimismMintableERC20 binds a generic wrapper to an already deployed contract.
func bindOptimismMintableERC20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OptimismMintableERC20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OptimismMintableERC20 *OptimismMintableERC20Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OptimismMintableERC20.Contract.OptimismMintableERC20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OptimismMintableERC20 *OptimismMintableERC20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.OptimismMintableERC20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OptimismMintableERC20 *OptimismMintableERC20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.OptimismMintableERC20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OptimismMintableERC20 *OptimismMintableERC20CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OptimismMintableERC20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.contract.Transact(opts, method, params...)
}

// BRIDGE is a free data retrieval call binding the contract method 0xee9a31a2.
//
// Solidity: function BRIDGE() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) BRIDGE(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "BRIDGE")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BRIDGE is a free data retrieval call binding the contract method 0xee9a31a2.
//
// Solidity: function BRIDGE() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) BRIDGE() (common.Address, error) {
	return _OptimismMintableERC20.Contract.BRIDGE(&_OptimismMintableERC20.CallOpts)
}

// BRIDGE is a free data retrieval call binding the contract method 0xee9a31a2.
//
// Solidity: function BRIDGE() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) BRIDGE() (common.Address, error) {
	return _OptimismMintableERC20.Contract.BRIDGE(&_OptimismMintableERC20.CallOpts)
}

// REMOTETOKEN is a free data retrieval call binding the contract method 0x033964be.
//
// Solidity: function REMOTE_TOKEN() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) REMOTETOKEN(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "REMOTE_TOKEN")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// REMOTETOKEN is a free data retrieval call binding the contract method 0x033964be.
//
// Solidity: function REMOTE_TOKEN() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) REMOTETOKEN() (common.Address, error) {
	return _OptimismMintableERC20.Contract.REMOTETOKEN(&_OptimismMintableERC20.CallOpts)
}

// REMOTETOKEN is a free data retrieval call binding the contract method 0x033964be.
//
// Solidity: function REMOTE_TOKEN() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) REMOTETOKEN() (common.Address, error) {
	return _OptimismMintableERC20.Contract.REMOTETOKEN(&_OptimismMintableERC20.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _OptimismMintableERC20.Contract.Allowance(&_OptimismMintableERC20.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _OptimismMintableERC20.Contract.Allowance(&_OptimismMintableERC20.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) BalanceOf(account common.Address) (*big.Int, error) {
	return _OptimismMintableERC20.Contract.BalanceOf(&_OptimismMintableERC20.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _OptimismMintableERC20.Contract.BalanceOf(&_OptimismMintableERC20.CallOpts, account)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) Bridge() (common.Address, error) {
	return _OptimismMintableERC20.Contract.Bridge(&_OptimismMintableERC20.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) Bridge() (common.Address, error) {
	return _OptimismMintableERC20.Contract.Bridge(&_OptimismMintableERC20.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) Decimals() (uint8, error) {
	return _OptimismMintableERC20.Contract.Decimals(&_OptimismMintableERC20.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) Decimals() (uint8, error) {
	return _OptimismMintableERC20.Contract.Decimals(&_OptimismMintableERC20.CallOpts)
}

// L1Token is a free data retrieval call binding the contract method 0xc01e1bd6.
//
// Solidity: function l1Token() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) L1Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "l1Token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L1Token is a free data retrieval call binding the contract method 0xc01e1bd6.
//
// Solidity: function l1Token() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) L1Token() (common.Address, error) {
	return _OptimismMintableERC20.Contract.L1Token(&_OptimismMintableERC20.CallOpts)
}

// L1Token is a free data retrieval call binding the contract method 0xc01e1bd6.
//
// Solidity: function l1Token() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) L1Token() (common.Address, error) {
	return _OptimismMintableERC20.Contract.L1Token(&_OptimismMintableERC20.CallOpts)
}

// L2Bridge is a free data retrieval call binding the contract method 0xae1f6aaf.
//
// Solidity: function l2Bridge() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) L2Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "l2Bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L2Bridge is a free data retrieval call binding the contract method 0xae1f6aaf.
//
// Solidity: function l2Bridge() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) L2Bridge() (common.Address, error) {
	return _OptimismMintableERC20.Contract.L2Bridge(&_OptimismMintableERC20.CallOpts)
}

// L2Bridge is a free data retrieval call binding the contract method 0xae1f6aaf.
//
// Solidity: function l2Bridge() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) L2Bridge() (common.Address, error) {
	return _OptimismMintableERC20.Contract.L2Bridge(&_OptimismMintableERC20.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) Name() (string, error) {
	return _OptimismMintableERC20.Contract.Name(&_OptimismMintableERC20.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) Name() (string, error) {
	return _OptimismMintableERC20.Contract.Name(&_OptimismMintableERC20.CallOpts)
}

// RemoteToken is a free data retrieval call binding the contract method 0xd6c0b2c4.
//
// Solidity: function remoteToken() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) RemoteToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "remoteToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RemoteToken is a free data retrieval call binding the contract method 0xd6c0b2c4.
//
// Solidity: function remoteToken() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) RemoteToken() (common.Address, error) {
	return _OptimismMintableERC20.Contract.RemoteToken(&_OptimismMintableERC20.CallOpts)
}

// RemoteToken is a free data retrieval call binding the contract method 0xd6c0b2c4.
//
// Solidity: function remoteToken() view returns(address)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) RemoteToken() (common.Address, error) {
	return _OptimismMintableERC20.Contract.RemoteToken(&_OptimismMintableERC20.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 _interfaceId) pure returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) SupportsInterface(opts *bind.CallOpts, _interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "supportsInterface", _interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 _interfaceId) pure returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) SupportsInterface(_interfaceId [4]byte) (bool, error) {
	return _OptimismMintableERC20.Contract.SupportsInterface(&_OptimismMintableERC20.CallOpts, _interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 _interfaceId) pure returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) SupportsInterface(_interfaceId [4]byte) (bool, error) {
	return _OptimismMintableERC20.Contract.SupportsInterface(&_OptimismMintableERC20.CallOpts, _interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) Symbol() (string, error) {
	return _OptimismMintableERC20.Contract.Symbol(&_OptimismMintableERC20.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) Symbol() (string, error) {
	return _OptimismMintableERC20.Contract.Symbol(&_OptimismMintableERC20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) TotalSupply() (*big.Int, error) {
	return _OptimismMintableERC20.Contract.TotalSupply(&_OptimismMintableERC20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) TotalSupply() (*big.Int, error) {
	return _OptimismMintableERC20.Contract.TotalSupply(&_OptimismMintableERC20.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_OptimismMintableERC20 *OptimismMintableERC20Caller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _OptimismMintableERC20.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) Version() (string, error) {
	return _OptimismMintableERC20.Contract.Version(&_OptimismMintableERC20.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_OptimismMintableERC20 *OptimismMintableERC20CallerSession) Version() (string, error) {
	return _OptimismMintableERC20.Contract.Version(&_OptimismMintableERC20.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Transactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.Approve(&_OptimismMintableERC20.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.Approve(&_OptimismMintableERC20.TransactOpts, spender, amount)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address _from, uint256 _amount) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20Transactor) Burn(opts *bind.TransactOpts, _from common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.contract.Transact(opts, "burn", _from, _amount)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address _from, uint256 _amount) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20Session) Burn(_from common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.Burn(&_OptimismMintableERC20.TransactOpts, _from, _amount)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address _from, uint256 _amount) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorSession) Burn(_from common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.Burn(&_OptimismMintableERC20.TransactOpts, _from, _amount)
}

// CreditGasFees is a paid mutator transaction binding the contract method 0x2e0f98ad.
//
// Solidity: function creditGasFees(address[] recipients, uint256[] amounts) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20Transactor) CreditGasFees(opts *bind.TransactOpts, recipients []common.Address, amounts []*big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.contract.Transact(opts, "creditGasFees", recipients, amounts)
}

// CreditGasFees is a paid mutator transaction binding the contract method 0x2e0f98ad.
//
// Solidity: function creditGasFees(address[] recipients, uint256[] amounts) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20Session) CreditGasFees(recipients []common.Address, amounts []*big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.CreditGasFees(&_OptimismMintableERC20.TransactOpts, recipients, amounts)
}

// CreditGasFees is a paid mutator transaction binding the contract method 0x2e0f98ad.
//
// Solidity: function creditGasFees(address[] recipients, uint256[] amounts) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorSession) CreditGasFees(recipients []common.Address, amounts []*big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.CreditGasFees(&_OptimismMintableERC20.TransactOpts, recipients, amounts)
}

// CreditGasFees0 is a paid mutator transaction binding the contract method 0x6a30b253.
//
// Solidity: function creditGasFees(address from, address feeRecipient, address , address communityFund, uint256 refund, uint256 tipTxFee, uint256 , uint256 baseTxFee) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20Transactor) CreditGasFees0(opts *bind.TransactOpts, from common.Address, feeRecipient common.Address, arg2 common.Address, communityFund common.Address, refund *big.Int, tipTxFee *big.Int, arg6 *big.Int, baseTxFee *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.contract.Transact(opts, "creditGasFees0", from, feeRecipient, arg2, communityFund, refund, tipTxFee, arg6, baseTxFee)
}

// CreditGasFees0 is a paid mutator transaction binding the contract method 0x6a30b253.
//
// Solidity: function creditGasFees(address from, address feeRecipient, address , address communityFund, uint256 refund, uint256 tipTxFee, uint256 , uint256 baseTxFee) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20Session) CreditGasFees0(from common.Address, feeRecipient common.Address, arg2 common.Address, communityFund common.Address, refund *big.Int, tipTxFee *big.Int, arg6 *big.Int, baseTxFee *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.CreditGasFees0(&_OptimismMintableERC20.TransactOpts, from, feeRecipient, arg2, communityFund, refund, tipTxFee, arg6, baseTxFee)
}

// CreditGasFees0 is a paid mutator transaction binding the contract method 0x6a30b253.
//
// Solidity: function creditGasFees(address from, address feeRecipient, address , address communityFund, uint256 refund, uint256 tipTxFee, uint256 , uint256 baseTxFee) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorSession) CreditGasFees0(from common.Address, feeRecipient common.Address, arg2 common.Address, communityFund common.Address, refund *big.Int, tipTxFee *big.Int, arg6 *big.Int, baseTxFee *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.CreditGasFees0(&_OptimismMintableERC20.TransactOpts, from, feeRecipient, arg2, communityFund, refund, tipTxFee, arg6, baseTxFee)
}

// DebitGasFees is a paid mutator transaction binding the contract method 0x58cf9672.
//
// Solidity: function debitGasFees(address from, uint256 value) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20Transactor) DebitGasFees(opts *bind.TransactOpts, from common.Address, value *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.contract.Transact(opts, "debitGasFees", from, value)
}

// DebitGasFees is a paid mutator transaction binding the contract method 0x58cf9672.
//
// Solidity: function debitGasFees(address from, uint256 value) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20Session) DebitGasFees(from common.Address, value *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.DebitGasFees(&_OptimismMintableERC20.TransactOpts, from, value)
}

// DebitGasFees is a paid mutator transaction binding the contract method 0x58cf9672.
//
// Solidity: function debitGasFees(address from, uint256 value) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorSession) DebitGasFees(from common.Address, value *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.DebitGasFees(&_OptimismMintableERC20.TransactOpts, from, value)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Transactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.DecreaseAllowance(&_OptimismMintableERC20.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.DecreaseAllowance(&_OptimismMintableERC20.TransactOpts, spender, subtractedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Transactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.IncreaseAllowance(&_OptimismMintableERC20.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.IncreaseAllowance(&_OptimismMintableERC20.TransactOpts, spender, addedValue)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address _to, uint256 _amount) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20Transactor) Mint(opts *bind.TransactOpts, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.contract.Transact(opts, "mint", _to, _amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address _to, uint256 _amount) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20Session) Mint(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.Mint(&_OptimismMintableERC20.TransactOpts, _to, _amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address _to, uint256 _amount) returns()
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorSession) Mint(_to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.Mint(&_OptimismMintableERC20.TransactOpts, _to, _amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Transactor) Transfer(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.contract.Transact(opts, "transfer", to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.Transfer(&_OptimismMintableERC20.TransactOpts, to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorSession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.Transfer(&_OptimismMintableERC20.TransactOpts, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Transactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.contract.Transact(opts, "transferFrom", from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20Session) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.TransferFrom(&_OptimismMintableERC20.TransactOpts, from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_OptimismMintableERC20 *OptimismMintableERC20TransactorSession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _OptimismMintableERC20.Contract.TransferFrom(&_OptimismMintableERC20.TransactOpts, from, to, amount)
}

// OptimismMintableERC20ApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the OptimismMintableERC20 contract.
type OptimismMintableERC20ApprovalIterator struct {
	Event *OptimismMintableERC20Approval // Event containing the contract specifics and raw log

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
func (it *OptimismMintableERC20ApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OptimismMintableERC20Approval)
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
		it.Event = new(OptimismMintableERC20Approval)
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
func (it *OptimismMintableERC20ApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OptimismMintableERC20ApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OptimismMintableERC20Approval represents a Approval event raised by the OptimismMintableERC20 contract.
type OptimismMintableERC20Approval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*OptimismMintableERC20ApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _OptimismMintableERC20.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &OptimismMintableERC20ApprovalIterator{contract: _OptimismMintableERC20.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *OptimismMintableERC20Approval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _OptimismMintableERC20.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OptimismMintableERC20Approval)
				if err := _OptimismMintableERC20.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) ParseApproval(log types.Log) (*OptimismMintableERC20Approval, error) {
	event := new(OptimismMintableERC20Approval)
	if err := _OptimismMintableERC20.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OptimismMintableERC20BurnIterator is returned from FilterBurn and is used to iterate over the raw logs and unpacked data for Burn events raised by the OptimismMintableERC20 contract.
type OptimismMintableERC20BurnIterator struct {
	Event *OptimismMintableERC20Burn // Event containing the contract specifics and raw log

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
func (it *OptimismMintableERC20BurnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OptimismMintableERC20Burn)
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
		it.Event = new(OptimismMintableERC20Burn)
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
func (it *OptimismMintableERC20BurnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OptimismMintableERC20BurnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OptimismMintableERC20Burn represents a Burn event raised by the OptimismMintableERC20 contract.
type OptimismMintableERC20Burn struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5.
//
// Solidity: event Burn(address indexed account, uint256 amount)
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) FilterBurn(opts *bind.FilterOpts, account []common.Address) (*OptimismMintableERC20BurnIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _OptimismMintableERC20.contract.FilterLogs(opts, "Burn", accountRule)
	if err != nil {
		return nil, err
	}
	return &OptimismMintableERC20BurnIterator{contract: _OptimismMintableERC20.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5.
//
// Solidity: event Burn(address indexed account, uint256 amount)
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *OptimismMintableERC20Burn, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _OptimismMintableERC20.contract.WatchLogs(opts, "Burn", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OptimismMintableERC20Burn)
				if err := _OptimismMintableERC20.contract.UnpackLog(event, "Burn", log); err != nil {
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
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) ParseBurn(log types.Log) (*OptimismMintableERC20Burn, error) {
	event := new(OptimismMintableERC20Burn)
	if err := _OptimismMintableERC20.contract.UnpackLog(event, "Burn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OptimismMintableERC20MintIterator is returned from FilterMint and is used to iterate over the raw logs and unpacked data for Mint events raised by the OptimismMintableERC20 contract.
type OptimismMintableERC20MintIterator struct {
	Event *OptimismMintableERC20Mint // Event containing the contract specifics and raw log

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
func (it *OptimismMintableERC20MintIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OptimismMintableERC20Mint)
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
		it.Event = new(OptimismMintableERC20Mint)
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
func (it *OptimismMintableERC20MintIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OptimismMintableERC20MintIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OptimismMintableERC20Mint represents a Mint event raised by the OptimismMintableERC20 contract.
type OptimismMintableERC20Mint struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMint is a free log retrieval operation binding the contract event 0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885.
//
// Solidity: event Mint(address indexed account, uint256 amount)
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) FilterMint(opts *bind.FilterOpts, account []common.Address) (*OptimismMintableERC20MintIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _OptimismMintableERC20.contract.FilterLogs(opts, "Mint", accountRule)
	if err != nil {
		return nil, err
	}
	return &OptimismMintableERC20MintIterator{contract: _OptimismMintableERC20.contract, event: "Mint", logs: logs, sub: sub}, nil
}

// WatchMint is a free log subscription operation binding the contract event 0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885.
//
// Solidity: event Mint(address indexed account, uint256 amount)
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) WatchMint(opts *bind.WatchOpts, sink chan<- *OptimismMintableERC20Mint, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _OptimismMintableERC20.contract.WatchLogs(opts, "Mint", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OptimismMintableERC20Mint)
				if err := _OptimismMintableERC20.contract.UnpackLog(event, "Mint", log); err != nil {
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
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) ParseMint(log types.Log) (*OptimismMintableERC20Mint, error) {
	event := new(OptimismMintableERC20Mint)
	if err := _OptimismMintableERC20.contract.UnpackLog(event, "Mint", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OptimismMintableERC20TransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the OptimismMintableERC20 contract.
type OptimismMintableERC20TransferIterator struct {
	Event *OptimismMintableERC20Transfer // Event containing the contract specifics and raw log

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
func (it *OptimismMintableERC20TransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OptimismMintableERC20Transfer)
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
		it.Event = new(OptimismMintableERC20Transfer)
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
func (it *OptimismMintableERC20TransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OptimismMintableERC20TransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OptimismMintableERC20Transfer represents a Transfer event raised by the OptimismMintableERC20 contract.
type OptimismMintableERC20Transfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*OptimismMintableERC20TransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _OptimismMintableERC20.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &OptimismMintableERC20TransferIterator{contract: _OptimismMintableERC20.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *OptimismMintableERC20Transfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _OptimismMintableERC20.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OptimismMintableERC20Transfer)
				if err := _OptimismMintableERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_OptimismMintableERC20 *OptimismMintableERC20Filterer) ParseTransfer(log types.Log) (*OptimismMintableERC20Transfer, error) {
	event := new(OptimismMintableERC20Transfer)
	if err := _OptimismMintableERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
