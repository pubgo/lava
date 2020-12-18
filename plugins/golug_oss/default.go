package golug_oss

import (
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/xerror"
)

var clientM sync.Map

func GetClient(names ...string) *oss.Bucket {
	var name = golug_consts.Default
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}
	val, ok := clientM.Load(name)
	if !ok {
		xerror.Next().Panic(xerror.Fmt("%s not found", name))
	}

	return val.(*oss.Bucket)
}

func initClient(name string, cfg ClientCfg) {
	client, err := oss.New(
		golug_config.Template(cfg.Endpoint),
		golug_config.Template(cfg.AccessKeyID),
		golug_config.Template(cfg.AccessKeySecret),
	)
	xerror.Panic(err)
	kk := xerror.PanicErr(client.Bucket(cfg.Bucket)).(*oss.Bucket)
	clientM.Store(name, kk)
}
