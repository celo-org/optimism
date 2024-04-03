// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-bindings/solc"
)

const MockSortedOraclesStorageLayoutJSON = "{\"storage\":[{\"astId\":1000,\"contract\":\"src/celo/testing/MockSortedOracles.sol:MockSortedOracles\",\"label\":\"numerators\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_mapping(t_address,t_uint256)\"},{\"astId\":1001,\"contract\":\"src/celo/testing/MockSortedOracles.sol:MockSortedOracles\",\"label\":\"medianTimestamp\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_mapping(t_address,t_uint256)\"},{\"astId\":1002,\"contract\":\"src/celo/testing/MockSortedOracles.sol:MockSortedOracles\",\"label\":\"numRates\",\"offset\":0,\"slot\":\"2\",\"type\":\"t_mapping(t_address,t_uint256)\"},{\"astId\":1003,\"contract\":\"src/celo/testing/MockSortedOracles.sol:MockSortedOracles\",\"label\":\"expired\",\"offset\":0,\"slot\":\"3\",\"type\":\"t_mapping(t_address,t_bool)\"}],\"types\":{\"t_address\":{\"encoding\":\"inplace\",\"label\":\"address\",\"numberOfBytes\":\"20\"},\"t_bool\":{\"encoding\":\"inplace\",\"label\":\"bool\",\"numberOfBytes\":\"1\"},\"t_mapping(t_address,t_bool)\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e bool)\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_bool\"},\"t_mapping(t_address,t_uint256)\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e uint256)\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_uint256\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"}}}"

var MockSortedOraclesStorageLayout = new(solc.StorageLayout)

var MockSortedOraclesDeployedBin = "0x608060405234801561001057600080fd5b50600436106100d45760003560e01c8063b325dd1f11610081578063f06a10e01161005b578063f06a10e0146102c6578063f7ca6963146102e9578063ffe736bf1461030957600080fd5b8063b325dd1f14610235578063bbc66a941461027e578063ef90e1b01461029e57600080fd5b806376ea210e116100b257806376ea210e1461017c5780638d24fc51146101c7578063918f86741461022457600080fd5b8063071b48fc146100d95780631b7aa2db1461010c578063495182a614610145575b600080fd5b6100f96100e7366004610406565b60016020526000908152604090205481565b6040519081526020015b60405180910390f35b61014361011a366004610428565b73ffffffffffffffffffffffffffffffffffffffff909116600090815260016020526040902055565b005b610143610153366004610428565b73ffffffffffffffffffffffffffffffffffffffff909116600090815260026020526040902055565b6101b761018a366004610428565b73ffffffffffffffffffffffffffffffffffffffff91909116600090815260208190526040902055600190565b6040519015158152602001610103565b6101436101d5366004610406565b73ffffffffffffffffffffffffffffffffffffffff16600090815260036020526040902080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00166001179055565b6100f969d3c21bcecceda100000081565b610143610243366004610406565b73ffffffffffffffffffffffffffffffffffffffff1660009081526001602052604090206fffffffffffffffffffffffffffffffff42169055565b6100f961028c366004610406565b60026020526000908152604090205481565b6102b16102ac366004610406565b61036f565b60408051928352602083019190915201610103565b6101b76102d4366004610406565b60036020526000908152604090205460ff1681565b6100f96102f7366004610406565b60006020819052908152604090205481565b610343610317366004610406565b73ffffffffffffffffffffffffffffffffffffffff811660009081526003602052604090205460ff1691565b60408051921515835273ffffffffffffffffffffffffffffffffffffffff909116602083015201610103565b73ffffffffffffffffffffffffffffffffffffffff81166000908152602081905260408120548190156103d257505073ffffffffffffffffffffffffffffffffffffffff166000908152602081905260409020549069d3c21bcecceda100000090565b506000928392509050565b803573ffffffffffffffffffffffffffffffffffffffff8116811461040157600080fd5b919050565b60006020828403121561041857600080fd5b610421826103dd565b9392505050565b6000806040838503121561043b57600080fd5b610444836103dd565b94602093909301359350505056fea164736f6c6343000813000a"


func init() {
	if err := json.Unmarshal([]byte(MockSortedOraclesStorageLayoutJSON), MockSortedOraclesStorageLayout); err != nil {
		panic(err)
	}

	layouts["MockSortedOracles"] = MockSortedOraclesStorageLayout
	deployedBytecodes["MockSortedOracles"] = MockSortedOraclesDeployedBin
	immutableReferences["MockSortedOracles"] = false
}
