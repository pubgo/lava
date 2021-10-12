package types

import (
	"github.com/hashicorp/hcl"
	"github.com/pelletier/go-toml"
	"github.com/pubgo/x/jsonx"
	"github.com/pubgo/xerror"
	"gopkg.in/yaml.v2"

	"github.com/pubgo/lava/errors"
)

func Decode(data []byte, c interface{}) (err error) {
	defer xerror.RespErr(&err)

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

	return errors.Unknown("config.watcher.decode", "data=>%s, c=>%T", data, c)
}
