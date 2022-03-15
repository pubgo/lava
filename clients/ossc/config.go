package ossc

import (
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pubgo/xerror"
)

var Name = "oss"

type Cfg struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	Bucket          string
}

func (c Cfg) Build() io.Closer {
	client, err := oss.New(c.Endpoint, c.AccessKeyID, c.AccessKeySecret)
	xerror.Panic(err)
	return &wrapper{client}
}

func DefaultCfg() Cfg {
	return Cfg{}
}
