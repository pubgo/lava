package resource

type Resource interface {
	Close() error
}

var _ Resource = (*resourceWrap)(nil)

type resourceWrap struct {
	kind string
	srv  Resource
}

func (r resourceWrap) Close() error { return r.srv.Close() }
