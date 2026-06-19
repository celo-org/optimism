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
	_ = abi.ConvertType
)

// EspressoTEEVerifierMetaData contains all meta data concerning the EspressoTEEVerifier contract.
var EspressoTEEVerifierMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"addGuardian\",\"inputs\":[{\"name\":\"guardian\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"deleteEnclaveHashes\",\"inputs\":[{\"name\":\"enclaveHashes\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"teeType\",\"type\":\"uint8\",\"internalType\":\"enumIEspressoTEEVerifier.TeeType\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"espressoNitroTEEVerifier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIEspressoNitroTEEVerifier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getGuardians\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"guardianCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_espressoNitroTEEVerifier\",\"type\":\"address\",\"internalType\":\"contractIEspressoNitroTEEVerifier\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isGuardian\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isSignerValid\",\"inputs\":[{\"name\":\"signer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"teeType\",\"type\":\"uint8\",\"internalType\":\"enumIEspressoTEEVerifier.TeeType\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerService\",\"inputs\":[{\"name\":\"verificationData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"teeType\",\"type\":\"uint8\",\"internalType\":\"enumIEspressoTEEVerifier.TeeType\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registeredEnclaveHashes\",\"inputs\":[{\"name\":\"enclaveHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"teeType\",\"type\":\"uint8\",\"internalType\":\"enumIEspressoTEEVerifier.TeeType\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"removeGuardian\",\"inputs\":[{\"name\":\"guardian\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setEnclaveHash\",\"inputs\":[{\"name\":\"enclaveHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"valid\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"teeType\",\"type\":\"uint8\",\"internalType\":\"enumIEspressoTEEVerifier.TeeType\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setEspressoNitroTEEVerifier\",\"inputs\":[{\"name\":\"_espressoNitroTEEVerifier\",\"type\":\"address\",\"internalType\":\"contractIEspressoNitroTEEVerifier\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setNitroEnclaveVerifier\",\"inputs\":[{\"name\":\"nitroVerifier\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"verify\",\"inputs\":[{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"userDataHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"teeType\",\"type\":\"uint8\",\"internalType\":\"enumIEspressoTEEVerifier.TeeType\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EspressoNitroTEEVerifierSet\",\"inputs\":[{\"name\":\"oldVerifier\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newVerifier\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"GuardianAdded\",\"inputs\":[{\"name\":\"guardian\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"GuardianRemoved\",\"inputs\":[{\"name\":\"guardian\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"InvalidGuardianAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidVerifierAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotGuardian\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"NotGuardianOrOwner\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnerCantBeGuardian\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UnsupportedTeeType\",\"inputs\":[{\"name\":\"teeType\",\"type\":\"uint8\",\"internalType\":\"enumIEspressoTEEVerifier.TeeType\"}]}]",
	Bin: "0x6080604052348015600e575f5ffd5b5060156019565b60c9565b7ff0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a00805468010000000000000000900460ff161560685760405163f92ee8a960e01b815260040160405180910390fd5b80546001600160401b039081161460c65780546001600160401b0319166001600160401b0390811782556040519081527fc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d29060200160405180910390a15b50565b612528806100d65f395ff3fe608060405234801561000f575f5ffd5b506004361061016e575f3560e01c806379ba5097116100d2578063a81d9c5c11610088578063dac79fc811610063578063dac79fc81461031e578063e30c397814610331578063f2fde38b14610339575f5ffd5b8063a81d9c5c146102bb578063cd8f6997146102ce578063d80a4c28146102e1575f5ffd5b80638da5cb5b116100b85780638da5cb5b14610268578063a526d83b14610295578063a628a19e146102a8575f5ffd5b806379ba50971461024557806384b0196e1461024d575f5ffd5b8063485cc955116101275780636d8f5aa91161010d5780636d8f5aa914610217578063714041561461022a578063715018a61461023d575f5ffd5b8063485cc955146101ee57806354387ad714610201575f5ffd5b80630f1f0f86116101575780630f1f0f86146101b3578063330282f5146101c85780633cbe6803146101db575f5ffd5b80630665f04b146101725780630c68ba2114610190575b5f5ffd5b61017a61034c565b6040516101879190611c67565b60405180910390f35b6101a361019e366004611ce0565b61037c565b6040519015158152602001610187565b6101c66101c1366004611d1b565b6103af565b005b6101c66101d6366004611ce0565b61050f565b6101a36101e9366004611d56565b6105fd565b6101c66101fc366004611d80565b6106c1565b6102096108ff565b604051908152602001610187565b6101a3610225366004611db7565b610929565b6101c6610238366004611ce0565b6109ab565b6101c6610a2b565b6101c6610a3e565b610255610ab6565b6040516101879796959493929190611e2d565b610270610bb0565b60405173ffffffffffffffffffffffffffffffffffffffff9091168152602001610187565b6101c66102a3366004611ce0565b610bf1565b6101c66102b6366004611ce0565b610d65565b6101a36102c9366004611f68565b610e0f565b6101c66102dc366004612027565b610f73565b7f89639f446056f5d7661bbd94e8ab0617a80058ed7b072845818d4b93332e48005473ffffffffffffffffffffffffffffffffffffffff16610270565b6101c661032c366004612122565b611023565b6102706110dd565b6101c6610347366004611ce0565b611105565b60606103777f0f4ac8aae5a4fa6a3612928fcd8255b475ff86b500ae30bb272e61542cfc6f00611177565b905090565b5f6103a9827f0f4ac8aae5a4fa6a3612928fcd8255b475ff86b500ae30bb272e61542cfc6f005b9061118a565b92915050565b6103d9337f0f4ac8aae5a4fa6a3612928fcd8255b475ff86b500ae30bb272e61542cfc6f006103a3565b15801561041957506103e9610bb0565b73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614155b15610457576040517fd53780c40000000000000000000000000000000000000000000000000000000081523360048201526024015b60405180910390fd5b610460816111b8565b5f7f89639f446056f5d7661bbd94e8ab0617a80058ed7b072845818d4b93332e480080546040517f93b5552e00000000000000000000000000000000000000000000000000000000815260048101879052851515602482015291925073ffffffffffffffffffffffffffffffffffffffff16906393b5552e906044015f604051808303815f87803b1580156104f3575f5ffd5b505af1158015610505573d5f5f3e3d5ffd5b5050505050505050565b610517611201565b73ffffffffffffffffffffffffffffffffffffffff8116610564576040517f10c40e8c00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b7f89639f446056f5d7661bbd94e8ab0617a80058ed7b072845818d4b93332e480080547fffffffffffffffffffffffff0000000000000000000000000000000000000000811673ffffffffffffffffffffffffffffffffffffffff848116918217845560405192169184919083907facbe6384d204f2883b25fa4d03777cb88af5b2afaaffbe06609d3e17dc68d096905f90a350505050565b5f610607826111b8565b7f89639f446056f5d7661bbd94e8ab0617a80058ed7b072845818d4b93332e480080546040517f966989ee0000000000000000000000000000000000000000000000000000000081526004810186905273ffffffffffffffffffffffffffffffffffffffff9091169063966989ee906024015b602060405180830381865afa158015610695573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906106b991906121a2565b949350505050565b5f6106ca611259565b805490915060ff68010000000000000000820416159067ffffffffffffffff165f811580156106f65750825b90505f8267ffffffffffffffff1660011480156107125750303b155b905081158015610720575080155b15610757576040517ff92ee8a900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b84547fffffffffffffffffffffffffffffffffffffffffffffffff000000000000000016600117855583156107b85784547fffffffffffffffffffffffffffffffffffffffffffffff00ffffffffffffffff16680100000000000000001785555b7f89639f446056f5d7661bbd94e8ab0617a80058ed7b072845818d4b93332e480080547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff881617815561082088611281565b6108946040518060400160405280601381526020017f457370726573736f5445455665726966696572000000000000000000000000008152506040518060400160405280600181526020017f3100000000000000000000000000000000000000000000000000000000000000815250611292565b5083156108f65784547fffffffffffffffffffffffffffffffffffffffffffffff00ffffffffffffffff168555604051600181527fc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d29060200160405180910390a15b50505050505050565b5f6103777f0f4ac8aae5a4fa6a3612928fcd8255b475ff86b500ae30bb272e61542cfc6f006112a4565b5f610933826111b8565b7f89639f446056f5d7661bbd94e8ab0617a80058ed7b072845818d4b93332e480080546040517f4460790300000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff86811660048301529091169063446079039060240161067a565b6109b3611201565b7f0f4ac8aae5a4fa6a3612928fcd8255b475ff86b500ae30bb272e61542cfc6f006109de81836112ad565b6109e6575050565b60405173ffffffffffffffffffffffffffffffffffffffff8316907fb8107d0c6b40be480ce3172ee66ba6d64b71f6b1685a851340036e6e2e3e3c52905f90a2505b50565b610a33611201565b610a3c5f6112ce565b565b3380610a486110dd565b73ffffffffffffffffffffffffffffffffffffffff1614610aad576040517f118cdaa700000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8216600482015260240161044e565b610a28816112ce565b5f60608082808083817fa16a46d94261c7517cc8ff89f61c0ce93598e3c849801011dee649a6a557d1008054909150158015610af457506001810154155b610b5a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601560248201527f4549503731323a20556e696e697469616c697a65640000000000000000000000604482015260640161044e565b610b6261131e565b610b6a6113f1565b604080515f808252602082019092527f0f000000000000000000000000000000000000000000000000000000000000009c939b5091995046985030975095509350915050565b5f807f9016d09d72d40fdae2fd8ceac6b6234c7706214fd39c1cd1e609a0528c1993005b5473ffffffffffffffffffffffffffffffffffffffff1692915050565b610bf9611201565b73ffffffffffffffffffffffffffffffffffffffff8116610c46576040517f1b08105400000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b610c4e610bb0565b73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161480610cb95750610c8a6110dd565b73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16145b15610cf0576040517f3af3c41c00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b7f0f4ac8aae5a4fa6a3612928fcd8255b475ff86b500ae30bb272e61542cfc6f00610d1b8183611442565b15610d615760405173ffffffffffffffffffffffffffffffffffffffff8316907f038596bb31e2e7d3d9f184d4c98b310103f6d7f5830e5eec32bffe6f1728f969905f90a25b5050565b610d6d611201565b7f89639f446056f5d7661bbd94e8ab0617a80058ed7b072845818d4b93332e4800546040517fa628a19e00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff83811660048301529091169063a628a19e906024015f604051808303815f87803b158015610df6575f5ffd5b505af1158015610e08573d5f5f3e3d5ffd5b5050505050565b5f610e19826111b8565b604080517ff2069c4294e295535d771d7fbbd52e3db330ca380d24e42af36427513a27649b602080830191909152818301869052825180830384018152606090920190925280519101207f89639f446056f5d7661bbd94e8ab0617a80058ed7b072845818d4b93332e4800905f610e8f82611463565b90505f610e9c82896114aa565b84546040517f4460790300000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8084166004830152929350911690634460790390602401602060405180830381865afa158015610f0b573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610f2f91906121a2565b610f65576040517f8baa579f00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b506001979650505050505050565b610f7b611201565b610f84816111b8565b7f89639f446056f5d7661bbd94e8ab0617a80058ed7b072845818d4b93332e480080546040517f6b8c01a600000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff90911690636b8c01a690610ffa9086906004016121bd565b5f604051808303815f87803b158015611011575f5ffd5b505af11580156108f6573d5f5f3e3d5ffd5b61102c816111b8565b7f89639f446056f5d7661bbd94e8ab0617a80058ed7b072845818d4b93332e480080546040517f0b1c4cde00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff90911690630b1c4cde906110a890899089908990899060040161223b565b5f604051808303815f87803b1580156110bf575f5ffd5b505af11580156110d1573d5f5f3e3d5ffd5b50505050505050505050565b5f807f237e158222e3e6968b72b9db0d8043aacf074ad9f650f0d1606b4d82ee432c00610bd4565b61110d611201565b611137817f0f4ac8aae5a4fa6a3612928fcd8255b475ff86b500ae30bb272e61542cfc6f006103a3565b1561116e576040517f3af3c41c00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b610a28816114d2565b60605f61118383611589565b9392505050565b73ffffffffffffffffffffffffffffffffffffffff81165f9081526001830160205260408120541515611183565b5f8180156111c8576111c861226c565b14610a2857806040517fdf84035d00000000000000000000000000000000000000000000000000000000815260040161044e9190612299565b3361120a610bb0565b73ffffffffffffffffffffffffffffffffffffffff1614610a3c576040517f118cdaa700000000000000000000000000000000000000000000000000000000815233600482015260240161044e565b5f807ff0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a006103a9565b6112896115e2565b610a2881611620565b61129a6115e2565b610d618282611631565b5f6103a9825490565b5f6111838373ffffffffffffffffffffffffffffffffffffffff84166116a3565b7f237e158222e3e6968b72b9db0d8043aacf074ad9f650f0d1606b4d82ee432c0080547fffffffffffffffffffffffff0000000000000000000000000000000000000000168155610d6182611786565b7fa16a46d94261c7517cc8ff89f61c0ce93598e3c849801011dee649a6a557d10280546060917fa16a46d94261c7517cc8ff89f61c0ce93598e3c849801011dee649a6a557d1009161136f906122d8565b80601f016020809104026020016040519081016040528092919081815260200182805461139b906122d8565b80156113e65780601f106113bd576101008083540402835291602001916113e6565b820191905f5260205f20905b8154815290600101906020018083116113c957829003601f168201915b505050505091505090565b7fa16a46d94261c7517cc8ff89f61c0ce93598e3c849801011dee649a6a557d10380546060917fa16a46d94261c7517cc8ff89f61c0ce93598e3c849801011dee649a6a557d1009161136f906122d8565b5f6111838373ffffffffffffffffffffffffffffffffffffffff841661181b565b5f6103a961146f611867565b836040517f19010000000000000000000000000000000000000000000000000000000000008152600281019290925260228201526042902090565b5f5f5f5f6114b88686611870565b9250925092506114c882826118b9565b5090949350505050565b6114da611201565b7f237e158222e3e6968b72b9db0d8043aacf074ad9f650f0d1606b4d82ee432c0080547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff83169081178255611543610bb0565b73ffffffffffffffffffffffffffffffffffffffff167f38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e2270060405160405180910390a35050565b6060815f018054806020026020016040519081016040528092919081815260200182805480156115d657602002820191905f5260205f20905b8154815260200190600101908083116115c2575b50505050509050919050565b6115ea6119bc565b610a3c576040517fd7e6bcf800000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6116286115e2565b610a28816119da565b6116396115e2565b7fa16a46d94261c7517cc8ff89f61c0ce93598e3c849801011dee649a6a557d1007fa16a46d94261c7517cc8ff89f61c0ce93598e3c849801011dee649a6a557d1026116858482612372565b50600381016116948382612372565b505f8082556001909101555050565b5f818152600183016020526040812054801561177d575f6116c5600183612489565b85549091505f906116d890600190612489565b9050808214611737575f865f0182815481106116f6576116f66124c1565b905f5260205f200154905080875f018481548110611716576117166124c1565b5f918252602080832090910192909255918252600188019052604090208390555b8554869080611748576117486124ee565b600190038181905f5260205f20015f90559055856001015f8681526020019081526020015f205f9055600193505050506103a9565b5f9150506103a9565b7f9016d09d72d40fdae2fd8ceac6b6234c7706214fd39c1cd1e609a0528c19930080547fffffffffffffffffffffffff0000000000000000000000000000000000000000811673ffffffffffffffffffffffffffffffffffffffff848116918217845560405192169182907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0905f90a3505050565b5f81815260018301602052604081205461186057508154600181810184555f8481526020808220909301849055845484825282860190935260409020919091556103a9565b505f6103a9565b5f610377611a31565b5f5f5f83516041036118a7576020840151604085015160608601515f1a61189988828585611aa4565b9550955095505050506118b2565b505081515f91506002905b9250925092565b5f8260038111156118cc576118cc61226c565b036118d5575050565b60018260038111156118e9576118e961226c565b03611920576040517ff645eedf00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60028260038111156119345761193461226c565b0361196e576040517ffce698f70000000000000000000000000000000000000000000000000000000081526004810182905260240161044e565b60038260038111156119825761198261226c565b03610d61576040517fd78bce0c0000000000000000000000000000000000000000000000000000000081526004810182905260240161044e565b5f6119c5611259565b5468010000000000000000900460ff16919050565b6119e26115e2565b73ffffffffffffffffffffffffffffffffffffffff8116610aad576040517f1e4fbdf70000000000000000000000000000000000000000000000000000000081525f600482015260240161044e565b5f7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f611a5b611b97565b611a63611c12565b60408051602081019490945283019190915260608201524660808201523060a082015260c00160405160208183030381529060405280519060200120905090565b5f80807f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0841115611add57505f91506003905082611b8d565b604080515f808252602082018084528a905260ff891692820192909252606081018790526080810186905260019060a0016020604051602081039080840390855afa158015611b2e573d5f5f3e3d5ffd5b50506040517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0015191505073ffffffffffffffffffffffffffffffffffffffff8116611b8457505f925060019150829050611b8d565b92505f91508190505b9450945094915050565b5f7fa16a46d94261c7517cc8ff89f61c0ce93598e3c849801011dee649a6a557d10081611bc261131e565b805190915015611bda57805160209091012092915050565b81548015611be9579392505050565b7fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470935050505090565b5f7fa16a46d94261c7517cc8ff89f61c0ce93598e3c849801011dee649a6a557d10081611c3d6113f1565b805190915015611c5557805160209091012092915050565b60018201548015611be9579392505050565b602080825282518282018190525f918401906040840190835b81811015611cb457835173ffffffffffffffffffffffffffffffffffffffff16835260209384019390920191600101611c80565b509095945050505050565b73ffffffffffffffffffffffffffffffffffffffff81168114610a28575f5ffd5b5f60208284031215611cf0575f5ffd5b813561118381611cbf565b8015158114610a28575f5ffd5b803560018110611d16575f5ffd5b919050565b5f5f5f60608486031215611d2d575f5ffd5b833592506020840135611d3f81611cfb565b9150611d4d60408501611d08565b90509250925092565b5f5f60408385031215611d67575f5ffd5b82359150611d7760208401611d08565b90509250929050565b5f5f60408385031215611d91575f5ffd5b8235611d9c81611cbf565b91506020830135611dac81611cbf565b809150509250929050565b5f5f60408385031215611dc8575f5ffd5b8235611dd381611cbf565b9150611d7760208401611d08565b5f81518084528060208401602086015e5f6020828601015260207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f83011685010191505092915050565b7fff000000000000000000000000000000000000000000000000000000000000008816815260e060208201525f611e6760e0830189611de1565b8281036040840152611e798189611de1565b6060840188905273ffffffffffffffffffffffffffffffffffffffff8716608085015260a0840186905283810360c0850152845180825260208087019350909101905f5b81811015611edb578351835260209384019390920191600101611ebd565b50909b9a5050505050505050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b604051601f82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016810167ffffffffffffffff81118282101715611f6057611f60611eec565b604052919050565b5f5f5f60608486031215611f7a575f5ffd5b833567ffffffffffffffff811115611f90575f5ffd5b8401601f81018613611fa0575f5ffd5b803567ffffffffffffffff811115611fba57611fba611eec565b611feb60207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f84011601611f19565b818152876020838501011115611fff575f5ffd5b816020840160208301375f602092820183015294508501359250611d4d905060408501611d08565b5f5f60408385031215612038575f5ffd5b823567ffffffffffffffff81111561204e575f5ffd5b8301601f8101851361205e575f5ffd5b803567ffffffffffffffff81111561207857612078611eec565b8060051b61208860208201611f19565b918252602081840181019290810190888411156120a3575f5ffd5b6020850194505b838510156120c9578435808352602095860195909350909101906120aa565b8096505050505050611d7760208401611d08565b5f5f83601f8401126120ed575f5ffd5b50813567ffffffffffffffff811115612104575f5ffd5b60208301915083602082850101111561211b575f5ffd5b9250929050565b5f5f5f5f5f60608688031215612136575f5ffd5b853567ffffffffffffffff81111561214c575f5ffd5b612158888289016120dd565b909650945050602086013567ffffffffffffffff811115612177575f5ffd5b612183888289016120dd565b9094509250612196905060408701611d08565b90509295509295909350565b5f602082840312156121b2575f5ffd5b815161118381611cfb565b602080825282518282018190525f918401906040840190835b81811015611cb45783518352602093840193909201916001016121d6565b81835281816020850137505f602082840101525f60207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f840116840101905092915050565b604081525f61224e6040830186886121f4565b82810360208401526122618185876121f4565b979650505050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffd5b60208101600183106122d2577f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffd5b91905290565b600181811c908216806122ec57607f821691505b602082108103612323577f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b50919050565b601f82111561236d57805f5260205f20601f840160051c8101602085101561234e5750805b601f840160051c820191505b81811015610e08575f815560010161235a565b505050565b815167ffffffffffffffff81111561238c5761238c611eec565b6123a08161239a84546122d8565b84612329565b6020601f8211600181146123f1575f83156123bb5750848201515b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff600385901b1c1916600184901b178455610e08565b5f848152602081207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08516915b8281101561243e578785015182556020948501946001909201910161241e565b508482101561247a57868401517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff600387901b60f8161c191681555b50505050600190811b01905550565b818103818111156103a9577f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603160045260245ffdfea164736f6c634300081c000a",
}

