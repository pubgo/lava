package golug_watcher

import (
	"context"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xprocess"
)

var _ Watcher = (*fileWatcher)(nil)

func newFileWatcher() *fileWatcher {
	watcher, err := fsnotify.NewWatcher()
	xerror.Exit(err)
	return &fileWatcher{watcher: watcher}
}

// fileWatcher ...
type fileWatcher struct {
	data    sync.Map
	watcher *fsnotify.Watcher
	cancel  context.CancelFunc
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

			fn, ok := t.data.Load(ns[1])
			if !ok {
				return
			}

			val := golug_config.UnMarshal(golug_config.GetCfg().Viper, event.Name)

			op := event.Op
			if op&fsnotify.Write == fsnotify.Write || op&fsnotify.Create == fsnotify.Create {
				resp := &Response{Key: ns[1], Value: []byte(golug_util.Marshal(val)), Event: "PUT"}
				xerror.Panic(fn.(CallBack)(resp))
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

func (t *fileWatcher) Watch(name string, h CallBack) (err error) {
	defer xerror.RespErr(&err)

	if h == nil {
		return xerror.New("[CallBack] is nil")
	}

	t.data.Store(name, h)
	return nil
}

func (t *fileWatcher) List() []string {
	var data []string
	t.data.Range(func(key, _ interface{}) bool { data = append(data, key.(string)); return true })
	return data
}

func (t *fileWatcher) Remove(name string) (err error) {
	t.data.Delete(name)
	return nil
}

func handlePath(name *string) (err error) {
	defer xerror.RespErr(&err)

	nm := filepath.Clean(*name)
	nm = xerror.PanicStr(filepath.EvalSymlinks(nm))
	*name = xerror.PanicStr(filepath.Abs(nm))

	return nil
}
