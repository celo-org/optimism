// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-bindings/solc"
)

const LegacyERC20ETHStorageLayoutJSON = "{\"storage\":[{\"astId\":1000,\"contract\":\"src/legacy/LegacyERC20ETH.sol:LegacyERC20ETH\",\"label\":\"_balances\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_mapping(t_address,t_uint256)\"},{\"astId\":1001,\"contract\":\"src/legacy/LegacyERC20ETH.sol:LegacyERC20ETH\",\"label\":\"_allowances\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_mapping(t_address,t_mapping(t_address,t_uint256))\"},{\"astId\":1002,\"contract\":\"src/legacy/LegacyERC20ETH.sol:LegacyERC20ETH\",\"label\":\"_totalSupply\",\"offset\":0,\"slot\":\"2\",\"type\":\"t_uint256\"},{\"astId\":1003,\"contract\":\"src/legacy/LegacyERC20ETH.sol:LegacyERC20ETH\",\"label\":\"_name\",\"offset\":0,\"slot\":\"3\",\"type\":\"t_string_storage\"},{\"astId\":1004,\"contract\":\"src/legacy/LegacyERC20ETH.sol:LegacyERC20ETH\",\"label\":\"_symbol\",\"offset\":0,\"slot\":\"4\",\"type\":\"t_string_storage\"}],\"types\":{\"t_address\":{\"encoding\":\"inplace\",\"label\":\"address\",\"numberOfBytes\":\"20\"},\"t_mapping(t_address,t_mapping(t_address,t_uint256))\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e mapping(address =\u003e uint256))\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_mapping(t_address,t_uint256)\"},\"t_mapping(t_address,t_uint256)\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e uint256)\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_uint256\"},\"t_string_storage\":{\"encoding\":\"bytes\",\"label\":\"string\",\"numberOfBytes\":\"32\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"}}}"

var LegacyERC20ETHStorageLayout = new(solc.StorageLayout)

