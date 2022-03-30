package cluster

import (
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/pubgo/lava/core/logging"
	"go.uber.org/zap"
	"log"
	"net"
	"path/filepath"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/serf/serf"

	"github.com/pubgo/lava/runtime"
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
	serfWANSnapshot   = "serf/remote.snapshot"
	raftState         = "raft/"
	snapshotsRetained = 2

	// raftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	raftLogCacheSize = 512

	// raftRemoveGracePeriod is how long we wait to allow a RemovePeer
	// to replicate to gracefully leave the cluster.
	raftRemoveGracePeriod = 5 * time.Second

	// serfEventChSize is the size of the buffered channel to get Serf
	// events. If this is exhausted we will block Serf and Memberlist.
	serfEventChSize = 2048

	// reconcileChSize is the size of the buffered channel reconcile updates
	// from Serf with the Catalog. If this is exhausted we will drop updates,
	// and wait for a periodic reconcile.
	reconcileChSize = 256

	LeaderTransferMinVersion = "1.6.0"
)

const (
	aclPolicyReplicationRoutineName       = "ACL policy replication"
	aclRoleReplicationRoutineName         = "ACL role replication"
	aclTokenReplicationRoutineName        = "ACL token replication"
	aclTokenReapingRoutineName            = "acl token reaping"
	aclUpgradeRoutineName                 = "legacy ACL token upgrade"
	caRootPruningRoutineName              = "CA root pruning"
	caRootMetricRoutineName               = "CA root expiration metric"
	caSigningMetricRoutineName            = "CA signing expiration metric"
	configReplicationRoutineName          = "config entry replication"
	federationStateReplicationRoutineName = "federation state replication"
	federationStateAntiEntropyRoutineName = "federation state anti-entropy"
	federationStatePruningRoutineName     = "federation state pruning"
	intentionMigrationRoutineName         = "intention config entry migration"
	secondaryCARootWatchRoutineName       = "secondary CA roots watch"
	intermediateCertRenewWatchRoutineName = "intermediate cert renew watch"
	backgroundCAInitializationRoutineName = "CA initialization"
	virtualIPCheckRoutineName             = "virtual IP version check"
)

var (
	ErrWANFederationDisabled = fmt.Errorf("WAN Federation is disabled")
)

const (
	PoolKindPartition = "partition"
	PoolKindSegment   = "segment"
)

