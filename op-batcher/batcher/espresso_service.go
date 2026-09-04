package batcher

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	espressoLightClient "github.com/EspressoSystems/espresso-network/sdks/go/light-client"
	espressoStreamers "github.com/EspressoSystems/espresso-streamers/op"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/hf/nitrite"

	"github.com/ethereum-optimism/optimism/op-batcher/enclave"
	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum-optimism/optimism/op-service/dial"
)

// EspressoBatcherConfig groups all Espresso-specific configuration the
// batcher consumes at runtime. It is embedded as a single field on
// BatcherConfig (see service.go) to keep the upstream Optimism
// BatcherConfig field block compact and minimize cherry-pick churn.
//
// Fields are populated from CLIConfig.Espresso (and a few RollupConfig
// fallbacks) by initEspresso below; ephemeral key material is generated
// by initKeyPair (TEE) or copied from CLIConfig.Espresso.TestingBatcherPrivateKey
// (devnet/test).
type EspressoBatcherConfig struct {
	Enabled                    bool
	PollInterval               time.Duration
	AttestationService         string
	CaffeinationHeightEspresso uint64
	// CaffeinationHeightL2 is the L2 batch position at which the Espresso
	// streamer should start emitting batches. Operational parameter for
	// starting batchers mid-chain (e.g. handing off from the fallback
	// batcher): waitForLocalSafeHead refuses to anchor the streamer until
	// the local-safe head reaches this height. When zero there is no floor:
	// the streamer anchors at the current local-safe head, which is correct
	// for steady-state restarts but provides no handoff protection, so
	// handoff starts must set --espresso.origin-height-l2 explicitly.
	CaffeinationHeightL2 uint64

	// Receipt verification tuning for the Espresso transaction submitter.
	VerifyReceiptMaxBlocks     uint64
	VerifyReceiptSafetyTimeout time.Duration
	VerifyReceiptRetryDelay    time.Duration

	// BatcherPublicKey/BatcherPrivateKey is the batcher's identity for the
	// Espresso authentication path. In TEE deployments the private key is
	// generated inside the enclave (initKeyPair) and the public key is
	// attested to via Nitro Enclave PCR0; outside TEE (devnet/test), the
	// configured TestingBatcherPrivateKey overrides them in initEspresso.
	BatcherPublicKey  *ecdsa.PublicKey
	BatcherPrivateKey *ecdsa.PrivateKey
}

// EspressoStreamer returns the Espresso batch streamer driven by this batcher.
func (bs *BatcherService) EspressoStreamer() *espressoStreamers.Streamer {
	return bs.driver.espressoStreamer
}

// initChainSigner builds the ChainSigner from the same signing configuration
// the txmgr consumes and stores it on the service. Espresso uses ChainSigner to
// sign batch authentication payloads sent to the BatchAuthenticator contract.
// The signer is only used for Sign (arbitrary-hash signing), so the chain ID and
// from address are taken from the already-built TxManager.
func (bs *BatcherService) initChainSigner(cfg *CLIConfig) error {
	if !cfg.Espresso.Enabled {
		return nil
	}
	tcfg := cfg.TxMgrConfig

	// Mirror the txmgr's backwards-compatible HD-path resolution.
	hdPath := tcfg.HDPath
	if hdPath == "" && tcfg.SequencerHDPath != "" {
		hdPath = tcfg.SequencerHDPath
	} else if hdPath == "" && tcfg.L2OutputHDPath != "" {
		hdPath = tcfg.L2OutputHDPath
	}

	factory, from, err := opcrypto.ChainSignerFactoryFromConfig(bs.Log, tcfg.PrivateKey, tcfg.Mnemonic, hdPath, tcfg.SignerCLIConfig)
	if err != nil {
		return fmt.Errorf("failed to init Espresso chain signer: %w", err)
	}
	bs.ChainSigner = factory(bs.TxManager.ChainID().ToBig(), from)
	return nil
}

// applyEspressoDriverSetup writes the Espresso-specific fields onto a
// DriverSetup populated with upstream-Optimism fields. Kept separate from the
// main initDriver struct literal so that the upstream block stays in upstream
// shape — minimizing cherry-pick churn when upstream renames or reorders
// fields.
func (bs *BatcherService) applyEspressoDriverSetup(ds *DriverSetup) {
	ds.Espresso.SequencerAddress = bs.TxManager.From()
	ds.Espresso.ChainSigner = bs.ChainSigner
	ds.Espresso.Client = bs.EspressoClient
	ds.Espresso.LightClient = bs.EspressoLightClient
	ds.Espresso.Attestation = bs.Attestation
}

