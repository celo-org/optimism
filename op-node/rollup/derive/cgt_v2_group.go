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
)

// Synthetic chain 1337 only. Production NUT bundles remain unresolved.
// Provenance and cross-client vectors are stored beside this fixture.
//
//go:embed testdata/cgtv2/cgtv2-fixture-nut-bundle.json
var cgtV2FixtureBundle []byte

const cgtV2FundingGas uint64 = 100_000
const cgtV2UpgradeGas uint64 = 56_300_000

var cgtV2FundingSource = (&UpgradeDepositSource{Intent: "Celo: CGTv2 Reserve Funding"}).SourceHash()

func cgtV2NUTBundle() (*nutBundle, error) {
	return readNUTBundle(forks.Name("cgtv2"), bytes.NewReader(cgtV2FixtureBundle))
}

// CgtV2UpgradeTransactions builds static upgrade deposits and dynamic reserve funding.
// The exact amount is authenticated by derivation from the EscrowSeeded receipt.
func CgtV2UpgradeTransactions(chainID, migratedAmount *big.Int) ([]hexutil.Bytes, uint64, error) {
	if chainID == nil || !chainID.IsUint64() {
		return nil, 0, fmt.Errorf("CGTv2 unsupported chain %v", chainID)
	}
	network, ok := cgtV2Params(bigs.Uint64Strict(chainID))
	if !ok {
		return nil, 0, fmt.Errorf("CGTv2 unsupported chain %v", chainID)
	}
	// A network profile does not authorize running fixture contracts on that network.
	if bigs.Uint64Strict(chainID) != cgtV2FixtureChainID {
		return nil, 0, fmt.Errorf("CGTv2 upgrade bundle is unresolved for chain %v", chainID)
	}
	reserve, ok := network.reserveSeed(migratedAmount)
	if !ok {
		return nil, 0, fmt.Errorf("CGTv2 migrated amount exceeds fixed total supply or is invalid")
	}
	bundle, err := cgtV2NUTBundle()
	if err != nil {
		return nil, 0, err
	}
	txs, err := bundle.toDepositTransactions()
	if err != nil {
		return nil, 0, err
	}
	target := predeploys.NativeAssetLiquidityAddr
	funding := &optypes.DepositTx{
		SourceHash: cgtV2FundingSource,
		From:       predeploys.LiquidityControllerAddr, To: &target,
		Mint: new(big.Int).Set(reserve), Value: reserve, Gas: cgtV2FundingGas,
		Data: crypto.Keccak256([]byte("deposit()"))[:4],
	}
	raw, err := funding.MarshalBinary()
	if err != nil {
		return nil, 0, err
	}
	return append(txs, raw), bundle.totalGas() + cgtV2FundingGas, nil
}

var cgtV2GroupSources = func() map[common.Hash]struct{} {
	bundle, err := cgtV2NUTBundle()
	if err != nil {
		panic(fmt.Sprintf("invalid embedded CGTv2 NUT fixture: %v", err))
	}
	sources := map[common.Hash]struct{}{cgtV2FundingSource: {}}
	for i, tx := range bundle.Transactions {
		intent := fmt.Sprintf("%s %d: %s", capitalizeForkName(bundle.ForkName), i, tx.Intent)
		sources[(&UpgradeDepositSource{Intent: intent}).SourceHash()] = struct{}{}
	}
	return sources
}()

func containsCgtV2Group(txs []hexutil.Bytes) bool {
	for _, raw := range txs {
		if len(raw) == 0 || raw[0] != optypes.DepositTxType {
			continue
		}
		tx, err := optypes.UnmarshalDepositTx(raw)
		if err != nil {
			continue
		}
		if _, ok := cgtV2GroupSources[tx.SourceHash]; ok {
			return true
		}
	}
	return false
}
