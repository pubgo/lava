package ctl_entry

import "github.com/pubgo/golug/golug_entry"

type Options struct{}
type Option func(opts *Options)
type Entry interface {
	golug_entry.Entry
	Register(fn func(), opts ...Option)
}
