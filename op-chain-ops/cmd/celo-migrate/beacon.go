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
)

const (
	// See - https://eth2book.info/capella/part3/containers/state/
	beaconChainGenesisTimeSeconds  = 1606824000
	beaconChainSlotDurationSeconds = 12
	beaconSlotsPerEpoch            = 32
)

type beaconClient struct {
	cl *http.Client
	// A beaconchain RPC API endpoint.
	beaconRPC string
	// A becaoncha.in api endpoint.
	beaconchainURL string
}

func NewBeaconClient(beaconRPC string, beaconchainURL string) *beaconClient {
	return &beaconClient{
		beaconRPC:      beaconRPC,
		beaconchainURL: beaconchainURL,
		cl:             &http.Client{},
	}
}

// MostRecentFinalizedL1BlockAtTime returns the hash of the most recent
// finalized L1 block at the L2 start time. It finds the epoch which started
// most recently before the L2 start time (or on the L2 start time) and then
// looks back from there to find the first finalized block.
func (c *beaconClient) MostRecentFinalizedL1BlockAtTime(l2StartTimeSeconds uint64) (common.Hash, error) {
	// Find the epoch starting at or before the L2 start time.
	epochNumber := SlotAtOrBefore(l2StartTimeSeconds) / beaconSlotsPerEpoch

	// This epoch is guaranteed to not be complete at L2 start time (if the L2 start time falls in the last
	// second of the epoch the epoch is still not complete, and if it was we'd be selecting the next epoch)
	// The previous epoch is the most recent completed epoch.
	// The one prior to that is the most recent justified epoch.
	// And the first block of the justified epoch (the epoch boundary block) will be finalized.
	// This is assuming that there was at least 2/3 participation in the completed and justified epochs.
	var epoch *Epoch
	var err error
	names := [2]string{"completed", "justified"}

	// Check the most recent completed and justified epochs had a participation rate of at least 0.67.
	for i := uint64(1); i <= 2; i++ {
		epoch, err = c.Epoch(context.Background(), epochNumber-i)
		if err != nil {
			return common.Hash{}, fmt.Errorf("error fetching epoch %d: %w", epochNumber-i, err)
		}
		if epoch.Data.Globalparticipationrate < 0.67 {
			return common.Hash{}, fmt.Errorf("most recent %s epoch before the L2 start time (%d) has less than 0.67 participation rate (%.2f)", names[i-1], l2StartTimeSeconds, epoch.Data.Globalparticipationrate)
		}
	}
	// Calculate the first slot of the most recent justified epoch.
	mostRecentFinalizedSlot := (epochNumber - 2) * beaconSlotsPerEpoch

	// Find the most recent actual finalized block, slots can be empty so we
	// search back if we encounter empty slots. We check up to 10 slots, if they
	// are all empty something serious is wrong with the L1 so we abort.
	var beaconBlock *BeaconBlock
	for i := uint64(0); i < 10; i++ {
		beaconBlock, err = c.BeaconBlock(context.Background(), mostRecentFinalizedSlot-i)
		if errors.Is(err, ethereum.NotFound) {
			// If there is not block for this slot then skip to the next.
			continue
		}
		if err != nil {
			return common.Hash{}, fmt.Errorf("error fetching beacon block at slot %d: %w", mostRecentFinalizedSlot, err)
		}
		if !beaconBlock.Finalized {
			return common.Hash{}, fmt.Errorf("expecting beacon block at slot %d to be finalized", mostRecentFinalizedSlot)
		}
		break // We found a good block.
	}
	if beaconBlock == nil {
		return common.Hash{}, fmt.Errorf("failed to find finalized block searching up to 10 slots back from the most recent finalized slot (%d) at the L2 fork time (%d)", mostRecentFinalizedSlot, l2StartTimeSeconds)
	}
	return common.HexToHash(beaconBlock.Data.Message.Body.ExecutionPayload.BlockHash), nil
}

// Gets the epoch from the beaconcha.in api
func (c *beaconClient) Epoch(ctx context.Context, num uint64) (epoch *Epoch, err error) {
	headers := http.Header{}
	headers.Add("Accept", "application/json")
	resp, err := c.cl.Get(fmt.Sprintf("%s/v1/epoch/%d", c.beaconchainURL, num))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch epoch: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch epoch, http status %d", resp.StatusCode)
	}
	epoch = &Epoch{}
	if err := json.NewDecoder(resp.Body).Decode(epoch); err != nil {
		return nil, fmt.Errorf("failed to decode epoch: %w", err)
	}
	return epoch, nil
}

// Gets the beacon block from the beacon rpc api
func (c *beaconClient) BeaconBlock(ctx context.Context, slot uint64) (block *BeaconBlock, err error) {
	headers := http.Header{}
	headers.Add("Accept", "application/json")
	resp, err := c.cl.Get(fmt.Sprintf("%s/eth/v2/beacon/blocks/%d", c.beaconRPC, slot))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch becon block: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("failed to fetch beacon block at slot %d: %w", slot, ethereum.NotFound)
		}
		return nil, fmt.Errorf("failed to fetch beacon block at slot %d, http status %d", slot, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
		return nil, fmt.Errorf("failed to decode beacon block: %w", err)
	}
	return block, nil
}

// Returns the start time of an epoch
func EpochStartTime(epoch uint64) uint64 {
	return beaconChainGenesisTimeSeconds + epoch*beaconSlotsPerEpoch*beaconChainSlotDurationSeconds
}

// Returns the number of the epoch starting at or before the given time.
func EpochAtOrBefore(unixTime uint64) uint64 {
	return SlotAtOrBefore(unixTime) / beaconSlotsPerEpoch
}

func SlotAtOrBefore(unixTime uint64) uint64 {
	// Get the slot at or before the given time.
	// Slot = (start - genesis) / slotDuration
	return (unixTime - beaconChainGenesisTimeSeconds) / beaconChainSlotDurationSeconds
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

type Epoch struct {
	Status string `json:"status"`
	Data   struct {
		Attestationscount       int       `json:"attestationscount"`
		Attesterslashingscount  int       `json:"attesterslashingscount"`
		Averagevalidatorbalance int64     `json:"averagevalidatorbalance"`
		Blockscount             int       `json:"blockscount"`
		Depositscount           int       `json:"depositscount"`
		Eligibleether           int64     `json:"eligibleether"`
		Epoch                   int       `json:"epoch"`
		Finalized               bool      `json:"finalized"`
		Globalparticipationrate float64   `json:"globalparticipationrate"`
		Missedblocks            int       `json:"missedblocks"`
		Orphanedblocks          int       `json:"orphanedblocks"`
		Proposedblocks          int       `json:"proposedblocks"`
		Proposerslashingscount  int       `json:"proposerslashingscount"`
		RewardsExported         bool      `json:"rewards_exported"`
		Scheduledblocks         int       `json:"scheduledblocks"`
		Totalvalidatorbalance   int64     `json:"totalvalidatorbalance"`
		Ts                      time.Time `json:"ts"`
		Validatorscount         int       `json:"validatorscount"`
		Voluntaryexitscount     int       `json:"voluntaryexitscount"`
		Votedether              int64     `json:"votedether"`
		Withdrawalcount         int       `json:"withdrawalcount"`
	} `json:"data"`
}
