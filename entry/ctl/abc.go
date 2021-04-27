package ctl

import "github.com/pubgo/lug/entry"

type Opts struct{}
type Opt func(opts *Opts)
type Entry interface {
	entry.Entry
	Register(fn func(), opts ...Opt)
}