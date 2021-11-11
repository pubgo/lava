package vm

type VM interface {
	Name() string
	Init() error
	Import()
}