var LegacyERC20ETHDeployedBin = "0x608060405234801561001057600080fd5b50600436106101985760003560e01c80636a30b253116100e3578063ae1f6aaf1161008c578063dd62ed3e11610066578063dd62ed3e146103e1578063e78cea9214610395578063ee9a31a21461042757600080fd5b8063ae1f6aaf14610395578063c01e1bd6146103bb578063d6c0b2c4146103bb57600080fd5b80639dc29fac116100bd5780639dc29fac1461035c578063a457c2d71461036f578063a9059cbb1461038257600080fd5b80636a30b2531461031957806370a082311461032c57806395d89b411461035457600080fd5b80632e0f98ad1161014557806340c10f191161011f57806340c10f19146102b757806354fd4d50146102ca57806358cf96721461030657600080fd5b80632e0f98ad1461025e578063313ce5671461027357806339509351146102a457600080fd5b8063095ea7b311610176578063095ea7b31461022657806318160ddd1461023957806323b872dd1461024b57600080fd5b806301ffc9a71461019d578063033964be146101c557806306fdde0314610211575b600080fd5b6101b06101ab366004610ed5565b61044e565b60405190151581526020015b60405180910390f35b6101ec7f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016101bc565b61021961053f565b6040516101bc9190610f1e565b6101b0610234366004610fba565b6105d1565b6002545b6040519081526020016101bc565b6101b0610259366004610fe4565b610661565b61027161026c36600461106c565b6106ec565b005b60405160ff7f00000000000000000000000000000000000000000000000000000000000000001681526020016101bc565b6101b06102b2366004610fba565b61084f565b6102716102c5366004610fba565b6108da565b6102196040518060400160405280600581526020017f312e332e3000000000000000000000000000000000000000000000000000000081525081565b610271610314366004610fba565b61093c565b6102716103273660046110d8565b6109b2565b61023d61033a366004611150565b73ffffffffffffffffffffffffffffffffffffffff163190565b610219610a42565b61027161036a366004610fba565b610a51565b6101b061037d366004610fba565b610ab3565b6101b0610390366004610fba565b610b3e565b7f00000000000000000000000000000000000000000000000000000000000000006101ec565b7f00000000000000000000000000000000000000000000000000000000000000006101ec565b61023d6103ef36600461116b565b73ffffffffffffffffffffffffffffffffffffffff918216600090815260016020908152604080832093909416825291909152205490565b6101ec7f000000000000000000000000000000000000000000000000000000000000000081565b60007f01ffc9a7000000000000000000000000000000000000000000000000000000007f1d1d8b63000000000000000000000000000000000000000000000000000000007fec4fc8e3000000000000000000000000000000000000000000000000000000007fffffffff00000000000000000000000000000000000000000000000000000000851683148061050757507fffffffff00000000000000000000000000000000000000000000000000000000858116908316145b8061053657507fffffffff00000000000000000000000000000000000000000000000000000000858116908216145b95945050505050565b60606003805461054e9061119e565b80601f016020809104026020016040519081016040528092919081815260200182805461057a9061119e565b80156105c75780601f1061059c576101008083540402835291602001916105c7565b820191906000526020600020905b8154815290600101906020018083116105aa57829003601f168201915b5050505050905090565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f4c656761637945524332304554483a20617070726f766520697320646973616260448201527f6c6564000000000000000000000000000000000000000000000000000000000060648201526000906084015b60405180910390fd5b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602860248201527f4c656761637945524332304554483a207472616e7366657246726f6d2069732060448201527f64697361626c65640000000000000000000000000000000000000000000000006064820152600090608401610658565b3315610754576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c000000000000000000000000000000006044820152606401610658565b8281146107e3576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602f60248201527f526563697069656e747320616e6420616d6f756e7473206d757374206265207460448201527f68652073616d65206c656e6774682e00000000000000000000000000000000006064820152608401610658565b60005b8381101561084857610836858583818110610803576108036111f1565b90506020020160208101906108189190611150565b84848481811061082a5761082a6111f1565b90506020020135610bc8565b806108408161124f565b9150506107e6565b5050505050565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f4c656761637945524332304554483a20696e637265617365416c6c6f77616e6360448201527f652069732064697361626c6564000000000000000000000000000000000000006064820152600090608401610658565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4c656761637945524332304554483a206d696e742069732064697361626c65646044820152606401610658565b33156109a4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c000000000000000000000000000000006044820152606401610658565b6109ae8282610ce8565b5050565b3315610a1a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f4f6e6c7920564d2063616e2063616c6c000000000000000000000000000000006044820152606401610658565b610a248885610bc8565b610a2e8784610bc8565b610a388582610bc8565b5050505050505050565b60606004805461054e9061119e565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4c656761637945524332304554483a206275726e2069732064697361626c65646044820152606401610658565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f4c656761637945524332304554483a206465637265617365416c6c6f77616e6360448201527f652069732064697361626c6564000000000000000000000000000000000000006064820152600090608401610658565b6040517f08c379a0000000000000000000000000000000000000000000000000000000008152602060048201526024808201527f4c656761637945524332304554483a207472616e73666572206973206469736160448201527f626c6564000000000000000000000000000000000000000000000000000000006064820152600090608401610658565b73ffffffffffffffffffffffffffffffffffffffff8216610c45576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f45524332303a206d696e7420746f20746865207a65726f2061646472657373006044820152606401610658565b8060026000828254610c579190611287565b909155505073ffffffffffffffffffffffffffffffffffffffff821660009081526020819052604081208054839290610c91908490611287565b909155505060405181815273ffffffffffffffffffffffffffffffffffffffff8316906000907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9060200160405180910390a35050565b73ffffffffffffffffffffffffffffffffffffffff8216610d8b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602160248201527f45524332303a206275726e2066726f6d20746865207a65726f2061646472657360448201527f73000000000000000000000000000000000000000000000000000000000000006064820152608401610658565b73ffffffffffffffffffffffffffffffffffffffff821660009081526020819052604090205481811015610e41576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f45524332303a206275726e20616d6f756e7420657863656564732062616c616e60448201527f63650000000000000000000000000000000000000000000000000000000000006064820152608401610658565b73ffffffffffffffffffffffffffffffffffffffff83166000908152602081905260408120838303905560028054849290610e7d90849061129f565b909155505060405182815260009073ffffffffffffffffffffffffffffffffffffffff8516907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9060200160405180910390a3505050565b600060208284031215610ee757600080fd5b81357fffffffff0000000000000000000000000000000000000000000000000000000081168114610f1757600080fd5b9392505050565b600060208083528351808285015260005b81811015610f4b57858101830151858201604001528201610f2f565b81811115610f5d576000604083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016929092016040019392505050565b803573ffffffffffffffffffffffffffffffffffffffff81168114610fb557600080fd5b919050565b60008060408385031215610fcd57600080fd5b610fd683610f91565b946020939093013593505050565b600080600060608486031215610ff957600080fd5b61100284610f91565b925061101060208501610f91565b9150604084013590509250925092565b60008083601f84011261103257600080fd5b50813567ffffffffffffffff81111561104a57600080fd5b6020830191508360208260051b850101111561106557600080fd5b9250929050565b6000806000806040858703121561108257600080fd5b843567ffffffffffffffff8082111561109a57600080fd5b6110a688838901611020565b909650945060208701359150808211156110bf57600080fd5b506110cc87828801611020565b95989497509550505050565b600080600080600080600080610100898b0312156110f557600080fd5b6110fe89610f91565b975061110c60208a01610f91565b965061111a60408a01610f91565b955061112860608a01610f91565b979a969950949760808101359660a0820135965060c0820135955060e0909101359350915050565b60006020828403121561116257600080fd5b610f1782610f91565b6000806040838503121561117e57600080fd5b61118783610f91565b915061119560208401610f91565b90509250929050565b600181811c908216806111b257607f821691505b6020821081036111eb577f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff820361128057611280611220565b5060010190565b6000821982111561129a5761129a611220565b500190565b6000828210156112b1576112b1611220565b50039056fea164736f6c634300080f000a"


func init() {
	if err := json.Unmarshal([]byte(LegacyERC20ETHStorageLayoutJSON), LegacyERC20ETHStorageLayout); err != nil {
		panic(err)
	}

	layouts["LegacyERC20ETH"] = LegacyERC20ETHStorageLayout
	deployedBytecodes["LegacyERC20ETH"] = LegacyERC20ETHDeployedBin
	immutableReferences["LegacyERC20ETH"] = true
}
