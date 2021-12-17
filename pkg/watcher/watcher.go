package watcher

import (
	"bytes"
	"context"
	"strings"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/pkg/ctxutil"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
)

var defaultWatcher Watcher = &nullWatcher{}
var logs = logz.Component(Name)

func Init(projects ...string) (err error) {
	defer xerror.RespErr(&err)

	_ = config.Decode(Name, &cfg)

	defaultWatcher = xerror.PanicErr(cfg.Build()).(Watcher)

	// 获取所有需要watch的项目
	if !strutil.Contains(projects, runenv.Project) {
		projects = append(projects, runenv.Project)
	}

	// 项目prefix
	for i := range projects {
		var name = projects[i]

		// get远程配置, 获取项目下所有配置
		xerror.Panic(defaultWatcher.GetCallback(
			ctxutil.Timeout(), name,
			func(resp *Response) {
				resp.Event = types.EventType_UPDATE
				onWatch(name, resp)
			},
		))

		// watch远程配置
		defaultWatcher.WatchCallback(
			context.Background(), name,
			func(resp *Response) { onWatch(name, resp) },
		)
	}
	return
}

func onWatch(name string, resp *Response) {
	var project = zap.String("watch-project", name)

	defer xerror.Resp(func(err xerror.XErr) {
		logs.WithErr(err, project, zap.Any("resp", resp)).Error("watch callback error")
	})

	// value为空就skip
	if cfg.SkipNull && len(bytes.TrimSpace(resp.Value)) == 0 {
		return
	}

	var key = KeyToDot(resp.Key)

	logs.Infow(
		"watch callback",
		project,
		zap.Any("key", key),
		zap.Any("event", resp.Event.String()),
		zap.Any("version", resp.Version),
		zap.Any("value", string(resp.Value)),
	)

	// 把数据更新到全局配置中
	// value必须是kv类型
	var dt = make(map[string]interface{})
	xerror.PanicF(types.Decode(resp.Value, &dt), "value必须是kv类型, key=>%s, value=>%s", resp.Key, resp.Value)

	resp.OnPut(func() {
		if name == runenv.Project {
			// 本项目配置, 去掉本项目前缀
			config.GetCfg().Set(trimProject(key), dt)
		} else {
			// 非本项目配置, 项目前缀要带上名字
			config.GetCfg().Set(key, dt)
		}
	})

	resp.OnDelete(func() {
		if name == runenv.Project {
			// 本项目配置, 去掉本项目前缀
			config.GetCfg().Set(trimProject(key), nil)
		} else {
			// 非本项目配置, 项目前缀要带上名字
			config.GetCfg().Set(key, nil)
		}
	})

	// 以name为前缀的所有的callbacks
	for k, v := range callbacks {
		// 检查是否是以key为前缀, `.`是连接符和分隔符
		if !strings.HasPrefix(key+".", k+".") {
			return
		}

		// 去掉watch前缀
		var watchKey = strings.Trim(strings.TrimPrefix(key, k), ".")

		// 执行watch callback
		for i := range v {
			logs.Logs("watch callback handle", func() error { return v[i](watchKey, resp) },
				project,
				zap.String("watch-key", k),
				zap.Any("watch-resp", resp),
				zap.Any("watch-stack", stack.Func(v[i])),
			)
		}
	}
}
