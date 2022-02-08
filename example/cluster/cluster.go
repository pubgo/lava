package cluster

import (
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"time"

	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/netutil"
)

var logs = logging.Component("cluster")

type Cluster struct {
	cfg              *memberlist.Config
	memberList       *memberlist.Memberlist
	eventDelegate    chan memberlist.NodeEvent
	metadataDelegate *NodeMetadataDelegate
	seedNodes        []string
	broadcasts       *memberlist.TransmitLimitedQueue
}

func NewCluster(cfg *Config) (*Cluster, error) {
	clusterLogger := zap.L().Named("cluster")

	var eventDelegate = make(chan memberlist.NodeEvent)
	nodeMetadataDelegate := NewNodeMetadataDelegate(nodeMetadata, clusterLogger)

	quic.ApplicationError{}
	quic.DialAddr()

	config := cfg.Build()
	config.Name = generateNodeName()
	config.Events = &memberlist.ChannelEventDelegate{Ch: eventDelegate}
	config.Delegate = nodeMetadataDelegate
	config.Transport = newNetTransport(clusterLogger, &netutil.Cfg{Addr: config.BindAddr, Port: config.BindPort})
	config.Logger = zap.NewStdLog(clusterLogger)

	member, err := memberlist.Create(config)
	if err != nil {
		clusterLogger.Error(err.Error(), zap.Any("config", config))
		return nil, err
	}

	member.SendBestEffort()

	member.SendReliable()

	// ping seed
	member.Ping()
	// 加入集群
	member.Join()

	// 广播本节点信息
	// members.LocalNode().Meta, err = nodeMetadata.Bytes()
	// if err != nil {
	// 	nodeLogger.Error("Failed to set node metadata", zap.Error(err))
	// }
	// members.UpdateNode(10 * time.Second)
	member.UpdateNode()

	return &Cluster{
		cfg:              config,
		memberList:       member,
		eventDelegate:    eventDelegate,
		metadataDelegate: nodeMetadataDelegate,
		broadcasts: &memberlist.TransmitLimitedQueue{
			NumNodes: func() int {
				return member.NumMembers()
			},
			RetransmitMult: 3,
		},
	}, nil
}

func (c *Cluster) NodeEvent() <-chan memberlist.NodeEvent { return c.eventDelegate }

func (c *Cluster) Join(seeds []string) (int, error) {
	c.broadcasts.QueueBroadcast()
	return c.memberList.Join(seeds)
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
