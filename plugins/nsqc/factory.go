package nsqc

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/logger"
	"github.com/pubgo/lug/pkg/typex"

	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"

	"runtime"
)

var clients typex.SMap

func Get(names ...string) *nsqClient {
	var name = consts.GetDefault(names...)
	val, ok := clients.Load(name)
	if !ok {
		return nil
	}

	return val.(*nsqClient)
}

func Update(name string, cfg Cfg) error {
	return try.Try(func() {
		xerror.Assert(name == "", "[name] should not be null")

		// 创建新的客户端
		client, err := cfg.Build()
		xerror.Panic(err)

		// 获取老的客户端
		oldClient, ok := clients.Load(name)
		if !ok || oldClient == nil {
			logs.Debug("create client", logger.Name(name))

			// 老客户端不存在就直接保存
			clients.Set(name, client)
			return
		}

		// 当old etcd client没有被使用的时候, 那么就关闭
		runtime.SetFinalizer(oldClient, func(cc *nsqClient) {
			logs.Info("old client gc", logger.Name(name), logger.UIntPrt(cc))
			//logs.Error("old client close error", logger.Name(name), logger.Err(err))
		})

		logs.Info("update client", logger.Name(name))
		// 老的客户端更新
		clients.Set(name, client)
	})
}
