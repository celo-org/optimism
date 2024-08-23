package main

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

var (
	contractCode         = []byte{0x01, 0x02}
	defaultBalance int64 = 123
	address              = common.HexToAddress("a")
)

func TestApplyAllocsToState(t *testing.T) {
	tests := []struct {
		name            string
		existingAccount *types.Account
		newAccount      types.Account
		allowlist       map[common.Address]bool
		wantErr         bool
	}{
		{
			name: "Write to empty account",
			newAccount: types.Account{
				Code:  contractCode,
				Nonce: 1,
			},
			wantErr: false,
		},
		{
			name: "Copy account with non-zero balance fails",
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
			existingAccount: &types.Account{
				Balance: big.NewInt(defaultBalance),
			},
			newAccount: types.Account{
				Code:  contractCode,
				Nonce: 5,
			},
			wantErr: false,
		},
		{
			name: "Write to account with existing nonce fails",
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
			existingAccount: &types.Account{
				Balance: big.NewInt(defaultBalance),
				Code:    bytes.Repeat([]byte{0x01}, 10),
			},
			newAccount: types.Account{
				Code:  contractCode,
				Nonce: 5,
			},
			allowlist: map[common.Address]bool{address: true},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := rawdb.NewMemoryDatabase()
			tdb := state.NewDatabase(db)
			sdb, _ := state.New(types.EmptyRootHash, tdb, nil)

			if tt.existingAccount != nil {
				sdb.CreateAccount(address)

				if tt.existingAccount.Balance != nil {
					sdb.SetBalance(address, uint256.MustFromBig(tt.existingAccount.Balance), tracing.BalanceChangeUnspecified)
				}
				if tt.existingAccount.Nonce != 0 {
					sdb.SetNonce(address, tt.existingAccount.Nonce)
				}
				if tt.existingAccount.Code != nil {
					sdb.SetCode(address, tt.existingAccount.Code)
				}
			}

			if err := applyAllocsToState(sdb, &core.Genesis{Alloc: types.GenesisAlloc{address: tt.newAccount}}, tt.allowlist); (err != nil) != tt.wantErr {
				t.Errorf("applyAllocsToState() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Don't check account state if an error was thrown
			if tt.wantErr {
				return
			}

			if !sdb.Exist(address) {
				t.Errorf("account does not exists as expected: %v", address.Hex())
			}

			assert.Equal(t, tt.newAccount.Nonce, sdb.GetNonce(address))
			assert.Equal(t, tt.newAccount.Code, sdb.GetCode(address))

			if tt.existingAccount != nil {
				assert.True(t, big.NewInt(defaultBalance).Cmp(sdb.GetBalance(address).ToBig()) == 0)
			} else {
				assert.True(t, big.NewInt(0).Cmp(sdb.GetBalance(address).ToBig()) == 0)
			}
		})
	}
}
