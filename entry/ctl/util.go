package ctl

import (
	"github.com/pubgo/lug/consts"

	"github.com/pubgo/xerror"
)

func register(t *ctlEntry, run Handler, optList ...Opt) options {
	var opts = options{handler: run}
	for i := range optList {
		optList[i](&opts)
	}

	if opts.Name == "" {
		opts.Name = consts.Default
	}

	xerror.Assert(t.handlers[opts.Name].handler != nil, "handler [%s] already exists", opts.Name)
	return opts
}
