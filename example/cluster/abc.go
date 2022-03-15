package cluster

import (
	"io"
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
	ExpiryTick               = time.Second * 1 // needs to be smaller than registry.RegisterTTL
	MaxPacketSize            = 512
)

type CommitLog interface {
	Delete() error
	NewReader(offset int64, maxBytes int32) (io.Reader, error)
	Truncate(int64) error
	NewestOffset() int64
	OldestOffset() int64
	Append([]byte) (int64, error)
}
