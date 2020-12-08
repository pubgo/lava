package golug_watcher

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/golug_plugin/plugins/golug_etcd"
	"github.com/pubgo/xerror"
)

var watchers []Watcher

func init() {
	xerror.Exit(dix_run.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		golug_config.GetCfg().WatchConfig()

		w := newFileWatcher()
		xerror.Exit(filepath.Walk(filepath.Join(golug_env.Home, "config"), func(path string, info os.FileInfo, err error) error {
			xerror.Panic(err)

			if info.IsDir() {
				return nil
			}

			// 配置文件类型检查
			if !strings.HasSuffix(info.Name(), golug_config.CfgType) {
				return nil
			}

			// 文件名字检查
			if info.Name() == golug_config.CfgName+"."+golug_config.CfgType {
				return nil
			}

			if len(strings.Split(info.Name(), ".")) != 3 {
				xerror.Exit(xerror.Fmt("config name error, %s", path))
			}

			xerror.Panic(w.watcher.Add(path))
			return nil
		}))

		if golug_env.IsDev() || golug_env.IsTest() {
			watchers = append(watchers, w)
		}

		if golug_config.GetCfg().GetBool("watcher.configs.etcd.enabled") {
			name := golug_config.GetCfg().GetString("watcher.configs.etcd.driver")
			watchers = append(watchers, newEtcdWatcher(golug_env.Project, golug_etcd.GetClient(name)))
		}

		Start()
	}))

	xerror.Exit(dix_run.WithAfterStop(func(ctx *dix_run.AfterStopCtx) { Close() }))
}

func AddWatcher(c Watcher) {
	watchers = append(watchers, c)
}

func getDefault() []Watcher {
	if len(watchers) != 0 {
		return watchers
	}

	xerror.Exit(errors.New("please init Watcher"))
	return nil
}

func Start() {
	for _, w := range getDefault() {
		xerror.ExitF(w.Start(), w.String())
	}
}

func Close() {
	for _, w := range getDefault() {
		xerror.ExitF(w.Close(), w.String())
	}
}

func Watch(name string, h CallBack) {
	for _, w := range getDefault() {
		xerror.ExitF(w.Watch(name, h), "name:%s watcher:%s", name, w.String())
	}
}

func Remove(name string) {
	for _, w := range getDefault() {
		xerror.ExitF(w.Remove(name), "name:%s watcher:%s", name, w.String())
	}
}

func List() []string {
	return getDefault()[0].List()
}
