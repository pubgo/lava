package file

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/golug/pkg/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
)

var _ golug_watcher.Watcher = (*fileWatcher)(nil)

func NewWatcher(cfgType, cfgName string, callback func(path string) map[string]interface{}) *fileWatcher {
	watcher, err := fsnotify.NewWatcher()
	xerror.Exit(err)

	if callback == nil {
		panic(xerror.New("[callback] is nil"))
	}

	w := &fileWatcher{watcher: watcher, callback: callback}

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
			xerror.Exit(xerror.Fmt("config name error, %s", path))
		}

		xerror.Panic(w.watcher.Add(path))
		return nil
	}))

	return w
}

// fileWatcher ...
type fileWatcher struct {
	callback func(path string) map[string]interface{}
	watcher  *fsnotify.Watcher
	cancel   context.CancelFunc
}

func (t *fileWatcher) String() string {
	return "file"
}

func (t *fileWatcher) Start() (err error) {
	defer xerror.RespErr(&err)
	t.cancel = xprocess.GoLoop(func(ctx context.Context) {
		select {
		case event, ok := <-t.watcher.Events:
			if !ok {
				xerror.Done()
			}

			name := filepath.Base(event.Name)
			ns := strings.Split(name, ".")
			if len(ns) != 3 {
				xlog.Warnf("config name format error, name:%s", event.Name)
				return
			}

			fn := golug_watcher.GetCallBack(ns[1])
			if fn == nil {
				return
			}

			op := event.Op
			if op&fsnotify.Write == fsnotify.Write || op&fsnotify.Create == fsnotify.Create {
				resp := &golug_watcher.Response{Key: ns[1], Value: []byte(golug_util.Marshal(t.callback(event.Name))), Event: "PUT"}
				xerror.Panic(fn(resp))
			}
		case err, ok := <-t.watcher.Errors:
			if !ok {
				return
			}

			if err == nil {
				return
			}

			xlog.Error("fileWatcher.Start handle error", xlog.Any("err", err))
		}
	})

	return nil
}

func (t *fileWatcher) Close() error {
	_ = t.watcher.Close()
	t.cancel()
	return nil
}
