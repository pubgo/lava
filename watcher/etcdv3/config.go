package etcdv3

import (
	"strings"
)

const Name = "etcd"

type Cfg struct {
	Prefix string `json:"prefix"`
	Name   string `json:"name"`
}

func handleKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.Trim(key, "/")
	return "/" + key + "/"
}
