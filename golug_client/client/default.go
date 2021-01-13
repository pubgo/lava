package client

import (
	"github.com/pubgo/golug/golug_balancer/p2c"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var timeout = 3 * time.Second
var clients sync.Map
var defaultDialOpts = []grpc.DialOption{
	grpc.WithInsecure(),
	grpc.WithBlock(),
	grpc.WithBalancerName(p2c.Name),
	grpc.WithKeepaliveParams(clientParameters),
}

// DefaultMaxRecvMsgSize maximum message that client can receive
// (4 MB).
var DefaultMaxRecvMsgSize = 1024 * 1024 * 4

// DefaultMaxSendMsgSize maximum message that client can send
// (4 MB).
var DefaultMaxSendMsgSize = 1024 * 1024 * 4

var clientParameters = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             2 * time.Second,  // wait 2 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

// max grpc conns for every grpc server
const maxConnsPerService = 30

// init grpc conns for every grpc server. set it equal to grpc server pod replicas
var InitServiceConnSize = 3

// interval time to auto update grpc conn pool
var PoolUpdateInterval = 60 * time.Second

// max reuse times for single grpc conn
const maxConcurrentStreams = 100