// EspressoTEEVerifierABI is the input ABI used to generate the binding from.
// Deprecated: Use EspressoTEEVerifierMetaData.ABI instead.
var EspressoTEEVerifierABI = EspressoTEEVerifierMetaData.ABI

// EspressoTEEVerifierBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use EspressoTEEVerifierMetaData.Bin instead.
var EspressoTEEVerifierBin = EspressoTEEVerifierMetaData.Bin

// DeployEspressoTEEVerifier deploys a new Ethereum contract, binding an instance of EspressoTEEVerifier to it.
func DeployEspressoTEEVerifier(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *EspressoTEEVerifier, error) {
	parsed, err := EspressoTEEVerifierMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(EspressoTEEVerifierBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &EspressoTEEVerifier{EspressoTEEVerifierCaller: EspressoTEEVerifierCaller{contract: contract}, EspressoTEEVerifierTransactor: EspressoTEEVerifierTransactor{contract: contract}, EspressoTEEVerifierFilterer: EspressoTEEVerifierFilterer{contract: contract}}, nil
}

// EspressoTEEVerifier is an auto generated Go binding around an Ethereum contract.
type EspressoTEEVerifier struct {
	EspressoTEEVerifierCaller     // Read-only binding to the contract
	EspressoTEEVerifierTransactor // Write-only binding to the contract
	EspressoTEEVerifierFilterer   // Log filterer for contract events
}

// EspressoTEEVerifierCaller is an auto generated read-only Go binding around an Ethereum contract.
type EspressoTEEVerifierCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EspressoTEEVerifierTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EspressoTEEVerifierTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EspressoTEEVerifierFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EspressoTEEVerifierFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EspressoTEEVerifierSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EspressoTEEVerifierSession struct {
	Contract     *EspressoTEEVerifier // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// EspressoTEEVerifierCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EspressoTEEVerifierCallerSession struct {
	Contract *EspressoTEEVerifierCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// EspressoTEEVerifierTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EspressoTEEVerifierTransactorSession struct {
	Contract     *EspressoTEEVerifierTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// EspressoTEEVerifierRaw is an auto generated low-level Go binding around an Ethereum contract.
type EspressoTEEVerifierRaw struct {
	Contract *EspressoTEEVerifier // Generic contract binding to access the raw methods on
}

// EspressoTEEVerifierCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EspressoTEEVerifierCallerRaw struct {
	Contract *EspressoTEEVerifierCaller // Generic read-only contract binding to access the raw methods on
}

// EspressoTEEVerifierTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EspressoTEEVerifierTransactorRaw struct {
	Contract *EspressoTEEVerifierTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEspressoTEEVerifier creates a new instance of EspressoTEEVerifier, bound to a specific deployed contract.
func NewEspressoTEEVerifier(address common.Address, backend bind.ContractBackend) (*EspressoTEEVerifier, error) {
	contract, err := bindEspressoTEEVerifier(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifier{EspressoTEEVerifierCaller: EspressoTEEVerifierCaller{contract: contract}, EspressoTEEVerifierTransactor: EspressoTEEVerifierTransactor{contract: contract}, EspressoTEEVerifierFilterer: EspressoTEEVerifierFilterer{contract: contract}}, nil
}

// NewEspressoTEEVerifierCaller creates a new read-only instance of EspressoTEEVerifier, bound to a specific deployed contract.
func NewEspressoTEEVerifierCaller(address common.Address, caller bind.ContractCaller) (*EspressoTEEVerifierCaller, error) {
	contract, err := bindEspressoTEEVerifier(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifierCaller{contract: contract}, nil
}

// NewEspressoTEEVerifierTransactor creates a new write-only instance of EspressoTEEVerifier, bound to a specific deployed contract.
func NewEspressoTEEVerifierTransactor(address common.Address, transactor bind.ContractTransactor) (*EspressoTEEVerifierTransactor, error) {
	contract, err := bindEspressoTEEVerifier(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifierTransactor{contract: contract}, nil
}

// NewEspressoTEEVerifierFilterer creates a new log filterer instance of EspressoTEEVerifier, bound to a specific deployed contract.
func NewEspressoTEEVerifierFilterer(address common.Address, filterer bind.ContractFilterer) (*EspressoTEEVerifierFilterer, error) {
	contract, err := bindEspressoTEEVerifier(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifierFilterer{contract: contract}, nil
}

// bindEspressoTEEVerifier binds a generic wrapper to an already deployed contract.
func bindEspressoTEEVerifier(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := EspressoTEEVerifierMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EspressoTEEVerifier *EspressoTEEVerifierRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _EspressoTEEVerifier.Contract.EspressoTEEVerifierCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EspressoTEEVerifier *EspressoTEEVerifierRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.EspressoTEEVerifierTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EspressoTEEVerifier *EspressoTEEVerifierRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.EspressoTEEVerifierTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _EspressoTEEVerifier.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.contract.Transact(opts, method, params...)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_EspressoTEEVerifier *EspressoTEEVerifierCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _EspressoTEEVerifier.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _EspressoTEEVerifier.Contract.Eip712Domain(&_EspressoTEEVerifier.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _EspressoTEEVerifier.Contract.Eip712Domain(&_EspressoTEEVerifier.CallOpts)
}

// EspressoNitroTEEVerifier is a free data retrieval call binding the contract method 0xd80a4c28.
//
// Solidity: function espressoNitroTEEVerifier() view returns(address)
func (_EspressoTEEVerifier *EspressoTEEVerifierCaller) EspressoNitroTEEVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _EspressoTEEVerifier.contract.Call(opts, &out, "espressoNitroTEEVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EspressoNitroTEEVerifier is a free data retrieval call binding the contract method 0xd80a4c28.
//
// Solidity: function espressoNitroTEEVerifier() view returns(address)
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) EspressoNitroTEEVerifier() (common.Address, error) {
	return _EspressoTEEVerifier.Contract.EspressoNitroTEEVerifier(&_EspressoTEEVerifier.CallOpts)
}

// EspressoNitroTEEVerifier is a free data retrieval call binding the contract method 0xd80a4c28.
//
// Solidity: function espressoNitroTEEVerifier() view returns(address)
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerSession) EspressoNitroTEEVerifier() (common.Address, error) {
	return _EspressoTEEVerifier.Contract.EspressoNitroTEEVerifier(&_EspressoTEEVerifier.CallOpts)
}

// GetGuardians is a free data retrieval call binding the contract method 0x0665f04b.
//
// Solidity: function getGuardians() view returns(address[])
func (_EspressoTEEVerifier *EspressoTEEVerifierCaller) GetGuardians(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _EspressoTEEVerifier.contract.Call(opts, &out, "getGuardians")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetGuardians is a free data retrieval call binding the contract method 0x0665f04b.
//
// Solidity: function getGuardians() view returns(address[])
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) GetGuardians() ([]common.Address, error) {
	return _EspressoTEEVerifier.Contract.GetGuardians(&_EspressoTEEVerifier.CallOpts)
}

// GetGuardians is a free data retrieval call binding the contract method 0x0665f04b.
//
// Solidity: function getGuardians() view returns(address[])
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerSession) GetGuardians() ([]common.Address, error) {
	return _EspressoTEEVerifier.Contract.GetGuardians(&_EspressoTEEVerifier.CallOpts)
}

// GuardianCount is a free data retrieval call binding the contract method 0x54387ad7.
//
// Solidity: function guardianCount() view returns(uint256)
func (_EspressoTEEVerifier *EspressoTEEVerifierCaller) GuardianCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _EspressoTEEVerifier.contract.Call(opts, &out, "guardianCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GuardianCount is a free data retrieval call binding the contract method 0x54387ad7.
//
// Solidity: function guardianCount() view returns(uint256)
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) GuardianCount() (*big.Int, error) {
	return _EspressoTEEVerifier.Contract.GuardianCount(&_EspressoTEEVerifier.CallOpts)
}

// GuardianCount is a free data retrieval call binding the contract method 0x54387ad7.
//
// Solidity: function guardianCount() view returns(uint256)
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerSession) GuardianCount() (*big.Int, error) {
	return _EspressoTEEVerifier.Contract.GuardianCount(&_EspressoTEEVerifier.CallOpts)
}

// IsGuardian is a free data retrieval call binding the contract method 0x0c68ba21.
//
// Solidity: function isGuardian(address account) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierCaller) IsGuardian(opts *bind.CallOpts, account common.Address) (bool, error) {
	var out []interface{}
	err := _EspressoTEEVerifier.contract.Call(opts, &out, "isGuardian", account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsGuardian is a free data retrieval call binding the contract method 0x0c68ba21.
//
// Solidity: function isGuardian(address account) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) IsGuardian(account common.Address) (bool, error) {
	return _EspressoTEEVerifier.Contract.IsGuardian(&_EspressoTEEVerifier.CallOpts, account)
}

// IsGuardian is a free data retrieval call binding the contract method 0x0c68ba21.
//
// Solidity: function isGuardian(address account) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerSession) IsGuardian(account common.Address) (bool, error) {
	return _EspressoTEEVerifier.Contract.IsGuardian(&_EspressoTEEVerifier.CallOpts, account)
}

// IsSignerValid is a free data retrieval call binding the contract method 0x6d8f5aa9.
//
// Solidity: function isSignerValid(address signer, uint8 teeType) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierCaller) IsSignerValid(opts *bind.CallOpts, signer common.Address, teeType uint8) (bool, error) {
	var out []interface{}
	err := _EspressoTEEVerifier.contract.Call(opts, &out, "isSignerValid", signer, teeType)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsSignerValid is a free data retrieval call binding the contract method 0x6d8f5aa9.
//
// Solidity: function isSignerValid(address signer, uint8 teeType) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) IsSignerValid(signer common.Address, teeType uint8) (bool, error) {
	return _EspressoTEEVerifier.Contract.IsSignerValid(&_EspressoTEEVerifier.CallOpts, signer, teeType)
}

// IsSignerValid is a free data retrieval call binding the contract method 0x6d8f5aa9.
//
// Solidity: function isSignerValid(address signer, uint8 teeType) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerSession) IsSignerValid(signer common.Address, teeType uint8) (bool, error) {
	return _EspressoTEEVerifier.Contract.IsSignerValid(&_EspressoTEEVerifier.CallOpts, signer, teeType)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_EspressoTEEVerifier *EspressoTEEVerifierCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _EspressoTEEVerifier.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) Owner() (common.Address, error) {
	return _EspressoTEEVerifier.Contract.Owner(&_EspressoTEEVerifier.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerSession) Owner() (common.Address, error) {
	return _EspressoTEEVerifier.Contract.Owner(&_EspressoTEEVerifier.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_EspressoTEEVerifier *EspressoTEEVerifierCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _EspressoTEEVerifier.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) PendingOwner() (common.Address, error) {
	return _EspressoTEEVerifier.Contract.PendingOwner(&_EspressoTEEVerifier.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerSession) PendingOwner() (common.Address, error) {
	return _EspressoTEEVerifier.Contract.PendingOwner(&_EspressoTEEVerifier.CallOpts)
}

// RegisteredEnclaveHashes is a free data retrieval call binding the contract method 0x3cbe6803.
//
// Solidity: function registeredEnclaveHashes(bytes32 enclaveHash, uint8 teeType) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierCaller) RegisteredEnclaveHashes(opts *bind.CallOpts, enclaveHash [32]byte, teeType uint8) (bool, error) {
	var out []interface{}
	err := _EspressoTEEVerifier.contract.Call(opts, &out, "registeredEnclaveHashes", enclaveHash, teeType)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// RegisteredEnclaveHashes is a free data retrieval call binding the contract method 0x3cbe6803.
//
// Solidity: function registeredEnclaveHashes(bytes32 enclaveHash, uint8 teeType) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) RegisteredEnclaveHashes(enclaveHash [32]byte, teeType uint8) (bool, error) {
	return _EspressoTEEVerifier.Contract.RegisteredEnclaveHashes(&_EspressoTEEVerifier.CallOpts, enclaveHash, teeType)
}

// RegisteredEnclaveHashes is a free data retrieval call binding the contract method 0x3cbe6803.
//
// Solidity: function registeredEnclaveHashes(bytes32 enclaveHash, uint8 teeType) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerSession) RegisteredEnclaveHashes(enclaveHash [32]byte, teeType uint8) (bool, error) {
	return _EspressoTEEVerifier.Contract.RegisteredEnclaveHashes(&_EspressoTEEVerifier.CallOpts, enclaveHash, teeType)
}

// Verify is a free data retrieval call binding the contract method 0xa81d9c5c.
//
// Solidity: function verify(bytes signature, bytes32 userDataHash, uint8 teeType) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierCaller) Verify(opts *bind.CallOpts, signature []byte, userDataHash [32]byte, teeType uint8) (bool, error) {
	var out []interface{}
	err := _EspressoTEEVerifier.contract.Call(opts, &out, "verify", signature, userDataHash, teeType)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Verify is a free data retrieval call binding the contract method 0xa81d9c5c.
//
// Solidity: function verify(bytes signature, bytes32 userDataHash, uint8 teeType) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) Verify(signature []byte, userDataHash [32]byte, teeType uint8) (bool, error) {
	return _EspressoTEEVerifier.Contract.Verify(&_EspressoTEEVerifier.CallOpts, signature, userDataHash, teeType)
}

