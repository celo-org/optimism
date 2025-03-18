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
	// Uncomment to see log output during tests
	// log.SetDefault(log.NewLogger(log.NewTerminalHandler(os.Stdout, false)))
	t.Run("DefaultTest", func(t *testing.T) {
		bc := NewBeaconClient("https://ethereum-beacon-api.publicnode.com")
		var chainID uint64 = Mainnet
		var targetTime uint64 = 1740759647 // this is the timestamp of slot 11161302 https://beaconcha.in/slot/11161302
		// var targetTime uint64 = 1741771358 // this is the timestamp of slot 11161302 https://beaconcha.in/slot/11161302
		var expectedSlot uint64 = 11161216 // this is the beginning slot of the epoch 2 before https://beaconcha.in/slot/11161216
		expectedL1BlockHash := common.HexToHash("0x6c71e110fc83faa393017681ab97b26621cda68d12e31d24917e210d169d7be5")
		t.Run("MidEpochSlot", func(t *testing.T) {
			checkFinalizedL1BlockSelection(t, bc, chainID, targetTime, expectedSlot, expectedL1BlockHash)
		})

		t.Run("MidEpochMidSlot", func(t *testing.T) {
			checkFinalizedL1BlockSelection(t, bc, chainID, targetTime+beaconChainSlotDurationSeconds/2, expectedSlot, expectedL1BlockHash)
		})

		epochFirstSlotTime := EpochStartTime(chainID, ContainingEpoch(chainID, targetTime))
		t.Run("StartEpochSlot", func(t *testing.T) {
			checkFinalizedL1BlockSelection(t, bc, chainID, epochFirstSlotTime, expectedSlot, expectedL1BlockHash)
		})

		epochLastSlotTime := epochFirstSlotTime + beaconChainSlotDurationSeconds*(beaconSlotsPerEpoch-1)
		t.Run("EndEpochSlot", func(t *testing.T) {
			checkFinalizedL1BlockSelection(t, bc, chainID, epochLastSlotTime, expectedSlot, expectedL1BlockHash)
		})

		// Execution payload block hash retrieved here - https://beaconcha.in/slot/11161184
		expectedL1BlockHashPrevEpoch := common.HexToHash("0xca6237f41190e1adfa52ba20fadb03ef14a7f57c516806edf7660f301b08de36")
		t.Run("PriorEpochEndSlot", func(t *testing.T) {
			checkFinalizedL1BlockSelection(t, bc, chainID, epochFirstSlotTime-1, expectedSlot-beaconSlotsPerEpoch, expectedL1BlockHashPrevEpoch)
		})

		// Execution payload block hash retrieved here - https://beaconcha.in/slot/11161248
		expectedL1BlockHashNextEpoch := common.HexToHash("0x7892cbc14d57ab7036a722306bcae1fc07a202fc5907a227cce539b5f1ca41b3")
		t.Run("NextEpochStartSlot", func(t *testing.T) {
			checkFinalizedL1BlockSelection(t, bc, chainID, epochLastSlotTime+beaconChainSlotDurationSeconds, expectedSlot+beaconSlotsPerEpoch, expectedL1BlockHashNextEpoch)
		})
	})

	// This test seeks to recreate the conditions of our first celo mainnet
	// migration dry run using holesky as the L1, in the end we chose not to use
	// holesky, and also this test does not work because the holesky beacon rpc
	// api does not support queries for old finality_checkpoints.
	t.Run("CeloMainnetDryRun1Holesky", func(t *testing.T) {
		t.Skip("Holesky public node does not support old historical queries for finality_checkpoints, so this test does not work any more")
		// celo l1 migraiton block 30819350
		// Timestamp for the last block of the l1 (one prior to the migration block) 1741771293
		// Timestamp targeted by the migration script for l2 block start 1741771353  (last block of celo l1 + one minute)
		bc := NewBeaconClient("https://ethereum-holesky-beacon-api.publicnode.com")
		var targetTime uint64 = 1741771353
		var expectedSlot uint64 = 3822400
		var expectedL1BlockHash common.Hash = common.HexToHash("0xd85604da94ed6263239dc6e48bbc508c69445f084a8d2bde2e4de6c536876b61")
		checkFinalizedL1BlockSelection(t, bc, Holesky, targetTime, expectedSlot, expectedL1BlockHash)

	})

	// This test seeks to recreate the conditions of our first celo mainnet
	// migration dry run using ethereum mainnet as the L1.
	t.Run("CeloMainnetDryRun1Mainnet", func(t *testing.T) {
		// celo l1 migraiton block 30824250
		// Timestamp for the last block of the l1 (one prior to the migration block) 1741771293
		// Timestamp targeted by the migration script for l2 block start 1741795853 (last block of celo l1 + one minute)
		bc := NewBeaconClient("https://ethereum-beacon-api.publicnode.com")
		var targetTime uint64 = 1741795853
		var expectedSlot uint64 = 11247584
		var expectedL1BlockHash common.Hash = common.HexToHash("0x4e2105c9e9d948efd49e611cbd6cfaee2053af2df20ed5c51880a6c310d0a360")
		checkFinalizedL1BlockSelection(t, bc, Mainnet, targetTime, expectedSlot, expectedL1BlockHash)

	})
}

func checkFinalizedL1BlockSelection(t *testing.T, bc *BeaconClient, chainID, targetTime, expectedSlot uint64, expectedBlockHash common.Hash) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	t.Cleanup(cancel)
	calculatedHash, err := bc.MostRecentFinalizedBlockAtTime(chainID, targetTime)
	require.NoError(t, err)
	assert.Equal(t, expectedBlockHash, calculatedHash)
	// Double check that we are targeting the correct slot
	block, err := bc.BeaconBlock(ctx, expectedSlot)
	require.NoError(t, err)
	assert.Equal(t, expectedBlockHash.String(), block.Data.Message.Body.ExecutionPayload.BlockHash)
}
