package ginEntry

import (
	"github.com/pubgo/lava/entry"
)

type options struct{}
type Opt func(opts *options)
type Entry interface {
	entry.Entry
	Register(srv interface{}, opts ...Opt)
}
