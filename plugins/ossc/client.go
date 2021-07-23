package ossc

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/atomic"
)

type Client struct {
	atomic.Value
}

func (t *Client) Bucket(name string) (*oss.Bucket, error) {
	return t.Get().Bucket(name)
}

func (t *Client) Get() *oss.Client {
	return t.Load().(*oss.Client)
}
