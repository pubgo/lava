package golug_cron

import (
	"errors"

	"github.com/pubgo/xerror"
	"github.com/robfig/cron/v3"
)

var defaultCron = New(cron.WithSeconds(), cron.WithChain(cron.Recover(cron.DefaultLogger)))

func SetDefault(c *cronManager) {
	defaultCron = c
}

func getDefault() *cronManager {
	if defaultCron != nil {
		return defaultCron
	}

	xerror.Exit(errors.New("please init cronManager"))
	return nil
}

func Start() error {
	return getDefault().Start()
}

func Stop() error {
	return getDefault().Stop()
}

func List() map[string]cron.Entry {
	return getDefault().List()
}

func Get(name string) cron.Entry {
	return getDefault().Get(name)
}

func Add(name string, spec string, cmd CallBack) error {
	return getDefault().Add(name, spec, cmd)
}

func Remove(name string) error {
	return getDefault().Remove(name)
}
