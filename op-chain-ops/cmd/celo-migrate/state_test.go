package main

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	accounts map[common.Address]*types.Account
}

func (db *mockDB) CreateAccount(addr common.Address) {
	account, ok := db.accounts[addr]
	db.accounts[addr] = &types.Account{
		Balance: big.NewInt(0),
	}
	if ok {
		db.accounts[addr].Balance = account.Balance
	}
}

func (db *mockDB) SubBalance(common.Address, *uint256.Int) {}

func (db *mockDB) AddBalance(common.Address, *uint256.Int) {}

func (db *mockDB) GetBalance(addr common.Address) *uint256.Int {
	account, ok := db.accounts[addr]
	if !ok {
		return common.U2560
	}
	return uint256.MustFromBig(account.Balance)
}

func (db *mockDB) GetNonce(addr common.Address) uint64 {
	account, ok := db.accounts[addr]
	if !ok {
		return 0
	}
	return account.Nonce
}

func (db *mockDB) SetNonce(addr common.Address, nonce uint64) {
	account, ok := db.accounts[addr]
	if !ok {
		return
	}
	account.Nonce = nonce
}

func (db *mockDB) GetCodeHash(common.Address) common.Hash {
	return common.Hash{}
}

func (db *mockDB) GetCode(addr common.Address) []byte {
	account, ok := db.accounts[addr]
	if !ok {
		return []byte{}
	}
	return account.Code
}

func (db *mockDB) SetCode(addr common.Address, code []byte) {
	account, ok := db.accounts[addr]
	if !ok {
		return
	}
	account.Code = code
}

func (db *mockDB) GetCodeSize(addr common.Address) int {
	account, ok := db.accounts[addr]
	if !ok {
		return 0
	}
	return len(account.Code)
}

func (db *mockDB) AddRefund(uint64) {}
func (db *mockDB) SubRefund(uint64) {}
func (db *mockDB) GetRefund() uint64 {
	return 0
}

func (db *mockDB) GetCommittedState(common.Address, common.Hash) common.Hash {
	return common.Hash{}
}
func (db *mockDB) GetState(common.Address, common.Hash) common.Hash {
	return common.Hash{}
}
func (db *mockDB) SetState(common.Address, common.Hash, common.Hash) {

}

func (db *mockDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	return common.Hash{}
}
func (db *mockDB) SetTransientState(addr common.Address, key, value common.Hash) {}

func (db *mockDB) SelfDestruct(common.Address) {}
func (db *mockDB) HasSelfDestructed(common.Address) bool {
	return false
}

func (db *mockDB) Selfdestruct6780(common.Address) {}

func (db *mockDB) Exist(addr common.Address) bool {
	_, exists := db.accounts[addr]
	return exists
}

func (db *mockDB) Empty(addr common.Address) bool {
	return !db.Exist(addr)
}

func (db *mockDB) AddressInAccessList(addr common.Address) bool {
	return false
}
func (db *mockDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	return false, false
}
func (db *mockDB) AddAddressToAccessList(addr common.Address) {}

func (db *mockDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {}
func (db *mockDB) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
}

func (db *mockDB) RevertToSnapshot(int) {}
func (db *mockDB) Snapshot() int {
	return 0
}

func (db *mockDB) AddLog(*types.Log)               {}
func (db *mockDB) AddPreimage(common.Hash, []byte) {}

var (
	contractCode         = []byte{0x01, 0x02}
	defaultBalance int64 = 123
)

func TestApplyAllocsToState(t *testing.T) {
	db := mockDB{
		accounts: map[common.Address]*types.Account{
			common.HexToAddress("a"): {
				Balance: big.NewInt(defaultBalance),
			},
			common.HexToAddress("b"): {
				Balance: big.NewInt(defaultBalance),
				Code:    bytes.Repeat([]byte{0x01}, 10),
			},
			common.HexToAddress("c"): {
				Balance: big.NewInt(defaultBalance),
				Nonce:   5,
			},
		},
	}

	tests := []struct {
		name             string
		addr             common.Address
		account          types.Account
		allowlist        map[common.Address]bool
		balanceInAccount bool
		wantErr          bool
	}{
		{
			name: "Write to empty account",
			addr: common.HexToAddress("01"),
			account: types.Account{
				Code:  contractCode,
				Nonce: 1,
			},
			balanceInAccount: false,
			wantErr:          false,
		},
		{
			name: "Write to account with only balance should overwrite and keep balance",
			addr: common.HexToAddress("a"),
			account: types.Account{
				Code:  contractCode,
				Nonce: 5,
			},
			balanceInAccount: true,
			wantErr:          false,
		},
		{
			name: "Write to account with existing nonce fails",
			addr: common.HexToAddress("c"),
			account: types.Account{
				Code:  contractCode,
				Nonce: 5,
			},
			wantErr: true,
		},
		{
			name: "Write to account with contract code fails",
			addr: common.HexToAddress("b"),
			account: types.Account{
				Code:  contractCode,
				Nonce: 5,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := applyAllocsToState(&db, &core.Genesis{Alloc: types.GenesisAlloc{tt.addr: tt.account}}, tt.allowlist); (err != nil) != tt.wantErr {
				t.Errorf("applyAllocsToState() error = %v, wantErr %v", err, tt.wantErr)

				account, exists := db.accounts[tt.addr]
				if !exists {
					t.Errorf("account does not exists as expected: %v", tt.addr.Hex())
				}

				assert.Equal(t, tt.account.Nonce, account.Nonce)
				assert.Equal(t, tt.account.Code, account.Code)

				if tt.balanceInAccount {
					assert.Equal(t, big.NewInt(defaultBalance), account.Balance)
				} else {
					assert.Equal(t, big.NewInt(0), account.Balance)
				}
			}
		})
	}
}
