package ossc

import (
	"github.com/pubgo/lava/pkg/utils"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/pubgo/lava/resource"
)

func Get(names ...string) *Client {
	val := resource.Get(Name, utils.GetDefault(names...))
	if val == nil {
		return nil
	}

	return val.(*Client)
}

var _ io.Closer = (*wrapper)(nil)

type wrapper struct {
	*oss.Client
}

func (w wrapper) Close() error { return nil }

type Client struct {
	resource.Resource
}

func (t *Client) Load() (*oss.Client, resource.Release) {
	var obj, r = t.Resource.LoadObj()
	return obj.(*wrapper).Client, r
}