// Verify is a free data retrieval call binding the contract method 0xa81d9c5c.
//
// Solidity: function verify(bytes signature, bytes32 userDataHash, uint8 teeType) view returns(bool)
func (_EspressoTEEVerifier *EspressoTEEVerifierCallerSession) Verify(signature []byte, userDataHash [32]byte, teeType uint8) (bool, error) {
	return _EspressoTEEVerifier.Contract.Verify(&_EspressoTEEVerifier.CallOpts, signature, userDataHash, teeType)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) AcceptOwnership() (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.AcceptOwnership(&_EspressoTEEVerifier.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.AcceptOwnership(&_EspressoTEEVerifier.TransactOpts)
}

// AddGuardian is a paid mutator transaction binding the contract method 0xa526d83b.
//
// Solidity: function addGuardian(address guardian) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) AddGuardian(opts *bind.TransactOpts, guardian common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "addGuardian", guardian)
}

// AddGuardian is a paid mutator transaction binding the contract method 0xa526d83b.
//
// Solidity: function addGuardian(address guardian) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) AddGuardian(guardian common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.AddGuardian(&_EspressoTEEVerifier.TransactOpts, guardian)
}

// AddGuardian is a paid mutator transaction binding the contract method 0xa526d83b.
//
// Solidity: function addGuardian(address guardian) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) AddGuardian(guardian common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.AddGuardian(&_EspressoTEEVerifier.TransactOpts, guardian)
}

