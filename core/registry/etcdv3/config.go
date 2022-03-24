package etcdv3

import (
	"time"
)

var (
	Name    = "etcdv3"
	prefix  = "/micro-registry"
	timeout = time.Second * 2
)

type Cfg struct {
	Prefix   string   `json:"prefix"`
	Name     string   `json:"name"`
	Projects []string `json:"projects"`
}
