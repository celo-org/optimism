package environment

import (
	espressoClient "github.com/EspressoSystems/espresso-network/sdks/go/client"
	"github.com/ethereum-optimism/optimism/espresso"
)

// mockEspressoDevNode is an EspressoDevNode backed by the in-memory mock Espresso
// client (espresso.MockEspressoClient) instead of a dockerized espresso-dev-node.
// The mock is owned by the e2esys.System (System.EspressoClient) and stopped when
// the system is closed, so Stop here is a no-op.
type mockEspressoDevNode struct {
	client *espresso.MockEspressoClient
}

var _ EspressoDevNode = (*mockEspressoDevNode)(nil)

func (m *mockEspressoDevNode) Client() espressoClient.EspressoClient {
	return m.client
}

// SequencerPort is unused by the in-memory mock; there is no listening port.
func (m *mockEspressoDevNode) SequencerPort() string { return "" }

// BuilderPort is unused by the in-memory mock; there is no listening port.
func (m *mockEspressoDevNode) BuilderPort() string { return "" }

// EspressoUrls is unused by the in-memory mock; there are no URLs.
func (m *mockEspressoDevNode) EspressoUrls() []string { return nil }

// Stop is a no-op; the mock client's lifecycle is owned by the e2esys.System.
func (m *mockEspressoDevNode) Stop() error { return nil }
