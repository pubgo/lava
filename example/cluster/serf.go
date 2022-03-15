package cluster

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

const (
	// StatusReap is used to update the status of a node if we
	// are handling a EventMemberReap
	StatusReap = serf.MemberStatus(-1)
)

var (
	brokerVerboseLogs bool

	ErrTopicExists            = errors.New("topic exists already")
	ErrInvalidArgument        = errors.New("no logger set")
	OffsetsTopicName          = "__consumer_offsets"
	OffsetsTopicNumPartitions = 50
)

const (
	serfLANSnapshot   = "serf/local.snapshot"
	raftState         = "raft/"
	raftLogCacheSize  = 512
	snapshotsRetained = 2
)

func init() {
	spew.Config.Indent = ""

	e := os.Getenv("JOCKODEBUG")
	if strings.Contains(e, "broker=1") {
		brokerVerboseLogs = true
	}
}

// DefaultConfig returns a Consul-flavored Serf default configuration,
// suitable as a basis for a LAN, WAN, segment, or area.
func DefaultConfig() *serf.Config {
	base := serf.DefaultConfig()

	// This effectively disables the annoying queue depth warnings.
	base.QueueDepthWarning = 1000000

	// This enables dynamic sizing of the message queue depth based on the
	// cluster size.
	base.MinQueueDepth = 4096

	// This gives leaves some time to propagate through the cluster before
	// we shut down. The value was chosen to be reasonably short, but to
	// allow a leave to get to over 99.99% of the cluster with 100k nodes
	// (using https://www.serf.io/docs/internals/simulator.html).
	base.LeavePropagateDelay = 3 * time.Second

	return base
}

func GetTags(serf *serf.Serf) map[string]string {
	tags := make(map[string]string)
	for tag, value := range serf.LocalMember().Tags {
		tags[tag] = value
	}

	return tags
}

func UpdateTag(serf *serf.Serf, tag, value string) {
	tags := GetTags(serf)
	tags[tag] = value

	serf.SetTags(tags)
}

func (c *Cluster) setupSerf(config *serf.Config, ch chan serf.Event, path string) (*serf.Serf, error) {
	config.Init()
	config.NodeName = c.config.NodeName
	config.Tags["role"] = "lava"
	config.Tags["id"] = fmt.Sprintf("%d", c.config.ID)
	config.Logger = log.NewStdLogger(log.New(log.DebugLevel, fmt.Sprintf("serf/%d: ", c.config.ID)))
	config.MemberlistConfig.Logger = log.NewStdLogger(log.New(log.DebugLevel, fmt.Sprintf("memberlist/%d: ", c.config.ID)))
	if c.config.Bootstrap {
		config.Tags["bootstrap"] = "1"
	}
	if c.config.BootstrapExpect != 0 {
		config.Tags["expect"] = fmt.Sprintf("%d", c.config.BootstrapExpect)
	}
	if c.config.NonVoter {
		config.Tags["non_voter"] = "1"
	}
	config.Tags["raft_addr"] = c.config.RaftAddr
	config.Tags["serf_lan_addr"] = fmt.Sprintf("%s:%d", c.config.SerfLANConfig.MemberlistConfig.BindAddr, c.config.SerfLANConfig.MemberlistConfig.BindPort)
	config.Tags["broker_addr"] = c.config.Addr
	config.EventCh = ch
	config.EnableNameConflictResolution = false
	if !c.config.DevMode {
		config.SnapshotPath = filepath.Join(c.config.DataDir, path)
	}
	return serf.Create(config)
}

func (c *Cluster) lanEventHandler() {
	for {
		select {
		case e := <-c.eventChLAN:
			switch e.EventType() {
			case serf.EventMemberJoin:
				c.lanNodeJoin(e.(serf.MemberEvent))
				c.localMemberEvent(e.(serf.MemberEvent))
			case serf.EventMemberReap:
				c.localMemberEvent(e.(serf.MemberEvent))
			case serf.EventMemberLeave, serf.EventMemberFailed:
				c.lanNodeFailed(e.(serf.MemberEvent))
				c.localMemberEvent(e.(serf.MemberEvent))
			}
		case <-c.shutdownCh:
			return
		}
	}
}

