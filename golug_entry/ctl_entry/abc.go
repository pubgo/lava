package ctl_entry

import "github.com/pubgo/golug/golug_entry"

type CtlOptions struct{}
type CtlOption func(opts *CtlOptions)
type CtlEntry interface {
	golug_entry.Entry
	Register(fn func(), opts ...CtlOption)
}