func init() {
	spew.Config.Indent = ""
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

const (
	// StatusReap is used to update the status of a node if we
	// are handling a EventMemberReap
	StatusReap = serf.MemberStatus(-1)
)

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

// nodeJoin is used to handle join events on the serf cluster
func (a *Cluster) nodeJoin(me serf.MemberEvent) {
	for _, m := range me.Members {
		ok, parts := isServer(m)
		if !ok {
			a.logger.WithField("member", m.Name).Warn("non-server in gossip pool")
			continue
		}
		a.logger.WithField("server", parts.Name).Info("adding server")

		// Check if this server is known
		found := false
		a.peerLock.Lock()
		existing := a.peers[parts.Region]
		for idx, e := range existing {
			if e.Name == parts.Name {
				existing[idx] = parts
				found = true
				break
			}
		}

		// Add ot the list if not known
		if !found {
			a.peers[parts.Region] = append(existing, parts)
		}

		// Check if a local peer
		if parts.Region == a.config.Region {
			a.localPeers[raft.ServerAddress(parts.Addr.String())] = parts
		}
		a.peerLock.Unlock()

		// If we still expecting to bootstrap, may need to handle this
		if a.config.BootstrapExpect != 0 {
			a.maybeBootstrap()
		}
	}
}

func (c *Cluster) setupSerf(config *serf.Config, ch chan serf.Event, path string) (*serf.Serf, error) {
	config.Init()
	config.NodeName = c.config.NodeName
	config.Tags["role"] = "lava"
	config.Tags["port"] = "8080"
	config.Tags["role"] = "consul"
	config.Tags["dc"] = s.config.Datacenter
	config.Tags["segment"] = opts.Segment
	config.Tags["id"] = string(s.config.NodeID)
	config.Tags["vsn"] = fmt.Sprintf("%d", s.config.ProtocolVersion)
	config.Tags["vsn_min"] = fmt.Sprintf("%d", ProtocolVersionMin)
	config.Tags["vsn_max"] = fmt.Sprintf("%d", ProtocolVersionMax)
	config.Tags["raft_vsn"] = fmt.Sprintf("%d", s.config.RaftConfig.ProtocolVersion)
	config.Tags["build"] = s.config.Build
	config.Tags["id"] = fmt.Sprintf("%d", c.config.ID)
	addr := opts.Listener.Addr().(*net.TCPAddr)
	conf.Tags["port"] = fmt.Sprintf("%d", addr.Port)
	if s.config.Bootstrap {
		conf.Tags["bootstrap"] = "1"
	}
	config.Logger = zap.NewStdLog(logging.Component("serf").L())
	config.MemberlistConfig.Logger = zap.NewStdLog(logging.Component("memberlist").L())
	config.Tags["serf_lan_addr"] = fmt.Sprintf("%s:%d", c.config.SerfLANConfig.MemberlistConfig.BindAddr, c.config.SerfLANConfig.MemberlistConfig.BindPort)
	config.EventCh = ch
	config.EnableNameConflictResolution = false
	if runtime.IsProd() || runtime.IsStag() {
		config.SnapshotPath = filepath.Join(c.config.DataDir, path)
	}
	return serf.Create(config)
}

func (s *Cluster) lanEventHandler() {
	for {
		select {
		case e := <-s.eventChLAN:
			switch e.EventType() {
			case serf.EventMemberJoin:
				s.lanNodeJoin(e.(serf.MemberEvent))
				s.localMemberEvent(e.(serf.MemberEvent))

			case serf.EventMemberLeave, serf.EventMemberFailed, serf.EventMemberReap:
				s.lanNodeFailed(e.(serf.MemberEvent))
				s.localMemberEvent(e.(serf.MemberEvent))

			case serf.EventUser:
				s.localEvent(e.(serf.UserEvent))
			case serf.EventMemberUpdate:
				s.lanNodeUpdate(e.(serf.MemberEvent))
				s.localMemberEvent(e.(serf.MemberEvent))
			case serf.EventQuery: // Ignore
			default:
				s.logger.Warn("Unhandled LAN Serf Event", "event", e)
			}

		case <-s.shutdownCh:
			return
		}
	}
}

func (b *Cluster) JoinLAN(addrs ...string) protocol.Error {
	if _, err := b.serf.Join(addrs, true); err != nil {
		return protocol.ErrUnknown.WithErr(err)
	}
	return protocol.ErrNone
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
			case serf.EventMemberUpdate:
				//	TODO update data
			}
		case <-c.shutdownCh:
			return
		}
	}
}

// lanNodeJoin is used to handle join events on the LAN pool.
func (c *Cluster) lanNodeJoin(me serf.MemberEvent) {
	for _, m := range me.Members {
		meta, ok := IsLavaNode(m)
		if !ok {
			continue
		}
		c.Log.Sugar().Infof("lava/%d: adding LAN server: %d", c.config.ID, meta.ID)
	}
}

func (c *Cluster) lanNodeFailed(me serf.MemberEvent) {
	for _, m := range me.Members {
		meta, ok := IsLavaNode(m)
		if !ok {
			continue
		}

		c.Log.Sugar().Infof("broker/%d: removing LAN server: %s", c.config.ID, m.Name)
	}
}

func (c *Cluster) localMemberEvent(me serf.MemberEvent) {
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

func (c *Cluster) LANMembers() []serf.Member {
	return c.serf.Members()
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
	meta, ok := IsLavaNode(member)
	if !ok {
		return nil
	}

	if meta.ID == c.config.ID {
		log.Debug.Printf("leader/%d: deregistering self should be done by follower", c.config.ID)
		return nil
	}

	if err := c.removeServer(member, meta); err != nil {
		return err
	}
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

	return nil
}

func (c *Cluster) handleFailedMember(m serf.Member) error {
	meta, ok := IsLavaNode(m)
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
	if m.Tags["role"] != "lava" {
		return nil, false
	}

	idStr := m.Tags["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, false
	}

	return &Node{
		ID:   id,
		Name: m.Tags["name"],
	}, true
}
