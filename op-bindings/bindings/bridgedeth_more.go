// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-bindings/solc"
)

const BridgedETHStorageLayoutJSON = "{\"storage\":[{\"astId\":1000,\"contract\":\"src/celo/BridgedETH.sol:BridgedETH\",\"label\":\"_balances\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_mapping(t_address,t_uint256)\"},{\"astId\":1001,\"contract\":\"src/celo/BridgedETH.sol:BridgedETH\",\"label\":\"_allowances\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_mapping(t_address,t_mapping(t_address,t_uint256))\"},{\"astId\":1002,\"contract\":\"src/celo/BridgedETH.sol:BridgedETH\",\"label\":\"_totalSupply\",\"offset\":0,\"slot\":\"2\",\"type\":\"t_uint256\"},{\"astId\":1003,\"contract\":\"src/celo/BridgedETH.sol:BridgedETH\",\"label\":\"_name\",\"offset\":0,\"slot\":\"3\",\"type\":\"t_string_storage\"},{\"astId\":1004,\"contract\":\"src/celo/BridgedETH.sol:BridgedETH\",\"label\":\"_symbol\",\"offset\":0,\"slot\":\"4\",\"type\":\"t_string_storage\"}],\"types\":{\"t_address\":{\"encoding\":\"inplace\",\"label\":\"address\",\"numberOfBytes\":\"20\"},\"t_mapping(t_address,t_mapping(t_address,t_uint256))\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e mapping(address =\u003e uint256))\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_mapping(t_address,t_uint256)\"},\"t_mapping(t_address,t_uint256)\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e uint256)\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_uint256\"},\"t_string_storage\":{\"encoding\":\"bytes\",\"label\":\"string\",\"numberOfBytes\":\"32\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"}}}"

var BridgedETHStorageLayout = new(solc.StorageLayout)

