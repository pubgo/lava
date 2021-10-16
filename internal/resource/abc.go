package resource

type Resource interface {
	Close() error
	Update(val interface{})
}

var _ Resource = (*resourceWrap)(nil)

type resourceWrap struct {
	Resource
	kind string
}
