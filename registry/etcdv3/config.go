package etcdv3

import (
	"time"
)

type Cfg struct {
	Prefix   string   `json:"prefix"`
	Driver   string   `json:"driver"`
	Projects []string `json:"projects"`
}

var (
	Name    = "etcdv3"
	prefix  = "/micro-registry"
	timeout = time.Second * 2
)
