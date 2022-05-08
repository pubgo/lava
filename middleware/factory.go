package middleware

import (
	"github.com/pubgo/xerror"
)

var factories = make(map[string]Middleware)

func Get(name string) Middleware  { return factories[name] }
func List() map[string]Middleware { return factories }
func Register(name string, m Middleware) {
	xerror.Assert(name == "" || m == nil, "[m,name] should not be null")
	xerror.Assert(factories[name] != nil, "[name] %s already exists", name)
	factories[name] = m
}
