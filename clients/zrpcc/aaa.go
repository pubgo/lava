package zrpcc

import (
	"context"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

const Name = "zrpcc"

type Client interface {
	Conn() (*nats.Conn, error)
	CallUnary(ctx context.Context, subject string, req, resp proto.Message) error
	Healthy(ctx context.Context) error
	Close() error
}
