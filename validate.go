package lava

import "github.com/go-playground/validator/v10"

type Validator interface {
	Validate(v *validator.Validate) error
}

// Initializer ...
type Initializer interface {
	Initialize()
}
