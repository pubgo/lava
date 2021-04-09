package ctlEntry

import "github.com/pubgo/lug/entry"

type Options struct{}
type Option func(opts *Options)
type Entry interface {
	entry.Entry
	Register(fn func(), opts ...Option)
}
