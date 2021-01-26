package golug_watcher_file

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/golug/pkg/golug_utils"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
)

var _ golug_watcher.Watcher = (*fileWatcher)(nil)

func newWatcher(name string) *fileWatcher {
	watcher, err := fsnotify.NewWatcher()
	xerror.ExitF(err, "file watcher init, name: %s", name)
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

	xerror.Exit(filepath.Walk(filepath.Join(golug_app.Home, "config"), func(path string, info os.FileInfo, err error) error {
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

func (t *fileWatcher) Name() string { return Name }

func (t *fileWatcher) Start() (err error) {
	defer xerror.RespErr(&err)

	t.init()

	t.cancel = xprocess.GoLoop(func(ctx context.Context) {
		select {
		case event, ok := <-t.watcher.Events:
			if !ok {
				xprocess.Break()
			}

			name := filepath.Base(event.Name)
			ns := strings.Split(name, ".")
			if len(ns) != 3 {
				xlog.Errorf("config name format error, name:%s", event.Name)
				return
			}

			fn := golug_watcher.GetWatch(ns[1])
			if fn == nil {
				xlog.Errorf("watcher callback is nil, name:%s", event.Name)
				return
			}

			op := event.Op
			if op&fsnotify.Write == fsnotify.Write || op&fsnotify.Create == fsnotify.Create {
				val := []byte(golug_utils.Marshal(t.callback(event.Name)))
				resp := &golug_watcher.Response{Key: ns[1], Value: val, Event: "PUT"}
				if err := fn(resp); err != nil {
					xlog.Error("watcher handle error", xlog.Any("err", err))
					return
				}
			}
		case err, ok := <-t.watcher.Errors:
			if !ok {
				xprocess.Break()
			}

			if err != nil {
				xlog.Error("watcher error", xlog.Any("err", err))
			}
		}
		return
	})

	return nil
}

func (t *fileWatcher) Close() error {
	return xerror.Try(func() {
		t.cancel()
		xerror.Panic(t.watcher.Close())
	})
}
