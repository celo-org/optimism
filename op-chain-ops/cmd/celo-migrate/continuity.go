package main

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

// RLPBlockRange is a range of blocks in RLP format
type RLPBlockRange struct {
	start    uint64
	hashes   [][]byte
	headers  [][]byte
	bodies   [][]byte
	receipts [][]byte
	tds      [][]byte
}

// RLPBlockElement contains all relevant block data in RLP format
type RLPBlockElement struct {
	decodedHeader *types.Header
	hash          []byte
	header        []byte
	body          []byte
	receipts      []byte
	td            []byte
}

// Follows checks if the current block has a number one greater than the previous block
// and if the parent hash of the current block matches the hash of the previous block.
func (e *RLPBlockElement) Follows(prev *RLPBlockElement) (err error) {
	if e.Number() != prev.Number()+1 {
		err = errors.Join(err, fmt.Errorf("header number mismatch indicating a gap in block numbers: expected %d, actual %d", prev.Number()+1, e.Number()))
	}
	// We compare the parent hash with the stored hash of the previous block because
	// at this point the header object will not calculate the correct hash since it
	// first needs to be transformed.
	if e.Header().ParentHash != common.BytesToHash(prev.hash) {
		err = errors.Join(err, fmt.Errorf("parent hash mismatch between blocks %d and %d", e.Number(), prev.Number()))
	}
	return err
}

func (e *RLPBlockElement) Header() *types.Header {
	return e.decodedHeader
}

func (e *RLPBlockElement) Number() uint64 {
	return e.Header().Number.Uint64()
}

func (r *RLPBlockRange) Element(i uint64) (*RLPBlockElement, error) {
	header := types.Header{}
	err := rlp.DecodeBytes(r.headers[i], &header)
	if err != nil {
		return nil, fmt.Errorf("can't decode header: %w", err)
	}
	return &RLPBlockElement{
		decodedHeader: &header,
		hash:          r.hashes[i],
		header:        r.headers[i],
		body:          r.bodies[i],
		receipts:      r.receipts[i],
		td:            r.tds[i],
	}, nil
}

// CheckContinuity checks if the block data in the range is continuous
// by comparing the header number and parent hash of each block with the previous block,
// and by checking if the number of elements retrieved from each table is the same.
// It takes in the last element in the preceding range, and returns the last element
// in the current range so that continuity can be checked across ranges.
func (r *RLPBlockRange) CheckContinuity(prevElement *RLPBlockElement, expectedLength uint64) (*RLPBlockElement, error) {
	log.Info("Checking data continuity for block range",
		"start", r.start,
		"end", r.start+expectedLength-1,
		"count", expectedLength,
		"prevElement", func() interface{} {
			if prevElement != nil {
				return prevElement.Number()
			}
			return "nil"
		}(),
	)

	if err := r.CheckLengths(expectedLength); err != nil {
		return nil, err
	}
	for i := range r.hashes {
		currElement, err := r.Element(uint64(i))
		if err != nil {
			return nil, err
		}
		if currElement.Number() != r.start+uint64(i) {
			return nil, fmt.Errorf("decoded header number mismatch indicating a gap in block numbers: expected %d, actual %d", r.start+uint64(i), currElement.Number())
		}
		if prevElement != nil {
			log.Debug("Checking continuity", "block", currElement.Number(), "prev", prevElement.Number())
			if err := currElement.Follows(prevElement); err != nil {
				return nil, err
			}
		}
		prevElement = currElement
	}
	return prevElement, nil
}

// CheckLengths makes sure the number of elements retrieved from each table is the same
func (r *RLPBlockRange) CheckLengths(expectedLength uint64) error {
	var err error
	if uint64(len(r.hashes)) != expectedLength {
		err = fmt.Errorf("Unexpected number of hashes for block range: expected %d, actual %d", expectedLength, len(r.hashes))
	}
	if uint64(len(r.bodies)) != expectedLength {
		err = errors.Join(err, fmt.Errorf("Unexpected number of bodies for block range: expected %d, actual %d", expectedLength, len(r.bodies)))
	}
	if uint64(len(r.headers)) != expectedLength {
		err = errors.Join(err, fmt.Errorf("Unexpected number of headers for block range: expected %d, actual %d", expectedLength, len(r.headers)))
	}
	if uint64(len(r.receipts)) != expectedLength {
		err = errors.Join(err, fmt.Errorf("Unexpected number of receipts for block range: expected %d, actual %d", expectedLength, len(r.receipts)))
	}
	if uint64(len(r.tds)) != expectedLength {
		err = errors.Join(err, fmt.Errorf("Unexpected number of total difficulties for block range: expected %d, actual %d", expectedLength, len(r.tds)))
	}
	return err
}
