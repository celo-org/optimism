package derive

import (
	"bytes"
	_ "embed"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	optypes "github.com/ethereum-optimism/optimism/op-core/types"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

// CeloTokenL1.TOTAL_MARKET_CAP: 1 billion CELO, with 18 decimals.
// Source: celo-contracts/v6.0.0 at 0296412d673a666af6236efa1f615032b1f1c424.
const cgtV2TotalSupply = "1000000000000000000000000000"

// TODO(celo-org/celo-blockchain-planning#1466): replace for each network after its L1 bridge deployment.
// This temporary address is not a deployed production bridge.
var cgtV2TemporaryL1Bridge = common.HexToAddress("0x00000000000000000000000000000000c6700001")

type cgtV2NetworkParams struct {
	bridge      common.Address
	totalSupply string
	bundle      []byte
	// Register new hashes before emitting the first retry request for them on L1.
	retryBundles []cgtV2RetryBundle
}

type cgtV2RetryBundle struct {
	// The protocol hash is pinned independently of the embedded artifact.
	hash   common.Hash
	bundle []byte
}

func (p cgtV2NetworkParams) retryBundle(hash common.Hash) *cgtV2RetryBundle {
	if hash == (common.Hash{}) {
		return nil
	}
	for i := range p.retryBundles {
		retry := &p.retryBundles[i]
		if retry.hash == hash {
			return retry
		}
	}
	return nil
}

// Production profiles never select the synthetic NUT bundle.
var cgtV2Networks = map[uint64]cgtV2NetworkParams{
	params.CeloMainnetChainID: {bridge: cgtV2TemporaryL1Bridge, totalSupply: cgtV2TotalSupply, bundle: cgtV2MainnetBundle},
	params.CeloSepoliaChainID: {bridge: cgtV2TemporaryL1Bridge, totalSupply: cgtV2TotalSupply, bundle: cgtV2SepoliaBundle},
	params.CeloChaosChainID:   {bridge: cgtV2TemporaryL1Bridge, totalSupply: cgtV2TotalSupply, bundle: cgtV2ChaosBundle},
}

// Synthetic profile matching the frozen contract/NUT fixtures, not a production default.
const cgtV2FixtureChainID uint64 = 1337

var cgtV2FixtureParams = cgtV2NetworkParams{
	bridge:      common.HexToAddress("0xcc"),
	totalSupply: "1000",
	retryBundles: []cgtV2RetryBundle{{
		hash:   common.HexToHash("0x437fabe614e74fd20231c33accdacbad38631aaefc16abdb3c38c3bc5be10f14"),
		bundle: cgtV2FixtureBundle,
	}},
}

func cgtV2Params(chainID uint64) (cgtV2NetworkParams, bool) {
	if chainID == cgtV2FixtureChainID {
		network := cgtV2FixtureParams
		network.bundle = cgtV2FixtureBundle
		return network, true
	}
	network, ok := cgtV2Networks[chainID]
	return network, ok
}

func (p cgtV2NetworkParams) reserveSeed(migratedAmount *big.Int) (*big.Int, bool) {
	supply, ok := new(big.Int).SetString(p.totalSupply, 10)
	if !ok || migratedAmount == nil || migratedAmount.Sign() < 0 || migratedAmount.Cmp(supply) > 0 {
		return nil, false
	}
	return supply.Sub(supply, migratedAmount), true
}

// Synthetic chain 1337 only. Production NUT bundles remain unresolved.
// Provenance and cross-client vectors are stored beside this fixture.
//
//go:embed testdata/cgtv2/cgtv2-fixture-nut-bundle.json
var cgtV2FixtureBundle []byte

//go:embed upgrades/cgtv2/mainnet.json
var cgtV2MainnetBundle []byte

//go:embed upgrades/cgtv2/sepolia.json
var cgtV2SepoliaBundle []byte

//go:embed upgrades/cgtv2/chaos.json
var cgtV2ChaosBundle []byte

var cgtV2Intents = [...]string{
	"L1BlockCGT deployment",
	"L2ToL1MessagePasserCGT deployment",
	"NativeAssetLiquidity deployment",
	"LiquidityController deployment",
	"CeloGasBridgeL2 deployment",
	"CeloSequencerFeeVault deployment",
	"L1BlockCGT upgrade",
	"L2ToL1MessagePasserCGT upgrade",
	"NativeAssetLiquidity upgrade",
	"LiquidityController upgrade and initialize",
	"CeloGasBridgeL2 upgrade and initialize",
	"CeloSequencerFeeVault upgrade and initialize",
	"CeloGasBridgeL2 minter authorization",
	"LiquidityController ownership transfer",
	"CGTv2 flag enable",
}

// Fixture gas limits: six 6M deployments, nine 500k direct calls,
// and 100k reserve funding. The v6 ProxyAdmin remains unchanged.
// TODO(celo-org/celo-blockchain-planning#1466): calibrate limits against final contracts and production state,
// then update the bundle and matching Go/Rust group allowances.
const cgtV2FundingGas uint64 = 100_000
const cgtV2UpgradeGas uint64 = 40_600_000

var cgtV2FundingSource = (&UpgradeDepositSource{Intent: "Celo: CGTv2 Reserve Funding"}).SourceHash()

func cgtV2NUTBundle(chainID uint64) (*nutBundle, error) {
	network, ok := cgtV2Params(chainID)
	if !ok {
		return nil, fmt.Errorf("CGTv2 unsupported chain %d", chainID)
	}
	return readCgtV2NUTBundle(chainID, network.bundle)
}

func readCgtV2NUTBundle(chainID uint64, artifact []byte) (*nutBundle, error) {
	if len(artifact) == 0 {
		return nil, fmt.Errorf("CGTv2 upgrade bundle is unresolved for chain %d", chainID)
	}
	bundle, err := readNUTBundle(forks.Name("cgtv2"), bytes.NewReader(artifact))
	if err != nil {
		return nil, err
	}
	if len(bundle.Transactions) == 0 {
		return nil, fmt.Errorf("CGTv2 upgrade bundle is unresolved for chain %d", chainID)
	}
	if len(bundle.Transactions) != len(cgtV2Intents) {
		return nil, fmt.Errorf("CGTv2 invalid static transaction count")
	}
	for i, tx := range bundle.Transactions {
		if tx.Intent != cgtV2Intents[i] || tx.GasLimit == 0 {
			return nil, fmt.Errorf("CGTv2 invalid static transaction %d", i)
		}
	}
	return bundle, nil
}

func cgtV2BundleGas(bundle *nutBundle) (uint64, error) {
	total := cgtV2FundingGas
	for _, tx := range bundle.Transactions {
		if tx.GasLimit > ^uint64(0)-total {
			return 0, fmt.Errorf("CGTv2 upgrade gas overflow")
		}
		total += tx.GasLimit
	}
	return total, nil
}

// CgtV2UpgradeGas returns the selected network bundle's checked one-time allowance.
func CgtV2UpgradeGas(chainID *big.Int) (uint64, error) {
	if chainID == nil || !chainID.IsUint64() {
		return 0, fmt.Errorf("CGTv2 unsupported chain %v", chainID)
	}
	bundle, err := cgtV2NUTBundle(bigs.Uint64Strict(chainID))
	if err != nil {
		return 0, err
	}
	return cgtV2BundleGas(bundle)
}

// CgtV2UpgradeTransactions builds the original upgrade deposits and dynamic reserve funding.
func CgtV2UpgradeTransactions(chainID, migratedAmount *big.Int) ([]hexutil.Bytes, uint64, error) {
	chain, network, err := cgtV2ParamsForChain(chainID)
	if err != nil {
		return nil, 0, err
	}
	return network.upgradeTransactions(chain, migratedAmount, nil)
}

func cgtV2ParamsForChain(chainID *big.Int) (uint64, cgtV2NetworkParams, error) {
	if chainID == nil || !chainID.IsUint64() {
		return 0, cgtV2NetworkParams{}, fmt.Errorf("CGTv2 unsupported chain %v", chainID)
	}
	chain := bigs.Uint64Strict(chainID)
	network, ok := cgtV2Params(chain)
	if !ok {
		return 0, cgtV2NetworkParams{}, fmt.Errorf("CGTv2 unsupported chain %v", chainID)
	}
	return chain, network, nil
}

func cgtV2StaticTransactions(chainID uint64, artifact []byte) ([]hexutil.Bytes, uint64, error) {
	bundle, err := readCgtV2NUTBundle(chainID, artifact)
	if err != nil {
		return nil, 0, err
	}
	gas, err := cgtV2BundleGas(bundle)
	if err != nil {
		return nil, 0, err
	}
	txs, err := bundle.toDepositTransactions()
	return txs, gas, err
}

func cgtV2BundleHash(static []hexutil.Bytes) (common.Hash, error) {
	raw, err := rlp.EncodeToBytes(static)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(raw), nil
}

func (r *cgtV2RetryBundle) transactions(chainID uint64) ([]hexutil.Bytes, uint64, error) {
	static, gas, err := cgtV2StaticTransactions(chainID, r.bundle)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid local CGTv2 retry artifact for %s: %w", r.hash, err)
	}
	hash, err := cgtV2BundleHash(static)
	if err != nil {
		return nil, 0, err
	}
	if hash != r.hash {
		return nil, 0, fmt.Errorf("local CGTv2 retry artifact does not match protocol hash %s", r.hash)
	}
	return static, gas, nil
}

