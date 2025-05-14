package p2p

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/p2p/gating"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/netutil"
	ds "github.com/ipfs/go-datastore"
	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/metrics"
	cmgr "github.com/libp2p/go-libp2p/p2p/net/connmgr"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
)

var DefaultBootnodes = []string{
	// OP Labs
	"enode://869d07b5932f17e8490990f75a3f94195e9504ddb6b85f7189e5a9c0a8fff8b00aecf6f3ac450ecba6cdabdb5858788a94bde2b613e0f2d82e9b395355f76d1a@34.65.67.101:30305",
	"enode://2d4e7e9d48f4dd4efe9342706dd1b0024681bd4c3300d021f86fc75eab7865d4e0cbec6fbc883f011cfd6a57423e7e2f6e104baad2b744c3cafaec6bc7dc92c1@34.65.43.171:30305",
	// Base
	"enr:-J24QNz9lbrKbN4iSmmjtnr7SjUMk4zB7f1krHZcTZx-JRKZd0kA2gjufUROD6T3sOWDVDnFJRvqBBo62zuF-hYCohOGAYiOoEyEgmlkgnY0gmlwhAPniryHb3BzdGFja4OFQgCJc2VjcDI1NmsxoQKNVFlCxh_B-716tTs-h1vMzZkSs1FTu_OYTNjgufplG4N0Y3CCJAaDdWRwgiQG",
	"enr:-J24QH-f1wt99sfpHy4c0QJM-NfmsIfmlLAMMcgZCUEgKG_BBYFc6FwYgaMJMQN5dsRBJApIok0jFn-9CS842lGpLmqGAYiOoDRAgmlkgnY0gmlwhLhIgb2Hb3BzdGFja4OFQgCJc2VjcDI1NmsxoQJ9FTIv8B9myn1MWaC_2lJ-sMoeCDkusCsk4BYHjjCq04N0Y3CCJAaDdWRwgiQG",
	"enr:-J24QDXyyxvQYsd0yfsN0cRr1lZ1N11zGTplMNlW4xNEc7LkPXh0NAJ9iSOVdRO95GPYAIc6xmyoCCG6_0JxdL3a0zaGAYiOoAjFgmlkgnY0gmlwhAPckbGHb3BzdGFja4OFQgCJc2VjcDI1NmsxoQJwoS7tzwxqXSyFL7g0JM-KWVbgvjfB8JA__T7yY_cYboN0Y3CCJAaDdWRwgiQG",
	"enr:-J24QHmGyBwUZXIcsGYMaUqGGSl4CFdx9Tozu-vQCn5bHIQbR7On7dZbU61vYvfrJr30t0iahSqhc64J46MnUO2JvQaGAYiOoCKKgmlkgnY0gmlwhAPnCzSHb3BzdGFja4OFQgCJc2VjcDI1NmsxoQINc4fSijfbNIiGhcgvwjsjxVFJHUstK9L1T8OTKUjgloN0Y3CCJAaDdWRwgiQG",
	"enr:-J24QG3ypT4xSu0gjb5PABCmVxZqBjVw9ca7pvsI8jl4KATYAnxBmfkaIuEqy9sKvDHKuNCsy57WwK9wTt2aQgcaDDyGAYiOoGAXgmlkgnY0gmlwhDbGmZaHb3BzdGFja4OFQgCJc2VjcDI1NmsxoQIeAK_--tcLEiu7HvoUlbV52MspE0uCocsx1f_rYvRenIN0Y3CCJAaDdWRwgiQG",
	// Conduit
	"enode://d25ce99435982b04d60c4b41ba256b84b888626db7bee45a9419382300fbe907359ae5ef250346785bff8d3b9d07cd3e017a27e2ee3cfda3bcbb0ba762ac9674@bootnode.conduit.xyz:0?discport=30301",
	"enode://2d4e7e9d48f4dd4efe9342706dd1b0024681bd4c3300d021f86fc75eab7865d4e0cbec6fbc883f011cfd6a57423e7e2f6e104baad2b744c3cafaec6bc7dc92c1@34.65.43.171:0?discport=30305",
	"enode://9d7a3efefe442351217e73b3a593bcb8efffb55b4807699972145324eab5e6b382152f8d24f6301baebbfb5ecd4127bd3faab2842c04cd432bdf50ba092f6645@34.65.109.126:0?discport=30305",
	// Uniswap Labs
	"enode://010800c668896c100e8d64abc388ac5a22a8134a96fb0107c5d0c56d79ba7225c12d9e9e012d3cc0ee2701d7f63dd45f8abf0bbcf6f3c541f91742b1d7a99355@3.134.214.169:9222",
	"enode://b97abcc7011d06299c4bc44742be4a0e631a1a2925a2992adcfe80ed86bec5ff0ddf1b90d015f2dbb5e305560e12c9873b2dad72d84d131ac4be9f2a4c74b763@52.14.30.39:9222",
	"enode://760230a662610620d6d2e4ad846a6dccbceaa4556872dfacf9cdca7c2f5b49e4c66e822ed2e8813debb5fb7391f0519b8d075e565a2a89c79a9e4092e81b3e5b@3.148.100.173:9222",
	// Celo
	"enr:-J64QJipvmFhMq6DVh6RR4HvIiiBtyy1NUg_QlnAAbf18SMqCxCPZtLgUiWED5p0HRVPv69Wth4YPsvdKXSUyh57mWuGAZXRp6HjgmlkgnY0gmlwhCJTtG-Hb3BzdGFja4TsyQIAiXNlY3AyNTZrMaECKPT8t_OMGwEgh_eu8l3LChJXzPHNxMqohYTcJUFhKQaDdGNwgiQGg3VkcIIkBg",
	"enr:-J64QCxBGS49IQbkbwsUuVWt9CkMctMCRe0b-4dqRsLr4QJ1S52urWPUk2uhBU5uerRGpxWTZZW5FtJC-9gSBHN3cSiGAZXRp4rbgmlkgnY0gmlwhCKph0CHb3BzdGFja4TsyQIAiXNlY3AyNTZrMaECqQd8PgMCBpVMXH8izBajLLUBMRKqiYXjV1-t2niEpQiDdGNwgiQGg3VkcIIkBg",
	"enr:-J64QLG71bmmljNbLFx3qim6zXohKA3jbK_4C4d1cwixI-7VMoBIlnM6kWZVvvdWcbjTQ6QXB1LAO39eZWC4Heztj1-GAZXRpzUGgmlkgnY0gmlwhCKpySSHb3BzdGFja4TsyQIAiXNlY3AyNTZrMaEDApsAenpWrLqo6lDsYs2ieUhL84Q_rhZG9pBWb3hKylCDdGNwgiQGg3VkcIIkBg",
	"enr:-J64QKFU-u1x1gt3WmNP88EDUMQ316ymbzdGy83QjkBDqVSsJBn6-nipuqYQDeHYoLBLVJUMdyAiwxVbbDm14qQSf5qGAZXRppmIgmlkgnY0gmlwhCJTfzOHb3BzdGFja4TsyQIAiXNlY3AyNTZrMaEC88lrc6V3LF77SNWjO_GT5YCA2Ca6fwPp1b3vIMBjSk-DdGNwgiQGg3VkcIIkBg",
	"enr:-J64QIXTVl0Opbdn20TSrkzpIZ4xQ54bERRlTmSeZ05dFLdlSbuRY7yn5tJeTPzsSldTw5V5E0qjEQcsfr20vMjTUDyGAZXRpiWygmlkgnY0gmlwhCPjrx6Hb3BzdGFja4TsyQIAiXNlY3AyNTZrMaED2qWtZdFrywlnz0eNnyBUS_G23mF2NORS3_e5RyefQfSDdGNwgiQGg3VkcIIkBg",
	"enr:-J64QFAsbeR4xRSyVyQOk7bILUCoMjI2EnbZvo4UAK3842HMYw41-UZXdnQJH8lwvzWn7qsY3Vu73NuxzxWKn4XB5wiGAZXRpYPAgmlkgnY0gmlwhCJSxmKHb3BzdGFja4TsyQIAiXNlY3AyNTZrMaEDx51ZbXcmg4flmWldI-lBwUwiB0UFLqZkKnHvffMaE4eDdGNwgiQGg3VkcIIkBg",
	"enr:-J64QFQSrL3mfG-i64T-5DgVE5V9dGKC5A0JrEvD6CRpZvuLK3feg4bPaqFWfqXyNN_6IgY2z1Jkr4Mf2Zx-GdWlWquGAZXQkMdSgmlkgnY0gmlwhCImtd-Hb3BzdGFja4TsyQIAiXNlY3AyNTZrMaEDQVEzYHXdCOtsdb_WOFXopL1v0Pka5KgbFJMPJnHhau6DdGNwgiQGg3VkcIIkBg",
	"enr:-J64QAp3g1m-5uX-_mBXWyo6ZQqAlnRcAt11Xwy0-ZzqaSrDSlg4adyOz6v9flzLgxYkVvXI50nJGs8GjLgT5bwDLtyGAZXQrD69gmlkgnY0gmlwhCJMJgaHb3BzdGFja4TsyQIAiXNlY3AyNTZrMaECq5mdt1EmXHFLFxNE3hly7XQ0gWLeRloERPVuULjP0EiDdGNwgiQGg3VkcIIkBg",
	"enr:-J64QFCZs1ePThNEsRxIIzbfDxYfap1nEyuPPpSUeeWOoPFWOp0zSEPwLEtXhG1eH-ipsB5CgtaVzcXOyT9hKeAeVVaGAZXQkaZ3gmlkgnY0gmlwhCO7ajaHb3BzdGFja4TsyQIAiXNlY3AyNTZrMaEDnYbZL7OKQpMwVG_hrvziZOH1XF1AJJtjFT5990QAX6ODdGNwgiQGg3VkcIIkBg",
	"enr:-J64QJ9LY8m9AjNgujuVT0juX8T6PHKojZEIqd-7_vhBasfiT2xUUJoUfWga_xVJGFECFcN6hPKB4TjihmYFxHXelwOGAZXQkclrgmlkgnY0gmlwhCJMELeHb3BzdGFja4TsyQIAiXNlY3AyNTZrMaEDyCwx8h3Vu7jcNWhv9npDUzgrQBfJ7HZgo4PMtbjjsEyDdGNwgiQGg3VkcIIkBg",
	"enr:-J64QGJFPZzLj2GLFgB4JhTde7rXChMNFERNbzrwYYTG7CY2SCSggFrU3VXczzWBvOoJWdbOMOzPuCI2klknGjruUxeGAZXQkf1LgmlkgnY0gmlwhGjHJzuHb3BzdGFja4TsyQIAiXNlY3AyNTZrMaEDO61fV62N5lQfAuNtgGuH5-nKbVM8lW6JpWswVK6F1giDdGNwgiQGg3VkcIIkBg",
	"enr:-J64QEXleDl25w0qEG__wmDgwnzB0F5zapu00D_jM4qkCbA3WIcLC8rXPm8dcrKdZNBuNXJOtNE6c2_ZDkuQMvIuhjCGAZXQwDjFgmlkgnY0gmlwhCKMdU-Hb3BzdGFja4TsyQIAiXNlY3AyNTZrMaECHezzuLmg0LgzLRUhjzvwzrlgaw7-GPNSxR7_wUu_H0-DdGNwgiQGg3VkcIIkBg",
}

