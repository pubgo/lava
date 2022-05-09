package inject

type Object interface {
	Name() string
	Type() string
}

type Field interface {
	Tag(name string) string
	Type() string
	Name() string
}
