package etcdv3

import (
	"strings"
)

const Name = "etcdv3"

type Cfg struct {
	Prefix string `json:"prefix"`
	Name   string `json:"name"`
}

// handleKey key=>/key/
func handleKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.Trim(key, "/")
	return "/" + key + "/"
}
