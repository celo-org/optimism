package derive

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestCgtV2HeaderBinding(t *testing.T) {
	for _, name := range []string{"valid", "wrong hash", "wrong number", "nil number", "nil header", "RPC error", "missing RPC", "oversized number"} {
		t.Run(name, func(t *testing.T) {
			_, source, _, origin := cgtTestDeriver(t, nil)
			var rpc CgtV2InputSource = source
			switch name {
			case "wrong hash":
				source.header.Extra = []byte{1}
			case "wrong number":
				source.header.Number = big.NewInt(12)
			case "nil number":
				source.header.Number = nil
			case "nil header":
				source.header = nil
			case "RPC error":
				source.err = errors.New("offline")
			case "missing RPC":
				rpc = nil
			case "oversized number":
				source.header.Number = new(big.Int).Lsh(big.NewInt(1), 65)
			}
			header, err := cgtV2Header(context.Background(), rpc, origin)
			if name == "valid" {
				require.NoError(t, err)
				require.IsType(t, &types.Header{}, header)
			} else {
				require.Error(t, err)
				if name == "RPC error" {
					require.ErrorIs(t, err, ErrTemporary)
				} else {
					require.ErrorIs(t, err, ErrCritical)
				}
			}
		})
	}
}
