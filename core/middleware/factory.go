package middleware

import (
	"github.com/pubgo/xerror"
)

func Register(name string, m Middleware) map[string]Middleware {
	defer xerror.RecoverAndExit()
	xerror.Assert(name == "" || m == nil, "[m,name] should not be null")
	return map[string]Middleware{name: m}
}
