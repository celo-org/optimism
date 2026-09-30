package batcher

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

// failingCallL1Client fails every contract call, so resolving the TEE verifier
// address fails.
type failingCallL1Client struct {
	fakeL1Client
}

func (f *failingCallL1Client) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return nil, errors.New("contract call failed")
}

func newLifecycleSubmitter(t *testing.T) *BatchSubmitter {
	l := &BatchSubmitter{}
	l.Log = testlog.Logger(t, log.LevelDebug)
	l.Config.NetworkTimeout = time.Second
	l.RollupConfig = &rollup.Config{
		BatchAuthenticatorAddress: common.HexToAddress("0x00000000000000000000000000000000000000aa"),
	}
	l.shutdownCtx, l.cancelShutdownCtx = context.WithCancel(context.Background())
	l.killCtx, l.cancelKillCtx = context.WithCancel(context.Background())
	l.wg = &sync.WaitGroup{}
	return l
}

// A start that fails partway must not publish a session: the next start would
// otherwise run clearState against this run's streamer and submitter.
func TestStartEspressoLoopsFailureLeavesNoSession(t *testing.T) {
	l := newLifecycleSubmitter(t)
	l.L1Client = &failingCallL1Client{}

	err := l.startEspressoLoops(&espressoStreamers.Streamer{}, make(chan txmgr.TxReceipt[txRef]), make(chan pubInfo, 1), make(chan int64, 1))
	require.ErrorContains(t, err, "could not resolve TEE verifier address")
	require.Nil(t, l.espresso)
}

// Stopping drops the whole session (streamer, submitter and verifier address),
// so no per-start Espresso state survives into the next start. A zero-value
// Streamer is Stop()-safe, which makes this directly testable.
func TestStopBatchSubmittingDropsEspressoSession(t *testing.T) {
	l := newLifecycleSubmitter(t)
	l.running = true
	l.espresso = &espressoSession{
		streamer:           &espressoStreamers.Streamer{},
		submitter:          &espressoTransactionSubmitter{},
		teeVerifierAddress: common.HexToAddress("0x00000000000000000000000000000000000000bb"),
	}

	require.NoError(t, l.StopBatchSubmitting(context.Background()))
	require.Nil(t, l.espresso)
}
