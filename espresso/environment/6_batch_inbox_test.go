package environment_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/setuputils"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum-optimism/optimism/op-service/txmgr/metrics"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// TestE2eDevnetWithoutAuthenticatingBatches verifies that batches posted without
// a corresponding BatchInfoAuthenticated event are rejected by the derivation pipeline.
//
// The batcher's BatchAuthenticatorAddress is zeroed out so it skips the
// authenticateBatchInfo call. Batches land on L1 (BatchInbox is an EOA, so txs
// succeed), but the derivation pipeline finds no matching auth event and ignores them.
//
// Arrange:
//
//	Start sequencer, batcher in Espresso mode and OP node.
//	Zero out the batcher's BatchAuthenticatorAddress so it skips authentication.
//
// Assert:
//
//	Assert that the batch transaction lands on L1 (BatchInbox is an EOA).
//	Assert that the derivation pipeline doesn't progress (no auth event).
func TestE2eDevnetWithoutAuthenticatingBatches(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)

	system, _, err := launcher.StartE2eDevnet(ctx, t,
		env.WithBatcherStoppedInitially(),
	)

	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	batchDriver := system.BatchSubmitter.TestDriver()
	// Set mock batcher authenticator address
	batchDriver.RollupConfig.BatchAuthenticatorAddress = common.Address{}

	// Substitute batcher's transaction manager with one that always sends transactions, even
	// if they won't succeed. Otherwise batcher wouldn't submit transactions that would revert to
	// batch inbox.
	// Use the Espresso batcher key (HD index 6) — the same key the primary batcher signs with.
	// This ensures the tx comes from an address that is NOT the SystemConfig batcher, so the
	// derivation pipeline's fallback authorization won't accept it either.
	txMgrCliConfig := setuputils.NewTxMgrConfig(system.NodeEndpoint(e2esys.RoleL1), system.Cfg.Secrets.AccountAtIdx(6))
	txMgrConfig, err := txmgr.NewConfig(txMgrCliConfig, log.Root())
	require.NoError(t, err)
	txMgrConfig.Backend = AlwaysSendingETHBackend{
		inner: txMgrConfig.Backend,
	}
	txMgr, err := txmgr.NewSimpleTxManagerFromConfig("always-sending", log.Root(), &metrics.NoopTxMetrics{}, txMgrConfig)
	require.NoError(t, err)
	batchDriver.Txmgr = txMgr

	// Start the batcher
	err = batchDriver.StartBatchSubmitting()
	require.NoError(t, err, "Couldn't start batcher")
	l1Client := system.NodeClient(e2esys.RoleL1)

	// Wait for batcher to submit a transaction to BatchInbox
	var batchInboxTxHash common.Hash
	for {
		l1Height, err := l1Client.BlockNumber(ctx)
		require.NoError(t, err)
		_, err = geth.FindBlock(l1Client,
			0,
			int(l1Height),
			time.Minute*2,
			func(block *types.Block) (bool, error) {
				for _, tx := range block.Transactions() {
					if *tx.To() == system.RollupConfig.BatchInboxAddress {
						batchInboxTxHash = tx.Hash()
						return true, nil
					}
				}
				return false, nil
			})
		if err == nil {
			break
		}
	}

	receipt, err := l1Client.TransactionReceipt(ctx, batchInboxTxHash)
	require.NoError(t, err)

	// BatchInbox is an EOA, so the transaction lands successfully on L1.
	// However, the derivation pipeline should reject it because there is no
	// BatchInfoAuthenticated event (the batcher skipped authentication).
	require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful, "transaction to EOA BatchInbox should succeed")

	_, err = geth.WaitForBlockToBeSafe(new(big.Int).SetUint64(1), system.NodeClient(e2esys.RoleVerif), time.Minute)
	require.Error(t, err)
}

// A wrapper for testing that proxies all calls to ETHBackend unchanged,
// except EstimateGas and CallContract calls, which always "succeed"
// without making any actual RPC calls.
//
// Wrapping SimpleTxManager's backend with it ensures that SimpleTxManager will always send
// transactions, even if they would be reverted. The reason for this behaviour is
// that SimpleTxManager will check whether transaction will be executed successfully
// before submitting it, either by calling CallContract if transaction request had
// set the gas cap, or by checking EstimateGas return value if transaction request
// doesn't have the gas cap set. Mocking these two methods to always succeed thus
// makes SimpleTxManager submit even invalid transactions, which it wouldn't normally do.
type AlwaysSendingETHBackend struct {
	inner txmgr.ETHBackend
}

// BlockNumber implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) BlockNumber(ctx context.Context) (uint64, error) {
	return m.inner.BlockNumber(ctx)
}

// CallContract implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return []byte{}, nil
}

// Close implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) Close() {
	m.inner.Close()
}

// EstimateGas implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return 1_000_000, nil
}

// HeaderByNumber implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return m.inner.HeaderByNumber(ctx, number)
}

// NonceAt implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return m.inner.NonceAt(ctx, account, blockNumber)
}

// PendingNonceAt implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return m.inner.PendingNonceAt(ctx, account)
}

// SendTransaction implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return m.inner.SendTransaction(ctx, tx)
}

// SuggestGasTipCap implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return m.inner.SuggestGasTipCap(ctx)
}

// TransactionReceipt implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return m.inner.TransactionReceipt(ctx, txHash)
}

// BlobBaseFee implements txmgr.ETHBackend.
func (m AlwaysSendingETHBackend) BlobBaseFee(ctx context.Context) (*big.Int, error) {
	return m.inner.BlobBaseFee(ctx)
}

// Ensure conformance to ETHBackend
var _ txmgr.ETHBackend = AlwaysSendingETHBackend{}