// lanNodeJoin is used to handle join events on the LAN pool.
func (c *Cluster) lanNodeJoin(me serf.MemberEvent) {
	for _, m := range me.Members {
		meta, ok := metadata.IsBroker(m)
		if !ok {
			continue
		}
		log.Info.Printf("broker/%d: adding LAN server: %s", c.config.ID, meta.ID)
		// update server lookup
		c.brokerLookup.AddBroker(meta)
		if c.config.BootstrapExpect != 0 {
			c.maybeBootstrap()
		}
	}
}

func (c *Cluster) lanNodeFailed(me serf.MemberEvent) {
	for _, m := range me.Members {
		meta, ok := metadata.IsBroker(m)
		if !ok {
			continue
		}
		log.Info.Printf("broker/%d: removing LAN server: %s", c.config.ID, m.Name)
		c.brokerLookup.RemoveBroker(meta)
	}
}

func (c *Cluster) localMemberEvent(me serf.MemberEvent) {
	if !c.isLeader() {
		return
	}

	isReap := me.EventType() == serf.EventMemberReap

	for _, m := range me.Members {
		if isReap {
			m.Status = StatusReap
		}
		select {
		case c.reconcileCh <- m:
		default:
		}
	}
}

func (c *Cluster) maybeBootstrap() {
	var index uint64
	var err error
	if c.config.DevMode {
		index, err = c.raftInmem.LastIndex()
	} else {
		index, err = c.raftStore.LastIndex()
	}
	if err != nil {
		log.Error.Printf("broker/%d: read last raft index error: %s", c.config.ID, err)
		return
	}
	if index != 0 {
		log.Info.Printf("broker/%d: raft data found, disabling bootstrap mode: index: %d, path: %s", c.config.ID, index, filepath.Join(c.config.DataDir, raftState))
		c.config.BootstrapExpect = 0
		return
	}

	members := c.LANMembers()
	brokers := make([]metadata.Broker, 0, len(members))
	for _, member := range members {
		meta, ok := metadata.IsBroker(member)
		if !ok {
			continue
		}
		if meta.Expect != 0 && meta.Expect != c.config.BootstrapExpect {
			log.Error.Printf("broker/%d: members expects conflicting node count: %s", c.config.ID, member.Name)
			return
		}
		if meta.Bootstrap {
			log.Error.Printf("broker/%d; member %s has bootstrap mode. expect disabled", c.config.ID, member.Name)
			return
		}
		brokers = append(brokers, *meta)
	}

	if len(brokers) < c.config.BootstrapExpect {
		log.Debug.Printf("broker/%d: maybe bootstrap: need more brokers: got: %d: expect: %d", c.config.ID, len(brokers), c.config.BootstrapExpect)
		return
	}

	var configuration raft.Configuration
	addrs := make([]string, 0, len(brokers))
	for _, meta := range brokers {
		addr := meta.RaftAddr
		addrs = append(addrs, addr)
		peer := raft.Server{
			ID:      raft.ServerID(meta.ID.String()),
			Address: raft.ServerAddress(addr),
		}
		configuration.Servers = append(configuration.Servers, peer)
	}

	log.Info.Printf("broker/%d: found expected number of peers, attempting bootstrap: addrs: %v", c.config.ID, addrs)
	future := c.raft.BootstrapCluster(configuration)
	if err := future.Error(); err != nil {
		log.Error.Printf("broker/%d: bootstrap cluster error: %s", c.config.ID, err)
	}
	c.config.BootstrapExpect = 0
}

