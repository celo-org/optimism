package pipeline

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// TestDeployEspressoSkipsWhenAlreadyDeployed checks that re-applying an intent keeps the
// existing BatchAuthenticator. The Env has no script host, so any deployment attempt would panic.
func TestDeployEspressoSkipsWhenAlreadyDeployed(t *testing.T) {
	chainID := common.HexToHash("0x01")
	existing := common.HexToAddress("0xba7c4a0000000000000000000000000000000001")

	intent := &state.Intent{Chains: []*state.ChainIntent{{ID: chainID, EspressoEnabled: true}}}
	st := &state.State{Chains: []*state.ChainState{{ID: chainID, BatchAuthenticatorAddress: existing}}}
	env := &Env{Logger: testlog.Logger(t, log.LevelInfo)}

	require.NoError(t, DeployEspresso(env, intent, st, chainID))
	require.Equal(t, existing, st.Chains[0].BatchAuthenticatorAddress)
}

func TestNitroEnclaveVerifier(t *testing.T) {
	verifier := common.HexToAddress("0x1234567890123456789012345678901234567890")

	tests := []struct {
		name      string
		address   string
		allowEnv  string
		allowMock bool
		want      common.Address
		wantErr   string
	}{
		{name: "unset without opt-in", wantErr: "is not set"},
		{name: "unset with env opt-in", allowEnv: "true"},
		{name: "unset with genesis opt-in", allowMock: true},
		{name: "valid address", address: verifier.Hex(), want: verifier},
		{name: "malformed address", address: "0x1234", allowEnv: "true", wantErr: "not a valid address"},
		{name: "zero address", address: common.Address{}.Hex(), allowMock: true, wantErr: "zero address"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NITRO_ENCLAVE_VERIFIER_ADDRESS", tt.address)
			t.Setenv(allowMockTEEVerifierEnv, tt.allowEnv)

			got, err := nitroEnclaveVerifier(&Env{AllowMockTEEVerifier: tt.allowMock})
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