// DeleteEnclaveHashes is a paid mutator transaction binding the contract method 0xcd8f6997.
//
// Solidity: function deleteEnclaveHashes(bytes32[] enclaveHashes, uint8 teeType) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) DeleteEnclaveHashes(opts *bind.TransactOpts, enclaveHashes [][32]byte, teeType uint8) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "deleteEnclaveHashes", enclaveHashes, teeType)
}

// DeleteEnclaveHashes is a paid mutator transaction binding the contract method 0xcd8f6997.
//
// Solidity: function deleteEnclaveHashes(bytes32[] enclaveHashes, uint8 teeType) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) DeleteEnclaveHashes(enclaveHashes [][32]byte, teeType uint8) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.DeleteEnclaveHashes(&_EspressoTEEVerifier.TransactOpts, enclaveHashes, teeType)
}

// DeleteEnclaveHashes is a paid mutator transaction binding the contract method 0xcd8f6997.
//
// Solidity: function deleteEnclaveHashes(bytes32[] enclaveHashes, uint8 teeType) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) DeleteEnclaveHashes(enclaveHashes [][32]byte, teeType uint8) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.DeleteEnclaveHashes(&_EspressoTEEVerifier.TransactOpts, enclaveHashes, teeType)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _owner, address _espressoNitroTEEVerifier) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) Initialize(opts *bind.TransactOpts, _owner common.Address, _espressoNitroTEEVerifier common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "initialize", _owner, _espressoNitroTEEVerifier)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _owner, address _espressoNitroTEEVerifier) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) Initialize(_owner common.Address, _espressoNitroTEEVerifier common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.Initialize(&_EspressoTEEVerifier.TransactOpts, _owner, _espressoNitroTEEVerifier)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _owner, address _espressoNitroTEEVerifier) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) Initialize(_owner common.Address, _espressoNitroTEEVerifier common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.Initialize(&_EspressoTEEVerifier.TransactOpts, _owner, _espressoNitroTEEVerifier)
}

