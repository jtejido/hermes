package cluster

import (
	"fmt"
	"github.com/jtejido/hermes/config"
	"github.com/jtejido/hermes/hermes"
	"net"
	"sort"
	"strings"
)

type (
	CacheCluster struct {
		*hermes.HTTPPool
		me     net.IP
		peers  []string
		config *config.Config
	}
)

// Peer status broadcasting thru Smudge. This is not peer discovery (nodes have to be explicitly listed on the config file)
func New(config *config.Config) (*CacheCluster, error) {

	cluster, err := newCluster(config)
	if err != nil {
		logger.Printf("error creating cluster %v", err)
		return nil, err
	}

	self := fmt.Sprintf("http://%s:%d", cluster.me, config.Peers.Listen)
	logger.Printf("initializing peers on %s", self)

	pool := hermes.NewHTTPPoolOpts(self, nil)
	cc := &CacheCluster{
		HTTPPool: pool,
		me:       cluster.me,
		config:   config,
	}

	logger.Printf("host %d : initializing peer discovery", config.Http.Host)

	go cluster.listenForUpdates(cc.updatePeers)

	return cc, nil
}

func (c *CacheCluster) updatePeers(peerAddresses []net.IP) {
	peers := make([]string, len(peerAddresses))
	for i, addr := range peerAddresses {
		peers[i] = fmt.Sprintf("http://%s:%d", addr, c.config.Peers.Listen)
	}

	// Working Locally, i.e., different ports instead of different IPs, you may comment the lines above and uncomment the line below. We don't have the setting to either work locally (and have a set of ports instead of IPs), or on LAN.
	// see comments in cluster.go.
	// peers := []string{fmt.Sprintf("http://%s:%d", c.me.String(), 9080), fmt.Sprintf("http://%s:%d", c.me.String(), 9081)}

	sort.Slice(peers, func(i, j int) bool { return peers[i] < peers[j] })

	if strings.Join(c.peers, ", ") == strings.Join(peers, ", ") {
		return
	}

	logger.Printf("host %d: %s set peers %s", c.config.Http.Host, c.me, strings.Join(peers, ", "))
	c.Set(peers...)
	c.peers = peers
}

func (c *CacheCluster) ListenOn() string {
	return fmt.Sprintf("0.0.0.0:%d", c.config.Peers.Listen)
}
