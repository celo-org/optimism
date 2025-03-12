package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

const (
	beaconChainSlotDurationSeconds = 12
	beaconSlotsPerEpoch            = 32
)

var (
	Mainnet                        uint64 = 1
	Holesky                        uint64 = 17000
	Sepolia                        uint64 = 11155111
	beaconChainGenesisTimesSeconds        = map[uint64]uint64{
		Mainnet: 1606824023,
		Holesky: 1695902400,
		Sepolia: 1655733600,
	}
)

type BeaconClient struct {
	cl *http.Client
	// A beaconchain RPC API endpoint.
	beaconRPC string
}

func NewBeaconClient(beaconRPC string) *BeaconClient {
	return &BeaconClient{
		beaconRPC: beaconRPC,
		cl:        &http.Client{},
	}
}

// Waits for the given epoch start time plus an extra 10 seconds just to ensure that infrastructure has had time to update.
func AwaitEpoch(chainID uint64, epoch uint64) {
	start := EpochStartTime(chainID, epoch)

	// Wait till the beginning of the epoch in which our point in time falls. We wait an extra 10 seconds to be sure that
	// the network has had time to update.
	waitTimeSeconds := int64(start) - time.Now().Unix() + 10
	if waitTimeSeconds > 0 {
		log.Info("Waiting %v + additional 10 seconds for start of epoch %d, to check for finalized slot at that time", waitTimeSeconds-10, epoch)
		time.Sleep(time.Duration(waitTimeSeconds) * time.Second)
	}
}

// findL1StartingBlock looks back or forward (depending on the epochIncrement)
// for a finalized L1 block that is up to maxSequencerDrift before the L2 fork
// block.
func (c *BeaconClient) findL1StartingBlock(chainID uint64, unixTime uint64, epochIncrement int64) (common.Hash, error) {
	var l1StartBlockHash common.Hash
	var l1StartBlockTime uint64

	epoch := int64(ContainingEpoch(chainID, unixTime))
	log.Info(fmt.Sprintf(
		"Searching for finalized block at time %d, chain (%d) beacon genesis time %d, containing epoch %d, epoch start time %d",
		unixTime, chainID, beaconChainGenesisTimesSeconds[chainID], epoch, EpochStartTime(chainID, uint64(epoch))))

	for ; ; epoch += epochIncrement {
		AwaitEpoch(chainID, uint64(epoch))
		slot := FirstSlotOfEpoch(uint64(epoch))
		finalityCheckpoints, err := c.FindFinalityCheckpointsForSlot(slot, 10)
		if err != nil {
			if errors.Is(err, ethereum.NotFound) {
				continue
			}
			return common.Hash{}, fmt.Errorf("failed fetching finality checkpoints for slot %d: %w", slot, err)
		}
		justifiedEpoch := uint64(finalityCheckpoints.Data.Finalized.Epoch)
		if !withinMaxSequencerDrift(EpochStartTime(chainID, justifiedEpoch), unixTime) {
			// We've gone too far back and not found a suitable block, so break.
			break
		}
		finalizedSlot := FirstSlotOfEpoch(justifiedEpoch)
		l1StartBlockHash, l1StartBlockTime, err = c.FindBlockForSlot(finalizedSlot, 10)
		if err != nil {
			if errors.Is(err, ethereum.NotFound) {
				continue
			}
			return common.Hash{}, fmt.Errorf("failed fetching L1 block for slot (%v): %w", finalizedSlot, err)
		}
		switch {
		case !withinMaxSequencerDrift(l1StartBlockTime, unixTime):
			// This block is too old to use with the L2 fork block. Nullify the hash
			// to signify that no suitable block was yet found.
			l1StartBlockHash = common.Hash{}
		case l1StartBlockTime >= unixTime:
			// We've gone too far forward and not found a suitable block, so nullify the hash and break.
			l1StartBlockHash = common.Hash{}
		default:
			log.Info(fmt.Sprintf("Found finalized L1 slot at %d", finalizedSlot))
		}
		break
	}
	return l1StartBlockHash, nil
}

// MostRecentFinalizedBlockAtTime returns the hash of the most recent finalized
// block that is up to maxSequencerDrift before unixTime. It starts by
// looking back from the epoch containing the given time, but if no finalized
// block is found it will consider future epochs that may have finalized a block
// ocurring before the given time.
func (c *BeaconClient) MostRecentFinalizedBlockAtTime(chainID uint64, unixTime uint64) (common.Hash, error) {
	if unixTime < beaconChainGenesisTimesSeconds[chainID] {
		return common.Hash{}, fmt.Errorf(
			"Searching for finalized block before time %d but %d is before network beacon chain genesis time %d",
			unixTime, unixTime, beaconChainGenesisTimesSeconds[chainID],
		)
	}
	hash, err := c.findL1StartingBlock(chainID, unixTime, -1)
	if err != nil {
		return common.Hash{}, err
	}
	if hash == (common.Hash{}) {
		// Try searching forward
		hash, err = c.findL1StartingBlock(chainID, unixTime, 1)
		if err != nil {
			return common.Hash{}, err
		}
	}
	return hash, nil
}

