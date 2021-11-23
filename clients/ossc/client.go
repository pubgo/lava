package ossc

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pubgo/lava/resource"
)

var _ resource.Resource = (*Client)(nil)

type Client struct {
	*oss.Client
}

func (t *Client) Close() error                 { return nil }
func (t *Client) UpdateResObj(val interface{}) { t.Client = val.(*Client).Client }
func (t *Client) Kind() string                 { return Name }
