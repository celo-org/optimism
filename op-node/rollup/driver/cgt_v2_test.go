package driver

import (
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/sources"
)

// Keep the driver's raw L1 client compatible with the header interface;
// the wrappers installed later do not expose these methods.
var _ derive.CgtV2InputSource = (*sources.L1Client)(nil)
