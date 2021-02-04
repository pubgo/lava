package etcdv3

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
)

// cfgMerge 合并etcd config
func cfgMerge(cfg clientv3.Config) (cfg1 clientv3.Config, err error) {
	cfg1 = DefaultCfg
	if err1 := mergo.Map(&cfg1, cfg, mergo.WithOverride, mergo.WithTypeCheck); err1 != nil {
		err = errors.Wrapf(err1, "[etcd] client config merge error")
	}
	return
}

func retry(c int, fn func() error) (err error) {
	for i := 0; i < c; i++ {
		if err = fn(); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	return
}

// InitFromEnv 从env中获取配置, 并初始化etcd client
func InitFromEnv() error {
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, etcdEnvPrefix) {
			continue
		}

		envs := strings.SplitN(env, "=", 2)
		if len(envs) != 2 {
			continue
		}

		// 从环境变量获取etcd配置
		name := strings.TrimPrefix(envs[0], etcdEnvPrefix)
		if name == "" {
			return fmt.Errorf("[etcd] name is null")
		}

		cfg, err := cfgFromURL(envs[1])
		if err != nil {
			return errors.Wrapf(err, "[etcd] parse etcd config from url error, url: %s", envs[1])
		}

		name = strings.ToLower(name)
		if err := InitClient(name, cfg); err != nil {
			return errors.Wrapf(err, "[etcd] init client error, name:%s, Config:%#v", name, cfg)
		}
	}
	return nil
}

// cfgFromURL 从url解析出etcd config
// [uri]: etcd://127.0.0.1:2379?timeout=1s&username=hello
func cfgFromURL(uri string) (cfg clientv3.Config, err error) {
	uri = strings.TrimSpace(uri)

	if uri == "" {
		err = fmt.Errorf("[etcd] uri is null")
		return
	}

	u, err := url.Parse(uri)
	if err != nil {
		err = errors.Wrapf(err, "can't parse %q as a URL", uri)
		return
	}

	addr := u.Host
	if addr == "" {
		err = fmt.Errorf("[etcd] url host is null")
		return
	}

	cfg, err = (config{}).fromQuery(u.Query())
	if err != nil {
		return cfg, errors.Wrapf(err, "[etcd] query parse error, query:%#v", u.Query())
	}

	if strings.Contains(addr, ",") {
		cfg.Endpoints = strings.Split(addr, ",")
	} else {
		cfg.Endpoints = append(cfg.Endpoints, addr)
	}

	return cfg, err
}
