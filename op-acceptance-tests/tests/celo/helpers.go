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

	// Skip if the FeeCurrencyDirectory has no code: eth_call on an empty account
	// silently returns empty bytes rather than reverting, so we have to check code
	// length explicitly.
	header, err := l2.InfoByLabel(ctx, "latest")
	t.Require().NoError(err, "fetch latest L2 header")
	code, err := l2.CodeAtHash(ctx, predeploys.FeeCurrencyDirectoryAddr, header.Hash())
	t.Require().NoError(err, "eth_getCode FeeCurrencyDirectory")
	if len(code) == 0 {
		t.Skip("FeeCurrencyDirectory predeploy missing — chain is not running with Celo predeploys")
	}

	// Confirm the test fee currency is actually registered. An unregistered token
	// returns a zero-valued config (oracle = address(0)) without reverting, so the
	// presence of a non-zero oracle is what proves registration. The first 32 bytes
	// of the ABI-encoded return are the oracle address (left-padded).
	getCurrencyConfig := w3.MustNewFunc(
		"getCurrencyConfig(address)",
		"(address oracle, uint256 intrinsicGas)",
	)
	data, err := getCurrencyConfig.EncodeArgs(predeploys.TestFeeCurrencyAddr)
	t.Require().NoError(err, "encode getCurrencyConfig")
	out, err := l2.Call(ctx, ethereum.CallMsg{To: &predeploys.FeeCurrencyDirectoryAddr, Data: data}, rpc.LatestBlockNumber)
	t.Require().NoError(err, "FeeCurrencyDirectory.getCurrencyConfig")
	if len(out) < 32 || common.BytesToAddress(out[:32]) == (common.Address{}) {
		t.Skip("test fee currency not registered with FeeCurrencyDirectory")
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