// RegisterService is a paid mutator transaction binding the contract method 0xdac79fc8.
//
// Solidity: function registerService(bytes verificationData, bytes data, uint8 teeType) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) RegisterService(opts *bind.TransactOpts, verificationData []byte, data []byte, teeType uint8) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "registerService", verificationData, data, teeType)
}

// RegisterService is a paid mutator transaction binding the contract method 0xdac79fc8.
//
// Solidity: function registerService(bytes verificationData, bytes data, uint8 teeType) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) RegisterService(verificationData []byte, data []byte, teeType uint8) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.RegisterService(&_EspressoTEEVerifier.TransactOpts, verificationData, data, teeType)
}

// RegisterService is a paid mutator transaction binding the contract method 0xdac79fc8.
//
// Solidity: function registerService(bytes verificationData, bytes data, uint8 teeType) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) RegisterService(verificationData []byte, data []byte, teeType uint8) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.RegisterService(&_EspressoTEEVerifier.TransactOpts, verificationData, data, teeType)
}

// RemoveGuardian is a paid mutator transaction binding the contract method 0x71404156.
//
// Solidity: function removeGuardian(address guardian) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) RemoveGuardian(opts *bind.TransactOpts, guardian common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "removeGuardian", guardian)
}

// RemoveGuardian is a paid mutator transaction binding the contract method 0x71404156.
//
// Solidity: function removeGuardian(address guardian) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) RemoveGuardian(guardian common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.RemoveGuardian(&_EspressoTEEVerifier.TransactOpts, guardian)
}

