package shared

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-chain-ops/script"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func versionReturnCode(t *testing.T, version string) []byte {
	t.Helper()
	stringType, err := abi.NewType("string", "", nil)
	require.NoError(t, err)
	data, err := (abi.Arguments{{Type: stringType}}).Pack(version)
	require.NoError(t, err)
	require.LessOrEqual(t, len(data), 255)
	// Copy the ABI response from the end of this runtime code and return it.
	code := []byte{0x60, byte(len(data)), 0x60, 0x0c, 0x60, 0x00, 0x39, 0x60, byte(len(data)), 0x60, 0x00, 0xf3}
	return append(code, data...)
}

func TestGetOPCMVersion(t *testing.T) {
	for _, tc := range []struct {
		name    string
		code    []byte
		version string
		wantErr string
	}{
		{name: "not deployed at fork block"},
		{name: "deployed OPCM", code: versionReturnCode(t, "8.0.1"), version: "8.0.1"},
		{name: "deployed empty version", code: versionReturnCode(t, ""), wantErr: "empty OPCM version"},
		{name: "malformed response", code: []byte{0x00}, wantErr: "invalid response length"},
		{name: "version call reverts", code: common.FromHex("0x60006000fd"), wantErr: "execution reverted"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			host := script.NewHost(testlog.Logger(t, log.LevelError), nil, nil, script.DefaultContext)
			addr := common.HexToAddress("0x1000")
			host.ImportAccount(addr, types.Account{Code: tc.code})
			version, err := getOPCMVersion(&HostCaller{Host: host}, addr)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.version, version)
		})
	}
}

func TestGetLastUsedOPCMVersion(t *testing.T) {
	for _, tc := range []struct {
		name       string
		code       []byte
		version    string
		hasVersion bool
		wantErr    string
	}{
		{name: "pre OPCM v6", code: common.FromHex("0x60006000fd")},
		{name: "current version", code: versionReturnCode(t, "8.0.1"), version: "8.0.1", hasVersion: true},
		{name: "malformed response", code: []byte{0x00}, wantErr: "invalid response length"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			host := script.NewHost(testlog.Logger(t, log.LevelError), nil, nil, script.DefaultContext)
			addr := common.HexToAddress("0x1000")
			host.ImportAccount(addr, types.Account{Code: tc.code})
			version, hasVersion, err := getLastUsedOPCMVersion(&HostCaller{Host: host}, addr)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.version, version)
			require.Equal(t, tc.hasVersion, hasVersion)
		})
	}
}
