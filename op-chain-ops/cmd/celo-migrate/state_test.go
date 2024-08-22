package main

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

var (
	contractCode         = []byte{0x01, 0x02}
	defaultBalance int64 = 123
)

func TestApplyAllocsToState(t *testing.T) {
	tests := []struct {
		name             string
		addr             common.Address
		existingAccount  *types.Account
		newAccount       types.Account
		allowlist        map[common.Address]bool
		balanceInAccount bool
		wantErr          bool
	}{
		{
			name: "Write to empty account",
			addr: common.HexToAddress("01"),
			newAccount: types.Account{
				Code:  contractCode,
				Nonce: 1,
			},
			balanceInAccount: false,
			wantErr:          false,
		},
		{
			name: "Copy account with non-zero balance fails",
			addr: common.HexToAddress("a"),
			existingAccount: &types.Account{
				Balance: big.NewInt(defaultBalance),
			},
			newAccount: types.Account{
				Balance: big.NewInt(1),
			},
			wantErr: true,
		},
		{
			name: "Write to account with only balance should overwrite and keep balance",
			addr: common.HexToAddress("a"),
			existingAccount: &types.Account{
				Balance: big.NewInt(defaultBalance),
			},
			newAccount: types.Account{
				Code:  contractCode,
				Nonce: 5,
			},
			balanceInAccount: true,
			wantErr:          false,
		},
		{
			name: "Write to account with existing nonce fails",
			addr: common.HexToAddress("c"),
			existingAccount: &types.Account{
				Balance: big.NewInt(defaultBalance),
				Nonce:   5,
			},
			newAccount: types.Account{
				Code:  contractCode,
				Nonce: 5,
			},
			wantErr: true,
		},
		{
			name: "Write to account with contract code fails",
			addr: common.HexToAddress("b"),
			existingAccount: &types.Account{
				Balance: big.NewInt(defaultBalance),
				Code:    bytes.Repeat([]byte{0x01}, 10),
			},
			newAccount: types.Account{
				Code:  contractCode,
				Nonce: 5,
			},
			wantErr: true,
		},
		{
			name: "Write account with allowlist overwrites",
			addr: common.HexToAddress("d"),
			existingAccount: &types.Account{
				Balance: big.NewInt(defaultBalance),
				Code:    bytes.Repeat([]byte{0x01}, 10),
			},
			newAccount: types.Account{
				Code:  contractCode,
				Nonce: 5,
			},
			allowlist:        map[common.Address]bool{common.HexToAddress("d"): true},
			balanceInAccount: true,
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := rawdb.NewMemoryDatabase()
			tdb := state.NewDatabase(db)
			sdb, _ := state.New(types.EmptyRootHash, tdb, nil)

			if tt.existingAccount != nil {
				sdb.CreateAccount(tt.addr)

				if tt.existingAccount.Balance != nil {
					sdb.SetBalance(tt.addr, uint256.MustFromBig(tt.existingAccount.Balance))
				}
				if tt.existingAccount.Nonce != 0 {
					sdb.SetNonce(tt.addr, tt.existingAccount.Nonce)
				}
				if tt.existingAccount.Code != nil {
					sdb.SetCode(tt.addr, tt.existingAccount.Code)
				}
			}

			if err := applyAllocsToState(sdb, &core.Genesis{Alloc: types.GenesisAlloc{tt.addr: tt.newAccount}}, tt.allowlist); (err != nil) != tt.wantErr {
				t.Errorf("applyAllocsToState() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Don't check account state if an error was thrown
			if tt.wantErr {
				return
			}

			if !sdb.Exist(tt.addr) {
				t.Errorf("account does not exists as expected: %v", tt.addr.Hex())
			}

			assert.Equal(t, tt.newAccount.Nonce, sdb.GetNonce(tt.addr))
			assert.Equal(t, tt.newAccount.Code, sdb.GetCode(tt.addr))

			if tt.balanceInAccount {
				assert.True(t, big.NewInt(defaultBalance).Cmp(sdb.GetBalance(tt.addr).ToBig()) == 0)
			} else {
				assert.True(t, big.NewInt(0).Cmp(sdb.GetBalance(tt.addr).ToBig()) == 0)
			}
		})
	}
}