// RemoveGuardian is a paid mutator transaction binding the contract method 0x71404156.
//
// Solidity: function removeGuardian(address guardian) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) RemoveGuardian(guardian common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.RemoveGuardian(&_EspressoTEEVerifier.TransactOpts, guardian)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) RenounceOwnership() (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.RenounceOwnership(&_EspressoTEEVerifier.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.RenounceOwnership(&_EspressoTEEVerifier.TransactOpts)
}

// SetEnclaveHash is a paid mutator transaction binding the contract method 0x0f1f0f86.
//
// Solidity: function setEnclaveHash(bytes32 enclaveHash, bool valid, uint8 teeType) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) SetEnclaveHash(opts *bind.TransactOpts, enclaveHash [32]byte, valid bool, teeType uint8) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "setEnclaveHash", enclaveHash, valid, teeType)
}

// SetEnclaveHash is a paid mutator transaction binding the contract method 0x0f1f0f86.
//
// Solidity: function setEnclaveHash(bytes32 enclaveHash, bool valid, uint8 teeType) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) SetEnclaveHash(enclaveHash [32]byte, valid bool, teeType uint8) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.SetEnclaveHash(&_EspressoTEEVerifier.TransactOpts, enclaveHash, valid, teeType)
}

// SetEnclaveHash is a paid mutator transaction binding the contract method 0x0f1f0f86.
//
// Solidity: function setEnclaveHash(bytes32 enclaveHash, bool valid, uint8 teeType) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) SetEnclaveHash(enclaveHash [32]byte, valid bool, teeType uint8) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.SetEnclaveHash(&_EspressoTEEVerifier.TransactOpts, enclaveHash, valid, teeType)
}

// SetEspressoNitroTEEVerifier is a paid mutator transaction binding the contract method 0x330282f5.
//
// Solidity: function setEspressoNitroTEEVerifier(address _espressoNitroTEEVerifier) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) SetEspressoNitroTEEVerifier(opts *bind.TransactOpts, _espressoNitroTEEVerifier common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "setEspressoNitroTEEVerifier", _espressoNitroTEEVerifier)
}

// SetEspressoNitroTEEVerifier is a paid mutator transaction binding the contract method 0x330282f5.
//
// Solidity: function setEspressoNitroTEEVerifier(address _espressoNitroTEEVerifier) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) SetEspressoNitroTEEVerifier(_espressoNitroTEEVerifier common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.SetEspressoNitroTEEVerifier(&_EspressoTEEVerifier.TransactOpts, _espressoNitroTEEVerifier)
}

// SetEspressoNitroTEEVerifier is a paid mutator transaction binding the contract method 0x330282f5.
//
// Solidity: function setEspressoNitroTEEVerifier(address _espressoNitroTEEVerifier) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) SetEspressoNitroTEEVerifier(_espressoNitroTEEVerifier common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.SetEspressoNitroTEEVerifier(&_EspressoTEEVerifier.TransactOpts, _espressoNitroTEEVerifier)
}

// SetNitroEnclaveVerifier is a paid mutator transaction binding the contract method 0xa628a19e.
//
// Solidity: function setNitroEnclaveVerifier(address nitroVerifier) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) SetNitroEnclaveVerifier(opts *bind.TransactOpts, nitroVerifier common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "setNitroEnclaveVerifier", nitroVerifier)
}

// SetNitroEnclaveVerifier is a paid mutator transaction binding the contract method 0xa628a19e.
//
// Solidity: function setNitroEnclaveVerifier(address nitroVerifier) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) SetNitroEnclaveVerifier(nitroVerifier common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.SetNitroEnclaveVerifier(&_EspressoTEEVerifier.TransactOpts, nitroVerifier)
}

// SetNitroEnclaveVerifier is a paid mutator transaction binding the contract method 0xa628a19e.
//
// Solidity: function setNitroEnclaveVerifier(address nitroVerifier) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) SetNitroEnclaveVerifier(nitroVerifier common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.SetNitroEnclaveVerifier(&_EspressoTEEVerifier.TransactOpts, nitroVerifier)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.TransferOwnership(&_EspressoTEEVerifier.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_EspressoTEEVerifier *EspressoTEEVerifierTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EspressoTEEVerifier.Contract.TransferOwnership(&_EspressoTEEVerifier.TransactOpts, newOwner)
}

// EspressoTEEVerifierEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierEIP712DomainChangedIterator struct {
	Event *EspressoTEEVerifierEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *EspressoTEEVerifierEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EspressoTEEVerifierEIP712DomainChanged)
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
		it.Event = new(EspressoTEEVerifierEIP712DomainChanged)
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
func (it *EspressoTEEVerifierEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EspressoTEEVerifierEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EspressoTEEVerifierEIP712DomainChanged represents a EIP712DomainChanged event raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*EspressoTEEVerifierEIP712DomainChangedIterator, error) {

	logs, sub, err := _EspressoTEEVerifier.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifierEIP712DomainChangedIterator{contract: _EspressoTEEVerifier.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *EspressoTEEVerifierEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _EspressoTEEVerifier.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EspressoTEEVerifierEIP712DomainChanged)
				if err := _EspressoTEEVerifier.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) ParseEIP712DomainChanged(log types.Log) (*EspressoTEEVerifierEIP712DomainChanged, error) {
	event := new(EspressoTEEVerifierEIP712DomainChanged)
	if err := _EspressoTEEVerifier.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EspressoTEEVerifierEspressoNitroTEEVerifierSetIterator is returned from FilterEspressoNitroTEEVerifierSet and is used to iterate over the raw logs and unpacked data for EspressoNitroTEEVerifierSet events raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierEspressoNitroTEEVerifierSetIterator struct {
	Event *EspressoTEEVerifierEspressoNitroTEEVerifierSet // Event containing the contract specifics and raw log

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
func (it *EspressoTEEVerifierEspressoNitroTEEVerifierSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EspressoTEEVerifierEspressoNitroTEEVerifierSet)
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
		it.Event = new(EspressoTEEVerifierEspressoNitroTEEVerifierSet)
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
func (it *EspressoTEEVerifierEspressoNitroTEEVerifierSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EspressoTEEVerifierEspressoNitroTEEVerifierSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EspressoTEEVerifierEspressoNitroTEEVerifierSet represents a EspressoNitroTEEVerifierSet event raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierEspressoNitroTEEVerifierSet struct {
	OldVerifier common.Address
	NewVerifier common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterEspressoNitroTEEVerifierSet is a free log retrieval operation binding the contract event 0xacbe6384d204f2883b25fa4d03777cb88af5b2afaaffbe06609d3e17dc68d096.
//
// Solidity: event EspressoNitroTEEVerifierSet(address indexed oldVerifier, address indexed newVerifier)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) FilterEspressoNitroTEEVerifierSet(opts *bind.FilterOpts, oldVerifier []common.Address, newVerifier []common.Address) (*EspressoTEEVerifierEspressoNitroTEEVerifierSetIterator, error) {

	var oldVerifierRule []interface{}
	for _, oldVerifierItem := range oldVerifier {
		oldVerifierRule = append(oldVerifierRule, oldVerifierItem)
	}
	var newVerifierRule []interface{}
	for _, newVerifierItem := range newVerifier {
		newVerifierRule = append(newVerifierRule, newVerifierItem)
	}

	logs, sub, err := _EspressoTEEVerifier.contract.FilterLogs(opts, "EspressoNitroTEEVerifierSet", oldVerifierRule, newVerifierRule)
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifierEspressoNitroTEEVerifierSetIterator{contract: _EspressoTEEVerifier.contract, event: "EspressoNitroTEEVerifierSet", logs: logs, sub: sub}, nil
}

// WatchEspressoNitroTEEVerifierSet is a free log subscription operation binding the contract event 0xacbe6384d204f2883b25fa4d03777cb88af5b2afaaffbe06609d3e17dc68d096.
//
// Solidity: event EspressoNitroTEEVerifierSet(address indexed oldVerifier, address indexed newVerifier)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) WatchEspressoNitroTEEVerifierSet(opts *bind.WatchOpts, sink chan<- *EspressoTEEVerifierEspressoNitroTEEVerifierSet, oldVerifier []common.Address, newVerifier []common.Address) (event.Subscription, error) {

	var oldVerifierRule []interface{}
	for _, oldVerifierItem := range oldVerifier {
		oldVerifierRule = append(oldVerifierRule, oldVerifierItem)
	}
	var newVerifierRule []interface{}
	for _, newVerifierItem := range newVerifier {
		newVerifierRule = append(newVerifierRule, newVerifierItem)
	}

	logs, sub, err := _EspressoTEEVerifier.contract.WatchLogs(opts, "EspressoNitroTEEVerifierSet", oldVerifierRule, newVerifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EspressoTEEVerifierEspressoNitroTEEVerifierSet)
				if err := _EspressoTEEVerifier.contract.UnpackLog(event, "EspressoNitroTEEVerifierSet", log); err != nil {
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

// ParseEspressoNitroTEEVerifierSet is a log parse operation binding the contract event 0xacbe6384d204f2883b25fa4d03777cb88af5b2afaaffbe06609d3e17dc68d096.
//
// Solidity: event EspressoNitroTEEVerifierSet(address indexed oldVerifier, address indexed newVerifier)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) ParseEspressoNitroTEEVerifierSet(log types.Log) (*EspressoTEEVerifierEspressoNitroTEEVerifierSet, error) {
	event := new(EspressoTEEVerifierEspressoNitroTEEVerifierSet)
	if err := _EspressoTEEVerifier.contract.UnpackLog(event, "EspressoNitroTEEVerifierSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EspressoTEEVerifierGuardianAddedIterator is returned from FilterGuardianAdded and is used to iterate over the raw logs and unpacked data for GuardianAdded events raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierGuardianAddedIterator struct {
	Event *EspressoTEEVerifierGuardianAdded // Event containing the contract specifics and raw log

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
func (it *EspressoTEEVerifierGuardianAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EspressoTEEVerifierGuardianAdded)
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
		it.Event = new(EspressoTEEVerifierGuardianAdded)
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
func (it *EspressoTEEVerifierGuardianAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EspressoTEEVerifierGuardianAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EspressoTEEVerifierGuardianAdded represents a GuardianAdded event raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierGuardianAdded struct {
	Guardian common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterGuardianAdded is a free log retrieval operation binding the contract event 0x038596bb31e2e7d3d9f184d4c98b310103f6d7f5830e5eec32bffe6f1728f969.
//
// Solidity: event GuardianAdded(address indexed guardian)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) FilterGuardianAdded(opts *bind.FilterOpts, guardian []common.Address) (*EspressoTEEVerifierGuardianAddedIterator, error) {

	var guardianRule []interface{}
	for _, guardianItem := range guardian {
		guardianRule = append(guardianRule, guardianItem)
	}

	logs, sub, err := _EspressoTEEVerifier.contract.FilterLogs(opts, "GuardianAdded", guardianRule)
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifierGuardianAddedIterator{contract: _EspressoTEEVerifier.contract, event: "GuardianAdded", logs: logs, sub: sub}, nil
}

// WatchGuardianAdded is a free log subscription operation binding the contract event 0x038596bb31e2e7d3d9f184d4c98b310103f6d7f5830e5eec32bffe6f1728f969.
//
// Solidity: event GuardianAdded(address indexed guardian)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) WatchGuardianAdded(opts *bind.WatchOpts, sink chan<- *EspressoTEEVerifierGuardianAdded, guardian []common.Address) (event.Subscription, error) {

	var guardianRule []interface{}
	for _, guardianItem := range guardian {
		guardianRule = append(guardianRule, guardianItem)
	}

	logs, sub, err := _EspressoTEEVerifier.contract.WatchLogs(opts, "GuardianAdded", guardianRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EspressoTEEVerifierGuardianAdded)
				if err := _EspressoTEEVerifier.contract.UnpackLog(event, "GuardianAdded", log); err != nil {
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

// ParseGuardianAdded is a log parse operation binding the contract event 0x038596bb31e2e7d3d9f184d4c98b310103f6d7f5830e5eec32bffe6f1728f969.
//
// Solidity: event GuardianAdded(address indexed guardian)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) ParseGuardianAdded(log types.Log) (*EspressoTEEVerifierGuardianAdded, error) {
	event := new(EspressoTEEVerifierGuardianAdded)
	if err := _EspressoTEEVerifier.contract.UnpackLog(event, "GuardianAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EspressoTEEVerifierGuardianRemovedIterator is returned from FilterGuardianRemoved and is used to iterate over the raw logs and unpacked data for GuardianRemoved events raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierGuardianRemovedIterator struct {
	Event *EspressoTEEVerifierGuardianRemoved // Event containing the contract specifics and raw log

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
func (it *EspressoTEEVerifierGuardianRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EspressoTEEVerifierGuardianRemoved)
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
		it.Event = new(EspressoTEEVerifierGuardianRemoved)
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
func (it *EspressoTEEVerifierGuardianRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EspressoTEEVerifierGuardianRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EspressoTEEVerifierGuardianRemoved represents a GuardianRemoved event raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierGuardianRemoved struct {
	Guardian common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterGuardianRemoved is a free log retrieval operation binding the contract event 0xb8107d0c6b40be480ce3172ee66ba6d64b71f6b1685a851340036e6e2e3e3c52.
//
// Solidity: event GuardianRemoved(address indexed guardian)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) FilterGuardianRemoved(opts *bind.FilterOpts, guardian []common.Address) (*EspressoTEEVerifierGuardianRemovedIterator, error) {

	var guardianRule []interface{}
	for _, guardianItem := range guardian {
		guardianRule = append(guardianRule, guardianItem)
	}

	logs, sub, err := _EspressoTEEVerifier.contract.FilterLogs(opts, "GuardianRemoved", guardianRule)
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifierGuardianRemovedIterator{contract: _EspressoTEEVerifier.contract, event: "GuardianRemoved", logs: logs, sub: sub}, nil
}

// WatchGuardianRemoved is a free log subscription operation binding the contract event 0xb8107d0c6b40be480ce3172ee66ba6d64b71f6b1685a851340036e6e2e3e3c52.
//
// Solidity: event GuardianRemoved(address indexed guardian)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) WatchGuardianRemoved(opts *bind.WatchOpts, sink chan<- *EspressoTEEVerifierGuardianRemoved, guardian []common.Address) (event.Subscription, error) {

	var guardianRule []interface{}
	for _, guardianItem := range guardian {
		guardianRule = append(guardianRule, guardianItem)
	}

	logs, sub, err := _EspressoTEEVerifier.contract.WatchLogs(opts, "GuardianRemoved", guardianRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EspressoTEEVerifierGuardianRemoved)
				if err := _EspressoTEEVerifier.contract.UnpackLog(event, "GuardianRemoved", log); err != nil {
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

// ParseGuardianRemoved is a log parse operation binding the contract event 0xb8107d0c6b40be480ce3172ee66ba6d64b71f6b1685a851340036e6e2e3e3c52.
//
// Solidity: event GuardianRemoved(address indexed guardian)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) ParseGuardianRemoved(log types.Log) (*EspressoTEEVerifierGuardianRemoved, error) {
	event := new(EspressoTEEVerifierGuardianRemoved)
	if err := _EspressoTEEVerifier.contract.UnpackLog(event, "GuardianRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EspressoTEEVerifierInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierInitializedIterator struct {
	Event *EspressoTEEVerifierInitialized // Event containing the contract specifics and raw log

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
func (it *EspressoTEEVerifierInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EspressoTEEVerifierInitialized)
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
		it.Event = new(EspressoTEEVerifierInitialized)
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
func (it *EspressoTEEVerifierInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EspressoTEEVerifierInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EspressoTEEVerifierInitialized represents a Initialized event raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) FilterInitialized(opts *bind.FilterOpts) (*EspressoTEEVerifierInitializedIterator, error) {

	logs, sub, err := _EspressoTEEVerifier.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifierInitializedIterator{contract: _EspressoTEEVerifier.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *EspressoTEEVerifierInitialized) (event.Subscription, error) {

	logs, sub, err := _EspressoTEEVerifier.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EspressoTEEVerifierInitialized)
				if err := _EspressoTEEVerifier.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) ParseInitialized(log types.Log) (*EspressoTEEVerifierInitialized, error) {
	event := new(EspressoTEEVerifierInitialized)
	if err := _EspressoTEEVerifier.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EspressoTEEVerifierOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierOwnershipTransferStartedIterator struct {
	Event *EspressoTEEVerifierOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *EspressoTEEVerifierOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EspressoTEEVerifierOwnershipTransferStarted)
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
		it.Event = new(EspressoTEEVerifierOwnershipTransferStarted)
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
func (it *EspressoTEEVerifierOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EspressoTEEVerifierOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EspressoTEEVerifierOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*EspressoTEEVerifierOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EspressoTEEVerifier.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifierOwnershipTransferStartedIterator{contract: _EspressoTEEVerifier.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *EspressoTEEVerifierOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EspressoTEEVerifier.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EspressoTEEVerifierOwnershipTransferStarted)
				if err := _EspressoTEEVerifier.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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

// ParseOwnershipTransferStarted is a log parse operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) ParseOwnershipTransferStarted(log types.Log) (*EspressoTEEVerifierOwnershipTransferStarted, error) {
	event := new(EspressoTEEVerifierOwnershipTransferStarted)
	if err := _EspressoTEEVerifier.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EspressoTEEVerifierOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierOwnershipTransferredIterator struct {
	Event *EspressoTEEVerifierOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *EspressoTEEVerifierOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EspressoTEEVerifierOwnershipTransferred)
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
		it.Event = new(EspressoTEEVerifierOwnershipTransferred)
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
func (it *EspressoTEEVerifierOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EspressoTEEVerifierOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EspressoTEEVerifierOwnershipTransferred represents a OwnershipTransferred event raised by the EspressoTEEVerifier contract.
type EspressoTEEVerifierOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*EspressoTEEVerifierOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EspressoTEEVerifier.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &EspressoTEEVerifierOwnershipTransferredIterator{contract: _EspressoTEEVerifier.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *EspressoTEEVerifierOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EspressoTEEVerifier.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EspressoTEEVerifierOwnershipTransferred)
				if err := _EspressoTEEVerifier.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_EspressoTEEVerifier *EspressoTEEVerifierFilterer) ParseOwnershipTransferred(log types.Log) (*EspressoTEEVerifierOwnershipTransferred, error) {
	event := new(EspressoTEEVerifierOwnershipTransferred)
	if err := _EspressoTEEVerifier.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
