package celo

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
)

func TestMain(m *testing.M) {
	presets.DoMain(m,
		presets.WithMinimal(),
		stack.MakeCommon(sysgo.WithDeployerOptions(
			sysgo.WithCelo(),
		)),
	)
}
