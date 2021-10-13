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
	"github.com/pubgo/lava/pkg/ctxutil"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/vars"
)

var defaultWatcher Watcher = &nullWatcher{}
var callbacks typex.Map

func Init() (err error) {
	defer xerror.RespErr(&err)

	if !config.Decode(Name, &cfg) {
		return
	}

	defaultWatcher = xerror.PanicErr(cfg.Build()).(Watcher)

	// 获取所有watch的项目
	projects := cfg.Projects
	if !strutil.Contains(projects, runenv.Project) {
		projects = append(projects, runenv.Project)
	}

	// 项目prefix
	for i := range projects {
		var name = projects[i]

		// get远程配置, 获取项目下所有配置
		xerror.Panic(defaultWatcher.GetCallback(
			ctxutil.Timeout(), name,
			func(resp *Response) { onWatch(name, resp) }),
		)

		// watch远程配置
		defaultWatcher.WatchCallback(
			context.Background(), name,
			func(resp *Response) { onWatch(name, resp) },
		)
	}

	vars.Watch(Name+"_callback", func() interface{} {
		var dt []string
		callbacks.Each(func(key string, _ interface{}) { dt = append(dt, key) })
		return dt
	})

	vars.Watch(Name, func() interface{} {
		var dt = make(map[string]string)
		for name, f := range factories {
			dt[name] = stack.Func(f)
		}
		return dt
	})

	return
}

func onWatch(name string, resp *Response) {
	var logs = zap.S().Named("watcher").With(zap.String("project", name))

	defer xerror.Resp(func(err xerror.XErr) {
		logs.Errorw(
			"watcher callback error",
			zap.Any("resp", resp),
			zap.Any("err", err),
			zap.Any("err_msg", err.Error()),
		)
	})

	// value为空就skip
	if cfg.SkipNull && len(bytes.TrimSpace(resp.Value)) == 0 {
		return
	}

	var key = KeyToDot(resp.Key)

	logs.Infow(
		"watcher callback",
		zap.Any("key", key),
		zap.Any("event", resp.Event.String()),
		zap.Any("version", resp.Version),
		zap.Any("value", string(resp.Value)),
	)

	// 把数据设置到全局配置管理中
	// value都必须是kv类型的数据
	var dt = make(map[string]interface{})
	xerror.PanicF(types.Decode(resp.Value, &dt), "value都必须是kv类型的数据, key=>%s, value=>%s", resp.Key, resp.Value)

	resp.OnPut(func() {
		if name == runenv.Project {
			// 本项目配置, 去掉本项目前缀
			config.GetCfg().Set(trimProject(key), dt)
		} else {
			config.GetCfg().Set(key, dt)
		}
	})

	resp.OnDelete(func() {
		if name == runenv.Project {
			// 本项目配置, 去掉本项目前缀
			config.GetCfg().Set(trimProject(key), nil)
		} else {
			config.GetCfg().Set(key, nil)
		}
	})

	// 过滤掉Exclude中的project, 不执行callback
	if strutil.Contains(cfg.Exclude, name) {
		return
	}

	// 以name为前缀的所有的callbacks
	callbacks.Each(func(k string, plg interface{}) {
		defer xerror.Resp(func(err xerror.XErr) {
			logs.Error("watch callback handle error",
				zap.String("watch-key", k),
				zap.Any("resp", resp),
				zap.Any("err", err),
				zap.Any("err_msg", err.Error()),
				zap.Any("stack", stack.Func(plg)),
			)
		})

		// 检查是否是以key为前缀, `.`是连接符和分隔符
		if !strings.HasPrefix(key+".", k+".") {
			return
		}

		// 执行watch callback
		var prefix = strings.Trim(strings.TrimPrefix(key, k), ".")
		xerror.Panic(plg.(func(name string, r *types.WatchResp) error)(prefix, resp))
	})
}