var BridgedETHDeployedBin = "0x608060405234801561001057600080fd5b50600436106101985760003560e01c80636a30b253116100e3578063ae1f6aaf1161008c578063dd62ed3e11610066578063dd62ed3e146103ef578063e78cea92146103a3578063ee9a31a21461043557600080fd5b8063ae1f6aaf146103a3578063c01e1bd6146103c9578063d6c0b2c4146103c957600080fd5b80639dc29fac116100bd5780639dc29fac1461036a578063a457c2d71461037d578063a9059cbb1461039057600080fd5b80636a30b2531461031957806370a082311461032c57806395d89b411461036257600080fd5b80632e0f98ad1161014557806340c10f191161011f57806340c10f19146102b757806354fd4d50146102ca57806358cf96721461030657600080fd5b80632e0f98ad1461025e578063313ce5671461027357806339509351146102a457600080fd5b8063095ea7b311610176578063095ea7b31461022657806318160ddd1461023957806323b872dd1461024b57600080fd5b806301ffc9a71461019d578063033964be146101c557806306fdde0314610211575b600080fd5b6101b06101ab366004611440565b61045c565b60405190151581526020015b60405180910390f35b6101ec7f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016101bc565b61021961054d565b6040516101bc9190611489565b6101b0610234366004611525565b6105df565b6002545b6040519081526020016101bc565b6101b061025936600461154f565b6105f7565b61027161026c3660046115d7565b61061b565b005b60405160ff7f00000000000000000000000000000000000000000000000000000000000000001681526020016101bc565b6101b06102b2366004611525565b610783565b6102716102c5366004611525565b6107cf565b6102196040518060400160405280600581526020017f312e332e3000000000000000000000000000000000000000000000000000000081525081565b610271610314366004611525565b6108f2565b610271610327366004611643565b610968565b61023d61033a3660046116bb565b73ffffffffffffffffffffffffffffffffffffffff1660009081526020819052604090205490565b6102196109f8565b610271610378366004611525565b610a07565b6101b061038b366004611525565b610b1e565b6101b061039e366004611525565b610bef565b7f00000000000000000000000000000000000000000000000000000000000000006101ec565b7f00000000000000000000000000000000000000000000000000000000000000006101ec565b61023d6103fd3660046116d6565b73ffffffffffffffffffffffffffffffffffffffff918216600090815260016020908152604080832093909416825291909152205490565b6101ec7f000000000000000000000000000000000000000000000000000000000000000081565b60007f01ffc9a7000000000000000000000000000000000000000000000000000000007f1d1d8b63000000000000000000000000000000000000000000000000000000007fec4fc8e3000000000000000000000000000000000000000000000000000000007fffffffff00000000000000000000000000000000000000000000000000000000851683148061051557507fffffffff00000000000000000000000000000000000000000000000000000000858116908316145b8061054457507fffffffff00000000000000000000000000000000000000000000000000000000858116908216145b95945050505050565b60606003805461055c90611709565b80601f016020809104026020016040519081016040528092919081815260200182805461058890611709565b80156105d55780601f106105aa576101008083540402835291602001916105d5565b820191906000526020600020905b8154815290600101906020018083116105b857829003601f168201915b5050505050905090565b6000336105ed818585610bfd565b5060019392505050565b600033610605858285610db1565b610610858585610e88565b506001949350505050565b3315610688576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c0000000000000000000000000000000060448201526064015b60405180910390fd5b828114610717576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602f60248201527f526563697069656e747320616e6420616d6f756e7473206d757374206265207460448201527f68652073616d65206c656e6774682e0000000000000000000000000000000000606482015260840161067f565b60005b8381101561077c5761076a8585838181106107375761073761175c565b905060200201602081019061074c91906116bb565b84848481811061075e5761075e61175c565b9050602002013561113b565b80610774816117ba565b91505061071a565b5050505050565b33600081815260016020908152604080832073ffffffffffffffffffffffffffffffffffffffff871684529091528120549091906105ed90829086906107ca9087906117f2565b610bfd565b3373ffffffffffffffffffffffffffffffffffffffff7f00000000000000000000000000000000000000000000000000000000000000001614610894576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603460248201527f4f7074696d69736d4d696e7461626c6545524332303a206f6e6c79206272696460448201527f67652063616e206d696e7420616e64206275726e000000000000000000000000606482015260840161067f565b61089e828261113b565b8173ffffffffffffffffffffffffffffffffffffffff167f0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885826040516108e691815260200190565b60405180910390a25050565b331561095a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c00000000000000000000000000000000604482015260640161067f565b610964828261125b565b5050565b33156109d0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c00000000000000000000000000000000604482015260640161067f565b6109da888561113b565b6109e4878461113b565b6109ee858261113b565b5050505050505050565b60606004805461055c90611709565b3373ffffffffffffffffffffffffffffffffffffffff7f00000000000000000000000000000000000000000000000000000000000000001614610acc576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603460248201527f4f7074696d69736d4d696e7461626c6545524332303a206f6e6c79206272696460448201527f67652063616e206d696e7420616e64206275726e000000000000000000000000606482015260840161067f565b610ad6828261125b565b8173ffffffffffffffffffffffffffffffffffffffff167fcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5826040516108e691815260200190565b33600081815260016020908152604080832073ffffffffffffffffffffffffffffffffffffffff8716845290915281205490919083811015610be2576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602560248201527f45524332303a2064656372656173656420616c6c6f77616e63652062656c6f7760448201527f207a65726f000000000000000000000000000000000000000000000000000000606482015260840161067f565b6106108286868403610bfd565b6000336105ed818585610e88565b73ffffffffffffffffffffffffffffffffffffffff8316610c9f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152602060048201526024808201527f45524332303a20617070726f76652066726f6d20746865207a65726f2061646460448201527f7265737300000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff8216610d42576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f45524332303a20617070726f766520746f20746865207a65726f20616464726560448201527f7373000000000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff83811660008181526001602090815260408083209487168084529482529182902085905590518481527f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92591015b60405180910390a3505050565b73ffffffffffffffffffffffffffffffffffffffff8381166000908152600160209081526040808320938616835292905220547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8114610e825781811015610e75576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f45524332303a20696e73756666696369656e7420616c6c6f77616e6365000000604482015260640161067f565b610e828484848403610bfd565b50505050565b73ffffffffffffffffffffffffffffffffffffffff8316610f2b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602560248201527f45524332303a207472616e736665722066726f6d20746865207a65726f20616460448201527f6472657373000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff8216610fce576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f45524332303a207472616e7366657220746f20746865207a65726f206164647260448201527f6573730000000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff831660009081526020819052604090205481811015611084576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f45524332303a207472616e7366657220616d6f756e742065786365656473206260448201527f616c616e63650000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff8085166000908152602081905260408082208585039055918516815290812080548492906110c89084906117f2565b925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef8460405161112e91815260200190565b60405180910390a3610e82565b73ffffffffffffffffffffffffffffffffffffffff82166111b8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f45524332303a206d696e7420746f20746865207a65726f206164647265737300604482015260640161067f565b80600260008282546111ca91906117f2565b909155505073ffffffffffffffffffffffffffffffffffffffff8216600090815260208190526040812080548392906112049084906117f2565b909155505060405181815273ffffffffffffffffffffffffffffffffffffffff8316906000907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9060200160405180910390a35050565b73ffffffffffffffffffffffffffffffffffffffff82166112fe576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602160248201527f45524332303a206275726e2066726f6d20746865207a65726f2061646472657360448201527f7300000000000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff8216600090815260208190526040902054818110156113b4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f45524332303a206275726e20616d6f756e7420657863656564732062616c616e60448201527f6365000000000000000000000000000000000000000000000000000000000000606482015260840161067f565b73ffffffffffffffffffffffffffffffffffffffff831660009081526020819052604081208383039055600280548492906113f090849061180a565b909155505060405182815260009073ffffffffffffffffffffffffffffffffffffffff8516907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef90602001610da4565b60006020828403121561145257600080fd5b81357fffffffff000000000000000000000000000000000000000000000000000000008116811461148257600080fd5b9392505050565b600060208083528351808285015260005b818110156114b65785810183015185820160400152820161149a565b818111156114c8576000604083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016929092016040019392505050565b803573ffffffffffffffffffffffffffffffffffffffff8116811461152057600080fd5b919050565b6000806040838503121561153857600080fd5b611541836114fc565b946020939093013593505050565b60008060006060848603121561156457600080fd5b61156d846114fc565b925061157b602085016114fc565b9150604084013590509250925092565b60008083601f84011261159d57600080fd5b50813567ffffffffffffffff8111156115b557600080fd5b6020830191508360208260051b85010111156115d057600080fd5b9250929050565b600080600080604085870312156115ed57600080fd5b843567ffffffffffffffff8082111561160557600080fd5b6116118883890161158b565b9096509450602087013591508082111561162a57600080fd5b506116378782880161158b565b95989497509550505050565b600080600080600080600080610100898b03121561166057600080fd5b611669896114fc565b975061167760208a016114fc565b965061168560408a016114fc565b955061169360608a016114fc565b979a969950949760808101359660a0820135965060c0820135955060e0909101359350915050565b6000602082840312156116cd57600080fd5b611482826114fc565b600080604083850312156116e957600080fd5b6116f2836114fc565b9150611700602084016114fc565b90509250929050565b600181811c9082168061171d57607f821691505b602082108103611756577f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82036117eb576117eb61178b565b5060010190565b600082198211156118055761180561178b565b500190565b60008282101561181c5761181c61178b565b50039056fea164736f6c634300080f000a"


func init() {
	if err := json.Unmarshal([]byte(BridgedETHStorageLayoutJSON), BridgedETHStorageLayout); err != nil {
		panic(err)
	}

	layouts["BridgedETH"] = BridgedETHStorageLayout
	deployedBytecodes["BridgedETH"] = BridgedETHDeployedBin
	immutableReferences["BridgedETH"] = true
}
