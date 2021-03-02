package golug_ctl

import "github.com/pubgo/golug/entry"

type Options struct{}
type Option func(opts *Options)
type Entry interface {
	entry.Entry
	Register(fn func(), opts ...Option)
}
