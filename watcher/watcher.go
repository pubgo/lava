package watcher

import (
	"bytes"
	"context"
	"strings"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/vars"

	"github.com/hashicorp/hcl"
	"github.com/pelletier/go-toml"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var defaultWatcher Watcher = &nullWatcher{}
var callbacks typex.Map

func Init() (err error) {
	defer xerror.RespErr(&err)

	if !config.GetCfg().Decode(Name, &cfg) {
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

		// 获取远程配置
		xerror.Panic(defaultWatcher.GetCallback(context.Background(), name, func(resp *Response) { onWatch(name, resp) }))

		// 配置远程watch
		defaultWatcher.WatchCallback(context.Background(), name, func(resp *Response) { onWatch(name, resp) })
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

func Watch(name string, plg plugin.Plugin) {
	name = KeyToDot(name)
	xerror.Assert(name == "" || plg == nil, "[name, plugin] should not be null")
	xerror.Assert(callbacks.Has(name), "plugin %s already exists", name)
	callbacks.Set(name, plg)
}

func onWatch(name string, resp *Response) {
	defer xerror.Resp(func(err xerror.XErr) {
		xlog.Error("onWatch error",
			zap.Any("err", err),
			zap.Any("resp", resp))
	})

	// value为空就skip
	if cfg.SkipNull && len(bytes.TrimSpace(resp.Value)) == 0 {
		return
	}

	var key = KeyToDot(resp.Key)

	// 把数据设置到全局配置管理中
	// value都必须是kv类型的数据
	var dt = make(map[string]interface{})
	xerror.PanicF(Decode(resp.Value, &dt), "value都必须是kv类型的数据, key:%s, value:%s", resp.Key, resp.Value)

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

	// 过滤掉Exclude中的project, 不进行plugin执行
	for _, exc := range cfg.Exclude {
		if strutil.Contains(cfg.Projects, exc) {
			return
		}
	}

	// 以name为前缀的所有的callbacks
	callbacks.Each(func(k string, plg interface{}) {
		defer xerror.Resp(func(err xerror.XErr) {
			xlog.Error("watch callback handle error",
				zap.String("watch_key", k),
				zap.Any("err", err))
		})

		// 检查是否是以key为前缀, `.`是连接符和分隔符
		if !strings.HasPrefix(key+".", k+".") {
			return
		}

		// 执行watch callback
		var name = strings.Trim(strings.TrimPrefix(key, k), ".")
		xerror.PanicF(plg.(plugin.Plugin).Watch(name, resp), "event: %#v", *resp)
	})
}

func Decode(data []byte, c interface{}) (err error) {
	defer func() {
		if err != nil {
			err = xerror.Fmt("Unmarshal Error, encoding\n")
		}
	}()

	// "yaml", "yml"
	if err = yaml.Unmarshal(data, &c); err == nil {
		return
	}

	// "json"
	if err = jsonx.Unmarshal(data, &c); err == nil {
		return
	}

	// "hcl"
	if err = hcl.Unmarshal(data, &c); err == nil {
		return
	}

	// "toml"
	if err = toml.Unmarshal(data, &c); err == nil {
		return
	}

	return
}
