package driver

import (
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/sources"
)

// Keep the driver's raw concrete clients compatible with the proof interface;
// the wrappers installed later do not expose these methods.
var (
	_ derive.CgtV2ProofClient = (*sources.L1Client)(nil)
	_ derive.CgtV2ProofClient = (*sources.EngineClient)(nil)
)