func (c *Cluster) reconcileReaped(known map[int32]struct{}) error {
	state := c.fsm.State()
	_, nodes, err := state.GetNodes()
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if _, ok := known[node.Node]; ok {
			continue
		}
		member := serf.Member{
			Tags: map[string]string{
				"id":   fmt.Sprintf("%d", node.Node),
				"role": "jocko",
			},
		}
		if err := c.handleReapMember(member); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cluster) reconcileMember(m serf.Member) error {
	var err error
	switch m.Status {
	case serf.StatusAlive:
		err = c.handleAliveMember(m)
	case serf.StatusFailed:
		err = c.handleFailedMember(m)
	case StatusReap:
		err = c.handleReapMember(m)
	case serf.StatusLeft:
		err = c.handleLeftMember(m)
	}
	if err != nil {
		log.Error.Printf("leader/%d: reconcile member: %s: error: %s", m.Name, c.config.ID, err)
	}
	return nil
}

func (c *Cluster) handleAliveMember(m serf.Member) error {
	meta, ok := metadata.IsBroker(m)
	if ok {
		if err := c.joinCluster(m, meta); err != nil {
			return err
		}
	}
	state := c.fsm.State()
	_, node, err := state.GetNode(meta.ID.Int32())
	if err != nil {
		return err
	}
	if node != nil {
		// TODO: should still register?
		return nil
	}

	log.Info.Printf("leader/%d: member joined, marking health alive: %s", c.config.ID, m.Name)
	req := structs.RegisterNodeRequest{
		Node: structs.Node{
			Node:    meta.ID.Int32(),
			Address: meta.BrokerAddr,
			Meta: map[string]string{
				"raft_addr":     meta.RaftAddr,
				"serf_lan_addr": meta.SerfLANAddr,
				"name":          meta.Name,
			},
			Check: &structs.HealthCheck{
				Node:    meta.ID.String(),
				CheckID: structs.SerfCheckID,
				Name:    structs.SerfCheckName,
				Status:  structs.HealthPassing,
				Output:  structs.SerfCheckAliveOutput,
			},
		},
	}
	_, err = c.raftApply(structs.RegisterNodeRequestType, &req)
	return err
}

func (c *Cluster) raftApply(t structs.MessageType, msg interface{}) (interface{}, error) {
	buf, err := structs.Encode(t, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %v", err)
	}
	future := c.raft.Apply(buf, 30*time.Second)
	if err := future.Error(); err != nil {
		return nil, err
	}
	return future.Response(), nil
}

func (c *Cluster) handleLeftMember(m serf.Member) error {
	return c.handleDeregisterMember("left", m)
}

func (c *Cluster) handleReapMember(member serf.Member) error {
	return c.handleDeregisterMember("reaped", member)
}

// handleDeregisterMember is used to deregister a mmeber for a given reason.
func (c *Cluster) handleDeregisterMember(reason string, member serf.Member) error {
	meta, ok := metadata.IsBroker(member)
	if !ok {
		return nil
	}

	if meta.ID.Int32() == c.config.ID {
		log.Debug.Printf("leader/%d: deregistering self should be done by follower", c.config.ID)
		return nil
	}

	if err := c.removeServer(member, meta); err != nil {
		return err
	}

	state := c.fsm.State()
	_, node, err := state.GetNode(meta.ID.Int32())
	if err != nil {
		return err
	}
	if node == nil {
		return nil
	}

	log.Info.Printf("leader/%d: member is deregistering: reason: %s; node: %s", c.config.ID, reason, meta.ID)
	req := structs.DeregisterNodeRequest{
		Node: structs.Node{Node: meta.ID.Int32()},
	}
	_, err = c.raftApply(structs.DeregisterNodeRequestType, &req)
	return err
}

