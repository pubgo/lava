package gnet

type Builder struct {
	srv interface{}
}

func (t *Builder) Get() interface{} {
	if t.srv == nil {
		panic("please init gnet")
	}

	return t.srv
}

func (t *Builder) Build(cfg Cfg) error {
	return nil
}

func New() Builder {
	return Builder{}
}
