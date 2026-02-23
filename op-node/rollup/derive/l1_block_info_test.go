package derive

import (
	crand "crypto/rand"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

var (
	MockDepositContractAddr               = common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeef00000000")
	_                       eth.BlockInfo = (*testutils.MockBlockInfo)(nil)
)

type infoTest struct {
	name    string
	mkInfo  func(rng *rand.Rand) *testutils.MockBlockInfo
	mkL1Cfg func(rng *rand.Rand, l1Info eth.BlockInfo) eth.SystemConfig
	seqNr   func(rng *rand.Rand) uint64
}

func randomL1Cfg(rng *rand.Rand, l1Info eth.BlockInfo) eth.SystemConfig {
	return eth.SystemConfig{
		BatcherAddr: testutils.RandomAddress(rng),
		Overhead:    [32]byte{},
		Scalar:      [32]byte{},
		GasLimit:    1234567,
	}
}

func TestParseL1InfoDepositTxData(t *testing.T) {
	randomSeqNr := func(rng *rand.Rand) uint64 {
		return rng.Uint64()
	}
	// Go 1.18 will have native fuzzing for us to use, until then, we cover just the below cases
	cases := []infoTest{
		{"random", testutils.MakeBlockInfo(nil), randomL1Cfg, randomSeqNr},
		{"zero basefee", testutils.MakeBlockInfo(func(l *testutils.MockBlockInfo) {
			l.InfoBaseFee = new(big.Int)
		}), randomL1Cfg, randomSeqNr},
		{"zero time", testutils.MakeBlockInfo(func(l *testutils.MockBlockInfo) {
			l.InfoTime = 0
		}), randomL1Cfg, randomSeqNr},
		{"zero num", testutils.MakeBlockInfo(func(l *testutils.MockBlockInfo) {
			l.InfoNum = 0
		}), randomL1Cfg, randomSeqNr},
		{"zero seq", testutils.MakeBlockInfo(nil), randomL1Cfg, func(rng *rand.Rand) uint64 {
			return 0
		}},
		{"all zero", func(rng *rand.Rand) *testutils.MockBlockInfo {
			return &testutils.MockBlockInfo{InfoBaseFee: new(big.Int)}
		}, randomL1Cfg, func(rng *rand.Rand) uint64 {
			return 0
		}},
	}
	var rollupCfg rollup.Config
	for i, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(int64(1234 + i)))
			info := testCase.mkInfo(rng)
			l1Cfg := testCase.mkL1Cfg(rng, info)
			seqNr := testCase.seqNr(rng)
			depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, l1Cfg, seqNr, info, 0)
			require.NoError(t, err)
			res, err := L1BlockInfoFromBytes(&rollupCfg, info.Time(), depTx.Data)
			require.NoError(t, err, "expected valid deposit info")
			assert.Equal(t, res.Number, info.NumberU64())
			assert.Equal(t, res.Time, info.Time())
			assert.True(t, res.BaseFee.Sign() >= 0)
			assert.Equal(t, res.BaseFee.Bytes(), info.BaseFee().Bytes())
			assert.Equal(t, res.BlockHash, info.Hash())
			assert.Equal(t, res.SequenceNumber, seqNr)
			assert.Equal(t, res.BatcherAddr, l1Cfg.BatcherAddr)
			assert.Equal(t, res.L1FeeOverhead, l1Cfg.Overhead)
			assert.Equal(t, res.L1FeeScalar, l1Cfg.Scalar)
		})
	}
	t.Run("no data", func(t *testing.T) {
		_, err := L1BlockInfoFromBytes(&rollupCfg, 0, nil)
		assert.Error(t, err)
	})
	t.Run("not enough data", func(t *testing.T) {
		_, err := L1BlockInfoFromBytes(&rollupCfg, 0, []byte{1, 2, 3, 4})
		assert.Error(t, err)
	})
	t.Run("too much data", func(t *testing.T) {
		_, err := L1BlockInfoFromBytes(&rollupCfg, 0, make([]byte, 4+32+32+32+32+32+1))
		assert.Error(t, err)
	})
	t.Run("invalid selector", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, 0)
		require.NoError(t, err)
		_, err = crand.Read(depTx.Data[0:4])
		require.NoError(t, err)
		_, err = L1BlockInfoFromBytes(&rollupCfg, info.Time(), depTx.Data)
		require.ErrorContains(t, err, "function signature")
	})
	t.Run("regolith", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{}
		rollupCfg.ActivateAtGenesis(forks.Regolith)
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, 0)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
	})
	t.Run("ecotone", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Ecotone)
		// run 1 block after ecotone transition
		timestamp := rollupCfg.Genesis.L2Time + rollupCfg.BlockTime
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, timestamp)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		require.Equal(t, L1InfoEcotoneLen, len(depTx.Data))
	})
	t.Run("activation-block ecotone", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Delta)
		ecotoneTime := rollupCfg.Genesis.L2Time + rollupCfg.BlockTime // activate ecotone just after genesis
		rollupCfg.EcotoneTime = &ecotoneTime
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, ecotoneTime)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		require.Equal(t, L1InfoBedrockLen, len(depTx.Data))
	})
	t.Run("genesis-block ecotone", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Ecotone)
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, rollupCfg.Genesis.L2Time)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		require.Equal(t, L1InfoEcotoneLen, len(depTx.Data))
	})
	t.Run("isthmus", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Isthmus)
		// run 1 block after isthmus transition
		timestamp := rollupCfg.Genesis.L2Time + rollupCfg.BlockTime
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, timestamp)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		require.Equal(t, L1InfoIsthmusLen, len(depTx.Data))
	})
	t.Run("activation-block isthmus", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Holocene)
		isthmusTime := rollupCfg.Genesis.L2Time + rollupCfg.BlockTime // activate isthmus just after genesis
		rollupCfg.IsthmusTime = &isthmusTime
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, isthmusTime)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		// Isthmus activates, but ecotone L1 info is still used at this upgrade block
		require.Equal(t, L1InfoEcotoneLen, len(depTx.Data))
		require.Equal(t, L1InfoFuncEcotoneBytes4, depTx.Data[:4])
	})
	t.Run("genesis-block isthmus", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Isthmus)
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, rollupCfg.Genesis.L2Time)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		require.Equal(t, L1InfoIsthmusLen, len(depTx.Data))
	})
	t.Run("jovian", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Jovian)
		// run 1 block after Jovian transition
		timestamp := rollupCfg.Genesis.L2Time + rollupCfg.BlockTime
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, timestamp)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		require.Equal(t, L1InfoJovianLen, len(depTx.Data))
		dafgs, err := types.ExtractDAFootprintGasScalar(depTx.Data)
		require.NoError(t, err)
		// randomL1Cfg has scalar 0, which should be translated to the default value.
		require.Equal(t, uint16(DAFootprintGasScalarDefault), dafgs)
	})
	t.Run("activation-block jovian", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Isthmus)
		jovianTime := rollupCfg.Genesis.L2Time + rollupCfg.BlockTime // activate jovian just after genesis
		rollupCfg.InteropTime = &jovianTime
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, jovianTime)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		// Jovian activates, but Isthmus L1 info is still used at this upgrade block
		require.Equal(t, L1InfoIsthmusLen, len(depTx.Data))
		require.Equal(t, L1InfoFuncIsthmusBytes4, depTx.Data[:4])
	})
	t.Run("genesis-block jovian", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Jovian)
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, rollupCfg.Genesis.L2Time)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		require.Equal(t, L1InfoJovianLen, len(depTx.Data))
	})
	t.Run("interop", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Interop)
		// run 1 block after interop transition
		timestamp := rollupCfg.Genesis.L2Time + rollupCfg.BlockTime
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, timestamp)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		require.Equal(t, L1InfoJovianLen, len(depTx.Data), "the length is same in interop")
	})
	t.Run("activation-block interop", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Jovian)
		interopTime := rollupCfg.Genesis.L2Time + rollupCfg.BlockTime // activate interop just after genesis
		rollupCfg.InteropTime = &interopTime
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, interopTime)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		require.Equal(t, L1InfoJovianLen, len(depTx.Data))
		require.Equal(t, L1InfoFuncJovianBytes4, depTx.Data[:4])
	})
	t.Run("genesis-block interop", func(t *testing.T) {
		rng := rand.New(rand.NewSource(1234))
		info := testutils.MakeBlockInfo(nil)(rng)
		rollupCfg := rollup.Config{BlockTime: 2, Genesis: rollup.Genesis{L2Time: 1000}}
		rollupCfg.ActivateAtGenesis(forks.Interop)
		depTx, err := L1InfoDeposit(&rollupCfg, params.MergedTestChainConfig, randomL1Cfg(rng, info), randomSeqNr(rng), info, rollupCfg.Genesis.L2Time)
		require.NoError(t, err)
		require.False(t, depTx.IsSystemTransaction)
		require.Equal(t, depTx.Gas, uint64(RegolithSystemTxGas))
		require.Equal(t, L1InfoJovianLen, len(depTx.Data))
	})
}

