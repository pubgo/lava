package stdout

import (
	"context"

	"github.com/pubgo/lug/abc/connector"

	"github.com/pubgo/x/q"
)

var _ connector.Connector = (*Connector)(nil)

type Connector struct{}

func (c *Connector) Read(ctx context.Context, cb func(interface{})) error {
	return nil
}

func (c *Connector) Write(ctx context.Context, data interface{}) error {
	q.Q(data)
	return nil
}

func (c *Connector) Close() error {
	return nil
}
