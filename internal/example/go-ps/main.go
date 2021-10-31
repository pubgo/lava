package main

import (
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"os"

	"github.com/mitchellh/go-ps"
)

func main() {
	defer xerror.RespExit()

	p, err := ps.FindProcess(os.Getpid())
	xerror.Panic(err)
	xerror.Assert(p == nil, "should have process")

	q.Q(p.Pid(), os.Getpid())
}
