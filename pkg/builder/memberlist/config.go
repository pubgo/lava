package memberlist

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/pubgo/xerror"
)

//https://github.com/asim/memberlist/blob/master/memberlist.go
//https://github.com/serialx/hashring
//https://github.com/octu0/example-memberlist
//https://github.com/pomerium/autocache/blob/master/autocache.go
//https://github.com/magisterquis/meshmembers/blob/master/meshmembers.go

var chanSize = 128

type Cfg struct {
	mc                      *memberlist.Config
	Name                    string
	BindAddr                string
	BindPort                int
	AdvertiseAddr           string
	AdvertisePort           int
	ProtocolVersion         uint8
	TCPTimeout              time.Duration
	IndirectChecks          int
	RetransmitMult          int
	SuspicionMult           int
	SuspicionMaxTimeoutMult int
	PushPullInterval        time.Duration
	ProbeInterval           time.Duration
	ProbeTimeout            time.Duration
	DisableTcpPings         bool
	AwarenessMaxMultiplier  int
	GossipInterval          time.Duration
	GossipNodes             int
	GossipToTheDeadTime     time.Duration
	GossipVerifyIncoming    bool
	GossipVerifyOutgoing    bool
	EnableCompression       bool
	SecretKey               []byte
	DelegateProtocolVersion uint8
	DelegateProtocolMin     uint8
	DelegateProtocolMax     uint8
	DNSConfigPath           string
	HandoffQueueDepth       int
	UDPBufferSize           int
}

func (t Cfg) Build() *memberlist.Memberlist {
	t.BindPort = 0
	t.Name += fmt.Sprintf("_%d", time.Now().Unix())
	t.mc.Events = &memberlist.ChannelEventDelegate{Ch: make(chan memberlist.NodeEvent, chanSize)}
	t.mc.Events = &eventDelegate{}
	t.mc.Delegate = &delegate{}
	t.mc.Logger = log.Default()

	ml, err := memberlist.Create(t.mc)
	xerror.Panic(err)
	return ml
}

func GetDefaultCfg() Cfg {
	return Cfg{
		mc: memberlist.DefaultLANConfig(),
	}
}