func (c *Cluster) joinCluster(m serf.Member, parts *metadata.Broker) error {
	if parts.Bootstrap {
		members := c.LANMembers()
		for _, member := range members {
			p, ok := metadata.IsBroker(member)
			if ok && member.Name != m.Name && p.Bootstrap {
				log.Error.Printf("leader/%d: multiple nodes in bootstrap mode. there can only be one.", c.config.ID)
				return nil
			}
		}
	}

	configFuture := c.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Error.Printf("leader/%d: get raft configuration error: %s", c.config.ID, err)
		return err
	}

	// Processing ourselves could result in trying to remove ourselves to
	// fix up our address, which would make us step down. This is only
	// safe to attempt if there are multiple servers available.
	if m.Name == c.config.NodeName {
		if l := len(configFuture.Configuration().Servers); l < 3 {
			log.Debug.Printf("leader/%d: skipping self join since cluster is too small: servers: %d", c.config.ID, l)
			return nil
		}
	}

	for _, server := range configFuture.Configuration().Servers {
		if server.Address == raft.ServerAddress(parts.RaftAddr) || server.ID == raft.ServerID(parts.ID.String()) {
			if server.Address == raft.ServerAddress(parts.RaftAddr) && server.ID == raft.ServerID(parts.ID.String()) {
				// no-op if this is being called on an existing server
				return nil
			}
			future := c.raft.RemoveServer(server.ID, 0, 0)
			if server.Address == raft.ServerAddress(parts.RaftAddr) {
				if err := future.Error(); err != nil {
					return fmt.Errorf("error removing server with duplicate address %q: %s", server.Address, err)
				}
				log.Info.Printf("removed server with duplicated address: %s", server.Address)
			} else {
				if err := future.Error(); err != nil {
					return fmt.Errorf("removing server with duplicate ID %q: %s", server.ID, err)
				}
				log.Info.Printf("removed server with duplicate ID: %s", server.ID)
			}
		}
	}

	if parts.NonVoter {
		addFuture := c.raft.AddNonvoter(raft.ServerID(parts.ID.String()), raft.ServerAddress(parts.RaftAddr), 0, 0)
		if err := addFuture.Error(); err != nil {
			log.Error.Printf("leader/%d: add raft peer error: %s", c.config.ID, err)
			return err
		}
	} else {
		log.Debug.Printf("leader/%d: join cluster: add voter: %s", c.config.ID, parts.ID)
		addFuture := c.raft.AddVoter(raft.ServerID(parts.ID.String()), raft.ServerAddress(parts.RaftAddr), 0, 0)
		if err := addFuture.Error(); err != nil {
			log.Error.Printf("leader/%d: add raft peer error: %s", c.config.ID, err)
			return err
		}
	}

	return nil
}

