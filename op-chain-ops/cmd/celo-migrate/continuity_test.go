package main

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

func TestCheckContinuity(t *testing.T) {
	decodedHeaders := []*types.Header{
		{Number: big.NewInt(0), ParentHash: common.Hash{}},
		{Number: big.NewInt(1), ParentHash: common.BytesToHash([]byte("hash0"))},
		{Number: big.NewInt(2), ParentHash: common.BytesToHash([]byte("hash1"))},
		{Number: big.NewInt(3), ParentHash: common.BytesToHash([]byte("hash2"))},
	}
	encodedHeaders := make([][]byte, len(decodedHeaders))
	for i, header := range decodedHeaders {
		encodedHeader, err := rlp.EncodeToBytes(header)
		if err != nil {
			t.Fatalf("Failed to encode header: %v", err)
		}
		encodedHeaders[i] = encodedHeader
	}

	tests := []struct {
		name           string
		blockRange     *RLPBlockRange
		prevElement    *RLPBlockElement
		expectedLength uint64
		expectErrorMsg string
	}{
		// Valid continuity tests
		{
			name: "Valid continuity w/ nil prevElement",
			blockRange: &RLPBlockRange{
				start:    0,
				hashes:   [][]byte{[]byte("hash0"), []byte("hash1"), []byte("hash2"), []byte("hash3")},
				headers:  encodedHeaders,
				bodies:   [][]byte{[]byte("body0"), []byte("body1"), []byte("body2"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt0"), []byte("receipt1"), []byte("receipt2"), []byte("receipt3")},
				tds:      [][]byte{[]byte("td0"), []byte("td1"), []byte("td2"), []byte("td3")},
			},
			prevElement:    nil,
			expectedLength: 4,
			expectErrorMsg: "",
		},
		{
			name: "Valid continuity w/ prevElement",
			blockRange: &RLPBlockRange{
				start:    1,
				hashes:   [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
				headers:  encodedHeaders[1:],
				bodies:   [][]byte{[]byte("body1"), []byte("body2"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt1"), []byte("receipt2"), []byte("receipt3")},
				tds:      [][]byte{[]byte("td1"), []byte("td2"), []byte("td3")},
			},
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: []byte("hash0")},
			expectedLength: 3,
			expectErrorMsg: "",
		},
		// Length mismatch tests
		{
			name: "Length mismatch from expected",
			blockRange: &RLPBlockRange{
				start:    1,
				hashes:   [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
				headers:  encodedHeaders[1:],
				bodies:   [][]byte{[]byte("body1"), []byte("body2"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt1"), []byte("receipt2"), []byte("receipt3")},
				tds:      [][]byte{[]byte("td1"), []byte("td2"), []byte("td3")},
			},
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: []byte("hash0")},
			expectedLength: 4,
			expectErrorMsg: "Expected count mismatch in block range hashes: expected 4, actual 3\nExpected count mismatch in block range bodies: expected 4, actual 3\nExpected count mismatch in block range headers: expected 4, actual 3\nExpected count mismatch in block range receipts: expected 4, actual 3\nExpected count mismatch in block range total difficulties: expected 4, actual 3",
		},
		{
			name: "Length mismatch in hashes",
			blockRange: &RLPBlockRange{
				start:    1,
				hashes:   [][]byte{[]byte("hash1"), []byte("hash2")},
				headers:  encodedHeaders[1:],
				bodies:   [][]byte{[]byte("body1"), []byte("body2"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt1"), []byte("receipt2"), []byte("receipt3")},
				tds:      [][]byte{[]byte("td1"), []byte("td2"), []byte("td3")},
			},
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: []byte("hash0")},
			expectedLength: 3,
			expectErrorMsg: "Expected count mismatch in block range hashes: expected 3, actual 2",
		},
		{
			name: "Length mismatch in headers",
			blockRange: &RLPBlockRange{
				start:    1,
				hashes:   [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
				headers:  encodedHeaders,
				bodies:   [][]byte{[]byte("body1"), []byte("body2"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt1"), []byte("receipt2"), []byte("receipt3")},
				tds:      [][]byte{[]byte("td1"), []byte("td2"), []byte("td3")},
			},
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: []byte("hash0")},
			expectedLength: 3,
			expectErrorMsg: "Expected count mismatch in block range headers: expected 3, actual 4",
		},
		{
			name: "Length mismatch in bodies",
			blockRange: &RLPBlockRange{
				start:    1,
				hashes:   [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
				headers:  encodedHeaders[1:],
				bodies:   [][]byte{[]byte("body1"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt1"), []byte("receipt2"), []byte("receipt3")},
				tds:      [][]byte{[]byte("td1"), []byte("td2"), []byte("td3")},
			},
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: []byte("hash0")},
			expectedLength: 3,
			expectErrorMsg: "Expected count mismatch in block range bodies: expected 3, actual 2",
		},
		{
			name: "Length mismatch in receipts",
			blockRange: &RLPBlockRange{
				start:    1,
				hashes:   [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
				headers:  encodedHeaders[1:],
				bodies:   [][]byte{[]byte("body1"), []byte("body2"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt1"), []byte("receipt2")},
				tds:      [][]byte{[]byte("td1"), []byte("td2"), []byte("td3")},
			},
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: []byte("hash0")},
			expectedLength: 3,
			expectErrorMsg: "Expected count mismatch in block range receipts: expected 3, actual 2",
		},
		{
			name: "Length mismatch in tds",
			blockRange: &RLPBlockRange{
				start:    1,
				hashes:   [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
				headers:  encodedHeaders[1:],
				bodies:   [][]byte{[]byte("body1"), []byte("body2"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt1"), []byte("receipt2"), []byte("receipt3")},
				tds:      [][]byte{[]byte("td1"), []byte("td2"), []byte("td3"), []byte("td4")},
			},
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: []byte("hash0")},
			expectedLength: 3,
			expectErrorMsg: "Expected count mismatch in block range total difficulties: expected 3, actual 4",
		},
		// Number mismatch tests
		{
			name: "Header number mismatch from range index",
			blockRange: &RLPBlockRange{
				start:    2,
				hashes:   [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
				headers:  encodedHeaders[1:],
				bodies:   [][]byte{[]byte("body1"), []byte("body2"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt1"), []byte("receipt2"), []byte("receipt3")},
				tds:      [][]byte{[]byte("td1"), []byte("td2"), []byte("td3")},
			},
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: []byte("hash0")},
			expectedLength: 3,
			expectErrorMsg: "decoded header number mismatch indicating a gap in block numbers: expected 2, actual 1",
		},
		{
			name: "Header number mismatch from prevElement number",
			blockRange: &RLPBlockRange{
				start:    1,
				hashes:   [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
				headers:  encodedHeaders[1:],
				bodies:   [][]byte{[]byte("body1"), []byte("body2"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt1"), []byte("receipt2"), []byte("receipt3")},
				tds:      [][]byte{[]byte("td1"), []byte("td2"), []byte("td3")},
			},
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[1], hash: []byte("hash1")},
			expectedLength: 3,
			expectErrorMsg: "header number mismatch indicating a gap in block numbers: expected 2, actual 1\nparent hash mismatch between blocks 1 and 1",
		},
		// Parent hash mismatch tests
		{
			name: "Parent hash mismatch",
			blockRange: &RLPBlockRange{
				start:    1,
				hashes:   [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
				headers:  encodedHeaders[1:],
				bodies:   [][]byte{[]byte("body1"), []byte("body2"), []byte("body3")},
				receipts: [][]byte{[]byte("receipt1"), []byte("receipt2"), []byte("receipt3")},
				tds:      [][]byte{[]byte("td1"), []byte("td2"), []byte("td3")},
			},
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: []byte("wrong-hash")},
			expectedLength: 3,
			expectErrorMsg: "parent hash mismatch between blocks 1 and 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.blockRange.CheckContinuity(tt.prevElement, tt.expectedLength)
			if tt.expectErrorMsg == "" {
				require.NoError(t, err, "CheckContinuity() unexpected error")
			} else {
				require.Error(t, err, "CheckContinuity() expected error")
				require.EqualError(t, err, tt.expectErrorMsg, "CheckContinuity() error message")
			}
		})
	}
}
