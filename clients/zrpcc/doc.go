// Package zrpcc provides a Lava-style zrpc client with lazy NATS connection,
// default middlewares (serviceinfo/metric/accesslog/recovery), and lifecycle APIs.
//
// Access logs are written by middleware_accesslog using fields such as
// request_id, operation (NATS subject), service, latency, and error details.
package zrpcc

import (
	"context"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/pkg/zrpc"
)

const Name = "zrpcc"

type Client interface {
	Conn() (*nats.Conn, error)
	CallUnary(ctx context.Context, subject string, req, resp proto.Message) error
	OpenStream(ctx context.Context, subject string) (*zrpc.ClientStream, error)
	Healthy(ctx context.Context) error
	Close() error
}