func withinMaxSequencerDrift(l1StartingBlockTime, l2StartBlockTime uint64) bool {
	// Used to check block validity, duplicate of the
	// value defined at op-node/rollup.maxSequencerDriftCelo.
	var maxSequencerDriftCelo uint64 = 2892
	return l2StartBlockTime-l1StartingBlockTime <= maxSequencerDriftCelo
}

// FindFinalityCheckpointForSlot returns the finality checkpoints for 'slot'
// searching up to 'tries' slots back if only empty slots are encountered.
func (c *BeaconClient) FindFinalityCheckpointsForSlot(slot uint64, tries uint64) (*FinalityCheckpoints, error) {
	var finalityCheckpoints *FinalityCheckpoints
	var err error
	for i := range tries {
		targetSlot := slot - i
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		finalityCheckpoints, err = c.FinalityCheckpoints(ctx, targetSlot)
		if errors.Is(err, ethereum.NotFound) {
			// If there is no checkpoint at this slot, skip to the previous slot.
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("error fetching finality checkpoints for slot %d: %w", targetSlot, err)
		}
		if finalityCheckpoints.Data.Finalized.Epoch == 0 {
			// In the case that finality has not yet been achieved, skip to the previous slot.
			continue
		}
		return finalityCheckpoints, nil
	}
	return nil, fmt.Errorf("finality_checkpoints %w searching up to %d slots back from slot (%d)", ethereum.NotFound, tries, slot)
}

// FindBlockForSlot returns the hash and timestamp of the block at the given slot,
// looking up to 'tries' slots back if only empty slots are encountered.
func (c *BeaconClient) FindBlockForSlot(slot uint64, tries uint64) (blockHash common.Hash, blockTime uint64, err error) {
	// Find the most recent actual finalized block, slots can be empty so we
	// search back if we encounter empty slots. We check up to 10 slots, if they
	// are all empty something serious is wrong with the L1 so we abort.
	var beaconBlock *BeaconBlock
	for i := range tries {
		targetSlot := slot - i
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		beaconBlock, err = c.BeaconBlock(ctx, targetSlot)
		if errors.Is(err, ethereum.NotFound) {
			// If there is not block for this slot then skip to the next.
			continue
		}
		if err != nil {
			return common.Hash{}, 0, fmt.Errorf("error fetching beacon block at slot %d: %w", targetSlot, err)
		}
		if !beaconBlock.Finalized {
			return common.Hash{}, 0, fmt.Errorf("expecting beacon block at slot %d to be finalized", targetSlot)
		}
		break // We found a good block.
	}
	if beaconBlock == nil {
		return common.Hash{}, 0, fmt.Errorf("finalized block %w searching up to %d slots back from slot (%d)", ethereum.NotFound, tries, slot)
	}
	return common.HexToHash(beaconBlock.Data.Message.Body.ExecutionPayload.BlockHash), uint64(beaconBlock.Data.Message.Body.ExecutionPayload.Timestamp), nil
}

func (c *BeaconClient) FinalityCheckpoints(ctx context.Context, slot uint64) (checkpoints *FinalityCheckpoints, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/eth/v1/beacon/states/%d/finality_checkpoints", c.beaconRPC, slot), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to get finality checkpoints for slot %d: %w", slot, err)
	}
	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch finality checkpoints for slot %d: %w", slot, err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("finality checkpoints for slot %d %w", slot, ethereum.NotFound)
		}
		return nil, fmt.Errorf("failed to fetch finality checkpoints for slot %d, http status %d: %w", slot, resp.StatusCode, err)
	}

	checkpoints = &FinalityCheckpoints{}
	if err := json.NewDecoder(resp.Body).Decode(checkpoints); err != nil {
		return nil, fmt.Errorf("failed to decode finality checkpoints: %w", err)
	}
	return checkpoints, nil
}

// BeaconBlock gets the beacon block from the beacon rpc api.
func (c *BeaconClient) BeaconBlock(ctx context.Context, slot uint64) (block *BeaconBlock, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/eth/v2/beacon/blocks/%d", c.beaconRPC, slot), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to get beacon block: %w", err)
	}
	resp, err := c.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch beacon block: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("beacon block at slot %d %w", slot, ethereum.NotFound)
		}
		return nil, fmt.Errorf("failed to fetch beacon block at slot %d, http status %d", slot, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
		return nil, fmt.Errorf("failed to decode beacon block: %w", err)
	}
	return block, nil
}

// EpochStartTime returns the start time of an epoch.
func EpochStartTime(chainID uint64, epoch uint64) uint64 {
	return beaconChainGenesisTimesSeconds[chainID] + (FirstSlotOfEpoch(epoch) * beaconChainSlotDurationSeconds)
}

