package pipeline

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

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
