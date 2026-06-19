// This file is dedicated to a series of tests related to soft confirmation
// integrity. This file contains tests related to, and ensuring that the
// integrity of the soft confirmations being provided by the underlying
// Rollup when compared against the confirmations being provided by
// Espresso/HotShot. The derivation from the L2 / L1 should not be compromised
// or result in different results than the derivation provided by the
// sequencer.
//
// Assumption: The rollup sequencer is correct, online, and honest. It
// produces a valid sequence of rollup blocks every few seconds or faster,
// and it never reorgs.
//
// The underlying documented definition of the soft confirmation integrity
// comes from this definition:
//	The integration must not weaken soft confirmations provided by a rollup's
//	sequencer. That is, while Espresso confirmations are valid even under a
//	weakened security assumption where the sequencer may be malicious, if we
//	consider the case with a stronger assumption where the sequencer is correct,
//	online, and honest, the rollup should finalize what the sequencer produces.

package environment_test

import (
	"context"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"math/big"
	"net"
	"net/url"
	"testing"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoCommon "github.com/EspressoSystems/espresso-network/sdks/go/types/common"
	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	op_crypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	op_signer "github.com/ethereum-optimism/optimism/op-service/signer"
	ethereum "github.com/ethereum/go-ethereum"
	geth_common "github.com/ethereum/go-ethereum/common"
	geth_types "github.com/ethereum/go-ethereum/core/types"
	geth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

// messageWithTimestamp is a struct that contains an entry of type T
// and a timestamp. It is used to store messages with their corresponding
// timestamps.
type messageWithTimestamp[T any] struct {
	entry     T
	timestamp time.Time
}

// recordTimestamp is a helper function that takes an entry of type T and
// returns a messageWithTimestamp[T] struct that contains the entry and
// the current timestamp.
func recordTimestamp[T any](entry T) messageWithTimestamp[T] {
	return messageWithTimestamp[T]{
		entry:     entry,
		timestamp: time.Now(),
	}
}

// produceTimestampedStream is a helper function that is designed to be run
// in a goroutine.  It consumes values coming from the input channel and
// outputs them to the output channel with a timestamp.
func produceTimestampedStream[T any](
	ctx context.Context,
	input <-chan T,
	output chan<- messageWithTimestamp[T],
) {
	for {
		select {
		case <-ctx.Done():
			return
		case header, ok := <-input:
			if !ok {
				// Channel is closed,
				// we should exit
				return
			}

			select {
			case <-ctx.Done():
			case output <- recordTimestamp(header):
			}
		}
	}
}

// timestampedHeaderStream is a struct that contains a subscription
// to the Ethereum client and a channel to receive timestamped headers.
type timestampedHeaderStream struct {
	sub ethereum.Subscription
	ch  chan messageWithTimestamp[*geth_types.Header]
}

// setupHeaderStreamSubscription sets up a subscription to the new head
// event on the given Ethereum client. It creates a channel to receive
// headers and a channel to receive timestamped headers. It starts a
// goroutine to produce timestamped headers from the received headers.
// It returns a timestampedHeaderStream struct containing the subscription
// and the channel for timestamped headers.
func setupHeaderStreamSubscription(ctx context.Context, t *testing.T, cli *ethclient.Client) (timestampedHeaderStream, error) {
	headerCh := make(chan *geth_types.Header)
	timestampedHeaderCh := make(chan messageWithTimestamp[*geth_types.Header])
	sub, err := cli.SubscribeNewHead(ctx, headerCh)
	if err != nil {
		return timestampedHeaderStream{sub: sub}, err
	}
	go produceTimestampedStream(ctx, headerCh, timestampedHeaderCh)

	return timestampedHeaderStream{sub: sub, ch: timestampedHeaderCh}, nil

}

// setupHeaderStreamSubscriptions sets up subscriptions to the new head
// event on the given Ethereum clients (sequencer and verifier).
func setupHeaderStreamSubscriptions(ctx context.Context, t *testing.T, l2Seq, l2Verif *ethclient.Client) (
	seqStream timestampedHeaderStream,
	verifStream timestampedHeaderStream,
) {

	seqStream, err := setupHeaderStreamSubscription(ctx, t, l2Seq)
	if have, want := err, error(nil); have != want {
		t.Fatalf("Failed to subscribe to sequencer new head:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	verifStream, err = setupHeaderStreamSubscription(ctx, t, l2Verif)
	if have, want := err, error(nil); have != want {
		t.Fatalf("Failed to subscribe to verifier new head:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	return seqStream, verifStream
}

// nextStreamEntries is a helper function that retrieves the next entries
// from the sequencer and verifier streams.
func nextStreamEntries[T any](ctx context.Context, seqCh, verifCh <-chan messageWithTimestamp[T]) (
	seqHeader, verifHeader messageWithTimestamp[T],
) {
	select {
	case <-ctx.Done():
		return

	case seqHeader = <-seqCh:
	}

	select {
	case <-ctx.Done():
		return
	case verifHeader = <-verifCh:
	}

	return seqHeader, verifHeader
}

// advanceStreamToHeight is a helper function that advances the
// timestampedHeaderStream to the specified height. It consumes headers
// from the stream until the block number of the header is greater than
// or equal to the specified height.
func advanceStreamToHeight(
	ctx context.Context,
	stream timestampedHeaderStream,
	start messageWithTimestamp[*geth_types.Header],
	height *big.Int,
) {
	for i := start.entry.Number; i.Cmp(height) < 0; {
		select {
		case <-ctx.Done():
			return
		case streamHeader, ok := <-stream.ch:
			if !ok {
				return
			}

			i = streamHeader.entry.Number
		}
	}
}

// EnsureStreamsAreSynced is a helper function that ensures that the
// sequencer and verifier streams are all at the same block height.
// It does this by advancing each stream to the largest block number
// among the two streams.
//
// Advancing the streams to the same block height is necessary as it ensures
// that we are comparing the same block across both streams.
//
// Advancing in this way does skip over existing blocks, so there is a
// potential for missing blocks in this way.
func ensureStreamsAreSynced(
	ctx context.Context,
	seqStream, verifStream timestampedHeaderStream,
) {
	seqHeader, verifHeader := nextStreamEntries(ctx, seqStream.ch, verifStream.ch)

	// Determine the largest block from the two streams
	var largestNumber = seqHeader.entry.Number
	if verifHeader.entry.Number.Cmp(largestNumber) > 0 {
		largestNumber = verifHeader.entry.Number
	}

	// Now advance all of these streams so that the last entry consumed
	// all point to the same block number.

	// Advance the Sequencer Stream
	advanceStreamToHeight(ctx, seqStream, seqHeader, largestNumber)
	advanceStreamToHeight(ctx, verifStream, verifHeader, largestNumber)
}

// verifyStreamSequenceForNextN is a helper function that verifies
// the sequence of blocks being produced by the sequencer and verifier
// streams all match for the next N blocks.
//
// It does this by waiting for the next entry from each stream and
// comparing their header values.
//
// The sequence being consumed should be ordered, and the same across both
// streams.
//
// The streams are assumed to be synced before this function is called.
// This means that they should be at the same block height before this
// verification is called, otherwise we may fail due to being on different
// block heights.
func verifyStreamSequenceForNextN(
	ctx context.Context,
	t *testing.T,
	seqStream, verifStream timestampedHeaderStream,
	count int,
) {
	for i := 0; i < count; i++ {
		// The easiest way to verify this is to just wait for each of these
		// streams entries in turn, then compare their header hashes.

		seqHeader, verifHeader := nextStreamEntries(ctx, seqStream.ch, verifStream.ch)

		// Alright, we should have both next headers now.
		// Let's compare them to make sure they are the same.
		select {
		case <-ctx.Done():
			t.Errorf("test was canceled by context while waiting to verify sequence entry %d", i)
			return
		default:
		}

		if have, want := seqHeader.entry.Hash(), verifHeader.entry.Hash(); have.Cmp(want) != 0 {
			t.Fatalf("Sequencer and Verifier headers do not match:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
			return
		}
	}
}

// SUBMIT_RANDOM_DATA_INTERVAL is the interval / frequency at which we
// will attempt to submit random data to the Espresso using the
// sequencer's namespace.
const SUBMIT_RANDOM_DATA_INTERVAL = 500 * time.Millisecond

// submitRandomDataToSequencerNamespace is a function that submits
// random data to the sequencer namespace at a specified interval.
func submitRandomDataToSequencerNamespace(ctx context.Context, espCli espressoClient.EspressoClient, namespace uint64) {
	// We only want to submit garbage data to the sequencer so quickly
	ticker := time.NewTicker(SUBMIT_RANDOM_DATA_INTERVAL)
	buffer := make([]byte, 1024*3)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		// Fill buffer with random data
		n, _ := crypto_rand.Read(buffer)

		// Submit garbage data to the sequencer namespace
		_, err := espCli.SubmitTransaction(ctx, espressoCommon.Transaction{
			Namespace: namespace,
			Payload:   espressoCommon.Bytes(buffer[:n]),
		})
		if err != nil {
			log.Error("Failed to submit random data to sequencer namespace", "namespace", namespace, "error", err)
		}
	}
}

// createMaliciousEspressoBatch creates a malicious Espresso batch by
// constructing a block with a deposit transaction. It uses the latest
// block from the sequencer to create a new block with a deposit
// transaction. The block is then converted to an Espresso batch using
// the derive.BlockToEspressoBatch function.
func createMaliciousEspressoBatch(ctx context.Context, cli *ethclient.Client, rollupCfg *rollup.Config) (*derive.EspressoBatch, error) {
	// / Determine what the latest block in the sequencer is, so we can
	// hope to create a valid transaction, to get something out of it.
	latestBlock, err := cli.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	latestHeader := latestBlock.Header()
	header := &geth_types.Header{
		ParentHash: latestBlock.Hash(),
		UncleHash:  latestHeader.UncleHash,
		Coinbase:   latestHeader.Coinbase,
		Root:       latestHeader.Root,
		Bloom:      latestHeader.Bloom,
		Difficulty: latestHeader.Difficulty,
		Number:     new(big.Int).Add(latestBlock.Number(), big.NewInt(1)),
		GasLimit:   latestHeader.GasLimit,
		GasUsed:    latestHeader.GasUsed,
		Time:       latestHeader.Time + 1,
		Extra:      latestHeader.Extra,
		MixDigest:  latestHeader.MixDigest,
		Nonce:      latestHeader.Nonce,
	}
	body := geth_types.Body{
		Transactions: []*geth_types.Transaction{
			geth_types.NewTx(
				&geth_types.DepositTx{
					Value: big.NewInt(1000),
				},
			),
		},
	}
	block := geth_types.NewBlockWithHeader(header).WithBody(body)

	return derive.BlockToEspressoBatch(rollupCfg, block)
}

// SUBMIT_VALID_DATA_WITH_WRONG_SIGNATURE_INTERVAlL is the interval / frequency
// at which we will attempt to submit valid data with the wrong signature to the
// Espresso using the sequencer's namespace.
const SUBMIT_VALID_DATA_WITH_WRONG_SIGNATURE_INTERVAlL = 500 * time.Millisecond

// Attack Espresso Integrity by Submitting Valid Data with the wrong
// Signature to the Sequencer's namespace.
func submitValidDataWithWrongSignature(ctx context.Context, rollupCfg *rollup.Config, l2Seq *ethclient.Client, espCli espressoClient.EspressoClient, namespace uint64) {
	// We only want to submit garbage data to the sequencer so quickly
	ticker := time.NewTicker(SUBMIT_VALID_DATA_WITH_WRONG_SIGNATURE_INTERVAlL)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		privateKey, err := geth_crypto.GenerateKey()
		if err != nil {
			continue
		}
		privateKeyString := hex.EncodeToString(geth_crypto.FromECDSA(privateKey))
		factory, _, err := op_crypto.ChainSignerFactoryFromConfig(nil, privateKeyString, "", "", op_signer.CLIConfig{})
		if err != nil {
			continue
		}
		randomChainSigner := factory(big.NewInt(int64(namespace)), geth_common.Address{})

		batch, err := createMaliciousEspressoBatch(ctx, l2Seq, rollupCfg)

		if err != nil {
			// Skip
			continue
		}

		txn, err := batch.ToEspressoTransaction(ctx, namespace, randomChainSigner)
		if err != nil {
			// Skip
			continue
		}

		// Submit garbage data to the sequencer namespace
		_, _ = espCli.SubmitTransaction(ctx, *txn)
	}
}

// fakeChainSigner is a fake implementation of the ChainSigner interface.
// It will create fake signatures for the transaction.
type fakeChainSigner struct{}

var _ op_crypto.ChainSigner = (*fakeChainSigner)(nil)

// Sign implements crypto.ChainSigner.
func (f *fakeChainSigner) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	sig := make([]byte, geth_crypto.SignatureLength)
	_, err := crypto_rand.Read(sig)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// SignTransaction implements crypto.ChainSigner.
func (f *fakeChainSigner) SignTransaction(
	ctx context.Context,
	addr geth_common.Address,
	tx *geth_types.Transaction,
) (*geth_types.Transaction, error) {
	// This is a fake implementation, and we're not expecting this method to be
	// called in this test, so this should be safe.
	panic("unimplemented")
}

// SUBMIT_VALID_DATA_WITH_RANDOM_SIGNATURE_INTERVAL is the interval / frequency
// at which we will attempt to submit valid data with a random signature to the
// Espresso using the sequencer's namespace.
const SUBMIT_VALID_DATA_WITH_RANDOM_SIGNATURE_INTERVAL = 100 * time.Millisecond

// Attack Espresso Integrity by Submitting A properly formatted
// transaction, with a random signature value to the Sequencer's
// namespace
func submitValidDataWithRandomSignature(
	ctx context.Context,
	rollupCfg *rollup.Config,
	l2Seq *ethclient.Client,
	espCli espressoClient.EspressoClient,
	namespace uint64,
) {
	// We only want to submit garbage data to the sequencer so quickly
	ticker := time.NewTicker(SUBMIT_VALID_DATA_WITH_RANDOM_SIGNATURE_INTERVAL)
	signer := new(fakeChainSigner)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		batch, err := createMaliciousEspressoBatch(ctx, l2Seq, rollupCfg)

		if err != nil {
			// Skip
			continue
		}

		txn, err := batch.ToEspressoTransaction(ctx, namespace, signer)
		if err != nil {
			// Skip
			continue
		}

		// Submit garbage data to the sequencer namespace
		_, _ = espCli.SubmitTransaction(ctx, *txn)
	}
}

// TestSequencerFeedConsistency is a test that ensures that the sequence of
// blocks being produced by the feeds from the Sequencer and another L2
// Verifier are consistent with each other.
//
// The criteria / goal of this test are outlined by the following requirement:
//
// Run the rollup and subscribe to the sequencer feed and a feed which derives
// the finalized block sequence from L1. All of these should yield the same
// blocks in the same order (but at different times).
func TestSequencerFeedConsistency(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)
	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t, env.WithL1FinalizedDistance(0))

	// Signal the testnet to shut down
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	l2Seq := system.NodeClient(e2esys.RoleSeq)
	l2Verif := system.NodeClient(e2esys.RoleVerif)

	seqStream, verifStream := setupHeaderStreamSubscriptions(ctx, t, l2Seq, l2Verif)
	defer seqStream.sub.Unsubscribe()
	defer verifStream.sub.Unsubscribe()

	// We need to sync these streams up.  We created them at different points
	// in their life times. so we need to wait for them all to be at the same
	// block height before we start comparing them.
	//
	// It is most likely going to be the case that the sequencer is ahead of
	// the verifier.  We will play it safe, and just make no assumptions by
	// grabbing the largest block
	ensureStreamsAreSynced(ctx, seqStream, verifStream)

	// Let's verify that these streams are producing the same blocks
	// in the same order. We will do this by waiting for a few blocks to
	verifyStreamSequenceForNextN(ctx, t, seqStream, verifStream, 100)
}

// TestSequencerFeedConsistencyWithAttackOnEspresso is a test that expands
// upon the previous test by introducing attacks against Espresso with the
// specific goal of arriving at a state where the Espresso feed is producing
// different blocks than the sequencer, for a variety of different potential
// reasons.
//
// These attacks are designed to cover some different use cases, and may
// reflect attempts of third parties to attack or manipulate the data being
// consumed for individual gain, or disruption.
//
// The criteria / goal of this test are outlined by the following requirement:
// Consider rollup-specific adversarial behavior which could break sequencer
// confirmations, such as an adversary sending non-sequencer blocks directly
// to Espresso. Such attacks should not cause Espresso to finalize something
// different than the sequencer feed.
func TestSequencerFeedConsistencyWithAttackOnEspresso(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	launcher := new(env.EspressoDevNodeLauncherDocker)
	system, espressoDevNode, err := launcher.StartE2eDevnet(ctx, t, env.WithL1FinalizedDistance(0))

	// Signal the testnet to shut down
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to start dev environment with espresso dev node:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	defer env.Stop(t, system)
	defer env.Stop(t, espressoDevNode)

	_, port, err := net.SplitHostPort(espressoDevNode.SequencerPort())
	if have, want := err, error(nil); have != want {
		t.Fatalf("failed to parse sequencer port URL:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
	}

	espressoSequencerURL := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort("localhost", port),
		Path:   "/",
	}

	l2Seq := system.NodeClient(e2esys.RoleSeq)
	espCli := espressoClient.NewClient(espressoSequencerURL.String())
	namespace := system.RollupConfig.L2ChainID.Uint64()

	// Attack Espresso Integrity by Submitting Garbage Data to the Same
	// namespace as the Sequencer's namespace.
	go submitRandomDataToSequencerNamespace(ctx, espCli, namespace)

	// Attack Espresso Integrity by Submitting Valid Data with the wrong
	// Signature to the Sequencer's namespace.
	go submitValidDataWithWrongSignature(ctx, system.RollupConfig, l2Seq, espCli, namespace)

	// Attack Espresso Integrity by Submitting A properly formatted
	// transaction, with a random signature value to the Sequencer's
	// namespace
	go submitValidDataWithRandomSignature(ctx, system.RollupConfig, l2Seq, espCli, namespace)

	l2Verif := system.NodeClient(e2esys.RoleVerif)

	seqStream, verifStream := setupHeaderStreamSubscriptions(ctx, t, l2Seq, l2Verif)
	defer seqStream.sub.Unsubscribe()
	defer verifStream.sub.Unsubscribe()

	// Sync the Streams to the same block height
	ensureStreamsAreSynced(ctx, seqStream, verifStream)

	// Verify the sequence of blocks being produced.
	verifyStreamSequenceForNextN(ctx, t, seqStream, verifStream, 100)
}