func (c *Cluster) handleFailedMember(m serf.Member) error {
	meta, ok := metadata.IsBroker(m)
	if !ok {
		return nil
	}

	req := structs.RegisterNodeRequest{
		Node: structs.Node{
			Node: meta.ID.Int32(),
			Check: &structs.HealthCheck{
				Node:    m.Tags["raft_addr"],
				CheckID: structs.SerfCheckID,
				Name:    structs.SerfCheckName,
				Status:  structs.HealthCritical,
				Output:  structs.SerfCheckFailedOutput,
			},
		},
	}
	if _, err := c.raftApply(structs.RegisterNodeRequestType, &req); err != nil {
		return err
	}

	// TODO should put all the following some where else. maybe onBrokerChange or handleBrokerChange

	state := c.fsm.State()

	_, partitions, err := state.GetPartitions()
	if err != nil {
		panic(err)
	}

	// need to reassign partitions
	_, partitions, err = state.PartitionsByLeader(meta.ID.Int32())
	if err != nil {
		return err
	}
	_, nodes, err := state.GetNodes()
	if err != nil {
		return err
	}

	// TODO: add an index for this. have same code in broker.go:handleMetadata(...)
	var passing []*structs.Node
	for _, n := range nodes {
		if n.Check.Status == structs.HealthPassing && n.ID != meta.ID.Int32() {
			passing = append(passing, n)
		}
	}

	// reassign consumer group coordinators
	_, groups, err := state.GetGroupsByCoordinator(meta.ID.Int32())
	if err != nil {
		return err
	}
	for _, group := range groups {
		i := rand.Intn(len(passing))
		node := passing[i]
		group.Coordinator = node.Node
		req := structs.RegisterGroupRequest{
			Group: *group,
		}
		if _, err = c.raftApply(structs.RegisterGroupRequestType, req); err != nil {
			return err
		}
	}

	leaderAndISRReq := &protocol.LeaderAndISRRequest{
		ControllerID:    c.config.ID,
		PartitionStates: make([]*protocol.PartitionState, 0, len(partitions)),
		// TODO: LiveLeaders, ControllerEpoch
	}
	for _, p := range partitions {
		i := rand.Intn(len(passing))
		// TODO: check that old leader won't be in this list, will have been deregistered removed from fsm
		node := passing[i]

		// TODO: need to check replication factor

		var ar []int32
		for _, r := range p.AR {
			if r != meta.ID.Int32() {
				ar = append(ar, r)
			}
		}
		var isr []int32
		for _, r := range p.ISR {
			if r != meta.ID.Int32() {
				isr = append(isr, r)
			}
		}

		// TODO: need to update epochs

		req := structs.RegisterPartitionRequest{
			Partition: structs.Partition{
				Topic:     p.Topic,
				ID:        p.Partition,
				Partition: p.Partition,
				Leader:    node.Node,
				AR:        ar,
				ISR:       isr,
			},
		}
		if _, err = c.raftApply(structs.RegisterPartitionRequestType, req); err != nil {
			return err
		}
		// TODO: need to send on leader and isr changes now i think
		leaderAndISRReq.PartitionStates = append(leaderAndISRReq.PartitionStates, &protocol.PartitionState{
			Topic:     p.Topic,
			Partition: p.Partition,
			// TODO: ControllerEpoch, LeaderEpoch, ZKVersion - lol
			Leader:   p.Leader,
			ISR:      p.ISR,
			Replicas: p.AR,
		})
	}

	// TODO: optimize this to send requests to only nodes affected
	for _, n := range passing {
		broker := c.brokerLookup.BrokerByID(raft.ServerID(fmt.Sprintf("%d", n.Node)))
		if broker == nil {
			// TODO: this probably shouldn't happen -- likely a root issue to fix
			log.Error.Printf("trying to assign partitions to unknown broker: %s", n)
			continue
		}
		conn, err := defaultDialer.Dial("tcp", broker.BrokerAddr)
		if err != nil {
			return err
		}
		_, err = conn.LeaderAndISR(leaderAndISRReq)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) removeServer(m serf.Member, meta *metadata.Broker) error {
	configFuture := c.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Error.Printf("leader/%d: get raft configuration error: %s", c.config.ID, err)
		return err
	}
	for _, server := range configFuture.Configuration().Servers {
		if server.ID != raft.ServerID(meta.ID.String()) {
			continue
		}
		log.Info.Printf("leader/%d: removing server by id: %s", c.config.ID, server.ID)
		future := c.raft.RemoveServer(raft.ServerID(meta.ID.String()), 0, 0)
		if err := future.Error(); err != nil {
			log.Error.Printf("leader/%d: remove server error: %s", c.config.ID, err)
			return err
		}
	}
	return nil
}

func IsLavaNode(m serf.Member) (*Node, bool) {
	if m.Tags["role"] != "jocko" {
		return nil, false
	}

	expect := 0
	expectStr, ok := m.Tags["expect"]
	var err error
	if ok {
		expect, err = strconv.Atoi(expectStr)
		if err != nil {
			return nil, false
		}
	}

	_, bootstrap := m.Tags["bootstrap"]
	_, nonVoter := m.Tags["non_voter"]

	idStr := m.Tags["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, false
	}

	return &Node{
		ID:          NodeID(id),
		Name:        m.Tags["name"],
		Bootstrap:   bootstrap,
		Expect:      expect,
		NonVoter:    nonVoter,
		Status:      m.Status,
		RaftAddr:    m.Tags["raft_addr"],
		SerfLANAddr: m.Tags["serf_lan_addr"],
		BrokerAddr:  m.Tags["broker_addr"],
	}, true
}
