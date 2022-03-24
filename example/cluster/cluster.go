package cluster

import (
	"crypto/rand"
	"fmt"
	"github.com/pubgo/lava/core/cmux"
	"math/big"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging"
)

var logs = logging.Component("cluster")

type Cluster struct {
	config       Config
	shutdownCh   chan struct{}
	raftNotifyCh <-chan bool
	// reconcileCh is used to pass events from the serf handler to the raft leader to update its state.
	reconcileCh chan serf.Member
	serf        *serf.Serf
	eventChLAN  chan serf.Event
	//eventChLAN:       make(chan serf.Event, 256),
	//		brokerLookup:     NewBrokerLookup(),
	//		replicaLookup:    NewReplicaLookup(),
	//		reconcileCh:      make(chan serf.Member, 32),

	Log              *logging.Logger
	nodeId           string
	AddrList         []string
	cfg              *memberlist.Config
	memberList       *memberlist.Memberlist
	eventDelegate    chan memberlist.NodeEvent
	metadataDelegate *NodeMetadataDelegate
	seedNodes        []string
	broadcast        *memberlist.TransmitLimitedQueue
	msgChan          chan []byte
}

func NewCluster(cfg *Config) (*Cluster, error) {
	clusterLogger := zap.L().Named("cluster")

	var queue = &memberlist.TransmitLimitedQueue{RetransmitMult: 3}

	var eventDelegate = make(chan memberlist.NodeEvent)
	var msgChan = make(chan []byte)

	config := cfg.Build()
	config.Name = generateNodeName()
	config.Events = &memberlist.ChannelEventDelegate{Ch: eventDelegate}
	config.Delegate = &delegate{queue: queue, msg: msgChan}
	config.Transport = newNetTransport(clusterLogger, &cmux.Mux{Addr: config.BindAddr, Port: config.BindPort})
	config.Logger = zap.NewStdLog(clusterLogger)

	member, err := memberlist.Create(config)
	if err != nil {
		clusterLogger.Error(err.Error(), zap.Any("config", config))
		return nil, err
	}

	queue.NumNodes = func() int {
		return member.NumMembers()
	}

	member.LocalNode()

	// 广播本节点信息
	// members.LocalNode().Meta, err = nodeMetadata.Bytes()
	// if err != nil {
	// 	nodeLogger.Error("Failed to set node metadata", zap.Error(err))
	// }
	// members.UpdateNode(10 * time.Second)

	var c = &Cluster{
		cfg:           config,
		memberList:    member,
		eventDelegate: eventDelegate,
		broadcast:     queue,
	}
	go c.lanEventHandler()

	return c, nil
}

// Join is used to have the broker join the gossip ring.
// The given address should be another broker listening on the Serf address.
func (c *Cluster) JoinLAN(addrs ...string) error {
	if _, err := c.serf.Join(addrs, true); err != nil {
		return err
	}
	return nil
}

func (c *Cluster) NodeEvent() <-chan memberlist.NodeEvent { return c.eventDelegate }
func (c *Cluster) MsgChan() <-chan []byte                 { return c.msgChan }

func (c *Cluster) Join(seeds []string) (int, error) {
	return c.memberList.Join(seeds)
}

// Broadcast 广播数据
func (c *Cluster) Broadcast(b memberlist.Broadcast) {
	c.broadcast.QueueBroadcast(b)
}

func (c *Cluster) BroadcastData(msg []byte) {
	c.broadcast.QueueBroadcast(&broadcast{msg: msg})
}

func (c *Cluster) Leave(timeout time.Duration) error {
	return c.memberList.Leave(timeout)
}

func (c *Cluster) LocalNodeName() string {
	return c.memberList.LocalNode().Name
}

func (c *Cluster) LocalNodeMetadata() (*NodeMetadata, error) {
	return NewNodeMetadataWithBytes(c.memberList.LocalNode().Meta)
}

func (c *Cluster) NodeMetadata(nodeName string) (*NodeMetadata, error) {
	nodes := c.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return NewNodeMetadataWithBytes(node.Meta)
		}
	}

	return nil, fmt.Errorf("ErrNodeDoesNotFound")
}

func (c *Cluster) NodeAddress(nodeName string) (string, error) {
	nodes := c.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return node.Addr.String(), nil
		}
	}

	return "", fmt.Errorf("ErrNodeDoesNotFound")
}

func (c *Cluster) NodePort(nodeName string) (uint16, error) {
	nodes := c.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return node.Port, nil
		}
	}

	return 0, fmt.Errorf("ErrNodeDoesNotFound")
}

func (c *Cluster) NodeState(nodeName string) (NodeState, error) {
	nodes := c.memberList.Members()
	for _, node := range nodes {
		if node.Name == nodeName {
			return makeNodeState(node.State), nil
		}
	}

	return NodeStateUnknown, fmt.Errorf("ErrNodeDoesNotFound")
}

func (c *Cluster) Nodes() []string {
	members := make([]string, 0)
	for _, member := range c.memberList.Members() {
		members = append(members, member.Name)
	}
	return members
}

func (c *Cluster) Start() error {
	return nil
}

func (c *Cluster) Stop() error {
	c.cfg.Transport.Shutdown()
	return nil
}

// mRandomNodes is used to select up to m random nodes. It is possible
// that less than m nodes are returned.
func (c *Cluster) mRandomNodes(m int, nodes []string) []string {
	n := len(nodes)
	mNodes := make([]string, 0, m)

OUTER:
	// Probe up to 3*n times, with large n this is not necessary
	// since k << n, but with small n we want search to be
	// exhaustive
	for i := 0; i < 3*n && len(mNodes) < m; i++ {
		// Get random node
		idx := randomOffset(n)
		node := nodes[idx]

		if node == c.memberList.LocalNode().Name {
			continue
		}

		// Check if we have this node already
		for j := 0; j < len(mNodes); j++ {
			if node == mNodes[j] {
				continue OUTER
			}
		}

		// Append the node
		mNodes = append(mNodes, node)
	}

	return mNodes
}

// Returns a random offset between 0 and n
func randomOffset(n int) int {
	if n == 0 {
		return 0
	}

	val, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		logs.S().Errorf("Failed to get a random offset: %v", err)
		return 0
	}

	return int(val.Int64())
}