// TestStripBPOBlobBaseFee verifies that stripBPOActivations produces the BlobBaseFee
// matching the actual Celo Sepolia L2 block (derived before BPO was known).
// Uses data from Sepolia L1 block 10253939 / Celo Sepolia L2 block 17727223.
func TestStripBPOBlobBaseFee(t *testing.T) {
	// Sepolia L1 block 10253939 header data (post-BPO2 activation).
	excessBlobGas := uint64(226664020)
	blockTime := uint64(1771010016)

	blockInfo := eth.HeaderBlockInfo(&types.Header{
		Time:          blockTime,
		ExcessBlobGas: &excessBlobGas,
	})

	// With BPO-stripped config, the blob base fee should match the value in the
	// corresponding Celo Sepolia L2 block (17727223), which was derived using
	// Prague blob parameters (before BPO was activated on the L2).
	strippedCfg := stripPreJovianBPOActivations(params.SepoliaChainConfig)
	derivedBlobBaseFee := blockInfo.BlobBaseFee(strippedCfg)
	expected, ok := new(big.Int).SetString("45441352348192177559", 10)
	require.True(t, ok)
	require.Equal(t, expected, derivedBlobBaseFee)
}

func TestIsPreJovianCeloChain(t *testing.T) {
	t.Run("celo mainnet returns true at block 0", func(t *testing.T) {
		assert.True(t, isPreJovianCeloChain(&rollup.Config{L2ChainID: big.NewInt(params.CeloMainnetChainID)}, 0))
	})
	t.Run("celo sepolia returns true at block 0", func(t *testing.T) {
		assert.True(t, isPreJovianCeloChain(&rollup.Config{L2ChainID: big.NewInt(params.CeloSepoliaChainID)}, 0))
	})
	t.Run("celo chaos returns true at block 0", func(t *testing.T) {
		assert.True(t, isPreJovianCeloChain(&rollup.Config{L2ChainID: big.NewInt(params.CeloChaosChainID)}, 0))
	})
	t.Run("default chain returns false at block 0", func(t *testing.T) {
		assert.False(t, isPreJovianCeloChain(&rollup.Config{L2ChainID: big.NewInt(999)}, 0))
	})
	t.Run("op mainnet returns false at block 0", func(t *testing.T) {
		assert.False(t, isPreJovianCeloChain(&rollup.Config{L2ChainID: big.NewInt(10)}, 0))
	})
}

