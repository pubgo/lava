package golug_watcher_file

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/golug/pkg/golug_utils"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
)

var _ golug_watcher.Watcher = (*fileWatcher)(nil)

func newWatcher(name string) *fileWatcher {
	watcher, err := fsnotify.NewWatcher()
	xerror.Exit(err)
	return &fileWatcher{name: name, watcher: watcher, callback: golug_config.UnMarshal}
}

// fileWatcher ...
type fileWatcher struct {
	name     string
	callback func(path string) map[string]interface{}
	watcher  *fsnotify.Watcher
	cancel   context.CancelFunc
}

func (t *fileWatcher) init() {
	cfgType := golug_config.CfgType
	cfgName := golug_config.CfgName

	xerror.Exit(filepath.Walk(filepath.Join(golug_env.Home, "config"), func(path string, info os.FileInfo, err error) error {
		xerror.Panic(err)

		if info.IsDir() {
			return nil
		}

		// 配置文件类型检查
		if !strings.HasSuffix(info.Name(), cfgType) {
			return nil
		}

		// 文件名字检查
		if info.Name() == cfgName+"."+cfgType {
			return nil
		}

		if len(strings.Split(info.Name(), ".")) != 3 {
			panic(xerror.Fmt("config name error, %s", path))
		}

		xerror.Panic(t.watcher.Add(path))
		return nil
	}))
}

func (t *fileWatcher) String() string {
	return "file"
}

func (t *fileWatcher) Start() (err error) {
	defer xerror.RespErr(&err)

	t.init()

	t.cancel = xprocess.GoLoop(func(ctx context.Context) error {
		select {
		case event, ok := <-t.watcher.Events:
			if !ok {
				return xprocess.Break
			}

			name := filepath.Base(event.Name)
			ns := strings.Split(name, ".")
			if len(ns) != 3 {
				xlog.Errorf("config name format error, name:%s", event.Name)
				return nil
			}

			fn := golug_watcher.GetCallBack(ns[1])
			if fn == nil {
				xlog.Errorf("[CallBack] is nil, name:%s", event.Name)
				return nil
			}

			op := event.Op
			if op&fsnotify.Write == fsnotify.Write || op&fsnotify.Create == fsnotify.Create {
				val := []byte(golug_utils.Marshal(t.callback(event.Name)))
				resp := &golug_watcher.Response{Key: ns[1], Value: val, Event: "PUT"}
				if err := fn(resp); err != nil {
					xlog.Errorf("%s handle error", xlog.Any("err", err))
					return nil
				}
			}
		case err, ok := <-t.watcher.Errors:
			if !ok {
				return xprocess.Break
			}

			if err != nil {
				xlog.Errorf("watcher error", xlog.Any("err", err))
			}
		}
		return nil
	})

	return nil
}

func (t *fileWatcher) Close() error {
	_ = t.watcher.Close()
	t.cancel()
	return nil
}