type HostMetrics interface {
	gating.UnbanMetrics
	gating.ConnectionGaterMetrics
}

// SetupP2P provides a host and discovery service for usage in the rollup node.
type SetupP2P interface {
	Check() error
	// Looks like this was started to prevent partially inited p2p.
	Disabled() bool
	// Host creates a libp2p host service. Returns nil, nil if p2p is disabled.
	Host(log log.Logger, reporter metrics.Reporter, metrics HostMetrics) (host.Host, error)
	// Discovery creates a disc-v5 service. Returns nil, nil, nil if discovery is disabled.
	Discovery(log log.Logger, rollupCfg *rollup.Config, tcpPort uint16) (*enode.LocalNode, *discover.UDPv5, error)
	TargetPeers() uint
	BanPeers() bool
	BanThreshold() float64
	BanDuration() time.Duration
	GossipSetupConfigurables
	ReqRespSyncEnabled() bool
}

// ScoringParams defines the various types of peer scoring parameters.
type ScoringParams struct {
	PeerScoring        pubsub.PeerScoreParams
	ApplicationScoring ApplicationScoreParams
}

// Config sets up a p2p host and discv5 service from configuration.
// This implements SetupP2P.
type Config struct {
	Priv *crypto.Secp256k1PrivateKey

	DisableP2P  bool
	NoDiscovery bool

	ScoringParams *ScoringParams

	// Whether to ban peers based on their [PeerScoring] score. Should be negative.
	BanningEnabled bool
	// Minimum score before peers are disconnected and banned
	BanningThreshold float64
	BanningDuration  time.Duration

	ListenIP      net.IP
	ListenTCPPort uint16

	// Port to bind discv5 to
	ListenUDPPort uint16

	AdvertiseIP      net.IP
	AdvertiseTCPPort uint16
	AdvertiseUDPPort uint16
	Bootnodes        []*enode.Node
	DiscoveryDB      *enode.DB
	NetRestrict      *netutil.Netlist

	StaticPeers []core.Multiaddr

	HostMux             []libp2p.Option
	HostSecurity        []libp2p.Option
	NoTransportSecurity bool

	PeersLo    uint
	PeersHi    uint
	PeersGrace time.Duration

	MeshD     int // topic stable mesh target count
	MeshDLo   int // topic stable mesh low watermark
	MeshDHi   int // topic stable mesh high watermark
	MeshDLazy int // gossip target

	// FloodPublish publishes messages from ourselves to peers outside of the gossip topic mesh but supporting the same topic.
	FloodPublish bool

	// If true a NAT manager will host a NAT port mapping that is updated with PMP and UPNP by libp2p/go-nat
	NAT bool

	UserAgent string

	TimeoutNegotiation time.Duration
	TimeoutAccept      time.Duration
	TimeoutDial        time.Duration

	// Underlying store that hosts connection-gater and peerstore data.
	Store ds.Batching

	EnableReqRespSync   bool
	SyncOnlyReqToStatic bool

	EnablePingService bool
}

