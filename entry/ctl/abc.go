package ctl

import "github.com/pubgo/lug/entry"

type opts struct{}
type Opt func(opts *opts)
type Entry interface {
	entry.Entry
	Register(fn func(), opts ...Opt)
}