func (p cgtV2NetworkParams) upgradeTransactions(chainID uint64, migratedAmount *big.Int, retry *cgtV2RetryBundle) ([]hexutil.Bytes, uint64, error) {
	reserve, ok := p.reserveSeed(migratedAmount)
	if !ok {
		return nil, 0, fmt.Errorf("CGTv2 migrated amount exceeds fixed total supply or is invalid")
	}
	var txs []hexutil.Bytes
	var gas uint64
	var err error
	if retry == nil {
		txs, gas, err = cgtV2StaticTransactions(chainID, p.bundle)
	} else {
		txs, gas, err = retry.transactions(chainID)
	}
	if err != nil {
		return nil, 0, err
	}
	raw, err := cgtV2FundingDeposit(reserve).MarshalBinary()
	if err != nil {
		return nil, 0, err
	}
	return append(txs, raw), gas, nil
}

func cgtV2FundingDeposit(reserve *big.Int) *optypes.DepositTx {
	target := predeploys.NativeAssetLiquidityAddr
	return &optypes.DepositTx{
		SourceHash: cgtV2FundingSource,
		From:       predeploys.LiquidityControllerAddr, To: &target,
		Mint: new(big.Int).Set(reserve), Value: reserve, Gas: cgtV2FundingGas,
		Data: crypto.Keccak256([]byte("deposit()"))[:4],
	}
}

