package resource

type Resource interface {
	Close() error
	UpdateResObj(val interface{})
	Kind() string
}

type Base struct{}

func (Base) Close() error                   { return nil }
func (s Base) UpdateResObj(val interface{}) {}
func (s Base) Kind() string                 { return "" }
