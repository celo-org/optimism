package batcher

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

// TestCheckEspressoDataAvailability: chains with Espresso scheduled are
// calldata-only (DEC-op-026) — post-Espresso derivation drops blob batches, so a
// blob/auto DA configuration must be rejected at startup.
func TestCheckEspressoDataAvailability(t *testing.T) {
	espressoTime := uint64(0)

	tests := []struct {
		name         string
		espressoTime *uint64
		daType       flags.DataAvailabilityType
		wantErr      bool
	}{
		{"espresso not scheduled: blobs allowed", nil, flags.BlobsType, false},
		{"espresso not scheduled: auto allowed", nil, flags.AutoType, false},
		{"espresso scheduled: calldata allowed", &espressoTime, flags.CalldataType, false},
		{"espresso scheduled: blobs rejected", &espressoTime, flags.BlobsType, true},
		{"espresso scheduled: auto rejected", &espressoTime, flags.AutoType, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs := &BatcherService{RollupConfig: &rollup.Config{
				EspressoTime: test.espressoTime,
			}}
			cfg := &CLIConfig{DataAvailabilityType: test.daType}
			err := bs.checkEspressoDataAvailability(cfg)
			if test.wantErr {
				require.ErrorContains(t, err, "calldata only")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestCheckFallbackAuthConfirmations: the NumConfirmations headroom bound only
// applies when a blob/auto DA batcher can actually emit auth→batch pairs — a
// BatchAuthenticator is configured AND the EspressoTime fork is scheduled.
//
// Cases pairing a set espressoTime with a non-calldata DA type describe batchers
// checkEspressoDataAvailability refuses to start. They only occur if the calldata-only
// restriction is lifted, and are kept so the bound stays covered if it is.
func TestCheckFallbackAuthConfirmations(t *testing.T) {
	espressoTime := uint64(0)
	authAddr := common.Address{0x01}

	tests := []struct {
		name          string
		authenticator common.Address
		espressoTime  *uint64
		daType        flags.DataAvailabilityType
		numConfs      uint64
		wantErr       bool
	}{
		{"no authenticator", common.Address{}, &espressoTime, flags.BlobsType, 100, false},
		{"espresso not scheduled", authAddr, nil, flags.BlobsType, 100, false},
		{"calldata exempt", authAddr, &espressoTime, flags.CalldataType, 100, false},
		{"blob within headroom", authAddr, &espressoTime, flags.BlobsType, 25, false},
		{"blob exceeds headroom", authAddr, &espressoTime, flags.BlobsType, 26, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs := &BatcherService{RollupConfig: &rollup.Config{
				BatchAuthenticatorAddress: test.authenticator,
				EspressoTime:              test.espressoTime,
			}}
			cfg := &CLIConfig{
				DataAvailabilityType: test.daType,
				TxMgrConfig:          txmgr.CLIConfig{NumConfirmations: test.numConfs},
			}
			err := bs.checkFallbackAuthConfirmations(cfg)
			if test.wantErr {
				require.ErrorContains(t, err, "NumConfirmations")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
