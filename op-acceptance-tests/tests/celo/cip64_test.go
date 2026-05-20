package celo

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
)

// TestCIP64_PayGasInTestFeeCurrency sends a CeloDynamicFeeTxV2 (CIP-64) with
// FeeCurrency set to the registered test fee currency. devAccounts[0] is the
// foundry test mnemonic account that L2Genesis pre-funds with 100k TEST and
// zero native CELO; the only way the tx can be paid for is via op-geth's
// fee-currency machinery.
func TestCIP64_PayGasInTestFeeCurrency(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMinimal(t)
	ensureCeloFeeCurrencyOrSkip(t, sys)

	priv, err := crypto.HexToECDSA(foundryDevAccount0PrivKeyHex)
	t.Require().NoError(err, "parse foundry dev priv key")
	sender := crypto.PubkeyToAddress(priv.PublicKey)

	l2 := sys.L2EL.Escape().L2EthClient()
	chainID := sys.L2Chain.ChainID().ToBig()
	recipient := sys.Wallet.NewEOA(sys.L2EL).Address()

	ctx, cancel := context.WithTimeout(t.Ctx(), 60*time.Second)
	defer cancel()

	nativeBefore, err := l2.BalanceAt(ctx, sender, nil)
	t.Require().NoError(err, "get native balance")
	testBefore := erc20BalanceOf(t, sys, predeploys.TestFeeCurrencyAddr, sender)
	t.Require().Truef(testBefore.Sign() > 0, "sender must have a positive TEST balance, got %s", testBefore)

	nonce, err := l2.PendingNonceAt(ctx, sender)
	t.Require().NoError(err, "get nonce")

	tx := types.NewTx(&types.CeloDynamicFeeTxV2{
		ChainID:     chainID,
		Nonce:       nonce,
		GasTipCap:   big.NewInt(1_000_000_000),  // 1 gwei
		GasFeeCap:   big.NewInt(10_000_000_000), // 10 gwei
		Gas:         200_000,
		To:          &recipient,
		Value:       big.NewInt(0),
		FeeCurrency: &predeploys.TestFeeCurrencyAddr,
	})
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), priv)
	t.Require().NoError(err, "sign CIP-64 tx")
	t.Require().NoError(l2.SendTransaction(ctx, signed), "send CIP-64 tx")

	var receipt *types.Receipt
	t.Require().Eventually(func() bool {
		r, err := l2.TransactionReceipt(ctx, signed.Hash())
		if err != nil {
			return false
		}
		receipt = r
		return true
	}, sys.L2EL.Escape().TransactionTimeout(), time.Second, "tx receipt")

	t.Require().Equal(uint64(1), receipt.Status, "CIP-64 tx must succeed")

	nativeAfter, err := l2.BalanceAt(ctx, sender, nil)
	t.Require().NoError(err, "get native balance after")
	testAfter := erc20BalanceOf(t, sys, predeploys.TestFeeCurrencyAddr, sender)

	// Native balance must not change: gas was paid in TEST, no value transferred.
	t.Require().Equal(0, nativeBefore.Cmp(nativeAfter),
		"native balance changed: before=%s after=%s", nativeBefore, nativeAfter)

	// TEST balance must drop — the fee-currency machinery debited gas + intrinsic-gas
	// surcharge in TEST tokens.
	t.Require().Truef(testAfter.Cmp(testBefore) < 0,
		"TEST balance must drop, got before=%s after=%s", testBefore, testAfter)
}
