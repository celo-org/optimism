package main

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

func makeRange(start int, bodies, receipts, tds, hashes, encodedHeaders [][]byte) *RLPBlockRange {
	return &RLPBlockRange{
		start:    uint64(start),
		hashes:   hashes,
		headers:  encodedHeaders,
		bodies:   bodies,
		receipts: receipts,
		tds:      tds,
	}
}

func TestCheckContinuity(t *testing.T) {
	hashes := [][]byte{[]byte("hash0"), []byte("hash1"), []byte("hash2"), []byte("hash3")}
	bodies := [][]byte{[]byte("body0"), []byte("body1"), []byte("body2"), []byte("body3")}
	receipts := [][]byte{[]byte("receipt0"), []byte("receipt1"), []byte("receipt2"), []byte("receipt3")}
	tds := [][]byte{[]byte("td0"), []byte("td1"), []byte("td2"), []byte("td3")}
	decodedHeaders := []*types.Header{
		{Number: big.NewInt(0), ParentHash: common.Hash{}},
		{Number: big.NewInt(1), ParentHash: common.BytesToHash(hashes[0])},
		{Number: big.NewInt(2), ParentHash: common.BytesToHash(hashes[1])},
		{Number: big.NewInt(3), ParentHash: common.BytesToHash(hashes[2])},
	}
	headers := make([][]byte, len(decodedHeaders))
	for i, header := range decodedHeaders {
		encodedHeader, err := rlp.EncodeToBytes(header)
		if err != nil {
			t.Fatalf("Failed to encode header: %v", err)
		}
		headers[i] = encodedHeader
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
			name:           "Valid continuity w/ nil prevElement",
			blockRange:     makeRange(0, bodies, receipts, tds, hashes, headers),
			prevElement:    nil,
			expectedLength: 4,
			expectErrorMsg: "",
		},
		{
			name:           "Valid continuity w/ prevElement",
			blockRange:     makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers[1:]),
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: hashes[0]},
			expectedLength: 3,
			expectErrorMsg: "",
		},
		// Length mismatch tests
		{
			name:           "Length mismatch from expected",
			blockRange:     makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers[1:]),
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: hashes[0]},
			expectedLength: 4,
			expectErrorMsg: "Unexpected number of hashes for block range: expected 4, actual 3\nUnexpected number of headers for block range: expected 4, actual 3\nUnexpected number of bodies for block range: expected 4, actual 3\nUnexpected number of receipts for block range: expected 4, actual 3\nUnexpected number of total difficulties for block range: expected 4, actual 3",
		},
		{
			name:           "Length mismatch in hashes",
			blockRange:     makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[2:], headers[1:]),
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: hashes[0]},
			expectedLength: 3,
			expectErrorMsg: "Unexpected number of hashes for block range: expected 3, actual 2",
		},
		{
			name:           "Length mismatch in headers",
			blockRange:     makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers),
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: hashes[0]},
			expectedLength: 3,
			expectErrorMsg: "Unexpected number of headers for block range: expected 3, actual 4",
		},
		{
			name:           "Length mismatch in bodies",
			blockRange:     makeRange(1, bodies[2:], receipts[1:], tds[1:], hashes[1:], headers[1:]),
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: hashes[0]},
			expectedLength: 3,
			expectErrorMsg: "Unexpected number of bodies for block range: expected 3, actual 2",
		},
		{
			name:           "Length mismatch in receipts",
			blockRange:     makeRange(1, bodies[1:], receipts[2:], tds[1:], hashes[1:], headers[1:]),
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: hashes[0]},
			expectedLength: 3,
			expectErrorMsg: "Unexpected number of receipts for block range: expected 3, actual 2",
		},
		{
			name:           "Length mismatch in tds",
			blockRange:     makeRange(1, bodies[1:], receipts[1:], tds[2:], hashes[1:], headers[1:]),
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: hashes[0]},
			expectedLength: 3,
			expectErrorMsg: "Unexpected number of total difficulties for block range: expected 3, actual 2",
		},
		// Number mismatch tests
		{
			name:           "Header number mismatch from range index",
			blockRange:     makeRange(2, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers[1:]),
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[0], hash: hashes[0]},
			expectedLength: 3,
			expectErrorMsg: "decoded header number mismatch indicating a gap in block numbers: expected 2, actual 1",
		},
		{
			name:           "Header number mismatch from prevElement number",
			blockRange:     makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers[1:]),
			prevElement:    &RLPBlockElement{decodedHeader: decodedHeaders[1], hash: hashes[1]},
			expectedLength: 3,
			expectErrorMsg: "header number mismatch indicating a gap in block numbers: expected 2, actual 1\nparent hash mismatch between blocks 1 and 1",
		},
		// Parent hash mismatch tests
		{
			name:           "Parent hash mismatch",
			blockRange:     makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers[1:]),
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
