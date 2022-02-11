package ossc

import (
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/pubgo/lava/resource"
)

var _ io.Closer = (*wrapper)(nil)

type wrapper struct {
	*oss.Client
}

func (w wrapper) Close() error { return nil }

var _ resource.Resource = (*Client)(nil)

type Client struct {
	resource.Resource
}

func (t *Client) Kind() string     { return Name }
func (t *Client) Get() *oss.Client { return t.GetObj().(*wrapper).Client }
func (t *Client) Load() (*oss.Client, resource.Release) {
	var obj, r = t.Resource.LoadObj()
	return obj.(*wrapper).Client, r
}
