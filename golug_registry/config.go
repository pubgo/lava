package golug_registry

import (
	"crypto/tls"
	"time"

	"github.com/pubgo/golug/golug_config"
)

var Name = "registry"

type Cfg struct {
	Project   string `json:"project"`
	Driver    string `json:"driver"`
	Name      string `json:"name"`
	Prefix    string
	Addrs     []string
	Timeout   time.Duration
	Secure    bool
	TTL       time.Duration
	TLSConfig *tls.Config
}

func GetCfg() (cfg map[string]Cfg) {
	golug_config.Decode(Name, &cfg)
	return
}

func GetDefaultCfg() Cfg {
	return Cfg{}
}
