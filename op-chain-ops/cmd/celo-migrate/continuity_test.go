package main

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

var (
	hashes   = [][]byte{[]byte("hash0"), []byte("hash1"), []byte("hash2"), []byte("hash3")}
	bodies   = [][]byte{[]byte("body0"), []byte("body1"), []byte("body2"), []byte("body3")}
	receipts = [][]byte{[]byte("receipt0"), []byte("receipt1"), []byte("receipt2"), []byte("receipt3")}
	tds      = [][]byte{[]byte("td0"), []byte("td1"), []byte("td2"), []byte("td3")}

	decodedHeaders = []*types.Header{
		{Number: big.NewInt(0), ParentHash: common.Hash{}},
		{Number: big.NewInt(1), ParentHash: common.BytesToHash(hashes[0])},
		{Number: big.NewInt(2), ParentHash: common.BytesToHash(hashes[1])},
		{Number: big.NewInt(3), ParentHash: common.BytesToHash(hashes[2])},
	}
)

func makeRange(start int, bodies, receipts, tds, hashes, encodedHeaders [][]byte) *RLPBlockRange {
	r := &RLPBlockRange{}
	r.start = uint64(start)
	r.hashes = append(r.hashes, hashes...)
	r.headers = append(r.headers, encodedHeaders...)
	r.bodies = append(r.bodies, bodies...)
	r.receipts = append(r.receipts, receipts...)
	r.tds = append(r.tds, tds...)
	return r
}

func TestCheckContinuity(t *testing.T) {
	headers := make([][]byte, len(decodedHeaders))
	for i, header := range decodedHeaders {
		encodedHeader, err := rlp.EncodeToBytes(header)
		if err != nil {
			t.Fatalf("Failed to encode header: %v", err)
		}
		headers[i] = encodedHeader
	}

	checkContinuity(
		t,
		"Valid continuity w/ nil prevElement",
		makeRange(0, bodies, receipts, tds, hashes, headers),
		"",
	)
	checkContinuity(
		t,
		"Valid continuity w/ prevElement",
		makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers[1:]),
		"",
	)
	checkContinuity(
		t,
		"Header number mismatch from range index",
		makeRange(2, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers[1:]),
		"decoded header number mismatch indicating a gap in block numbers: expected 2, actual 1",
	)

	// Set the second header in the range to have a bad parent hash.
	r := makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers[1:])
	badParentHash := *decodedHeaders[2]
	badParentHash.ParentHash = common.Hash{}
	encodedBadParentHash, err := rlp.EncodeToBytes(&badParentHash)
	require.NoError(t, err)
	r.headers[1] = encodedBadParentHash
	checkContinuity(t, "Parent hash mismatch", r, "parent hash mismatch between blocks 2 and 1")
}

func checkContinuity(t *testing.T, name string, blockRange *RLPBlockRange, expectErrorMsg string) {
	t.Run(name, func(t *testing.T) {
		err := blockRange.CheckContinuity()
		if expectErrorMsg == "" {
			require.NoError(t, err, "CheckContinuity() unexpected error")
		} else {
			require.Error(t, err, "CheckContinuity() expected error")
			require.EqualError(t, err, expectErrorMsg, "CheckContinuity() error message")
		}
	})
}

func TestCheckLengths(t *testing.T) {
	headers := make([][]byte, len(hashes))
	checkLengths(
		t,
		"Length mismatch from expected",
		makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers[1:]),
		4,
		"Unexpected number of hashes for block range: expected 4, actual 3\nUnexpected number of bodies for block range: expected 4, actual 3\nUnexpected number of headers for block range: expected 4, actual 3\nUnexpected number of receipts for block range: expected 4, actual 3\nUnexpected number of total difficulties for block range: expected 4, actual 3",
	)
	checkLengths(
		t,
		"Length mismatch in hashes",
		makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[2:], headers[1:]),
		3,
		"Unexpected number of hashes for block range: expected 3, actual 2",
	)
	checkLengths(
		t,
		"Length mismatch in headers",
		makeRange(1, bodies[1:], receipts[1:], tds[1:], hashes[1:], headers),
		3,
		"Unexpected number of headers for block range: expected 3, actual 4",
	)
	checkLengths(
		t,
		"Length mismatch in bodies",
		makeRange(1, bodies[2:], receipts[1:], tds[1:], hashes[1:], headers[1:]),
		3,
		"Unexpected number of bodies for block range: expected 3, actual 2",
	)
	checkLengths(
		t,
		"Length mismatch in receipts",
		makeRange(1, bodies[1:], receipts[2:], tds[1:], hashes[1:], headers[1:]),
		3,
		"Unexpected number of receipts for block range: expected 3, actual 2",
	)
	checkLengths(
		t,
		"Length mismatch in tds",
		makeRange(1, bodies[1:], receipts[1:], tds[2:], hashes[1:], headers[1:]),
		3,
		"Unexpected number of total difficulties for block range: expected 3, actual 2",
	)
}

func checkLengths(t *testing.T, name string, blockRange *RLPBlockRange, expectedLength int, expectErrorMsg string) {
	t.Run(name, func(t *testing.T) {
		err := blockRange.CheckLengths(expectedLength)
		if expectErrorMsg == "" {
			require.NoError(t, err, "CheckLengths() unexpected error")
		} else {
			require.Error(t, err, "CheckLengths() expected error")
			require.EqualError(t, err, expectErrorMsg, "CheckLengths() error message")
		}
	})
}
