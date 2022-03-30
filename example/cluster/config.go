package cluster

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/fastrand"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/utils"
)

const (
	nodeNamePrefix = "lava"
)

func generateNodeName() string {
	return fmt.Sprintf("%s-%s", nodeNamePrefix, fastrand.String(8))
}

type Config struct {
	DataDir             string
	SerfLANConfig       serf.Config
	ID                  int
	NodeName            string
	Addr                string
	AdvertiseAddr       string
	AdvertisePort       int
	TCPTimeout          time.Duration
	IndirectChecks      int
	RetransmitMult      int
	SuspicionMult       int
	PushPullInterval    time.Duration
	ProbeInterval       time.Duration
	ProbeTimeout        time.Duration
	GossipInterval      time.Duration
	GossipToTheDeadTime time.Duration
	SecretKey           string   `yaml:"secretKey"`
	Seeds               []string `yaml:"seeds"`
}

func (t *Config) OnNodeEvent(func(e *ClusterEvent, c *Cluster)) {}

func (t *Config) OnDelegate() {
}

func (t *Config) Build() *memberlist.Config {
	xerror.Assert(t.Addr == "", "Config.Addr should not be null")

	t.SecretKey = utils.FirstNotEmpty(
		func() string { return t.SecretKey },
		func() string { return env.Get("secret-key") },
		func() string { return defaultSecretKey },
	)

	var host, port, err = net.SplitHostPort(t.Addr)
	xerror.Panic(err)
	xerror.Assert(port == "", "port should not be null")

	if host == "" {
		host = "0.0.0.0"
	}

	p, err := strconv.Atoi(port)
	xerror.Panic(err, "port parse error")

	cfg := memberlist.DefaultLocalConfig()
	cfg.Name = generateNodeName()
	cfg.BindAddr = host
	cfg.BindPort = p
	cfg.SecretKey = []byte(t.SecretKey)

	if cfg.AdvertiseAddr == "" {
		cfg.AdvertiseAddr = host
	}

	if cfg.AdvertisePort == 0 {
		cfg.AdvertisePort = p
	}

	merge.Struct(cfg, t)
	return cfg
}

func DefaultCfg() *Config {
	return &Config{
		SecretKey:           defaultSecretKey,
		Addr:                ":8080",
		TCPTimeout:          time.Second,
		IndirectChecks:      1,
		RetransmitMult:      2,
		SuspicionMult:       3,
		PushPullInterval:    15 * time.Second,
		ProbeTimeout:        200 * time.Millisecond,
		ProbeInterval:       time.Second,
		GossipInterval:      100 * time.Millisecond,
		GossipToTheDeadTime: 15 * time.Second,
	}
}
