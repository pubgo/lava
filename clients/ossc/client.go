package ossc

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pubgo/lava/resource"
	"io"
)

var _ io.Closer = (*wrapper)(nil)

type wrapper struct {
	*oss.Client
}

func (w wrapper) Close() error { return nil }

var _ resource.Resource = (*Client)(nil)

type Client struct {
	v *wrapper
}

func (t *Client) Unwrap() io.Closer               { return t.v }
func (t *Client) UpdateObj(val resource.Resource) { t.v = val.(*Client).v }
func (t *Client) Kind() string                    { return Name }
