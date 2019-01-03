package cluster

import (
	"github.com/clockworksoul/smudge"
	"github.com/jtejido/hermes/config"
	"net"
)

type (
	Cluster struct {
		me     net.IP
		config *config.Config
	}

	statusListener struct {
		update updateFunc
	}

	updateFunc func(peers []net.IP)
)

func newCluster(config *config.Config) (*Cluster, error) {

	me := net.ParseIP("127.0.0.1")

	c := &Cluster{
		me:     me,
		config: config,
	}

	logger.Printf("host %d : cluster running on port %d with heartbeat every %d ms", config.Http.Host, config.Peers.PeerListenPort, config.Peers.Heartbeat)
	smudge.SetListenIP(me)
	smudge.SetListenPort(config.Peers.PeerListenPort)
	smudge.SetHeartbeatMillis(config.Peers.Heartbeat)
	smudge.SetLogThreshold(smudge.LogInfo)
	smudge.SetMulticastEnabled(config.Peers.MulticastEnabled)
	smudge.SetMulticastAddress(config.Peers.MulticastAddress)
	smudge.SetMulticastPort(config.Peers.MulticastPort)
	smudge.SetClusterName(config.Peers.MulticastClusterName)
	smudge.SetMulticastAnnounceIntervalSeconds(config.Peers.MulticastAnnounceInterval)

	return c, nil

}

func (c *Cluster) listenForUpdates(update updateFunc) {
	smudge.AddStatusListener(&statusListener{update})

	if err := c.setInitialNodes(); err != nil {
		logger.Printf("host %d : error setting initial nodes %v", c.config.Peers.Listen, err)

		return
	}

	smudge.Begin()
}

func (c *Cluster) setInitialNodes() error {

	port := uint16(smudge.GetListenPort())

	// They are anyone else other than you.
	for _, n := range c.config.Peers.Nodes {
		ip := net.ParseIP(n)
		if !ip.Equal(c.me) {
			node, err := smudge.CreateNodeByIP(ip, port)

			if err != nil {
				return err
			}

			smudge.AddNode(node)
		}
	}

	// For local, it's just 127.0.0.1, comment the lines above and uncomment below
	// see comments in cachecluster.go
	// node, _ := smudge.CreateNodeByIP(c.me, port)
	// smudge.AddNode(node)

	return nil
}

func (s *statusListener) OnChange(node *smudge.Node, status smudge.NodeStatus) {
	nodes := smudge.HealthyNodes()
	peers := make([]net.IP, len(nodes))
	for i, node := range nodes {
		peers[i] = node.IP()
	}
	s.update(peers)
}