func TestStripPreJovianBPOActivations(t *testing.T) {
	osakaTime := uint64(1000)
	bpo1Time := uint64(2000)
	bpo2Time := uint64(3000)
	bpo3Time := uint64(4000)
	bpo4Time := uint64(5000)
	bpo5Time := uint64(6000)
	pragueTime := uint64(500)

	t.Run("strips osaka and bpo times", func(t *testing.T) {
		cfg := &params.ChainConfig{
			OsakaTime:  &osakaTime,
			BPO1Time:   &bpo1Time,
			BPO2Time:   &bpo2Time,
			BPO3Time:   &bpo3Time,
			BPO4Time:   &bpo4Time,
			BPO5Time:   &bpo5Time,
			PragueTime: &pragueTime,
			BlobScheduleConfig: &params.BlobScheduleConfig{
				Osaka:  &params.BlobConfig{Target: 6, Max: 9},
				BPO1:   &params.BlobConfig{Target: 8, Max: 12},
				BPO2:   &params.BlobConfig{Target: 10, Max: 15},
				BPO3:   &params.BlobConfig{Target: 12, Max: 18},
				BPO4:   &params.BlobConfig{Target: 14, Max: 21},
				BPO5:   &params.BlobConfig{Target: 16, Max: 24},
				Prague: &params.BlobConfig{Target: 3, Max: 6},
			},
		}
		stripped := stripPreJovianBPOActivations(cfg)

		// BPO/Osaka times should be nil
		assert.Nil(t, stripped.OsakaTime)
		assert.Nil(t, stripped.BPO1Time)
		assert.Nil(t, stripped.BPO2Time)

		// Other times should be preserved
		assert.NotNil(t, stripped.BPO3Time)
		assert.NotNil(t, stripped.BPO4Time)
		assert.NotNil(t, stripped.BPO5Time)
		assert.NotNil(t, stripped.PragueTime)

		// BlobScheduleConfig BPO/Osaka entries should be nil
		require.NotNil(t, stripped.BlobScheduleConfig)
		assert.Nil(t, stripped.BlobScheduleConfig.Osaka)
		assert.Nil(t, stripped.BlobScheduleConfig.BPO1)
		assert.Nil(t, stripped.BlobScheduleConfig.BPO2)

		// Other BlobScheduleConfig entries should be preserved
		assert.NotNil(t, stripped.BlobScheduleConfig.BPO3)
		assert.NotNil(t, stripped.BlobScheduleConfig.BPO4)
		assert.NotNil(t, stripped.BlobScheduleConfig.BPO5)
		require.NotNil(t, stripped.BlobScheduleConfig.Prague)

		// Original should be unmodified
		require.NotNil(t, cfg.OsakaTime)
		require.NotNil(t, cfg.BlobScheduleConfig)
		require.NotNil(t, cfg.BlobScheduleConfig.Osaka)
		require.NotNil(t, cfg.BPO1Time)
		require.NotNil(t, cfg.BlobScheduleConfig.BPO1)
		require.NotNil(t, cfg.BPO2Time)
		require.NotNil(t, cfg.BlobScheduleConfig.BPO2)
	})

	t.Run("nil blob schedule config is safe", func(t *testing.T) {
		cfg := &params.ChainConfig{
			OsakaTime:          &osakaTime,
			BPO1Time:           &bpo1Time,
			BlobScheduleConfig: nil,
		}

		stripped := stripPreJovianBPOActivations(cfg)
		assert.Nil(t, stripped.OsakaTime)
		assert.Nil(t, stripped.BPO1Time)
		assert.Nil(t, stripped.BlobScheduleConfig)
	})
}
