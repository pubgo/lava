package golug_entry

type CtlOptions struct{}
type CtlOption func(opts *CtlOptions)
type CtlEntry interface {
	Entry
	Main(fn func(), opts ...CtlOption)
}
