package resty

import (
	"context"
	"time"

	"github.com/pubgo/funk/v2/result"
)

const (
	defaultRetryCount    = 3
	defaultRetryInterval = 10 * time.Millisecond
	defaultHTTPTimeout   = 2 * time.Second
	defaultContentType   = "application/json"
	defaultTimeout       = 10 * time.Second
	Name                 = "resty"
)

type IClient interface {
	Do(ctx context.Context, req *Request) result.Result[*Response]
}
