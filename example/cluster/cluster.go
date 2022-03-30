package cluster

import (
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"github.com/pubgo/lava/core/logging"
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