// initKeyPair generates an ephemeral ECDSA key pair for the batcher's
// Espresso authentication path. In TEE deployments this key is attested
// to via Nitro Enclave PCR0; outside TEE (devnet/test), the configured
// TestingBatcherPrivateKey overrides this key in initEspresso.
func (bs *BatcherService) initKeyPair() error {
	key, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate key pair for batcher: %w", err)
	}
	bs.Espresso.BatcherPrivateKey = key
	bs.Espresso.BatcherPublicKey = &key.PublicKey
	return nil
}

// initEspresso configures the Espresso TEE-batcher integration on the
// BatcherService. When --espresso.enabled is false this is a no-op. When
// enabled, it wires up the Espresso query-service client, light client,
// ephemeral key pair, and Nitro Enclave attestation (if running in TEE).
func (bs *BatcherService) initEspresso(ctx context.Context, cfg *CLIConfig) error {
	if !cfg.Espresso.Enabled {
		return nil
	}

	if cfg.Espresso.L1URL == "" {
		log.Warn("Espresso L1 URL not provided, using batcher's L1EthRpc")
		cfg.Espresso.L1URL = cfg.L1EthRpc
	}
	if err := cfg.Espresso.Check(); err != nil {
		return fmt.Errorf("invalid Espresso config: %w", err)
	}

	bs.Espresso.Enabled = true
	bs.Espresso.PollInterval = cfg.Espresso.PollInterval
	bs.Espresso.AttestationService = cfg.Espresso.EspressoAttestationService
	bs.Espresso.CaffeinationHeightEspresso = cfg.Espresso.CaffeinationHeightEspresso
	bs.Espresso.CaffeinationHeightL2 = cfg.Espresso.CaffeinationHeightL2
	bs.Espresso.VerifyReceiptMaxBlocks = cfg.Espresso.VerifyReceiptMaxBlocks
	bs.Espresso.VerifyReceiptSafetyTimeout = cfg.Espresso.VerifyReceiptSafetyTimeout
	bs.Espresso.VerifyReceiptRetryDelay = cfg.Espresso.VerifyReceiptRetryDelay

	client, err := espressoClient.NewMultipleNodesClient(cfg.Espresso.QueryServiceURLs)
	if err != nil {
		return fmt.Errorf("failed to create Espresso client: %w", err)
	}
	bs.EspressoClient = client

	// Light-client reads go through the batcher's L1 client unless
	// --espresso.l1-url points at a different RPC endpoint, in which case a
	// dedicated client is dialed for them (and closed in Stop).
	lightClientBackend := bs.L1Client
	if cfg.Espresso.L1URL != cfg.L1EthRpc {
		espressoL1, err := dial.DialEthClientWithTimeout(ctx, dial.DefaultDialTimeout, bs.Log, cfg.Espresso.L1URL)
		if err != nil {
			return fmt.Errorf("failed to dial Espresso L1 RPC: %w", err)
		}
		bs.EspressoL1Client = espressoL1
		lightClientBackend = espressoL1
	}
	lightClient, err := espressoLightClient.NewLightclientCaller(cfg.Espresso.LightClientAddr, lightClientBackend)
	if err != nil {
		return fmt.Errorf("failed to create Espresso light client: %w", err)
	}
	bs.EspressoLightClient = lightClient

	if err := bs.initKeyPair(); err != nil {
		return fmt.Errorf("failed to create key pair for batcher: %w", err)
	}

	// try to generate attestationBytes on public key when start batcher
	attestationBytes, err := enclave.AttestationWithPublicKey(bs.Espresso.BatcherPublicKey)
	if err != nil {
		bs.Log.Info("Not running in enclave, skipping attestation", "info", err)

		// Replace ephemeral keys with configured keys, as in devnet they'll be pre-approved for batching
		privateKey := cfg.Espresso.TestingBatcherPrivateKey
		if privateKey == nil {
			return fmt.Errorf("when not running in enclave, testing batcher private key should be set")
		}

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("error casting public key to ECDSA")
		}

		bs.Espresso.BatcherPrivateKey = privateKey
		bs.Espresso.BatcherPublicKey = publicKeyECDSA
	} else {
		// output length of attestation
		bs.Log.Info("Successfully got attestation. Attestation length", "length", len(attestationBytes))
		_, err := nitrite.Verify(attestationBytes, nitrite.VerifyOptions{})
		if err != nil {
			return fmt.Errorf("Couldn't verify attestation: %w", err)
		}
		bs.Attestation = attestationBytes
	}

	return nil
}
