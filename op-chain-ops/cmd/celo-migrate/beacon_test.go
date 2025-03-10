package main

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinalizedL1BlockSelection(t *testing.T) {
	bc := NewBeaconClient("https://ethereum-beacon-api.publicnode.com")
	var targetTime uint64 = 1740759647 // this is the timestamp of slot 11161302 https://beaconcha.in/slot/11161302
	var expectedSlot uint64 = 11161216 // this is the beginning slot of the epoch 2 before https://beaconcha.in/slot/11161216
	expectedL1BlockHash := common.HexToHash("0x6c71e110fc83faa393017681ab97b26621cda68d12e31d24917e210d169d7be5")

	// We expect the same block to be chosen regardless of which slot in an
	// epoch the target time falls in. The only thing that would change the
	// chosen block would be if the target time fell in a different epoch.

	t.Run("MidEpochSlot", func(t *testing.T) {
		checkFinalizedL1BlockSelection(t, bc, targetTime, expectedSlot, expectedL1BlockHash)
	})

	t.Run("MidEpochMidSlot", func(t *testing.T) {
		checkFinalizedL1BlockSelection(t, bc, targetTime+beaconChainSlotDurationSeconds/2, expectedSlot, expectedL1BlockHash)
	})

	epochFirstSlotTime := EpochStartTime(Mainnet, ContainingEpoch(Mainnet, targetTime))
	t.Run("StartEpochSlot", func(t *testing.T) {
		checkFinalizedL1BlockSelection(t, bc, epochFirstSlotTime, expectedSlot, expectedL1BlockHash)
	})

	epochLastSlotTime := epochFirstSlotTime + beaconChainSlotDurationSeconds*(beaconSlotsPerEpoch-1)
	t.Run("EndEpochSlot", func(t *testing.T) {
		checkFinalizedL1BlockSelection(t, bc, epochLastSlotTime, expectedSlot, expectedL1BlockHash)
	})

	// Execution payload block hash retrieved here - https://beaconcha.in/slot/11161184
	expectedL1BlockHashPrevEpoch := common.HexToHash("0xca6237f41190e1adfa52ba20fadb03ef14a7f57c516806edf7660f301b08de36")
	t.Run("PriorEpochEndSlot", func(t *testing.T) {
		checkFinalizedL1BlockSelection(t, bc, epochFirstSlotTime-1, expectedSlot-beaconSlotsPerEpoch, expectedL1BlockHashPrevEpoch)
	})

	// Execution payload block hash retrieved here - https://beaconcha.in/slot/11161248
	expectedL1BlockHashNextEpoch := common.HexToHash("0x7892cbc14d57ab7036a722306bcae1fc07a202fc5907a227cce539b5f1ca41b3")
	t.Run("NextEpochStartSlot", func(t *testing.T) {
		checkFinalizedL1BlockSelection(t, bc, epochLastSlotTime+beaconChainSlotDurationSeconds, expectedSlot+beaconSlotsPerEpoch, expectedL1BlockHashNextEpoch)
	})
}

func checkFinalizedL1BlockSelection(t *testing.T, bc *BeaconClient, targetTime, expectedSlot uint64, expectedBlockHash common.Hash) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	calculatedHash, err := bc.MostRecentFinalizedBlockAtTime(Mainnet, targetTime)
	require.NoError(t, err)
	assert.Equal(t, expectedBlockHash, calculatedHash)
	// Double check that we are targeting the correct slot
	block, err := bc.BeaconBlock(ctx, expectedSlot)
	require.NoError(t, err)
	assert.Equal(t, expectedBlockHash.String(), block.Data.Message.Body.ExecutionPayload.BlockHash)
}