func DefaultConnManager(conf *Config) (connmgr.ConnManager, error) {
	return cmgr.NewConnManager(
		int(conf.PeersLo),
		int(conf.PeersHi),
		cmgr.WithGracePeriod(conf.PeersGrace),
		cmgr.WithSilencePeriod(time.Minute),
		cmgr.WithEmergencyTrim(true))
}

func (conf *Config) TargetPeers() uint {
	return conf.PeersLo
}

func (conf *Config) Disabled() bool {
	return conf.DisableP2P
}

func (conf *Config) PeerScoringParams() *ScoringParams {
	if conf.ScoringParams == nil {
		return nil
	}
	return conf.ScoringParams
}

func (conf *Config) BanPeers() bool {
	return conf.BanningEnabled
}

func (conf *Config) BanThreshold() float64 {
	return conf.BanningThreshold
}

func (conf *Config) BanDuration() time.Duration {
	return conf.BanningDuration
}

func (conf *Config) ReqRespSyncEnabled() bool {
	return conf.EnableReqRespSync
}

const maxMeshParam = 1000

func (conf *Config) Check() error {
	if conf.DisableP2P {
		if len(conf.StaticPeers) > 0 {
			return errors.New("both --p2p.static and --p2p.disable are specified")
		}
		return nil
	}
	if conf.Store == nil {
		return errors.New("p2p requires a persistent or in-memory peerstore, but found none")
	}
	if !conf.NoDiscovery {
		if conf.DiscoveryDB == nil {
			return errors.New("discovery requires a persistent or in-memory discv5 db, but found none")
		}
	}
	if conf.PeersLo == 0 || conf.PeersHi == 0 || conf.PeersLo > conf.PeersHi {
		return fmt.Errorf("peers lo/hi tides are invalid: %d, %d", conf.PeersLo, conf.PeersHi)
	}
	if conf.MeshD <= 0 || conf.MeshD > maxMeshParam {
		return fmt.Errorf("mesh D param must not be 0 or exceed %d, but got %d", maxMeshParam, conf.MeshD)
	}
	if conf.MeshDLo <= 0 || conf.MeshDLo > maxMeshParam {
		return fmt.Errorf("mesh Dlo param must not be 0 or exceed %d, but got %d", maxMeshParam, conf.MeshDLo)
	}
	if conf.MeshDHi <= 0 || conf.MeshDHi > maxMeshParam {
		return fmt.Errorf("mesh Dhi param must not be 0 or exceed %d, but got %d", maxMeshParam, conf.MeshDHi)
	}
	if conf.MeshDLazy <= 0 || conf.MeshDLazy > maxMeshParam {
		return fmt.Errorf("mesh Dlazy param must not be 0 or exceed %d, but got %d", maxMeshParam, conf.MeshDLazy)
	}
	return nil
}
