package watcher_builder

import (
	"bytes"
	"context"
	"strings"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/core/watcher"
	"github.com/pubgo/lava/core/watcher/watcher_driver/noop"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/pkg/ctxutil"
	"github.com/pubgo/lava/runtime"
)

var defaultWatcher watcher.Watcher = &noop.NullWatcher{}
var logs = logging.Component(watcher.Name)
var cfg = watcher.DefaultCfg()

// Init 初始化watcher
func Init(conf config.Config) {
	defer xerror.RespExit()

	if c := conf.GetMap(watcher.Name); c != nil {
		xerror.Panic(c.Decode(&cfg))
	}

	defaultWatcher = xerror.PanicErr(build(cfg.DriverCfg)).(watcher.Watcher)
	// 依赖注入
	inject.Inject(defaultWatcher)
	defaultWatcher.Init()

	// 获取所有需要watch的项目
	if !strutil.Contains(cfg.Projects, runtime.Name()) {
		cfg.Projects = append(cfg.Projects, runtime.Name())
	}

	// 项目prefix
	for i := range cfg.Projects {
		var project = cfg.Projects[i]

		// get远程配置, 启动时, 获取项目下所有配置
		xerror.Panic(defaultWatcher.GetCallback(ctxutil.Timeout(), project, func(resp *watcher.Response) { onInit(project, resp) }))

		// watch远程配置
		defaultWatcher.WatchCallback(context.Background(), project, func(resp *watcher.Response) { onWatch(project, resp) })
	}
	return
}

// onInit 初始化, 获取远程配置
func onInit(name string, resp *watcher.Response) {
	// 过滤空值
	if cfg.SkipNull && len(bytes.TrimSpace(resp.Value)) == 0 {
		return
	}

	var key = watcher.KeyToDot(resp.Key)

	logutil.OkOrPanic(logs.L(), "watch get callback", func() error {
		// 把远程配置更新到内存配置中
		// value必须是json类型
		var dt = make(map[string]interface{})
		xerror.PanicF(resp.Decode(&dt), "value必须是json类型, key=>%s, value=>%s", resp.Key, resp.Value)

		// 内存配置, 去掉本项目前缀, 其他项目保留项目前缀
		config.GetCfg().Set(trimProject(key), dt)
		return nil
	},
		zap.String("watch-project", name),
		zap.Any("key", key),
		zap.Any("event", resp.Event.String()),
		zap.Any("version", resp.Version),
		zap.Any("value", string(resp.Value)),
		zap.Any("resp", resp),
	)

}

func onWatch(name string, resp *watcher.Response) {
	var project = zap.String("watch-project", name)

	defer xerror.Resp(func(err xerror.XErr) {
		logs.WithErr(err).With(
			zap.Any("project", project),
			zap.Any("resp", resp),
		).Error("watch callback error")
	})

	// value为空就skip
	if cfg.SkipNull && len(bytes.TrimSpace(resp.Value)) == 0 {
		return
	}

	var key = watcher.KeyToDot(resp.Key)

	logs.L().With(
		zap.Any("project", project),
		zap.Any("key", key),
		zap.Any("event", resp.Event.String()),
		zap.Any("version", resp.Version),
		zap.Any("value", string(resp.Value)),
	).Info("watch callback")

	// 把数据更新到全局配置中
	// value必须是json类型
	var dt = make(map[string]interface{})
	xerror.PanicF(resp.Decode(&dt), "value必须是json类型, key=>%s, value=>%s", resp.Key, resp.Value)

	resp.OnPut(func() {
		// 本项目配置, 去掉本项目前缀
		config.GetCfg().Set(trimProject(key), dt)
	})

	resp.OnDelete(func() {
		// 本项目配置, 去掉本项目前缀
		config.GetCfg().Set(trimProject(key), nil)
	})

	// 以name为前缀的所有的callbacks
	for k, v := range watcher.GetWatchers() {
		// 检查是否是以key为前缀, `.`是连接符和分隔符
		if !strings.HasPrefix(key+".", k+".") {
			return
		}

		// 去掉watch前缀
		var watchKey = strings.Trim(strings.TrimPrefix(key, k), ".")

		// 执行watch callback
		for i := range v {
			var h = v[i]
			logutil.LogOrErr(logs.L(), "watch callback handle", func() error { return h(watchKey, resp) },
				project,
				zap.String("watch-key", k),
				zap.Any("watch-resp", resp),
				zap.Any("watch-stack", stack.Func(h)))
		}
	}
}

func build(data config.CfgMap) (_ watcher.Watcher, err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "watcher driver is null")
	xerror.Assert(watcher.GetFactory(driver) == nil, "watcher driver [%s] not found", driver)

	fc := watcher.GetFactory(driver)
	return fc(data)
}
