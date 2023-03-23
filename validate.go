package lava

type Validator interface {
	Validate() error
}

// Initializer ...
type Initializer interface {
	Initialize()
}