func SlotTime(chainID uint64, slot uint64) uint64 {
	return beaconChainGenesisTimesSeconds[chainID] + (slot * beaconChainSlotDurationSeconds)
}

// FirstSlotOfEpoch returns the number of the first slot of an epoch.
func FirstSlotOfEpoch(epoch uint64) uint64 {
	return epoch * beaconSlotsPerEpoch
}

// ContainingEpoch returns the number of the epoch whithin which the given time falls.
func ContainingEpoch(chainID uint64, unixTime uint64) uint64 {
	return ContainingSlot(chainID, unixTime) / beaconSlotsPerEpoch
}

// ContainingSlot returns the slot within which the given time falls.
func ContainingSlot(chainID uint64, unixTime uint64) uint64 {
	// Get the slot at or before the given time.
	// Slot = (start - genesis) / slotDuration
	return (unixTime - beaconChainGenesisTimesSeconds[chainID]) / beaconChainSlotDurationSeconds
}

type BeaconBlock struct {
	Version             string `json:"version"`
	ExecutionOptimistic bool   `json:"execution_optimistic"`
	Finalized           bool   `json:"finalized"`
	Data                struct {
		Message struct {
			Slot          eth.Uint64String `json:"slot"`
			ProposerIndex string           `json:"proposer_index"`
			ParentRoot    string           `json:"parent_root"`
			StateRoot     string           `json:"state_root"`
			Body          struct {
				RandaoReveal string `json:"randao_reveal"`
				Eth1Data     struct {
					DepositRoot  string `json:"deposit_root"`
					DepositCount string `json:"deposit_count"`
					BlockHash    string `json:"block_hash"`
				} `json:"eth1_data"`
				Graffiti          string        `json:"graffiti"`
				ProposerSlashings []interface{} `json:"proposer_slashings"`
				AttesterSlashings []interface{} `json:"attester_slashings"`
				Attestations      []struct {
					AggregationBits string `json:"aggregation_bits"`
					Data            struct {
						Slot            string `json:"slot"`
						Index           string `json:"index"`
						BeaconBlockRoot string `json:"beacon_block_root"`
						Source          struct {
							Epoch string `json:"epoch"`
							Root  string `json:"root"`
						} `json:"source"`
						Target struct {
							Epoch string `json:"epoch"`
							Root  string `json:"root"`
						} `json:"target"`
					} `json:"data"`
					Signature string `json:"signature"`
				} `json:"attestations"`
				Deposits       []interface{} `json:"deposits"`
				VoluntaryExits []interface{} `json:"voluntary_exits"`
				SyncAggregate  struct {
					SyncCommitteeBits      string `json:"sync_committee_bits"`
					SyncCommitteeSignature string `json:"sync_committee_signature"`
				} `json:"sync_aggregate"`
				ExecutionPayload struct {
					ParentHash    string           `json:"parent_hash"`
					FeeRecipient  string           `json:"fee_recipient"`
					StateRoot     string           `json:"state_root"`
					ReceiptsRoot  string           `json:"receipts_root"`
					LogsBloom     string           `json:"logs_bloom"`
					PrevRandao    string           `json:"prev_randao"`
					BlockNumber   eth.Uint64String `json:"block_number"`
					GasLimit      string           `json:"gas_limit"`
					GasUsed       string           `json:"gas_used"`
					Timestamp     eth.Uint64String `json:"timestamp"`
					ExtraData     string           `json:"extra_data"`
					BaseFeePerGas string           `json:"base_fee_per_gas"`
					BlockHash     string           `json:"block_hash"`
					Transactions  []string         `json:"transactions"`
					Withdrawals   []struct {
						Index          string `json:"index"`
						ValidatorIndex string `json:"validator_index"`
						Address        string `json:"address"`
						Amount         string `json:"amount"`
					} `json:"withdrawals"`
					BlobGasUsed   string `json:"blob_gas_used"`
					ExcessBlobGas string `json:"excess_blob_gas"`
				} `json:"execution_payload"`
				BlsToExecutionChanges []interface{} `json:"bls_to_execution_changes"`
				BlobKzgCommitments    []string      `json:"blob_kzg_commitments"`
			} `json:"body"`
		} `json:"message"`
		Signature string `json:"signature"`
	} `json:"data"`
}

type FinalityCheckpoints struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		PreviousJustified struct {
			Epoch eth.Uint64String `json:"epoch"`
			Root  string           `json:"root"`
		} `json:"previous_justified"`
		CurrentJustified struct {
			Epoch eth.Uint64String `json:"epoch"`
			Root  string           `json:"root"`
		} `json:"current_justified"`
		Finalized struct {
			Epoch eth.Uint64String `json:"epoch"`
			Root  string           `json:"root"`
		} `json:"finalized"`
	} `json:"data"`
}
