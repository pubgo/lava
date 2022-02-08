package cluster

import (
	"time"
)

const (
	defaultSecretKey         = "stEL9:6CUP{mOyQ&!b@4cDI_;'RK~.7#"
	DefaultPushPullInterval  = 60 * time.Second
	DefaultGossipInterval    = 200 * time.Millisecond
	DefaultTcpTimeout        = 10 * time.Second
	DefaultProbeTimeout      = 500 * time.Millisecond
	DefaultProbeInterval     = 1 * time.Second
	DefaultReconnectInterval = 10 * time.Second
	DefaultReconnectTimeout  = 6 * time.Hour
	DefaultRefreshInterval   = 15 * time.Second
	MaxGossipPacketSize      = 1400
)
