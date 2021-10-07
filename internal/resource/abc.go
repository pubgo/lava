package resource

type Resource interface {
	Close() error
}

type resourceWrap struct {
	kind string
	srv  Resource
}
