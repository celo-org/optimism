package celo

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
)

// Foundry "test test test test test test test test test test test junk" mnemonic,
// path m/44'/60'/0'/0/0 — same address that L2Genesis.s.sol pre-funds with 100k TEST.
const foundryDevAccount0PrivKeyHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

// ensureCeloFeeCurrencyOrSkip checks that the FeeCurrencyDirectory predeploy is in
// place and that the test fee currency is registered. Skips the test otherwise so
// the gate can be run against environments without the Celo predeploys.
func ensureCeloFeeCurrencyOrSkip(t devtest.T, sys *presets.Minimal) {
	l2 := sys.L2EL.Escape().L2EthClient()

	ctx, cancel := context.WithTimeout(t.Ctx(), 20*time.Second)
	defer cancel()

	// Probe the FeeCurrencyDirectory by calling getCurrencyConfig(testFeeCurrency).
	// On a non-Celo devnet the call has no code to execute and reverts; on a Celo
	// devnet without the test currency registered, the directory itself reverts
	// with "currency not registered". Both cases skip.
	getCurrencyConfig := w3.MustNewFunc(
		"getCurrencyConfig(address)",
		"(address oracle, uint256 intrinsicGas)",
	)
	data, err := getCurrencyConfig.EncodeArgs(predeploys.TestFeeCurrencyAddr)
	t.Require().NoError(err, "encode getCurrencyConfig")

	if _, err := l2.Call(ctx, ethereum.CallMsg{To: &predeploys.FeeCurrencyDirectoryAddr, Data: data}, rpc.LatestBlockNumber); err != nil {
		t.Skip("test fee currency not registered with FeeCurrencyDirectory: " + err.Error())
	}
}

// erc20BalanceOf reads the ERC-20 balance of `account` for the token at `token`.
func erc20BalanceOf(t devtest.T, sys *presets.Minimal, token, account common.Address) *big.Int {
	l2 := sys.L2EL.Escape().L2EthClient()

	ctx, cancel := context.WithTimeout(t.Ctx(), 20*time.Second)
	defer cancel()

	balanceOf := w3.MustNewFunc("balanceOf(address)", "uint256")
	data, err := balanceOf.EncodeArgs(account)
	t.Require().NoError(err, "encode balanceOf")

	out, err := l2.Call(ctx, ethereum.CallMsg{To: &token, Data: data}, rpc.LatestBlockNumber)
	t.Require().NoError(err, "call balanceOf")

	var bal *big.Int
	t.Require().NoError(balanceOf.DecodeReturns(out, &bal), "decode balanceOf")
	return bal
}