var cgtV2GroupSources = func() map[common.Hash]struct{} {
	sources := map[common.Hash]struct{}{cgtV2FundingSource: {}}
	for i, intent := range cgtV2Intents {
		qualified := fmt.Sprintf("Cgtv2 %d: %s", i, intent)
		sources[(&UpgradeDepositSource{Intent: qualified}).SourceHash()] = struct{}{}
	}
	return sources
}()

func containsCgtV2Group(txs []hexutil.Bytes) bool {
	return cgtV2GroupStart(txs) >= 0
}

func cgtV2GroupStart(txs []hexutil.Bytes) int {
	for i, raw := range txs {
		if len(raw) == 0 || raw[0] != optypes.DepositTxType {
			continue
		}
		tx, err := optypes.UnmarshalDepositTx(raw)
		if err != nil {
			continue
		}
		if _, ok := cgtV2GroupSources[tx.SourceHash]; ok {
			return i
		}
	}
	return -1
}

// cgtV2PayloadUpgradeGas matches the complete suffix against a pinned artifact.
func cgtV2PayloadUpgradeGas(chainID *big.Int, txs []hexutil.Bytes) (uint64, error) {
	start := cgtV2GroupStart(txs)
	if start < 0 {
		return 0, nil
	}
	if start == 0 || len(txs)-start != len(cgtV2Intents)+1 {
		return 0, fmt.Errorf("invalid CGTv2 upgrade group length")
	}
	chain, network, err := cgtV2ParamsForChain(chainID)
	if err != nil {
		return 0, err
	}
	last := txs[len(txs)-1]
	funding, err := optypes.UnmarshalDepositTx(last)
	if err != nil {
		return 0, err
	}
	if _, ok := network.reserveSeed(funding.Value); !ok {
		return 0, fmt.Errorf("invalid CGTv2 reserve funding amount")
	}
	expected, err := cgtV2FundingDeposit(funding.Value).MarshalBinary()
	if err != nil || !bytes.Equal(last, expected) {
		return 0, fmt.Errorf("invalid CGTv2 reserve funding deposit")
	}

	hash, err := cgtV2BundleHash(txs[start : len(txs)-1])
	if err != nil {
		return 0, err
	}
	// The original artifact remains valid even when also registered for retries.
	if original, gas, err := cgtV2StaticTransactions(chain, network.bundle); err == nil {
		if originalHash, err := cgtV2BundleHash(original); err == nil && originalHash == hash {
			return gas, nil
		}
	}
	// The receipt scan enforces eligibility. The parent payload identifies the artifact by its bytes.
	if retry := network.retryBundle(hash); retry != nil {
		_, gas, err := retry.transactions(chain)
		return gas, err
	}
	return 0, fmt.Errorf("CGTv2 upgrade group does not match a pinned bundle")
}
