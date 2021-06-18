package watcher

import (
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/pelletier/go-toml"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/encoding"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
	"gopkg.in/yaml.v2"
)

const Name = "watcher"

var cfg = GetDefaultCfg()

type Cfg struct {
	Prefix        string   `json:"prefix"`
	SkipNullValue bool     `json:"skip_null_value"`
	Driver        string   `json:"driver"`
	Projects      []string `json:"projects"`
}

func (cfg Cfg) Build() (_ Watcher, err error) {
	defer xerror.RespErr(&err)

	driver := cfg.Driver
	xerror.Assert(driver == "", "watcher driver is null")
	xerror.Assert(factories[driver] == nil, "watcher driver [%s] not found", driver)

	fc := factories[driver]
	return fc(config.GetCfg().GetStringMap(Name))
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Prefix: "/watcher",
		Driver: "etcd",
	}
}

func trimProject(key string) string {
	return strings.Trim(strings.TrimPrefix(key, runenv.Project), ".")
}

// KeyToDot /projectName/foo/bar -->  projectName.foo.bar
func KeyToDot(prefix ...string) string {
	var p string
	if len(prefix) > 0 {
		p = strings.Join(prefix, ".")
	}

	p = strings.ReplaceAll(strings.ReplaceAll(p, "/", "."), "..", ".")
	p = strings.Trim(p, ".")

	return p
}

func unmarshal(in []byte, c interface{}) (err error) {
	defer func() {
		if err != nil {
			err = xerror.Fmt("Unmarshal Error, encoding: %v\n", encoding.Keys())
		}
	}()

	// "yaml", "yml"
	if err = yaml.Unmarshal(in, &c); err == nil {
		return
	}

	// "json"
	if err = jsonx.Unmarshal(in, &c); err == nil {
		return
	}

	// "hcl"
	if err = hcl.Unmarshal(in, &c); err == nil {
		return
	}

	// "toml"
	if err = toml.Unmarshal(in, &c); err == nil {
		return
	}

	return
}
