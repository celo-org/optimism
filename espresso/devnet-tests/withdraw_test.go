package devnet_tests

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	nodebindings "github.com/ethereum-optimism/optimism/op-node/bindings"
	nodepreview "github.com/ethereum-optimism/optimism/op-node/bindings/preview"
	"github.com/ethereum-optimism/optimism/op-node/withdrawals"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/stretchr/testify/require"
)

// opSuccinctProveABI is the part of the OPSuccinctFaultDisputeGame ABI that
// FaultDisputeGame lacks. Its status() and resolve() share FaultDisputeGame's
// selectors, so the test uses that binding for them.
const opSuccinctProveABI = `[{"type":"function","name":"prove","stateMutability":"nonpayable","inputs":[{"name":"proofBytes","type":"bytes"}],"outputs":[{"name":"","type":"uint8"}]}]`

// l2BlockFromExtraData extracts the L2 block number from a dispute game's
// extra data, whose first 32 bytes hold it as a big-endian uint256.
func l2BlockFromExtraData(extraData []byte) (*big.Int, error) {
	if len(extraData) < 32 {
		return nil, fmt.Errorf("extraData too short: got %d bytes, need at least 32", len(extraData))
	}
	return new(big.Int).SetBytes(extraData[:32]), nil
}

// TestWithdrawal withdraws ETH from L2 to L1 end to end: it initiates a
// withdrawal on L2, waits for an op-succinct dispute game covering it, proves
// the game (the mock verifier accepts an empty proof) and the withdrawal, then
// finalizes the withdrawal once the game resolves and the delays pass.
func TestWithdrawal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Minute)
	defer cancel()

	d := NewDevnet(ctx, t, WithProofs())
	require.NoError(t, d.Up())
	defer func() {
		require.NoError(t, d.Down())
	}()

	require.NoError(t, d.WaitForBatcher(ctx))

	alice := crypto.PubkeyToAddress(d.secrets.Alice.PublicKey)
	callOpts := &bind.CallOpts{Context: ctx}

	l1ChainID, err := d.L1.ChainID(ctx)
	require.NoError(t, err)
	aliceL1, err := bind.NewKeyedTransactorWithChainID(d.secrets.Alice, l1ChainID)
	require.NoError(t, err)
	aliceL1.Context = ctx
	// l1Transactor returns a fresh copy of Alice's L1 transactor, so that
	// setting Value or GasLimit for one transaction does not leak into others.
	l1Transactor := func() *bind.TransactOpts {
		opts := *aliceL1
		return &opts
	}

	systemConfig, _, err := d.SystemConfig(ctx)
	require.NoError(t, err)
	factoryAddr, err := systemConfig.DisputeGameFactory(callOpts)
	require.NoError(t, err)
	portalAddr, err := systemConfig.OptimismPortal(callOpts)
	require.NoError(t, err)

	factory, err := nodebindings.NewDisputeGameFactoryCaller(factoryAddr, d.L1)
	require.NoError(t, err)
	portal2, err := nodepreview.NewOptimismPortal2(portalAddr, d.L1)
	require.NoError(t, err)
	gameType, err := portal2.RespectedGameType(callOpts)
	require.NoError(t, err)

	// Step 1: wait for the proposer to create a first game.
	t.Log("waiting for the proposer to create a first game")
	require.Eventually(t, func() bool {
		count, err := factory.GameCount(callOpts)
		if err != nil {
			t.Logf("error getting the game count: %v", err)
			return false
		}
		return count.Sign() > 0
	}, 3*time.Minute, 2*time.Second, "proposer didn't start")

	// Step 2: fund the bridge with a deposit, then initiate the withdrawal on L2.
	withdrawalAmount := big.NewInt(1e18)

	t.Log("depositing ETH to the L1 bridge")
	depositOpts := l1Transactor()
	depositOpts.Value = new(big.Int).Mul(withdrawalAmount, big.NewInt(2))
	depositOpts.GasLimit = 500_000
	portal, err := bindings.NewOptimismPortal(portalAddr, d.L1)
	require.NoError(t, err)
	depositTx, err := portal.DepositTransaction(depositOpts, common.Address{}, depositOpts.Value, 21_000, false, nil)
	require.NoError(t, err)
	_, err = wait.ForReceiptOK(ctx, d.L1, depositTx.Hash())
	require.NoError(t, err)

	t.Log("initiating the withdrawal on L2")
	messagePasser, err := bindings.NewL2ToL1MessagePasser(predeploys.L2ToL1MessagePasserAddr, d.L2Seq)
	require.NoError(t, err)
	l2ChainID, err := d.L2Seq.ChainID(ctx)
	require.NoError(t, err)
	l2Opts, err := bind.NewKeyedTransactorWithChainID(d.secrets.Alice, l2ChainID)
	require.NoError(t, err)
	l2Opts.Context = ctx
	l2Opts.Value = withdrawalAmount
	withdrawTx, err := messagePasser.InitiateWithdrawal(l2Opts, alice, big.NewInt(21_000), nil)
	require.NoError(t, err)
	withdrawReceipt, err := wait.ForReceiptOK(ctx, d.L2Verif, withdrawTx.Hash())
	require.NoError(t, err)
	t.Logf("withdrawal initiated in L2 block %d", withdrawReceipt.BlockNumber)

	// Step 3: wait for a dispute game covering the withdrawal block.
	t.Logf("waiting for a dispute game covering L2 block %d", withdrawReceipt.BlockNumber)
	var (
		gameIndex   *big.Int
		gameProxy   common.Address
		gameL2Block *big.Int
	)
	require.Eventually(t, func() bool {
		count, err := factory.GameCount(callOpts)
		if err != nil || count.Sign() == 0 {
			return false
		}
		games, err := factory.FindLatestGames(callOpts, gameType, new(big.Int).Sub(count, common.Big1), common.Big1)
		if err != nil || len(games) == 0 {
			return false
		}
		l2Block, err := l2BlockFromExtraData(games[0].ExtraData)
		if err != nil {
			t.Logf("error parsing the game's extra data: %v", err)
			return false
		}
		t.Logf("latest game: index=%d, L2 block=%d (need >= %d)", games[0].Index, l2Block, withdrawReceipt.BlockNumber)
		if l2Block.Cmp(withdrawReceipt.BlockNumber) < 0 {
			return false
		}
		info, err := factory.GameAtIndex(callOpts, games[0].Index)
		if err != nil {
			t.Logf("error getting the game info: %v", err)
			return false
		}
		gameIndex, gameProxy, gameL2Block = games[0].Index, info.Proxy, l2Block
		return true
	}, 10*time.Minute, 5*time.Second, "no game covering the withdrawal block")

	// Step 4: prove the dispute game; the mock verifier accepts an empty proof.
	t.Log("proving the dispute game")
	proveABI, err := abi.JSON(strings.NewReader(opSuccinctProveABI))
	require.NoError(t, err)
	proveTx, err := bind.NewBoundContract(gameProxy, proveABI, d.L1, d.L1, d.L1).Transact(l1Transactor(), "prove", []byte{})
	require.NoError(t, err)
	_, err = wait.ForReceiptOK(ctx, d.L1, proveTx.Hash())
	require.NoError(t, err)

	// Step 5: prove the withdrawal against the game's output root.
	t.Log("proving the withdrawal")
	l2Header, err := d.L2Seq.HeaderByNumber(ctx, gameL2Block)
	require.NoError(t, err)
	params, err := withdrawals.ProveWithdrawalParametersForBlock(ctx, gethclient.New(d.L2Seq.Client()), d.L2Seq, withdrawTx.Hash(), l2Header, gameIndex)
	require.NoError(t, err)

	withdrawal := nodepreview.TypesWithdrawalTransaction{
		Nonce:    params.Nonce,
		Sender:   params.Sender,
		Target:   params.Target,
		Value:    params.Value,
		GasLimit: params.GasLimit,
		Data:     params.Data,
	}
	outputProof := nodepreview.TypesOutputRootProof{
		Version:                  params.OutputRootProof.Version,
		StateRoot:                params.OutputRootProof.StateRoot,
		MessagePasserStorageRoot: params.OutputRootProof.MessagePasserStorageRoot,
		LatestBlockhash:          params.OutputRootProof.LatestBlockhash,
	}
	proveWithdrawalOpts := l1Transactor()
	proveWithdrawalOpts.GasLimit = 500_000
	proveWithdrawalTx, err := portal2.ProveWithdrawalTransaction(proveWithdrawalOpts, withdrawal, gameIndex, outputProof, params.WithdrawalProof)
	require.NoError(t, err)
	proveReceipt, err := wait.ForReceiptOK(ctx, d.L1, proveWithdrawalTx.Hash())
	require.NoError(t, err)

	// Step 6: wait for the game to resolve and for the proof maturity and
	// dispute game finality delays to pass.
	t.Log("waiting for the game to resolve")
	game, err := bindings.NewFaultDisputeGame(gameProxy, d.L1)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		status, err := game.Status(callOpts)
		if err != nil {
			t.Logf("error getting the game status: %v", err)
			return false
		}
		if status != 0 {
			t.Logf("game resolved with status %d", status)
			return true
		}
		// Resolve it ourselves if the challenger has not; this fails until
		// the challenge window has passed.
		resolveTx, err := game.Resolve(l1Transactor())
		if err != nil {
			t.Logf("resolve not possible yet: %v", err)
			return false
		}
		if _, err := wait.ForReceiptOK(ctx, d.L1, resolveTx.Hash()); err != nil {
			t.Logf("resolve tx failed: %v", err)
		}
		return false // Recheck the status on the next iteration.
	}, 5*time.Minute, 5*time.Second, "game not resolved")

	maturityDelay, err := portal2.ProofMaturityDelaySeconds(callOpts)
	require.NoError(t, err)
	finalityDelay, err := portal2.DisputeGameFinalityDelaySeconds(callOpts)
	require.NoError(t, err)
	proveBlock, err := d.L1.HeaderByNumber(ctx, proveReceipt.BlockNumber)
	require.NoError(t, err)
	targetTime := proveBlock.Time + bigs.Uint64Strict(maturityDelay) + bigs.Uint64Strict(finalityDelay) + 10

	t.Log("waiting for the proof maturity and dispute game finality delays")
	require.Eventually(t, func() bool {
		header, err := d.L1.HeaderByNumber(ctx, nil)
		return err == nil && header.Time >= targetTime
	}, 3*time.Minute, 2*time.Second, "timeout waiting for the delays")

	// Step 7: finalize the withdrawal.
	t.Log("finalizing the withdrawal")
	balanceBefore, err := d.L1.BalanceAt(ctx, alice, nil)
	require.NoError(t, err)

	finalizeOpts := l1Transactor()
	finalizeOpts.GasLimit = 300_000
	finalizeTx, err := portal2.FinalizeWithdrawalTransaction(finalizeOpts, withdrawal)
	require.NoError(t, err)
	finalizeReceipt, err := wait.ForReceiptOK(ctx, d.L1, finalizeTx.Hash())
	require.NoError(t, err)

	// Alice receives the withdrawal amount, minus the finalization gas.
	balanceAfter, err := d.L1.BalanceAt(ctx, alice, nil)
	require.NoError(t, err)
	gasCost := new(big.Int).Mul(new(big.Int).SetUint64(finalizeReceipt.GasUsed), finalizeReceipt.EffectiveGasPrice)
	balanceChange := new(big.Int).Sub(balanceAfter, balanceBefore)
	expectedChange := new(big.Int).Sub(withdrawalAmount, gasCost)
	require.True(t, balanceChange.Cmp(expectedChange) >= 0,
		"Alice's L1 balance should increase by at least the withdrawal amount minus gas: got %s, want >= %s", balanceChange, expectedChange)
}
