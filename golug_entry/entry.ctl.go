package golug_entry

type CtlOptions struct{}
type CtlOption func(opts *CtlOptions)
type CtlEntry interface {
	Entry
	Register(fn func(), opts ...CtlOption)
}
