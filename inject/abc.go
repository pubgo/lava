package inject

type Named interface {
	ModuleUniqueName() string
}

func Id(name string) Named {
	return &namedImpl{name: name}
}

type namedImpl struct {
	name string
}

func (t *namedImpl) ModuleUniqueName() string {
	return t.name
}
