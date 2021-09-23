package connector

import (
	"context"
)

const Name = "connector"

type Connector interface {
	Close() error
	Read(ctx context.Context, cb func(interface{})) error
	Write(ctx context.Context, data interface{}) error
}
