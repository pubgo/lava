package ossc

import (
	"github.com/pubgo/lava/resource"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var _ io.Closer = (*wrapper)(nil)

type wrapper struct {
	*oss.Client
}

func (w wrapper) Close() error { return nil }

type Client struct {
	resource.Resource
}

func (t *Client) Load() *oss.Client {
	var obj = t.Resource.GetRes()
	return obj.(*wrapper).Client
}
