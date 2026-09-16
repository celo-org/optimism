package batcher

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// mockFixedTimeL1Client is an L1Client whose tip header carries a fixed
// timestamp. The embedded bind.ContractBackend satisfies the rest of the
// L1Client interface; only HeaderByNumber is exercised by the gate.
type mockFixedTimeL1Client struct {
	bind.ContractBackend
	time uint64
}

func (f *mockFixedTimeL1Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return &types.Header{Number: big.NewInt(0), Time: f.time}, nil
}

func (f *mockFixedTimeL1Client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return 0, nil
}

func u64(v uint64) *uint64 { return &v }

type fallbackAuthGateCase struct {
	name         string
	authAddr     common.Address
	espressoTime *uint64
	tipTime      uint64
	want         bool
}

// TestIsFallbackAuthRequired_ForkBoundary locks the batcher's own side of the
// Espresso fork boundary: the fallback batcher must begin authenticating exactly
// at EspressoTime activation and never before, and only when a BatchAuthenticator
// is configured. This is the counterpart to the derivation-side
// TestDataAndHashesFromTxsForkBoundary — a silent divergence here would strand
// batches (auth expected by the verifier but never posted, or posted needlessly).
func TestIsFallbackAuthRequired_ForkBoundary(t *testing.T) {
	const espressoTime = 1000
	authAddr := common.HexToAddress("0x00000000000000000000000000000000000000aa")

	tests := []fallbackAuthGateCase{
		{
			name:         "no authenticator configured is never gated",
			authAddr:     common.Address{},
			espressoTime: u64(espressoTime),
			tipTime:      espressoTime, // even at/after activation
			want:         false,
		},
		{
			name:         "espresso not scheduled is never gated",
			authAddr:     authAddr,
			espressoTime: nil,
			tipTime:      espressoTime,
			want:         false,
		},
		{
			name:         "one second before activation is not gated",
			authAddr:     authAddr,
			espressoTime: u64(espressoTime),
			tipTime:      espressoTime - 1,
			want:         false,
		},
		{
			name:         "exactly at activation is gated",
			authAddr:     authAddr,
			espressoTime: u64(espressoTime),
			tipTime:      espressoTime,
			want:         true,
		},
		{
			name:         "after activation is gated",
			authAddr:     authAddr,
			espressoTime: u64(espressoTime),
			tipTime:      espressoTime + 1,
			want:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := &BatchSubmitter{}
			l.Log = testlog.Logger(t, log.LevelDebug)
			l.Metr = metrics.NoopMetrics
			l.RollupConfig = &rollup.Config{
				BatchAuthenticatorAddress: test.authAddr,
				EspressoTime:              test.espressoTime,
			}
			l.Config.NetworkTimeout = time.Second
			l.L1Client = &mockFixedTimeL1Client{time: test.tipTime}

			got, err := l.isFallbackAuthRequired(context.Background())
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}
